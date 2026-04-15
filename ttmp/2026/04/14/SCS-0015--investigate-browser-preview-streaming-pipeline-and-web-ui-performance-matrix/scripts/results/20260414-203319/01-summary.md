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

- label: scs0015-browser-desktop-recording-one-tab-upstream-timing
- server_url: http://127.0.0.1:7777
- server_pid: 681602
- avg_cpu: 150.70%
- max_cpu: 394.00%
- metrics_dir: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/results/20260414-203319/metrics

## Metric deltas

~~~text
screencast_studio_eventhub_events_delivered_total{event_type="preview.state"}	delta=60
screencast_studio_eventhub_events_delivered_total{event_type="session.log"}	delta=2
screencast_studio_eventhub_events_delivered_total{event_type="session.state"}	delta=5
screencast_studio_eventhub_events_delivered_total{event_type="telemetry.audio_meter"}	delta=45
screencast_studio_eventhub_events_delivered_total{event_type="telemetry.disk_status"}	delta=3
screencast_studio_eventhub_events_published_total{event_type="preview.state"}	delta=60
screencast_studio_eventhub_events_published_total{event_type="session.log"}	delta=2
screencast_studio_eventhub_events_published_total{event_type="session.state"}	delta=5
screencast_studio_eventhub_events_published_total{event_type="telemetry.audio_meter"}	delta=45
screencast_studio_eventhub_events_published_total{event_type="telemetry.disk_status"}	delta=3
screencast_studio_eventhub_publish_nanoseconds_total{event_type="preview.state"}	delta=535814
screencast_studio_eventhub_publish_nanoseconds_total{event_type="session.log"}	delta=7802
screencast_studio_eventhub_publish_nanoseconds_total{event_type="session.state"}	delta=21497
screencast_studio_eventhub_publish_nanoseconds_total{event_type="telemetry.audio_meter"}	delta=6452838
screencast_studio_eventhub_publish_nanoseconds_total{event_type="telemetry.disk_status"}	delta=23904
screencast_studio_eventhub_subscribers	last=1
screencast_studio_preview_frame_store_nanoseconds_total{source_type="display"}	delta=3691230
screencast_studio_preview_frame_updates_total{source_type="display"}	delta=60
screencast_studio_preview_http_bytes_served_total{source_type="display"}	delta=10258225
screencast_studio_preview_http_clients{source_type="display"}	last=1
screencast_studio_preview_http_flush_nanoseconds_total{source_type="display"}	delta=511999
screencast_studio_preview_http_flushes_total{source_type="display"}	delta=60
screencast_studio_preview_http_frames_served_total{source_type="display"}	delta=60
screencast_studio_preview_http_idle_iterations_total{source_type="display"}	delta=11
screencast_studio_preview_http_loop_iterations_total{source_type="display"}	delta=71
screencast_studio_preview_http_streams_finished_total{source_type="display",reason="client_done"}	delta=0
screencast_studio_preview_http_streams_started_total{source_type="display"}	delta=0
screencast_studio_preview_http_write_nanoseconds_total{source_type="display"}	delta=7028353
screencast_studio_preview_latest_frame_copy_nanoseconds_total{source_type="display"}	delta=1796290
screencast_studio_preview_state_publish_nanoseconds_total{source_type="display"}	delta=790719
screencast_studio_websocket_connections	last=1
screencast_studio_websocket_event_write_errors_total{event_type="preview.state"}	delta=0
screencast_studio_websocket_events_written_total{event_type="preview.list"}	delta=0
screencast_studio_websocket_events_written_total{event_type="preview.state"}	delta=60
screencast_studio_websocket_events_written_total{event_type="session.log"}	delta=2
screencast_studio_websocket_events_written_total{event_type="session.state"}	delta=5
screencast_studio_websocket_events_written_total{event_type="telemetry.audio_meter"}	delta=45
screencast_studio_websocket_events_written_total{event_type="telemetry.disk_status"}	delta=3
~~~

## Files

- server.pidstat.log
- metrics/
- metric-deltas.txt
- previews-*.json
- recordings-current-*.json
