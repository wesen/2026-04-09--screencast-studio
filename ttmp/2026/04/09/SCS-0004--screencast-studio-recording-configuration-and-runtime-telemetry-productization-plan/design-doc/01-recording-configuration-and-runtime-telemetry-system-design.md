---
Title: Recording Configuration and Runtime Telemetry System Design
Ticket: SCS-0004
Status: active
Topics:
    - frontend
    - backend
    - ui
    - audio
    - video
    - product
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/handlers_api.go
      Note: |-
        Current REST entrypoint for setup and recording APIs that may expand for richer configuration
        Current recording and setup API surface that may expand for productized config
    - Path: internal/web/handlers_ws.go
      Note: |-
        Current websocket stream that should carry runtime telemetry events
        Websocket event stream that should carry telemetry
    - Path: pkg/dsl/types.go
      Note: DSL and normalized config types that determine where name and destination data should live
    - Path: proto/screencast/studio/v1/web.proto
      Note: |-
        Shared transport schema that should remain the source of truth for added fields and events
        Shared contract that should be extended for config and telemetry messages
    - Path: ui/src/components/studio/MicPanel.tsx
      Note: |-
        Existing microphone panel that needs real input enumeration and live telemetry
        Microphone panel that needs real devices and live meter support
    - Path: ui/src/components/studio/OutputPanel.tsx
      Note: |-
        Existing output controls that need to become real product controls instead of placeholders
        Recording controls and destination UI that need productization
    - Path: ui/src/components/studio/StatusPanel.tsx
      Note: |-
        Existing status panel that still renders placeholder disk telemetry
        Status panel that needs real disk telemetry
    - Path: ui/src/pages/StudioPage.tsx
      Note: |-
        Current app-level owner for recording, normalization, and preview orchestration
        Mounted page that will own configuration and telemetry wiring
ExternalSources: []
Summary: Detailed design and implementation guide for turning recording configuration, naming, destination previews, and runtime telemetry into real backend-driven product behavior.
LastUpdated: 2026-04-09T21:00:00-04:00
WhatFor: Guide an intern through productizing the recording controls and telemetry surfaces without reintroducing fake UI state.
WhenToUse: Read before implementing recording name fields, destination previews, audio metering, or disk/runtime telemetry.
---


# Recording Configuration and Runtime Telemetry System Design

## Executive Summary

The current web UI can start and stop a recording, but the recording-control surface is still not a finished product. Several controls remain presentation-only. The output panel exposes format, frame rate, quality, and destination widgets, yet the actual recording start mutation still sends only the raw DSL. The microphone panel renders controls, yet it does not use discovered audio devices or real meter telemetry. The status panel renders a disk area, but it does not consume real disk telemetry. A user also cannot clearly see or edit the recording name, and cannot see the final resolved output paths before hitting Record.

This ticket fixes that. The goal is to make the recording configuration surface honest and operational. A user should be able to set a recording name, choose or view the destination directory, inspect the resolved output filenames the system will generate, choose a discovered audio input, observe live metering, and understand disk/runtime status before and during recording.

This is not a pure frontend ticket. The backend must expose or compute the data the UI needs. The correct implementation is therefore an end-to-end one: product model, shared protobuf contract, backend mapping and telemetry collection, then frontend rendering and interaction.

## Problem Statement

### Current Behavior

The current UI has good structure but incomplete product semantics.

- The mounted UI can normalize DSL and start/stop recording.
- The preview lifecycle works.
- Websocket session and log streaming work.
- The output controls are not the source of truth for the actual recording request.
- The microphone panel does not use real discovered devices or meter data.
- The status panel does not use real disk telemetry.
- There is no user-facing session name field.
- There is no reliable preview of resolved output file paths.

### Why This Matters

For a recording tool, the questions a user asks before pressing Record are straightforward:

- What will this recording be called?
- Where will it go?
- What files will be created?
- Which microphone am I using?
- Is the mic actually receiving signal?
- Do I have enough disk space?

If the UI cannot answer those questions clearly, it is still a prototype, not a finished product.

### Risks If Left Unfixed

- Users will not trust the output controls because they do not actually affect recording.
- Users will not know which files are going to be written until after recording.
- Users will not know whether the chosen audio input is live.
- The system will continue carrying UI-only recording semantics that do not correspond to backend behavior.

## Proposed Solution

### High-Level Approach

Extend the existing recording configuration model so the backend can compute and expose user-facing recording details before recording begins, and add real runtime telemetry over the websocket channel.

The solution has five parts:

1. Add a product-level recording configuration model.
2. Extend the protobuf contract.
3. Make the backend compute and validate resolved output destinations.
4. Add backend telemetry publishers for audio levels and disk state.
5. Render the new configuration and telemetry in the mounted UI.

### Target User Flow

```text
Open Studio
  -> See discovered/default setup
  -> Edit recording name
  -> Inspect destination directory
  -> See resolved output file list
  -> Pick microphone input
  -> Confirm live meter activity
  -> Confirm disk status
  -> Record
  -> See recording runtime status update live
```

### Proposed Product Model

The product model should distinguish between three categories of data.

#### Category 1: Editable Configuration

- recording name
- destination root directory
- selected audio input
- gain
- output container / quality / frame rate if these remain exposed in UI

#### Category 2: Backend-Derived Preview Data

- resolved output filenames
- resolved full paths
- low-disk warnings
- validation errors on destination fields

#### Category 3: Live Runtime Telemetry

- audio meter level
- disk telemetry
- maybe future write-rate or dropped-frame information

This separation is important. The UI should not derive final file paths on its own. It should ask the backend for them.

## Detailed Design

## 1. Recording Name

The UI needs a real “name” field. This should not just be decorative. It should affect output filename resolution.

### Proposed Semantics

- The field is user-editable in the main recording controls area.
- It becomes part of the data used to render destination templates.
- It is sent to the backend as part of the recording-config input surface.

### Example

If the user enters `Interview-01`, the per-source outputs might become:

- `recordings/Interview-01/Full Desktop.mov`
- `recordings/Interview-01/Built-in Mic.wav`

or if a flatter naming policy is chosen:

- `recordings/Interview-01-desktop.mov`
- `recordings/Interview-01-audio-mix.wav`

The key point is that the backend, not the UI, resolves the exact final paths.

## 2. Destination Directory

The existing `Save to` selector is not real. The productized version should expose a real destination-root concept.

### Recommended Approach

Use one editable destination-root field in the UI. This can begin as a plain text field if no cross-platform directory-picker is available in the browser context.

The backend should validate:

- empty path
- non-writable path if detectable
- relative path policy
- path traversal or unsafe normalization if relevant

### Recommended UI Behavior

- show current destination root
- show validation state
- show resolved outputs below it

Pseudocode:

```ts
const preview = await previewRecordingConfig({
  dsl,
  recordingName,
  destinationRoot,
  audioInputId,
  gain,
});

render(preview.outputs);
```

## 3. Resolved Output Preview

Users need to know what files will be written before recording starts.

### Recommended Backend Behavior

The backend should expose a resolved preview of:

- output kind
- source name
- source id
- final filename
- full path

This can likely fit best into:

- normalize response, if normalize is extended to include recording-config inputs
- or compile response, if compile is treated as the authoritative preview
- or a dedicated `preview recording config` endpoint if separation is cleaner

### Suggested JSON/Protobuf Shape

```proto
message ResolvedOutput {
  string kind = 1;
  string source_id = 2;
  string source_name = 3;
  string file_name = 4;
  string full_path = 5;
}
```

### UI Rendering

Render a small table or list:

```text
Outputs
  Desktop        recordings/Interview-01/Full Desktop.mov
  Audio Mix      recordings/Interview-01/audio-mix.wav
```

This should be visible without forcing the user into Raw DSL mode.

## 4. Audio Input Discovery And Metering

The microphone panel currently renders fixed choices and fake or unavailable meter behavior. That is not enough for a finished product.

### Discovery

Use the backend discovery snapshot as the source of truth for audio inputs.

The UI should render:

- discovered device name
- maybe driver or state if useful
- current selected device

### Metering

The backend should produce short rolling meter samples for the active input. The frontend should consume that via websocket.

Do not sample too fast. A reasonable cadence for v1 is something like 5 to 10 updates per second.

### Meter Event Shape

```proto
message AudioMeterEvent {
  string input_id = 1;
  float left = 2;
  float right = 3;
  bool clipped = 4;
}
```

### Ownership

The backend owns the meter.
The websocket owns delivery.
The frontend owns only the latest displayed meter state.

## 5. Disk Telemetry

The status panel should stop showing placeholder disk UI.

### Recommended Data Model

```proto
message DiskTelemetryEvent {
  string root_path = 1;
  uint64 bytes_total = 2;
  uint64 bytes_available = 3;
  uint64 bytes_estimated_required = 4;
  bool low_space = 5;
}
```

### Notes

- The backend may not know the final total required bytes precisely.
- An estimate is acceptable if clearly identified.
- If the chosen destination is invalid or unavailable, the backend should send a degraded status instead of nothing.

## 6. Runtime Coordination

Any backend telemetry worker must be owned clearly and cancellable cleanly.

### Rule

Every goroutine started for metering or disk polling must:

- accept a context
- be started from one clear owner
- stop when the session or websocket scope ends

This is especially important because the codebase already cares about goroutine ownership and cancellation discipline.

Diagram:

```text
recording session start
  |
  +--> audio meter worker (ctx-bound)
  |
  +--> disk telemetry worker (ctx-bound)
  |
  +--> websocket event publish path
  |
recording stop / disconnect
  |
  +--> cancel ctx
        |
        +--> all telemetry workers exit
```

## API References

Current relevant files:

- REST:
  - `internal/web/handlers_api.go`
- websocket:
  - `internal/web/handlers_ws.go`
- shared schema:
  - `proto/screencast/studio/v1/web.proto`
- page owner:
  - `ui/src/pages/StudioPage.tsx`
- recording controls:
  - `ui/src/components/studio/OutputPanel.tsx`
- microphone UI:
  - `ui/src/components/studio/MicPanel.tsx`
- status UI:
  - `ui/src/components/studio/StatusPanel.tsx`

## Design Decisions

### Decision 1: Keep The Backend As The Source Of Truth For Resolved Output Paths

Rationale:

- avoids duplicate template logic in the browser
- ensures the preview reflects what the runtime will actually do
- keeps name and destination validation centralized

### Decision 2: Use Websocket Events For Live Telemetry

Rationale:

- meter and disk updates are naturally streaming state
- avoids polling loops in the browser
- fits the existing event architecture

### Decision 3: Keep The UI Editable Model Separate From Backend-Derived Preview Data

Rationale:

- prevents the UI from guessing final paths
- makes loading and error states clearer
- keeps the app architecture easier to explain to an intern

### Decision 4: Do Not Reintroduce Fake Runtime State

Rationale:

- the cleanup ticket removed fake meter and disk state on purpose
- this ticket should replace those with real data, not a better simulation

## Alternatives Considered

### Alternative 1: Compute Resolved Output Paths Entirely In The Frontend

Rejected because:

- it duplicates DSL/template behavior
- it risks diverging from backend runtime behavior
- it becomes hard to validate safely

### Alternative 2: Keep Audio Metering Out Of Scope

Rejected because:

- the current UI already visually implies microphone monitoring
- a recording product without a signal indicator is incomplete

### Alternative 3: Add Destination Directory Only And Defer Telemetry

Rejected because:

- it would still leave the status and microphone panels half-finished
- the user explicitly wants the telemetry and output naming behavior too

## Implementation Plan

### Step 1: Freeze The Product Model

- write down the exact editable config vs preview vs telemetry data split
- decide where recording name and destination live structurally

### Step 2: Extend Protobuf

- add new recording configuration preview messages
- add telemetry event messages
- regenerate code

### Step 3: Backend Destination Preview

- add or extend backend APIs to return resolved output previews
- validate destination inputs

### Step 4: Backend Telemetry Workers

- implement audio metering
- implement disk telemetry
- publish websocket events

### Step 5: Frontend UI Wiring

- add recording name field
- add destination-root field
- render resolved outputs
- render live meter and disk telemetry

### Step 6: Tests And Smoke Validation

- Go tests for validation and event mapping
- frontend tests and stories
- live smoke against the real server

## Open Questions

- Should destination path selection remain a text field for v1, or do we want a browser-assisted picker if available?
- Should output preview be part of normalize, compile, or a dedicated endpoint?
- Should disk telemetry be scoped to the destination root only, or to the actual resolved output directory per recording?
- Should meter telemetry be available only when the studio is open, or throughout the recording lifecycle?

## Intern Notes

If you are the intern implementing this, keep these rules in mind:

- Do not add fake telemetry to make the UI “look alive.”
- Do not compute final output names in the frontend if the backend can do it.
- Keep every new telemetry goroutine cancellable and clearly owned.
- Update the protobuf schema first if shared payloads change.
- Re-test the mounted app against the real server, not only Storybook.
