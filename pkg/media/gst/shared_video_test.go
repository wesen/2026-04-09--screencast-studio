package gst

import (
	"testing"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
)

func TestPreviewProfileForCameraPreservesHigherQuality(t *testing.T) {
	source := dsl.EffectiveVideoSource{
		Type: "camera",
		Capture: dsl.VideoCaptureSettings{
			FPS:  30,
			Size: "1280x720",
		},
	}

	got := previewProfileForSource(source)
	if got.MaxWidth != 1280 {
		t.Fatalf("expected camera preview max width 1280, got %d", got.MaxWidth)
	}
	if got.FPS != 10 {
		t.Fatalf("expected camera preview fps clamped to 10, got %d", got.FPS)
	}
	if got.JPEGQuality != 85 {
		t.Fatalf("expected camera preview jpeg quality 85, got %d", got.JPEGQuality)
	}

	width, height := previewTargetDimensions(source, got.MaxWidth)
	if width != 1280 || height != 720 {
		t.Fatalf("expected camera preview target dimensions 1280x720, got %dx%d", width, height)
	}
}

func TestPreviewProfileUsesSmallerWindowWidthWithoutUpscaling(t *testing.T) {
	source := dsl.EffectiveVideoSource{
		Type: "window",
		Target: dsl.VideoTarget{
			Rect: &dsl.Rect{W: 800, H: 600},
		},
		Capture: dsl.VideoCaptureSettings{FPS: 24},
	}

	got := previewProfileForSource(source)
	if got.MaxWidth != 1280 {
		t.Fatalf("expected window preview max width 1280, got %d", got.MaxWidth)
	}
	if got.FPS != 10 {
		t.Fatalf("expected window preview fps clamped to 10, got %d", got.FPS)
	}
	if got.JPEGQuality != 80 {
		t.Fatalf("expected window preview jpeg quality 80, got %d", got.JPEGQuality)
	}

	width, height := previewTargetDimensions(source, got.MaxWidth)
	if width != 800 || height != 600 {
		t.Fatalf("expected smaller window preview dimensions 800x600, got %dx%d", width, height)
	}
}

func TestPreviewFPSBounds(t *testing.T) {
	if got := previewFPS(0); got != 10 {
		t.Fatalf("expected default preview fps 10, got %d", got)
	}
	if got := previewFPS(5); got != 5 {
		t.Fatalf("expected preview fps 5, got %d", got)
	}
	if got := previewFPS(30); got != 10 {
		t.Fatalf("expected preview fps clamp 10, got %d", got)
	}
	if got := previewFPSForProfile(30, 4); got != 4 {
		t.Fatalf("expected recording preview fps clamp 4, got %d", got)
	}
}

func TestPreviewProfileWhileRecordingUsesConstrainedProfile(t *testing.T) {
	source := dsl.EffectiveVideoSource{
		Type:    "region",
		Target:  dsl.VideoTarget{Rect: &dsl.Rect{W: 2880, H: 960}},
		Capture: dsl.VideoCaptureSettings{FPS: 24},
	}

	got := previewProfileForSourceWhileRecording(source)
	if got.MaxWidth != 640 {
		t.Fatalf("expected recording preview max width 640, got %d", got.MaxWidth)
	}
	if got.FPS != 4 {
		t.Fatalf("expected recording preview fps 4, got %d", got.FPS)
	}
	if got.JPEGQuality != 50 {
		t.Fatalf("expected recording preview jpeg quality 50, got %d", got.JPEGQuality)
	}
}

func TestPreviewStagesForLayout(t *testing.T) {
	gotScaleFirst := previewStagesForLayout(sharedPreviewLayoutScaleFirst)
	wantScaleFirst := []sharedPreviewStage{
		sharedPreviewStageQueue,
		sharedPreviewStageScale,
		sharedPreviewStageScaleCaps,
		sharedPreviewStageRate,
		sharedPreviewStageRateCaps,
		sharedPreviewStageJPEG,
		sharedPreviewStageSink,
	}
	if len(gotScaleFirst) != len(wantScaleFirst) {
		t.Fatalf("expected %d scale-first stages, got %d", len(wantScaleFirst), len(gotScaleFirst))
	}
	for i := range wantScaleFirst {
		if gotScaleFirst[i] != wantScaleFirst[i] {
			t.Fatalf("scale-first stage %d: expected %s, got %s", i, wantScaleFirst[i], gotScaleFirst[i])
		}
	}

	gotRateFirst := previewStagesForLayout(sharedPreviewLayoutRateFirst)
	wantRateFirst := []sharedPreviewStage{
		sharedPreviewStageQueue,
		sharedPreviewStageRate,
		sharedPreviewStageRateCaps,
		sharedPreviewStageScale,
		sharedPreviewStageScaleCaps,
		sharedPreviewStageJPEG,
		sharedPreviewStageSink,
	}
	if len(gotRateFirst) != len(wantRateFirst) {
		t.Fatalf("expected %d rate-first stages, got %d", len(wantRateFirst), len(gotRateFirst))
	}
	for i := range wantRateFirst {
		if gotRateFirst[i] != wantRateFirst[i] {
			t.Fatalf("rate-first stage %d: expected %s, got %s", i, wantRateFirst[i], gotRateFirst[i])
		}
	}
}

func TestPreviewTargetDimensionsPreserveAspectRatioWhenScalingDown(t *testing.T) {
	source := dsl.EffectiveVideoSource{
		Type: "region",
		Target: dsl.VideoTarget{
			Rect: &dsl.Rect{W: 2880, H: 960},
		},
	}

	width, height := previewTargetDimensions(source, 1280)
	if width != 1280 || height != 427 {
		t.Fatalf("expected scaled region preview dimensions 1280x427, got %dx%d", width, height)
	}
	if got := previewScaleCaps(width, height); got != "video/x-raw,width=1280,height=427,pixel-aspect-ratio=1/1" {
		t.Fatalf("unexpected preview caps: %q", got)
	}
}
