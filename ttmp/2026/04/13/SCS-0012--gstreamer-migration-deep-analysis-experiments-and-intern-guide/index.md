---
Title: 'GStreamer Migration: Deep Analysis, Experiments, and Intern Guide'
Ticket: SCS-0012
Status: active
Topics:
    - screencast-studio
    - backend
    - gstreamer
    - audio
    - video
    - transcription
    - screenshots
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/preview_manager.go
      Note: Preview lifecycle manager to be preserved
    - Path: internal/web/preview_runner.go
      Note: Current FFmpeg preview runner to be replaced with GStreamer
    - Path: pkg/recording/ffmpeg.go
      Note: Current FFmpeg argument builders that will be replaced
    - Path: pkg/recording/run.go
      Note: Current recording subprocess supervisor to be replaced
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-13T14:15:54.675319496-04:00
WhatFor: ""
WhenToUse: ""
---





# GStreamer Migration: Deep Analysis, Experiments, and Intern Guide

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **System explanation + full postmortem**: `design-doc/03-screencast-studio-system-explanation-and-gstreamer-migration-postmortem-for-interns.md`
- **Broad migration analysis**: `design-doc/01-gstreamer-migration-analysis-and-intern-guide.md`
- **Phase 4 shared-capture guide**: `design-doc/02-phase-4-shared-capture-architecture-and-intern-implementation-guide.md`
- **Implementation diary**: `reference/01-diary.md`
- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- screencast-studio
- backend
- gstreamer
- audio
- video
- transcription
- screenshots

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
