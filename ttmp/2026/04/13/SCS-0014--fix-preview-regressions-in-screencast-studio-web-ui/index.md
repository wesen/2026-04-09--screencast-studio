---
Title: Fix preview regressions in screencast-studio web UI
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - video
    - backend
    - ui
    - analysis
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/discovery/service.go
      Note: Camera discovery currently enumerates every /dev/video* node directly
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go
      Note: Shared preview source signatures and preview JPEG path explain several observed preview regressions
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go
      Note: Preview limit accounting and async release behavior explain the preview-limit race
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/conversion.ts
      Note: Source picker creates duplicate logical source IDs for the same physical device rather than deduplicating it
ExternalSources: []
Summary: Ticket for the preview regressions observed in the running screencast-studio web UI, now including implementation progress on camera deduplication, preview-quality tuning, preview-limit race hardening, a full postmortem on the window/region full-display bug, standalone recording-performance measurements for the shared GStreamer recording path, and a staged benchmark of the appsink-to-Go-to-appsrc bridge overhead.
LastUpdated: 2026-04-13T23:36:00-04:00
WhatFor: Track and document the preview bugs observed during live manual testing of the running screencast-studio web UI.
WhenToUse: Start here when triaging or fixing the preview regressions reproduced in the running app and documented from real logs and UI behavior.
---

# Fix preview regressions in screencast-studio web UI

## Overview

This ticket captures the concrete preview regressions observed while manually testing the running `screencast-studio` web UI against the real backend. The evidence comes from:

- live `tmux` server logs,
- the current `/api/discovery` snapshot,
- the visible UI state in the reported screenshot,
- and the current code paths in discovery, preview management, shared video capture, and the source picker.

The goal is not only to list symptoms. Each bug report also includes a fixing analysis and an implementation plan so the ticket is directly actionable.

## Bug Reports

- `design-doc/01-webcam-preview-remove-readd-device-churn-and-duplicate-camera-source-bug-report.md`
- `design-doc/02-webcam-preview-quality-and-format-regression-bug-report.md`
- `design-doc/03-second-desktop-capture-collapses-onto-root-x11-display-bug-report.md`
- `design-doc/04-preview-limit-race-and-stale-preview-recovery-bug-report.md`
- `design-doc/05-window-and-region-preview-full-display-postmortem.md`

## Status

Current status: **active**

Current deliverable status:
- Ticket created
- Four bug reports written
- Diary started and updated with implementation steps
- Camera discovery deduplication and duplicate-camera UI prevention implemented
- Preview quality/profile tuning implemented
- Preview-limit race/stopping-preview reuse hardening implemented
- Full postmortem written for the window/region full-display bug
- Standalone recording-performance harnesses and saved result directories added under `scripts/`
- Measurements show the recording CPU jump is dominated by the x264 encode path and is amplified further by the current shared raw-consumer/appsrc bridge topology
- A second staged bridge-overhead benchmark now isolates `appsink`, Go buffer copy, async queueing, `appsrc`, and `x264`
- A same-session reconciliation run now compares the direct GStreamer, shared-runtime, and staged bridge benchmarks side by side and shows recorder-only CPU is broadly aligned across those three paths; the larger remaining cost spike is preview + recorder together
- A dedicated shared-source preview/recorder interplay benchmark now shows that preview + recorder together is dramatically more expensive than recorder-only, and that cheaper preview settings help but do not remove the spike
- A follow-up preview-branch ablation benchmark now isolates second-branch cost, JPEG cost, raw frame-copy cost, the current preview path, and a cheap preview profile while recording
- `docmgr doctor` passed cleanly
- Bug report bundle uploaded to reMarkable and verified in `/ai/2026/04/13/SCS-0014`

## Tasks

See `tasks.md` for the ticket checklist.

## Changelog

See `changelog.md` for the delivery record.
