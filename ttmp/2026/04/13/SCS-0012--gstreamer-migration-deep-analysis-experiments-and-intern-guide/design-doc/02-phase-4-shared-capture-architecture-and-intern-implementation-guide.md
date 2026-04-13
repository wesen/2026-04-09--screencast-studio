---
Title: Phase 4 Shared Capture Architecture and Intern Implementation Guide
Ticket: SCS-0012
Status: active
Topics:
    - screencast-studio
    - backend
    - gstreamer
    - audio
    - video
    - transcription
    - screenshots
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/handlers_api.go
      Note: Recording start/stop and live audio effects endpoints
    - Path: internal/web/preview_manager.go
      Note: |-
        Preview lifecycle, leases, caching, and current suspend/restore behavior
        Preview lifecycle
    - Path: internal/web/server.go
      Note: |-
        Current recording-preview handoff logic that still exists because shared capture is not done yet
        Server-level preview handoff logic and Phase 4 removal constraints explained in the intern guide
    - Path: pkg/app/application.go
      Note: |-
        Application defaults and recording runtime wiring
        Application runtime defaults and recording handoff explained in the intern guide
    - Path: pkg/media/gst/preview.go
      Note: |-
        Current native GStreamer preview runtime using appsink JPEG frames
        Current native preview runtime architecture referenced in the intern guide
    - Path: pkg/media/gst/recording.go
      Note: |-
        Current native GStreamer recording runtime for video/audio jobs and live controls
        Current native recording runtime and Phase 4 constraints explained in the intern guide
    - Path: pkg/media/types.go
      Note: |-
        Media runtime seam shared by FFmpeg and GStreamer implementations
        Runtime seam and media session interfaces explained in the intern guide
    - Path: ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/16-web-gst-default-runtime-e2e/main.go
      Note: Real-defaults harness that proved duplicate capture without suspend/restore is still unsafe
    - Path: ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/17-go-gst-shared-video-tee-experiment/main.go
      Note: |-
        Focused tee-stop experiment showing MP4 finalization versus preview continuity tradeoff
        Experiment evidence section for tee stop/finalization tradeoffs
    - Path: ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/18-go-gst-shared-source-appsink-appsrc-bridge/main.go
      Note: |-
        Focused shared-source bridge experiment narrowing the remaining issue to appsrc segment/timestamp handling
        Experiment evidence section for shared-source bridge architecture
ExternalSources: []
Summary: A detailed intern-facing architecture and implementation guide for the Phase 4 shared capture problem in the Screencast Studio GStreamer migration. Explains the system, the current runtime seam, the exact Phase 4 failure mode, the experiment evidence, the recommended direction, and a step-by-step implementation plan.
LastUpdated: 2026-04-13T18:05:00-04:00
WhatFor: Onboard a new intern to the Screencast Studio media architecture and explain, in detail, why shared capture is the central remaining problem in the FFmpeg to GStreamer migration and how to approach it safely.
WhenToUse: Read this before attempting any Phase 4 work, before removing preview suspend/restore, or before designing a shared source registry, tee graph, or appsink/appsrc bridge for recording.
---


# Phase 4 Shared Capture Architecture and Intern Implementation Guide

## Executive Summary

This document explains the most important remaining architecture problem in the Screencast Studio FFmpeg-to-GStreamer migration: **how to share a single live video capture source between preview and recording without corrupting recording finalization or breaking live preview**.

If you are a new intern, the short version is this:

- The product already works well with GStreamer for **preview**, **recording**, **screenshots**, **live audio controls**, and **audio meter events**.
- The app currently still keeps an old workaround from the FFmpeg era: when recording starts, active previews are suspended, and when recording finishes, previews are restored.
- We tried deleting that workaround once GStreamer became the default runtime. That looked promising at first, but it failed during real recording finalization.
- The deeper problem is not “use GStreamer instead of FFmpeg.” The deeper problem is **source ownership and shutdown semantics for live media graphs**.
- Phase 4 is therefore not a cleanup-only phase. It is a real architecture phase.

The rest of this document teaches you:

- what Screencast Studio is and how its layers fit together,
- what parts of the system are stable and what parts are in flux,
- what exactly fails when preview and recording capture the same source independently,
- what experiments were run to isolate the failure,
- what design options exist,
- which direction is currently recommended,
- and how to implement it incrementally without destabilizing the app.

---

## Audience and How To Read This Guide

This guide is written for a new engineer or intern who is:

- comfortable in Go,
- not yet comfortable with GStreamer,
- and not yet familiar with Screencast Studio’s internal structure.

You should read it in this order:

1. **Sections 1–4** to understand what the app does and how requests flow through it.
2. **Sections 5–7** to understand the current GStreamer runtime seam and what already works.
3. **Sections 8–10** to understand the exact Phase 4 problem and the experiment evidence.
4. **Sections 11–15** for the proposed architecture, implementation plan, and debugging advice.

If you need the broader migration context as well, read this document together with:

- `ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/design-doc/01-gstreamer-migration-analysis-and-intern-guide.md`
- `ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/reference/01-diary.md`

---

## 1. What Screencast Studio Is

Screencast Studio is a desktop-oriented recording application. The user describes what they want to capture through a DSL, and the backend turns that into a running media system.

From the user’s perspective, the app can:

- preview a display, a window, a screen region, or a webcam,
- record one or more video sources,
- record and mix one or more audio sources,
- expose those operations through a web UI,
- and publish state and telemetry over WebSockets.

Historically, the media layer was implemented with FFmpeg subprocesses. The migration goal is to move that media layer to native GStreamer pipelines built in Go, while preserving the DSL, the web API, and the user-facing behavior.

That last clause is important: **the migration is intentionally backend-focused**. The product contract is mostly stable; the runtime internals are what changed.

### 1.1 Core user-facing features

The system today, after Phases 0–3, supports:

- preview of display / region / window / camera sources,
- video recording via native GStreamer,
- audio recording and mixing via native GStreamer,
- screenshot capture from preview sessions,
- live audio gain / compressor updates during recording,
- audio meter events through the existing websocket channel,
- max-duration stop,
- and proper session state transitions.

### 1.2 What is still unfinished

The big unfinished architecture item is **shared capture**:

- keeping preview active during recording,
- without duplicate capture of the same source,
- while still allowing clean MP4 finalization and clean shutdown.

That is the heart of Phase 4.

---

## 2. System Tour: Layers, Responsibilities, and Stable Boundaries

The easiest way to understand Screencast Studio is to look at it as five layers.

```text
+--------------------------------------------------------------+
| Browser UI                                                   |
| - DSL editor                                                 |
| - Preview panes                                              |
| - Recording controls                                         |
| - Logs / telemetry / audio meter                             |
+-------------------------------+------------------------------+
                                |
                                | HTTP + WebSocket
                                v
+--------------------------------------------------------------+
| internal/web                                                  |
| - HTTP handlers                                               |
| - PreviewManager                                              |
| - RecordingManager                                            |
| - EventHub / TelemetryManager                                 |
+-------------------------------+------------------------------+
                                |
                                | application service methods
                                v
+--------------------------------------------------------------+
| pkg/app                                                       |
| - Normalize DSL                                               |
| - Compile DSL                                                 |
| - Start recording via media runtime                           |
+-------------------------------+------------------------------+
                                |
                                | media runtime seam
                                v
+--------------------------------------------------------------+
| pkg/media                                                     |
| - interfaces for PreviewRuntime / RecordingRuntime            |
| - FFmpeg adapters                                             |
| - GStreamer runtimes                                          |
+-------------------------------+------------------------------+
                                |
                                | actual media engine
                                v
+--------------------------------------------------------------+
| FFmpeg subprocesses (legacy) / GStreamer pipelines (native)   |
+--------------------------------------------------------------+
```

### 2.1 Stable versus unstable layers

One of the most important migration insights is that not all layers are equally risky.

**Mostly stable layers:**

- DSL schema and normalization
- DSL compilation into jobs
- HTTP routes and browser contract
- session/event model at the web layer
- preview lease semantics at the manager level

**Still evolving layers:**

- native GStreamer runtime internals
- how live sources are owned and shared
- how recording stop/finalization interacts with preview continuity
- how Phase 4 removes FFmpeg-era handoff logic safely

### 2.2 The most important architectural boundary

The migration introduced `pkg/media/types.go`, which defines a runtime seam.

That seam allows the higher layers to think in terms of:

- “start a preview session,”
- “start a recording session,”
- “wait for results,”
- “adjust audio controls,”

without caring whether the implementation is FFmpeg or GStreamer.

That means most of the remaining architecture work should stay inside or near the media runtime layer.

---

## 3. File Map: Where To Read and Why

This section is intentionally repetitive. If you are new to the repo, you should keep it open while reading code.

### 3.1 Control-plane files

These files describe the application contract and should be treated as relatively stable.

- `pkg/dsl/types.go`
  - Defines the normalized DSL model.
  - Read this to understand source types, capture settings, output settings, and destination templates.

- `pkg/dsl/compile.go`
  - Turns normalized DSL into a `CompiledPlan`.
  - The key outputs for the runtime are `VideoJobs`, `AudioJobs`, and `Outputs`.

- `pkg/app/application.go`
  - The application service layer.
  - This is where the default recording runtime is selected and where `RecordPlan(...)` hands work to the media runtime.

### 3.2 Web-layer lifecycle files

These files are crucial to understanding what the browser expects and why some old behavior still exists.

- `internal/web/preview_manager.go`
  - Manages previews by source signature.
  - Stores latest preview frames.
  - Tracks leases.
  - Provides `SuspendAll(...)` and `RestoreSuspended(...)`, which are the FFmpeg-era workaround still preserved today.

- `internal/web/server.go`
  - Wires together preview and recording managers.
  - Contains the preview handoff logic that stores suspended previews at recording start and restores them after recording.

- `internal/web/handlers_api.go`
  - Recording start/stop handlers.
  - Also contains the live audio effects endpoint.

- `internal/web/handlers_preview.go`
  - Preview ensure/release/list handlers.
  - Also contains the screenshot endpoint.

- `internal/web/session_manager.go`
  - Receives recording events and translates them into current session state and websocket publication.

### 3.3 Media runtime seam files

These are the most important migration files.

- `pkg/media/types.go`
  - Defines `PreviewRuntime`, `PreviewSession`, `RecordingRuntime`, `RecordingSession`, and event/result types.

- `pkg/media/ffmpeg/preview.go`
  - FFmpeg adapter for preview.
  - Useful as a conceptual reference for how preview used to be supervised.

- `pkg/media/ffmpeg/recording.go`
  - FFmpeg adapter for recording.
  - Useful as a reference for session supervision behavior.

### 3.4 Native GStreamer runtime files

These files are where new architecture work will eventually land.

- `pkg/media/gst/preview.go`
  - Current native GStreamer preview runtime.
  - Uses appsink and stores latest JPEG frames.

- `pkg/media/gst/recording.go`
  - Current native GStreamer recording runtime.
  - Builds one pipeline per video job and one per audio mix job.
  - Handles session state, EOS stop, live audio controls, and audio meter events.

- `pkg/media/gst/pipeline.go`
  - Small helper functions for element linking and caps filters.

- `pkg/media/gst/bus.go`
  - GLib main loop backed bus watch helper.

### 3.5 Phase 4 evidence files

These are not production files, but they are critical reading.

- `scripts/16-web-gst-default-runtime-e2e/main.go`
  - Proved that removing preview suspend/restore while keeping duplicate live capture is unsafe.

- `scripts/17-go-gst-shared-video-tee-experiment/main.go`
  - Proved that tee branch stop/finalization for MP4 is more subtle than it first appears.

- `scripts/18-go-gst-shared-source-appsink-appsrc-bridge/main.go`
  - Proved that a bridge architecture may preserve preview continuity, but still has unresolved `appsrc` segment/timestamp handling.

---

## 4. The End-to-End Control Flow Today

To solve the Phase 4 problem, you need to understand two different flows:

- preview flow
- recording flow

### 4.1 Preview flow

A simplified version of preview today looks like this:

```text
Browser
  -> POST /api/previews/ensure
     -> PreviewManager.Ensure(dsl, sourceID)
        -> app.NormalizeDSL(...)
        -> resolve source
        -> compute preview signature
        -> reuse existing preview or create new one
        -> runtime.StartPreview(...)
           -> pkg/media/gst/preview.go
              -> build GStreamer preview pipeline
              -> appsink callback pushes JPEG bytes into manager
  -> Browser requests /api/previews/{id}/mjpeg
     -> PreviewManager serves latest frames as multipart MJPEG
```

### 4.2 Recording flow

A simplified version of recording today looks like this:

```text
Browser
  -> POST /api/recordings/start
     -> handlers_api.go
        -> PreviewManager.SuspendAll(...)   [still active today]
        -> RecordingManager.Start(...)
           -> app.CompileDSL(...)
           -> app.RecordPlan(...)
              -> recording runtime StartRecording(...)
                 -> pkg/media/gst/recording.go
                    -> create one worker per video job
                    -> create one worker per audio mix job
                    -> publish lifecycle events
  -> POST /api/recordings/stop
     -> session cancel
     -> recording workers send EOS and wait for finalization
     -> RecordingManager marks session finished/failed
     -> server.go restores suspended previews
```

### 4.3 Why preview is still suspended during recording

Because preview and recording are still implemented as **independent live captures of the same source**.

That is the key reason the handoff exists.

Without a true shared-source design, preview and recording are not two consumers of one source; they are two separate source pipelines aimed at the same device or screen.

That was tolerable in FFmpeg with explicit suspend/restore. It is still the current stable behavior in GStreamer.

---

## 5. GStreamer Concepts You Must Understand For Phase 4

If you are new to GStreamer, these are the specific concepts that matter most for this ticket.

### 5.1 Live sources

Examples in this repo:

- `ximagesrc`
- `v4l2src`
- `pulsesrc`

These are not finite files. They continuously produce buffers until stopped.

For a live source, stop behavior is usually not “close file and exit.” It is:

- send EOS to the right downstream path,
- allow encoders/muxers to flush,
- avoid killing unrelated consumers,
- then set state to NULL.

That last part is precisely where shared capture becomes hard.

### 5.2 Tee

A `tee` duplicates a stream into multiple branches.

```text
            +--> preview branch
source --> tee
            +--> recording branch
```

This sounds like the whole Phase 4 answer, but it is only half of it. A tee solves **sharing**. It does not automatically solve **independent stop/finalization semantics**.

### 5.3 EOS

EOS is an end-of-stream signal, but where you inject it matters.

If you inject EOS too far upstream in a shared graph, you may end the whole pipeline.
If you inject it too far downstream, your muxer may never finalize the file correctly.

This is the central media-engineering problem that surfaced in Phase 4.

### 5.4 appsink and appsrc

- `appsink` lets GStreamer hand samples to Go code.
- `appsrc` lets Go code push samples into a separate GStreamer pipeline.

An appsink/appsrc bridge is attractive because it decouples:

- shared source lifetime,
- preview lifetime,
- and recording finalization.

But it introduces a new class of problems:

- caps negotiation,
- timestamps,
- segment format,
- backpressure,
- and possibly copying raw frame data.

### 5.5 MP4 muxers are not magical

`mp4mux` needs a clean end-of-stream to write a valid MP4 trailer / `moov` atom.

If you interrupt the branch incorrectly, the file may exist but be invalid.

That is why “preview stayed alive” is not sufficient validation.
The recording file must also be valid.

---

## 6. What Has Already Been Achieved In The Migration

Before you try to solve Phase 4, be clear about what is already done. This prevents you from reinventing finished work.

### 6.1 Runtime seam

Completed:

- `pkg/media/types.go`
- FFmpeg adapters under `pkg/media/ffmpeg/`
- GStreamer runtime package under `pkg/media/gst/`

This means higher-level application code no longer talks directly to FFmpeg builders.

### 6.2 Native preview runtime

Completed in `pkg/media/gst/preview.go`:

- display preview
- region preview
- camera preview
- window preview, using geometry-first fallback instead of fragile XID-only capture
- latest-frame storage
- screenshot support through latest frame reuse

### 6.3 Native recording runtime

Completed in `pkg/media/gst/recording.go`:

- video recording pipelines for video jobs
- audio recording/mixing pipelines for audio jobs
- lifecycle states and events
- max-duration handling
- graceful stop via EOS
- live audio gain / compressor controls
- audio meter event plumbing

### 6.4 Web integration

Completed in `internal/web`:

- preview manager runtime seam
- screenshot endpoint
- audio effects endpoint
- websocket meter publication
- default runtime selection changed to GStreamer in stable code paths

### 6.5 What remains architecturally hard

The remaining hard problem is not “can we record?” or “can we preview?”
It is:

- **can we preview and record the same source at the same time, safely, through one shared capture abstraction?**

---

## 7. The Exact Phase 4 Problem Statement

Here is the Phase 4 problem in one sentence:

> Remove preview suspend/restore by making preview and recording share a single live capture source, while preserving correct recording finalization and stable preview continuity.

That sentence has four requirements hidden inside it.

### 7.1 Requirement A: one source capture

If preview and recording want the same display/window/camera, we should not start two separate live capture sources.

Why this matters:

- duplicate live capture is wasteful,
- duplicate capture may contend with devices or screen APIs,
- duplicate graphs can create shutdown/finalization races,
- and the browser-visible preview state becomes harder to reason about.

### 7.2 Requirement B: preview stays alive

When recording starts, the user should not see their preview disappear and come back.

That was an acceptable transitional workaround, but it is not the desired architecture.

### 7.3 Requirement C: recording finalizes correctly

A recording that fails to produce a valid MP4 is worse than a preview that temporarily disappears.

This is why we rolled back the naive Phase 4 attempt.

### 7.4 Requirement D: managers and APIs remain stable if possible

We want the browser, handlers, and managers to change as little as possible.
The best Phase 4 is one where:

- `PreviewManager` still manages preview leases,
- `RecordingManager` still manages recording sessions,
- and the big changes live mostly inside `pkg/media/gst`.

---

## 8. Why The Naive “Just Delete Suspend/Restore” Attempt Failed

We tried the obvious thing once native GStreamer preview and recording both existed:

- default app runtime -> GStreamer
- default preview runtime -> GStreamer
- remove suspend/restore
- keep preview active during recording

At first glance, this looked good.

The default-runtime harness (`scripts/16-web-gst-default-runtime-e2e/main.go`) showed that:

- preview could stay visible during recording,
- screenshot could work during recording,
- audio effect updates still worked,
- and audio meter websocket events still flowed.

Those are all encouraging signals.

But the harness also showed the real failure:

```text
timed out waiting for recording EOS
```

At the application level, that surfaced as recording finalization failure.

### 8.1 Why that failure matters more than the preview success

A preview is ephemeral. A recording file is the product.

A migration step that keeps preview alive but intermittently produces invalid or unfinalized output is not safe to ship.

That is why the server-side handoff was restored after the experiment.

### 8.2 The hidden architectural lesson

The lesson was not “GStreamer is broken.”
The lesson was:

> As long as preview and recording remain separate live captures of the same source, removing the handoff is premature.

That is what transformed Phase 4 from a cleanup task into an actual design task.

---

## 9. Experiment Evidence: What We Learned From Scripts 17 and 18

This section is important because it grounds the design in observed runtime behavior instead of wishful thinking.

## 9.1 Script 17: Shared tee branch-stop experiment

File:

- `scripts/17-go-gst-shared-video-tee-experiment/main.go`

Goal:

- create one live source,
- split it with `tee`,
- keep a preview branch alive,
- try different stop/EOS targets for the recording branch,
- see which ones both preserve preview and finalize MP4.

### 9.1.1 Simplified graph

```text
videotestsrc
  -> videoconvert
  -> tee
     -> preview queue -> videorate -> jpegenc -> appsink
     -> record queue  -> videorate -> x264enc -> mp4mux -> filesink
```

### 9.1.2 What happened

**EOS at queue-sink-pad / videorate / encoder:**

- MP4 became valid,
- but EOS propagated in a way that effectively ended the shared pipeline too,
- so preview continuity was lost.

Observed signal:

```text
BUS eos from shared-video-tee
```

**EOS at mux / filesink:**

- preview stayed alive longer,
- but the MP4 was invalid.

Observed signal:

```text
moov atom not found
Invalid data found when processing input
```

**Branch removal without proper branch-finalization:**

- preview could survive,
- but the recording file was invalid or empty.

### 9.1.3 Architectural meaning

A plain tee is not enough.
It solves duplication of the source, but not independent branch finalization for MP4.

That is a major design result.

## 9.2 Script 18: Shared source bridge experiment

File:

- `scripts/18-go-gst-shared-source-appsink-appsrc-bridge/main.go`

Goal:

- keep one shared source pipeline,
- feed preview directly from it,
- feed recording through a separate encoder pipeline using `appsink -> appsrc`,
- let the recording pipeline finalize independently.

### 9.2.1 Simplified graph

```text
Shared source pipeline:

source -> convert -> caps -> tee
                     -> preview queue -> jpegenc -> preview appsink
                     -> record queue  -> raw record appsink

Separate recorder pipeline:

appsrc -> videoconvert -> x264enc -> mp4mux -> filesink
```

### 9.2.2 Why this architecture is attractive

It separates the problem into two pieces:

- shared source lifetime and preview continuity stay in one pipeline,
- recording finalization happens in a different pipeline that can receive EOS safely.

In theory, that is cleaner than trying to finalize a single branch inside a long-lived shared tee graph.

### 9.2.3 What happened

This architecture got further, but still failed.

Important observations:

- preview continuity improved,
- recorder shutdown no longer obviously poisoned the shared source pipeline,
- but the recorder side hit `appsrc` time/segment issues.

Observed symptoms included:

```text
record bus error from record-appsrc: Internal data stream error.
```

and later:

```text
gst_segment_to_running_time: assertion 'segment->format == format' failed
output size: 850 bytes
duration=N/A
```

### 9.2.4 Architectural meaning

This does **not** mean the bridge architecture is wrong.
It means the bridge architecture is **not fully solved yet**.

The unresolved problem is narrower now:

- `appsrc` caps, timing, and segment semantics need a correct pattern.

That is a much better place to be than the earlier naive default-runtime experiment, because the shared-source lifetime problem appears separable from the recording finalization problem.

---

## 10. Design Goals and Non-Negotiable Invariants

Before discussing the proposed architecture, we need clear rules.

### 10.1 Design goals

The new design should:

- share one live capture source per source signature,
- let previews come and go independently,
- let recordings start and stop independently,
- keep preview active during recording,
- produce valid finalized files,
- fit into the existing runtime seam,
- and be testable in small slices.

### 10.2 Invariants

These are the things you should treat as contract-level truths.

#### Invariant 1: the recording file must finalize cleanly

If a design improves preview continuity but weakens finalization guarantees, it is not acceptable.

#### Invariant 2: shared capture must be keyed by a stable source signature

A display capture and a region capture are not always the same thing.
A window capture resolved to geometry may or may not be share-compatible depending on the normalized target.

A proper registry key needs to reflect the real capture shape.

#### Invariant 3: manager-level behavior should stay simple

The more complexity you can keep inside `pkg/media/gst`, the safer the migration.

#### Invariant 4: validation must be end-to-end

A pipeline that looks correct in isolation is not enough.
You must validate through:

- runtime behavior,
- HTTP handlers,
- and actual finalized output files.

---

## 11. Proposed Solution Direction

This section is intentionally honest: the design direction is stronger than a rough idea, but weaker than a finished implementation. That is okay.

## 11.1 Recommendation: a shared video source registry with pluggable consumer branches

The best current direction is:

- introduce a **CaptureRegistry** inside the GStreamer runtime layer,
- let it own long-lived shared **video source graphs** keyed by source signature,
- allow multiple consumer attachments,
- keep preview as one kind of consumer,
- and use a separate strategy for recording consumers that avoids the branch-finalization trap.

There are two viable recording-consumer strategies:

### Option A: pure tee branch recording consumer

```text
shared source -> tee -> preview branch
                   -> recording branch inside same pipeline
```

Pros:

- conceptually simple
- avoids appsink/appsrc copying
- everything stays inside one pipeline

Cons:

- branch-local MP4 finalization is currently the hard unsolved part
- experiments show easy EOS points poison the shared pipeline

### Option B: tee to raw appsink, then separate recorder pipeline via appsrc

```text
shared source -> tee -> preview branch
                   -> raw appsink bridge -> appsrc -> encoder pipeline
```

Pros:

- recording finalization can happen in a separate pipeline
- source/preview lifetime stays independent
- maps well to “shared capture source, independent recording session”

Cons:

- raw frame copying / backpressure overhead
- `appsrc` time/segment semantics need to be solved correctly
- more moving parts

## 11.2 Current recommendation between A and B

Based on current evidence, **Option B is the more promising direction**.

Reason:

- Option A already demonstrated a structural problem with MP4 finalization in a shared tee graph.
- Option B demonstrated a narrower integration bug rather than a structural impossibility.

That means if you are continuing Phase 4 work, you should currently treat Option B as the leading design hypothesis.

---

## 12. Proposed Internal Architecture

Here is the architecture this guide recommends you implement incrementally.

## 12.1 Core objects

### CaptureRegistry

A long-lived runtime-level registry.

Responsibilities:

- look up or create shared video source graphs,
- reference count active consumers,
- stop and remove a source graph when the last consumer detaches,
- isolate shared-source lifecycle from preview-manager bookkeeping.

Possible shape:

```go
type CaptureRegistry struct {
    mu      sync.Mutex
    sources map[string]*SharedVideoSource
}
```

### SharedVideoSource

Owns one live capture pipeline for one normalized source signature.

Responsibilities:

- create the source pipeline,
- expose consumer attachment APIs,
- own shared source lifecycle,
- publish source-level logs/errors,
- and destroy the pipeline when unused.

Possible shape:

```go
type SharedVideoSource struct {
    signature string
    source    dsl.EffectiveVideoSource

    pipeline *gst.Pipeline
    tee      *gst.Element
    watch    *busWatch

    mu        sync.Mutex
    refCount  int
    consumers map[string]ConsumerHandle
}
```

### PreviewConsumer

A consumer that wants JPEG preview frames.

Responsibilities:

- attach a preview branch,
- encode JPEG,
- expose latest frame,
- stop independently without killing the shared source.

### RecordingConsumer

A consumer that wants finalized file output.

Responsibilities differ depending on which recording strategy is chosen:

- pure tee branch recording, or
- appsink/appsrc bridge recording.

For now, assume bridge-based recording unless later evidence changes that.

---

## 12.2 Proposed topology

### Shared source side

```text
                         +--> preview consumer A
source -> normalize -> tee
                         +--> preview consumer B
                         +--> record-bridge consumer C (raw appsink)
```

### Recorder side (bridge approach)

```text
record-bridge consumer raw frames
    -> appsrc
    -> videoconvert
    -> x264enc
    -> mp4mux
    -> filesink
```

### Why this split is attractive

Because it assigns responsibilities cleanly:

- shared source graph: “capture and distribute live frames”
- preview consumer: “produce JPEG frames”
- recorder pipeline: “encode and finalize a file”

That is much closer to the real problem decomposition than the current per-job independent-capture model.

---

## 13. Suggested API Shape Inside `pkg/media/gst`

This section is not a final interface contract. It is a concrete starting point.

## 13.1 Registry methods

```go
type CaptureRegistry interface {
    AcquireVideoSource(ctx context.Context, source dsl.EffectiveVideoSource) (*SharedVideoSource, error)
    ReleaseVideoSource(signature string) error
}
```

## 13.2 Shared source methods

```go
type SharedVideoSource struct {
    // ...
}

func (s *SharedVideoSource) AttachPreviewConsumer(opts PreviewConsumerOptions) (PreviewConsumerHandle, error)
func (s *SharedVideoSource) AttachRecordingBridge(opts RecordingBridgeOptions) (RecordingBridgeHandle, error)
func (s *SharedVideoSource) CloseWhenUnused() error
```

## 13.3 Preview runtime integration

`pkg/media/gst/preview.go` would stop creating a whole standalone source pipeline for every preview session.
Instead, it would:

1. acquire the shared source from the registry,
2. attach a preview consumer branch,
3. expose `LatestFrame()` and `TakeScreenshot()` as it already does,
4. detach on stop.

## 13.4 Recording runtime integration

`pkg/media/gst/recording.go` would stop creating a whole standalone source pipeline for each video job.
Instead, for each video job, it would:

1. acquire the shared source,
2. attach a recording bridge consumer,
3. build or reuse the independent encoder pipeline,
4. stop that recording pipeline independently,
5. detach from the shared source afterward.

---

## 14. Pseudocode For The Recommended Incremental Plan

This section is deliberately concrete.

## 14.1 Acquire or reuse a shared source

```go
func (r *CaptureRegistry) AcquireVideoSource(ctx context.Context, source dsl.EffectiveVideoSource) (*SharedVideoSource, error) {
    signature := computeSharedSourceSignature(source)

    r.mu.Lock()
    if existing := r.sources[signature]; existing != nil {
        existing.refCount++
        r.mu.Unlock()
        return existing, nil
    }
    r.mu.Unlock()

    shared, err := newSharedVideoSource(ctx, source)
    if err != nil {
        return nil, err
    }

    r.mu.Lock()
    defer r.mu.Unlock()
    if existing := r.sources[signature]; existing != nil {
        // lost race; use existing and discard new one
        shared.CloseImmediately()
        existing.refCount++
        return existing, nil
    }
    shared.refCount = 1
    r.sources[signature] = shared
    return shared, nil
}
```

## 14.2 Attach preview consumer

```go
func (s *SharedVideoSource) AttachPreviewConsumer(opts PreviewConsumerOptions) (PreviewConsumerHandle, error) {
    // request tee pad
    // create queue -> videorate -> jpegenc -> appsink
    // add branch to shared pipeline
    // link tee pad to branch queue
    // sync state with parent
    // return handle exposing LatestFrame / Stop
}
```

## 14.3 Attach recording bridge consumer

```go
func (s *SharedVideoSource) AttachRecordingBridge(opts RecordingBridgeOptions) (RecordingBridgeHandle, error) {
    // request tee pad
    // create queue -> capsfilter(raw fixed format) -> appsink
    // appsink callback forwards raw samples to recording bridge object
    // sync branch state with parent
    // return handle with Stop() and sample callback lifecycle
}
```

## 14.4 Start recorder pipeline from bridge

```go
func StartBridgeRecorder(ctx context.Context, bridge RecordingBridgeHandle, outputPath string) (*BridgeRecorder, error) {
    // build appsrc -> videoconvert -> x264enc -> mp4mux -> filesink
    // configure fixed caps, live mode, timestamps, segment format
    // on each bridge sample, push raw frame to appsrc
    // on Stop(): end appsrc stream, wait for EOS on recorder pipeline
}
```

## 14.5 Stop flow

```go
func (r *BridgeRecorder) Stop(ctx context.Context) error {
    // 1. stop accepting new bridge frames
    // 2. call appsrc.EndStream()
    // 3. wait for recorder-pipeline EOS
    // 4. set recorder pipeline NULL
    // 5. detach recording bridge consumer from shared source
    // 6. release shared source from registry
    return nil
}
```

---

## 15. HTTP/API Surface That Must Stay Stable

One way interns get lost is by changing the wrong layer. This section exists to stop that.

## 15.1 Preview endpoints

Current important routes include:

- `POST /api/previews/ensure`
- `POST /api/previews/release`
- `GET /api/previews`
- `GET /api/previews/{id}/mjpeg`
- `GET /api/previews/{id}/screenshot`

These endpoints should ideally not need semantic changes for Phase 4.

## 15.2 Recording endpoints

Current important routes include:

- `POST /api/recordings/start`
- `POST /api/recordings/stop`
- `GET /api/recordings/current`
- `POST /api/audio/effects`

These should also remain stable.

## 15.3 WebSocket events

Important event flows include:

- session state changes
- process logs
- preview logs
- audio meter events

Again, these are higher-level product contracts and should remain stable if possible.

### 15.4 What should change at the web layer later

Only once shared capture works reliably should we remove:

- `PreviewManager.SuspendAll(...)`
- `PreviewManager.RestoreSuspended(...)`
- preview handoff storage in `internal/web/server.go`
- preview restore after recording finish

Until then, these are stable safety rails, not dead code.

---

## 16. Alternatives Considered

This is the part many design docs skip. Do not skip it.

## 16.1 Keep duplicate capture and live with suspend/restore forever

Pros:

- simplest operationally
- already works

Cons:

- not the desired UX
- wastes resources
- keeps FFmpeg-era architecture assumptions alive
- prevents true shared capture and more advanced live graph behavior

Decision:

- acceptable as temporary stable state,
- not acceptable as final Phase 4 outcome.

## 16.2 Pure tee-based shared graph with recording branch finalization in-place

Pros:

- elegant on paper
- one shared graph per source
- no appsrc copying

Cons:

- experiments showed that easy EOS points poison the shared pipeline
- branch-local finalization for MP4 is tricky and unresolved

Decision:

- not ruled out forever,
- but currently not the leading direction.

## 16.3 Shared source plus appsink/appsrc bridge to separate recording pipeline

Pros:

- clean separation between source lifetime and file finalization
- preview continuity becomes much easier to reason about
- better matches product-level lifecycle semantics

Cons:

- `appsrc` segment/timestamp handling still unresolved in our experiments
- may copy raw frames in Go

Decision:

- currently the best next-step architecture hypothesis.

## 16.4 Record to a different container first, remux after stop

Idea:

- use a more forgiving live-recording container internally,
- remux to MP4 after stop.

Pros:

- may reduce sensitivity to branch-finalization behavior

Cons:

- more complexity
- adds post-processing latency
- drifts away from the current direct-output contract

Decision:

- worth keeping in mind as fallback,
- but not the first thing to try.

---

## 17. Step-by-Step Implementation Plan For The Next Intern

This is the section you should execute against.

## 17.1 Step 1: do not delete stable handoff logic yet

Do **not** start by removing code from `internal/web/server.go`.

Instead:

- keep preview handoff intact,
- add new shared-capture plumbing behind the runtime seam,
- validate it in isolation first.

## 17.2 Step 2: implement a minimal internal `CaptureRegistry`

Scope it narrowly:

- video sources only,
- no audio sharing yet,
- no browser/API changes,
- no FFmpeg deletion yet.

Deliverables:

- new GStreamer-internal registry object
- keyed by normalized source signature
- can acquire/release a shared source graph

## 17.3 Step 3: move preview to shared source first

Preview is easier than recording.

Goal:

- keep preview runtime semantics the same,
- but source capture should come from the registry instead of building standalone source pipelines.

Validation:

- existing preview E2E harnesses should still pass.

## 17.4 Step 4: build a dedicated recording-bridge prototype inside `pkg/media/gst`

Do this in code the way script 18 did it in experiments.

But fix the missing part:

- correct `appsrc` time/segment behavior,
- correct backpressure policy,
- correct stop behavior.

Validation target:

- preview stays active,
- recording finalizes to valid MP4,
- stopping recording does not poison preview.

## 17.5 Step 5: add a narrow internal smoke harness before web-level integration

Do not jump straight to HTTP integration.

Create a small runtime harness that proves:

- shared source acquired once,
- preview branch attached,
- recorder bridge attached,
- valid MP4 output,
- preview continues after recorder stop,
- shared source stops only when last consumer detaches.

## 17.6 Step 6: integrate into `RecordingRuntime`

Only after the bridge works reliably should `pkg/media/gst/recording.go` use shared sources for video jobs.

Audio jobs can remain separate for now.

## 17.7 Step 7: rerun the real-defaults web harness

Use the existing Phase 4 harness idea again, but now against the true shared-source implementation.

You want all of these to be true at once:

- preview stays active during recording
- screenshot works during recording
- audio effects still work
- audio meter still works
- video file finalizes correctly
- preview still exists after recording completes

## 17.8 Step 8: only then remove preview suspend/restore

Only after the shared-source runtime path is proven should you remove:

- `SuspendAll(...)`
- `RestoreSuspended(...)`
- preview handoff storage
- preview restore callbacks on recording finish

This is the order that minimizes risk.

---

## 18. Validation Plan

This migration work must be evidence-driven.

## 18.1 Runtime-level validation

Use focused harnesses for:

- shared source + preview only
- shared source + one recording consumer
- shared source + preview + one recording consumer
- attach/detach ordering
- EOS stop correctness

## 18.2 Web-level validation

Use or extend:

- `scripts/11-web-gst-preview-e2e/`
- `scripts/14-web-gst-recording-e2e/`
- `scripts/15-web-gst-phase3-e2e/`
- `scripts/16-web-gst-default-runtime-e2e/`

## 18.3 Output validation

Never trust “pipeline stopped cleanly” by itself.

Always validate output files with:

- `ffprobe`
- non-zero file size
- reasonable duration
- no `moov atom not found`

## 18.4 Failure signatures to watch for

These strings matter:

```text
timed out waiting for recording EOS
BUS eos from shared-video-tee
moov atom not found
Internal data stream error
gst_segment_to_running_time: assertion 'segment->format == format' failed
```

If you see one of those, record exactly where in the stop path it happened.

---

## 19. Debugging Guide For The Intern

When a shared capture experiment fails, you need a method, not guesswork.

## 19.1 Questions to ask immediately

- Did preview continue?
- Did the recorder pipeline receive EOS?
- Did the shared source pipeline receive EOS unexpectedly?
- Did the muxer finalize the file?
- Did the file pass `ffprobe`?
- Did failure happen before stop, during stop, or after stop?

## 19.2 If preview dies unexpectedly

Suspect:

- EOS injected too far upstream
- tee branch stop poisoning shared graph
- branch removal order incorrect

## 19.3 If MP4 is invalid

Suspect:

- muxer never got a proper EOS/finalization path
- branch removed too early
- state was set to NULL before muxer finished flushing

## 19.4 If appsrc complains about internal data stream errors

Suspect:

- caps mismatch
- segment format mismatch
- timestamps or duration missing / inconsistent
- pushing after end-of-stream
- appsrc not configured the way the downstream pipeline expects

## 19.5 If you are stuck

Reduce the graph.

Do not debug a full browser-driven path first.
Use a tiny runtime experiment with:

- one source,
- one preview branch,
- one recording path,
- one stop mode,
- and explicit logs.

---

## 20. Suggested Diagrams For Future Revisions

If you extend this document later, consider converting these into actual drawings.

### 20.1 Current stable architecture

```text
preview request -> PreviewManager -> standalone preview pipeline
record request  -> RecordingManager -> standalone recording pipeline(s)

During recording:
  previews suspended
  recordings run
After recording:
  previews restored
```

### 20.2 Desired architecture

```text
PreviewManager ----+
                   |
RecordingRuntime --+--> CaptureRegistry --> SharedVideoSource --> tee
                                                         |         |
                                                         |         +--> preview consumer(s)
                                                         |
                                                         +------------> recording bridge consumer
                                                                               |
                                                                               v
                                                                       appsrc recorder pipeline
```

### 20.3 Unsafe architecture we already disproved

```text
PreviewManager -> independent source capture A
RecordingRuntime -> independent source capture B

Result:
  looks fine at first,
  but can fail during EOS/finalization in the real app path
```

---

## 21. Open Questions

These are still unresolved and should stay visible.

1. What is the correct go-gst `appsrc` pattern for bridging raw live frames into an H.264/MP4 recording pipeline here?
2. Can pure tee-branch finalization still be made safe with pad probes, valves, or a different branch topology?
3. Should shared capture initially support only display/region/window, with camera sharing added later?
4. Should the first shared-capture milestone target only preview + video recording, leaving audio fully separate?
5. Is a more forgiving intermediate recording container worth considering if `mp4mux` remains awkward in a shared branch model?

---

## 22. Practical Reading Checklist For A New Intern

Before writing code, read these in this order:

1. `pkg/media/types.go`
2. `pkg/app/application.go`
3. `internal/web/preview_manager.go`
4. `internal/web/server.go`
5. `pkg/media/gst/preview.go`
6. `pkg/media/gst/recording.go`
7. `scripts/16-web-gst-default-runtime-e2e/main.go`
8. `scripts/17-go-gst-shared-video-tee-experiment/main.go`
9. `scripts/18-go-gst-shared-source-appsink-appsrc-bridge/main.go`
10. `reference/01-diary.md` Steps 20 and 21

Then rerun these commands yourself:

```bash
go test ./... -count=1
./ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/17-go-gst-shared-video-tee-experiment.sh
./ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/18-go-gst-shared-source-appsink-appsrc-bridge.sh
```

Only after you have done that should you start changing Phase 4 production code.

---

## 23. Bottom Line

If you only remember one thing from this guide, remember this:

> Phase 4 is not “remove preview suspend/restore.”
> Phase 4 is “give preview and recording a true shared source architecture with correct independent stop semantics.”

That is the whole game.

Everything else is implementation detail.

---

## References

### Ticket documents

- `ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/design-doc/01-gstreamer-migration-analysis-and-intern-guide.md`
- `ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/reference/01-diary.md`
- `ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/tasks.md`
- `ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/changelog.md`

### Key source files

- `pkg/media/types.go`
- `pkg/app/application.go`
- `internal/web/preview_manager.go`
- `internal/web/server.go`
- `internal/web/handlers_api.go`
- `pkg/media/gst/preview.go`
- `pkg/media/gst/recording.go`

### Key experiment harnesses

- `scripts/16-web-gst-default-runtime-e2e/main.go`
- `scripts/17-go-gst-shared-video-tee-experiment/main.go`
- `scripts/18-go-gst-shared-source-appsink-appsrc-bridge/main.go`
