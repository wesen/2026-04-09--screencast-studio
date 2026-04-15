---
Title: Prometheus metrics architecture and field guide
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
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/metrics/metrics.go
      Note: In-process metrics registry and Prometheus text exposition implementation
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_metrics.go
      Note: HTTP export endpoint for the metrics registry
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_metrics.go
      Note: Preview HTTP, PreviewManager, and timing metric families
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/event_metrics.go
      Note: EventHub and websocket metric families
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/recording.go
      Note: GStreamer audio-level parse failure metrics
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_recording_bridge.go
      Note: Shared bridge recorder metric families
ExternalSources: []
Summary: Field guide to the project's Prometheus-style metrics: how they are implemented, what they measure, and how they were used during the performance investigation.
LastUpdated: 2026-04-14T21:00:00-04:00
WhatFor: Preserve a continuation-friendly explanation of the current metrics architecture and each major metric family.
WhenToUse: Read when adding new metrics, interpreting `/metrics`, or wiring future Grafana/Prometheus dashboards.
---

# Prometheus metrics architecture and field guide

## Goal

Explain how Screencast Studio’s Prometheus-style metrics work, what metric families currently exist, what they are intended to measure, and how the investigation has used them so far.

## Context

The project uses a small in-process metrics package instead of a larger external Prometheus client library. That choice was sufficient for the investigation because the first goal was not “enterprise observability.” The goal was to expose a stable, grep-friendly `/metrics` surface quickly so the performance investigation could correlate server CPU with concrete counts, gauges, and timing accumulators.

Over time, the metric set evolved from a very small backend observability slice into a more detailed browser-path investigation surface. The metrics now cover:

- preview HTTP serving,
- PreviewManager internal timing and lifecycle,
- EventHub and websocket fanout,
- GStreamer recording parse failures,
- and the shared video recording bridge.

## Quick Reference

## 1. Architecture overview

### In-process registry

The core implementation lives in:

- `pkg/metrics/metrics.go`

It provides:

- `CounterVec`
- `GaugeVec`
- a default registry
- Prometheus text exposition via `WritePrometheus`

Important characteristics:

- metrics are registered process-locally
- counters and gauges are stored in maps keyed by normalized label sets
- values are stored with atomics
- exposition is deterministic and sorted by metric name and label set
- no histograms or summaries currently exist

### HTTP export

The export endpoint lives in:

- `internal/web/handlers_metrics.go`

and is mounted at:

- `/metrics`

It renders the contents of the default registry in standard Prometheus text format.

### Why this architecture was good enough

It was fast to extend, easy to inspect, and easy to consume from ticket-local scripts. That was more important than a feature-complete client library during the investigation phase.

## 2. How to interpret the metrics

### Most metrics are cumulative counters

That means short experiments should not read raw counter values in isolation. Instead, they should compare the first and last snapshot in a run directory and compute a delta.

That is exactly why the browser sampler scripts save multiple `.prom` snapshots and emit a `metric-deltas.txt` summary.

### Gauges are usually “current state” values

Examples:

- active preview HTTP clients
- current EventHub subscribers
- current websocket connections

These are typically interpreted as the **last observed** value in a run.

### Timing metrics are cumulative nanosecond counters, not distributions

The project currently uses cumulative nanosecond totals rather than histograms.

That means the usual interpretation is:

- total nanoseconds over the run
- divided by a relevant event count if per-event or per-frame cost is needed

This was sufficient for the investigation because the main question was usually “is this path even remotely expensive enough?”

## 3. Metric families by subsystem

### A. Preview HTTP serving metrics

Defined in:
- `internal/web/preview_metrics.go`

Produced mainly by:
- `internal/web/handlers_preview.go`

Metric families:

- `screencast_studio_preview_http_clients`
  - gauge
  - labels: `source_type`
  - meaning: active MJPEG clients currently connected

- `screencast_studio_preview_http_streams_started_total`
  - counter
  - labels: `source_type`
  - meaning: how many MJPEG streams were started

- `screencast_studio_preview_http_streams_finished_total`
  - counter
  - labels: `source_type`, `reason`
  - meaning: why MJPEG streams ended

- `screencast_studio_preview_http_frames_served_total`
  - counter
  - labels: `source_type`
  - meaning: JPEG frames written to HTTP clients

- `screencast_studio_preview_http_bytes_served_total`
  - counter
  - labels: `source_type`
  - meaning: total multipart/JPEG bytes written

- `screencast_studio_preview_http_flushes_total`
  - counter
  - labels: `source_type`
  - meaning: total flush calls while serving MJPEG

- `screencast_studio_preview_http_loop_iterations_total`
  - counter
  - labels: `source_type`
  - meaning: total handler-loop iterations

- `screencast_studio_preview_http_idle_iterations_total`
  - counter
  - labels: `source_type`
  - meaning: iterations that did not serve a new frame

- `screencast_studio_preview_http_write_nanoseconds_total`
  - counter
  - labels: `source_type`
  - meaning: cumulative time spent writing multipart headers and JPEG bytes

- `screencast_studio_preview_http_flush_nanoseconds_total`
  - counter
  - labels: `source_type`
  - meaning: cumulative time spent in `Flush()`

### Why these were added

These metrics were added to answer increasingly narrow questions:

1. are real browser runs serving lots more frames/bytes than synthetic clients?
2. is the final MJPEG write/flush loop itself expensive enough to explain the hot phase?

The later result was particularly important: write/flush time turned out to be tiny relative to the observed CPU spikes.

### B. PreviewManager lifecycle and timing metrics

Defined in:
- `internal/web/preview_metrics.go`

Produced mainly by:
- `internal/web/preview_manager.go`

Metric families:

- `screencast_studio_preview_frame_updates_total`
  - counter
  - labels: `source_type`
  - meaning: stored frame updates in PreviewManager

- `screencast_studio_preview_frame_store_nanoseconds_total`
  - counter
  - labels: `source_type`
  - meaning: cumulative time to store a preview frame, including cached copy and `preview.state` publish path

- `screencast_studio_preview_latest_frame_copy_nanoseconds_total`
  - counter
  - labels: `source_type`
  - meaning: cumulative time spent copying cached frames back out to callers

- `screencast_studio_preview_state_publish_nanoseconds_total`
  - counter
  - labels: `source_type`
  - meaning: cumulative time spent publishing `preview.state`

- `screencast_studio_preview_ensures_total`
  - counter
  - labels: `source_type`, `result`
  - meaning: ensure attempts and whether they created, reused, or failed

- `screencast_studio_preview_releases_total`
  - counter
  - labels: `source_type`, `result`
  - meaning: release requests and outcomes

### Why these were added

These metrics were added after MJPEG timing suggested the final HTTP write path was too cheap. They allowed the project to check the next immediate Go-side layer and showed that PreviewManager copy/store/publication work was also too small to explain the hot phase.

### C. EventHub and websocket metrics

Defined in:
- `internal/web/event_metrics.go`

Produced mainly by:
- `internal/web/event_hub.go`
- `internal/web/handlers_ws.go`

Metric families:

- `screencast_studio_eventhub_subscribers`
  - gauge
  - labels: none
  - meaning: number of EventHub subscribers

- `screencast_studio_eventhub_events_published_total`
  - counter
  - labels: `event_type`
  - meaning: events published into the hub

- `screencast_studio_eventhub_events_delivered_total`
  - counter
  - labels: `event_type`
  - meaning: events delivered into subscriber channels

- `screencast_studio_eventhub_events_dropped_total`
  - counter
  - labels: `event_type`
  - meaning: events dropped because subscriber channels were full

- `screencast_studio_eventhub_publish_nanoseconds_total`
  - counter
  - labels: `event_type`
  - meaning: cumulative time spent inside `EventHub.Publish`

- `screencast_studio_websocket_connections`
  - gauge
  - labels: none
  - meaning: active websocket connections

- `screencast_studio_websocket_events_written_total`
  - counter
  - labels: `event_type`
  - meaning: websocket events successfully written

- `screencast_studio_websocket_event_write_errors_total`
  - counter
  - labels: `event_type`
  - meaning: websocket event write failures

### Why these were added

These metrics were added because the browser path includes `/ws`, while plain MJPEG baselines do not. They made it possible to measure synthetic websocket ablations and later to show that EventHub publish time was too small to be the dominant hot path.

### D. GStreamer recording parse-failure metrics

Defined in:
- `pkg/media/gst/recording.go`

Metric family:

- `screencast_studio_gst_audio_level_parse_failures_total`
  - counter
  - labels: `reason`, `rms_type`
  - meaning: audio-level parse failures from the recording runtime

### Why this exists

This metric was added to replace noisy repeated log spam with something that can be counted and monitored more safely. It is an observability metric more than a browser-performance metric.

### E. Shared bridge recorder metrics

Defined in:
- `pkg/media/gst/shared_video_recording_bridge.go`

Metric families:

- `screencast_studio_gst_shared_bridge_recorder_samples_received_total`
- `screencast_studio_gst_shared_bridge_recorder_buffers_copied_total`
- `screencast_studio_gst_shared_bridge_recorder_enqueued_total`
- `screencast_studio_gst_shared_bridge_recorder_dropped_total`
- `screencast_studio_gst_shared_bridge_recorder_worker_handled_total`
- `screencast_studio_gst_shared_bridge_recorder_appsrc_pushed_total`

All currently use:
- labels: `source_type`

### Why these exist

These metrics support visibility into the shared video recording bridge, especially:

- how many buffers are seen,
- how many are copied,
- whether the async queue is dropping,
- and whether appsrc push is keeping up.

They were originally motivated by recording-path debugging and remain relevant whenever shared bridge backpressure is suspected.

## 4. Label strategy and why it matters

The project deliberately kept labels low-cardinality.

Common labels include:
- `source_type`
- `reason`
- `result`
- `event_type`
- `rms_type`

The project intentionally avoided labels like:
- preview ID
- session ID
- source ID
- browser tab identity

Why:
- easier aggregation
- smaller outputs
- safer long-lived metrics use
- better fit for repeated ticket-local comparisons and future Grafana dashboards

This choice was one of the best observability decisions in the project so far.

## 5. How the investigation actually used the metrics

### Use case 1: compare fresh-server and real-browser runs

The sampler scripts saved repeated `.prom` snapshots and compared the first and last values. That made cumulative counters useful in short-lived experiments.

### Use case 2: decide whether byte volume alone explains the spike

Frame/byte deltas showed that browser recording runs were hot even without proportionally huge served-byte deltas.

### Use case 3: decide whether websocket fanout explains the spike

EventHub and websocket counters made the synthetic websocket ablation measurable and showed that websocket fanout alone did not explain the full browser gap.

### Use case 4: decide whether final MJPEG write/flush explains the spike

Timing counters showed final write/flush time was far too small to explain the late-run `~300–400%` CPU band.

### Use case 5: decide whether PreviewManager/EventHub internals explain the spike

More timing counters showed that PreviewManager copy/store/publication and EventHub publish costs were also too small, which justified escalating to lower-level profiling.

## 6. Current limitations of the metrics system

### No histograms

All timing metrics are cumulative counters, so percentiles and distribution shape are not available.

### No automatic per-run reset

The registry is process-global, so per-run interpretation depends on delta sampling.

### No built-in external scraping or retention

The current setup is perfect for ticket-local scripts and manual analysis, but a full Prometheus+Grafana deployment would still need extra wiring and dashboard design.

### Not a replacement for profilers

The current metrics can rule out many app-level suspects, but they cannot fully explain lower-level CPU in GStreamer, CGO, or libc. That is why SCS-0016 exists.

## Usage Examples

### Example: interpreting a timing metric

If a run shows:

- `preview_http_write_nanoseconds_total delta = 7,028,353`
- `preview_http_frames_served_total delta = 60`

then the approximate write cost per frame is:

```text
7,028,353 ns / 60 ≈ 117,139 ns ≈ 0.117 ms/frame
```

That kind of calculation was used repeatedly to decide whether a path was even plausibly expensive enough.

### Example: interpreting a gauge

If a run ends with:

```text
screencast_studio_websocket_connections last=1
```

that means the last sampled state saw one active websocket connection. It does not mean the total number of websocket connections over the whole run was one.

### Example: adding a new metric family safely

Follow the current pattern:

1. define it in the relevant subsystem file
2. keep labels low-cardinality
3. expose it automatically through `/metrics`
4. update tests
5. consume it through first/last snapshot deltas in ticket-local scripts

## Related

- `reference/02-performance-investigation-approaches-and-tricks-report.md`
- `pkg/metrics/metrics.go`
- `internal/web/handlers_metrics.go`
