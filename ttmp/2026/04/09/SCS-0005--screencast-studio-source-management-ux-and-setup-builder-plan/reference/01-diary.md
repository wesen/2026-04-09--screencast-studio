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
LastUpdated: 2026-04-09T21:55:00-04:00
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
