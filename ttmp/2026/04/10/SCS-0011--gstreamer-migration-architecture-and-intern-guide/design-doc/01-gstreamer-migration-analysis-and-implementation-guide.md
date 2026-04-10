---
Title: GStreamer migration analysis and implementation guide
Ticket: SCS-0011
Status: active
Topics:
    - screencast-studio
    - backend
    - frontend
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/handlers_api.go
      Note: Recording start API and preview suspend orchestration
    - Path: internal/web/preview_manager.go
      Note: Preview lifecycle
    - Path: internal/web/preview_runner.go
      Note: Current preview FFmpeg process execution and MJPEG parsing
    - Path: internal/web/server.go
      Note: Serve runtime wiring and preview handoff restoration after recording
    - Path: internal/web/session_manager.go
      Note: Recording session state and browser-facing lifecycle management
    - Path: pkg/app/application.go
      Note: Application boundary between DSL compilation and runtime execution
    - Path: pkg/dsl/compile.go
      Note: Compiled plan generation for video jobs
    - Path: pkg/dsl/normalize.go
      Note: Normalized DSL model and effective source/output settings
    - Path: pkg/recording/ffmpeg.go
      Note: Current FFmpeg argument construction for preview
    - Path: pkg/recording/run.go
      Note: Current recording subprocess supervision and stop sequencing
    - Path: proto/screencast/studio/v1/web.proto
      Note: Stable browser/server protobuf API contract
    - Path: ui/src/pages/StudioPage.tsx
      Note: Primary browser control flow for setup
ExternalSources: []
Summary: Detailed analysis and phased implementation guide for replacing the current FFmpeg-based media runtime with GStreamer while preserving the existing planning and web contracts.
LastUpdated: 2026-04-10T13:07:13.40611929-04:00
WhatFor: Explain the current FFmpeg-based architecture, identify migration constraints, and provide a detailed phased plan for moving screencast-studio to GStreamer.
WhenToUse: Use when onboarding engineers to the media backend, planning the GStreamer migration, or reviewing the future runtime design.
---


# GStreamer migration analysis and implementation guide

## Executive Summary

The current screencast-studio runtime is split into a clean planning layer and a brittle execution layer. The planning layer is relatively healthy: the DSL is normalized into an `EffectiveConfig`, compiled into a `CompiledPlan`, and exposed consistently through HTTP and protobuf contracts. The execution layer is where complexity has accumulated. Recording is implemented as one FFmpeg subprocess per output, preview is implemented as a separate FFmpeg subprocess per preview source, and the server/runtime layer must manually coordinate all of those independent processes through cancellation, `stdin` shutdown, frame polling, websocket event fanout, and preview suspend/restore behavior.

That split is the key fact an intern needs to understand. This is not a “rewrite the whole app” problem. It is a “replace the media runtime while preserving the existing planning, web API, and browser orchestration contracts” problem. The most practical migration is therefore to keep the DSL, compile model, and browser/server control flow stable, and replace the FFmpeg-specific process builders and runners with a GStreamer-backed media runtime that offers the same higher-level capabilities:

- compile DSL into executable jobs,
- start and stop a recording session,
- stream live preview frames into the browser,
- surface process or pipeline logs to the UI,
- preserve current session and preview lifecycle state semantics,
- keep `preview` and `recording` ownership boundaries intact.

The main recommendation in this document is to introduce a small media-runtime seam and migrate in phases. Phase 1 should create explicit runtime interfaces and a `gstreamer` package while leaving FFmpeg as the active implementation. Phase 2 should port preview first, because preview currently drives the most visible device-contention problems. Phase 3 should port recording and audio mixing. Phase 4 should remove the FFmpeg-specific code paths once parity, tests, and operational confidence are established.

## Problem Statement

The user request behind this ticket is clear: “port this over to use gstreamer, the ffmpeg process juggling is quite brittle and complex.” That statement matches the codebase.

Observed facts:

1. Recording is orchestrated as a set of independent subprocesses created from `CompiledPlan` jobs in [run.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go#L50). Each worker is a wrapper around `exec.Command("ffmpeg", ...)`, its pipes, and its stop logic.
2. Preview is orchestrated separately through `FFmpegPreviewRunner`, which launches another independent FFmpeg process and parses an MJPEG stream from stdout in [preview_runner.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_runner.go#L28).
3. The browser can request preview and recording from the same DSL and source set. The server has already needed extra logic to suspend and restore previews during recording start/stop in [handlers_api.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go#L75) and [server.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go#L88).
4. The preview and recording layers both have to solve their own subprocess lifecycle concerns. Recording does this in `ManagedProcess` and `stopProcesses` in [run.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go#L37) and [run.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go#L332). Preview does a separate version of the same thing in [preview_runner.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_runner.go#L48).
5. Media capture details are encoded directly as FFmpeg argv strings in [ffmpeg.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/ffmpeg.go#L14), [ffmpeg.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/ffmpeg.go#L36), and [ffmpeg.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/ffmpeg.go#L58).

The migration problem is therefore:

- how to replace the FFmpeg-specific runtime with GStreamer,
- without changing the externally visible DSL contract too early,
- without rewriting the whole web app,
- without regressing session state, preview state, logs, or planned-output previews,
- and without adding another brittle orchestration layer on top of the current one.

In plain terms: the codebase needs a new media engine, not a new product model.

## Scope

This document covers:

- the current runtime architecture,
- the exact files an intern must understand before changing anything,
- the migration constraints that must be preserved,
- a proposed GStreamer architecture,
- phased implementation guidance,
- testing and rollout advice.

This document does not cover:

- a complete implementation of GStreamer bindings,
- a finished choice between `gst-launch` subprocesses and native Go GStreamer bindings,
- non-Linux portability,
- a redesign of the DSL or UI product model.

## Current-State Analysis

### 1. The planning layer is already separated from execution

The application entry point is intentionally narrow. The `Application` service exposes discovery, DSL normalization, compilation, and recording in [application.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go#L45), [application.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go#L126), [application.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go#L150), and [application.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go#L168).

Important architectural fact:

- `CompileDSL` does not execute media. It normalizes DSL and builds a `CompiledPlan` at a point in time.
- `RecordPlan` takes an already compiled plan and hands it to the recording runtime.

That means the migration should preserve this shape.

Current compile flow:

```text
DSL text
  -> ParseAndNormalize()
  -> EffectiveConfig
  -> BuildPlan(now)
  -> CompiledPlan
  -> RecordPlan(plan)
  -> recording.Run(plan)
```

Relevant references:

- [normalize.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/normalize.go#L14)
- [compile.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/compile.go#L9)
- [application.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go#L203)

### 2. The DSL models execution intent, not FFmpeg internals

The normalized DSL carries source types, targets, capture settings, output settings, and destination templates in [normalize.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/normalize.go#L32) through [normalize.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/normalize.go#L92). The compiler produces `VideoJob`, `AudioMixJob`, and `PlannedOutput` records in [compile.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/compile.go#L10) through [compile.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/compile.go#L99).

This is a major asset for the migration. The DSL already describes:

- what to capture,
- where it comes from,
- what output settings matter,
- what files should exist when the session is done.

It does not require FFmpeg-specific concepts to be exposed to the user. That means GStreamer can be introduced as an execution backend while leaving the DSL stable for the first several phases.

### 3. FFmpeg-specific media graph construction lives in one file, but it leaks outward

The file [ffmpeg.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/ffmpeg.go#L14) is the current translation layer from normalized job description to concrete media runtime commands.

It builds:

- preview command lines in `BuildPreviewArgs`,
- video recording command lines in `buildVideoRecordArgs`,
- audio mixing command lines in `buildAudioMixArgs`,
- source-specific input arguments in `appendVideoInputArgs`.

That centralization is good, but the runtime still leaks FFmpeg assumptions in two ways:

1. The recording runner logs and starts `ffmpeg` directly in [run.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go#L382) through [run.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go#L410).
2. The preview runner also shells out to `ffmpeg` directly in [preview_runner.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_runner.go#L48).

This means there is no single “media backend” seam yet. The FFmpeg coupling exists in both `pkg/recording` and `internal/web`.

### 4. Recording is a multi-process supervisor, not a single pipeline

The heart of the brittleness is in [run.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go#L50). `recording.Run` does not create one capture graph. It creates many independent `ManagedProcess` values, one per output, and then supervises them with a session state machine.

Important runtime characteristics:

- one video job becomes one FFmpeg process,
- one audio mix job becomes one FFmpeg process,
- each process has its own stdout/stderr readers,
- stop happens by writing `q\n` to stdin in [run.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go#L544),
- cancellation and hard timeout are handled outside the media engine by Go concurrency,
- session state transitions are driven by subprocess exit events.

This is the current recording architecture:

```text
CompiledPlan
  -> video jobs: N
  -> audio jobs: M

for each job:
  build ffmpeg args
  start subprocess
  drain stdout/stderr
  wait for process exit

session supervisor:
  watch worker exit events
  watch context cancellation
  watch hard timeout
  write "q" to stdin on stop
  aggregate final state
```

This architecture explains why the code feels complex:

- capture logic and process management are intertwined,
- multiple outputs compete for the same devices independently,
- stop behavior depends on external process cooperation,
- there is no shared graph for preview and recording,
- errors surface as per-process failures rather than per-session pipeline diagnostics.

### 5. Preview is a second, separate media runtime

Preview has its own abstraction boundary in `PreviewRunner`, but its implementation is another FFmpeg subprocess wrapper in [preview_runner.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_runner.go#L22). It shells out to `ffmpeg`, reads MJPEG frames from stdout, reads log lines from stderr, and waits for process exit.

Preview manager behavior in [preview_manager.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go#L103) is important for migration because it is not merely a frame cache. It also owns:

- deduplication by source signature,
- lease counting,
- preview state publication,
- latest-frame caching for MJPEG HTTP streaming,
- suspend/restore across recording starts and stops,
- bounded shutdown semantics.

The preview path therefore has two responsibilities:

1. media acquisition and frame production,
2. serve/runtime state management.

A GStreamer migration should replace only the first responsibility at first. The second responsibility is already useful and should likely remain.

### 6. The web server/runtime already has a stable application boundary

The web runtime talks to an `ApplicationService` interface in [application.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/application.go#L11). That service already provides:

- discovery,
- DSL normalization,
- DSL compilation,
- recording from a compiled plan.

The server itself wires managers and HTTP handlers in [server.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go#L48) through [server.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go#L75), and routes in [routes.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/routes.go#L9) through [routes.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/routes.go#L23).

The most important server-level media-specific behavior is preview handoff on recording start and finish:

- suspend previews before recording in [handlers_api.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go#L96),
- restore previews after recording in [server.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go#L126).

This is the symptom of a deeper media-backend problem: preview and recording are independent device consumers.

### 7. The browser already assumes stable API and event contracts

The protobuf contract in [web.proto](/home/manuel/code/wesen/2026-04-09--screencast-studio/proto/screencast/studio/v1/web.proto#L59) through [web.proto](/home/manuel/code/wesen/2026-04-09--screencast-studio/proto/screencast/studio/v1/web.proto#L253) is quite rich. It includes:

- normalize and compile responses,
- recording session state,
- preview descriptors and preview list,
- audio and disk telemetry,
- server event streaming via websocket.

The React page in [StudioPage.tsx](/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx#L246) uses these contracts heavily. It:

- normalizes and compiles DSL,
- ensures and releases previews,
- starts and stops recordings,
- syncs structured editor state back into DSL text,
- consumes session and preview events,
- renders preview streams from `/api/previews/<id>/mjpeg`.

The migration implication is simple:

- the web API and websocket event shapes should remain stable if at all possible,
- preview transport can continue to be MJPEG-over-HTTP for now even if GStreamer is used internally,
- the browser should not need a major redesign for the backend migration.

## Gap Analysis

The current FFmpeg design is workable but mismatched to the product’s real needs.

### Gap 1: shared-device coordination is not a first-class concept

Current behavior:

- preview opens a device independently,
- recording opens the same device independently,
- the server now suspends previews as a workaround.

Why this is a gap:

- the product wants “same live scene, different consumers” semantics,
- the backend currently models “many unrelated capture processes.”

GStreamer advantage:

- a single capture branch can be tee’d into multiple sinks,
- preview and recording can share the same source pipeline,
- device contention becomes a graph design problem instead of an HTTP orchestration problem.

### Gap 2: stop semantics depend on process protocol, not pipeline control

Current behavior:

- recording stop writes `q\n` to FFmpeg stdin,
- preview stop cancels the process context,
- the session logic then waits for subprocess termination.

Why this is a gap:

- there is no common stop contract,
- graceful drain depends on external process conventions,
- final state depends on race timing between context cancellation and process exit.

GStreamer advantage:

- the runtime can own the pipeline object,
- stop can be expressed as EOS or state transition,
- lifecycle and finalization can stay within one runtime.

### Gap 3: media backend concerns are duplicated

Current behavior:

- `pkg/recording/run.go` has one subprocess orchestration model,
- `internal/web/preview_runner.go` has another.

Why this is a gap:

- the code repeats process-start, pipe-drain, wait, and cancellation patterns,
- improvements in one area do not automatically help the other.

GStreamer advantage:

- preview and recording can sit on top of a common runtime,
- one capture graph can feed both a frame sink and file sinks,
- logs and state can be normalized by one backend.

### Gap 4: output-level isolation is useful, but process-level isolation is expensive

Current behavior:

- one output equals one process.

Why this is good:

- failures are isolated,
- output paths are explicit,
- the compile model is simple.

Why this is bad:

- duplicate capture work,
- duplicate device opens,
- duplicate timestamp and synchronization behavior,
- more moving parts to supervise.

GStreamer migration must preserve the useful part:

- “compiled outputs are separate artifacts”

while removing the expensive part:

- “every output requires its own top-level media process.”

## Migration Goals And Non-Negotiable Invariants

Any GStreamer migration should preserve the following invariants.

### Product invariants

- The DSL remains the user-facing source of truth.
- Compile preview still returns planned outputs before recording starts.
- The browser continues to work through the current HTTP and websocket surfaces.
- Preview and recording can coexist from the user’s perspective without surprising regressions.

### Runtime invariants

- Session state transitions remain meaningful: `starting`, `running`, `stopping`, `finished`, `failed`.
- Preview state transitions remain meaningful: `starting`, `running`, `stopping`, `finished`, `failed`.
- The server can still shut down cleanly with bounded timeouts.
- Errors still surface in logs and user-visible state.

### Migration invariants

- Do not rewrite the DSL and the browser at the same time as the media backend.
- Introduce one seam for the media runtime rather than scattering GStreamer code directly into every manager.
- Keep preview migration and recording migration separable.

## Proposed Solution

### High-level architecture

Introduce a new media runtime layer that sits below `Application.RecordPlan` and below the preview runner abstraction, but above any concrete FFmpeg or GStreamer implementation.

Proposed shape:

```text
DSL / Compile layer
  pkg/dsl/*
  pkg/app/application.go

Media runtime abstraction
  pkg/media/runtime.go
  pkg/media/types.go

Implementations
  pkg/media/ffmpeg/*
  pkg/media/gstreamer/*

Web/server state managers
  internal/web/session_manager.go
  internal/web/preview_manager.go

Browser contracts
  proto/.../web.proto
  ui/src/pages/StudioPage.tsx
```

This is the migration target:

```text
browser
  -> HTTP / websocket
  -> web server
  -> ApplicationService
  -> CompiledPlan
  -> MediaRuntime
      -> preview session(s)
      -> recording session(s)
      -> frame callbacks
      -> structured logs
```

### Why a runtime seam is worth adding

This is one place where a thin abstraction is justified. The current code already has two backend-specific execution paths. A runtime seam is not “backwards compatibility.” It is the minimum structure needed to swap the current backend out safely.

Without this seam, a GStreamer migration would require:

- touching `pkg/recording/run.go`,
- touching `internal/web/preview_runner.go`,
- touching session lifecycle code,
- touching preview state code,
- and repeatedly threading GStreamer-specific assumptions through the app.

With the seam, only the media backend needs to vary.

## Proposed Runtime Interfaces

The exact interface names can change, but the responsibilities should look like this.

### 1. Recording runtime

```go
type RecordingRuntime interface {
    StartRecording(ctx context.Context, plan *dsl.CompiledPlan, opts RecordingOptions) (RecordingSession, error)
}

type RecordingSession interface {
    Wait() (*RecordingResult, error)
    Stop(ctx context.Context) error
}

type RecordingOptions struct {
    MaxDuration time.Duration
    EventSink   func(RecordingEvent)
    Logger      func(StructuredLog)
}
```

### 2. Preview runtime

```go
type PreviewRuntime interface {
    StartPreview(ctx context.Context, source dsl.EffectiveVideoSource, opts PreviewOptions) (PreviewSession, error)
}

type PreviewSession interface {
    Wait() error
    Stop(ctx context.Context) error
}

type PreviewOptions struct {
    OnFrame func([]byte)
    OnLog   func(StructuredLog)
}
```

### 3. Shared media event and log types

```go
type StructuredLog struct {
    Timestamp time.Time
    Label     string
    Stream    string
    Message   string
}

type RecordingEvent struct {
    Type       string
    State      string
    Process    string
    OutputPath string
    Reason     string
    Timestamp  time.Time
}
```

This keeps the server/session/preview managers stable while allowing:

- `ffmpeg` runtime now,
- `gstreamer` runtime next,
- and possibly hybrid rollout if needed.

## GStreamer Design Direction

### Recommendation: use native Go bindings or a controlled Go wrapper, not `gst-launch-1.0` shelling

There are two broad implementation options:

1. Shell out to `gst-launch-1.0` similarly to FFmpeg.
2. Use a native GStreamer API from Go, or a dedicated Go-side wrapper process that exposes structured control.

Recommendation:

- avoid a straight `gst-launch-1.0` subprocess migration if the goal is to reduce brittleness;
- prefer a native or semi-native runtime that can:
  - build pipelines programmatically,
  - listen for EOS and errors through a bus,
  - control start/stop without stdin tricks,
  - expose tees and appsinks cleanly for preview frames.

Reasoning:

- replacing FFmpeg subprocess juggling with GStreamer subprocess juggling only changes syntax, not architecture;
- the hard part here is not “different command lines,” it is “owning the media graph and lifecycle in-process.”

### Preview architecture in GStreamer

Preview is the best first target because it already has a narrow interface and it is the main source of device-contention pain.

Desired preview shape:

```text
source
  -> colorspace/format normalization
  -> scale / framerate reduction
  -> jpegenc
  -> appsink
  -> Go callback -> PreviewManager.storePreviewFrame()
```

The `PreviewManager` can stay almost entirely intact if the preview runtime continues to deliver:

- frame bytes,
- log or error events,
- a clean stop signal,
- a terminal error if the pipeline fails.

### Recording architecture in GStreamer

Recording should evolve from “one process per output” into “one runtime graph per session with controlled fanout.”

Possible target shape:

```text
video source A --\
video source B ----> per-source branch(es) -> encoder -> muxer -> filesink
video source C --/

audio source(s) -> mix branch -> encoder -> filesink
```

or, if the product wants stronger isolation during the first migration phase:

```text
one GStreamer pipeline per compiled output
but with shared runtime control and structured stop/error handling
```

The second option is less elegant but lower-risk as an intermediate step. It still removes FFmpeg-specific stop semantics and gives a more structured control surface, even if device-sharing is not fully optimized on day one.

### Recommended migration strategy for graph sharing

Do not try to solve “one perfect shared graph for all previews and recordings” on day one.

Instead:

1. Port preview to a dedicated GStreamer preview runtime.
2. Port recording to a dedicated GStreamer recording runtime with one pipeline per output or per source-group.
3. Once parity is achieved, consider a second-stage optimization that shares capture branches between preview and recording.

This keeps the migration tractable.

## Detailed File-Level Analysis

### [pkg/app/application.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go)

This is the orchestration boundary between “product model” and “runtime implementation.”

Key points:

- `NormalizeDSL` and `CompileDSL` are stable entry points and should remain so.
- `RecordPlan` is where a runtime seam can be inserted.
- Today it calls `recording.Run(...)` directly in [application.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go#L168).

Recommended change:

- replace direct `recording.Run(...)` calls with a configured `RecordingRuntime`.

### [pkg/dsl/normalize.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/normalize.go)

This file is not media-backend-specific. It validates the DSL and produces executable intent.

Key reason to preserve it:

- it defines the product’s current contract.

Possible future enhancement:

- add optional backend capability warnings if GStreamer and FFmpeg differ in support for certain settings.

### [pkg/dsl/compile.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/compile.go)

This file turns normalized config into outputs and jobs.

Migration guidance:

- keep `PlannedOutput` stable,
- keep `VideoJob` / `AudioMixJob` until the new runtime is ready,
- if the GStreamer runtime needs richer graph descriptions later, add new internal runtime planning structures after this layer rather than replacing the DSL compiler immediately.

### [pkg/recording/ffmpeg.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/ffmpeg.go)

This is the FFmpeg translation layer.

Migration guidance:

- eventually move this under `pkg/media/ffmpeg/`,
- create a sibling `pkg/media/gstreamer/`,
- avoid leaving FFmpeg-specific names in the long-term neutral package namespace.

### [pkg/recording/run.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go)

This file contains the session supervisor and the process wrapper. It is both valuable and problematic.

Valuable:

- it already expresses the session lifecycle clearly,
- it emits structured events the server depends on,
- it centralizes stop and timeout handling.

Problematic:

- it assumes subprocess-based workers,
- it assumes FFmpeg `stdin` semantics,
- it equates worker control with OS process control.

Migration guidance:

- preserve the session-state logic conceptually,
- move process-specific behavior behind a runtime-specific session object,
- let `Stop()` become pipeline-aware rather than stdin-aware.

### [internal/web/preview_runner.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_runner.go)

This file is the preview-specific backend seam today.

Migration guidance:

- keep the `PreviewRunner` interface concept,
- replace `FFmpegPreviewRunner` with a `GStreamerPreviewRunner`,
- ensure the new implementation still pushes JPEG frame bytes or another browser-consumable transport that `PreviewManager` can expose.

### [internal/web/preview_manager.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go)

This file should mostly survive the migration.

Why:

- it owns preview identity, leases, latest-frame caching, suspend/restore, and HTTP-facing state semantics,
- those are product/runtime orchestration concerns, not FFmpeg concerns.

Potential refactors:

- rename the runner field to `runtime` or keep `runner` if it stays preview-specific,
- consider adding explicit stop timeout handling at the preview session layer if the GStreamer runtime can block.

### [internal/web/session_manager.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/session_manager.go)

This file is also worth preserving.

Why:

- it owns the browser-facing session model,
- it publishes session and log events,
- it does not care whether the backend is FFmpeg or GStreamer so long as the event contract remains meaningful.

Migration guidance:

- do not rewrite this file first,
- instead adapt the runtime event model so this manager can continue to consume it.

### [internal/web/server.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go)

This file should not become more media-aware during migration.

Guidance:

- keep preview handoff behavior while preview and recording still contend,
- if a later GStreamer design allows true shared capture, then this handoff can be simplified or removed,
- but do that only after runtime sharing is proven.

### [internal/web/handlers_api.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go)

These handlers define the browser contract.

Guidance:

- keep the HTTP verbs and payloads stable,
- avoid leaking GStreamer-specific details into HTTP responses,
- preserve current error semantics where possible.

### [ui/src/pages/StudioPage.tsx](/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx)

This file is long, but its core responsibilities are stable:

- request discovery,
- maintain structured/DSL editor state,
- ensure previews,
- start and stop recording,
- show logs and outputs.

Migration guidance:

- keep this page mostly unchanged in the first phases,
- only revisit it if preview transport changes materially.

## Proposed Implementation Phases

### Phase 0: preparation and non-functional scaffolding

Objective:

- create the migration seam without changing runtime behavior.

Tasks:

1. Introduce a runtime package such as `pkg/media`.
2. Define recording and preview runtime interfaces.
3. Add an FFmpeg-backed implementation under `pkg/media/ffmpeg`.
4. Change `pkg/app.Application` and preview wiring to depend on interfaces, not direct FFmpeg helpers.

Pseudocode:

```go
type Application struct {
    recordingRuntime media.RecordingRuntime
    previewRuntime   media.PreviewRuntime
}

func New() *Application {
    return &Application{
        recordingRuntime: ffmpeg.NewRecordingRuntime(),
        previewRuntime:   ffmpeg.NewPreviewRuntime(),
    }
}
```

Why this phase matters:

- it lowers the cost of the later migration,
- it proves the interface is viable while the system still behaves the same.

### Phase 1: GStreamer preview runtime

Objective:

- replace FFmpeg preview first.

Tasks:

1. Implement `GStreamerPreviewRunner` or `GStreamerPreviewRuntime`.
2. Produce JPEG frames or an equivalent frame callback into `PreviewManager`.
3. Preserve preview log and terminal error handling.
4. Keep MJPEG HTTP output stable for the browser.

Pseudocode:

```go
func (r *GStreamerPreviewRuntime) StartPreview(ctx context.Context, src dsl.EffectiveVideoSource, opts PreviewOptions) (PreviewSession, error) {
    pipeline := buildPreviewPipeline(src)
    wireBusHandlers(pipeline, opts.OnLog)
    wireAppSink(pipeline, func(buf []byte) {
        opts.OnFrame(buf)
    })
    pipeline.SetState(PLAYING)
    return newPreviewSession(pipeline), nil
}
```

Acceptance criteria:

- camera preview works,
- display preview works,
- region preview works,
- preview suspend/restore still works,
- browser preview UI is unchanged.

### Phase 2: GStreamer recording runtime

Objective:

- replace FFmpeg recording while preserving `CompiledPlan`.

Tasks:

1. Implement a `GStreamerRecordingRuntime`.
2. Map each `VideoJob` and `AudioMixJob` into pipeline definitions.
3. Emit runtime events compatible with `RecordingManager`.
4. Implement graceful stop using EOS or state transitions.

Important decision point:

- either one pipeline per output for lower-risk parity,
- or a shared graph for higher payoff but higher complexity.

Recommended first step:

- one pipeline per output or per source-group, controlled by Go-side session objects.

Why:

- it reduces migration risk,
- it avoids solving graph sharing too early,
- it still removes FFmpeg-specific brittleness.

### Phase 3: unify preview and recording source ownership

Objective:

- reduce or eliminate preview suspend/restore if GStreamer sharing makes it unnecessary.

Tasks:

1. Audit whether preview and recording can share source branches safely.
2. If yes, introduce a capture graph registry keyed by source signature.
3. Fan out shared source branches to preview sinks and recording sinks.

Diagram:

```text
capture registry
  source signature -> live source graph
      -> preview tee branch -> jpeg/appsink
      -> recording tee branch -> encoder/mux/filesink
```

Caution:

- do not start here,
- only do this after preview and recording are already stable on GStreamer independently.

### Phase 4: remove FFmpeg-specific code

Objective:

- delete dead code and simplify the codebase.

Tasks:

1. Remove `pkg/recording/ffmpeg.go`.
2. Remove direct `exec.Command("ffmpeg", ...)` paths.
3. Collapse stop logic onto the new runtime session interfaces.
4. Revisit preview handoff code and delete it if no longer needed.

## Testing And Validation Strategy

### Unit tests

Add or preserve tests for:

- DSL normalization,
- compile outputs,
- preview manager lease and suspend/restore behavior,
- recording manager state transitions,
- media-runtime interface adapters.

Suggested new tests:

1. preview runtime emits frames into `PreviewManager` without leaking backend details,
2. recording runtime stop transitions session state correctly on cancellation,
3. recording runtime reports structured errors in a way `RecordingManager` can surface.

### Integration tests

Existing `internal/web/server_test.go` is already a useful seam because it validates browser-facing behavior with fake application and fake preview runners.

Migration guidance:

- add fake GStreamer runtime implementations behind the new interfaces,
- preserve the same server tests so the API contract is checked independently of the real backend.

### Manual smoke tests

The intern should verify at least:

1. display preview start/release,
2. camera preview start/release,
3. display + camera preview together,
4. start recording with active previews,
5. stop recording,
6. bounded recording timeout,
7. browser refresh while recording,
8. server shutdown during recording,
9. planned output names match produced files,
10. logs surface meaningful pipeline failures.

## Risks

### Risk 1: Go GStreamer bindings increase build complexity

Why it matters:

- FFmpeg subprocesses are operationally simple to install against,
- native GStreamer bindings can introduce cgo and system dependency concerns.

Mitigation:

- evaluate dependency model early,
- prototype the bindings before large-scale refactoring,
- document dev and CI setup explicitly.

### Risk 2: a naive GStreamer port preserves the same architectural problems

Why it matters:

- if the code just shells out to `gst-launch-1.0`, lifecycle complexity may remain high.

Mitigation:

- prefer owning pipeline lifecycle in-process or through a structured wrapper,
- do not migrate syntax without changing control semantics.

### Risk 3: preview migration may force browser transport decisions

Why it matters:

- the browser currently expects MJPEG over HTTP.

Mitigation:

- keep MJPEG generation server-side for phase 1,
- postpone any WebRTC/WebSocket/frame transport redesign until after backend stability.

### Risk 4: audio is harder than video

Why it matters:

- today’s FFmpeg audio mix path is simple but centralized,
- GStreamer audio mixing and negotiation can be more subtle.

Mitigation:

- treat audio mixing as its own milestone,
- do not assume preview parity implies recording parity.

## Alternatives Considered

### Alternative 1: keep FFmpeg and just refactor the Go process supervisor

Pros:

- smallest dependency delta,
- less build-system change.

Cons:

- does not solve the fundamental “many unrelated media processes” shape,
- still depends on FFmpeg shutdown and device semantics,
- still duplicates preview and recording backend logic.

Assessment:

- useful as cleanup,
- not a real answer to the user’s migration goal.

### Alternative 2: shell out to `gst-launch-1.0`

Pros:

- faster initial experimentation,
- lower up-front binding complexity.

Cons:

- too similar to the current subprocess architecture,
- likely preserves much of the brittleness,
- harder to express structured runtime control.

Assessment:

- acceptable for prototypes,
- not the recommended production target.

### Alternative 3: rewrite the whole application around GStreamer immediately

Pros:

- clean slate.

Cons:

- very high regression risk,
- throws away working planning and web layers,
- too much surface area for an intern-led migration.

Assessment:

- not recommended.

## Open Questions

1. Which Go GStreamer binding or wrapper strategy should be used in this repository?
2. Is Linux-only support acceptable for the first GStreamer milestone?
3. Should the first recording migration preserve one-output-per-pipeline semantics, or invest immediately in shared graph fanout?
4. How much of preview suspend/restore should survive phase 2 as a fallback safety mechanism?
5. Do we want to preserve the `pkg/recording` package name, or split neutral runtime abstractions into a new `pkg/media` package and leave `pkg/recording` as product-level session orchestration?

## Intern Implementation Guide

This section is intentionally direct. If you are the intern implementing this migration, follow this order.

### Step 1: read these files in order

1. [application.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go)
2. [normalize.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/normalize.go)
3. [compile.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/compile.go)
4. [ffmpeg.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/ffmpeg.go)
5. [run.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go)
6. [preview_runner.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_runner.go)
7. [preview_manager.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go)
8. [session_manager.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/session_manager.go)
9. [server.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go)
10. [handlers_api.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go)
11. [handlers_preview.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go)
12. [web.proto](/home/manuel/code/wesen/2026-04-09--screencast-studio/proto/screencast/studio/v1/web.proto)
13. [StudioPage.tsx](/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx)

### Step 2: do not change the browser contract first

The browser is not the main problem. The media backend is the problem. Keep the browser stable while you change the runtime below it.

### Step 3: create the runtime seam before writing GStreamer code

If you start by dropping GStreamer directly into `run.go` and `preview_runner.go`, the migration will sprawl. Create the backend seam first.

### Step 4: port preview before recording

Preview is the smallest end-to-end path:

- normalize DSL,
- resolve source,
- start runtime,
- push frames,
- stop runtime,
- surface errors.

If you cannot make preview stable, you are not ready to migrate recording.

### Step 5: treat session/preview managers as consumers, not rewrite targets

These managers are already valuable. Adapt the runtime to them first. Rewrite them only if the new backend proves that the old state model is wrong.

## Pseudocode Walkthrough

### Current preview flow

```go
cfg := app.NormalizeDSL(dslBody)
source := findPreviewSource(cfg, sourceID)
signature := computePreviewSignature(source)

if preview already exists:
    increment leases
    return snapshot

start ffmpeg preview process
read stderr -> publish preview.log
read jpeg stdout -> storePreviewFrame
wait for exit -> finishPreview
```

### Proposed preview flow

```go
cfg := app.NormalizeDSL(dslBody)
source := findPreviewSource(cfg, sourceID)
signature := computePreviewSignature(source)

if preview already exists:
    increment leases
    return snapshot

session := gstreamerPreviewRuntime.StartPreview(ctx, source, {
    OnFrame: storePreviewFrame,
    OnLog:   publishPreviewLog,
})

wait for session.Wait()
finishPreview(session result)
```

### Proposed recording flow

```go
plan := app.CompileDSL(dslBody)

session := recordingRuntime.StartRecording(ctx, plan, {
    MaxDuration: maxDuration,
    EventSink:   publishSessionEvent,
    Logger:      publishStructuredLog,
})

on cancel:
    session.Stop(graceCtx)

result, err := session.Wait()
publish final session state
```

## Recommended Initial Package Layout

```text
pkg/media/
  types.go
  recording.go
  preview.go

pkg/media/ffmpeg/
  recording_runtime.go
  preview_runtime.go
  argv_builders.go

pkg/media/gstreamer/
  recording_runtime.go
  preview_runtime.go
  pipeline_builders.go
  bus.go
  appsink.go
```

If you adopt this layout, then:

- `pkg/app` depends on `pkg/media`,
- `internal/web` depends on `pkg/media` preview interfaces indirectly through runtime wiring,
- FFmpeg and GStreamer stay implementation details.

## References

- [pkg/app/application.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go)
- [pkg/dsl/normalize.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/normalize.go)
- [pkg/dsl/compile.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/compile.go)
- [pkg/recording/ffmpeg.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/ffmpeg.go)
- [pkg/recording/run.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go)
- [internal/web/application.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/application.go)
- [internal/web/preview_runner.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_runner.go)
- [internal/web/preview_manager.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go)
- [internal/web/session_manager.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/session_manager.go)
- [internal/web/server.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go)
- [internal/web/handlers_api.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go)
- [internal/web/handlers_preview.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go)
- [internal/web/handlers_ws.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go)
- [internal/web/routes.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/routes.go)
- [proto/screencast/studio/v1/web.proto](/home/manuel/code/wesen/2026-04-09--screencast-studio/proto/screencast/studio/v1/web.proto)
- [ui/src/pages/StudioPage.tsx](/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx)

## Proposed Solution

<!-- Describe the proposed solution in detail -->

## Design Decisions

<!-- Document key design decisions and rationale -->

## Alternatives Considered

<!-- List alternative approaches that were considered and why they were rejected -->

## Implementation Plan

<!-- Outline the steps to implement this design -->

## Open Questions

<!-- List any unresolved questions or concerns -->

## References

<!-- Link to related documents, RFCs, or external resources -->
