---
Title: go-manual perf stat threadgroup summary
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

# go-manual perf stat threadgroup summary

- thread_samples: 7
- threads_status_min_max: 1..25
- threads_ps_min_max: 24..25
- attached_tids: 24

| metric | value | unit |
|---|---:|---|
| task-clock | 26480.36 | msec |
| context-switches | 14953 |  |
| cpu-migrations | 2306 |  |
| page-faults | 134832 |  |
| cycles | 63853466536 |  |
| instructions | 114809329662 |  |
| branches | 9030640305 |  |
| branch-misses | 53353356 |  |
