---
Title: Frontend Cleanup and Backend Alignment System Design
Ticket: SCS-0003
Status: active
Topics:
    - frontend
    - backend
    - ui
    - architecture
    - dsl
    - video
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/api_types.go
      Note: |-
        Canonical HTTP payload schema for the frontend to mirror
        Canonical backend payload contract for the cleanup
    - Path: internal/web/handlers_api.go
      Note: Canonical REST route surface for setup and recording transport
    - Path: internal/web/handlers_preview.go
      Note: Canonical preview lifecycle contract
    - Path: internal/web/handlers_ws.go
      Note: |-
        Canonical WebSocket event envelope and bootstrap messages
        Canonical websocket event surface for the cleanup
    - Path: ttmp/2026/04/09/SCS-0002--screencast-studio-web-control-frontend-architecture-and-implementation-plan/design-doc/02-frontend-assessment-and-improvement-guide.md
      Note: |-
        Assessment baseline that motivated this cleanup ticket
        Precursor frontend assessment that this cleanup ticket operationalizes
    - Path: ui/src/App.tsx
      Note: |-
        Current frontend entrypoint mounts the older StudioApp shell
        Current entrypoint must be changed to mount the surviving cleanup shell
    - Path: ui/src/api/discoveryApi.ts
      Note: Discovery query layer currently targets stale endpoints
    - Path: ui/src/api/previewsApi.ts
      Note: |-
        Preview API contract currently does not match backend ensure and release semantics
        Preview lifecycle must match backend ensure and release semantics
    - Path: ui/src/api/recordingApi.ts
      Note: Recording API contract must be realigned with backend envelopes
    - Path: ui/src/api/setupsApi.ts
      Note: |-
        Setup query layer uses deprecated route names and stale request semantics
        Deprecated route naming and stale contract should be replaced
    - Path: ui/src/api/types.ts
      Note: |-
        Frontend transport types have drifted from the backend and need replacement
        Transport types must be replaced with backend-mirrored definitions
    - Path: ui/src/components/studio/StudioApp.tsx
      Note: |-
        Current mounted shell still simulates recording behavior and duplicates orchestration
        Duplicate mounted shell slated for removal or reduction
    - Path: ui/src/features/session/wsClient.ts
      Note: |-
        WebSocket client needs typed decoding and explicit lifecycle ownership
        Websocket decoding and ownership need centralized cleanup
    - Path: ui/src/mocks/handlers.ts
      Note: |-
        Mock handlers currently hide backend integration drift and should be rewritten
        Mock handlers should be rewritten only after real transport alignment
    - Path: ui/src/pages/StudioPage.tsx
      Note: |-
        Alternate shell contains better composition ideas but is not the mounted entrypoint
        Likely long-term shell and orchestration target
ExternalSources: []
Summary: Detailed architecture and implementation guide for cleaning up the frontend, removing duplicate and stale code, and making the UI a strict client of the current Go backend without compatibility layers.
LastUpdated: 2026-04-09T19:18:00-04:00
WhatFor: Guide an intern or follow-on engineer through a deliberate frontend cleanup that removes stale paths and aligns the UI with the backend transport and runtime model.
WhenToUse: Read before starting the frontend cleanup implementation, reviewing frontend cleanup PRs, or deciding which existing UI code should be retained versus deleted.
---


# Frontend Cleanup and Backend Alignment System Design

## Executive Summary

This ticket is a cleanup and realignment ticket, not a feature-expansion ticket. The goal is to turn the current `ui/` codebase from a partially integrated product mock into a real frontend for the existing backend. The key outcome is architectural clarity: one mounted app shell, one transport contract, one WebSocket event model, one source of truth for setup and session state, and no preservation of stale or deprecated frontend paths.

The current frontend already contains useful work. The visual language is strong, the component library is promising, and Storybook is a good investment. The problem is the integration layer. The frontend currently contains duplicate shells, stale API assumptions, mock handlers that hide real backend drift, and client-owned demo state that should be replaced by backend-aligned state. Those problems should be fixed by deleting or rewriting the incorrect paths, not by layering compatibility adapters on top.

This document assumes the backend HTTP and WebSocket surface in `internal/web/` is now the source of truth. The cleanup plan deliberately avoids backwards compatibility. If the frontend contains deprecated route names, duplicate orchestration paths, or state models that no longer match the product, they should be removed.

## Problem Statement

The frontend in `ui/` is currently in an awkward middle state:

- the visual shell is farther along than the transport integration
- the current mounted app and the more complete page composition are split across two top-level shells
- the frontend transport layer does not match the implemented backend contract
- the WebSocket client only partially understands server events
- the mock layer encodes stale routes and stale payloads
- several pieces of client state still simulate product behavior instead of reflecting real backend state

This creates four concrete risks.

### Risk 1: A UI That Looks Done But Is Not Correct

Because the interface renders and the stories build, the frontend can appear more complete than it really is. But a user clicking through the UI is not yet interacting with the same contract exposed by the backend. That leads to false confidence and delayed discovery of integration bugs.

### Risk 2: Duplicate Orchestration Will Compound Future Work

The presence of both `ui/src/components/studio/StudioApp.tsx` and `ui/src/pages/StudioPage.tsx` means future features can accidentally be added twice, or added to the wrong shell. If this is not collapsed now, every later improvement will cost more.

### Risk 3: The Mock Layer Rewards The Wrong Behavior

`ui/src/mocks/handlers.ts` currently acts as a parallel API fiction. That is dangerous. Instead of helping isolated development, it rewards stale assumptions and hides real transport mismatches.

### Risk 4: The State Model Is Still Too Demo-Oriented

The current slices are not yet organized around the actual system model:

- DSL text
- normalized config
- compiled plan
- discovery snapshot
- preview sessions
- recording session
- WebSocket connectivity

Instead, some state still resembles a frontend-authored demo of what the product could look like. That state needs to be replaced by a product-aligned model.

## Cleanup Goal

The cleanup target is precise:

- keep useful presentational components and styling
- delete duplicate or deprecated orchestration paths
- replace stale transport types and endpoints
- rewrite mocks to match reality
- make the UI a strict client of the backend

The cleanup target is not:

- preserve old route names
- support both old and new payload shapes
- keep duplicate shells alive “for now”
- carry temporary demo state longer than necessary

## Proposed Solution

The correct solution is a replace-in-place cleanup with explicit deletion.

At a high level:

1. Use the backend in `internal/web/` as the only source of truth for transport.
2. Collapse the frontend to a single mounted shell based on the better page composition.
3. Rebuild the frontend state model around real backend concepts.
4. Make WebSocket event decoding explicit and typed.
5. Rewrite mocks after the real transport layer is correct.
6. Remove stale files, route assumptions, and compatibility ideas as soon as the replacement is ready.

## Target Architecture

### Top-Level Shape

```text
ui/src
  app/
    store.ts
    hooks.ts
    providers.tsx
  api/
    baseApi.ts
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
    dsl-editor/
    log-panel/
    studio/
  mocks/
    handlers.ts
```

### Ownership Model

```text
backend
  -> owns discovery snapshot
  -> owns normalize and compile results
  -> owns recording session state
  -> owns preview lifecycle
  -> owns websocket event stream

frontend
  -> owns page composition
  -> owns local draft text before submit
  -> owns UI-only controls like active tab and panel visibility
  -> owns query/mutation lifecycle state
```

This division is the core of the cleanup. Anything that crosses that line incorrectly should be removed or rewritten.

## Detailed Design

## 1. One Shell Only

The cleanup should standardize on one mounted screen shell. The recommended destination is `ui/src/pages/StudioPage.tsx`, because it already contains the right overall composition:

- studio view
- logs view
- raw DSL editing view

The current `ui/src/App.tsx` mounts `StudioApp`, not `StudioPage`. That should change. Once `StudioPage` is made real, `StudioApp` should either be deleted or reduced to a small wrapper that disappears quickly.

### Why `StudioPage` Should Win

- It is closer to the actual product surface.
- It already organizes the interface as a page instead of a monolithic component.
- It gives the cleanup a natural place to put orchestration code.

### What Must Not Happen

- Do not keep both shells alive “temporarily” for long.
- Do not duplicate logic between `StudioApp` and `StudioPage`.
- Do not leave the entrypoint ambiguous.

### Migration Sketch

```text
today:
  App -> StudioApp

cleanup:
  App -> StudioPageContainer -> StudioPage

delete:
  StudioApp or fold it down to purely presentational pieces
```

## 2. Frontend Transport Must Mirror The Backend Exactly

The frontend transport layer should mirror:

- `internal/web/api_types.go`
- `internal/web/handlers_api.go`
- `internal/web/handlers_preview.go`
- `internal/web/handlers_ws.go`

This is the highest priority change.

### Current Problems

- `ui/src/api/setupsApi.ts` uses stale plural route naming
- `ui/src/api/discoveryApi.ts` assumes discovery subroutes that do not exist
- `ui/src/api/previewsApi.ts` does not match preview ensure and release requirements
- `ui/src/api/recordingApi.ts` does not respect current session envelope semantics
- `ui/src/api/types.ts` is not a faithful mirror of the backend payloads

### Cleanup Rule

If a frontend endpoint or type does not match the backend, replace it. Do not preserve the old name or shape.

### Suggested Type Layout

```ts
export interface DiscoveryResponse {
  displays: DisplayDescriptor[];
  windows: WindowDescriptor[];
  cameras: CameraDescriptor[];
  audio: AudioInputDescriptor[];
}

export interface NormalizeResponse {
  session_id: string;
  warnings: string[];
  config: EffectiveConfig;
}

export interface CompileResponse {
  session_id: string;
  warnings: string[];
  outputs: PlannedOutput[];
  video_jobs: VideoJob[];
  audio_jobs: AudioMixJob[];
}

export interface SessionEnvelope {
  session: RecordingSession;
}

export interface PreviewEnvelope {
  preview: PreviewDescriptor;
}

export interface PreviewListResponse {
  previews: PreviewDescriptor[];
}
```

### API Reference Table

| Operation | Method | Route | Frontend file |
| --- | --- | --- | --- |
| discovery snapshot | `GET` | `/api/discovery` | `ui/src/api/discoveryApi.ts` |
| normalize setup | `POST` | `/api/setup/normalize` | `ui/src/api/setupApi.ts` |
| compile setup | `POST` | `/api/setup/compile` | `ui/src/api/setupApi.ts` |
| current session | `GET` | `/api/recordings/current` | `ui/src/api/recordingApi.ts` |
| start session | `POST` | `/api/recordings/start` | `ui/src/api/recordingApi.ts` |
| stop session | `POST` | `/api/recordings/stop` | `ui/src/api/recordingApi.ts` |
| list previews | `GET` | `/api/previews` | `ui/src/api/previewApi.ts` |
| ensure preview | `POST` | `/api/previews/ensure` | `ui/src/api/previewApi.ts` |
| release preview | `POST` | `/api/previews/release` | `ui/src/api/previewApi.ts` |

## 3. Replace Demo State With Product State

The frontend state model should be rebuilt around the actual product flow:

```text
raw DSL text
  -> normalize
  -> effective config
  -> compile
  -> compiled plan
  -> ensure previews
  -> start recording
  -> observe session and logs through websocket
```

### Recommended Slice Layout

- `editor`
  - `dslText`
  - `dirty`
  - `normalizeWarnings`
  - `compileWarnings`
  - `lastNormalize`
  - `lastCompile`
- `transport`
  - `wsConnected`
  - `lastEventAt`
  - `connectionError`
- `recording`
  - `currentSession`
  - `recentLogs`
- `previews`
  - `byId`
  - `bySourceId`
- `ui`
  - `activeTab`
  - `expandedPanels`
  - `selectedSourceId`

### What To Remove

- simulated elapsed timer increments in the mounted shell
- simulated microphone levels that pretend to come from the server
- synthetic disk usage increments that create fake session behavior
- state that only exists to support a deprecated shell

### Why

The cleanup should make it obvious which data is real and which data is merely local UI state. Any demo-only state that looks like real runtime state is confusing and should go.

## 4. WebSocket Decoding Must Be Typed And Centralized

`ui/src/features/session/wsClient.ts` currently combines useful reconnect behavior with under-specified payload handling. That should be split into clearer responsibilities:

- socket lifecycle and reconnect
- event decoding
- dispatch into store

### Recommended Event Model

```ts
type ServerEvent =
  | { type: 'session.state'; payload: RecordingSession }
  | { type: 'session.log'; payload: ProcessLog }
  | { type: 'preview.list'; payload: PreviewListResponse }
  | { type: 'preview.state'; payload: PreviewDescriptor };
```

### Recommended Event Handling Pseudocode

```ts
function handleServerEvent(event: ServerEvent) {
  switch (event.type) {
    case 'session.state':
      dispatch(recordingSessionReplaced(event.payload));
      return;
    case 'session.log':
      dispatch(recordingLogAdded(event.payload));
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

### Ownership Rule

The websocket client should be started and stopped by one app-level owner. It should not be silently immortal because one component decided never to disconnect on unmount.

## 5. Preview Lifecycle Must Be Modeled Correctly

The backend preview flow is already explicit:

- ensure preview with `{ dsl, source_id }`
- receive a `preview_id`
- render MJPEG stream at `/api/previews/:preview_id/mjpeg`
- release preview by `preview_id`

The frontend should use that exact model.

### Recommended Container Behavior

```ts
async function onPreviewVisible(sourceId: string) {
  const result = await ensurePreview({ dsl, source_id: sourceId }).unwrap();
  dispatch(previewUpserted(result.preview));
}

function previewUrl(previewId: string) {
  return `/api/previews/${previewId}/mjpeg`;
}
```

### Component Rule

`PreviewStream` should remain mostly presentational. It should not invent preview state by itself. It should accept a concrete preview descriptor or preview URL and render that state clearly.

## 6. Mocks Must Follow Reality, Not Invent It

The current mock layer should be treated as stale and replaced after the real API layer is corrected.

### Rules

- no fake route names that the backend does not expose
- no fake response envelopes that flatten real nested payloads
- no fake preview semantics that ignore `preview_id`
- no fake validation endpoints unless the backend adds them

### Development Sequence

```text
1. fix real transport
2. test against real server
3. rewrite mocks
4. use Storybook and local mocks again
```

If the mocks are updated earlier, they will likely re-encode the current frontend assumptions instead of the backend truth.

## 7. Remove Deprecated And Unclear Code Aggressively

This ticket explicitly does not require backwards compatibility. That has implementation consequences.

### Remove Or Rename

- `ui/src/api/setupsApi.ts`
  - replace with `setupApi.ts`
- old transport types in `ui/src/api/types.ts`
  - replace with a backend-mirrored schema
- duplicated shell logic split across `StudioApp` and `StudioPage`
  - keep one, delete the other
- mock routes that do not exist
- generated frontend output checked into source if it is not intentionally committed

### Do Not Add

- compatibility aliases for old endpoint names
- TypeScript union types that support both old and new payload shapes
- temporary adapters whose only job is to preserve stale internal callers
- parallel frontend flows for “legacy” and “new”

The correct cleanup strategy is subtractive. If a path is wrong, remove it.

## Design Decisions

### Decision 1: No Backwards Compatibility Layer

Rationale:

- the codebase is still young
- preserving stale paths will increase maintenance cost immediately
- the user explicitly does not want compatibility work here

### Decision 2: Backend Contract Wins

Rationale:

- the backend is already implemented
- the frontend should consume the product runtime, not redefine it
- this removes ambiguity and speeds review

### Decision 3: One Mounted App Shell

Rationale:

- two shells guarantee duplicated work
- page-level orchestration is easier to review than component-level orchestration spread across duplicates

### Decision 4: Keep Presentational Work Where It Is Good

Rationale:

- the visual layer is not the main problem
- rewrites are expensive and unnecessary when the component work can be retained

### Decision 5: Mocks Are Secondary To Real Integration

Rationale:

- a stale mock layer is worse than no mock layer
- the real server contract is now available locally and should be used first

## Alternatives Considered

### Alternative 1: Add Compatibility Adapters And Migrate Slowly

Rejected because:

- it prolongs ambiguity
- it rewards stale callers
- it directly conflicts with the requirement to avoid backwards compatibility

### Alternative 2: Rewrite The Entire Frontend From Scratch

Rejected because:

- useful presentational work would be lost
- the core problem is alignment and ownership, not total visual failure
- a rewrite would delay learning from the current implementation

### Alternative 3: Keep Both Shells Until The End

Rejected because:

- every new feature would need careful duplicate updates
- reviewers would have to reason about two code paths for the same screen
- the app entrypoint would remain conceptually unclear

## Implementation Plan

## Phase 1: Freeze The Transport Contract

Goal:

- rewrite `ui/src/api/` so it exactly matches the backend

Main files:

- `ui/src/api/types.ts`
- `ui/src/api/discoveryApi.ts`
- `ui/src/api/setupsApi.ts` or renamed `setupApi.ts`
- `ui/src/api/recordingApi.ts`
- `ui/src/api/previewsApi.ts`

Acceptance criteria:

- no frontend endpoint references a nonexistent backend route
- request and response types mirror backend payloads
- basic API smoke tests hit the real server successfully

## Phase 2: Collapse To One Shell

Goal:

- mount `StudioPage`
- move orchestration there
- delete or demote `StudioApp`

Main files:

- `ui/src/App.tsx`
- `ui/src/components/studio/StudioApp.tsx`
- `ui/src/pages/StudioPage.tsx`

Acceptance criteria:

- only one mounted shell owns the screen
- no duplicated recording-control code remains
- studio, logs, and DSL views all render from the same page

## Phase 3: Replace Demo State

Goal:

- move from demo-oriented slices to product-oriented slices

Main files:

- `ui/src/app/store.ts`
- `ui/src/features/studio-draft/studioDraftSlice.ts`
- new slices under `ui/src/features/`

Acceptance criteria:

- fake runtime metrics are removed
- real query and websocket state drives the UI
- UI-only state is clearly separated from server state

## Phase 4: Rebuild WebSocket Integration

Goal:

- explicitly decode server events and update the store from them

Main files:

- `ui/src/features/session/wsClient.ts`
- new typed event helpers under `ui/src/api/` or `ui/src/features/transport/`

Acceptance criteria:

- `session.state`, `session.log`, `preview.list`, and `preview.state` all update the store correctly
- reconnect still works
- lifecycle ownership is documented and explicit

## Phase 5: Repair Preview Integration

Goal:

- ensure source cards request previews using the real backend lifecycle

Main files:

- `ui/src/api/previewsApi.ts`
- `ui/src/components/preview/PreviewStream.tsx`
- source-grid and source-card containers

Acceptance criteria:

- preview creation uses `{ dsl, source_id }`
- preview release uses `preview_id`
- visible preview panels reflect real preview descriptors

## Phase 6: Rewrite Mocks And Tooling Hygiene

Goal:

- make isolated development truthful again

Main files:

- `ui/src/mocks/handlers.ts`
- `ui/package.json`
- ESLint config files to add

Acceptance criteria:

- `pnpm --dir ui lint` passes
- MSW matches the real contract
- any generated output inclusion is intentional and documented

## Commit Boundaries

Suggested review-friendly commit boundaries:

1. `ui: align transport types and RTK query endpoints with backend`
2. `ui: mount StudioPage and remove duplicate shell logic`
3. `ui: replace demo session state with backend-driven state`
4. `ui: rebuild websocket and preview integration`
5. `ui: rewrite mocks and add frontend linting`

## Review Checklist

Use this checklist during implementation and review:

- Does this route exist in `internal/web/`?
- Does this payload shape match `api_types.go`?
- Is this data owned by the server or by the UI?
- Is this code path duplicated somewhere else?
- If this file is deprecated, can it be deleted now?
- Have we avoided compatibility aliases and adapters?

## Open Questions

- Should the frontend eventually derive structured form state from the DSL, or should DSL-first remain the main editing model for v1?
- Should preview acquisition be automatic for visible cards or explicit by user action in v1?
- Should transport types eventually be generated from a shared schema, or is manual mirroring sufficient for the project at this stage?

## References

- `ui/src/App.tsx`
- `ui/src/main.tsx`
- `ui/src/app/store.ts`
- `ui/src/components/studio/StudioApp.tsx`
- `ui/src/pages/StudioPage.tsx`
- `ui/src/api/types.ts`
- `ui/src/api/discoveryApi.ts`
- `ui/src/api/setupsApi.ts`
- `ui/src/api/recordingApi.ts`
- `ui/src/api/previewsApi.ts`
- `ui/src/features/session/wsClient.ts`
- `ui/src/mocks/handlers.ts`
- `internal/web/api_types.go`
- `internal/web/handlers_api.go`
- `internal/web/handlers_preview.go`
- `internal/web/handlers_ws.go`
- `ttmp/2026/04/09/SCS-0002--screencast-studio-web-control-frontend-architecture-and-implementation-plan/design-doc/02-frontend-assessment-and-improvement-guide.md`
