---
Title: Webcam preview quality and format regression bug report
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
    - Path: pkg/dsl/normalize.go
      Note: Camera defaults preserve source capture size, but preview rendering currently ignores user-facing expectations about fidelity
    - Path: pkg/media/gst/shared_video.go
      Note: |-
        Preview branch currently hard-clamps preview width, frame rate, and JPEG quality
        Hard-codes preview width
ExternalSources: []
Summary: Bug report for the webcam preview looking compressed and lower quality than expected because the current preview branch forces 640px width, 5 fps, and JPEG quality 50 for all preview streams.
LastUpdated: 2026-04-13T22:00:00-04:00
WhatFor: Explain why the webcam preview looked visibly degraded during manual testing and recommend a fix path.
WhenToUse: Read when improving preview fidelity or investigating why webcam previews look compressed in the web UI.
---


# Webcam preview quality and format regression bug report

## Executive Summary

The webcam preview in the running web UI looked visibly compressed and lower quality than expected. This was not just subjective perception. The current preview branch in the shared GStreamer source forces all previews through a low-fidelity transform:

- scale to width 640,
- reduce to 5 fps,
- encode as JPEG with quality 50.

That design may have been acceptable as a conservative early preview path, but for camera sources it produces a clearly degraded image that looks much worse than the source settings shown in the UI. The bug is therefore an intentional but now user-visible quality regression: the preview path is too aggressively downsampled for the current product expectations.

## Problem Statement

The user reported that the webcam “format was off” and “looked compressed.” The screenshot shows a visibly blocky, low-fidelity webcam preview. The current code confirms that the preview branch is not trying to preserve camera fidelity; it is explicitly optimizing for a small low-bandwidth preview.

The problem is not that the preview is technically broken. The problem is that the preview quality is poor enough to undermine trust in the setup.

## Evidence

### Code evidence: hard-coded preview fidelity

In `pkg/media/gst/shared_video.go:449-495`, every preview consumer applies the same transformation chain:

- `videoscale`
- `capsfilter("video/x-raw,width=640")`
- `videorate`
- `capsfilter("video/x-raw,framerate=5/1")`
- `jpegenc`
- `jpegenc.Set("quality", 50)`

Key lines:

- `capsScale, err := newCapsFilter("video/x-raw,width=640")`
- `capsRate, err := newCapsFilter("video/x-raw,framerate=5/1")`
- `jpegenc.Set("quality", 50)`

This directly explains the compressed appearance.

### UI/context evidence

The default camera settings in the editor are much higher quality than the preview branch preserves. The setup defaults include camera size `1280x720` and quality `80` (`ui/src/features/editor/editorSlice.ts` default DSL and `pkg/dsl/normalize.go` defaulting behavior). A user looking at the setup would reasonably expect the preview to look closer to that source fidelity than to a 640px/5fps JPEG preview.

## Likely Root Cause

This is a mismatch between an early implementation tradeoff and current user expectations.

### Why the current behavior exists

The preview path likely chose:

- low resolution,
- low frame rate,
- and medium-low JPEG quality

for good reasons early on:

- smaller payloads,
- less CPU,
- fewer browser/network issues,
- simpler MJPEG preview transport.

### Why it is now a bug

The app is using live previews as a primary setup tool, not as an incidental debug stream. For camera framing, focus, lighting, and mirror checks, the preview needs to be “good enough to judge the source.” At 640px / 5fps / JPEG quality 50, it no longer reliably meets that standard.

## Fixing Analysis

There are two main ways to improve this.

### Option A: raise the existing hard-coded preview settings

Examples:

- width 960 or 1280 instead of 640
- 10 fps instead of 5 fps
- JPEG quality 75 or 85 instead of 50

Pros:

- smallest code change
- preserves current transport shape
- likely enough to remove the “looks broken” complaint

Cons:

- still one-size-fits-all
- still not source-aware

### Option B: make preview fidelity source-aware and/or configurable

Examples:

- camera previews use better settings than display previews
- preview settings derive from capture size/output quality with safe caps
- browser/UI may later expose a "preview quality" preference

Pros:

- better product fit
- avoids punishing webcam previews just because display previews need conservative defaults

Cons:

- more behavior to design and test

### Recommendation

Do a two-step fix:

1. **Short-term:** raise the defaults enough that webcam previews stop looking broken.
2. **Medium-term:** make preview fidelity source-aware, especially for cameras.

## Implementation Plan

### Phase 1: Raise preview quality defaults

Target file:

- `pkg/media/gst/shared_video.go`

Suggested changes:

- increase preview scale width from `640` to at least `960`
- increase preview frame rate from `5/1` to `10/1`
- increase JPEG quality from `50` to `75` or `80`

Validation:

- webcam preview should look visibly sharper in the browser
- CPU/network impact should remain acceptable during normal multi-preview usage

### Phase 2: Make preview settings source-aware

Target file:

- `pkg/media/gst/shared_video.go`
- possibly a new helper like `previewCapsForSource(...)`

Suggested logic:

- display previews can stay relatively conservative
- camera previews should preserve more detail and motion
- source capture size can inform the preview scaling policy

Example policy:

```go
if source.Type == "camera" {
    width = min(requestedWidth, 960)
    fps = min(source.Capture.FPS, 10)
    jpegQuality = 80
} else {
    width = 640
    fps = 5
    jpegQuality = 60
}
```

### Phase 3: Add a preview-quality regression harness

Add a ticket-local test or script that at least verifies the negotiated preview path configuration for camera sources and makes future fidelity regressions easy to catch in review.

## Alternatives Considered

### Alternative: leave preview low-fidelity and explain it in the UI

Rejected. The user already interpreted it as broken, not just low quality. Explanation text will not solve that product problem.

### Alternative: switch preview away from JPEG entirely right now

Too large for this bug ticket. It may be worth exploring later, but the current issue can be improved substantially without changing the transport model.

## Validation Checklist

- camera preview is visibly less blocky than before
- preview motion is smoother than 5 fps
- preview still streams reliably in the browser
- enabling multiple previews does not create unacceptable CPU or latency spikes

## References

- `pkg/media/gst/shared_video.go:449-495`
- `pkg/dsl/normalize.go:95-162`
- `ui/src/features/editor/editorSlice.ts`
