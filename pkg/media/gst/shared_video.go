package gst

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/media"
)

var defaultCaptureRegistry = newCaptureRegistry()

type captureRegistry struct {
	mu      sync.Mutex
	sources map[string]*sharedVideoSource
}

func newCaptureRegistry() *captureRegistry {
	return &captureRegistry{sources: map[string]*sharedVideoSource{}}
}

func (r *captureRegistry) acquireVideoSource(ctx context.Context, source dsl.EffectiveVideoSource) (*sharedVideoSource, error) {
	if err := initGStreamer(); err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}

	resolvedSource, err := resolveWindowSource(ctx, source)
	if err != nil {
		return nil, err
	}
	signature := computeSharedVideoSourceSignature(resolvedSource)

	r.mu.Lock()
	defer r.mu.Unlock()

	if existing := r.sources[signature]; existing != nil && !existing.isClosed() {
		existing.refCount++
		log.Info().
			Str("event", "capture.gst.shared.acquire.reused").
			Str("signature", signature).
			Str("source_id", resolvedSource.ID).
			Str("source_type", resolvedSource.Type).
			Int("ref_count", existing.refCount).
			Msg("reused shared gstreamer video source")
		return existing, nil
	}

	shared, err := newSharedVideoSource(resolvedSource, signature, r)
	if err != nil {
		return nil, err
	}
	shared.refCount = 1
	r.sources[signature] = shared
	log.Info().
		Str("event", "capture.gst.shared.acquire.created").
		Str("signature", signature).
		Str("source_id", resolvedSource.ID).
		Str("source_type", resolvedSource.Type).
		Msg("created shared gstreamer video source")
	return shared, nil
}

func (r *captureRegistry) remove(signature string, shared *sharedVideoSource) {
	if r == nil || signature == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if current := r.sources[signature]; current == shared {
		delete(r.sources, signature)
	}
}

type sharedVideoSource struct {
	registry  *captureRegistry
	signature string
	source    dsl.EffectiveVideoSource
	pipeline  *gst.Pipeline
	tee       *gst.Element

	mu        sync.Mutex
	watch     *busWatch
	resultCh  chan error
	stopWait  chan struct{}
	started   bool
	closed    bool
	refCount  int
	consumers map[string]*sharedPreviewConsumer
	counter   atomic.Uint64
}

func newSharedVideoSource(source dsl.EffectiveVideoSource, signature string, registry *captureRegistry) (*sharedVideoSource, error) {
	pipeline, tee, err := buildSharedVideoSourcePipeline(source)
	if err != nil {
		return nil, err
	}
	return &sharedVideoSource{
		registry:  registry,
		signature: signature,
		source:    source,
		pipeline:  pipeline,
		tee:       tee,
		stopWait:  make(chan struct{}),
		consumers: map[string]*sharedPreviewConsumer{},
	}, nil
}

func (s *sharedVideoSource) isClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

func (s *sharedVideoSource) releaseReference() {
	if s == nil {
		return
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	if s.refCount > 0 {
		s.refCount--
	}
	shouldShutdown := s.refCount == 0 && len(s.consumers) == 0
	s.mu.Unlock()
	if shouldShutdown {
		s.shutdown(nil)
	}
}

func (s *sharedVideoSource) attachPreviewConsumer(opts media.PreviewOptions) (*sharedPreviewConsumer, error) {
	if s == nil {
		return nil, errors.New("shared video source is nil")
	}
	if opts.OnFrame == nil {
		opts.OnFrame = func([]byte) {}
	}
	if opts.OnLog == nil {
		opts.OnLog = func(string, string) {}
	}

	consumer, err := buildSharedPreviewConsumer(s, opts)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, errors.New("shared video source already closed")
	}
	if err := s.pipeline.AddMany(consumer.elements...); err != nil {
		s.mu.Unlock()
		return nil, fmt.Errorf("add preview branch to pipeline: %w", err)
	}
	if err := linkElements(consumer.elements...); err != nil {
		s.mu.Unlock()
		_ = s.pipeline.RemoveMany(consumer.elements...)
		return nil, err
	}
	teePad := s.tee.GetRequestPad("src_%u")
	if teePad == nil {
		s.mu.Unlock()
		_ = s.pipeline.RemoveMany(consumer.elements...)
		return nil, errors.New("request tee src pad for preview branch")
	}
	queueSinkPad := consumer.queue.GetStaticPad("sink")
	if queueSinkPad == nil {
		s.tee.ReleaseRequestPad(teePad)
		s.mu.Unlock()
		_ = s.pipeline.RemoveMany(consumer.elements...)
		return nil, errors.New("preview queue missing sink pad")
	}
	if ret := teePad.Link(queueSinkPad); ret != gst.PadLinkOK {
		s.tee.ReleaseRequestPad(teePad)
		s.mu.Unlock()
		_ = s.pipeline.RemoveMany(consumer.elements...)
		return nil, fmt.Errorf("link tee to preview queue: %s", ret.String())
	}
	consumer.teePad = teePad

	if !s.started {
		if err := s.startLocked(); err != nil {
			teePad.Unlink(queueSinkPad)
			s.tee.ReleaseRequestPad(teePad)
			for _, element := range consumer.elements {
				_ = element.SetState(gst.StateNull)
			}
			_ = s.pipeline.RemoveMany(consumer.elements...)
			s.mu.Unlock()
			return nil, err
		}
	} else {
		for _, element := range consumer.elements {
			if !element.SyncStateWithParent() {
				teePad.Unlink(queueSinkPad)
				s.tee.ReleaseRequestPad(teePad)
				for _, added := range consumer.elements {
					_ = added.SetState(gst.StateNull)
				}
				_ = s.pipeline.RemoveMany(consumer.elements...)
				s.mu.Unlock()
				return nil, fmt.Errorf("sync preview branch element %s with parent", element.GetName())
			}
		}
	}

	s.consumers[consumer.id] = consumer
	log.Info().
		Str("event", "capture.gst.shared.preview.attach").
		Str("signature", s.signature).
		Str("consumer_id", consumer.id).
		Str("source_id", s.source.ID).
		Int("consumer_count", len(s.consumers)).
		Int("ref_count", s.refCount).
		Msg("attached shared preview consumer")
	s.mu.Unlock()
	return consumer, nil
}

func (s *sharedVideoSource) startLocked() error {
	bus := s.pipeline.GetPipelineBus()
	resultCh := make(chan error, 1)
	watch, err := startBusWatch(bus, func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageError:
			gstErr := msg.ParseError()
			select {
			case resultCh <- fmt.Errorf("shared video source pipeline error: %w", gstErr):
			default:
			}
			return false
		case gst.MessageEOS:
			select {
			case resultCh <- nil:
			default:
			}
			return false
		default:
			return true
		}
	})
	if err != nil {
		return err
	}
	if err := s.pipeline.SetState(gst.StatePlaying); err != nil {
		watch.Stop()
		return fmt.Errorf("start shared video source pipeline: %w", err)
	}
	s.watch = watch
	s.resultCh = resultCh
	s.started = true
	go s.wait()
	log.Info().
		Str("event", "capture.gst.shared.start").
		Str("signature", s.signature).
		Str("source_id", s.source.ID).
		Str("source_type", s.source.Type).
		Msg("started shared gstreamer video source pipeline")
	return nil
}

func (s *sharedVideoSource) wait() {
	select {
	case err := <-s.resultCh:
		s.failAll(err)
	case <-s.stopWait:
		return
	}
}

func (s *sharedVideoSource) failAll(err error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	consumers := make([]*sharedPreviewConsumer, 0, len(s.consumers))
	for _, consumer := range s.consumers {
		consumers = append(consumers, consumer)
	}
	s.consumers = map[string]*sharedPreviewConsumer{}
	s.refCount = 0
	signalStop := s.stopWait
	watch := s.watch
	pipeline := s.pipeline
	s.mu.Unlock()

	closeSharedStop(signalStop)
	if s.registry != nil {
		s.registry.remove(s.signature, s)
	}
	for _, consumer := range consumers {
		consumer.finish(err)
	}
	if watch != nil {
		watch.Stop()
	}
	if pipeline != nil {
		_ = pipeline.BlockSetState(gst.StateNull)
	}
	log.Warn().
		Str("event", "capture.gst.shared.fail_all").
		Str("signature", s.signature).
		Str("source_id", s.source.ID).
		Err(err).
		Msg("shared gstreamer video source terminated")
}

func (s *sharedVideoSource) detachPreviewConsumer(consumerID string, err error) {
	if s == nil || consumerID == "" {
		return
	}

	s.mu.Lock()
	consumer, ok := s.consumers[consumerID]
	if !ok {
		s.mu.Unlock()
		return
	}
	delete(s.consumers, consumerID)
	if s.refCount > 0 {
		s.refCount--
	}
	remainingConsumers := len(s.consumers)
	shouldShutdown := !s.closed && s.refCount == 0 && remainingConsumers == 0
	s.mu.Unlock()

	s.teardownPreviewConsumer(consumer, err)
	log.Info().
		Str("event", "capture.gst.shared.preview.detach").
		Str("signature", s.signature).
		Str("consumer_id", consumerID).
		Str("source_id", s.source.ID).
		Int("consumer_count", remainingConsumers).
		Msg("detached shared preview consumer")
	if shouldShutdown {
		s.shutdown(nil)
	}
}

func (s *sharedVideoSource) teardownPreviewConsumer(consumer *sharedPreviewConsumer, err error) {
	if s == nil || consumer == nil {
		return
	}
	if consumer.teePad != nil && consumer.queue != nil {
		if queueSinkPad := consumer.queue.GetStaticPad("sink"); queueSinkPad != nil {
			consumer.teePad.Unlink(queueSinkPad)
		}
		s.tee.ReleaseRequestPad(consumer.teePad)
	}
	for _, element := range consumer.elements {
		_ = element.SetState(gst.StateNull)
	}
	if s.pipeline != nil {
		_ = s.pipeline.RemoveMany(consumer.elements...)
	}
	consumer.finish(err)
}

func (s *sharedVideoSource) shutdown(err error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	consumers := make([]*sharedPreviewConsumer, 0, len(s.consumers))
	for _, consumer := range s.consumers {
		consumers = append(consumers, consumer)
	}
	s.consumers = map[string]*sharedPreviewConsumer{}
	s.refCount = 0
	signalStop := s.stopWait
	watch := s.watch
	pipeline := s.pipeline
	s.mu.Unlock()

	closeSharedStop(signalStop)
	if s.registry != nil {
		s.registry.remove(s.signature, s)
	}
	for _, consumer := range consumers {
		s.teardownPreviewConsumer(consumer, err)
	}
	if watch != nil {
		watch.Stop()
	}
	if pipeline != nil {
		_ = pipeline.BlockSetState(gst.StateNull)
	}
	log.Info().
		Str("event", "capture.gst.shared.shutdown").
		Str("signature", s.signature).
		Str("source_id", s.source.ID).
		Msg("shutdown shared gstreamer video source")
}

type sharedPreviewConsumer struct {
	id       string
	source   *sharedVideoSource
	queue    *gst.Element
	elements []*gst.Element
	sink     *app.Sink
	teePad   *gst.Pad
	onLog    func(string, string)

	done      chan struct{}
	closeOnce sync.Once

	mu          sync.RWMutex
	latestFrame []byte
	waitErr     error
}

func buildSharedPreviewConsumer(source *sharedVideoSource, opts media.PreviewOptions) (*sharedPreviewConsumer, error) {
	consumerID := fmt.Sprintf("preview-%s-%d", trimSharedSignature(source.signature), source.counter.Add(1))
	queue, err := gst.NewElementWithName("queue", consumerID+"-queue")
	if err != nil {
		return nil, fmt.Errorf("create preview queue: %w", err)
	}
	queue.Set("leaky", 2)
	queue.Set("max-size-buffers", 2)

	videoscale, err := gst.NewElementWithName("videoscale", consumerID+"-videoscale")
	if err != nil {
		return nil, fmt.Errorf("create preview videoscale: %w", err)
	}
	capsScale, err := newCapsFilter("video/x-raw,width=640")
	if err != nil {
		return nil, err
	}
	capsScale.SetProperty("name", consumerID+"-scale-caps")
	videorate, err := gst.NewElementWithName("videorate", consumerID+"-videorate")
	if err != nil {
		return nil, fmt.Errorf("create preview videorate: %w", err)
	}
	capsRate, err := newCapsFilter("video/x-raw,framerate=5/1")
	if err != nil {
		return nil, err
	}
	capsRate.SetProperty("name", consumerID+"-rate-caps")
	jpegenc, err := gst.NewElementWithName("jpegenc", consumerID+"-jpegenc")
	if err != nil {
		return nil, fmt.Errorf("create preview jpegenc: %w", err)
	}
	jpegenc.Set("quality", 50)
	sink, err := app.NewAppSink()
	if err != nil {
		return nil, fmt.Errorf("create preview appsink: %w", err)
	}
	sink.SetProperty("name", consumerID+"-appsink")
	sink.SetCaps(gst.NewCapsFromString("image/jpeg"))
	sink.SetProperty("max-buffers", 2)
	sink.SetProperty("drop", true)

	consumer := &sharedPreviewConsumer{
		id:       consumerID,
		source:   source,
		queue:    queue,
		elements: []*gst.Element{queue, videoscale, capsScale, videorate, capsRate, jpegenc, sink.Element},
		sink:     sink,
		onLog:    opts.OnLog,
		done:     make(chan struct{}),
	}

	sink.SetCallbacks(&app.SinkCallbacks{
		NewSampleFunc: func(s *app.Sink) gst.FlowReturn {
			sample := s.PullSample()
			if sample == nil {
				return gst.FlowEOS
			}
			buffer := sample.GetBuffer()
			if buffer == nil {
				return gst.FlowError
			}
			mapInfo := buffer.Map(gst.MapRead)
			if mapInfo == nil {
				return gst.FlowError
			}
			defer buffer.Unmap()

			frame := mapInfo.Bytes()
			consumer.setLatestFrame(frame)
			opts.OnFrame(frame)
			return gst.FlowOK
		},
	})

	return consumer, nil
}

func (c *sharedPreviewConsumer) finish(err error) {
	if c == nil {
		return
	}
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.waitErr = err
		c.mu.Unlock()
		if err != nil && c.onLog != nil {
			c.onLog("stderr", err.Error())
		}
		close(c.done)
	})
}

func (c *sharedPreviewConsumer) LatestFrame() ([]byte, error) {
	if c == nil {
		return nil, errors.New("shared preview consumer is nil")
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.latestFrame) == 0 {
		return nil, errors.New("preview frame not available yet")
	}
	return append([]byte(nil), c.latestFrame...), nil
}

func (c *sharedPreviewConsumer) WaitErr() error {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.waitErr
}

func (c *sharedPreviewConsumer) setLatestFrame(frame []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.latestFrame = append([]byte(nil), frame...)
}

func buildSharedVideoSourcePipeline(source dsl.EffectiveVideoSource) (*gst.Pipeline, *gst.Element, error) {
	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return nil, nil, fmt.Errorf("create shared video pipeline: %w", err)
	}
	elements, tee, err := buildSharedVideoSourceElements(source)
	if err != nil {
		return nil, nil, err
	}
	if err := pipeline.AddMany(elements...); err != nil {
		return nil, nil, fmt.Errorf("add shared video source elements: %w", err)
	}
	if err := linkElements(elements...); err != nil {
		_ = pipeline.BlockSetState(gst.StateNull)
		return nil, nil, err
	}
	return pipeline, tee, nil
}

func buildSharedVideoSourceElements(source dsl.EffectiveVideoSource) ([]*gst.Element, *gst.Element, error) {
	elements := []*gst.Element{}
	sourceElements, err := buildSourceElements(source)
	if err != nil {
		return nil, nil, err
	}
	elements = append(elements, sourceElements...)

	videoconvert, err := gst.NewElement("videoconvert")
	if err != nil {
		return nil, nil, fmt.Errorf("create shared videoconvert: %w", err)
	}
	elements = append(elements, videoconvert)

	if source.Type == "camera" && boolValue(source.Capture.Mirror, false) {
		videoflip, err := gst.NewElement("videoflip")
		if err != nil {
			return nil, nil, fmt.Errorf("create shared videoflip: %w", err)
		}
		videoflip.Set("method", "horizontal-flip")
		elements = append(elements, videoflip)
	}

	tee, err := gst.NewElement("tee")
	if err != nil {
		return nil, nil, fmt.Errorf("create shared tee: %w", err)
	}
	tee.Set("allow-not-linked", true)
	elements = append(elements, tee)
	return elements, tee, nil
}

func computeSharedVideoSourceSignature(source dsl.EffectiveVideoSource) string {
	parts := []string{
		"type=" + strings.TrimSpace(source.Type),
		"display=" + strings.TrimSpace(source.Target.Display),
		"device=" + strings.TrimSpace(source.Target.Device),
		"window=" + strings.TrimSpace(source.Target.WindowID),
		"size=" + strings.TrimSpace(source.Capture.Size),
		"cursor=" + strconv.FormatBool(boolValue(source.Capture.Cursor, true)),
		"mirror=" + strconv.FormatBool(boolValue(source.Capture.Mirror, false)),
	}
	if source.Target.Rect != nil {
		rect := source.Target.Rect
		parts = append(parts,
			fmt.Sprintf("rect=%d,%d,%d,%d", rect.X, rect.Y, rect.W, rect.H),
		)
	}
	return strings.Join(parts, "|")
}

func trimSharedSignature(signature string) string {
	signature = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		default:
			return '-'
		}
	}, signature)
	if len(signature) <= 24 {
		return signature
	}
	return signature[:24]
}

func closeSharedStop(ch chan struct{}) {
	if ch == nil {
		return
	}
	defer func() {
		_ = recover()
	}()
	close(ch)
}
