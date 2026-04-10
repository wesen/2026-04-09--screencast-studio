---
Title: GStreamer migration architecture and intern guide
Ticket: SCS-0011
Status: active
Topics:
    - screencast-studio
    - backend
    - frontend
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Ticket workspace for the GStreamer migration research, including a detailed intern-facing architecture and implementation guide."
LastUpdated: 2026-04-10T13:07:13.293419975-04:00
WhatFor: "Track research and design work for replacing the current FFmpeg subprocess-based media runtime with a GStreamer-based runtime."
WhenToUse: "Use when planning, implementing, or reviewing the screencast-studio media backend migration to GStreamer."
---

# GStreamer migration architecture and intern guide

## Overview

This ticket captures the architecture analysis and migration plan for replacing the current FFmpeg-heavy recording and preview runtime with a GStreamer-based runtime. The current system works, but it does so by building many independent FFmpeg subprocesses, supervising them manually, and coordinating them through cancellation, stdin signaling, and HTTP/websocket state propagation. That is functional, but it is brittle in exactly the places a live capture application tends to hurt: device contention, preview and recording duplication, stop sequencing, and error propagation.

The goal of this ticket is not to implement GStreamer immediately. The goal is to produce a detailed migration guide that a new intern can use to understand the current system, see where the complexity comes from, and execute a careful migration without breaking the browser-facing API or the existing DSL/configuration model.

## Key Links

- Design doc: [01-gstreamer-migration-analysis-and-implementation-guide.md](./design-doc/01-gstreamer-migration-analysis-and-implementation-guide.md)
- Diary: [01-investigation-diary.md](./reference/01-investigation-diary.md)
- Tasks: [tasks.md](./tasks.md)
- Changelog: [changelog.md](./changelog.md)

## Status

Current status: **active**

## Topics

- screencast-studio
- backend
- frontend

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
