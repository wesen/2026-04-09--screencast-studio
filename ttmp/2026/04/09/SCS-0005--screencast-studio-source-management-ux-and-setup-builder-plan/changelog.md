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
