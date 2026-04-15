package web

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gorilla/websocket"
	studiov1 "github.com/wesen/2026-04-09--screencast-studio/gen/go/proto/screencast/studio/v1"
	apppkg "github.com/wesen/2026-04-09--screencast-studio/pkg/app"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/discovery"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/media"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/recording"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type fakeApplication struct {
	discoverySnapshot *discovery.Snapshot
	normalizeConfig   *dsl.EffectiveConfig
	normalizeErr      error
	compilePlan       *dsl.CompiledPlan
	compileErr        error
	recordDelay       time.Duration
	recordErr         error
	recordStarted     chan struct{}
}

type fakePreviewRunner struct {
	started   chan struct{}
	waitDelay time.Duration
	runs      atomic.Int32
}

type fakePreviewSession struct {
	ctx       context.Context
	waitDelay time.Duration
}

func boolPtr(v bool) *bool {
	return &v
}

func decodeProtoResponse(t *testing.T, body []byte, msg proto.Message) {
	t.Helper()
	if err := protojson.Unmarshal(body, msg); err != nil {
		t.Fatalf("unmarshal proto response: %v", err)
	}
}

func (f *fakeApplication) DiscoverySnapshot(ctx context.Context) (*discovery.Snapshot, error) {
	return f.discoverySnapshot, nil
}

func (f *fakeApplication) NormalizeDSL(ctx context.Context, body []byte) (*dsl.EffectiveConfig, error) {
	if f.normalizeErr != nil {
		return nil, f.normalizeErr
	}
	return f.normalizeConfig, nil
}

func (f *fakeApplication) CompileDSL(ctx context.Context, body []byte) (*dsl.CompiledPlan, error) {
	if f.compileErr != nil {
		return nil, f.compileErr
	}
	return f.compilePlan, nil
}

func (f *fakeApplication) SetRecordingAudioGain(ctx context.Context, sessionID, sourceID string, gain float64) error {
	_ = ctx
	_ = sessionID
	_ = sourceID
	_ = gain
	return nil
}

func (f *fakeApplication) SetRecordingCompressorEnabled(ctx context.Context, sessionID string, enabled bool) error {
	_ = ctx
	_ = sessionID
	_ = enabled
	return nil
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
			Message:      "fake media runtime log line",
		})
	}
	select {
	case <-ctx.Done():
	case <-time.After(f.recordDelay):
	}
	if f.recordErr != nil {
		return nil, f.recordErr
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

func (f *fakePreviewRunner) StartPreview(ctx context.Context, source dsl.EffectiveVideoSource, opts media.PreviewOptions) (media.PreviewSession, error) {
	f.runs.Add(1)
	if f.started != nil {
		select {
		case f.started <- struct{}{}:
		default:
		}
	}
	if opts.OnLog != nil {
		opts.OnLog("stderr", "fake preview started")
	}
	if opts.OnFrame != nil {
		opts.OnFrame([]byte{0xff, 0xd8, 0xff, 0xd9})
	}
	return &fakePreviewSession{ctx: ctx, waitDelay: f.waitDelay}, nil
}

func (s *fakePreviewSession) Wait() error {
	if s == nil || s.ctx == nil {
		return nil
	}
	<-s.ctx.Done()
	if s.waitDelay > 0 {
		time.Sleep(s.waitDelay)
	}
	return nil
}

func (s *fakePreviewSession) Stop(ctx context.Context) error {
	return nil
}

func (s *fakePreviewSession) LatestFrame() ([]byte, error) {
	return []byte{0xff, 0xd8, 0xff, 0xd9}, nil
}

func (s *fakePreviewSession) TakeScreenshot(ctx context.Context, opts media.ScreenshotOptions) ([]byte, error) {
	return s.LatestFrame()
}

func TestComputePreviewSignatureIsStableForEquivalentSources(t *testing.T) {
	t.Parallel()

	sourceA := dsl.EffectiveVideoSource{
		ID:      "camera-1",
		Name:    "Camera 1",
		Type:    "camera",
		Enabled: true,
		Target: dsl.VideoTarget{
			Device: "/dev/video0",
			Rect: &dsl.Rect{
				X: 1,
				Y: 2,
				W: 3,
				H: 4,
			},
		},
		Capture: dsl.VideoCaptureSettings{
			FPS:          30,
			Cursor:       boolPtr(false),
			FollowResize: boolPtr(false),
			Mirror:       boolPtr(true),
			Size:         "1280x720",
		},
		Output: dsl.VideoOutputSettings{
			Container:  "mov",
			VideoCodec: "h264",
			Quality:    80,
		},
		DestinationTemplate: "default",
	}
	sourceB := dsl.EffectiveVideoSource{
		ID:      "camera-1",
		Name:    "Camera 1",
		Type:    "camera",
		Enabled: true,
		Target: dsl.VideoTarget{
			Device: "/dev/video0",
			Rect: &dsl.Rect{
				X: 1,
				Y: 2,
				W: 3,
				H: 4,
			},
		},
		Capture: dsl.VideoCaptureSettings{
			FPS:          30,
			Cursor:       boolPtr(false),
			FollowResize: boolPtr(false),
			Mirror:       boolPtr(true),
			Size:         "1280x720",
		},
		Output: dsl.VideoOutputSettings{
			Container:  "mov",
			VideoCodec: "h264",
			Quality:    80,
		},
		DestinationTemplate: "default",
	}

	if gotA, gotB := computePreviewSignature(sourceA), computePreviewSignature(sourceB); gotA != gotB {
		t.Fatalf("preview signature mismatch for equivalent sources: %q != %q", gotA, gotB)
	}
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	server := NewServer(context.Background(), &fakeApplication{}, Config{
		Addr:         ":0",
		PreviewLimit: 7,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	payload := &studiov1.HealthResponse{}
	decodeProtoResponse(t, rec.Body.Bytes(), payload)
	if payload.Service != "screencast-studio" {
		t.Fatalf("service = %q, want screencast-studio", payload.Service)
	}
	if payload.PreviewLimit != 7 {
		t.Fatalf("preview_limit = %d, want 7", payload.PreviewLimit)
	}
	if payload.InitialDsl != nil {
		t.Fatalf("initial_dsl = %q, want nil", payload.GetInitialDsl())
	}
	if payload.InitialDslPath != nil {
		t.Fatalf("initial_dsl_path = %q, want nil", payload.GetInitialDslPath())
	}
}

func TestHealthzIncludesInitialDSL(t *testing.T) {
	t.Parallel()

	server := NewServer(context.Background(), &fakeApplication{}, Config{
		Addr:           ":0",
		PreviewLimit:   4,
		InitialDSL:     "schema: recorder.config/v1\nsession_id: preload\n",
		InitialDSLPath: "./examples/preload.yaml",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	payload := &studiov1.HealthResponse{}
	decodeProtoResponse(t, rec.Body.Bytes(), payload)
	if payload.GetInitialDsl() != "schema: recorder.config/v1\nsession_id: preload\n" {
		t.Fatalf("initial_dsl = %q", payload.GetInitialDsl())
	}
	if payload.GetInitialDslPath() != "./examples/preload.yaml" {
		t.Fatalf("initial_dsl_path = %q", payload.GetInitialDslPath())
	}
}

func TestDiscoveryEndpoint(t *testing.T) {
	t.Parallel()

	server := NewServer(context.Background(), &fakeApplication{
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

	payload := &studiov1.DiscoveryResponse{}
	decodeProtoResponse(t, rec.Body.Bytes(), payload)
	if len(payload.Displays) != 1 || payload.Displays[0].Id != "display-1" {
		t.Fatalf("unexpected displays payload: %+v", payload.Displays)
	}
	if len(payload.Windows) != 1 || payload.Windows[0].Id != "0x01" {
		t.Fatalf("unexpected windows payload: %+v", payload.Windows)
	}
}

func TestNormalizeAndCompileEndpoints(t *testing.T) {
	t.Parallel()

	server := NewServer(context.Background(), &fakeApplication{
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

	compilePayload := &studiov1.CompileResponse{}
	decodeProtoResponse(t, compileRec.Body.Bytes(), compilePayload)
	if compilePayload.SessionId != "session-1" {
		t.Fatalf("sessionId = %q, want session-1", compilePayload.SessionId)
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
	server := NewServer(context.Background(), fakeApp, Config{})

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

	currentPayload := &studiov1.SessionEnvelope{}
	decodeProtoResponse(t, currentRec.Body.Bytes(), currentPayload)
	if !currentPayload.Session.Active {
		t.Fatalf("expected active session, got %+v", currentPayload.Session)
	}
	if currentPayload.Session.SessionId != "session-2" {
		t.Fatalf("sessionId = %q, want session-2", currentPayload.Session.SessionId)
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
		decodeProtoResponse(t, currentRec.Body.Bytes(), currentPayload)
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
					Capture:             dsl.VideoCaptureSettings{FPS: 5},
					Output:              dsl.VideoOutputSettings{Container: "mkv", VideoCodec: "h264", Quality: 75},
					DestinationTemplate: "default",
				},
			},
		},
	}
	runner := &fakePreviewRunner{started: make(chan struct{}, 1)}
	server := NewServer(context.Background(), fakeApp, Config{PreviewLimit: 2})
	server.previews = NewPreviewManager(context.Background(), fakeApp, server.events.Publish, 2, runner)

	ensureReq := httptest.NewRequest(http.MethodPost, "/api/previews/ensure", bytes.NewBufferString(`{"dsl":"test","sourceId":"display-1"}`))
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

	ensurePayload := &studiov1.PreviewEnvelope{}
	decodeProtoResponse(t, ensureRec.Body.Bytes(), ensurePayload)

	listReq := httptest.NewRequest(http.MethodGet, "/api/previews", nil)
	listRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listRec.Code, http.StatusOK)
	}

	listPayload := &studiov1.PreviewListResponse{}
	decodeProtoResponse(t, listRec.Body.Bytes(), listPayload)
	if len(listPayload.Previews) != 1 {
		t.Fatalf("expected one preview, got %+v", listPayload.Previews)
	}

	releaseReq := httptest.NewRequest(http.MethodPost, "/api/previews/release", bytes.NewBufferString(`{"previewId":"`+ensurePayload.Preview.Id+`"}`))
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
	server := NewServer(context.Background(), fakeApp, Config{PreviewLimit: 2})
	server.previews = NewPreviewManager(context.Background(), fakeApp, server.events.Publish, 2, runner)

	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	ensureReqBody := bytes.NewBufferString(`{"dsl":"test","sourceId":"display-1"}`)
	ensureResp, err := http.Post(ts.URL+"/api/previews/ensure", "application/json", ensureReqBody)
	if err != nil {
		t.Fatalf("ensure request: %v", err)
	}
	defer ensureResp.Body.Close()

	ensurePayload := &studiov1.PreviewEnvelope{}
	body, err := io.ReadAll(ensureResp.Body)
	if err != nil {
		t.Fatalf("read ensure response: %v", err)
	}
	decodeProtoResponse(t, body, ensurePayload)

	select {
	case <-runner.started:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for preview runner")
	}

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/previews/"+ensurePayload.Preview.Id+"/mjpeg", nil)
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

	streamBody, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(streamBody), "--frame") {
		t.Fatalf("expected mjpeg boundary in body, got %q", string(streamBody))
	}

	metricsResp, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatalf("metrics request: %v", err)
	}
	defer metricsResp.Body.Close()
	metricsBody, err := io.ReadAll(metricsResp.Body)
	if err != nil {
		t.Fatalf("read metrics response: %v", err)
	}
	metricsText := string(metricsBody)
	if !strings.Contains(metricsText, `screencast_studio_preview_http_frames_served_total{source_type="display"} `) {
		t.Fatalf("expected preview frames served metric in metrics body, got %s", metricsText)
	}
	if !strings.Contains(metricsText, `screencast_studio_preview_http_write_nanoseconds_total{source_type="display"} `) {
		t.Fatalf("expected preview write timing metric in metrics body, got %s", metricsText)
	}
	if !strings.Contains(metricsText, `screencast_studio_preview_http_flush_nanoseconds_total{source_type="display"} `) {
		t.Fatalf("expected preview flush timing metric in metrics body, got %s", metricsText)
	}
}

func TestRecordingStartSuspendsAndRestoresPreviews(t *testing.T) {
	t.Parallel()

	fakeApp := &fakeApplication{
		normalizeConfig: &dsl.EffectiveConfig{
			Schema:               dsl.SchemaVersion,
			SessionID:            "session-preview-recording",
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
		compilePlan: &dsl.CompiledPlan{
			SessionID: "session-preview-recording",
			Outputs:   []dsl.PlannedOutput{{Kind: "video", Name: "Display", Path: "/tmp/display.mkv"}},
		},
		recordDelay:   10 * time.Second,
		recordStarted: make(chan struct{}, 1),
	}
	runner := &fakePreviewRunner{started: make(chan struct{}, 4)}
	server := NewServer(context.Background(), fakeApp, Config{PreviewLimit: 2})
	server.previews = NewPreviewManager(context.Background(), fakeApp, server.events.Publish, 2, runner)

	ensureReq := httptest.NewRequest(http.MethodPost, "/api/previews/ensure", bytes.NewBufferString(`{"dsl":"test","sourceId":"display-1"}`))
	ensureRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(ensureRec, ensureReq)
	if ensureRec.Code != http.StatusOK {
		t.Fatalf("ensure status = %d, want %d", ensureRec.Code, http.StatusOK)
	}

	select {
	case <-runner.started:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for initial preview to start")
	}

	startReq := httptest.NewRequest(http.MethodPost, "/api/recordings/start", bytes.NewBufferString(`{"dsl":"test"}`))
	startRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusOK {
		t.Fatalf("start status = %d, want %d", startRec.Code, http.StatusOK)
	}

	select {
	case <-fakeApp.recordStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for recording to start")
	}

	previews := server.previews.List()
	if len(previews) != 1 {
		t.Fatalf("expected preview to remain active during recording, got %+v", previews)
	}
	if previews[0].SourceID != "display-1" {
		t.Fatalf("preview source_id = %q, want display-1", previews[0].SourceID)
	}

	stopReq := httptest.NewRequest(http.MethodPost, "/api/recordings/stop", nil)
	stopRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(stopRec, stopReq)
	if stopRec.Code != http.StatusOK {
		t.Fatalf("stop status = %d, want %d", stopRec.Code, http.StatusOK)
	}

	if got := runner.runs.Load(); got != 1 {
		t.Fatalf("expected preview to stay on the shared source without restart, runs=%d", got)
	}
	previews = server.previews.List()
	if len(previews) != 1 {
		t.Fatalf("expected preview to remain active after recording stop, got %+v", previews)
	}
}

func TestRecordingStartFailureRestoresSuspendedPreviews(t *testing.T) {
	t.Parallel()

	fakeApp := &fakeApplication{
		normalizeConfig: &dsl.EffectiveConfig{
			Schema:               dsl.SchemaVersion,
			SessionID:            "session-preview-recording-failure",
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
		compileErr: context.Canceled,
	}
	runner := &fakePreviewRunner{started: make(chan struct{}, 4)}
	server := NewServer(context.Background(), fakeApp, Config{PreviewLimit: 2})
	server.previews = NewPreviewManager(context.Background(), fakeApp, server.events.Publish, 2, runner)

	ensureReq := httptest.NewRequest(http.MethodPost, "/api/previews/ensure", bytes.NewBufferString(`{"dsl":"test","sourceId":"display-1"}`))
	ensureRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(ensureRec, ensureReq)
	if ensureRec.Code != http.StatusOK {
		t.Fatalf("ensure status = %d, want %d", ensureRec.Code, http.StatusOK)
	}

	select {
	case <-runner.started:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for initial preview to start")
	}

	startReq := httptest.NewRequest(http.MethodPost, "/api/recordings/start", bytes.NewBufferString(`{"dsl":"test"}`))
	startRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusBadRequest {
		t.Fatalf("start status = %d, want %d", startRec.Code, http.StatusBadRequest)
	}

	if got := runner.runs.Load(); got != 1 {
		t.Fatalf("expected preview to remain active after start failure without restart, runs=%d", got)
	}
	if previews := server.previews.List(); len(previews) != 1 {
		t.Fatalf("expected preview to remain active after start failure, got %+v", previews)
	}
}

func TestWebsocketEndpoint(t *testing.T) {
	t.Parallel()

	server := NewServer(context.Background(), &fakeApplication{
		compilePlan: &dsl.CompiledPlan{
			SessionID: "session-ws",
			Outputs: []dsl.PlannedOutput{
				{Kind: "video", Name: "Display", Path: "/tmp/session-ws/display.mkv"},
			},
			AudioJobs: []dsl.AudioMixJob{
				{
					Name: "audio-mix",
					Sources: []dsl.EffectiveAudioSource{
						{ID: "mic-1", Name: "Mic", Device: "default", Enabled: true},
					},
					OutputPath: "/tmp/session-ws/audio.wav",
				},
			},
		},
	}, Config{})
	server.telemetry.UpdateFromPlan(&dsl.CompiledPlan{
		SessionID: "session-ws",
		Outputs: []dsl.PlannedOutput{
			{Kind: "video", Name: "Display", Path: "/tmp/session-ws/display.mkv"},
		},
		AudioJobs: []dsl.AudioMixJob{
			{
				Name: "audio-mix",
				Sources: []dsl.EffectiveAudioSource{
					{ID: "mic-1", Name: "Mic", Device: "default", Enabled: true},
				},
				OutputPath: "/tmp/session-ws/audio.wav",
			},
		},
	})
	server.telemetry.setAudioMeter(audioMeterSnapshot{
		DeviceID:   "default",
		LeftLevel:  0.4,
		RightLevel: 0.35,
		Available:  true,
	})
	server.telemetry.setDiskStatus(diskTelemetrySnapshot{
		Path:        "/tmp/session-ws",
		Filesystem:  "/tmp",
		UsedPercent: 20,
		FreeGiB:     10,
		TotalGiB:    50,
		Available:   true,
	})
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	_, sessionMessage, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read session event: %v", err)
	}
	sessionEvent := &studiov1.ServerEvent{}
	decodeProtoResponse(t, sessionMessage, sessionEvent)
	if sessionEvent.GetSessionState() == nil {
		t.Fatalf("first websocket event = %+v, want sessionState", sessionEvent)
	}

	_, previewMessage, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read preview event: %v", err)
	}
	previewEvent := &studiov1.ServerEvent{}
	decodeProtoResponse(t, previewMessage, previewEvent)
	if previewEvent.GetPreviewList() == nil {
		t.Fatalf("second websocket event = %+v, want previewList", previewEvent)
	}

	_, audioMessage, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read audio meter event: %v", err)
	}
	audioEvent := &studiov1.ServerEvent{}
	decodeProtoResponse(t, audioMessage, audioEvent)
	if audioEvent.GetAudioMeter() == nil {
		t.Fatalf("third websocket event = %+v, want audioMeter", audioEvent)
	}

	_, diskMessage, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read disk status event: %v", err)
	}
	diskEvent := &studiov1.ServerEvent{}
	decodeProtoResponse(t, diskMessage, diskEvent)
	if diskEvent.GetDiskStatus() == nil {
		t.Fatalf("fourth websocket event = %+v, want diskStatus", diskEvent)
	}
}

func TestServeSPAFromFS(t *testing.T) {
	t.Parallel()

	root := fstest.MapFS{
		"index.html":    &fstest.MapFile{Data: []byte("<html>studio</html>")},
		"assets/app.js": &fstest.MapFile{Data: []byte("console.log('ok')")},
	}

	for _, testCase := range []struct {
		name       string
		requestURL string
		wantBody   string
	}{
		{
			name:       "root serves index",
			requestURL: "/",
			wantBody:   "<html>studio</html>",
		},
		{
			name:       "asset serves exact file",
			requestURL: "/assets/app.js",
			wantBody:   "console.log('ok')",
		},
		{
			name:       "deep route falls back to index",
			requestURL: "/studio/sources",
			wantBody:   "<html>studio</html>",
		},
	} {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, testCase.requestURL, nil)
			rec := httptest.NewRecorder()

			if ok := serveSPAFromFS(rec, req, root); !ok {
				t.Fatalf("serveSPAFromFS returned false")
			}

			if body := rec.Body.String(); body != testCase.wantBody {
				t.Fatalf("body = %q, want %q", body, testCase.wantBody)
			}
		})
	}
}

func TestServeSPAFromFSReturnsFalseWhenAssetMissing(t *testing.T) {
	t.Parallel()

	var root fs.FS = fstest.MapFS{}
	req := httptest.NewRequest(http.MethodGet, "/missing.js", nil)
	rec := httptest.NewRecorder()

	if ok := serveSPAFromFS(rec, req, root); ok {
		t.Fatalf("serveSPAFromFS returned true, want false")
	}
}
