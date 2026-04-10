---
Title: Preview handoff during recording implementation guide
Ticket: SCS-0008
Status: active
Topics:
    - screencast-studio
    - backend
DocType: design-doc
Intent: short-term
Owners: []
RelatedFiles:
    - Path: internal/web/handlers_api.go
      Note: Recording start endpoint currently starts recording without preview coordination
    - Path: internal/web/preview_manager.go
      Note: Preview lifecycle owner that will suspend and restore preview workers
    - Path: internal/web/server.go
      Note: Server owns both managers and is the right place to coordinate handoff state
    - Path: internal/web/session_manager.go
      Note: Recording session completion hook needs to trigger preview restore
    - Path: pkg/recording/ffmpeg.go
      Note: Preview and recording both build camera inputs from the same v4l2 source shape
    - Path: ui/src/pages/StudioPage.tsx
      Note: UI keeps desired previews active independently of recording state
ExternalSources: []
Summary: "Small implementation guide for stopping active previews before recording and restoring them after recording completion."
LastUpdated: 2026-04-10T12:45:00-04:00
WhatFor: "Guide a small backend fix for the camera preview contention bug."
WhenToUse: "Use when implementing or reviewing preview/recording coordination."
---

# Preview handoff during recording implementation guide

## Problem

The current runtime launches previews and recordings as separate `ffmpeg` processes. For camera sources, both paths open the same v4l2 device. If a preview is already alive for `/dev/video0` and recording starts, the recording `ffmpeg` process can fail immediately with `Device or resource busy`.

This is not a frontend rendering bug. It is a runtime ownership bug. The product currently treats preview and recording as independent consumers of the same camera handle.

## Current code path

- `ui/src/pages/StudioPage.tsx`
  - Keeps previews alive for armed sources while the Studio tab is active.
  - Starts recording without first releasing previews.
- `internal/web/handlers_api.go`
  - Accepts `/api/recordings/start` and directly delegates to `RecordingManager.Start(...)`.
- `internal/web/preview_manager.go`
  - Owns active preview workers and their cancellation functions.
- `internal/web/session_manager.go`
  - Owns the current recording session and knows when the session really finishes.
- `pkg/recording/ffmpeg.go`
  - Builds the same `-f v4l2 ... -i /dev/videoX` input shape for both preview and recording.

## Implementation approach

Use a server-owned handoff, not a UI workaround.

1. Before recording starts, ask `PreviewManager` to suspend all active previews and wait for them to exit.
2. Record the set of preview source IDs that were active at suspend time, plus the DSL used to start the recording.
3. Start the recording session.
4. If recording start fails, immediately restore the suspended previews.
5. When the recording session really finishes, restore the suspended previews from the stored DSL and source IDs.

This keeps the behavior deterministic:

- preview owns the device before recording,
- then recording owns the device,
- then preview regains ownership after recording completion.

## Why this is the right size fix

This bug does not require a full shared-capture pipeline, a v4l2 loopback fanout, or frontend-only state juggling. The smallest coherent fix is to make the backend explicitly hand device ownership from preview to recording and back again.

That approach also handles non-UI recording termination paths. A user can stop recording, a bounded recording can finish naturally, or a recording can fail; the restore logic still runs from the recording lifecycle rather than assuming the UI will clean up in the right order.

## Main changes

- `PreviewManager.SuspendAll(ctx, reason)`
  - cancel current preview workers
  - wait for them to exit
  - return the source IDs that should be restored later
- `PreviewManager.RestoreSuspended(ctx, dslBody, plan)`
  - re-run preview ensure for the suspended source IDs
- `Server`
  - store one preview handoff for the active recording session
  - restore previews on recording finish
- `RecordingManager`
  - invoke a post-finish callback once the session state is finalized

## Review checklist

- Confirm preview suspension waits for preview workers to exit before recording start proceeds.
- Confirm failed recording starts restore the previews that were just suspended.
- Confirm recording completion restores previews only after the session is fully finished.
- Confirm runtime shutdown does not restore previews if the server context is already canceled.
- Confirm tests cover suspend/restore behavior and the full start-stop handoff path.
