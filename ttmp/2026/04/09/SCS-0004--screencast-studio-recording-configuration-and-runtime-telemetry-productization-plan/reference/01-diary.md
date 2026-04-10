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
