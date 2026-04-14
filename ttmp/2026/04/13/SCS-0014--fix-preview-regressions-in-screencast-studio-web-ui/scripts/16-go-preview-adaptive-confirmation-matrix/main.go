package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
)

type rect struct{ X, Y, W, H int }

type previewOrder string

const (
	previewDisabled   previewOrder = "disabled"
	previewScaleFirst previewOrder = "scale-first"
	previewRateFirst  previewOrder = "rate-first"
)

type previewProfile struct {
	Enabled bool
	Order   previewOrder
	Width   int
	Height  int
	FPS     int
	Quality int
	Label   string
}

type previewMetrics struct {
	SamplesPulled atomic.Uint64
	FramesCopied  atomic.Uint64
	BytesCopied   atomic.Uint64
}

type recorderMetrics struct {
	SamplesPulled atomic.Uint64
	BuffersCopied atomic.Uint64
	Enqueued      atomic.Uint64
	Dropped       atomic.Uint64
	WorkerHandled atomic.Uint64
	AppsrcPushed  atomic.Uint64
}

type combinedMetrics struct {
	Preview  previewMetrics
	Recorder recorderMetrics
}

type appsrcBridge struct {
	pipeline      *gst.Pipeline
	appsrc        *app.Source
	frameDuration time.Duration
	frameCount    uint64
	mu            sync.Mutex
	closed        bool
}

func main() {
	gst.Init(nil)

	scenario := envOr("SCENARIO", "preview-rate-first-constrained-plus-recorder")
	displayName := envOr("DISPLAY", ":0")
	region := parseRegion(envOr("REGION", "0,540,1920,540"))
	rootW := intFromEnv("ROOT_WIDTH", 1920)
	rootH := intFromEnv("ROOT_HEIGHT", 1080)
	fps := intFromEnv("FPS", 24)
	duration := durationFromEnv("DURATION", 6*time.Second)
	outputPath := envOr("OUTPUT_PATH", filepath.Join(os.TempDir(), "scs-preview-adaptive-confirmation.mp4"))

	preview, recorderEnabled := scenarioProfile(scenario, region)

	fmt.Printf("scenario=%s\n", scenario)
	fmt.Printf("display=%s\n", displayName)
	fmt.Printf("root=%dx%d\n", rootW, rootH)
	fmt.Printf("region=%d,%d,%d,%d\n", region.X, region.Y, region.W, region.H)
	fmt.Printf("fps=%d\n", fps)
	fmt.Printf("duration=%s\n", duration)
	fmt.Printf("preview_enabled=%t\n", preview.Enabled)
	fmt.Printf("preview_label=%s\n", preview.Label)
	fmt.Printf("preview_order=%s\n", preview.Order)
	fmt.Printf("preview_width=%d\n", preview.Width)
	fmt.Printf("preview_height=%d\n", preview.Height)
	fmt.Printf("preview_fps=%d\n", preview.FPS)
	fmt.Printf("preview_quality=%d\n", preview.Quality)
	fmt.Printf("output=%s\n", outputPath)

	pipeline, metrics, bridge, workerDone, sampleCh, err := buildPipeline(displayName, rootW, rootH, region, fps, preview, recorderEnabled, outputPath)
	must(err, "build pipeline")
	defer func() { _ = pipeline.BlockSetState(gst.StateNull) }()
	if bridge != nil {
		defer func() { _ = bridge.pipeline.BlockSetState(gst.StateNull) }()
	}

	must(pipeline.SetState(gst.StatePlaying), "start source pipeline")
	if bridge != nil {
		must(bridge.pipeline.SetState(gst.StatePlaying), "start recorder bridge pipeline")
	}

	time.Sleep(duration)
	pipeline.SendEvent(gst.NewEOSEvent())
	must(waitPipeline(pipeline), "wait source pipeline EOS")
	if sampleCh != nil {
		close(sampleCh)
		<-workerDone
	}
	if bridge != nil {
		must(bridge.stop(), "stop recorder bridge")
	}

	fmt.Printf("preview_samples_pulled=%d\n", metrics.Preview.SamplesPulled.Load())
	fmt.Printf("preview_frames_copied=%d\n", metrics.Preview.FramesCopied.Load())
	fmt.Printf("preview_bytes_copied=%d\n", metrics.Preview.BytesCopied.Load())
	fmt.Printf("recorder_samples_pulled=%d\n", metrics.Recorder.SamplesPulled.Load())
	fmt.Printf("recorder_buffers_copied=%d\n", metrics.Recorder.BuffersCopied.Load())
	fmt.Printf("recorder_enqueued=%d\n", metrics.Recorder.Enqueued.Load())
	fmt.Printf("recorder_dropped=%d\n", metrics.Recorder.Dropped.Load())
	fmt.Printf("recorder_worker_handled=%d\n", metrics.Recorder.WorkerHandled.Load())
	fmt.Printf("recorder_appsrc_pushed=%d\n", metrics.Recorder.AppsrcPushed.Load())
}

func scenarioProfile(scenario string, region rect) (previewProfile, bool) {
	switch scenario {
	case "recorder-only":
		return previewProfile{Enabled: false, Order: previewDisabled, Label: "none"}, true
	case "preview-scale-first-current-plus-recorder":
		p := currentProfile(region)
		p.Order = previewScaleFirst
		p.Label = "current"
		return p, true
	case "preview-rate-first-current-plus-recorder":
		p := currentProfile(region)
		p.Order = previewRateFirst
		p.Label = "current"
		return p, true
	case "preview-scale-first-constrained-plus-recorder":
		p := constrainedProfile(region)
		p.Order = previewScaleFirst
		p.Label = "constrained"
		return p, true
	case "preview-rate-first-constrained-plus-recorder":
		p := constrainedProfile(region)
		p.Order = previewRateFirst
		p.Label = "constrained"
		return p, true
	default:
		fatal("unsupported scenario=%s", scenario)
		return previewProfile{}, false
	}
}

func buildPipeline(displayName string, rootW, rootH int, region rect, fps int, preview previewProfile, recorderEnabled bool, outputPath string) (*gst.Pipeline, *combinedMetrics, *appsrcBridge, chan struct{}, chan *gst.Buffer, error) {
	metrics := &combinedMetrics{}
	pipeline, err := gst.NewPipeline("preview-adaptive-confirmation")
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	ximagesrc, err := gst.NewElement("ximagesrc")
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	ximagesrc.Set("display-name", displayName)
	ximagesrc.Set("use-damage", false)
	ximagesrc.Set("show-pointer", true)
	crop, err := gst.NewElement("videocrop")
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	crop.Set("left", region.X)
	crop.Set("top", region.Y)
	crop.Set("right", maxInt(rootW-(region.X+region.W), 0))
	crop.Set("bottom", maxInt(rootH-(region.Y+region.H), 0))
	videoconvert, err := gst.NewElement("videoconvert")
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	tee, err := gst.NewElement("tee")
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if err := pipeline.AddMany(ximagesrc, crop, videoconvert, tee); err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if err := gst.ElementLinkMany(ximagesrc, crop, videoconvert, tee); err != nil {
		return nil, nil, nil, nil, nil, err
	}

	var bridge *appsrcBridge
	var workerDone chan struct{}
	var sampleCh chan *gst.Buffer

	if preview.Enabled {
		previewElems, err := buildPreviewBranch(preview, metrics)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if err := pipeline.AddMany(previewElems...); err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if err := gst.ElementLinkMany(previewElems...); err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if err := linkTeeBranch(tee, previewElems[0]); err != nil {
			return nil, nil, nil, nil, nil, err
		}
	}

	if recorderEnabled {
		bridge, err = newAppsrcBridge(region.W, region.H, fps, outputPath)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		sampleCh = make(chan *gst.Buffer, 16)
		workerDone = make(chan struct{})
		go pushBuffers(sampleCh, workerDone, bridge, &metrics.Recorder)
		recorderElems, err := buildRecorderRawBranch(region, fps, metrics, sampleCh)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if err := pipeline.AddMany(recorderElems...); err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if err := gst.ElementLinkMany(recorderElems...); err != nil {
			return nil, nil, nil, nil, nil, err
		}
		if err := linkTeeBranch(tee, recorderElems[0]); err != nil {
			return nil, nil, nil, nil, nil, err
		}
	}

	return pipeline, metrics, bridge, workerDone, sampleCh, nil
}

func buildPreviewBranch(preview previewProfile, metrics *combinedMetrics) ([]*gst.Element, error) {
	queue, err := gst.NewElement("queue")
	if err != nil {
		return nil, err
	}
	queue.Set("leaky", 2)
	queue.Set("max-size-buffers", 2)
	videoscale, err := gst.NewElement("videoscale")
	if err != nil {
		return nil, err
	}
	capsScale, err := newCapsFilter(fmt.Sprintf("video/x-raw,width=%d,height=%d,pixel-aspect-ratio=1/1", preview.Width, preview.Height))
	if err != nil {
		return nil, err
	}
	videorate, err := gst.NewElement("videorate")
	if err != nil {
		return nil, err
	}
	capsRate, err := newCapsFilter(fmt.Sprintf("video/x-raw,framerate=%d/1", preview.FPS))
	if err != nil {
		return nil, err
	}
	jpegenc, err := gst.NewElement("jpegenc")
	if err != nil {
		return nil, err
	}
	jpegenc.Set("quality", preview.Quality)
	sink, err := app.NewAppSink()
	if err != nil {
		return nil, err
	}
	sink.SetCaps(gst.NewCapsFromString("image/jpeg"))
	sink.SetProperty("max-buffers", 2)
	sink.SetProperty("drop", true)
	sink.SetCallbacks(&app.SinkCallbacks{NewSampleFunc: func(s *app.Sink) gst.FlowReturn {
		sample := s.PullSample()
		if sample == nil {
			return gst.FlowEOS
		}
		metrics.Preview.SamplesPulled.Add(1)
		buffer := sample.GetBuffer()
		if buffer == nil {
			return gst.FlowError
		}
		mapInfo := buffer.Map(gst.MapRead)
		if mapInfo == nil {
			return gst.FlowError
		}
		frame := append([]byte(nil), mapInfo.Bytes()...)
		buffer.Unmap()
		metrics.Preview.FramesCopied.Add(1)
		metrics.Preview.BytesCopied.Add(uint64(len(frame)))
		return gst.FlowOK
	}})

	switch preview.Order {
	case previewScaleFirst:
		return []*gst.Element{queue, videoscale, capsScale, videorate, capsRate, jpegenc, sink.Element}, nil
	case previewRateFirst:
		return []*gst.Element{queue, videorate, capsRate, videoscale, capsScale, jpegenc, sink.Element}, nil
	default:
		return nil, fmt.Errorf("unsupported preview order %s", preview.Order)
	}
}

func buildRecorderRawBranch(region rect, fps int, metrics *combinedMetrics, sampleCh chan *gst.Buffer) ([]*gst.Element, error) {
	queue, err := gst.NewElement("queue")
	if err != nil {
		return nil, err
	}
	queue.Set("leaky", 2)
	queue.Set("max-size-buffers", 4)
	videoconvert, err := gst.NewElement("videoconvert")
	if err != nil {
		return nil, err
	}
	videoscale, err := gst.NewElement("videoscale")
	if err != nil {
		return nil, err
	}
	videorate, err := gst.NewElement("videorate")
	if err != nil {
		return nil, err
	}
	capsRate, err := newCapsFilter(fmt.Sprintf("video/x-raw,format=I420,width=%d,height=%d,framerate=%d/1,pixel-aspect-ratio=1/1", region.W, region.H, fps))
	if err != nil {
		return nil, err
	}
	sink, err := app.NewAppSink()
	if err != nil {
		return nil, err
	}
	sink.SetProperty("max-buffers", 2)
	sink.SetProperty("drop", true)
	sink.SetCallbacks(&app.SinkCallbacks{NewSampleFunc: func(s *app.Sink) gst.FlowReturn {
		sample := s.PullSample()
		if sample == nil {
			return gst.FlowEOS
		}
		metrics.Recorder.SamplesPulled.Add(1)
		buffer := sample.GetBuffer()
		if buffer == nil {
			return gst.FlowError
		}
		copied := buffer.Copy()
		if copied == nil {
			return gst.FlowError
		}
		metrics.Recorder.BuffersCopied.Add(1)
		select {
		case sampleCh <- copied:
			metrics.Recorder.Enqueued.Add(1)
		default:
			metrics.Recorder.Dropped.Add(1)
		}
		return gst.FlowOK
	}})
	return []*gst.Element{queue, videoconvert, videoscale, videorate, capsRate, sink.Element}, nil
}

func linkTeeBranch(tee *gst.Element, first *gst.Element) error {
	teePad := tee.GetRequestPad("src_%u")
	if teePad == nil {
		return fmt.Errorf("request tee src pad")
	}
	sinkPad := first.GetStaticPad("sink")
	if sinkPad == nil {
		return fmt.Errorf("branch sink pad")
	}
	if ret := teePad.Link(sinkPad); ret != gst.PadLinkOK {
		return fmt.Errorf("link tee branch: %s", ret)
	}
	return nil
}

func newAppsrcBridge(width, height, fps int, outputPath string) (*appsrcBridge, error) {
	pipeline, err := gst.NewPipeline("preview-adaptive-confirmation-bridge")
	if err != nil {
		return nil, err
	}
	src, err := app.NewAppSrc()
	if err != nil {
		return nil, err
	}
	src.SetCaps(gst.NewCapsFromString(fmt.Sprintf("video/x-raw,format=I420,width=%d,height=%d,framerate=%d/1,pixel-aspect-ratio=1/1", width, height, fps)))
	src.SetStreamType(app.AppStreamTypeStream)
	src.Set("block", true)
	src.Set("emit-signals", false)
	src.Set("format", int(gst.FormatTime))
	src.Set("is-live", true)
	src.SetFormat(gst.FormatTime)
	src.SetDoTimestamp(true)
	src.SetAutomaticEOS(false)
	videoconvert, err := gst.NewElement("videoconvert")
	if err != nil {
		return nil, err
	}
	x264enc, err := gst.NewElement("x264enc")
	if err != nil {
		return nil, err
	}
	x264enc.Set("bitrate", 2500)
	x264enc.Set("bframes", 0)
	x264enc.Set("tune", 4)
	x264enc.Set("speed-preset", 3)
	h264parse, err := gst.NewElement("h264parse")
	if err != nil {
		return nil, err
	}
	mux, err := gst.NewElement("mp4mux")
	if err != nil {
		return nil, err
	}
	filesink, err := gst.NewElement("filesink")
	if err != nil {
		return nil, err
	}
	filesink.Set("location", outputPath)
	if err := pipeline.AddMany(src.Element, videoconvert, x264enc, h264parse, mux, filesink); err != nil {
		return nil, err
	}
	if err := gst.ElementLinkMany(src.Element, videoconvert, x264enc, h264parse, mux, filesink); err != nil {
		return nil, err
	}
	return &appsrcBridge{pipeline: pipeline, appsrc: src, frameDuration: time.Second / time.Duration(fps)}, nil
}

func (b *appsrcBridge) pushBuffer(buf *gst.Buffer, metrics *recorderMetrics) error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	frameIndex := b.frameCount
	b.frameCount++
	frameDuration := b.frameDuration
	b.mu.Unlock()
	buf.SetPresentationTimestamp(gst.ClockTime(time.Duration(frameIndex) * frameDuration))
	buf.SetDuration(gst.ClockTime(frameDuration))
	ret := b.appsrc.PushBuffer(buf)
	if ret != gst.FlowOK {
		return fmt.Errorf("appsrc push returned %s", ret)
	}
	metrics.AppsrcPushed.Add(1)
	return nil
}

func (b *appsrcBridge) stop() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	b.mu.Unlock()
	ret := b.appsrc.EndStream()
	if ret != gst.FlowOK && ret != gst.FlowEOS {
		return fmt.Errorf("appsrc end stream returned %s", ret)
	}
	if err := waitPipeline(b.pipeline); err != nil {
		return err
	}
	return b.pipeline.BlockSetState(gst.StateNull)
}

func pushBuffers(ch <-chan *gst.Buffer, done chan<- struct{}, bridge *appsrcBridge, metrics *recorderMetrics) {
	defer close(done)
	for buf := range ch {
		if buf == nil {
			continue
		}
		metrics.WorkerHandled.Add(1)
		if err := bridge.pushBuffer(buf, metrics); err != nil {
			fmt.Fprintf(os.Stderr, "bridge push error: %v\n", err)
			return
		}
	}
}

func waitPipeline(pipeline *gst.Pipeline) error {
	bus := pipeline.GetPipelineBus()
	if bus == nil {
		return fmt.Errorf("pipeline bus is nil")
	}
	for {
		msg := bus.TimedPop(gst.ClockTime(500 * time.Millisecond))
		if msg == nil {
			continue
		}
		switch msg.Type() {
		case gst.MessageEOS:
			return nil
		case gst.MessageError:
			return msg.ParseError()
		}
	}
}

func newCapsFilter(spec string) (*gst.Element, error) {
	filter, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, err
	}
	filter.SetProperty("caps", gst.NewCapsFromString(spec))
	return filter, nil
}

func currentProfile(region rect) previewProfile {
	width, height := targetDimensions(region.W, region.H, 1280)
	return previewProfile{Enabled: true, Width: width, Height: height, FPS: 10, Quality: 80}
}

func constrainedProfile(region rect) previewProfile {
	width, height := targetDimensions(region.W, region.H, 640)
	return previewProfile{Enabled: true, Width: width, Height: height, FPS: 4, Quality: 50}
}

func targetDimensions(width, height, maxWidth int) (int, int) {
	if width <= 0 || height <= 0 {
		return maxWidth, 0
	}
	if width <= maxWidth {
		return width, height
	}
	scaledHeight := int(float64(height)*(float64(maxWidth)/float64(width)) + 0.5)
	if scaledHeight <= 0 {
		scaledHeight = 1
	}
	return maxWidth, scaledHeight
}

func parseRegion(value string) rect {
	parts := strings.Split(strings.TrimSpace(value), ",")
	if len(parts) != 4 {
		fatal("REGION must be x,y,w,h")
	}
	vals := [4]int{}
	for i, part := range parts {
		var n int
		_, err := fmt.Sscanf(strings.TrimSpace(part), "%d", &n)
		if err != nil {
			fatal("parse REGION component %q: %v", part, err)
		}
		vals[i] = n
	}
	return rect{vals[0], vals[1], vals[2], vals[3]}
}

func durationFromEnv(key string, def time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return def
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		fatal("parse %s=%q: %v", key, value, err)
	}
	return d
}

func intFromEnv(key string, def int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return def
	}
	var n int
	_, err := fmt.Sscanf(value, "%d", &n)
	if err != nil {
		fatal("parse %s=%q: %v", key, value, err)
	}
	return n
}

func envOr(key, def string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return def
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func must(err error, label string) {
	if err != nil {
		fatal("%s: %v", label, err)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
