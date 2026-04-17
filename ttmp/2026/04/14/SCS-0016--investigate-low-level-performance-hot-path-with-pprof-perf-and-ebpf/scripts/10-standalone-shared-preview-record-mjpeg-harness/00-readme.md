---
Title: 10 standalone shared preview + record + MJPEG harness
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
RelatedFiles: []
ExternalSources: []
Summary: Ticket-local script/readme artifact for SCS-0016.
LastUpdated: 2026-04-15T01:45:00-04:00
WhatFor: Preserve this local script/readme artifact as part of the reproducible SCS-0016 investigation surface.
WhenToUse: Read when rerunning or reviewing this specific script directory inside SCS-0016.
---

# 10 standalone shared preview + record + MJPEG harness

This harness is a self-contained standalone binary for the lower-level SCS-0016 investigation.

It intentionally does **not** reuse the web server stack. Instead it isolates:

- one shared GST source
- one preview branch via the exported GST preview runtime
- one recording branch via the exported GST recording runtime
- Go-side JPEG frame copying into a local frame store
- a minimal local MJPEG HTTP handler
- an optional separate child process that connects to the MJPEG endpoint and drains the stream

It intentionally omits:

- PreviewManager
- preview.state publication
- EventHub
- websocket handling
- frontend / browser app logic
- the main web server

## Why this exists

This is the missing middle slice between:

- pure native GST/bridge experiments
- and the full real web/browser path

It lets us answer whether the combination of:

- shared source
- preview branch
- recording branch
- Go frame copies
- MJPEG HTTP serving

is already enough to reproduce the hot path without PreviewManager/websocket/server machinery.

## Files

- `main.go` — standalone harness with `harness` and `mjpeg-client` modes
- `run.sh` — convenience wrapper that runs the harness and captures pidstat plus a markdown summary

## Example

```bash
bash ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/10-standalone-shared-preview-record-mjpeg-harness/run.sh
```

For a region source:

```bash
RECT=0,0,1280,720 SOURCE_TYPE=region bash ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/10-standalone-shared-preview-record-mjpeg-harness/run.sh
```
