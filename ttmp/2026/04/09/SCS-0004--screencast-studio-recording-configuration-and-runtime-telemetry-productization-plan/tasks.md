# Tasks

## Goal

Make the recording configuration and telemetry parts of the web UI real: destination directory, recording name, resolved output filenames, audio input selection, gain, live audio meter, and disk/runtime telemetry.

## Phase 1: Freeze The Current Product Gap

- [x] Audit the current controls in `ui/src/components/studio/OutputPanel.tsx`, `ui/src/components/studio/MicPanel.tsx`, and `ui/src/pages/StudioPage.tsx`.
- [x] Record which controls are presentation-only versus backend-driven.
- [x] Enumerate exactly which values must become first-class runtime configuration:
  - recording name
  - destination root directory
  - per-source filename preview
  - audio input choice
  - gain
  - meter levels
  - disk telemetry
- [x] Record the current backend capabilities and limits so the intern does not design imaginary behavior.

Current-vs-target summary:

- `OutputPanel` is mostly presentation-only today. `format`, `fps`, `quality`, `audio`, and `multiTrack` are driven by `studioDraft`, but `record` still sends only `dsl`, so those controls do not affect the runtime. `Save to` is fully fake.
- `MicPanel` renders a real control shell, but input choices are hardcoded and `gain` is only local Redux UI state. Live meter data is not connected.
- `StatusPanel` is mostly honest after cleanup, but disk telemetry is still absent and explicitly rendered as unavailable.
- The backend already knows how to normalize DSL, compile a real output plan, resolve concrete output paths, discover audio inputs, and stream session/log events over websocket.
- The backend does not yet publish audio-meter or disk-telemetry events.

Acceptance criteria:

- There is a crisp “current vs target” list for every control and telemetry surface.
- The intern can point to which fields are currently fake, missing, or only partially wired.

## Phase 2: Define The Product Model

- [x] Define the user-facing recording configuration model:
  - session or recording name
  - output root directory
  - resolved output files
  - audio input selection
  - gain
  - telemetry feed
- [x] Decide which values live in DSL, which live in a runtime request, and which are derived read-only display values.
- [x] Decide whether the UI should mutate raw DSL directly, maintain a structured draft model, or send an overlay config to the backend.
- [x] Document the canonical resolution flow for output filenames before recording begins.

Chosen product model:

- `setupDraft.sessionId` is the recording name in structured mode.
- `setupDraft.destinationTemplates` remains the canonical naming template source. The mounted UI will expose a simplified destination-root field only for the builder-supported template shape and will rewrite the canonical templates, not bypass them.
- `setupDraft.audioSources[0]` is the primary microphone control surface in v1. Device selection and gain update the structured draft and therefore the rendered DSL.
- The compile endpoint remains the authority for output preview. The UI should display `CompileResponse.outputs`, not guess filenames.
- Raw DSL remains the advanced override. If the active DSL shape is not builder-compatible, the structured recording controls should lock rather than pretending they still own the configuration.
- Protobuf expansion is required for runtime telemetry events, not for the initial recording-name or output-preview slice because those can flow through the existing DSL + compile contract.

Acceptance criteria:

- The intern knows what data structure owns each field.
- The product model is stable enough to implement without adding another temporary layer.

## Phase 3: Extend The Shared Contract

- [ ] Add protobuf messages or fields for telemetry and any later recording-configuration inputs that cannot cleanly flow through DSL.
- [ ] Add protobuf messages or events for runtime telemetry:
  - audio meter samples
  - disk telemetry
  - output path preview if delivered from the backend
- [ ] Regenerate Go and TypeScript protobuf outputs.
- [ ] Update any mapping helpers in `internal/web/`.

Acceptance criteria:

- The missing configuration and telemetry shapes are defined in protobuf.
- The transport contract remains schema-first instead of ad hoc.

## Phase 4: Backend Recording-Configuration Surface

- [ ] Implement any backend endpoint or request-shape expansion needed for:
  - recording name
  - destination root directory
  - output path preview
- [ ] Ensure the backend resolves output files using the same rules the runtime will actually use.
- [ ] Decide how destination previews are produced:
  - normalize response
  - compile response
  - dedicated preview endpoint
- [ ] Add validation for invalid or unsafe destination inputs.

Acceptance criteria:

- The UI can ask the backend what it would write before recording starts.
- Name and destination handling are validated by the backend, not guessed in the UI.

## Phase 5: Backend Telemetry Surface

- [ ] Implement live audio meter collection for the active input path.
- [ ] Implement disk telemetry collection for the chosen output destination or current filesystem.
- [ ] Publish telemetry on the websocket event stream at a controlled cadence.
- [ ] Ensure every goroutine involved in telemetry uses context cancellation and has clear ownership.
- [ ] Bound memory and event volume so telemetry does not flood the UI.

Acceptance criteria:

- The backend publishes real meter and disk telemetry.
- Telemetry shuts down cleanly with the session or connection lifecycle.

## Phase 6: Frontend Recording Configuration UI

- [ ] Add a recording name field to the mounted UI.
- [ ] Replace the fake `Save to` selector with a real destination directory control.
- [ ] Show resolved destination file paths and names in the UI before recording.
- [ ] Ensure the current configuration view reflects backend-resolved paths rather than frontend guesses.
- [ ] Keep the UI explicit about pending/invalid states while the backend recomputes previews.

Acceptance criteria:

- A user can see and edit the name and destination fields.
- A user can see what files will be produced before recording begins.

## Phase 7: Frontend Audio Telemetry And Device UX

- [ ] Replace hardcoded microphone choices with discovered audio devices.
- [ ] Wire gain changes to real backend-facing configuration.
- [ ] Render live meter samples in `MicPanel`.
- [ ] Ensure the panel gracefully shows “unavailable” only when telemetry is actually unavailable.

Acceptance criteria:

- Audio input choices come from backend discovery.
- The meter is real, not simulated.

## Phase 8: Frontend Status Telemetry

- [ ] Replace placeholder disk telemetry in `StatusPanel` with real backend-fed status.
- [ ] Show destination and capacity context clearly enough that a user can reason about risk before recording.
- [ ] Ensure the UI degrades cleanly when telemetry cannot be collected.

Acceptance criteria:

- Disk/status UI reflects real state.
- Placeholder `n/a` states only appear when data is genuinely unavailable.

## Phase 9: Validation And Smoke Tests

- [ ] Add Go tests for new request validation and websocket telemetry mapping.
- [ ] Add frontend tests or Storybook states for:
  - valid destination preview
  - invalid destination
  - meter active
  - meter unavailable
  - low disk warning
- [ ] Run a manual smoke test that verifies:
  - changing recording name changes resolved output names
  - changing destination changes resolved output paths
  - the meter moves while audio is active
  - the disk panel updates meaningfully

Acceptance criteria:

- The ticket includes a repeatable validation recipe.
- The intern can prove the UI is no longer faking these surfaces.

## Suggested Commit Boundaries

- [ ] Commit 1: protobuf contract expansion for config and telemetry
- [ ] Commit 2: backend destination preview and validation
- [ ] Commit 3: backend telemetry collection and websocket events
- [ ] Commit 4: frontend recording configuration UI
- [ ] Commit 5: frontend meter and disk telemetry rendering
- [ ] Commit 6: tests, docs, and smoke validation
