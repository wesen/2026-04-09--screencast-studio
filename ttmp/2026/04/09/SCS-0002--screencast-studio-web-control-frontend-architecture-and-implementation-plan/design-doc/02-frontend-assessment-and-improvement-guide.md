---
Title: Frontend Assessment and Improvement Guide
Ticket: SCS-0002
Status: active
Topics:
    - backend
    - frontend
    - video
    - audio
    - dsl
    - cli
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/api_types.go
      Note: |-
        Backend transport types inspected as the canonical HTTP and websocket contract
        Canonical backend transport schema used as frontend source of truth
    - Path: internal/web/handlers_api.go
      Note: Backend REST routes inspected to compare actual endpoints to frontend assumptions
    - Path: internal/web/handlers_preview.go
      Note: Preview request and response semantics inspected for frontend integration review
    - Path: internal/web/handlers_ws.go
      Note: Backend websocket event envelope and bootstrap events inspected for client alignment
    - Path: ui/src/api/types.ts
      Note: |-
        Frontend transport types inspected for backend contract drift
        Frontend transport types drift from the current Go contract
    - Path: ui/src/components/studio/StudioApp.tsx
      Note: |-
        Main mounted app shell inspected for recording and websocket behavior
        Mounted shell still simulates recording-adjacent state and needs cleanup
    - Path: ui/src/features/session/wsClient.ts
      Note: |-
        WebSocket client inspected for event-contract alignment and lifecycle ownership
        WebSocket client reviewed for event handling and lifecycle ownership
    - Path: ui/src/mocks/handlers.ts
      Note: |-
        Mock API layer inspected for stale endpoint definitions that hide integration bugs
        Mock layer currently hides backend integration drift
    - Path: ui/src/pages/StudioPage.tsx
      Note: |-
        Alternate app shell inspected for duplication and incomplete handlers
        Alternate shell demonstrates desired layout but duplicates orchestration
ExternalSources: []
Summary: Evidence-based review of the current ui/ frontend, with findings, strengths, an architecture cleanup plan, and concrete next steps for the intern.
LastUpdated: 2026-04-09T18:15:00-04:00
WhatFor: Help the intern understand what is good, what is broken, and how to turn the current frontend into a real client of the backend instead of a mostly disconnected prototype.
WhenToUse: Read before doing more frontend work, before reviewing frontend pull requests, or when aligning the UI with the current Go web API.
---


# Frontend Assessment and Improvement Guide

## Executive Summary

The current `ui/` frontend is in an in-between state. It is not a failed start, but it is also not yet a correct implementation of the web ticket. The strongest part of the work is the component and styling layer. The visual system, Storybook setup, and many of the UI building blocks are usable and worth keeping. The weakest part is the integration layer. The frontend currently behaves more like a rich product mock than like a real client of the Go backend.

That distinction matters. The backend now exposes a concrete contract for discovery, DSL normalization and compilation, recording session control, preview management, and WebSocket event delivery. The frontend does not yet consume that contract correctly. Several API slices target endpoints that do not exist, several request and response types are stale, the main app shell still simulates recording-adjacent behavior locally, and the mock layer reinforces that drift instead of catching it.

The correct next move is not to throw the UI away. The correct next move is to preserve the component library and visual language, then re-center the frontend around the real backend contract. This document explains exactly what to keep, what to replace, and how to proceed in a way that is understandable to a new intern.

## Findings First

### Critical Findings

- The frontend HTTP contract has drifted from the real backend contract. This is the most important issue in the codebase today.
- The mounted app shell still uses local simulated state for recording-adjacent behavior instead of driving the UI from live server state.
- There are two overlapping app-shell directions, `StudioApp` and `StudioPage`, which creates duplicated behavior and no clear ownership boundary.
- The WebSocket client only partially understands the current server event protocol and ignores important event types.
- The MSW mock handlers are based on stale routes and stale payload shapes, so mocked development can appear healthy while real integration is broken.

### Medium Findings

- The frontend type layer in `ui/src/api/types.ts` appears to have been designed from an earlier backend idea rather than generated or derived from the current Go transport schema.
- The Redux slices still model a product mock and demo session state rather than a DSL-backed studio editing model.
- `ui/package.json` defines a lint script, but linting is not actually configured, so basic static hygiene is missing.
- Built artifacts exist inside `ui/dist/` and `ui/storybook-static/`, which should usually be treated as generated outputs rather than source.

### Low Findings

- There are useful pieces of exploratory work that should be retained, especially the retro control-surface styling, Storybook coverage, and several reusable panel components.
- The overall module layout in `ui/src/components/`, `ui/src/features/`, and `ui/src/api/` is directionally correct, even though the ownership boundaries inside those directories need cleanup.

## Scope Of This Review

This review covers the current frontend as it exists in `ui/`, with special attention to:

- whether the frontend matches the current backend contract
- whether the UI state model matches the real product model
- whether component work should be kept or replaced
- what sequence of changes the intern should make next

The review does not attempt to redesign the backend. It treats the Go web layer as the current source of truth for transport behavior:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/api_types.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go`

## Validation Evidence

The assessment is based on direct inspection of the frontend and backend files plus basic frontend validation commands.

Commands run:

```bash
pnpm --dir ui build
pnpm --dir ui lint
pnpm --dir ui build-storybook
```

Observed results:

- `pnpm --dir ui build` passed.
- `pnpm --dir ui build-storybook` passed.
- `pnpm --dir ui lint` failed because ESLint configuration is missing even though a lint script exists in `ui/package.json`.

This matters because it tells us the frontend is syntactically buildable, but not yet held to a meaningful static quality bar.

## Current Strengths

Before criticizing the implementation, it is important to say clearly what is good.

### 1. The Visual Direction Is Strong

The current CSS token and layout work successfully captures the intended “hardware control panel” feel from the imported JSX mock. The theme is opinionated, readable, and consistent rather than generic.

Important files:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/styles/tokens.css`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/styles/studio.css`

This work should be preserved.

### 2. Storybook Was A Good Call

The existence of Storybook is a real strength. It means components can be reviewed and refined independently of backend availability. For a control-heavy UI like this, that is useful.

Important files:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/.storybook/main.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/.storybook/preview.ts`

The correct fix is not “remove Storybook.” The correct fix is “make the real app architecture as disciplined as the component work.”

### 3. The Component Breakdown Is Serviceable

Several component boundaries are reasonable and worth keeping:

- `SourceCard`
- `OutputPanel`
- `MicPanel`
- `LogPanel`
- `DSLEditor`
- `PreviewStream`

Representative files:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/source-card/SourceCard.tsx`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/OutputPanel.tsx`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/dsl-editor/DSLEditor.tsx`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/log-panel/LogPanel.tsx`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/preview/PreviewStream.tsx`

These components should be re-wired, not discarded.

## Detailed Current-State Analysis

## 1. Frontend Workspace And Tooling

The current frontend workspace is a Vite + React + TypeScript application with Redux Toolkit, RTK Query, Storybook, and MSW. That stack is fine for this project. It matches the constraints in the repo instructions and is a reasonable fit for a control-oriented frontend.

The relevant package manifest is:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/package.json`

The immediate tooling problem is not the selected stack. The immediate tooling problem is inconsistency:

- lint is declared but not configured
- build artifacts appear to live inside the workspace
- the mock layer is not enforced to match the backend contract

Recommended interpretation:

- keep the stack
- add a real lint configuration
- treat mock and build infrastructure as part of the architecture, not as optional extras

## 2. App Shell Architecture

There are currently two different “top-level UI direction” files:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/StudioApp.tsx`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`

This is a design smell.

`StudioApp.tsx` appears to be the mounted, product-facing shell. It owns:

- WebSocket connection startup
- simulated elapsed timer behavior
- simulated disk usage behavior
- simulated mic meter behavior
- local recording toggles with TODOs for real API calls

`StudioPage.tsx` appears to be an alternate shell that introduces tabs, log view, and raw DSL editing, but it is full of placeholder handlers such as `console.log(...)` and does not look like the true application entry point.

That means the codebase currently has:

- one shell that is mounted but still simulating behavior
- another shell that explores the desired screen composition but is not yet real

The result is architectural duplication without a clear winner.

### Recommendation

Pick one shell. Do not continue developing both.

The cleanest choice is:

- keep `StudioPage` as the primary screen composition model because it already includes studio, logs, and raw DSL editing as first-class areas
- move useful behavior from `StudioApp` into container hooks or a single page-level controller
- delete or demote the other shell once the migration is complete

If you do not collapse the shell architecture soon, every feature will continue to be implemented twice.

## 3. Backend Contract Alignment

This is the single most important section of the review.

The backend contract is now concrete. The frontend contract assumptions in `ui/src/api/` are not aligned with it.

### The Real Backend Contract

Current server behavior is defined by:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/api_types.go`

Representative REST routes:

- `GET /api/discovery`
- `POST /api/setup/normalize`
- `POST /api/setup/compile`
- `GET /api/recordings/current`
- `POST /api/recordings/start`
- `POST /api/recordings/stop`
- `GET /api/previews`
- `POST /api/previews/ensure`
- `POST /api/previews/release`
- `GET /api/previews/:id/mjpeg`

Representative websocket behavior:

- websocket endpoint is `/ws`
- messages are sent as `ServerEvent` envelopes
- the server sends bootstrap events for `session.state` and `preview.list`

### The Current Frontend Assumptions

Problem files:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/discoveryApi.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/setupsApi.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/recordingApi.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/previewsApi.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/types.ts`

The specific drift is substantial:

- frontend uses `/api/setups/...` while backend exposes `/api/setup/...`
- frontend expects `/api/setups/example` and `/api/setups/validate`, which do not exist
- frontend expects discovery subroutes like `/api/discovery/:kind` and `/api/discovery/refresh`, which do not exist
- frontend models preview ensure as `{ source_id }`, but backend requires `{ dsl, source_id }`
- frontend models preview release differently than backend, which requires `preview_id`
- frontend types assume payload shapes that do not match backend envelopes

This is not a small bug. This means the UI is not yet wired to the real server.

### Why This Matters

If the transport types are wrong, then:

- RTK Query hooks cannot be trusted
- mocks give false confidence
- integration bugs show up late
- the frontend becomes harder to review because there is no clear source of truth

### Required Fix

The frontend transport layer must be rebuilt from the backend contract outward.

A good rule for the intern:

- `internal/web/api_types.go` is the contract to mirror
- `ui/src/api/types.ts` must be updated to match it exactly
- mock handlers must be updated only after the real transport layer is correct

Pseudo-workflow:

```text
1. read Go transport structs and handlers
2. rewrite TypeScript transport types
3. rewrite RTK Query endpoints
4. remove routes that do not exist
5. update UI consumers
6. update MSW handlers to match reality
7. test against the real server before trusting mocks
```

## 4. State Ownership And Product Modeling

The Redux setup is structurally promising, but the current state contents still reflect a product mock rather than the actual system.

Important files:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/studio-draft/studioDraftSlice.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/session/sessionSlice.ts`

### Current Situation

`studioDraftSlice.ts` models a set of UI-facing source cards with synthetic properties such as scenes, armed/solo toggles, demo selections, and optimistic local editing state. That is fine for early mock work, but it is not yet the right long-term product model.

The real product model is closer to:

- one editable DSL document
- one normalized effective config
- one compiled plan
- one discovery snapshot
- zero or more preview sessions derived from the current DSL
- zero or one recording session derived from the current DSL

The frontend should eventually reflect those concepts directly.

### Better Model

Instead of treating the draft as a standalone UI invention, model it as a projection of the DSL and its normalized/compiled results.

Suggested state partitions:

- `editor`
  - raw DSL text
  - dirty flag
  - last normalize result
  - last compile result
- `discovery`
  - latest server snapshot
  - loading and error state
- `recording`
  - current server session
  - start and stop mutation status
- `previews`
  - preview descriptors keyed by preview ID and source ID
- `transport`
  - websocket connected flag
  - last event timestamp

That produces a cleaner architecture than a giant hand-maintained “studio draft” object that pretends to already know the whole product.

### State Diagram

```text
raw DSL text
   |
   v
normalize request
   |
   +--> normalized config
   |
   v
compile request
   |
   +--> compiled outputs and jobs
   |
   +--> preview ensure requests
   |
   +--> recording start request
```

The key design idea is that the UI is editing a setup, not editing a collection of arbitrary frontend-only objects.

## 5. WebSocket Layer

The current WebSocket client is a partial implementation:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/session/wsClient.ts`

The reconnection logic is useful. The contract handling is not yet correct enough.

### Problems

- It assumes event payloads that do not match the server.
- It ignores `preview.list`, even though the backend sends it immediately on connection.
- It treats `preview.state` as future work rather than part of the current runtime behavior.
- It uses a singleton client with no clear ownership boundary.
- The mounted shell intentionally leaves the connection alive on unmount, which makes lifecycle reasoning harder.

### Better Model

The websocket client should not interpret events ad hoc in random component code. It should be a transport adapter with a small, explicit decoding layer.

Suggested event handling architecture:

```text
WebSocket
  -> parse ServerEvent envelope
  -> narrow by event.type
  -> dispatch typed Redux action
  -> reducers update canonical slices
  -> selectors drive UI
```

Pseudocode:

```ts
type ServerEvent =
  | { type: 'session.state'; payload: RecordingSession }
  | { type: 'session.log'; payload: ProcessLog }
  | { type: 'preview.list'; payload: PreviewListResponse }
  | { type: 'preview.state'; payload: PreviewDescriptor }

function handleServerEvent(event: ServerEvent) {
  switch (event.type) {
    case 'session.state':
      dispatch(sessionReplaced(event.payload));
      return;
    case 'session.log':
      dispatch(sessionLogAdded(event.payload));
      return;
    case 'preview.list':
      dispatch(previewsReplaced(event.payload.previews));
      return;
    case 'preview.state':
      dispatch(previewUpserted(event.payload));
      return;
  }
}
```

The important point is that the event protocol should be expressed explicitly in code, not left as “some JSON we hope matches.”

## 6. Preview Integration

The backend preview flow is clear:

- frontend provides the current DSL and a source ID
- server ensures a preview worker exists
- frontend renders `/api/previews/:id/mjpeg`
- frontend releases the preview by `preview_id`

Relevant backend file:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`

The frontend currently does not model that lifecycle cleanly enough. In particular, it is easy to mix up:

- source ID
- preview ID
- preview state
- preview stream URL

### Recommended Frontend Preview Model

Each source card should not create its own private preview protocol. It should ask a shared preview slice for the preview associated with a source.

Pseudocode:

```ts
async function ensurePreviewForSource(sourceId: string, dsl: string) {
  const response = await ensurePreviewMutation({ source_id: sourceId, dsl }).unwrap();
  dispatch(previewUpserted(response.preview));
}

function previewStreamUrl(previewId: string) {
  return `/api/previews/${previewId}/mjpeg`;
}
```

Component ownership:

- container decides whether preview should exist
- preview slice stores preview descriptors
- `PreviewStream` only renders an image stream for a known preview ID

That keeps view code simple.

## 7. Mock Layer Drift

The current MSW handlers are actively hiding backend/frontend contract errors.

Problem file:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/mocks/handlers.ts`

Examples of stale behavior:

- mocking `/api/setups/...` instead of `/api/setup/...`
- mocking discovery subroutes that do not exist
- returning preview payloads that do not match backend envelopes
- returning recording payloads without the real session envelope shape

This is dangerous because it produces a false sense of progress. A mock layer should help catch interface mismatches, not normalize them away.

### Rule For The Intern

Never update the mock layer first.

Correct sequence:

```text
1. fix real TypeScript transport types
2. fix RTK Query endpoints
3. manually test against the real Go server
4. update MSW handlers to match the verified real contract
5. use Storybook/MSW for stable isolated UI work
```

## 8. The Current Frontend Is More Prototype Than Product

This is the conceptual diagnosis that ties the findings together.

`StudioApp.tsx` simulates elapsed time, disk usage, and microphone behavior locally. `StudioPage.tsx` includes tabs and editing affordances but still uses many placeholder handlers. That means the frontend currently demonstrates the product concept, but it does not yet implement the product truthfully.

That is okay for an early phase. It is not okay as the steady state.

The architecture should now move from:

```text
UI-first mock with backend-shaped intentions
```

to:

```text
backend-driven application with reusable UI components
```

That is the central improvement goal.

## Proposed Cleanup Architecture

## Design Principles

- Keep the visual system and component library.
- Make the Go backend the source of truth for transport.
- Model frontend state around DSL, discovery, previews, and recording, not around demo-only objects.
- Collapse duplicate app shells into one page-level controller.
- Make mocks follow the real contract instead of defining their own.

## Target Frontend Architecture

```text
ui/src
  app/
    store.ts
    hooks.ts
  api/
    transport-types.ts
    discoveryApi.ts
    setupApi.ts
    recordingApi.ts
    previewApi.ts
    wsEvents.ts
  features/
    editor/
    discovery/
    recording/
    previews/
    transport/
  pages/
    StudioPage.tsx
  components/
    primitives/
    preview/
    source-card/
    studio/
    dsl-editor/
    log-panel/
```

Important design choice:

- the page owns orchestration
- the components stay mostly presentational
- the API layer mirrors the Go backend closely

## Proposed Data Flow

```text
on page load
  -> fetch discovery
  -> fetch current recording session
  -> connect websocket
  -> receive session.state + preview.list bootstrap
  -> render current UI from store

when DSL changes
  -> update editor slice
  -> debounce normalize/compile if desired
  -> show warnings and plan summary

when preview requested
  -> call ensurePreview({ dsl, source_id })
  -> receive preview descriptor
  -> render MJPEG stream by preview ID

when record clicked
  -> call startRecording({ dsl, ...options })
  -> websocket updates session state and logs
  -> current session query and websocket remain consistent
```

## Concrete Recovery Plan For The Intern

## Phase A: Freeze The Contract

Goal: make the frontend type and endpoint layer match the Go backend exactly.

Tasks:

- Rewrite `ui/src/api/types.ts` to match `internal/web/api_types.go`.
- Rename `setupsApi.ts` to `setupApi.ts` if needed for terminology alignment.
- Remove nonexistent endpoints from RTK Query slices.
- Update response handling to respect backend envelopes.

Acceptance criteria:

- every frontend endpoint maps to a real backend route
- every frontend transport type corresponds to a real backend payload
- no UI code depends on routes that do not exist

## Phase B: Pick One App Shell

Goal: eliminate duplicated screen orchestration.

Tasks:

- Choose `StudioPage.tsx` as the single main shell.
- Move useful mounted behavior from `StudioApp.tsx` into page-level hooks or containers.
- Delete or archive the losing shell once behavior is migrated.

Acceptance criteria:

- one mounted page owns the studio screen
- no duplicate recording-control logic remains
- tabs, DSL editor, logs, and controls live in one coherent shell

## Phase C: Rebuild State Around Real Domain Concepts

Goal: replace product-mock slices with backend-aligned state.

Tasks:

- Introduce slices or RTK Query selectors for:
  - discovery
  - editor DSL text
  - normalized config
  - compiled plan
  - preview descriptors
  - recording session
- Reduce `studioDraftSlice` until it only represents true client-owned UI state.

Acceptance criteria:

- source lists come from discovery or normalized config, not hardcoded demo state
- recording UI is driven by real session state
- preview state is keyed by real preview IDs

## Phase D: Fix WebSocket Ownership And Decoding

Goal: make websocket behavior explicit, typed, and maintainable.

Tasks:

- Add typed event definitions matching backend event names.
- Decode `preview.list`, `preview.state`, `session.state`, and `session.log`.
- Replace singleton ambiguity with a clear app-lifecycle owner.
- Ensure disconnect logic is explicit.

Acceptance criteria:

- websocket bootstrap events populate the store correctly
- reconnect behavior is preserved
- no event handling relies on guessed payload shapes

## Phase E: Repair The Mock Layer

Goal: make mock-driven development trustworthy again.

Tasks:

- Rewrite `ui/src/mocks/handlers.ts` to match the real backend contract.
- Remove fake endpoints that do not exist.
- Add realistic envelope shapes.

Acceptance criteria:

- MSW responses are valid examples of the real server contract
- Storybook and local mocked work do not hide real integration bugs

## Phase F: Add Missing Frontend Hygiene

Goal: stop the codebase from drifting again.

Tasks:

- add ESLint configuration
- add a small number of strict TypeScript and React rules
- decide whether generated `dist/` and `storybook-static/` should remain in the repo
- add a frontend validation section to CI later

Acceptance criteria:

- `pnpm --dir ui lint` passes
- the frontend has a documented validation loop
- generated outputs are handled intentionally

## Suggested File-Level Plan

If an intern is starting from scratch on the cleanup, this order is sensible:

1. Read:
   - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/api_types.go`
   - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go`
   - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`
   - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go`
2. Rewrite:
   - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/types.ts`
   - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/discoveryApi.ts`
   - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/setupsApi.ts`
   - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/recordingApi.ts`
   - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/previewsApi.ts`
3. Collapse app shell duplication:
   - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/StudioApp.tsx`
   - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`
4. Fix websocket event decoding:
   - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/session/wsClient.ts`
5. Repair mocks:
   - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/mocks/handlers.ts`
6. Add frontend hygiene:
   - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/package.json`
   - ESLint config files to be added under `ui/`

## Review Guidance For The Intern

When you work through this cleanup, do not ask “does the screen look right?” first. Ask these questions first:

- Does this UI action call a real backend route?
- Does the request payload match what the server actually expects?
- Does the response type match what the server actually returns?
- Is this state truly owned by the client, or am I inventing local state that should come from the server?
- Am I adding a second orchestration path where one already exists?

Only after those questions are answered should you spend time on polish.

## A Simple Mental Model

If you are new to this codebase, use this picture:

```text
component library = mostly good
screen composition = promising but duplicated
state model = halfway between demo and product
transport layer = currently wrong in important ways
mocks = misleading
next step = integrate for real, not redesign visually
```

That is the shortest accurate summary of the frontend today.

## Alternatives Considered

### Alternative 1: Rewrite The Frontend From Scratch

Rejected because:

- too much useful component and styling work would be lost
- the main problems are contract and ownership problems, not visual problems
- a rewrite would likely repeat the same drift unless the contract issue is fixed first

### Alternative 2: Keep The Current Mock-Driven Flow And Patch Integration Later

Rejected because:

- the mock layer is already encoding the wrong API
- more mock-first work would deepen the cleanup cost
- it would keep rewarding frontend-only invention instead of real integration

### Alternative 3: Treat The Current UI As “Good Enough” And Move On

Rejected because:

- the backend/frontend mismatch is too large
- future bug reports would become harder to diagnose
- the web ticket would look complete while still being structurally wrong

## Open Questions

- Should the frontend eventually edit structured form state first and generate DSL, or should raw DSL remain the primary editing surface in version 1?
- Do we want preview acquisition to be entirely explicit in the UI, or should the page auto-ensure previews for visible sources?
- Should transport types eventually be code-generated from a shared schema, or is manual mirroring sufficient for this project size?
- Should Storybook remain MSW-backed only, or should some stories consume exported fixture objects that are also used in integration tests?

## Recommended Immediate Next Steps

The next concrete work items for the intern should be:

1. Rewrite the frontend transport types and endpoint definitions to match the Go backend.
2. Collapse to a single app shell.
3. Remove local simulated recording behavior.
4. Make websocket event handling match the current server events.
5. Update MSW handlers after the real contract is correct.
6. Add ESLint and front-end validation commands to the review checklist.

## References

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/package.json`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/types.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/discoveryApi.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/setupsApi.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/recordingApi.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/previewsApi.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/StudioApp.tsx`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/studio-draft/studioDraftSlice.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/session/sessionSlice.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/session/wsClient.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/mocks/handlers.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/styles/tokens.css`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/styles/studio.css`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/api_types.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go`
