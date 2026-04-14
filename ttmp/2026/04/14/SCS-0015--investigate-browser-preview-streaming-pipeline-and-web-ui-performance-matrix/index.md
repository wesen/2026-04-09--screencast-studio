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
LastUpdated: 2026-04-14T17:48:00-04:00
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
- `reference/02-browser-preview-streaming-lab-report.md`

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
- a first desktop preview HTTP-client baseline matrix was added and run under `scripts/03-desktop-preview-http-client-matrix/results/20260414-161024/`
- a larger fresh-server HTTP-client matrix with recording now exists under `scripts/05-desktop-preview-http-client-recording-matrix/results/20260414-163154/`
- real browser scenario scripts now exist under `scripts/08-playwright-browser-matrix/`
- live browser-backed measurements now exist for desktop one-tab, desktop two-tab, and desktop-plus-camera one-tab scenarios under `scripts/results/20260414-163610/`, `20260414-163951/`, `20260414-164457/`, `20260414-164535/`, `20260414-164657/`, and `20260414-164720/`
- strongest current finding: the **browser-connected recording** path is the missing hot slice, with desktop one-tab preview+recording around `410.60%` avg CPU and desktop two-tab preview+recording around `432.97%`, far above the fresh-server plain-MJPEG-client recording baseline around `158–165%`
- a newer focused ablation now exists at `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173541/`
- that ablation showed one MJPEG client plus one synthetic websocket consumer only moved avg CPU from `166.56%` to `170.48%`, which lowers confidence that websocket fanout alone explains the browser spike
- a first human-readable findings note now exists at `scripts/09-browser-preview-matrix-findings-summary.md`
- a focused ablation summary note now exists at `scripts/13-mjpeg-websocket-ablation-summary.md`
- an ongoing detailed lab report now exists at `reference/02-browser-preview-streaming-lab-report.md` and backfills the current experiments in detail
- camera-only browser-tab scenarios are still pending, but the current priority is the desktop preview+recording one-tab repro rather than scenario expansion

## Tasks

See `tasks.md` for the checklist.

## Changelog

See `changelog.md` for the delivery record.
