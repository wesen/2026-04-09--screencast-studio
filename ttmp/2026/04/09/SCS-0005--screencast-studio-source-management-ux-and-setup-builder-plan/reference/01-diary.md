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
LastUpdated: 2026-04-09T22:26:00-04:00
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

## Step 4: Make The Mounted Source Cards Actually Edit The Setup

The third slice focused on the mounted source cards themselves. Before this change, the cards still exposed a fake scene selector that looked interactive but did not represent real source editing. I replaced that with controls that operate on the structured setup draft and immediately rewrite the canonical DSL.

### What changed

- The mounted source grid now renders from `setupDraft.videoSources` instead of only from `normalizedConfig.videoSources`.
- `StudioPage` now treats the setup draft as the owner of:
  - rename
  - enable/disable
  - remove
  - reorder
- After each edit, `StudioPage` renders the next DSL text and feeds it back into the existing normalize path.
- `SourceCard` no longer uses the placeholder scene dropdown in the mounted product path.
- `SourceCard` now shows:
  - a text input for the source name
  - a short target detail line
  - enable/disable
  - move up
  - move down
  - remove in the title bar

### Important UI decision

I kept the target-specific re-selection work out of this slice. The mounted cards now support honest editing for source identity and ordering, but they still do not yet let the user retarget:

- display source target
- window source target
- camera device target
- region rectangle

That keeps this slice coherent while still removing the fake scene-selector behavior.

### Smoke validation

I ran another live smoke test against the real server and verified:

- the mounted source card shows a real editable name field
- editing the name updates the raw DSL after the normalize debounce

Observed example:

- source card name changed from `Full Desktop` to `Primary Desktop`
- the Raw DSL tab then showed:
  - `name: "Primary Desktop"`

That confirms the mounted card edits are now writing through the same canonical path as the rest of the app.

### Validation commands

```bash
pnpm --dir ui build
pnpm --dir ui lint
docmgr doctor --ticket SCS-0005 --stale-after 30
```

### Next step

The next slice should finish the remaining target-editing gaps:

- choose a different discovered display target
- choose a different window target
- choose a different camera device
- edit region rectangles directly

## Step 5: Add Real Target Editors And Remove Dead Solo UI

This slice finished most of the remaining source-editing work for v1 by adding target-specific editors where the backend/source model can support them today.

### What changed

- Added companion editor content inside mounted source cards rather than growing `SourceCard` into a giant source-type switch.
- Implemented real target editing for:
  - window sources via a discovered-window `<select>`
  - camera sources via a discovered-camera `<select>`
  - region sources via:
    - direct numeric `x/y/w/h` editing
    - preset buttons derived from discovered display geometry
- Removed the dead `solo` field from the mounted source model and Storybook fixtures.

### Why the companion-editor approach

The mounted card still owns common controls:

- name
- enabled state
- remove
- reorder

But the target editor is injected by `StudioPage`, which has the actual draft source plus the latest discovery payload. That keeps the generic card reusable and keeps the source-type branching near the real setup orchestration.

### Important display-source limitation

I did **not** fake per-monitor display selection for `display` sources.

Current backend/runtime reality:

- display sources still capture a root X11 display string like `:0.0`
- discovered monitors are monitor descriptors, not a first-class runtime target in the DSL

So the mounted editor now shows an explicit note for display sources instead of pretending monitor selection works. That is the correct product behavior until the backend target model changes.

### Smoke validation

I ran a live smoke test against the real server and verified the new window target editor:

1. add a new window source from discovery
2. use the inline `Window` selector on that source card
3. switch to `Raw DSL`

Observed result:

- the source `name` changed to the newly selected window title
- `target.window_id` changed to the selected discovered window ID

That confirms the target editor writes through the structured setup and canonical DSL path.

### Validation commands

```bash
pnpm --dir ui build
pnpm --dir ui lint
docmgr doctor --ticket SCS-0005 --stale-after 30
```

### Remaining work after this slice

- decide where raw DSL should live in the finished product
- add structured/raw conflict handling and unsupported-shape behavior
- make previews react more clearly to source reconfiguration
- add validation states and stronger smoke coverage

## Step 6: Lock The Remaining Direction Before Implementing It

Before continuing, the remaining ambiguous product behavior was narrowed down with two explicit decisions.

### Raw DSL decision

The Raw DSL tab will remain, but it is no longer treated as a live peer editor. The intended behavior is now:

- structured editing is the primary path
- structured edits regenerate Raw DSL
- Raw DSL edits are local draft changes until the user clicks `Apply DSL`
- if applied Raw DSL is builder-compatible, structured editing rehydrates from it
- if applied Raw DSL is not builder-compatible, the config stays active but the structured editor locks behind a clear banner

This keeps the product simple and avoids trying to live-sync both directions on every keystroke.

### Preview reconfiguration decision

For preview behavior during source reconfiguration, the chosen policy is the simplest reliable one:

- hard cutover
- release old preview
- allow the normal ensure loop to create a new preview
- flicker is acceptable
- preview continuity is less important than correctness and flexibility

That means the remaining preview work can stay in the frontend orchestration rather than needing a more elaborate reconciliation layer.

### Why these decisions matter

These decisions remove the two biggest remaining sources of complexity in `SCS-0005`:

- bidirectional builder/raw synchronization rules
- trying to be clever about preview reuse during edits

With those choices fixed, the next two slices are now clear:

1. Raw DSL advanced apply and builder lockout
2. Simple preview hard-cutover on meaningful source reconfiguration

## Step 7: Turn Raw DSL Into An Explicit Advanced Apply Flow

The next slice made Raw DSL stop behaving like a hidden second primary editor. Before this change, the raw tab still used a blur-driven local buffer that wrote straight back into the canonical DSL state. That was convenient early on, but it blurred ownership and made it too easy for advanced-only setups to desynchronize the builder.

### Product behavior after this slice

The product now behaves like this:

- structured edits still regenerate the canonical applied DSL immediately
- Raw DSL edits are draft-only until the user clicks `Apply DSL`
- the Studio tab, previews, and record path continue to use the applied DSL
- `Apply DSL` normalizes through the backend
- normalized advanced DSL is then checked for builder compatibility by round-tripping it through the structured setup draft renderer
- compatible advanced DSL rehydrates the structured builder
- incompatible advanced DSL remains active but the Studio builder locks and shows a clear banner

### What I implemented

- Extended `ui/src/features/editor/editorSlice.ts` with:
  - `dslText` as the applied canonical DSL
  - `rawDslText` as the advanced-mode draft
  - lock state plus lock reason for structured editing
- Added `ui/src/features/setup-draft/compatibility.ts` to compare the applied normalized config against the builder round-trip result.
- Reworked `ui/src/components/dsl-editor/DSLEditor.tsx` into a controlled editor with:
  - `Apply DSL`
  - `Reset`
  - no blur-driven implicit sync
- Reworked `ui/src/pages/StudioPage.tsx` so:
  - startup hydration from normalized config happens only once for the builder
  - Raw DSL apply explicitly drives normalize + compatibility check
  - builder-compatible advanced DSL rehydrates `setup-draft`
  - incompatible advanced DSL keeps the applied config live but renders the Studio source grid read-only from `normalizedConfig`

### Why this approach is better

This keeps the ownership model simple:

- the builder is the normal path
- Raw DSL is an advanced override
- unsupported advanced setups do not force the code into a confusing partial-sync state

That is much easier to reason about than trying to keep two live editors mutually authoritative at all times.

### Validation

Commands run:

```bash
pnpm --dir ui build
pnpm --dir ui lint
```

### Review focus

Review these files together:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/editor/editorSlice.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/compatibility.ts`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/dsl-editor/DSLEditor.tsx`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`

### Next step

The next slice should make preview ownership react explicitly to meaningful source reconfiguration by releasing and recreating previews on target changes instead of leaving same-ID sources on stale preview leases.

## Step 8: Add Hard-Cutover Preview Restart On Meaningful Source Edits

The next slice made preview lifecycle ownership explicit instead of trying to infer it indirectly from source IDs alone. Before this change, target edits like window changes or region rectangle changes kept the same source ID, so the existing ensure loop had no reason to request a fresh preview. That left room for stale previews to remain attached to updated sources.

### Product behavior after this slice

Preview behavior is now intentionally simple:

- previews are only desired for enabled Studio-tab sources
- rename-only edits do not restart previews
- meaningful target edits bump the preview generation for that source
- if an owned preview exists, it is released first
- the normal ensure loop then creates a new preview for the updated source
- if an older in-flight ensure completes after the source changed, that stale preview is released instead of being tracked

This is a hard cutover. Flicker is acceptable. Correctness is more important than continuity.

### What I implemented

- Reworked `ui/src/pages/StudioPage.tsx` preview orchestration so it now has:
  - a preview generation ref per source
  - an explicit `restartPreviewForSource(...)` path
  - separate handling for releasing owned previews vs releasing stale detached previews
  - a local preview-sync nonce to force the ensure loop to reevaluate same-ID source edits
- Updated the desired preview set so only enabled sources are previewed.
- Hooked preview restart into meaningful source changes:
  - window target selection
  - camera device selection
  - region rectangle edits and presets
- Left rename-only edits alone so they do not churn previews.

### Live smoke result

I ran the real server in tmux and exercised the mounted app through Playwright. The important evidence came from the server log when changing a window source from one discovered window to another:

- `POST /api/previews/release`
- immediately followed by `POST /api/previews/ensure`

That confirms the hard-cutover path is actually firing on same-ID source reconfiguration instead of silently holding onto the old preview lease.

### Validation

Commands run:

```bash
pnpm --dir ui build
pnpm --dir ui lint
docmgr doctor --ticket SCS-0005 --stale-after 30
```

Live smoke commands used:

```bash
lsof-who -p 18080 -k || true
tmux new-session -d -s scs-0005-preview-smoke 'cd /home/manuel/code/wesen/2026-04-09--screencast-studio && go run ./cmd/screencast-studio serve --addr :18080 --static-dir ui/dist'
tmux capture-pane -pt scs-0005-preview-smoke
curl -sSf http://127.0.0.1:18080/api/previews
tmux kill-session -t scs-0005-preview-smoke
```

### Review focus

Review these sections in:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`

Pay particular attention to:

- `desiredPreviewSourceIds`
- `releaseOwnedPreviewForSource`
- `releaseDetachedPreview`
- `restartPreviewForSource`
- the generation check inside the ensure loop

### Next step

The remaining work in `SCS-0005` is now mostly validation, unsupported-source guardrails, and deciding how far the structured editor should go before deferring to Raw DSL.
