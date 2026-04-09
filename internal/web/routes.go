package web

import (
	"encoding/json"
	"net/http"
)

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/api/healthz", s.handleHealthz)
	s.mux.HandleFunc("/ws", s.handleWebsocketPlaceholder)
	s.mux.HandleFunc("/", s.handleRoot)
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
