---
Title: Go bridge overhead measurements summary
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - video
    - performance
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Summary of the staged standalone benchmark that isolates the raw appsink, Go callback/copy, async queue, appsrc bridge, and x264 stages.
LastUpdated: 2026-04-13T23:35:00-04:00
WhatFor: Explain what the staged bridge-overhead benchmark suggests about where recording CPU is actually being spent.
WhenToUse: Read when investigating whether the current shared raw-consumer to Go to appsrc bridge is itself the main CPU problem.
---

# Go bridge overhead measurements summary

This note summarizes the staged standalone bridge-overhead benchmark captured under:

- `09-go-bridge-overhead-matrix/results/20260413-232943/`

## Goal

Measure the overhead of the path that looks like:

```text
normalized raw source -> appsink -> Go callback/copy/queue -> appsrc -> downstream sink
```

The purpose was to stop inferring bridge overhead indirectly and instead test the stages one by one.

## Tested scenarios

Using the real `2880x960 @ 24 fps` bottom-half region shape, the benchmark measured:

1. `normalized-fakesink`
2. `appsink-discard`
3. `appsink-copy-discard`
4. `appsink-copy-async-discard`
5. `appsink-copy-async-appsrc-fakesink`
6. `appsink-copy-async-appsrc-x264`

## Results

Source: `09-go-bridge-overhead-matrix/results/20260413-232943/01-summary.md`

- normalized raw pipeline to `fakesink`: **24.50% avg CPU**, **26.00% max CPU**
- `appsink` discard: **25.33% avg CPU**, **30.00% max CPU**
- `appsink` + `buffer.Copy()` discard: **28.33% avg CPU**, **32.00% max CPU**
- `appsink` + copy + async queue discard: **25.17% avg CPU**, **28.00% max CPU**
- `appsink` + copy + async queue + `appsrc -> fakesink`: **24.33% avg CPU**, **29.00% max CPU**
- `appsink` + copy + async queue + `appsrc -> x264`: **77.83% avg CPU**, **99.00% max CPU**

## Main interpretation

This benchmark suggests that the standalone **Go bridge by itself is not the dominant cost**.

The important observation is that all of these cases remain clustered near the normalized baseline until x264 is introduced:

- raw normalized baseline: ~24.5%
- appsink callback: ~25.3%
- copy in Go: ~28.3%
- copy + async queue: ~25.2%
- copy + async queue + appsrc -> fakesink: ~24.3%

The large jump only appears in the `appsrc -> x264` case.

That implies:

1. `appsink` callback overhead is small relative to the overall workload for this capture shape.
2. `buffer.Copy()` adds some cost, but not enough to explain the full recording spike on its own.
3. The async queue and `appsrc -> fakesink` path do not look catastrophically expensive in isolation.
4. The encoder is still the main cost center in this staged benchmark.

## Important nuance

These staged results do **not** fully match the earlier shared-runtime benchmark from `07-go-shared-recording-performance-matrix`, where recorder-only measured much higher CPU (`139.57% avg CPU`).

That discrepancy means one of the following is true:

- the full shared runtime is adding substantial overhead beyond the isolated staged bridge,
- the earlier benchmark captured more than the encoder cost alone,
- or the two benchmarks are measuring slightly different pipeline shapes/behaviors.

So the new result is **not** “the bridge is free.” It is narrower:

> in this standalone staged benchmark, the bridge machinery does not appear to be the dominant CPU problem until x264 is present.

## Practical takeaway

The current best interpretation is:

- the recording spike is still heavily tied to **x264**,
- the pure direct encode benchmark already proved encoder settings matter a lot,
- and this staged benchmark suggests the **bridge alone is probably not the whole story**.

That means the next useful step is to reconcile this staged benchmark with the earlier full shared-runtime benchmark, not to assume one or the other is the final truth.

## Recommended next steps

1. Re-run the earlier full shared-runtime benchmark (`07-go-shared-recording-performance-matrix`) and compare it directly against this staged bridge matrix.
2. Add x264 preset variations to the staged matrix so we can quantify how much of the `77.83%` is encoder tuning versus bridge cost.
3. Compare the staged benchmark against the exact production path in `pkg/media/gst/shared_video_recording_bridge.go` and note any differences in queueing, caps negotiation, or extra logging.
4. Measure whether the full shared runtime introduces extra cost through tee management, registry lifecycle, or preview-related plumbing even when no preview consumer is attached.
