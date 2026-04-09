package web

import (
	"encoding/json"
	"io"
	"net/http"
)

func (s *Server) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	snapshot, err := s.app.DiscoverySnapshot(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "discovery_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, mapDiscoveryResponse(snapshot))
}

func (s *Server) handleSessionCurrent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, apiSessionEnvelope{
		Session: mapRecordingSessionResponse(s.recordings.Current()),
	})
}

func (s *Server) handleNormalizeSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	request, ok := decodeDSLRequest(w, r)
	if !ok {
		return
	}

	cfg, err := s.app.NormalizeDSL(r.Context(), []byte(request.DSL))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_dsl", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, apiNormalizeResponse{
		SessionID: cfg.SessionID,
		Warnings:  append([]string(nil), cfg.Warnings...),
		Config:    mapEffectiveConfig(cfg),
	})
}

func (s *Server) handleCompileSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	request, ok := decodeDSLRequest(w, r)
	if !ok {
		return
	}

	plan, err := s.app.CompileDSL(r.Context(), []byte(request.DSL))
	if err != nil {
		writeError(w, http.StatusBadRequest, "compile_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, mapCompileResponse(plan))
}

func (s *Server) handleRecordingStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	var request apiRecordingStartRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	if request.DSL == "" {
		writeError(w, http.StatusBadRequest, "missing_dsl", "dsl is required")
		return
	}

	state, err := s.recordings.Start(
		[]byte(request.DSL),
		durationFromSeconds(request.GracePeriodSeconds),
		durationFromSeconds(request.MaxDurationSeconds),
	)
	if err != nil {
		if err == ErrRecordingAlreadyActive {
			writeJSON(w, http.StatusConflict, apiSessionEnvelope{
				Session: mapRecordingSessionResponse(state),
			})
			return
		}
		writeError(w, http.StatusBadRequest, "recording_start_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, apiSessionEnvelope{
		Session: mapRecordingSessionResponse(state),
	})
}

func (s *Server) handleRecordingStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, apiSessionEnvelope{
		Session: mapRecordingSessionResponse(s.recordings.Stop()),
	})
}

func decodeDSLRequest(w http.ResponseWriter, r *http.Request) (*apiDSLRequest, bool) {
	var request apiDSLRequest
	if !decodeJSON(w, r, &request) {
		return nil, false
	}
	if request.DSL == "" {
		writeError(w, http.StatusBadRequest, "missing_dsl", "dsl is required")
		return nil, false
	}
	return &request, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", err.Error())
		return false
	}
	if err := json.Unmarshal(body, out); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return false
	}
	return true
}
