---
Title: Scripts README
Ticket: SCS-0001
Status: active
Topics:
    - backend
    - video
    - cli
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/direct-ffmpeg-region-smoke.sh
      Note: Direct FFmpeg baseline repro
    - Path: ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/inspect-record-processes.sh
      Note: Process inspection helper
    - Path: ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/smoke-record-region.sh
      Note: CLI smoke repro for bounded recording
ExternalSources: []
Summary: Quick index for the ticket-local repro and inspection scripts used during the CLI recording investigation.
LastUpdated: 2026-04-09T15:00:00-04:00
WhatFor: Provide a lightweight index of the scripts stored with the ticket.
WhenToUse: Read when rerunning the recording investigation or locating the exact shell repros used during development.
---

# Scripts

These scripts capture the concrete shell repros used during the CLI recording runtime investigation.

- `smoke-record-region.sh` creates a temporary DSL file and runs `screencast-studio record --duration 1` against a small X11 region.
- `direct-ffmpeg-region-smoke.sh` runs a direct `ffmpeg` capture against the same X11 region to separate runtime bugs from host-environment capture issues.
- `inspect-record-processes.sh` lists relevant `screencast-studio`, `ffmpeg`, and smoke-test processes during debugging.

Run them from the repository root.
