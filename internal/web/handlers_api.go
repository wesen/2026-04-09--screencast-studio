package web

import (
	"encoding/json"
	"net/http"

	studiov1 "github.com/wesen/2026-04-09--screencast-studio/gen/go/proto/screencast/studio/v1"
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

	writeProtoJSON(w, http.StatusOK, mapDiscoveryResponse(snapshot))
}

func (s *Server) handleSessionCurrent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	writeProtoJSON(w, http.StatusOK, mapSessionEnvelope(s.recordings.Current()))
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

	cfg, err := s.app.NormalizeDSL(r.Context(), []byte(request.GetDsl()))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_dsl", err.Error())
		return
	}

	writeProtoJSON(w, http.StatusOK, mapNormalizeResponse(cfg))
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

	plan, err := s.app.CompileDSL(r.Context(), []byte(request.GetDsl()))
	if err != nil {
		writeError(w, http.StatusBadRequest, "compile_failed", err.Error())
		return
	}
	s.telemetry.UpdateFromPlan(plan)

	writeProtoJSON(w, http.StatusOK, mapCompileResponse(plan))
}

func (s *Server) handleRecordingStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	var request studiov1.RecordingStartRequest
	if !decodeProtoJSON(w, r, &request) {
		return
	}
	if request.GetDsl() == "" {
		writeError(w, http.StatusBadRequest, "missing_dsl", "dsl is required")
		return
	}

	current := s.recordings.Current()
	if current.Active {
		writeProtoJSON(w, http.StatusConflict, mapSessionEnvelope(current))
		return
	}

	state, err := s.recordings.Start(
		[]byte(request.GetDsl()),
		durationFromSeconds(int(request.GetGracePeriodSeconds())),
		durationFromSeconds(int(request.GetMaxDurationSeconds())),
	)
	if err != nil {
		if err == ErrRecordingAlreadyActive {
			writeProtoJSON(w, http.StatusConflict, mapSessionEnvelope(state))
			return
		}
		writeError(w, http.StatusBadRequest, "recording_start_failed", err.Error())
		return
	}

	writeProtoJSON(w, http.StatusOK, mapSessionEnvelope(state))
}

func (s *Server) handleRecordingStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	writeProtoJSON(w, http.StatusOK, mapSessionEnvelope(s.recordings.Stop()))
}

type audioEffectsRequest struct {
	SourceID          string   `json:"source_id"`
	Gain              *float64 `json:"gain,omitempty"`
	CompressorEnabled *bool    `json:"compressor_enabled,omitempty"`
}

func (s *Server) handleAudioEffects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	current := s.recordings.Current()
	if !current.Active || current.SessionID == "" {
		writeError(w, http.StatusConflict, "recording_not_active", "recording is not active")
		return
	}
	var request audioEffectsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_audio_effects_request", err.Error())
		return
	}
	if request.Gain == nil && request.CompressorEnabled == nil {
		writeError(w, http.StatusBadRequest, "invalid_audio_effects_request", "at least one effect field is required")
		return
	}
	if request.Gain != nil {
		if err := s.app.SetRecordingAudioGain(r.Context(), current.SessionID, request.SourceID, *request.Gain); err != nil {
			writeError(w, http.StatusBadRequest, "audio_gain_update_failed", err.Error())
			return
		}
	}
	if request.CompressorEnabled != nil {
		if err := s.app.SetRecordingCompressorEnabled(r.Context(), current.SessionID, *request.CompressorEnabled); err != nil {
			writeError(w, http.StatusBadRequest, "audio_compressor_update_failed", err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func decodeDSLRequest(w http.ResponseWriter, r *http.Request) (*studiov1.DslRequest, bool) {
	var request studiov1.DslRequest
	if !decodeProtoJSON(w, r, &request) {
		return nil, false
	}
	if request.GetDsl() == "" {
		writeError(w, http.StatusBadRequest, "missing_dsl", "dsl is required")
		return nil, false
	}
	return &request, true
}
