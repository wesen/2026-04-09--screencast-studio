---
Title: Diary
Ticket: SCS-0009
Status: active
Topics:
    - screencast-studio
    - backend
    - frontend
DocType: reference
Intent: short-term
Owners: []
RelatedFiles:
    - Path: pkg/cli/serve.go
      Note: Serve preload flag and startup validation
    - Path: internal/web/server.go
      Note: Server bootstrap payload source
    - Path: ui/src/pages/StudioPage.tsx
      Note: Frontend bootstrap application timing
ExternalSources: []
Summary: Diary for the serve-time DSL preload implementation.
LastUpdated: 2026-04-10T13:05:00-04:00
WhatFor: Capture what changed, why, what worked, and how to review the serve preload feature.
WhenToUse: Use when reviewing or resuming ticket SCS-0009.
---

# Diary

## Goal

Capture the work to preload a DSL file into `screencast-studio serve` so the web UI starts from a real setup file instead of the built-in demo DSL.

## Step 1: Create the ticket and choose the bootstrap shape

I started by checking whether serve mode already had any file preload support. It did not. The CLI supports files for `record` and `setup compile`, but `serve` only takes network/runtime flags and the frontend always starts from its own hard-coded default DSL.

The key design choice in this step was to keep the feature small by treating preload as bootstrap data rather than as a late overwrite. The health/bootstrap request already happens early and already carries lightweight server config, so extending that payload is simpler than inventing a new startup endpoint just for preload state.

### Prompt Context

**User prompt (verbatim):** "yes, add a ticket + implementation plan + build it"

**Assistant interpretation:** Create a new ticket workspace, write a small implementation plan for serve-time DSL preload, implement the feature, and commit it with the usual diary workflow.

**Inferred user intent:** Make serve mode usable with a real setup file from the first browser load, while preserving the ticket/documentation discipline used for earlier changes.

**Commit (code):** N/A

### What I did
- Verified that `serve` has no `--file` preload support.
- Verified that the frontend currently starts from a built-in default DSL.
- Chose a startup bootstrap design based on extending the existing health response with optional preload DSL metadata.
- Created ticket `SCS-0009` and wrote the implementation guide.

### Why
- The feature spans CLI startup, web API bootstrap, and frontend hydration timing.
- A preload that lands too late can let the default DSL hydrate first, which would create confusing mixed state.

### What worked
- The current codebase already has a clear health/bootstrap request and a clear editor/store bootstrap path.
- The feature can be added without a new backwards-compatibility layer.

### What didn't work
- N/A

### What I learned
- The hard part is not loading the file; it is applying it before the first normalize/compile cycle.

### What was tricky to build
- The main tricky point was avoiding a race where the default frontend DSL normalizes before the preloaded DSL arrives. That would make the builder and preview setup boot from the wrong config even if the raw editor text were later replaced.

### What warrants a second pair of eyes
- The frontend bootstrap timing.
- Whether health response is the right place for preload metadata long-term.

### What should be done in the future
- Update this diary after the implementation and tests are committed.

### Code review instructions
- Start with the design doc, then review the startup path from CLI to health/bootstrap to UI.

### Technical details
- Ticket workspace: `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/10/SCS-0009--serve-preload-dsl-file-into-web-ui`
