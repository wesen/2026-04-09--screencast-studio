---
Title: Serve DSL preload implementation guide
Ticket: SCS-0009
Status: active
Topics:
    - screencast-studio
    - backend
    - frontend
DocType: design-doc
Intent: short-term
Owners: []
RelatedFiles:
    - Path: pkg/cli/serve.go
      Note: Serve CLI flags and startup validation live here
    - Path: internal/web/server.go
      Note: Server configuration carries preload bootstrap data
    - Path: internal/web/handlers_api.go
      Note: Health/bootstrap response is served here
    - Path: internal/web/pb_mapping.go
      Note: Health response mapping must expose preload fields
    - Path: proto/screencast/studio/v1/web.proto
      Note: Bootstrap payload schema changes live here
    - Path: ui/src/pages/StudioPage.tsx
      Note: Frontend bootstrap timing and hydration logic live here
    - Path: ui/src/api/discoveryApi.ts
      Note: Health/bootstrap query already exists and can carry preload info
    - Path: ui/src/features/editor/editorSlice.ts
      Note: The built-in default DSL currently wins at startup
ExternalSources: []
Summary: "Small guide for adding a serve --file preload path and hydrating the web UI from that file."
LastUpdated: 2026-04-10T13:05:00-04:00
WhatFor: "Guide the implementation of serve-time DSL preloading."
WhenToUse: "Use when implementing or reviewing startup-time setup hydration."
---

# Serve DSL preload implementation guide

## Problem

Today `screencast-studio serve` always boots the web app with the frontend’s built-in demo DSL. The CLI can compile or record from files, but serve mode cannot preload a setup file into the web UI.

That means users cannot launch the web control surface already pointed at a real setup. They have to paste or rebuild the DSL in the browser after startup.

## Desired behavior

This should work:

```bash
screencast-studio serve --file ./examples/my-setup.yaml
```

When the browser opens:

- the editor should contain the file contents,
- normalize/compile should run against that DSL,
- the structured builder should hydrate from that setup,
- the initial preview selection should come from that setup rather than the hard-coded demo.

## Implementation approach

Use the existing health/bootstrap request as the startup transport.

1. Add `--file` to the `serve` command.
2. Read the DSL file and validate it at startup.
3. Store the raw DSL text and optional path in `web.Config`.
4. Extend `HealthResponse` with optional preload fields.
5. Return those preload fields from `/api/healthz`.
6. In the frontend, wait for `healthz` before the first normalize/compile pass.
7. If preload DSL is present, apply it to the editor state before bootstrapping normalize/compile.

## Why this shape

The health query already happens very early in the page lifecycle and already carries lightweight server bootstrap data like preview limit. Reusing it keeps the feature small and avoids introducing a second startup endpoint.

The important constraint is timing. The preload must be applied before the first normalize/compile cycle, otherwise the default DSL can hydrate the builder first and leave the page in a mixed state.

## Main code changes

- `pkg/cli/serve.go`
  - add `--file`
  - load file bytes
  - validate with the application before starting the server
- `internal/web/server.go`
  - carry preload DSL text/path in config
- `internal/web/handlers_api.go` and `internal/web/pb_mapping.go`
  - expose preload fields in health/bootstrap
- `proto/screencast/studio/v1/web.proto`
  - extend `HealthResponse`
- `ui/src/pages/StudioPage.tsx`
  - delay first normalize/compile until bootstrap is ready
  - apply preload DSL once

## Review checklist

- Confirm invalid preload files fail before `ListenAndServe`.
- Confirm no preload keeps today’s default behavior.
- Confirm preload DSL wins over the built-in editor default on first page load.
- Confirm setup draft hydration uses the preloaded config, not the default config.
