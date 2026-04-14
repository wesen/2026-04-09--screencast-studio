package gst

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-gst/go-gst/gst"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/discovery"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/media"
)

type PreviewRuntime struct {
	registry *captureRegistry
}

var (
	gstInitOnce sync.Once
	gstInitErr  error
)

func NewPreviewRuntime() *PreviewRuntime {
	return &PreviewRuntime{registry: defaultCaptureRegistry}
}

func (r *PreviewRuntime) StartPreview(ctx context.Context, source dsl.EffectiveVideoSource, opts media.PreviewOptions) (media.PreviewSession, error) {
	if err := initGStreamer(); err != nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if r == nil || r.registry == nil {
		r = &PreviewRuntime{registry: defaultCaptureRegistry}
	}
	if opts.OnFrame == nil {
		opts.OnFrame = func([]byte) {}
	}
	if opts.OnLog == nil {
		opts.OnLog = func(string, string) {}
	}

	previewCtx, cancel := context.WithCancel(ctx)
	shared, err := r.registry.acquireVideoSource(previewCtx, source)
	if err != nil {
		cancel()
		return nil, err
	}
	consumer, err := shared.attachPreviewConsumer(opts)
	if err != nil {
		shared.releaseReference()
		cancel()
		return nil, err
	}

	session := &previewSession{
		cancel:   cancel,
		source:   shared,
		consumer: consumer,
	}
	go func() {
		select {
		case <-previewCtx.Done():
			shared.detachPreviewConsumer(consumer.id, nil)
		case <-consumer.done:
		}
	}()

	log.Info().
		Str("event", "preview.gst.start").
		Str("source_id", shared.source.ID).
		Str("source_name", shared.source.Name).
		Str("source_type", shared.source.Type).
		Str("shared_signature", shared.signature).
		Msg("started gstreamer preview session on shared source")
	return session, nil
}

type previewSession struct {
	cancel   context.CancelFunc
	source   *sharedVideoSource
	consumer *sharedPreviewConsumer
}

func (s *previewSession) Wait() error {
	if s == nil || s.consumer == nil {
		return nil
	}
	<-s.consumer.done
	return s.consumer.WaitErr()
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
	if s.consumer == nil {
		return nil
	}
	select {
	case <-s.consumer.done:
		return s.consumer.WaitErr()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *previewSession) LatestFrame() ([]byte, error) {
	if s == nil || s.consumer == nil {
		return nil, errors.New("preview session is nil")
	}
	return s.consumer.LatestFrame()
}

func (s *previewSession) TakeScreenshot(ctx context.Context, opts media.ScreenshotOptions) ([]byte, error) {
	_ = ctx
	_ = opts
	return s.LatestFrame()
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

		elements := []*gst.Element{ximagesrc}
		switch source.Type {
		case "region", "window":
			if source.Target.Rect == nil {
				return nil, fmt.Errorf("%s source missing target.rect", source.Type)
			}
			crop, err := buildVideoCrop(source)
			if err != nil {
				return nil, err
			}
			elements = append(elements, crop)
		}
		return elements, nil
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

func buildVideoCrop(source dsl.EffectiveVideoSource) (*gst.Element, error) {
	if source.Target.Rect == nil {
		return nil, fmt.Errorf("%s source missing target.rect", source.Type)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	rootWidth, rootHeight, err := discovery.RootGeometry(ctx, source.Target.Display)
	if err != nil {
		return nil, fmt.Errorf("resolve root geometry for %s source %s: %w", source.Type, source.ID, err)
	}
	rect := source.Target.Rect
	left := maxInt(rect.X, 0)
	top := maxInt(rect.Y, 0)
	right := maxInt(rootWidth-(rect.X+rect.W), 0)
	bottom := maxInt(rootHeight-(rect.Y+rect.H), 0)
	crop, err := gst.NewElement("videocrop")
	if err != nil {
		return nil, fmt.Errorf("create videocrop: %w", err)
	}
	crop.Set("left", left)
	crop.Set("top", top)
	crop.Set("right", right)
	crop.Set("bottom", bottom)
	log.Info().
		Str("event", "preview.gst.crop.configure").
		Str("source_id", source.ID).
		Str("source_type", source.Type).
		Int("root_width", rootWidth).
		Int("root_height", rootHeight).
		Int("rect_x", rect.X).
		Int("rect_y", rect.Y).
		Int("rect_w", rect.W).
		Int("rect_h", rect.H).
		Int("crop_left", left).
		Int("crop_top", top).
		Int("crop_right", right).
		Int("crop_bottom", bottom).
		Msg("configured videocrop for preview source")
	return crop, nil
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
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
