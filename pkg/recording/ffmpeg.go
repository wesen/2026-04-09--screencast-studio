package recording

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
)

func buildVideoRecordArgs(job dsl.VideoJob, maxDuration time.Duration) ([]string, error) {
	args := []string{"-hide_banner", "-loglevel", "error", "-y"}
	if maxDuration > 0 {
		args = append(args, "-t", formatDuration(maxDuration))
	}
	var err error
	args, err = appendVideoInputArgs(args, job.Source)
	if err != nil {
		return nil, err
	}
	filters := []string{}
	if job.Source.Type == "camera" && boolValue(job.Source.Capture.Mirror, false) {
		filters = append(filters, "hflip")
	}
	if len(filters) > 0 {
		args = append(args, "-vf", strings.Join(filters, ","))
	}
	args = append(args, encodeVideoArgs(job.Source.Output)...)
	args = append(args, job.OutputPath)
	return args, nil
}

func buildAudioMixArgs(job dsl.AudioMixJob, maxDuration time.Duration) ([]string, error) {
	if len(job.Sources) == 0 {
		return nil, errors.New("no audio sources")
	}
	args := []string{"-hide_banner", "-loglevel", "error", "-y"}
	for _, src := range job.Sources {
		if maxDuration > 0 {
			args = append(args, "-t", formatDuration(maxDuration))
		}
		args = append(args,
			"-f", "pulse",
			"-sample_rate", fmt.Sprintf("%d", job.Output.SampleRateHz),
			"-channels", fmt.Sprintf("%d", job.Output.Channels),
			"-i", src.Device,
		)
	}
	filterParts := make([]string, 0, len(job.Sources)+1)
	mixInputs := make([]string, 0, len(job.Sources))
	for i, src := range job.Sources {
		label := fmt.Sprintf("a%d", i)
		filterParts = append(filterParts, fmt.Sprintf("[%d:a]volume=%s[%s]", i, trimFloat(src.Settings.Gain), label))
		mixInputs = append(mixInputs, fmt.Sprintf("[%s]", label))
	}
	if len(job.Sources) == 1 {
		filterParts = append(filterParts, fmt.Sprintf("%sanull[aout]", mixInputs[0]))
	} else {
		filterParts = append(filterParts, fmt.Sprintf("%samix=inputs=%d:normalize=0[aout]", strings.Join(mixInputs, ""), len(job.Sources)))
	}
	args = append(args,
		"-filter_complex", strings.Join(filterParts, ";"),
		"-map", "[aout]",
		"-ar", fmt.Sprintf("%d", job.Output.SampleRateHz),
		"-ac", fmt.Sprintf("%d", job.Output.Channels),
	)
	switch strings.ToLower(strings.TrimSpace(job.Output.Codec)) {
	case "", "pcm_s16le", "wav":
		args = append(args, "-c:a", "pcm_s16le")
	case "aac":
		args = append(args, "-c:a", "aac", "-b:a", "192k")
	case "opus":
		args = append(args, "-c:a", "libopus", "-b:a", "160k")
	case "mp3":
		args = append(args, "-c:a", "libmp3lame", "-b:a", "192k")
	default:
		return nil, fmt.Errorf("unsupported audio codec %q", job.Output.Codec)
	}
	args = append(args, job.OutputPath)
	return args, nil
}

func appendVideoInputArgs(args []string, src dsl.EffectiveVideoSource) ([]string, error) {
	switch src.Type {
	case "display":
		if strings.TrimSpace(src.Target.Display) == "" {
			return nil, errors.New("display source missing target.display")
		}
		args = append(args,
			"-f", "x11grab",
			"-framerate", fmt.Sprintf("%d", src.Capture.FPS),
			"-draw_mouse", boolFlag(boolValue(src.Capture.Cursor, true)),
			"-i", src.Target.Display,
		)
		return args, nil
	case "region":
		if src.Target.Rect == nil {
			return nil, errors.New("missing region rect")
		}
		if strings.TrimSpace(src.Target.Display) == "" {
			return nil, errors.New("region source missing target.display")
		}
		args = append(args,
			"-f", "x11grab",
			"-framerate", fmt.Sprintf("%d", src.Capture.FPS),
			"-video_size", fmt.Sprintf("%dx%d", src.Target.Rect.W, src.Target.Rect.H),
			"-draw_mouse", boolFlag(boolValue(src.Capture.Cursor, true)),
			"-i", fmt.Sprintf("%s+%d,%d", src.Target.Display, src.Target.Rect.X, src.Target.Rect.Y),
		)
		return args, nil
	case "window":
		if strings.TrimSpace(src.Target.Display) == "" {
			return nil, errors.New("window source missing target.display")
		}
		if strings.TrimSpace(src.Target.WindowID) == "" {
			return nil, errors.New("window source missing target.window_id")
		}
		args = append(args,
			"-f", "x11grab",
			"-framerate", fmt.Sprintf("%d", src.Capture.FPS),
			"-draw_mouse", boolFlag(boolValue(src.Capture.Cursor, true)),
			"-window_id", src.Target.WindowID,
			"-i", src.Target.Display,
		)
		return args, nil
	case "camera":
		if strings.TrimSpace(src.Target.Device) == "" {
			return nil, errors.New("camera source missing target.device")
		}
		args = append(args, "-f", "v4l2")
		if src.Capture.Size != "" {
			args = append(args, "-video_size", src.Capture.Size)
		}
		args = append(args,
			"-framerate", fmt.Sprintf("%d", src.Capture.FPS),
			"-i", src.Target.Device,
		)
		return args, nil
	default:
		return nil, fmt.Errorf("unsupported video source type %q", src.Type)
	}
}

func encodeVideoArgs(out dsl.VideoOutputSettings) []string {
	switch strings.ToLower(strings.TrimSpace(out.VideoCodec)) {
	case "", "h264", "libx264":
		return []string{"-c:v", "libx264", "-preset", "veryfast", "-crf", fmt.Sprintf("%d", qualityToCRF(out.Quality)), "-pix_fmt", "yuv420p"}
	case "ffv1":
		return []string{"-c:v", "ffv1"}
	case "mpeg4":
		return []string{"-c:v", "mpeg4", "-q:v", fmt.Sprintf("%d", qualityToQScale(out.Quality))}
	default:
		return []string{"-c:v", "libx264", "-preset", "veryfast", "-crf", fmt.Sprintf("%d", qualityToCRF(out.Quality)), "-pix_fmt", "yuv420p"}
	}
}

func qualityToCRF(q int) int {
	if q < 1 {
		q = 75
	}
	if q > 100 {
		q = 100
	}
	return 35 - int(math.Round(float64(q-1)*18.0/99.0))
}

func qualityToQScale(q int) int {
	if q < 1 {
		q = 75
	}
	if q > 100 {
		q = 100
	}
	return 31 - int(math.Round(float64(q-1)*26.0/99.0))
}

func boolValue(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}

func boolFlag(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func trimFloat(v float64) string {
	if v == 0 {
		v = 1.0
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", v), "0"), ".")
}

func formatDuration(d time.Duration) string {
	seconds := d.Seconds()
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", seconds), "0"), ".")
}
