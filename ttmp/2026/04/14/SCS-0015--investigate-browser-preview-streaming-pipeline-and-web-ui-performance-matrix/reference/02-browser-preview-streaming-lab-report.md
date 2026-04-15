---
Title: Browser preview streaming lab report
Ticket: SCS-0015
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - performance
    - preview
    - browser
    - recording
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/event_hub.go
      Note: Browser-connected event fanout path summarized in the lab report
    - Path: internal/web/event_metrics.go
      Note: EventHub and websocket metric families added during the focused ablation slice
    - Path: internal/web/handlers_preview.go
      Note: |-
        MJPEG serving loop under investigation in the lab report, now including loop/write/flush timing metrics
        MJPEG loop under investigation now exports timing counters for the next rerun
    - Path: internal/web/handlers_ws.go
      Note: Real browser path includes websocket event delivery beyond plain MJPEG serving
    - Path: internal/web/preview_manager.go
      Note: Preview frame caching and per-frame preview.state publication are central to the current hypothesis
    - Path: internal/web/preview_metrics.go
      Note: New loop/idle/write/flush timing metric families support INST-01
    - Path: internal/web/session_manager.go
      Note: Recording-time websocket event publication is part of the current working explanation
    - Path: pkg/media/gst/shared_video.go
      Note: Preview frame copy path and adaptive preview recipe are discussed in the lab report
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/05-desktop-preview-http-client-recording-matrix/run.sh
      Note: Fresh-server matrix harness backfilled in the lab report
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/07-live-server-browser-scenario-sample.sh
      Note: Live browser-backed sampler backfilled in the lab report
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/run.sh
      Note: Focused MJPEG-vs-websocket ablation harness backfilled in EXP-12
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/ws_client/main.go
      Note: Synthetic websocket consumer used in EXP-12
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/13-mjpeg-websocket-ablation-summary.md
      Note: Short human-readable summary of the focused ablation result
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-202519/01-summary.md
      Note: First real-browser rerun with MJPEG timing metrics
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-202519/metric-deltas.txt
      Note: Timing deltas that strongly lower confidence in the final MJPEG write/flush loop as the dominant hot path
ExternalSources: []
Summary: Ongoing lab report for the browser preview streaming investigation, with backfilled experiments, exact result directories, scenario definitions, outcomes, caveats, and the current working explanation for the desktop preview-plus-recording CPU spike.
LastUpdated: 2026-04-14T20:36:00-04:00
WhatFor: Keep a continuation-friendly experimental record for SCS-0015 that can be updated as new browser-path measurements and instrumentation land.
WhenToUse: Read this when continuing the browser preview investigation and you need the exact experiments already run, the result locations, and the current evidence-backed interpretation.
---





# Browser preview streaming lab report

## Status

This is the **ongoing lab report** for SCS-0015. It is intentionally more experiment-log-oriented than the design doc and less final-polish than the main performance report. The purpose is to preserve the actual runs, scenario definitions, commands, directories, caveats, and current working interpretation in one place as the investigation continues.

This version backfills the current experiment set through the first browser-backed matrix pass and the next timing-instrumentation slice.

## Core question

Why does the real Studio page drive the server so much hotter than earlier backend/API-only or curl-like preview measurements suggested, especially for the high-signal case:

- **desktop preview + recording + one real browser tab**

## Current best answer, in one paragraph

The browser-connected recording spike appears to be **real and not explained by simple MJPEG byte volume alone**. The fresh-server plain-MJPEG-client recording matrix stayed around `158–165%` avg CPU, while the real browser-backed desktop preview+recording run reached about `410.60%` avg CPU and the two-tab variant reached about `432.97%`. A newer focused ablation also showed that adding a plain `/ws` consumer on top of one MJPEG client only nudged the fresh-server desktop recording case from `166.56%` to `170.48%`, which means websocket/event fanout by itself is **not** enough to explain the full browser gap. The current evidence points more toward a combination of shared-source preview + recording interaction upstream, MJPEG serving work, multiple Go-side frame copies, and browser-specific behavior that is still not reproduced by a dumb MJPEG-plus-websocket client pair.

## Environment and methodology

### Repository and ticket

- repo: `/home/manuel/code/wesen/2026-04-09--screencast-studio`
- ticket: `SCS-0015`
- ticket root: `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/`

### Measurement surfaces used so far

1. **Fresh dedicated-server matrices**
   - build a dedicated server binary
   - run one scenario per fresh port and process
   - sample server CPU with `pidstat`
   - sample `/metrics`
   - optionally attach plain MJPEG consumers via `curl`

2. **Live browser-backed runs on the real local Studio page**
   - use the server running on `http://127.0.0.1:7777`
   - drive scenarios through the actual browser UI via Playwright snippets
   - sample the real listening server PID with `pidstat`
   - sample `/metrics`
   - save periodic `/api/previews` and `/api/recordings/current` snapshots

3. **Supporting artifacts**
   - Playwright state captures
   - browser-tool network summaries
   - per-scenario ffprobe output when recordings are produced

4. **New handler timing metrics for the next rerun**
   - `screencast_studio_preview_http_loop_iterations_total`
   - `screencast_studio_preview_http_idle_iterations_total`
   - `screencast_studio_preview_http_write_nanoseconds_total`
   - `screencast_studio_preview_http_flush_nanoseconds_total`

These timing metrics were added after the websocket-ablation slice so the next real-browser rerun can distinguish “browser makes MJPEG write/flush much hotter” from “the missing cost is mostly elsewhere.”

### Important caveat about comparability

The fresh dedicated-server runs and the live browser-backed runs are intentionally **not identical**. The fresh matrix isolates server-side preview stream fan-out with dumb HTTP consumers. The browser-backed runs include the real Studio page, `/ws`, frontend preview ownership, and whatever server-side event traffic that implies. The gap between those two kinds of runs is part of the point of the investigation, not a flaw in the dataset.

## Experiment inventory

### EXP-00: Metrics export smoke check

Purpose:
- prove that the new browser-preview metric families are live in the runtime before heavier matrix work

Helper script:
- `scripts/02-sample-preview-metrics.sh`

Saved result directory:
- `scripts/results/20260414-160358/`

Outcome:
- raw `.prom` snapshots included the new preview-serving metric families
- this validated the first observability surface before the matrix harnesses were built

### EXP-01: Desktop preview HTTP-client baseline matrix

Purpose:
- measure the server-side MJPEG path in a controlled way before involving a real browser
- answer: how much does server CPU move for the same desktop preview workload when we have `0`, `1`, or `2` plain MJPEG clients?

Harness:
- `scripts/03-desktop-preview-http-client-matrix/run.sh`

Command used:

```bash
DURATION=4 bash ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/03-desktop-preview-http-client-matrix/run.sh
```

Saved result directory:
- `scripts/03-desktop-preview-http-client-matrix/results/20260414-161024/`

Results:

| Scenario | Avg CPU | Max CPU |
| --- | ---: | ---: |
| no-client | 11.67% | 15.00% |
| one-client | 11.50% | 13.00% |
| two-clients | 15.50% | 18.00% |

Interpretation:
- the upstream preview itself already costs CPU even with no attached MJPEG client
- one short run with one client was similar to the no-client case
- two clients clearly raised CPU
- but this was still only a **server-side MJPEG approximation**, not the full browser path

### EXP-02: Desktop preview HTTP-client matrix with recording

Purpose:
- extend the controlled MJPEG baseline to preview-only and preview-plus-recording cases under `0`, `1`, and `2` plain MJPEG consumers
- test whether simple MJPEG client fan-out is enough to explain the browser-path spike

Harness:
- `scripts/05-desktop-preview-http-client-recording-matrix/run.sh`

Command used:

```bash
DURATION=6 bash ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/05-desktop-preview-http-client-recording-matrix/run.sh
```

Saved result directory:
- `scripts/05-desktop-preview-http-client-recording-matrix/results/20260414-163154/`

Results:

| Scenario | Clients | Recording | Avg CPU | Max CPU |
| --- | ---: | ---: | ---: | ---: |
| preview-no-client | 0 | no | 15.67% | 18.00% |
| preview-one-client | 1 | no | 18.11% | 21.00% |
| preview-two-clients | 2 | no | 19.22% | 22.00% |
| record-no-client | 0 | yes | 162.22% | 463.00% |
| record-one-client | 1 | yes | 158.56% | 458.00% |
| record-two-clients | 2 | yes | 165.00% | 464.00% |

Recording output evidence:
- `record-no-client/output/Full Desktop.mov`
- `record-one-client/output/Full Desktop.mov`
- `record-two-clients/output/Full Desktop.mov`

Representative ffprobe data:
- codec: `h264`
- size: `2880x1920`
- fps: `24/1`
- duration: about `6.29s`

Interpretation:
- plain MJPEG client fan-out matters somewhat in preview-only cases
- but in preview+recording, `0`, `1`, and `2` plain clients all stay in roughly the same `~159–165%` band
- this strongly suggests that **simple MJPEG HTTP-client fan-out alone is not the main explanation** for the user-observed `~400%` browser-path spike

### EXP-03: Real browser-backed desktop, one tab, preview only

Purpose:
- establish the real one-tab desktop preview baseline on the actual Studio page

Browser helper scripts used:
- `scripts/08-playwright-browser-matrix/01-open-studio-and-wait-desktop.js`
- `scripts/07-live-server-browser-scenario-sample.sh`

Saved result directory:
- `scripts/results/20260414-163610/`

Summary result:
- avg CPU: `14.20%`
- max CPU: `15.00%`

Saved metric-delta interpretation from first/last `.prom` snapshots:
- active clients at end: `display=1`
- frames served delta: `71`
- bytes served delta: `23,498,705`

Interpretation:
- the real browser one-tab preview-only case is not dramatically different from the plain preview baseline
- the browser does not look inherently catastrophic in preview-only mode for desktop-only

### EXP-04: Real browser-backed desktop, one tab, preview + recording

Purpose:
- measure the high-signal repro directly in the real Studio page

Browser helper scripts used:
- `scripts/08-playwright-browser-matrix/01-open-studio-and-wait-desktop.js`
- `scripts/08-playwright-browser-matrix/03-start-recording.js`
- `scripts/08-playwright-browser-matrix/04-stop-recording.js`
- `scripts/07-live-server-browser-scenario-sample.sh`

Saved result directory:
- `scripts/results/20260414-163951/`

Summary result:
- avg CPU: `410.60%`
- max CPU: `489.00%`

Saved metric-delta interpretation from first/last `.prom` snapshots:
- active clients at end: `display=1`
- frames served delta: `28`
- bytes served delta: `3,198,438`

Interpretation:
- this is the decisive run that proves the browser-connected hot slice is real
- the jump from the plain-client baseline (`158.56%`) to the real browser run (`410.60%`) is too large to dismiss as noise
- the relatively small frame/byte deltas during recording are an important clue: **the extra CPU is not explained simply by sending more JPEG bytes**

### EXP-05: Real browser-backed desktop, two tabs, preview only

Purpose:
- measure the effect of duplicate browser listeners in a preview-only desktop scenario

Browser helper scripts used:
- `scripts/08-playwright-browser-matrix/01-open-studio-and-wait-desktop.js`
- `scripts/08-playwright-browser-matrix/02-open-second-desktop-tab.js`
- `scripts/07-live-server-browser-scenario-sample.sh`

Saved result directory:
- `scripts/results/20260414-164457/`

Summary result:
- avg CPU: `12.69%`
- max CPU: `14.00%`

Metric-delta interpretation from first/last `.prom` snapshots:
- active clients at end: `display=2`
- frames served delta: `142`
- bytes served delta: `46,059,020`

Interpretation:
- two real tabs did double the served display frames in the saved metric deltas
- but preview-only CPU still remained modest
- this implies that duplicate browser listeners matter, but **preview-only duplication is not enough by itself to create the giant hot slice**

### EXP-06: Real browser-backed desktop, two tabs, preview + recording

Purpose:
- measure whether duplicate browser tabs amplify the already-hot desktop recording case

Browser helper scripts used:
- `scripts/08-playwright-browser-matrix/01-open-studio-and-wait-desktop.js`
- `scripts/08-playwright-browser-matrix/02-open-second-desktop-tab.js`
- `scripts/08-playwright-browser-matrix/03-start-recording.js`
- `scripts/08-playwright-browser-matrix/04-stop-recording.js`
- `scripts/07-live-server-browser-scenario-sample.sh`

Saved result directory:
- `scripts/results/20260414-164535/`

Summary result:
- avg CPU: `432.97%`
- max CPU: `571.00%`

Metric-delta interpretation from first/last `.prom` snapshots:
- active clients at end: `display=2`
- frames served delta: `57`
- bytes served delta: `6,146,580`

Interpretation:
- two tabs are hotter than one in the desktop recording case (`432.97%` vs `410.60%`)
- but the difference is much smaller than the jump from plain client baseline to one real browser tab
- this means duplicate tabs are **real amplification**, but **not the main explanation**

### EXP-07: Real browser-backed desktop + camera, one tab, preview only

Purpose:
- observe how a second active preview source changes the browser-backed baseline

Browser helper scripts used:
- `scripts/08-playwright-browser-matrix/05-add-camera-if-needed.js`
- `scripts/08-playwright-browser-matrix/06-capture-browser-preview-state.js`
- `scripts/07-live-server-browser-scenario-sample.sh`

Saved result directory:
- `scripts/results/20260414-164657/`

Saved browser state artifact:
- `scripts/11-browser-playwright-state-desktop-camera.json`

Summary result:
- avg CPU: `20.10%`
- max CPU: `28.00%`

Metric-delta interpretation from first/last `.prom` snapshots:
- active clients at end: `display=1`, `camera=1`
- display frames served delta: `71`
- display bytes served delta: `16,902,658`
- camera frames served delta: `64`
- camera bytes served delta: `9,204,346`

Interpretation:
- adding a second real preview source does raise preview-only CPU
- this is expected and useful context, but the browser-only desktop recording scenario still remains the cleanest main repro to dig into first

### EXP-08: Real browser-backed desktop + camera, one tab, preview + recording

Purpose:
- observe how the browser-backed hot slice behaves when both desktop and camera previews are live

Browser helper scripts used:
- `scripts/08-playwright-browser-matrix/05-add-camera-if-needed.js`
- `scripts/08-playwright-browser-matrix/03-start-recording.js`
- `scripts/08-playwright-browser-matrix/04-stop-recording.js`
- `scripts/07-live-server-browser-scenario-sample.sh`

Saved result directory:
- `scripts/results/20260414-164720/`

Summary result:
- avg CPU: `343.71%`
- max CPU: `411.00%`

Metric-delta interpretation from first/last `.prom` snapshots:
- active clients at end: `display=1`, `camera=1`
- display frames served delta: `27`
- display bytes served delta: `2,490,413`
- camera frames served delta: `43`
- camera bytes served delta: `2,544,617`

Interpretation:
- this remains very hot
- but it does not exceed the desktop-only two-tab case
- the result is useful context, yet it does not displace **desktop preview + recording + one browser tab** as the highest-value main repro

### EXP-09: Improved browser sampler validation with direct metric deltas

Purpose:
- validate the improved live browser sampler after it was extended to emit `metric-deltas.txt`

Harness:
- `scripts/07-live-server-browser-scenario-sample.sh`

Saved result directory:
- `scripts/results/20260414-165126/`

Representative result:
- avg CPU: `18.17%`
- max CPU: `24.00%`
- direct `metric-deltas.txt` emitted successfully

Important failure encountered before the fix:

```text
ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/07-live-server-browser-scenario-sample.sh: line 93: bad substitution: no closing "`" in `text
```

Cause:
- the script wrote a markdown fence of `````text````` inside an unquoted heredoc, which bash treated as command-substitution syntax

Fix:
- switch that block to `~~~text`
- rerun validation successfully

### EXP-10: Browser network-summary artifact

Purpose:
- preserve the browser-tool network view from the live scenarios as supporting evidence

Artifact:
- `scripts/10-browser-session-network-summary.txt`

Observed caveat:
- the browser tool’s saved network summary did not clearly expose the long-lived MJPEG GET stream lines even though preview images were definitely loaded
- for this pass, the server-side metrics were the more reliable evidence surface for active preview stream behavior

### EXP-11: Post-run preview-release cleanup observation

Purpose:
- note a relevant cleanup/runtime observation encountered after the browser-backed runs

Observation:
- after closing browser tabs, `/api/previews` still showed active preview leases in at least one post-run cleanup check
- manual `POST /api/previews/release` calls were needed to return to an empty preview list in that cleanup sequence

Caveat:
- this is **not yet fully isolated** as a product bug versus a browser-tool/session artifact
- it is worth keeping in mind because stale browser listeners and stale preview ownership are central to this ticket’s theme

### EXP-12: Desktop preview + recording MJPEG-vs-MJPEG+WS ablation

Purpose:
- isolate how much of the desktop preview+recording cost can be explained by adding a plain websocket consumer on top of one MJPEG preview client
- test the then-current hypothesis that websocket/event fanout might explain a large part of the browser-vs-curl gap

Harness:
- `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/run.sh`
- `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/ws_client/main.go`

Command used:

```bash
DURATION=6 bash ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/run.sh
```

Saved result directory:
- `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173541/`

Scenarios:
- `mjpeg-only`
- `mjpeg-plus-ws`

Results:

| Scenario | Avg CPU | Max CPU |
| --- | ---: | ---: |
| mjpeg-only | 166.56% | 469.00% |
| mjpeg-plus-ws | 170.48% | 481.00% |

Key metric deltas from `mjpeg-plus-ws`:
- `preview.state` published: `33`
- `preview.state` delivered: `23`
- websocket `preview.state` written: `23`
- websocket total messages observed by the synthetic client: `54`

Saved websocket client summary:
- `mjpeg-plus-ws/ws-client-summary.json`

Important harness bug encountered and fixed before trusting the result:

- The first run under `results/20260414-173359/` incorrectly started recording too early in the `mjpeg-only` case, before the preview path had clearly produced an initial frame.
- That made the first `11.00%` `mjpeg-only` number invalid.
- I fixed the harness by waiting for an initial preview screenshot before starting the measurement in both scenarios, then reran the matrix successfully.

Interpretation:
- websocket/event fanout is **real**, measurable, and now directly visible in metrics
- but adding a plain `/ws` consumer on top of one MJPEG client only increased average CPU by about `3.92` percentage points in this focused fresh-server ablation (`166.56%` → `170.48%`)
- that is far too small to explain the jump from `~158–166%` fresh-server MJPEG cases to `~410%` real one-tab browser recording
- this materially lowers confidence that websocket fanout by itself is the dominant explanation for the browser-path spike

### INST-01: Added MJPEG handler timing metrics

Purpose:
- instrument the browser-facing MJPEG handler before changing behavior again
- prepare the next one-tab real-browser rerun so it can answer whether the hot path is dominated by MJPEG write/flush time or by something else

Code commit:
- `9fd8754ab4db6aab3ce0bd174c2ac006d957b1dd` — `Add MJPEG handler timing metrics`

Files changed:
- `internal/web/preview_metrics.go`
- `internal/web/handlers_preview.go`
- `internal/web/metrics_test.go`
- `internal/web/server_test.go`

New metrics added:
- `screencast_studio_preview_http_loop_iterations_total`
- `screencast_studio_preview_http_idle_iterations_total`
- `screencast_studio_preview_http_write_nanoseconds_total`
- `screencast_studio_preview_http_flush_nanoseconds_total`

Validation:

```bash
gofmt -w internal/web/preview_metrics.go internal/web/handlers_preview.go internal/web/metrics_test.go internal/web/server_test.go
go test ./internal/web ./pkg/metrics -count=1
go test ./... -count=1
```

Interpretation:
- this is an instrumentation-only step, not a new runtime result yet
- it is the correct next move after the websocket ablation because it measures before it perturbs
- the next real-browser desktop preview+recording rerun should now make it possible to compare:
  - total write nanoseconds,
  - total flush nanoseconds,
  - total loop iterations,
  - total idle iterations,
  against the earlier browser and synthetic-client cases

### EXP-13: Real browser-backed desktop preview + recording rerun with MJPEG timing metrics

Purpose:
- rerun the highest-value browser scenario after the new MJPEG timing metrics landed
- answer whether the browser-path spike appears to live inside MJPEG write/flush work itself

Browser helper scripts used:
- `scripts/08-playwright-browser-matrix/01-open-studio-and-wait-desktop.js`
- `scripts/08-playwright-browser-matrix/03-start-recording.js`
- `scripts/08-playwright-browser-matrix/04-stop-recording.js`
- `scripts/07-live-server-browser-scenario-sample.sh`

Saved result directory:
- `scripts/results/20260414-202519/`

Summary result:
- avg CPU: `163.34%`
- max CPU: `407.92%`

Important comparability caveat:
- this rerun is **not** directly comparable to the earlier `410.60% avg CPU` browser run as a pure average, because the sampled window clearly included several preview-only seconds before recording fully ramped up
- the per-second `pidstat` trace shows a low preview-only band first, then a rapid climb during recording, including `377.00%` and `407.92%` late in the sampled window
- this rerun is therefore more useful for **timing-metric interpretation** than for replacing the earlier average-CPU headline number

Key metric deltas:
- `preview_http_frames_served_total`: `55`
- `preview_http_bytes_served_total`: `10,503,752`
- `preview_http_loop_iterations_total`: `71`
- `preview_http_idle_iterations_total`: `16`
- `preview_http_write_nanoseconds_total`: `8,724,144`
- `preview_http_flush_nanoseconds_total`: `908,218`
- `websocket_events_written_total{event_type="preview.state"}`: `54`
- `websocket_events_written_total{event_type="telemetry.audio_meter"}`: `62`

Interpretation:
- the new timing counters are the most important part of this rerun
- cumulative MJPEG write time across the sampled run was only about `8.7ms` total, and cumulative flush time was only about `0.9ms` total
- on a rough per-served-frame basis, that is approximately:
  - write time: `~0.159ms/frame`
  - flush time: `~0.017ms/frame`
- those numbers are **far too small** to explain a server process that still reached `~378–408%` CPU late in the run
- this strongly lowers confidence that the main browser-path cost is the HTTP MJPEG write/flush loop itself
- the missing hot slice now looks even more likely to be upstream of the final HTTP write path: shared preview+recording interaction, Go-side frame-copy/publication work before the final write, browser-connected lifecycle behavior, or some combination of those

### EXP-14: Real browser-backed desktop preview + recording rerun with preview-manager and EventHub timing metrics

Purpose:
- move one step upstream of the final HTTP write path
- measure whether PreviewManager cached-frame copying, `preview.state` publication, or EventHub publish time is large enough to explain the browser-path hot phase better than the final MJPEG write loop did

Browser helper scripts used:
- `scripts/08-playwright-browser-matrix/01-open-studio-and-wait-desktop.js`
- `scripts/08-playwright-browser-matrix/03-start-recording.js`
- `scripts/08-playwright-browser-matrix/04-stop-recording.js`
- `scripts/07-live-server-browser-scenario-sample.sh`

Saved result directory:
- `scripts/results/20260414-203319/`

Summary result:
- avg CPU: `150.70%`
- max CPU: `394.00%`

Important comparability caveat:
- just like EXP-13, this sampled window included preview-only warmup before the recording phase fully ramped, so the average CPU number is not a replacement for the earlier `410.60%` one-tab browser headline
- the important value of this rerun is the new upstream timing deltas plus the late-run climb into the hot band (`312–394%`)

Key timing deltas:
- `preview_http_write_nanoseconds_total`: `7,028,353`
- `preview_http_flush_nanoseconds_total`: `511,999`
- `preview_frame_store_nanoseconds_total`: `3,691,230`
- `preview_latest_frame_copy_nanoseconds_total`: `1,796,290`
- `preview_state_publish_nanoseconds_total`: `790,719`
- `eventhub_publish_nanoseconds_total{event_type="preview.state"}`: `535,814`
- `eventhub_publish_nanoseconds_total{event_type="telemetry.audio_meter"}`: `6,452,838`

Approximate per-event/per-frame interpretation:
- MJPEG write time per served frame: `~0.117ms`
- MJPEG flush time per served frame: `~0.009ms`
- PreviewManager frame-store time per frame update: `~0.062ms`
- PreviewManager latest-frame copy time per served frame: `~0.030ms`
- PreviewManager `preview.state` publish time per frame update: `~0.013ms`
- EventHub `preview.state` publish time per event: `~0.009ms`
- EventHub `telemetry.audio_meter` publish time per event: `~0.143ms`

Interpretation:
- this rerun narrows the hypothesis again
- not only is the final MJPEG write/flush loop too cheap to explain the hot phase, but the immediate Go-side work just upstream of it is also tiny:
  - PreviewManager frame-store/copy/publication work is measured in **hundredths of a millisecond per event/frame**
  - EventHub `preview.state` publish time is also tiny
- even the largest newly visible event-path cost here, `telemetry.audio_meter` EventHub publish time, is still only about `6.45ms` cumulative across the whole sampled run
- those numbers are nowhere near large enough to explain a process that still climbed into the `~312–394%` range late in the run
- the dominant unexplained browser-path cost now looks even more likely to live **upstream of PreviewManager itself**, likely around the shared GStreamer preview consumer/appsink callback boundary, the transition from GStreamer buffers into Go JPEG byte slices, or some browser-connected interaction that those new counters still do not see

## Aggregated result tables

### Fresh-server MJPEG-client matrix

| Scenario | Avg CPU |
| --- | ---: |
| preview-no-client | 15.67% |
| preview-one-client | 18.11% |
| preview-two-clients | 19.22% |
| record-no-client | 162.22% |
| record-one-client | 158.56% |
| record-two-clients | 165.00% |

### Real browser-backed matrix

| Scenario | Avg CPU |
| --- | ---: |
| desktop, 1 tab, preview only | 14.20% |
| desktop, 1 tab, preview + recording | 410.60% |
| desktop, 2 tabs, preview only | 12.69% |
| desktop, 2 tabs, preview + recording | 432.97% |
| desktop + camera, 1 tab, preview only | 20.10% |
| desktop + camera, 1 tab, preview + recording | 343.71% |

### Focused MJPEG-vs-websocket ablation

| Scenario | Avg CPU |
| --- | ---: |
| mjpeg-only | 166.56% |
| mjpeg-plus-ws | 170.48% |

## Current working interpretation

### Strongest conclusion

The **real browser-connected recording path** is the missing hot slice.

The best side-by-side comparison remains:

- fresh server + desktop recording + one plain MJPEG client → `158.56%` avg CPU
- live Studio page + desktop recording + one real browser tab → `410.60%` avg CPU

### What is probably *not* the whole explanation

It is probably **not** just:
- “more JPEG bytes were sent”
- or “more MJPEG clients were attached”

Why:
- recording runs with real browser tabs showed modest per-run served-frame and served-byte deltas compared with preview-only runs
- two tabs are hotter than one, but the much bigger jump is from plain client baseline to one real browser tab

### Code-backed explanation that currently looks strongest

The real browser path activates additional server work that the plain-client baseline did not pay for.

#### 1. MJPEG serving loop

`internal/web/handlers_preview.go`:
- `handlePreviewMJPEG` polls every `100ms`
- on a new frame it writes multipart boundaries + headers + JPEG payload + CRLF
- increments preview-serving metrics
- flushes after each frame

This is real cost, but the result data suggests it is not the full explanation by itself.

#### 2. Multiple Go-side frame copies

`pkg/media/gst/shared_video.go` and `internal/web/preview_manager.go` show that preview frames are copied several times:
- GStreamer appsink callback calls `consumer.setLatestFrame(frame)`
- `setLatestFrame` copies into `sharedPreviewConsumer.latestFrame`
- `PreviewManager.storePreviewFrame` copies again into `managedPreview.latestFrame`
- `PreviewManager.LatestFrame` copies again for the HTTP handler
- then `handlePreviewMJPEG` writes the frame to the client

This makes the preview path allocation/copy-heavy even before considering browser-side behavior.

#### 3. Browser-only websocket event traffic

This was the strongest suspect before the focused websocket ablation, but the latest results reduce confidence that it is the dominant explanation by itself.

`internal/web/preview_manager.go` currently does:
- `m.publishPreviewState(snapshot)` on every stored preview frame update

That sends `preview.state` events into:
- `internal/web/event_hub.go`
- `internal/web/handlers_ws.go`
- and the real browser client in `ui/src/features/session/wsClient.ts`

The real browser path also receives recording/session telemetry over `/ws`, including audio-meter updates during recording from `internal/web/session_manager.go`.

However, the focused fresh-server ablation that compared one MJPEG client versus one MJPEG client plus one synthetic websocket consumer only moved avg CPU from `166.56%` to `170.48%`. That means websocket/event fanout is a **real contributor**, but it does **not** appear large enough by itself to explain the jump to `~410%` in the real browser-tab run.

### Concise current hypothesis

The `~400%` desktop preview+recording jump is probably explained by a **combination** of:

1. expensive upstream shared preview + recording interaction,
2. MJPEG preview serving and flush work,
3. multiple Go-side preview frame copies,
4. browser-specific behavior that is still not reproduced by a plain MJPEG-plus-websocket synthetic client,
5. and possibly some websocket/event-path contribution, but likely not enough by itself.

## Why desktop preview + recording is the best main repro now

This scenario is good enough that we do **not** need broader scenario expansion before going deeper.

Reasons:
- it is the cleanest browser-backed hot slice already measured
- it isolates the browser-vs-plain-client gap clearly
- it avoids the additional interpretive noise of camera-specific behavior
- it is already sufficient to justify deeper instrumentation and A/B tests

## Recommended next experiments

Keep using the same high-signal scenario:
- **desktop preview + recording + one real browser tab**

Then do focused A/B isolation:

1. **Measure with browser tab connected but websocket disabled**
   - keep MJPEG, remove `/ws` load if possible
   - estimate how much of the gap is websocket/event traffic

2. **Throttle or suppress per-frame `preview.state` publication**
   - this is the strongest current server-side suspect

3. **Temporarily disable recording audio-meter websocket publication**
   - isolate its contribution to the browser-only path

4. **Add deeper server metrics in `handlePreviewMJPEG`**
   - write time
   - flush time
   - blocked write counts
   - loop iterations with no new frame vs with new frame

5. **Optionally compare visible tab vs headless browser consumer**
   - only if needed after the server-side event-path isolation work

## Raw result locations

### Ticket-local scripts and summaries

- `scripts/03-desktop-preview-http-client-matrix/run.sh`
- `scripts/04-desktop-preview-http-client-baseline-summary.md`
- `scripts/05-desktop-preview-http-client-recording-matrix/run.sh`
- `scripts/07-live-server-browser-scenario-sample.sh`
- `scripts/08-playwright-browser-matrix/`
- `scripts/09-browser-preview-matrix-findings-summary.md`
- `scripts/10-browser-session-network-summary.txt`
- `scripts/11-browser-playwright-state-desktop-camera.json`

### Result directories

- `scripts/results/20260414-160358/`
- `scripts/03-desktop-preview-http-client-matrix/results/20260414-161024/`
- `scripts/05-desktop-preview-http-client-recording-matrix/results/20260414-163154/`
- `scripts/results/20260414-163610/`
- `scripts/results/20260414-163951/`
- `scripts/results/20260414-164457/`
- `scripts/results/20260414-164535/`
- `scripts/results/20260414-164657/`
- `scripts/results/20260414-164720/`
- `scripts/results/20260414-165126/`

## Review guidance

If you are continuing this investigation, start in this order:

1. `reference/02-browser-preview-streaming-lab-report.md`
2. `scripts/09-browser-preview-matrix-findings-summary.md`
3. `design/02-browser-preview-streaming-performance-report.md`
4. the actual runtime paths:
   - `internal/web/handlers_preview.go`
   - `internal/web/preview_manager.go`
   - `internal/web/handlers_ws.go`
   - `internal/web/event_hub.go`
   - `internal/web/session_manager.go`
   - `pkg/media/gst/shared_video.go`

## Bottom line

At this point, the investigation has already done enough to justify a narrow next phase:

> Stop broadening scenarios for now and dig deeper into **desktop preview + recording + one real browser tab**.

That scenario already reproduces the real browser-connected hot slice cleanly enough to guide the next round of instrumentation and A/B testing.
