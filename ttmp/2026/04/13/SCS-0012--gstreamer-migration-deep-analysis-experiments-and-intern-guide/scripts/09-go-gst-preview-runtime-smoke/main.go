package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
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
	fmt.Printf("=== GStreamer Preview Runtime Smoke Test ===\n")
	fmt.Printf("source type: %s\n", source.Type)
	fmt.Printf("source id:   %s\n", source.ID)
	fmt.Printf("display:     %s\n", source.Target.Display)
	if source.Target.Device != "" {
		fmt.Printf("device:      %s\n", source.Target.Device)
	}
	if source.Target.WindowID != "" {
		fmt.Printf("window id:   %s\n", source.Target.WindowID)
	}
	if source.Target.Rect != nil {
		fmt.Printf("region:      %d,%d %dx%d\n", source.Target.Rect.X, source.Target.Rect.Y, source.Target.Rect.W, source.Target.Rect.H)
	}
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var frameCount atomic.Int64
	var lastFrameBytes atomic.Int64
	session, err := runtime.StartPreview(ctx, source, media.PreviewOptions{
		OnFrame: func(frame []byte) {
			next := frameCount.Add(1)
			lastFrameBytes.Store(int64(len(frame)))
			fmt.Printf("frame %03d: %d bytes\n", next, len(frame))
		},
		OnLog: func(stream, message string) {
			fmt.Printf("[%s] %s\n", stream, message)
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "start preview: %v\n", err)
		os.Exit(1)
	}

	if err := session.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "preview wait: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("frames captured: %d\n", frameCount.Load())
	fmt.Printf("last frame size: %d bytes\n", lastFrameBytes.Load())
}

func buildSourceFromEnv() dsl.EffectiveVideoSource {
	sourceType := envOr("SOURCE_TYPE", "display")
	source := dsl.EffectiveVideoSource{
		ID:      "smoke-preview",
		Name:    "Smoke Preview",
		Type:    sourceType,
		Enabled: true,
		Target: dsl.VideoTarget{
			Display:  envOr("DISPLAY_NAME", envOr("DISPLAY", ":0.0")),
			Device:   envOr("DEVICE", ""),
			WindowID: envOr("WINDOW_ID", ""),
		},
		Capture: dsl.VideoCaptureSettings{FPS: 30},
	}
	if rect := parseRect(envOr("REGION", "")); rect != nil {
		source.Target.Rect = rect
		if source.Type == "display" {
			source.Type = "region"
		}
	}
	if source.Type == "camera" && source.Target.Device == "" {
		source.Target.Device = "/dev/video0"
	}
	return source
}

func parseRect(value string) *dsl.Rect {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	if len(parts) != 4 {
		panic("REGION must be x,y,w,h")
	}
	vals := make([]int, 4)
	for i, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			panic(err)
		}
		vals[i] = n
	}
	return &dsl.Rect{X: vals[0], Y: vals[1], W: vals[2], H: vals[3]}
}

func envOr(key, def string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return def
}
