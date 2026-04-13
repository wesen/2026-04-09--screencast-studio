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
	codec := envOr("CODEC", "wav")
	outPath := envOr("OUT_PATH", filepath.Join(os.TempDir(), defaultOutputName(codec)))
	sources := buildSourcesFromEnv()
	output := dsl.AudioOutputSettings{Codec: codec, SampleRateHz: intEnvOr("SAMPLE_RATE_HZ", 48000), Channels: intEnvOr("CHANNELS", 2)}
	plan := &dsl.CompiledPlan{
		SessionID: "gst-audio-recording-smoke",
		AudioJobs: []dsl.AudioMixJob{{Name: "audio-mix", Sources: sources, Output: output, OutputPath: outPath}},
		Outputs:   []dsl.PlannedOutput{{Kind: "audio", Name: "audio-mix", Path: outPath}},
	}

	fmt.Println("=== GStreamer Audio Recording Runtime Smoke Test ===")
	fmt.Printf("codec:      %s\n", codec)
	fmt.Printf("output:     %s\n", outPath)
	fmt.Printf("sources:    %d\n", len(sources))
	for i, src := range sources {
		fmt.Printf("  [%d] device=%s gain=%.2f\n", i, src.Device, src.Settings.Gain)
	}
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
	fmt.Println()
	fmt.Printf("result state: %s\n", result.State)
	fmt.Printf("result reason: %s\n", result.Reason)
	info, err := os.Stat(outPath)
	if err != nil {
		fatal("stat output: %v", err)
	}
	fmt.Printf("output size:   %d bytes\n", info.Size())
	if info.Size() == 0 {
		fatal("output file is empty")
	}
}

func buildSourcesFromEnv() []dsl.EffectiveAudioSource {
	devicesValue := envOr("DEVICES", envOr("DEVICE", "default"))
	deviceParts := strings.Split(devicesValue, ",")
	gainsValue := envOr("GAINS", "")
	gainParts := []string{}
	if gainsValue != "" {
		gainParts = strings.Split(gainsValue, ",")
	}
	sources := make([]dsl.EffectiveAudioSource, 0, len(deviceParts))
	for i, part := range deviceParts {
		device := strings.TrimSpace(part)
		if device == "" {
			continue
		}
		gain := 1.0
		if i < len(gainParts) {
			parsed, err := strconv.ParseFloat(strings.TrimSpace(gainParts[i]), 64)
			if err != nil {
				fatal("parse gain %q: %v", gainParts[i], err)
			}
			gain = parsed
		}
		sources = append(sources, dsl.EffectiveAudioSource{
			ID:      fmt.Sprintf("audio-%d", i+1),
			Name:    fmt.Sprintf("Audio %d", i+1),
			Enabled: true,
			Device:  device,
			Settings: dsl.AudioSourceSettings{
				Gain: gain,
			},
		})
	}
	if len(sources) == 0 {
		fatal("no audio sources configured")
	}
	return sources
}

func defaultOutputName(codec string) string {
	switch strings.ToLower(strings.TrimSpace(codec)) {
	case "opus":
		return "gst-audio-smoke.ogg"
	default:
		return "gst-audio-smoke.wav"
	}
}

func envOr(key, def string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return def
}

func intEnvOr(key string, def int) int {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			fatal("parse %s: %v", key, err)
		}
		return parsed
	}
	return def
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
