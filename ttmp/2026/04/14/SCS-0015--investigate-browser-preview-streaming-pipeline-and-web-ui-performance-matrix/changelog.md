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

Implemented the next SCS-0015 script bundle in commit `ad0523a5d8d8f1a891d9187577402f528491b691` (`Add browser preview performance matrix harnesses`).

This slice added four major pieces:

1. a larger fresh-server desktop preview matrix that includes recording,
2. a live-server sampler for real browser scenarios,
3. a reusable Playwright script bundle for one-tab, two-tab, recording, and camera-add scenarios,
4. a small cleanup to the earlier HTTP-client baseline harness so it no longer writes a built server binary into the saved result tree.

The new measurement scripts are:

- `scripts/05-desktop-preview-http-client-recording-matrix/run.sh`
- `scripts/07-live-server-browser-scenario-sample.sh`
- `scripts/08-playwright-browser-matrix/00-usage.txt`
- `scripts/08-playwright-browser-matrix/01-open-studio-and-wait-desktop.js`
- `scripts/08-playwright-browser-matrix/02-open-second-desktop-tab.js`
- `scripts/08-playwright-browser-matrix/03-start-recording.js`
- `scripts/08-playwright-browser-matrix/04-stop-recording.js`
- `scripts/08-playwright-browser-matrix/05-add-camera-if-needed.js`
- `scripts/08-playwright-browser-matrix/06-capture-browser-preview-state.js`

I then used those scripts to produce the first larger result set.

### Fresh-server HTTP-client matrix with recording

Saved under:

- `scripts/05-desktop-preview-http-client-recording-matrix/results/20260414-163154/`

Key CPU results:

- `preview-no-client` → `15.67%`
- `preview-one-client` → `18.11%`
- `preview-two-clients` → `19.22%`
- `record-no-client` → `162.22%`
- `record-one-client` → `158.56%`
- `record-two-clients` → `165.00%`

That matrix says plain MJPEG client fan-out alone does **not** explain the user-observed browser-path `~400%` CPU behavior.

### Live browser-backed measurements against the real Studio page

Saved under:

- `scripts/results/20260414-163610/` — desktop, one tab, preview only
- `scripts/results/20260414-163951/` — desktop, one tab, preview + recording
- `scripts/results/20260414-164457/` — desktop, two tabs, preview only
- `scripts/results/20260414-164535/` — desktop, two tabs, preview + recording
- `scripts/results/20260414-164657/` — desktop + camera, one tab, preview only
- `scripts/results/20260414-164720/` — desktop + camera, one tab, preview + recording
- `scripts/results/20260414-165126/` — short validation rerun of the improved browser sampler with direct metric-delta output

Key CPU results from the browser-backed scenarios:

- desktop, one tab, preview only → `14.20%`
- desktop, one tab, preview + recording → `410.60%`
- desktop, two tabs, preview only → `12.69%`
- desktop, two tabs, preview + recording → `432.97%`
- desktop + camera, one tab, preview only → `20.10%`
- desktop + camera, one tab, preview + recording → `343.71%`

This is the strongest current SCS-0015 finding: the **real browser-connected recording path** is dramatically hotter than the fresh-server plain-MJPEG-client baseline.

The clearest side-by-side comparison is:

- fresh server, desktop preview + recording + one plain MJPEG client → `158.56%` avg CPU
- live server, desktop preview + recording + one real browser tab → `410.60%` avg CPU

I also saved additional human-readable/browser-evidence files:

- `scripts/09-browser-preview-matrix-findings-summary.md`
- `scripts/10-browser-session-network-summary.txt`
- `scripts/11-browser-playwright-state-desktop-camera.json`

The findings note records another important nuance: during the browser recording runs, the per-run MJPEG frame/byte deltas were relatively modest compared with preview-only runs. That means the browser-path heat is **not** explained simply by “the server had to send vastly more JPEG data.” The likely explanation is a combination of browser-connected lifecycle behavior plus the already-expensive upstream preview-plus-recording interaction.

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/05-desktop-preview-http-client-recording-matrix/run.sh — larger fresh-server matrix for preview-only and preview-plus-recording across 0/1/2 MJPEG clients
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/07-live-server-browser-scenario-sample.sh — live server sampler for browser-backed scenarios, including per-run preview-metric deltas in the newer runs
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/08-playwright-browser-matrix/01-open-studio-and-wait-desktop.js — opens the Studio page and waits for the desktop preview to load
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/08-playwright-browser-matrix/02-open-second-desktop-tab.js — second-tab helper for the multi-tab matrix slice
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/08-playwright-browser-matrix/03-start-recording.js — browser-side recording start helper
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/08-playwright-browser-matrix/04-stop-recording.js — browser-side recording stop helper
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/08-playwright-browser-matrix/05-add-camera-if-needed.js — browser-side camera-add helper used for the desktop+camera scenarios
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/09-browser-preview-matrix-findings-summary.md — first human-readable combined findings note for the larger matrix pass
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/10-browser-session-network-summary.txt — raw browser-tool network summary saved as a ticket artifact
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/11-browser-playwright-state-desktop-camera.json — saved browser-side preview state showing both desktop and camera previews loaded successfully

Added a dedicated ongoing lab report document and backfilled the current SCS-0015 experiment set in detail in the docs commit `Add browser preview lab report`.

The new document is:

- `reference/02-browser-preview-streaming-lab-report.md`

This lab report is meant to stay “live” as the investigation continues. Unlike the more polished design/report documents, it preserves the actual experiment ledger in a continuation-friendly form:

- the purpose of each experiment,
- the exact harness or helper scripts used,
- the commands run,
- the saved result directories,
- the measured CPU summaries,
- important caveats,
- and the current working interpretation of the `~400%` desktop preview+recording browser-path spike.

The report explicitly backfills the current experiment inventory, including:

- metrics export smoke validation,
- the first desktop HTTP-client baseline matrix,
- the larger fresh-server HTTP-client matrix with recording,
- the live browser-backed desktop one-tab and two-tab measurements,
- the desktop-plus-camera one-tab measurements,
- the improved browser sampler validation after the heredoc bug fix,
- and the browser-tool network-summary and Playwright-state artifacts.

The lab report also records the strongest current explanation in one place:

- simple MJPEG client fan-out alone does not explain the browser-path spike,
- the real browser-connected recording path is substantially hotter than the plain-client baseline,
- and the most likely explanation is a combination of shared preview+recording cost, MJPEG serving work, multiple Go-side frame copies, and browser-only websocket/event fanout during recording.

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/reference/02-browser-preview-streaming-lab-report.md — ongoing detailed experiment ledger and working interpretation for SCS-0015

Implemented the next SCS-0015 code slice in commit `ede87bb2fb824ee4f427813ba29daedf1b280b22` (`Add websocket and eventhub metrics`).

This slice added low-cardinality server-side observability around the websocket/event path that had become the main suspected browser-only cost center:

- `screencast_studio_eventhub_subscribers`
- `screencast_studio_eventhub_events_published_total`
- `screencast_studio_eventhub_events_delivered_total`
- `screencast_studio_eventhub_events_dropped_total`
- `screencast_studio_websocket_connections`
- `screencast_studio_websocket_events_written_total`
- `screencast_studio_websocket_event_write_errors_total`

It also updated the live browser sampler so `metric-deltas.txt` now includes:

- preview HTTP metrics,
- preview frame updates,
- EventHub metrics,
- websocket metrics.

Focused validation for that slice was:

```bash
gofmt -w internal/web/event_metrics.go internal/web/event_hub.go internal/web/handlers_ws.go internal/web/metrics_test.go internal/web/event_hub_test.go
go test ./internal/web ./pkg/metrics -count=1
go test ./... -count=1
```

Then I added a new focused ablation harness in commit `23234634f3e73c822c58b7af9dc91756519cf2f0` (`Add MJPEG websocket ablation harness`).

The new harness lives at:

- `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/run.sh`
- `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/ws_client/main.go`

It isolates the current high-signal scenario — desktop preview + recording — into two fresh-server cases:

- one MJPEG client only
- one MJPEG client plus one synthetic `/ws` consumer

The first attempted run under:

- `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173359/`

turned out to be invalid for comparison because the `mjpeg-only` case started recording before the preview path had clearly produced an initial frame. That made its `11.00%` result unusable. I fixed the harness by waiting for an initial preview screenshot before measurement in both scenarios and reran it.

The trusted rerun is:

- `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173541/`

Key CPU results from the trusted rerun:

- `mjpeg-only` → `166.56%` avg CPU
- `mjpeg-plus-ws` → `170.48%` avg CPU

The websocket-enabled case produced real websocket/event traffic, including:

- `preview.state` published: `33`
- `preview.state` delivered: `23`
- websocket `preview.state` writes: `23`
- websocket client total messages observed: `54`

The new short summary note is:

- `scripts/13-mjpeg-websocket-ablation-summary.md`

This is an important refinement of the current hypothesis: websocket/event fanout is **real**, but by itself it only raised avg CPU by about `3.92` points in the focused fresh-server ablation. That is far too small to explain the jump from the fresh-server desktop recording band (`~158–166%`) to the real browser one-tab desktop recording case (`410.60%`).

So the browser-path spike is still real, but the evidence now says websocket fanout alone is not the dominant explanation.

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/event_metrics.go — new EventHub and websocket metric-family definitions
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/event_hub.go — EventHub now counts published, delivered, dropped events and tracks subscriber count
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go — websocket handler now tracks active connections, event writes, and write errors
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/event_hub_test.go — focused EventHub metrics regression test
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/run.sh — focused MJPEG-vs-websocket ablation harness for the desktop preview+recording scenario
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/ws_client/main.go — synthetic websocket consumer used by the ablation harness
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173541/01-summary.md — trusted ablation rerun summary
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/13-mjpeg-websocket-ablation-summary.md — short human-readable summary of the ablation result

Implemented the next SCS-0015 instrumentation slice in commit `9fd8754ab4db6aab3ce0bd174c2ac006d957b1dd` (`Add MJPEG handler timing metrics`).

This slice deliberately followed the websocket ablation with **measurement before behavior change**. The goal was to instrument the browser-facing MJPEG handler so the next high-signal real-browser rerun can answer whether the browser-path spike is dominated by MJPEG write/flush cost or whether the missing heat still mostly lives elsewhere.

The new metric families are:

- `screencast_studio_preview_http_loop_iterations_total`
- `screencast_studio_preview_http_idle_iterations_total`
- `screencast_studio_preview_http_write_nanoseconds_total`
- `screencast_studio_preview_http_flush_nanoseconds_total`

The implementation adds low-cardinality timing counters around the existing `handlePreviewMJPEG(...)` loop in `internal/web/handlers_preview.go`:

- every handler loop increments a loop counter,
- iterations that do not serve a new frame increment an idle counter,
- successful multipart writes accumulate write nanoseconds,
- successful flushes accumulate flush nanoseconds.

I also extended tests so the timing metrics are visible at `/metrics` and so an end-to-end preview request proves the new MJPEG timing families appear after a real stream is served.

Validation for this slice was:

```bash
gofmt -w internal/web/preview_metrics.go internal/web/handlers_preview.go internal/web/metrics_test.go internal/web/server_test.go
go test ./internal/web ./pkg/metrics -count=1
go test ./... -count=1
```

This slice does **not** claim a new performance conclusion yet. It is preparation for the next rerun of the highest-value scenario:

- desktop preview + recording + one real browser tab

That rerun should now let us compare write/flush/loop timing deltas directly instead of guessing whether the browser’s extra heat is coming from the MJPEG handler itself.

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_metrics.go — preview metric-family definitions now include loop/idle/write/flush timing counters
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go — MJPEG handler now accumulates loop, idle, write-duration, and flush-duration metrics
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/metrics_test.go — `/metrics` endpoint test now asserts the new timing metric families
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server_test.go — preview MJPEG endpoint test now checks that the timing metrics appear after serving a real stream

Used the recovered browser tool to run the next high-signal real-browser validation with the new MJPEG timing metrics enabled.

Saved result directory:

- `scripts/results/20260414-202519/`

This rerun used the real Studio page with one browser tab, desktop preview, and recording. The average CPU number from this run (`163.34%`) is **not** directly comparable to the earlier `410.60%` one-tab browser run because the sampled window clearly included several preview-only seconds before recording fully ramped. The per-second `pidstat` trace still climbed late into the very hot band:

- `377.00%`
- `407.92%`

The most important new evidence from this rerun is the timing-metric deltas:

- `screencast_studio_preview_http_write_nanoseconds_total{source_type="display"}` → `8,724,144`
- `screencast_studio_preview_http_flush_nanoseconds_total{source_type="display"}` → `908,218`
- `screencast_studio_preview_http_frames_served_total{source_type="display"}` → `55`
- `screencast_studio_preview_http_loop_iterations_total{source_type="display"}` → `71`
- `screencast_studio_preview_http_idle_iterations_total{source_type="display"}` → `16`

Those numbers imply extremely small final-write costs per served frame:

- write: about `0.159ms/frame`
- flush: about `0.017ms/frame`

That is the key refinement from this run: the final MJPEG HTTP write/flush loop does **not** currently look expensive enough to explain a process that still reached `~378–408%` CPU late in the run.

So the current hypothesis sharpens again:

- the browser-path spike is still real,
- but the dominant hot slice is probably **upstream of the final HTTP write/flush step**, likely in shared preview+recording interaction, Go-side frame-copy/publication work before the write, browser-connected lifecycle behavior, or some combination of those.

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-202519/01-summary.md — first real-browser rerun with MJPEG timing metrics enabled
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-202519/metric-deltas.txt — timing, preview, websocket, and EventHub deltas for the rerun
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-202519/server.pidstat.log — per-second CPU trace showing the late-run climb into the hot band
