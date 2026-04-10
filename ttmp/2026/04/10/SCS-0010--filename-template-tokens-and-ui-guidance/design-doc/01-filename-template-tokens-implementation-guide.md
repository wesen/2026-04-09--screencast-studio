---
Title: Filename template tokens implementation guide
Ticket: SCS-0010
Status: active
Topics:
    - screencast-studio
    - backend
    - frontend
DocType: design-doc
Intent: short-term
Owners: []
RelatedFiles:
    - Path: pkg/dsl/templates.go
      Note: Destination-template renderer and token expansion
    - Path: ui/src/features/setup-draft/conversion.ts
      Note: Builder-managed destination-template rewriting
    - Path: ui/src/components/studio/OutputPanel.tsx
      Note: Output naming controls and inline token documentation
ExternalSources: []
Summary: Implementation guide for filename-template token support and mounted UI guidance.
LastUpdated: 2026-04-10T14:05:00-04:00
WhatFor: Explain how to add incremental filename rendering and expose token-oriented output naming controls in structured mode.
WhenToUse: Use when implementing or reviewing ticket SCS-0010.
---

# Filename template tokens implementation guide

## Goal

Add filename-oriented template support that is discoverable from the mounted UI without introducing a new output-naming model.

## Desired Outcome

The user should be able to stay in structured mode, set a destination root, add a filename suffix like `-{date}-{index}`, and immediately see the resulting planned outputs. The backend should render those templates consistently during compile preview and actual recording.

## Implementation Shape

### 1. Backend template rendering

- Keep `destination_templates` as the canonical source of naming rules.
- Preserve existing `{date}`, `{time}`, and `{timestamp}` expansion.
- Add `{index}` to resolve to the first free number for the final candidate path.
- Apply `{index}` after the other placeholders have rendered so filesystem lookup happens against a concrete path.

### 2. Structured-mode builder support

- Keep the builder-managed template shape narrow:
  - per-source: `<root>/{session_id}/{source_name}<suffix>.{ext}`
  - audio-mix: `<root>/{session_id}/audio-mix<suffix>.{ext}`
- Parse the current templates back into:
  - destination root
  - filename suffix
- Treat any non-matching template shape as advanced DSL and keep structured editing locked for that surface.

### 3. UI guidance

- Add a filename suffix input beside the existing destination-root controls.
- Show concise inline help that explains:
  - video file naming shape
  - audio-mix naming shape
  - available tokens: `{date}`, `{time}`, `{timestamp}`, `{index}`
- Rely on the planned output list as the authoritative preview instead of inventing a second preview component.

## Validation Plan

- Add backend tests for `{index}` collision avoidance.
- Run Go tests.
- Run frontend typecheck/build validation.

## Review Notes

- Start with backend rendering in `pkg/dsl/templates.go`.
- Then review builder-template parsing and rewriting in `ui/src/features/setup-draft/conversion.ts`.
- Finish with the mounted output panel UI and token help text.
