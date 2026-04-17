---
Title: gst-launch perf stat threadgroup summary
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

# gst-launch perf stat threadgroup summary

- thread_samples: 7
- threads_status_min_max: 1..22
- threads_ps_min_max: 22..22
- attached_tids: 22

| metric | value | unit |
|---|---:|---|
| task-clock | 11327.53 | msec |
| context-switches | 9813 |  |
| cpu-migrations | 2580 |  |
| page-faults | 232 |  |
| cycles | 30699553389 |  |
| instructions | 61726971637 |  |
| branches | 4399512165 |  |
| branch-misses | 31429866 |  |
