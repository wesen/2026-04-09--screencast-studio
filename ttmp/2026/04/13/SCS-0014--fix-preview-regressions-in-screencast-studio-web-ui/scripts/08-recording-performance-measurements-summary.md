---
Title: Recording performance measurements summary
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - video
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Human-readable summary of the standalone recording CPU measurements captured for SCS-0014.
LastUpdated: 2026-04-13T23:12:00-04:00
WhatFor: Explain the saved recording CPU benchmark results and their main conclusion.
WhenToUse: Read when you want the headline result before drilling into the raw per-run logs.
---

# Recording Performance Measurements Summary

This note summarizes the standalone recording-performance measurements captured for SCS-0014 without modifying the main server.

## Goal

Determine whether the large CPU jump observed when starting recording is primarily caused by:

- raw capture,
- preview JPEG generation,
- x264 encode settings,
- or the app's shared-source bridge topology.

## Saved benchmark harnesses

- `06-gst-recording-performance-matrix/`
  - pure `gst-launch-1.0` measurements
  - results: `06-gst-recording-performance-matrix/results/20260413-230721/`
- `07-go-shared-recording-performance-matrix/`
  - current shared-source Go bridge measurements
  - results: `07-go-shared-recording-performance-matrix/results/20260413-230755/`

## Test context

- Display: `:0`
- Root geometry: `2880x1920`
- Measured region: `0,960,2880,960` (bottom half of the desktop)
- Recording FPS: `24`
- Preview-like FPS: `10`

## Results

### 06: Pure GStreamer (`gst-launch-1.0`)

Source: `06-gst-recording-performance-matrix/results/20260413-230721/01-summary.md`

- capture to fakesink: **23.17% avg CPU**, **31.00% max CPU**
- preview-like JPEG pipeline: **8.67% avg CPU**, **10.00% max CPU**
- direct record with current x264 settings (`speed-preset=3`): **86.50% avg CPU**, **91.00% max CPU**
- direct record with faster x264 settings (`speed-preset=1`): **49.83% avg CPU**, **52.00% max CPU**

Interpretation:

- The encoder is the dominant CPU consumer in the pure pipeline.
- The current x264 preset used by the app is much more expensive than the faster preset.
- Preview JPEG work is comparatively cheap for this region.

### 07: Current shared Go bridge path

Source: `07-go-shared-recording-performance-matrix/results/20260413-230755/01-summary.md`

- preview only: **10.67% avg CPU**, **14.00% max CPU**
- recorder only: **139.57% avg CPU**, **237.00% max CPU**
- preview + recorder: **151.62% avg CPU**, **265.00% max CPU**

Interpretation:

- The shared bridge path is substantially more expensive than the pure direct-encode `gst-launch-1.0` pipeline.
- The recording start spike is not just “GStreamer is doing something”; it is especially tied to the current **shared raw consumer + appsrc bridge + x264 encode** setup.
- Preview adds some overhead, but the big jump comes from recording/encoding rather than preview generation.

## Current takeaway

The data strongly suggests that the recording CPU spike is real and is largely explained by the current encoding path:

1. **x264 settings matter a lot** even in pure GStreamer.
2. The initial benchmark suggested the **shared bridge path might add significant extra overhead** compared with a direct GStreamer pipeline.
3. That conclusion needed a reconciliation pass before treating it as final.

## Reconciliation update

A later same-session rerun compared the three relevant benchmark families directly under the same `2880x960 @ 24 fps` region shape:

- `06` direct record, current x264 preset: **94.33% avg CPU**
- `07` shared runtime recorder-only: **94.00% avg CPU**
- `09` staged `appsink -> Go -> appsrc -> x264`: **91.50% avg CPU**
- `07` shared runtime preview + recorder: **131.00% avg CPU**

That changed the interpretation.

The earlier very large gap between the full shared-runtime recorder-only benchmark and the staged bridge+x264 benchmark did **not** reproduce in the same-session reconciliation run. The more stable current interpretation is:

- recorder-only cost is largely aligned across direct encode, the real shared runtime, and the staged bridge+x264 case,
- so the bridge alone does **not** currently look like the dominant extra recorder-only cost,
- the remaining clearly expensive combined case is **preview + recorder together**.

That means the highest-confidence current priorities are:

1. continue tuning x264 presets/bitrates/FPS,
2. investigate why preview + recorder together rises well above recorder-only,
3. only then decide whether deeper bridge surgery is justified.

## Recommended next experiments

1. Add a shared-bridge benchmark matrix for multiple x264 presets/bitrates.
2. Measure whether lowering FPS from `24` to `15` or `10` meaningfully reduces CPU for large region captures.
3. Compare the current bridge path against a direct tee-to-encoder branch benchmark, even if only as an experiment, to estimate bridge-copy overhead.
4. Measure the cost of recording a smaller region versus the full `2880x960` bottom-half capture.
