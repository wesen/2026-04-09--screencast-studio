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
    - Path: internal/web/preview_manager.go
      Note: |-
        Preview lifecycle, reuse, and cached-frame fan-out behavior shape the browser streaming path
        Preview cache and fan-out behavior shaped the first-step reasoning
    - Path: internal/web/preview_metrics.go
      Note: Metric-family definitions were added in the first code slice
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/01-restart-scs-web-ui.sh
      Note: Added and validated in the second helper-script slice
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/02-sample-preview-metrics.sh
      Note: Added and validated in the second helper-script slice
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/03-desktop-preview-http-client-matrix/run.sh
      Note: Added and validated as the first matrix harness slice
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/04-desktop-preview-http-client-baseline-summary.md
      Note: Human-readable interpretation of the first 0/1/2-client desktop baseline
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
LastUpdated: 2026-04-14T15:42:00-04:00
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
