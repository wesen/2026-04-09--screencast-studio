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
LastUpdated: 2026-04-14T20:46:00-04:00
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

## Current status

Current status: **active**

Current deliverable status:

- ticket created
- detailed low-level profiling plan written
- diary started
- next implementation slice chosen: optional local pprof enablement plus reproducible profile-capture scripts

## Tasks

See `tasks.md` for the checklist.

## Changelog

See `changelog.md` for the delivery record.
