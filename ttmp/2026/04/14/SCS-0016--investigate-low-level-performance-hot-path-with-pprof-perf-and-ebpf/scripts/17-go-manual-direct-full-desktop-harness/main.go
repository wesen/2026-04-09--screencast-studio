package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-gst/go-gst/gst"
)

type summary struct {
	StartedAt       time.Time `json:"started_at"`
	FinishedAt      time.Time `json:"finished_at"`
	DisplayName     string    `json:"display_name"`
	FPS             int       `json:"fps"`
	Bitrate         int       `json:"bitrate"`
	Container       string    `json:"container"`
	OutputPath      string    `json:"output_path"`
	DurationSeconds int       `json:"duration_seconds"`
	Result          string    `json:"result,omitempty"`
	Error           string    `json:"error,omitempty"`
}

func main() {
	displayName := flag.String("display-name", strings.TrimSpace(envOr("DISPLAY", ":0")), "X11 display name")
	fps := flag.Int("fps", 24, "capture fps")
	bitrate := flag.Int("bitrate", 6920, "x264 bitrate in kbps")
	container := flag.String("container", "mov", "mov|qt|mp4")
	durationSeconds := flag.Int("duration-seconds", 8, "seconds to record before EOS")
	outputPath := flag.String("output-path", filepath.Join(os.TempDir(), "scs-manual-direct-full-desktop.mov"), "output path")
	dotDir := flag.String("dot-dir", "", "directory for GStreamer .dot graph dumps")
	flag.Parse()

	s := summary{
		StartedAt:       time.Now(),
		DisplayName:     *displayName,
		FPS:             *fps,
		Bitrate:         *bitrate,
		Container:       *container,
		OutputPath:      *outputPath,
		DurationSeconds: *durationSeconds,
	}
	defer func() {
		s.FinishedAt = time.Now()
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(&s)
	}()

	if s.FPS <= 0 {
		s.Error = "fps must be > 0"
		return
	}
	if s.DurationSeconds <= 0 {
		s.Error = "duration-seconds must be > 0"
		return
	}
	if err := os.MkdirAll(filepath.Dir(s.OutputPath), 0o755); err != nil {
		s.Error = fmt.Sprintf("mkdir output dir: %v", err)
		return
	}
	if strings.TrimSpace(*dotDir) != "" {
		if err := os.MkdirAll(strings.TrimSpace(*dotDir), 0o755); err != nil {
			s.Error = fmt.Sprintf("mkdir dot dir: %v", err)
			return
		}
		_ = os.Setenv("GST_DEBUG_DUMP_DOT_DIR", strings.TrimSpace(*dotDir))
	}

	gst.Init(nil)

	pipeline, err := buildPipeline(s.DisplayName, s.FPS, s.Bitrate, s.Container, s.OutputPath)
	if err != nil {
		s.Error = err.Error()
		return
	}
	defer pipeline.BlockSetState(gst.StateNull)

	bus := pipeline.GetPipelineBus()
	if bus == nil {
		s.Error = "pipeline bus is nil"
		return
	}

	dumpGraph(pipeline, "manual-direct-full-desktop-pre-play")
	if err := pipeline.SetState(gst.StatePlaying); err != nil {
		s.Error = fmt.Sprintf("start pipeline: %v", err)
		return
	}
	time.Sleep(1 * time.Second)
	dumpGraph(pipeline, "manual-direct-full-desktop-playing")

	time.Sleep(time.Duration(s.DurationSeconds-1) * time.Second)
	pipeline.SendEvent(gst.NewEOSEvent())

	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		msg := bus.TimedPopFiltered(gst.ClockTime(500*time.Millisecond), gst.MessageEOS|gst.MessageError)
		if msg == nil {
			continue
		}
		switch msg.Type() {
		case gst.MessageEOS:
			dumpGraph(pipeline, "manual-direct-full-desktop-eos")
			s.Result = "eos"
			return
		case gst.MessageError:
			dumpGraph(pipeline, "manual-direct-full-desktop-error")
			gerr := msg.ParseError()
			if gerr != nil {
				s.Error = gerr.Error()
			} else {
				s.Error = "gstreamer error"
			}
			return
		}
	}

	s.Error = "timed out waiting for EOS"
}

func buildPipeline(displayName string, fps int, bitrate int, container string, outputPath string) (*gst.Pipeline, error) {
	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return nil, fmt.Errorf("create pipeline: %w", err)
	}

	ximagesrc, err := gst.NewElement("ximagesrc")
	if err != nil {
		return nil, fmt.Errorf("create ximagesrc: %w", err)
	}
	if strings.TrimSpace(displayName) != "" {
		ximagesrc.Set("display-name", strings.TrimSpace(displayName))
	}
	ximagesrc.Set("show-pointer", true)
	ximagesrc.Set("use-damage", false)

	videoconvert, err := gst.NewElement("videoconvert")
	if err != nil {
		return nil, fmt.Errorf("create videoconvert: %w", err)
	}
	videorate, err := gst.NewElement("videorate")
	if err != nil {
		return nil, fmt.Errorf("create videorate: %w", err)
	}
	capsfilter, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, fmt.Errorf("create capsfilter: %w", err)
	}
	capsfilter.Set("caps", gst.NewCapsFromString(fmt.Sprintf("video/x-raw,format=I420,framerate=%d/1,pixel-aspect-ratio=1/1", fps)))

	x264enc, err := gst.NewElement("x264enc")
	if err != nil {
		return nil, fmt.Errorf("create x264enc: %w", err)
	}
	x264enc.Set("bitrate", bitrate)
	x264enc.Set("bframes", 0)
	x264enc.Set("tune", 4)
	x264enc.Set("speed-preset", 3)

	h264parse, err := gst.NewElement("h264parse")
	if err != nil {
		return nil, fmt.Errorf("create h264parse: %w", err)
	}

	muxName := "qtmux"
	switch strings.ToLower(strings.TrimSpace(container)) {
	case "", "mov", "qt":
		muxName = "qtmux"
	case "mp4":
		muxName = "mp4mux"
	default:
		return nil, fmt.Errorf("unsupported container %q", container)
	}
	mux, err := gst.NewElement(muxName)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", muxName, err)
	}
	filesink, err := gst.NewElement("filesink")
	if err != nil {
		return nil, fmt.Errorf("create filesink: %w", err)
	}
	filesink.Set("location", outputPath)

	elements := []*gst.Element{ximagesrc, videoconvert, videorate, capsfilter, x264enc, h264parse, mux, filesink}
	if err := pipeline.AddMany(elements...); err != nil {
		return nil, fmt.Errorf("add elements: %w", err)
	}
	for i := 0; i < len(elements)-1; i++ {
		if err := elements[i].Link(elements[i+1]); err != nil {
			pipeline.BlockSetState(gst.StateNull)
			return nil, fmt.Errorf("link %s -> %s: %w", elements[i].GetName(), elements[i+1].GetName(), err)
		}
	}

	return pipeline, nil
}

func dumpGraph(pipeline *gst.Pipeline, name string) {
	if pipeline == nil || strings.TrimSpace(name) == "" {
		return
	}
	pipeline.DebugBinToDotFileWithTs(gst.DebugGraphShowAll, name)
}

func envOr(key, def string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return def
}
