---
Title: Serve cancellation and subprocess lifecycle hardening
Ticket: SCS-0007
Status: active
Topics:
    - screencast-studio
    - backend
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Ticket workspace for designing proper serve-mode cancellation, subprocess cleanup, and lifecycle observability."
LastUpdated: 2026-04-10T09:31:48.028102181-04:00
WhatFor: "Track investigation, design, and eventual implementation planning for robust Ctrl-C and serve shutdown behavior."
WhenToUse: "Use when working on serve lifecycle management, subprocess cleanup, or shutdown diagnostics."
---

# Serve cancellation and subprocess lifecycle hardening

## Overview

This ticket tracks the work needed to make `screencast-studio serve` shut down cleanly and observably. The motivating user report is that `Ctrl-C` can appear to hang and can leave `ffmpeg` running in the background. The current repository already contains some cancellation logic, but the runtime ownership model is not yet fully unified across the HTTP server, recording manager, preview manager, and telemetry manager.

The goal of this ticket is to replace ad hoc shutdown fixes with an evidence-based implementation plan. The main design deliverable explains the current architecture, identifies the ownership and cancellation gaps, and proposes a phased implementation plan that a new engineer can follow safely.

## Key Links

- Design doc: [`design-doc/01-serve-cancellation-subprocess-shutdown-and-observability-implementation-guide.md`](./design-doc/01-serve-cancellation-subprocess-shutdown-and-observability-implementation-guide.md)
- Diary: [`reference/01-investigation-diary.md`](./reference/01-investigation-diary.md)
- Tasks: [`tasks.md`](./tasks.md)
- Changelog: [`changelog.md`](./changelog.md)

## Current Status

Current status: **active**

Completed in this ticket so far:

- ticket workspace created,
- current cancellation architecture inspected,
- intern-oriented design guide written,
- investigation diary written.

Still pending:

- relate files,
- run `docmgr doctor`,
- upload the bundle to reMarkable,
- implement the actual runtime changes in a later coding pass.

## Recommended Reading Order

For a new engineer:

1. Read the design doc executive summary and current-state architecture sections.
2. Read the diary for the exact commands, failed attempts, and rationale for stepping back.
3. Review the key code files referenced in the design doc.
4. Use the task list as the implementation checklist.

## Structure

- `design-doc/` — primary architecture and implementation guidance
- `reference/` — chronological diary and future quick references
- `scripts/` — any future reproduction helpers or shutdown test harnesses
- `sources/` — future saved logs or external references if needed
- `archive/` — future deprecated drafts or superseded materials
