package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/media"
	gstreamer "github.com/wesen/2026-04-09--screencast-studio/pkg/media/gst"
)

func main() {
	runtime := gstreamer.NewRecordingRuntime()
	source := buildSourceFromEnv()
	outPath := envOr("OUT_PATH", filepath.Join(os.TempDir(), fmt.Sprintf("gst-recording-%s.mp4", source.Type)))
	plan := &dsl.CompiledPlan{
		SessionID: "gst-recording-smoke",
		VideoJobs: []dsl.VideoJob{{Source: source, OutputPath: outPath}},
		Outputs:   []dsl.PlannedOutput{{Kind: "video", SourceID: source.ID, Name: source.Name, Path: outPath}},
	}

	fmt.Println("=== GStreamer Recording Runtime Smoke Test ===")
	fmt.Printf("source type: %s\n", source.Type)
	fmt.Printf("output path: %s\n", outPath)
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	session, err := runtime.StartRecording(ctx, plan, media.RecordingOptions{
		EventSink: func(event media.RecordingEvent) {
			fmt.Printf("event: type=%s state=%s reason=%q label=%q path=%q\n", event.Type, event.State, event.Reason, event.ProcessLabel, event.OutputPath)
		},
	})
	if err != nil {
		fatal("start recording: %v", err)
	}

	result, err := session.Wait()
	if err != nil {
		fatal("wait recording: %v", err)
	}
	if result == nil {
		fatal("recording result is nil")
	}

	fmt.Println()
	fmt.Printf("result state: %s\n", result.State)
	fmt.Printf("result reason: %s\n", result.Reason)
	fmt.Printf("started at:    %s\n", result.StartedAt.Format(time.RFC3339Nano))
	fmt.Printf("finished at:   %s\n", result.FinishedAt.Format(time.RFC3339Nano))

	info, err := os.Stat(outPath)
	if err != nil {
		fatal("stat output: %v", err)
	}
	fmt.Printf("output size:   %d bytes\n", info.Size())
	if info.Size() == 0 {
		fatal("output file is empty")
	}
}

func buildSourceFromEnv() dsl.EffectiveVideoSource {
	sourceType := envOr("SOURCE_TYPE", "display")
	container := envOr("CONTAINER", "mp4")
	source := dsl.EffectiveVideoSource{
		ID:      "recording-source-1",
		Name:    "Recording Source",
		Type:    sourceType,
		Enabled: true,
		Target: dsl.VideoTarget{
			Display:  envOr("DISPLAY_NAME", envOr("DISPLAY", ":0")),
			Device:   envOr("DEVICE", ""),
			WindowID: envOr("WINDOW_ID", ""),
		},
		Capture: dsl.VideoCaptureSettings{FPS: 10, Size: envOr("SIZE", "")},
		Output:  dsl.VideoOutputSettings{Container: container, VideoCodec: "h264", Quality: 75},
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
		fatal("REGION must be x,y,w,h")
	}
	vals := make([]int, 4)
	for i, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			fatal("parse REGION component: %v", err)
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

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
