package main

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

//go:embed web/* examples/*
var embeddedFiles embed.FS

const schemaVersion = "recorder.config/v1"

type Config struct {
	Schema                string            `json:"schema"`
	SessionID             string            `json:"session_id"`
	DestinationTemplates  map[string]string `json:"destination_templates"`
	ScreenCaptureDefaults VideoDefaults     `json:"screen_capture_defaults"`
	CameraCaptureDefaults VideoDefaults     `json:"camera_capture_defaults"`
	AudioDefaults         AudioDefaults     `json:"audio_defaults"`
	AudioMix              AudioMixConfig    `json:"audio_mix"`
	VideoSources          []VideoSource     `json:"video_sources"`
	AudioSources          []AudioSource     `json:"audio_sources"`
}

type VideoDefaults struct {
	Capture VideoCaptureSettings `json:"capture"`
	Output  VideoOutputSettings  `json:"output"`
}

type VideoCaptureSettings struct {
	FPS          int    `json:"fps"`
	Cursor       *bool  `json:"cursor"`
	FollowResize *bool  `json:"follow_resize"`
	Mirror       *bool  `json:"mirror"`
	Size         string `json:"size"`
}

type VideoOutputSettings struct {
	Container  string `json:"container"`
	VideoCodec string `json:"video_codec"`
	Quality    int    `json:"quality"`
}

type VideoSource struct {
	ID                  string        `json:"id"`
	Name                string        `json:"name"`
	Type                string        `json:"type"`
	Enabled             *bool         `json:"enabled"`
	Target              VideoTarget   `json:"target"`
	Settings            VideoDefaults `json:"settings"`
	DestinationTemplate string        `json:"destination_template"`
}

type VideoTarget struct {
	Display  string `json:"display"`
	WindowID string `json:"window_id"`
	Device   string `json:"device"`
	Rect     *Rect  `json:"rect"`
}

type Rect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type AudioDefaults struct {
	Output AudioOutputSettings `json:"output"`
}

type AudioOutputSettings struct {
	Codec        string `json:"codec"`
	SampleRateHz int    `json:"sample_rate_hz"`
	Channels     int    `json:"channels"`
}

type AudioMixConfig struct {
	DestinationTemplate string `json:"destination_template"`
}

type AudioSource struct {
	ID       string              `json:"id"`
	Name     string              `json:"name"`
	Device   string              `json:"device"`
	Enabled  *bool               `json:"enabled"`
	Settings AudioSourceSettings `json:"settings"`
}

type AudioSourceSettings struct {
	Gain      float64 `json:"gain"`
	NoiseGate bool    `json:"noise_gate"`
	Denoise   bool    `json:"denoise"`
}

type EffectiveConfig struct {
	Schema               string
	SessionID            string
	DestinationTemplates map[string]string
	AudioMixTemplate     string
	AudioOutput          AudioOutputSettings
	VideoSources         []EffectiveVideoSource
	AudioSources         []EffectiveAudioSource
	RawDSL               string
	Warnings             []string
}

type EffectiveVideoSource struct {
	ID                  string
	Name                string
	Type                string
	Enabled             bool
	Target              VideoTarget
	Capture             VideoCaptureSettings
	Output              VideoOutputSettings
	DestinationTemplate string
}

type EffectiveAudioSource struct {
	ID       string
	Name     string
	Enabled  bool
	Device   string
	Settings AudioSourceSettings
}

type PreviewInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	PreviewURL string `json:"preview_url"`
}

type OutputFile struct {
	Kind     string `json:"kind"`
	SourceID string `json:"source_id,omitempty"`
	Name     string `json:"name"`
	Path     string `json:"path"`
}

type ManagedProcess struct {
	Label      string
	OutputPath string
	Cmd        *exec.Cmd
	Stdin      io.WriteCloser
	done       chan error
}

type CaptureSession struct {
	StartedAt time.Time
	Processes []*ManagedProcess
	Outputs   []OutputFile
	waitDone  chan struct{}
}

type App struct {
	mu      sync.RWMutex
	cfg     *EffectiveConfig
	capture *CaptureSession
	logs    []string
}

func main() {
	app := &App{}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/example", app.handleExample)
	mux.HandleFunc("/api/config/apply", app.handleApplyConfig)
	mux.HandleFunc("/api/config", app.handleGetConfig)
	mux.HandleFunc("/api/state", app.handleState)
	mux.HandleFunc("/api/capture/start", app.handleStartCapture)
	mux.HandleFunc("/api/capture/stop", app.handleStopCapture)
	mux.HandleFunc("/api/preview/", app.handlePreview)

	webFS, err := fs.Sub(embeddedFiles, "web")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", http.FileServer(http.FS(webFS)))

	addr := getenv("ADDR", ":8080")
	srv := &http.Server{
		Addr:              addr,
		Handler:           withLogging(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	app.logf("server starting on %s", addr)
	log.Printf("listening on http://localhost%s", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(started).Round(time.Millisecond))
	})
}

func (a *App) handleExample(w http.ResponseWriter, r *http.Request) {
	b, err := embeddedFiles.ReadFile("examples/example.yaml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(b)
}

func (a *App) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	cfg := a.cfg
	a.mu.RUnlock()
	if cfg == nil {
		writeJSON(w, http.StatusOK, map[string]any{"configured": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"configured":    true,
		"session_id":    cfg.SessionID,
		"dsl":           cfg.RawDSL,
		"video_sources": buildPreviewInfos(cfg),
		"warnings":      cfg.Warnings,
	})
}

func (a *App) handleApplyConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	a.mu.RLock()
	recording := a.capture != nil
	a.mu.RUnlock()
	if recording {
		http.Error(w, "cannot apply config while recording", http.StatusConflict)
		return
	}

	b, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(bytes.TrimSpace(b)) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	cfg, err := parseAndNormalizeConfig(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	a.mu.Lock()
	a.cfg = cfg
	a.mu.Unlock()
	a.logf("applied config %s with %d video source(s) and %d audio source(s)", cfg.SessionID, len(cfg.VideoSources), len(cfg.AudioSources))

	resp := map[string]any{
		"ok":            true,
		"session_id":    cfg.SessionID,
		"video_sources": buildPreviewInfos(cfg),
		"warnings":      cfg.Warnings,
	}
	audio := make([]map[string]string, 0, len(cfg.AudioSources))
	for _, src := range cfg.AudioSources {
		if !src.Enabled {
			continue
		}
		audio = append(audio, map[string]string{"id": src.ID, "name": src.Name, "device": src.Device})
	}
	resp["audio_sources"] = audio
	writeJSON(w, http.StatusOK, resp)
}

func (a *App) handleState(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	cfg := a.cfg
	capture := a.capture
	logsCopy := append([]string(nil), a.logs...)
	a.mu.RUnlock()

	resp := map[string]any{
		"configured": cfg != nil,
		"recording":  capture != nil,
		"logs":       logsCopy,
	}
	if cfg != nil {
		resp["session_id"] = cfg.SessionID
		resp["video_sources"] = buildPreviewInfos(cfg)
		resp["warnings"] = cfg.Warnings
	}
	if capture != nil {
		resp["started_at"] = capture.StartedAt.Format(time.RFC3339)
		resp["outputs"] = capture.Outputs
	}
	writeJSON(w, http.StatusOK, resp)
}

func (a *App) handleStartCapture(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	a.mu.RLock()
	cfg := a.cfg
	recording := a.capture != nil
	a.mu.RUnlock()
	if cfg == nil {
		http.Error(w, "no config loaded", http.StatusBadRequest)
		return
	}
	if recording {
		http.Error(w, "capture already running", http.StatusConflict)
		return
	}
	sess, err := startCaptureSession(cfg, a.logf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.mu.Lock()
	if a.capture != nil {
		a.mu.Unlock()
		_ = sess.Stop(5 * time.Second)
		http.Error(w, "capture already running", http.StatusConflict)
		return
	}
	a.capture = sess
	a.mu.Unlock()
	a.logf("capture started")
	go a.watchCaptureSession(sess)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"recording":  true,
		"started_at": sess.StartedAt.Format(time.RFC3339),
		"outputs":    sess.Outputs,
	})
}

func (a *App) handleStopCapture(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	a.mu.Lock()
	sess := a.capture
	a.capture = nil
	a.mu.Unlock()
	if sess == nil {
		http.Error(w, "capture is not running", http.StatusConflict)
		return
	}
	if err := sess.Stop(8 * time.Second); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.logf("capture stopped")
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "recording": false})
}

func (a *App) handlePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/preview/")
	if id == "" {
		http.Error(w, "missing source id", http.StatusBadRequest)
		return
	}

	a.mu.RLock()
	cfg := a.cfg
	a.mu.RUnlock()
	if cfg == nil {
		http.Error(w, "no config loaded", http.StatusBadRequest)
		return
	}

	var src *EffectiveVideoSource
	for i := range cfg.VideoSources {
		if cfg.VideoSources[i].ID == id {
			src = &cfg.VideoSources[i]
			break
		}
	}
	if src == nil || !src.Enabled {
		http.Error(w, "unknown source", http.StatusNotFound)
		return
	}

	args, err := buildPreviewArgs(*src)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := exec.CommandContext(r.Context(), "ffmpeg", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := cmd.Start(); err != nil {
		http.Error(w, fmt.Sprintf("failed to start ffmpeg: %v", err), http.StatusInternalServerError)
		return
	}
	a.logf("preview started for %s", src.Name)
	go drainScanner(stderr, func(line string) {
		if strings.TrimSpace(line) != "" {
			a.logf("preview %s: %s", src.ID, line)
		}
	})

	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
		a.logf("preview stopped for %s", src.Name)
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")

	reader := bufio.NewReader(stdout)
	for {
		frame, err := readJPEGFrame(reader)
		if err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, context.Canceled) {
				a.logf("preview read error for %s: %v", src.ID, err)
			}
			return
		}
		if _, err := fmt.Fprintf(w, "--frame\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(frame)); err != nil {
			return
		}
		if _, err := w.Write(frame); err != nil {
			return
		}
		if _, err := w.Write([]byte("\r\n")); err != nil {
			return
		}
		flusher.Flush()
	}
}

func (a *App) watchCaptureSession(sess *CaptureSession) {
	<-sess.waitDone
	a.mu.Lock()
	if a.capture == sess {
		a.capture = nil
		a.logf("capture session ended")
	}
	a.mu.Unlock()
}

func (a *App) logf(format string, args ...any) {
	line := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
	log.Println(line)
	a.mu.Lock()
	defer a.mu.Unlock()
	a.logs = append(a.logs, line)
	if len(a.logs) > 200 {
		a.logs = append([]string(nil), a.logs[len(a.logs)-200:]...)
	}
}

func parseAndNormalizeConfig(body []byte) (*EffectiveConfig, error) {
	var cfg Config
	if err := decodeDSL(body, &cfg); err != nil {
		return nil, fmt.Errorf("invalid DSL: %w", err)
	}
	if cfg.Schema == "" {
		cfg.Schema = schemaVersion
	}
	if cfg.Schema != schemaVersion {
		return nil, fmt.Errorf("unsupported schema %q", cfg.Schema)
	}
	if len(cfg.DestinationTemplates) == 0 {
		return nil, errors.New("destination_templates is required")
	}
	if len(cfg.VideoSources) == 0 && len(cfg.AudioSources) == 0 {
		return nil, errors.New("at least one video or audio source is required")
	}

	eff := &EffectiveConfig{
		Schema:               cfg.Schema,
		SessionID:            cfg.SessionID,
		DestinationTemplates: map[string]string{},
		RawDSL:               string(body),
	}
	if eff.SessionID == "" {
		eff.SessionID = "session-" + time.Now().Format("20060102-150405")
	}
	for k, v := range cfg.DestinationTemplates {
		eff.DestinationTemplates[k] = v
	}
	eff.AudioOutput = cfg.AudioDefaults.Output
	if eff.AudioOutput.Codec == "" {
		eff.AudioOutput.Codec = "pcm_s16le"
	}
	if eff.AudioOutput.SampleRateHz <= 0 {
		eff.AudioOutput.SampleRateHz = 48000
	}
	if eff.AudioOutput.Channels <= 0 {
		eff.AudioOutput.Channels = 2
	}
	if cfg.AudioMix.DestinationTemplate != "" {
		if _, ok := eff.DestinationTemplates[cfg.AudioMix.DestinationTemplate]; !ok {
			return nil, fmt.Errorf("audio_mix.destination_template %q not found", cfg.AudioMix.DestinationTemplate)
		}
		eff.AudioMixTemplate = cfg.AudioMix.DestinationTemplate
	} else if _, ok := eff.DestinationTemplates["audio_mix"]; ok {
		eff.AudioMixTemplate = "audio_mix"
	}

	seen := map[string]struct{}{}
	for i, src := range cfg.VideoSources {
		v, warnings, err := normalizeVideoSource(src, i, cfg.ScreenCaptureDefaults, cfg.CameraCaptureDefaults, eff.DestinationTemplates)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[v.ID]; ok {
			return nil, fmt.Errorf("duplicate source id %q", v.ID)
		}
		seen[v.ID] = struct{}{}
		eff.VideoSources = append(eff.VideoSources, v)
		eff.Warnings = append(eff.Warnings, warnings...)
	}
	for i, src := range cfg.AudioSources {
		v, warnings, err := normalizeAudioSource(src, i)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[v.ID]; ok {
			return nil, fmt.Errorf("duplicate source id %q", v.ID)
		}
		seen[v.ID] = struct{}{}
		eff.AudioSources = append(eff.AudioSources, v)
		eff.Warnings = append(eff.Warnings, warnings...)
	}
	sort.Strings(eff.Warnings)
	return eff, nil
}

func decodeDSL(body []byte, out any) error {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return errors.New("empty config")
	}
	if json.Valid(trimmed) {
		return json.Unmarshal(trimmed, out)
	}
	script := `import json, sys
import yaml
obj = yaml.safe_load(sys.stdin.read())
json.dump(obj, sys.stdout)`
	cmd := exec.Command("python3", "-c", script)
	cmd.Stdin = bytes.NewReader(body)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return errors.New(msg)
	}
	return json.Unmarshal(stdout.Bytes(), out)
}

func normalizeVideoSource(src VideoSource, idx int, screenDefaults, cameraDefaults VideoDefaults, templates map[string]string) (EffectiveVideoSource, []string, error) {
	warnings := []string{}
	eff := EffectiveVideoSource{
		ID:      src.ID,
		Name:    src.Name,
		Type:    strings.ToLower(strings.TrimSpace(src.Type)),
		Enabled: boolValue(src.Enabled, true),
		Target:  src.Target,
	}
	if eff.ID == "" {
		base := src.Name
		if base == "" {
			base = fmt.Sprintf("source-%d", idx+1)
		}
		eff.ID = slugify(base)
	}
	if eff.Name == "" {
		eff.Name = strings.ReplaceAll(eff.ID, "_", " ")
	}

	var defaults VideoDefaults
	switch eff.Type {
	case "display", "window", "region":
		defaults = screenDefaults
	case "camera":
		defaults = cameraDefaults
	default:
		return eff, nil, fmt.Errorf("video_sources[%d]: unsupported type %q", idx, src.Type)
	}
	eff.Capture = defaults.Capture
	mergeVideoCapture(&eff.Capture, src.Settings.Capture)
	eff.Output = defaults.Output
	mergeVideoOutput(&eff.Output, src.Settings.Output)
	if eff.Capture.FPS <= 0 {
		eff.Capture.FPS = 24
	}
	if eff.Output.Container == "" {
		eff.Output.Container = "mov"
	}
	if eff.Output.VideoCodec == "" {
		eff.Output.VideoCodec = "h264"
	}
	if eff.Output.Quality <= 0 {
		eff.Output.Quality = 75
	}
	eff.DestinationTemplate = src.DestinationTemplate
	if eff.DestinationTemplate == "" {
		return eff, nil, fmt.Errorf("video_sources[%s]: destination_template is required", eff.ID)
	}
	if _, ok := templates[eff.DestinationTemplate]; !ok {
		return eff, nil, fmt.Errorf("video_sources[%s]: destination_template %q not found", eff.ID, eff.DestinationTemplate)
	}

	if eff.Capture.FollowResize != nil && *eff.Capture.FollowResize {
		warnings = append(warnings, fmt.Sprintf("video source %s: follow_resize is not implemented in this minimal runtime", eff.ID))
	}

	switch eff.Type {
	case "display":
		if eff.Target.Display == "" {
			eff.Target.Display = defaultDisplay()
		}
	case "region":
		if eff.Target.Display == "" {
			eff.Target.Display = defaultDisplay()
		}
		if eff.Target.Rect == nil || eff.Target.Rect.W <= 0 || eff.Target.Rect.H <= 0 {
			return eff, nil, fmt.Errorf("video_sources[%s]: region sources require target.rect with positive w/h", eff.ID)
		}
	case "window":
		if eff.Target.Display == "" {
			eff.Target.Display = defaultDisplay()
		}
		if strings.TrimSpace(eff.Target.WindowID) == "" {
			return eff, nil, fmt.Errorf("video_sources[%s]: window sources require target.window_id", eff.ID)
		}
	case "camera":
		if strings.TrimSpace(eff.Target.Device) == "" {
			return eff, nil, fmt.Errorf("video_sources[%s]: camera sources require target.device", eff.ID)
		}
	}
	return eff, warnings, nil
}

func normalizeAudioSource(src AudioSource, idx int) (EffectiveAudioSource, []string, error) {
	warnings := []string{}
	eff := EffectiveAudioSource{
		ID:       src.ID,
		Name:     src.Name,
		Enabled:  boolValue(src.Enabled, true),
		Device:   strings.TrimSpace(src.Device),
		Settings: src.Settings,
	}
	if eff.ID == "" {
		base := src.Name
		if base == "" {
			base = fmt.Sprintf("audio-%d", idx+1)
		}
		eff.ID = slugify(base)
	}
	if eff.Name == "" {
		eff.Name = strings.ReplaceAll(eff.ID, "_", " ")
	}
	if eff.Device == "" {
		return eff, nil, fmt.Errorf("audio_sources[%s]: device is required", eff.ID)
	}
	if eff.Settings.Gain == 0 {
		eff.Settings.Gain = 1.0
	}
	if eff.Settings.NoiseGate {
		warnings = append(warnings, fmt.Sprintf("audio source %s: noise_gate is not implemented in this minimal runtime", eff.ID))
	}
	if eff.Settings.Denoise {
		warnings = append(warnings, fmt.Sprintf("audio source %s: denoise is not implemented in this minimal runtime", eff.ID))
	}
	return eff, warnings, nil
}

func mergeVideoCapture(dst *VideoCaptureSettings, src VideoCaptureSettings) {
	if src.FPS > 0 {
		dst.FPS = src.FPS
	}
	if src.Cursor != nil {
		dst.Cursor = src.Cursor
	}
	if src.FollowResize != nil {
		dst.FollowResize = src.FollowResize
	}
	if src.Mirror != nil {
		dst.Mirror = src.Mirror
	}
	if src.Size != "" {
		dst.Size = src.Size
	}
}

func mergeVideoOutput(dst *VideoOutputSettings, src VideoOutputSettings) {
	if src.Container != "" {
		dst.Container = src.Container
	}
	if src.VideoCodec != "" {
		dst.VideoCodec = src.VideoCodec
	}
	if src.Quality > 0 {
		dst.Quality = src.Quality
	}
}

func buildPreviewInfos(cfg *EffectiveConfig) []PreviewInfo {
	out := make([]PreviewInfo, 0, len(cfg.VideoSources))
	for _, src := range cfg.VideoSources {
		if !src.Enabled {
			continue
		}
		out = append(out, PreviewInfo{ID: src.ID, Name: src.Name, Type: src.Type, PreviewURL: "/api/preview/" + src.ID})
	}
	return out
}

func startCaptureSession(cfg *EffectiveConfig, logger func(string, ...any)) (*CaptureSession, error) {
	sess := &CaptureSession{StartedAt: time.Now(), waitDone: make(chan struct{})}
	cleanup := func() { _ = sess.Stop(5 * time.Second) }

	startedAt := sess.StartedAt
	for _, src := range cfg.VideoSources {
		if !src.Enabled {
			continue
		}
		ext := videoExtension(src.Output.Container)
		outPath, err := renderDestination(cfg.DestinationTemplates[src.DestinationTemplate], renderVars{
			SessionID:  cfg.SessionID,
			SourceID:   src.ID,
			SourceName: src.Name,
			SourceType: src.Type,
			Ext:        ext,
			Now:        startedAt,
		})
		if err != nil {
			cleanup()
			return nil, err
		}
		args, err := buildRecordVideoArgs(src, outPath)
		if err != nil {
			cleanup()
			return nil, err
		}
		proc, err := startManagedProcess(src.Name, args, outPath, logger)
		if err != nil {
			cleanup()
			return nil, err
		}
		sess.Processes = append(sess.Processes, proc)
		sess.Outputs = append(sess.Outputs, OutputFile{Kind: "video", SourceID: src.ID, Name: src.Name, Path: outPath})
	}

	enabledAudio := make([]EffectiveAudioSource, 0, len(cfg.AudioSources))
	for _, src := range cfg.AudioSources {
		if src.Enabled {
			enabledAudio = append(enabledAudio, src)
		}
	}
	if len(enabledAudio) > 0 {
		templateName := cfg.AudioMixTemplate
		if templateName == "" {
			if _, ok := cfg.DestinationTemplates["audio_mix"]; ok {
				templateName = "audio_mix"
			} else {
				for k := range cfg.DestinationTemplates {
					templateName = k
					break
				}
			}
		}
		if templateName == "" {
			cleanup()
			return nil, errors.New("no destination template available for mixed audio")
		}
		ext := audioExtension(cfg.AudioOutput.Codec)
		outPath, err := renderDestination(cfg.DestinationTemplates[templateName], renderVars{
			SessionID:  cfg.SessionID,
			SourceID:   "audio_mix",
			SourceName: "audio-mix",
			SourceType: "audio",
			Ext:        ext,
			Now:        startedAt,
		})
		if err != nil {
			cleanup()
			return nil, err
		}
		args, err := buildRecordAudioArgs(enabledAudio, cfg.AudioOutput, outPath)
		if err != nil {
			cleanup()
			return nil, err
		}
		proc, err := startManagedProcess("audio-mix", args, outPath, logger)
		if err != nil {
			cleanup()
			return nil, err
		}
		sess.Processes = append(sess.Processes, proc)
		sess.Outputs = append(sess.Outputs, OutputFile{Kind: "audio", Name: "audio-mix", Path: outPath})
	}

	if len(sess.Processes) == 0 {
		return nil, errors.New("no enabled sources to capture")
	}
	go func() {
		for _, proc := range sess.Processes {
			<-proc.done
		}
		close(sess.waitDone)
	}()
	return sess, nil
}

func (s *CaptureSession) Stop(timeout time.Duration) error {
	if s == nil {
		return nil
	}
	errs := []string{}
	for _, proc := range s.Processes {
		if err := proc.Stop(timeout); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", proc.Label, err))
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func startManagedProcess(label string, args []string, outputPath string, logger func(string, ...any)) (*ManagedProcess, error) {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return nil, err
	}
	cmd := exec.Command("ffmpeg", args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	proc := &ManagedProcess{Label: label, OutputPath: outputPath, Cmd: cmd, Stdin: stdin, done: make(chan error, 1)}
	go drainScanner(stderr, func(line string) {
		if strings.TrimSpace(line) != "" {
			logger("%s: %s", label, line)
		}
	})
	go func() {
		err := cmd.Wait()
		proc.done <- err
		close(proc.done)
	}()
	return proc, nil
}

func (p *ManagedProcess) Stop(timeout time.Duration) error {
	if p == nil {
		return nil
	}
	if p.Stdin != nil {
		_, _ = io.WriteString(p.Stdin, "q\n")
		_ = p.Stdin.Close()
	}
	select {
	case err, ok := <-p.done:
		if !ok || err == nil {
			return nil
		}
		return err
	case <-time.After(timeout):
		if p.Cmd != nil && p.Cmd.Process != nil {
			_ = p.Cmd.Process.Kill()
		}
		err, ok := <-p.done
		if !ok || err == nil {
			return nil
		}
		return err
	}
}

func buildPreviewArgs(src EffectiveVideoSource) ([]string, error) {
	args := []string{"-hide_banner", "-loglevel", "error", "-nostdin"}
	var err error
	args, err = appendVideoInputArgs(args, src)
	if err != nil {
		return nil, err
	}
	filters := []string{"fps=5", "scale=640:-1:force_original_aspect_ratio=decrease"}
	if src.Type == "camera" && boolValue(src.Capture.Mirror, false) {
		filters = append(filters, "hflip")
	}
	args = append(args, "-an", "-vf", strings.Join(filters, ","), "-q:v", "7", "-f", "image2pipe", "-vcodec", "mjpeg", "pipe:1")
	return args, nil
}

func buildRecordVideoArgs(src EffectiveVideoSource, outputPath string) ([]string, error) {
	args := []string{"-hide_banner", "-loglevel", "error", "-y"}
	var err error
	args, err = appendVideoInputArgs(args, src)
	if err != nil {
		return nil, err
	}
	filters := []string{}
	if src.Type == "camera" && boolValue(src.Capture.Mirror, false) {
		filters = append(filters, "hflip")
	}
	if len(filters) > 0 {
		args = append(args, "-vf", strings.Join(filters, ","))
	}
	args = append(args, encodeVideoArgs(src.Output)...)
	args = append(args, outputPath)
	return args, nil
}

func buildRecordAudioArgs(sources []EffectiveAudioSource, settings AudioOutputSettings, outputPath string) ([]string, error) {
	if len(sources) == 0 {
		return nil, errors.New("no audio sources")
	}
	args := []string{"-hide_banner", "-loglevel", "error", "-y"}
	for _, src := range sources {
		args = append(args, "-f", "pulse", "-sample_rate", fmt.Sprintf("%d", settings.SampleRateHz), "-channels", fmt.Sprintf("%d", settings.Channels), "-i", src.Device)
	}
	filterParts := make([]string, 0, len(sources)+1)
	mixInputs := make([]string, 0, len(sources))
	for i, src := range sources {
		label := fmt.Sprintf("a%d", i)
		filterParts = append(filterParts, fmt.Sprintf("[%d:a]volume=%s[%s]", i, trimFloat(src.Settings.Gain), label))
		mixInputs = append(mixInputs, fmt.Sprintf("[%s]", label))
	}
	if len(sources) == 1 {
		filterParts = append(filterParts, fmt.Sprintf("%sanull[aout]", mixInputs[0]))
	} else {
		filterParts = append(filterParts, fmt.Sprintf("%samix=inputs=%d:normalize=0[aout]", strings.Join(mixInputs, ""), len(sources)))
	}
	args = append(args, "-filter_complex", strings.Join(filterParts, ";"), "-map", "[aout]", "-ar", fmt.Sprintf("%d", settings.SampleRateHz), "-ac", fmt.Sprintf("%d", settings.Channels))
	switch strings.ToLower(strings.TrimSpace(settings.Codec)) {
	case "", "pcm_s16le", "wav":
		args = append(args, "-c:a", "pcm_s16le")
	case "aac":
		args = append(args, "-c:a", "aac", "-b:a", "192k")
	case "opus":
		args = append(args, "-c:a", "libopus", "-b:a", "160k")
	case "mp3":
		args = append(args, "-c:a", "libmp3lame", "-b:a", "192k")
	default:
		return nil, fmt.Errorf("unsupported audio codec %q", settings.Codec)
	}
	args = append(args, outputPath)
	return args, nil
}

func appendVideoInputArgs(args []string, src EffectiveVideoSource) ([]string, error) {
	switch src.Type {
	case "display":
		display := src.Target.Display
		if display == "" {
			display = defaultDisplay()
		}
		args = append(args, "-f", "x11grab", "-framerate", fmt.Sprintf("%d", src.Capture.FPS), "-draw_mouse", boolFlag(boolValue(src.Capture.Cursor, true)), "-i", display)
		return args, nil
	case "region":
		if src.Target.Rect == nil {
			return nil, errors.New("missing region rect")
		}
		display := src.Target.Display
		if display == "" {
			display = defaultDisplay()
		}
		args = append(args, "-f", "x11grab", "-framerate", fmt.Sprintf("%d", src.Capture.FPS), "-video_size", fmt.Sprintf("%dx%d", src.Target.Rect.W, src.Target.Rect.H), "-draw_mouse", boolFlag(boolValue(src.Capture.Cursor, true)), "-i", fmt.Sprintf("%s+%d,%d", display, src.Target.Rect.X, src.Target.Rect.Y))
		return args, nil
	case "window":
		display := src.Target.Display
		if display == "" {
			display = defaultDisplay()
		}
		args = append(args, "-f", "x11grab", "-framerate", fmt.Sprintf("%d", src.Capture.FPS), "-draw_mouse", boolFlag(boolValue(src.Capture.Cursor, true)), "-window_id", src.Target.WindowID, "-i", display)
		return args, nil
	case "camera":
		args = append(args, "-f", "v4l2")
		if src.Capture.Size != "" {
			args = append(args, "-video_size", src.Capture.Size)
		}
		args = append(args, "-framerate", fmt.Sprintf("%d", src.Capture.FPS), "-i", src.Target.Device)
		return args, nil
	default:
		return nil, fmt.Errorf("unsupported video source type %q", src.Type)
	}
}

func encodeVideoArgs(out VideoOutputSettings) []string {
	switch strings.ToLower(strings.TrimSpace(out.VideoCodec)) {
	case "", "h264", "libx264":
		return []string{"-c:v", "libx264", "-preset", "veryfast", "-crf", fmt.Sprintf("%d", qualityToCRF(out.Quality)), "-pix_fmt", "yuv420p"}
	case "ffv1":
		return []string{"-c:v", "ffv1"}
	case "mpeg4":
		return []string{"-c:v", "mpeg4", "-q:v", fmt.Sprintf("%d", qualityToQScale(out.Quality))}
	default:
		return []string{"-c:v", "libx264", "-preset", "veryfast", "-crf", fmt.Sprintf("%d", qualityToCRF(out.Quality)), "-pix_fmt", "yuv420p"}
	}
}

func qualityToCRF(q int) int {
	if q < 1 {
		q = 75
	}
	if q > 100 {
		q = 100
	}
	return 35 - int(math.Round(float64(q-1)*18.0/99.0))
}

func qualityToQScale(q int) int {
	if q < 1 {
		q = 75
	}
	if q > 100 {
		q = 100
	}
	return 31 - int(math.Round(float64(q-1)*26.0/99.0))
}

type renderVars struct {
	SessionID  string
	SourceID   string
	SourceName string
	SourceType string
	Ext        string
	Now        time.Time
}

func renderDestination(tmpl string, vars renderVars) (string, error) {
	if strings.TrimSpace(tmpl) == "" {
		return "", errors.New("empty destination template")
	}
	path := strings.NewReplacer(
		"{session_id}", safePathSegment(vars.SessionID),
		"{source_id}", safePathSegment(vars.SourceID),
		"{source_name}", safePathSegment(vars.SourceName),
		"{source_type}", safePathSegment(vars.SourceType),
		"{ext}", vars.Ext,
		"{date}", vars.Now.Format("2006-01-02"),
		"{time}", vars.Now.Format("15-04-05"),
		"{timestamp}", vars.Now.Format("20060102-150405"),
	).Replace(tmpl)
	path = expandHome(path)
	if strings.TrimSpace(path) == "" {
		return "", errors.New("destination template rendered to empty path")
	}
	return filepath.Clean(path), nil
}

func videoExtension(container string) string {
	switch strings.ToLower(strings.TrimSpace(container)) {
	case "mov":
		return "mov"
	case "mp4":
		return "mp4"
	case "mkv":
		return "mkv"
	case "avi":
		return "avi"
	case "webm":
		return "webm"
	default:
		return "mov"
	}
}

func audioExtension(codec string) string {
	switch strings.ToLower(strings.TrimSpace(codec)) {
	case "", "pcm_s16le", "wav":
		return "wav"
	case "aac":
		return "m4a"
	case "opus":
		return "ogg"
	case "mp3":
		return "mp3"
	default:
		return "wav"
	}
}

func readJPEGFrame(r *bufio.Reader) ([]byte, error) {
	var buf bytes.Buffer
	var prev byte
	found := false
	for {
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		if prev == 0xff && b == 0xd8 {
			buf.WriteByte(prev)
			buf.WriteByte(b)
			found = true
			break
		}
		prev = b
	}
	if !found {
		return nil, io.EOF
	}
	prev = 0
	for {
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		buf.WriteByte(b)
		if prev == 0xff && b == 0xd9 {
			return buf.Bytes(), nil
		}
		prev = b
	}
}

func drainScanner(r io.Reader, fn func(string)) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1024), 1024*1024)
	for sc.Scan() {
		fn(sc.Text())
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func defaultDisplay() string {
	if d := strings.TrimSpace(os.Getenv("DISPLAY")); d != "" {
		return d
	}
	return ":0.0"
}

func boolValue(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}

func boolFlag(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func trimFloat(v float64) string {
	if v == 0 {
		v = 1.0
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", v), "0"), ".")
}

func expandHome(path string) string {
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func safePathSegment(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, string(filepath.Separator), "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")
	if s == "" {
		return "unnamed"
	}
	return s
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	if s == "" {
		return "source"
	}
	return s
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
