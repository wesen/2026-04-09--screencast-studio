---
Title: 21 perf stat threadgroup compare go vs gst-launch
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

# 21 perf stat threadgroup compare go vs gst-launch

| scenario | codec | width | height | fps | duration | threads_status_min_max | attached_tids | task-clock | context-switches | cpu-migrations | page-faults | cycles | instructions |
|---|---|---:|---:|---|---:|---|---:|---:|---:|---:|---:|---:|---:|
| go-manual | h264 | 2880 | 1920 | 24/1 | 8.000000 | 1..25 | 24 | 26480.36 | 14953 | 2306 | 134832 | 63853466536 | 114809329662 |
| gst-launch | h264 | 2880 | 1920 | 24/1 | 8.000000 | 1..22 | 22 | 11327.53 | 9813 | 2580 | 232 | 30699553389 | 61726971637 |
