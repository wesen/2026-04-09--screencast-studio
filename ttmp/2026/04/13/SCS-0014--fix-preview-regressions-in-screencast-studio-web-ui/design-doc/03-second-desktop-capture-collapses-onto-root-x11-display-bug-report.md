---
Title: Second desktop capture collapses onto root X11 display bug report
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - video
    - backend
    - ui
    - analysis
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/media/gst/shared_video.go
      Note: |-
        Shared display source signature is based on the root display target, so distinct monitor sources can collapse together
        Shared-source signatures collapse plain display captures with the same root display target
    - Path: ui/src/features/setup-draft/conversion.ts
      Note: |-
        Display source drafts currently normalize to the same default display target
        Display source drafts still normalize to the default root display target
    - Path: ui/src/pages/StudioPage.tsx
      Note: |-
        UI already warns that full-display sources still target the root X11 display
        Explicitly warns that display sources still use the root X11 target
ExternalSources: []
Summary: Bug report for the second desktop/display preview not behaving as an independent source because full-display capture still targets the root X11 display (`:0.0`) rather than monitor-specific geometry or connector identity.
LastUpdated: 2026-04-13T22:00:00-04:00
WhatFor: Explain why the second desktop capture did not work independently and propose a path to a real per-display model.
WhenToUse: Read when fixing monitor-specific capture or debugging why display previews collapse into the same root X11 capture source.
---


# Second desktop capture collapses onto root X11 display bug report

## Executive Summary

The user reported that the second desktop capture did not work at the end of the session. The logs confirm that the backend did not create an independent per-display capture. Instead, the `eDP-1` display preview reused the existing shared source for the root X11 display `:0.0`. This is a known limitation in the UI, but it is still a bug from a user perspective because the app presents “Full Desktop” and `eDP-1` as distinct sources while the backend still collapses them onto the same capture target.

The recommended fix is to stop modeling full-display sources as a single root-display capture and instead map each display source to monitor-specific geometry or an equivalent backend target model.

## Problem Statement

Observed behavior:

- the user enabled a second desktop/display source,
- expected an independent desktop preview,
- but did not get a distinct capture.

Backend log evidence showed the `eDP-1` source reusing the already-running shared display capture instead of creating a separate monitor-targeted capture.

## Evidence

### UI warning evidence

The UI already contains a warning acknowledging the limitation (`ui/src/pages/StudioPage.tsx:1139-1143`):

```text
Full-display sources still use the runtime's root X11 target (`:0.0`).
Per-monitor display selection needs a backend target-model change...
```

That warning is honest, but it also confirms the root cause.

### Source-creation evidence

`createDisplaySourceDraft(...)` currently hardcodes the display target to `DEFAULT_DISPLAY_TARGET` rather than a monitor-specific geometry or connector identity (`ui/src/features/setup-draft/conversion.ts:235-252`).

Key line:

- `target: { displayId: DEFAULT_DISPLAY_TARGET }`

### Shared-source signature evidence

The shared source signature is computed from `source.Target.Display` plus optional rect/window/device information (`pkg/media/gst/shared_video.go:619-635`). For plain display sources, there is no rect or connector in the signature, only the root display target.

### Runtime log evidence

The live log sequence for `source_id=edp-1` showed:

```text
created preview worker ... source_id=edp-1 source_name=eDP-1 source_type=display
reused shared gstreamer video source ... signature=type=display|display=:0.0|...
attached shared preview consumer ... source_id=desktop-1
started gstreamer preview session ... source_id=desktop-1 source_name="Full Desktop"
```

That is the clearest possible evidence that the second display preview collapsed onto the same root display source.

## Likely Root Cause

This is not a transient runtime race. It is a source-model limitation.

### Current model

The current display source model effectively says:

- all display previews target `:0.0`
- only region/window sources have distinct geometry
- therefore all full-display sources on the same root display collapse together

### Why that breaks user expectations

The discovery layer exposes named monitors such as `eDP-1`. The source picker lets users create display sources named after those monitors. But the runtime does not yet preserve that distinction for preview capture.

That means the UI is conceptually ahead of the runtime.

## Fixing Analysis

There are two plausible fix levels.

### Option A: disable or strongly constrain per-display UI until the backend is real

This would be the safest short-term honesty fix:

- if monitor-specific capture is not real, do not make the user think it is.

Pros:

- prevents misleading behavior immediately
- low engineering risk

Cons:

- does not actually add per-display capture
- removes or degrades a potentially useful UI path

### Option B: make display sources geometry-backed

This is the correct feature fix.

Instead of letting a display source mean only “`display=:0.0`”, translate each chosen display into an explicit region matching that monitor's geometry.

Pros:

- aligns runtime behavior with the source picker
- naturally integrates with the existing shared-source signature model because rects are already part of the signature
- avoids inventing a separate monitor-specific backend element immediately

Cons:

- needs careful normalization and discovery plumbing
- requires a clear decision about how to treat monitor reconfiguration

### Recommendation

Use **geometry-backed display capture** as the real fix. The system already knows monitor geometry in discovery, and region capture already works. That makes this a good match for the current runtime design.

## Implementation Plan

### Phase 1: Normalize display sources to connector-specific geometry

Target files:

- `pkg/discovery/service.go`
- `ui/src/features/setup-draft/conversion.ts`
- `pkg/dsl/normalize.go`

Plan:

- preserve the chosen display connector/ID from the UI
- resolve it to explicit geometry
- carry that geometry into the effective source model for display capture

### Phase 2: Include monitor-specific geometry in the capture signature

Target file:

- `pkg/media/gst/shared_video.go`

Plan:

- if a display source resolves to a rect, that rect should be part of the signature
- distinct monitors should then get distinct shared capture sources instead of collapsing into one root-display capture

### Phase 3: Update the UI messaging

Target file:

- `ui/src/pages/StudioPage.tsx`

Plan:

- once the backend really supports per-display capture, remove or revise the current warning text
- if the fix is partial at first, make the warning more specific instead of leaving the UI in a confusing half-supported state

## Alternatives Considered

### Alternative: keep the current behavior and leave the warning

Rejected for bug-fix purposes. The current behavior is internally acknowledged but still user-hostile.

### Alternative: invent a brand-new monitor-specific capture backend immediately

Probably unnecessary. Geometry-backed capture is a better first fix because it reuses existing region-capture semantics.

## Validation Checklist

- adding `Full Desktop` and `eDP-1` produces distinct preview signatures
- `eDP-1` preview no longer reuses the root full-desktop shared source
- two display previews can coexist when they represent distinct geometries
- the UI no longer needs the current "root X11 target" warning once the fix is complete

## References

- `ui/src/pages/StudioPage.tsx:1139-1143`
- `ui/src/features/setup-draft/conversion.ts:235-252`
- `pkg/media/gst/shared_video.go:619-635`
- live logs captured from the running `scs-web-ui` tmux session on 2026-04-13
