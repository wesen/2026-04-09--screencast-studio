package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/media"
	gstreamer "github.com/wesen/2026-04-09--screencast-studio/pkg/media/gst"
)

func main() {
	runtime := gstreamer.NewPreviewRuntime()
	source := buildSourceFromEnv()

	fmt.Println("=== GStreamer Shared Preview Runtime Smoke Test ===")
	fmt.Printf("source type: %s\n", source.Type)
	fmt.Printf("display:     %s\n", source.Target.Display)
	if source.Target.Rect != nil {
		fmt.Printf("region:      %d,%d %dx%d\n", source.Target.Rect.X, source.Target.Rect.Y, source.Target.Rect.W, source.Target.Rect.H)
	}
	fmt.Println()

	var preview1Frames atomic.Uint64
	var preview2Frames atomic.Uint64
	var preview3Frames atomic.Uint64

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	session1, err := runtime.StartPreview(ctx1, source, media.PreviewOptions{
		OnFrame: func([]byte) { preview1Frames.Add(1) },
	})
	must(err, "start preview 1")
	defer func() { _ = session1.Stop(context.Background()) }()

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	session2, err := runtime.StartPreview(ctx2, source, media.PreviewOptions{
		OnFrame: func([]byte) { preview2Frames.Add(1) },
	})
	must(err, "start preview 2")
	defer func() { _ = session2.Stop(context.Background()) }()

	waitForFrames("preview 1", &preview1Frames, 1, 6*time.Second)
	waitForFrames("preview 2", &preview2Frames, 1, 6*time.Second)
	fmt.Printf("initial frames: preview1=%d preview2=%d\n", preview1Frames.Load(), preview2Frames.Load())

	beforeStopPreview2 := preview2Frames.Load()
	must(session1.Stop(context.Background()), "stop preview 1")
	fmt.Println("preview 1 stopped")
	waitForFrames("preview 2 after preview 1 stop", &preview2Frames, beforeStopPreview2+2, 6*time.Second)
	fmt.Printf("preview 2 continued: preview2=%d\n", preview2Frames.Load())

	must(session2.Stop(context.Background()), "stop preview 2")
	fmt.Println("preview 2 stopped")

	time.Sleep(500 * time.Millisecond)

	ctx3, cancel3 := context.WithCancel(context.Background())
	defer cancel3()
	session3, err := runtime.StartPreview(ctx3, source, media.PreviewOptions{
		OnFrame: func([]byte) { preview3Frames.Add(1) },
	})
	must(err, "start preview 3")
	defer func() { _ = session3.Stop(context.Background()) }()
	waitForFrames("preview 3 after recreate", &preview3Frames, 1, 6*time.Second)
	fmt.Printf("preview 3 recreated successfully: preview3=%d\n", preview3Frames.Load())

	must(session3.Stop(context.Background()), "stop preview 3")
	fmt.Println("shared preview runtime smoke test passed")
}

func buildSourceFromEnv() dsl.EffectiveVideoSource {
	display := envOr("DISPLAY", ":0")
	region := envOr("REGION", "0,0,640,480")
	rect := parseRect(region)
	return dsl.EffectiveVideoSource{
		ID:      "shared-preview-source",
		Name:    "Shared Preview Source",
		Type:    "region",
		Enabled: true,
		Target: dsl.VideoTarget{
			Display: display,
			Rect:    rect,
		},
		Capture: dsl.VideoCaptureSettings{FPS: 10, Cursor: boolPtr(true)},
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
