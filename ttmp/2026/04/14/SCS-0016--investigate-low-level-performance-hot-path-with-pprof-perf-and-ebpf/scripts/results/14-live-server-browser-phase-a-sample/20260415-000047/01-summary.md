---
Title: 14 live server browser phase a sample summary
Ticket: SCS-0016
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
Summary: One live browser-driven server-sampling run against the real Studio page while Phase A debug ablation flags are active.
LastUpdated: 2026-04-15T00:00:58-04:00
WhatFor: Preserve pidstat, metrics, and preview/recording API snapshots for the real browser Phase A perturbation run.
WhenToUse: Read when comparing Phase A ablation against earlier unperturbed real browser runs.
---

# 14 live server browser phase a sample summary

- label: audio-disabled-browser-preview-recording-one-tab
- server_url: http://127.0.0.1:7777
- server_pid: 986187
- avg_cpu: 509.50%
- max_cpu: 558.00%
- metrics_dir: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/14-live-server-browser-phase-a-sample/20260415-000047/metrics

## Metric deltas

~~~text
screencast_studio_eventhub_events_delivered_total{event_type="preview.state"}	delta=29
screencast_studio_eventhub_events_delivered_total{event_type="session.log"}	delta=0
screencast_studio_eventhub_events_delivered_total{event_type="session.state"}	delta=0
screencast_studio_eventhub_events_delivered_total{event_type="telemetry.audio_meter"}	delta=7
screencast_studio_eventhub_events_delivered_total{event_type="telemetry.disk_status"}	delta=4
screencast_studio_eventhub_events_published_total{event_type="preview.state"}	delta=29
screencast_studio_eventhub_events_published_total{event_type="session.log"}	delta=0
screencast_studio_eventhub_events_published_total{event_type="session.state"}	delta=0
screencast_studio_eventhub_events_published_total{event_type="telemetry.audio_meter"}	delta=7
screencast_studio_eventhub_events_published_total{event_type="telemetry.disk_status"}	delta=4
screencast_studio_eventhub_publish_nanoseconds_total{event_type="preview.state"}	delta=1031553
screencast_studio_eventhub_publish_nanoseconds_total{event_type="session.log"}	delta=0
screencast_studio_eventhub_publish_nanoseconds_total{event_type="session.state"}	delta=0
screencast_studio_eventhub_publish_nanoseconds_total{event_type="telemetry.audio_meter"}	delta=154678
screencast_studio_eventhub_publish_nanoseconds_total{event_type="telemetry.disk_status"}	delta=39306
screencast_studio_eventhub_subscribers	last=1
screencast_studio_preview_frame_store_nanoseconds_total{source_type="display"}	delta=2058937
screencast_studio_preview_frame_updates_total{source_type="display"}	delta=29
screencast_studio_preview_http_bytes_served_total{source_type="display"}	delta=3209104
screencast_studio_preview_http_clients{source_type="display"}	last=1
screencast_studio_preview_http_flush_nanoseconds_total{source_type="display"}	delta=1588460
screencast_studio_preview_http_flushes_total{source_type="display"}	delta=30
screencast_studio_preview_http_frames_served_total{source_type="display"}	delta=30
screencast_studio_preview_http_idle_iterations_total{source_type="display"}	delta=43
screencast_studio_preview_http_loop_iterations_total{source_type="display"}	delta=73
screencast_studio_preview_http_streams_finished_total{source_type="display",reason="client_done"}	delta=0
screencast_studio_preview_http_streams_started_total{source_type="display"}	delta=0
screencast_studio_preview_http_write_nanoseconds_total{source_type="display"}	delta=34923105
screencast_studio_preview_latest_frame_copy_nanoseconds_total{source_type="display"}	delta=13999978
screencast_studio_preview_state_publish_nanoseconds_total{source_type="display"}	delta=1272519
screencast_studio_websocket_connections	last=1
screencast_studio_websocket_event_write_errors_total{event_type="telemetry.disk_status"}	delta=0
screencast_studio_websocket_events_written_total{event_type="preview.list"}	delta=0
screencast_studio_websocket_events_written_total{event_type="preview.state"}	delta=29
screencast_studio_websocket_events_written_total{event_type="session.log"}	delta=0
screencast_studio_websocket_events_written_total{event_type="session.state"}	delta=0
screencast_studio_websocket_events_written_total{event_type="telemetry.audio_meter"}	delta=7
screencast_studio_websocket_events_written_total{event_type="telemetry.disk_status"}	delta=4
~~~

## Files

- server.pidstat.log
- metrics/
- metric-deltas.txt
- previews-*.json
- recordings-current-*.json
