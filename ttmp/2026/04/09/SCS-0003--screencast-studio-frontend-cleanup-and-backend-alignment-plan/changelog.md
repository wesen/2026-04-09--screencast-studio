# Changelog

## 2026-04-09

- Initial workspace created


## 2026-04-09

Created the dedicated frontend cleanup ticket and wrote the initial architecture and implementation guide for removing stale frontend integration paths, collapsing duplicate shells, and aligning the UI strictly with the backend contract without compatibility shims.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0003--screencast-studio-frontend-cleanup-and-backend-alignment-plan/design-doc/01-frontend-cleanup-and-backend-alignment-system-design.md — Main cleanup architecture and implementation guide
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0003--screencast-studio-frontend-cleanup-and-backend-alignment-plan/reference/01-diary.md — Diary documenting why this cleanup ticket was created and what evidence it builds on
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0003--screencast-studio-frontend-cleanup-and-backend-alignment-plan/tasks.md — Detailed phased task list for the cleanup implementation

## 2026-04-09

Completed the first frontend cleanup slice by replacing the stale transport contract, renaming the setup API layer, aligning discovery, recording, preview, and websocket payload types with the backend, and updating the mock layer and log view to compile against the real transport schema.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/types.ts — Replaced stale frontend transport definitions with backend-aligned types
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/setupApi.ts — New setup API layer using the real `/api/setup/*` routes
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/discoveryApi.ts — Removed unsupported discovery subroutes
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/recordingApi.ts — Recording API now uses session envelopes
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/previewsApi.ts — Preview API now uses backend preview envelopes and `preview_id` release semantics
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/session/wsClient.ts — Websocket client now decodes the current backend event payload shapes
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/mocks/handlers.ts — Mock handlers updated to stop reinforcing deprecated routes

## 2026-04-09

Completed the mounted-shell consolidation by making `StudioPage` the app entrypoint, moving the remaining shell-level orchestration there, deleting `StudioApp`, and updating Storybook to present `StudioPage` as the main top-level screen.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/App.tsx — App now mounts the surviving shell directly
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx — StudioPage now owns the top-level shell orchestration
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/Introduction.mdx — Storybook introduction now points at StudioPage

## 2026-04-09

Started the deeper state cleanup by moving DSL text and page tab state into explicit slices, seeding session state from the real current-session query, deriving elapsed time from backend timestamps, and removing the fake shell-level elapsed and disk-growth timers.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/editor/editorSlice.ts — New editor-owned DSL state
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/studio-ui/studioUiSlice.ts — New UI-owned tab state
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/app/store.ts — Store now includes the new page-level state slices
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx — Page now uses backend-seeded session state and explicit slices instead of local timers
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/OutputPanel.tsx — Pause control can now be explicitly disabled when unsupported

## 2026-04-09

Wired the page’s compile and recording controls to the real backend mutations so the raw DSL editor and transport controls no longer use placeholder behavior.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx — Page now compiles DSL and starts or stops recording through RTK Query mutations
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/editor/editorSlice.ts — Editor slice now tracks compile warnings, compile errors, and compile loading state
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/OutputPanel.tsx — Transport controls now support a busy state during start and stop transitions

## 2026-04-09

Added a real ESLint configuration for the frontend workspace and verified that both lint and build now pass under `ui/`.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/.eslintrc.cjs — New ESLint configuration for the frontend workspace
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/.eslintignore — Ignore rules for generated and dependency directories
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/package.json — Existing lint script is now backed by a real config

## 2026-04-09

Replaced the mounted page’s synthetic source grid with backend-normalized setup data, added a dedicated setup slice for normalize results, decoupled source-card components from `studioDraftSlice`, and revalidated lint, build, and Storybook against the shared source-card type.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx — Page now derives visible sources from normalized DSL data instead of the demo source list
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup/setupSlice.ts — New slice for normalized config, warnings, and normalization errors
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/source-card/types.ts — Shared source-card type used by the page, grid, status panel, and stories
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/SourceGrid.tsx — Grid now accepts the shared source-card model and supports read-only rendering
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/StatusPanel.tsx — Status panel no longer depends on `studioDraftSlice` source types
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/SourceCard.stories.tsx — Storybook updated for the shared source-card model
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/SourceGrid.stories.tsx — Storybook updated for the read-only normalized source-grid mode
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/StatusPanel.stories.tsx — Storybook updated for the shared source-card model

## 2026-04-09

Added a dedicated protobuf migration design note inside `SCS-0003`, reordered the remaining implementation phases around shared schema generation for REST and websocket transport, and explicitly pivoted away from continuing the handwritten transport cleanup path.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0003--screencast-studio-frontend-cleanup-and-backend-alignment-plan/design-doc/02-protobuf-transport-migration-plan.md — New detailed design and implementation guide for the protobuf migration
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0003--screencast-studio-frontend-cleanup-and-backend-alignment-plan/tasks.md — Task plan now prioritizes protobuf schema, generation, and REST plus websocket migration before further preview cleanup
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0003--screencast-studio-frontend-cleanup-and-backend-alignment-plan/reference/01-diary.md — Diary records why the cleanup plan pivoted away from more handwritten transport work

## 2026-04-09

Implemented the shared protobuf transport contract for both REST and websocket traffic, generated Go and TypeScript code with Buf, removed the old handwritten Go transport file, and switched the frontend API and websocket decode paths to generated protobuf schemas.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/proto/screencast/studio/v1/web.proto — Shared transport schema for REST and websocket messages
- /home/manuel/code/wesen/2026-04-09--screencast-studio/buf.yaml — Buf module configuration for schema generation
- /home/manuel/code/wesen/2026-04-09--screencast-studio/buf.gen.yaml — Buf code generation configuration for Go and TypeScript outputs
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/pb_mapping.go — Explicit domain-to-protobuf mapping helpers for the web boundary
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/protojson.go — Shared `protojson` encode and decode helpers for REST and websocket messages
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go — Websocket stream now emits generated `ServerEvent` messages
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/proto.ts — Frontend protobuf JSON helper boundary
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/session/wsClient.ts — Websocket decode now uses the generated `ServerEvent` schema
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/gen/proto/screencast/studio/v1/web_pb.ts — Generated TypeScript transport types and schemas

## 2026-04-09

Implemented the first real preview-integration slice on top of the protobuf transport by adding explicit owned preview leases in the frontend state model, ensuring and releasing previews from `StudioPage`, and rendering source cards through the presentational `PreviewStream` component instead of preview-adjacent demo placeholders.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx — Page now owns preview leasing and release against the real preview API
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/previews/previewSlice.ts — Preview slice now tracks server descriptors plus client-owned preview leases by source
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/preview/PreviewStream.tsx — Preview stream now renders supplied preview state and stream URLs without inventing lifecycle rules
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/source-card/SourceCard.tsx — Source cards now render the real preview stream component
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/discoveryApi.ts — Added health query usage so preview leasing can respect the backend preview limit
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go — Preview validation messages no longer refer to stale snake_case field names

## 2026-04-09

Finished the cleanup and hygiene pass by deleting the remaining demo-only source state from `studioDraftSlice`, removing the mounted page’s fake mic and disk telemetry, fixing the default DSL so the real page boots into a valid backend shape, and documenting a successful live smoke test against the Go server including preview ensure and release behavior.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/studio-draft/studioDraftSlice.ts — Removed unused source-list and fake meter state, leaving only UI-owned output and microphone controls
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/editor/editorSlice.ts — Default DSL now matches the backend’s normalized config shape
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/MicPanel.tsx — Microphone panel now renders “unavailable” instead of invented live meters
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/StatusPanel.tsx — Status panel now renders “unavailable” instead of invented disk telemetry
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/DSLEditor.stories.tsx — Storybook DSL story now uses the same valid default DSL as the mounted app
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0003--screencast-studio-frontend-cleanup-and-backend-alignment-plan/reference/01-diary.md — Diary now records the final validation checklist and live smoke results
