---
Title: 07 live server browser scenario sample summary
Ticket: SCS-0015
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - performance
    - preview
    - browser
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: One live browser-driven server-sampling run against the real Studio page.
LastUpdated: 2026-04-14T16:40:00-04:00
WhatFor: Preserve pidstat, metrics, and preview/recording API snapshots while a browser scenario is active.
WhenToUse: Read when comparing real browser-tab scenarios against HTTP-client baselines.
---

# 07 live server browser scenario sample summary

- label: desktop-two-tab-preview-only
- server_url: http://127.0.0.1:7777
- server_pid: 321005
- avg_cpu: 12.69%
- max_cpu: 14.00%
- metrics_dir: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-164457/metrics

## Files

- server.pidstat.log
- metrics/
- previews-*.json
- recordings-current-*.json
