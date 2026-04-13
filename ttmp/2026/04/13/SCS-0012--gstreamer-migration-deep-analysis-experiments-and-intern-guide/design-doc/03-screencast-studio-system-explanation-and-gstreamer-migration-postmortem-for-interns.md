---
Title: Screencast Studio System Explanation and GStreamer Migration Postmortem for Interns
Ticket: SCS-0012
Status: active
Topics:
    - screencast-studio
    - backend
    - gstreamer
    - ffmpeg
    - audio
    - video
    - screenshots
    - postmortem
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/handlers_api.go
      Note: |-
        Main HTTP recording/audio control endpoints and the current no-suspend recording start behavior
        Explains current recording/audio-control endpoints and the no-suspend recording start behavior
    - Path: internal/web/preview_manager.go
      Note: |-
        Preview lifecycle manager, lease model, frame cache, screenshot path, and preview identity model
        Explains preview leasing
    - Path: internal/web/server.go
      Note: |-
        Server wiring and still-present legacy preview handoff cleanup candidate
        Explains current server wiring and the remaining preview handoff cleanup candidate
    - Path: pkg/app/application.go
      Note: |-
        Application-level wiring from compiled plans into the selected recording runtime
        Explains application-to-runtime handoff from compiled plans into media sessions
    - Path: pkg/dsl/compile.go
      Note: The stable compiler that turns the DSL config into executable video and audio jobs
    - Path: pkg/media/gst/recording.go
      Note: |-
        Current GStreamer recording runtime that now routes video jobs through the shared bridge path
        Explains how shared bridge recording was integrated into the real video worker path
    - Path: pkg/media/gst/shared_video.go
      Note: |-
        Shared capture registry and tee-based source fan-out for preview and recording consumers
        Explains the shared source registry and tee-based consumer attachment model
    - Path: pkg/media/gst/shared_video_recording_bridge.go
      Note: |-
        The shared-source recording bridge and the appsrc MP4 finalization fix that unlocked Phase 4
        Explains the bridge recorder design and the h264parse fix
    - Path: pkg/media/types.go
      Note: |-
        Media runtime seam used to decouple the app from FFmpeg vs GStreamer implementation details
        Explains the media runtime seam and the preview/recording session interfaces
    - Path: ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/16-web-gst-default-runtime-e2e/main.go
      Note: Real default-runtime end-to-end validation of no-suspend shared capture
    - Path: ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/17-go-gst-shared-video-tee-experiment/main.go
      Note: Tee-finalization failure evidence and why naive EOS-on-branch did not solve MP4 recording
    - Path: ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/18-go-gst-shared-source-appsink-appsrc-bridge/main.go
      Note: Intermediate bridge experiment that narrowed the failure toward appsrc timing and recorder finalization
    - Path: ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/20-go-gst-shared-bridge-recorder-smoke/main.go
      Note: |-
        Focused harness proving preview continuity plus valid MP4 output for the shared bridge recorder
        Experiment evidence for preview continuity plus valid finalized MP4 output
    - Path: ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/21-go-gst-appsrc-mp4-recorder-smoke/main.go
      Note: |-
        Minimal reproduction that proved the appsrc MP4 path needed h264parse
        Minimal reproduction that proved the recorder-shape bug was not shared-capture-specific
ExternalSources: []
Summary: A detailed intern-facing system explanation, architecture tour, and postmortem of the Screencast Studio FFmpeg-to-GStreamer migration. Explains the stable layers, the runtime seam, the web API, the shared-capture problem, the failed experiments, the root causes, the eventual recorder fix, and the current architecture after the breakthrough.
LastUpdated: 2026-04-13T22:00:00-04:00
WhatFor: Onboard a new intern who needs both a system map and an honest postmortem of what was hard in the GStreamer migration, especially around shared capture, appsrc recording, and preview/recording interaction.
WhenToUse: Read this before touching preview, recording, shared capture, or FFmpeg-removal work. Use it when you need to understand not just what the current design is, but why it ended up this way.
---


# Screencast Studio System Explanation and GStreamer Migration Postmortem for Interns

## 0. Executive Summary

This document is the "you just joined the project and need the whole story" report.

If you only remember five things from this postmortem, remember these:

1. **Screencast Studio is not just a recorder.** It is a small system with a DSL compiler, a web/API layer, preview lifecycle management, recording session management, and a media runtime underneath.
2. **The DSL/compiler/web layers were mostly stable throughout the migration.** The hard work was in the media runtime layer.
3. **The hardest problem was not “replace FFmpeg with GStreamer.”** The hardest problem was **shared live capture**: letting preview and recording use the same live source without breaking preview continuity or recording finalization.
4. **A huge amount of time was lost productively, not wastefully.** The "struggles" were mostly necessary experiments that separated architecture problems from recorder-shape problems.
5. **The decisive technical breakthrough was small but subtle:** the `appsrc -> x264enc -> mp4mux` path needed `h264parse` in between. Without that parser, the recorder produced tiny invalid MP4 files and misleading segment/timing errors.

This report explains the system, the migration strategy, what failed, why it failed, what experiments taught us, what finally worked, and what cleanup is still left.

---

## 1. Audience, Reading Strategy, and Expectations

This report is written for a new intern or engineer who:

- knows Go reasonably well,
- is not yet fluent in GStreamer,
- and has not internalized the architecture of Screencast Studio.

You should expect three different kinds of material here:

- **System explanation:** what the application is, what components exist, and how a request flows through the stack.
- **Design explanation:** why the current architecture is shaped the way it is.
- **Postmortem explanation:** what went wrong, what assumptions failed, and how the team learned its way to the right fix.

A good way to read the document is:

1. Read Sections 2–6 to understand the stable system.
2. Read Sections 7–10 to understand the migration strategy and why Phase 4 mattered.
3. Read Sections 11–16 as the real postmortem of the struggles.
4. Keep Sections 17–20 open as a debugging and implementation reference.

---

## 2. What Screencast Studio Actually Is

Screencast Studio is a desktop recording application whose backend is driven by a YAML-like DSL. A user describes the desired capture setup — such as which display or region to capture, which microphone to use, where outputs should go, and what codecs to use — and the system turns that description into a running preview or recording session.

That sentence sounds simple, but it hides several layers that matter a lot to this migration:

- a **discovery layer** that knows what displays, windows, cameras, and audio devices exist,
- a **DSL normalization layer** that turns user input into an effective config,
- a **compiler layer** that turns config into executable jobs,
- a **web layer** that exposes those operations through HTTP and WebSockets,
- a **preview manager** that owns long-lived preview sessions,
- a **recording manager** that owns long-lived recording sessions,
- and a **media runtime** that does the real media work.

For a long time, the media runtime was mostly FFmpeg subprocess orchestration. The migration goal was to replace that with native GStreamer graph management without destabilizing the layers above it.

### 2.1 Mental model

Think of the app as this pipeline:

```text
User / Web UI
    ↓
HTTP + WebSocket API
    ↓
Application service
    ↓
DSL normalize + compile
    ↓
Preview manager / Recording manager
    ↓
Media runtime seam
    ↓
FFmpeg or GStreamer implementation
    ↓
Actual OS media sources (X11, PulseAudio/PipeWire, cameras)
```

The migration succeeded only because the team resisted the temptation to rewrite all of these layers at once.

---

## 3. The Stable Core: DSL and Plan Compilation

The most important architectural fact for an intern is that the **DSL and plan compiler are intentionally media-backend-agnostic**.

### 3.1 Why the DSL mattered

The DSL is the project’s contract with the rest of the application. It lets the web layer and the application layer speak in domain terms like:

- video source
- audio source
- destination template
- session id
- output settings
- audio mix

rather than in FFmpeg flags or GStreamer element strings.

That meant the migration could focus on replacing the media execution machinery while keeping the user-facing and web-facing model stable.

### 3.2 Key file references

- `pkg/dsl/types.go`
  - defines the DSL-level data model
- `pkg/dsl/normalize.go`
  - fills defaults and produces an effective configuration
- `pkg/dsl/compile.go`
  - turns the effective config into executable jobs

### 3.3 The compiled plan model

The key compiler function is in:

- `pkg/dsl/compile.go`

The compiler walks enabled video sources and enabled audio sources and produces:

- `VideoJob`s
- `AudioMixJob`s
- `PlannedOutput`s

In simplified pseudocode, the compiler works like this:

```go
func BuildPlan(cfg EffectiveConfig) CompiledPlan {
    outputs := []PlannedOutput{}
    videoJobs := []VideoJob{}
    audioJobs := []AudioMixJob{}

    for each enabled video source:
        outputPath := renderDestinationTemplate(...)
        outputs += PlannedOutput(kind="video", path=outputPath)
        videoJobs += VideoJob(source=src, outputPath=outputPath)

    if any enabled audio sources exist:
        outputPath := renderAudioMixDestination(...)
        outputs += PlannedOutput(kind="audio", path=outputPath)
        audioJobs += AudioMixJob(sources=enabledAudio, outputPath=outputPath)

    return CompiledPlan{videoJobs, audioJobs, outputs, warnings}
}
```

### 3.4 Why this helped the migration

This compiler produced a **stable execution plan** that both FFmpeg and GStreamer runtimes could consume. That is one of the main reasons the migration could be incremental.

---

## 4. The Media Runtime Seam

The migration became tractable only after introducing a formal media runtime seam.

### 4.1 Key file

- `pkg/media/types.go`

This file defines the interfaces that the rest of the application depends on:

- `PreviewRuntime`
- `PreviewSession`
- `RecordingRuntime`
- `RecordingSession`

These interfaces matter more than they look, because they define **what the upper layers are allowed to know** about media execution.

### 4.2 Important interface summary

From `pkg/media/types.go`:

- `PreviewRuntime.StartPreview(...)`
- `PreviewSession.Wait()`
- `PreviewSession.Stop(...)`
- `PreviewSession.LatestFrame()`
- `PreviewSession.TakeScreenshot(...)`
- `RecordingRuntime.StartRecording(...)`
- `RecordingSession.Wait()`
- `RecordingSession.Stop(...)`
- `RecordingSession.SetAudioGain(...)`
- `RecordingSession.SetAudioCompressorEnabled(...)`

This was the seam that allowed the codebase to have:

- an FFmpeg-backed implementation during early migration phases,
- and a GStreamer-backed implementation later,
- without rewriting the whole server or app layer each time.

### 4.3 Why this seam was strategically correct

Without the seam, the migration would have become a giant “delete FFmpeg, add GStreamer, hope everything still works” rewrite. With the seam, each new runtime slice could be validated independently:

- preview first,
- recording next,
- screenshots,
- audio controls,
- audio meter,
- shared capture last.

That last ordering mattered a lot, because the shared-capture problem turned out to be much harder than the basic preview and recording problems.

---

## 5. The Application Layer: Where Plans Become Running Work

The application layer is the bridge between compiled plans and runtime execution.

### 5.1 Key file

- `pkg/app/application.go`

### 5.2 What the application layer does

The `Application` struct:

- exposes discovery,
- normalizes DSL,
- compiles DSL,
- and calls the recording runtime with a compiled plan.

In practice, it is the point where the project converts:

```text
"Here is the user’s intent"
```

into:

```text
"Here is a specific runtime call with a concrete execution plan"
```

### 5.3 Why this layer is important to the intern

If you are debugging a recording request end-to-end, the application layer is where you confirm whether a problem is:

- an input/config problem,
- a compiler problem,
- or a runtime/media problem.

### 5.4 Simplified pseudocode

```go
func (a *Application) RecordPlan(ctx, plan, opts) {
    session := recordingRuntime.StartRecording(ctx, plan, mediaOpts)
    result := session.Wait()
    return summary(result)
}
```

This simplicity is intentional. The application layer should not know whether the runtime uses subprocesses, native pipelines, or magic.

---

## 6. The Web/API Layer: What the Browser Talks To

The web layer matters because many of the migration struggles only showed up when the runtime was exercised through the real preview and recording endpoints rather than through isolated smoke tests.

### 6.1 Key files

- `internal/web/routes.go`
- `internal/web/handlers_api.go`
- `internal/web/handlers_preview.go`
- `internal/web/server.go`
- `proto/screencast/studio/v1/web.proto`

### 6.2 Important HTTP endpoints

From `internal/web/routes.go`, the main routes are:

- `GET /api/healthz`
- `GET /api/discovery`
- `GET /api/session`
- `POST /api/setup/normalize`
- `POST /api/setup/compile`
- `POST /api/recordings/start`
- `POST /api/recordings/stop`
- `GET /api/recordings/current`
- `POST /api/audio/effects`
- `POST /api/previews/ensure`
- `POST /api/previews/release`
- `GET /api/previews`
- `GET /api/previews/{id}/mjpeg`
- `GET /api/previews/{id}/screenshot`
- `GET /ws`

### 6.3 Why these routes mattered in the migration

A common mistake in media migrations is to prove that a low-level pipeline works in isolation and assume the product now works. That assumption is almost always false.

In this project, the team repeatedly had to validate not only:

- “does a GStreamer pipeline run?”

but also:

- “does the preview manager still behave correctly?”
- “does the web API still expose the right lifecycle behavior?”
- “does screenshot retrieval work through the preview session abstraction?”
- “do live audio control requests arrive while recording is still active?”
- “does the WebSocket audio meter still show up in real time?”

### 6.4 Important API contract references

The protobuf file `proto/screencast/studio/v1/web.proto` defines:

- discovery responses,
- normalized config and compile responses,
- recording start request,
- session envelope,
- preview descriptor and preview list,
- audio meter events,
- and server event multiplexing over WebSockets.

That matters because the migration was not allowed to casually break or redesign these payloads.

---

## 7. The Preview Side of the World

Preview was simpler than recording, but only relatively simpler.

### 7.1 Key files

- `internal/web/preview_manager.go`
- `pkg/media/gst/preview.go`
- `pkg/media/gst/shared_video.go`

### 7.2 Preview manager responsibilities

The preview manager owns concepts that are not just media concerns:

- preview identity,
- preview leases,
- latest-frame cache,
- preview list snapshots,
- screenshot lookup by preview id,
- and state transitions visible to the web layer.

This is important: a preview is not just a GStreamer pipeline. It is a **managed session** with API-visible state.

### 7.3 The lease model

The manager uses a signature-based identity model so equivalent preview requests can reuse the same preview session.

Conceptually:

```go
if preview for this source signature already exists:
    increment lease count
    return existing preview descriptor
else:
    create preview
```

This was one of the first hints that shared capture might be the right long-term architecture not just for preview reuse, but for preview + recording coexistence too.

### 7.4 Screenshot path

The preview screenshot endpoint is intentionally simple at the web layer:

```text
GET /api/previews/{id}/screenshot
```

Internally that becomes:

```text
PreviewManager.TakeScreenshot(previewID)
    → PreviewSession.TakeScreenshot(...)
    → GStreamer preview session returns latest JPEG frame
```

That simplicity was valuable because it gave Phase 4 a concrete test: if shared capture is truly working, screenshots should continue to work while recording is active.

---

## 8. The Legacy FFmpeg World and Why It Looked Easier Than It Was

Before the migration, FFmpeg did not just act as a media engine. It also hid many architectural problems behind process boundaries.

### 8.1 Key legacy files

- `pkg/recording/ffmpeg.go`
- `pkg/recording/run.go`
- `internal/web/preview_runner.go`

### 8.2 What FFmpeg gave the old design “for free”

The FFmpeg-based design made a lot of operations feel simple because each preview or recording was effectively its own external process.

That gave apparent advantages:

- separate process boundaries,
- simple stop semantics,
- isolated stdout/stderr logs,
- fewer concerns about graph lifetime inside one process,
- and fewer in-process ownership issues.

### 8.3 Why that became a problem during migration

Once GStreamer moved in-process, the team had to deal explicitly with things FFmpeg had hidden:

- graph lifetime,
- shared source ownership,
- EOS ordering,
- branch finalization,
- buffer timestamp semantics,
- and how preview state and recording state interact when they are no longer separate OS processes.

This is the deeper reason why Phase 4 was never “cleanup.” It was a real architecture phase.

---

## 9. The Phase Structure of the Migration

The migration only worked because it was phased.

### 9.1 High-level phase story

- **Phase 0:** introduce a runtime seam so FFmpeg and GStreamer can coexist.
- **Phase 1:** make GStreamer preview work.
- **Phase 2:** make GStreamer recording work for video and audio.
- **Phase 3:** add feature parity pieces like screenshots, live audio controls, and audio meter plumbing.
- **Phase 4:** solve shared capture honestly.
- **Phase 5:** later work like transcription.

### 9.2 Why Phase 4 was different

Phases 1–3 were mostly about proving that GStreamer could do the same jobs FFmpeg had been doing.

Phase 4 was about something more subtle:

- preview and recording should stop acting like two unrelated consumers that happen to point at the same hardware or desktop source,
- and should instead share one capture graph safely.

That was the phase where many of the struggles happened.

---

## 10. The Real Shared-Capture Goal

The long-term target behavior was always this:

```text
One live source capture
    ├── preview branch
    └── recording branch
```

The user-visible consequences should be:

- preview stays visible while recording starts,
- preview does not have to be suspended,
- recording finalizes correctly,
- screenshots keep working,
- audio effects remain adjustable,
- audio meter events continue to flow,
- and no duplicate source capture causes conflicts or finalization breakage.

The mistake would have been to describe this as a cosmetic UX improvement. It was not. It changed source ownership and shutdown semantics.

---

## 11. Postmortem Part I: The First Wrong Mental Model

The first wrong mental model was:

> "Once GStreamer is the default preview runtime and default recording runtime, we can probably remove preview suspend/restore."

That was understandable, but incomplete.

### 11.1 Why the assumption sounded reasonable

By that point, individual slices already worked:

- preview worked,
- recording worked,
- screenshots worked,
- audio meter worked,
- audio effects worked,
- and GStreamer was already the default runtime.

So it was tempting to think the remaining work was mainly deleting legacy coordination logic.

### 11.2 What actually happened

A real-defaults harness showed that removing suspend/restore naively could leave preview alive, but recording finalization could still fail.

Observed failures included:

- `timed out waiting for recording EOS`
- invalid output files
- conflicts that only appeared when preview and recording were both active in real application flow

### 11.3 Lesson

The lesson was:

- **feature parity of independent preview and recording paths is not the same thing as shared-capture safety**.

That distinction shaped the rest of Phase 4.

---

## 12. Postmortem Part II: Why the Pure Tee Finalization Approach Failed

One possible design was to keep one shared source pipeline and let recording be another branch of the same tee. That sounds attractive because it is conceptually clean.

### 12.1 The pure tee idea

```text
shared source → tee
    ├── preview branch → appsink/JPEG
    └── record branch → encoder → muxer → file
```

Then the recording branch would be attached when recording starts and detached when recording stops.

### 12.2 Why this was harder than it looked

MP4 finalization depends on a clean EOS/finalization path through the muxer. But in a tee-shared graph, sending EOS “to the recording branch” is not trivial.

The team tried several EOS placements. The results were bad in different ways:

- EOS far enough upstream to finalize MP4 could poison or terminate the shared pipeline itself.
- EOS closer to the muxer could preserve preview better but still produce invalid MP4 output.
- Removing a branch without the right finalization sequence kept preview alive but produced broken files.

### 12.3 Key experiment evidence

The focused tee experiment in:

- `scripts/17-go-gst-shared-video-tee-experiment/main.go`

showed outcomes like:

- `BUS eos from shared-video-tee`
- `moov atom not found`
- `Invalid data found when processing input`

### 12.4 Lesson

The lesson was not “tees are bad.” The lesson was:

- **for this app and this MP4 finalization goal, making recording just another detachable tee branch was structurally awkward**.

That pushed the design toward a bridge architecture.

---

## 13. Postmortem Part III: The Shared Appsink/Appsrc Bridge Idea

The next architecture was more promising:

- one shared source pipeline,
- preview as a branch of that source,
- but recording as a **separate pipeline** fed by samples from the shared source.

### 13.1 The bridge concept

```text
shared source pipeline
    ├── preview branch → jpeg/appsink
    └── raw branch → appsink
                    ↓ application callback
separate recorder pipeline
    appsrc → encoder → muxer → file
```

This approach was attractive because it decoupled:

- preview continuity,
- from recorder finalization.

Instead of trying to surgically EOS one branch of a shared muxing graph, the app could let preview remain untouched while a completely separate recorder pipeline was started and stopped.

### 13.2 Why the idea was sound

Architecturally, this cleanly separated concerns:

- the shared source owns capture,
- the preview consumer owns preview rendering,
- the recorder pipeline owns file finalization.

That is a much healthier ownership split than trying to let one tee branch both belong to the shared graph and also have independent recording lifecycle semantics.

### 13.3 Why it still failed at first

Even though the architecture was promising, the first implementations still produced broken outputs.

Observed failures included:

- `record bus error from record-appsrc: Internal data stream error.`
- `gst_segment_to_running_time: assertion 'segment->format == format' failed`
- tiny outputs like `850 bytes`, `852 bytes`, `854 bytes`
- `duration=N/A`

At first glance, this still looked like the bridge architecture might be wrong.

But that would have been the wrong conclusion.

---

## 14. Postmortem Part IV: The Critical Diagnostic Move

This was the most important diagnostic move in the entire migration:

> build a minimal synthetic `appsrc` MP4 recorder with no shared source at all.

### 14.1 Why this was so important

The bridge path still had many moving parts:

- shared source registry,
- tee branch attach/detach,
- raw normalization,
- appsink callback,
- buffer copy,
- appsrc push,
- encoder,
- muxer,
- filesink.

If you try to debug all of those together, you can easily blame the wrong subsystem.

### 14.2 The new minimal reproduction

A new experiment was created:

- `scripts/21-go-gst-appsrc-mp4-recorder-smoke/main.go`

It generated synthetic frames and pushed them through:

```text
appsrc -> videoconvert -> x264enc -> mp4mux -> filesink
```

### 14.3 The result

It failed with the **same signature** as the shared bridge path:

- `gst_segment_to_running_time: assertion 'segment->format == format' failed`
- invalid MP4 around ~850 bytes

### 14.4 Why this changed everything

This proved that the remaining bug was **not fundamentally a shared-capture problem**.

It was a narrower problem:

- the exact appsrc-driven MP4 recorder shape was wrong.

That is a huge difference.

### 14.5 Lesson

If you remember one debugging lesson from this whole postmortem, remember this:

- **when a complex architecture seems broken, first ask whether the lowest-level reproducer is also broken.**

That question saved a lot of wasted redesign.

---

## 15. Postmortem Part V: What the Team Tried Before the Real Fix

Before the decisive fix, several recorder-side adjustments were explored. They were not wasted. They narrowed the problem.

### 15.1 Raw branch normalization

The raw branch originally surfaced suspicious caps. The recorder path was seeing shapes and formats that did not look like the intended recording target. So the raw consumer branch was normalized explicitly with a transform chain like:

```text
queue
→ videoconvert
→ videoscale
→ videorate
→ capsfilter(format=I420,width=640,height=480,framerate=10/1,pixel-aspect-ratio=1/1)
→ appsink
```

This was crucial because it proved:

- buffers were really flowing,
- caps were stable,
- and the problem was not just malformed raw input.

### 15.2 Appsrc property tuning

Recorder-side `appsrc` configuration was also tightened:

- `format = GST_FORMAT_TIME`
- `is-live = true`
- explicit timestamps and durations on buffers
- deterministic frame timing

These changes did not fix the invalid MP4 issue, but they were still valuable because they eliminated several plausible timing/segment hypotheses.

### 15.3 What those attempts proved

After normalization and instrumentation, the team could say with confidence:

- the shared source was alive,
- preview continuity worked,
- the recorder was getting frames,
- and the recorder was pushing many buffers.

That narrowed the failure to the recorder pipeline shape itself.

---

## 16. Postmortem Part VI: The Actual Recorder Fix

The decisive fix was adding:

- `h264parse`

between:

- `x264enc`
- and `mp4mux`

### 16.1 The corrected shape

```text
appsrc
→ videoconvert
→ x264enc
→ h264parse
→ mp4mux
→ filesink
```

### 16.2 Why this mattered

With the parser present, the synthetic appsrc recorder started finalizing valid MP4 output immediately.

That strongly suggested that the parser was providing something the muxer/finalization path needed for this appsrc-fed encoded stream shape.

### 16.3 The practical lesson

Even if a direct-source recording pipeline appears to work without a parser, an appsrc-fed encoded path may still need one.

This is a classic multimedia integration lesson:

- **the fact that two pipelines contain the same major elements does not mean they have the same negotiation and finalization behavior.**

### 16.4 The evidence

The synthetic script `21` went from invalid tiny files to valid MP4 with a real duration.

Then the same fix was applied to:

- `pkg/media/gst/shared_video_recording_bridge.go`

and the production shared bridge harness began producing valid MP4 while preview stayed alive.

That was the turning point.

---

## 17. The Architecture After the Breakthrough

The current architecture is much closer to the intended final design.

### 17.1 Shared source registry

Key file:

- `pkg/media/gst/shared_video.go`

This registry:

- computes a source signature,
- reuses existing live capture for equivalent sources,
- tracks preview consumers and raw consumers,
- reference-counts source users,
- and shuts the shared source down when the last consumer is gone.

### 17.2 Preview behavior

Preview now attaches as a consumer branch to the shared source.

Conceptually:

```text
captureRegistry.acquireVideoSource(signature)
    → sharedVideoSource
        → tee
            → preview consumer branch
```

### 17.3 Recording behavior

Video recording now uses the shared bridge path.

Conceptually:

```text
captureRegistry.acquireVideoSource(signature)
    → sharedVideoSource
        → raw consumer branch → appsink callback
            → appsrc recorder pipeline
                → videoconvert
                → x264enc
                → h264parse
                → mp4mux
                → filesink
```

Audio remains on the stable dedicated GStreamer path in `pkg/media/gst/recording.go`.

### 17.4 High-level diagram

```text
                    +----------------------------------+
                    |   sharedVideoSource pipeline     |
OS source ----------> source elems -> convert -> tee  |
                    +----------------+-----------------+
                                     |
                   +-----------------+------------------+
                   |                                    |
                   v                                    v
        preview consumer branch              raw consumer branch
        queue -> jpeg/appsink                queue -> normalize -> appsink
                   |                                    |
                   v                                    v
             PreviewSession                     Go callback / sample copy
                                                        |
                                                        v
                                     +----------------------------------+
                                     | separate bridge recorder pipeline |
                                     | appsrc -> x264enc -> h264parse   |
                                     |        -> mp4mux -> filesink     |
                                     +----------------------------------+
```

---

## 18. How Recording Start Works Now

### 18.1 Recording path summary

Key file:

- `pkg/media/gst/recording.go`

The recording runtime now starts workers for:

- each video job,
- each audio mix job,
- and supervises them under one session model.

### 18.2 Important nuance

The video worker model did **not** need to be redesigned at the session layer. Instead, the video worker’s implementation changed underneath.

This is good architecture. The upper session model still says:

- start worker,
- emit events,
- wait for worker,
- stop worker on cancel,
- aggregate result.

But the worker now uses shared-capture bridge recording instead of its old standalone pipeline path.

### 18.3 Pseudocode

```go
func startVideoRecordingWorker(ctx, job) {
    source := resolveWindowSource(job.Source)
    recorder := StartExperimentalSharedVideoRecorder(ctx, source, ...)

    go func() {
        resultCh <- recorder.Wait()
    }()

    return worker{
        stopFn: func(timeout) { recorder.Stop(timeoutCtx) },
    }
}
```

This preserved the higher-level recording lifecycle while swapping out the implementation detail that actually mattered.

---

## 19. How the Web Layer Changed

The biggest web-layer behavioral change was simple to describe but important to justify:

- `POST /api/recordings/start` no longer suspends previews first.

### 19.1 Key file

- `internal/web/handlers_api.go`

### 19.2 Why this was safe now

It became safe only after all of the following were true:

- shared source registry worked,
- preview sharing worked,
- bridge recording worked,
- MP4 finalization worked,
- default runtime end-to-end validation proved preview continuity,
- screenshots worked during recording,
- live audio effects still worked during recording,
- audio meter events still flowed.

### 19.3 What is still not fully cleaned up

There is still leftover preview handoff bookkeeping in:

- `internal/web/server.go`

That is now mostly legacy cleanup work. It is important to tell the intern this explicitly:

- **some code still exists because cleanup lags proof.**

That is normal and healthy. Cleanup should follow evidence, not precede it.

---

## 20. API-Level Behavior the Intern Should Know

These are the behaviors you should assume are important to preserve.

### 20.1 Preview API

- `POST /api/previews/ensure`
  - idempotent-ish by source signature because the preview manager reuses previews
- `POST /api/previews/release`
  - decreases lease count and eventually stops preview
- `GET /api/previews`
  - returns active preview list
- `GET /api/previews/{id}/screenshot`
  - should work while preview is running and, now, also while recording is active

### 20.2 Recording API

- `POST /api/recordings/start`
  - starts recording without suspending preview now
- `POST /api/recordings/stop`
  - requests recording stop
- `GET /api/recordings/current`
  - exposes active/finished session state
- `POST /api/audio/effects`
  - updates live gain / compressor settings while recording is active

### 20.3 WebSocket API

- `GET /ws`
  - carries preview state events, logs, recording session events, audio meter events, and disk telemetry

### 20.4 Why these endpoints belong in a postmortem

Because the migration failed if it only proved media correctness without preserving user-visible system behavior.

---

## 21. A Chronological List of the Biggest Struggles

This section is intentionally blunt.

### Struggle 1: Thinking preview + recording feature parity meant shared capture was "almost done"

Why it happened:
- many independent slices were already green

Why it was wrong:
- independent success is not shared-lifecycle success

### Struggle 2: Underestimating muxer/finalization semantics

Why it happened:
- the team had already solved startup and raw frame flow

Why it was wrong:
- finalization is its own problem, especially for MP4

### Struggle 3: Treating tee branch EOS as if it were branch-local and harmless

Why it happened:
- conceptually, a recording branch feels detachable

Why it was wrong:
- EOS and muxer finalization in a shared live graph are not that simple

### Struggle 4: Risk of blaming shared capture architecture for what was really an appsrc recorder-shape problem

Why it happened:
- the failing path lived inside the shared bridge implementation

Why it was wrong:
- the minimal appsrc reproducer proved the same bug without any shared source at all

### Struggle 5: Real-default harnesses had their own timing races

Why it happened:
- end-to-end tests can fail for harness reasons, not product reasons

Why it mattered:
- an intern must learn to distinguish a flaky validator from a broken architecture

---

## 22. Root Cause Analysis

There was not one root cause. There were layers of causes.

### 22.1 Architectural root cause

The original application behavior assumed preview and recording could be managed independently and then coordinated with suspend/restore when necessary. That works tolerably with separate FFmpeg subprocesses. It is much less natural for in-process GStreamer graphs.

### 22.2 Media-graph root cause

A shared live source with a recording branch that needs independent MP4 finalization creates a lifecycle mismatch. Preview wants continuity. Recording wants isolated finalization semantics.

### 22.3 Recorder-shape root cause

The appsrc-driven H.264-to-MP4 path needed `h264parse`. Without it, the recorder produced misleading segment/timing failures and invalid tiny MP4 outputs.

### 22.4 Process root cause

It was initially tempting to treat successful isolated preview and successful isolated recording as evidence that shared capture would be easy. The migration had to learn, painfully but honestly, that lifecycle composition is its own engineering problem.

---

## 23. What the Team Did Well

This is important in a postmortem. Not everything was failure.

### 23.1 Strong phased strategy

The runtime seam prevented a rewrite disaster.

### 23.2 Honest rollback behavior

When naive removal of suspend/restore failed, the code was rolled back to a stable state instead of pretending the migration was done.

### 23.3 Focused experiments

Scripts `17`, `18`, `20`, and `21` were excellent engineering moves because each isolated a narrower question.

### 23.4 Good evidence discipline

The team repeatedly required:

- real output files,
- `ffprobe` validation,
- web-level validation,
- and not just “pipeline started” logs.

This discipline is one of the reasons the migration succeeded.

---

## 24. What the Team Could Easily Have Done Wrong, But Did Not

A new intern should appreciate how easy it would have been to take bad shortcuts here.

The team could have:

- deleted suspend/restore early and shipped a flaky system,
- removed FFmpeg before shared capture was solved,
- declared the bridge architecture wrong without creating a minimal appsrc reproducer,
- or hidden the problem behind a "works on my machine" narrative.

Instead, the team:

- preserved stable behavior during uncertainty,
- kept experimental paths isolated,
- and only switched defaults once the real-default harnesses were persuasive.

That is good engineering practice.

---

## 25. File-by-File Reading Guide for a New Intern

If you want to understand the current system with minimal confusion, read files in this order.

### Step 1: Domain and compiler

- `pkg/dsl/types.go`
- `pkg/dsl/normalize.go`
- `pkg/dsl/compile.go`

Questions to answer:
- what can the user describe?
- what becomes a `VideoJob`?
- what becomes an `AudioMixJob`?

### Step 2: Runtime seam

- `pkg/media/types.go`
- `pkg/app/application.go`

Questions to answer:
- what do the upper layers require from a preview runtime?
- what do they require from a recording runtime?

### Step 3: Web lifecycle

- `internal/web/routes.go`
- `internal/web/handlers_api.go`
- `internal/web/handlers_preview.go`
- `internal/web/preview_manager.go`

Questions to answer:
- how are previews identified?
- how are screenshots served?
- how does recording start interact with the web layer now?

### Step 4: Current shared capture implementation

- `pkg/media/gst/shared_video.go`
- `pkg/media/gst/shared_video_recording_bridge.go`
- `pkg/media/gst/recording.go`
- `pkg/media/gst/preview.go`

Questions to answer:
- how is a shared source acquired?
- how do preview and raw consumers attach?
- where does the recording pipeline now start?
- where was `h264parse` inserted?

### Step 5: Historical evidence

- `scripts/17-go-gst-shared-video-tee-experiment/main.go`
- `scripts/18-go-gst-shared-source-appsink-appsrc-bridge/main.go`
- `scripts/20-go-gst-shared-bridge-recorder-smoke/main.go`
- `scripts/21-go-gst-appsrc-mp4-recorder-smoke/main.go`
- `reference/01-diary.md`

Questions to answer:
- what hypothesis did each experiment test?
- what exact failure did each one isolate?

---

## 26. Debugging Playbook for Future Interns

When preview or recording breaks again, use this order.

### 26.1 First classify the failure

Ask:

- Is this a DSL/compile problem?
- A web lifecycle problem?
- A shared-source ownership problem?
- A recorder finalization problem?
- A harness timing problem?

### 26.2 Then pick the smallest relevant harness

Use:

- `scripts/21` for pure appsrc MP4 issues
- `scripts/20` for shared bridge recorder issues
- `scripts/19` for shared preview issues
- `scripts/16` for real default-runtime no-suspend behavior

### 26.3 Validate final outputs, not just logs

Always check:

```bash
ffprobe -hide_banner -loglevel error -show_entries format=duration,size <file>
```

### 26.4 Do not trust startup alone

A graph that starts is not necessarily a graph that:

- finalizes correctly,
- preserves preview continuity,
- or works through the real app/server seam.

---

## 27. Pseudocode Summary of the Current End-to-End Flow

### 27.1 Preview ensure

```go
HTTP POST /api/previews/ensure
    → PreviewManager.Ensure(dsl, sourceID)
        → app.NormalizeDSL(...)
        → find source in effective config
        → compute source signature
        → if preview exists: reuse and lease++
        → else:
            runtime.StartPreview(source)
            store preview session
            publish preview state
```

### 27.2 Recording start

```go
HTTP POST /api/recordings/start
    → RecordingManager.Start(dsl, maxDuration, gracePeriod)
        → app.CompileDSL(...)
        → app.RecordPlan(...)
            → recordingRuntime.StartRecording(plan)
                → for video job:
                    acquire shared source
                    attach raw consumer
                    start appsrc recorder bridge
                → for audio job:
                    start stable audio pipeline
```

### 27.3 Screenshot during recording

```go
HTTP GET /api/previews/{id}/screenshot
    → PreviewManager.TakeScreenshot(previewID)
        → PreviewSession.TakeScreenshot(...)
            → latest preview JPEG frame
```

The important thing now is that this works **while recording is active** because preview is no longer being suspended.

---

## 28. Remaining Cleanup and Why It Is Separate From the Postmortem Breakthrough

As of this report, the hard architecture work is mostly done, but cleanup is not fully done.

### 28.1 Still-open cleanup topics

- remove now-stale preview handoff bookkeeping from `internal/web/server.go`
- decide how aggressively to remove `PreviewManager.SuspendAll / RestoreSuspended`
- later remove FFmpeg-specific code once all safety checks are complete

### 28.2 Why these are cleanup, not blockers

Because the big question is already answered:

- can the default app/server path keep preview alive during recording while producing valid finalized output?

The answer is now yes.

That changes the remaining work from architectural risk to cleanup and simplification.

---

## 29. Final Lessons for the Intern

I want to end with practical lessons, not just project-specific facts.

### 29.1 Lesson: isolate the smallest reproducible failing subsystem

This was the difference between endlessly doubting shared capture and finding the real recorder bug.

### 29.2 Lesson: do not confuse “works in isolation” with “works in composition”

Preview and recording can each work alone while still failing together.

### 29.3 Lesson: multimedia work is often about shutdown and ownership, not just startup

Many teams focus too much on whether frames are flowing. Real products also care about:

- finalization,
- continuity,
- partial failure behavior,
- and API-visible lifecycle semantics.

### 29.4 Lesson: temporary stability rails are not shameful

Keeping suspend/restore during uncertainty was the correct call. It bought time to learn the problem honestly.

### 29.5 Lesson: a parser element can be the difference between “everything is broken” and “the design is correct”

This migration’s `h264parse` lesson is exactly the kind of subtle media-engineering fact that makes postmortems worth writing down.

---

## 30. Suggested Next Reading After This Report

After this document, read:

1. `design-doc/01-gstreamer-migration-analysis-and-intern-guide.md`
2. `design-doc/02-phase-4-shared-capture-architecture-and-intern-implementation-guide.md`
3. `reference/01-diary.md`
4. `scripts/20-go-gst-shared-bridge-recorder-smoke/main.go`
5. `scripts/21-go-gst-appsrc-mp4-recorder-smoke/main.go`

That reading path will take you from:

- general architecture,
- to the hardest shared-capture problem,
- to the exact experiments that solved it.

---

## Appendix A: Short Glossary

- **DSL:** The user-facing configuration language.
- **Effective config:** Normalized DSL with defaults applied.
- **Compiled plan:** Executable jobs derived from the effective config.
- **Preview session:** A managed live preview object exposed through the web layer.
- **Recording session:** A managed recording lifecycle object.
- **Shared source:** A single live capture graph reused by multiple consumers.
- **Consumer:** A preview or raw branch attached to a shared source.
- **Appsink:** GStreamer element that lets the application pull samples out of a pipeline.
- **Appsrc:** GStreamer element that lets the application push samples into a pipeline.
- **EOS:** End of stream; a key event for clean finalization.
- **Muxer:** Element that writes encoded streams into a container such as MP4.

## Appendix B: One-Page Architecture Diagram

```text
                         Screencast Studio System

  +-------------------+      +-------------------------+
  |  Browser / UI     |<---->| HTTP + WebSocket API    |
  +-------------------+      | internal/web/*          |
                             +-----------+-------------+
                                         |
                                         v
                             +-------------------------+
                             | Application Service     |
                             | pkg/app/application.go  |
                             +-----------+-------------+
                                         |
                   +---------------------+----------------------+
                   |                                            |
                   v                                            v
       +---------------------------+               +---------------------------+
       | DSL Normalize / Compile   |               | Discovery                 |
       | pkg/dsl/*                 |               | pkg/discovery/*           |
       +-------------+-------------+               +---------------------------+
                     |
                     v
       +---------------------------+
       | CompiledPlan              |
       | video jobs + audio jobs   |
       +-------------+-------------+
                     |
         +-----------+-----------+
         |                       |
         v                       v
+-------------------+   +-------------------+
| Preview Manager   |   | Recording Manager |
| preview ids       |   | session states    |
| leases            |   | event fanout      |
| screenshots       |   | stop / wait       |
+---------+---------+   +---------+---------+
          |                       |
          +-----------+-----------+
                      |
                      v
          +-------------------------------+
          | Media Runtime Seam            |
          | pkg/media/types.go            |
          +---------------+---------------+
                          |
                          v
          +-----------------------------------------------+
          | GStreamer Runtime                             |
          | preview.go / recording.go / shared_video.go   |
          | shared_video_recording_bridge.go              |
          +-------------------+---------------------------+
                              |
                              v
          +-----------------------------------------------+
          | OS media sources                              |
          | X11, windows, displays, PulseAudio/PipeWire, |
          | cameras                                       |
          +-----------------------------------------------+
```
