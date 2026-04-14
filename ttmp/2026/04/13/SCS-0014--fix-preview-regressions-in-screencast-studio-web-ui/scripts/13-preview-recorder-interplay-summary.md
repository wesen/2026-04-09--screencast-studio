---
Title: Preview recorder interplay measurements summary
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
Summary: Summary of the standalone benchmark that measures preview-only, recorder-only, and combined preview-plus-recorder cost on the same shared source shape.
LastUpdated: 2026-04-14T00:10:00-04:00
WhatFor: Explain the saved benchmark results for the interaction between preview work and recording work.
WhenToUse: Read when investigating why preview plus recording is much more expensive than recorder-only.
---

# Preview recorder interplay measurements summary

This note summarizes the standalone preview/recorder interplay benchmark captured under:

- `12-go-preview-recorder-interplay-matrix/results/20260414-070646/`

## Goal

Measure how much extra CPU comes from keeping the preview branch alive while recording is running, using a standalone shared-source benchmark that is closer to the real current architecture than the earlier isolated bridge decomposition.

## Scenarios

The benchmark used the real `2880x960 @ 24 fps` bottom-half region shape and measured:

1. `preview-current-only`
2. `recorder-current-only`
3. `preview-current-plus-recorder`
4. `preview-cheap-plus-recorder`

The `preview-current-*` scenarios approximate the current region preview settings:

- width `1280`
- fps `10`
- JPEG quality `80`

The `preview-cheap-plus-recorder` scenario approximates a cheaper preview profile:

- width `640`
- fps `5`
- JPEG quality `50`

## Results

Source: `12-go-preview-recorder-interplay-matrix/results/20260414-070646/01-summary.md`

- preview-only: **12.17% avg CPU**, **14.00% max CPU**
- recorder-only: **94.00% avg CPU**, **120.00% max CPU**
- current preview + recorder: **188.43% avg CPU**, **492.00% max CPU**
- cheap preview + recorder: **170.00% avg CPU**, **427.00% max CPU**

Additional counters from the run:

- current preview-only copied about **17.3 MB** of JPEG bytes in 6 seconds
- current preview + recorder copied about **17.6 MB** of JPEG bytes while still pushing **144** recorder frames
- cheap preview + recorder dropped preview byte-copy load substantially to about **1.6 MB**, but CPU still stayed very high

## Main interpretation

This benchmark strongly suggests that the remaining elevated cost is really the **combined preview + recorder case**, not just recorder-only encoding.

Important observations:

1. Recorder-only remains in the same rough range as the reconciled earlier benchmarks (~94% avg CPU).
2. Preview-only by itself is relatively cheap (~12%).
3. But preview + recorder together is **far more expensive than recorder-only**, and noticeably worse than a naive “94 + 12” mental model.
4. Making the preview branch cheaper helps, but only partially:
   - current preview + recorder: **188.43%**
   - cheap preview + recorder: **170.00%**

That means the extra cost is not explained only by JPEG byte-copy volume. The combined live branches appear to contend in a way that amplifies total CPU usage.

## Practical takeaway

The most important current performance conclusion is now:

> The major remaining cost spike is not simply the recorder bridge by itself. It is the **interaction of an active preview branch with the recording branch** on the same shared source.

That shifts the optimization target.

The next likely areas to investigate are:

- whether the shared source plus two live downstream branches are causing extra conversion/rate/scale work that can be reduced,
- whether the preview branch should become cheaper automatically while recording is active,
- whether preview JPEG generation should be suspended or degraded while recording on high-resolution regions,
- or whether some work can be shared more efficiently before the preview and recording branches diverge.

## Recommended next experiments

1. Compare preview + recorder with preview branch disabled dynamically after warmup.
2. Sweep preview profile settings more aggressively while recording:
   - width,
   - fps,
   - JPEG quality.
3. Check whether preview branch color conversion / scaling can be moved or simplified to reduce duplicated work.
4. Measure the same interplay on a smaller region to understand scaling behavior.
