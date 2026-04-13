package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
)

const (
	width  = 320
	height = 240
	fps    = 10
	frames = 30
)

func main() {
	gst.Init(nil)

	outPath := filepath.Join(os.TempDir(), "gst-appsrc-mp4-recorder-smoke.mp4")
	_ = os.Remove(outPath)

	pipeline, src, err := buildPipeline(outPath)
	must(err, "build pipeline")
	defer func() { _ = pipeline.BlockSetState(gst.StateNull) }()

	loop := glib.NewMainLoop(glib.MainContextDefault(), false)
	defer loop.Quit()

	resultCh := make(chan error, 1)
	if !pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageError:
			resultCh <- fmt.Errorf("pipeline error: %w", msg.ParseError())
			loop.Quit()
			return false
		case gst.MessageEOS:
			resultCh <- nil
			loop.Quit()
			return false
		default:
			return true
		}
	}) {
		fatal("failed to add bus watch")
	}

	if err := pipeline.SetState(gst.StatePlaying); err != nil {
		fatal("set pipeline playing: %v", err)
	}

	go func() {
		defer func() {
			if ret := src.EndStream(); ret != gst.FlowOK && ret != gst.FlowEOS {
				fmt.Fprintf(os.Stderr, "end stream returned %s\n", ret)
			}
		}()
		for i := 0; i < frames; i++ {
			buf := gst.NewBufferWithSize(width * height * 4)
			buf.SetPresentationTimestamp(gst.ClockTime(time.Duration(i) * time.Second / fps))
			buf.SetDuration(gst.ClockTime(time.Second / fps))
			pixels := makeRGBAFrame(i)
			buf.Map(gst.MapWrite).WriteData(pixels)
			buf.Unmap()
			ret := src.PushBuffer(buf)
			fmt.Printf("push frame=%d ret=%s pts=%s\n", i, ret, time.Duration(buf.PresentationTimestamp()))
			if ret != gst.FlowOK {
				resultCh <- fmt.Errorf("push frame %d returned %s", i, ret)
				loop.Quit()
				return
			}
			time.Sleep(time.Second / fps)
		}
	}()

	loop.Run()

	select {
	case err := <-resultCh:
		must(err, "pipeline run")
	case <-time.After(10 * time.Second):
		fatal("timed out waiting for EOS")
	}

	info, err := os.Stat(outPath)
	must(err, "stat output")
	fmt.Printf("output size: %d bytes\n", info.Size())
	probe := exec.Command("ffprobe", "-hide_banner", "-loglevel", "error", "-show_entries", "format=duration,size", "-of", "default=noprint_wrappers=1:nokey=0", outPath)
	probe.Stdout = os.Stdout
	probe.Stderr = os.Stderr
	if err := probe.Run(); err != nil {
		fmt.Printf("ffprobe failed: %v\n", err)
	}
}

func buildPipeline(outPath string) (*gst.Pipeline, *app.Source, error) {
	pipeline, err := gst.NewPipeline("appsrc-mp4-smoke")
	if err != nil {
		return nil, nil, err
	}
	src, err := app.NewAppSrc()
	if err != nil {
		return nil, nil, err
	}
	src.SetProperty("name", "test-appsrc")
	src.SetCaps(gst.NewCapsFromString(fmt.Sprintf("video/x-raw,format=RGBA,width=%d,height=%d,framerate=%d/1,pixel-aspect-ratio=1/1", width, height, fps)))
	src.SetStreamType(app.AppStreamTypeStream)
	src.Set("format", int(gst.FormatTime))
	src.Set("block", true)
	src.Set("emit-signals", false)
	src.SetFormat(gst.FormatTime)
	src.SetAutomaticEOS(false)

	videoconvert, err := gst.NewElementWithName("videoconvert", "test-videoconvert")
	if err != nil {
		return nil, nil, err
	}
	x264enc, err := gst.NewElementWithName("x264enc", "test-x264enc")
	if err != nil {
		return nil, nil, err
	}
	x264enc.Set("bitrate", 2500)
	x264enc.Set("bframes", 0)
	x264enc.Set("tune", 4)
	x264enc.Set("speed-preset", 3)
	h264parse, err := gst.NewElementWithName("h264parse", "test-h264parse")
	if err != nil {
		return nil, nil, err
	}
	mux, err := gst.NewElementWithName("mp4mux", "test-mp4mux")
	if err != nil {
		return nil, nil, err
	}
	filesink, err := gst.NewElementWithName("filesink", "test-filesink")
	if err != nil {
		return nil, nil, err
	}
	filesink.Set("location", outPath)

	if err := pipeline.AddMany(src.Element, videoconvert, x264enc, h264parse, mux, filesink); err != nil {
		return nil, nil, err
	}
	if err := gst.ElementLinkMany(src.Element, videoconvert, x264enc, h264parse, mux, filesink); err != nil {
		return nil, nil, err
	}
	return pipeline, src, nil
}

func makeRGBAFrame(i int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	c := color.RGBA{R: uint8((i * 17) % 255), G: uint8((i * 31) % 255), B: uint8((i * 47) % 255), A: 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	return img.Pix
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
