package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/media"
	gstreamer "github.com/wesen/2026-04-09--screencast-studio/pkg/media/gst"
)

func main() {
	source := buildSourceFromEnv()
	previewRuntime := gstreamer.NewPreviewRuntime()
	outPath := filepath.Join(os.TempDir(), "gst-shared-bridge-recorder-smoke.mp4")
	_ = os.Remove(outPath)

	var previewFrames atomic.Uint64
	previewCtx, previewCancel := context.WithCancel(context.Background())
	defer previewCancel()
	previewSession, err := previewRuntime.StartPreview(previewCtx, source, media.PreviewOptions{
		OnFrame: func([]byte) { previewFrames.Add(1) },
	})
	must(err, "start preview session")
	defer func() { _ = previewSession.Stop(context.Background()) }()

	waitForFrames("preview before recorder start", &previewFrames, 2, 6*time.Second)
	beforeRecording := previewFrames.Load()

	recorderCtx, recorderCancel := context.WithCancel(context.Background())
	defer recorderCancel()
	recorder, err := gstreamer.StartExperimentalSharedVideoRecorder(recorderCtx, source, gstreamer.ExperimentalSharedVideoRecorderOptions{
		OutputPath: outPath,
		Container:  "mp4",
		FPS:        10,
		OnLog: func(stream, message string) {
			fmt.Printf("bridge log [%s]: %s\n", stream, message)
		},
	})
	must(err, "start shared bridge recorder")

	fmt.Println("recorder started; letting it run for 3 seconds...")
	time.Sleep(3 * time.Second)
	waitForFrames("preview during recording", &previewFrames, beforeRecording+2, 6*time.Second)
	fmt.Printf("preview frames before stop: %d\n", previewFrames.Load())

	must(recorder.Stop(context.Background()), "stop shared bridge recorder")
	fmt.Println("recorder stopped")
	waitForFrames("preview after recorder stop", &previewFrames, previewFrames.Load()+2, 6*time.Second)
	fmt.Printf("preview frames after recorder stop: %d\n", previewFrames.Load())

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

func buildSourceFromEnv() dsl.EffectiveVideoSource {
	return dsl.EffectiveVideoSource{
		ID:      "shared-bridge-source",
		Name:    "Shared Bridge Source",
		Type:    "region",
		Enabled: true,
		Target: dsl.VideoTarget{
			Display: envOr("DISPLAY", ":0"),
			Rect:    parseRect(envOr("REGION", "0,0,640,480")),
		},
		Capture: dsl.VideoCaptureSettings{FPS: 10, Cursor: boolPtr(true)},
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

func boolPtr(v bool) *bool { return &v }

func envOr(key, def string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return def
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
