package web

import (
	"strings"

	appmetrics "github.com/wesen/2026-04-09--screencast-studio/pkg/metrics"
)

var (
	eventHubSubscribers = appmetrics.MustRegisterGaugeVec(
		"screencast_studio_eventhub_subscribers",
		"Current number of active EventHub subscribers.",
	)
	eventHubEventsPublished = appmetrics.MustRegisterCounterVec(
		"screencast_studio_eventhub_events_published_total",
		"Total server events published into the EventHub.",
		"event_type",
	)
	eventHubEventsDelivered = appmetrics.MustRegisterCounterVec(
		"screencast_studio_eventhub_events_delivered_total",
		"Total server events delivered into subscriber channels by the EventHub.",
		"event_type",
	)
	eventHubEventsDropped = appmetrics.MustRegisterCounterVec(
		"screencast_studio_eventhub_events_dropped_total",
		"Total server events dropped by the EventHub because a subscriber channel was full.",
		"event_type",
	)
	websocketConnections = appmetrics.MustRegisterGaugeVec(
		"screencast_studio_websocket_connections",
		"Current number of active websocket connections to the server.",
	)
	websocketEventsWritten = appmetrics.MustRegisterCounterVec(
		"screencast_studio_websocket_events_written_total",
		"Total server events written to websocket clients.",
		"event_type",
	)
	websocketEventWriteErrors = appmetrics.MustRegisterCounterVec(
		"screencast_studio_websocket_event_write_errors_total",
		"Total websocket event write attempts that returned an error.",
		"event_type",
	)
)

func eventMetricLabels(eventType string) map[string]string {
	eventType = strings.TrimSpace(eventType)
	if eventType == "" {
		eventType = "unknown"
	}
	return map[string]string{"event_type": eventType}
}
