package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
)

func main() {
	gst.Init(nil)

	outPath := filepath.Join(os.TempDir(), "gst-shared-source-appsink-appsrc-bridge.mp4")
	_ = os.Remove(outPath)
	fmt.Printf("output: %s\n", outPath)

	recorder, err := newRecorder(outPath)
	must(err, "new recorder")
	must(recorder.pipeline.SetState(gst.StatePlaying), "start recorder pipeline")

	sourceGraph, err := newSourceGraph(recorder)
	must(err, "new source graph")
	must(sourceGraph.pipeline.SetState(gst.StatePlaying), "start source pipeline")
	fmt.Println("pipelines started")

	start := time.Now()
	for time.Since(start) < 3*time.Second {
		drainBus("source", sourceGraph.pipeline.GetPipelineBus())
		drainBus("record", recorder.pipeline.GetPipelineBus())
		time.Sleep(50 * time.Millisecond)
	}
	before := sourceGraph.previewFrames.Load()
	fmt.Printf("preview frames before stop: %d\n", before)

	fmt.Println("ending recorder appsrc stream")
	recorder.closed.Store(true)
	if ret := recorder.appsrc.EndStream(); ret != gst.FlowOK {
		fmt.Printf("warning: appsrc end stream returned %s\n", ret)
	}

	waitForEOS("record", recorder.pipeline.GetPipelineBus(), 5*time.Second)
	must(recorder.pipeline.BlockSetState(gst.StateNull), "stop recorder pipeline")
	fmt.Println("recorder pipeline stopped")

	start = time.Now()
	for time.Since(start) < 2*time.Second {
		drainBus("source", sourceGraph.pipeline.GetPipelineBus())
		time.Sleep(50 * time.Millisecond)
	}
	after := sourceGraph.previewFrames.Load()
	fmt.Printf("preview frames after recorder stop: %d\n", after)
	if after <= before {
		fmt.Fprintf(os.Stderr, "preview did not continue after recorder stop\n")
		os.Exit(1)
	}

	must(sourceGraph.pipeline.BlockSetState(gst.StateNull), "stop source pipeline")

	info, err := os.Stat(outPath)
	must(err, "stat output")
	fmt.Printf("output size: %d bytes\n", info.Size())
	probe := exec.Command("ffprobe", "-hide_banner", "-loglevel", "error", "-show_entries", "format=duration,size", "-of", "default=noprint_wrappers=1:nokey=0", outPath)
	probe.Stdout = os.Stdout
	probe.Stderr = os.Stderr
	must(probe.Run(), "ffprobe output")
}

type recorder struct {
	pipeline   *gst.Pipeline
	appsrc     *app.Source
	capsOnce   sync.Once
	frameCount uint64
	closed     atomic.Bool
}

func newRecorder(outPath string) (*recorder, error) {
	pipeline, err := gst.NewPipeline("record-bridge")
	if err != nil {
		return nil, err
	}
	appsrc, err := app.NewAppSrc()
	if err != nil {
		return nil, err
	}
	appsrc.SetProperty("name", "record-appsrc")
	appsrc.SetFormat(gst.FormatTime)
	appsrc.SetLive(true)
	appsrc.SetDoTimestamp(true)
	appsrc.Set("block", true)
	appsrc.SetCaps(gst.NewCapsFromString("video/x-raw,format=I420,width=640,height=480,framerate=10/1"))
	videoconvert, err := gst.NewElementWithName("videoconvert", "record-videoconvert")
	if err != nil {
		return nil, err
	}
	x264enc, err := gst.NewElementWithName("x264enc", "record-x264enc")
	if err != nil {
		return nil, err
	}
	x264enc.Set("bitrate", 2500)
	x264enc.Set("bframes", 0)
	x264enc.Set("tune", 4)
	x264enc.Set("speed-preset", 3)
	mux, err := gst.NewElementWithName("mp4mux", "record-mux")
	if err != nil {
		return nil, err
	}
	filesink, err := gst.NewElementWithName("filesink", "record-filesink")
	if err != nil {
		return nil, err
	}
	filesink.Set("location", outPath)

	if err := pipeline.AddMany(appsrc.Element, videoconvert, x264enc, mux, filesink); err != nil {
		return nil, err
	}
	if err := gst.ElementLinkMany(appsrc.Element, videoconvert, x264enc, mux, filesink); err != nil {
		return nil, err
	}
	return &recorder{pipeline: pipeline, appsrc: appsrc}, nil
}

func (r *recorder) pushSample(sample *gst.Sample) gst.FlowReturn {
	if sample == nil {
		return gst.FlowError
	}
	if r.closed.Load() {
		return gst.FlowOK
	}

	r.capsOnce.Do(func() {
		if caps := sample.GetCaps(); caps != nil {
			fmt.Printf("record sample caps: %s\n", caps.String())
			r.appsrc.SetCaps(caps)
		}
	})
	buffer := sample.GetBuffer()
	if buffer == nil {
		return gst.FlowError
	}
	mapInfo := buffer.Map(gst.MapRead)
	if mapInfo == nil {
		return gst.FlowError
	}
	payload := append([]byte(nil), mapInfo.Bytes()...)
	buffer.Unmap()

	outBuffer := gst.NewBufferFromBytes(payload)
	frameIndex := r.frameCount
	r.frameCount++
	frameDur := 100 * time.Millisecond
	outBuffer.SetPresentationTimestamp(gst.ClockTime(time.Duration(frameIndex) * frameDur))
	outBuffer.SetDuration(gst.ClockTime(frameDur))
	ret := r.appsrc.PushBuffer(outBuffer)
	if ret != gst.FlowOK {
		fmt.Printf("appsrc push buffer returned %s\n", ret)
	}
	return ret
}

type sourceGraph struct {
	pipeline      *gst.Pipeline
	previewFrames atomic.Uint64
}

func newSourceGraph(rec *recorder) (*sourceGraph, error) {
	pipeline, err := gst.NewPipeline("source-bridge")
	if err != nil {
		return nil, err
	}
	graph := &sourceGraph{pipeline: pipeline}

	src, err := gst.NewElementWithName("videotestsrc", "src")
	if err != nil {
		return nil, err
	}
	src.Set("is-live", true)
	src.Set("pattern", 0)
	convert, err := gst.NewElementWithName("videoconvert", "convert")
	if err != nil {
		return nil, err
	}
	rawCaps, err := gst.NewElementWithName("capsfilter", "raw-caps")
	if err != nil {
		return nil, err
	}
	rawCaps.Set("caps", gst.NewCapsFromString("video/x-raw,format=I420,width=640,height=480,framerate=10/1"))
	tee, err := gst.NewElementWithName("tee", "tee")
	if err != nil {
		return nil, err
	}
	if err := pipeline.AddMany(src, convert, rawCaps, tee); err != nil {
		return nil, err
	}
	if err := gst.ElementLinkMany(src, convert, rawCaps, tee); err != nil {
		return nil, err
	}

	previewQueue, err := gst.NewElementWithName("queue", "preview-queue")
	if err != nil {
		return nil, err
	}
	previewQueue.Set("leaky", 2)
	previewQueue.Set("max-size-buffers", 2)
	previewRate, err := gst.NewElementWithName("videorate", "preview-videorate")
	if err != nil {
		return nil, err
	}
	previewCaps, err := gst.NewElementWithName("capsfilter", "preview-rate-caps")
	if err != nil {
		return nil, err
	}
	previewCaps.Set("caps", gst.NewCapsFromString("video/x-raw,framerate=5/1"))
	jpegenc, err := gst.NewElementWithName("jpegenc", "preview-jpegenc")
	if err != nil {
		return nil, err
	}
	previewSink, err := app.NewAppSink()
	if err != nil {
		return nil, err
	}
	previewSink.SetProperty("name", "preview-appsink")
	previewSink.SetCaps(gst.NewCapsFromString("image/jpeg"))
	previewSink.SetProperty("max-buffers", 2)
	previewSink.SetProperty("drop", true)
	previewSink.SetCallbacks(&app.SinkCallbacks{NewSampleFunc: func(s *app.Sink) gst.FlowReturn {
		sample := s.PullSample()
		if sample == nil {
			return gst.FlowEOS
		}
		graph.previewFrames.Add(1)
		return gst.FlowOK
	}})

	recordQueue, err := gst.NewElementWithName("queue", "record-queue")
	if err != nil {
		return nil, err
	}
	recordSink, err := app.NewAppSink()
	if err != nil {
		return nil, err
	}
	recordSink.SetProperty("name", "record-appsink")
	recordSink.SetCaps(gst.NewCapsFromString("video/x-raw,format=I420,width=640,height=480,framerate=10/1"))
	recordSink.SetProperty("max-buffers", 4)
	recordSink.SetProperty("drop", false)
	recordSink.SetCallbacks(&app.SinkCallbacks{NewSampleFunc: func(s *app.Sink) gst.FlowReturn {
		sample := s.PullSample()
		if sample == nil {
			return gst.FlowEOS
		}
		return rec.pushSample(sample)
	}})

	if err := pipeline.AddMany(previewQueue, previewRate, previewCaps, jpegenc, previewSink.Element, recordQueue, recordSink.Element); err != nil {
		return nil, err
	}
	if err := gst.ElementLinkMany(previewQueue, previewRate, previewCaps, jpegenc, previewSink.Element); err != nil {
		return nil, err
	}
	if err := gst.ElementLinkMany(recordQueue, recordSink.Element); err != nil {
		return nil, err
	}
	teePreviewPad := tee.GetRequestPad("src_%u")
	if teePreviewPad == nil {
		return nil, fmt.Errorf("preview tee pad request failed")
	}
	if ret := teePreviewPad.Link(previewQueue.GetStaticPad("sink")); ret != gst.PadLinkOK {
		return nil, fmt.Errorf("link preview branch: %s", ret.String())
	}
	teeRecordPad := tee.GetRequestPad("src_%u")
	if teeRecordPad == nil {
		return nil, fmt.Errorf("record tee pad request failed")
	}
	if ret := teeRecordPad.Link(recordQueue.GetStaticPad("sink")); ret != gst.PadLinkOK {
		return nil, fmt.Errorf("link record branch: %s", ret.String())
	}
	return graph, nil
}

func waitForEOS(label string, bus *gst.Bus, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for {
		msg := bus.TimedPop(gst.ClockTime(200 * time.Millisecond))
		if msg != nil {
			switch msg.Type() {
			case gst.MessageEOS:
				fmt.Printf("%s bus eos from %s\n", label, msg.Source())
				return
			case gst.MessageError:
				gerr := msg.ParseError()
				fmt.Fprintf(os.Stderr, "%s bus error from %s: %v\n", label, msg.Source(), gerr)
				if debug := gerr.DebugString(); debug != "" {
					fmt.Fprintf(os.Stderr, "%s bus debug: %s\n", label, debug)
				}
				os.Exit(1)
			case gst.MessageStateChanged:
				old, newState := msg.ParseStateChanged()
				name := msg.Source()
				if name == "record-bridge" || name == "record-mux" || name == "record-filesink" {
					fmt.Printf("%s bus state changed %s: %s -> %s\n", label, name, old, newState)
				}
			}
		}
		if time.Now().After(deadline) {
			fmt.Fprintf(os.Stderr, "timed out waiting for %s EOS\n", label)
			os.Exit(1)
		}
	}
}

func drainBus(label string, bus *gst.Bus) {
	for {
		msg := bus.Pop()
		if msg == nil {
			return
		}
		switch msg.Type() {
		case gst.MessageError:
			gerr := msg.ParseError()
			fmt.Fprintf(os.Stderr, "%s bus error from %s: %v\n", label, msg.Source(), gerr)
			if debug := gerr.DebugString(); debug != "" {
				fmt.Fprintf(os.Stderr, "%s bus debug: %s\n", label, debug)
			}
			os.Exit(1)
		case gst.MessageStateChanged:
			old, newState := msg.ParseStateChanged()
			name := msg.Source()
			if name == "source-bridge" || name == "preview-appsink" || name == "record-appsink" || name == "record-bridge" || name == "record-mux" || name == "record-filesink" {
				fmt.Printf("%s bus state changed %s: %s -> %s\n", label, name, old, newState)
			}
		case gst.MessageEOS:
			fmt.Printf("%s bus eos from %s\n", label, msg.Source())
		}
	}
}

func must(err error, label string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", label, err)
		os.Exit(1)
	}
}
