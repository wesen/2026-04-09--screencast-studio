---
Title: Screencast Studio Web Control Frontend System Design
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
    - Path: cmd/screencast-studio/main.go
      Note: Existing binary entrypoint that currently only exposes the CLI tree
    - Path: jank-prototype/main.go
      Note: |-
        Prototype HTTP server and preview/capture handlers being replaced
        Prototype HTTP server and preview/capture baseline
    - Path: jank-prototype/web/app.js
      Note: |-
        Prototype polling-based frontend logic
        Prototype browser logic and polling baseline
    - Path: jank-prototype/web/index.html
      Note: |-
        Prototype HTML structure and UI scope
        Prototype page structure baseline
    - Path: pkg/app/application.go
      Note: |-
        Current domain boundary for discovery, compile, and record operations
        Current application boundary the web transport should wrap
    - Path: pkg/cli/root.go
      Note: Existing command tree that the eventual server command should integrate with
    - Path: pkg/discovery/service.go
      Note: Machine discovery surface the web API should wrap
    - Path: pkg/dsl/types.go
      Note: |-
        Setup DSL and compiled-plan data structures the web UI should edit and review
        Current setup and compiled-plan schema boundary
    - Path: pkg/recording/run.go
      Note: |-
        Stateful recording runtime the web transport must observe, not reimplement
        Current stateful runtime to observe through web transport
    - Path: ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/design-doc/01-screencast-studio-system-design.md
      Note: CLI-first architecture ticket that established the current backend contracts
    - Path: ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/sources/local/screencast-studio-v2.jsx.jsx
      Note: |-
        Imported visual target for the eventual control surface
        Imported UI target for the web control surface
ExternalSources: []
Summary: Detailed design and implementation guide for the second-ticket web control frontend, including API contracts, frontend state design, preview transport, Go embedding strategy, and phased implementation guidance for an intern.
LastUpdated: 2026-04-09T15:07:00-04:00
WhatFor: Exhaustive architecture and implementation guide for the browser-based control surface that will sit on top of the CLI-first screencast studio backend.
WhenToUse: Use when building the web control frontend, adding an HTTP/WebSocket transport layer to the current backend, or onboarding a new engineer to the second implementation phase.
---


# Screencast Studio Web Control Frontend System Design

## Executive Summary

This ticket defines the second major phase of the screencast studio project: a local web control surface that sits on top of the now-working CLI-first backend. The first ticket established the core system boundaries: machine discovery, setup DSL normalization, compile-time plan generation, and a stateful recording runtime. This ticket should not reopen those core decisions. Its job is to add a browser-facing transport and UI that use those existing backend contracts cleanly.

The most important principle for this phase is that the web app is a client of the backend domain layer, not a second implementation of recorder logic. The browser should never assemble FFmpeg arguments, infer platform details, or decide session-state transitions. The browser should manage user intent, issue explicit API requests, render compile results, show live previews, and subscribe to runtime events. The Go backend should remain the sole authority on discovery, plan compilation, preview lifecycle, and recording session state.

For version 1 of the web ticket, the recommended architecture is:

- React + TypeScript + Vite for the SPA.
- Redux Toolkit plus RTK Query for networked state and cache orchestration.
- Bootstrap plus project-local CSS variables for layout and the imported retro UI language.
- Go server command that exposes `/api`, `/ws`, and `/static` while serving the SPA at `/`.
- `go generate` plus `go:embed` for a production single-binary build.
- Request/response APIs for discovery, setup load/save/compile, and recording commands.
- WebSocket events for runtime session state, logs, meters, and preview lifecycle notifications.
- MJPEG previews for source cards in version 1, with a backend-managed preview process lifecycle.

If an intern follows the plan in this document, they should be able to produce a browser UI that feels close to the imported JSX mock while preserving the architecture discipline established in the CLI-first milestone.

## Problem Statement

The current repository has two extremes and no good middle ground.

The prototype server in [jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go) already demonstrates a tiny browser control loop. It exposes `/api/example`, `/api/config/apply`, `/api/config`, `/api/state`, `/api/capture/start`, `/api/capture/stop`, and `/api/preview/...` as raw HTTP handlers ([jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L185)). Its frontend in [jank-prototype/web/app.js](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/app.js) simply polls state, mutates a textarea, and repaints preview cards. That proved the product idea, but it is too coupled, too ad hoc, and too fragile to serve as the long-term frontend architecture.

The current production direction, on the other hand, is a CLI-first backend with no web transport at all. The current application surface in [pkg/app/application.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go) exposes clean methods for discovery, setup compilation, and recording execution. The setup DSL and compiled plan are explicit in [pkg/dsl/types.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/types.go). The recording runtime is explicit and stateful in [pkg/recording/run.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go). That backend is much healthier than the prototype, but there is no browser-facing contract yet.

The user-visible gap is obvious:

- the system already knows how to discover displays, windows, cameras, and audio devices;
- the system already knows how to validate a setup and compile a concrete plan;
- the system already knows how to execute a recording session;
- but there is no structured web API, no browser state model, no preview transport policy, and no modern UI implementation.

The imported UI mock in [screencast-studio-v2.jsx.jsx](/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/sources/local/screencast-studio-v2.jsx.jsx) makes the intended product shape very clear. The user should manipulate a studio-like surface with source cards, source selection, per-source status, output parameters, mic controls, and global record controls. That product model deserves a real frontend architecture rather than a polling demo.

## Scope

### In scope

- Define the web control surface architecture on top of the existing backend.
- Define the Go HTTP/WebSocket transport needed to expose discovery, setup, compile, previews, and recording state.
- Define the React/TypeScript frontend architecture and package layout.
- Define state ownership between browser, transport layer, domain layer, and runtime.
- Define version 1 preview strategy, event stream strategy, and deployment strategy.
- Provide file-level implementation guidance for an intern.

### Explicitly out of scope

- Replacing the current DSL or recorder runtime architecture.
- Reintroducing frontend-owned FFmpeg logic.
- Multi-user remote collaboration.
- Timeline editing, postproduction, or compositing.
- Cross-platform browser guarantees beyond the current Linux/X11 target.
- Public cloud deployment.

## Current-State Analysis

### What exists today

The project currently has three strong building blocks for the web phase.

First, the CLI application boundary already exists in [pkg/app/application.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go). That file exposes:

- `DiscoveryList(...)` for machine inventory;
- `CompileFile(...)` for plan validation and compile review;
- `RecordFile(...)` for recording execution and final session summary.

Second, the domain model already exists. The DSL types in [pkg/dsl/types.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/types.go) distinguish between user-authored config and runtime-executable plan artifacts. That means the browser can safely edit setup intent without leaking runtime state into saved files.

Third, the recording runtime already owns explicit session behavior. The stateful session runtime and its explicit state machine live in [pkg/recording/session.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/session.go) and [pkg/recording/run.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go). The web ticket should reuse that lifecycle rather than rebuild it in the browser.

### What the prototype web layer teaches us

The prototype still provides useful evidence about what the browser needs to do. In [jank-prototype/web/index.html](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/index.html), the UI had five visible concerns:

- setup editing;
- recording status;
- preview cards;
- warnings;
- logs.

In [jank-prototype/web/app.js](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/app.js), the browser:

- fetched a sample setup from `/api/example`;
- posted raw DSL text to `/api/config/apply`;
- polled `/api/state`;
- started and stopped capture through simple POST endpoints;
- rendered image previews by repeatedly changing `<img src="/api/preview/...?...">`.

That is exactly the shape we do not want to repeat. The problems are:

- polling instead of event subscription;
- textarea-centric editing instead of a structured studio view model;
- preview lifecycle coupled to page refresh behavior;
- one big page script instead of isolated components and state slices.

### What the imported JSX mock adds

The imported JSX mock from the first ticket is the actual product target, not the prototype HTML. It introduces three important UI concepts the prototype never modeled well:

1. Source cards with per-source controls and local status.
2. A global output-parameters surface that is not just “the YAML text area.”
3. Dedicated microphone and status panels that belong to runtime observation, not static config entry.

Those three concepts imply three frontend subsystems:

- a studio-editor model;
- a transport/cache layer;
- a session-observation model.

Any frontend design that does not explicitly model those subsystems will regress back toward the prototype’s monolithic `app.js`.

## Design Goals

1. Keep the browser as a client of the existing domain layer instead of a second runtime.
2. Make the frontend implementation modular enough that an intern can work on one surface area at a time.
3. Keep the development loop fast: Vite for frontend development, Go API on a separate port, proxy in dev, embed in production.
4. Keep the production packaging simple: one Go binary serving the SPA and the API.
5. Make runtime state observable through events rather than through aggressive polling.
6. Keep preview transport simple for version 1, but avoid request-scoped process lifecycles.
7. Preserve the distinctive retro studio visual direction from the imported mock while still using Bootstrap for layout and form primitives.

## Proposed Solution

### High-level architecture

The web system should consist of five layers:

```text
React SPA (ui/)
  |-- studio editor
  |-- source cards
  |-- output panel
  |-- status/mic/log panels
  v
Transport layer
  |-- RTK Query for REST data
  |-- WebSocket client for session/runtime events
  v
Go web API
  |-- HTTP handlers under /api
  |-- WebSocket endpoint under /ws
  |-- preview streaming endpoints under /api/previews/*
  v
Application/domain layer
  |-- pkg/app
  |-- pkg/dsl
  |-- pkg/discovery
  |-- pkg/recording
  v
Platform/FFmpeg adapters
```

### Development and production topology

The recommended topology follows the `go-web-frontend-embed` pattern:

```text
Development:
  Vite dev server      -> localhost:3000
  Go API/WebSocket     -> localhost:3001
  Vite proxy           -> /api and /ws forwarded to Go

Production:
  one Go binary
    -> serves SPA at /
    -> serves API at /api
    -> serves WebSocket at /ws
    -> serves static assets under /static
```

This is the right choice because it optimizes the inner loop without complicating deployment. The frontend gets HMR in development, while the production build remains a single binary with embedded assets.

### Recommended repository layout

The repo should gain these new areas:

```text
ui/
  package.json
  pnpm-lock.yaml
  vite.config.ts
  tsconfig.json
  src/
    main.tsx
    app/
    pages/
    components/
    features/
    api/
    store/
    styles/

internal/web/
  embed.go
  embed_none.go
  generate.go
  generate_build.go
  spa.go
  server.go
  api/
    routes.go
    handlers_*.go
  ws/
    hub.go
    messages.go

cmd/screencast-studio/
  main.go
```

Why `internal/web/` and `ui/`?

- `ui/` is the natural place for the frontend toolchain.
- `internal/web/` keeps serving, embedding, generation, and HTTP-specific backend code together.
- `pkg/app`, `pkg/dsl`, `pkg/discovery`, and `pkg/recording` stay transport-agnostic.

### Frontend state ownership

One of the most important architectural decisions is to separate three kinds of state.

#### 1. Server-owned truth

This includes:

- discovery inventory;
- compile results;
- active recording session state;
- runtime logs;
- preview availability;
- output manifests.

This state should come from REST or WebSocket data. The browser should cache it, not invent it.

#### 2. Browser-owned draft state

This includes:

- which source card is selected in the UI;
- unsaved edits to a draft studio setup;
- whether the user is in “raw YAML” mode or “form” mode;
- panel visibility, temporary dialogs, and filters.

This belongs in Redux slices or local component state depending on scope.

#### 3. Derived presentation state

This includes:

- compile diff banners like “draft differs from last compiled setup”;
- whether the Record button should be enabled;
- whether a source card should show “preview unavailable”;
- whether the UI should show a reconnect banner for `/ws`.

This should be computed from server state plus draft state, not stored redundantly.

## Frontend Architecture

### Stack

The frontend should use:

- React
- TypeScript
- pnpm
- Redux Toolkit
- RTK Query
- Bootstrap

This is the preferred stack from the repository instructions, and it is appropriate here because:

- RTK Query will simplify API caching and loading/error state;
- Bootstrap gives a stable layout baseline;
- custom CSS variables can preserve the imported mock’s visual identity.

### Visual direction

The imported JSX mock should be treated as a visual source of truth, not just inspiration. That means:

- warm off-white/cream background tones;
- black and muted gray linework;
- condensed retro mono/window chrome styling;
- visible “panel” metaphors;
- deliberate recording state colors such as red/amber/green;
- non-generic source cards with preview frames.

Bootstrap should provide spacing, grid, buttons, and forms, but the final UI should not look like stock Bootstrap. Define CSS variables in one place, then style Bootstrap primitives and custom components around them.

Example token sketch:

```css
:root {
  --studio-bg: #e8e4dc;
  --studio-surface: #f5f0e8;
  --studio-ink: #1a1a1a;
  --studio-muted: #8a8a7a;
  --studio-line: #2c2c2c;
  --studio-red: #c04040;
  --studio-green: #5a8a5a;
  --studio-amber: #b89840;
  --studio-font-mono: "Chicago", "Geneva", "Monaco", monospace;
}
```

### Suggested frontend module breakdown

```text
ui/src/
  main.tsx
  App.tsx
  app/
    router.tsx
    layout.tsx
  store/
    store.ts
    hooks.ts
  api/
    baseApi.ts
    discoveryApi.ts
    setupsApi.ts
    recordingApi.ts
    previewsApi.ts
  features/studio-draft/
    studioDraftSlice.ts
    selectors.ts
    adapters.ts
  features/session/
    sessionSlice.ts
    wsClient.ts
  components/source-card/
    SourceCard.tsx
    SourcePreview.tsx
    SourceToolbar.tsx
  components/output-panel/
    OutputPanel.tsx
  components/mic-panel/
    MicPanel.tsx
  components/status-panel/
    StatusPanel.tsx
  components/log-panel/
    LogPanel.tsx
  components/dsl-editor/
    DSLEditor.tsx
  styles/
    tokens.css
    studio.css
```

### Redux and RTK Query responsibilities

Use RTK Query for server-driven resources:

- discovery catalog
- compiled plan review
- saved setup fetch/save
- preview metadata
- current session snapshot

Use classic Redux slices for:

- studio draft editing
- WebSocket-fed session event accumulation
- UI preferences and panel state

Suggested slices:

```text
discoveryApi    -> GET /api/discovery/*
setupsApi       -> load/save/compile setup drafts
recordingApi    -> start/stop recording commands
sessionSlice    -> WebSocket-fed runtime events and current session status
studioDraftSlice-> editable browser-side studio draft
uiSlice         -> dialogs, selected source, active tab, layout preferences
```

## Backend Web Transport Design

### Core rule

The web transport must call the existing application/domain layer instead of bypassing it. HTTP handlers should be thin wrappers around `pkg/app` and new preview/session services.

That means:

- `pkg/app` remains the entrypoint for discovery, compile, and record operations;
- new HTTP handlers should adapt request bodies into domain calls;
- new WebSocket logic should subscribe to domain/runtime events;
- the browser should never call `pkg/recording` concepts directly.

### Proposed REST API

Version 1 should expose explicit, typed endpoints instead of mirroring the prototype exactly.

#### Discovery

```text
GET /api/discovery
GET /api/discovery/displays
GET /api/discovery/windows
GET /api/discovery/cameras
GET /api/discovery/audio-inputs
POST /api/discovery/refresh
```

Example response:

```json
{
  "generated_at": "2026-04-09T15:00:00Z",
  "displays": [{ "id": "display-1", "name": "Display 1", "width": 1920, "height": 1080 }],
  "windows": [{ "id": "0x3a00007", "title": "Browser Window", "x": 10, "y": 20, "width": 1200, "height": 800 }],
  "cameras": [{ "id": "cam-1", "device": "/dev/video0", "label": "Built-in Camera" }],
  "audio_inputs": [{ "id": "audio-1", "name": "Built-in Mic", "driver": "PipeWire" }]
}
```

#### Setup draft and compile review

```text
GET  /api/setups/example
POST /api/setups/normalize
POST /api/setups/compile
POST /api/setups/validate
```

Request shape for compile:

```json
{
  "dsl_format": "yaml",
  "dsl": "schema: recorder.config/v1\n..."
}
```

Response shape:

```json
{
  "session_id": "demo",
  "warnings": [],
  "outputs": [
    { "kind": "video", "source_id": "desktop", "name": "Full Desktop", "path": "recordings/demo/Full Desktop.mov" }
  ],
  "video_jobs": 4,
  "audio_jobs": 1
}
```

#### Recording

```text
POST /api/recordings/start
POST /api/recordings/stop
GET  /api/recordings/current
GET  /api/recordings/history   (optional later)
```

Start request:

```json
{
  "dsl_format": "yaml",
  "dsl": "schema: recorder.config/v1\n...",
  "max_duration_seconds": 0
}
```

Current-session response:

```json
{
  "active": true,
  "session_id": "demo",
  "state": "running",
  "reason": "all workers started",
  "started_at": "2026-04-09T15:00:00Z",
  "outputs": [...],
  "warnings": [...]
}
```

#### Preview endpoints

```text
POST /api/previews/ensure
POST /api/previews/release
GET  /api/previews/:sourceID.mjpeg
GET  /api/previews
```

The browser should not create preview processes implicitly by changing an `<img>` URL. The browser should first declare which preview streams it needs, and the backend should maintain them per source.

### Proposed WebSocket API

Use one WebSocket endpoint:

```text
GET /ws
```

It should carry server-to-client events such as:

```json
{ "type": "session.state", "payload": { "session_id": "demo", "state": "running", "reason": "all workers started" } }
{ "type": "session.log", "payload": { "session_id": "demo", "level": "info", "line": "desktop stderr: ..." } }
{ "type": "session.output", "payload": { "session_id": "demo", "kind": "video", "path": "recordings/demo/Full Desktop.mov" } }
{ "type": "preview.state", "payload": { "source_id": "desktop", "state": "ready" } }
{ "type": "meter.audio", "payload": { "device_id": "default", "level": 0.42 } }
```

The browser can also send a tiny set of control messages later if needed, but version 1 should keep mutation traffic on REST and reserve WebSocket primarily for observation.

## Preview Strategy

### Version 1 preview policy

Use MJPEG for source-card previews in version 1.

Why this is acceptable:

- implementation complexity is low;
- browser support is trivial through `<img>` tags;
- it fits the retro control-surface product shape;
- it keeps the first web ticket focused on system architecture rather than media transport sophistication.

Why MJPEG should still be backend-managed:

- preview lifecycles should be source-centric, not request-centric;
- multiple browser repaints should not create duplicate FFmpeg processes;
- the server needs room for reference counting, throttling, and cleanup.

### Preview manager behavior

Pseudocode:

```go
type PreviewManager struct {
    previews map[string]*PreviewWorker
}

func (pm *PreviewManager) Ensure(sourceID string) (*PreviewDescriptor, error) {
    if worker, ok := pm.previews[sourceID]; ok {
        worker.RefCount++
        return worker.Descriptor(), nil
    }
    worker := startPreviewWorker(sourceID)
    pm.previews[sourceID] = worker
    return worker.Descriptor(), nil
}

func (pm *PreviewManager) Release(sourceID string) {
    worker := pm.previews[sourceID]
    worker.RefCount--
    if worker.RefCount <= 0 {
        worker.Stop()
        delete(pm.previews, sourceID)
    }
}
```

This manager should live in the Go transport/runtime boundary, not in the browser.

## Backend Implementation Plan

### Phase 1: Add the web server shell

Create the Go-side server package and SPA serving contract.

Files:

- `internal/web/server.go`
- `internal/web/spa.go`
- `internal/web/embed.go`
- `internal/web/embed_none.go`
- `internal/web/generate.go`
- `internal/web/generate_build.go`

Definition of done:

- `go run ./cmd/screencast-studio serve` starts a server.
- `/api/healthz` works.
- `/` serves the SPA in dev and prod modes.
- `/api` routes are not shadowed by static serving.
- `/static` is reserved for static assets.

### Phase 2: Add the frontend scaffold

Files:

- `ui/package.json`
- `ui/pnpm-lock.yaml`
- `ui/vite.config.ts`
- `ui/src/main.tsx`
- `ui/src/App.tsx`
- `ui/src/store/store.ts`
- `ui/src/api/baseApi.ts`
- `ui/src/styles/tokens.css`
- `ui/src/styles/studio.css`

Definition of done:

- `pnpm -C ui install`
- `pnpm -C ui dev`
- Vite proxies `/api` and `/ws` to the Go backend
- a basic studio shell renders

### Phase 3: Wrap the current backend with HTTP endpoints

Files:

- `internal/web/api/routes.go`
- `internal/web/api/discovery_handlers.go`
- `internal/web/api/setup_handlers.go`
- `internal/web/api/recording_handlers.go`
- `internal/web/api/types.go`

Definition of done:

- browser can fetch discovery data
- browser can submit DSL for compile review
- browser can start and stop a recording session through HTTP

### Phase 4: Add WebSocket session events

Files:

- `internal/web/ws/hub.go`
- `internal/web/ws/messages.go`
- `internal/web/ws/handler.go`
- small application/runtime hooks to publish session changes

Definition of done:

- browser receives recording state changes without polling
- browser receives live log lines
- browser reconnect behavior is acceptable

### Phase 5: Add preview management

Files:

- `internal/web/previews/manager.go`
- `internal/web/previews/worker.go`
- `internal/web/api/preview_handlers.go`

Definition of done:

- source cards can request and release previews
- preview worker lifetimes are not tied to single HTTP requests
- two cards watching the same source do not start two preview processes

### Phase 6: Build the real studio UI

Files:

- `ui/src/components/source-card/*`
- `ui/src/components/output-panel/*`
- `ui/src/components/mic-panel/*`
- `ui/src/components/status-panel/*`
- `ui/src/components/log-panel/*`
- `ui/src/features/studio-draft/*`
- `ui/src/features/session/*`

Definition of done:

- source cards show discovery-backed choices and preview state
- the output panel edits the draft configuration model
- the status panel reflects live runtime state
- the user can compile, review, and record from the browser

## Testing and Validation Strategy

### Backend tests

- HTTP handler tests for each `/api` endpoint
- WebSocket message contract tests
- preview manager tests for reference counting and cleanup
- integration test ensuring `/` serves HTML and `/api/*` remains separate

### Frontend tests

- slice tests for draft transformations
- component tests for source cards, output panel, and session panel
- API contract tests using mocked responses
- browser integration tests for compile, start, and stop flows

### Manual validation checklist

1. Start Go API on `:3001`.
2. Start Vite on `:3000`.
3. Open the studio UI.
4. Confirm discovery data renders.
5. Change a setup draft in the UI.
6. Run compile and verify outputs/warnings.
7. Start a recording.
8. Confirm session state and logs stream without polling.
9. Confirm previews appear and do not multiply FFmpeg workers.
10. Stop recording and verify final outputs.

## Design Decisions

### REST for commands, WebSocket for observation

This split is intentional:

- REST is clearer for discrete mutations such as compile, start, and stop.
- WebSocket is better for event streams such as session state, logs, meters, and preview availability.

This keeps browser logic simple and avoids turning every interaction into ad hoc socket messages.

### Browser edits draft DSL, not raw FFmpeg intent

The frontend should own a draft setup model that maps cleanly onto the existing DSL. The user may still want a raw YAML view, but the primary UX should be structured editing, not text editing.

That keeps the frontend aligned with the imported mock and avoids making the browser a string-manipulation shell around YAML.

### Keep `pkg/` domain packages transport-agnostic

Do not move HTTP or WebSocket concerns into `pkg/app`, `pkg/dsl`, or `pkg/recording`. New transport code belongs under `internal/web/`.

That makes the backend reusable from both CLI and web flows and preserves the clean architecture gained in the first ticket.

## Alternatives Considered

### Alternative 1: Reuse the prototype HTTP API almost as-is

Rejected because the prototype API mirrors its monolithic implementation. It would force the new browser UI to inherit polling, request-scoped previews, and raw DSL-first interactions.

### Alternative 2: Put all frontend logic inside one static page again

Rejected because the imported UI target already exceeds what a single handwritten `app.js` can maintain cleanly.

### Alternative 3: WebRTC previews in version 1

Rejected for now because the system needs a stable control surface first. MJPEG is simpler, cheaper to reason about, and easier for an intern to implement safely.

### Alternative 4: Expose the recorder runtime directly to the browser

Rejected because it would couple the browser to low-level runtime policy and undo the main architectural improvement of the first ticket.

## Risks and Open Questions

### Risks

- preview worker lifecycle can become a source of leaks if reference counting is sloppy;
- WebSocket reconnection logic can create subtle stale-state bugs if session snapshots are not replayable;
- editing a structured studio draft and a raw YAML view in the same UI can create synchronization complexity;
- the current runtime was designed for CLI-first use, so some event publication hooks may need small refactors.

### Open questions

- Should the first web iteration include YAML editing as a primary tab, or only as an advanced/debug view?
- Should meters be true live audio levels in version 1, or is a “device active / inactive” signal enough?
- Should preview quality and frame rate be configurable per source in the UI, or fixed by backend policy for version 1?

## Pseudocode Walkthrough

### Browser startup

```text
page loads
  -> fetch discovery snapshot
  -> fetch current recording session snapshot
  -> fetch example or last draft setup
  -> connect WebSocket
  -> render source cards, output panel, and status panel
```

### Compile flow

```text
user edits draft
  -> browser updates studioDraftSlice
  -> user clicks Compile
  -> POST /api/setups/compile with current draft
  -> backend normalizes + compiles
  -> browser renders warnings, outputs, and plan review
```

### Record flow

```text
user clicks Record
  -> POST /api/recordings/start with current draft
  -> backend compiles and starts stateful session
  -> runtime publishes events
  -> WebSocket pushes session.state and session.log events
  -> browser updates sessionSlice
  -> UI reflects recording progress and status
```

### Preview flow

```text
source card mounts
  -> browser ensures preview for source
  -> backend starts or reuses preview worker
  -> browser displays MJPEG stream URL
source card unmounts
  -> browser releases preview
  -> backend decrements refcount and stops worker if unused
```

## Intern-Facing File-Level Guidance

If you are the new engineer implementing this ticket, read files in this order:

1. [pkg/app/application.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go)
2. [pkg/dsl/types.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/types.go)
3. [pkg/recording/session.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/session.go)
4. [pkg/recording/run.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go)
5. [jank-prototype/web/app.js](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/app.js)
6. [screencast-studio-v2.jsx.jsx](/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/sources/local/screencast-studio-v2.jsx.jsx)

Then implement in this order:

1. server shell and embed strategy
2. frontend scaffold and Vite proxy
3. discovery + compile REST endpoints
4. record start/stop REST endpoints
5. WebSocket event transport
6. preview manager
7. source-card UI and output panel
8. status/log/mic panels

Do not start with CSS polish or preview performance tweaks. First make the transport boundaries correct.

## References

- [pkg/app/application.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go)
- [pkg/cli/root.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/root.go)
- [pkg/dsl/types.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/types.go)
- [pkg/recording/run.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go)
- [jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go)
- [jank-prototype/web/app.js](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/app.js)
- [jank-prototype/web/index.html](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/index.html)
- [SCS-0001 design doc](/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/design-doc/01-screencast-studio-system-design.md)
