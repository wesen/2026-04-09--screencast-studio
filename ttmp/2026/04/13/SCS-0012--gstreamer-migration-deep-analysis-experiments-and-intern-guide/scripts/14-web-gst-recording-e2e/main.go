package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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

type runtimeBackedApp struct {
	normalizeConfig  *dsl.EffectiveConfig
	compilePlan      *dsl.CompiledPlan
	recordingRuntime media.RecordingRuntime
}

func (a *runtimeBackedApp) DiscoverySnapshot(ctx context.Context) (*discovery.Snapshot, error) {
	return &discovery.Snapshot{}, nil
}

func (a *runtimeBackedApp) NormalizeDSL(ctx context.Context, body []byte) (*dsl.EffectiveConfig, error) {
	return a.normalizeConfig, nil
}

func (a *runtimeBackedApp) CompileDSL(ctx context.Context, body []byte) (*dsl.CompiledPlan, error) {
	return a.compilePlan, nil
}

func (a *runtimeBackedApp) SetRecordingAudioGain(ctx context.Context, sessionID, sourceID string, gain float64) error {
	_ = ctx
	if runtime, ok := a.recordingRuntime.(*gstreamer.RecordingRuntime); ok {
		return runtime.SetAudioGain(sessionID, sourceID, gain)
	}
	return fmt.Errorf("recording runtime does not support audio gain control")
}

func (a *runtimeBackedApp) SetRecordingCompressorEnabled(ctx context.Context, sessionID string, enabled bool) error {
	_ = ctx
	if runtime, ok := a.recordingRuntime.(*gstreamer.RecordingRuntime); ok {
		return runtime.SetAudioCompressorEnabled(sessionID, enabled)
	}
	return fmt.Errorf("recording runtime does not support compressor control")
}

func (a *runtimeBackedApp) RecordPlan(ctx context.Context, plan *dsl.CompiledPlan, options apppkg.RecordOptions) (*apppkg.RecordSummary, error) {
	session, err := a.recordingRuntime.StartRecording(ctx, plan, media.RecordingOptions{
		GracePeriod: options.GracePeriod,
		MaxDuration: options.MaxDuration,
		EventSink: func(event media.RecordingEvent) {
			if options.EventSink == nil {
				return
			}
			options.EventSink(recording.RunEvent{
				Type:         recording.RunEventType(event.Type),
				Timestamp:    event.Timestamp,
				State:        recording.SessionState(event.State),
				Reason:       event.Reason,
				ProcessLabel: event.ProcessLabel,
				OutputPath:   event.OutputPath,
				Stream:       event.Stream,
				Message:      event.Message,
			})
		},
	})
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
	fmt.Println("=== Web GStreamer Recording E2E ===")
	fmt.Println()
	runVideoStopWithPreviewCase()
	fmt.Println()
	runAudioMaxDurationCase()
	fmt.Println()
	runParentCancelCase()
	fmt.Println()
	fmt.Println("Recording E2E validation complete.")
}

func runVideoStopWithPreviewCase() {
	fmt.Println("-- Case 1: video recording with active preview + explicit stop --")
	outPath := filepath.Join(os.TempDir(), "scs-gst-web-video-stop.mp4")
	_ = os.Remove(outPath)
	cfg := &dsl.EffectiveConfig{Schema: dsl.SchemaVersion, SessionID: "web-gst-video-stop", DestinationTemplates: map[string]string{"default": outPath}, VideoSources: []dsl.EffectiveVideoSource{{ID: "display-1", Name: "Display 1", Type: "region", Enabled: true, Target: dsl.VideoTarget{Display: envOr("DISPLAY", ":0"), Rect: &dsl.Rect{X: 0, Y: 0, W: 640, H: 480}}, Capture: dsl.VideoCaptureSettings{FPS: 10}, Output: dsl.VideoOutputSettings{Container: "mp4", VideoCodec: "h264", Quality: 75}, DestinationTemplate: "default"}}}
	plan := &dsl.CompiledPlan{SessionID: "web-gst-video-stop", VideoJobs: []dsl.VideoJob{{Source: cfg.VideoSources[0], OutputPath: outPath}}, Outputs: []dsl.PlannedOutput{{Kind: "video", SourceID: "display-1", Name: "Display 1", Path: outPath}}}
	serverURL, closeFn, _ := newServer(cfg, plan)
	defer closeFn()

	previewID := ensurePreview(serverURL, "display-1")
	waitPreviewCount(serverURL, 1, 4*time.Second, "preview active")
	startRecording(serverURL, 0)
	waitPreviewCount(serverURL, 0, 4*time.Second, "preview suspended")
	time.Sleep(1200 * time.Millisecond)
	stopRecording(serverURL)
	waitSessionInactive(serverURL, 8*time.Second, "video stop session finished")
	waitPreviewCount(serverURL, 1, 6*time.Second, "preview restored")
	validateMediaFile(outPath, "video-stop mp4")
	releasePreview(serverURL, previewID)
}

func runAudioMaxDurationCase() {
	fmt.Println("-- Case 2: audio recording with max-duration timeout --")
	outPath := filepath.Join(os.TempDir(), "scs-gst-web-audio-timeout.wav")
	_ = os.Remove(outPath)
	cfg := &dsl.EffectiveConfig{Schema: dsl.SchemaVersion, SessionID: "web-gst-audio-timeout", DestinationTemplates: map[string]string{"default": outPath}}
	plan := &dsl.CompiledPlan{SessionID: "web-gst-audio-timeout", AudioJobs: []dsl.AudioMixJob{{Name: "audio-mix", Sources: []dsl.EffectiveAudioSource{{ID: "mic-1", Name: "Mic", Enabled: true, Device: "default", Settings: dsl.AudioSourceSettings{Gain: 1.0}}}, Output: dsl.AudioOutputSettings{Codec: "wav", SampleRateHz: 48000, Channels: 2}, OutputPath: outPath}}, Outputs: []dsl.PlannedOutput{{Kind: "audio", Name: "audio-mix", Path: outPath}}}
	serverURL, closeFn, _ := newServer(cfg, plan)
	defer closeFn()

	startRecording(serverURL, 2)
	session := waitSessionInactive(serverURL, 8*time.Second, "audio timeout session finished")
	if !strings.Contains(session.GetReason(), "max duration") {
		fatal("expected max duration reason, got %q", session.GetReason())
	}
	validateMediaFile(outPath, "audio-timeout wav")
}

func runParentCancelCase() {
	fmt.Println("-- Case 3: recording canceled by parent runtime context --")
	outPath := filepath.Join(os.TempDir(), "scs-gst-web-parent-cancel.mp4")
	_ = os.Remove(outPath)
	cfg := &dsl.EffectiveConfig{Schema: dsl.SchemaVersion, SessionID: "web-gst-parent-cancel", DestinationTemplates: map[string]string{"default": outPath}, VideoSources: []dsl.EffectiveVideoSource{{ID: "display-1", Name: "Display 1", Type: "region", Enabled: true, Target: dsl.VideoTarget{Display: envOr("DISPLAY", ":0"), Rect: &dsl.Rect{X: 0, Y: 0, W: 640, H: 480}}, Capture: dsl.VideoCaptureSettings{FPS: 10}, Output: dsl.VideoOutputSettings{Container: "mp4", VideoCodec: "h264", Quality: 75}, DestinationTemplate: "default"}}}
	plan := &dsl.CompiledPlan{SessionID: "web-gst-parent-cancel", VideoJobs: []dsl.VideoJob{{Source: cfg.VideoSources[0], OutputPath: outPath}}, Outputs: []dsl.PlannedOutput{{Kind: "video", SourceID: "display-1", Name: "Display 1", Path: outPath}}}
	parentCtx, cancel := context.WithCancel(context.Background())
	serverURL, closeFn, _ := newServerWithParent(parentCtx, cfg, plan)
	defer closeFn()

	startRecording(serverURL, 0)
	time.Sleep(1200 * time.Millisecond)
	cancel()
	session := waitSessionInactive(serverURL, 8*time.Second, "parent cancel session finished")
	if session.GetReason() == "" {
		fatal("expected parent cancel reason, got empty")
	}
	validateMediaFile(outPath, "parent-cancel mp4")
}

func newServer(cfg *dsl.EffectiveConfig, plan *dsl.CompiledPlan) (string, func(), *runtimeBackedApp) {
	return newServerWithParent(context.Background(), cfg, plan)
}

func newServerWithParent(parent context.Context, cfg *dsl.EffectiveConfig, plan *dsl.CompiledPlan) (string, func(), *runtimeBackedApp) {
	app := &runtimeBackedApp{normalizeConfig: cfg, compilePlan: plan, recordingRuntime: gstreamer.NewRecordingRuntime()}
	server := web.NewServerWithOptions(parent, app, web.Config{PreviewLimit: 2}, web.WithPreviewRuntime(gstreamer.NewPreviewRuntime()))
	ts := httptest.NewServer(server.Handler())
	return ts.URL, ts.Close, app
}

func ensurePreview(baseURL, sourceID string) string {
	resp, err := http.Post(baseURL+"/api/previews/ensure", "application/json", bytes.NewBufferString(fmt.Sprintf(`{"dsl":"test","sourceId":%q}`, sourceID)))
	must(err, "ensure preview request")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fatal("ensure preview status %d body=%s", resp.StatusCode, string(body))
	}
	var envelope studiov1.PreviewEnvelope
	must(protojson.Unmarshal(body, &envelope), "decode preview ensure")
	fmt.Printf("preview ensured: %s\n", envelope.GetPreview().GetId())
	return envelope.GetPreview().GetId()
}

func releasePreview(baseURL, previewID string) {
	resp, err := http.Post(baseURL+"/api/previews/release", "application/json", bytes.NewBufferString(fmt.Sprintf(`{"previewId":%q}`, previewID)))
	must(err, "release preview request")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fatal("release preview status %d body=%s", resp.StatusCode, string(body))
	}
	fmt.Printf("preview released: %s\n", previewID)
}

func startRecording(baseURL string, maxDurationSeconds int) {
	resp, err := http.Post(baseURL+"/api/recordings/start", "application/json", bytes.NewBufferString(fmt.Sprintf(`{"dsl":"test","maxDurationSeconds":%d}`, maxDurationSeconds)))
	must(err, "start recording request")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fatal("start recording status %d body=%s", resp.StatusCode, string(body))
	}
	fmt.Println("recording started")
}

func stopRecording(baseURL string) {
	resp, err := http.Post(baseURL+"/api/recordings/stop", "application/json", bytes.NewBufferString(`{}`))
	must(err, "stop recording request")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fatal("stop recording status %d body=%s", resp.StatusCode, string(body))
	}
	fmt.Println("recording stop requested")
}

func waitPreviewCount(baseURL string, want int, timeout time.Duration, label string) {
	deadline := time.Now().Add(timeout)
	for {
		resp, err := http.Get(baseURL + "/api/previews")
		must(err, "list previews request")
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fatal("list previews status %d body=%s", resp.StatusCode, string(body))
		}
		var list studiov1.PreviewListResponse
		must(protojson.Unmarshal(body, &list), "decode preview list")
		if len(list.GetPreviews()) == want {
			fmt.Printf("%s: preview count=%d\n", label, want)
			return
		}
		if time.Now().After(deadline) {
			fatal("%s: expected preview count %d got %d body=%s", label, want, len(list.GetPreviews()), string(body))
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func waitSessionInactive(baseURL string, timeout time.Duration, label string) *studiov1.RecordingSession {
	deadline := time.Now().Add(timeout)
	for {
		resp, err := http.Get(baseURL + "/api/recordings/current")
		must(err, "get current session request")
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fatal("current session status %d body=%s", resp.StatusCode, string(body))
		}
		var envelope studiov1.SessionEnvelope
		must(protojson.Unmarshal(body, &envelope), "decode session envelope")
		session := envelope.GetSession()
		if session != nil && !session.GetActive() && session.GetSessionId() != "" {
			fmt.Printf("%s: state=%s reason=%q\n", label, session.GetState(), session.GetReason())
			return session
		}
		if time.Now().After(deadline) {
			fatal("%s: session did not become inactive body=%s", label, string(body))
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func validateMediaFile(path, label string) {
	info, err := os.Stat(path)
	must(err, "stat media file")
	if info.Size() == 0 {
		fatal("%s: output file is empty", label)
	}
	fmt.Printf("%s: %d bytes\n", label, info.Size())
	out, err := exec.Command("ffprobe", "-hide_banner", "-loglevel", "error", "-show_entries", "format=duration,size", "-of", "default=noprint_wrappers=1:nokey=0", path).CombinedOutput()
	must(err, "ffprobe media file")
	fmt.Printf("%s ffprobe:\n%s", label, string(out))
	if !strings.Contains(string(out), "duration=") || !strings.Contains(string(out), "size=") {
		fatal("%s: ffprobe output missing expected fields: %s", label, string(out))
	}
}

func envOr(key, def string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return def
}

func must(err error, msg string) {
	if err != nil {
		fatal("%s: %v", msg, err)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
