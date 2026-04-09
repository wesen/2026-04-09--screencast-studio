package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apppkg "github.com/wesen/2026-04-09--screencast-studio/pkg/app"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/discovery"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/recording"
)

type fakeApplication struct {
	discoverySnapshot *discovery.Snapshot
	normalizeConfig   *dsl.EffectiveConfig
	compilePlan       *dsl.CompiledPlan
	recordDelay       time.Duration
	recordStarted     chan struct{}
}

func (f *fakeApplication) DiscoverySnapshot(ctx context.Context) (*discovery.Snapshot, error) {
	return f.discoverySnapshot, nil
}

func (f *fakeApplication) NormalizeDSL(ctx context.Context, body []byte) (*dsl.EffectiveConfig, error) {
	return f.normalizeConfig, nil
}

func (f *fakeApplication) CompileDSL(ctx context.Context, body []byte) (*dsl.CompiledPlan, error) {
	return f.compilePlan, nil
}

func (f *fakeApplication) RecordPlan(ctx context.Context, plan *dsl.CompiledPlan, options apppkg.RecordOptions) (*apppkg.RecordSummary, error) {
	if f.recordStarted != nil {
		select {
		case f.recordStarted <- struct{}{}:
		default:
		}
	}
	if options.EventSink != nil {
		options.EventSink(recording.RunEvent{
			Type:      recording.RunEventStateChanged,
			Timestamp: time.Now(),
			State:     recording.StateRunning,
			Reason:    "fake recording running",
		})
		options.EventSink(recording.RunEvent{
			Type:         recording.RunEventProcessLog,
			Timestamp:    time.Now(),
			ProcessLabel: "display-1",
			Stream:       "stderr",
			Message:      "fake ffmpeg log line",
		})
	}
	select {
	case <-ctx.Done():
	case <-time.After(f.recordDelay):
	}
	finishedAt := time.Now()
	if options.EventSink != nil {
		options.EventSink(recording.RunEvent{
			Type:      recording.RunEventStateChanged,
			Timestamp: finishedAt,
			State:     recording.StateFinished,
			Reason:    "fake recording finished",
		})
	}
	return &apppkg.RecordSummary{
		SessionID:  plan.SessionID,
		State:      string(recording.StateFinished),
		Reason:     "fake recording finished",
		Outputs:    append([]dsl.PlannedOutput(nil), plan.Outputs...),
		Warnings:   append([]string(nil), plan.Warnings...),
		StartedAt:  finishedAt.Add(-1 * time.Second),
		FinishedAt: finishedAt,
	}, nil
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	server := NewServer(&fakeApplication{}, Config{
		Addr:         ":0",
		PreviewLimit: 7,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload apiHealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Service != "screencast-studio" {
		t.Fatalf("service = %q, want screencast-studio", payload.Service)
	}
	if payload.PreviewLimit != 7 {
		t.Fatalf("preview_limit = %d, want 7", payload.PreviewLimit)
	}
}

func TestDiscoveryEndpoint(t *testing.T) {
	t.Parallel()

	server := NewServer(&fakeApplication{
		discoverySnapshot: &discovery.Snapshot{
			Displays: []discovery.Display{{ID: "display-1", Name: "Primary"}},
			Windows:  []discovery.Window{{ID: "0x01", Title: "Editor"}},
			Cameras:  []discovery.Camera{{ID: "camera-1", Device: "/dev/video0"}},
			Audio:    []discovery.AudioInput{{ID: "mic-1", Name: "Mic"}},
		},
	}, Config{})

	req := httptest.NewRequest(http.MethodGet, "/api/discovery", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload apiDiscoveryResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(payload.Displays) != 1 || payload.Displays[0].ID != "display-1" {
		t.Fatalf("unexpected displays payload: %+v", payload.Displays)
	}
	if len(payload.Windows) != 1 || payload.Windows[0].ID != "0x01" {
		t.Fatalf("unexpected windows payload: %+v", payload.Windows)
	}
}

func TestNormalizeAndCompileEndpoints(t *testing.T) {
	t.Parallel()

	server := NewServer(&fakeApplication{
		normalizeConfig: &dsl.EffectiveConfig{
			Schema:               dsl.SchemaVersion,
			SessionID:            "session-1",
			DestinationTemplates: map[string]string{"default": "/tmp/{session_id}.mkv"},
			Warnings:             []string{"warning-1"},
		},
		compilePlan: &dsl.CompiledPlan{
			SessionID: "session-1",
			Outputs:   []dsl.PlannedOutput{{Kind: "video", Name: "Display", Path: "/tmp/display.mkv"}},
			Warnings:  []string{"warning-1"},
		},
	}, Config{})

	body := []byte(`{"dsl":"schema: recorder.config/v1"}`)

	normalizeReq := httptest.NewRequest(http.MethodPost, "/api/setup/normalize", bytes.NewReader(body))
	normalizeRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(normalizeRec, normalizeReq)
	if normalizeRec.Code != http.StatusOK {
		t.Fatalf("normalize status = %d, want %d", normalizeRec.Code, http.StatusOK)
	}

	compileReq := httptest.NewRequest(http.MethodPost, "/api/setup/compile", bytes.NewReader(body))
	compileRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(compileRec, compileReq)
	if compileRec.Code != http.StatusOK {
		t.Fatalf("compile status = %d, want %d", compileRec.Code, http.StatusOK)
	}

	var compilePayload apiCompileResponse
	if err := json.Unmarshal(compileRec.Body.Bytes(), &compilePayload); err != nil {
		t.Fatalf("unmarshal compile response: %v", err)
	}
	if compilePayload.SessionID != "session-1" {
		t.Fatalf("session_id = %q, want session-1", compilePayload.SessionID)
	}
}

func TestRecordingLifecycleEndpoints(t *testing.T) {
	t.Parallel()

	fakeApp := &fakeApplication{
		compilePlan: &dsl.CompiledPlan{
			SessionID: "session-2",
			Outputs:   []dsl.PlannedOutput{{Kind: "video", Name: "Display", Path: "/tmp/display.mkv"}},
			Warnings:  []string{"warning-1"},
		},
		recordDelay:   10 * time.Second,
		recordStarted: make(chan struct{}, 1),
	}
	server := NewServer(fakeApp, Config{})

	startReq := httptest.NewRequest(http.MethodPost, "/api/recordings/start", bytes.NewBufferString(`{"dsl":"test"}`))
	startRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusOK {
		t.Fatalf("start status = %d, want %d", startRec.Code, http.StatusOK)
	}

	select {
	case <-fakeApp.recordStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for fake recording to start")
	}

	currentReq := httptest.NewRequest(http.MethodGet, "/api/recordings/current", nil)
	currentRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(currentRec, currentReq)
	if currentRec.Code != http.StatusOK {
		t.Fatalf("current status = %d, want %d", currentRec.Code, http.StatusOK)
	}

	var currentPayload apiSessionEnvelope
	if err := json.Unmarshal(currentRec.Body.Bytes(), &currentPayload); err != nil {
		t.Fatalf("unmarshal current response: %v", err)
	}
	if !currentPayload.Session.Active {
		t.Fatalf("expected active session, got %+v", currentPayload.Session)
	}
	if currentPayload.Session.SessionID != "session-2" {
		t.Fatalf("session_id = %q, want session-2", currentPayload.Session.SessionID)
	}

	stopReq := httptest.NewRequest(http.MethodPost, "/api/recordings/stop", nil)
	stopRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(stopRec, stopReq)
	if stopRec.Code != http.StatusOK {
		t.Fatalf("stop status = %d, want %d", stopRec.Code, http.StatusOK)
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		currentReq = httptest.NewRequest(http.MethodGet, "/api/session", nil)
		currentRec = httptest.NewRecorder()
		server.Handler().ServeHTTP(currentRec, currentReq)
		if currentRec.Code != http.StatusOK {
			t.Fatalf("session status = %d, want %d", currentRec.Code, http.StatusOK)
		}
		if err := json.Unmarshal(currentRec.Body.Bytes(), &currentPayload); err != nil {
			t.Fatalf("unmarshal session response: %v", err)
		}
		if !currentPayload.Session.Active {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("session never became inactive: %+v", currentPayload.Session)
		}
		time.Sleep(10 * time.Millisecond)
	}

	if len(currentPayload.Session.Logs) == 0 {
		t.Fatalf("expected session logs, got %+v", currentPayload.Session)
	}
}

func TestWebsocketPlaceholder(t *testing.T) {
	t.Parallel()

	server := NewServer(&fakeApplication{}, Config{})

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotImplemented)
	}
}
