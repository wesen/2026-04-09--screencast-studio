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

## Step 2: Freeze The Current Product Gap And Choose A Simple Ownership Model

The first implementation step was analysis rather than code. I audited the mounted recording controls, the compile/normalize path, and the protobuf/websocket surface to decide where the productized configuration should live. The important outcome was deciding not to invent another overlay config transport for the initial slice.

The chosen model is deliberately simple: recording name, destination naming, microphone choice, and gain should live in the structured DSL-backed setup draft when the builder is active. Planned outputs should come from the existing compile endpoint. Protobuf expansion should be reserved for live telemetry and any truly runtime-only state.

### Prompt Context

**User prompt (verbatim):** "go ahead."

**Assistant interpretation:** Start implementing `SCS-0004` task by task, beginning with the current-state audit and product-model decisions, and keep the ticket diary updated while working.

**Inferred user intent:** Move from planning into real implementation, but do it in disciplined slices so the architecture stays coherent and reviewable.

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
- Updated the design doc with the current-reality audit and the ownership decision.

### Why

- The backend already resolves real output paths from DSL. That should be reused rather than duplicated.
- The frontend already has a structured setup draft. Reusing it is much simpler than adding another temporary config layer.
- Meter and disk telemetry are truly runtime-only and therefore still belong on the protobuf websocket path.

### What worked

- The current codebase was already structured enough that the product-model decision could be grounded in real file ownership instead of speculation.
- The compile path was already good enough to become the authoritative output preview.

### What didn't work

- N/A

### What I learned

- `studioDraft` was the last major fake-state pocket in the mounted app.
- The easiest clean path is not “invent a richer recording API”; it is “make the mounted UI edit the same DSL-backed state the runtime already understands.”

### What was tricky to build

- The tricky part here was avoiding a bad architecture fork. There was a tempting path where recording name, destination, mic input, and gain would be sent as a second overlay object while the builder continued generating DSL independently. That would have increased drift risk immediately because compile preview, preview ensuring, and recording start would all need to reconcile DSL plus overlay state.

### What warrants a second pair of eyes

- The decision to use `setupDraft.audioSources[0]` as the v1 microphone control surface should be revisited once audio-source management grows more sophisticated.
- The structured-builder policy for rewriting destination templates should be checked against any advanced DSL expectations before it is treated as final UX.

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

## Step 3: Replace Fake Recording Controls With DSL-Backed State And Compile Preview

The first code slice focused on making the mounted recording controls honest without widening the backend contract yet. The core change was deleting the fake `studioDraft` ownership model and moving the visible recording configuration back onto the structured DSL draft that already powers the builder. That let the mounted UI edit real state and keep using the existing normalize/compile flow instead of inventing a parallel recording-config transport.

This slice deliberately stopped short of telemetry. Name, destination-root editing, planned output preview, discovered microphone choices, and gain are now real enough to drive the DSL and therefore the runtime preview/record path. Live meter and disk telemetry remain separate follow-up work because they genuinely need runtime-side event publishing.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Keep implementing `SCS-0004` in concrete slices, commit at useful boundaries, and maintain a detailed diary.

**Inferred user intent:** Turn the visible recording controls into real product behavior without losing architectural clarity.

### What I did

- Deleted the obsolete fake Redux slice:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/studio-draft/studioDraftSlice.ts`
- Extended structured setup state and helpers:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/setupDraftSlice.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/conversion.ts`
- Added compile-preview state to the setup slice:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup/setupSlice.ts`
- Reworked the mounted page:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`
- Reworked the recording widgets:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/OutputPanel.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/MicPanel.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/StatusPanel.tsx`
- Updated Storybook wiring and stories:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/StudioPage.stories.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/OutputPanel.stories.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/MicPanel.stories.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/StatusPanel.stories.tsx`

### Why

- The previous `studioDraft` slice was the last major fake-state pocket in the mounted app.
- The runtime already understands the DSL and the compile endpoint already resolves planned outputs. Reusing those paths is simpler and less risky than layering another temporary request model on top.
- Tightening component prop types around plain display data instead of protobuf message instances keeps the UI boundary cleaner and made Storybook easier to validate.

### What worked

- Recording name now maps to `setupDraft.sessionId`.
- Destination-root editing now rewrites the builder-supported templates directly.
- Planned outputs are rendered from backend compile results instead of frontend guesses.
- Microphone choices now come from discovery rather than hardcoded labels.
- Gain changes now rewrite the primary audio source in the DSL-backed setup draft.
- Storybook, lint, UI build, Go build, Go test, and `docmgr doctor` all passed.
- The live browser smoke confirmed:
  - editing `Name` updated planned outputs to `captures/session-check/...`
  - editing `Save to` updated both planned outputs and the status destination line
  - the microphone combobox contained discovered PulseAudio device IDs and switched successfully

### What didn't work

- The first build attempt failed because `OutputPanel` accepted generated protobuf `PlannedOutput` message types in its Storybook props. The exact TypeScript error was:

```text
Type '{ kind: string; sourceId: string; name: string; path: string; }' is not assignable to type 'PlannedOutput'.
Property '$typeName' is missing in type ...
```

- The fix was to narrow the component to a plain display-oriented output-preview type instead of using generated transport types directly.
- The first Playwright attempt was blocked by a stale shared profile lock, and a second retry briefly hit a closed MCP transport before the browser lane recovered.

### What I learned

- The existing compile endpoint is enough to power the pre-record output preview; a new preview endpoint is not needed for the first productization slice.
- The component layer should not take protobuf-generated message types unless it genuinely needs transport semantics. Plain view-model props are cleaner.

### What was tricky to build

- The sharp edge was state ownership, not UI markup. `StudioPage` already coordinates raw DSL apply, structured draft edits, preview lifecycles, normalize calls, and session transport. Adding recording configuration on top of that could have easily reintroduced a split-brain model. The fix was to keep one rule: structured edits mutate the structured draft, then the page re-renders canonical DSL from that draft, and normalize/compile continue to treat that DSL as the source of truth.

### What warrants a second pair of eyes

- The current destination-root editor only supports the builder-managed default template shape (`per_source` and `audio_mix`). That limitation is intentional but should be reviewed against any advanced template expectations.
- `Multi-track` is still not a real runtime capability in this slice; the UI remains visible but not product-complete there.

### What should be done in the future

- Add real telemetry protobuf events and backend publishers for audio meter and disk status.
- Decide whether format/fps/quality should be generalized into first-class builder defaults rather than “apply to current sources” semantics in the mounted page.
- Add a broader browser-level smoke once the telemetry slice exists so the mic/status panels can be validated end to end.

### Code review instructions

- Start with the mounted orchestration:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`
- Then review the draft helpers:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/setupDraftSlice.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/conversion.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup/setupSlice.ts`
- Finally review the rendered widgets:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/OutputPanel.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/MicPanel.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/StatusPanel.tsx`

### Technical details

Validation commands used in this step:

```bash
pnpm --dir ui lint
pnpm --dir ui build
CI=1 pnpm --dir ui build-storybook
go test ./...
go build ./...
docmgr doctor --ticket SCS-0004 --stale-after 30
lsof-who -p 18082 -k || true
tmux new-session -d -s scs-0004 'cd /home/manuel/code/wesen/2026-04-09--screencast-studio && go run ./cmd/screencast-studio serve --addr :18082 --static-dir ui/dist'
curl -sSf http://127.0.0.1:18082/api/healthz
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
