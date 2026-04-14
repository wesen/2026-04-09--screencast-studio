package web

import (
	"net/http"

	appmetrics "github.com/wesen/2026-04-09--screencast-studio/pkg/metrics"
)

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "metrics endpoint only supports GET")
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	if err := appmetrics.DefaultRegistry().WritePrometheus(w); err != nil {
		writeError(w, http.StatusInternalServerError, "metrics_write_failed", "failed to render metrics")
		return
	}
}
