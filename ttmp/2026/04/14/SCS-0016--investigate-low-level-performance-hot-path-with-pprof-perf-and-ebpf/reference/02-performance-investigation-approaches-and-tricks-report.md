---
Title: Performance investigation approaches and tricks report
Ticket: SCS-0016
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - backend
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/run.sh
      Note: Representative standalone harness from the earlier performance work
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/05-desktop-preview-http-client-recording-matrix/run.sh
      Note: Fresh-server MJPEG baseline matrix harness
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/07-live-server-browser-scenario-sample.sh
      Note: Live browser-backed sampler that made metric-delta comparison routine
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/run.sh
      Note: Focused synthetic ablation harness that narrowed websocket suspicion
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/02-capture-pprof-cpu-profile.sh
      Note: First lower-level profiling helper after app-level narrowing ran out of explanatory power
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/reference/02-browser-preview-streaming-lab-report.md
      Note: Detailed experiment ledger backing the approaches summarized here
ExternalSources: []
Summary: Project report describing the main approaches, tricks, and narrowing strategies used to investigate screencast-studio performance issues so far.
LastUpdated: 2026-04-14T21:00:00-04:00
WhatFor: Preserve the actual investigation playbook so future profiling work reuses the effective techniques and avoids the dead ends already explored.
WhenToUse: Read before starting a new performance investigation in screencast-studio or when handing the project to a new investigator.
---

# Performance investigation approaches and tricks report

## Goal

Summarize the practical approaches, tricks, and methodological choices that have worked so far for diagnosing Screencast Studio performance issues, especially around the shared GStreamer preview+recording path and the later browser-connected hot path.

## Context

The performance investigation in this project did not start with low-level profilers. It started with a much more basic problem: the team did not yet know which layer was responsible for the observed heat. Was the cost in GStreamer recording itself, the shared preview branch, the appsink/appsrc bridge, MJPEG fan-out, websocket traffic, the real browser tab, or some combination of those?

Because the initial uncertainty was high, the most effective strategy was to progress from **broad and honest measurement** toward **narrower ablations** and only then toward **lower-level profiling**. That staged approach matters. It prevented the investigation from jumping too early into opaque profiler output before the higher-level question had been sharpened enough to make profiler results meaningful.

## Quick Reference

## The main investigation principles that worked

### 1. Start with the smallest honest question, not the fanciest tool

The first useful performance investigations asked questions like:

- how expensive is recorder-only work with no preview consumer?
- how expensive is preview-only work with no recorder?
- what happens when preview and recorder share the same source?
- does a plain MJPEG client materially change server CPU?
- does a real browser tab behave differently from a plain HTTP client?

This sounds obvious, but it is the foundation for every later narrowing step. The project got its biggest clarity gains from small, explicit scenario comparisons rather than from clever instrumentation alone.

### 2. Keep results reproducible and ticket-local

A major trick that paid off repeatedly was saving every meaningful experiment under a numbered ticket-local `scripts/` tree with raw artifacts and human-readable summaries. That made the work continuation-friendly and prevented “I think we saw something like this yesterday” drift.

Useful recurring pattern:

- `scripts/NN-something/run.sh`
- `scripts/NN-something/results/YYYYMMDD-HHMMSS/`
- raw outputs: `pidstat`, `.prom`, `ffprobe`, stdout/stderr, JSON snapshots
- human summary: `01-summary.md`

This mattered because many conclusions only became convincing after comparing several saved runs side by side.

### 3. Separate fresh-server baselines from live shared-server browser runs

One of the most important methodological decisions in the browser work was to **stop pretending all clients were the same**.

The investigation explicitly split:

- **fresh dedicated-server runs** with plain MJPEG clients
- **live browser-backed runs** against the real `:7777` server and the actual Studio page

That separation revealed the central finding of SCS-0015: plain MJPEG fan-out alone did **not** explain the much hotter real browser recording path. Without that split, the project might have continued to misclassify the browser issue as “just more preview bytes.”

### 4. Do not trust averages alone

Another trick that proved essential: always inspect the per-second trace and the scenario window, not just the summary average.

The later browser reruns made this especially clear. Some sampled windows included preview-only warmup before recording fully ramped. If those runs had been interpreted only through the average CPU headline, they would have looked less alarming than they actually were. The per-second `pidstat` traces showed the real story: late-run climbs back into the `~300–400%` range.

### 5. Use paired before/after or A/B experiments wherever possible

The project got much better answers from **paired comparisons** than from isolated runs.

Examples that worked well:

- direct recorder-only vs shared-runtime recorder-only
- preview-only vs preview+recorder
- current preview branch vs cheap-preview/adaptive-preview branch
- one MJPEG client vs one MJPEG client plus one synthetic websocket consumer
- no browser vs one real browser tab vs two real browser tabs

This A/B style prevented the investigation from overfitting to noisy absolute numbers.

## The concrete approaches that worked best

### Approach A: standalone runtime harnesses before touching the app path

Earlier work under SCS-0014 used standalone Go/GStreamer harnesses to isolate:

- recorder-only cost
- preview-only cost
- preview+recorder interaction
- shared bridge overhead
- adaptive preview policy effects

This was a good first move because it removed browser lifecycle noise and answered architectural questions cheaply.

### Approach B: fresh-server HTTP-client baselines

Once the browser path became suspect, the next good trick was to create fresh-server MJPEG-client baselines rather than jumping straight to the real browser.

Why it worked:
- it isolated server-side MJPEG fan-out honestly
- it avoided pretending that a curl-like client was a browser benchmark
- it gave a clean reference band for later comparison

This is the step that established the key contrast:

- plain MJPEG client recording runs stayed around the `~158–166%` band
- the real browser one-tab recording run could reach the `~400%` band

### Approach C: real-browser automation with controlled tab hygiene

Another practical trick was using ticket-local Playwright snippets with explicit tab hygiene:

- open only the tab(s) needed for the scenario
- close extra tabs after each run
- release previews afterward

This mattered because stale browser listeners and extra MJPEG consumers could otherwise contaminate the measurement.

### Approach D: metric deltas from first/last `.prom` snapshots

A high-value trick was computing metric deltas from saved first/last `/metrics` snapshots instead of only looking at raw cumulative counters.

This turned cumulative metrics into per-run evidence and made it possible to say things like:

- how many preview frames were actually served during this run?
- how many websocket events were actually published?
- how much cumulative write time happened in the MJPEG loop?

Without that delta layer, the metrics would have been much less useful for short experiments.

### Approach E: ablate one hypothesis at a time

The websocket ablation is the clearest example.

Instead of changing multiple things at once, the project created a focused experiment:

- one MJPEG client only
- one MJPEG client plus one synthetic websocket consumer

That single ablation materially lowered confidence in “websocket fanout is the main culprit” because it showed only a small CPU increase.

### Approach F: add timing metrics only after a suspect survives coarse measurement

The project avoided adding timing metrics everywhere at once. Instead, it added them where the previous evidence made them worth it:

1. final MJPEG write/flush timing
2. PreviewManager store/copy/publication timing
3. EventHub publish timing

That staged approach kept the metrics understandable and preserved a clean chain of reasoning.

### Approach G: interpret “negative findings” as progress

A very important investigation habit in this project was treating disproven hypotheses as real deliverables.

Examples:
- final MJPEG write/flush is not expensive enough to explain the spike
- PreviewManager frame-store/copy is not expensive enough
- EventHub `preview.state` publish is not expensive enough
- plain websocket fanout alone is not expensive enough

Those are not failures. They are narrowing steps, and they are exactly what justified moving to lower-level profiling.

## The tricks that saved time

### Trick 1: poison assumptions with a direct baseline

Whenever a theory sounded plausible, the fastest way to test it was usually to build a tiny baseline or ablation rather than debate it abstractly.

### Trick 2: keep labels low-cardinality in metrics

This made metrics easier to compare across many runs and kept the outputs readable.

### Trick 3: save the raw artifacts, not just the conclusion

Later reinterpretation often depended on raw `pidstat`, `.prom`, or ffprobe outputs.

### Trick 4: be explicit about scope and comparability

The investigation repeatedly distinguished between:

- standalone runtime experiments
- fresh dedicated-server MJPEG baselines
- real browser-backed runs

That prevented invalid apples-to-oranges conclusions.

### Trick 5: use the same “high-signal repro” repeatedly

The repeated target scenario:

```text
desktop preview + recording + one real browser tab
```

became extremely valuable because every new slice could be compared against the same known-problem case.

## The mistakes and near-mistakes worth remembering

### 1. Warmup windows can corrupt a headline average

A run that includes preview-only warmup before recording fully ramps may still be useful, but not for the same conclusion as a tightly aligned hot-window sample.

### 2. Synthetic clients are useful, but only for narrow questions

A synthetic MJPEG client or websocket consumer is great for an ablation. It is not a full browser substitute.

### 3. App-level metrics eventually stop being enough

Once upper-layer timing costs are ruled out, continuing to add only app metrics can create the illusion of progress without actually revealing the lower-layer hotspot.

## Current state of the investigation playbook

The current state machine for future investigations now looks like this:

1. **Scenario definition**
   - pin down the smallest honest repro
2. **Fresh baseline**
   - isolate the coarse layer first
3. **Live real-path validation**
   - prove the user-observed behavior really exists
4. **Focused ablation**
   - test one suspect at a time
5. **Targeted timing metrics**
   - only where the previous evidence points
6. **Lower-level profiling**
   - only after app-level narrowing has done its job

That sequence is probably the single most useful outcome of the performance work so far.

## Usage Examples

### Example: how to investigate a new “browser recording is hot” claim

1. Reuse the current high-signal browser repro.
2. Run a fresh-server baseline that isolates the same server-side transport.
3. Compare live browser and synthetic deltas, not just average CPU.
4. If the browser run is still unexplained, add one targeted ablation.
5. If upper-layer metrics stay tiny, move to pprof/perf rather than inventing more application counters blindly.

### Example: how to brief a new investigator

Tell them three things first:

- keep every experiment ticket-local and reproducible,
- do not trust average CPU without checking the window,
- negative findings are progress if they rule out a plausible hot path.

## Related

- `reference/03-prometheus-metrics-architecture-and-field-guide.md`
- `ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/reference/02-browser-preview-streaming-lab-report.md`
