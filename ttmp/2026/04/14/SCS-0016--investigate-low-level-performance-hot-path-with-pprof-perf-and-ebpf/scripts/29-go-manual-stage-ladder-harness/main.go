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
	Stage           string    `json:"stage"`
	Encoder         string    `json:"encoder"`
	X264SpeedPreset int       `json:"x264_speed_preset"`
	X264Tune        int       `json:"x264_tune"`
	X264BFrames     int       `json:"x264_bframes"`
	X264Trellis     bool      `json:"x264_trellis"`
	FPS             int       `json:"fps"`
	Bitrate         int       `json:"bitrate"`
	Container       string    `json:"container"`
	OutputPath      string    `json:"output_path,omitempty"`
	DurationSeconds int       `json:"duration_seconds"`
	Result          string    `json:"result,omitempty"`
	Error           string    `json:"error,omitempty"`
}

func main() {
	displayName := flag.String("display-name", strings.TrimSpace(envOr("DISPLAY", ":0")), "X11 display name")
	stage := flag.String("stage", "capture", "capture|convert|rate-caps|encode|parse|mux-file")
	fps := flag.Int("fps", 24, "capture fps")
	bitrate := flag.Int("bitrate", 6920, "target bitrate in kbps")
	encoder := flag.String("encoder", "x264enc", "x264enc|openh264enc|vaapih264enc")
	x264SpeedPreset := flag.Int("x264-speed-preset", 3, "x264enc speed-preset enum")
	x264Tune := flag.Int("x264-tune", 4, "x264enc tune flags value")
	x264BFrames := flag.Int("x264-bframes", 0, "x264enc bframes")
	x264Trellis := flag.Bool("x264-trellis", true, "x264enc trellis")
	container := flag.String("container", "mov", "mov|qt|mp4")
	durationSeconds := flag.Int("duration-seconds", 8, "seconds to record before EOS")
	outputPath := flag.String("output-path", filepath.Join(os.TempDir(), "scs-go-stage-ladder.mov"), "output path for mux-file stage")
	dotDir := flag.String("dot-dir", "", "directory for GStreamer .dot graph dumps")
	flag.Parse()

	s := summary{
		StartedAt:       time.Now(),
		DisplayName:     *displayName,
		Stage:           strings.TrimSpace(*stage),
		Encoder:         strings.TrimSpace(*encoder),
		X264SpeedPreset: *x264SpeedPreset,
		X264Tune:        *x264Tune,
		X264BFrames:     *x264BFrames,
		X264Trellis:     *x264Trellis,
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

	if s.Stage == "" {
		s.Stage = "capture"
	}
	if s.FPS <= 0 {
		s.Error = "fps must be > 0"
		return
	}
	if s.DurationSeconds <= 0 {
		s.Error = "duration-seconds must be > 0"
		return
	}
	if strings.TrimSpace(*dotDir) != "" {
		if err := os.MkdirAll(strings.TrimSpace(*dotDir), 0o755); err != nil {
			s.Error = fmt.Sprintf("mkdir dot dir: %v", err)
			return
		}
		_ = os.Setenv("GST_DEBUG_DUMP_DOT_DIR", strings.TrimSpace(*dotDir))
	}
	if s.Stage == "mux-file" {
		if err := os.MkdirAll(filepath.Dir(s.OutputPath), 0o755); err != nil {
			s.Error = fmt.Sprintf("mkdir output dir: %v", err)
			return
		}
	}

	gst.Init(nil)
	pipeline, err := buildPipeline(s.Stage, s.Encoder, s.DisplayName, s.FPS, s.Bitrate, s.X264SpeedPreset, s.X264Tune, s.X264BFrames, s.X264Trellis, s.Container, s.OutputPath)
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

	dumpGraph(pipeline, "go-stage-ladder-"+slug(s.Stage)+"-pre-play")
	if err := pipeline.SetState(gst.StatePlaying); err != nil {
		s.Error = fmt.Sprintf("start pipeline: %v", err)
		return
	}
	time.Sleep(1 * time.Second)
	dumpGraph(pipeline, "go-stage-ladder-"+slug(s.Stage)+"-playing")
	if s.DurationSeconds > 1 {
		time.Sleep(time.Duration(s.DurationSeconds-1) * time.Second)
	}
	pipeline.SendEvent(gst.NewEOSEvent())

	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		msg := bus.TimedPopFiltered(gst.ClockTime(500*time.Millisecond), gst.MessageEOS|gst.MessageError)
		if msg == nil {
			continue
		}
		switch msg.Type() {
		case gst.MessageEOS:
			dumpGraph(pipeline, "go-stage-ladder-"+slug(s.Stage)+"-eos")
			s.Result = "eos"
			return
		case gst.MessageError:
			dumpGraph(pipeline, "go-stage-ladder-"+slug(s.Stage)+"-error")
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

func buildPipeline(stage, encoderName, displayName string, fps, bitrate, x264SpeedPreset, x264Tune, x264BFrames int, x264Trellis bool, container, outputPath string) (*gst.Pipeline, error) {
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

	elems := []*gst.Element{ximagesrc}
	current := ximagesrc

	appendElem := func(name string) (*gst.Element, error) {
		e, err := gst.NewElement(name)
		if err != nil {
			return nil, fmt.Errorf("create %s: %w", name, err)
		}
		elems = append(elems, e)
		current = e
		return e, nil
	}
	_ = current

	stage = strings.TrimSpace(stage)
	switch stage {
	case "capture":
		fakesink, err := appendElem("fakesink")
		if err != nil {
			return nil, err
		}
		fakesink.Set("sync", false)
	case "convert", "rate-caps", "encode", "parse", "mux-file":
		videoconvert, err := appendElem("videoconvert")
		if err != nil {
			return nil, err
		}
		_ = videoconvert
		if stage == "convert" {
			fakesink, err := appendElem("fakesink")
			if err != nil {
				return nil, err
			}
			fakesink.Set("sync", false)
			break
		}
		videorate, err := appendElem("videorate")
		if err != nil {
			return nil, err
		}
		_ = videorate
		capsfilter, err := appendElem("capsfilter")
		if err != nil {
			return nil, err
		}
		capsfilter.Set("caps", gst.NewCapsFromString(fmt.Sprintf("video/x-raw,format=I420,framerate=%d/1,pixel-aspect-ratio=1/1", fps)))
		if stage == "rate-caps" {
			fakesink, err := appendElem("fakesink")
			if err != nil {
				return nil, err
			}
			fakesink.Set("sync", false)
			break
		}
		encoderName = strings.TrimSpace(encoderName)
		if encoderName == "" {
			encoderName = "x264enc"
		}
		encoderElem, err := appendElem(encoderName)
		if err != nil {
			return nil, err
		}
		if err := applyEncoderSettings(encoderElem, encoderName, fps, bitrate, x264SpeedPreset, x264Tune, x264BFrames, x264Trellis); err != nil {
			return nil, err
		}
		if stage == "encode" {
			fakesink, err := appendElem("fakesink")
			if err != nil {
				return nil, err
			}
			fakesink.Set("sync", false)
			break
		}
		h264parse, err := appendElem("h264parse")
		if err != nil {
			return nil, err
		}
		_ = h264parse
		if stage == "parse" {
			fakesink, err := appendElem("fakesink")
			if err != nil {
				return nil, err
			}
			fakesink.Set("sync", false)
			break
		}
		muxName := "qtmux"
		if strings.EqualFold(container, "mp4") {
			muxName = "mp4mux"
		}
		mux, err := appendElem(muxName)
		if err != nil {
			return nil, err
		}
		_ = mux
		filesink, err := appendElem("filesink")
		if err != nil {
			return nil, err
		}
		filesink.Set("location", outputPath)
	default:
		return nil, fmt.Errorf("unsupported stage %q", stage)
	}

	if err := pipeline.AddMany(elems...); err != nil {
		return nil, fmt.Errorf("add elements: %w", err)
	}
	for i := 0; i < len(elems)-1; i++ {
		if err := elems[i].Link(elems[i+1]); err != nil {
			pipeline.BlockSetState(gst.StateNull)
			return nil, fmt.Errorf("link %s -> %s: %w", elems[i].GetName(), elems[i+1].GetName(), err)
		}
	}
	return pipeline, nil
}

func applyEncoderSettings(elem *gst.Element, encoderName string, fps, bitrate, x264SpeedPreset, x264Tune, x264BFrames int, x264Trellis bool) error {
	if elem == nil {
		return fmt.Errorf("encoder element is nil")
	}
	switch strings.TrimSpace(encoderName) {
	case "", "x264enc":
		elem.Set("bitrate", bitrate)
		elem.Set("bframes", x264BFrames)
		elem.Set("tune", x264Tune)
		elem.Set("speed-preset", x264SpeedPreset)
		elem.Set("trellis", x264Trellis)
		return nil
	case "openh264enc":
		elem.Set("bitrate", bitrate*1000)
		elem.Set("rate-control", 1)
		elem.Set("usage-type", 1)
		elem.Set("complexity", 0)
		elem.Set("gop-size", fps)
		return nil
	case "vaapih264enc":
		elem.Set("bitrate", bitrate)
		elem.Set("rate-control", 2)
		elem.Set("keyframe-period", fps)
		elem.Set("max-bframes", 0)
		elem.Set("quality-level", 7)
		return nil
	default:
		return fmt.Errorf("unsupported encoder %q", encoderName)
	}
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

func slug(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
