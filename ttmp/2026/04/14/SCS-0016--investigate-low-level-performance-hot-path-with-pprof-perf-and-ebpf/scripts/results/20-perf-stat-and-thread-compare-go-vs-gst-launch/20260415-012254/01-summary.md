---
Title: 20 perf stat and thread compare go vs gst-launch
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

# 20 perf stat and thread compare go vs gst-launch

| scenario | codec | width | height | fps | duration | threads_status_min_max | task-clock | context-switches | cpu-migrations | page-faults | cycles | instructions |
|---|---|---:|---:|---|---:|---|---:|---:|---:|---:|---:|---:|
| go-manual | h264 | 2880 | 1920 | 24/1 | 7.916667 | 1..25 | 154.16 | 207 | 14 | 142 | 325761451 | 437244422 |
| gst-launch | h264 | 2880 | 1920 | 24/1 | 8.000000 | 1..22 | 13078.58 | 11912 | 3223 | 12591 | 34876202265 | 69614828852 |
