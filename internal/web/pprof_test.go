package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPprofMuxServesIndex(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/?debug=1", nil)
	rec := httptest.NewRecorder()

	newPprofMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("expected non-empty pprof index response body")
	}
}
