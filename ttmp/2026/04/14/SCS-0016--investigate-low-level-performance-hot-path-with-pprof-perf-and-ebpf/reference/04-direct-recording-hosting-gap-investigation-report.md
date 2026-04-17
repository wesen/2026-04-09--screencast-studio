---
Title: Direct recording hosting-gap investigation report
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
    - Path: pkg/media/gst/bus.go
      Note: GLib bus-watch integration discussed as a possible hosting boundary in the report
    - Path: pkg/media/gst/recording.go
      Note: Direct recording pipeline builder and the I420 caps fix discussed in the hosting-gap report
    - Path: ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/17-go-manual-direct-full-desktop-harness/main.go
      Note: Minimal direct Go harness used as the main non-app hosting control in the report
    - Path: ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/16-gst-launch-direct-full-desktop-match-app/20260415-004852/01-summary.md
      Note: Matched gst-launch control evidence cited in the report
    - Path: ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/18-direct-harness-ab-matrix/20260415-010905/02-summary.md
      Note: A/B matrix evidence for the remaining Go-hosted gap
    - Path: ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/19-perf-compare-go-manual-vs-gst-launch/20260415-011757/01-summary.md
      Note: Mixed-stack perf comparison evidence cited in the report
    - Path: ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/21-perf-stat-threadgroup-compare-go-vs-gst-launch/20260415-012510/01-summary.md
      Note: Threadgroup perf-stat comparison that surfaced the large page-fault delta
ExternalSources: []
Summary: Evidence-backed report on the remaining performance gap between a Go-hosted direct GStreamer recorder and an otherwise matched gst-launch control, including what has been ruled out and what the next investigation targets should be.
LastUpdated: 2026-04-15T01:40:00-04:00
WhatFor: Give future investigators and reviewers one detailed, continuation-friendly report that explains the current hypothesis boundary for the Go-hosted direct recording performance gap.
WhenToUse: Read this before doing more perf, scheduler, allocator, memory-fault, or eBPF work on SCS-0016.
---


# Direct recording hosting-gap investigation report

## Goal

Explain, with code-backed and artifact-backed evidence, what the current SCS-0016 investigation has actually proved about the large CPU gap between a Go-hosted direct GStreamer recording pipeline and an otherwise matched `gst-launch-1.0` control. The report is intended to replace vague folklore like “maybe the DSL is wrong” or “maybe frames are flowing through Go” with a narrower, evidence-backed problem statement.

## Context

The immediate background is SCS-0015. That earlier ticket had already ruled out several browser/server-layer suspects for the real one-tab desktop preview + recording spike. By the time SCS-0016 began, the main unresolved question had shifted from “which high-level web path is hot?” to “what lower-level runtime boundary is still making the system expensive?”

Early SCS-0016 work then made the problem even narrower:

1. a standalone direct full-desktop Go harness reproduced the late high-CPU ramp,
2. a matched `gst-launch-1.0` full-desktop control was much cooler,
3. the copied/app-like Go direct harness initially negotiated `Y444` instead of `I420`,
4. that graph bug was fixed,
5. but a large Go-hosted vs `gst-launch` gap still remained.

This report is about that narrowed “hosting gap” specifically.

## Quick Reference

### Current best conclusion

The remaining direct-recording hot path is **not** best explained by the app DSL building the wrong graph, by final MJPEG/server/browser work, or by ordinary Go userland processing each frame. The strongest current evidence points instead to a **Go-hosted process effect** that changes native runtime behavior relative to `gst-launch`, with the strongest current signal being a huge increase in **page faults** in the Go-hosted case.

### What we can currently say with high confidence

1. The direct recording pipeline in the app is built imperatively in Go, not by a generic media graph DSL.
2. The initial copied/app-like direct harness really did have a graph mistake: it pinned only framerate and let GStreamer negotiate `Y444` into `x264enc`.
3. After fixing that to `I420`, the copied/app-like Go direct harness and the fully manual Go direct harness realize the **same effective graph**.
4. Despite that, both Go-hosted direct harnesses remain much hotter than the matched `gst-launch` control.
5. Mixed-stack perf shows the running direct pipeline is dominated by **native x264 + GStreamer push-path work**, while visible Go/cgo frames are tiny by comparison.
6. Thread counts are similar between the Go-hosted and `gst-launch` cases, but the Go-hosted case shows dramatically more **page faults**.

### Most important saved artifacts

- Matched `gst-launch` control:
  - `scripts/results/16-gst-launch-direct-full-desktop-match-app/20260415-004852/01-summary.md`
- Pre-fix copied/app-like Go harness (`Y444`):
  - `scripts/15-go-direct-record-full-desktop-harness/results/20260415-005501/01-summary.md`
  - `scripts/15-go-direct-record-full-desktop-harness/results/20260415-005501/dot/...-playing.dot`
- Post-fix copied/app-like Go harness (`I420`):
  - `scripts/15-go-direct-record-full-desktop-harness/results/20260415-010455/01-summary.md`
  - `scripts/15-go-direct-record-full-desktop-harness/results/20260415-010455/dot/...-playing.dot`
- Manual Go harness:
  - `scripts/17-go-manual-direct-full-desktop-harness/results/20260415-005343/01-summary.md`
- 3x3 A/B matrix:
  - `scripts/results/18-direct-harness-ab-matrix/20260415-010905/02-summary.md`
- Mixed-stack perf compare:
  - `scripts/results/19-perf-compare-go-manual-vs-gst-launch/20260415-011757/01-summary.md`
- Threadgroup `perf stat` compare:
  - `scripts/results/21-perf-stat-threadgroup-compare-go-vs-gst-launch/20260415-012510/01-summary.md`

## Problem Statement and Scope

The exact question in scope is:

> Why is an otherwise matched **Go-hosted direct GStreamer desktop recorder** substantially hotter than an equivalent **`gst-launch-1.0` direct recorder**, even after the direct-path `Y444` caps bug has been fixed?

This report is **not** trying to explain:

- the old browser preview path by itself,
- websocket fanout,
- MJPEG serving,
- the shared preview bridge,
- or the direct effect of appsink/appsrc bridging.

Those were all part of the broader narrowing process, but the evidence summarized here concerns the **direct full-desktop recording control path**.

## Current-State Architecture Boundary

### The app/runtime direct video pipeline is built imperatively in Go

The app’s direct video recorder is built in `pkg/media/gst/recording.go`. The core builder starts at `buildVideoRecordingPipeline(...)` (`pkg/media/gst/recording.go:526`) and explicitly creates and links native elements: source elements, `videoconvert`, `videorate`, a raw capsfilter, `x264enc`, and the mux/filesink path (`pkg/media/gst/recording.go:526-585`).

The current direct path now explicitly constrains the raw encoder input with:

- `video/x-raw,format=I420,framerate=%d/1,pixel-aspect-ratio=1/1`

That comes from the capsfilter construction at `pkg/media/gst/recording.go:562`.

The `x264enc` settings are also explicit and simple:

- bitrate: `pkg/media/gst/recording.go:570`
- `bframes=0`: `pkg/media/gst/recording.go:571`
- `tune=4` (`zerolatency`): `pkg/media/gst/recording.go:572`
- `speed-preset=3` (`veryfast`): `pkg/media/gst/recording.go:573`

### The app/runtime has GLib bus-watch machinery, but the manual harness does not depend on it for steady-state work

The production runtime does have GLib main-loop bus-watch integration in `pkg/media/gst/bus.go`, where `startBusWatch(...)` adds a watch, creates a `glib.NewMainLoop(...)`, and runs it in a goroutine (`pkg/media/gst/bus.go:16-30`). That machinery is relevant because it is one of the few obvious Go/GLib hosting boundaries in the native runtime.

However, the manual direct harness deliberately minimizes the runtime control plane. It does not route frames into Go. Instead it:

1. builds the pipeline,
2. sets `PLAYING`,
3. sleeps,
4. sends EOS (`scripts/17-go-manual-direct-full-desktop-harness/main.go:98`),
5. waits for bus EOS/error with `TimedPopFiltered(...)` (`scripts/17-go-manual-direct-full-desktop-harness/main.go:102`),
6. optionally dumps dot graphs (`scripts/17-go-manual-direct-full-desktop-harness/main.go:207`).

That matters because it gives us a Go-hosted control with very little obvious Go-side steady-state work.

## High-Signal Repro and Success Criteria

### High-signal repro

For the direct hosting-gap slice, the stable control workload is:

- full desktop root capture on `:0`
- resolution `2880x1920`
- `24 fps`
- H.264 via `x264enc`
- quality `75` / bitrate `6920`
- `mov` / `qtmux`
- roughly `8 seconds`

### What counts as a successful capture

For this report, a successful run means all of the following are true:

1. the output file finalizes successfully and `ffprobe` reports valid H.264 video,
2. the intended resolution and frame rate are present,
3. the graph dump or command line reflects the intended direct pipeline shape,
4. CPU or perf artifacts are saved under the ticket-local `scripts/results/` tree,
5. the comparison uses the same source size and encode settings on both sides.

These criteria matter because earlier SCS-0016 eliminations proved that a “low CPU” number can be misleading if the pipeline actually errored or terminated incorrectly.

## Experiment Narrative and Findings

### 1. The matched `gst-launch-1.0` control is much cooler

The saved matched control run is:

- `scripts/results/16-gst-launch-direct-full-desktop-match-app/20260415-004852/01-summary.md`

That artifact reports:

- average CPU `128.00%` (`.../01-summary.md:10`)
- max CPU `137.00%` (`.../01-summary.md:11`)
- valid `2880x1920` / `24/1` output (`.../01-summary.md:21-22`)

This was the first major pivot in the investigation. It proved that the direct full-desktop encode workload is not *inherently* a guaranteed `~400–600%` CPU path on this machine. A cooler native control existed.

### 2. The copied/app-like Go harness initially had a real graph bug: `Y444`

The first copied/app-like harness run is:

- `scripts/15-go-direct-record-full-desktop-harness/results/20260415-005501/01-summary.md`

That run reported:

- average CPU `245.88%` (`.../01-summary.md:9`)
- max CPU `597.00%` (`.../01-summary.md:10`)

The corresponding playing dot graph showed the real mistake. Instead of feeding `I420` into the encoder, it fed `Y444`:

- encoder input `format: Y444` at `.../20260415-005501/dot/...-playing.dot:124`
- upstream negotiated `format: Y444` on the same branch at `...-playing.dot:147` and `...-playing.dot:170`

This explained why the copied/app-like Go path could be much more expensive than expected even before we started talking about Go-hosting effects.

### 3. The fully manual Go graph was cooler, which sharpened the comparison

The first manual direct harness run is:

- `scripts/17-go-manual-direct-full-desktop-harness/results/20260415-005343/01-summary.md`

It reported:

- average CPU `172.75%` (`.../01-summary.md:9`)
- max CPU `334.00%` (`.../01-summary.md:10`)
- valid `2880x1920` / `24/1` output (`.../01-summary.md:21-22`)

This was an important intermediate result because it suggested two distinct hypotheses:

1. maybe the app/copy-of-app path had a graph-shape problem,
2. maybe Go-hosting itself still had some residual effect beyond graph shape.

### 4. Fixing the app-like path to `I420` was correct, but not sufficient

The app/runtime direct path was fixed at the source by changing the capsfilter in `pkg/media/gst/recording.go:562` to explicitly pin `format=I420`.

The post-fix copied/app-like harness rerun is:

- `scripts/15-go-direct-record-full-desktop-harness/results/20260415-010455/01-summary.md`

Its graph dump now shows:

- encoder input `format: I420` at `.../20260415-010455/dot/...-playing.dot:124`
- upstream negotiated `format: I420` at `...-playing.dot:147` and `...-playing.dot:170`

So the graph fix was real and should be kept.

However, the rerun still reported a hot path:

- average CPU `320.62%` (`.../01-summary.md:9`)
- max CPU `540.59%` (`.../01-summary.md:10`)

That means the `Y444` bug was real but **not** the whole story.

### 5. After the `I420` fix, the two Go-hosted graphs converged

A normalized text diff of the playing dot files for the fixed copied/app-like harness and the fully manual harness returned `IDENTICAL_AFTER_NORMALIZATION` during the investigation. The practical point is simpler than the mechanics of the diff:

- same source,
- same transform chain,
- same raw caps,
- same encoder settings,
- same parser/mux/filesink path.

In other words, once the `I420` fix landed, the remaining discrepancy could no longer be blamed on “the DSL built the wrong graph” or “the manual graph is materially different.”

### 6. The 3x3 A/B matrix still showed a Go-hosted gap, even with graph identity

The saved A/B matrix summary is:

- `scripts/results/18-direct-harness-ab-matrix/20260415-010905/02-summary.md`

It reports:

- copied/app-like harness mean avg CPU `381.29` with peak max `633.00` (`.../02-summary.md:8`)
- manual harness mean avg CPU `327.73` with peak max `547.00` (`.../02-summary.md:9`)

This did two useful things:

1. it showed that the copied/app-like path is still usually hotter than the manual path,
2. but it also showed that **both** Go-hosted versions remain much hotter than the matched `gst-launch` control.

So the remaining problem is no longer merely “app helper path bad, manual Go path good.”

### 7. Mixed-stack perf says the direct running pipeline is mostly native x264 + GStreamer work

The saved perf compare summary is:

- `scripts/results/19-perf-compare-go-manual-vs-gst-launch/20260415-011757/01-summary.md`

For the Go manual harness, the top entries are:

- `libx264.so.164 x264_8_trellis_coefn` at `42.48%` / `38.23%` (`.../01-summary.md:11`)
- `libc.so.6 clone3` at `24.31%` (`.../01-summary.md:13`)
- `libgstreamer gst_pad_push` at `12.91%` (`.../01-summary.md:16`)

For `gst-launch`, the top entries are still the same basic native family:

- `libx264.so.164 x264_8_trellis_coefn` at `42.09%` / `41.44%` (`.../01-summary.md:28`)
- `libc.so.6 clone3` at `34.93%` (`.../01-summary.md:29`)
- `libgstreamer gst_pad_push` at `19.29%` (`.../01-summary.md:32`)

The important interpretive sentence is already recorded in the same summary file:

- directly visible Go/cgo frames in the Go-hosted manual harness are tiny, e.g. `runtime.asmcgocall.abi0` around `0.37%` (`.../01-summary.md:45`)

The raw dso-sorted report backs that up:

- `runtime.asmcgocall.abi0` at `0.37%` in `.../go-manual/perf-report-dso-symbol.txt:1930`
- `gst_pad_push` as a major native hot function at `.../perf-report-dso-symbol.txt:510`
- `x264_8_trellis_coefn` dominating the profile at `.../perf-report-dso-symbol.txt:12`

That does **not** mean Go is irrelevant. It does mean that the direct pipeline does **not** look like a per-frame Go processing path.

### 8. Thread counts are close; page faults are not

The best current quantitative comparison is the threadgroup-based `perf stat` run:

- `scripts/results/21-perf-stat-threadgroup-compare-go-vs-gst-launch/20260415-012510/01-summary.md`

The per-scenario summaries show:

#### Go manual harness

- threads observed `1..25` (`.../go-manual/01-summary.md:4`)
- attached tids `24` (`.../go-manual/01-summary.md:6`)
- task-clock `26480.36 ms` (`.../go-manual/01-summary.md:10`)
- page-faults `134832` (`.../go-manual/01-summary.md:13`)

#### `gst-launch`

- threads observed `1..22` (`.../gst-launch/01-summary.md:4`)
- attached tids `22` (`.../gst-launch/01-summary.md:6`)
- task-clock `11327.53 ms` (`.../gst-launch/01-summary.md:10`)
- page-faults `232` (`.../gst-launch/01-summary.md:13`)

This is the strongest current clue in the entire report.

The thread counts are similar. Context switches and migrations are not so wildly different that they independently explain the gap. But page faults differ by **orders of magnitude**.

That strongly suggests the Go-hosted case is changing native runtime behavior in a memory-related way rather than merely creating more Go application work.

## What This Investigation Has Ruled Out

### Ruled out: “the browser/web path is required for the hot spike”

Earlier SCS-0016 work had already shown that no-preview, no-audio, direct-recording-only cases could still stay hot. This report does not repeat all of that evidence, but it depends on that earlier narrowing.

### Ruled out: “the app DSL is constructing the wrong media graph”

The direct path is not built by a generic media graph DSL in the first place. It is built imperatively in Go (`pkg/media/gst/recording.go:526-585`). Once the `I420` caps fix landed, the manual and copied/app-like playing dot graphs converged.

### Ruled out: “frames are obviously traveling through Go code on every frame in the direct path”

The manual direct harness does not route frame payloads through Go. Its steady-state Go involvement is mostly control-plane logic around EOS and bus polling (`scripts/17-go-manual-direct-full-desktop-harness/main.go:98-102`). The mixed-stack perf report shows only tiny visible Go/cgo frames relative to the native x264/GStreamer hot path.

### Ruled out: “thread count explosion is the main explanation”

The warmed-up threadgroup compare shows roughly 24 attached tids in the Go-hosted manual harness and 22 in the `gst-launch` process. That is a real difference, but it is not the kind of enormous fan-out that would by itself explain the gap.

## What Still Looks Likely

### Most likely current explanation

The strongest current model is:

> the Go-hosted process changes native runtime behavior enough that the direct full-desktop x264 pipeline performs substantially more memory-related work than the equivalent `gst-launch` process, even though the realized graph and dominant native hot functions are broadly similar.

### Why page faults matter so much here

The extreme fault delta suggests at least one of these broad classes of explanation:

1. **allocator / memory-layout differences** between the Go-hosted process and `gst-launch`,
2. **anonymous page first-touch / compaction / THP interaction** inside hot native encoder paths,
3. **buffer-pool / allocation reuse differences** triggered indirectly by hosting/runtime conditions,
4. **thread-start / stack / faulting behavior** that is not visible in graph dumps but is visible in kernel/native counters.

This does **not** yet prove which one is correct. It does tell us where the next work should focus.

## eBPF-Relevant Unanswered Questions

These are the questions that still justify lower-level tracing after the current pprof/perf work:

1. Are the extra Go-hosted page faults mostly **minor** faults, mostly **major** faults, or dominated by a smaller subset of threads?
2. Are the hottest Go-hosted native threads spending materially more time in memory-management paths such as page fault handling or compaction?
3. Is there an off-CPU difference between the Go-hosted and `gst-launch` cases that the current on-CPU-heavy perf reports are smoothing over?
4. Are particular worker threads in the Go-hosted case repeatedly faulting/allocating while similarly named threads in `gst-launch` are not?
5. Does changing THP policy, allocator settings, or warmup behavior materially alter the page-fault delta?

## Recommended Next Steps

### Keep the `I420` fix

The `I420` direct-path caps fix in `pkg/media/gst/recording.go:562` was correct and should remain. The pre-fix `Y444` graph was a real bug, not noise.

### Do not spend more time blaming the DSL or graph builder

At this point, the investigation would lose time by circling back to generic “maybe Go built the graph wrong” theories. The direct graph has already been examined, corrected, dumped, and compared.

### Next measurement targets

1. Extend the threadgroup comparison to capture **minor vs major faults** explicitly.
2. Save `/proc/<pid>/status`, `/proc/<pid>/smaps_rollup`, and THP-related context during the hot run.
3. Compare allocator- and paging-sensitive environment perturbations one at a time, for example:
   - THP policy,
   - `MALLOC_ARENA_MAX`,
   - controlled warmup / pre-faulting.
4. Use eBPF only for these now-targeted questions, not as a generic fishing expedition.

### What should *not* be the next product code change

The current evidence does **not** justify more product-path rewrites to preview management, browser streaming, or arbitrary pipeline graph changes. The next useful changes should be investigation-oriented and memory/fault-oriented.

## References

### Code

- `pkg/media/gst/recording.go:526-585` — direct video recording pipeline builder
- `pkg/media/gst/recording.go:562-573` — fixed raw caps plus explicit x264 settings
- `pkg/media/gst/bus.go:16-40` — GLib bus-watch integration in the production runtime
- `scripts/17-go-manual-direct-full-desktop-harness/main.go:98-102,207` — minimal Go-hosted direct harness control plane

### Key artifacts

- `scripts/results/16-gst-launch-direct-full-desktop-match-app/20260415-004852/01-summary.md:10-22`
- `scripts/15-go-direct-record-full-desktop-harness/results/20260415-005501/01-summary.md:9-22`
- `scripts/15-go-direct-record-full-desktop-harness/results/20260415-005501/dot/0.00.01.015137962-direct-full-desktop-playing.dot:124,147,170`
- `scripts/15-go-direct-record-full-desktop-harness/results/20260415-010455/01-summary.md:9-22`
- `scripts/15-go-direct-record-full-desktop-harness/results/20260415-010455/dot/0.00.01.020657234-direct-full-desktop-playing.dot:124,147,170`
- `scripts/17-go-manual-direct-full-desktop-harness/results/20260415-005343/01-summary.md:9-22`
- `scripts/results/18-direct-harness-ab-matrix/20260415-010905/02-summary.md:6-9`
- `scripts/results/19-perf-compare-go-manual-vs-gst-launch/20260415-011757/01-summary.md:11-16,28-32,45`
- `scripts/results/19-perf-compare-go-manual-vs-gst-launch/20260415-011757/go-manual/perf-report-dso-symbol.txt:12,122,510,1930`
- `scripts/results/21-perf-stat-threadgroup-compare-go-vs-gst-launch/20260415-012510/01-summary.md:3-5`
- `scripts/results/21-perf-stat-threadgroup-compare-go-vs-gst-launch/20260415-012510/go-manual/01-summary.md:4-13`
- `scripts/results/21-perf-stat-threadgroup-compare-go-vs-gst-launch/20260415-012510/gst-launch/01-summary.md:4-13`

## Usage Examples

### Use this report as a handoff brief

If a new investigator joins the ticket, point them to this file first, then to the raw artifacts in the References section.

### Use this report to scope future experiments

If the next question is “what should we measure next?”, use the `Recommended Next Steps` and `eBPF-Relevant Unanswered Questions` sections as the decision boundary.

## Related

- `design-doc/01-low-level-profiling-plan.md`
- `reference/01-investigation-diary.md`
- `reference/02-performance-investigation-approaches-and-tricks-report.md`
- `reference/05-online-research-query-packet-for-go-hosted-gstreamer-performance.md`
