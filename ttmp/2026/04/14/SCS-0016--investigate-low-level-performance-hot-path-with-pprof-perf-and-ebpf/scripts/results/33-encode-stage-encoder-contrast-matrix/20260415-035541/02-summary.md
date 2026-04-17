---
Title: 33 encode stage encoder contrast matrix summary
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
Summary: Focused encode-stage comparison across x264enc, openh264enc, and vaapih264enc for Go, Python, and gst-launch hosts.
LastUpdated: 2026-04-15T04:20:00-04:00
WhatFor: Preserve the encoder-only contrast matrix used to test whether the Go-hosted anomaly is specific to x264enc or broader across encoder implementations.
WhenToUse: Read this when continuing the encoder-boundary investigation after the small-graph ladder localized the first strong divergence to encode.
---

# 33 encode stage encoder contrast matrix

## x264enc

| host | avg_cpu | max_cpu | page-faults | minor-faults | major-faults | result | error |
|---|---:|---:|---:|---:|---:|---|---|
| go | 168.83 | 340.00 | 276940 | 276936 | 0 | eos |  |
| python | 127.17 | 133.00 | 61 | 61 | 0 | eos |  |
| gst-launch | 129.67 | 139.00 | 2 | 2 | 0 | gst-launch |  |

## openh264enc

| host | avg_cpu | max_cpu | page-faults | minor-faults | major-faults | result | error |
|---|---:|---:|---:|---:|---:|---|---|
| go | 58.00 | 59.00 | 53 | 53 | 0 | eos |  |
| python | 87.50 | 90.00 | 110 | 110 | 0 | eos |  |
| gst-launch | 88.11 | 90.00 | 53 | 53 | 0 | gst-launch |  |

