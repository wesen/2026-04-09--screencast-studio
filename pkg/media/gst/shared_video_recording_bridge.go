package gst

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	appmetrics "github.com/wesen/2026-04-09--screencast-studio/pkg/metrics"
)

type ExperimentalSharedVideoRecorderOptions struct {
	OutputPath string
	Container  string
	FPS        int
	OnLog      func(string, string)
}

var (
	sharedBridgeRecorderSamplesReceived = appmetrics.MustRegisterCounterVec(
		"screencast_studio_gst_shared_bridge_recorder_samples_received_total",
		"Total samples received by the shared video recorder bridge from the upstream appsink callback.",
		"source_type",
	)
	sharedBridgeRecorderBuffersCopied = appmetrics.MustRegisterCounterVec(
		"screencast_studio_gst_shared_bridge_recorder_buffers_copied_total",
		"Total GStreamer buffers copied by the shared video recorder bridge before enqueueing.",
		"source_type",
	)
	sharedBridgeRecorderEnqueued = appmetrics.MustRegisterCounterVec(
		"screencast_studio_gst_shared_bridge_recorder_enqueued_total",
		"Total copied buffers successfully enqueued into the shared video recorder bridge async queue.",
		"source_type",
	)
	sharedBridgeRecorderDropped = appmetrics.MustRegisterCounterVec(
		"screencast_studio_gst_shared_bridge_recorder_dropped_total",
		"Total copied buffers dropped because the shared video recorder bridge async queue was full or closed.",
		"source_type",
	)
	sharedBridgeRecorderWorkerHandled = appmetrics.MustRegisterCounterVec(
		"screencast_studio_gst_shared_bridge_recorder_worker_handled_total",
		"Total buffers handled by the shared video recorder bridge worker goroutine.",
		"source_type",
	)
	sharedBridgeRecorderAppsrcPushed = appmetrics.MustRegisterCounterVec(
		"screencast_studio_gst_shared_bridge_recorder_appsrc_pushed_total",
		"Total buffers successfully pushed from the shared video recorder bridge worker into appsrc.",
		"source_type",
	)
)

type ExperimentalSharedVideoRecorder struct {
	cancel context.CancelFunc
	source *sharedVideoSource
	raw    *sharedRawConsumer
	opts   ExperimentalSharedVideoRecorderOptions

	mu         sync.Mutex
	bridge     *bridgeVideoRecorder
	sampleCh   chan *gst.Buffer
	workerDone chan struct{}
	done       chan struct{}
	err        error
	once       sync.Once
	closed     atomic.Bool
	labels     map[string]string
}

func StartExperimentalSharedVideoRecorder(ctx context.Context, source dsl.EffectiveVideoSource, opts ExperimentalSharedVideoRecorderOptions) (*ExperimentalSharedVideoRecorder, error) {
	if err := initGStreamer(); err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(opts.OutputPath) == "" {
		return nil, errors.New("experimental shared video recorder output path is required")
	}
	if strings.TrimSpace(opts.Container) == "" {
		opts.Container = "mp4"
	}
	if opts.FPS <= 0 {
		if source.Capture.FPS > 0 {
			opts.FPS = source.Capture.FPS
		} else {
			opts.FPS = 10
		}
	}
	if opts.OnLog == nil {
		opts.OnLog = func(string, string) {}
	}
	if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0o755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	recorderCtx, cancel := context.WithCancel(ctx)
	shared, err := defaultCaptureRegistry.acquireVideoSource(recorderCtx, source)
	if err != nil {
		cancel()
		return nil, err
	}
	recorder := &ExperimentalSharedVideoRecorder{
		cancel:     cancel,
		source:     shared,
		opts:       opts,
		sampleCh:   make(chan *gst.Buffer, 16),
		workerDone: make(chan struct{}),
		done:       make(chan struct{}),
		labels: map[string]string{
			"source_type": strings.TrimSpace(shared.source.Type),
		},
	}
	if recorder.labels["source_type"] == "" {
		recorder.labels["source_type"] = "unknown"
	}

	width, height := sharedRecordingTargetSize(shared.source)
	raw, err := shared.attachRawConsumer(sharedRawConsumerOptions{
		FPS:    opts.FPS,
		Width:  width,
		Height: height,
		Format: "I420",
		OnLog: func(stream, message string) {
			opts.OnLog(stream, message)
		},
		OnSample: recorder.handleSample,
	})
	if err != nil {
		shared.releaseReference()
		cancel()
		return nil, err
	}
	recorder.raw = raw

	go recorder.runSamplePump()
	go func() {
		select {
		case <-recorderCtx.Done():
			recorder.closeWithError(nil)
		case <-recorder.done:
		}
	}()

	log.Info().
		Str("event", "capture.gst.shared.recording_bridge.start").
		Str("signature", shared.signature).
		Str("source_id", shared.source.ID).
		Str("output_path", opts.OutputPath).
		Int("fps", opts.FPS).
		Msg("started experimental shared video recording bridge")
	return recorder, nil
}

func (r *ExperimentalSharedVideoRecorder) handleSample(sample *gst.Sample) gst.FlowReturn {
	if r == nil || sample == nil {
		return gst.FlowError
	}
	if r.closed.Load() {
		return gst.FlowOK
	}
	sharedBridgeRecorderSamplesReceived.Inc(r.labels)
	buffer := sample.GetBuffer()
	if buffer == nil {
		r.closeWithError(errors.New("shared recorder sample missing buffer"))
		return gst.FlowError
	}

	r.mu.Lock()
	bridge := r.bridge
	if bridge == nil {
		var err error
		bridge, err = newBridgeVideoRecorder(sample.GetCaps(), r.opts)
		if err != nil {
			r.mu.Unlock()
			r.closeWithError(err)
			return gst.FlowError
		}
		r.bridge = bridge
	}
	r.mu.Unlock()

	copyBuffer := buffer.Copy()
	sharedBridgeRecorderBuffersCopied.Inc(r.labels)
	if r.enqueueBuffer(copyBuffer) {
		sharedBridgeRecorderEnqueued.Inc(r.labels)
		return gst.FlowOK
	}
	sharedBridgeRecorderDropped.Inc(r.labels)
	log.Warn().
		Str("event", "capture.gst.shared.bridge_recorder.drop").
		Str("output_path", r.opts.OutputPath).
		Msg("dropping shared bridge recorder sample because the async queue is full or closed")
	return gst.FlowOK
}

func (r *ExperimentalSharedVideoRecorder) enqueueBuffer(buffer *gst.Buffer) (ok bool) {
	if r == nil || buffer == nil || r.closed.Load() {
		return false
	}
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	select {
	case r.sampleCh <- buffer:
		return true
	default:
		return false
	}
}

func (r *ExperimentalSharedVideoRecorder) runSamplePump() {
	defer close(r.workerDone)
	for buffer := range r.sampleCh {
		if buffer == nil {
			continue
		}
		r.mu.Lock()
		bridge := r.bridge
		r.mu.Unlock()
		if bridge == nil {
			continue
		}
		sharedBridgeRecorderWorkerHandled.Inc(r.labels)
		ret, err := bridge.pushBuffer(buffer)
		if err != nil {
			go r.closeWithError(err)
			return
		}
		if ret != gst.FlowOK && ret != gst.FlowEOS {
			go r.closeWithError(fmt.Errorf("bridge recorder push buffer returned %s", ret))
			return
		}
		if ret == gst.FlowOK {
			sharedBridgeRecorderAppsrcPushed.Inc(r.labels)
		}
	}
}

func (r *ExperimentalSharedVideoRecorder) Stop(ctx context.Context) error {
	if r == nil {
		return nil
	}
	r.closeWithError(nil)
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-r.done:
		return r.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *ExperimentalSharedVideoRecorder) Wait() error {
	if r == nil {
		return nil
	}
	<-r.done
	return r.Err()
}

func (r *ExperimentalSharedVideoRecorder) Err() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.err
}

func (r *ExperimentalSharedVideoRecorder) closeWithError(err error) {
	if r == nil {
		return
	}
	r.once.Do(func() {
		r.closed.Store(true)
		if r.cancel != nil {
			r.cancel()
		}
		if r.raw != nil && r.source != nil {
			r.source.detachRawConsumer(r.raw.id, nil)
		}
		close(r.sampleCh)
		<-r.workerDone
		r.mu.Lock()
		bridge := r.bridge
		if r.err == nil {
			r.err = err
		}
		r.mu.Unlock()
		if bridge != nil {
			stopErr := bridge.stop(10 * time.Second)
			if stopErr != nil {
				r.mu.Lock()
				if r.err == nil {
					r.err = stopErr
				}
				r.mu.Unlock()
			}
		}
		close(r.done)
	})
}

type sharedRawConsumer struct {
	id       string
	source   *sharedVideoSource
	queue    *gst.Element
	elements []*gst.Element
	sink     *app.Sink
	teePad   *gst.Pad
	onLog    func(string, string)

	done      chan struct{}
	closeOnce sync.Once
	mu        sync.RWMutex
	waitErr   error
}

type sharedRawConsumerOptions struct {
	FPS      int
	Width    int
	Height   int
	Format   string
	OnSample func(*gst.Sample) gst.FlowReturn
	OnLog    func(string, string)
}

func (s *sharedVideoSource) attachRawConsumer(opts sharedRawConsumerOptions) (*sharedRawConsumer, error) {
	if s == nil {
		return nil, errors.New("shared video source is nil")
	}
	if opts.OnSample == nil {
		return nil, errors.New("shared raw consumer sample callback is required")
	}
	if opts.FPS <= 0 {
		opts.FPS = 10
	}
	if opts.OnLog == nil {
		opts.OnLog = func(string, string) {}
	}
	if strings.TrimSpace(opts.Format) == "" {
		opts.Format = "I420"
	}

	consumer, err := buildSharedRawConsumer(s, opts)
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
		return nil, fmt.Errorf("add raw branch to pipeline: %w", err)
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
		return nil, errors.New("request tee src pad for raw branch")
	}
	queueSinkPad := consumer.queue.GetStaticPad("sink")
	if queueSinkPad == nil {
		s.tee.ReleaseRequestPad(teePad)
		s.mu.Unlock()
		_ = s.pipeline.RemoveMany(consumer.elements...)
		return nil, errors.New("raw queue missing sink pad")
	}
	if ret := teePad.Link(queueSinkPad); ret != gst.PadLinkOK {
		s.tee.ReleaseRequestPad(teePad)
		s.mu.Unlock()
		_ = s.pipeline.RemoveMany(consumer.elements...)
		return nil, fmt.Errorf("link tee to raw queue: %s", ret.String())
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
				return nil, fmt.Errorf("sync raw branch element %s with parent", element.GetName())
			}
		}
	}

	s.rawConsumers[consumer.id] = consumer
	s.syncPreviewProfilesLocked()
	log.Info().
		Str("event", "capture.gst.shared.raw.attach").
		Str("signature", s.signature).
		Str("consumer_id", consumer.id).
		Str("source_id", s.source.ID).
		Int("raw_consumer_count", len(s.rawConsumers)).
		Int("ref_count", s.refCount).
		Msg("attached shared raw consumer")
	s.mu.Unlock()
	return consumer, nil
}

func (s *sharedVideoSource) detachRawConsumer(consumerID string, err error) {
	if s == nil || consumerID == "" {
		return
	}

	s.mu.Lock()
	consumer, ok := s.rawConsumers[consumerID]
	if !ok {
		s.mu.Unlock()
		return
	}
	delete(s.rawConsumers, consumerID)
	if s.refCount > 0 {
		s.refCount--
	}
	remainingRaw := len(s.rawConsumers)
	s.syncPreviewProfilesLocked()
	shouldShutdown := !s.closed && s.refCount == 0 && len(s.consumers) == 0 && remainingRaw == 0
	s.mu.Unlock()

	s.teardownRawConsumer(consumer, err)
	log.Info().
		Str("event", "capture.gst.shared.raw.detach").
		Str("signature", s.signature).
		Str("consumer_id", consumerID).
		Str("source_id", s.source.ID).
		Int("raw_consumer_count", remainingRaw).
		Msg("detached shared raw consumer")
	if shouldShutdown {
		s.shutdown(nil)
	}
}

func (s *sharedVideoSource) teardownRawConsumer(consumer *sharedRawConsumer, err error) {
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

func buildSharedRawConsumer(source *sharedVideoSource, opts sharedRawConsumerOptions) (*sharedRawConsumer, error) {
	consumerID := fmt.Sprintf("raw-%s-%d", trimSharedSignature(source.signature), source.counter.Add(1))
	queue, err := gst.NewElementWithName("queue", consumerID+"-queue")
	if err != nil {
		return nil, fmt.Errorf("create raw queue: %w", err)
	}
	queue.Set("leaky", 2)
	queue.Set("max-size-buffers", 4)
	videoconvert, err := gst.NewElementWithName("videoconvert", consumerID+"-videoconvert")
	if err != nil {
		return nil, fmt.Errorf("create raw videoconvert: %w", err)
	}
	videoscale, err := gst.NewElementWithName("videoscale", consumerID+"-videoscale")
	if err != nil {
		return nil, fmt.Errorf("create raw videoscale: %w", err)
	}
	videorate, err := gst.NewElementWithName("videorate", consumerID+"-videorate")
	if err != nil {
		return nil, fmt.Errorf("create raw videorate: %w", err)
	}
	capsSpec := fmt.Sprintf("video/x-raw,format=%s,framerate=%d/1,pixel-aspect-ratio=1/1", opts.Format, opts.FPS)
	if opts.Width > 0 && opts.Height > 0 {
		capsSpec = fmt.Sprintf("%s,width=%d,height=%d", capsSpec, opts.Width, opts.Height)
	}
	capsRate, err := newCapsFilter(capsSpec)
	if err != nil {
		return nil, err
	}
	capsRate.SetProperty("name", consumerID+"-normalized-caps")
	sink, err := app.NewAppSink()
	if err != nil {
		return nil, fmt.Errorf("create raw appsink: %w", err)
	}
	sink.SetProperty("name", consumerID+"-appsink")
	sink.SetProperty("max-buffers", 2)
	sink.SetProperty("drop", true)

	consumer := &sharedRawConsumer{
		id:       consumerID,
		source:   source,
		queue:    queue,
		elements: []*gst.Element{queue, videoconvert, videoscale, videorate, capsRate, sink.Element},
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
			return opts.OnSample(sample)
		},
	})
	return consumer, nil
}

func (c *sharedRawConsumer) finish(err error) {
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

func (c *sharedRawConsumer) WaitErr() error {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.waitErr
}

type bridgeVideoRecorder struct {
	pipeline      *gst.Pipeline
	appsrc        *app.Source
	watch         *busWatch
	resultCh      chan error
	frameDuration time.Duration
	frameCount    uint64
	pushCount     atomic.Uint64
	mu            sync.Mutex
	closed        bool
}

func newBridgeVideoRecorder(caps *gst.Caps, opts ExperimentalSharedVideoRecorderOptions) (*bridgeVideoRecorder, error) {
	if caps == nil {
		return nil, errors.New("bridge recorder requires sample caps")
	}
	pipeline, err := gst.NewPipeline("shared-video-record-bridge")
	if err != nil {
		return nil, fmt.Errorf("create bridge recorder pipeline: %w", err)
	}
	appsrc, err := app.NewAppSrc()
	if err != nil {
		return nil, fmt.Errorf("create bridge appsrc: %w", err)
	}
	appsrc.SetProperty("name", "shared-record-appsrc")
	appsrc.SetCaps(caps)
	appsrc.SetStreamType(app.AppStreamTypeStream)
	appsrc.Set("block", true)
	appsrc.Set("emit-signals", false)
	appsrc.Set("format", int(gst.FormatTime))
	appsrc.Set("is-live", true)
	appsrc.SetFormat(gst.FormatTime)
	appsrc.SetDoTimestamp(true)
	appsrc.SetAutomaticEOS(false)

	videoconvert, err := gst.NewElementWithName("videoconvert", "shared-record-videoconvert")
	if err != nil {
		return nil, err
	}
	x264enc, err := gst.NewElementWithName("x264enc", "shared-record-x264enc")
	if err != nil {
		return nil, err
	}
	x264enc.Set("bitrate", 2500)
	x264enc.Set("bframes", 0)
	x264enc.Set("tune", 4)
	x264enc.Set("speed-preset", 3)
	h264parse, err := gst.NewElementWithName("h264parse", "shared-record-h264parse")
	if err != nil {
		return nil, err
	}
	mux, err := newVideoMuxer(opts.Container)
	if err != nil {
		return nil, err
	}
	filesink, err := gst.NewElementWithName("filesink", "shared-record-filesink")
	if err != nil {
		return nil, err
	}
	filesink.Set("location", opts.OutputPath)

	if err := pipeline.AddMany(appsrc.Element, videoconvert, x264enc, h264parse, mux, filesink); err != nil {
		return nil, err
	}
	if err := gst.ElementLinkMany(appsrc.Element, videoconvert, x264enc, h264parse, mux, filesink); err != nil {
		return nil, err
	}

	resultCh := make(chan error, 1)
	watch, err := startBusWatch(pipeline.GetPipelineBus(), func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageError:
			gstErr := msg.ParseError()
			select {
			case resultCh <- fmt.Errorf("shared bridge recorder pipeline error: %w", gstErr):
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
		return nil, err
	}
	if err := pipeline.SetState(gst.StatePlaying); err != nil {
		watch.Stop()
		_ = pipeline.BlockSetState(gst.StateNull)
		return nil, fmt.Errorf("start bridge recorder pipeline: %w", err)
	}

	fps := opts.FPS
	if fps <= 0 {
		fps = 10
	}
	frameDuration := time.Second / time.Duration(fps)
	log.Info().
		Str("event", "capture.gst.shared.bridge_recorder.start").
		Str("output_path", opts.OutputPath).
		Str("caps", caps.String()).
		Int("fps", fps).
		Msg("started experimental shared bridge recorder pipeline")
	return &bridgeVideoRecorder{pipeline: pipeline, appsrc: appsrc, watch: watch, resultCh: resultCh, frameDuration: frameDuration}, nil
}

func (r *bridgeVideoRecorder) pushBuffer(copyBuffer *gst.Buffer) (gst.FlowReturn, error) {
	if r == nil || copyBuffer == nil {
		return gst.FlowError, errors.New("bridge recorder buffer is nil")
	}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return gst.FlowEOS, nil
	}
	frameIndex := r.frameCount
	r.frameCount++
	frameDuration := r.frameDuration
	r.mu.Unlock()

	copyBuffer.SetPresentationTimestamp(gst.ClockTime(time.Duration(frameIndex) * frameDuration))
	copyBuffer.SetDuration(gst.ClockTime(frameDuration))
	ret := r.appsrc.PushBuffer(copyBuffer)
	if ret != gst.FlowOK {
		return ret, fmt.Errorf("bridge recorder push buffer returned %s", ret)
	}
	count := r.pushCount.Add(1)
	if count == 1 || count%10 == 0 {
		log.Info().
			Str("event", "capture.gst.shared.bridge_recorder.push").
			Uint64("buffer_count", count).
			Dur("pts", time.Duration(copyBuffer.PresentationTimestamp())).
			Dur("duration", time.Duration(copyBuffer.Duration())).
			Msg("pushed buffer into experimental shared bridge recorder")
	}
	return ret, nil
}

func (r *bridgeVideoRecorder) stop(timeout time.Duration) error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true
	r.mu.Unlock()

	log.Info().
		Str("event", "capture.gst.shared.bridge_recorder.stop").
		Uint64("buffer_count", r.pushCount.Load()).
		Msg("stopping experimental shared bridge recorder")
	ret := r.appsrc.EndStream()
	if ret != gst.FlowOK && ret != gst.FlowEOS {
		return fmt.Errorf("bridge recorder end stream returned %s", ret)
	}
	select {
	case err := <-r.resultCh:
		if r.watch != nil {
			r.watch.Stop()
		}
		if r.pipeline != nil {
			_ = r.pipeline.BlockSetState(gst.StateNull)
		}
		return err
	case <-time.After(timeout):
		if r.watch != nil {
			r.watch.Stop()
		}
		if r.pipeline != nil {
			_ = r.pipeline.BlockSetState(gst.StateNull)
		}
		return errors.New("timed out waiting for shared bridge recorder EOS")
	}
}

func sharedRecordingTargetSize(source dsl.EffectiveVideoSource) (int, int) {
	if source.Target.Rect != nil && source.Target.Rect.W > 0 && source.Target.Rect.H > 0 {
		return source.Target.Rect.W, source.Target.Rect.H
	}
	if strings.TrimSpace(source.Capture.Size) != "" {
		if width, height, err := parseSize(source.Capture.Size); err == nil {
			return width, height
		}
	}
	return 0, 0
}
