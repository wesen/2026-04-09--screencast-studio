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
      Note: Preview lifecycle manager after suspend/restore removal
    - Path: pkg/media/gst/shared_video.go
      Note: Shared capture registry and tee-backed preview source implementation
    - Path: pkg/media/gst/shared_video_recording_bridge.go
      Note: Shared-source bridge recorder used to keep preview alive during recording
    - Path: pkg/media/gst/recording.go
      Note: Current native recording runtime
ExternalSources: []
Summary: "Phases 0-4 are complete. The active runtime architecture is now GStreamer-only with shared capture and no preview suspend/restore path. Phase 5 is the planned live transcription feature set."
LastUpdated: 2026-04-13T19:20:00-04:00
WhatFor: "Track the completed migration from legacy multi-runtime behavior to the current shared-capture GStreamer runtime, and capture the remaining live-transcription work."
WhenToUse: "Start here when you need the current ticket status, the key implementation docs, or the next planned phase after the completed Phase 4 cleanup."
---





# GStreamer Migration: Deep Analysis, Experiments, and Intern Guide

## Overview

This ticket now documents a completed Phase 4 migration: the live repo/runtime is GStreamer-only, preview and recording share capture sources without suspend/restore, and the old FFmpeg execution path has been removed. The next planned implementation phase is live audio transcription.

## Key Links

- **System explanation + full postmortem**: `design-doc/03-screencast-studio-system-explanation-and-gstreamer-migration-postmortem-for-interns.md`
- **Broad migration analysis**: `design-doc/01-gstreamer-migration-analysis-and-intern-guide.md`
- **Phase 4 shared-capture guide**: `design-doc/02-phase-4-shared-capture-architecture-and-intern-implementation-guide.md`
- **Implementation diary**: `reference/01-diary.md`
- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

Phase status summary:
- **Phases 0-4:** complete
- **Phase 5:** planned next
- **Active runtime model:** GStreamer shared capture with no preview suspend/restore

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
