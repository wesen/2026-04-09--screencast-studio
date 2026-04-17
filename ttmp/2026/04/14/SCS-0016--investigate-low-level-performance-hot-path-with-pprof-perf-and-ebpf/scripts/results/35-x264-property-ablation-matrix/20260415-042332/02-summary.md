---
Title: 35 x264 property ablation matrix summary
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
Summary: Focused encode-stage x264enc ablation across tune, trellis, and speed-preset for Go, Python, and gst-launch hosts.
LastUpdated: 2026-04-15T05:00:00-04:00
WhatFor: Preserve the x264enc-property matrix used to narrow which x264 settings keep or reduce the Go-hosted anomaly.
WhenToUse: Read this when continuing the x264-specific branch of the hosting-gap investigation.
---

# 35 x264 property ablation matrix

## baseline (speed-preset=3, tune=4, bframes=0, trellis=true)

| host | avg_cpu | max_cpu | page-faults | minor-faults | major-faults | result | error |
|---|---:|---:|---:|---:|---:|---|---|
| go | 165.88 | 319.00 | 233962 | 233959 | 0 | eos |  |
| python | 145.50 | 155.00 | 63 | 63 | 0 | eos |  |
| gst-launch | 144.11 | 160.00 | 2 | 2 | 0 | gst-launch |  |

## no_tune (speed-preset=3, tune=0, bframes=0, trellis=true)

| host | avg_cpu | max_cpu | page-faults | minor-faults | major-faults | result | error |
|---|---:|---:|---:|---:|---:|---|---|
| go | 301.67 | 583.00 | 187211 | 187211 | 0 | eos |  |
| python | 146.50 | 181.00 | 10607 | 10606 | 0 | eos |  |
| gst-launch | 128.11 | 156.00 | 3534 | 3534 | 0 | gst-launch |  |

## no_trellis (speed-preset=3, tune=4, bframes=0, trellis=false)

| host | avg_cpu | max_cpu | page-faults | minor-faults | major-faults | result | error |
|---|---:|---:|---:|---:|---:|---|---|
| go | 343.56 | 593.00 | 179717 | 179717 | 0 | eos |  |
| python | 139.33 | 154.00 | 62 | 62 | 0 | eos |  |
| gst-launch | 137.44 | 154.00 | 0 | 0 | 0 | gst-launch |  |

## ultrafast (speed-preset=1, tune=4, bframes=0, trellis=false)

| host | avg_cpu | max_cpu | page-faults | minor-faults | major-faults | result | error |
|---|---:|---:|---:|---:|---:|---|---|
| go | 332.00 | 576.00 | 136321 | 136321 | 0 | eos |  |
| python | 83.17 | 96.00 | 95 | 95 | 0 | eos |  |
| gst-launch | 79.33 | 87.00 | 23 | 23 | 0 | gst-launch |  |

