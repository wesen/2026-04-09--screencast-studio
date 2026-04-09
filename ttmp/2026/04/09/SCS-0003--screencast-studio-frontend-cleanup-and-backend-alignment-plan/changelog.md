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
