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

- scenario: mjpeg-plus-ws
- with_ws: 1
- server: http://127.0.0.1:7831
- avg_cpu: 170.48%
- max_cpu: 481.00%
- preview_id: preview-909758a3cc65
- metrics_dir: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173541/mjpeg-plus-ws/metrics
- video_path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173541/mjpeg-plus-ws/output/Full Desktop.mov

## Metric deltas

~~~text
screencast_studio_eventhub_events_delivered_total{event_type="preview.state"}	delta=23
screencast_studio_eventhub_events_delivered_total{event_type="session.log"}	delta=1
screencast_studio_eventhub_events_delivered_total{event_type="session.state"}	delta=4
screencast_studio_eventhub_events_delivered_total{event_type="telemetry.disk_status"}	delta=2
screencast_studio_eventhub_events_published_total{event_type="preview.state"}	delta=33
screencast_studio_eventhub_events_published_total{event_type="session.log"}	delta=1
screencast_studio_eventhub_events_published_total{event_type="session.state"}	delta=4
screencast_studio_eventhub_events_published_total{event_type="telemetry.disk_status"}	delta=3
screencast_studio_eventhub_subscribers	last=1
screencast_studio_preview_frame_updates_total{source_type="display"}	delta=33
screencast_studio_preview_http_bytes_served_total{source_type="display"}	delta=6527372
screencast_studio_preview_http_clients{source_type="display"}	last=1
screencast_studio_preview_http_flushes_total{source_type="display"}	delta=33
screencast_studio_preview_http_frames_served_total{source_type="display"}	delta=33
screencast_studio_preview_http_streams_started_total{source_type="display"}	delta=0
screencast_studio_websocket_connections	last=1
screencast_studio_websocket_events_written_total{event_type="preview.list"}	delta=1
screencast_studio_websocket_events_written_total{event_type="preview.state"}	delta=23
screencast_studio_websocket_events_written_total{event_type="session.log"}	delta=1
screencast_studio_websocket_events_written_total{event_type="session.state"}	delta=5
screencast_studio_websocket_events_written_total{event_type="telemetry.audio_meter"}	delta=1
screencast_studio_websocket_events_written_total{event_type="telemetry.disk_status"}	delta=3
~~~
