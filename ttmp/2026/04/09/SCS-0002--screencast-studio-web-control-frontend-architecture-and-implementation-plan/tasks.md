# Tasks

## Done

- [x] Create the second ticket workspace for the web control frontend
- [x] Inspect the current CLI/runtime code, prototype HTTP UI, and imported JSX mock
- [x] Write the detailed web frontend architecture and implementation guide
- [x] Validate the ticket with `docmgr doctor --ticket SCS-0002 --stale-after 30`
- [x] Upload the ticket bundle to reMarkable and verify the remote listing

## Implementation Plan For Intern

This section is the concrete implementation sequence for the second ticket. The order matters. Do not start by building React screens in isolation. The browser depends on backend transport primitives, and those primitives depend on clear session and preview ownership.

The definition of success for this ticket is:

- a Go server process can expose the existing discovery, compile, and record capabilities over stable HTTP and WebSocket interfaces
- a React frontend can connect to that server, load system state, edit a setup draft, request previews, and start or stop a recording session
- the implementation uses the existing DSL and recording plan as the single source of truth rather than inventing a second frontend-only model
- the browser experience is good enough to replace the current prototype for day-to-day interactive control

## Phase 0: Read And Prepare

- [ ] Read the main web design doc from start to finish:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0002--screencast-studio-web-control-frontend-architecture-and-implementation-plan/design-doc/01-screencast-studio-web-control-frontend-system-design.md`
- [ ] Read the CLI-first backend design from the first ticket:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/design-doc/01-screencast-studio-system-design.md`
- [ ] Inspect the current code paths that the web stack must wrap:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/discovery/service.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/types.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/session.go`
- [ ] Inspect the existing prototype only to understand product intent, not as an implementation template:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/app.js`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/index.html`
- [ ] Write a short implementation note in the diary summarizing what is already reusable versus what must be rewritten for the web stack.

Acceptance criteria:

- you can explain, in your own words, the difference between raw DSL, compiled plan, runtime session, and preview session
- you know which package should remain domain logic and which package should become HTTP transport

## Phase 1: Add The Server Shell In Go

- [x] Create the backend web package layout described in the design doc.
- [x] Add an HTTP server entrypoint that lives alongside the current CLI instead of replacing it.
- [x] Add a `serve` CLI command using glazed command patterns.
- [x] Define server configuration fields for:
  - bind address
  - static asset mode for development versus embedded production assets
  - log level
  - preview limits
- [x] Wire the `serve` command into the existing application boundary rather than bypassing `pkg/app`.
- [x] Serve a minimal JSON health endpoint and a placeholder WebSocket endpoint before adding business logic.

Suggested file targets:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/cmd/screencast-studio/main.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/root.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/serve.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/routes.go`

Acceptance criteria:

- `go run ./cmd/screencast-studio serve` starts a server
- `GET /api/healthz` returns a simple success payload
- the command shuts down cleanly on Ctrl-C

## Phase 2: Expose Read-Only Discovery And Session APIs

- [x] Add JSON endpoints for discovery:
  - displays
  - windows, if supported by the current platform
  - cameras
  - audio devices
- [x] Add an endpoint for fetching current runtime session state, even if no session is running.
- [x] Normalize all HTTP responses into stable payload structs instead of returning internal structs directly.
- [x] Add API request and response types in a dedicated package or file set.
- [x] Add tests for handler success and failure paths.

Suggested endpoints:

- [ ] `GET /api/discovery`
- [ ] `GET /api/session`

Acceptance criteria:

- the browser can fetch discovery state with one initial request
- the payload shape is documented and covered by tests
- API structs do not leak transport-irrelevant internal fields

## Phase 3: Add DSL Draft, Normalize, And Compile APIs

- [x] Implement an endpoint that accepts a DSL document and returns normalized setup data.
- [x] Implement an endpoint that accepts a DSL document and returns compile output, warnings, and derived outputs.
- [x] Keep the canonical compile logic in the backend domain packages.
- [x] Avoid adding frontend-specific compilation shortcuts.
- [x] Include useful error locations and messages for invalid DSL.
- [ ] Add round-trip fixtures for valid and invalid setups.

Suggested endpoints:

- [ ] `POST /api/setup/normalize`
- [ ] `POST /api/setup/compile`

Acceptance criteria:

- the frontend can submit a draft setup and get back compile results without starting a recording
- invalid DSL errors are actionable enough to surface in the UI
- compile warnings are preserved, not dropped

## Phase 4: Add Recording Control APIs

- [x] Implement a backend session coordinator for the web server that owns one active recording session at a time for version 1.
- [x] Add `start` and `stop` endpoints that operate on the existing recording runtime.
- [ ] Map runtime state transitions into stable API responses and WebSocket events.
- [x] Ensure concurrent start attempts fail clearly.
- [x] Ensure stop requests are idempotent.
- [x] Add integration-style tests that start a fake or test-double session coordinator where appropriate.

Suggested endpoints:

- [ ] `POST /api/recordings/start`
- [ ] `POST /api/recordings/stop`
- [ ] `GET /api/recordings/current`

Acceptance criteria:

- starting a recording from HTTP triggers the same runtime primitives used by the CLI
- stopping a recording transitions through the formal runtime states cleanly
- repeated stop requests do not crash or deadlock

## Phase 5: Add WebSocket Event Transport

- [ ] Implement `/ws` for browser subscriptions to runtime and preview events.
- [ ] Define concrete event envelopes with `type`, `timestamp`, and `payload`.
- [ ] Publish session state changes to connected clients.
- [ ] Publish selected FFmpeg log output or summarized process output where useful for UI diagnostics.
- [ ] Publish preview lifecycle changes.
- [ ] Handle disconnects and reconnects without leaking goroutines.

Acceptance criteria:

- a connected browser receives state transitions during start, run, and stop
- WebSocket disconnects do not destabilize the active recording session
- the event contract is documented in the design doc or a follow-up API note

## Phase 6: Add Preview Lifecycle Management In Go

- [ ] Implement preview leasing or reference counting so browser tabs do not create unbounded preview workers.
- [ ] Add endpoints to ensure, list, and release previews.
- [ ] Start with MJPEG preview transport for version 1.
- [ ] Keep preview workers isolated from recording workers.
- [ ] Limit the number of simultaneous previews and surface errors clearly.
- [ ] Add metrics or logs that make preview lifecycle visible during debugging.

Suggested endpoints:

- [ ] `POST /api/previews/ensure`
- [ ] `POST /api/previews/release`
- [ ] `GET /api/previews`
- [ ] `GET /api/previews/:id/mjpeg`

Acceptance criteria:

- previews can be started and stopped intentionally
- closing the browser or releasing a preview eventually tears down preview resources
- preview failures are observable in logs and API/WebSocket output

## Phase 7: Scaffold The React Frontend

- [x] Create the `ui/` workspace with Vite, TypeScript, Redux Toolkit, RTK Query, and Bootstrap.
- [x] Configure local development so Vite proxies `/api` and `/ws` to the Go server.
- [x] Define the frontend folder layout from the design doc:
  - app shell
  - API layer
  - feature modules
  - shared components
  - styling tokens
- [x] Add one smoke page that loads discovery data from the live Go backend.
- [x] Add a production build target that the Go server can eventually embed.

Suggested file targets:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/package.json`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/vite.config.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/main.tsx`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/app/store.ts`

Acceptance criteria:

- `pnpm install` and `pnpm dev` work in `ui/`
- the frontend can fetch real backend data through the dev proxy
- the production build emits static assets without manual patching

## Phase 8: Build The Main Operator Screen

- [x] Build the source-card layout modeled on the imported JSX mock.
- [x] Show display, region, window, camera, and audio sources in clearly separated UI sections.
- [x] Show preview panes for active sources.
- [x] Show operator controls for:
  - arm or disarm source
  - choose input target
  - edit region parameters
  - choose recording format and destination template
  - start and stop recording
- [x] Show recording session status, elapsed time, and warnings.
- [x] Make the UI usable at laptop widths as well as large desktop screens.

Acceptance criteria:

- the operator can create a plausible recording setup without editing raw YAML
- the page still exposes raw DSL or compile output somewhere for debugging
- the screen remains comprehensible while a recording is active

## Phase 9: Connect The UI To Draft And Compile Flows

- [x] Define a frontend draft state that mirrors editable setup intent without pretending to be the final compiled plan.
- [x] Add conversion logic between UI draft state and DSL request payloads.
- [x] Recompile on explicit actions first, then consider auto-compile later.
- [x] Show compile warnings and errors inline.
- [x] Add loading, empty, and failure states for every async panel.

Acceptance criteria:

- UI edits can be compiled repeatedly without page reloads
- users can distinguish between draft data, compiled data, and active runtime state
- compile errors do not wipe out the current draft

## Phase 10: Connect The UI To Recording And Preview Flows

- [ ] Connect preview cards to preview ensure and release calls.
- [ ] Connect the record button to the recording start endpoint.
- [ ] Connect the stop button to the recording stop endpoint.
- [ ] Subscribe the UI to WebSocket updates for session state, preview state, and logs.
- [ ] Ensure the UI recovers cleanly if the WebSocket reconnects mid-session.

Acceptance criteria:

- a user can start a preview, see frames, then release it
- a user can start a recording, watch state updates, and stop it
- the UI does not show stale “recording” state after reconnect or refresh

## Phase 11: Production Packaging

- [ ] Add a production frontend build step to the Go build workflow.
- [ ] Embed built static assets with `go:embed`.
- [ ] Serve frontend assets under `/static/` and the SPA shell under `/`.
- [ ] Preserve API routes under `/api` and WebSocket under `/ws`.
- [ ] Document the development and production workflows for future contributors.

Acceptance criteria:

- one production Go binary serves the web app and API together
- local developer workflow remains two-process and easy to reason about
- asset embedding does not break `go test ./...` or `go build ./...`

## Phase 12: Testing, Hardening, And Review

- [ ] Add backend handler tests for all major endpoints.
- [ ] Add runtime tests around session state to cover web-triggered start and stop flows.
- [ ] Add frontend tests for the core state transitions and operator interactions.
- [ ] Run manual smoke tests with a real display capture and at least one preview stream.
- [ ] Record validation commands and outcomes in the diary.
- [ ] Update the design doc if implementation reality diverges from the original plan.

Validation checklist:

- [ ] `go test ./...`
- [ ] `go build ./...`
- [ ] `pnpm test`
- [ ] `pnpm build`
- [ ] manual browser smoke test against local backend
- [ ] `docmgr doctor --ticket SCS-0002 --stale-after 30`

## Guardrails

- [ ] Do not duplicate DSL compilation logic in the browser.
- [ ] Do not let HTTP handlers reach deep into FFmpeg process details directly; go through application and runtime boundaries.
- [ ] Do not add backwards-compatibility layers unless explicitly requested.
- [ ] Do not make previews and recordings share the same unmanaged process pool.
- [ ] Keep the diary updated after each meaningful implementation milestone.

## Suggested Milestone Commits

- [ ] Milestone 1: server shell, health endpoint, and basic discovery APIs
- [ ] Milestone 2: setup normalize/compile APIs and stable payload contracts
- [ ] Milestone 3: recording start/stop/current APIs and WebSocket events
- [ ] Milestone 4: preview lifecycle and MJPEG transport
- [ ] Milestone 5: React scaffold and live backend integration
- [ ] Milestone 6: main operator screen with draft, compile, preview, and recording controls
- [ ] Milestone 7: production asset embedding, tests, and documentation cleanup

- [ ] Validate the ticket with `docmgr doctor --ticket SCS-0002 --stale-after 30`
- [ ] Upload the ticket bundle to reMarkable and verify the remote listing
- [ ] Start implementation from the server shell and SPA scaffold phases
