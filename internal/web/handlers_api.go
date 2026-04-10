package web

import (
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
