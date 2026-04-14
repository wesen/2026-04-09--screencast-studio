---
Title: Browser preview streaming pipeline analysis and performance matrix plan
Ticket: SCS-0015
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - video
    - backend
    - ui
    - analysis
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/handlers_metrics.go
      Note: Prometheus-style export endpoint for the new browser-preview metrics
    - Path: internal/web/handlers_preview.go
      Note: |-
        HTTP preview ensure/release endpoints and MJPEG streaming loop define the browser-facing preview transport
        Browser-facing MJPEG preview transport and streaming loop
    - Path: internal/web/preview_manager.go
      Note: |-
        Preview lifecycle, reuse, frame storage, and per-preview state live here
        Preview lifecycle and cached JPEG frame fan-out boundary
    - Path: internal/web/preview_metrics.go
      Note: First browser-preview metric families and low-cardinality label choices live here
    - Path: pkg/media/gst/preview.go
      Note: |-
        GStreamer preview runtime acquires shared sources and attaches preview consumers
        Shared-source preview runtime entrypoint
    - Path: pkg/media/gst/shared_video.go
      Note: |-
        Shared preview branch builds JPEG appsink consumers and therefore defines the upstream browser-preview payload cost
        Shared preview branch performs JPEG encoding and appsink delivery
    - Path: pkg/metrics/metrics.go
      Note: |-
        New in-process Prometheus-style metrics registry defines current observability ceiling and extension points
        Current Prometheus-style metrics registry to extend for browser-preview observability
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/01-restart-scs-web-ui.sh
      Note: Local restart helper for later browser-driven matrix runs
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/02-sample-preview-metrics.sh
      Note: Raw metrics sampler for time-windowed preview measurements
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/03-desktop-preview-http-client-matrix/run.sh
      Note: First dedicated MJPEG-client baseline matrix harness for the browser-preview investigation
    - Path: ui/src/components/preview/PreviewStream.tsx
      Note: |-
        Browser preview rendering uses plain img tags pointed at MJPEG endpoints
        Browser preview renderer uses img tags pointed at MJPEG endpoints
    - Path: ui/src/pages/StudioPage.tsx
      Note: |-
        Frontend preview ensure/release lifecycle and tab-driven preview ownership are orchestrated here
        Frontend preview ensure/release lifecycle and ownership rules
ExternalSources: []
Summary: Evidence-backed current-state analysis and experiment plan for investigating server CPU cost in the browser preview streaming path, especially the gap between API-only preview/recording measurements and the much hotter real web UI path.
LastUpdated: 2026-04-14T15:40:00-04:00
WhatFor: Orient future investigation of browser-connected preview streaming overhead, define the missing performance matrices, and document what should be measured next.
WhenToUse: Read this before adding browser-preview metrics, building new measurement harnesses, or interpreting server CPU spikes observed only when the real Studio page is open.
---





# Browser preview streaming pipeline analysis and performance matrix plan

## Executive summary

The project already has strong measurement coverage for the shared GStreamer capture and recording path, including recorder-only, preview-only, preview-plus-recorder, staged bridge-overhead, adaptive-preview, and live app-path API-driven measurements. However, those experiments do not yet fully isolate the **browser-connected preview streaming path**.

That gap matters because the user observed an important discrepancy: when using the real web UI, pressing record can drive the server CPU far higher than the earlier standalone or API-only matrix runs suggested, including spikes around `400%` when desktop and camera are armed and even very high CPU for desktop-only recording through the Studio page. That implies there is another significant cost center in the end-to-end path: not just capture + encode, but also **preview serving to browsers**, **frontend preview lifecycle behavior**, or **browser-connected preview fan-out effects**.

This ticket therefore focuses on the browser-streaming part of the pipeline rather than on the recorder architecture alone. The investigation should answer four questions:

1. How much server CPU comes from **keeping browser preview clients attached**?
2. How much of that cost is due to the current **MJPEG over `<img>`** transport?
3. Does the real Studio page create more preview work than the API-only harnesses because of **ensure/release timing, multiple listeners, or tab behavior**?
4. Can Prometheus-style metrics plus targeted browser-driven experiments produce a trustworthy explanation and optimization plan?

## Problem statement and scope

### Problem

The existing performance work established useful baseline conclusions about the shared GStreamer runtime, but it did not fully explain the much hotter server CPU seen when the real browser UI is connected and recording is started from the Studio page.

### In scope

This ticket investigates:

- backend MJPEG preview serving cost,
- browser-connected preview client behavior,
- frontend preview lifecycle orchestration,
- differences between API-only and browser-driven measurement paths,
- camera-versus-desktop browser preview cost,
- metrics needed for Prometheus/Grafana-style observability,
- and a new performance matrix focused on browser streaming.

### Out of scope

This ticket is **not** primarily about:

- redesigning the recorder bridge,
- transcription,
- feature UX unrelated to preview streaming,
- or replacing the whole preview transport yet.

Those topics may appear as alternatives or follow-ups, but the first goal here is diagnosis.

## Current-state architecture

### 1. Browser previews are plain MJPEG HTTP streams, not WebRTC or WebSocket media

The server registers `/api/previews/{id}/mjpeg` under the generic preview route path (`internal/web/routes.go:19-25`). The HTTP handler in `internal/web/handlers_preview.go:79-149` serves a multipart MJPEG stream by polling `s.previews.LatestFrame(previewID)` every `100ms`, writing JPEG frame boundaries and bytes directly to the response, and flushing after each new frame.

Important consequences:

- each connected browser preview client holds an active HTTP stream,
- the server repeatedly checks preview state on a fixed ticker,
- and each client receives a separate stream response even though the underlying preview frame bytes are shared in memory.

This is simple and robust, but it may also be expensive in the combined case of:

- multiple active previews,
- multiple browser tabs,
- concurrent recording,
- and camera + desktop combinations.

### 2. PreviewManager stores one latest JPEG frame per preview and fans it out to HTTP readers

`internal/web/preview_manager.go:95-220` shows the main preview lifecycle:

- `Ensure(...)` normalizes DSL, finds the source, computes a preview signature, reuses or creates a preview session, and installs an `OnFrame` callback.
- The `OnFrame` callback stores the latest JPEG bytes via `storePreviewFrame(...)`.
- `LatestFrame(...)` then returns a copy of that cached JPEG frame plus a frame sequence.

This means the browser-facing HTTP loop is not reading frames directly from GStreamer. Instead, the path is:

```text
shared GStreamer source
  -> shared preview branch
  -> jpegenc
  -> appsink
  -> Go callback stores latest JPEG bytes in PreviewManager
  -> /api/previews/{id}/mjpeg loop reads cached frame
  -> HTTP multipart MJPEG to browser <img>
```

That makes the browser-streaming portion a distinct subsystem with its own potential costs:

- repeated HTTP writes per connected client,
- repeated frame copies into user-space buffers,
- flush frequency,
- and possible fan-out multiplication when multiple clients watch the same preview.

### 3. The upstream preview branch is already JPEG-encoding inside GStreamer

The shared preview consumer in `pkg/media/gst/shared_video.go:492-560` builds the preview branch with:

- `queue` (leaky, `max-size-buffers=2`),
- `videorate`,
- `videoscale`,
- `jpegenc`,
- and an `appsink` with `image/jpeg` caps.

So the browser transport is not sending raw frames. It is sending JPEG frames that are already compressed before entering Go. That is important because browser-path overhead can come from at least three places:

1. upstream preview branch work (`videorate`, `videoscale`, `jpegenc`),
2. Go-side copying / caching / per-client writes,
3. and browser-side rendering / decode.

The previous adaptive-preview work reduced part of (1), but it did not directly measure (2) with real attached browsers.

### 4. The frontend explicitly ensures and releases previews based on visible armed Studio sources

`ui/src/pages/StudioPage.tsx:347-352` computes `desiredPreviewSourceIds` from the active Studio tab, armed sources, and preview limit. Then `ui/src/pages/StudioPage.tsx:531-592` performs preview ensure/release work:

- ensure previews for desired sources not already owned,
- track preview ownership in Redux,
- release detached or stale previews,
- and release all owned previews on unmount (`594-600`).

This means the real browser path can differ from API-only harnesses in ways the earlier matrices did not cover:

- timing of preview creation and destruction,
- lifecycle churn when tab/source state changes,
- duplicate listeners from multiple tabs,
- and the fact that the browser itself stays attached to MJPEG endpoints continuously.

### 5. The preview UI uses simple `<img src="/api/previews/{id}/mjpeg">`

`ui/src/components/preview/PreviewStream.tsx:66-97` renders previews with a normal `<img>` element pointed at the MJPEG URL. This keeps the client implementation simple, but it also means:

- each open preview is a long-lived HTTP download,
- browser tab duplication can create extra preview consumers at the HTTP layer,
- and browser-visible preview behavior is tightly coupled to the MJPEG endpoint.

This is likely the single most important architectural fact for the new ticket.

### 6. WebSocket traffic is present, but not the media transport

`ui/src/features/session/wsClient.ts:1-96` shows that `/ws` carries logs, preview state metadata, session state, audio meter events, and disk status. It does **not** carry video frames. Therefore, if the browser-connected path is much hotter than API-only measurements, the likely media-specific culprit is still the MJPEG path rather than the WebSocket channel itself.

### 7. Metrics are now possible, but still too limited for this question

The new metrics layer consists of:

- the in-process registry in `pkg/metrics/metrics.go:13-188`,
- and `/metrics` rendering in `internal/web/handlers_metrics.go:9-20`.

Current production metrics mainly cover the shared bridge recorder and audio-level parse failures. That is a good start, but it is not yet enough to explain browser preview cost. We currently lack first-class metrics for:

- active MJPEG clients,
- preview bytes served,
- preview frames served per client,
- preview ensure/release rate,
- active preview sessions per source type,
- and queue-depth-like gauges for the browser-facing portion.

## Gap analysis

The current experiment suite answered many recorder/runtime questions, but left these gaps:

### Gap 1: no browser-attached matrix

The previous matrices mostly tested:

- direct GStreamer pipelines,
- standalone Go shared-runtime paths,
- and API-driven server interactions.

They did **not** fully model:

- one real browser tab with live MJPEG previews,
- multiple tabs,
- or Studio-page-driven preview lifecycle behavior.

### Gap 2: no dedicated camera browser-preview matrix

The existing performance work focused heavily on the real `2880x960` region/display path. It did not create a comparable focused matrix for the **camera preview transport** in the real browser-connected setup.

### Gap 3: no per-client preview serving metrics

Without metrics for MJPEG clients, bytes, and frames, Prometheus/Grafana cannot yet explain whether the server CPU spike aligns with:

- preview fan-out,
- specific source types,
- or recording-state transitions.

### Gap 4: no browser-tab hygiene measurement

We now know stale browser tabs or Playwright tabs can keep preview listeners alive. The previous work did not formally benchmark:

- zero browser clients,
- one browser client,
- two browser tabs of the same page,
- and combinations of desktop + camera clients.

That should now become an explicit part of the matrix.

## Proposed investigation architecture

This ticket should proceed in three tracks.

### Track A: observability extensions

Add preview-serving metrics so `/metrics` can answer browser-path questions.

Recommended first metrics:

```text
screencast_studio_preview_http_clients
screencast_studio_preview_http_streams_started_total
screencast_studio_preview_http_streams_finished_total
screencast_studio_preview_http_frames_served_total
screencast_studio_preview_http_bytes_served_total
screencast_studio_preview_http_flushes_total
screencast_studio_preview_active_total
screencast_studio_preview_ensure_total
screencast_studio_preview_release_total
screencast_studio_preview_frame_updates_total
```

Suggested labels:

```text
preview_id
source_type
source_id
result
client_kind   // optional, if we can identify browser/test clients cleanly
```

Start simple. If label cardinality becomes risky, collapse `preview_id`/`source_id` into source type and aggregate-only counters for the first pass.

### Track B: browser-driven measurement matrix

Create a new experiment suite that uses a real browser session as part of the measurement path.

Minimum scenario matrix:

1. desktop preview only, one tab
2. desktop preview + recording, one tab
3. camera preview only, one tab
4. camera preview + recording, one tab
5. desktop + camera preview only, one tab
6. desktop + camera preview + recording, one tab
7. desktop preview only, two tabs
8. desktop + camera preview, two tabs

For each scenario, capture:

- server PID CPU (`pidstat`),
- `/metrics` snapshots before/during/after,
- preview list snapshots,
- browser network requests,
- optional browser process CPU if useful,
- and resulting recording validity where recording is involved.

### Track C: comparison harnesses

Split measurements so we can compare:

1. backend only / no browser attached,
2. API-only harness,
3. browser attached with one tab,
4. browser attached with multiple tabs.

That is the clearest way to isolate the browser-streaming delta.

## Pseudocode and key flows

### Proposed server metrics wrapping for MJPEG handler

```go
func (s *Server) handlePreviewMJPEG(w http.ResponseWriter, r *http.Request) {
    labels := previewLabels(previewID, snapshot.SourceType)
    previewHTTPStreamsStartedTotal.Inc(labels)
    previewHTTPClients.Inc(labels)
    defer previewHTTPClients.Dec(labels)
    defer previewHTTPStreamsFinishedTotal.Inc(labels)

    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    var lastSeq uint64
    for {
        frame, seq, snapshot, ok := s.previews.LatestFrame(previewID)
        if !ok {
            return
        }
        if len(frame) > 0 && seq != lastSeq {
            lastSeq = seq
            n := writeMultipartJPEG(w, frame)
            previewHTTPFramesServedTotal.Inc(labels)
            previewHTTPBytesServedTotal.Add(labels, uint64(n))
            previewHTTPFlushesTotal.Inc(labels)
            flusher.Flush()
        }
        select {
        case <-r.Context().Done():
            return
        case <-ticker.C:
        }
    }
}
```

### Proposed browser-driven matrix harness shape

```text
for scenario in matrix:
  restart server in clean state
  start pidstat for server pid
  start metrics sampler loop (curl /metrics every second)
  launch browser with exactly N tabs
  drive Studio UI into requested sources / recording state
  hold steady-state window for M seconds
  capture network summary and /api/previews snapshot
  if recording scenario: stop recording and ffprobe outputs
  stop samplers
  summarize CPU + metrics + preview-client counts
```

## Phased implementation plan

### Phase 1: ticket scaffolding and current-state documentation

- create SCS-0015 docs,
- relate the key files,
- record current evidence and hypotheses,
- and define the browser-streaming measurement matrix.

### Phase 2: preview-serving metrics

Target files:

- `pkg/metrics/metrics.go`
- `internal/web/handlers_metrics.go`
- `internal/web/handlers_preview.go`
- optionally `internal/web/preview_manager.go`

Deliverables:

- counters and gauges for preview serving,
- low-cardinality labels where practical,
- and one or two tests that prove `/metrics` exposes the new families.

### Phase 3: new ticket-local browser matrix scripts

Expected new scripts under this ticket’s `scripts/` directory:

```text
01-restart-scs-web-ui.sh
02-browser-preview-metrics-sampler.sh
03-browser-preview-desktop-matrix/
04-browser-preview-camera-matrix/
05-browser-preview-desktop-camera-matrix/
06-browser-preview-multi-tab-matrix/
07-browser-preview-report-summary.md
```

The exact numbering can change, but the idea is to save every measurement harness and every raw result under the ticket tree.

### Phase 4: analysis and report

Write a proper report that includes:

- matrix results,
- explanation of where the browser delta appears,
- metrics plots / tables,
- comparison against earlier SCS-0014 backend-focused measurements,
- and optimization options ranked by likely impact and implementation risk.

## Testing and validation strategy

### Code validation

After adding metrics:

```bash
gofmt -w internal/web/handlers_preview.go pkg/metrics/metrics.go internal/web/handlers_metrics.go
go test ./internal/web ./pkg/metrics -count=1
go test ./... -count=1
```

### Runtime validation

For each browser-driven matrix run, save:

- `pidstat` output,
- `/metrics` snapshots,
- browser network request summaries,
- preview list snapshots,
- and `ffprobe` output when recordings are produced.

### Interpretation rule

Do not attribute CPU to “the browser streaming path” unless the comparison shows a meaningful delta between:

- no browser client,
- one browser client,
- and two browser clients,

while holding the backend capture/recording scenario otherwise constant.

## Risks, alternatives, and open questions

### Risks

1. **Metric cardinality:** per-preview labels can explode if we are not careful.
2. **Measurement pollution:** stale browser tabs or Playwright tabs can keep extra MJPEG listeners alive and distort results.
3. **Mixed bottlenecks:** browser-path CPU might combine preview serving cost with recorder-state cost, making attribution noisy unless the matrix is carefully staged.

### Alternatives

1. Replace MJPEG with another transport later, but that should come after this investigation unless the current results become overwhelmingly obvious.
2. Use browser-devtools or Chromium tracing as a second-stage tool if server-side metrics still leave ambiguity.
3. Add `node_exporter` / `process-exporter` and Grafana dashboards once the in-app preview metrics exist, so app counters can be correlated with process CPU.

### Open questions

1. Is the dominant browser-path cost on the server side the **HTTP/MJPEG fan-out**, or simply keeping the preview branch alive while a real browser is connected?
2. Does a camera preview produce meaningfully different browser-serving cost than desktop preview?
3. How much of the web-UI-only server spike comes from **one** tab versus duplicate tabs/listeners?
4. Is the right first optimization to further constrain preview profile during recording, or to optimize the MJPEG serving path itself?

## References

- `internal/web/routes.go:9-25`
- `internal/web/handlers_preview.go:14-149`
- `internal/web/preview_manager.go:95-220`
- `pkg/media/gst/preview.go:50-83`
- `pkg/media/gst/shared_video.go:492-560`
- `ui/src/pages/StudioPage.tsx:327-352`
- `ui/src/pages/StudioPage.tsx:467-600`
- `ui/src/components/preview/PreviewStream.tsx:12-97`
- `ui/src/features/session/wsClient.ts:1-96`
- `pkg/metrics/metrics.go:13-188`
- `internal/web/handlers_metrics.go:9-20`
