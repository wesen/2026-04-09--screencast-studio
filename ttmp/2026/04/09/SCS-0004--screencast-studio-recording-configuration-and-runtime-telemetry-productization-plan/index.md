---
Title: Screencast Studio Recording Configuration and Runtime Telemetry Productization Plan
Ticket: SCS-0004
Status: active
Topics:
    - frontend
    - backend
    - ui
    - audio
    - video
    - product
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ui/src/components/studio/OutputPanel.tsx
      Note: Current recording controls and destination UI that still contain placeholder behavior
    - Path: ui/src/components/studio/MicPanel.tsx
      Note: Current microphone panel that lacks real audio telemetry and discovered input options
    - Path: ui/src/pages/StudioPage.tsx
      Note: Current page-level orchestration that must be extended to use real recording configuration and telemetry
    - Path: internal/web/handlers_api.go
      Note: Current REST boundary for recording start and setup normalization that will likely expand for richer configuration flows
    - Path: internal/web/handlers_ws.go
      Note: Current websocket event stream that is the natural path for runtime telemetry delivery
ExternalSources: []
Summary: Follow-on productization ticket for making recording configuration, naming, destination paths, and runtime telemetry real in the mounted UI instead of placeholder controls.
LastUpdated: 2026-04-09T20:55:00-04:00
WhatFor: Track the work needed to turn the recording controls and telemetry panels into real product behavior.
WhenToUse: Read when implementing output configuration, session naming, destination previews, audio metering, or disk/runtime telemetry in the web UI.
---

# Screencast Studio Recording Configuration and Runtime Telemetry Productization Plan

## Overview

This ticket exists because the current UI now has a clean transport and preview foundation, but the recording-control side is still only partially productized. The output controls render well, yet several of them still operate as presentation-only widgets. The microphone panel exists, but it does not expose real discovered inputs or real live metering. The output panel exposes a `Save to` control, but that control is not wired to anything meaningful in the backend. There is also no explicit recording name field, no destination path preview, and no reliable way for a user to see exactly where the files will be written before they hit Record.

This ticket closes that gap. It focuses on making the recording configuration surface honest and useful: naming, destination directory, derived filenames, audio device choice, gain wiring, live metering, and disk/runtime telemetry.

## Key Links

- Main design guide:
  - `design-doc/01-recording-configuration-and-runtime-telemetry-system-design.md`
- Diary:
  - `reference/01-diary.md`
- Task list:
  - `tasks.md`
- Predecessor cleanup ticket:
  - `../SCS-0003--screencast-studio-frontend-cleanup-and-backend-alignment-plan/index.md`

## Status

Current status: **active**

## Topics

- frontend
- backend
- ui
- audio
- video
- product

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design-doc/ - Architecture and implementation guides
- reference/ - Investigation diary and quick references
- playbooks/ - Future operator workflows and manual test guides
- scripts/ - Ticket-local repro or helper scripts
- various/ - Notes and scratch artifacts
- archive/ - Deprecated or superseded material
