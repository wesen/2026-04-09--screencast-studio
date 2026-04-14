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

- label: desktop-camera-one-tab-preview-only-with-deltas
- server_url: http://127.0.0.1:7777
- server_pid: 321005
- avg_cpu: 18.17%
- max_cpu: 24.00%
- metrics_dir: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-165126/metrics

## Metric deltas

~~~text
screencast_studio_preview_http_bytes_served_total{source_type="camera"}	delta=3815988
screencast_studio_preview_http_bytes_served_total{source_type="display"}	delta=8108165
screencast_studio_preview_http_clients{source_type="camera"}	last=1
screencast_studio_preview_http_clients{source_type="display"}	last=1
screencast_studio_preview_http_flushes_total{source_type="camera"}	delta=27
screencast_studio_preview_http_flushes_total{source_type="display"}	delta=31
screencast_studio_preview_http_frames_served_total{source_type="camera"}	delta=27
screencast_studio_preview_http_frames_served_total{source_type="display"}	delta=31
screencast_studio_preview_http_streams_finished_total{source_type="display",reason="client_done"}	delta=0
screencast_studio_preview_http_streams_started_total{source_type="camera"}	delta=0
screencast_studio_preview_http_streams_started_total{source_type="display"}	delta=0
~~~

## Files

- server.pidstat.log
- metrics/
- metric-deltas.txt
- previews-*.json
- recordings-current-*.json
