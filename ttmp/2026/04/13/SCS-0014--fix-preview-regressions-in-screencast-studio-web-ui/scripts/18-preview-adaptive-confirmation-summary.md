---
Title: Preview adaptive confirmation summary
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - performance
    - preview
    - recording
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources:
    - local:preview-and-recording-performance-improvement-report.md
    - local:preview-and-recording-performance-improvement-diary.md
Summary: Summary of standalone confirmation experiments for the imported adaptive-preview mitigation proposal, run independently of the main runtime.
LastUpdated: 2026-04-14T14:25:00-04:00
WhatFor: Capture what the isolated confirmation experiments actually showed about constrained preview profiles and rate-first preview ordering.
WhenToUse: Read when deciding whether to prototype adaptive preview behavior in the real runtime or when reviewing the imported research notes.
---

# Preview adaptive confirmation summary

This note summarizes two new standalone experiment families that were run to test the imported mitigation proposal independently of the main `screencast-studio` runtime.

## Goal

The imported report proposed two changes for the preview-plus-recording CPU spike:

1. constrain preview while recording is active, and
2. reorder the preview branch so `videorate` drops surplus frames before `videoscale` and `jpegenc`.

The goal of these experiments was not to trust those notes blindly. It was to reproduce the claim in isolation on the real capture machine with a standalone harness.

## New standalone experiments

### 16: direct confirmation matrix

Saved under:

- `scripts/16-go-preview-adaptive-confirmation-matrix/`

This harness tested the actual preview path with JPEG + Go-side byte copy, while keeping the recorder path fixed.

Scenarios:

- `recorder-only`
- `preview-scale-first-current-plus-recorder`
- `preview-rate-first-current-plus-recorder`
- `preview-scale-first-constrained-plus-recorder`
- `preview-rate-first-constrained-plus-recorder`

The constrained profile matched the imported proposal for screen-like sources:

- `640 max width`
- `4 fps`
- `jpeg quality 50`

The first single run was useful but noisy. It suggested the combined direction was promising, but it was not strong enough by itself to treat as final.

### 17: repeatability wrapper

Saved under:

- `scripts/17-go-preview-adaptive-repeatability-matrix/`

This wrapper re-ran the standalone harness across three rounds with alternating scenario order to reduce single-run noise.

That result is the better basis for interpretation.

## Most useful repeated results

Source: `scripts/17-go-preview-adaptive-repeatability-matrix/results/20260414-135308/01-summary.md`

Average CPU means across three rounds:

- `recorder-only`: **91.00%**
- `preview-scale-first-current-plus-recorder`: **108.28%**
- `preview-rate-first-current-plus-recorder`: **106.45%**
- `preview-scale-first-constrained-plus-recorder`: **104.95%**
- `preview-rate-first-constrained-plus-recorder`: **87.61%**

## What this says

### 1. The imported direction is broadly supported

The best repeated result was the **combined imported mitigation direction**:

- constrained recording-time preview profile
- plus rate-first preview ordering

That was the only preview+recorder variant that landed essentially in the same band as recorder-only.

### 2. Rate-first alone looks modest, not revolutionary

Comparing the two current-profile cases:

- scale-first current: **108.28%**
- rate-first current: **106.45%**

That is directionally better, but only modestly so in the repeated results.

So the imported claim that ordering matters is still plausible, but the more important practical effect is probably not ordering by itself.

### 3. Constrained profile alone did not look as convincing as the combined fix

Comparing:

- scale-first constrained: **104.95%**
- rate-first constrained: **87.61%**

This suggests the constrained profile by itself is not the full answer in this standalone setup. The strongest effect appears when the constrained profile is paired with rate-first ordering.

### 4. The combined adaptive variant is the one worth prototyping in the real runtime

The repeated standalone evidence does **not** say that every part of the imported proposal wins strongly in isolation.

What it does say is:

> the combined adaptive-preview direction is promising enough to justify a real runtime prototype.

## Important caveats

### Run-to-run variance is real

Even with repeated runs, some scenarios still show noticeable spread. This work remains benchmark-driven rather than mathematically clean.

### These are still standalone harnesses

They are intentionally independent of the production runtime. That is good for isolation, but it also means the next validation step still needs to happen in the real runtime path.

## Practical conclusion

The imported research helped in a useful way.

After standalone confirmation, the best current reading is:

- **do not** blindly assume rate-first ordering alone solves the issue,
- **do not** treat a constrained profile alone as fully proven,
- **do** treat the **combined adaptive-preview approach** as the strongest next production-side prototype.

## Recommended next step

Prototype these two changes together in the real runtime:

1. when recording attaches to a shared source, switch preview to a constrained recording-time profile,
2. build the preview branch as `queue -> videorate -> fps caps -> videoscale -> size caps -> jpegenc -> appsink`.

Then validate in both ways:

- rerun the live-app CPU measurements,
- and confirm that preview quality during recording is still acceptable.
