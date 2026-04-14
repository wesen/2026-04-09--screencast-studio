---
Title: Webcam preview remove-readd device churn and duplicate camera source bug report
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
    - Path: pkg/discovery/service.go
      Note: |-
        Camera discovery currently exposes every /dev/video* node directly as a separate camera choice
        Enumerates every /dev/videoN node directly
    - Path: pkg/media/gst/shared_video.go
      Note: |-
        Shared preview source identity includes the device node, so rebinding from /dev/video0 to /dev/video1 changes the capture signature
        Uses raw device node in the shared preview signature
    - Path: ui/src/features/setup-draft/conversion.ts
      Note: |-
        Source picker generates duplicate logical camera sources by name with suffixed IDs
        Creates duplicate logical camera sources with suffixed IDs for the same physical label
ExternalSources: []
Summary: Bug report for the webcam preview failing to come back cleanly after remove/re-add operations, driven by duplicate physical-camera discovery entries, unstable /dev/videoN rebinding, and duplicate logical camera sources in the UI.
LastUpdated: 2026-04-13T22:00:00-04:00
WhatFor: Capture the observed webcam re-add failure, the likely root causes, and a concrete fix plan.
WhenToUse: Read when fixing webcam source stability, duplicate camera choices, or source re-add behavior in the web UI.
---


# Webcam preview remove-readd device churn and duplicate camera source bug report

## Executive Summary

During live testing, the built-in webcam preview could be removed and then fail to come back cleanly. The logs show that the same physical laptop camera was sometimes captured as `/dev/video0` and then later as `/dev/video1`, and the UI simultaneously allowed multiple logical sources for that same camera with suffixed source IDs such as `...-2` and `...-3`. The result is not one simple failure but a cluster of related identity problems: unstable device-node selection, duplicate user-visible camera choices, duplicate logical source entries, and confusing preview recovery after rapid add/remove actions.

The recommended fix is to treat camera identity at a more stable level than raw `/dev/videoN`, deduplicate camera choices at discovery and/or UI levels, and make preview recovery deterministic even when the underlying V4L2 node order changes.

## Problem Statement

Observed user-facing behavior:

- a webcam source was added,
- then removed,
- then added again,
- and eventually "didn't come back" in a stable or trustworthy way.

The screenshot also showed multiple logical sources referring to the same physical camera, and the planned outputs section implied duplicate output paths for that same camera name.

This is a serious product bug because the user is reasoning about “my laptop camera” as one source, while the backend and UI currently expose multiple unstable identities for it.

## Evidence

### Runtime evidence from live server logs

The live `tmux` logs showed the camera first using `/dev/video0`:

```text
created shared gstreamer video source ... signature=type=camera|...|device=/dev/video0|... source_id=laptop-camera-...
```

and then, after release/re-add, using `/dev/video1` for what appears to be the same hardware label:

```text
2026-04-13T21:33:06.827536492-04:00 INF created shared gstreamer video source ... device=/dev/video1 ... source_id=laptop-camera-...
```

The current discovery snapshot also reports both nodes as separate cameras with the same label:

```json
{
  "id": "/dev/video0",
  "label": "Laptop Camera: Laptop Camera (usb-0000:00:14.0-7)"
}
{
  "id": "/dev/video1",
  "label": "Laptop Camera: Laptop Camera (usb-0000:00:14.0-7)"
}
```

### Discovery code evidence

`ListCameras()` appends every `/dev/video*` node directly to the camera list without deduplication by physical card or capability (`pkg/discovery/service.go:201-233`).

Key lines:

- `device := strings.TrimSpace(line)`
- `if !strings.HasPrefix(device, "/dev/video") { continue }`
- `cameras = append(cameras, Camera{ ID: device, Label: currentLabel, Device: device, CardName: currentCard })`

That means `/dev/video0` and `/dev/video1` for the same laptop camera are presented as two separate choices.

### UI source-creation evidence

The source picker creates a camera source using the selected camera's device ID but generates the source ID only from the camera label/name (`ui/src/features/setup-draft/conversion.ts:275-290`).

Key lines:

- `const name = camera.label || camera.device || 'Camera'`
- `id: nextUniqueSourceId(slugify(name), draft.videoSources)`
- `target: { deviceId: camera.device }`

This allows repeated additions of the same physical camera to become multiple logical sources with suffixed IDs like:

- `laptop-camera-...`
- `laptop-camera-...-2`
- `laptop-camera-...-3`

### Shared-source identity evidence

The shared preview source signature includes the raw device node (`pkg/media/gst/shared_video.go:619-635`):

```go
"device=" + strings.TrimSpace(source.Target.Device)
```

So `/dev/video0` and `/dev/video1` are treated as distinct capture identities even if they belong to the same physical camera.

## Likely Root Cause

This is an identity-model bug spanning three layers.

### Layer 1: discovery identity is too raw

The discovery layer equates “camera” with “every `/dev/videoN` node” rather than “the physical camera the user thinks they selected.” Many webcams expose multiple video nodes; not all should become separate user-facing sources.

### Layer 2: UI source creation allows duplicate logical sources

The UI permits repeated additions of the same physical camera, and because source IDs are generated from the label, the same camera becomes multiple suffixed logical sources instead of being recognized as already present.

### Layer 3: preview source identity changes when the device node changes

Even if the user thinks they are re-adding the same camera, the preview backend treats `/dev/video0` and `/dev/video1` as different shared sources. That makes preview recovery depend on unstable kernel/device-node ordering.

## Why this matters operationally

This is not just cosmetic confusion.

It can lead to:

- lost or stale previews,
- duplicate logical sources in the setup,
- duplicate planned outputs with the same human-readable camera name,
- and brittle behavior after unplug/replug, remove/re-add, or backend device re-enumeration.

## Fixing Analysis

The fix should not be a one-line patch in just one layer.

### Recommended fix shape

1. **Deduplicate discovery results** so the source picker does not present every raw `/dev/videoN` node as a separate end-user camera by default.
2. **Prevent accidental duplicate logical camera sources** in the structured editor unless the product explicitly wants multi-node or multi-format camera capture.
3. **Stabilize camera identity** using a more physical/human concept than raw `/dev/videoN` where possible.

### Good enough versus ideal

#### Good enough short-term

- Deduplicate discovery by `CardName`/label and prefer one canonical node.
- Block adding the same physical camera twice in the UI.
- Leave the backend signature device-based for the moment.

This would likely remove most user-visible confusion quickly.

#### Better medium-term

- Introduce a stable camera identity model with both:
  - `physical_id` (user-facing identity), and
  - `device_node` (runtime opening target)
- Update preview/recovery logic to survive node rebinding when the physical camera is unchanged.

This is a better long-term fix, but it touches more code paths.

### Recommendation

Do the short-term dedupe/blocking fix first, then decide whether a stronger physical-camera identity model is needed once the basic UX is stable.

## Implementation Plan

### Phase 1: Deduplicate discovery output

Target file:

- `pkg/discovery/service.go`

Plan:

- keep reading `v4l2-ctl --list-devices`
- group `/dev/video*` nodes by card/label
- choose one canonical node for the general camera list
- optionally retain alternates in metadata for future advanced use

Validation:

- `/api/discovery` should return one logical camera for the laptop camera instead of both `/dev/video0` and `/dev/video1`

### Phase 2: Prevent duplicate camera additions in the UI

Target file:

- `ui/src/features/setup-draft/conversion.ts`
- possibly `ui/src/pages/StudioPage.tsx` source picker behavior

Plan:

- before adding a camera source, check whether a source already targets that same camera/device/physical identity
- if yes, either:
  - focus the existing source, or
  - show a small UI warning instead of adding another duplicate

Validation:

- repeatedly picking the same laptop camera should not create `...-2` and `...-3` sources unless explicitly allowed by design

### Phase 3: Review output-path collision behavior

Target files:

- destination template handling in the UI defaults
- compile/output naming logic if needed

Plan:

- once duplicate camera sources are blocked, verify the default `{source_name}` output path is no longer colliding in common flows
- if duplicates remain allowed for some reason, the output naming model must include a stable disambiguator

Validation:

- planned outputs should not contain the same camera-named path twice in a standard setup

## Alternatives Considered

### Alternative: leave discovery alone and only block duplicates in the UI

This would help, but the raw API would still present confusing duplicate cameras and would likely keep surprising future consumers.

### Alternative: keep duplicate camera entries as an advanced feature

That is only defensible if the UI clearly distinguishes their capabilities and intended use. Right now it does not.

## Validation Checklist

- `/api/discovery` shows one logical laptop camera entry by default
- adding/removing/re-adding the laptop camera does not produce suffixed duplicate source IDs
- preview returns reliably after remove/re-add
- planned outputs do not show duplicate paths for the same camera in normal workflows

## References

- `pkg/discovery/service.go:201-233`
- `ui/src/features/setup-draft/conversion.ts:275-290`
- `pkg/media/gst/shared_video.go:619-635`
- live logs captured from the running `scs-web-ui` tmux session on 2026-04-13
