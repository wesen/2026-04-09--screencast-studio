package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"time"

	studiov1 "github.com/wesen/2026-04-09--screencast-studio/gen/go/proto/screencast/studio/v1"
	"github.com/wesen/2026-04-09--screencast-studio/internal/web"
	apppkg "github.com/wesen/2026-04-09--screencast-studio/pkg/app"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/discovery"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	gstreamer "github.com/wesen/2026-04-09--screencast-studio/pkg/media/gst"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/recording"
	"google.golang.org/protobuf/encoding/protojson"
)

type fakeApp struct {
	normalizeConfig *dsl.EffectiveConfig
	compilePlan     *dsl.CompiledPlan
	recordDelay     time.Duration
}

func (f *fakeApp) DiscoverySnapshot(ctx context.Context) (*discovery.Snapshot, error) {
	return &discovery.Snapshot{}, nil
}

func (f *fakeApp) NormalizeDSL(ctx context.Context, body []byte) (*dsl.EffectiveConfig, error) {
	return f.normalizeConfig, nil
}

func (f *fakeApp) CompileDSL(ctx context.Context, body []byte) (*dsl.CompiledPlan, error) {
	return f.compilePlan, nil
}

func (f *fakeApp) RecordPlan(ctx context.Context, plan *dsl.CompiledPlan, options apppkg.RecordOptions) (*apppkg.RecordSummary, error) {
	started := time.Now()
	if options.EventSink != nil {
		options.EventSink(recording.RunEvent{Type: recording.RunEventStateChanged, Timestamp: started, State: recording.StateRunning, Reason: "fake recording running"})
	}
	select {
	case <-ctx.Done():
		finished := time.Now()
		if options.EventSink != nil {
			options.EventSink(recording.RunEvent{Type: recording.RunEventStateChanged, Timestamp: finished, State: recording.StateFinished, Reason: ctx.Err().Error()})
		}
		return &apppkg.RecordSummary{SessionID: plan.SessionID, State: string(recording.StateFinished), Reason: ctx.Err().Error(), Outputs: append([]dsl.PlannedOutput(nil), plan.Outputs...), StartedAt: started, FinishedAt: finished}, nil
	case <-time.After(f.recordDelay):
	}
	finished := time.Now()
	if options.EventSink != nil {
		options.EventSink(recording.RunEvent{Type: recording.RunEventStateChanged, Timestamp: finished, State: recording.StateFinished, Reason: "fake recording finished"})
	}
	return &apppkg.RecordSummary{SessionID: plan.SessionID, State: string(recording.StateFinished), Reason: "fake recording finished", Outputs: append([]dsl.PlannedOutput(nil), plan.Outputs...), StartedAt: started, FinishedAt: finished}, nil
}

func main() {
	source := buildSourceFromEnv()
	cfg := &dsl.EffectiveConfig{
		Schema:               dsl.SchemaVersion,
		SessionID:            "gst-web-preview-session",
		DestinationTemplates: map[string]string{"default": "/tmp/out.mkv"},
		VideoSources:         []dsl.EffectiveVideoSource{source},
	}
	plan := &dsl.CompiledPlan{
		SessionID: "gst-web-preview-session",
		Outputs:   []dsl.PlannedOutput{{Kind: "video", Name: source.Name, Path: "/tmp/fake-recording.mkv"}},
	}
	app := &fakeApp{normalizeConfig: cfg, compilePlan: plan, recordDelay: 1200 * time.Millisecond}
	server := web.NewServerWithOptions(context.Background(), app, web.Config{PreviewLimit: 2}, web.WithPreviewRuntime(gstreamer.NewPreviewRuntime()))
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	fmt.Println("=== Web GStreamer Preview E2E ===")
	fmt.Printf("source type: %s\n", source.Type)
	fmt.Printf("source id:   %s\n", source.ID)
	fmt.Println()

	previewID := ensurePreview(ts.URL, source.ID)
	readMJPEG(ts.URL, previewID, "initial")
	waitForPreviewCount(ts.URL, 1, 4*time.Second, "preview active after ensure")

	fmt.Println("starting fake recording to validate preview suspend/restore...")
	startRecording(ts.URL)
	waitForPreviewCount(ts.URL, 0, 4*time.Second, "preview suspended during recording")
	waitForPreviewCount(ts.URL, 1, 6*time.Second, "preview restored after recording")
	readMJPEG(ts.URL, previewID, "restored")
	releasePreview(ts.URL, previewID)

	fmt.Println()
	fmt.Println("E2E validation complete.")
}

func ensurePreview(baseURL, sourceID string) string {
	resp, err := http.Post(baseURL+"/api/previews/ensure", "application/json", bytes.NewBufferString(fmt.Sprintf(`{"dsl":"test","sourceId":%q}`, sourceID)))
	must(err, "ensure preview request")
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	must(err, "read ensure preview body")
	if resp.StatusCode != http.StatusOK {
		fatal("ensure preview status %d body=%s", resp.StatusCode, string(body))
	}
	var envelope studiov1.PreviewEnvelope
	must(protojson.Unmarshal(body, &envelope), "decode ensure preview body")
	if envelope.GetPreview() == nil || envelope.GetPreview().GetId() == "" {
		fatal("missing preview id in ensure response: %s", string(body))
	}
	fmt.Printf("ensure preview: %s\n", envelope.GetPreview().GetId())
	return envelope.GetPreview().GetId()
}

func releasePreview(baseURL, previewID string) {
	resp, err := http.Post(baseURL+"/api/previews/release", "application/json", bytes.NewBufferString(fmt.Sprintf(`{"previewId":%q}`, previewID)))
	must(err, "release preview request")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fatal("release preview status %d body=%s", resp.StatusCode, string(body))
	}
	fmt.Printf("release preview: %s\n", previewID)
}

func startRecording(baseURL string) {
	resp, err := http.Post(baseURL+"/api/recordings/start", "application/json", bytes.NewBufferString(`{"dsl":"test"}`))
	must(err, "start recording request")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fatal("start recording status %d body=%s", resp.StatusCode, string(body))
	}
}

func readMJPEG(baseURL, previewID, label string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/previews/"+previewID+"/mjpeg", nil)
	must(err, "create mjpeg request")
	resp, err := http.DefaultClient.Do(req)
	must(err, "perform mjpeg request")
	defer resp.Body.Close()
	buf := make([]byte, 8192)
	n, err := resp.Body.Read(buf)
	if err != nil && err != io.EOF {
		must(err, "read mjpeg body")
	}
	chunk := string(buf[:n])
	if !strings.Contains(chunk, "--frame") {
		fatal("mjpeg response for %s preview missing boundary: %q", label, chunk)
	}
	fmt.Printf("mjpeg %s: received frame boundary\n", label)
}

func waitForPreviewCount(baseURL string, want int, timeout time.Duration, label string) {
	deadline := time.Now().Add(timeout)
	for {
		count, body := getPreviewCount(baseURL)
		if count == want {
			fmt.Printf("%s: preview count=%d\n", label, count)
			return
		}
		if time.Now().After(deadline) {
			fatal("%s: expected preview count %d, got %d body=%s", label, want, count, body)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func getPreviewCount(baseURL string) (int, string) {
	resp, err := http.Get(baseURL + "/api/previews")
	must(err, "list previews request")
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	must(err, "read preview list body")
	if resp.StatusCode != http.StatusOK {
		fatal("list previews status %d body=%s", resp.StatusCode, string(body))
	}
	var list studiov1.PreviewListResponse
	must(protojson.Unmarshal(body, &list), "decode preview list body")
	return len(list.GetPreviews()), string(body)
}

func buildSourceFromEnv() dsl.EffectiveVideoSource {
	sourceType := envOr("SOURCE_TYPE", "display")
	source := dsl.EffectiveVideoSource{
		ID:      "preview-source-1",
		Name:    "Preview Source",
		Type:    sourceType,
		Enabled: true,
		Target: dsl.VideoTarget{
			Display:  envOr("DISPLAY_NAME", envOr("DISPLAY", ":0")),
			Device:   envOr("DEVICE", ""),
			WindowID: envOr("WINDOW_ID", ""),
		},
		Capture:             dsl.VideoCaptureSettings{FPS: 24},
		Output:              dsl.VideoOutputSettings{Container: "mkv", VideoCodec: "h264", Quality: 75},
		DestinationTemplate: "default",
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
		must(err, "parse REGION component")
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

func must(err error, msg string) {
	if err != nil {
		fatal("%s: %v", msg, err)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
