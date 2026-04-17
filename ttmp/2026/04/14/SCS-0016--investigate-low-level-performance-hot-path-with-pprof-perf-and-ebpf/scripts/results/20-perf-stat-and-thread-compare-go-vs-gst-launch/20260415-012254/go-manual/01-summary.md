---
Title: go-manual perf stat summary
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

# go-manual perf stat summary

- thread_samples: 5
- threads_status_min_max: 1..25
- threads_ps_min_max: 24..25

| metric | value | unit |
|---|---:|---|
| task-clock | 154.16 | msec |
| context-switches | 207 |  |
| cpu-migrations | 14 |  |
| page-faults | 142 |  |
| cycles | 325761451 |  |
| instructions | 437244422 |  |
| branches | 71904759 |  |
| branch-misses | 204103 |  |
