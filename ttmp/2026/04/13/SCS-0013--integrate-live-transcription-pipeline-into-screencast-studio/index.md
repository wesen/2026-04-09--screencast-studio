---
Title: Integrate live transcription pipeline into screencast-studio
Ticket: SCS-0013
Status: active
Topics:
    - screencast-studio
    - transcription
    - gstreamer
    - audio
    - websocket
    - go
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/recording.go
      Note: Existing recording audio graph where transcription branching should happen
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go
      Note: Existing websocket transport for browser-facing live events
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/proto/screencast/studio/v1/web.proto
      Note: Existing server event schema that will need transcription updates
    - Path: /home/manuel/code/wesen/2026-04-13--transcription-go/server/server.py
      Note: Prototype ASR service exposing batch, chunk, and websocket streaming endpoints
    - Path: /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/runner.go
      Note: Prototype live orchestration and accumulator model that informed this ticket
ExternalSources: []
Summary: This ticket captures how to integrate the separate transcription-go prototype into screencast-studio. The primary deliverable is a detailed intern-facing guide recommending a protocol-level transcription seam, a GStreamer appsink branch for normalized PCM, server-mediated websocket updates, and a phased rollout toward live transcript UI and transcript sidecar outputs.
LastUpdated: 2026-04-13T20:50:00-04:00
WhatFor: Plan and document a safe, reviewable integration of live transcription into screencast-studio using the ideas and protocols proven in transcription-go.
WhenToUse: Start here when you need the overall ticket context, key file map, or the main design guide for screencast-studio transcription integration.
---

# Integrate live transcription pipeline into screencast-studio

## Overview

This ticket studies the separate `../2026-04-13--transcription-go` repository and translates its useful architecture into a concrete integration plan for `screencast-studio`. The focus is not merely “run ASR somehow”; it is how to connect the current GStreamer recording graph, the current server websocket/event model, and the browser UI to a session-oriented transcription backend in a way that remains understandable for a new engineer.

## Key Links

- **Primary design guide**: `design-doc/01-screencast-studio-live-transcription-integration-architecture-and-intern-implementation-guide.md`
- **Investigation diary**: `reference/01-investigation-diary.md`
- **Tasks**: `tasks.md`
- **Changelog**: `changelog.md`

## Status

Current status: **active**

Current deliverable status:
- Ticket created
- Evidence gathered from both repositories
- Primary design/implementation guide written
- Diary started
- `docmgr doctor` passed cleanly
- Bundle uploaded to reMarkable and verified in `/ai/2026/04/13/SCS-0013`

## Topics

- screencast-studio
- transcription
- gstreamer
- audio
- websocket
- go

## Tasks

See [tasks.md](./tasks.md) for the full research and implementation checklist.

## Changelog

See [changelog.md](./changelog.md) for the record of what was added and why.

## Structure

- `design-doc/` - architecture and implementation guidance
- `reference/` - investigation diary and quick-reference context
- `scripts/` - future ticket-local validation harnesses
- `sources/` - imported or generated source material if needed later
- `archive/` - deprecated or historical artifacts if the ticket grows
