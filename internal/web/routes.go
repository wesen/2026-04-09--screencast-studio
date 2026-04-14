package web

import (
	"encoding/json"
	"net/http"
	"time"
)

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/api/healthz", s.handleHealthz)
	s.mux.HandleFunc("/api/discovery", s.handleDiscovery)
	s.mux.HandleFunc("/api/session", s.handleSessionCurrent)
	s.mux.HandleFunc("/api/setup/normalize", s.handleNormalizeSetup)
	s.mux.HandleFunc("/api/setup/compile", s.handleCompileSetup)
	s.mux.HandleFunc("/api/recordings/start", s.handleRecordingStart)
	s.mux.HandleFunc("/api/recordings/stop", s.handleRecordingStop)
	s.mux.HandleFunc("/api/recordings/current", s.handleSessionCurrent)
	s.mux.HandleFunc("/api/audio/effects", s.handleAudioEffects)
	s.mux.HandleFunc("/api/previews/ensure", s.handlePreviewEnsure)
	s.mux.HandleFunc("/api/previews/release", s.handlePreviewRelease)
	s.mux.HandleFunc("/api/previews", s.handlePreviewList)
	s.mux.HandleFunc("/api/previews/", s.handlePreviewMJPEG)
	s.mux.HandleFunc("/ws", s.handleWebsocket)
	s.mux.HandleFunc("/metrics", s.handleMetrics)
	s.mux.HandleFunc("/", s.handleRoot)
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, statusCode int, code, message string) {
	writeJSON(w, statusCode, apiErrorResponse{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}

func durationFromSeconds(seconds int) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}
