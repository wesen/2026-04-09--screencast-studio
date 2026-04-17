package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-gst/go-gst/gst"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/discovery"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
)

type summary struct {
	StartedAt       time.Time `json:"started_at"`
	FinishedAt      time.Time `json:"finished_at"`
	SourceType      string    `json:"source_type"`
	DisplayName     string    `json:"display_name"`
	Rect            string    `json:"rect,omitempty"`
	FPS             int       `json:"fps"`
	Quality         int       `json:"quality"`
	Bitrate         int       `json:"bitrate"`
	Container       string    `json:"container"`
	OutputPath      string    `json:"output_path"`
	DurationSeconds int       `json:"duration_seconds"`
	Result          string    `json:"result,omitempty"`
	Error           string    `json:"error,omitempty"`
}

func main() {
	sourceType := flag.String("source-type", "display", "display|region|window")
	displayName := flag.String("display-name", strings.TrimSpace(envOr("DISPLAY", ":0")), "X11 display name")
	rectValue := flag.String("rect", "", "optional region/window rect x,y,w,h")
	windowID := flag.String("window-id", "", "window id for window source")
	fps := flag.Int("fps", 24, "capture fps")
	quality := flag.Int("quality", 75, "recording quality")
	container := flag.String("container", "mov", "mov|qt|mp4")
	durationSeconds := flag.Int("duration-seconds", 8, "seconds to record before EOS")
	outputPath := flag.String("output-path", filepath.Join(os.TempDir(), "scs-direct-full-desktop.mov"), "output path")
	dotDir := flag.String("dot-dir", "", "directory for GStreamer .dot graph dumps")
	flag.Parse()

	s := summary{
		StartedAt:       time.Now(),
		SourceType:      strings.TrimSpace(*sourceType),
		DisplayName:     strings.TrimSpace(*displayName),
		Rect:            strings.TrimSpace(*rectValue),
		FPS:             *fps,
		Quality:         *quality,
		Bitrate:         qualityToBitrate(*quality),
		Container:       strings.TrimSpace(*container),
		OutputPath:      *outputPath,
		DurationSeconds: *durationSeconds,
	}
	defer func() {
		s.FinishedAt = time.Now()
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(&s)
	}()

	if s.SourceType == "" {
		s.SourceType = "display"
	}
	if s.FPS <= 0 {
		s.Error = "fps must be > 0"
		return
	}
	if s.DurationSeconds <= 0 {
		s.Error = "duration-seconds must be > 0"
		return
	}
	if s.Quality <= 0 {
		s.Quality = 75
		s.Bitrate = qualityToBitrate(s.Quality)
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

	source, err := buildSource(s.SourceType, s.DisplayName, strings.TrimSpace(*windowID), s.Rect, s.FPS, s.Container, s.Quality)
	if err != nil {
		s.Error = err.Error()
		return
	}
	pipeline, err := buildVideoRecordingPipeline(source, s.OutputPath)
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

	dumpGraph(pipeline, "direct-full-desktop-pre-play")
	if err := pipeline.SetState(gst.StatePlaying); err != nil {
		s.Error = fmt.Sprintf("start pipeline: %v", err)
		return
	}
	time.Sleep(1 * time.Second)
	dumpGraph(pipeline, "direct-full-desktop-playing")

	time.Sleep(time.Duration(s.DurationSeconds-1) * time.Second)
	sendEOS(pipeline)

	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		msg := bus.TimedPopFiltered(gst.ClockTime(500*time.Millisecond), gst.MessageEOS|gst.MessageError)
		if msg == nil {
			continue
		}
		switch msg.Type() {
		case gst.MessageEOS:
			dumpGraph(pipeline, "direct-full-desktop-eos")
			s.Result = "eos"
			return
		case gst.MessageError:
			dumpGraph(pipeline, "direct-full-desktop-error")
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

func buildSource(sourceType, displayName, windowID, rectValue string, fps int, container string, quality int) (dsl.EffectiveVideoSource, error) {
	source := dsl.EffectiveVideoSource{
		ID:      "direct-record-source-1",
		Name:    "Direct Record Source",
		Type:    strings.TrimSpace(sourceType),
		Enabled: true,
		Target: dsl.VideoTarget{
			Display:  strings.TrimSpace(displayName),
			WindowID: strings.TrimSpace(windowID),
		},
		Capture: dsl.VideoCaptureSettings{
			FPS:    fps,
			Cursor: boolPtr(true),
		},
		Output: dsl.VideoOutputSettings{
			Container:  strings.TrimSpace(container),
			VideoCodec: "h264",
			Quality:    quality,
		},
	}
	if source.Target.Display == "" {
		source.Target.Display = ":0"
	}
	if strings.TrimSpace(rectValue) != "" {
		rect, err := parseRect(rectValue)
		if err != nil {
			return source, err
		}
		source.Target.Rect = rect
		if source.Type == "display" {
			source.Type = "region"
		}
	}
	switch source.Type {
	case "display":
		return source, nil
	case "region":
		if source.Target.Rect == nil {
			return source, errors.New("region source requires --rect x,y,w,h")
		}
		return source, nil
	case "window":
		if source.Target.WindowID == "" {
			return source, errors.New("window source requires --window-id")
		}
		if source.Target.Rect == nil {
			return source, errors.New("window source currently requires --rect x,y,w,h")
		}
		return source, nil
	default:
		return source, fmt.Errorf("unsupported source type %q", source.Type)
	}
}

func buildVideoRecordingPipeline(source dsl.EffectiveVideoSource, outputPath string) (*gst.Pipeline, error) {
	pipeline, err := gst.NewPipeline("")
	if err != nil {
		return nil, fmt.Errorf("create pipeline: %w", err)
	}

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

	fps := source.Capture.FPS
	if fps <= 0 {
		fps = 24
	}
	videorate, err := gst.NewElement("videorate")
	if err != nil {
		return nil, fmt.Errorf("create videorate: %w", err)
	}
	fpsCaps, err := newCapsFilter(fmt.Sprintf("video/x-raw,format=I420,framerate=%d/1,pixel-aspect-ratio=1/1", fps))
	if err != nil {
		return nil, err
	}
	x264enc, err := gst.NewElement("x264enc")
	if err != nil {
		return nil, fmt.Errorf("create x264enc: %w", err)
	}
	x264enc.Set("bitrate", qualityToBitrate(source.Output.Quality))
	x264enc.Set("bframes", 0)
	x264enc.Set("tune", 4)
	x264enc.Set("speed-preset", 3)
	h264parse, err := gst.NewElement("h264parse")
	if err != nil {
		return nil, fmt.Errorf("create h264parse: %w", err)
	}
	mux, err := newVideoMuxer(source.Output.Container)
	if err != nil {
		return nil, err
	}
	filesink, err := gst.NewElement("filesink")
	if err != nil {
		return nil, fmt.Errorf("create filesink: %w", err)
	}
	filesink.Set("location", outputPath)

	elements = append(elements, videorate, fpsCaps, x264enc, h264parse, mux, filesink)
	pipeline.AddMany(elements...)
	if err := linkElements(elements...); err != nil {
		pipeline.BlockSetState(gst.StateNull)
		return nil, err
	}
	return pipeline, nil
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
			crop, err := buildVideoCrop(source)
			if err != nil {
				return nil, err
			}
			elements = append(elements, crop)
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
	return crop, nil
}

func newCapsFilter(caps string) (*gst.Element, error) {
	element, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, fmt.Errorf("create capsfilter: %w", err)
	}
	element.Set("caps", gst.NewCapsFromString(caps))
	return element, nil
}

func linkElements(elements ...*gst.Element) error {
	for i := 0; i < len(elements)-1; i++ {
		if err := elements[i].Link(elements[i+1]); err != nil {
			return fmt.Errorf("link %s -> %s: %w", elements[i].GetName(), elements[i+1].GetName(), err)
		}
	}
	return nil
}

func newVideoMuxer(container string) (*gst.Element, error) {
	switch strings.ToLower(strings.TrimSpace(container)) {
	case "", "mp4":
		mux, err := gst.NewElement("mp4mux")
		if err != nil {
			return nil, fmt.Errorf("create mp4mux: %w", err)
		}
		return mux, nil
	case "mov", "qt":
		mux, err := gst.NewElement("qtmux")
		if err != nil {
			return nil, fmt.Errorf("create qtmux: %w", err)
		}
		return mux, nil
	default:
		return nil, fmt.Errorf("unsupported gstreamer video container %q", container)
	}
}

func qualityToBitrate(quality int) int {
	if quality < 1 {
		quality = 75
	}
	if quality > 100 {
		quality = 100
	}
	return 1000 + (quality-1)*80
}

func sendEOS(pipeline *gst.Pipeline) {
	if pipeline == nil {
		return
	}
	pipeline.SendEvent(gst.NewEOSEvent())
}

func parseRect(value string) (*dsl.Rect, error) {
	parts := strings.Split(strings.TrimSpace(value), ",")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid rect %q, want x,y,w,h", value)
	}
	vals := make([]int, 0, 4)
	for _, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("parse rect component %q: %w", part, err)
		}
		vals = append(vals, n)
	}
	if vals[2] <= 0 || vals[3] <= 0 {
		return nil, fmt.Errorf("invalid rect %q, width and height must be > 0", value)
	}
	return &dsl.Rect{X: vals[0], Y: vals[1], W: vals[2], H: vals[3]}, nil
}

func boolPtr(v bool) *bool { return &v }

func boolValue(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
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
