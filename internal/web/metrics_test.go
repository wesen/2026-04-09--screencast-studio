package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appmetrics "github.com/wesen/2026-04-09--screencast-studio/pkg/metrics"
)

func TestMetricsEndpoint(t *testing.T) {
	counter := appmetrics.MustRegisterCounterVec(
		"screencast_studio_test_metrics_endpoint_total",
		"Test-only counter to verify metrics endpoint rendering.",
		"kind",
	)
	gauge := appmetrics.MustRegisterGaugeVec(
		"screencast_studio_test_metrics_endpoint_gauge",
		"Test-only gauge to verify metrics endpoint rendering.",
		"kind",
	)
	counter.Inc(map[string]string{"kind": "server_test"})
	gauge.Set(map[string]string{"kind": "server_test"}, 2)
	previewHTTPClients.Set(map[string]string{"source_type": "display"}, 1)
	previewHTTPStreamsStarted.Inc(map[string]string{"source_type": "display"})
	previewHTTPLoopIterations.Inc(map[string]string{"source_type": "display"})
	previewHTTPIdleIterations.Inc(map[string]string{"source_type": "display"})
	previewHTTPWriteNanoseconds.Add(map[string]string{"source_type": "display"}, 123)
	previewHTTPFlushNanoseconds.Add(map[string]string{"source_type": "display"}, 45)
	eventHubSubscribers.Set(nil, 1)
	eventHubEventsPublished.Inc(map[string]string{"event_type": "test.event"})
	websocketConnections.Set(nil, 1)
	websocketEventsWritten.Inc(map[string]string{"event_type": "test.event"})

	server := NewServer(context.Background(), &fakeApplication{}, Config{})
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "text/plain") {
		t.Fatalf("content-type = %q, want text/plain", got)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "screencast_studio_test_metrics_endpoint_total") {
		t.Fatalf("metrics body missing test metric: %s", body)
	}
	if !strings.Contains(body, `kind="server_test"`) {
		t.Fatalf("metrics body missing expected label set: %s", body)
	}
	if !strings.Contains(body, "# TYPE screencast_studio_test_metrics_endpoint_gauge gauge") {
		t.Fatalf("metrics body missing gauge type line: %s", body)
	}
	if !strings.Contains(body, `screencast_studio_test_metrics_endpoint_gauge{kind="server_test"} 2`) {
		t.Fatalf("metrics body missing gauge sample: %s", body)
	}
	if !strings.Contains(body, `screencast_studio_preview_http_clients{source_type="display"} 1`) {
		t.Fatalf("metrics body missing preview client gauge: %s", body)
	}
	if !strings.Contains(body, `screencast_studio_preview_http_streams_started_total{source_type="display"} 1`) {
		t.Fatalf("metrics body missing preview stream counter: %s", body)
	}
	if !strings.Contains(body, `screencast_studio_preview_http_loop_iterations_total{source_type="display"} 1`) {
		t.Fatalf("metrics body missing preview loop counter: %s", body)
	}
	if !strings.Contains(body, `screencast_studio_preview_http_idle_iterations_total{source_type="display"} 1`) {
		t.Fatalf("metrics body missing preview idle counter: %s", body)
	}
	if !strings.Contains(body, `screencast_studio_preview_http_write_nanoseconds_total{source_type="display"} 123`) {
		t.Fatalf("metrics body missing preview write duration counter: %s", body)
	}
	if !strings.Contains(body, `screencast_studio_preview_http_flush_nanoseconds_total{source_type="display"} 45`) {
		t.Fatalf("metrics body missing preview flush duration counter: %s", body)
	}
	if !strings.Contains(body, `screencast_studio_eventhub_subscribers 1`) {
		t.Fatalf("metrics body missing event hub subscriber gauge: %s", body)
	}
	if !strings.Contains(body, `screencast_studio_eventhub_events_published_total{event_type="test.event"} 1`) {
		t.Fatalf("metrics body missing event hub published counter: %s", body)
	}
	if !strings.Contains(body, `screencast_studio_websocket_connections 1`) {
		t.Fatalf("metrics body missing websocket connections gauge: %s", body)
	}
	if !strings.Contains(body, `screencast_studio_websocket_events_written_total{event_type="test.event"} 1`) {
		t.Fatalf("metrics body missing websocket events written counter: %s", body)
	}
}
