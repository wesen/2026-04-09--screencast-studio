package web

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	studiov1 "github.com/wesen/2026-04-09--screencast-studio/gen/go/proto/screencast/studio/v1"
)

func (s *Server) handlePreviewEnsure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	var request studiov1.PreviewEnsureRequest
	if !decodeProtoJSON(w, r, &request) {
		return
	}
	if request.GetDsl() == "" || request.GetSourceId() == "" {
		writeError(w, http.StatusBadRequest, "invalid_preview_request", "dsl and source_id are required")
		return
	}

	preview, err := s.previews.Ensure(r.Context(), []byte(request.GetDsl()), request.GetSourceId())
	if err != nil {
		switch err {
		case ErrPreviewLimitExceeded:
			writeError(w, http.StatusConflict, "preview_limit_exceeded", err.Error())
		case ErrPreviewSourceNotFound:
			writeError(w, http.StatusNotFound, "preview_source_not_found", err.Error())
		default:
			writeError(w, http.StatusBadRequest, "preview_ensure_failed", err.Error())
		}
		return
	}

	writeProtoJSON(w, http.StatusOK, &studiov1.PreviewEnvelope{Preview: mapPreviewResponse(preview)})
}

func (s *Server) handlePreviewRelease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	var request studiov1.PreviewReleaseRequest
	if !decodeProtoJSON(w, r, &request) {
		return
	}
	if request.GetPreviewId() == "" {
		writeError(w, http.StatusBadRequest, "missing_preview_id", "preview_id is required")
		return
	}

	preview, err := s.previews.Release(request.GetPreviewId())
	if err != nil {
		writeError(w, http.StatusNotFound, "preview_not_found", err.Error())
		return
	}

	writeProtoJSON(w, http.StatusOK, &studiov1.PreviewEnvelope{Preview: mapPreviewResponse(preview)})
}

func (s *Server) handlePreviewList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	previews := s.previews.List()
	writeProtoJSON(w, http.StatusOK, mapPreviewListResponse(previews))
}

func (s *Server) handlePreviewMJPEG(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	previewID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/previews/"), "/mjpeg")
	if previewID == "" {
		writeError(w, http.StatusBadRequest, "missing_preview_id", "missing preview id")
		return
	}

	if _, ok := s.previews.Snapshot(previewID); !ok {
		writeError(w, http.StatusNotFound, "preview_not_found", "preview not found")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming_not_supported", "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var lastSeq uint64
	for {
		frame, seq, snapshot, ok := s.previews.LatestFrame(previewID)
		if !ok {
			return
		}
		if len(frame) > 0 && seq != lastSeq {
			lastSeq = seq
			if _, err := w.Write([]byte("--frame\r\nContent-Type: image/jpeg\r\nContent-Length: ")); err != nil {
				return
			}
			if _, err := w.Write([]byte(strconv.Itoa(len(frame)))); err != nil {
				return
			}
			if _, err := w.Write([]byte("\r\n\r\n")); err != nil {
				return
			}
			if _, err := io.Copy(w, bytes.NewReader(frame)); err != nil {
				return
			}
			if _, err := w.Write([]byte("\r\n")); err != nil {
				return
			}
			flusher.Flush()
		}
		if (snapshot.State == "failed" || snapshot.State == "finished") && !snapshot.HasFrame {
			return
		}
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
		}
	}
}
