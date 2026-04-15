---
Title: Low-level profiling plan
Ticket: SCS-0016
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - backend
    - analysis
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/event_hub.go
      Note: |-
        Earlier timing work already narrowed EventHub publish overhead
        SCS-0015 already narrowed EventHub publish overhead before this ticket started
    - Path: internal/web/handlers_preview.go
      Note: |-
        Earlier timing work already narrowed the final MJPEG write loop
        SCS-0015 already narrowed the final MJPEG write loop before this ticket started
    - Path: internal/web/preview_manager.go
      Note: |-
        Earlier timing work already narrowed PreviewManager cached-frame overhead
        SCS-0015 already narrowed PreviewManager cached-frame overhead before this ticket started
    - Path: pkg/cli/serve.go
      Note: |-
        Candidate location for optional pprof enablement flags
        Likely place to add optional local profiling enablement
    - Path: pkg/media/gst/shared_video.go
      Note: |-
        Likely next hot boundary around appsink callbacks, frame mapping, and Go handoff
        Likely next lower-level boundary around appsink callbacks and buffer handoff
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/reference/02-browser-preview-streaming-lab-report.md
      Note: Current evidence trail that motivates this lower-level profiling phase
ExternalSources: []
Summary: Detailed plan for lower-level profiling of the browser-connected recording hot path using pprof first, then perf and eBPF only as needed.
LastUpdated: 2026-04-14T20:46:00-04:00
WhatFor: Turn the narrowed SCS-0015 browser-path mystery into a disciplined lower-level profiling campaign.
WhenToUse: Read before adding profiling endpoints, scripts, or low-level profiler runs so the work stays staged and question-driven.
---


# Low-level profiling plan

## Executive Summary

SCS-0015 already did the right app-level narrowing. The real browser-connected hot path is still real, but several higher-level suspects now look too cheap to explain it: final MJPEG write/flush, PreviewManager cached-frame copy/store, `preview.state` publication, and EventHub publish time. That is exactly the point where lower-level profilers become useful rather than premature.

This ticket will therefore proceed in stages. First, add an **optional local pprof path** and use it to capture CPU profiles during the highest-value repro: **desktop preview + recording + one real browser tab**. If pprof clearly explains the hot phase in Go userland, stay there. If pprof mostly points at `runtime.cgocall` or leaves the core cost unexplained, move to **`perf`** so Go, CGO, libc, GStreamer, and kernel stacks can be seen together. Only if there are still unanswered, targeted questions after that should we use **eBPF**, and then only for narrow scheduler/off-CPU/syscall/socket questions.

## Problem Statement

The current browser-path evidence says:

- the real Studio-page one-tab recording case can still climb into the `~380–410%` CPU band,
- but final MJPEG write/flush work is tiny,
- PreviewManager store/copy/publication work is tiny,
- and EventHub publish time is also tiny.

That means the remaining hot path is likely below the current app-level metrics surface. Plausible remaining boundaries include:

- GStreamer `appsink` callback cost,
- CGO transitions,
- memcpy-heavy buffer mapping/copying,
- scheduler/runtime behavior under combined preview+recording load,
- or some lower-layer interaction between browser-connected preview consumption and the shared recording path.

The project now needs profiler evidence that can actually see those layers.

## Proposed Solution

### Phase 1: add optional pprof support and reproducible capture scripts

Add a deliberately opt-in profiling path for local investigation. The current leading implementation idea is:

- add a `--pprof-addr` flag to `screencast-studio serve` with an empty default,
- if set, start a separate debug server exposing standard Go `net/http/pprof` handlers,
- keep it disabled by default so normal local usage is unchanged,
- add ticket-local scripts under `scripts/` that:
  - restart the app with pprof enabled,
  - drive or coordinate the existing high-signal browser repro,
  - capture CPU profiles during the hot recording window,
  - save both raw profile artifacts and human-readable summaries.

The first question pprof should answer is straightforward:

> Is the hot phase primarily visible in Go code, or does it mostly disappear into CGO/runtime boundaries?

### Phase 2: move to perf only if pprof is not explanatory enough

If pprof mainly shows:

- `runtime.cgocall`,
- too much time in opaque runtime frames,
- or otherwise fails to localize the dominant hot cost,

then add reproducible `perf` capture scripts for the same high-signal repro. The point of `perf` is that it can see a mixed stack:

- Go,
- CGO,
- libc,
- GStreamer,
- kernel/syscall layers.

The output artifacts should include:

- raw `perf.data`,
- saved `perf report` text,
- and, if practical, stack-collapsed/flamegraph-style outputs.

### Phase 3: use eBPF only for targeted unanswered questions

eBPF should not be the default first move. It is best reserved for narrow questions such as:

- is there significant off-CPU time,
- are scheduler delays or wakeups the real issue,
- are socket/syscall patterns weird under browser load,
- is there hidden blocking that plain CPU profiles do not show.

So eBPF is the fallback for **specific** remaining questions after pprof/perf, not a generic exploratory shotgun step.

## Design Decisions

### 1. Split this work into a new ticket instead of bloating SCS-0015

Rationale:
- SCS-0015 is already the right place for browser-path measurement matrices and app-level reasoning.
- Lower-level profiling has different artifacts, tools, risks, and interpretation rules.
- Keeping it separate will make both tickets easier to continue and review.

### 2. Start with pprof, not perf or eBPF

Rationale:
- easiest to add and capture reproducibly,
- lowest-friction answer to whether the hot phase is still in Go userland,
- good enough if the hotspot is actually in Go logic.

### 3. Keep profiling opt-in and local-only by default

Rationale:
- profiling endpoints and extra debug servers should not be enabled accidentally,
- lower-level investigation often needs experimental or privileged tooling,
- the normal `serve` path should remain stable and unsurprising.

### 4. Reuse the same high-signal repro instead of broadening scenarios immediately

Rationale:
- SCS-0015 already established the strongest scenario: **desktop preview + recording + one real browser tab**.
- Lower-level tools are easiest to compare when the repro stays stable.
- Broadening scenarios too early would make profiler interpretation noisier.

## Alternatives Considered

### Alternative A: Keep adding more app-level counters only

Rejected because:
- app-level counters already ruled out several plausible upper-layer suspects,
- the remaining cost likely lives beneath the current visibility surface.

### Alternative B: Jump straight to perf and skip pprof

Rejected for now because:
- pprof is easier to wire and lower-friction,
- if the hotspot is still mostly Go-side, pprof may answer the question cheaply.

### Alternative C: Jump straight to eBPF

Rejected for now because:
- eBPF is most useful once the question is narrowed,
- right now the first missing answer is still simply “where is the CPU going?”
- pprof and perf are better first tools for that question.

## Implementation Plan

### Slice 1: ticket setup and reproducibility plan

- create ticket docs and tasks
- write down the exact high-signal repro
- decide artifact layout under `scripts/`

### Slice 2: optional pprof enablement

- add `--pprof-addr` or equivalent local-only profiling flag
- start a separate pprof HTTP server when enabled
- add tests if reasonable for route/server startup behavior

### Slice 3: pprof capture scripts

- restart server with pprof enabled
- capture CPU profile during the hot recording window
- save raw profile and textual summaries under ticket-local `scripts/`

### Slice 4: interpret pprof result honestly

- if pprof explains the hot path, stop and summarize
- if pprof mostly points at CGO/runtime boundaries, proceed to perf

### Slice 5: perf capture scripts and analysis

- add reproducible `perf record` flow for the same repro
- save raw and summarized artifacts
- decide whether the dominant cost is Go, CGO, GStreamer, libc, or kernel

### Slice 6: eBPF only if still needed

- choose one narrow question
- add the smallest targeted eBPF capture that answers it
- do not broaden tooling until the previous tool clearly fails to answer the question

## Open Questions

- Should pprof live on the main server mux under `/debug/pprof`, or on a separate debug address? The current preference is a separate opt-in address.
- Will pprof be sufficient, or will the hot phase mostly collapse into `runtime.cgocall` and force a quick move to `perf`?
- Does the machine already have the necessary `perf` permissions and symbol fidelity for mixed Go/GStreamer profiling, or will setup work be needed first?
- If we need eBPF later, which tool family is already available on this machine: `bpftrace`, BCC tools, or something else?

## References

- `ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/reference/02-browser-preview-streaming-lab-report.md`
