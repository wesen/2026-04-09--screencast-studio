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
	previewHTTPLoopIterations = appmetrics.MustRegisterCounterVec(
		"screencast_studio_preview_http_loop_iterations_total",
		"Total iterations of the HTTP MJPEG preview handler loop.",
		"source_type",
	)
	previewHTTPIdleIterations = appmetrics.MustRegisterCounterVec(
		"screencast_studio_preview_http_idle_iterations_total",
		"Total HTTP MJPEG preview handler loop iterations that did not serve a new frame.",
		"source_type",
	)
	previewHTTPWriteNanoseconds = appmetrics.MustRegisterCounterVec(
		"screencast_studio_preview_http_write_nanoseconds_total",
		"Total time spent writing MJPEG multipart headers and JPEG bytes to HTTP preview clients, in nanoseconds.",
		"source_type",
	)
	previewHTTPFlushNanoseconds = appmetrics.MustRegisterCounterVec(
		"screencast_studio_preview_http_flush_nanoseconds_total",
		"Total time spent in HTTP flush calls while serving MJPEG preview streams, in nanoseconds.",
		"source_type",
	)
	previewFrameUpdates = appmetrics.MustRegisterCounterVec(
		"screencast_studio_preview_frame_updates_total",
		"Total preview frame updates stored by PreviewManager.",
		"source_type",
	)
	previewFrameStoreNanoseconds = appmetrics.MustRegisterCounterVec(
		"screencast_studio_preview_frame_store_nanoseconds_total",
		"Total time spent storing preview frames in PreviewManager, including cached-frame copy and preview.state publication, in nanoseconds.",
		"source_type",
	)
	previewLatestFrameCopyNanoseconds = appmetrics.MustRegisterCounterVec(
		"screencast_studio_preview_latest_frame_copy_nanoseconds_total",
		"Total time spent copying cached preview frames out of PreviewManager.LatestFrame, in nanoseconds.",
		"source_type",
	)
	previewStatePublishNanoseconds = appmetrics.MustRegisterCounterVec(
		"screencast_studio_preview_state_publish_nanoseconds_total",
		"Total time spent publishing preview.state events from PreviewManager, in nanoseconds.",
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
