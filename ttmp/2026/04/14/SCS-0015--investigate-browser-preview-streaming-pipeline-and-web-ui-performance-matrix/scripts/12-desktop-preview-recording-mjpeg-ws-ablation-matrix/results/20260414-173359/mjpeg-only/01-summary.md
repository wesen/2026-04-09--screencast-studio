---
Title: 12 desktop preview recording mjpeg websocket ablation summary
Ticket: SCS-0015
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - performance
    - preview
    - browser
    - websocket
    - recording
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: One focused desktop preview+recording ablation scenario comparing MJPEG-only versus MJPEG-plus-websocket server load.
LastUpdated: 2026-04-14T17:40:00-04:00
WhatFor: Preserve a focused server-side ablation run for the desktop preview+recording browser hypothesis.
WhenToUse: Read when comparing the likely websocket/event contribution against the earlier browser-path findings.
---

# 12 desktop preview recording mjpeg websocket ablation summary

- scenario: mjpeg-only
- with_ws: 0
- server: http://127.0.0.1:7830
- avg_cpu: 11.00%
- max_cpu: 59.00%
- preview_id: preview-909758a3cc65
- metrics_dir: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173359/mjpeg-only/metrics
- video_path: 

## Metric deltas

~~~text
screencast_studio_eventhub_events_published_total{event_type="preview.state"}	delta=0
screencast_studio_eventhub_events_published_total{event_type="session.log"}	delta=0
screencast_studio_eventhub_events_published_total{event_type="session.state"}	delta=0
screencast_studio_eventhub_events_published_total{event_type="telemetry.disk_status"}	delta=3
screencast_studio_preview_http_clients{source_type="display"}	last=1
screencast_studio_preview_http_streams_started_total{source_type="display"}	delta=0
~~~
