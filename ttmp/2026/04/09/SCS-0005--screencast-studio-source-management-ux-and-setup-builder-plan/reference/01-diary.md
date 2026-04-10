---
Title: Diary
Ticket: SCS-0005
Status: active
Topics:
    - frontend
    - backend
    - ui
    - dsl
    - video
    - product
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ui/src/components/studio/SourceGrid.tsx
      Note: Current read-only source-grid path that motivated this ticket
    - Path: ui/src/components/source-card/SourceCard.tsx
      Note: Current source-card rendering path that lacks full setup editing
    - Path: ui/src/pages/StudioPage.tsx
      Note: Current app-level owner for structured source management work
ExternalSources: []
Summary: Chronological record of creating the source-management and setup-builder follow-on ticket and implementing it slice by slice.
LastUpdated: 2026-04-09T22:15:00-04:00
WhatFor: Preserve why the source-management work was split into its own ticket and what it should cover.
WhenToUse: Read when starting implementation or reviewing the intended scope of the structured source builder.
---

# Diary

## Goal

Document why the source-management and setup-builder ticket was created and what the initial implementation scope is.

## Step 1: Create The Source Management Ticket

This ticket was created after identifying that the mounted UI now displays real sources and real previews, but still does not let a user build a capture setup in a structured way. The source grid is now real enough to show the setup, but not yet real enough to create or manage it.

That gap is large enough to deserve its own ticket because it involves:

- discovery
- structured draft state
- source editing UX
- raw DSL synchronization
- preview integration during source changes

### Prompt Context

**User prompt (verbatim):** "yes, and also create tickets for the source management and the better run lifecycle."

**Assistant interpretation:** Create a dedicated follow-on ticket for source management and the structured setup-builder experience, with a full intern-facing design and task breakdown.

### What I did

- Created `SCS-0005`.
- Added the main design guide and this diary.
- Wrote the initial task plan covering structured draft state, discovery-backed source creation, source editing, DSL synchronization, preview integration, and validation.

### Why

- This work is product-level and substantial enough that it should not be hidden inside a generic frontend bucket.
- Source management is central to the user experience and will affect multiple parts of the app architecture.

### Code review instructions

- Start with:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0005--screencast-studio-source-management-ux-and-setup-builder-plan/design-doc/01-source-management-ux-and-setup-builder-system-design.md`
- Then compare that guide against:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/SourceGrid.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/source-card/SourceCard.tsx`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`

### Technical details

Commands used in this step:

```bash
docmgr ticket create-ticket --ticket SCS-0005 --title "Screencast Studio Source Management UX and Setup Builder Plan" --topics frontend,backend,ui,dsl,video,product
docmgr doc add --ticket SCS-0005 --doc-type design-doc --title "Source Management UX and Setup Builder System Design"
docmgr doc add --ticket SCS-0005 --doc-type reference --title "Diary"
```

## Step 2: Audit The Current Path And Introduce A Structured Setup Draft

The first implementation slice was deliberately architectural. Before adding any new source-picker UI, I audited the current mounted source path and added a dedicated structured setup-draft feature in the frontend.

### What I inspected

- `ui/src/pages/StudioPage.tsx`
- `ui/src/components/studio/SourceGrid.tsx`
- `ui/src/components/source-card/SourceCard.tsx`
- `ui/src/api/discoveryApi.ts`
- `ui/src/api/setupApi.ts`
- `proto/screencast/studio/v1/web.proto`
- `internal/web/pb_mapping.go`
- `pkg/dsl/types.go`

### Findings

- The mounted app currently renders video sources from `normalizedConfig.videoSources`.
- The source grid and source cards already have editing-shaped props, but the mounted page keeps the grid read-only.
- The protobuf transport already exposes the backend truth needed for a real builder:
  - discovery descriptors
  - normalized effective video sources
  - normalized effective audio sources
- The v1 source kinds are stable:
  - display
  - window
  - region
  - camera
- Audio sources belong in the structured document even if the first mounted-builder UI still focuses on video source cards.

### Design decision

I added a new `setup-draft` feature instead of pushing this behavior into `setup` or `studioDraft`.

- `setup` remains the normalized backend result.
- `setup-draft` becomes the editable structured setup document.
- `studioDraft` continues to represent the existing output-panel controls.

That separation keeps the source-builder logic from contaminating unrelated state.

### What I implemented

- `ui/src/features/setup-draft/types.ts`
  - explicit structured source types for display, window, region, camera, and audio
- `ui/src/features/setup-draft/conversion.ts`
  - conversion from normalized backend config into the structured draft model
- `ui/src/features/setup-draft/setupDraftSlice.ts`
  - reducer and selectors for future source add/edit/remove/reorder operations
- `ui/src/app/store.ts`
  - reducer registration
- `ui/src/pages/StudioPage.tsx`
  - hydration from normalized config into the new draft slice

### Why this matters

Without this layer, every future source-management feature would either:

- mutate raw DSL directly from UI widgets, or
- keep depending on derived normalized source rendering with nowhere to store edits

Neither is a good foundation. The new draft feature gives the setup builder a real home.

### Validation

```bash
pnpm --dir ui build
pnpm --dir ui lint
docmgr doctor --ticket SCS-0005 --stale-after 30
```

### Next step

The next slice should make the builder visible by wiring discovery-backed source creation into the mounted Studio page and using the draft state as the owner of new sources.

## Step 3: Add Discovery-Backed Source Creation In The Mounted App

The next slice turned the source builder from architecture into behavior. The mounted `StudioPage` now has a real add-source flow backed by discovery data and the new setup-draft state.

### What I implemented

- Extended the setup-draft document so it carries the config-level fields needed to render canonical DSL:
  - schema
  - destination templates
  - audio mix template
  - audio output settings
- Added draft-rendering logic that emits full DSL from the structured setup document.
- Added source-factory helpers for:
  - display sources
  - window sources
  - camera sources
  - region sources using preset rectangles
- Added a new `SourcePicker` component in `ui/src/components/studio/SourcePicker.tsx`.
- Changed `SourceGrid` so the mounted page can expose `Add Source` without forcing all source cards into edit mode yet.
- Wired `StudioPage` to:
  - fetch discovery data
  - open the picker from the mounted source grid
  - create a draft source from the chosen resource
  - append that source to the setup draft
  - render the next DSL text
  - feed that DSL back into the existing normalize pipeline

### Why this approach

I deliberately reused the existing normalize path instead of inventing a second client-side source-of-truth for execution. That keeps:

- backend normalization
- preview leasing
- compile behavior
- raw DSL visibility

on the same path they already use today. The structured builder mutates the setup document, but the canonical runtime path still flows through DSL and backend normalization.

### Region behavior in this slice

For regions, the picker currently supports preset rectangles per discovered display:

- full display
- top half
- bottom half
- left half
- right half

This is enough to make the mounted app meaningfully useful before the dedicated region-rectangle editor lands in the later editing slice.

### Important limitation recorded here

The current backend runtime still treats `target.display` as an X11 display string, not a monitor identifier. That means the new display-source picker currently uses discovery data for naming and selection flow, but still emits `:0.0` as the underlying display target. Region sources are more precise because they render absolute rectangles from discovered monitor geometry.

That limitation should be revisited later, but it does not block the source-builder flow from becoming useful now.

### Smoke validation

I ran a live smoke test against the real server:

```bash
tmux new-session -d -s scs-scs0005-smoke 'cd /home/manuel/code/wesen/2026-04-09--screencast-studio && go run ./cmd/screencast-studio serve --addr :18080 --static-dir ui/dist'
```

Then, in the browser:

- opened the mounted Studio page
- clicked `Add Source`
- chose `Window`
- selected a discovered window
- switched to `Raw DSL`

Observed result:

- the header changed from `1 source armed` to `2 sources armed`
- the raw DSL contained a second `video_sources` entry with:
  - `type: "window"`
  - `window_id: "0x..."`
  - `destination_template: "per_source"`

That confirms the mounted source-creation flow now drives the real normalized setup path.

### Validation commands

```bash
pnpm --dir ui build
pnpm --dir ui lint
docmgr doctor --ticket SCS-0005 --stale-after 30
```

### Next step

The next slice should make existing source cards editable for rename/enable/remove/reorder and keep the structured setup + DSL text coherent when those edits happen.
