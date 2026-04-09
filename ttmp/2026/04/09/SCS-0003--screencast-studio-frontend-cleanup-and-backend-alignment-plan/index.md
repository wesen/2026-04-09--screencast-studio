---
Title: Screencast Studio Frontend Cleanup and Backend Alignment Plan
Ticket: SCS-0003
Status: active
Topics:
    - frontend
    - backend
    - ui
    - architecture
    - dsl
    - video
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-09T19:18:00-04:00
WhatFor: Track the dedicated frontend cleanup effort and link to the architecture guide, diary, and task plan for replacing stale frontend integration paths.
WhenToUse: Read when starting or reviewing the frontend cleanup effort, or when orienting a new engineer on why this follow-up ticket exists.
---

# Screencast Studio Frontend Cleanup and Backend Alignment Plan

## Overview

This ticket exists because the current frontend in `ui/` is no longer just “under construction.” It now needs explicit cleanup. The backend web transport exists and is concrete, but the frontend still contains stale API assumptions, duplicate top-level shells, mock handlers that hide drift, and local demo state that should be replaced by backend-driven state.

This ticket is the cleanup and replacement plan. It assumes no backwards compatibility work is needed. The objective is to keep the good visual and component work, while deleting or rewriting the frontend integration code that no longer matches the backend or the intended product model.

## Key Links

- Main design guide:
  - `design-doc/01-frontend-cleanup-and-backend-alignment-system-design.md`
- Diary:
  - `reference/01-diary.md`
- Task list:
  - `tasks.md`
- Precursor assessment:
  - `../SCS-0002--screencast-studio-web-control-frontend-architecture-and-implementation-plan/design-doc/02-frontend-assessment-and-improvement-guide.md`

## Status

Current status: **active**

## Topics

- frontend
- backend
- ui
- architecture
- dsl
- video

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design-doc/ - Architecture and cleanup design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
