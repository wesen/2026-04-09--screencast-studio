package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
)

func main() {
	gst.Init(nil)

	outPath := filepath.Join(os.TempDir(), "gst-shared-video-tee-experiment.mp4")
	_ = os.Remove(outPath)
	fmt.Printf("output: %s\n", outPath)

	pipeline, err := gst.NewPipeline("shared-video-tee")
	must(err, "new pipeline")

	src, err := gst.NewElementWithName("videotestsrc", "src")
	must(err, "videotestsrc")
	src.Set("is-live", true)
	src.Set("pattern", 0)

	convert, err := gst.NewElementWithName("videoconvert", "convert")
	must(err, "videoconvert")
	tee, err := gst.NewElementWithName("tee", "tee")
	must(err, "tee")

	must(pipeline.AddMany(src, convert, tee), "add source chain")
	must(gst.ElementLinkMany(src, convert, tee), "link source chain")

	previewBranch, err := newPreviewBranch()
	must(err, "new preview branch")
	must(addBranch(pipeline, tee, previewBranch), "add preview branch")

	recordBranch, err := newRecordingBranch(outPath)
	must(err, "new recording branch")
	must(addBranch(pipeline, tee, recordBranch), "add recording branch")

	bus := pipeline.GetPipelineBus()
	must(pipeline.SetState(gst.StatePlaying), "set pipeline playing")
	fmt.Println("pipeline started")

	start := time.Now()
	for time.Since(start) < 3*time.Second {
		drainBus(bus)
		time.Sleep(50 * time.Millisecond)
	}
	fmt.Printf("preview frames before stop: %d\n", previewBranch.frames.Load())

	stopMode := os.Getenv("STOP_MODE")
	if stopMode == "" {
		stopMode = "queue-sink-pad"
	}
	fmt.Printf("stop mode: %s\n", stopMode)
	if !sendRecordingStop(recordBranch, stopMode) {
		fmt.Printf("warning: stop mode %s rejected EOS\n", stopMode)
	}

	start = time.Now()
	for time.Since(start) < 3*time.Second {
		drainBus(bus)
		time.Sleep(50 * time.Millisecond)
	}
	fmt.Printf("preview frames after branch EOS: %d\n", previewBranch.frames.Load())

	fmt.Println("tearing down recording branch")
	must(removeBranch(pipeline, tee, recordBranch), "remove recording branch")

	start = time.Now()
	for time.Since(start) < 2*time.Second {
		drainBus(bus)
		time.Sleep(50 * time.Millisecond)
	}
	fmt.Printf("preview frames after branch removal: %d\n", previewBranch.frames.Load())

	fmt.Println("stopping pipeline")
	pipeline.BlockSetState(gst.StateNull)

	if info, err := os.Stat(outPath); err == nil {
		fmt.Printf("output size: %d bytes\n", info.Size())
		probe := exec.Command("ffprobe", "-hide_banner", "-loglevel", "error", "-show_entries", "format=duration,size", "-of", "default=noprint_wrappers=1:nokey=0", outPath)
		probe.Stdout = os.Stdout
		probe.Stderr = os.Stderr
		if err := probe.Run(); err != nil {
			fmt.Printf("ffprobe failed: %v\n", err)
		}
	} else {
		fmt.Printf("output stat failed: %v\n", err)
	}
}

type previewBranch struct {
	elements []*gst.Element
	queue    *gst.Element
	sink     *app.Sink
	teePad   *gst.Pad
	frames   atomic.Uint64
}

type recordingBranch struct {
	elements  []*gst.Element
	queue     *gst.Element
	mux       *gst.Element
	filesink  *gst.Element
	x264enc   *gst.Element
	videorate *gst.Element
	teePad    *gst.Pad
}

func newPreviewBranch() (*previewBranch, error) {
	queue, err := gst.NewElementWithName("queue", "preview-queue")
	if err != nil {
		return nil, err
	}
	queue.Set("leaky", 2)
	queue.Set("max-size-buffers", 2)
	videorate, err := gst.NewElementWithName("videorate", "preview-videorate")
	if err != nil {
		return nil, err
	}
	capsRate, err := gst.NewElementWithName("capsfilter", "preview-rate-caps")
	if err != nil {
		return nil, err
	}
	capsRate.Set("caps", gst.NewCapsFromString("video/x-raw,framerate=5/1"))
	jpegenc, err := gst.NewElementWithName("jpegenc", "preview-jpegenc")
	if err != nil {
		return nil, err
	}
	sink, err := app.NewAppSink()
	if err != nil {
		return nil, err
	}
	sink.SetProperty("name", "preview-appsink")
	sink.SetCaps(gst.NewCapsFromString("image/jpeg"))
	sink.SetProperty("max-buffers", 2)
	sink.SetProperty("drop", true)

	branch := &previewBranch{elements: []*gst.Element{queue, videorate, capsRate, jpegenc, sink.Element}, queue: queue, sink: sink}
	sink.SetCallbacks(&app.SinkCallbacks{NewSampleFunc: func(s *app.Sink) gst.FlowReturn {
		sample := s.PullSample()
		if sample == nil {
			return gst.FlowEOS
		}
		branch.frames.Add(1)
		return gst.FlowOK
	}})
	return branch, nil
}

func newRecordingBranch(outPath string) (*recordingBranch, error) {
	queue, err := gst.NewElementWithName("queue", "record-queue")
	if err != nil {
		return nil, err
	}
	videorate, err := gst.NewElementWithName("videorate", "record-videorate")
	if err != nil {
		return nil, err
	}
	capsRate, err := gst.NewElementWithName("capsfilter", "record-rate-caps")
	if err != nil {
		return nil, err
	}
	capsRate.Set("caps", gst.NewCapsFromString("video/x-raw,framerate=10/1"))
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
	return &recordingBranch{elements: []*gst.Element{queue, videorate, capsRate, x264enc, mux, filesink}, queue: queue, videorate: videorate, x264enc: x264enc, mux: mux, filesink: filesink}, nil
}

func addBranch(pipeline *gst.Pipeline, tee *gst.Element, branch interface {
	getElements() []*gst.Element
	getQueue() *gst.Element
	setTeePad(*gst.Pad)
}) error {
	elements := branch.getElements()
	if err := pipeline.AddMany(elements...); err != nil {
		return err
	}
	if err := gst.ElementLinkMany(elements...); err != nil {
		return err
	}
	teePad := tee.GetRequestPad("src_%u")
	if teePad == nil {
		return fmt.Errorf("tee request pad failed")
	}
	sinkPad := branch.getQueue().GetStaticPad("sink")
	if sinkPad == nil {
		return fmt.Errorf("branch queue sink pad missing")
	}
	if ret := teePad.Link(sinkPad); ret != gst.PadLinkOK {
		return fmt.Errorf("tee link failed: %s", ret.String())
	}
	branch.setTeePad(teePad)
	for _, element := range elements {
		if !element.SyncStateWithParent() {
			return fmt.Errorf("sync state with parent failed for %s", element.GetName())
		}
	}
	return nil
}

func removeBranch(pipeline *gst.Pipeline, tee *gst.Element, branch interface {
	getElements() []*gst.Element
	getQueue() *gst.Element
	getTeePad() *gst.Pad
}) error {
	if branch.getTeePad() != nil {
		branch.getTeePad().Unlink(branch.getQueue().GetStaticPad("sink"))
		tee.ReleaseRequestPad(branch.getTeePad())
	}
	for _, element := range branch.getElements() {
		element.SetState(gst.StateNull)
	}
	return pipeline.RemoveMany(branch.getElements()...)
}

func sendRecordingStop(branch *recordingBranch, mode string) bool {
	switch mode {
	case "queue-sink-pad":
		fmt.Println("sending EOS to recording queue sink pad")
		return branch.queue.GetStaticPad("sink").SendEvent(gst.NewEOSEvent())
	case "queue-src-pad":
		fmt.Println("sending EOS to recording queue src pad")
		return branch.queue.GetStaticPad("src").SendEvent(gst.NewEOSEvent())
	case "videorate":
		fmt.Println("sending EOS to videorate element")
		return branch.videorate.SendEvent(gst.NewEOSEvent())
	case "encoder":
		fmt.Println("sending EOS to x264enc element")
		return branch.x264enc.SendEvent(gst.NewEOSEvent())
	case "mux":
		fmt.Println("sending EOS to mp4mux element")
		return branch.mux.SendEvent(gst.NewEOSEvent())
	case "filesink":
		fmt.Println("sending EOS to filesink element")
		return branch.filesink.SendEvent(gst.NewEOSEvent())
	default:
		fmt.Printf("unknown stop mode %q, defaulting to queue-sink-pad\n", mode)
		return branch.queue.GetStaticPad("sink").SendEvent(gst.NewEOSEvent())
	}
}

func drainBus(bus *gst.Bus) {
	for {
		msg := bus.Pop()
		if msg == nil {
			return
		}
		switch msg.Type() {
		case gst.MessageError:
			fmt.Printf("BUS error from %s: %v\n", msg.Source(), msg.ParseError())
		case gst.MessageEOS:
			fmt.Printf("BUS eos from %s\n", msg.Source())
		case gst.MessageStateChanged:
			old, newState := msg.ParseStateChanged()
			name := msg.Source()
			if name == "shared-video-tee" || name == "record-filesink" || name == "record-mux" || name == "preview-appsink" {
				fmt.Printf("BUS state changed %s: %s -> %s\n", name, old, newState)
			}
		}
	}
}

func (b *previewBranch) getElements() []*gst.Element { return b.elements }
func (b *previewBranch) getQueue() *gst.Element      { return b.queue }
func (b *previewBranch) getTeePad() *gst.Pad         { return b.teePad }
func (b *previewBranch) setTeePad(p *gst.Pad)        { b.teePad = p }

func (b *recordingBranch) getElements() []*gst.Element { return b.elements }
func (b *recordingBranch) getQueue() *gst.Element      { return b.queue }
func (b *recordingBranch) getTeePad() *gst.Pad         { return b.teePad }
func (b *recordingBranch) setTeePad(p *gst.Pad)        { b.teePad = p }

func must(err error, label string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", label, err)
		os.Exit(1)
	}
}
