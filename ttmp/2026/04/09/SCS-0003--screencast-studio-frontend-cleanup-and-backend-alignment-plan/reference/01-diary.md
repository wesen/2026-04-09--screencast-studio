---
Title: Diary
Ticket: SCS-0003
Status: active
Topics:
    - frontend
    - backend
    - ui
    - architecture
    - dsl
    - video
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ui/src/App.tsx
      Note: Existing entrypoint inspected to define the cleanup target shell
    - Path: ui/src/components/studio/StudioApp.tsx
      Note: Existing mounted shell inspected as cleanup and deletion candidate
    - Path: ui/src/pages/StudioPage.tsx
      Note: Existing page composition inspected as the likely long-term shell
    - Path: ui/src/api/types.ts
      Note: Existing transport types inspected as stale frontend contract
    - Path: ui/src/mocks/handlers.ts
      Note: Existing mock handlers inspected as stale routes that hide drift
    - Path: internal/web/api_types.go
      Note: Backend transport definitions used as the canonical cleanup target
    - Path: ttmp/2026/04/09/SCS-0002--screencast-studio-web-control-frontend-architecture-and-implementation-plan/design-doc/02-frontend-assessment-and-improvement-guide.md
      Note: Prior frontend review used as the direct precursor to this cleanup ticket
ExternalSources: []
Summary: Chronological record of creating the dedicated frontend cleanup ticket and defining a no-compatibility plan to realign the UI with the backend.
LastUpdated: 2026-04-09T18:05:00-04:00
WhatFor: Preserve the reasoning, evidence, and document-creation steps behind the dedicated frontend cleanup ticket.
WhenToUse: Read when continuing the cleanup ticket, reviewing why the ticket exists, or checking which existing frontend files were considered the primary cleanup targets.
---

# Diary

## Goal

Record why the dedicated cleanup ticket was created, what existing evidence it builds on, and how the cleanup architecture and task plan were assembled.

## Step 1: Create A Dedicated Frontend Cleanup Ticket

This step split the frontend cleanup effort into its own ticket instead of continuing to overload the original web frontend ticket. The reason for doing this separately is that the work is no longer “build the web frontend.” The work is now “clean up the existing frontend implementation, delete stale and duplicate paths, and align the result with the backend we already have.”

The most important decision in this step was to treat backwards compatibility as unnecessary. The cleanup plan assumes that the codebase is still young enough that preserving stale route names, duplicate shells, or old payload shapes would only increase confusion. The right move is to replace the incorrect paths and delete them.

### Prompt Context

**User prompt (verbatim):** "Create a new ticket to cleanup the frontend and to properly align it with the backend, and remove any unclear / deprecated / duplicated / unnecessary code. No backwards compatibiliity needed.

Add a detailed architecture and implementation guide"

**Assistant interpretation:** Create a new docmgr ticket specifically for frontend cleanup, produce a detailed architecture and implementation guide for the cleanup, and frame the work around removing duplicate and stale code instead of preserving it.

### What I did

- Checked the current docmgr status and ticket list.
- Confirmed that `SCS-0001` and `SCS-0002` already exist and that the next appropriate ticket ID is `SCS-0003`.
- Created ticket `SCS-0003` with topics focused on frontend, backend, UI, architecture, DSL, and video.
- Added the primary design document.
- Added the diary document.

### Why

- The cleanup work is substantial enough that it deserves its own implementation plan, review checklist, and progress tracking.
- The prior ticket already documents what is wrong with the frontend. The new ticket documents how to fix it.
- Splitting this work out gives the team a clean place to track deletion-oriented cleanup without mixing it into earlier architecture exploration.

### What worked

- The earlier frontend assessment in `SCS-0002` already supplied enough evidence to define a clear cleanup ticket.
- The current backend surface in `internal/web/` is explicit enough to act as a stable transport target for the frontend cleanup.

### What didn't work

- The newly created ticket documents started from generic templates and needed full replacement, not incremental edits.

### What I learned

- The cleanup ticket needs a different tone than the original implementation ticket. It must be much more explicit about deletion, replacement, and refusal to support deprecated paths.
- The most useful form of guidance for the intern is not just “fix the API.” It is “keep the presentational layer, replace the integration layer, and delete the stale code when the new path lands.”

### What warrants a second pair of eyes

- Whether `StudioPage.tsx` should definitely be the one surviving shell, or whether some parts of `StudioApp.tsx` should be extracted differently before deletion.
- Whether generated frontend outputs currently in the repo are intentionally committed or should be removed as part of the cleanup.

### Code review instructions

- Start with the new design guide:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0003--screencast-studio-frontend-cleanup-and-backend-alignment-plan/design-doc/01-frontend-cleanup-and-backend-alignment-system-design.md`
- Then compare the cleanup guidance against:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/App.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/StudioApp.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/types.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/mocks/handlers.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/api_types.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0002--screencast-studio-web-control-frontend-architecture-and-implementation-plan/design-doc/02-frontend-assessment-and-improvement-guide.md`

### Technical details

Commands used in this step:

```bash
docmgr status --summary-only
docmgr ticket list
docmgr ticket create-ticket --ticket SCS-0003 --title "Screencast Studio Frontend Cleanup and Backend Alignment Plan" --topics frontend,backend,ui,architecture,dsl,video
docmgr doc add --ticket SCS-0003 --doc-type design-doc --title "Frontend Cleanup and Backend Alignment System Design"
docmgr doc add --ticket SCS-0003 --doc-type reference --title "Diary"
```

## Step 2: Replace The Stale Frontend Transport Contract

This step implemented the first actual cleanup slice in the codebase. The focus was the transport layer, not the shell. The goal was to make the frontend API definitions, transport types, and mock payloads reflect the backend that now exists, while avoiding any compatibility layer for old route names or old payload shapes.

The biggest concrete changes in this step were:

- replacing the stale transport schema in `ui/src/api/types.ts`
- deleting `ui/src/api/setupsApi.ts` and replacing it with `ui/src/api/setupApi.ts`
- removing discovery routes that the backend does not expose
- changing recording and preview APIs to use the backend envelope types
- updating the WebSocket client and session slice enough to compile against the new event payloads
- rewriting the MSW transport layer so it no longer encodes obviously deprecated routes

This step did not yet solve the duplicate shell problem. It was intentionally limited to the contract layer so the cleanup could land in a small, reviewable commit.

### Prompt Context

**User prompt (verbatim):** "ok, now work on it task by task, committing at appropriate intervals, and keeping a detailed diary as you work"

**Assistant interpretation:** Begin executing the cleanup ticket in small slices, validate each slice, document the work in the ticket diary, and commit at sensible boundaries.

### What I did

- Inspected the current frontend API files and compared them to:
  - `internal/web/api_types.go`
  - `internal/web/handlers_api.go`
  - `internal/web/handlers_preview.go`
  - `internal/web/handlers_ws.go`
- Replaced the old TypeScript transport schema with a backend-mirrored schema in `ui/src/api/types.ts`.
- Deleted `ui/src/api/setupsApi.ts`.
- Added `ui/src/api/setupApi.ts` using `/api/setup/normalize` and `/api/setup/compile`.
- Removed stale discovery routes and stale setup routes from the RTK Query layer.
- Updated `ui/src/api/recordingApi.ts` and `ui/src/api/previewsApi.ts` to use backend envelopes.
- Updated `ui/src/features/session/sessionSlice.ts` and `ui/src/features/session/wsClient.ts` so they compile against the real websocket payload types.
- Updated `ui/src/mocks/data.ts` and `ui/src/mocks/handlers.ts` so the mock layer no longer reinforces the worst transport drift.
- Updated `LogPanel` and its story to use real process-log shapes rather than a frontend-invented log type.
- Made the minimal compile-preserving adjustments in `StudioApp.tsx` and `StudioPage.tsx` after removing stale session helpers.
- Ran `pnpm --dir ui build`.

### Why

- The transport contract was the highest-risk source of drift.
- The RTK Query hooks were mostly unused, which made this a safe first cleanup slice.
- Fixing the API layer first reduces ambiguity for every later shell and state cleanup.
- Keeping this as a small slice makes the later shell consolidation easier to review.

### What worked

- The backend contract was explicit enough that the transport rewrite was straightforward.
- The frontend hooks were still lightly used, so there were relatively few consumer updates needed.
- The UI build passed after the transport and story updates.

### What didn't work

- Changing the transport and session types exposed stale `LogPanel` story imports immediately.
- The first build failed because `LogPanel` and its story still referenced the old `LogEntry` type from the session slice.

### Exact failure

```text
src/components/log-panel/LogPanel.tsx(3,15): error TS2305: Module '"@/features/session/sessionSlice"' has no exported member 'LogEntry'.
src/stories/LogPanel.stories.tsx(3,15): error TS2305: Module '"../features/session/sessionSlice"' has no exported member 'LogEntry'.
```

After fixing that, the next build failed on a syntax mistake introduced while converting the story helper:

```text
src/stories/LogPanel.stories.tsx(18,1): error TS1005: ',' expected.
```

Both were corrected and the build passed afterward.

### What I learned

- The transport layer was drifting in more places than just the API files. The mock layer and story data were carrying old assumptions too.
- The frontend’s current structure does allow contract cleanup in isolation, which is a good sign for the next cleanup slices.
- The shell cleanup should happen next, because transport is now much less ambiguous.

### What warrants a second pair of eyes

- Whether the current level of mock cleanup is enough for this slice, or whether Storybook should be explicitly revalidated before the next commit.
- Whether `sessionSlice.ts` should keep its `handleWsEvent` helper at all, given that the WebSocket client currently dispatches directly.

### What should be done in the future

- Collapse the app to a single mounted shell.
- Remove the remaining shell-level simulated runtime behavior.
- Continue replacing demo state with backend-driven state.
- Revisit the mock layer again once the shell and preview integrations are real.

### Code review instructions

- Review the transport changes first:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/types.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/discoveryApi.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/setupApi.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/recordingApi.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/previewsApi.ts`
- Compare them against:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/api_types.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go`
- Then review the compile-preserving consumer updates:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/session/sessionSlice.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/session/wsClient.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/mocks/handlers.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/log-panel/LogPanel.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/LogPanel.stories.tsx`

### Technical details

Commands used in this step:

```bash
git status --short
sed -n '1,240p' ui/src/api/types.ts
sed -n '1,240p' ui/src/api/discoveryApi.ts
sed -n '1,260p' ui/src/api/setupsApi.ts
sed -n '1,260p' ui/src/api/recordingApi.ts
sed -n '1,260p' ui/src/api/previewsApi.ts
sed -n '1,240p' ui/src/api/baseApi.ts
sed -n '1,260p' ui/src/features/session/sessionSlice.ts
sed -n '1,220p' ui/src/features/session/wsClient.ts
sed -n '1,260p' ui/src/mocks/data.ts
sed -n '1,260p' ui/src/mocks/handlers.ts
pnpm --dir ui build
```

## Step 7: Replace The Shared REST And Websocket Contract With Protobuf

This step pivoted the cleanup effort away from more handwritten transport maintenance and moved the web boundary to a single protobuf-defined schema shared by Go and TypeScript. The key idea was to keep the transport as JSON while replacing the duplicated contract declarations with generated code.

The most important architectural choice in this step was to keep protobuf at the web boundary instead of dragging generated messages into the domain packages. `pkg/discovery`, `pkg/dsl`, and the recording packages remain domain-shaped. `internal/web/` now owns the explicit mapping into generated transport messages.

### Prompt Context

**User prompt (verbatim):** "Alright, add a new implementation plan / analysis for switching to protobuf, within this ticket, and then tasks, and then tackle REST + WS protobuf use before we go on much further. That should simplify our code quite a bit."

**Assistant interpretation:** Add a dedicated protobuf migration design note to `SCS-0003`, update the task plan to prioritize the migration, and then implement the shared-schema move now instead of continuing with more handwritten transport cleanup.

### What I did

- Added the protobuf migration design guide:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0003--screencast-studio-frontend-cleanup-and-backend-alignment-plan/design-doc/02-protobuf-transport-migration-plan.md`
- Updated `tasks.md` so protobuf schema, generation, and REST plus websocket migration became the next implementation phases.
- Added protobuf generation plumbing:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/buf.yaml`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/buf.gen.yaml`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/proto/screencast/studio/v1/web.proto`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/generate.go`
- Generated and committed code for both languages:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/gen/go/proto/screencast/studio/v1/web.pb.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/gen/proto/screencast/studio/v1/web_pb.ts`
- Added runtime dependencies:
  - `google.golang.org/protobuf` in `go.mod` / `go.sum`
  - `@bufbuild/protobuf` in `ui/package.json` / `ui/pnpm-lock.yaml`
- Replaced the old handwritten Go transport layer:
  - deleted `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/api_types.go`
  - added `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/api_errors.go`
  - added `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/pb_mapping.go`
  - added `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/protojson.go`
- Migrated backend handlers and tests to protobuf JSON:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server_test.go`
- Migrated the frontend transport boundary to generated schemas:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/proto.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/discoveryApi.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/setupApi.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/recordingApi.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/previewsApi.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/session/wsClient.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/previews/previewSlice.ts`
- Reworked mocks and stories to emit protobuf-shaped JSON instead of stale handwritten payloads.

### Why

- The frontend cleanup had reached the point where transport duplication was the main remaining source of confusion.
- Websocket runtime validation was about to become a second schema language in the UI.
- The app is still small enough that moving to a single schema now is cheaper than maintaining one more round of duplicated types.

### What worked

- Keeping JSON on the wire made the migration straightforward to test in the browser-oriented code paths.
- `protojson` plus explicit mapping helpers gave the Go side a clear transport boundary.
- `@bufbuild/protobuf` gave the frontend a small, predictable runtime decode path for both REST and websocket messages.
- The websocket event contract became much clearer once `ServerEvent` used a protobuf `oneof` instead of string-dispatch plus free-form payloads.

### What didn't work

- The initial UI compile surfaced a few remaining assumptions about the old local transport types.
- A few places still treated generated messages like arbitrary local interfaces and needed to move to `create(...)`, `fromJson(...)`, or updated camelCase field names.

### What I learned

- The best dividing line is “protobuf only at the transport boundary.” Pulling protobuf into the domain layer would have made the code harder to explain, not easier.
- Generated TypeScript messages are usable in Redux and RTK Query as long as the code is disciplined about where decode and initialization happen.
- The schema-first websocket path is much easier to review than local type guards spread through the UI.

### Validation

The following commands passed after the migration:

```bash
buf generate
go test ./...
go build ./...
pnpm --dir ui build
pnpm --dir ui lint
pnpm --dir ui build-storybook
```

### What warrants a second pair of eyes

- Whether `ui/src/api/types.ts` should stay as a thin re-export layer or be deleted entirely once call sites import generated types directly.
- Whether the internal websocket event hub should stay string-based internally or switch to a typed Go event wrapper now that the transport schema is formalized.
- Whether `ui/dist/` and `ui/storybook-static/` should remain committed now that generated client code and schema outputs are in the repo as well.

### What should be done in the future

- Do a real manual smoke test against the running Go server to confirm preview and recording flows under the new protobuf contract.
- Finish the preview-integration cleanup on top of the new schema instead of adding any more handwritten transport logic.
- Remove any remaining convenience layers that no longer provide enough value after the generated-code switch.

### Code review instructions

- Start with the schema:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/proto/screencast/studio/v1/web.proto`
- Then review the Go web boundary:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/pb_mapping.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/protojson.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go`
- Then review the frontend boundary:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/proto.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/recordingApi.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/previewsApi.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/session/wsClient.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/session/sessionSlice.ts`
- Finally review the generated outputs only for shape sanity:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/gen/go/proto/screencast/studio/v1/web.pb.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/gen/proto/screencast/studio/v1/web_pb.ts`

### Technical details

Commands used in this step:

```bash
buf generate
go test ./internal/web
go test ./...
go build ./...
pnpm --dir ui add @bufbuild/protobuf
pnpm --dir ui build
pnpm --dir ui lint
pnpm --dir ui build-storybook
```

## Step 8: Put Real Preview Leasing On Top Of The Shared Schema

This step finally connected the source grid to the real preview lifecycle instead of leaving preview support as an API layer plus isolated components with no clear owner. The goal here was not to build a perfect long-term preview manager in the frontend. The goal was to establish one explicit ownership path for preview ensure and release behavior that respects the backend contract and makes the current UI honest.

The main architectural decision was to keep preview ownership in `StudioPage`, because the page is the current owner of normalized setup state, the active studio tab, and the mounted websocket connection. That means the page already knows when sources are visible and when the studio view is active. The preview component itself stays presentational.

### Prompt Context

**User prompt (verbatim):** "ok, continue"

**Assistant interpretation:** Continue executing `SCS-0003` task-by-task after the protobuf migration, starting with the next uncompleted phase rather than branching into unrelated frontend work.

### What I did

- Added a `getHealth` query to:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/discoveryApi.ts`
- Refactored the preview slice so it tracks both:
  - preview descriptors by preview ID
  - client-owned preview leases by source ID
- Added owned-preview reducers and selectors in:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/previews/previewSlice.ts`
- Updated `StudioPage` so it:
  - computes the desired visible preview sources
  - respects the backend preview limit
  - ensures previews using `{ dsl, sourceId }`
  - releases previews using `previewId`
  - logs ensure and release failures into the existing process-log panel
  - releases owned previews on unmount
- Updated the source-card model to carry preview state and preview URLs:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/source-card/types.ts`
- Switched source cards from the fake preview placeholder to the presentational preview component:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/source-card/SourceCard.tsx`
- Made `PreviewStream` display the supplied preview state more clearly instead of only showing a generic unavailable message:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/preview/PreviewStream.tsx`
- Cleaned up stale backend error text that still referred to snake_case preview field names:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`

### Why

- The transport and websocket layers were already correct, but the mounted studio screen still was not using the real preview lifecycle.
- Preview ownership was otherwise about to become ambiguous again, split between page code, source-card code, and free-form local state.
- The user explicitly wants the frontend aligned to the backend, and preview lifecycle is one of the most concrete places where “aligned” has to mean “real ensure and release semantics.”

### What worked

- The protobuf transport made the page-level preview code much simpler than the old handwritten-transport version would have been.
- A dedicated preview slice with explicit owned-lease tracking gave the page a stable place to store “which previews did this UI instance ensure?”
- Keeping `PreviewStream` presentational avoided pushing network or lifecycle behavior down into the card component tree.

### What didn't work

- The first version of the unmount cleanup used a dependency-bound effect that would have attempted cleanup on dependency changes, not only on unmount.
- That needed to be corrected by storing the latest owned-preview map in a ref and using a dedicated unmount cleanup effect.

### What I learned

- Preview descriptors from the server and preview leases owned by this frontend are related, but they are not the same state. Treating them separately makes the code easier to reason about.
- The correct UI rule is “render the preview if the page currently owns a lease for that source,” not “render any descriptor in the store that happens to share a source ID.”
- The server-side preview limit belongs in the frontend behavior too. Hard-coding a UI assumption there would just reintroduce drift.

### Validation

The following commands passed after this step:

```bash
pnpm --dir ui build
pnpm --dir ui lint
go test ./internal/web ./ui/...
```

Note: `go test ./ui/...` does not match any Go packages, so the meaningful validation here is `go test ./internal/web` plus the frontend build and lint commands.

### What warrants a second pair of eyes

- Whether the page should keep previews alive when the user switches away from the Studio tab, or whether releasing immediately on tab switch is the correct resource policy for v1.
- Whether a future manual smoke test shows that limiting previews to the first visible sources is the right product behavior when more sources exist than the backend preview limit allows.

### What should be done in the future

- Run a real browser-backed smoke test against the Go server to confirm preview start, frame arrival, and release behavior through the actual MJPEG stream.
- Continue with the remaining hygiene and dead-code cleanup now that preview lifecycle is no longer a major missing piece.

### Code review instructions

- Review the page-level preview ownership first:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`
- Then review the preview state model:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/previews/previewSlice.ts`
- Then review the rendering boundary:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/source-card/SourceCard.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/preview/PreviewStream.tsx`
- Finally review the small backend cleanup:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`

### Technical details

Commands used in this step:

```bash
rg -n "previewId|preview_id|source_id|sourceId|getWsClient|handleWsEvent|setupsApi|StudioApp" ui internal/web
pnpm --dir ui build
pnpm --dir ui lint
go test ./internal/web ./ui/...
```

## Step 3: Collapse The App To One Mounted Shell

This step removed the most obvious frontend duplication by making `StudioPage` the single mounted shell and deleting `StudioApp`. The purpose of this slice was not to finish the runtime integration. The purpose was to establish one place where future state and transport cleanup can happen.

Before this step:

- `App.tsx` mounted `StudioApp`
- `StudioPage.tsx` existed in parallel with overlapping layout ideas
- `StudioApp` owned websocket startup and some screen orchestration
- Storybook exposed both shells

After this step:

- `App.tsx` mounts `StudioPage`
- `StudioPage` owns the top-level shell orchestration
- `StudioApp.tsx` is deleted
- `StudioApp.stories.tsx` is deleted
- Storybook documentation now refers to `StudioPage` as the main screen

This matters because later state cleanup becomes much easier once there is only one surviving shell.

### What I did

- Deleted:
  - `ui/src/components/studio/StudioApp.tsx`
  - `ui/src/stories/StudioApp.stories.tsx`
- Updated `ui/src/App.tsx` to mount `StudioPage`.
- Moved the shell orchestration that still existed in `StudioApp` into `StudioPage`, including:
  - websocket connect and disconnect ownership
  - source-grid handlers
  - output-panel settings handlers
  - mic-panel settings handlers
  - local elapsed and disk simulation that still remains for now
- Updated the Storybook introduction so it references `StudioPage` instead of `StudioApp`.
- Ran:
  - `pnpm --dir ui build`
  - `pnpm --dir ui build-storybook`

### Why

- There should be exactly one mounted shell for the studio screen.
- Keeping both shells alive would make every later cleanup step more expensive.
- The page-level shell is the right place to centralize UI orchestration before deeper state cleanup.

### What worked

- `StudioPage` already had the right overall layout structure, so it was the natural surviving shell.
- The migration was mostly a matter of moving handlers and ownership, not redesigning components.
- Both the app build and the Storybook build passed after the consolidation.

### What didn't work

- Nothing failed hard in this slice, but Storybook still reports its existing large-chunk warnings and `eval` warnings from Storybook internals. Those are not regressions introduced by this slice.

### What I learned

- The shell duplication was making the codebase look more complex than it really is.
- Once `StudioPage` became the mounted shell, the next cleanup steps became much easier to reason about.
- There is still simulated session-adjacent behavior left in the surviving shell, which is the next thing to remove.

### What warrants a second pair of eyes

- Whether the remaining local elapsed and disk simulation should be removed immediately in the next slice or only as part of the larger state-model cleanup.
- Whether the websocket client should keep its singleton wrapper now that ownership is clearly page-level.

### What should be done in the future

- Remove the remaining shell-level simulated runtime state.
- Replace the shell’s current demo-oriented control behavior with backend-driven state.
- Continue simplifying the store so the page is easier to reason about.

### Code review instructions

- Review:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/App.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/Introduction.mdx`
- Confirm deletion of:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/StudioApp.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/StudioApp.stories.tsx`

### Technical details

Commands used in this step:

```bash
rg -n "StudioApp|StudioPage" ui/src ui/.storybook ui/src/stories
sed -n '1,240p' ui/src/App.tsx
sed -n '1,220p' ui/src/pages/StudioPage.tsx
sed -n '1,220p' ui/src/stories/Introduction.mdx
pnpm --dir ui build
pnpm --dir ui build-storybook
```

## Step 4: Move Page State Into Explicit Slices And Remove Fake Runtime Timers

This step began the deeper state cleanup. The main goal was to stop treating the page shell as a small self-contained app with its own fake runtime timers and instead make the page consume more explicit state.

The biggest changes in this step were:

- creating an editor slice for DSL text
- creating a studio UI slice for active tab state
- wiring `StudioPage` to read those slices instead of owning that state locally
- seeding the session slice from the real `/api/recordings/current` query
- replacing the fake elapsed timer with a value derived from `session.started_at` and `session.finished_at`
- removing the fake disk growth interval and the page-local pause state

This is not the full state cleanup yet. `studioDraftSlice.ts` still exists and still drives the source-card model. But the page is now materially less demo-driven than it was before this step.

### What I did

- Added:
  - `ui/src/features/editor/editorSlice.ts`
  - `ui/src/features/studio-ui/studioUiSlice.ts`
- Updated `ui/src/app/store.ts` to include the new reducers.
- Updated `ui/src/pages/StudioPage.tsx` to:
  - use Redux-backed DSL text
  - use Redux-backed active-tab state
  - query the current recording session via RTK Query
  - dispatch `setSession(...)` when current session data arrives
  - derive elapsed time from backend timestamps instead of incrementing a fake counter
  - stop simulating disk usage growth and local pause behavior
- Updated `ui/src/components/studio/OutputPanel.tsx` to support a disabled pause control when the backend does not support pause.
- Updated `ui/src/stories/StudioPage.stories.tsx` so Storybook provides the new reducers.
- Ran:
  - `pnpm --dir ui build`
  - `pnpm --dir ui build-storybook`

### Why

- `dslText` and active tab are genuinely UI-owned state and should live in explicit frontend state, not as incidental local variables in the page.
- `elapsed` is derived from real session timestamps, so the page should compute it from backend data rather than inventing it locally.
- The backend does not currently support pause, so keeping a fake page-level pause toggle was misleading.
- This cleanup sets up the next slice, where more of the remaining demo state can be replaced with backend-driven behavior.

### What worked

- The new slices fit naturally into the existing store.
- Moving DSL text and active-tab state out of the component reduced the page’s local ambiguity.
- Deriving elapsed from session timestamps removed a fake behavior without requiring a large component rewrite.
- Both the app build and Storybook build remained green after the change.

### What didn't work

- Storybook still emits the same pre-existing large-chunk and `eval` warnings from its build pipeline. Those warnings are not caused by this slice, but they remain visible in validation output.

### What I learned

- The easiest demo behaviors to remove are the ones that can be replaced by derived values, not by whole new backend APIs.
- `StudioPage` is now a better place to continue cleanup because its local state is shrinking.
- The next major state cleanup question is what to do with `studioDraftSlice.ts`, which still mixes real settings controls with a synthetic source model.

### What warrants a second pair of eyes

- Whether `sessionSlice.ts` should own the current-session query result more directly instead of being updated from the page.
- Whether `studioDraftSlice.ts` should be split into a UI-only settings slice and a source-model slice, or retired more aggressively once discovery-backed sources land.

### What should be done in the future

- Continue replacing the synthetic source model with backend-aligned source and preview state.
- Wire real start/stop recording actions through the page now that the transport layer and page shell are in better shape.
- Revisit the WebSocket client singleton and preview ownership after the recording controls are real.

### Code review instructions

- Review the new slices:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/editor/editorSlice.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/studio-ui/studioUiSlice.ts`
- Review the page/store integration:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/app/store.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/OutputPanel.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/StudioPage.stories.tsx`

### Technical details

Commands used in this step:

```bash
sed -n '1,220p' ui/src/pages/StudioPage.tsx
sed -n '1,220p' ui/src/app/store.ts
sed -n '1,260p' ui/src/stories/OutputPanel.stories.tsx
pnpm --dir ui build
pnpm --dir ui build-storybook
```

## Step 5: Wire Real Compile And Recording Controls

This step replaced the remaining no-op control behavior for the raw DSL compile button and the record button. The page now uses the backend mutations instead of pretending those actions exist.

The important idea in this step was to keep the scope narrow:

- compile the current DSL via the real compile endpoint
- start and stop recording via the real recording endpoints
- surface compile warnings and compile errors in the editor
- surface recording-start and recording-stop failures in the log panel

This slice does not yet solve preview acquisition or the synthetic source model. It only makes the existing top-level controls stop lying.

### What I did

- Extended `ui/src/features/editor/editorSlice.ts` to track:
  - compile warnings
  - compile errors
  - compile in-flight state
- Updated `ui/src/pages/StudioPage.tsx` to use:
  - `useCompileSetupMutation`
  - `useStartRecordingMutation`
  - `useStopRecordingMutation`
- Changed the record button handler so it:
  - starts recording with the current DSL when idle
  - stops the current session when active
  - dispatches backend session envelopes back into the store
- Changed the compile button handler so it:
  - calls the real compile endpoint
  - records warnings on success
  - records error messages on failure
- Updated `ui/src/components/studio/OutputPanel.tsx` to support a `transportBusy` flag so the record button can be disabled while a transition is in flight.
- Kept the page error path simple by pushing backend mutation errors into the session log view as `ui/stderr` entries.
- Ran:
  - `pnpm --dir ui build`
  - `pnpm --dir ui build-storybook`

### Why

- The existing buttons still represented fake functionality.
- Backend-aligned transport without real control wiring still leaves the page misleading.
- This slice improves correctness without requiring the preview or source-model cleanup to land first.

### What worked

- The compile and recording mutations integrated cleanly with the current page structure.
- Editor warnings and errors are now represented explicitly instead of being mixed with recording-session warnings.
- The UI build and Storybook build both remained green after the control integration.

### What didn't work

- Nothing failed structurally in this slice.
- Storybook still emits its existing `eval` and chunk-size warnings, which remain unrelated to this feature work.

### What I learned

- The page is now at a much better point for manual smoke testing because its main controls actually exercise backend endpoints.
- The next major architectural blocker is no longer “buttons are fake.” It is “the source and preview model is still synthetic.”

### What warrants a second pair of eyes

- Whether compile warnings should stay in the editor slice or move into a more general setup slice later.
- Whether recording-mutation errors should stay in the log panel or also be shown more visibly in the transport panel.

### What should be done in the future

- Replace the synthetic source model with backend-aligned source and preview state.
- Wire preview lifecycle into the page and source cards.
- Add a manual smoke test against the real Go server now that compile and record controls are live.

### Code review instructions

- Review:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/editor/editorSlice.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/OutputPanel.tsx`
- Verify these calls against:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/setupApi.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/recordingApi.ts`

### Technical details

Commands used in this step:

```bash
sed -n '1,260p' ui/src/components/dsl-editor/DSLEditor.tsx
sed -n '1,260p' ui/src/features/session/sessionSlice.ts
sed -n '1,240p' ui/src/components/studio/OutputPanel.tsx
pnpm --dir ui build
pnpm --dir ui build-storybook
```

## Step 6: Add Real Frontend Linting

This step fixed the frontend hygiene gap where `ui/package.json` declared a lint command but the workspace had no ESLint configuration at all.

The goal was not to build a perfect lint regime. The goal was to make the workspace honest: if `pnpm --dir ui lint` exists, it should work and enforce a reasonable baseline for the rest of the cleanup.

### What I did

- Confirmed the existing lint failure was caused by a missing ESLint config rather than rule violations.
- Added:
  - `ui/.eslintrc.cjs`
  - `ui/.eslintignore`
- Configured ESLint with:
  - `eslint:recommended`
  - `plugin:@typescript-eslint/recommended`
  - `plugin:react-hooks/recommended`
  - `react-refresh/only-export-components`
- Ignored generated directories:
  - `dist`
  - `storybook-static`
  - `node_modules`
- Ran:
  - `pnpm --dir ui lint`
  - `pnpm --dir ui build`

### Why

- The workspace already claimed to support linting.
- Without linting, the cleanup work can drift quickly and reviewers lose a low-cost validation step.
- Adding linting before the later preview and source-model cleanup helps prevent avoidable regressions.

### What worked

- The frontend passed lint immediately once the baseline config existed.
- No rule changes or source fixes were needed beyond the configuration itself.
- The build stayed green after adding the lint setup.

### What didn't work

- Nothing failed beyond the original missing-config issue.

### Exact original failure

```text
ESLint couldn't find a configuration file.
```

### What I learned

- The main frontend hygiene problem at this point was missing infrastructure, not a large volume of actual lint violations.
- The repo can now use lint as a real gate for later cleanup slices.

### What warrants a second pair of eyes

- Whether we want a stricter lint profile later, especially around `console.log`, unused exports, or import ordering.
- Whether generated build output under `ui/dist` and `ui/storybook-static` should be removed from version control in a later cleanup slice.

### What should be done in the future

- Decide how generated frontend output should be handled in source control.
- Consider adding frontend build and lint commands to CI once the main cleanup stabilizes.

### Code review instructions

- Review:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/.eslintrc.cjs`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/.eslintignore`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/package.json`

### Technical details

Commands used in this step:

```bash
pnpm --dir ui lint
pnpm --dir ui build
```

## Step 7: Replace The Synthetic Source Grid With Normalized Setup Data

This step removed the most visible remaining fake runtime model from the mounted page. Before this step, `StudioPage` still rendered the source grid from `studioDraftSlice`, which meant the main capture surface was showing synthetic source IDs and demo-only source state even though compile and record actions were already using the real DSL.

The goal of this slice was not to finish preview integration. The goal was to make the page read source cards from the backend-shaped normalized DSL so that later preview work can target real `source_id` values instead of client-invented numeric IDs.

### What I did

- Added `ui/src/features/setup/setupSlice.ts` to hold:
  - the current normalized config
  - normalization warnings
  - normalization errors
  - normalization loading state
- Registered the setup slice in `ui/src/app/store.ts`.
- Refactored `StudioPage.tsx` so it:
  - calls `useNormalizeSetupMutation`
  - debounces normalization after DSL changes
  - stores normalize results in the setup slice
  - derives source cards from `normalizedConfig.video_sources`
  - merges normalize warnings and compile warnings for the raw DSL editor
  - merges normalize errors and compile errors for the raw DSL editor
- Introduced a shared `StudioSource` type under `ui/src/components/source-card/types.ts`.
- Updated `SourceGrid`, `SourceCard`, `FakeScreen`, and `StatusPanel` to use the shared source-card type instead of depending on `studioDraftSlice` types.
- Made the mounted source grid read-only for now by passing `editable={false}` from `StudioPage`.
- Updated the source-card, source-grid, and status-panel stories to use the new shared source type and to cover the read-only normalized mode.
- Ran:
  - `pnpm --dir ui lint`
  - `pnpm --dir ui build`
  - `pnpm --dir ui build-storybook`

### Why

- The source grid was still the largest fake-data path in the mounted shell.
- Preview APIs require real `source_id` values, so leaving the grid on demo IDs would have made the next step more confusing.
- Moving the page to normalized setup data makes the frontend architecture much easier to explain: the editor owns text, the setup slice owns normalized DSL state, and the page renders source cards from that normalized state.

### What worked

- The existing backend normalize endpoint already returns the exact source fields the page needs: `id`, `name`, `type`, and `enabled`.
- The shared `StudioSource` type made it straightforward to decouple presentational components from `studioDraftSlice`.
- All frontend validation commands passed after updating the remaining story type drift.

### What didn't work

- The first `pnpm --dir ui build` failed because `StatusPanel.stories.tsx` was still typed against the old `studioDraftSlice` `Source` shape.

### Exact failure

```text
src/stories/StatusPanel.stories.tsx(28,20): error TS2741: Property 'sourceId' is missing in type 'Source' but required in type 'StudioSource'.
```

After updating that story to use the shared source-card type, build passed again.

### What I learned

- The mounted screen is now materially closer to the backend than the stories were. Storybook was still carrying the older assumptions in a few places.
- A normalized setup slice is the right boundary for the next preview step because it gives the page a stable place to read `source_id` values from.
- `studioDraftSlice` is now more clearly limited to UI-side output and microphone controls, which makes its remaining cleanup scope much smaller.

### What warrants a second pair of eyes

- Whether normalization should stay page-driven with a debounce, or move later into a more explicit editor orchestration layer.
- Whether the read-only normalized source cards should display more target detail than `source.name` once preview integration is in place.

### What should be done in the future

- Add preview ownership keyed by normalized `source_id`.
- Continue shrinking `studioDraftSlice` now that source cards no longer depend on it.
- Revisit whether output settings should also move closer to normalized DSL state later.

### Code review instructions

- Review the normalization flow:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup/setupSlice.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/setupApi.ts`
- Then review the presentational decoupling:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/source-card/types.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/source-card/SourceCard.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/SourceGrid.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/StatusPanel.tsx`
- Finally review the story updates:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/SourceCard.stories.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/SourceGrid.stories.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/StatusPanel.stories.tsx`

### Technical details

Commands used in this step:

```bash
sed -n '1,260p' ui/src/pages/StudioPage.tsx
sed -n '1,220p' ui/src/features/setup/setupSlice.ts
sed -n '1,220p' ui/src/stories/SourceGrid.stories.tsx
sed -n '1,220p' ui/src/stories/SourceCard.stories.tsx
sed -n '1,220p' ui/src/stories/StatusPanel.stories.tsx
pnpm --dir ui lint
pnpm --dir ui build
pnpm --dir ui build-storybook
```

## Step 8: Pivot The Remaining Transport Cleanup To Protobuf

This step changed the planned implementation order for the rest of the ticket. The original next move was to keep cleaning up websocket handling and preview ownership directly on top of handwritten REST and websocket transport types. That would have improved the frontend architecture, but it would also have hardened the wrong contract boundary.

The key realization was that the same transport schema is currently being maintained by hand in:

- `internal/web/api_types.go`
- `ui/src/api/types.ts`
- whatever runtime validation logic the frontend adds when raw websocket JSON needs to be trusted

That is too much duplication for a cleanup ticket whose explicit goal is to remove unclear and unnecessary code. Instead of adding a better handwritten websocket decoder, the right move is to switch the shared REST and websocket contract to protobuf-generated Go and TypeScript types before the preview and websocket cleanup goes much further.

### Prompt Context

**User prompt (verbatim):** "Alright, add a new implementation plan / analysis for switching to protobuf, within this ticket, and then tasks, and then tackle REST + WS protobuf use before we go on much further. That should simplify our code quite a bit."

**Assistant interpretation:** Pause the handwritten transport cleanup, document a schema-first protobuf migration inside `SCS-0003`, update the task plan to reflect that pivot, and then implement protobuf for shared REST and websocket transport before continuing the remaining frontend cleanup.

### What I did

- Read the `protobuf-go-ts-schema-exchange` skill and its supporting references.
- Checked the repo for existing protobuf or Buf tooling and confirmed there was none yet.
- Confirmed the local environment already has:
  - `buf`
  - `protoc`
- Created a new design document:
  - `ttmp/2026/04/09/SCS-0003--screencast-studio-frontend-cleanup-and-backend-alignment-plan/design-doc/02-protobuf-transport-migration-plan.md`
- Rewrote the task plan so the next implementation phases are now:
  - shared protobuf schema
  - Go REST and websocket migration
  - frontend REST and websocket decode migration
  - preview integration cleanup on top of the shared schema
- Abandoned an in-progress non-protobuf websocket and preview refactor by restoring those local UI files to the last committed state. That was intentional cleanup, not lost work. The aborted edits were exploratory and no longer fit the new direction.

### Why

- A protobuf-generated shared contract is a better simplification than more handwritten TypeScript interfaces plus websocket validators.
- REST and websocket payloads share too many entities to keep evolving separately.
- The cleanup ticket explicitly says not to add compatibility layers or preserve unnecessary code paths. A schema-first contract fits that goal better than a larger manually synchronized transport layer.

### What worked

- The protobuf skill already matched the exact migration we need: shared Go and TS schema generation with JSON transport.
- The repo already has `buf` and `protoc`, which means the migration can be implemented directly instead of only documented.
- Restoring the half-started non-protobuf websocket refactor put the tree back on a clean committed baseline before the real transport migration begins.

### What didn't work

- The first attempt to replace the newly created protobuf design document in one large patch failed because the template content did not match my patch context exactly.

### Exact failure

```text
apply_patch verification failed: Failed to find expected lines in .../02-protobuf-transport-migration-plan.md
```

I resolved that by deleting the template file content and recreating the full document explicitly with one clean patch.

### What I learned

- The next biggest simplification opportunity is not in component structure anymore. It is at the transport contract boundary.
- The current codebase is still early enough that adopting protobuf-generated shared transport should reduce complexity rather than add bureaucracy.
- The right migration unit is “REST plus websocket together,” not “websocket first, REST later.”

### What warrants a second pair of eyes

- Whether to keep API error payloads handwritten for the first protobuf slice or include them in the schema immediately.
- Whether generation should be wired into `go generate ./...` right away or first land with an explicit `buf generate` workflow.

### What should be done in the future

- Implement the shared protobuf schema and Buf generation config.
- Migrate Go REST and websocket transport to generated types.
- Migrate the frontend RTK Query and websocket layers to generated decode helpers.
- Only then continue the preview ownership cleanup.

### Code review instructions

- Read the new protobuf migration guide:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0003--screencast-studio-frontend-cleanup-and-backend-alignment-plan/design-doc/02-protobuf-transport-migration-plan.md`
- Then compare it to the current handwritten transport surface:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/api_types.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/types.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/session/wsClient.ts`
- Finally review the updated task ordering:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0003--screencast-studio-frontend-cleanup-and-backend-alignment-plan/tasks.md`

### Technical details

Commands used in this step:

```bash
sed -n '1,220p' /home/manuel/.codex/skills/protobuf-go-ts-schema-exchange/SKILL.md
sed -n '1,240p' /home/manuel/.codex/skills/protobuf-go-ts-schema-exchange/references/templates.md
sed -n '1,240p' /home/manuel/.codex/skills/protobuf-go-ts-schema-exchange/references/validation.md
rg --files | rg '(^|/)(buf\\.ya?ml|buf\\.gen\\.ya?ml|.*\\.proto|package.json|pnpm-workspace.yaml)$'
which buf
which protoc
git restore -- ui/src/api/types.ts ui/src/app/store.ts ui/src/components/preview/PreviewStream.tsx ui/src/components/source-card/SourceCard.tsx ui/src/components/source-card/types.ts ui/src/features/session/sessionSlice.ts ui/src/features/session/wsClient.ts
docmgr doc add --ticket SCS-0003 --doc-type design-doc --title "Protobuf Transport Migration Plan"
```
