---
Title: Investigate low-level performance hot path with pprof, perf, and eBPF
Ticket: SCS-0016
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - backend
    - analysis
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go
      Note: Browser-facing MJPEG handler already ruled out as the dominant final-write hot path in SCS-0015
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go
      Note: PreviewManager timing instrumentation already narrowed several Go-side suspects before this ticket was created
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/event_hub.go
      Note: EventHub timing instrumentation already narrowed server-event fanout as the dominant cause
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go
      Note: Likely next low-level boundary around appsink callbacks and buffer handoff from GStreamer into Go
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/serve.go
      Note: Likely place to add optional low-level profiling flags for local investigation
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/reference/02-browser-preview-streaming-lab-report.md
      Note: Upstream ticket that narrowed the problem enough to justify this lower-level profiling track
ExternalSources: []
Summary: Separate ticket for low-level CPU profiling of the real browser-connected recording hot path using Go pprof first, then perf and eBPF if needed.
LastUpdated: 2026-04-14T22:54:41-04:00
WhatFor: Keep low-level profiling work separate from SCS-0015's higher-level browser-path measurement and interpretation work.
WhenToUse: Start here when working on pprof, perf, eBPF, flamegraphs, or lower-level runtime evidence for the browser-connected hot path.
---

# Investigate low-level performance hot path with pprof, perf, and eBPF

## Overview

SCS-0015 narrowed the browser-connected hot path significantly. The current evidence says the real Studio-page recording spike is not dominated by final MJPEG HTTP write/flush time, PreviewManager cached-frame copy/store, `preview.state` publication, or EventHub publish cost. That means the remaining unexplained cost is likely lower in the stack: around CGO, GStreamer buffer handoff, appsink callbacks, memcpy-heavy transitions, scheduler behavior, or some other boundary that the current app-level counters cannot see clearly.

This ticket exists to keep that lower-level investigation separate and disciplined. The plan is to start with **Go pprof** because it is the lowest-friction way to answer whether the CPU is really in Go userland. If pprof points mostly at `runtime.cgocall` or otherwise fails to explain the hot phase, the next step is **`perf`** so we can see Go, CGO, libc, GStreamer, and kernel time together. If ambiguity still remains after that, the fallback is **eBPF-based** tracing for scheduler, syscall, off-CPU, or socket-level behavior.

## Primary documents

- `design-doc/01-low-level-profiling-plan.md`
- `reference/01-investigation-diary.md`
- `reference/02-performance-investigation-approaches-and-tricks-report.md`
- `reference/03-prometheus-metrics-architecture-and-field-guide.md`

## Current status

Current status: **active**

Current deliverable status:

- ticket created
- detailed low-level profiling plan written
- diary started
- project report written for the investigation approaches and tricks used so far
- separate field guide written for the Prometheus-style metrics architecture and current metric families
- optional local pprof enablement has been added to `serve` via a separate debug address
- pprof restart/capture helper scripts now exist under `scripts/`
- the first real browser-connected CPU profile capture is saved under `scripts/results/02-capture-pprof-cpu-profile/20260414-204800/`
- the first pprof result mostly points at external/native code (`runtime._ExternalCode`, `<unknown>`, `libgstvideo`, `libc`, `libjpeg`) rather than clear Go-userland hotspots
- ticket-local profiler prereq checks are now recorded under `scripts/results/03-check-profiler-prereqs/20260414-211300/`
- the latest prereq check found `perf_event_paranoid=1` and a working `perf stat` path for the current user, while `bpftrace` still requires root
- a reproducible `perf` capture script now exists under `scripts/04-capture-perf-cpu-profile.sh`
- the first mixed-stack `perf` capture is saved under `scripts/results/04-capture-perf-cpu-profile/20260414-224952/`
- the first `perf` result points strongly at native recording work in `libx264` plus visible GStreamer push-path frames, while Go symbolization for the main binary is still weaker than ideal because the process is running from a temporary `go run` executable
- the exact browser-driving helpers and the Go-address resolver used for this slice have now been backfilled into the ticket-local `scripts/` directory with numbered prefixes
- two new project reports now exist: one on investigation approaches/tricks and one on the Prometheus-style metrics architecture

## Tasks

See `tasks.md` for the checklist.

## Changelog

See `changelog.md` for the delivery record.
