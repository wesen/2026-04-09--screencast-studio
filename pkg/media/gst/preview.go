package gst

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/discovery"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/media"
)

type PreviewRuntime struct{}

var (
	gstInitOnce sync.Once
	gstInitErr  error
)

func NewPreviewRuntime() *PreviewRuntime {
	return &PreviewRuntime{}
}

func (PreviewRuntime) StartPreview(ctx context.Context, source dsl.EffectiveVideoSource, opts media.PreviewOptions) (media.PreviewSession, error) {
	if err := initGStreamer(); err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	previewCtx, cancel := context.WithCancel(ctx)
	if opts.OnFrame == nil {
		opts.OnFrame = func([]byte) {}
	}
	if opts.OnLog == nil {
		opts.OnLog = func(string, string) {}
	}

	resolvedSource, err := resolveWindowSource(previewCtx, source)
	if err != nil {
		cancel()
		return nil, err
	}

	pipeline, sink, err := buildPreviewPipeline(resolvedSource)
	if err != nil {
		cancel()
		return nil, err
	}

	session := &previewSession{
		cancel:   cancel,
		pipeline: pipeline,
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
			session.setLatestFrame(frame)
			opts.OnFrame(frame)
			return gst.FlowOK
		},
	})

	bus := pipeline.GetPipelineBus()
	resultCh := make(chan error, 1)
	watch, err := startBusWatch(bus, func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageError:
			gstErr := msg.ParseError()
			err := fmt.Errorf("preview pipeline error: %w", gstErr)
			opts.OnLog("stderr", err.Error())
			select {
			case resultCh <- err:
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
		cancel()
		pipeline.BlockSetState(gst.StateNull)
		return nil, err
	}

	if err := pipeline.SetState(gst.StatePlaying); err != nil {
		watch.Stop()
		cancel()
		pipeline.BlockSetState(gst.StateNull)
		return nil, fmt.Errorf("start preview pipeline: %w", err)
	}

	log.Info().
		Str("event", "preview.gst.start").
		Str("source_id", resolvedSource.ID).
		Str("source_name", resolvedSource.Name).
		Str("source_type", resolvedSource.Type).
		Msg("started gstreamer preview pipeline")

	go session.wait(previewCtx, watch, resultCh)
	return session, nil
}

type previewSession struct {
	cancel   context.CancelFunc
	pipeline *gst.Pipeline
	done     chan struct{}

	mu          sync.RWMutex
	latestFrame []byte
	waitErr     error
}

func (s *previewSession) Wait() error {
	if s == nil {
		return nil
	}
	if s.done != nil {
		<-s.done
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.waitErr
}

func (s *previewSession) Stop(ctx context.Context) error {
	if s == nil {
		return nil
	}
	if s.cancel != nil {
		s.cancel()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if s.done == nil {
		return nil
	}
	select {
	case <-s.done:
		return s.Wait()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *previewSession) LatestFrame() ([]byte, error) {
	if s == nil {
		return nil, errors.New("preview session is nil")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.latestFrame) == 0 {
		return nil, errors.New("preview frame not available yet")
	}
	return append([]byte(nil), s.latestFrame...), nil
}

func (s *previewSession) TakeScreenshot(ctx context.Context, opts media.ScreenshotOptions) ([]byte, error) {
	_ = ctx
	_ = opts
	return s.LatestFrame()
}

func (s *previewSession) wait(ctx context.Context, watch *busWatch, resultCh <-chan error) {
	defer close(s.done)
	defer watch.Stop()
	defer s.pipeline.BlockSetState(gst.StateNull)

	select {
	case err := <-resultCh:
		s.setWaitResult(err)
	case <-ctx.Done():
		sendEOS(s.pipeline)
		select {
		case err := <-resultCh:
			if ctx.Err() != nil && err == nil {
				s.setWaitResult(nil)
			} else {
				s.setWaitResult(err)
			}
		case <-time.After(750 * time.Millisecond):
			s.setWaitResult(nil)
		}
	}
}

func (s *previewSession) setLatestFrame(frame []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latestFrame = append([]byte(nil), frame...)
}

func (s *previewSession) setWaitResult(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.waitErr = err
}

func buildPreviewPipeline(source dsl.EffectiveVideoSource) (*gst.Pipeline, *app.Sink, error) {
	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return nil, nil, fmt.Errorf("create pipeline: %w", err)
	}

	elements, err := buildPreviewElements(source)
	if err != nil {
		return nil, nil, err
	}

	sink, err := app.NewAppSink()
	if err != nil {
		return nil, nil, fmt.Errorf("create appsink: %w", err)
	}
	sink.SetCaps(gst.NewCapsFromString("image/jpeg"))
	sink.SetProperty("max-buffers", 2)
	sink.SetProperty("drop", true)

	toAdd := append([]*gst.Element(nil), elements...)
	toAdd = append(toAdd, sink.Element)
	pipeline.AddMany(toAdd...)
	if err := linkElements(append(elements, sink.Element)...); err != nil {
		pipeline.BlockSetState(gst.StateNull)
		return nil, nil, err
	}
	return pipeline, sink, nil
}

func buildPreviewElements(source dsl.EffectiveVideoSource) ([]*gst.Element, error) {
	elements := []*gst.Element{}

	sourceElements, err := buildSourceElements(source)
	if err != nil {
		return nil, err
	}
	elements = append(elements, sourceElements...)

	videoconvert, err := gst.NewElement("videoconvert")
	if err != nil {
		return nil, fmt.Errorf("create videoconvert: %w", err)
	}
	elements = append(elements, videoconvert)

	if source.Type == "camera" && boolValue(source.Capture.Mirror, false) {
		videoflip, err := gst.NewElement("videoflip")
		if err != nil {
			return nil, fmt.Errorf("create videoflip: %w", err)
		}
		videoflip.Set("method", "horizontal-flip")
		elements = append(elements, videoflip)
	}

	videoscale, err := gst.NewElement("videoscale")
	if err != nil {
		return nil, fmt.Errorf("create videoscale: %w", err)
	}
	capsScale, err := newCapsFilter("video/x-raw,width=640")
	if err != nil {
		return nil, err
	}
	videorate, err := gst.NewElement("videorate")
	if err != nil {
		return nil, fmt.Errorf("create videorate: %w", err)
	}
	capsRate, err := newCapsFilter("video/x-raw,framerate=5/1")
	if err != nil {
		return nil, err
	}
	jpegenc, err := gst.NewElement("jpegenc")
	if err != nil {
		return nil, fmt.Errorf("create jpegenc: %w", err)
	}
	jpegenc.Set("quality", 50)

	elements = append(elements, videoscale, capsScale, videorate, capsRate, jpegenc)
	return elements, nil
}

func buildSourceElements(source dsl.EffectiveVideoSource) ([]*gst.Element, error) {
	switch source.Type {
	case "display", "region", "window":
		ximagesrc, err := gst.NewElement("ximagesrc")
		if err != nil {
			return nil, fmt.Errorf("create ximagesrc: %w", err)
		}
		if display := strings.TrimSpace(source.Target.Display); display != "" {
			ximagesrc.Set("display-name", display)
		}
		ximagesrc.Set("show-pointer", boolValue(source.Capture.Cursor, true))
		ximagesrc.Set("use-damage", false)

		switch source.Type {
		case "region", "window":
			if source.Target.Rect == nil {
				return nil, fmt.Errorf("%s source missing target.rect", source.Type)
			}
			rect := source.Target.Rect
			ximagesrc.Set("startx", rect.X)
			ximagesrc.Set("starty", rect.Y)
			ximagesrc.Set("endx", rect.X+rect.W-1)
			ximagesrc.Set("endy", rect.Y+rect.H-1)
		}
		return []*gst.Element{ximagesrc}, nil
	case "camera":
		if strings.TrimSpace(source.Target.Device) == "" {
			return nil, errors.New("camera source missing target.device")
		}
		v4l2src, err := gst.NewElement("v4l2src")
		if err != nil {
			return nil, fmt.Errorf("create v4l2src: %w", err)
		}
		v4l2src.Set("device", source.Target.Device)
		elements := []*gst.Element{v4l2src}
		if strings.TrimSpace(source.Capture.Size) != "" {
			width, height, err := parseSize(source.Capture.Size)
			if err != nil {
				return nil, err
			}
			caps, err := newCapsFilter(fmt.Sprintf("video/x-raw,width=%d,height=%d", width, height))
			if err != nil {
				return nil, err
			}
			elements = append(elements, caps)
		}
		return elements, nil
	default:
		return nil, fmt.Errorf("unsupported video source type %q", source.Type)
	}
}

func parseSize(size string) (int, int, error) {
	parts := strings.Split(strings.TrimSpace(size), "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid size %q", size)
	}
	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse width from %q: %w", size, err)
	}
	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse height from %q: %w", size, err)
	}
	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("invalid size %q", size)
	}
	return width, height, nil
}

func sendEOS(pipeline *gst.Pipeline) {
	if pipeline == nil {
		return
	}
	pipeline.SendEvent(gst.NewEOSEvent())
}

func resolveWindowSource(ctx context.Context, source dsl.EffectiveVideoSource) (dsl.EffectiveVideoSource, error) {
	if source.Type != "window" {
		return source, nil
	}
	if source.Target.Rect != nil && source.Target.Rect.W > 0 && source.Target.Rect.H > 0 {
		return source, nil
	}
	if strings.TrimSpace(source.Target.WindowID) == "" {
		return source, errors.New("window source missing target.window_id")
	}
	x, y, width, height, err := discovery.WindowGeometry(ctx, source.Target.WindowID)
	if err != nil {
		return source, fmt.Errorf("resolve window geometry for %s: %w", source.Target.WindowID, err)
	}
	resolved := source
	resolved.Target.Rect = &dsl.Rect{X: x, Y: y, W: width, H: height}
	log.Info().
		Str("event", "preview.gst.window_geometry.resolved").
		Str("source_id", source.ID).
		Str("window_id", source.Target.WindowID).
		Int("x", x).
		Int("y", y).
		Int("width", width).
		Int("height", height).
		Msg("resolved window preview geometry for region-style capture")
	return resolved, nil
}

func boolValue(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}

func initGStreamer() error {
	gstInitOnce.Do(func() {
		gst.Init(nil)
	})
	return gstInitErr
}
