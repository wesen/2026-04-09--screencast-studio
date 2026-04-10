# Changelog

## 2026-04-09

- Initial workspace created

## 2026-04-09

Created the dedicated source-management and setup-builder ticket, wrote the main design guide, and added a phased task plan covering discovery-backed source creation, source editing, structured draft state, DSL synchronization, and preview-safe source changes.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0005--screencast-studio-source-management-ux-and-setup-builder-plan/design-doc/01-source-management-ux-and-setup-builder-system-design.md — Main source-management design and implementation guide
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0005--screencast-studio-source-management-ux-and-setup-builder-plan/reference/01-diary.md — Diary explaining why this ticket exists
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0005--screencast-studio-source-management-ux-and-setup-builder-plan/tasks.md — Detailed phased implementation tasks

## 2026-04-09

Implemented the first source-management foundation slice:

- audited the mounted source-display path and the backend source contract
- introduced a dedicated `setup-draft` frontend feature
- defined explicit structured source types
- added conversion from normalized backend config into draft state
- wired `StudioPage` to hydrate the structured draft

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/types.ts — Explicit structured source model
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/conversion.ts — Hydration from backend-normalized config
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/setupDraftSlice.ts — Reducer and selectors for source-management state
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/app/store.ts — Store registration for the new slice
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx — Draft hydration from normalized config

## 2026-04-09

Implemented the discovery-backed source-creation slice:

- extended the setup-draft document so it can render full DSL
- added source-factory helpers for display/window/camera/region creation
- added a mounted `SourcePicker` flow driven by discovery data
- wired `StudioPage` to append selected sources and rewrite DSL text through the existing normalize path
- confirmed in a live smoke test that selecting a discovered window creates a new `window` source in raw DSL

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/conversion.ts — DSL rendering plus source factory helpers
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/types.ts — Expanded setup document with config-level fields
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/setupDraftSlice.ts — Expanded draft state selectors
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/SourcePicker.tsx — New mounted source-picker component
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/SourceGrid.tsx — Add-source affordance now available without full edit mode
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx — Discovery-backed source creation orchestration

## 2026-04-09

Implemented the mounted source-editing slice:

- source cards now edit the setup draft instead of exposing the fake scene selector
- rename, enable/disable, remove, and reorder all rewrite the canonical DSL through the existing normalize path
- the mounted grid now renders from the setup draft so structured edits appear immediately
- confirmed in a live smoke test that renaming a source updates the raw DSL

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/source-card/SourceCard.tsx — Real mounted editing controls
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/source-card/types.ts — Studio source metadata now includes editable detail text
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/SourceGrid.tsx — Mounted grid now forwards reorder actions
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx — Draft-owned source edits and DSL synchronization
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/styles/studio.css — Input/detail styling for mounted source cards

## 2026-04-09

Implemented the target-editor cleanup slice:

- added real target editors for window, camera, and region sources
- added region preset buttons derived from discovered display geometry
- removed the dead `solo` concept from the mounted source model and stories
- made the display-source limitation explicit instead of pretending monitor selection works
- confirmed in a live smoke test that changing a window source target updates both the source name and `window_id` in Raw DSL

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx — Target-specific editor rendering and source update flow
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/source-card/SourceCard.tsx — Generic card now renders injected editor content
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/SourceGrid.tsx — Source grid now forwards editor render content
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/conversion.ts — Region preset geometry helper export
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/styles/studio.css — Styling for the companion target editor

## 2026-04-09

Implemented the Raw DSL advanced-mode ownership slice:

- split editor state into applied DSL vs raw draft DSL
- changed the raw editor to explicit `Apply DSL` / `Reset` controls
- routed Raw DSL apply through backend normalize plus a builder compatibility round-trip
- locked the structured builder when advanced DSL uses unsupported shapes
- rendered the Studio source grid read-only from the applied normalized config while locked

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/editor/editorSlice.ts — Applied vs raw DSL state and structured-editor lock metadata
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/compatibility.ts — Builder compatibility comparison helper
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/dsl-editor/DSLEditor.tsx — Explicit advanced-mode apply/reset controls
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx — Apply flow, builder lock banner, and read-only Studio fallback

## 2026-04-09

Implemented the preview hard-cutover slice:

- preview ownership is now generation-aware for same-ID source edits
- meaningful source target changes release the old preview and then re-ensure a new one
- stale in-flight preview responses are released instead of being reattached
- only enabled sources are included in the desired preview set
- a live smoke test confirmed `release` followed by `ensure` when switching a window source target

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx — Preview generation tracking, restart path, and enabled-source preview policy
