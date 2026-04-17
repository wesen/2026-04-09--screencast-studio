---
Title: 34 Go x264 CGO flag matrix summary
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
Summary: Focused Go-only x264enc encode-stage comparison across different CGO_CFLAGS settings.
LastUpdated: 2026-04-15T04:35:00-04:00
WhatFor: Preserve the build-flag control used to test whether the Go x264enc anomaly is sensitive to cgo C optimization levels.
WhenToUse: Read this when evaluating whether the x264enc anomaly is likely in thin cgo wrapper compilation or elsewhere in the Go-hosted process behavior.
---

# 34 Go x264 CGO flag matrix

| scenario | cgo_cflags | avg_cpu | max_cpu | page-faults | minor-faults | major-faults | result | error |
|---|---|---:|---:|---:|---:|---:|---|---|
| default |  | 254.33 | 413.00 | 284806 | 284806 | 0 | eos |  |
| cgo_o2 | -O2 | 349.47 | 527.00 | 262226 | 262226 | 0 | eos |  |
| cgo_o3 | -O3 | 272.02 | 444.00 | 283258 | 283258 | 0 | eos |  |
