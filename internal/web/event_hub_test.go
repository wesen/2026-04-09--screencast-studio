package web

import (
	"strings"
	"testing"

	appmetrics "github.com/wesen/2026-04-09--screencast-studio/pkg/metrics"
)

func TestEventHubPublishesDeliveryAndDropMetrics(t *testing.T) {
	hub := NewEventHub()
	ch, unsubscribe := hub.Subscribe(1)
	defer unsubscribe()

	hub.Publish(ServerEvent{Type: "test.event.hub"})
	hub.Publish(ServerEvent{Type: "test.event.hub"})

	select {
	case <-ch:
	default:
		t.Fatalf("expected first event to be buffered")
	}

	var b strings.Builder
	if err := appmetrics.DefaultRegistry().WritePrometheus(&b); err != nil {
		t.Fatalf("write metrics: %v", err)
	}
	body := b.String()
	if !strings.Contains(body, `screencast_studio_eventhub_events_published_total{event_type="test.event.hub"} 2`) {
		t.Fatalf("metrics body missing published count: %s", body)
	}
	if !strings.Contains(body, `screencast_studio_eventhub_events_delivered_total{event_type="test.event.hub"} 1`) {
		t.Fatalf("metrics body missing delivered count: %s", body)
	}
	if !strings.Contains(body, `screencast_studio_eventhub_events_dropped_total{event_type="test.event.hub"} 1`) {
		t.Fatalf("metrics body missing dropped count: %s", body)
	}
}
