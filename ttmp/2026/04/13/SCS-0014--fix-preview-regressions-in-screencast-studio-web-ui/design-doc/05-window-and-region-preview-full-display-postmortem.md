---
Title: Window and region preview full-display bug postmortem
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - video
    - backend
    - x11
    - analysis
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/media/gst/preview.go
      Note: |-
        The original capture path relied on ximagesrc startx/starty/endx/endy for region and window sources.
        The fix path switches to full-root X11 capture plus explicit videocrop based on resolved geometry.
    - Path: pkg/media/gst/shared_video.go
      Note: |-
        Shared preview consumers exposed the bug visually and also had a separate scaling/aspect-ratio issue that initially obscured the root cause.
    - Path: pkg/discovery/service.go
      Note: |-
        Root display geometry is needed to convert absolute region/window rectangles into explicit videocrop margins.
    - Path: ui/src/components/preview/PreviewStream.tsx
      Note: |-
        The frontend was investigated as a possible source of the bug, but the final evidence pointed to backend capture content rather than card wiring.
ExternalSources: []
Summary: Postmortem of the bug where selected window and predefined region previews still looked like the full display, including the false leads, the standalone ximagesrc experiments, the actual root cause, and the rationale for switching to full-root capture plus videocrop.
LastUpdated: 2026-04-13T22:40:00-04:00
WhatFor: Explain the complete debugging journey and capture the lessons needed to maintain or extend X11/GStreamer preview capture safely.
WhenToUse: Read when debugging incorrect X11 region/window capture, evaluating ximagesrc semantics, or deciding whether to keep GStreamer capture or replace it with custom X11 bindings.
---

# Window and region preview full-display bug postmortem

## Executive Summary

The user reported a severe correctness bug: choosing a **window** or a predefined **region** in the Screencast Studio UI still resulted in a preview that looked like the **full display**. At first glance, this looked like a frontend bug, a preview-card wiring bug, or a DSL-generation bug. The investigation proved otherwise.

The UI was emitting correct source types, the backend was resolving correct window and region rectangles, and the preview manager was assigning distinct preview IDs. The real failure was deeper: on this machine, relying on `ximagesrc` region coordinates (`startx`, `starty`, `endx`, `endy`) did not reliably produce the selected pixels. Instead, it could produce a frame with the *requested dimensions* while still showing the *full desktop content squeezed into that frame*.

The correct fix direction was therefore not “change the UI labels” or “change the preview card,” but “stop trusting ximagesrc coordinate cropping for correctness-sensitive region/window capture.” The recommended runtime model is to capture the full X11 root and crop explicitly with `videocrop` using the resolved absolute rectangle.

## Problem Statement

Observed behavior:

- the user selected a **window** source,
- or selected a **predefined region** such as `Top Half` or `Bottom Half`,
- but the preview still visually resembled the **full display**.

This violated the most basic expectation of the setup UI: that the preview image should show the actual selected source.

## Initial Hypotheses

The bug initially had several plausible explanations.

### Hypothesis 1: the UI was creating the wrong source kind

Maybe the source picker was presenting “Window” or “Region” choices, but the structured draft or DSL renderer was silently emitting a full-display source.

### Hypothesis 2: the backend was dropping `window_id` or `rect`

Maybe the DSL was correct, but normalization or runtime source construction was collapsing the effective source onto the root display.

### Hypothesis 3: the preview cards were wired to the wrong MJPEG stream

Maybe the backend had distinct previews, but the UI card for the region/window source was accidentally rendering the full-display preview URL.

### Hypothesis 4: the GStreamer preview branch was distorting the image enough to masquerade as full-screen

This turned out to be partly true, but only as a secondary issue.

## Investigation Timeline

## 1. Verified frontend source creation and raw DSL

The first task was to determine whether the browser was even asking for the right thing.

### Findings

The Raw DSL tab showed correct entries for window sources, including:

- `type: "window"`
- a real `window_id`

And region sources showed:

- `type: "region"`
- a real `rect`

This ruled out the simplest “the UI is sending the wrong DSL” explanation.

## 2. Verified backend window/region resolution in live logs

Next, the live `tmux` logs were inspected while the user-facing app created previews.

### Findings

For a window source, the backend logged a resolved geometry such as:

```text
resolved window preview geometry for region-style capture ... x=6 y=30 width=1431 height=1884
```

For a region source, the shared-source signature included a concrete rect such as:

```text
type=region|display=:0.0|...|rect=0,0,2880,960
```

This proved that the backend was *not* losing the selected rectangle.

## 3. Verified distinct preview IDs and stream URLs in the browser

The next suspicion was that the UI might be showing the wrong stream on the wrong card.

### Findings

The DOM inspection showed separate MJPEG URLs, for example:

- full desktop → `/api/previews/preview-18daabfb09b6/mjpeg`
- region → `/api/previews/preview-eca7f98f9b75/mjpeg`

So the cards were not trivially wired to the same backend preview ID.

## 4. Found a secondary bug: preview scaling / aspect-ratio distortion

Direct screenshot pulls from `/api/previews/{id}/screenshot` showed clearly wrong dimensions such as:

- desktop preview → `960x1920`
- region preview → `1280x1920`

This exposed a separate bug in the shared preview consumer path: it only constrained width and let GStreamer renegotiate the rest, which produced distorted portrait-like results.

### Fix applied

The preview branch was changed to preserve aspect ratio by explicitly calculating width and height when source dimensions were known.

### Outcome

This improved output dimensions substantially, but the underlying “still shows full display content” complaint remained.

That was an important lesson: not every visual symptom was caused by the same bug.

## 5. Ran standalone ximagesrc experiments outside the app

This was the decisive step.

Instead of continuing to reason from app behavior, standalone `gst-launch-1.0` pipelines were used.

### Experiment A: ximagesrc with direct region coordinates

A bottom-half capture was requested directly from `ximagesrc` using `startx/starty/endx/endy`.

Representative command:

```bash
gst-launch-1.0 -q \
  ximagesrc use-damage=false startx=0 starty=960 endx=2879 endy=1919 num-buffers=1 \
  ! videoconvert ! jpegenc ! multifilesink location=/tmp/gst-region-bottom-half.jpg
```

### Result

The file had the *expected dimensions* (`2880x960`), but visual inspection showed it still resembled the **full desktop squeezed into that aspect ratio**, not a true bottom-half crop.

That was the smoking gun: the app was not inventing the bug. `ximagesrc` region-coordinate capture itself was unreliable in this setup.

### Experiment B: full-root capture plus videocrop

A second standalone pipeline captured the full root image and cropped explicitly with `videocrop`.

Representative command:

```bash
gst-launch-1.0 -q \
  ximagesrc use-damage=false num-buffers=1 \
  ! videocrop top=960 bottom=0 left=0 right=0 \
  ! videoconvert ! jpegenc ! multifilesink location=/tmp/gst-region-bottom-half-videocrop.jpg
```

### Result

This also produced `2880x960`, but this time the image was a **true bottom-half crop**.

That experiment narrowed the root cause from “maybe a UI bug” to “do not trust ximagesrc coordinate cropping on this machine for correctness-sensitive region/window capture.”

## Root Cause

The root cause was a bad assumption in the runtime design:

> If `ximagesrc` is given `startx/starty/endx/endy`, it will reliably emit the selected sub-region pixels.

That assumption did not hold on this machine.

### More precise failure mode

`ximagesrc` could produce a frame whose **output dimensions matched the requested region**, while the **visual content still corresponded to the full display**. In practice, this meant the user saw a preview that looked like the whole desktop, merely squeezed into the geometry of the selected window or region.

### Why it was hard to diagnose

Several layers each looked “correct” in isolation:

- the source picker created the right source kind,
- the DSL looked correct,
- normalization preserved the target,
- backend logs reported correct rects,
- preview cards pointed to different MJPEG URLs,
- and the broken preview still had plausible dimensions.

The bug only became undeniable once the capture primitive itself was tested outside the app.

## Fixing Analysis

There were two main fix paths.

### Option A: keep ximagesrc coordinate cropping and keep debugging around it

This would mean trying to understand whether some exact combination of X11 setup, damage tracking, rotation, root-window semantics, or GStreamer negotiation was causing the unexpected behavior.

Pros:

- smaller theoretical code diff if a single knob fixed it

Cons:

- low confidence
- fragile machine-specific behavior
- more time spent reverse-engineering ximagesrc semantics instead of delivering a stable preview

### Option B: stop trusting ximagesrc coordinate cropping and crop explicitly

This means:

1. capture the full X11 root image,
2. resolve absolute region/window rects,
3. convert those rects into `videocrop` margins using the real root display size,
4. let downstream preview/recording branches consume the explicitly cropped frames.

Pros:

- aligns behavior with the successful standalone experiment
- easier to reason about than implicit source-element cropping semantics
- preserves the rest of the GStreamer runtime architecture

Cons:

- more pixels are captured before cropping
- slightly more work per frame
- requires reliable root-geometry discovery

### Recommendation

Use **Option B**. It is the most robust fix while keeping the existing GStreamer-based runtime intact.

## Implementation Plan

### Phase 1: add root-display geometry discovery

Target file:

- `pkg/discovery/service.go`

Plan:

- add a helper to query root window width/height for a given X11 display
- allow display-specific command environment when needed

### Phase 2: change region/window capture construction

Target file:

- `pkg/media/gst/preview.go`

Plan:

- for `display`, keep plain root capture
- for `region` and `window`, resolve or reuse absolute rects
- capture the full root with `ximagesrc`
- apply `videocrop` using margins derived from root size and target rect

### Phase 3: validate both app and standalone behavior

Validation should include:

- standalone `gst-launch-1.0` proof for root+videocrop
- live preview screenshot pulls from `/api/previews/{id}/screenshot`
- visual browser verification that `Full Desktop` and `Bottom Half` no longer look identical

## Alternatives Considered

### Alternative: replace capture with custom X11 bindings immediately

This is still a plausible long-term fallback, but it is not the first recommended fix.

Why not immediately:

- the broader preview/recording/shared-source architecture is already working well in GStreamer
- the problem is localized to the X11 capture primitive
- explicit `videocrop` already demonstrated a working path with much lower integration cost

### When custom X11 capture becomes justified

A switch to X11 bindings becomes reasonable if:

- root+videocrop still proves unreliable,
- per-monitor targeting remains painful,
- or `ximagesrc` continues to surprise us in correctness-critical ways.

## Validation Checklist

- `window` preview no longer visually duplicates the full desktop preview
- `top half` and `bottom half` previews clearly differ from each other and from the full desktop
- direct preview screenshots differ in both dimensions and content
- live logs show explicit crop configuration for region/window sources
- no regressions in full-display preview or camera preview paths

## Lessons Learned

### 1. Correct metadata does not imply correct pixels

The logs can show the right rect and the runtime can still display the wrong content. Pixel-level validation matters.

### 2. Standalone experiments are worth the time

The `gst-launch-1.0` experiments were the turning point. Without them, it would have been easy to keep blaming the UI.

### 3. Separate multiple visual bugs instead of merging them mentally

The aspect-ratio/scaling bug and the wrong-content bug were both real, but they were not the same problem.

### 4. Prefer explicit transforms over implicit source semantics when correctness matters

If a region/window preview must be exact, explicit cropping is easier to trust than “whatever the source element means by region capture.”

## References

- `pkg/media/gst/preview.go`
- `pkg/media/gst/shared_video.go`
- `pkg/discovery/service.go`
- standalone `gst-launch-1.0` experiments captured during SCS-0014 work
- live tmux logs from `scs-web-ui`
- user screenshot: `/tmp/pi-clipboard-76a216ad-65d9-4220-a1f6-b2ba9ba89046.png`
