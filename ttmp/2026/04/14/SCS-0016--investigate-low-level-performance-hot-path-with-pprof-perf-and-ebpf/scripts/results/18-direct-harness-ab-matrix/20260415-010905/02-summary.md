---
Title: 18 direct harness ab matrix
Ticket: SCS-0016
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - backend
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Ticket-local result summary artifact for SCS-0016.
LastUpdated: 2026-04-15T01:45:00-04:00
WhatFor: Preserve this local result summary so the SCS-0016 evidence trail stays reproducible and reviewable.
WhenToUse: Read when reviewing or comparing this specific saved result inside SCS-0016.
---

# 18 direct harness ab matrix

- repeats: 3 copied, 3 manual
- duration_seconds: 8

| harness | runs | avg_cpu_mean | avg_cpu_min | avg_cpu_max | max_cpu_peak |
|---|---:|---:|---:|---:|---:|
| 15-copied-direct | 3 | 381.29 | 378.86 | 382.56 | 633.00 |
| 17-manual-direct | 3 | 327.73 | 269.47 | 378.17 | 547.00 |

## Interpretation

- Both Go-hosted full-desktop direct pipelines remain much hotter than the matched `gst-launch` control.
- The copied/app-like direct harness is consistently hotter than the fully manual Go graph in this run set.
- The remaining gap after fixing the accidental `Y444` negotiation suggests there is still additional overhead or behavioral difference beyond just the raw-format caps mistake.
- The `I420` fix was still correct and should be kept because the realized graph now matches the intended raw format into `x264enc`.
