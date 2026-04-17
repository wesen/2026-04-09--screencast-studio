---
Title: 15 go direct record full desktop harness
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

# 15 go direct record full desktop harness

Tiny standalone Go harness for the direct full-desktop recording control case.

It intentionally omits:
- preview
- MJPEG
- websocket
- shared source bridge
- audio

It records:
- `ximagesrc -> videoconvert -> videorate -> x264enc -> h264parse -> qtmux/mp4mux -> filesink`

## Example

```bash
bash ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/15-go-direct-record-full-desktop-harness/run.sh
```
