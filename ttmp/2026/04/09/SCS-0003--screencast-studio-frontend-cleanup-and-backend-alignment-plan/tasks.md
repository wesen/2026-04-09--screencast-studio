# Tasks

## Goal

Clean up the existing frontend so it becomes a strict client of the current backend, while deleting stale, duplicated, unclear, or unnecessary frontend code. No compatibility layer should be added.

## Phase 1: Freeze The Real Contract

- [x] Read `internal/web/api_types.go`, `internal/web/handlers_api.go`, `internal/web/handlers_preview.go`, and `internal/web/handlers_ws.go` and write down the exact route and payload surface the frontend must follow.
- [x] Compare that surface against `ui/src/api/types.ts`, `ui/src/api/discoveryApi.ts`, `ui/src/api/setupsApi.ts`, `ui/src/api/recordingApi.ts`, and `ui/src/api/previewsApi.ts`.
- [x] Identify every stale route, stale field, stale envelope, and stale event assumption.
- [x] Record the replacement mapping from old frontend assumptions to the correct backend contract.

Acceptance criteria:

- There is a complete backend-to-frontend contract checklist.
- It is clear which frontend API files will be rewritten, renamed, or deleted.

## Phase 2: Replace The Transport Layer

- [x] Rewrite `ui/src/api/types.ts` so it mirrors backend payloads rather than older frontend assumptions.
- [x] Replace or rename `ui/src/api/setupsApi.ts` with `setupApi.ts` semantics that match `/api/setup/...`.
- [x] Fix discovery, recording, and preview endpoints to match the backend exactly.
- [x] Remove calls to routes that do not exist.
- [x] Update any consumers that rely on the old payload shapes.

Acceptance criteria:

- All frontend API files target real backend routes.
- All request and response shapes match backend envelopes.
- No compatibility aliases or fallback parsing paths remain.

## Phase 3: Collapse To One Mounted Shell

- [x] Decide whether `StudioPage.tsx` is the surviving shell. The default expectation is yes.
- [x] Move real orchestration into the surviving shell or a page-level container.
- [x] Change `ui/src/App.tsx` to mount the new shell.
- [x] Delete or demote `ui/src/components/studio/StudioApp.tsx` once the replacement is in place.
- [x] Remove duplicated handler logic and placeholder shell-specific state.

Acceptance criteria:

- One mounted shell owns the studio screen.
- No duplicate screen orchestration remains.
- App entrypoint ownership is obvious from `App.tsx`.

## Phase 4: Replace Demo State With Real State

- [x] Audit `ui/src/features/studio-draft/studioDraftSlice.ts` and `ui/src/features/session/sessionSlice.ts`.
- [x] Decide which state is truly UI-owned and which state should come from backend queries or websocket events.
- [x] Create or refactor slices around:
  - editor DSL text
  - discovery snapshot
  - current recording session
  - preview descriptors
  - websocket connectivity
  - UI-only tab and panel state
- [x] Remove simulated elapsed time, mic level, and disk growth from the mounted shell.

Acceptance criteria:

- Real runtime state is not simulated locally.
- UI-only state is clearly separated from server-backed state.
- Deprecated demo-only state is deleted.

## Phase 5: Introduce Shared Protobuf Transport Schemas

- [x] Add a new protobuf design note to the ticket and align the implementation order around it.
- [x] Create `proto/screencast/studio/v1/web.proto` for the shared REST and websocket contract.
- [x] Add `buf.yaml` and `buf.gen.yaml` using Buf v2 remote plugins.
- [x] Generate Go code under `gen/go/proto/...`.
- [x] Generate TypeScript code under `ui/src/gen/proto/...`.
- [x] Add any missing runtime dependencies for generated Go and TypeScript protobuf code.

Acceptance criteria:

- REST and websocket transport messages are defined once in `proto/`.
- Go and TypeScript generated outputs can be reproduced with a committed command.
- The repo no longer needs handwritten shared contract declarations as the long-term source of truth.

## Phase 6: Migrate Go REST And Websocket Transport To Generated Types

- [x] Add Go-side `protojson` encode and decode helpers for REST messages.
- [x] Replace shared REST request and response structs in `internal/web/api_types.go` with generated protobuf messages or mapping helpers built around them.
- [x] Migrate health, discovery, setup, recording, and preview REST handlers to generated protobuf transport messages.
- [x] Replace websocket bootstrap and fan-out events with generated protobuf `ServerEvent` messages.
- [x] Update Go tests to assert the new lowerCamelCase protobuf JSON contract.

Acceptance criteria:

- Go REST handlers read and write generated protobuf JSON messages.
- Websocket events are emitted from generated protobuf event messages.
- The backend transport shape is defined by protobuf plus explicit domain-to-proto mapping helpers, not handwritten parallel structs.

## Phase 7: Migrate Frontend Transport And Websocket Decode To Generated Types

- [x] Replace `ui/src/api/types.ts` as the source of truth for shared REST and websocket payloads.
- [x] Add small frontend protobuf JSON helpers for `fromJson(...)` and `toJson(...)`.
- [x] Update RTK Query endpoints in `ui/src/api/` to send and decode generated protobuf JSON.
- [x] Refactor `ui/src/features/session/wsClient.ts` so websocket lifecycle is explicit and event decoding uses the generated protobuf `ServerEvent` schema.
- [x] Ensure one app-level owner starts and stops the websocket connection.
- [x] Remove handwritten websocket schema guards and stale singleton assumptions if they are no longer needed.

Acceptance criteria:

- Frontend REST and websocket transport logic uses generated protobuf schemas.
- Websocket event handling is runtime-decoded from the shared schema rather than inferred from raw JSON.
- One clear frontend owner controls websocket lifecycle.

## Phase 8: Repair Preview Integration On Top Of The Shared Schema

- [x] Ensure preview creation uses the protobuf-generated `{ dsl, sourceId }` request shape.
- [x] Ensure preview release uses the protobuf-generated `previewId` request shape.
- [x] Add or refactor preview state so source-card or source-grid containers lease previews through one shared state model.
- [x] Keep `PreviewStream` presentational and focused on rendering the supplied preview state.
- [x] Remove any preview code that assumes stale route names, stale field names, or source-id-only lifecycle shortcuts.

Acceptance criteria:

- Preview lifecycle matches the backend exactly through the generated protobuf transport contract.
- Preview panels reflect real preview descriptors and preview IDs.

## Phase 9: Rewrite The Mock Layer

- [x] Replace stale MSW routes in `ui/src/mocks/handlers.ts`.
- [x] Remove fake endpoints that the backend does not expose.
- [x] Update mock payloads so they mirror real backend envelopes.
- [x] Re-test Storybook or local mock mode after transport alignment is complete.

Acceptance criteria:

- Mock handlers validate the real contract instead of masking drift.
- Storybook and mocked development remain useful after the cleanup.

## Phase 10: Add Frontend Hygiene

- [x] Add an ESLint configuration under `ui/`.
- [x] Make `pnpm --dir ui lint` pass.
- [x] Decide whether `ui/dist/` and `ui/storybook-static/` should remain checked in.
- [x] Add a short frontend validation checklist to the ticket diary or follow-on docs.

Acceptance criteria:

- Frontend linting is real, not only declared.
- Generated output handling is intentional.

## Phase 11: Delete The Dead Code

- [x] Remove deprecated API helpers, stale types, and duplicate shell code once replacements are in place.
- [x] Remove unused imports, selectors, reducers, stories, or helpers discovered during cleanup.
- [x] Do a final repo scan for files and symbols that only supported the deleted paths.

Acceptance criteria:

- The frontend no longer contains obvious compatibility shims or dead legacy paths.
- The surviving architecture is smaller and easier to explain.

## Suggested Commit Boundaries

- [x] Commit 1: transport types and endpoint cleanup
- [x] Commit 2: mounted shell consolidation
- [x] Commit 3: state model cleanup
- [ ] Commit 4: protobuf schema and generation plumbing
- [ ] Commit 5: Go REST and websocket protobuf migration
- [ ] Commit 6: frontend protobuf transport and websocket decode migration
- [x] Commit 7: preview integration cleanup on top of protobuf
- [ ] Commit 8: final dead-code deletion and hygiene cleanup

## Validation Checklist

- [x] `pnpm --dir ui build`
- [x] `pnpm --dir ui lint`
- [x] `pnpm --dir ui build-storybook`
- [x] `go test ./...`
- [x] `go build ./...`
- [x] manual smoke test against the real Go server
- [x] review for remaining duplicate shells or stale routes
