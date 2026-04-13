package web

import (
	"context"
	"testing"
	"time"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
)

func TestRecordingManagerShutdownWaitsForActiveSession(t *testing.T) {
	t.Parallel()

	fakeApp := &fakeApplication{
		compilePlan: &dsl.CompiledPlan{
			SessionID: "session-shutdown",
			Outputs: []dsl.PlannedOutput{
				{Kind: "video", Name: "Display", Path: "/tmp/session-shutdown/display.mkv"},
			},
		},
		recordDelay:   10 * time.Second,
		recordStarted: make(chan struct{}, 1),
	}

	manager := NewRecordingManager(context.Background(), fakeApp, nil)
	if _, err := manager.Start([]byte("test"), 0, 0); err != nil {
		t.Fatalf("start recording session: %v", err)
	}

	select {
	case <-fakeApp.recordStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for fake recording to start")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := manager.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown recording manager: %v", err)
	}

	snapshot := manager.Current()
	if snapshot.Active {
		t.Fatalf("expected inactive session after shutdown, got %+v", snapshot)
	}
	if snapshot.SessionID != "session-shutdown" {
		t.Fatalf("session_id = %q, want session-shutdown", snapshot.SessionID)
	}
}

func TestRecordingManagerShutdownTimesOutWhenSessionDoesNotFinish(t *testing.T) {
	t.Parallel()

	manager := NewRecordingManager(context.Background(), &fakeApplication{}, nil)
	done := make(chan struct{})
	manager.current = &managedRecording{
		cancel: func() {},
		done:   done,
		state: recordingSessionState{
			Active:    true,
			SessionID: "session-timeout",
			State:     "running",
		},
	}
	defer close(done)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if err := manager.Shutdown(shutdownCtx); err == nil {
		t.Fatal("expected shutdown timeout error, got nil")
	}
}

func TestPreviewManagerShutdownCancelsActivePreviews(t *testing.T) {
	t.Parallel()

	fakeApp := &fakeApplication{
		normalizeConfig: &dsl.EffectiveConfig{
			Schema:               dsl.SchemaVersion,
			SessionID:            "session-preview-shutdown",
			DestinationTemplates: map[string]string{"default": "/tmp/out.mkv"},
			VideoSources: []dsl.EffectiveVideoSource{
				{
					ID:                  "display-1",
					Name:                "Display 1",
					Type:                "display",
					Enabled:             true,
					Target:              dsl.VideoTarget{Display: ":0.0"},
					Capture:             dsl.VideoCaptureSettings{FPS: 5},
					Output:              dsl.VideoOutputSettings{Container: "mkv", VideoCodec: "h264", Quality: 75},
					DestinationTemplate: "default",
				},
			},
		},
	}
	runner := &fakePreviewRunner{started: make(chan struct{}, 1)}
	manager := NewPreviewManager(context.Background(), fakeApp, nil, 2, runner)

	if _, err := manager.Ensure(context.Background(), []byte("test"), "display-1"); err != nil {
		t.Fatalf("ensure preview: %v", err)
	}

	select {
	case <-runner.started:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for fake preview runner to start")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := manager.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown preview manager: %v", err)
	}

	if previews := manager.List(); len(previews) != 0 {
		t.Fatalf("expected no previews after shutdown, got %+v", previews)
	}
}

func TestPreviewManagerShutdownTimesOutWhenPreviewDoesNotFinish(t *testing.T) {
	t.Parallel()

	manager := NewPreviewManager(context.Background(), &fakeApplication{}, nil, 2, nil)
	done := make(chan struct{})
	preview := &managedPreview{
		id:     "preview-timeout",
		source: dsl.EffectiveVideoSource{ID: "display-1", Name: "Display 1", Type: "display"},
		cancel: func() {},
		done:   done,
		state:  "running",
	}
	manager.byID[preview.id] = preview
	manager.bySignature["sig"] = preview
	defer close(done)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if err := manager.Shutdown(shutdownCtx); err == nil {
		t.Fatal("expected preview shutdown timeout error, got nil")
	}
}

func TestPreviewManagerSuspendAndRestore(t *testing.T) {
	t.Parallel()

	fakeApp := &fakeApplication{
		normalizeConfig: &dsl.EffectiveConfig{
			Schema:               dsl.SchemaVersion,
			SessionID:            "session-preview-suspend-restore",
			DestinationTemplates: map[string]string{"default": "/tmp/out.mkv"},
			VideoSources: []dsl.EffectiveVideoSource{
				{
					ID:                  "display-1",
					Name:                "Display 1",
					Type:                "display",
					Enabled:             true,
					Target:              dsl.VideoTarget{Display: ":0.0"},
					Capture:             dsl.VideoCaptureSettings{FPS: 5},
					Output:              dsl.VideoOutputSettings{Container: "mkv", VideoCodec: "h264", Quality: 75},
					DestinationTemplate: "default",
				},
			},
		},
	}
	runner := &fakePreviewRunner{started: make(chan struct{}, 4)}
	manager := NewPreviewManager(context.Background(), fakeApp, nil, 2, runner)

	if _, err := manager.Ensure(context.Background(), []byte("test"), "display-1"); err != nil {
		t.Fatalf("ensure preview: %v", err)
	}

	select {
	case <-runner.started:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for preview runner to start")
	}

	plan, err := manager.SuspendAll(context.Background(), "recording starting")
	if err != nil {
		t.Fatalf("suspend previews: %v", err)
	}
	if len(plan.SourceIDs) != 1 || plan.SourceIDs[0] != "display-1" {
		t.Fatalf("unexpected suspend plan: %+v", plan)
	}
	if previews := manager.List(); len(previews) != 0 {
		t.Fatalf("expected no previews after suspend, got %+v", previews)
	}

	if err := manager.RestoreSuspended(context.Background(), []byte("test"), plan); err != nil {
		t.Fatalf("restore previews: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		if runner.runs.Load() >= 2 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected restored preview run, got %d", runner.runs.Load())
		}
		time.Sleep(10 * time.Millisecond)
	}
}
