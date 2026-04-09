---
Title: Screencast Studio System Design
Ticket: SCS-0001
Status: active
Topics:
    - backend
    - frontend
    - video
    - audio
    - dsl
    - cli
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/screencast-studio/main.go
      Note: Root binary entrypoint for the new CLI
    - Path: go.mod
      Note: Root module for the new CLI-first implementation
    - Path: jank-prototype/dsl.md
      Note: Current DSL examples that motivate the intermediate spec
    - Path: jank-prototype/main.go
      Note: Prototype backend monolith and FFmpeg orchestration baseline
    - Path: jank-prototype/research.md
      Note: X11 window enumeration and capture research
    - Path: jank-prototype/web/app.js
      Note: Prototype frontend rendering and polling baseline
    - Path: jank-prototype/web/index.html
      Note: Prototype UI structure baseline
    - Path: pkg/app/application.go
      Note: Application boundary that now compiles DSL files into execution plans
    - Path: pkg/cli/record.go
      Note: Top-level record command wiring
    - Path: pkg/cli/root.go
      Note: Root Glazed/Cobra command tree
    - Path: pkg/cli/setup/compile.go
      Note: Setup compile command that exposes compiled outputs on the CLI
    - Path: pkg/discovery/service.go
      Note: Concrete command-backed discovery for displays
    - Path: pkg/discovery/types.go
      Note: Typed discovery descriptors used by the CLI milestone
    - Path: pkg/dsl/compile.go
      Note: Compiled plan generation from the normalized setup DSL
    - Path: pkg/dsl/normalize.go
      Note: Setup DSL normalization and validation rules
    - Path: ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/sources/local/screencast-studio-v2.jsx.jsx
      Note: Imported UI mock defining the target control surface
ExternalSources: []
Summary: Detailed architecture and implementation guide for a CLI-first Go screencast studio built around discovery, compile, and record flows with an intermediate setup DSL.
LastUpdated: 2026-04-09T13:12:49.178772118-04:00
WhatFor: Exhaustive architecture and implementation guide for the screencast studio backend, DSL, discovery, compile, and capture runtime, with web explicitly deferred.
WhenToUse: Use when implementing the CLI-first recorder, reviewing the proposed architecture, or onboarding a new engineer to the project.
---




# Screencast Studio System Design

## Executive Summary

This document proposes a clean replacement for the prototype in [jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go), [jank-prototype/web/app.js](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/app.js), and [jank-prototype/web/index.html](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/index.html). The new system starts as a local Go CLI with Glazed commands and an intermediate DSL that separates user intent from runtime execution details. The browser control surface remains part of the longer-term design, but it is deferred to a follow-up ticket after the CLI primitives are proven.

The architectural center of gravity is the setup DSL compiler. The CLI or future UI should edit a draft setup. The backend should normalize that draft into a canonical `StudioSpec`. The compiler should resolve that spec plus the current machine inventory into a `CompiledPlan`. The runtime should execute that plan by supervising recorder workers. This separation keeps the command-line workflow simple now, makes the backend testable, and prevents OS-specific details from leaking into saved setups.

For the current milestone, the most pragmatic media strategy is:

- Use platform-specific discovery adapters for displays, windows, cameras, and audio devices.
- Use FFmpeg subprocesses for recording workers.
- Support `discover`, `compile`, and `record` as first-class CLI verbs.
- Persist the setup DSL as YAML or JSON.
- Defer preview transport, browser state, and web event delivery until the CLI primitives are stable.

This is intentionally not a backwards-compatibility design. The prototype is evidence and inspiration, not a compatibility target. We should preserve the good ideas, not the monolith.

## Problem Statement

The prototype proves that the basic idea works: accept a DSL, preview video sources, and spawn FFmpeg processes to record video and mixed audio. However, it also combines too many concerns in one process and one file. The current prototype:

- mixes HTTP handlers, config parsing, normalization, process supervision, preview streaming, and output path rendering in one Go file ([jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L31));
- uses one static HTML page and one handwritten JS file instead of a structured frontend ([jank-prototype/web/index.html](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/index.html#L10), [jank-prototype/web/app.js](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/app.js#L1));
- starts one preview FFmpeg process per browser request rather than maintaining a source-centric preview lifecycle ([jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L385));
- encodes capture behavior directly in handler code instead of in a reusable runtime plan ([jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L769));
- already exposes future scope pressure around region presets, window selection, audio features, and follow-resize behavior ([jank-prototype/dsl.md](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/dsl.md#L165), [jank-prototype/README.md](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/README.md#L16)).

The desired product is closer to a small local OBS-like controller:

- The user can discover and select display, region, window, camera, and audio sources.
- The user can arm or disarm sources, adjust output parameters, and record separate files according to templates.
- The backend should be structured well enough that an intern can add features without opening a giant ball of stateful handler code.
- The first delivered milestone should be runnable and testable from the terminal before any web work starts.

## Scope

### In scope

- Local-first screencast studio backend in Go.
- Glazed commands and flags for process startup and maintenance commands.
- Intermediate DSL plus compiled runtime plan.
- Live discovery for displays, windows, cameras, and microphones.
- Separate file outputs by template, plus mixed or selected audio outputs.
- CLI-first user flows for source discovery, setup compilation, and recording execution.

### Explicitly out of scope for version 1

- Cross-platform parity on day one.
- Nonlocal multi-user control.
- Full timeline editing.
- A general-purpose scene compositor.
- Legacy compatibility with the exact prototype HTTP schema.
- The web control frontend. That work should move to a follow-up ticket once the CLI primitives are stable.

## Current-State Analysis

### What exists today

The prototype exposes a handful of handlers on one `http.ServeMux`, stores all mutable app state in one `App` struct, and serves a static web folder from the embedded filesystem ([jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L181)). The backend accepts a raw YAML or JSON config body on `/api/config/apply`, normalizes it immediately, and stores only the resulting effective config in memory ([jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L248), [jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L504)).

The frontend is intentionally simple. The HTML declares panels for the DSL textarea, status, preview grid, warnings, and logs ([jank-prototype/web/index.html](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/index.html#L23)). The JS polls `/api/state`, renders preview cards from server-provided preview URLs, and posts start and stop actions back to the server ([jank-prototype/web/app.js](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/app.js#L46), [jank-prototype/web/app.js](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/app.js#L156)).

The prototype DSL already contains the right shape of the problem. It describes destination templates, screen defaults, camera defaults, audio defaults, and per-source configuration blocks ([jank-prototype/dsl.md](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/dsl.md#L1)). This is the strongest part of the current work and should become the center of the new architecture.

### What the imported UI mock adds

The imported JSX file shows the longer-term product shape more clearly than the current prototype UI. In [sources/local/screencast-studio-v2.jsx.jsx](/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/sources/local/screencast-studio-v2.jsx.jsx#L201), the main surface is a grid of source cards with preview thumbnails, a source-type selector, and armed and solo toggles. In [sources/local/screencast-studio-v2.jsx.jsx](/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/sources/local/screencast-studio-v2.jsx.jsx#L265), the mock introduces an output-parameter panel for format, frame rate, quality, audio settings, multi-track mode, and destination. In [sources/local/screencast-studio-v2.jsx.jsx](/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/sources/local/screencast-studio-v2.jsx.jsx#L321), it adds a microphone panel and status panel with metering, disk usage, and recording state.

That mock still matters because it exposes the real product model. The system is not merely "paste YAML into a text box." It is "manipulate a studio setup, observe it, and run it." However, the browser realization of that model is now explicitly deferred.

### Key prototype limitations

- Preview lifecycle is request-driven rather than source-driven. Each browser preview request starts its own FFmpeg process ([jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L422)).
- The normalized config is already halfway to a compiler pipeline, but the compiler result is not named or reused as a first-class plan ([jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L522)).
- Recording logic directly loops over effective sources and spawns processes inline, which makes policy changes hard to test ([jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L769)).
- Platform details are hardcoded in argument builders. X11, V4L2, and PulseAudio are coupled directly to the recording and preview code ([jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L1009), [jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L972)).
- The current frontend has no persistent model for discovery, draft editing, or runtime events. It only polls and repaints ([jank-prototype/web/app.js](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/app.js#L156)).

## Design Goals

1. Keep the media runtime simple enough that the team can ship version 1 quickly.
2. Keep the domain model explicit enough that the system does not collapse into handler glue.
3. Let the frontend edit user intent, not raw process flags.
4. Make platform-specific behavior pluggable behind discovery and capture interfaces.
5. Use Glazed for CLI flags and command structure, not ad hoc environment variables.
6. Make previews cheap enough for a source grid but structured enough to replace later if needed.
7. Persist setups and templates cleanly so the user can reopen a previous studio state.

## Proposed Solution

### High-level architecture

The application should be split into four current layers plus one deferred layer:

1. CLI and process bootstrap.
2. Domain model and DSL compiler.
3. Discovery services.
4. Recording runtime services.
5. Deferred web control surface.

```text
CLI (discover / compile / record)
  |-- Glazed flags and output rendering
  v
Go application
  |-- DSL parser + normalizer
  |-- compiler
  |-- discovery catalog
  |-- recording manager
  v
Platform adapters
  |-- X11 display/window discovery and capture selectors
  |-- V4L2 camera inventory
  |-- PulseAudio device inventory
  |-- FFmpeg process runners

Deferred follow-up ticket:
  Browser UI -> HTTP/WebSocket API -> same compiler/runtime core
```

### The core architectural decision

The core architectural decision is to make the intermediate DSL compiler the source of truth. The UI should never assemble FFmpeg arguments directly. Instead:

```text
Machine inventory + user draft
  -> normalize into StudioSpec
  -> compile into CompiledPlan
  -> start Session from CompiledPlan
  -> emit runtime events back to frontend
```

This is the right boundary because the same `CompiledPlan` can be:

- validated before recording;
- rendered as a review summary in the UI;
- executed by the runtime;
- serialized for debugging;
- tested without a browser.

## DSL And Domain Model

### Why two data structures are required

One DSL is not enough. The user-facing saved setup and the runtime execution graph solve different problems.

- `StudioSpec` should capture user intent and saved configuration.
- `CompiledPlan` should capture concrete runtime actions after selectors are resolved and defaults are applied.

This distinction prevents the saved DSL from accumulating transient runtime junk such as PID values, output file paths, or live preview URLs.

### Proposed Go types

```go
type StudioSpec struct {
    Schema    string
    Metadata  SetupMetadata
    Defaults  DefaultsBlock
    Templates map[string]DestinationTemplate
    Video     []VideoSourceSpec
    Audio     []AudioSourceSpec
    Preview   PreviewPolicy
    Record    RecordPolicy
}

type VideoSourceSpec struct {
    ID       string
    Label    string
    Kind     VideoSourceKind
    Enabled  bool
    Armed    bool
    Solo     bool
    Selector VideoSelector
    Capture  VideoCaptureSpec
    Output   VideoOutputSpec
    Preview  PreviewSpec
}

type AudioSourceSpec struct {
    ID       string
    Label    string
    Enabled  bool
    Armed    bool
    DeviceID string
    Capture  AudioCaptureSpec
    Output   AudioOutputSpec
    Meter    MeterSpec
}

type CompiledPlan struct {
    SessionID       string
    VideoJobs       []VideoRecordJob
    AudioJobs       []AudioRecordJob
    PreviewJobs     []PreviewJob
    OutputManifest  []PlannedOutput
    ResolvedSources ResolvedInventory
    Warnings        []string
}
```

### Selectors versus resolved sources

Selectors should be human-authored or UI-authored. Resolved sources should be runtime artifacts.

Examples:

- A display selector can store `display_id`.
- A window selector can store `window_id`, `title_contains`, or a stable adapter-specific reference.
- A region selector can store either a literal rectangle or a preset token like `top_half`.
- A camera selector can store `device_id`.
- An audio selector can store `device_id`.

At compile time, selectors resolve against a fresh inventory snapshot:

```go
func Compile(ctx context.Context, spec StudioSpec, inv InventorySnapshot) (*CompiledPlan, error) {
    normalized := NormalizeSpec(spec)
    resolved, warnings, err := ResolveSelectors(normalized, inv)
    if err != nil {
        return nil, errors.Wrap(err, "resolve selectors")
    }
    plan := BuildJobs(normalized, resolved)
    plan.Warnings = append(plan.Warnings, warnings...)
    return plan, nil
}
```

### YAML shape

The current YAML examples in [jank-prototype/dsl.md](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/dsl.md#L1) are directionally correct, but I recommend renaming the top-level groups so the saved setup reads like a studio description rather than an FFmpeg recipe.

```yaml
schema: screencast.studio/v1alpha1

metadata:
  name: "Product Demo"
  recording_root: "~/Recordings/ScreencastStudio"

defaults:
  preview:
    fps: 5
    max_width: 640
  video:
    fps: 24
    cursor: true
    container: mov
    codec: h264
    quality: 75
  audio:
    codec: pcm_s16le
    sample_rate_hz: 48000
    channels: 2

templates:
  per-source: "{date}/{source_name}.{ext}"
  mixed-audio: "{date}/audio/mixed.{ext}"

video:
  - id: display-1
    label: "Display 1"
    kind: display
    armed: true
    selector:
      display_id: "1"
    output:
      template: per-source

  - id: top-half
    label: "Top Half"
    kind: region
    armed: true
    selector:
      display_id: "1"
      preset: top_half
    output:
      template: per-source

audio:
  - id: built-in-mic
    label: "Built-in Mic"
    armed: true
    device_id: "built_in_mic"
    capture:
      gain: 1.0
    output:
      template: mixed-audio
```

### Why this is better than the prototype shape

- It makes preview policy explicit instead of implicit.
- It separates metadata, defaults, templates, and sources more cleanly.
- It lets the frontend round-trip the setup without understanding platform-specific command flags.
- It keeps room for UI-only fields like `armed` and `solo`, which the imported JSX clearly expects ([sources/local/screencast-studio-v2.jsx.jsx](/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/sources/local/screencast-studio-v2.jsx.jsx#L175)).

## Proposed Package Layout

```text
cmd/
  screencast-studio/
    main.go

pkg/
  cli/
    root.go
    discover.go
    compile.go
    record.go
    setup_validate.go
    setup_export.go
    discovery_list.go

  app/
    application.go

  dsl/
    spec.go
    normalize.go
    validate.go
    compile.go
    templates.go

  discovery/
    catalog.go
    inventory.go
    types.go
    x11/
      displays.go
      windows.go
    v4l2/
      cameras.go
    pulse/
      inputs.go

  runtime/
    session.go
    session_store.go
    recording_manager.go
    process_supervisor.go
    ffmpeg/
      record_args.go
```

This layout follows the real system boundaries. It avoids a package named `default`, uses the top-level module only, and gives each future engineer a stable place to add code.

## CLI Design With Glazed

The Go backend should expose one root binary and a small set of Glazed subcommands. All operational flags should be Glazed fields, not environment variables. Logging should use zerolog with a `--log-level` flag as required by the local project conventions.

### Recommended commands

```text
screencast-studio setup validate
screencast-studio setup print-example
screencast-studio setup compile
screencast-studio discovery list
screencast-studio record
```

### `record` command

Purpose: load a DSL file, compile it against current discovery results, print the resolved plan if requested, and execute the capture jobs.

Recommended flags:

- `--file`
- `--recordings-root`
- `--dry-run`
- `--print-plan`
- `--continue-on-nonfatal-warning`
- `--log-level`

### `setup validate` command

Purpose: validate and normalize a saved setup file without starting the server.

Recommended flags:

- `--file`
- `--format yaml|json`
- `--strict`
- `--log-level`

### `discovery list` command

Purpose: print displays, windows, cameras, or microphones in a structured form for debugging and smoke tests.

Recommended flags:

- `--kind display|window|camera|audio`
- `--output json|yaml|table`
- `--log-level`

## CLI Contracts

The CLI is the current control plane. It should expose machine-readable output for inspection and keep human-facing behavior predictable.

### `discovery list`

`discovery list` should support:

- `--kind display|window|camera|audio|all`
- `--output json|yaml|table`
- `--show-raw`

Example JSON shape:

```json
{
  "displays": [{"id":"1","label":"Display 1","bounds":{"x":0,"y":0,"w":1920,"h":1080}}],
  "windows": [{"id":"0x3a00007","title":"Firefox","bounds":{"x":1920,"y":24,"w":1280,"h":1016}}],
  "cameras": [{"id":"/dev/video0","label":"Built-in Camera"}],
  "audio_inputs": [{"id":"alsa_input.usb-1","label":"Built-in Mic"}]
}
```

### `setup compile`

`setup compile` should:

- load a DSL file;
- normalize it;
- resolve selectors against current inventory;
- print the normalized spec and compiled plan summary.

Recommended response shape for structured output:

```json
{
  "spec": { "...": "normalized StudioSpec" },
  "plan": {
    "session_id": "session-20260409-131500",
    "video_jobs": 4,
    "audio_jobs": 1,
    "outputs": [
      {"source_id":"display-1","kind":"video","path":"/Recordings/.../Display-1.mov"}
    ],
    "warnings": []
  }
}
```

### `record`

`record` should:

- call the same compile pipeline;
- refuse to run on fatal validation errors;
- print the plan before execution when requested;
- start capture workers;
- stream progress and log lines to stderr or structured output;
- exit non-zero if any required job fails.

## Discovery Architecture

The research note in [jank-prototype/research.md](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/research.md#L26) gives a clear version 1 window-discovery strategy on X11:

- read `_NET_CLIENT_LIST` or `_NET_CLIENT_LIST_STACKING`;
- resolve titles from `_NET_WM_NAME`;
- compute root coordinates with `XTranslateCoordinates`;
- optionally expand decorations using `_NET_FRAME_EXTENTS`.

That leads directly to the following interface:

```go
type InventoryProvider interface {
    Snapshot(ctx context.Context) (*InventorySnapshot, error)
}

type InventorySnapshot struct {
    TakenAt    time.Time
    Displays   []DisplayDescriptor
    Windows    []WindowDescriptor
    Cameras    []CameraDescriptor
    AudioInputs []AudioInputDescriptor
}
```

### Version 1 platform stance

Version 1 should explicitly target Linux with X11, V4L2, and PulseAudio because that is what the prototype and research already support ([jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L1011), [jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L1035), [jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L977)). Do not promise Wayland or macOS support in the version 1 architecture doc unless the team explicitly commits to it later.

### Region presets

The DSL examples already hint at presets such as `top_half` ([jank-prototype/dsl.md](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/dsl.md#L165)). Region handling should therefore distinguish between:

- literal rectangles;
- display-relative presets;
- future interactive drag selections.

Suggested compiler helper:

```go
func ResolveRegion(selector RegionSelector, displays []DisplayDescriptor) (Rect, error)
```

## Runtime Architecture

### Session model

A running session should be a first-class object with immutable plan input and mutable runtime state.

```go
type Session struct {
    ID          string
    Plan        *CompiledPlan
    StartedAt   time.Time
    State       SessionState
    Outputs     []OutputFile
    Previews    map[string]*PreviewWorker
    Recorders   map[string]*RecorderWorker
    EventBus    *EventBus
}
```

### Recording manager

The recording manager should take a `CompiledPlan` and start recorder workers. Each recorder worker wraps exactly one FFmpeg subprocess and exposes:

- start;
- stop;
- wait;
- last error;
- output path;
- health events.

This is an evolution of the current managed-process concept in [jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L159), but moved into a dedicated runtime package with explicit interfaces and event emission.

### Errgroup and shutdown

All long-running runtime orchestration should use `errgroup`. That gives clean cancellation semantics and follows the local project guidance.

```go
group, ctx := errgroup.WithContext(parentCtx)

for _, job := range plan.VideoJobs {
    job := job
    group.Go(func() error {
        return recordingManager.RunVideoJob(ctx, job)
    })
}

if err := group.Wait(); err != nil {
    eventBus.Publish(SessionFailed{Err: err.Error()})
}
```

## Deferred Web Follow-Up

The imported JSX mock remains the reference for the eventual control surface, but that work should move to a second ticket after the CLI milestone is complete. The web ticket should reuse the compiler and runtime packages created here rather than introducing a separate execution path.

## Media Transport Decisions

### Recording

Use FFmpeg subprocesses for version 1 recording because:

- the prototype already demonstrates the basic builders for display, region, window, camera, and mixed audio capture ([jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L953), [jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L972));
- FFmpeg is easier to supervise from Go than building a native media graph from scratch;
- the output-template model maps naturally to one-worker-per-output.

### Audio metering

Do not try to solve browser audio preview or live metering in this ticket. Focus on correct capture and file output first.

## Path Templates And Output Policy

The current prototype already has a usable template renderer with placeholders such as `{session_id}`, `{source_name}`, `{source_type}`, `{date}`, `{time}`, and `{timestamp}` ([jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go#L1080)). That should become a dedicated package with unit tests and better validation.

Rules for the new system:

- Validate all template keys at setup-compile time.
- Sanitize path segments before rendering output paths.
- Resolve `~` only inside the backend, not in the browser.
- Show the predicted output paths in the UI before recording starts.

Planned API:

```go
type PlannedOutput struct {
    SourceID string
    Kind     string
    Template string
    Path     string
}
```

## Detailed Runtime Flows

### Boot flow

```text
record command starts
  -> parse flags
  -> load DSL file
  -> run discovery snapshot
  -> compile setup against latest inventory
  -> print plan summary
  -> execute recorder workers
```

### Compile flow

```text
setup compile command runs
  -> backend normalizes spec
  -> compiler resolves selectors against inventory
  -> command prints normalized spec + plan summary + warnings
```

### Record start flow

```text
user runs record
  -> command compiles latest setup
  -> recording manager creates recorder workers
  -> command waits on errgroup
  -> command exits with success or failure
```

### Record stop flow

```text
user sends interrupt or recording duration elapses
  -> record command cancels errgroup context
  -> recorder workers send graceful stop to ffmpeg
  -> outputs are finalized
  -> command prints final output manifest
```

## Implementation Plan

### Phase 1: Foundations and root CLI

Create the root binary, top-level Go module, Glazed command tree, zerolog setup, and a minimal app container.

Files to create first:

- `go.mod`
- `cmd/screencast-studio/main.go`
- `pkg/cli/root.go`
- `pkg/cli/discover.go`
- `pkg/cli/compile.go`
- `pkg/cli/record.go`
- `pkg/app/application.go`

Definition of done:

- `go run ./cmd/screencast-studio --help` works.
- `go run ./cmd/screencast-studio discovery list --help` works.
- `go run ./cmd/screencast-studio setup compile --help` works.
- `go run ./cmd/screencast-studio record --help` works.
- logging is structured and respects `--log-level`.

### Phase 2: DSL, validation, and compile pipeline

Extract the current config structs and normalization logic from the prototype into dedicated packages. Rename types around `StudioSpec` and `CompiledPlan`.

Files:

- `pkg/dsl/spec.go`
- `pkg/dsl/normalize.go`
- `pkg/dsl/compile.go`
- `pkg/dsl/templates.go`
- `pkg/dsl/validate_test.go`

Definition of done:

- saved setup files can be normalized and compiled without the server running;
- template rendering and selector validation have unit tests;
- unsupported fields produce warnings or errors intentionally.

### Phase 3: Discovery services

Implement inventory providers for X11 displays and windows, V4L2 cameras, and PulseAudio inputs.

Files:

- `pkg/discovery/types.go`
- `pkg/discovery/x11/windows.go`
- `pkg/discovery/v4l2/cameras.go`
- `pkg/discovery/pulse/inputs.go`

Definition of done:

- `discovery list --kind window --output json` prints usable selectors;
- region presets can be resolved relative to discovered displays;
- inventory snapshots can be cached and refreshed safely.

### Phase 4: Recording runtime

Create recording managers plus FFmpeg argument builders. Reuse the prototype builders as reference, but move them behind typed jobs.

Files:

- `pkg/runtime/session.go`
- `pkg/runtime/recording_manager.go`
- `pkg/runtime/ffmpeg/record_args.go`

Definition of done:

- a compiled plan can start and stop cleanly;
- recorder workers create expected files;
- `record` runs the same compile pipeline used by `setup compile`.

### Phase 5: Packaging and smoke tests

Add smoke tests and document operational commands.

Files:

- `Makefile` targets
- CI workflow files
- `playbooks/` docs if needed later

Definition of done:

- `go build ./...` succeeds;
- smoke test covers discovery, setup compile, record, and stop.

## Current Delivery Sequence

This ticket should now be executed in the following order:

1. Establish the root Go module and Glazed CLI.
2. Extract the DSL and compile pipeline.
3. Implement discovery commands.
4. Implement `record` on top of the compiled plan.
5. Add smoke tests and commit the working CLI milestone.
6. Open a second ticket for the web control surface that reuses these packages.

## Testing Strategy

### Unit tests

- DSL normalization.
- selector resolution.
- output-template rendering.
- FFmpeg argument generation.
- event serialization.

### Integration tests

- fake or stubbed FFmpeg worker invocation;
- record start and stop lifecycle;
- command contract tests for `setup compile` and `record`.

### Manual smoke test

1. Run `discovery list` and confirm the expected sources are present.
2. Run `setup compile --file ...` and review the compiled plan.
3. Run `record --file ...`.
4. Stop recording and verify output files landed at the predicted template paths.

## Risks And Tradeoffs

### Risk: window selectors can go stale

Window titles and XIDs can change. The mitigation is to store selectors separately from resolved runtime targets and re-resolve before recording.

### Risk: audio sync and mixing policy can drift

Multiple audio devices can introduce timing issues. Version 1 should keep audio policy simple and explicit. Do not quietly invent complex sync logic without product requirements.

### Tradeoff: FFmpeg subprocesses versus native media graph

FFmpeg subprocesses are not the fanciest architecture, but they are the right version 1 architecture. They let the Go service focus on orchestration and state, which is where the system complexity actually lives.

## Alternatives Considered

### Alternative 1: keep the prototype and refactor in place

Rejected. The prototype file already combines too many concerns. Pulling clean seams out of it while preserving behavior would cost nearly as much as building the correct package structure directly.

### Alternative 2: build a fully in-process Go media pipeline

Rejected for version 1. That is too much low-level media complexity for the current scope.

### Alternative 3: build the web frontend before the CLI works

Rejected. That would force the browser to become the main debugging surface before the primitives are stable. The better sequence is to prove `discover`, `compile`, and `record` from the terminal first.

## Open Questions

1. Is version 1 explicitly Linux and X11 only, or should the public README phrase it as experimental platform support?
2. Does product scope require a merged composite output file, or are separate per-source files plus mixed audio sufficient?
3. Should region presets remain preset-only in the CLI milestone?
4. What should the second web ticket cover first: previews, setup editing, or runtime control?

## Recommended First Week For A New Intern

1. Read the prototype files listed in the References section.
2. Build the DSL and compiler package first.
3. Add `setup validate` and `setup compile` commands before touching the frontend.
4. Implement discovery adapters next so the compiler can resolve selectors.
5. Implement `record` only after the compiler and discovery pieces are stable.
6. Only after the CLI milestone is stable should the intern open the web-control follow-up ticket.

This ordering matters. If the intern starts with the browser first, they will invent unstable state shapes and push OS-specific behavior into the wrong layer.

## References

- Prototype overview: [jank-prototype/README.md](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/README.md)
- Prototype backend: [jank-prototype/main.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go)
- Prototype frontend: [jank-prototype/web/index.html](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/index.html)
- Prototype frontend logic: [jank-prototype/web/app.js](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/app.js)
- DSL examples: [jank-prototype/dsl.md](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/dsl.md)
- X11 window research: [jank-prototype/research.md](/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/research.md)
- Imported UI reference: [sources/local/screencast-studio-v2.jsx.jsx](/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/sources/local/screencast-studio-v2.jsx.jsx)

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
