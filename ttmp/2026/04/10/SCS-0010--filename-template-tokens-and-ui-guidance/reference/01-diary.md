---
Title: Diary
Ticket: SCS-0010
Status: active
Topics:
    - screencast-studio
    - backend
    - frontend
DocType: reference
Intent: short-term
Owners: []
RelatedFiles:
    - Path: pkg/dsl/templates.go
      Note: Destination-template token rendering and incremental numbering
    - Path: ui/src/features/setup-draft/conversion.ts
      Note: Builder-managed filename suffix parsing and rewriting
    - Path: ui/src/components/studio/OutputPanel.tsx
      Note: Output naming controls and inline documentation
ExternalSources: []
Summary: Diary for filename-template tokens and mounted-UI guidance.
LastUpdated: 2026-04-10T14:25:00-04:00
WhatFor: Capture what changed, why, what worked, and how to review ticket SCS-0010.
WhenToUse: Use when reviewing or resuming ticket SCS-0010.
---

# Diary

## Goal

Capture the work to make filename template tokens easier to use by adding incremental numbering in the backend and exposed filename-token guidance in the mounted UI.

## Step 1: Create the ticket and keep the feature inside destination templates

I started by reading the existing destination-template renderer and the mounted output controls. The key constraint was to avoid inventing a second naming model. The codebase already treats `destination_templates` as the canonical output-naming source, and the UI already previews resolved outputs from the backend compile path, so the clean solution was to extend that path rather than adding frontend-only naming logic.

That led to a narrow design: add one new backend placeholder, `{index}`, and extend the builder-supported template shape from “root only” to “root plus filename suffix”. That gives the user a practical surface for patterns like `-{date}-{index}` without opening the full raw DSL for the common case.

### Prompt Context

**User prompt (verbatim):** "ok, create a new ticket that handles adding templates to the filenames (date fields, incremental number) and also showing some documentation in the UI to help the user. 

Add an implementation guide, and then implement"

**Assistant interpretation:** Create a new ticket workspace, write a small implementation guide, then implement filename-template support that exposes date/time tokens and incremental numbering while documenting the feature in the mounted UI.

**Inferred user intent:** Make output naming easier to discover and safer to use from the main studio interface, not just from raw DSL editing.

**Commit (code):** N/A

### What I did
- Reviewed `pkg/dsl/templates.go` and the mounted output controls in `ui/src/components/studio/OutputPanel.tsx`.
- Confirmed that date/time placeholders already exist in the backend renderer.
- Chose a builder-managed naming shape based on:
  - destination root
  - filename suffix appended after `{source_name}` or `audio-mix`
- Created ticket `SCS-0010` and wrote the implementation guide.

### Why
- The planned-output list already acts as the authoritative naming preview.
- Reusing `destination_templates` keeps structured mode and raw DSL consistent.

### What worked
- The existing template shape for `per_source` and `audio_mix` was regular enough to parse and rewrite safely for the common case.

### What didn't work
- N/A

### What I learned
- The fastest path was not “full filename-template editing”; it was “suffix editing on top of the existing canonical template shape”.

### What was tricky to build
- The design pressure was between flexibility and keeping structured mode honest. A fully arbitrary filename-template editor would have required a new abstraction or builder-specific template semantics. Restricting structured mode to a shared suffix shape kept the feature understandable and avoided a hidden compatibility layer.

### What warrants a second pair of eyes
- Whether the suffix-only structured model is the right long-term compromise, or whether users will eventually need prefix and nested-path controls too.

### What should be done in the future
- Revisit the structured naming model only if users explicitly need more than suffix tokens.

### Code review instructions
- Start with the implementation guide and confirm the intended builder-managed template shape.

### Technical details
- Ticket workspace: `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/10/SCS-0010--filename-template-tokens-and-ui-guidance`

## Step 2: Implement `{index}` rendering and mounted UI guidance

I implemented the feature in both the backend renderer and the mounted studio UI. The backend now resolves `{index}` by scanning for the first free candidate path after the other placeholders have rendered. The frontend parses the builder-managed destination templates back into a destination root and filename suffix, so a user can type values like `-{date}-{index}` in structured mode and immediately see the rendered output filenames in the planned-output list.

The UI work stayed intentionally small. I added a filename input next to the existing destination root, preserved the existing advanced-template lock behavior for non-builder shapes, and added inline documentation that shows both the naming shape and the supported tokens. I also refreshed the embedded web assets so `go run ./cmd/screencast-studio` will serve the updated mounted UI instead of stale bundled JavaScript.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Complete the backend and frontend changes, validate them, and record the final implementation details.

**Inferred user intent:** Make filename token support usable from the main UI and keep the code/documentation trail clean.

**Commit (code):** `bab0885` — `output: add filename template tokens`

### What I did
- Added `{index}` collision-avoidance rendering in `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/templates.go`.
- Added a regression test for date-plus-index filename rendering in `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/templates_test.go`.
- Extended builder-managed template parsing and rewriting in `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/conversion.ts`.
- Added a filename suffix control and token help text in `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/OutputPanel.tsx`.
- Wired the new field through `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`.
- Updated output panel stories in `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/OutputPanel.stories.tsx`.
- Refreshed embedded assets with `go generate ./internal/web`.
- Ran:

```bash
gofmt -w pkg/dsl/templates.go pkg/dsl/templates_test.go
go test ./pkg/dsl ./internal/web -count=1
go test ./... -count=1
pnpm --dir ui build
go generate ./internal/web
```

### Why
- `{index}` solves the missing incremental-number part of the filename-template request.
- Builder-managed suffix editing lets the user access the common tokenized naming case without leaving structured mode.
- Inline UI text is the lowest-friction place to document the tokens because the output controls are where the user is already thinking about filenames.

### What worked
- The backend renderer could add `{index}` without changing the compile/recording API shape.
- Regex-based parsing of the existing `per_source` and `audio_mix` templates was enough to support a destination-root-plus-suffix model.
- The planned output list already gave the user a concrete preview of the generated filenames, so no extra preview system was needed.
- `go generate ./internal/web` refreshed the embedded frontend bundle after the UI change.

### What didn't work
- My first validation pass failed because of a typo in the new standard-library error alias:

```text
# github.com/wesen/2026-04-09--screencast-studio/pkg/dsl
pkg/dsl/templates.go:4:2: "errors" imported as stderrors and not used
pkg/dsl/templates.go:65:5: undefined: sterrors
FAIL	github.com/wesen/2026-04-09--screencast-studio/pkg/dsl [build failed]
FAIL	github.com/wesen/2026-04-09--screencast-studio/internal/web [build failed]
FAIL
```

- I fixed that import typo and reran the same validation commands instead of changing the feature shape further.

### What I learned
- The code already had the right architecture for discoverable output naming: the missing piece was exposure in structured mode, not a new backend API.
- Refreshing embedded assets is still easy to forget after frontend changes, so it is worth keeping in the validation checklist for `serve`-mode UI work.

### What was tricky to build
- The sharp edge was making the structured UI more capable without pretending to support arbitrary template shapes. The parsing logic had to accept builder-managed patterns with shared suffixes while still rejecting advanced DSL shapes cleanly. That is why the code treats both root and suffix as derived values from the canonical templates instead of storing extra frontend-only state.

### What warrants a second pair of eyes
- The `{index}` filesystem lookup semantics, especially whether starting at `1` matches user expectation.
- The exact copy in the inline token documentation.
- Whether trimming leading/trailing whitespace from the filename suffix is the right structured-mode policy.

### What should be done in the future
- Consider adding a focused frontend test harness if the project starts relying on more behavior in `conversion.ts`.
- Consider adding duplicate planned-output path detection if output-name collisions become a recurring source of support issues.

### Code review instructions
- Start with `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/templates.go` and `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/templates_test.go`.
- Then review `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/conversion.ts`.
- Finish with `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/OutputPanel.tsx` and `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`.
- Validate by running the commands listed above, then launch `go run ./cmd/screencast-studio serve` and check that filename suffix edits change the planned output paths.

### Technical details
- Structured-mode naming shape:

```text
per_source: <root>/{session_id}/{source_name}<suffix>.{ext}
audio_mix:  <root>/{session_id}/audio-mix<suffix>.{ext}
```

- Supported inline tokens:

```text
{date}       -> 2006-01-02
{time}       -> 15-04-05
{timestamp}  -> 20060102-150405
{index}      -> 1, 2, 3, ... first free filename
```
