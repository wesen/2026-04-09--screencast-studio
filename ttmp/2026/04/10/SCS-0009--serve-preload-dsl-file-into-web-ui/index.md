---
Title: Serve preload DSL file into web UI
Ticket: SCS-0009
Status: active
Topics:
    - screencast-studio
    - backend
    - frontend
DocType: index
Intent: short-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Ticket workspace for adding a serve-time DSL preload flag and bootstrapping the web UI from that file."
LastUpdated: 2026-04-10T13:05:00-04:00
WhatFor: "Track the work needed to launch screencast-studio serve with a preloaded setup file."
WhenToUse: "Use when working on serve bootstrap behavior, startup config loading, or initial UI hydration."
---

# Serve preload DSL file into web UI

## Overview

This ticket adds support for launching `screencast-studio serve` with a setup file already loaded into the web application. The goal is that the first UI state comes from the file passed on the command line rather than from the hard-coded demo DSL embedded in the frontend.

The implementation needs coordinated backend and frontend work. The CLI must accept a file path and validate it at startup, the web API must expose the preloaded DSL as bootstrap data, and the frontend must apply that bootstrap data before its first normalize/compile cycle so the builder and preview logic start from the file content instead of the fallback default.

## Key Links

- Design doc: [`design-doc/01-serve-dsl-preload-implementation-guide.md`](./design-doc/01-serve-dsl-preload-implementation-guide.md)
- Diary: [`reference/01-diary.md`](./reference/01-diary.md)
- Tasks: [`tasks.md`](./tasks.md)
- Changelog: [`changelog.md`](./changelog.md)
