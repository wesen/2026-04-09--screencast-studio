---
Title: Filename template tokens and UI guidance
Ticket: SCS-0010
Status: active
Topics:
    - screencast-studio
    - backend
    - frontend
DocType: index
Intent: short-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Ticket workspace for adding filename-oriented template tokens, incremental numbering, and UI guidance for output naming."
LastUpdated: 2026-04-10T14:05:00-04:00
WhatFor: "Track work on output filename template customization and user-facing documentation in the mounted studio UI."
WhenToUse: "Use when working on destination template rendering, output naming UX, or filename token help."
---

# Filename template tokens and UI guidance

## Overview

This ticket improves output naming in two places at once. First, the backend destination-template renderer needs to support an incremental filename token so a user can build collision-resistant names directly into the DSL. Second, the mounted UI needs a builder-supported way to edit the common filename-template shape and explain which tokens are available without forcing the user into Raw DSL.

The implementation should stay inside the existing `destination_templates` model. The structured UI should continue to rewrite the canonical templates rather than introducing a parallel naming configuration. Planned outputs remain the authoritative preview of what the naming rules produce.

## Key Links

- Design doc: [`design-doc/01-filename-template-tokens-implementation-guide.md`](./design-doc/01-filename-template-tokens-implementation-guide.md)
- Diary: [`reference/01-diary.md`](./reference/01-diary.md)
- Tasks: [`tasks.md`](./tasks.md)
- Changelog: [`changelog.md`](./changelog.md)
