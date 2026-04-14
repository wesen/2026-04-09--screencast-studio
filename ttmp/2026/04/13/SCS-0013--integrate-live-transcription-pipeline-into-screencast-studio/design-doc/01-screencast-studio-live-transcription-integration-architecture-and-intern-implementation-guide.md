---
Title: Screencast Studio Live Transcription Integration Architecture and Intern Implementation Guide
Ticket: SCS-0013
Status: active
Topics:
    - screencast-studio
    - transcription
    - gstreamer
    - audio
    - websocket
    - go
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../2026-04-13--transcription-go/internal/live/runner.go
      Note: |-
        Prototype live orchestration showing chunk and websocket transports, accumulators, and sinks
        Prototype transport split
    - Path: ../../../../../../../2026-04-13--transcription-go/server/live_sessions.py
      Note: |-
        Prototype session registry, sequence checks, flush/stop semantics, and idle cleanup rules
        Prototype session registry and sequence/flush/stop rules to preserve conceptually
    - Path: ../../../../../../../2026-04-13--transcription-go/server/server.py
      Note: |-
        Prototype ASR service exposing batch, chunk, and websocket streaming endpoints
        Prototype ASR backend contract including batch
    - Path: internal/web/handlers_ws.go
      Note: |-
        Existing browser websocket bootstrap and event stream
        Current browser websocket bootstrap and live event transport that should carry transcript updates
    - Path: internal/web/session_manager.go
      Note: Current recording event handling and websocket publication path that should carry transcription updates
    - Path: pkg/media/gst/recording.go
      Note: |-
        Current audio recording pipeline, mixer, compressor, level element, and event emission seam where transcription audio should branch
        Current GStreamer audio graph
    - Path: proto/screencast/studio/v1/web.proto
      Note: |-
        Current websocket event schema that needs transcription event additions
        Current websocket schema that needs transcription event additions
    - Path: ui/src/features/session/wsClient.ts
      Note: |-
        Existing frontend websocket client and reducer dispatch path
        Current frontend websocket decode/dispatch path for live state
ExternalSources: []
Summary: Detailed intern-facing architecture and implementation guide for integrating the transcription-go live transcription pipeline into screencast-studio. Recommends a protocol-level integration via a new transcription seam, a GStreamer appsink transcription branch, server-mediated websocket updates to the browser, and phased rollout from backend seam to live UI.
LastUpdated: 2026-04-13T20:45:00-04:00
WhatFor: Guide implementation of live transcription in screencast-studio using lessons and protocols from the separate transcription-go prototype, while keeping the current recording architecture stable and understandable for a new engineer.
WhenToUse: Read this before designing or implementing transcription in screencast-studio, especially if you need to understand the current GStreamer recording graph, the existing browser websocket path, the prototype transcription service, and the recommended staged integration plan.
---


# Screencast Studio Live Transcription Integration Architecture and Intern Implementation Guide

This document explains how to integrate the separate `transcription-go` prototype into `screencast-studio` without turning the codebase into an unreviewable pile of one-off adapters. It is written for a new intern. That means it starts by orienting you to the existing system, not by assuming you already know why the GStreamer migration mattered, how the web event path works, or what parts of the transcription prototype are reusable versus prototype-only.

The short version is this: `screencast-studio` now has the right media foundation for live transcription because its audio graph is in-process and programmable. The `transcription-go` repository now has the right ASR service ideas because it already models session-oriented streaming, partial versus final transcript updates, and output sinks. The correct integration is therefore **not** to shell out to the `transcribe` CLI from inside `screencast-studio`, and **not** to import the prototype's `internal/...` Go packages directly. The correct integration is to add a **new transcription seam** to `screencast-studio`, feed it normalized PCM from the GStreamer recording graph, and speak the proven service protocol from `transcription-go` through a dedicated backend adapter.

If you remember only one sentence from this document, remember this one:

> **`screencast-studio` should own capture, session lifecycle, browser updates, and UX; the `transcription-go` service should own ASR decoding.**

---

## 1. Executive Summary

The current `screencast-studio` runtime is finally in a position where live transcription makes architectural sense. After the GStreamer migration, the recording audio path is a real in-process graph: source branches feed an `audiomixer`, then post-mix processing runs through compressor, level metering, and encoder/file output (`pkg/media/gst/recording.go:576-709`). The web layer already has an event hub and browser websocket transport (`internal/web/event_hub.go:26-48`, `internal/web/handlers_ws.go:16-74`), and the frontend already consumes live server events for session state and audio meter updates (`ui/src/features/session/wsClient.ts:13-73`, `ui/src/features/session/sessionSlice.ts:10-109`).

Separately, the `transcription-go` prototype already proved the ASR-side ideas we need. It has:

- a Dagger-managed long-running ASR server (`internal/server/dagger.go:24-118`),
- a batch API and a chunk API (`internal/asr/client.go:84-120`, `server/server.py:74-185`),
- a real websocket streaming API (`server/server.py:187-314`, `internal/live/wsclient.go:17-118`),
- explicit partial-vs-final transcript state (`internal/live/types.go:5-23`, `internal/live/accumulator.go:14-152`),
- and live sinks for subtitle/SQLite-style outputs (`internal/live/sinks.go:8-35`, `internal/live/subtitle_sink.go:1-38`, `internal/live/sqlite_sink.go:1-28`).

### Recommendation

The recommended design is:

1. **Add a transcription seam inside `screencast-studio`** rather than hard-coding one backend.
2. **Branch normalized mono PCM out of the GStreamer audio graph** using a tee + appsink.
3. **Send that PCM to a session-oriented transcription backend** using the websocket protocol already proven in `transcription-go`.
4. **Convert backend transcript updates into screencast-studio websocket events** over the app's existing `/ws` channel.
5. **Keep browser traffic server-mediated**. The browser should keep talking only to `screencast-studio`, not directly to the ASR backend.
6. **Start with a backend adapter that speaks the `transcription-go` service protocol**. Do not try to import its `internal/...` packages directly across module boundaries.

### Most important design decision

The fastest-looking path is not the best path. The old Phase 5 note in SCS-0012 suggested a local whisper CLI process per chunk. After studying `transcription-go`, that should be treated as a fallback or throwaway spike, not the target architecture. The better long-term design is a **session-oriented transcription backend with partial and final results**, because that matches the app's live nature and avoids making transcript state management a giant pile of chunk-boundary hacks.

---

## 2. Problem Statement and Scope

### 2.1 The actual problem

The problem is not merely “transcribe some audio.” The problem is:

- capture audio during a live recording,
- preserve the current recording behavior,
- produce useful transcript updates during the recording,
- publish those updates to the web UI,
- and end with a coherent final transcript artifact.

That means the feature spans at least four layers:

1. the GStreamer recording graph,
2. the Go runtime and session lifecycle,
3. the server websocket/event model,
4. the React/Redux UI state.

Any plan that only talks about ASR inference and ignores the other three layers is incomplete.

### 2.2 In scope

This ticket covers:

- how `screencast-studio` currently records audio,
- how `transcription-go` currently models batch and live transcription,
- what integration seam should exist in `screencast-studio`,
- how transcript events should move to the browser,
- a phased implementation plan with file references,
- and a validation strategy for the live path.

### 2.3 Out of scope

This ticket does **not** try to solve:

- diarization / speaker separation,
- cloud multi-tenant deployment,
- subtitle editing UI,
- multilingual model selection,
- or a full transcript search/export product surface.

Those may come later. This ticket is specifically about integrating the current transcription prototype into the current recording application in a safe and understandable way.

---

## 3. What a New Intern Needs to Understand First

Before you touch implementation, you need a correct mental model of both systems.

### 3.1 `screencast-studio` today: the important layers

At a high level, the app looks like this:

```text
Browser UI (React/Redux)
        |
        | HTTP + /ws
        v
internal/web/*
  - handlers_api.go
  - handlers_ws.go
  - session_manager.go
  - preview_manager.go
        |
        v
pkg/app/application.go
        |
        v
pkg/media/gst/*
  - preview runtime
  - recording runtime
  - shared capture graph
        |
        v
GStreamer pipelines + devices
```

The important thing to notice is that `pkg/app/application.go` already provides a runtime seam for recording (`pkg/app/application.go:23-44`, `196-245`). That seam was used to move from FFmpeg to GStreamer. Transcription should follow the same philosophy: put backend-specific behavior behind a seam instead of hard-coding it into web handlers.

### 3.2 The current audio recording path in `screencast-studio`

The most important file for transcription integration is `pkg/media/gst/recording.go`.

The current audio pipeline is built in two layers:

1. `buildAudioRecordingPipeline(job)` creates the top-level mixer graph (`pkg/media/gst/recording.go:576-624`).
2. `buildAudioOutputChain(out, outputPath)` creates the post-mixer processing chain (`pkg/media/gst/recording.go:660-709`).

Today the shape is effectively:

```text
pulsesrc(source1) -> caps -> audioconvert -> audioresample -> volume ---\
pulsesrc(source2) -> caps -> audioconvert -> audioresample -> volume ----> audiomixer
...                                                                   ---/

audiomixer
  -> audioconvert
  -> audioresample
  -> encoder caps
  -> audiodynamic (compressor)
  -> level
  -> wavenc/opusenc
  -> filesink
```

Important details:

- per-source gain already exists via the `volume` element in each source branch (`pkg/media/gst/recording.go:626-658`),
- global compressor control already exists via `audiodynamic` (`pkg/media/gst/recording.go:672-679`),
- the audio meter already exists via the `level` element (`pkg/media/gst/recording.go:680-685`),
- and `audioLevelEventFromMessage(...)` already turns GStreamer bus messages into web-consumable audio level events (`pkg/media/gst/recording.go:773-826`).

This is exactly why live transcription belongs in this graph now: the app already has an authoritative live mixed-audio point.

### 3.3 The current web event path in `screencast-studio`

A new intern should also understand that `screencast-studio` already has a working app-level websocket event path.

Key files:

- `internal/web/event_hub.go` — generic in-process pub/sub for server events
- `internal/web/handlers_ws.go` — upgrades `/ws`, sends bootstrap state, then forwards events
- `internal/web/pb_mapping.go` — maps internal `ServerEvent` to protobuf `ServerEvent`
- `proto/screencast/studio/v1/web.proto` — wire schema consumed by web UI
- `ui/src/features/session/wsClient.ts` — browser websocket client
- `ui/src/features/session/sessionSlice.ts` — Redux store updates

That flow looks like this:

```text
recording runtime event
      |
      v
RecordingManager / TelemetryManager
      |
      v
EventHub.Publish(ServerEvent)
      |
      v
handleWebsocket() writes protobuf JSON to browser
      |
      v
WsClient decodes ServerEvent and dispatches Redux actions
```

You can see this clearly in the code:

- websocket subscribe/publish path: `internal/web/event_hub.go:26-48`
- websocket handler bootstrap + event forwarding: `internal/web/handlers_ws.go:16-74`
- current event schema: `proto/screencast/studio/v1/web.proto:223-252`
- frontend event handling: `ui/src/features/session/wsClient.ts:47-73`

This matters because it means we do **not** need a second browser-facing transport for transcription. We already have one.

### 3.4 The `transcription-go` prototype: what it actually is

The `transcription-go` repository is not just a CLI. It actually contains three distinct ideas:

1. **A batch transcription tool**
2. **A long-running ASR backend service**
3. **A live transcription session model**

#### 3.4.1 Service lifecycle

The backend service is started through Dagger in `internal/server/dagger.go:24-118`.

Important facts:

- it starts a Python container,
- installs server requirements,
- loads the ASR model once,
- exposes the service over a host tunnel,
- and health-checks it before use.

This is operationally important: a warm model server is much better than spawning a new ASR process for every 3-second chunk.

#### 3.4.2 Service endpoints

The prototype service exposes:

- `GET /health`
- `POST /transcribe/full`
- `POST /transcribe/chunk`
- `WS /transcribe/stream`

See `server/server.py:1-185` for batch/chunk and `server/server.py:187-314` for websocket streaming.

#### 3.4.3 Session-oriented live model

The prototype's best live ideas are not in the CLI wrapper. They are in:

- `server/live_sessions.py`
- `server/live_decoder.py`
- `internal/live/types.go`
- `internal/live/accumulator.go`
- `internal/live/wsclient.go`
- `internal/live/stream_receiver.go`

Those files define the concepts that matter:

- `start` / `audio` / `flush` / `stop` websocket messages,
- session IDs,
- sequence validation,
- partial vs final transcript events,
- committed vs pending transcript state,
- and final artifact sinks.

### 3.5 The most important limitation: you cannot just import the prototype internals

This point is easy to miss and very important.

`transcription-go` is a separate Go module (`/home/manuel/code/wesen/2026-04-13--transcription-go/go.mod`) and almost all of its useful Go code lives under `internal/...`.

Examples:

- `internal/asr/client.go`
- `internal/live/runner.go`
- `internal/live/wsclient.go`
- `internal/output/sqlite.go`

Because these are `internal` packages in a different module, `screencast-studio` cannot cleanly import them directly under normal Go visibility rules.

That means the integration choices are:

1. copy/adapt code into `screencast-studio`,
2. extract shared public packages upstream into `transcription-go`,
3. or integrate at the **protocol level** instead of the Go-package level.

The recommended short-term answer is **protocol-level integration**.

---

## 4. Gap Analysis

Now that we have the two systems mapped, the gaps are straightforward.

### 4.1 What `screencast-studio` already has

`Screencast-studio` already has:

- a stable recording lifecycle,
- a live mixed audio graph,
- per-source gain control,
- global compressor control,
- browser websocket delivery,
- a frontend live state store,
- and recording-session scoping.

### 4.2 What `screencast-studio` does not yet have

It does **not** yet have:

- a transcription backend seam,
- a way to branch mixed audio into a Go callback,
- a transcript event type in `media.RecordingEvent` or the web protobuf schema,
- a server-side transcript snapshot store for reconnect/bootstrap,
- or a UI surface for partial/final transcript text.

### 4.3 What `transcription-go` already has

It already has:

- a reusable service protocol,
- a session-oriented websocket ASR transport,
- partial/final result semantics,
- and transcript accumulation/output logic.

### 4.4 What `transcription-go` does not solve for us automatically

It does **not** automatically solve:

- how the GStreamer graph feeds it,
- how `screencast-studio` session IDs map to ASR session IDs,
- how browser reconnect bootstrap works,
- how transcript state should be shown in the UI,
- or how recording/transcription failures interact.

That integration logic belongs in `screencast-studio`.

---

## 5. Recommended Architecture

## 5.1 Design principles

The recommended architecture follows five rules.

### Rule 1: `screencast-studio` should own capture and product UX

The app already knows when recording starts, stops, fails, and what session ID the browser cares about. The transcription backend does not.

### Rule 2: the ASR backend should be replaceable

Do not hard-code “Nemotron over Dagger” into the recording runtime. Create a seam such as `pkg/transcription` so the app can later swap backend modes.

### Rule 3: the audio graph should emit normalized PCM, not ad hoc temp files

The best match between the two systems is the `transcription-go` websocket streaming API because `screencast-studio` can produce live PCM directly from GStreamer.

### Rule 4: the browser should still only talk to `screencast-studio`

Do not make the browser open a second websocket to the ASR backend. That would split state ownership and create auth/lifecycle complexity.

### Rule 5: start with stable final words, then expose partials

The server-side model should support both partial and final events from day one, but the UI can initially emphasize committed/final transcript content to reduce confusion.

---

## 5.2 Proposed top-level integration shape

```text
                   ┌──────────────────────────────┐
                   │  Browser UI                  │
                   │  - existing /ws connection   │
                   │  - transcript panel          │
                   └──────────────┬───────────────┘
                                  │
                                  │ protobuf JSON ServerEvent
                                  │
                   ┌──────────────▼───────────────┐
                   │ internal/web                 │
                   │ - RecordingManager           │
                   │ - TranscriptionManager       │
                   │ - handlers_ws.go             │
                   └──────────────┬───────────────┘
                                  │
                                  │ recording + transcription events
                                  │
                   ┌──────────────▼───────────────┐
                   │ pkg/app/application.go       │
                   │ - recording runtime seam     │
                   │ - transcription seam         │
                   └──────────────┬───────────────┘
                                  │
                                  │ PCM frames + lifecycle
                                  │
                   ┌──────────────▼───────────────┐
                   │ pkg/media/gst/recording.go   │
                   │ - audiomixer                 │
                   │ - compressor                 │
                   │ - tee                        │
                   │ - filesink branch            │
                   │ - transcription appsink      │
                   └──────────────┬───────────────┘
                                  │
                                  │ websocket or chunk protocol
                                  │
                   ┌──────────────▼───────────────┐
                   │ transcription backend        │
                   │ (transcription-go service)   │
                   │ - /transcribe/stream         │
                   │ - partial/final words        │
                   └──────────────────────────────┘
```

---

## 5.3 The new seam: `pkg/transcription`

Create a new package in `screencast-studio`, likely `pkg/transcription`, with interfaces like these:

```go
type Backend interface {
    StartSession(ctx context.Context, cfg SessionConfig) (Session, error)
}

type Session interface {
    PushPCM(ctx context.Context, chunk PCMChunk) error
    Flush(ctx context.Context) error
    Stop(ctx context.Context) error
    Events() <-chan Update
}

type UpdateType string

const (
    UpdatePartial UpdateType = "partial"
    UpdateFinal   UpdateType = "final"
    UpdateStopped UpdateType = "stopped"
    UpdateError   UpdateType = "error"
)

type Update struct {
    Type         UpdateType
    SessionID    string
    Sequence     int
    UpToTime     float64
    Words        []Word
    Text         string
    ProcessingMS int
    Err          error
}
```

Why a seam is important:

- the first backend may speak the `transcription-go` websocket service,
- a fallback backend may later use HTTP chunk uploads,
- a dev backend may later be a fake transcript generator for tests,
- and a future local-whisper backend could exist without touching the GStreamer graph.

### Important recommendation

Do **not** make `pkg/media/gst/recording.go` know about Dagger, Python, or Nemotron. It should only know it has a callback/session that consumes normalized PCM.

---

## 5.4 Where the tee should go in the audio graph

This design detail matters.

The existing Phase 5 note from SCS-0012 said:

> “After `audiomixer`, add a tee: branch 1 goes to recording, branch 2 goes to capsfilter(16kHz/mono) -> appsink.”

That is directionally correct, but we can improve it.

### Recommended tee position

Put the tee **after source gains and after the global compressor**, but before final encoding and file sink.

Recommended shape:

```text
source branches
  -> volume per source
  -> audiomixer
  -> audioconvert
  -> audioresample
  -> encoder caps
  -> audiodynamic (compressor)
  -> tee
      ├─ branch A: level -> wavenc/opusenc -> filesink
      └─ branch B: audioconvert -> audioresample ->
                   capsfilter(audio/x-raw,format=S16LE,rate=16000,channels=1) -> appsink
```

### Why this tee position is better than teeing immediately after `audiomixer`

Because then transcription hears:

- per-source gain adjustments,
- the same mix the recording uses,
- and the same compressor behavior the user has enabled.

That means the transcript follows the actual recording audio more closely.

### Why `level` should stay on the recording branch

The `level` element currently exists to emit meter messages for the browser (`pkg/media/gst/recording.go:680-685`, `773-826`). It is not needed on the transcription branch. Keeping it only on the recording branch keeps the existing meter behavior stable.

---

## 5.5 Backend transport recommendation

### Recommended production target: websocket streaming

The `transcription-go` websocket API is the best match for `screencast-studio` because it already models:

- `start`
- `audio`
- `flush`
- `stop`
- `partial`
- `final_words`
- `stopped`
- `error`

See:

- client API: `internal/live/wsclient.go:17-118`
- receiver mapping: `internal/live/stream_receiver.go:8-57`
- server behavior: `server/server.py:187-314`
- session registry: `server/live_sessions.py:82-122`

### Transitional fallback: chunk HTTP uploads

The chunk transport in `internal/asr/client.go:102-120` and `server/server.py:133-185` is useful as a fallback or early validation path because it is easier to debug. But it is not the best long-term live design because it pushes more chunk-boundary and finalization logic into the client side.

### Practical recommendation

Implement the seam so that it supports both modes:

- `ws` = preferred production/live mode
- `chunk` = debug fallback

That mirrors the prototype's `TransportWS` and `TransportChunk` split (`internal/live/runner.go:19-20`, `72-116`).

---

## 5.6 Transcript state model inside `screencast-studio`

One of the best ideas in the prototype is the explicit distinction between pending and committed words.

From `transcription-go`:

- event types: `partial`, `final_words` (`internal/live/types.go:7-11`)
- accumulator state: `Committed`, `Pending`, `LastFinalTime` (`internal/live/types.go:18-23`)
- dedupe and monotonic finalization logic: `internal/live/accumulator.go:14-152`

You should copy this **concept**, not necessarily the exact package.

### Recommended server-side state in `screencast-studio`

Add a dedicated transcription state holder, either as:

- a new `TranscriptionManager`, or
- an extension of the existing recording session manager.

I recommend a **new `TranscriptionManager`** because transcript state is richer than the existing recording session state.

Example state:

```go
type TranscriptionSnapshot struct {
    SessionID      string
    Active         bool
    Backend        string
    Transport      string
    CommittedWords []Word
    PendingWords   []Word
    LastFinalTime  float64
    UpdatedAt      time.Time
    Error          string
}
```

### Why a manager is useful

It solves three product problems cleanly:

1. **websocket bootstrap on reconnect** — the server can send the current transcript snapshot immediately;
2. **UI refresh continuity** — the browser does not lose transcript state on reconnect;
3. **future persistence/export** — final transcript artifacts can be written from one coherent source of truth.

---

## 5.7 Browser event model

The browser already understands a protobuf `ServerEvent` union (`proto/screencast/studio/v1/web.proto:242-252`). The natural extension is to add transcription-specific messages there.

### Recommended protobuf additions

Something like:

```proto
message TranscriptWord {
  string word = 1;
  double start = 2;
  double end = 3;
}

message TranscriptionUpdateEvent {
  string session_id = 1;
  string update_type = 2; // partial | final | stopped | error
  repeated TranscriptWord committed_words = 3;
  repeated TranscriptWord pending_words = 4;
  double last_final_time = 5;
  string text = 6;
  int32 sequence = 7;
  int32 processing_ms = 8;
  string reason = 9;
}
```

and then in `ServerEvent`:

```proto
TranscriptionUpdateEvent transcription = 17;
```

### Why send both committed and pending

Because the UI can then choose how smart it wants to be:

- V1: show committed words only
- V1.1: show committed words plus gray pending tail
- V2: animate partial updates more richly

This avoids forcing UI state reconstruction from tiny patch messages.

---

## 5.8 Persistence and output artifacts

The prototype has useful output ideas:

- segment grouping: `internal/output/format.go:16-69`
- SRT/VTT/TXT writers: `internal/output/format.go:72-109`
- SQLite schema: `internal/output/sqlite.go:10-109`
- live rewrite sinks: `internal/live/subtitle_sink.go:13-38`, `internal/live/sqlite_sink.go:10-28`

### Recommendation for `screencast-studio` V1

Do **not** start by making transcript outputs part of the DSL compiler.

Instead:

- gate transcription behind a runtime option or HTTP setting,
- derive transcript artifact paths from the recording session output directory,
- write sidecar files when the recording stops.

For example:

```text
recordings/<session>/Display 1.mp4
recordings/<session>/audio-mix.wav
recordings/<session>/transcript.txt
recordings/<session>/transcript.srt
recordings/<session>/transcript.vtt
recordings/<session>/transcript.db
```

### Why not change the DSL immediately

Because Phase 5 already touches:

- media runtime,
- server event schema,
- frontend state,
- and potentially backend service lifecycle.

Adding DSL/schema/compiler changes at the same time increases review difficulty and rollout risk. Make transcript outputs deterministic sidecars first; make them first-class DSL outputs later if product needs it.

---

## 6. Concrete API and Runtime Sketches

## 6.1 Recording runtime integration sketch

The GStreamer recording runtime should own the transcription branch because it owns the audio graph.

Pseudocode:

```go
func buildAudioRecordingPipeline(job dsl.AudioMixJob, tx transcription.Session) (...) {
    mixer := gst.NewElement("audiomixer")

    // existing per-source branches
    for _, src := range job.Sources {
        branch := buildAudioSourceBranch(src, job.Output)
        link(branch -> mixer)
    }

    // existing post-mix shared processing
    postConvert := gst.NewElement("audioconvert")
    postResample := gst.NewElement("audioresample")
    postCaps := caps("audio/x-raw,format=S16LE,rate=48000,channels=2")
    compressor := gst.NewElement("audiodynamic")
    tee := gst.NewElement("tee")

    // recording branch
    recordQueue := gst.NewElement("queue")
    level := gst.NewElement("level")
    encoder := chooseEncoder(job.Output)
    sink := filesink(job.OutputPath)

    // transcription branch
    txQueue := gst.NewElement("queue")
    txConvert := gst.NewElement("audioconvert")
    txResample := gst.NewElement("audioresample")
    txCaps := caps("audio/x-raw,format=S16LE,rate=16000,channels=1")
    txSink := appsink()
    txSink.OnSample(func(buf []byte, pts, dur time.Duration) {
        tx.PushPCM(ctx, PCMChunk{Bytes: buf, PTS: pts, Duration: dur})
    })

    // link and request tee pads
    link(mixer -> postConvert -> postResample -> postCaps -> compressor -> tee)
    tee(recordQueue -> level -> encoder -> sink)
    tee(txQueue -> txConvert -> txResample -> txCaps -> txSink)
}
```

## 6.2 Transcription backend adapter sketch

```go
type WSBackend struct {
    endpoint string
}

func (b *WSBackend) StartSession(ctx context.Context, cfg SessionConfig) (Session, error) {
    c := newWSClient(b.endpoint)
    if err := c.Connect(ctx); err != nil { return nil, err }
    if err := c.Start(cfg.SessionID); err != nil { return nil, err }

    s := &wsSession{client: c, updates: make(chan Update, 64)}
    go s.readLoop()
    return s, nil
}

func (s *wsSession) PushPCM(ctx context.Context, chunk PCMChunk) error {
    return s.client.SendAudio(chunk)
}

func (s *wsSession) Flush(ctx context.Context) error {
    return s.client.Flush()
}

func (s *wsSession) Stop(ctx context.Context) error {
    return s.client.Stop()
}
```

## 6.3 Server-side publication sketch

```go
func (m *TranscriptionManager) Apply(update transcription.Update) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.state = accumulate(m.state, update)
    snapshot := cloneSnapshot(m.state)

    m.publish(ServerEvent{
        Type:      "transcription.update",
        Timestamp: time.Now(),
        Payload:   mapTranscriptionUpdate(snapshot, update),
    })
}
```

## 6.4 Frontend reducer sketch

```ts
switch (event.kind.case) {
  case 'transcription':
    dispatch(applyTranscriptionUpdate(event.kind.value));
    break;
}
```

and state like:

```ts
interface TranscriptState {
  committedWords: TranscriptWord[];
  pendingWords: TranscriptWord[];
  lastFinalTime: number;
  active: boolean;
  reason?: string;
}
```

---

## 7. Phased Implementation Plan

This section is intentionally concrete. If an intern followed these phases in order, the work would be reviewable and debuggable.

## Phase 0: Add the seam and event vocabulary

### Goal

Create stable application-level places for transcription concepts before touching the GStreamer graph.

### Files to add or change

Add:

- `pkg/transcription/types.go`
- `pkg/transcription/backend.go`
- `pkg/transcription/fake.go` (for tests)

Change:

- `proto/screencast/studio/v1/web.proto`
- `internal/web/pb_mapping.go`
- `ui/src/features/session/wsClient.ts`
- `ui/src/features/session/sessionSlice.ts`

### Deliverables

- transcription protobuf messages exist
- frontend can parse a transcription event even if nothing emits one yet
- fake backend exists for integration tests

## Phase 1: Add server-side transcript state management

### Goal

Store current transcript state and send it to websocket clients.

### Files to add or change

Add:

- `internal/web/transcription_manager.go`

Change:

- `internal/web/server.go` — add manager to `Server`
- `internal/web/handlers_ws.go` — send current transcript bootstrap state
- `internal/web/application.go` — if new app-facing methods are needed

### Important behavior

On websocket connect, the client should receive:

1. current recording state
2. preview list
3. audio meter
4. disk telemetry
5. current transcript snapshot

That keeps reconnect behavior sane.

## Phase 2: Add the GStreamer transcription branch

### Goal

Get live normalized PCM out of the recording graph without breaking existing recording output.

### Files to change

- `pkg/media/gst/recording.go`
- potentially a new helper file such as `pkg/media/gst/transcription_branch.go`

### Validation target

A focused smoke test should prove:

- recording still finalizes correctly,
- audio meter still works,
- appsink receives 16kHz mono S16LE audio,
- and transcription-disabled sessions behave exactly as before.

### Recommended script

Create a new ticket-local script in SCS-0013 such as:

- `scripts/01-go-gst-transcription-branch-smoke/`

## Phase 3: Add the backend adapter

### Goal

Connect the appsink PCM branch to the external transcription service.

### Files to add

- `pkg/transcription/ws_backend.go`
- `pkg/transcription/chunk_backend.go`
- `pkg/transcription/session.go`

### Important design choice

Start with the websocket backend as the target, but keep a chunk backend for debugging. This mirrors the prototype's successful split between `ws` and `chunk` transports (`internal/live/runner.go:19-20`, `113-116`, `122-185`, `185-335`).

## Phase 4: Feed transcription updates into the web event stream

### Goal

Turn backend transcript updates into browser-consumable app websocket events.

### Files to change

- `pkg/media/types.go`
- `pkg/recording/events.go`
- `pkg/app/application.go`
- `internal/web/session_manager.go`
- `internal/web/pb_mapping.go`

### Implementation note

You have two reasonable choices:

1. extend `media.RecordingEvent` / `recording.RunEvent` with transcription event types,
2. or introduce a parallel server-side callback path from the transcription session.

Recommendation: use the existing recording event sink unless it becomes awkward. It keeps lifecycle correlation simple.

## Phase 5: Add UI presentation

### Goal

Show transcript text live in the studio UI.

### Files to change

- `ui/src/features/session/sessionSlice.ts`
- `ui/src/features/session/wsClient.ts`
- `ui/src/pages/StudioPage.tsx`
- likely new UI components under `ui/src/components/studio/`

### V1 recommendation

Show:

- committed transcript text as the main body,
- optional pending tail in muted styling,
- backend status / active flag / last update time.

Do **not** over-design the editor UX in the first pass.

## Phase 6: Final artifact writing

### Goal

Persist transcript outputs when recording finishes.

### Files to add or change

Add:

- `pkg/transcription/output.go`
- or `pkg/transcription/output/*.go`

Potentially copy/adapt concepts from:

- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/output/format.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/output/sqlite.go`

### Recommendation

Initially write sidecar files under the recording output directory. Delay DSL/compiler changes unless a concrete product requirement emerges.

---

## 8. Testing and Validation Strategy

A live transcription feature can look “demo successful” while still being architecturally broken. Validate in layers.

### 8.1 Unit-level tests

Add fake tests for:

- transcript accumulator ordering,
- partial/final dedupe,
- server protobuf mapping,
- frontend reducer updates,
- backend adapter error mapping.

### 8.2 Pipeline-level tests

Use fake GStreamer audio sources where possible.

Examples:

- `audiotestsrc` instead of microphone
- appsink callback assertions for sample format/rate/channel count

Validate:

- appsink gets `S16LE`, `16000`, `1 channel`
- recording output still finalizes
- meter events still arrive

### 8.3 Backend integration tests

Against a fake transcription backend:

- session start succeeds
- partial update emitted
- final update emitted after flush
- stop closes session
- out-of-order sequence errors handled gracefully

### 8.4 Web/server integration tests

Add an E2E similar in spirit to existing phase scripts:

- start recording
- confirm transcript websocket events flow
- adjust live gain mid-recording
- verify transcript continues
- stop recording
- verify final transcript artifact exists

### 8.5 Manual validation checklist

1. Start recording with transcription enabled.
2. Speak continuously for ~20 seconds.
3. Confirm UI shows incremental transcript updates.
4. Change source gain while recording.
5. Confirm transcript continues rather than silently dying.
6. Stop recording.
7. Confirm transcript output files exist and are non-empty.
8. Reload the browser during recording and verify transcript bootstrap works.

---

## 9. Risks, Alternatives, and Open Questions

## 9.1 Biggest risks

### Risk: transcription branch destabilizes recording EOS/finalization

This is the same kind of architectural mistake that bit the video shared-capture work. Any new tee/appsink branch must be validated carefully so it does not prevent recording shutdown.

### Risk: backend startup latency makes recording feel broken

If screencast-studio itself tries to cold-start a Dagger/Nemotron backend on record button press, startup could feel slow or flaky. This is why the transcription seam should support an externally managed backend and why transcription should be explicitly enabled.

### Risk: transcript reconnect/bootstrap is underspecified

If only delta events are sent and no snapshot exists, browser reloads mid-recording will show empty transcript state. Do not skip the snapshot design.

### Risk: direct code reuse from `transcription-go` becomes messy

Because its useful packages are `internal/...`, naive reuse will lead to awkward copying or illegal imports. Decide early whether this ticket is protocol-level integration only or also an upstream extraction effort.

## 9.2 Alternatives considered

### Alternative A: shell out to `transcribe` CLI during recording

Rejected because:

- it is the wrong control boundary,
- it duplicates lifecycle management,
- it makes partial/final events awkward,
- and it gives the browser no clean live state model.

### Alternative B: browser talks directly to transcription backend

Rejected because:

- screencast-studio already owns the live product session,
- it splits state ownership,
- it complicates reconnect/auth/config,
- and it bypasses the app's existing websocket/event infrastructure.

### Alternative C: use chunk HTTP uploads only

Acceptable as a debug fallback, but not recommended as the final design because session-oriented streaming gives better semantics for live partial/final results.

## 9.3 Open questions

1. Should screencast-studio start the ASR backend itself, or should the backend be externally configured by endpoint for V1?
2. Should transcript outputs be persisted in V1, or is live UI-only good enough first?
3. Should transcript events flow through `RecordingEvent`/`RunEvent`, or through a dedicated callback channel?
4. Should transcript words be sent to the UI as full snapshots or as patches? I recommend snapshots first.
5. Should SCS-0012 Phase 5 be updated to point at this ticket as the more precise architecture for transcription work?

---

## 10. Recommended Task Breakdown for the New Ticket

If this work becomes an implementation ticket sequence, I would break it down like this:

1. **Create transcription seam + protobuf event types**
2. **Add TranscriptionManager + websocket bootstrap support**
3. **Add GStreamer transcription appsink branch**
4. **Add websocket backend adapter for transcription-go service**
5. **Publish transcription updates into server `/ws`**
6. **Add minimal transcript UI panel**
7. **Add transcript sidecar output writing**
8. **Add fake/integration tests and operator docs**

That order preserves the rule that every step should be testable before the next one starts.

---

## 11. References and File Map

### `screencast-studio`

- `pkg/media/gst/recording.go:456-495` — audio worker startup and event watch loop
- `pkg/media/gst/recording.go:576-624` — current audio recording pipeline construction
- `pkg/media/gst/recording.go:660-709` — current post-mix output chain with compressor, level, encoder, filesink
- `pkg/media/gst/recording.go:773-826` — audio level bus-message parsing
- `pkg/app/application.go:196-245` — recording runtime start path and event conversion
- `internal/web/application.go:11-18` — app/web service seam
- `internal/web/handlers_api.go:75-161` — recording start and audio effects HTTP APIs
- `internal/web/session_manager.go:277-290` — current audio-meter publication path
- `internal/web/handlers_ws.go:16-74` — websocket bootstrap and event streaming
- `internal/web/event_hub.go:26-48` — in-process event fan-out
- `proto/screencast/studio/v1/web.proto:223-252` — current `AudioMeterEvent` and `ServerEvent` schema
- `ui/src/features/session/wsClient.ts:13-73` — browser websocket decode and dispatch path
- `ui/src/features/session/sessionSlice.ts:10-109` — current live session Redux state

### `transcription-go`

- `go.mod` — separate module; useful internals live under `internal/...`
- `internal/server/dagger.go:24-118` — Dagger-managed ASR server lifecycle
- `server/server.py:1-185` — health, full-file, and chunk endpoints
- `server/server.py:187-314` — websocket streaming endpoint
- `server/live_sessions.py:11-122` — live session registry and protocol rules
- `server/live_decoder.py:13-109` — buffered PCM decoder, partial flush model
- `internal/asr/client.go:84-120` — batch and chunk HTTP client contracts
- `internal/live/wsclient.go:17-118` — websocket client protocol
- `internal/live/types.go:5-23` — transcript event and state model
- `internal/live/accumulator.go:14-152` — committed/pending transcript merge logic
- `internal/live/runner.go:19-116` — live orchestration and transport selection
- `internal/live/stream_receiver.go:8-57` — mapping websocket messages into transcript updates
- `internal/output/format.go:16-109` — subtitle segmentation and writers
- `internal/output/sqlite.go:10-109` — transcript SQLite schema and writing

---

## 12. Final Recommendation

The best integration plan is not “copy the CLI” and not “bolt a whisper subprocess onto the side.” The best integration plan is to treat `transcription-go` as a **backend protocol and architecture prototype**, then integrate it into `screencast-studio` through a proper seam:

- GStreamer produces normalized live PCM,
- a transcription backend adapter consumes it,
- the app owns transcript state,
- and the browser receives transcript updates through the existing `/ws` path.

That approach preserves the hard-won structure of the current recording system and gives the team a live transcription feature that looks like it belongs in the app, rather than like a temporary experiment stapled onto it.
