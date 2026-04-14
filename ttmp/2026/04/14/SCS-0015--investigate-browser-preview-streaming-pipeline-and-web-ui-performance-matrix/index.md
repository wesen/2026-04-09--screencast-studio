---
Title: Investigate browser preview streaming pipeline and web UI performance matrix
Ticket: SCS-0015
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
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go
      Note: Browser-facing MJPEG transport is implemented here
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go
      Note: Preview state, cached frames, and fan-out behavior live here
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go
      Note: The upstream shared preview branch already performs JPEG encoding before Go handles the frame bytes
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx
      Note: Frontend preview ownership and lifecycle are orchestrated here
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/preview/PreviewStream.tsx
      Note: Browser rendering uses img tags pointed at MJPEG URLs
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/metrics/metrics.go
      Note: New metrics registry will be extended for browser-preview observability
ExternalSources: []
Summary: Ticket for investigating the browser-connected preview streaming path and measuring why the real Studio page can drive the server much hotter than earlier backend/API-only performance matrices suggested.
LastUpdated: 2026-04-14T15:44:00-04:00
WhatFor: Track the browser preview streaming investigation, its measurement plan, and the final report.
WhenToUse: Start here when working on browser-preview serving overhead, frontend preview lifecycle performance, or MJPEG/metrics measurement work.
---

# Investigate browser preview streaming pipeline and web UI performance matrix

## Overview

This ticket isolates the browser-connected preview path as its own investigation track.

Earlier work in SCS-0014 already measured much of the backend/shared-runtime behavior around preview and recording. The remaining gap is the **real browser UI path**:

- MJPEG preview serving from the backend,
- browser tabs holding live preview listeners,
- frontend preview ensure/release behavior,
- and the possibility that the Studio page keeps the server much hotter than API-only harnesses suggested.

The user specifically reported that pressing record in the real web UI can drive server CPU far higher than the earlier standalone/backend matrices implied, including very high CPU even for desktop-only recording through the UI. This ticket is the structured follow-up to explain that discrepancy.

## Primary documents

- `design/01-browser-preview-streaming-pipeline-analysis-and-performance-matrix-plan.md`
- `design/02-browser-preview-streaming-performance-report.md`
- `reference/01-investigation-diary.md`

## Current status

Current status: **active**

Current deliverable status:

- ticket created
- primary analysis/plan document written
- diary started
- final report scaffold created
- current browser preview transport and frontend lifecycle mapped from code
- browser-specific measurement matrix defined
- preview-serving metrics have been added for active MJPEG clients, stream starts/finishes, frames, bytes, flushes, frame updates, and preview ensure/release events
- focused metrics tests and `/metrics` validation are complete
- initial runtime helper scripts now exist for local server restart and `/metrics` sampling, with the first saved smoke result under `scripts/results/20260414-160358/`
- a first desktop preview HTTP-client baseline matrix has been added and run under `scripts/03-desktop-preview-http-client-matrix/results/20260414-161024/`
- early baseline result: 0 and 1 MJPEG client looked similar in this short run, while 2 clients pushed server CPU higher
- real browser-tab matrix harnesses are still pending

## Tasks

See `tasks.md` for the checklist.

## Changelog

See `changelog.md` for the delivery record.
