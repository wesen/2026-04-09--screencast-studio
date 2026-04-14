---
Title: Live app adaptive preview runtime summary
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
ExternalSources: []
Summary: Summary of the live app-path before/after measurement for the adaptive preview runtime prototype.
LastUpdated: 2026-04-14T14:55:00-04:00
WhatFor: Capture whether the adaptive preview runtime change improved CPU in the real server/API path rather than only in standalone harnesses.
WhenToUse: Read when deciding whether the adaptive preview prototype is promising enough to keep pushing or whether more iteration is needed.
---

# Live app adaptive preview runtime summary

This note summarizes the first real app-path before/after measurement for the adaptive preview runtime prototype.

## Measurement method

The new script:

- `scripts/19-live-app-preview-recording-cpu-measure/run.sh`

runs a real `screencast-studio serve` process from a specified repo/revision, drives preview + recording through the HTTP API, measures the server process with `pidstat`, and saves screenshots plus `ffprobe` output.

I used it twice:

1. **before** — pre-adaptive revision `1554243058be6ecd73651a240fd1b7fc8272e286`
2. **after** — current adaptive-preview revision `e27ebdfa5bdda660bdd0caa00ee926e7c4c3435b`

Both runs used the same live server path and the same `2880x960 @ 24 fps` bottom-half region scenario.

## Results

### Before

From:

- `scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142838/01-summary.md`

Key result:

- avg CPU: **188.27%**
- max CPU: **329.00%**

### After

From:

- `scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142808/01-summary.md`

Key result:

- avg CPU: **170.82%**
- max CPU: **324.00%**

## Interpretation

This is a real improvement in the production server/API path.

Measured against the pre-adaptive revision, the adaptive-preview prototype reduced average server CPU by about:

- **17.45 percentage points absolute**
- about **9.3% relative** (`(188.27 - 170.82) / 188.27`)

That is not a complete solution, but it is strong enough to say the adaptive-preview direction is helping outside the synthetic harnesses.

## Important caveats

### 1. CPU is still high

The combined preview + recording case remains expensive even after the adaptive-preview change. The prototype improved the result, but did not make the problem disappear.

### 2. This script is strongest as a CPU comparison, not as a freeze detector

The saved screenshot hashes were not meant to be a definitive preview-freeze test. In both the before and after runs, the `pre` and `during` screenshots happened to hash the same, which could simply mean the visible scene content stayed stable at the sampled instant. For freeze diagnosis, the dedicated preview-freeze tooling remains better.

### 3. Preview quality acceptability still needs direct product judgment

The CPU comparison is useful, but it does not answer whether the constrained recording-time preview still looks good enough to users.

## Bottom line

The adaptive-preview runtime prototype is now supported by three layers of evidence:

1. standalone interaction benchmarks,
2. standalone adaptive confirmation / repeatability runs,
3. and now a real app-path before/after CPU improvement.

That means the prototype is worth keeping and iterating on rather than rolling back immediately.

## Recommended next step

The next most valuable step is now qualitative and product-facing:

- inspect the live preview during recording with the constrained profile,
- decide whether its quality is acceptable,
- and then decide whether the adaptive behavior should remain the default or become configurable.
