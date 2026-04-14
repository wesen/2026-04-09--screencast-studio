package main

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
)

type metrics struct {
	samplesPulled atomic.Uint64
	buffersCopied atomic.Uint64
	enqueued      atomic.Uint64
	dropped       atomic.Uint64
	workerHandled atomic.Uint64
	appsrcPushed  atomic.Uint64
}

type bridgeSinkMode string

const (
	modeDiscard    bridgeSinkMode = "discard"
	modeAppsrcFake bridgeSinkMode = "appsrc-fakesink"
	modeAppsrcX264 bridgeSinkMode = "appsrc-x264"
)

type appsrcBridge struct {
	mode          bridgeSinkMode
	pipeline      *gst.Pipeline
	appsrc        *app.Source
	resultCh      chan error
	frameDuration time.Duration
	frameCount    uint64
	mu            sync.Mutex
	closed        bool
}

func main() {
	gst.Init(nil)

	scenario := envOr("SCENARIO", "appsink-discard")
	displayName := envOr("DISPLAY", ":0")
	region := parseRegion(envOr("REGION", "0,540,1920,540"))
	rootW := intFromEnv("ROOT_WIDTH", 1920)
	rootH := intFromEnv("ROOT_HEIGHT", 1080)
	fps := intFromEnv("FPS", 24)
	duration := durationFromEnv("DURATION", 6*time.Second)
	outputPath := envOr("OUTPUT_PATH", filepath.Join(os.TempDir(), "scs-go-bridge-overhead.mp4"))

	fmt.Printf("scenario=%s\n", scenario)
	fmt.Printf("display=%s\n", displayName)
	fmt.Printf("root=%dx%d\n", rootW, rootH)
	fmt.Printf("region=%d,%d,%d,%d\n", region.X, region.Y, region.W, region.H)
	fmt.Printf("fps=%d\n", fps)
	fmt.Printf("duration=%s\n", duration)
	fmt.Printf("output=%s\n", outputPath)

	if region.X < 0 || region.Y < 0 || region.X+region.W > rootW || region.Y+region.H > rootH {
		fatal("invalid region %v for root %dx%d", region, rootW, rootH)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := &metrics{}
	var sampleCh chan *gst.Buffer
	var workerDone chan struct{}
	var bridge *appsrcBridge
	var err error

	switch scenario {
	case "normalized-fakesink":
		// no-op, direct sink path handled in pipeline builder
	case "appsink-discard", "appsink-copy-discard":
		// no-op
	case "appsink-copy-async-discard":
		sampleCh = make(chan *gst.Buffer, 16)
		workerDone = make(chan struct{})
		go drainBuffers(sampleCh, workerDone, m)
	case "appsink-copy-async-appsrc-fakesink":
		sampleCh = make(chan *gst.Buffer, 16)
		workerDone = make(chan struct{})
		bridge, err = newAppsrcBridge(modeAppsrcFake, region.W, region.H, fps, outputPath)
		must(err, "create appsrc fakesink bridge")
		go pushBuffers(sampleCh, workerDone, bridge, m)
	case "appsink-copy-async-appsrc-x264":
		sampleCh = make(chan *gst.Buffer, 16)
		workerDone = make(chan struct{})
		bridge, err = newAppsrcBridge(modeAppsrcX264, region.W, region.H, fps, outputPath)
		must(err, "create appsrc x264 bridge")
		go pushBuffers(sampleCh, workerDone, bridge, m)
	default:
		fatal("unsupported scenario=%s", scenario)
	}

	pipeline, err := buildSourcePipeline(displayName, rootW, rootH, region, fps, scenario, m, sampleCh)
	must(err, "build source pipeline")

	must(pipeline.SetState(gst.StatePlaying), "start source pipeline")
	if bridge != nil {
		must(bridge.pipeline.SetState(gst.StatePlaying), "start bridge pipeline")
	}

	go func() {
		timer := time.NewTimer(duration)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			pipeline.SendEvent(gst.NewEOSEvent())
		}
	}()

	must(waitPipeline(pipeline), "wait source pipeline")
	cancel()
	_ = pipeline.BlockSetState(gst.StateNull)

	if sampleCh != nil {
		close(sampleCh)
		<-workerDone
	}
	if bridge != nil {
		must(bridge.stop(), "stop bridge pipeline")
	}

	fmt.Printf("samples_pulled=%d\n", m.samplesPulled.Load())
	fmt.Printf("buffers_copied=%d\n", m.buffersCopied.Load())
	fmt.Printf("enqueued=%d\n", m.enqueued.Load())
	fmt.Printf("dropped=%d\n", m.dropped.Load())
	fmt.Printf("worker_handled=%d\n", m.workerHandled.Load())
	fmt.Printf("appsrc_pushed=%d\n", m.appsrcPushed.Load())
}

type rect struct{ X, Y, W, H int }

func buildSourcePipeline(displayName string, rootW, rootH int, region rect, fps int, scenario string, m *metrics, sampleCh chan *gst.Buffer) (*gst.Pipeline, error) {
	pipeline, err := gst.NewPipeline("bridge-overhead-source")
	if err != nil {
		return nil, err
	}
	ximagesrc, err := gst.NewElement("ximagesrc")
	if err != nil {
		return nil, err
	}
	ximagesrc.Set("display-name", displayName)
	ximagesrc.Set("use-damage", false)
	ximagesrc.Set("show-pointer", true)

	crop, err := gst.NewElement("videocrop")
	if err != nil {
		return nil, err
	}
	crop.Set("left", region.X)
	crop.Set("top", region.Y)
	crop.Set("right", maxInt(rootW-(region.X+region.W), 0))
	crop.Set("bottom", maxInt(rootH-(region.Y+region.H), 0))

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
	capsFilter, err := newCapsFilter(fmt.Sprintf("video/x-raw,format=I420,width=%d,height=%d,framerate=%d/1,pixel-aspect-ratio=1/1", region.W, region.H, fps))
	if err != nil {
		return nil, err
	}

	elements := []*gst.Element{ximagesrc, crop, videoconvert, videoscale, videorate, capsFilter}

	switch scenario {
	case "normalized-fakesink":
		fakesink, err := gst.NewElement("fakesink")
		if err != nil {
			return nil, err
		}
		fakesink.Set("sync", false)
		elements = append(elements, fakesink)
	case "appsink-discard", "appsink-copy-discard", "appsink-copy-async-discard", "appsink-copy-async-appsrc-fakesink", "appsink-copy-async-appsrc-x264":
		sink, err := app.NewAppSink()
		if err != nil {
			return nil, err
		}
		sink.SetProperty("name", "bridge-overhead-appsink")
		sink.SetProperty("max-buffers", 2)
		sink.SetProperty("drop", true)
		sink.SetCallbacks(&app.SinkCallbacks{NewSampleFunc: func(s *app.Sink) gst.FlowReturn {
			sample := s.PullSample()
			if sample == nil {
				return gst.FlowEOS
			}
			m.samplesPulled.Add(1)
			buffer := sample.GetBuffer()
			if buffer == nil {
				return gst.FlowError
			}
			switch scenario {
			case "appsink-discard":
				return gst.FlowOK
			case "appsink-copy-discard":
				copied := buffer.Copy()
				if copied == nil {
					return gst.FlowError
				}
				m.buffersCopied.Add(1)
				return gst.FlowOK
			default:
				copied := buffer.Copy()
				if copied == nil {
					return gst.FlowError
				}
				m.buffersCopied.Add(1)
				select {
				case sampleCh <- copied:
					m.enqueued.Add(1)
				default:
					m.dropped.Add(1)
				}
				return gst.FlowOK
			}
		}})
		elements = append(elements, sink.Element)
	default:
		return nil, fmt.Errorf("unsupported scenario %s", scenario)
	}

	if err := pipeline.AddMany(elements...); err != nil {
		return nil, err
	}
	if err := gst.ElementLinkMany(elements...); err != nil {
		return nil, err
	}
	return pipeline, nil
}

func newAppsrcBridge(mode bridgeSinkMode, width, height, fps int, outputPath string) (*appsrcBridge, error) {
	pipeline, err := gst.NewPipeline("bridge-overhead-appsrc")
	if err != nil {
		return nil, err
	}
	src, err := app.NewAppSrc()
	if err != nil {
		return nil, err
	}
	src.SetProperty("name", "bridge-overhead-appsrc")
	src.SetCaps(gst.NewCapsFromString(fmt.Sprintf("video/x-raw,format=I420,width=%d,height=%d,framerate=%d/1,pixel-aspect-ratio=1/1", width, height, fps)))
	src.SetStreamType(app.AppStreamTypeStream)
	src.Set("block", true)
	src.Set("emit-signals", false)
	src.Set("format", int(gst.FormatTime))
	src.Set("is-live", true)
	src.SetFormat(gst.FormatTime)
	src.SetDoTimestamp(true)
	src.SetAutomaticEOS(false)

	elements := []*gst.Element{src.Element}
	switch mode {
	case modeAppsrcFake:
		fakesink, err := gst.NewElement("fakesink")
		if err != nil {
			return nil, err
		}
		fakesink.Set("sync", false)
		elements = append(elements, fakesink)
	case modeAppsrcX264:
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
		elements = append(elements, videoconvert, x264enc, h264parse, mux, filesink)
	default:
		return nil, fmt.Errorf("unsupported bridge mode %s", mode)
	}

	if err := pipeline.AddMany(elements...); err != nil {
		return nil, err
	}
	if err := gst.ElementLinkMany(elements...); err != nil {
		return nil, err
	}
	return &appsrcBridge{
		mode:          mode,
		pipeline:      pipeline,
		appsrc:        src,
		resultCh:      make(chan error, 1),
		frameDuration: time.Second / time.Duration(fps),
	}, nil
}

func (b *appsrcBridge) pushBuffer(buf *gst.Buffer, m *metrics) error {
	if b == nil || buf == nil {
		return nil
	}
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
	m.appsrcPushed.Add(1)
	return nil
}

func (b *appsrcBridge) stop() error {
	if b == nil {
		return nil
	}
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
		_ = b.pipeline.BlockSetState(gst.StateNull)
		return err
	}
	return b.pipeline.BlockSetState(gst.StateNull)
}

func drainBuffers(ch <-chan *gst.Buffer, done chan<- struct{}, m *metrics) {
	defer close(done)
	for buf := range ch {
		if buf == nil {
			continue
		}
		m.workerHandled.Add(1)
	}
}

func pushBuffers(ch <-chan *gst.Buffer, done chan<- struct{}, bridge *appsrcBridge, m *metrics) {
	defer close(done)
	for buf := range ch {
		if buf == nil {
			continue
		}
		m.workerHandled.Add(1)
		if err := bridge.pushBuffer(buf, m); err != nil {
			fmt.Fprintf(os.Stderr, "bridge push error: %v\n", err)
			return
		}
	}
}

func waitPipeline(pipeline *gst.Pipeline) error {
	if pipeline == nil {
		return nil
	}
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
	return rect{X: vals[0], Y: vals[1], W: vals[2], H: vals[3]}
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
