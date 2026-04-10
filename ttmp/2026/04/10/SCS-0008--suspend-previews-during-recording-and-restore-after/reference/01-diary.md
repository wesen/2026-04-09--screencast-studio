---
Title: Diary
Ticket: SCS-0008
Status: active
Topics:
    - screencast-studio
    - backend
DocType: reference
Intent: short-term
Owners: []
RelatedFiles:
    - Path: internal/web/handlers_api.go
      Note: Recording start path that needs preview coordination
    - Path: internal/web/preview_manager.go
      Note: Preview suspend and restore logic lives here
    - Path: internal/web/server.go
      Note: Server-level preview handoff state and restore callback live here
    - Path: internal/web/session_manager.go
      Note: Recording finish callback must trigger preview restore
ExternalSources: []
Summary: Diary for the preview handoff implementation work.
LastUpdated: 2026-04-10T12:45:00-04:00
WhatFor: Capture what changed, why, what worked, and how to review the preview handoff fix.
WhenToUse: Use when reviewing or resuming ticket SCS-0008.
---

# Diary

## Goal

Capture the work needed to suspend live previews before recording starts and restore them after the recording session finishes so camera devices are not opened by competing `ffmpeg` workers.

## Step 1: Create the ticket and implementation plan

I started by turning the bug report into a focused ticket because the fix needs code, tests, and a small amount of design guidance. The main implementation decision was to keep the fix in the backend runtime layer rather than relying on the UI to release and later recreate previews in exactly the right order.

That decision came directly from the current architecture. The server owns both the preview manager and the recording manager, and the recording manager is the only place that knows when a recording session is truly finished. That makes the server/manager layer the correct place to coordinate preview handoff.

### Prompt Context

**User prompt (verbatim):** "create a new ticket to address the issue by closing the previews before recording, and restarting them after. 

Create a small implementation guide and then implement, commit, keep a diary."

**Assistant interpretation:** Create a new ticket workspace, write a concise implementation guide for the preview handoff behavior, implement the backend fix, commit the changes, and record the work in a diary.

**Inferred user intent:** Fix the camera contention bug in a disciplined way that leaves behind clear ticket-local documentation and an implementation record for later review.

**Commit (code):** N/A

### What I did
- Created ticket `SCS-0008` under `ttmp/2026/04/10/`.
- Wrote a small implementation guide describing the failure mode and the intended backend handoff.
- Chose a backend coordination approach centered on `Server`, `PreviewManager`, and `RecordingManager`.

### Why
- The bug spans preview lifecycle, recording lifecycle, and session completion.
- A ticket-local guide makes the implementation easier to review and resume.

### What worked
- The current runtime architecture is already split into managers that make the handoff fix straightforward.
- The failure report and code inspection were enough to justify a narrow, explicit fix.

### What didn't work
- N/A

### What I learned
- The backend already has the right ownership seams for a handoff fix.
- The UI should not be the sole owner of preview restart behavior for a recording lifecycle bug.

### What was tricky to build
- The trickiest design constraint was picking a place that can both suspend previews before start and restore them after actual finish. Starting and stopping happen in request handlers, but true recording completion happens inside the recording manager, so the fix has to bridge both phases.

### What warrants a second pair of eyes
- The exact restore trigger after recording completion.
- The behavior during runtime shutdown, where preview restore should not resurrect workers after the server is already canceling.

### What should be done in the future
- Update this diary after the implementation and tests are committed.

### Code review instructions
- Start with the design doc.
- Then review the server/manager coordination code once the implementation step is added here.

### Technical details
- Ticket workspace: `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/10/SCS-0008--suspend-previews-during-recording-and-restore-after`
