---
Title: 17 preview adaptive repeatability matrix run summary
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
Summary: Repeated standalone confirmation runs for adaptive preview mitigation scenarios to reduce single-run noise.
LastUpdated: 2026-04-14T14:15:00-04:00
WhatFor: Preserve repeated CPU measurements for preview-profile and preview-ordering mitigation scenarios independent of the main runtime.
WhenToUse: Read when deciding whether the adaptive preview mitigation direction remains plausible after repeatability checks.
---

# 17 preview adaptive repeatability matrix

## Per-scenario aggregates

## recorder-only
- runs: 3
- avg_cpu_mean: 91.00%
- avg_cpu_min: 87.83%
- avg_cpu_max: 95.67%
- avg_cpu_stdev: 3.37
- max_cpu_mean: 128.33%
- run_details:
  - round1: avg=95.67% max=140.00%
  - round2: avg=87.83% max=115.00%
  - round3: avg=89.50% max=130.00%

## preview-scale-first-current-plus-recorder
- runs: 3
- avg_cpu_mean: 108.28%
- avg_cpu_min: 101.83%
- avg_cpu_max: 116.00%
- avg_cpu_stdev: 5.85
- max_cpu_mean: 143.33%
- run_details:
  - round1: avg=116.00% max=147.00%
  - round2: avg=101.83% max=131.00%
  - round3: avg=107.00% max=152.00%

## preview-rate-first-current-plus-recorder
- runs: 3
- avg_cpu_mean: 106.45%
- avg_cpu_min: 97.17%
- avg_cpu_max: 124.00%
- avg_cpu_stdev: 12.42
- max_cpu_mean: 149.67%
- run_details:
  - round1: avg=124.00% max=194.00%
  - round2: avg=98.17% max=127.00%
  - round3: avg=97.17% max=128.00%

## preview-scale-first-constrained-plus-recorder
- runs: 3
- avg_cpu_mean: 104.95%
- avg_cpu_min: 88.50%
- avg_cpu_max: 121.67%
- avg_cpu_stdev: 13.54
- max_cpu_mean: 148.33%
- run_details:
  - round1: avg=121.67% max=174.00%
  - round2: avg=88.50% max=111.00%
  - round3: avg=104.67% max=160.00%

## preview-rate-first-constrained-plus-recorder
- runs: 3
- avg_cpu_mean: 87.61%
- avg_cpu_min: 75.00%
- avg_cpu_max: 98.33%
- avg_cpu_stdev: 9.62
- max_cpu_mean: 113.33%
- run_details:
  - round1: avg=89.50% max=116.00%
  - round2: avg=98.33% max=133.00%
  - round3: avg=75.00% max=91.00%

## Raw CSV

```csv
round,scenario,avg_cpu,max_cpu
round1,recorder-only,95.67,140.00
round1,preview-scale-first-current-plus-recorder,116.00,147.00
round1,preview-rate-first-current-plus-recorder,124.00,194.00
round1,preview-scale-first-constrained-plus-recorder,121.67,174.00
round1,preview-rate-first-constrained-plus-recorder,89.50,116.00
round2,recorder-only,87.83,115.00
round2,preview-rate-first-current-plus-recorder,98.17,127.00
round2,preview-scale-first-current-plus-recorder,101.83,131.00
round2,preview-rate-first-constrained-plus-recorder,98.33,133.00
round2,preview-scale-first-constrained-plus-recorder,88.50,111.00
round3,recorder-only,89.50,130.00
round3,preview-scale-first-current-plus-recorder,107.00,152.00
round3,preview-rate-first-current-plus-recorder,97.17,128.00
round3,preview-scale-first-constrained-plus-recorder,104.67,160.00
round3,preview-rate-first-constrained-plus-recorder,75.00,91.00
```
