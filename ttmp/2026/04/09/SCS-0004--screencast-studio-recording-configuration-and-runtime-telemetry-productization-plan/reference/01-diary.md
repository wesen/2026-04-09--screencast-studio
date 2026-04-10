---
Title: Diary
Ticket: SCS-0004
Status: active
Topics:
    - frontend
    - backend
    - ui
    - audio
    - video
    - product
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ui/src/components/studio/OutputPanel.tsx
      Note: Current placeholder output configuration UI that motivated this ticket
    - Path: ui/src/components/studio/MicPanel.tsx
      Note: Current microphone UI that motivated the telemetry part of the ticket
    - Path: ui/src/components/studio/StatusPanel.tsx
      Note: Current status UI that still exposes placeholder disk telemetry
    - Path: ui/src/pages/StudioPage.tsx
      Note: Current mounted page and orchestration owner for the future implementation
ExternalSources: []
Summary: Chronological record of creating the recording-configuration and telemetry productization ticket and defining its implementation direction.
LastUpdated: 2026-04-09T21:02:00-04:00
WhatFor: Preserve why this follow-on ticket exists and what the intended scope is.
WhenToUse: Read when beginning implementation, reviewing scope, or checking why the recording controls were split into their own productization ticket.
---

# Diary

## Goal

Document why the recording configuration and telemetry ticket was created, what gaps it is responsible for, and how the initial implementation plan was assembled.

## Step 1: Create The Productization Ticket

This ticket was created immediately after reviewing what is still missing from the web UI once the cleanup work was done. The frontend is now structurally clean, but the recording-control portion is still not a finished product. The output controls, microphone panel, and status panel are all visibly present, yet several of them still expose placeholder or partially wired behavior.

The user specifically called out:

- live audio metering
- destination directory display
- destination names
- a name field in the UI

That request clarified that the next meaningful step is not another cleanup ticket. It is a productization ticket for the recording configuration surface and runtime telemetry.

### Prompt Context

**User prompt (verbatim):** "Make a new ticket and add a detailed implementation plan.

In fact, I also want to properly show the destination directory and the destination names, and have a name field in the UI.

Actually before you start:
What other things in the UI are not functional / should be added to make this a finished product."

**Assistant interpretation:** First identify the major remaining product gaps in the UI, then open a dedicated ticket for the recording-configuration and telemetry slice, and write a detailed intern-facing design and implementation guide plus tasks.

### What I did

- Reviewed the current mounted UI and backend web surface.
- Identified the recording configuration and telemetry slice as a distinct follow-on product ticket.
- Created `SCS-0004`.
- Added the main design guide and this diary.
- Broke the work down into explicit phases covering product model, protobuf, backend preview logic, telemetry workers, frontend wiring, and validation.

### Why

- The UI cleanup ticket intentionally removed fake telemetry. That now makes the missing product work very obvious and easy to scope correctly.
- Recording name, destination preview, microphone metering, and disk telemetry all belong to one coherent product surface.
- This is the right level of scope for an intern: broad enough to be meaningful, but still organized around one user-facing area of the application.

### What worked

- The earlier cleanup work makes this ticket much easier to define than it would have been before the transport and shell were stabilized.
- The current protobuf contract and websocket path already give this ticket a clean extension point.

### What warrants a second pair of eyes

- Whether recording name and destination should be modeled as DSL fields, runtime-only overlay fields, or both.
- Whether output preview should be delivered from normalize, compile, or a dedicated endpoint.

### Code review instructions

- Start with the main design guide:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0004--screencast-studio-recording-configuration-and-runtime-telemetry-productization-plan/design-doc/01-recording-configuration-and-runtime-telemetry-system-design.md`
- Then compare the problem statement against:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/OutputPanel.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/MicPanel.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/StatusPanel.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`

### Technical details

Commands used in this step:

```bash
docmgr ticket create-ticket --ticket SCS-0004 --title "Screencast Studio Recording Configuration and Runtime Telemetry Productization Plan" --topics frontend,backend,ui,audio,video,product
docmgr doc add --ticket SCS-0004 --doc-type design-doc --title "Recording Configuration and Runtime Telemetry System Design"
docmgr doc add --ticket SCS-0004 --doc-type reference --title "Diary"
```

## Step 2: Audit The Current Controls And Freeze The Product Model

The first implementation step was not code. It was freezing the actual product gap and choosing a simple ownership model before changing the frontend. That matters here because the codebase already has three nearby concepts: raw DSL, a structured setup draft, and a fake `studioDraft` UI slice. Without deciding which one should win, the implementation would drift into another temporary state layer.

The chosen direction is to keep the product model simple. Recording name, destination naming, microphone choice, and gain should flow through the structured DSL draft when the builder is active. The backend compile path should remain the authority for output preview. Protobuf expansion should be reserved for runtime telemetry and any genuinely runtime-only state, not used as an excuse to create a second recording-config overlay prematurely.

### Prompt Context

**User prompt (verbatim):** "go ahead."

**Assistant interpretation:** Start implementing `SCS-0004` task by task, beginning with the current-state audit and product-model decisions, and keep the ticket diary up to date while working.

**Inferred user intent:** Move from planning into real implementation, but do it in disciplined slices so the resulting code and ticket history stay understandable.

### What I did

- Audited the current frontend control surfaces:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/OutputPanel.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/MicPanel.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/StatusPanel.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`
- Audited the existing backend and protobuf surfaces:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/proto/screencast/studio/v1/web.proto`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/compile.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/types.go`
- Updated the ticket tasks with a precise current-vs-target summary.
- Updated the design doc with the current reality audit and the ownership decision.

### Why

- The backend already resolves real output paths from DSL. That is a strong existing primitive and should be reused.
- The frontend already has a structured setup draft. Reusing it is much simpler than adding yet another config overlay layer.
- Meter and disk telemetry really are runtime-only and belong on the protobuf websocket path, so that remains the correct place to extend the shared contract later.

### What worked

- The current codebase is already organized enough that the product-model decision can be grounded in real file ownership instead of speculation.
- The compile path already gives the exact output preview needed for the name and destination UI.

### What didn't work

- N/A

### What I learned

- `studioDraft` is currently the last major fake state pocket in the mounted app.
- The easiest path forward is not “invent a richer recording API”; it is “make the mounted UI edit the same DSL-backed state the runtime already understands”.

### What was tricky to build

- The main tricky part here was not syntax. It was avoiding a bad architecture fork. There was a tempting path where recording name, destination, mic input, and gain would be sent as a second overlay object while the builder continued generating DSL independently. That would have increased drift risk immediately because compile preview, preview ensuring, and recording start would all need to reconcile DSL plus overlay state.

### What warrants a second pair of eyes

- The decision to use `setupDraft.audioSources[0]` as the v1 microphone control surface should be reviewed once source-management work for audio grows more sophisticated.
- The exact structured-builder policy for rewriting destination templates should be checked against any advanced DSL examples before locking the UX completely.

### What should be done in the future

- Implement the first code slice by replacing fake `studioDraft` ownership with real structured-draft-backed recording configuration.
- Extend protobuf only for telemetry events and any data that truly does not belong in DSL.

### Code review instructions

- Start with the task and design-doc updates:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0004--screencast-studio-recording-configuration-and-runtime-telemetry-productization-plan/tasks.md`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0004--screencast-studio-recording-configuration-and-runtime-telemetry-productization-plan/design-doc/01-recording-configuration-and-runtime-telemetry-system-design.md`
- Then compare the chosen model to:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/setupDraftSlice.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/studio-draft/studioDraftSlice.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/compile.go`

### Technical details

Commands used in this step:

```bash
sed -n '1,260p' proto/screencast/studio/v1/web.proto
sed -n '1,260p' internal/web/handlers_api.go
sed -n '1,260p' internal/web/pb_mapping.go
sed -n '1,260p' pkg/dsl/compile.go
sed -n '1,260p' pkg/dsl/types.go
sed -n '1,260p' ui/src/components/studio/OutputPanel.tsx
sed -n '1,260p' ui/src/components/studio/MicPanel.tsx
sed -n '1,240p' ui/src/components/studio/StatusPanel.tsx
sed -n '1,260p' ui/src/pages/StudioPage.tsx
```
