package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
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

type fakePreviewRunner struct {
	started chan struct{}
	runs    atomic.Int32
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

func (f *fakePreviewRunner) Run(ctx context.Context, source dsl.EffectiveVideoSource, onFrame func([]byte), onLog func(string, string)) error {
	f.runs.Add(1)
	if f.started != nil {
		select {
		case f.started <- struct{}{}:
		default:
		}
	}
	onLog("stderr", "fake preview started")
	onFrame([]byte{0xff, 0xd8, 0xff, 0xd9})
	<-ctx.Done()
	return nil
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

func TestPreviewLifecycleEndpoints(t *testing.T) {
	t.Parallel()

	fakeApp := &fakeApplication{
		normalizeConfig: &dsl.EffectiveConfig{
			Schema:               dsl.SchemaVersion,
			SessionID:            "session-preview",
			DestinationTemplates: map[string]string{"default": "/tmp/out.mkv"},
			VideoSources: []dsl.EffectiveVideoSource{
				{
					ID:      "display-1",
					Name:    "Display 1",
					Type:    "display",
					Enabled: true,
					Target: dsl.VideoTarget{
						Display: ":0.0",
					},
					Capture: dsl.VideoCaptureSettings{FPS: 5},
					Output:  dsl.VideoOutputSettings{Container: "mkv", VideoCodec: "h264", Quality: 75},
					DestinationTemplate: "default",
				},
			},
		},
	}
	runner := &fakePreviewRunner{started: make(chan struct{}, 1)}
	server := NewServer(fakeApp, Config{PreviewLimit: 2})
	server.previews = NewPreviewManager(fakeApp, server.events.Publish, 2, runner)

	ensureReq := httptest.NewRequest(http.MethodPost, "/api/previews/ensure", bytes.NewBufferString(`{"dsl":"test","source_id":"display-1"}`))
	ensureRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(ensureRec, ensureReq)
	if ensureRec.Code != http.StatusOK {
		t.Fatalf("ensure status = %d, want %d", ensureRec.Code, http.StatusOK)
	}

	select {
	case <-runner.started:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for preview to start")
	}

	var ensurePayload apiPreviewEnvelope
	if err := json.Unmarshal(ensureRec.Body.Bytes(), &ensurePayload); err != nil {
		t.Fatalf("unmarshal ensure response: %v", err)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/previews", nil)
	listRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listRec.Code, http.StatusOK)
	}

	var listPayload apiPreviewListResponse
	if err := json.Unmarshal(listRec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("unmarshal preview list: %v", err)
	}
	if len(listPayload.Previews) != 1 {
		t.Fatalf("expected one preview, got %+v", listPayload.Previews)
	}

	releaseReq := httptest.NewRequest(http.MethodPost, "/api/previews/release", bytes.NewBufferString(`{"preview_id":"`+ensurePayload.Preview.ID+`"}`))
	releaseRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(releaseRec, releaseReq)
	if releaseRec.Code != http.StatusOK {
		t.Fatalf("release status = %d, want %d", releaseRec.Code, http.StatusOK)
	}
}

func TestPreviewMJPEGStream(t *testing.T) {
	t.Parallel()

	fakeApp := &fakeApplication{
		normalizeConfig: &dsl.EffectiveConfig{
			Schema:               dsl.SchemaVersion,
			SessionID:            "session-preview",
			DestinationTemplates: map[string]string{"default": "/tmp/out.mkv"},
			VideoSources: []dsl.EffectiveVideoSource{
				{
					ID:      "display-1",
					Name:    "Display 1",
					Type:    "display",
					Enabled: true,
					Target:  dsl.VideoTarget{Display: ":0.0"},
					Capture: dsl.VideoCaptureSettings{FPS: 5},
					Output:  dsl.VideoOutputSettings{Container: "mkv", VideoCodec: "h264", Quality: 75},
					DestinationTemplate: "default",
				},
			},
		},
	}
	runner := &fakePreviewRunner{started: make(chan struct{}, 1)}
	server := NewServer(fakeApp, Config{PreviewLimit: 2})
	server.previews = NewPreviewManager(fakeApp, server.events.Publish, 2, runner)

	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	ensureReqBody := bytes.NewBufferString(`{"dsl":"test","source_id":"display-1"}`)
	ensureResp, err := http.Post(ts.URL+"/api/previews/ensure", "application/json", ensureReqBody)
	if err != nil {
		t.Fatalf("ensure request: %v", err)
	}
	defer ensureResp.Body.Close()

	var ensurePayload apiPreviewEnvelope
	if err := json.NewDecoder(ensureResp.Body).Decode(&ensurePayload); err != nil {
		t.Fatalf("decode ensure response: %v", err)
	}

	select {
	case <-runner.started:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for preview runner")
	}

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/previews/"+ensurePayload.Preview.ID+"/mjpeg", nil)
	if err != nil {
		t.Fatalf("create mjpeg request: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("mjpeg request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "--frame") {
		t.Fatalf("expected mjpeg boundary in body, got %q", string(body))
	}
}

func TestWebsocketEndpoint(t *testing.T) {
	t.Parallel()

	server := NewServer(&fakeApplication{}, Config{})
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	var sessionEvent ServerEvent
	if err := conn.ReadJSON(&sessionEvent); err != nil {
		t.Fatalf("read session event: %v", err)
	}
	if sessionEvent.Type != "session.state" {
		t.Fatalf("first websocket event = %q, want session.state", sessionEvent.Type)
	}

	var previewEvent ServerEvent
	if err := conn.ReadJSON(&previewEvent); err != nil {
		t.Fatalf("read preview event: %v", err)
	}
	if previewEvent.Type != "preview.list" {
		t.Fatalf("second websocket event = %q, want preview.list", previewEvent.Type)
	}
}
