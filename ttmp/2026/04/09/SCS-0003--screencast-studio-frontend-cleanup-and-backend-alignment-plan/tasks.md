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

- [ ] Audit `ui/src/features/studio-draft/studioDraftSlice.ts` and `ui/src/features/session/sessionSlice.ts`.
- [ ] Decide which state is truly UI-owned and which state should come from backend queries or websocket events.
- [ ] Create or refactor slices around:
  - editor DSL text
  - discovery snapshot
  - current recording session
  - preview descriptors
  - websocket connectivity
  - UI-only tab and panel state
- [ ] Remove simulated elapsed time, mic level, and disk growth from the mounted shell.

Acceptance criteria:

- Real runtime state is not simulated locally.
- UI-only state is clearly separated from server-backed state.
- Deprecated demo-only state is deleted.

## Phase 5: Rebuild WebSocket Handling

- [ ] Define the typed server event model used by the frontend.
- [ ] Refactor `ui/src/features/session/wsClient.ts` so socket lifecycle and event decoding are explicit.
- [ ] Handle at least `session.state`, `session.log`, `preview.list`, and `preview.state`.
- [ ] Ensure one app-level owner starts and stops the websocket connection.
- [ ] Remove ambiguous singleton behavior if it no longer matches the final ownership model.

Acceptance criteria:

- WebSocket state updates flow through typed reducers or explicit actions.
- Reconnect behavior still works.
- Event handling is aligned with `internal/web/handlers_ws.go`.

## Phase 6: Repair Preview Integration

- [ ] Ensure preview creation uses `{ dsl, source_id }`.
- [ ] Ensure preview release uses `preview_id`.
- [ ] Make source-card or source-grid containers request previews through a shared preview state model.
- [ ] Keep `PreviewStream` presentational and focused on rendering the given stream state.
- [ ] Remove any preview code that assumes stale route names or source-id-only lifecycle semantics.

Acceptance criteria:

- Preview lifecycle matches the backend exactly.
- Preview panels reflect real preview descriptors and preview IDs.

## Phase 7: Rewrite The Mock Layer

- [x] Replace stale MSW routes in `ui/src/mocks/handlers.ts`.
- [x] Remove fake endpoints that the backend does not expose.
- [x] Update mock payloads so they mirror real backend envelopes.
- [ ] Re-test Storybook or local mock mode after transport alignment is complete.

Acceptance criteria:

- Mock handlers validate the real contract instead of masking drift.
- Storybook and mocked development remain useful after the cleanup.

## Phase 8: Add Frontend Hygiene

- [ ] Add an ESLint configuration under `ui/`.
- [ ] Make `pnpm --dir ui lint` pass.
- [ ] Decide whether `ui/dist/` and `ui/storybook-static/` should remain checked in.
- [ ] Add a short frontend validation checklist to the ticket diary or follow-on docs.

Acceptance criteria:

- Frontend linting is real, not only declared.
- Generated output handling is intentional.

## Phase 9: Delete The Dead Code

- [ ] Remove deprecated API helpers, stale types, and duplicate shell code once replacements are in place.
- [ ] Remove unused imports, selectors, reducers, stories, or helpers discovered during cleanup.
- [ ] Do a final repo scan for files and symbols that only supported the deleted paths.

Acceptance criteria:

- The frontend no longer contains obvious compatibility shims or dead legacy paths.
- The surviving architecture is smaller and easier to explain.

## Suggested Commit Boundaries

- [x] Commit 1: transport types and endpoint cleanup
- [x] Commit 2: mounted shell consolidation
- [ ] Commit 3: state model cleanup
- [ ] Commit 4: websocket and preview cleanup
- [ ] Commit 5: mocks, linting, and dead-code deletion

## Validation Checklist

- [x] `pnpm --dir ui build`
- [ ] `pnpm --dir ui lint`
- [x] `pnpm --dir ui build-storybook`
- [ ] manual smoke test against the real Go server
- [ ] review for remaining duplicate shells or stale routes
