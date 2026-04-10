# Tasks

## Goal

Build a structured source-management experience that lets a user create, edit, reorder, and enable capture sources from discovery data without requiring raw DSL for ordinary use.

## Phase 1: Freeze The Existing Source Model

- [x] Audit the current source display path in `StudioPage`, `SourceGrid`, and `SourceCard`.
- [x] Audit the current discovery response shape and the current DSL source model.
- [x] Record the exact source kinds that must be supported in v1:
  - display
  - window
  - region
  - camera
  - audio source where relevant in the builder flow
- [x] Document the minimum editable fields per source kind.

Acceptance criteria:

- The intern can explain how discovery data maps to DSL source types today.
- There is a stable table of editable fields per source kind.

## Phase 2: Define The Structured Setup Draft Model

- [x] Define a frontend draft model for structured source editing.
- [x] Decide whether the draft model should mirror DSL closely or use a slightly more UI-friendly intermediate model.
- [x] Define conversion rules:
  - discovery -> draft source
  - draft source -> DSL
  - normalized backend config -> hydrated editor state when reopening
- [x] Decide where raw DSL remains visible in the product.

Acceptance criteria:

- The draft model is explicit and can round-trip to DSL without hidden fields.
- Raw DSL is positioned as advanced mode, not the only real editor.

## Phase 3: Discovery-Driven Source Creation

- [x] Add a source-picker flow that uses backend discovery data.
- [x] Let the user add displays, windows, cameras, and regions from actual discovered resources.
- [x] Define how region creation works:
  - choose a display
  - choose preset region
  - or input/edit a custom rectangle
- [x] Make the “Add Source” control real in the mounted app.

Acceptance criteria:

- A user can add sources from real discovered resources.
- The mounted app no longer treats source creation as a storybook-only affordance.

## Phase 4: Source Editing

- [ ] Implement per-source editing controls:
  - display selection
  - window selection
  - camera device selection
  - region rectangle editing
  - source name
  - enabled toggle
- [x] Allow reordering sources.
- [x] Allow removing sources.
- [x] Decide whether “solo” survives as a real feature or should be removed until it has backend meaning.

Acceptance criteria:

- Source cards or a companion editor actually modify the structured setup.
- No major source-edit operation requires raw DSL.

## Phase 5: Structured Draft To DSL Synchronization

- [x] Implement robust conversion from the structured setup draft into canonical DSL text.
- [x] Keep the Raw DSL tab synchronized with the structured editor.
- [x] Handle conflicts or parse failures when the user edits raw DSL directly.
- [x] Decide whether structured mode becomes read-only when raw DSL enters unsupported shapes.

Acceptance criteria:

- Structured editing and raw DSL remain coherent.
- The user does not lose work when switching between modes.

### Approved Policy

- Raw DSL remains visible as an advanced mode.
- Structured edits always regenerate Raw DSL text.
- Raw DSL edits do not affect the running setup until the user clicks `Apply DSL`.
- `Apply DSL` behavior:
  - if normalize fails, stay in Raw DSL and show errors
  - if normalize succeeds and the result is representable by the builder, hydrate structured state from that config
  - if normalize succeeds but uses unsupported shapes, keep the applied config active but lock structured editing behind a clear banner

### Concrete Implementation Tasks

- [x] Add separate `applied` vs `raw draft` DSL state in the frontend editor slice.
- [x] Change the Raw DSL tab from blur-driven implicit sync to explicit `Apply DSL` / `Reset` controls.
- [x] Make the Studio tab use the applied config, not un-applied Raw DSL edits.
- [x] Add a builder-compatibility check for applied Raw DSL.
- [x] Lock structured editing when advanced DSL is active but not builder-compatible.
- [x] Show a clear Studio-tab banner explaining why structured editing is unavailable.

## Phase 6: Preview Integration With Source Editing

- [ ] Ensure previews react correctly as sources are added, removed, renamed, or reconfigured.
- [ ] Make preview ownership robust when source IDs change.
- [ ] Decide whether previews should pause while a source is being edited or immediately re-ensure.

Acceptance criteria:

- Source editing does not break preview lifecycle ownership.
- Preview behavior remains understandable during edits.

### Approved Policy

- Use the simplest reliable approach: hard cutover on meaningful reconfiguration.
- Flicker is acceptable.
- Cosmetic edits like rename should not restart previews.
- Target/capture edits should release the old preview and let the ensure loop create a new one.

### Concrete Implementation Tasks

- [ ] Preview only enabled sources in the mounted Studio page.
- [ ] Introduce an explicit `restart preview for source` path in `StudioPage`.
- [ ] Restart previews on:
  - window target changes
  - camera device changes
  - region rectangle changes
  - other preview-relevant source changes as needed
- [ ] Do not restart previews on rename-only edits.
- [ ] Verify that the hard-cutover path does not leave orphaned preview leases.

## Phase 7: Validation And UX Guardrails

- [ ] Surface backend validation errors in the structured editor, not only in raw logs.
- [ ] Show incomplete-source warnings before recording.
- [ ] Add clear affordances for unsupported or advanced-only DSL features.

Acceptance criteria:

- The structured editor fails clearly when a source is incomplete or invalid.
- The user understands when they must use raw DSL for advanced cases.

## Phase 8: Tests And Smoke Validation

- [ ] Add frontend tests or Storybook states for each source kind and editing path.
- [ ] Add integration coverage for discovery -> source add -> normalize -> preview.
- [ ] Run a manual smoke test that verifies:
  - add display source
  - add window source
  - add camera source
  - add or edit region source
  - rename and reorder sources
  - remove a source
  - switch to raw DSL and back

Acceptance criteria:

- The ticket includes a repeatable validation plan.
- The source-management flow is no longer “read-only plus raw DSL”.

## Suggested Commit Boundaries

- [x] Commit 1: structured setup draft model and conversion rules
- [x] Commit 2: source-picker and source creation flow
- [x] Commit 3: source editing, reorder, and removal
- [x] Commit 4: structured/raw DSL synchronization and validation
- [ ] Commit 5: preview integration adjustments
- [ ] Commit 6: tests, stories, and smoke validation

## Progress Notes

- 2026-04-09: Completed the initial audit and introduced the new `setup-draft` feature with explicit structured source types, reducer actions, and hydration from normalized backend config.
- 2026-04-09: Added a mounted source-picker that uses discovery data for display, window, camera, and preset region creation, then renders the updated setup draft back into DSL text.
- 2026-04-09: Replaced the fake scene selector on mounted source cards with real rename, enable/disable, remove, and reorder actions backed by the setup draft and DSL sync. Target re-selection is still pending.
- 2026-04-09: Added real target editors for window sources, camera sources, and region rectangles/presets. Removed the dead `solo` concept from the mounted source model. Full per-monitor display selection remains blocked on a backend target-model change.
- 2026-04-09: Agreed product policy for the remaining work: Raw DSL becomes an explicit advanced apply flow, and preview reconfiguration uses simple hard-cutover release/re-ensure behavior.
- 2026-04-09: Implemented Raw DSL as an explicit advanced apply flow with separate applied vs draft text, builder-compatibility lockout, and a read-only Studio fallback for unsupported advanced shapes.
