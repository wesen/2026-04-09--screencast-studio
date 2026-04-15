---
Title: Browser preview streaming performance report
Ticket: SCS-0015
Status: draft
Topics:
    - screencast-studio
    - gstreamer
    - video
    - backend
    - ui
    - analysis
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/handlers_preview.go
      Note: New MJPEG timing metrics now support the next high-signal rerun
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/05-desktop-preview-http-client-recording-matrix/run.sh
      Note: Fresh-server preview and recording MJPEG-client matrix harness
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/07-live-server-browser-scenario-sample.sh
      Note: Live browser-backed server sampler used for the real Studio-page measurements
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/09-browser-preview-matrix-findings-summary.md
      Note: Human-readable first findings summary for the larger matrix pass
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173541/01-summary.md
      Note: Focused websocket ablation summary that refined the browser-path hypothesis
ExternalSources: []
Summary: First substantive report draft for the browser preview streaming investigation, based on the fresh-server HTTP-client matrix and the first live browser-backed measurement pass.
LastUpdated: 2026-04-14T18:08:00-04:00
WhatFor: Hold the matrix results and engineering conclusions for SCS-0015 as the browser-connected preview investigation progresses.
WhenToUse: Read after the first measurement pass to understand what is already proven, what remains pending, and what optimization directions are now justified.
---




# Browser preview streaming performance report

## Status

This report is now partially populated from the first real measurement pass. It is still a draft because camera-only browser scenarios and deeper handler-path instrumentation are not finished yet.

## Executive summary

The strongest current SCS-0015 finding is that the **real browser-connected recording path** is dramatically hotter than the fresh-server plain-MJPEG-client baseline.

The clearest side-by-side comparison is:

- fresh dedicated server, desktop preview + recording + one plain MJPEG client → `158.56%` avg CPU
- live shared `:7777` server, desktop preview + recording + one real browser tab → `410.60%` avg CPU
- live shared `:7777` server, desktop preview + recording + two real browser tabs → `432.97%` avg CPU

This means the missing hot slice was real: earlier API-only and curl-like measurements were not enough to explain the browser-connected Studio-page behavior the user reported.

At the same time, the browser-backed recording runs did **not** show proportionally huge MJPEG frame/byte deltas. A newer focused ablation also showed that adding a plain websocket consumer on top of one MJPEG client only moved avg CPU from `166.56%` to `170.48%`. That means the browser-path heat is not explained simply by “the server had to send vastly more JPEG bytes,” and websocket/event fanout alone is also insufficient. The current evidence points more toward a combination of:

- expensive upstream preview + recording interaction,
- browser-connected consumer behavior that differs materially from a dumb `curl` reader,
- and browser-specific lifecycle/rendering/stream-consumer behavior that is still not reproduced by a plain MJPEG-plus-websocket synthetic client.

## Measurement environment

- Repository: `/home/manuel/code/wesen/2026-04-09--screencast-studio`
- Date: `2026-04-14`
- Dedicated fresh-server matrix used temporary built binaries on fresh ports.
- Live browser-backed runs used the real local server on `http://127.0.0.1:7777`.
- CPU measurement used `pidstat` against the real listening server PID.
- Metrics sampling used the ticket-local `/metrics` sampler.
- Browser interaction used the ticket-local Playwright code snippets executed through the browser tool.

## Scenario matrix

### Fresh dedicated-server HTTP-client matrix

Saved under:

- `scripts/05-desktop-preview-http-client-recording-matrix/results/20260414-163154/`

Measured scenarios:

| Scenario | Avg CPU |
| --- | ---: |
| preview-no-client | 15.67% |
| preview-one-client | 18.11% |
| preview-two-clients | 19.22% |
| record-no-client | 162.22% |
| record-one-client | 158.56% |
| record-two-clients | 165.00% |

### Live browser-backed scenarios

Saved under:

- `scripts/results/20260414-163610/` — desktop, one tab, preview only
- `scripts/results/20260414-163951/` — desktop, one tab, preview + recording
- `scripts/results/20260414-164457/` — desktop, two tabs, preview only
- `scripts/results/20260414-164535/` — desktop, two tabs, preview + recording
- `scripts/results/20260414-164657/` — desktop + camera, one tab, preview only
- `scripts/results/20260414-164720/` — desktop + camera, one tab, preview + recording

Measured scenarios:

| Scenario | Avg CPU |
| --- | ---: |
| desktop, one tab, preview only | 14.20% |
| desktop, one tab, preview + recording | 410.60% |
| desktop, two tabs, preview only | 12.69% |
| desktop, two tabs, preview + recording | 432.97% |
| desktop + camera, one tab, preview only | 20.10% |
| desktop + camera, one tab, preview + recording | 343.71% |

## Raw result locations

### Script bundles

- `scripts/05-desktop-preview-http-client-recording-matrix/run.sh`
- `scripts/07-live-server-browser-scenario-sample.sh`
- `scripts/08-playwright-browser-matrix/`

### Human-readable summary artifacts

- `scripts/09-browser-preview-matrix-findings-summary.md`
- `scripts/10-browser-session-network-summary.txt`
- `scripts/11-browser-playwright-state-desktop-camera.json`
- `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173541/01-summary.md`

## Key findings

### 1. Plain MJPEG client fan-out is not enough to explain the browser-path spike

The fresh-server recording results stayed in a narrow band:

- `record-no-client` → `162.22%`
- `record-one-client` → `158.56%`
- `record-two-clients` → `165.00%`

That means simple HTTP-client fan-out alone does not explain the much hotter real browser runs.

### 2. The browser-connected recording path is the missing hot slice

The desktop one-tab recording run on the real Studio page reached `410.60%` avg CPU, and the desktop two-tab recording run reached `432.97%` avg CPU.

That is a major gap relative to the fresh-server plain-client matrix and matches the earlier user report that the real web UI could push the server into a much hotter regime.

### 3. More browser tabs amplify the hot slice, but they do not fully explain it

Going from one browser tab to two browser tabs raised desktop preview+recording from `410.60%` to `432.97%` avg CPU. That matters, but the bigger story is that even **one** real tab is already far hotter than the plain-client baseline.

### 4. The browser recording runs were hot even without huge MJPEG-byte deltas

Per-run deltas computed from the first and last saved metric snapshots showed that browser recording runs served relatively modest numbers of frames and bytes compared with preview-only runs.

Examples:

- desktop, one tab, preview only:
  - frames served delta: `71`
  - bytes served delta: `23,498,705`
- desktop, one tab, preview + recording:
  - frames served delta: `28`
  - bytes served delta: `3,198,438`

That means the browser-path heat is not explained simply by larger MJPEG output volume.

### 5. A synthetic websocket consumer is not enough to reproduce the browser spike

The focused fresh-server ablation at `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173541/` compared:

- one MJPEG client only → `166.56%` avg CPU
- one MJPEG client plus one synthetic websocket consumer → `170.48%` avg CPU

That websocket-enabled case did generate real websocket/event traffic, including `preview.state` deliveries and websocket writes, but the CPU increase was only about `3.92` points. This materially lowers confidence that websocket/event fanout by itself is the dominant explanation for the real browser-tab result.

### 6. Desktop + camera changes the mix but does not invalidate the browser finding

Desktop + camera one-tab preview-only was hotter than desktop-only preview-only (`20.10%` vs `14.20%`).

Desktop + camera one-tab preview+recording reached `343.71%`, which is still extremely hot even though it did not exceed the desktop-only two-tab recording case.

## Comparison against earlier SCS-0014 backend-focused measurements

Earlier SCS-0014 work already showed that preview + recorder interaction on the shared source was expensive. But the most representative backend/API-only and curl-like measurements still did not show the same scale as the real browser path.

The new SCS-0015 matrix explains why the user’s `~400%` CPU report did not line up with the earlier isolated results:

- the backend-only work was measuring an important part of the system,
- but the **browser-connected** path remained under-measured,
- and that path really does enter a much hotter server CPU regime when recording is started from the Studio page.

## Metrics interpretation

The current preview-serving metrics are useful, but there is one important caveat:

- the counters are cumulative across the live server lifetime,
- so per-run interpretation depends on comparing the first and last snapshot inside a result directory.

The improved browser sampler now emits `metric-deltas.txt` directly for newer runs. For the earlier browser-backed runs in this pass, the deltas were computed manually from the saved `.prom` snapshots.

The most helpful current metrics are:

- `screencast_studio_preview_http_clients`
- `screencast_studio_preview_http_frames_served_total`
- `screencast_studio_preview_http_bytes_served_total`
- `screencast_studio_preview_http_flushes_total`
- `screencast_studio_preview_http_loop_iterations_total`
- `screencast_studio_preview_http_idle_iterations_total`
- `screencast_studio_preview_http_write_nanoseconds_total`
- `screencast_studio_preview_http_flush_nanoseconds_total`

These newer timing metrics were added after the websocket-ablation pass so the next real browser rerun can answer a narrower question: is the browser-specific hot slice spending its extra time inside the MJPEG write/flush loop, or is the real browser gap still mostly elsewhere?

## Likely bottlenecks

Current best interpretation:

1. **Upstream preview + recording interaction is still a major cost center.**
2. **Real browser-connected preview consumption is materially different from a dumb HTTP reader.**
3. **Simple MJPEG fan-out is only part of the story.**
4. **Synthetic websocket/event fanout alone is also only part of the story.**
5. **Served-byte volume is not the whole explanation.**

That points away from a simplistic “just lower JPEG quality” answer and toward deeper investigation of the browser-facing preview loop and the browser-connected shared-source workload.

## Recommended optimizations

Ranked by current confidence and likely value:

1. **Use the newly added MJPEG timing metrics in a real browser rerun**
   - compare write/flush/loop deltas in the high-signal one-tab desktop preview+recording scenario,
   - then decide whether deeper handler instrumentation is still necessary.
2. **If needed, add even deeper handler/path instrumentation**
   - per-stream skip/drop reasons,
   - blocked write/flush timing,
   - maybe frame-age or stale-frame counters.
3. **Rerun the real browser-tab desktop preview+recording case with the new event/websocket metrics enabled**
   - this is now higher priority than broadening the scenario matrix.
3. **Compare visible browser tabs with a more minimal/headless browser consumer**
   - to separate UI-rendering effects from server-side streaming behavior.
4. **Continue tuning the preview profile during recording**
   - but do not assume JPEG byte reduction alone will solve the full browser-path spike.

## Risks and open questions

- Camera-only browser scenarios are still missing.
- The browser tool’s request summary did not directly expose the long-lived MJPEG GET lines in the saved network note, so server-side metrics were the more reliable proof surface in this pass.
- We still need a more explicit model of where the extra server CPU time is being spent inside the browser-connected recording path.
