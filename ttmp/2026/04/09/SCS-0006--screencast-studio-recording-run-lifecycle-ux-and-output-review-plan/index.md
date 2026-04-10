---
Title: Screencast Studio Recording Run Lifecycle UX and Output Review Plan
Ticket: SCS-0006
Status: active
Topics:
    - frontend
    - backend
    - ui
    - video
    - product
    - recording
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ui/src/pages/StudioPage.tsx
      Note: Current page owner for recording actions, logs, and tab-level lifecycle behavior
    - Path: ui/src/components/log-panel/LogPanel.tsx
      Note: Existing log panel that should become part of a fuller runtime lifecycle experience
    - Path: ui/src/components/studio/OutputPanel.tsx
      Note: Existing transport controls that need better run lifecycle UX around start, stop, and outputs
    - Path: ui/src/components/studio/StatusPanel.tsx
      Note: Existing status panel that should be part of a better run-state experience
    - Path: internal/web/session_manager.go
      Note: Backend session state source that drives the current lifecycle UX
    - Path: pkg/recording/events.go
      Note: Runtime event model that may need extension for better product-facing run lifecycle UX
ExternalSources: []
Summary: Follow-on product ticket for making recording start, running, stop, failure, and output review feel like a finished user workflow instead of a thin log-driven control strip.
LastUpdated: 2026-04-09T21:14:00-04:00
WhatFor: Track the work needed to productize the run lifecycle, result review, and output visibility surfaces of the application.
WhenToUse: Read when implementing better start/stop UX, output summaries, failure surfacing, or post-record review affordances.
---

# Screencast Studio Recording Run Lifecycle UX and Output Review Plan

## Overview

This ticket exists because the mounted UI can now start and stop recordings, stream logs, and expose session state, but the overall run lifecycle still feels like an engineering tool rather than a finished product. A user can press Record, but the app does not yet guide them clearly through “starting,” “recording,” “stopping,” “finished,” or “failed.” It also does not yet expose the produced outputs in a rich enough way once a run completes.

This ticket focuses on the lifecycle and result-review experience around a recording run.

## Key Links

- Main design guide:
  - `design-doc/01-recording-run-lifecycle-ux-and-output-review-system-design.md`
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
- video
- product
- recording

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design-doc/ - Architecture and implementation guides
- reference/ - Investigation diary and quick references
- playbooks/ - Future operator runbooks and validation guides
- scripts/ - Ticket-local helpers
- various/ - Notes and scratch artifacts
- archive/ - Deprecated or superseded material
