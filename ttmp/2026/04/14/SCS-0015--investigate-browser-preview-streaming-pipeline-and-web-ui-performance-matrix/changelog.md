# Changelog

## 2026-04-14

Created ticket **SCS-0015** to investigate the browser-connected preview streaming path and the CPU gap between earlier backend/API-only performance work and the much hotter real Studio page behavior reported by the user.

Wrote the first primary analysis document:

- `design/01-browser-preview-streaming-pipeline-analysis-and-performance-matrix-plan.md`

That document maps the current path from the shared GStreamer preview branch through PreviewManager’s cached JPEG frames, through the HTTP MJPEG handler, and finally to the frontend `<img>` preview renderer. It also defines the missing measurement scenarios that should be added next: browser-attached desktop-only, camera-only, desktop-plus-camera, and multi-tab matrices.

Started the ticket diary:

- `reference/01-investigation-diary.md`

Added a final report scaffold:

- `design/02-browser-preview-streaming-performance-report.md`

The current conclusion at this stage is not that the browser path is already proven guilty. It is that the codebase clearly contains a separate browser-facing media transport boundary — MJPEG over `/api/previews/{id}/mjpeg` — that has not yet been measured with the same rigor as the shared runtime itself. This ticket exists to close that gap.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go — browser-facing MJPEG transport implementation
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go — preview reuse, cached-frame storage, and lifecycle behavior
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/preview.go — GStreamer preview runtime acquires shared sources and attaches preview consumers
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go — upstream preview branch performs JPEG encoding into appsink consumers
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx — frontend preview ensure/release lifecycle and preview ownership
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/preview/PreviewStream.tsx — frontend renders MJPEG previews through img tags
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/metrics/metrics.go — metrics registry that will be extended for browser-preview observability

Implemented the first SCS-0015 code slice in commit `1c99094caf1a3661562c26332b2e57fd257de2a4` (`Add preview serving and runtime metrics`).

This slice did three things:

1. extended the in-process metrics registry to support both counters and gauges,
2. wired new browser-preview metrics into the MJPEG serving path and PreviewManager lifecycle,
3. kept label cardinality deliberately low by using `source_type` plus small bounded `reason` / `result` enums instead of per-preview or per-source IDs.

The new metric families now include:

- `screencast_studio_preview_http_clients`
- `screencast_studio_preview_http_streams_started_total`
- `screencast_studio_preview_http_streams_finished_total`
- `screencast_studio_preview_http_frames_served_total`
- `screencast_studio_preview_http_bytes_served_total`
- `screencast_studio_preview_http_flushes_total`
- `screencast_studio_preview_frame_updates_total`
- `screencast_studio_preview_ensures_total`
- `screencast_studio_preview_releases_total`

The focused validation for this slice was:

```bash
gofmt -w pkg/metrics/metrics.go pkg/metrics/metrics_test.go internal/web/preview_metrics.go internal/web/handlers_preview.go internal/web/preview_manager.go internal/web/metrics_test.go internal/web/handlers_metrics.go internal/web/routes.go pkg/media/gst/recording.go pkg/media/gst/shared_video_recording_bridge.go
go test ./pkg/metrics ./internal/web ./pkg/media/gst -count=1
go test ./... -count=1
```

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_metrics.go — new metric-family definitions for browser preview serving and PreviewManager lifecycle events
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_metrics.go — `/metrics` endpoint rendering remains the export surface for the new metrics
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go — MJPEG handler now tracks active clients, stream starts/finishes, frames, bytes, and flushes
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go — preview ensures, releases, and frame updates now emit metrics
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/metrics_test.go — focused endpoint test now asserts both gauge and preview-serving metric visibility
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/metrics/metrics_test.go — new registry test verifies counter and gauge rendering
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/recording.go — audio-level parse failures now contribute to exported metrics
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_recording_bridge.go — shared bridge recorder counters now contribute to exported metrics

Implemented the next SCS-0015 helper-script slice in commit `fb87ab7f33c0eb2e379d83fe0e9b16c54aae70e9` (`Add browser preview metrics helper scripts`).

This slice added the first ticket-local runtime helpers under `scripts/`:

- `scripts/01-restart-scs-web-ui.sh` — restarts the local server in the `scs-web-ui` tmux session and waits for `/api/healthz`
- `scripts/02-sample-preview-metrics.sh` — snapshots `/metrics` repeatedly into a timestamped result directory with a manifest and summary file

I validated both scripts live. The restart helper brought the server back up successfully on `:7777`, and the sampler saved the first smoke result under:

- `scripts/results/20260414-160358/`

The saved raw `.prom` snapshots already show the new preview-serving metric families in the export output, which is enough to confirm the first measurement surface is live before the heavier browser-driven matrix harnesses exist.

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/01-restart-scs-web-ui.sh — local tmux restart helper for the live server
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/02-sample-preview-metrics.sh — ticket-local metrics sampler for later matrix runs
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-160358/01-manifest.tsv — first saved metrics-sampling manifest
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-160358/02-summary.txt — first saved metrics-sampling summary
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-160358/raw/001-1776197038.prom — smoke snapshot proving the new preview-serving metric families are exported

Implemented the first actual matrix harness slice in commit `70b22663dda79cb399b67c89b450df397ad611c9` (`Add desktop preview HTTP client matrix`).

This slice is intentionally framed as a **server-side MJPEG HTTP-client baseline**, not yet the full real-browser-tab matrix. The goal was to isolate whether preview-stream client fan-out alone already changes server CPU before the browser automation layer is added.

The new harness lives at:

- `scripts/03-desktop-preview-http-client-matrix/run.sh`

It builds a dedicated server binary once, then runs three clean scenarios against separate ports and fresh server processes:

- `no-client`
- `one-client`
- `two-clients`

Each scenario:

- ensures one desktop preview through the real API,
- samples `/metrics` repeatedly,
- samples server CPU with `pidstat`,
- and, when clients are enabled, holds open one or more MJPEG preview streams using `curl` as the stream consumer.

The first saved run is:

- `scripts/03-desktop-preview-http-client-matrix/results/20260414-161024/`

A short human-readable summary was also added at:

- `scripts/04-desktop-preview-http-client-baseline-summary.md`

The early result from this first `DURATION=4` baseline is:

- `no-client` → `11.67%` avg CPU
- `one-client` → `11.50%` avg CPU
- `two-clients` → `15.50%` avg CPU

That is not yet enough to explain the full web-UI problem, but it does support one important direction: the shared desktop preview itself already costs CPU, and multiple MJPEG clients can raise that cost further.

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/03-desktop-preview-http-client-matrix/run.sh — first dedicated server-side preview-stream fan-out matrix harness
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/03-desktop-preview-http-client-matrix/results/20260414-161024/01-summary.md — top-level summary for the first matrix run
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/03-desktop-preview-http-client-matrix/results/20260414-161024/no-client/01-summary.md — baseline scenario with preview active but no MJPEG client
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/03-desktop-preview-http-client-matrix/results/20260414-161024/one-client/01-summary.md — one-client desktop preview baseline
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/03-desktop-preview-http-client-matrix/results/20260414-161024/two-clients/01-summary.md — two-client desktop preview baseline
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/04-desktop-preview-http-client-baseline-summary.md — human-readable interpretation note for the first baseline run
