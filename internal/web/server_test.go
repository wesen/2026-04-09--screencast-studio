package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/app"
)

func TestHealthz(t *testing.T) {
	t.Parallel()

	server := NewServer(app.New(), Config{
		Addr:         ":0",
		PreviewLimit: 7,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload["service"] != "screencast-studio" {
		t.Fatalf("service = %v, want screencast-studio", payload["service"])
	}
	if got, ok := payload["preview_limit"].(float64); !ok || int(got) != 7 {
		t.Fatalf("preview_limit = %v, want 7", payload["preview_limit"])
	}
}

func TestWebsocketPlaceholder(t *testing.T) {
	t.Parallel()

	server := NewServer(app.New(), Config{})

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotImplemented)
	}
}
