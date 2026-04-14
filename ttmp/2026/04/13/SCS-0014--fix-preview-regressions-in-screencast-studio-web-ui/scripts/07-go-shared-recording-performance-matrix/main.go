package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/media"
	gstreamer "github.com/wesen/2026-04-09--screencast-studio/pkg/media/gst"
)

func main() {
	scenario := envOr("SCENARIO", "preview-plus-recorder")
	duration := durationFromEnv("DURATION", 6*time.Second)
	postPreview := durationFromEnv("POST_PREVIEW", 2*time.Second)
	outputPath := envOr("OUTPUT_PATH", filepath.Join(os.TempDir(), "scs-go-shared-recording-bench.mp4"))
	source := buildSourceFromEnv()

	fmt.Printf("scenario=%s\n", scenario)
	fmt.Printf("duration=%s\n", duration)
	fmt.Printf("output=%s\n", outputPath)
	fmt.Printf("source_type=%s region=%+v fps=%d\n", source.Type, source.Target.Rect, source.Capture.FPS)

	previewRuntime := gstreamer.NewPreviewRuntime()
	var previewFrames atomic.Uint64
	var previewSession media.PreviewSession
	var err error
	stopPreview := func() {
		if previewSession != nil {
			_ = previewSession.Stop(context.Background())
		}
	}
	defer stopPreview()

	startPreview := func(label string) {
		ctx, _ := context.WithCancel(context.Background())
		previewSession, err = previewRuntime.StartPreview(ctx, source, media.PreviewOptions{
			OnFrame: func([]byte) { previewFrames.Add(1) },
			OnLog: func(stream, message string) {
				fmt.Printf("preview-log[%s]: %s\n", stream, message)
			},
		})
		must(err, "start preview")
		waitForFrames(label, &previewFrames, 2, 6*time.Second)
	}

	switch scenario {
	case "preview-only":
		startPreview("preview-only warmup")
		time.Sleep(duration)
		fmt.Printf("preview_frames=%d\n", previewFrames.Load())
	case "recorder-only":
		recorder := startRecorder(source, outputPath)
		time.Sleep(duration)
		must(recorder.Stop(context.Background()), "stop recorder")
		fmt.Printf("recorder_stopped=true\n")
	case "preview-plus-recorder":
		startPreview("preview warmup")
		before := previewFrames.Load()
		recorder := startRecorder(source, outputPath)
		time.Sleep(duration)
		must(recorder.Stop(context.Background()), "stop recorder")
		fmt.Printf("preview_frames_during_record=%d\n", previewFrames.Load()-before)
		time.Sleep(postPreview)
		fmt.Printf("preview_frames_total=%d\n", previewFrames.Load())
	default:
		fatal("unsupported SCENARIO=%s", scenario)
	}
}

func startRecorder(source dsl.EffectiveVideoSource, outputPath string) *gstreamer.ExperimentalSharedVideoRecorder {
	_ = os.Remove(outputPath)
	recorder, err := gstreamer.StartExperimentalSharedVideoRecorder(context.Background(), source, gstreamer.ExperimentalSharedVideoRecorderOptions{
		OutputPath: outputPath,
		Container:  "mp4",
		FPS:        source.Capture.FPS,
		OnLog: func(stream, message string) {
			fmt.Printf("record-log[%s]: %s\n", stream, message)
		},
	})
	must(err, "start recorder")
	return recorder
}

func buildSourceFromEnv() dsl.EffectiveVideoSource {
	return dsl.EffectiveVideoSource{
		ID:      "perf-bench-source",
		Name:    "Performance Bench Source",
		Type:    envOr("SOURCE_TYPE", "region"),
		Enabled: true,
		Target: dsl.VideoTarget{
			Display: envOr("DISPLAY", ":0"),
			Rect:    parseRect(envOr("REGION", "0,540,1920,540")),
		},
		Capture: dsl.VideoCaptureSettings{FPS: intFromEnv("FPS", 24), Cursor: boolPtr(true)},
		Output:  dsl.VideoOutputSettings{Container: "mp4", VideoCodec: "h264", Quality: 75},
	}
}

func parseRect(value string) *dsl.Rect {
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
	return &dsl.Rect{X: vals[0], Y: vals[1], W: vals[2], H: vals[3]}
}

func waitForFrames(label string, counter *atomic.Uint64, min uint64, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if counter.Load() >= min {
			fmt.Printf("%s reached %d frames\n", label, counter.Load())
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	fatal("timed out waiting for %s to reach %d frames (got %d)", label, min, counter.Load())
}

func envOr(key, def string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return def
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

func boolPtr(v bool) *bool { return &v }

func must(err error, label string) {
	if err != nil {
		fatal("%s: %v", label, err)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
