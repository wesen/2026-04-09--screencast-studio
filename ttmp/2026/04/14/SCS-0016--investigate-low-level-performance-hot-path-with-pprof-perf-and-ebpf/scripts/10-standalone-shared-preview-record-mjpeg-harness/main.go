package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
	"github.com/wesen/2026-04-09--screencast-studio/pkg/media"
	gstmedia "github.com/wesen/2026-04-09--screencast-studio/pkg/media/gst"
)

type frameStore struct {
	mu          sync.RWMutex
	latest      []byte
	seq         uint64
	lastFrameAt time.Time
	framesSeen  uint64
	bytesSeen   uint64
}

func (s *frameStore) store(frame []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latest = append([]byte(nil), frame...)
	s.seq++
	s.lastFrameAt = time.Now()
	s.framesSeen++
	s.bytesSeen += uint64(len(frame))
}

func (s *frameStore) latestFrame() ([]byte, uint64, time.Time, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.latest) == 0 {
		return nil, 0, time.Time{}, false
	}
	return append([]byte(nil), s.latest...), s.seq, s.lastFrameAt, true
}

type mjpegStats struct {
	mu           sync.Mutex
	clients      int
	streams      int
	framesServed uint64
	bytesServed  uint64
}

func (s *mjpegStats) startClient() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients++
	s.streams++
}

func (s *mjpegStats) endClient() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.clients > 0 {
		s.clients--
	}
}

func (s *mjpegStats) addFrame(bytes int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.framesServed++
	s.bytesServed += uint64(bytes)
}

type harnessSummary struct {
	Mode               string    `json:"mode"`
	StartedAt          time.Time `json:"started_at"`
	FinishedAt         time.Time `json:"finished_at"`
	SourceType         string    `json:"source_type"`
	OutputPath         string    `json:"output_path,omitempty"`
	HTTPAddr           string    `json:"http_addr,omitempty"`
	MJPEGURL           string    `json:"mjpeg_url,omitempty"`
	ClientEnabled      bool      `json:"client_enabled"`
	ClientSeconds      int       `json:"client_seconds,omitempty"`
	WarmupSeconds      int       `json:"warmup_seconds,omitempty"`
	RecordSeconds      int       `json:"record_seconds,omitempty"`
	PreviewFramesSeen  uint64    `json:"preview_frames_seen,omitempty"`
	PreviewBytesSeen   uint64    `json:"preview_bytes_seen,omitempty"`
	MJPEGClients       int       `json:"mjpeg_clients,omitempty"`
	MJPEGStreams       int       `json:"mjpeg_streams,omitempty"`
	MJPEGFramesServed  uint64    `json:"mjpeg_frames_served,omitempty"`
	MJPEGBytesServed   uint64    `json:"mjpeg_bytes_served,omitempty"`
	PreviewLastFrameAt string    `json:"preview_last_frame_at,omitempty"`
	RecordingState     string    `json:"recording_state,omitempty"`
	RecordingReason    string    `json:"recording_reason,omitempty"`
	Error              string    `json:"error,omitempty"`
}

func main() {
	mode := flag.String("mode", "harness", "harness or mjpeg-client")
	httpAddr := flag.String("http-addr", "127.0.0.1:7791", "http listen address for harness mode")
	mjpegURL := flag.String("mjpeg-url", "", "mjpeg url for client mode")
	clientSeconds := flag.Int("client-seconds", 12, "seconds for client mode to drain mjpeg")
	spawnClient := flag.Bool("spawn-client", true, "spawn a separate child client process in harness mode")
	warmupSeconds := flag.Int("warmup-seconds", 2, "seconds to keep preview+mjpeg active before starting recording")
	recordSeconds := flag.Int("record-seconds", 8, "seconds to record after warmup")
	outputPath := flag.String("output-path", filepath.Join(os.TempDir(), "scs-standalone-shared-preview-record.mp4"), "recording output path")
	sourceType := flag.String("source-type", "display", "display|region|window|camera")
	sourceID := flag.String("source-id", "standalone-source-1", "source id")
	sourceName := flag.String("source-name", "Standalone Source", "source name")
	displayName := flag.String("display-name", strings.TrimSpace(os.Getenv("DISPLAY")), "x11 display name")
	windowID := flag.String("window-id", "", "x11 window id for window source")
	device := flag.String("device", "", "camera device path for camera source")
	rect := flag.String("rect", "", "region/window rectangle as x,y,w,h")
	fps := flag.Int("fps", 24, "capture fps")
	previewSize := flag.String("preview-size", "960x640", "preview target size hint for camera sources")
	container := flag.String("container", "mp4", "recording container")
	quality := flag.Int("quality", 70, "recording quality")
	child := flag.Bool("child", false, "internal child marker")
	flag.Parse()

	if *mode == "mjpeg-client" {
		if err := runMJPEGClient(*mjpegURL, time.Duration(*clientSeconds)*time.Second); err != nil {
			fmt.Fprintf(os.Stderr, "mjpeg client failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	summary := harnessSummary{
		Mode:          *mode,
		StartedAt:     time.Now(),
		SourceType:    *sourceType,
		OutputPath:    *outputPath,
		HTTPAddr:      *httpAddr,
		ClientEnabled: *spawnClient,
		ClientSeconds: *clientSeconds,
		WarmupSeconds: *warmupSeconds,
		RecordSeconds: *recordSeconds,
	}
	defer func() {
		summary.FinishedAt = time.Now()
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(summary)
	}()

	source, err := buildSource(*sourceType, *sourceID, *sourceName, *displayName, *windowID, *device, *rect, *fps, *previewSize, *container, *quality)
	if err != nil {
		summary.Error = err.Error()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := &frameStore{}
	previewRuntime := gstmedia.NewPreviewRuntime()
	previewSession, err := previewRuntime.StartPreview(ctx, source, media.PreviewOptions{
		OnFrame: func(frame []byte) {
			store.store(frame)
		},
		OnLog: func(stream, message string) {
			log.Debug().Str("stream", stream).Msg(message)
		},
	})
	if err != nil {
		summary.Error = fmt.Sprintf("start preview: %v", err)
		return
	}
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer stopCancel()
		_ = previewSession.Stop(stopCtx)
	}()

	if err := waitForFirstFrame(store, 10*time.Second); err != nil {
		summary.Error = fmt.Sprintf("wait for first frame: %v", err)
		return
	}

	stats := &mjpegStats{}
	server := &http.Server{Addr: *httpAddr, Handler: buildMux(store, stats)}
	listener, err := net.Listen("tcp", *httpAddr)
	if err != nil {
		summary.Error = fmt.Sprintf("listen http: %v", err)
		return
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "http serve failed: %v\n", serveErr)
		}
	}()

	summary.MJPEGURL = "http://" + listener.Addr().String() + "/mjpeg"

	var clientCmd *exec.Cmd
	if *spawnClient && !*child {
		exe, exeErr := os.Executable()
		if exeErr != nil {
			summary.Error = fmt.Sprintf("resolve executable for child client: %v", exeErr)
			return
		}
		clientCmd = exec.Command(exe,
			"--mode=mjpeg-client",
			"--mjpeg-url", summary.MJPEGURL,
			"--client-seconds", strconv.Itoa(*warmupSeconds+*recordSeconds+2),
		)
		clientCmd.Stdout = os.Stderr
		clientCmd.Stderr = os.Stderr
		if err := clientCmd.Start(); err != nil {
			summary.Error = fmt.Sprintf("start child mjpeg client: %v", err)
			return
		}
		defer func() {
			_ = clientCmd.Wait()
		}()
	}

	if *warmupSeconds > 0 {
		time.Sleep(time.Duration(*warmupSeconds) * time.Second)
	}

	plan := &dsl.CompiledPlan{
		SessionID: "standalone-shared-preview-record",
		VideoJobs: []dsl.VideoJob{{Source: source, OutputPath: *outputPath}},
		Outputs: []dsl.PlannedOutput{{Kind: "video", SourceID: source.ID, Name: source.Name, Path: *outputPath}},
	}
	recordingRuntime := gstmedia.NewRecordingRuntime()
	recordingSession, err := recordingRuntime.StartRecording(ctx, plan, media.RecordingOptions{})
	if err != nil {
		summary.Error = fmt.Sprintf("start recording: %v", err)
		return
	}

	time.Sleep(time.Duration(*recordSeconds) * time.Second)
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()
	if err := recordingSession.Stop(stopCtx); err != nil {
		summary.Error = fmt.Sprintf("stop recording: %v", err)
		return
	}
	result, err := recordingSession.Wait()
	if err != nil {
		summary.Error = fmt.Sprintf("wait recording: %v", err)
	}
	if result != nil {
		summary.RecordingState = string(result.State)
		summary.RecordingReason = result.Reason
	}

	store.mu.RLock()
	summary.PreviewFramesSeen = store.framesSeen
	summary.PreviewBytesSeen = store.bytesSeen
	if !store.lastFrameAt.IsZero() {
		summary.PreviewLastFrameAt = store.lastFrameAt.Format(time.RFC3339Nano)
	}
	store.mu.RUnlock()

	stats.mu.Lock()
	summary.MJPEGClients = stats.clients
	summary.MJPEGStreams = stats.streams
	summary.MJPEGFramesServed = stats.framesServed
	summary.MJPEGBytesServed = stats.bytesServed
	stats.mu.Unlock()
}

func runMJPEGClient(url string, duration time.Duration) error {
	if strings.TrimSpace(url) == "" {
		return errors.New("mjpeg url is required")
	}
	if duration <= 0 {
		duration = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("unexpected status %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(req.Context().Err(), context.DeadlineExceeded) {
		return err
	}
	return nil
}

func waitForFirstFrame(store *frameStore, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, _, _, ok := store.latestFrame(); ok {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return errors.New("timed out waiting for first preview frame")
}

func buildMux(store *frameStore, stats *mjpegStats) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	mux.HandleFunc("/mjpeg", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}
		stats.startClient()
		defer stats.endClient()

		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		var lastSeq uint64
		for {
			frame, seq, _, ok := store.latestFrame()
			if ok && len(frame) > 0 && seq != lastSeq {
				lastSeq = seq
				written := 0
				n, err := w.Write([]byte("--frame\r\nContent-Type: image/jpeg\r\nContent-Length: "))
				written += n
				if err != nil {
					return
				}
				n, err = w.Write([]byte(strconv.Itoa(len(frame))))
				written += n
				if err != nil {
					return
				}
				n, err = w.Write([]byte("\r\n\r\n"))
				written += n
				if err != nil {
					return
				}
				copied, err := io.Copy(w, bytes.NewReader(frame))
				written += int(copied)
				if err != nil {
					return
				}
				n, err = w.Write([]byte("\r\n"))
				written += n
				if err != nil {
					return
				}
				stats.addFrame(written)
				flusher.Flush()
			}
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
			}
		}
	})
	return mux
}

func buildSource(sourceType, sourceID, sourceName, displayName, windowID, device, rect string, fps int, previewSize, container string, quality int) (dsl.EffectiveVideoSource, error) {
	source := dsl.EffectiveVideoSource{
		ID:      sourceID,
		Name:    sourceName,
		Type:    strings.TrimSpace(sourceType),
		Enabled: true,
		Target: dsl.VideoTarget{
			Display: strings.TrimSpace(displayName),
			WindowID: strings.TrimSpace(windowID),
			Device: strings.TrimSpace(device),
		},
		Capture: dsl.VideoCaptureSettings{
			FPS:     fps,
			Size:    previewSize,
			Cursor:  boolPtr(true),
		},
		Output: dsl.VideoOutputSettings{
			Container: strings.TrimSpace(container),
			VideoCodec: "h264",
			Quality: quality,
		},
	}
	if strings.TrimSpace(rect) != "" {
		r, err := parseRect(rect)
		if err != nil {
			return source, err
		}
		source.Target.Rect = r
	}
	switch source.Type {
	case "display":
		if source.Target.Display == "" {
			return source, errors.New("display source requires --display-name or DISPLAY")
		}
	case "region":
		if source.Target.Display == "" {
			return source, errors.New("region source requires --display-name or DISPLAY")
		}
		if source.Target.Rect == nil {
			return source, errors.New("region source requires --rect x,y,w,h")
		}
	case "window":
		if source.Target.Display == "" {
			return source, errors.New("window source requires --display-name or DISPLAY")
		}
		if source.Target.WindowID == "" {
			return source, errors.New("window source requires --window-id")
		}
		if source.Target.Rect == nil {
			return source, errors.New("window source currently requires --rect x,y,w,h for reproducible geometry")
		}
	case "camera":
		if source.Target.Device == "" {
			return source, errors.New("camera source requires --device")
		}
	default:
		return source, fmt.Errorf("unsupported source type %q", source.Type)
	}
	return source, nil
}

func parseRect(value string) (*dsl.Rect, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid rect %q, want x,y,w,h", value)
	}
	vals := make([]int, 0, 4)
	for _, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("parse rect component %q: %w", part, err)
		}
		vals = append(vals, n)
	}
	if vals[2] <= 0 || vals[3] <= 0 {
		return nil, fmt.Errorf("invalid rect %q, width and height must be > 0", value)
	}
	return &dsl.Rect{X: vals[0], Y: vals[1], W: vals[2], H: vals[3]}, nil
}

func boolPtr(v bool) *bool { return &v }
