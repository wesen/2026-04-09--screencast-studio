package web

import (
	"strings"

	appmetrics "github.com/wesen/2026-04-09--screencast-studio/pkg/metrics"
)

var (
	previewHTTPClients = appmetrics.MustRegisterGaugeVec(
		"screencast_studio_preview_http_clients",
		"Current number of active HTTP MJPEG preview clients connected to the server.",
		"source_type",
	)
	previewHTTPStreamsStarted = appmetrics.MustRegisterCounterVec(
		"screencast_studio_preview_http_streams_started_total",
		"Total HTTP MJPEG preview streams started by the server.",
		"source_type",
	)
	previewHTTPStreamsFinished = appmetrics.MustRegisterCounterVec(
		"screencast_studio_preview_http_streams_finished_total",
		"Total HTTP MJPEG preview streams finished by the server.",
		"source_type",
		"reason",
	)
	previewHTTPFramesServed = appmetrics.MustRegisterCounterVec(
		"screencast_studio_preview_http_frames_served_total",
		"Total JPEG preview frames written to HTTP MJPEG clients.",
		"source_type",
	)
	previewHTTPBytesServed = appmetrics.MustRegisterCounterVec(
		"screencast_studio_preview_http_bytes_served_total",
		"Total bytes written to HTTP MJPEG preview clients, including multipart framing.",
		"source_type",
	)
	previewHTTPFlushes = appmetrics.MustRegisterCounterVec(
		"screencast_studio_preview_http_flushes_total",
		"Total HTTP flush calls performed while serving MJPEG preview streams.",
		"source_type",
	)
	previewFrameUpdates = appmetrics.MustRegisterCounterVec(
		"screencast_studio_preview_frame_updates_total",
		"Total preview frame updates stored by PreviewManager.",
		"source_type",
	)
	previewEnsures = appmetrics.MustRegisterCounterVec(
		"screencast_studio_preview_ensures_total",
		"Total preview ensure attempts handled by PreviewManager.",
		"source_type",
		"result",
	)
	previewReleases = appmetrics.MustRegisterCounterVec(
		"screencast_studio_preview_releases_total",
		"Total preview release requests handled by PreviewManager.",
		"source_type",
		"result",
	)
)

func previewMetricLabels(sourceType string) map[string]string {
	sourceType = strings.TrimSpace(sourceType)
	if sourceType == "" {
		sourceType = "unknown"
	}
	return map[string]string{"source_type": sourceType}
}

func previewMetricLabelsWithResult(sourceType, result string) map[string]string {
	labels := previewMetricLabels(sourceType)
	labels["result"] = strings.TrimSpace(result)
	if labels["result"] == "" {
		labels["result"] = "unknown"
	}
	return labels
}

func previewMetricLabelsWithReason(sourceType, reason string) map[string]string {
	labels := previewMetricLabels(sourceType)
	labels["reason"] = strings.TrimSpace(reason)
	if labels["reason"] == "" {
		labels["reason"] = "unknown"
	}
	return labels
}
