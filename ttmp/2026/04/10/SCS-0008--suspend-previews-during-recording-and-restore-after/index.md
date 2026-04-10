---
Title: Suspend previews during recording and restore after
Ticket: SCS-0008
Status: active
Topics:
    - screencast-studio
    - backend
DocType: index
Intent: short-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Ticket workspace for suspending live previews before recording starts and restoring them after the recording session finishes."
LastUpdated: 2026-04-10T12:45:00-04:00
WhatFor: "Track the preview handoff bug where preview ffmpeg workers contend with recording ffmpeg workers for camera devices."
WhenToUse: "Use when working on preview lifecycle, recording start/stop coordination, or related regression tests."
---

# Suspend previews during recording and restore after

## Overview

This ticket addresses the current runtime bug where preview workers can continue holding camera devices while a recording session starts. On Linux v4l2 devices that causes recording startup to fail with `Device or resource busy` because preview and recording are separate `ffmpeg` processes that both try to open the same `/dev/video*` node.

The implementation approach in this ticket is intentionally small and explicit: suspend all active previews before starting a recording, remember which preview sources were active, and restore that set once the recording session fully finishes. This keeps preview/device ownership clear without introducing a new video fanout layer.

## Key Links

- Design doc: [`design-doc/01-preview-handoff-during-recording-implementation-guide.md`](./design-doc/01-preview-handoff-during-recording-implementation-guide.md)
- Diary: [`reference/01-diary.md`](./reference/01-diary.md)
- Tasks: [`tasks.md`](./tasks.md)
- Changelog: [`changelog.md`](./changelog.md)
