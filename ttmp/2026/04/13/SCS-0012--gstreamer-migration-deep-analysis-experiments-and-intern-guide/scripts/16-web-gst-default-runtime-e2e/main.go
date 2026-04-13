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

	"github.com/gorilla/websocket"
	studiov1 "github.com/wesen/2026-04-09--screencast-studio/gen/go/proto/screencast/studio/v1"
	"github.com/wesen/2026-04-09--screencast-studio/internal/web"
	apppkg "github.com/wesen/2026-04-09--screencast-studio/pkg/app"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	fmt.Println("=== Web GStreamer Default Runtime E2E ===")
	root := filepath.Join(os.TempDir(), "scs-phase4-default-runtime")
	_ = os.RemoveAll(root)
	must(os.MkdirAll(root, 0o755), "mkdir root")
	dslBody := buildDSL(root)

	app := apppkg.New()
	server := web.NewServer(context.Background(), app, web.Config{PreviewLimit: 2})
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	previewID := ensurePreview(ts.URL, dslBody, "display-1")
	waitPreviewCount(ts.URL, 1, 4*time.Second, "preview active before recording")
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	must(err, "dial websocket")
	defer conn.Close()
	drainInitialEvents(conn)

	outputs := compileOutputs(ts.URL, dslBody)
	videoPath := outputs["video"]
	audioPath := outputs["audio"]
	startRecording(ts.URL, dslBody, 4)
	waitPreviewCount(ts.URL, 1, 4*time.Second, "preview still active during recording")
	shot := waitForScreenshot(ts.URL, previewID, 4*time.Second)
	fmt.Printf("screenshot during recording: %d bytes\n", len(shot))
	waitForAudioMeter(conn, 4*time.Second)
	postJSON(ts.URL+"/api/audio/effects", `{"source_id":"mic-1","gain":1.5,"compressor_enabled":true}`)
	fmt.Println("audio effects update accepted during recording")
	waitSessionInactive(ts.URL, 20*time.Second)
	waitPreviewCount(ts.URL, 1, 4*time.Second, "preview still active after recording")
	ensureSamePreviewID(ts.URL, previewID)
	validateMediaFile(videoPath, "default-runtime video")
	validateMediaFile(audioPath, "default-runtime audio")
	fmt.Println("Default runtime E2E complete.")
}

func buildDSL(root string) string {
	return fmt.Sprintf(`schema: "recorder.config/v1"
session_id: "phase4-default-runtime"

destination_templates:
  video_out: "%s/{session_id}/{source_name}.{ext}"
  audio_out: "%s/{session_id}/audio-mix.{ext}"

audio_defaults:
  output:
    codec: "pcm_s16le"
    sample_rate_hz: 48000
    channels: 2

audio_mix:
  destination_template: "audio_out"

video_sources:
  - id: "display-1"
    name: "Display 1"
    type: "region"
    enabled: true
    target:
      display: "%s"
      rect:
        x: 0
        y: 0
        w: 640
        h: 480
    settings:
      capture:
        fps: 10
        cursor: true
      output:
        container: "mp4"
        video_codec: "h264"
        quality: 75
    destination_template: "video_out"

audio_sources:
  - id: "mic-1"
    name: "default"
    device: "default"
    enabled: true
    settings:
      gain: 1.0
      noise_gate: false
      denoise: false
`, root, root, envOr("DISPLAY", ":0"))
}

func ensurePreview(baseURL, dslBody, sourceID string) string {
	resp, err := http.Post(baseURL+"/api/previews/ensure", "application/json", bytes.NewBufferString(fmt.Sprintf(`{"dsl":%q,"sourceId":%q}`, dslBody, sourceID)))
	must(err, "ensure preview")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fatal("ensure preview status %d body=%s", resp.StatusCode, string(body))
	}
	var env studiov1.PreviewEnvelope
	must(protojson.Unmarshal(body, &env), "decode preview envelope")
	fmt.Printf("preview ensured: %s\n", env.GetPreview().GetId())
	return env.GetPreview().GetId()
}

func compileOutputs(baseURL, dslBody string) map[string]string {
	resp, err := http.Post(baseURL+"/api/setup/compile", "application/json", bytes.NewBufferString(fmt.Sprintf(`{"dsl":%q}`, dslBody)))
	must(err, "compile request")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fatal("compile status %d body=%s", resp.StatusCode, string(body))
	}
	var result studiov1.CompileResponse
	must(protojson.Unmarshal(body, &result), "decode compile response")
	out := map[string]string{}
	for _, entry := range result.GetOutputs() {
		switch entry.GetKind() {
		case "video":
			out["video"] = entry.GetPath()
		case "audio":
			out["audio"] = entry.GetPath()
		}
	}
	if out["video"] == "" || out["audio"] == "" {
		fatal("missing expected outputs in compile response: %+v", out)
	}
	return out
}

func startRecording(baseURL, dslBody string, maxDurationSeconds int) {
	resp, err := http.Post(baseURL+"/api/recordings/start", "application/json", bytes.NewBufferString(fmt.Sprintf(`{"dsl":%q,"maxDurationSeconds":%d}`, dslBody, maxDurationSeconds)))
	must(err, "start recording")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fatal("start recording status %d body=%s", resp.StatusCode, string(body))
	}
	fmt.Println("recording started")
}

func waitPreviewCount(baseURL string, want int, timeout time.Duration, label string) {
	deadline := time.Now().Add(timeout)
	for {
		resp, err := http.Get(baseURL + "/api/previews")
		must(err, "list previews")
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var list studiov1.PreviewListResponse
		must(protojson.Unmarshal(body, &list), "decode preview list")
		if len(list.GetPreviews()) == want {
			fmt.Printf("%s: preview count=%d\n", label, want)
			return
		}
		if time.Now().After(deadline) {
			fatal("%s: expected preview count %d got %d", label, want, len(list.GetPreviews()))
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func ensureSamePreviewID(baseURL, previewID string) {
	resp, err := http.Get(baseURL + "/api/previews")
	must(err, "list previews for id check")
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var list studiov1.PreviewListResponse
	must(protojson.Unmarshal(body, &list), "decode preview list for id check")
	if len(list.GetPreviews()) != 1 || list.GetPreviews()[0].GetId() != previewID {
		fatal("expected same preview id %q after recording, got %+v", previewID, list.GetPreviews())
	}
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
			fatal("timed out waiting for screenshot during recording")
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func drainInitialEvents(conn *websocket.Conn) {
	for i := 0; i < 4; i++ {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		if _, _, err := conn.ReadMessage(); err != nil {
			fatal("drain initial event %d: %v", i, err)
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
			fatal("wait for audio meter: %v", err)
		}
		var event studiov1.ServerEvent
		if err := protojson.Unmarshal(msg, &event); err != nil {
			continue
		}
		if meter := event.GetAudioMeter(); meter != nil && meter.GetAvailable() {
			fmt.Printf("audio meter during recording: left=%.3f right=%.3f\n", meter.GetLeftLevel(), meter.GetRightLevel())
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
			fmt.Printf("recording finished: state=%s reason=%q\n", env.GetSession().GetState(), env.GetSession().GetReason())
			return
		}
		if time.Now().After(deadline) {
			fatal("recording did not finish in time body=%s", string(body))
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func validateMediaFile(path, label string) {
	info, err := os.Stat(path)
	must(err, "stat media file")
	if info.Size() == 0 {
		fatal("%s: file is empty", label)
	}
	out, err := exec.Command("ffprobe", "-hide_banner", "-loglevel", "error", "-show_entries", "format=duration,size", "-of", "default=noprint_wrappers=1:nokey=0", path).CombinedOutput()
	must(err, "ffprobe media file")
	fmt.Printf("%s (%d bytes):\n%s", label, info.Size(), string(out))
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
func fatal(format string, args ...any) { fmt.Fprintf(os.Stderr, format+"\n", args...); os.Exit(1) }
