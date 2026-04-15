---
Title: Investigation diary
Ticket: SCS-0015
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - video
    - backend
    - ui
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/handlers_preview.go
      Note: |-
        Browser-facing MJPEG streaming handler is the main backend boundary for this ticket
        Main backend boundary for the new browser streaming investigation
        MJPEG handler timing metrics added in Step 8 (commit 9fd8754ab4db6aab3ce0bd174c2ac006d957b1dd)
    - Path: internal/web/preview_manager.go
      Note: |-
        Preview lifecycle, reuse, and cached-frame fan-out behavior shape the browser streaming path
        Preview cache and fan-out behavior shaped the first-step reasoning
    - Path: internal/web/preview_metrics.go
      Note: |-
        Metric-family definitions were added in the first code slice
        Preview timing metric families added in Step 8 (commit 9fd8754ab4db6aab3ce0bd174c2ac006d957b1dd)
    - Path: internal/web/server_test.go
      Note: End-to-end MJPEG metrics visibility coverage extended in Step 8
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/reference/02-browser-preview-streaming-lab-report.md
      Note: New ongoing lab report created and backfilled in Step 6
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/01-restart-scs-web-ui.sh
      Note: Added and validated in the second helper-script slice
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/02-sample-preview-metrics.sh
      Note: Added and validated in the second helper-script slice
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/03-desktop-preview-http-client-matrix/run.sh
      Note: Added and validated as the first matrix harness slice
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/04-desktop-preview-http-client-baseline-summary.md
      Note: Human-readable interpretation of the first 0/1/2-client desktop baseline
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/05-desktop-preview-http-client-recording-matrix/run.sh
      Note: Added and run for the broader fresh-server matrix
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/07-live-server-browser-scenario-sample.sh
      Note: Used for all real browser-backed server measurements
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/08-playwright-browser-matrix/05-add-camera-if-needed.js
      Note: Used to create the desktop-plus-camera browser scenarios
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/09-browser-preview-matrix-findings-summary.md
      Note: Saved the first larger findings note after the matrix run
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/run.sh
      Note: Added and used during Step 7
    - Path: ui/src/components/preview/PreviewStream.tsx
      Note: |-
        The actual browser media transport is an img tag pointed at the MJPEG endpoint
        Proof that the browser path is plain img-plus-MJPEG rather than a richer transport
    - Path: ui/src/pages/StudioPage.tsx
      Note: |-
        Frontend preview ensure/release lifecycle likely explains differences between API-only and browser-driven runs
        Frontend preview lifecycle likely explains differences between browser-driven and API-only runs
ExternalSources: []
Summary: Diary for the browser preview streaming and web-UI performance investigation ticket.
LastUpdated: 2026-04-14T18:08:00-04:00
WhatFor: Record the chronology of how the browser preview streaming performance investigation was set up and how its conclusions are derived.
WhenToUse: Read when continuing SCS-0015 and you need the exact reasoning, commands, and file boundaries used to define the new browser-streaming research track.
---









# Diary

## Goal

Capture the setup and evidence-backed planning for investigating the browser preview streaming path, especially the CPU gap between backend/API-only performance experiments and the much hotter real Studio page behavior.

## Step 1: Created a dedicated ticket for the browser-connected preview path

The previous performance work in SCS-0014 narrowed the strongest remaining cost center to the combined preview-plus-recorder case, but the user then reported an even hotter effect in the real web UI: server CPU shooting much higher when recording is started from the Studio page, including cases where desktop-only recording still runs much hotter through the browser path than earlier isolated matrices suggested. That was enough evidence to split browser streaming into its own dedicated ticket instead of continuing to overload SCS-0014.

I created SCS-0015 specifically to investigate the browser preview transport and the real web-UI measurement gap. Since the user explicitly said they had closed the server while this setup work happened, I kept the first step offline-friendly: ticket scaffolding, current-state architecture mapping, and a concrete performance-matrix plan rather than pretending we could already produce runtime measurements.

### Prompt Context

**User prompt (verbatim):** "Ok, let's create a new ticket to investigate the browser streaming part of the pipeline, and do a performance matrix analysis just as we did before, with a plan/analysis document, standalone experiments, and a proper report."

**Assistant interpretation:** Create a new docmgr ticket focused on the browser-connected preview streaming path, with the same rigor as the earlier performance tickets: design doc, diary, scripts, measurements, and a report.

**Inferred user intent:** Separate the unexplained browser/UI-specific CPU spike into its own investigation track so it can be measured and optimized systematically rather than mixed into the older recorder-focused work.

### What I did

- Ran:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
docmgr status --summary-only
docmgr ticket create-ticket --ticket SCS-0015 --title "Investigate browser preview streaming pipeline and web UI performance matrix" --topics screencast-studio,gstreamer,video,backend,ui,analysis
```

- Confirmed the created workspace path:

```text
/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/
```

- Wrote the new primary design doc:
  - `design/01-browser-preview-streaming-pipeline-analysis-and-performance-matrix-plan.md`
- Wrote this diary:
  - `reference/01-investigation-diary.md`
- Planned a future proper report as part of the tasks/checklist rather than pretending it already exists.

### Why

The earlier tickets already did a strong job investigating backend preview/recording interaction. What was still missing was the explicit browser-connected path:

- the MJPEG HTTP stream,
- the `<img>` rendering model,
- preview ensure/release behavior driven by the Studio page,
- and possible amplification from stale or duplicate browser tabs.

That is different enough from backend-only measurement that it deserves its own deliverable and experiment suite.

### What worked

- The new ticket workspace was created successfully.
- The user’s clarified scope was specific enough that the first design doc could be evidence-based even before restarting the server.
- The existing codebase already contains strong file boundaries for this investigation: preview handler, preview manager, GStreamer preview branch, frontend ensure/release logic, and preview rendering.

### What didn't work

Immediately after ticket creation, `docmgr doc add --ticket SCS-0015 ...` failed with:

```text
Error: failed to find ticket directory: ticket not found: SCS-0015
```

The workspace itself had been created correctly, so I continued by writing the initial docs directly into the created ticket path.

### What I learned

The browser-streaming investigation can be scoped much more cleanly now than it could earlier. We already know enough to state a concrete hypothesis:

- the project has good backend/shared-runtime measurements,
- but it does not yet have a proper matrix for **browser-attached MJPEG preview clients**,
- and that missing slice is probably why the real Studio page still surprises us.

### What was tricky to build

The tricky part in this setup step was resisting the urge to over-claim. The user reported real browser-path CPU spikes, but with the server intentionally shut down during setup, it would have been misleading to pretend we had already confirmed the new hypothesis with fresh measurements. The right move here was to create a continuation-friendly investigation package first, then restart the runtime later for carefully staged experiments.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/design/01-browser-preview-streaming-pipeline-analysis-and-performance-matrix-plan.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`

The main review question is whether the new ticket scope is correctly isolating the browser-connected delta rather than drifting back into generic recorder optimization.

### What should be done in the future

- Extend `/metrics` with browser-preview-serving metrics.
- Build a browser-driven matrix harness under this ticket’s `scripts/` directory.
- Compare no-browser, one-tab, and multi-tab runs while holding backend capture/recording constant.
- Write the final report after results exist.

### Code review instructions

Start with the new design doc, then verify the main browser-path evidence in:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/preview/PreviewStream.tsx`

### Technical details

The key current-state evidence gathered for the new ticket includes:

- `internal/web/handlers_preview.go:79-149` — MJPEG is served as `multipart/x-mixed-replace` with a `100ms` polling ticker.
- `internal/web/preview_manager.go:95-220` — PreviewManager caches the latest JPEG frame and exposes it to HTTP readers.
- `pkg/media/gst/shared_video.go:492-560` — The shared preview branch already JPEG-encodes frames before they reach Go.
- `ui/src/pages/StudioPage.tsx:531-592` — The Studio page ensures and releases previews based on active desired sources.
- `ui/src/components/preview/PreviewStream.tsx:79-95` — The browser consumes previews via plain `<img src="/api/previews/{id}/mjpeg">`.

## Step 2: Added the First Browser-Preview Metrics Slice and Committed It

Once the ticket scope was clear, I started with observability rather than with another blind benchmark run. The browser-streaming question needs time-series evidence, so the first code slice was to make `/metrics` materially more useful for the preview-serving path. That meant both extending the tiny metrics package itself and wiring preview-serving counters/gauges into the real HTTP MJPEG handler and PreviewManager.

I also used this slice to fold the initial metrics foundation into a clean commit, because the earlier runtime metrics changes were still only local modifications. After the code was in place, I ran focused tests plus full repo tests and committed the slice as one code change before touching ticket bookkeeping.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Begin the new browser-streaming investigation incrementally, keep the work grouped into reviewable slices, and keep the diary and ticket state synchronized as code lands.

**Inferred user intent:** Build the browser-preview investigation the same way the earlier performance work was built: one concrete measurement-enabling step at a time, with real commits and continuation-friendly documentation.

**Commit (code):** `1c99094caf1a3661562c26332b2e57fd257de2a4` — `Add preview serving and runtime metrics`

### What I did

- Extended `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/metrics/metrics.go` to support both counters and gauges.
- Added `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/metrics/metrics_test.go` to verify mixed counter/gauge rendering.
- Added `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_metrics.go` defining the first browser-preview metric families.
- Wired the MJPEG handler in `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go` to track:
  - active HTTP clients,
  - stream starts,
  - stream finishes by reason,
  - frames served,
  - bytes served,
  - flush count.
- Wired `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go` to track:
  - preview ensure results,
  - preview release results,
  - preview frame updates.
- Kept labels intentionally low-cardinality:
  - `source_type`
  - bounded `result`
  - bounded `reason`
- Validated with:

```bash
gofmt -w pkg/metrics/metrics.go pkg/metrics/metrics_test.go internal/web/preview_metrics.go internal/web/handlers_preview.go internal/web/preview_manager.go internal/web/metrics_test.go internal/web/handlers_metrics.go internal/web/routes.go pkg/media/gst/recording.go pkg/media/gst/shared_video_recording_bridge.go
go test ./pkg/metrics ./internal/web ./pkg/media/gst -count=1
go test ./... -count=1
```

### Why

Before building the next browser-driven performance matrix, I needed a better measurement surface than raw `pidstat` alone. Counters and gauges tied to the preview-serving path let us ask better questions later:

- how many MJPEG clients are actually attached,
- how many frames were served,
- how many bytes were written,
- and whether the browser path is hotter because of preview fan-out, not just because recording is active.

### What worked

- The minimal metrics registry was easy to extend to gauges without introducing external dependencies.
- The browser-serving metric families fit naturally into the MJPEG handler and PreviewManager boundaries.
- The entire repo test suite still passed after the instrumentation.

### What didn't work

Nothing materially failed in this slice after implementation. The main non-code complication was that the metrics foundation changes already existed locally and needed to be folded into a clean commit rather than left drifting uncommitted.

### What I learned

The first metrics slice is enough to make Prometheus/Grafana-style investigation realistic. We still cannot answer every browser-path question yet, but we now have the minimum observability needed to correlate CPU with:

- active preview clients,
- preview-serving volume,
- and preview lifecycle events.

### What was tricky to build

The sharp edge here was **label discipline**. It would have been easy to emit per-preview or per-source IDs because those are available in the code, but that would make the new metrics much harder to reason about and potentially much noisier in Prometheus. I deliberately kept the first slice to `source_type` plus small bounded `reason`/`result` enums so the metrics stay useful for plotting and aggregation.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/metrics/metrics.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go`

The main review question is whether the chosen label set is the right first compromise between useful detail and low cardinality.

### What should be done in the future

- Restart the local server and verify the new metrics families in a real runtime.
- Add the browser-driven measurement scripts under this ticket’s `scripts/` directory.
- Use the new metrics to compare no-browser, one-tab, and multi-tab runs.

### Code review instructions

Start with the new metrics definitions and then trace where they are emitted:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_metrics.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/metrics/metrics.go`

Re-run:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
go test ./pkg/metrics ./internal/web ./pkg/media/gst -count=1
go test ./... -count=1
```

### Technical details

The first browser-preview metric slice now exposes these families:

```text
screencast_studio_preview_http_clients
screencast_studio_preview_http_streams_started_total
screencast_studio_preview_http_streams_finished_total
screencast_studio_preview_http_frames_served_total
screencast_studio_preview_http_bytes_served_total
screencast_studio_preview_http_flushes_total
screencast_studio_preview_frame_updates_total
screencast_studio_preview_ensures_total
screencast_studio_preview_releases_total
```

## Step 3: Added Ticket-Local Restart and Metrics-Sampling Scripts and Validated Them Live

After the preview-serving metrics were committed, the next missing piece was a reusable way to bring the local server up cleanly and capture `/metrics` snapshots into ticket-local artifacts. I did not want the first browser-driven matrix run to depend on ad hoc shell history, so I created two helper scripts under the new ticket before attempting any heavier browser measurements.

This slice is intentionally small: one script restarts `scs-web-ui` in `tmux`, and the second repeatedly samples `/metrics` into a timestamped result directory. I validated both immediately so the saved files are already known-good building blocks rather than speculative scaffolding.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue the browser-streaming investigation incrementally, with reusable experiment assets saved into the ticket and validated as they are added.

**Inferred user intent:** Build the browser-preview performance work the same disciplined way as the earlier runtime investigations: save the helper scripts first, then use them as stable building blocks for the real matrix runs.

**Commit (code):** `fb87ab7f33c0eb2e379d83fe0e9b16c54aae70e9` — `Add browser preview metrics helper scripts`

### What I did

- Added:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/01-restart-scs-web-ui.sh`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/02-sample-preview-metrics.sh`
- Made both executable.
- Ran the restart helper, which recreated the `scs-web-ui` tmux session and brought the server back up successfully on `:7777`.
- Ran the metrics sampler for a short smoke window:

```bash
DURATION_SECONDS=3 INTERVAL_SECONDS=1 bash ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/02-sample-preview-metrics.sh
```

- Saved the first result directory at:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-160358/`
- Confirmed the raw `.prom` snapshots contained the new preview-serving metric families.

### Why

Before adding the real browser-driven matrix harnesses, I wanted two stable building blocks:

1. a clean server restart path,
2. and a repeatable way to capture metrics snapshots into ticket-local artifacts.

That reduces the chance that later matrix results become hard to reproduce or impossible to compare cleanly.

### What worked

- The restart helper brought the server back up and printed the current health plus recent tmux pane output.
- The metrics sampler created a timestamped result directory with a manifest, summary, and raw `.prom` snapshots.
- The first raw snapshot already showed the new preview metric families, confirming the observability slice is live in the runtime, not just in tests.

### What didn't work

I initially mistyped the first raw snapshot filename while reading it back manually. The sampler itself was fine; the issue was my own follow-up inspection command.

### What I learned

The new ticket now has the minimum runtime scaffolding required to begin real browser-path measurement work without improvising the setup each time. That is especially useful because the next matrix runs will need clean result directories and comparable startup conditions.

### What was tricky to build

The subtle part was deciding how much structure to add to the sampler before the first real matrix run exists. I deliberately kept it simple: raw Prometheus snapshots plus a manifest and summary. That preserves the source data losslessly and avoids overcommitting too early to one parsing or plotting format.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/01-restart-scs-web-ui.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/02-sample-preview-metrics.sh`

The review question is whether the metrics sampler should stay raw-snapshot-first or whether the next slice should immediately add a parsed long-form time-series export for easier plotting.

### What should be done in the future

- Add the first real browser-driven matrix harness, starting with desktop preview only.
- Pair the metrics sampler with CPU sampling and browser/tab control.
- Save those heavier results under this same ticket-local `scripts/results/` structure.

### Code review instructions

Re-run:

```bash
bash ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/01-restart-scs-web-ui.sh
DURATION_SECONDS=3 INTERVAL_SECONDS=1 bash ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/02-sample-preview-metrics.sh
```

Then inspect the emitted result directory under:

```text
ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/
```

### Technical details

The first saved result directory is:

```text
scripts/results/20260414-160358/
```

The first raw snapshot there already includes headers for:

```text
screencast_studio_preview_http_clients
screencast_studio_preview_http_streams_started_total
screencast_studio_preview_http_streams_finished_total
screencast_studio_preview_http_frames_served_total
screencast_studio_preview_http_bytes_served_total
screencast_studio_preview_http_flushes_total
```

## Step 4: Added the First Desktop Preview HTTP-Client Baseline Matrix

With preview-serving metrics and helper scripts in place, I wanted the first actual measurement slice to stay as controlled as possible. Rather than jumping immediately to full browser-tab automation, I built a server-side baseline harness around the MJPEG transport itself. That gives us a useful decomposition step: how much does server CPU move when the same desktop preview is streamed to 0, 1, or 2 HTTP MJPEG clients?

I was careful to describe this honestly in the ticket docs. This is **not yet** the full browser-tab matrix. It is an HTTP-client approximation of the browser-facing preview transport. That still matters because the server-side streaming cost should already show up there even before Chromium rendering enters the picture.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Keep moving the new ticket forward with real measurement harnesses, but do so in a staged way that separates server-side preview streaming cost from later full browser automation.

**Inferred user intent:** Build up the browser-preview investigation incrementally so each measurement slice has a clear scope and does not blur together multiple possible bottlenecks.

**Commit (code):** `70b22663dda79cb399b67c89b450df397ad611c9` — `Add desktop preview HTTP client matrix`

### What I did

- Added:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/03-desktop-preview-http-client-matrix/run.sh`
- Ran it with:

```bash
DURATION=4 bash ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/03-desktop-preview-http-client-matrix/run.sh
```

- Saved the first run under:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/03-desktop-preview-http-client-matrix/results/20260414-161024/`
- Added a short interpretation note:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/04-desktop-preview-http-client-baseline-summary.md`

### Why

This harness is the right intermediate step because it isolates one question cleanly:

> Does simply having more MJPEG stream consumers raise server CPU for the same desktop preview workload?

That gives us a server-side preview-stream baseline before layering in:

- real browser tabs,
- frontend preview lifecycle churn,
- and recording.

### What worked

- The harness ran cleanly against fresh server processes on separate ports for each scenario.
- Each scenario saved its own `pidstat`, metrics snapshots, API artifacts, and summary.
- The first short run already produced a useful directional signal:
  - `no-client` → `11.67%` avg CPU
  - `one-client` → `11.50%` avg CPU
  - `two-clients` → `15.50%` avg CPU

### What didn't work

I caught one packaging mistake while first committing this slice: the initial local commit accidentally included the built `server-bin` artifact from the run directory. I rewrote that local commit before documenting it so the history stayed clean.

### What I learned

The first baseline suggests three things:

1. keeping the upstream preview alive already costs real CPU even without a connected MJPEG client,
2. one short-lived client did not materially change the average CPU in this first run,
3. two clients did increase CPU, which supports the idea that preview-stream fan-out can matter.

That is not yet the whole web-UI story, but it is already a useful narrowing step.

### What was tricky to build

The tricky part was choosing a measurement slice that was honest about what it **was not**. It would have been easy to overstate this as a browser benchmark, but that would blur an important distinction. This harness measures the server-side MJPEG streaming path with HTTP consumers. The later browser-tab matrix still needs to exist, because browser lifecycle behavior and real tab multiplicity may add more cost or different behavior on top.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/03-desktop-preview-http-client-matrix/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/04-desktop-preview-http-client-baseline-summary.md`

The main review question is whether the current 0/1/2-client baseline is the right precursor to the real browser-tab matrix or whether one more recording-inclusive HTTP-client baseline should come first.

### What should be done in the future

- Add the next desktop preview matrix slice with recording enabled.
- Add a real browser-tab harness so the browser UI path can be compared against this HTTP-client baseline.
- Extend the same pattern later to camera-only and desktop-plus-camera scenarios.

### Code review instructions

Re-run:

```bash
DURATION=4 bash ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/03-desktop-preview-http-client-matrix/run.sh
```

Then inspect:

- `.../scripts/03-desktop-preview-http-client-matrix/results/<timestamp>/01-summary.md`
- `.../no-client/01-summary.md`
- `.../one-client/01-summary.md`
- `.../two-clients/01-summary.md`

### Technical details

The first saved baseline run currently reports:

```text
no-client  -> 11.67% avg CPU
one-client -> 11.50% avg CPU
two-clients -> 15.50% avg CPU
```

## Step 5: Built the Larger Matrix and Proved the Browser-Connected Recording Slice Is Real

Once the first HTTP-client baseline was in place, the next question was obvious: does the real browser path actually behave materially differently from a plain MJPEG consumer, or was the earlier approximation already close enough? To answer that honestly, I built a larger matrix in two layers.

First, I added a fresh-server matrix that extends the HTTP-client baseline to include recording. That isolates what happens with preview-only and preview-plus-recording under `0`, `1`, and `2` plain MJPEG consumers. Second, I added a live browser sampler plus ticket-local Playwright scripts so I could measure the actual Studio page in one-tab, two-tab, and desktop-plus-camera scenarios.

This turned out to be the decisive measurement pass. The fresh-server plain-MJPEG-client recording cases stayed around `158–165%` avg CPU, but the **real browser-connected recording** cases were dramatically hotter: about `410.60%` avg CPU for desktop one-tab preview+recording and `432.97%` for desktop two-tab preview+recording.

### Prompt Context

**User prompt (verbatim):** "Yes, create the matrix and do extensive performance testing"

**Assistant interpretation:** Expand SCS-0015 from planning and light baselines into a real broader measurement campaign, with new harnesses, real result directories, and the browser-connected path measured directly rather than inferred.

**Inferred user intent:** Stop treating the browser path as a hypothesis and gather enough evidence to identify whether the hot slice is MJPEG fan-out, browser lifecycle behavior, or the combination of browser connection plus recording.

**Commit (code):** `ad0523a5d8d8f1a891d9187577402f528491b691` — `Add browser preview performance matrix harnesses`

### What I did

Added new matrix and browser helper scripts:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/05-desktop-preview-http-client-recording-matrix/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/07-live-server-browser-scenario-sample.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/08-playwright-browser-matrix/00-usage.txt`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/08-playwright-browser-matrix/01-open-studio-and-wait-desktop.js`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/08-playwright-browser-matrix/02-open-second-desktop-tab.js`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/08-playwright-browser-matrix/03-start-recording.js`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/08-playwright-browser-matrix/04-stop-recording.js`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/08-playwright-browser-matrix/05-add-camera-if-needed.js`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/08-playwright-browser-matrix/06-capture-browser-preview-state.js`

Ran the larger fresh-server HTTP-client matrix:

```bash
DURATION=6 bash ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/05-desktop-preview-http-client-recording-matrix/run.sh
```

Saved under:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/05-desktop-preview-http-client-recording-matrix/results/20260414-163154/`

Then ran real browser-backed scenarios against the live `:7777` server:

- one-tab desktop preview-only
- one-tab desktop preview+recording
- two-tab desktop preview-only
- two-tab desktop preview+recording
- one-tab desktop+camera preview-only
- one-tab desktop+camera preview+recording

Saved under:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-163610/`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-163951/`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-164457/`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-164535/`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-164657/`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-164720/`

Also saved the first combined findings note and browser evidence files:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/09-browser-preview-matrix-findings-summary.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/10-browser-session-network-summary.txt`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/11-browser-playwright-state-desktop-camera.json`

### Why

The first HTTP-client baseline was useful, but it still left the user’s key complaint open: the real Studio page looked much hotter than a plain MJPEG consumer. The only honest way to close that gap was to measure both:

1. a larger fresh-server plain-client matrix,
2. and the real browser-connected page.

That gives us a genuine comparison instead of an argument by analogy.

### What worked

The biggest thing that worked is that the matrix finally produced a decisive contrast.

Fresh-server plain-MJPEG-client results:

- `preview-no-client` → `15.67%`
- `preview-one-client` → `18.11%`
- `preview-two-clients` → `19.22%`
- `record-no-client` → `162.22%`
- `record-one-client` → `158.56%`
- `record-two-clients` → `165.00%`

Real browser-backed results:

- desktop, one tab, preview only → `14.20%`
- desktop, one tab, preview + recording → `410.60%`
- desktop, two tabs, preview only → `12.69%`
- desktop, two tabs, preview + recording → `432.97%`
- desktop + camera, one tab, preview only → `20.10%`
- desktop + camera, one tab, preview + recording → `343.71%`

This is the first solid proof that the browser-connected recording slice is not just a small variation on the curl/MJPEG baseline.

### What didn't work

While improving the live browser sampler, I introduced a shell heredoc bug: the markdown fence `````text````` inside an unquoted heredoc was interpreted as command substitution syntax and produced:

```text
bad substitution: no closing "`" in `text
```

I fixed that by changing the emitted fence style to `~~~text` and reran a validation sample successfully.

### What I learned

The strongest current learning is this:

> The browser-connected recording path is the missing hot slice, and simple MJPEG fan-out alone is not enough to explain it.

The clearest side-by-side comparison is:

- fresh server, desktop preview + recording + one plain MJPEG client → `158.56%` avg CPU
- live server, desktop preview + recording + one real browser tab → `410.60%` avg CPU

I also learned something subtler from the metric snapshots: during the browser recording runs, the per-run MJPEG frame/byte deltas were actually fairly modest. That means the `~400%` server heat is **not** explained simply by “the server had to push vastly more JPEG bytes.”

### What was tricky to build

The hardest part in this step was keeping the measurement story honest across two different kinds of runs:

- fresh dedicated-server runs with plain HTTP consumers,
- live shared-session runs with the real Studio page.

Those are intentionally different. The first isolates server-side stream fan-out. The second exposes the real browser-connected behavior. The ticket notes needed to preserve that distinction rather than collapsing them into one pretend-uniform matrix.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/05-desktop-preview-http-client-recording-matrix/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/07-live-server-browser-scenario-sample.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/09-browser-preview-matrix-findings-summary.md`

The main review question is whether the current evidence is strong enough to prioritize deeper browser-path handler instrumentation over more matrix expansion first.

### What should be done in the future

- Add camera-only one-tab preview and preview+recording scenarios so the matrix is even more complete.
- Add more browser-path instrumentation if needed:
  - handler write/flush timing,
  - per-stream skip/drop counters,
  - possibly browser-side state capture per run.
- Compare visible-tab versus headless browser consumption if we need to separate UI rendering cost from stream-consumer behavior.

### Code review instructions

Re-run the fresh-server matrix:

```bash
DURATION=6 bash ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/05-desktop-preview-http-client-recording-matrix/run.sh
```

Then use the Playwright scripts plus the live sampler to recreate the browser scenarios:

```text
scripts/08-playwright-browser-matrix/01-open-studio-and-wait-desktop.js
scripts/08-playwright-browser-matrix/02-open-second-desktop-tab.js
scripts/08-playwright-browser-matrix/03-start-recording.js
scripts/08-playwright-browser-matrix/04-stop-recording.js
scripts/08-playwright-browser-matrix/05-add-camera-if-needed.js
scripts/07-live-server-browser-scenario-sample.sh
```

### Technical details

The clearest current numeric contrast is:

```text
fresh server + 1 plain MJPEG client + desktop recording  -> 158.56% avg CPU
live Studio page + 1 real browser tab + desktop recording -> 410.60% avg CPU
live Studio page + 2 real browser tabs + desktop recording -> 432.97% avg CPU
```

## Step 6: Backfilled a Dedicated Ongoing Lab Report for SCS-0015

After the first browser-backed matrix pass, the ticket already had enough moving pieces that the design doc, report draft, changelog, and diary were no longer the best place to hold the raw experimental story in one continuation-friendly view. The user asked specifically for a lab report document that we keep as we investigate, so I created that document and backfilled the current experiment set in detail rather than leaving the evidence spread across result directories and summary notes.

The goal of this step was not to invent new measurements. It was to preserve the measurements we already ran in a form that makes future investigation easier: one place that records the exact experiment purpose, script used, command shape, result directory, CPU summary, caveats, and the current working explanation for the `~400%` desktop preview-plus-recording browser spike.

### Prompt Context

**User prompt (verbatim):** "ok, write up your current experiments and results in detail in a lab report document that we keep as we investigate (backfill in detail), commit appropriately, then moving forward keep track of your experiments, update your diary."

**Assistant interpretation:** Create a durable ticket-local lab report for SCS-0015, backfill the experiments already run into it in detail, update the diary and ticket bookkeeping, and treat that report plus the diary as the ongoing record from here forward.

**Inferred user intent:** Turn the current measurement work into a maintained research artifact so future investigation does not depend on reconstructing the history from scattered result directories, chat messages, or half-polished summaries.

**Commit (docs):** `Add browser preview lab report`

### What I did

- Added a new ongoing lab report document:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/reference/02-browser-preview-streaming-lab-report.md`
- Backfilled the current experiment inventory into that document, including:
  - metrics smoke validation,
  - the first desktop HTTP-client baseline matrix,
  - the larger fresh-server HTTP-client matrix with recording,
  - the live browser-backed desktop one-tab and two-tab runs,
  - the desktop-plus-camera one-tab runs,
  - the improved sampler validation after the heredoc bug fix,
  - the browser-tool network artifact,
  - and the saved Playwright browser-state artifact.
- Updated ticket docs to point at the new lab report:
  - `index.md`
  - `tasks.md`
  - `changelog.md`
  - `reference/01-investigation-diary.md`
- Kept the lab report framed explicitly as a **living experiment ledger**, not as the final polished report.

### Why

By this point in SCS-0015, the experiment set was already large enough that “just read the result directories” was no longer a good continuation strategy. The lab report is the right middle layer between:

- raw result trees,
- and the higher-level performance report.

It preserves the details needed for debugging and future A/B work without forcing every future reader to manually piece the story back together.

### What worked

- The new lab report now gives the ticket one canonical place to read the current experimental record.
- The report backfills the strongest result clearly:
  - fresh server + one plain MJPEG client + desktop recording → `158.56%`
  - real Studio page + one real browser tab + desktop recording → `410.60%`
- The report also captures the subtler current interpretation that served MJPEG bytes alone do not explain the browser-path spike.

### What didn't work

There was no new implementation failure in this documentation step itself. The main limitation is still analytical rather than editorial: we do not yet have the next round of server-path A/B instrumentation to prove exactly how much of the hot slice comes from websocket event fanout versus MJPEG serving versus upstream shared-capture interaction.

### What I learned

The ticket is now at a stage where the **desktop preview + recording + one real browser tab** scenario is already strong enough to carry the next investigation phase by itself. The lab report made that easier to see because it puts the fresh-server and real-browser runs side by side in one place.

### What was tricky to build

The main tricky part was deciding what belongs in the lab report versus the diary versus the main performance report. I treated the lab report as the place for experiment-by-experiment preservation, exact result locations, and current working hypotheses. I left the main performance report as the more structured summary document and the diary as the chronological narrative of what I changed and why.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/reference/02-browser-preview-streaming-lab-report.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/design/02-browser-preview-streaming-performance-report.md`

The main review question is whether the current lab-report structure is the right level of detail for future experimental additions, especially once we start adding deeper A/B instrumentation around websocket versus MJPEG cost.

### What should be done in the future

- Keep the new lab report updated after each meaningful experimental slice.
- Keep the diary updated in parallel for chronology and implementation/debugging narrative.
- Use the desktop preview + recording + one real browser tab scenario as the main repro while we instrument deeper.

### Code review instructions

Start with:

- `reference/02-browser-preview-streaming-lab-report.md`

Then compare it against:

- `scripts/09-browser-preview-matrix-findings-summary.md`
- `design/02-browser-preview-streaming-performance-report.md`
- the saved result directories listed in the lab report

### Technical details

The lab report now explicitly records these key directories and their role:

```text
scripts/03-desktop-preview-http-client-matrix/results/20260414-161024/
scripts/05-desktop-preview-http-client-recording-matrix/results/20260414-163154/
scripts/results/20260414-163610/
scripts/results/20260414-163951/
scripts/results/20260414-164457/
scripts/results/20260414-164535/
scripts/results/20260414-164657/
scripts/results/20260414-164720/
scripts/results/20260414-165126/
```

## Step 7: Added EventHub/WebSocket Metrics and Ran a Focused MJPEG-vs-WebSocket Ablation

After the larger browser matrix pass and the lab-report backfill, the next question was narrower: if websocket event traffic is the key browser-only extra load, can I reproduce a large chunk of the browser-path spike by taking the same desktop preview+recording workload and adding a synthetic `/ws` consumer on top of one MJPEG client? Rather than broadening scenarios, I chose the cleanest server-side A/B I could build around the current hypothesis.

This step ended up being especially useful because it both improved observability and changed the working theory. I added explicit EventHub and websocket metrics first, then built a fresh-server ablation harness with two scenarios: `mjpeg-only` and `mjpeg-plus-ws`. The trusted rerun showed that websocket/event fanout is real, but by itself it only moved avg CPU from `166.56%` to `170.48%`. That is nowhere near enough to explain the real browser one-tab desktop preview+recording result of `410.60%`.

### Prompt Context

**User prompt (verbatim):** (same as Step 6)

**Assistant interpretation:** Continue the browser-preview investigation in small, reviewable slices, keep the new lab report and diary updated, and dig deeper into the strongest current desktop preview+recording repro instead of broadening the scenario matrix.

**Inferred user intent:** Turn the current explanation into something more evidence-based by testing the strongest remaining hypothesis directly and preserving the result carefully.

**Commit (code):** `ede87bb2fb824ee4f427813ba29daedf1b280b22` — `Add websocket and eventhub metrics`

**Commit (code):** `23234634f3e73c822c58b7af9dc91756519cf2f0` — `Add MJPEG websocket ablation harness`

### What I did

First, I added low-cardinality observability around the event/websocket path:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/event_metrics.go`
- updated `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/event_hub.go`
- updated `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go`
- added `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/event_hub_test.go`
- updated `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/metrics_test.go`

The new metric families are:

```text
screencast_studio_eventhub_subscribers
screencast_studio_eventhub_events_published_total
screencast_studio_eventhub_events_delivered_total
screencast_studio_eventhub_events_dropped_total
screencast_studio_websocket_connections
screencast_studio_websocket_events_written_total
screencast_studio_websocket_event_write_errors_total
```

I also updated the live browser sampler so `metric-deltas.txt` includes:

- preview HTTP metrics,
- preview frame updates,
- EventHub metrics,
- websocket metrics.

Then I added the focused ablation harness:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/ws_client/main.go`

The harness runs a fresh server for two scenarios:

- one MJPEG client only
- one MJPEG client plus one synthetic websocket consumer

I first hit a harness problem: the initial run under:

- `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173359/`

produced an invalid `mjpeg-only` comparison because recording started before the preview path had clearly produced an initial frame. I fixed the harness by waiting for an initial preview screenshot before measurement in both scenarios and reran it.

The trusted rerun is:

- `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173541/`

I also wrote a short summary note:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/13-mjpeg-websocket-ablation-summary.md`

### Why

This was the cleanest way to test the current websocket hypothesis without depending on a flaky browser-tool session. The big browser result (`410.60%`) was still real, but I needed to know whether the missing cost was plausibly “MJPEG plus websocket fanout” or whether something more browser-specific was still hiding behind the real tab.

### What worked

The trusted rerun gave a clear result:

- `mjpeg-only` → `166.56%` avg CPU
- `mjpeg-plus-ws` → `170.48%` avg CPU

The websocket-enabled case also proved the event/websocket path was active and measurable:

- `preview.state` published: `33`
- `preview.state` delivered: `23`
- websocket `preview.state` writes: `23`
- websocket client total messages observed: `54`

So the observability slice worked, and the focused ablation answered the main question it was built to answer.

### What didn't work

I hit two concrete problems in this step.

First, the browser tool itself remained unreliable after a profile-lock issue. Even after clearing stale Playwright-MCP Chrome processes, the direct browser tool session did not become usable again for a new real-browser rerun during this slice.

Second, my first ablation run was invalid because of a harness race. The first run under `20260414-173359` let the `mjpeg-only` case start recording too early, before the preview path had clearly produced an initial frame. That yielded a bogus `11.00%` result. I fixed the harness by waiting for an initial preview screenshot before measurement in both scenarios and then reran the matrix.

### What I learned

This step materially changed the explanation.

Before the ablation, websocket/event fanout looked like the strongest server-side suspect for the browser-only gap. After the ablation, I now think it is a **real contributor but not the dominant one by itself**. A plain websocket consumer on top of one MJPEG client only raised avg CPU by about `3.92` points in the trusted fresh-server rerun. That is far too small to explain the jump from the `~158–166%` fresh-server desktop-recording band to the `410.60%` real browser one-tab case.

### What was tricky to build

The hardest part was making the ablation fair. The first version looked simple, but it hid an important timing problem: if the preview path does not clearly have a first frame before recording starts, the comparison can be misleading. The fix was to make both scenarios wait for an initial preview screenshot before the actual timed measurement starts.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/event_metrics.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/event_hub.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/13-mjpeg-websocket-ablation-summary.md`

The main review question is whether the current ablation is a fair enough synthetic proxy to lower confidence in websocket fanout as the primary explanation, or whether another synthetic browser-consumer behavior still needs to be added before drawing that conclusion strongly.

### What should be done in the future

- Rerun the real browser one-tab desktop preview+recording case with the new websocket/event metrics once the browser tool path is healthy again.
- If needed, add deeper handler-path timing metrics around MJPEG writes/flushes and any other browser-specific behavior that still differentiates a real tab from a plain MJPEG-plus-websocket synthetic client.

### Code review instructions

Start with the new metrics and harness files:

- `internal/web/event_metrics.go`
- `internal/web/event_hub.go`
- `internal/web/handlers_ws.go`
- `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/run.sh`
- `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/ws_client/main.go`

Validation commands used:

```bash
gofmt -w internal/web/event_metrics.go internal/web/event_hub.go internal/web/handlers_ws.go internal/web/metrics_test.go internal/web/event_hub_test.go
go test ./internal/web ./pkg/metrics -count=1
go test ./... -count=1
DURATION=6 bash ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/run.sh
```

### Technical details

Trusted ablation rerun:

```text
scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173541/
```

Key result:

```text
mjpeg-only    -> 166.56% avg CPU
mjpeg-plus-ws -> 170.48% avg CPU
```

## Step 8: Added MJPEG Handler Timing Metrics Before Changing Behavior Again

After the websocket ablation, the most important thing was to stop guessing. We had already learned that websocket/event fanout was too small by itself to explain the browser-path spike, but that still left an open question: is the real browser spending much more time in the MJPEG write/flush loop than the synthetic clients, or is most of the missing heat still elsewhere? The right next move was to instrument first and only perturb behavior later.

So this step is intentionally a measurement-only slice. I added low-cardinality timing counters to the browser-facing MJPEG handler for loop iterations, idle iterations, cumulative write time, and cumulative flush time. This does not answer the performance question by itself yet, but it prepares the next real one-tab browser rerun to answer it much more directly.

### Prompt Context

**User prompt (verbatim):** "ok, go ahead. continue committing and keeping a diary as you go"

**Assistant interpretation:** Implement the next instrumentation slice now, keep the work split into commits, and update the SCS-0015 diary and ticket docs as the code lands.

**Inferred user intent:** Continue the investigation in disciplined, reviewable steps, preferring measurement before speculative optimization and preserving the reasoning trail in the ticket.

**Commit (code):** `9fd8754ab4db6aab3ce0bd174c2ac006d957b1dd` — `Add MJPEG handler timing metrics`

### What I did

- Added new preview HTTP timing metric families in:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_metrics.go`
- Instrumented the MJPEG handler in:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`
- Extended focused metric visibility checks in:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/metrics_test.go`
- Extended preview-endpoint coverage in:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server_test.go`

The new metric families are:

```text
screencast_studio_preview_http_loop_iterations_total
screencast_studio_preview_http_idle_iterations_total
screencast_studio_preview_http_write_nanoseconds_total
screencast_studio_preview_http_flush_nanoseconds_total
```

Validation commands used:

```bash
gofmt -w internal/web/preview_metrics.go internal/web/handlers_preview.go internal/web/metrics_test.go internal/web/server_test.go
go test ./internal/web ./pkg/metrics -count=1
go test ./... -count=1
```

### Why

The websocket ablation reduced confidence in the websocket hypothesis, but it did not tell us whether the browser-specific heat is concentrated inside the MJPEG handler itself. Adding timing counters to the existing MJPEG loop is the cleanest next diagnostic step because it gives the next rerun a way to distinguish:

- many iterations but little write/flush cost,
- or genuinely expensive write/flush behavior under the real browser.

### What worked

- The timing instrumentation fit cleanly into the existing low-cardinality metrics model using only `source_type` labels.
- Focused tests and the full repo test suite both passed.
- The preview endpoint test now proves that the new timing metric families appear after a real MJPEG stream has been served.

### What didn't work

No new runtime failure happened in this slice. The only practical limitation is that I have not yet rerun the real one-tab browser scenario with these metrics enabled, so this step is infrastructure, not a new performance result.

### What I learned

The handler is easy to instrument incrementally without changing runtime behavior. That means we can keep the investigation honest: first measure write/flush/loop behavior, then decide whether a behavior change like preview-state throttling is actually justified.

### What was tricky to build

The tricky part was choosing metrics that are useful without becoming noisy or misleading. I avoided per-preview IDs and avoided adding a larger metrics primitive just for this slice. Instead, I stuck to cumulative counters for loop count, idle count, write nanoseconds, and flush nanoseconds. That keeps the metrics cheap and easy to compare across the kind of short-lived fresh-server reruns we are already doing in this ticket.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_metrics.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server_test.go`

The main review question is whether these timing counters are sufficient to answer the current fork in the investigation, or whether the next rerun will still need even more detailed handler instrumentation such as blocked-write reasons or stale-frame counters.

### What should be done in the future

- Rerun the high-signal desktop preview + recording + one real browser tab scenario with the new timing metrics enabled.
- Compare the MJPEG timing deltas against the earlier synthetic-client and websocket-ablation runs.
- Only after that decide whether to try a behavior change like throttling `preview.state`.

### Code review instructions

Start in:

- `internal/web/preview_metrics.go`
- `internal/web/handlers_preview.go`

Then validate through:

- `internal/web/metrics_test.go`
- `internal/web/server_test.go`

Commands used:

```bash
gofmt -w internal/web/preview_metrics.go internal/web/handlers_preview.go internal/web/metrics_test.go internal/web/server_test.go
go test ./internal/web ./pkg/metrics -count=1
go test ./... -count=1
```

### Technical details

The instrumentation points are deliberately simple:

- one counter increment per handler loop iteration,
- one counter increment for idle iterations that do not serve a new frame,
- one cumulative nanosecond counter around multipart write work,
- one cumulative nanosecond counter around `flusher.Flush()`.

This should make the next real-browser rerun much more informative without changing preview behavior yet.
