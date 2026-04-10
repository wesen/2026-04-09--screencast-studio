---
Title: Screencast Studio Source Management UX and Setup Builder Plan
Ticket: SCS-0005
Status: active
Topics:
    - frontend
    - backend
    - ui
    - dsl
    - video
    - product
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ui/src/components/studio/SourceGrid.tsx
      Note: Current source-grid UI that is still largely read-only in the mounted app
    - Path: ui/src/components/source-card/SourceCard.tsx
      Note: Current source-card rendering that lacks full interactive source-management behavior
    - Path: ui/src/pages/StudioPage.tsx
      Note: Current page owner that derives sources from normalized DSL but does not yet provide structured setup editing
    - Path: ui/src/api/discoveryApi.ts
      Note: Existing discovery transport that should power the setup builder
    - Path: pkg/dsl/types.go
      Note: Current DSL model that the setup builder must construct without drifting from backend truth
ExternalSources: []
Summary: Follow-on product ticket for making source creation, selection, editing, and setup-building a real structured UI instead of a read-only normalized source view plus raw DSL editing.
LastUpdated: 2026-04-09T21:05:00-04:00
WhatFor: Track the work needed to turn source management into a real product workflow.
WhenToUse: Read when implementing discovery-driven source creation, source editing, or a structured setup builder over the DSL model.
---

# Screencast Studio Source Management UX and Setup Builder Plan

## Overview

This ticket exists because the current web UI now displays real sources and previews, but it still does not let the user build or manage sources in a structured way. The mounted app derives sources from normalized DSL and renders them read-only. That is useful for validation and previewing, but it is not enough for a finished product. A real user should be able to browse discovered displays, windows, regions, cameras, and audio inputs, add them to the setup, edit their properties, reorder them, enable or disable them, and only drop to raw DSL when they explicitly want advanced control.

This ticket is the structured setup-builder ticket. It is not about recording telemetry or run lifecycle messaging. It is about how a user defines what they want to capture.

## Key Links

- Main design guide:
  - `design-doc/01-source-management-ux-and-setup-builder-system-design.md`
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
- dsl
- video
- product

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design-doc/ - Architecture and implementation guides
- reference/ - Investigation diary and quick references
- playbooks/ - Future setup-building flows and smoke guides
- scripts/ - Ticket-local helpers
- various/ - Notes and scratch artifacts
- archive/ - Deprecated or superseded material
