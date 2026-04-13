package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	studiov1 "github.com/wesen/2026-04-09--screencast-studio/gen/go/proto/screencast/studio/v1"
	"github.com/wesen/2026-04-09--screencast-studio/internal/web"
	apppkg "github.com/wesen/2026-04-09--screencast-studio/pkg/app"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/discovery"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/media"
	gstreamer "github.com/wesen/2026-04-09--screencast-studio/pkg/media/gst"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/recording"
	"google.golang.org/protobuf/encoding/protojson"
)

type phase3App struct {
	normalizeConfig  *dsl.EffectiveConfig
	compilePlan      *dsl.CompiledPlan
	recordingRuntime *gstreamer.RecordingRuntime
}

func (a *phase3App) DiscoverySnapshot(ctx context.Context) (*discovery.Snapshot, error) {
	return &discovery.Snapshot{}, nil
}
func (a *phase3App) NormalizeDSL(ctx context.Context, body []byte) (*dsl.EffectiveConfig, error) {
	return a.normalizeConfig, nil
}
func (a *phase3App) CompileDSL(ctx context.Context, body []byte) (*dsl.CompiledPlan, error) {
	return a.compilePlan, nil
}
func (a *phase3App) SetRecordingAudioGain(ctx context.Context, sessionID, sourceID string, gain float64) error {
	return a.recordingRuntime.SetAudioGain(sessionID, sourceID, gain)
}
func (a *phase3App) SetRecordingCompressorEnabled(ctx context.Context, sessionID string, enabled bool) error {
	return a.recordingRuntime.SetAudioCompressorEnabled(sessionID, enabled)
}
func (a *phase3App) RecordPlan(ctx context.Context, plan *dsl.CompiledPlan, options apppkg.RecordOptions) (*apppkg.RecordSummary, error) {
	session, err := a.recordingRuntime.StartRecording(ctx, plan, media.RecordingOptions{GracePeriod: options.GracePeriod, MaxDuration: options.MaxDuration, EventSink: func(event media.RecordingEvent) {
		if options.EventSink == nil {
			return
		}
		options.EventSink(recording.RunEvent{Type: recording.RunEventType(event.Type), Timestamp: event.Timestamp, State: recording.SessionState(event.State), Reason: event.Reason, ProcessLabel: event.ProcessLabel, OutputPath: event.OutputPath, Stream: event.Stream, Message: event.Message, DeviceID: event.DeviceID, LeftLevel: event.LeftLevel, RightLevel: event.RightLevel, Available: event.Available})
	}})
	summary := &apppkg.RecordSummary{SessionID: plan.SessionID, Outputs: append([]dsl.PlannedOutput(nil), plan.Outputs...), Warnings: append([]string(nil), plan.Warnings...)}
	if err != nil {
		return summary, err
	}
	result, err := session.Wait()
	if result != nil {
		summary.State = string(result.State)
		summary.Reason = result.Reason
		summary.StartedAt = result.StartedAt
		summary.FinishedAt = result.FinishedAt
	}
	return summary, err
}

func main() {
	fmt.Println("=== Web GStreamer Phase 3 E2E ===")
	fmt.Println()
	runScreenshotCase()
	fmt.Println()
	runAudioEffectsAndMeterCase()
	fmt.Println()
	fmt.Println("Phase 3 E2E validation complete.")
}

func runScreenshotCase() {
	fmt.Println("-- Case 1: screenshot endpoint returns JPEG --")
	cfg := &dsl.EffectiveConfig{Schema: dsl.SchemaVersion, SessionID: "phase3-screenshot", DestinationTemplates: map[string]string{"default": "/tmp/unused.mp4"}, VideoSources: []dsl.EffectiveVideoSource{{ID: "display-1", Name: "Display 1", Type: "region", Enabled: true, Target: dsl.VideoTarget{Display: envOr("DISPLAY", ":0"), Rect: &dsl.Rect{X: 0, Y: 0, W: 640, H: 480}}, Capture: dsl.VideoCaptureSettings{FPS: 10}, Output: dsl.VideoOutputSettings{Container: "mp4", VideoCodec: "h264", Quality: 75}, DestinationTemplate: "default"}}}
	plan := &dsl.CompiledPlan{SessionID: "phase3-screenshot"}
	baseURL, closeFn := newPhase3Server(cfg, plan)
	defer closeFn()

	previewID := ensurePreview(baseURL, "display-1")
	waitPreviewCount(baseURL, 1, 4*time.Second)
	shot := waitForScreenshot(baseURL, previewID, 4*time.Second)
	if len(shot) < 4 || shot[0] != 0xff || shot[1] != 0xd8 {
		fatal("screenshot does not look like JPEG, first bytes=%v", shot[:min(len(shot), 4)])
	}
	fmt.Printf("screenshot bytes: %d\n", len(shot))
}

func runAudioEffectsAndMeterCase() {
	fmt.Println("-- Case 2: audio effects endpoint + audio meter websocket events --")
	outPath := filepath.Join(os.TempDir(), "scs-gst-phase3-audio.wav")
	_ = os.Remove(outPath)
	cfg := &dsl.EffectiveConfig{Schema: dsl.SchemaVersion, SessionID: "phase3-audio", DestinationTemplates: map[string]string{"default": outPath}}
	plan := &dsl.CompiledPlan{SessionID: "phase3-audio", AudioJobs: []dsl.AudioMixJob{{Name: "audio-mix", Sources: []dsl.EffectiveAudioSource{{ID: "mic-1", Name: "Mic", Enabled: true, Device: "default", Settings: dsl.AudioSourceSettings{Gain: 1.0}}}, Output: dsl.AudioOutputSettings{Codec: "wav", SampleRateHz: 48000, Channels: 2}, OutputPath: outPath}}, Outputs: []dsl.PlannedOutput{{Kind: "audio", Name: "audio-mix", Path: outPath}}}
	baseURL, closeFn := newPhase3Server(cfg, plan)
	defer closeFn()

	wsURL := "ws" + strings.TrimPrefix(baseURL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	must(err, "dial websocket")
	defer conn.Close()
	drainInitialEvents(conn)

	startRecording(baseURL, 4)
	waitForAudioMeter(conn, 4*time.Second)
	postJSON(baseURL+"/api/audio/effects", `{"source_id":"mic-1","gain":1.5,"compressor_enabled":true}`)
	fmt.Println("audio effects update accepted")
	waitSessionInactive(baseURL, 8*time.Second)
	info, err := os.Stat(outPath)
	must(err, "stat audio output")
	if info.Size() == 0 {
		fatal("phase3 audio output is empty")
	}
	fmt.Printf("audio output bytes: %d\n", info.Size())
}

func newPhase3Server(cfg *dsl.EffectiveConfig, plan *dsl.CompiledPlan) (string, func()) {
	app := &phase3App{normalizeConfig: cfg, compilePlan: plan, recordingRuntime: gstreamer.NewRecordingRuntime()}
	server := web.NewServerWithOptions(context.Background(), app, web.Config{PreviewLimit: 2}, web.WithPreviewRuntime(gstreamer.NewPreviewRuntime()))
	ts := httptest.NewServer(server.Handler())
	return ts.URL, ts.Close
}

func ensurePreview(baseURL, sourceID string) string {
	resp, err := http.Post(baseURL+"/api/previews/ensure", "application/json", bytes.NewBufferString(fmt.Sprintf(`{"dsl":"test","sourceId":%q}`, sourceID)))
	must(err, "ensure preview")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fatal("ensure preview status %d body=%s", resp.StatusCode, string(body))
	}
	var env studiov1.PreviewEnvelope
	must(protojson.Unmarshal(body, &env), "decode preview envelope")
	return env.GetPreview().GetId()
}

func waitForScreenshot(baseURL, previewID string, timeout time.Duration) []byte {
	deadline := time.Now().Add(timeout)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/previews/"+previewID+"/screenshot", nil)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			cancel()
			if resp.StatusCode == http.StatusOK {
				return body
			}
		} else {
			cancel()
		}
		if time.Now().After(deadline) {
			fatal("timed out waiting for preview screenshot")
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func waitPreviewCount(baseURL string, want int, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for {
		resp, err := http.Get(baseURL + "/api/previews")
		must(err, "list previews")
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var list studiov1.PreviewListResponse
		must(protojson.Unmarshal(body, &list), "decode preview list")
		if len(list.GetPreviews()) == want {
			return
		}
		if time.Now().After(deadline) {
			fatal("expected preview count %d got %d", want, len(list.GetPreviews()))
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func startRecording(baseURL string, maxDurationSeconds int) {
	resp, err := http.Post(baseURL+"/api/recordings/start", "application/json", bytes.NewBufferString(fmt.Sprintf(`{"dsl":"test","maxDurationSeconds":%d}`, maxDurationSeconds)))
	must(err, "start recording")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fatal("start recording status %d body=%s", resp.StatusCode, string(body))
	}
}

func waitSessionInactive(baseURL string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for {
		resp, err := http.Get(baseURL + "/api/recordings/current")
		must(err, "current session")
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var env studiov1.SessionEnvelope
		must(protojson.Unmarshal(body, &env), "decode session envelope")
		if env.GetSession() != nil && !env.GetSession().GetActive() && env.GetSession().GetSessionId() != "" {
			return
		}
		if time.Now().After(deadline) {
			fatal("session did not finish in time body=%s", string(body))
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func drainInitialEvents(conn *websocket.Conn) {
	for i := 0; i < 4; i++ {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if _, _, err := conn.ReadMessage(); err != nil {
			fatal("drain initial websocket event %d: %v", i, err)
		}
	}
	_ = conn.SetReadDeadline(time.Time{})
}

func waitForAudioMeter(conn *websocket.Conn, timeout time.Duration) {
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	defer conn.SetReadDeadline(time.Time{})
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fatal("timed out waiting for audio meter websocket event: %v", err)
		}
		var event studiov1.ServerEvent
		if err := protojson.Unmarshal(msg, &event); err != nil {
			continue
		}
		if meter := event.GetAudioMeter(); meter != nil && meter.GetAvailable() {
			fmt.Printf("audio meter: left=%.3f right=%.3f\n", meter.GetLeftLevel(), meter.GetRightLevel())
			return
		}
	}
}

func postJSON(url, body string) {
	resp, err := http.Post(url, "application/json", bytes.NewBufferString(body))
	must(err, "post json")
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fatal("post json status %d body=%s", resp.StatusCode, string(respBody))
	}
}

func envOr(key, def string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return def
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func must(err error, msg string) {
	if err != nil {
		fatal("%s: %v", msg, err)
	}
}
func fatal(format string, args ...any) { fmt.Fprintf(os.Stderr, format+"\n", args...); os.Exit(1) }
