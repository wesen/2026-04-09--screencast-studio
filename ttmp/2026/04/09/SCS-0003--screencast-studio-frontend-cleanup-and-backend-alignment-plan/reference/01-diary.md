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
LastUpdated: 2026-04-09T19:18:00-04:00
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
