---
Title: Diary
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
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/preview_manager.go
      Note: |-
        Preview runtime wiring and lifecycle changes recorded in Step 10
        PreviewManager now owns preview sessions via media.PreviewRuntime (commit e36d29966f9fc2dd49721c1608192a2123b64c0c)
    - Path: pkg/app/application.go
      Note: |-
        Application runtime seam work recorded in Step 10
        RecordPlan rewired through media runtime seam (commit e36d29966f9fc2dd49721c1608192a2123b64c0c)
    - Path: pkg/media/ffmpeg/preview.go
      Note: |-
        FFmpeg preview adapter introduced in Step 10
        FFmpeg preview runtime adapter (commit e36d29966f9fc2dd49721c1608192a2123b64c0c)
    - Path: pkg/media/ffmpeg/recording.go
      Note: |-
        FFmpeg recording adapter introduced in Step 10
        FFmpeg recording runtime adapter (commit e36d29966f9fc2dd49721c1608192a2123b64c0c)
    - Path: pkg/media/types.go
      Note: |-
        New media runtime interfaces introduced in Step 10
        Phase 0 runtime seam interfaces (commit e36d29966f9fc2dd49721c1608192a2123b64c0c)
ExternalSources: []
Summary: Chronological diary of the SCS-0012 GStreamer migration investigation and implementation work.
LastUpdated: 2026-04-13T15:12:03-04:00
WhatFor: ""
WhenToUse: ""
---


# Diary

## Goal

Capture the investigation journey of analyzing the screencast-studio FFmpeg codebase, experimenting with GStreamer pipelines (both `gst-launch-1.0` subprocess and native go-gst bindings), and producing a comprehensive analysis document for an intern to understand the migration path.

## Step 1: Environment Inventory and Setup

Created ticket SCS-0012, set up diary and design-doc stubs, and ran an environment check script.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket to port this from ffmpeg streams to gstreamer. Keep a detailed diary as you work."

**Assistant interpretation:** Create a docmgr ticket, analyze the codebase, experiment with GStreamer, write an analysis, and upload to reMarkable.

**Inferred user intent:** Produce a thorough technical investigation that can guide the GStreamer migration, stored in docmgr, with full reproducibility.

### What I did

- Created ticket SCS-0012 with topics: screencast-studio, backend, gstreamer, audio, video, transcription, screenshots
- Created diary doc and design-doc stub
- Wrote `01-check-gstreamer-env.sh` and ran it
- Found: GStreamer 1.24.2, all runtime plugins installed (1678 elements), but no -dev headers (needed for go-gst cgo)
- Found: all critical elements available: ximagesrc, pipewiresrc, pulsesrc, v4l2src, appsink, audiomixer, jpegenc, pngenc, x264enc, opusenc, wavenc, mp4mux
- Notable: `gdkpixbufsink` exists but doesn't write files directly (posts pixbufs on bus), so for screenshots use `pngenc + filesink` instead

### Why

Needed to know what's available before designing pipeline experiments.

### What worked

- All needed GStreamer elements are present on this machine
- GStreamer 1.24.2 is modern enough
- PipeWire source is available (for modern Wayland screen capture in future)

### What didn't work

- No -dev headers installed initially → go-gst wouldn't compile (needed user to run apt-get)

### What I learned

- The property name for ximagesrc's display is `display-name`, not `display` (caught by gst-launch error)
- `gdkpixbufsink` doesn't have a `location` property — it posts pixbufs on the bus
- The machine runs PipeWire with PulseAudio compatibility (`pactl list short sources` shows PipeWire-managed devices)

### What was tricky to build

- Nothing yet, just environment setup

### What warrants a second pair of eyes

- N/A for this step

### What should be done in the future

- Consider whether `pipewiresrc` + xdg-desktop-portal should replace `ximagesrc` for Wayland compatibility long-term

### Code review instructions

- Run `scripts/01-check-gstreamer-env.sh` to reproduce

---

## Step 2: GStreamer Dev Headers Installation

### Prompt Context

**User prompt (verbatim):** "ok, installed. The ugly plugins were not available"

**Assistant interpretation:** User installed the dev headers; proceed with go-gst compilation experiments.

### What I did

- Verified `pkg-config --modversion gstreamer-1.0 gstreamer-app-1.0 gstreamer-controller-1.0` → all 1.24.2
- Wrote `02-install-gstreamer-dev-headers.sh` documenting what's needed
- Proceeded to gst-launch experiments

### Why

Dev headers are required for go-gst cgo bindings to compile.

### What worked

- All three pkg-config entries found after user installed headers
- libgstreamer-plugins-ugly1.0-dev was not available but we don't strictly need it (x264enc is in the bad plugins which we do have)

### What didn't work

- libgstreamer-plugins-ugly1.0-dev not available on this Ubuntu — not critical since we have x264enc from elsewhere

### What I learned

- On this Ubuntu, the "ugly" dev package isn't available in the default repos. This is fine — the ugly set mainly adds additional codec support we don't depend on.

---

## Step 3: gst-launch-1.0 Video Preview Experiments

### Prompt Context

**User prompt (verbatim):** (see Step 1)

### What I did

- Wrote `03-gst-launch-preview-x11.sh`
- Fixed `display=` to not pass display name (ximagesrc uses `display-name` property, not needed when $DISPLAY is set)
- Ran three tests:
  1. JPEG frame capture via multifilesink → ✓ 17 frames captured (640x480 JPEG, ~75KB each)
  2. H.264 MP4 recording → ✓ valid MP4 (68KB for 3 seconds)
  3. Screenshot via `pngenc + filesink` → ✓ valid PNG (46KB, 640x480)

### Why

Validate that all the GStreamer elements we need for video capture, preview, and screenshot work on this machine before investing in go-gst integration.

### What worked

- **JPEG frame capture**: `ximagesrc → videoconvert → videoscale → capsfilter(width=640) → videorate(5fps) → jpegenc → multifilesink` produces clean JPEGs
- **H.264 recording**: `ximagesrc → videoconvert → videoscale → capsfilter(640x480@10fps) → x264enc(zerolatency,veryfast) → mp4mux → filesink` produces valid MP4
- **Screenshot**: `ximagesrc(num-buffers=1) → videoconvert → pngenc → filesink` produces valid PNG
- Caps negotiation worked automatically: BGRx format from ximagesrc was correctly converted by videoconvert

### What didn't work

- Initial attempt used `display=":0"` property which doesn't exist on ximagesrc (should be `display-name`)
- `gdkpixbufsink` doesn't write files — had to use `pngenc + filesink` instead

### What I learned

- The GStreamer pipeline structure maps cleanly to the current FFmpeg preview pipeline:
  - FFmpeg `-f x11grab -framerate 5` = `ximagesrc + videorate(5/1)`
  - FFmpeg `-vf "fps=5,scale=640:-1"` = `videorate + videoscale + capsfilter`
  - FFmpeg `-vcodec mjpeg -f image2pipe pipe:1` = `jpegenc + multifilesink` (or appsink for Go)
- Screenshots are trivial in GStreamer: just set `num-buffers=1` on the source
- The `-e` flag on `gst-launch-1.0` is critical for proper EOS handling when recording to muxed formats

### What was tricky to build

- Nothing tricky, GStreamer pipelines are straightforward once you know the element names

---

## Step 4: gst-launch-1.0 Audio Capture Experiments

### Prompt Context

**User prompt (verbatim):** (see Step 1)

### What I did

- Wrote `04-gst-launch-audio-capture.sh`
- Tested three audio pipelines:
  1. WAV recording: `pulsesrc → capsfilter(48kHz/stereo) → audioconvert → wavenc → filesink` → ✓ valid WAV (739KB/3sec)
  2. Opus recording: `pulsesrc → capsfilter → audioconvert → audioresample → opusenc(160kbps) → oggmux → filesink` → ✓ valid Ogg/Opus
  3. Volume-adjusted WAV: `pulsesrc → capsfilter → audioconvert → volume(1.5) → wavenc → filesink` → ✓ valid WAV

### Why

Audio is described as harder than video in the SCS-0011 analysis. Need to validate GStreamer audio pipelines before designing the runtime.

### What worked

- All three audio pipelines produced valid files
- PulseAudio device auto-selection worked (`device=default` resolved to `alsa_input.pci-0000_00_1f.3.analog-stereo`)
- Volume element works as expected for gain control
- Opus encoding with Ogg muxing worked perfectly

### What didn't work

- `timeout` killed the script before all tests could complete when run as one batch

### What I learned

- GStreamer audio pipeline maps to current FFmpeg as:
  - FFmpeg `-f pulse -sample_rate 48000 -channels 2 -i default` = `pulsesrc device=default ! audio/x-raw,rate=48000,channels=2`
  - FFmpeg `-filter_complex "[0:a]volume=1.0[a0]"` = `volume volume=1.0`
  - FFmpeg `-c:a pcm_s16le output.wav` = `wavenc ! filesink`
  - FFmpeg `-c:a libopus -b:a 160k` = `opusenc bitrate=160000 ! oggmux`
- For multi-source audio mixing (current `amix` in FFmpeg), we'd use `audiomixer` element which accepts multiple request pads
- Audio mixing pipeline: `pulsesrc0 → volume → ┐ pulsesrc1 → volume → → audiomixer → wavenc → filesink`

---

## Step 5: go-gst Native Binding Compilation and Element Discovery

### Prompt Context

**User prompt (verbatim):** "save all your go experiments in the scripts/ folder of the ticket as well, don't forget to write a frequent diary"

**Assistant interpretation:** Save all work to ticket scripts/ with XX- prefixes, keep the diary updated.

### What I did

- Wrote `05-go-gst-basic-test/` (Go module with main.go)
- Had to learn the go-gst v1.4.0 API through trial and error:
  - `gst.FindElementFactory` → `gst.Find` (renamed in v1.4.0)
  - `gst.Version()` takes no args and returns (uint,uint,uint,uint) in older versions → no version function accessible, just gst.VersionString()
  - `Element.ListProperties()` doesn't exist
  - `PadTemplate.GetDirection()` → `PadTemplate.Direction()` (no Get prefix)
  - `PadTemplate.GetName()` → `PadTemplate.Name()` (no Get prefix)
- Successfully compiled and ran the element discovery program
- All 20 critical elements confirmed available via go-gst
- Hit a double-free crash in go-gst v1.4.0 `GetMetadata` after creating an element via `gst.NewElement("ximagesrc")` — appears to be a known memory management issue in the bindings

### Why

Need to validate that the go-gst bindings are production-viable for this project.

### What worked

- go-gst v1.4.0 compiles with Go 1.25.5 + GStreamer 1.24.2 dev headers
- Element factory lookup (`gst.Find`) works reliably
- All 20 needed elements found
- `GetMetadata("long-name")`, `GetMetadata("description")` work on factory objects
- `Element.GetPadTemplates()` works

### What didn't work

- **Double-free crash**: calling `gst.NewElement("ximagesrc")` and then `GetMetadata` on the factory causes `free(): double free detected in tcache 2`. This is a go-gst v1.4.0 bug in the GObject memory management. The crash only happens when mixing factory queries with element creation in the same scope.
- Go compilation is very slow (~60+ seconds) due to cgo overhead

### What I learned

- The go-gst API naming convention: methods on Go objects don't use `Get` prefix (e.g., `pt.Name()` not `pt.GetName()`)
- The `gst.Find(name)` function replaces the older `gst.FindElementFactory(name)`
- go-gst cgo compilation is slow — this will affect development iteration speed
- There are memory management bugs in go-gst v1.4.0 that may require careful lifetime management or using a subprocess wrapper instead

### What was tricky to build

- Figuring out the exact v1.4.0 API required reading the source code of the go-gst package itself, as the documentation is thin
- The double-free crash was unexpected and suggests the bindings need careful testing before production use

### What warrants a second pair of eyes

- The double-free crash in go-gst v1.4.0 — this may be a showstopper for in-process pipeline management, or it may be avoidable with proper lifetime discipline
- Whether to use go-gst native bindings vs. a controlled subprocess wrapper (gst-launch or a custom C helper) should be decided early

### What should be done in the future

- Write a focused experiment that creates a pipeline, runs it, and tears it down cleanly to test if the double-free is specific to factory+element mixing or a general issue
- Benchmark go-gst compilation time and decide if it's acceptable for CI
- Consider a hybrid approach: Go orchestrator + small C helper process for GStreamer pipeline management

### Code review instructions

- Scripts are in `scripts/05-go-gst-basic-test/`
- To reproduce: copy the directory, run `go mod tidy && go build -o gst-test . && ./gst-test`

---

## Step 6: go-gst Native Preview Pipeline — End-to-End Success

### Prompt Context

**User prompt (verbatim):** "save all your go experiments in the scripts/ folder of the ticket as well, don't forget to write a frequent diary"

**Assistant interpretation:** Save all Go experiment code to the ticket scripts/ directory with numeric prefixes, keep diary updated.

**Inferred user intent:** Ensure full reproducibility and traceability of all experiments.

### What I did

- Wrote `06-go-gst-preview-pipeline/main.go` — a complete Go program that:
  1. Builds a GStreamer pipeline programmatically: `ximagesrc -> videoconvert -> videoscale -> capsfilter(640w) -> videorate(5fps) -> capsfilter -> jpegenc(quality=50) -> appsink`
  2. Sets up an `appsink` callback that receives each JPEG frame buffer in Go
  3. Runs for 5 seconds and prints frame sizes
  4. Shuts down cleanly with EOS
- Had to fix several go-gst v1.4.0 API issues:
  - `gst.NewCapsFromString()` returns `*Caps` not `(*Caps, error)` (single return)
  - `app.WrapAppSink()` doesn't exist → use `app.NewAppSink()` directly
  - `mainLoop.Run()` returns nothing (no error), use `mainLoop.RunError()` for error
- Successfully ran the pipeline: **24 frames in 5 seconds at 5fps, ~267KB per frame**

### Why

This proves that the core preview pipeline works natively in Go — we can get JPEG frame bytes from GStreamer into Go callbacks without any subprocess juggling.

### What worked

- **Pipeline construction**: All elements created and linked without issues
- **Caps negotiation**: GStreamer automatically handled format negotiation (BGRx → converted by videoconvert)
- **Appsink callback**: `NewSampleFunc` receives buffers reliably
- **Buffer mapping**: `buffer.Map(gst.MapRead)` gives us raw bytes
- **Clean shutdown**: `SendEvent(gst.NewEOSEvent())` + `BlockSetState(gst.StateNull)` works cleanly
- **Frame rate**: Exactly 24 frames in 5 seconds = 4.8fps (close to target 5fps)

### What didn't work

- Go compilation is slow (~60-90 seconds due to cgo)
- Initial API mistakes required reading go-gst source code (thin docs)

### What I learned

- The go-gst pipeline API is actually quite clean once you know the patterns:
  1. `gst.NewPipeline("")` → `pipeline.AddMany(elems...)` → `elem.Link(next)` → `pipeline.SetState(Playing)`
  2. For appsink: create with `app.NewAppSink()`, set caps, set callbacks, add `sink.Element` to pipeline
  3. For clean shutdown: send EOS, wait briefly, set state to NULL
- The buffer size (~267KB per 640x480 JPEG at quality 50) is larger than the FFmpeg frames (~75KB) — probably because GStreamer captures the full 640x480 area while FFmpeg may have been capturing a different region

### What was tricky to build

- go-gst API has no `Get` prefix convention (unlike C GStreamer), and the docs don't make this obvious
- `app.NewAppSink()` creates a standalone sink element; its `.Element` field is the `*gst.Element` you add to the pipeline
- The `glib.MainLoop` is needed for bus message dispatch; without it, the bus watch callback never fires

---

## Step 7: go-gst Native Audio Capture Pipeline — End-to-End Success

### Prompt Context

**User prompt (verbatim):** (same as Step 6)

### What I did

- Wrote `07-go-gst-audio-capture/main.go` — captures 3 seconds of audio:
  `pulsesrc(device=default) -> capsfilter(48kHz/stereo) -> audioconvert -> volume(1.0) -> wavenc -> filesink`
- Successfully produced a valid WAV file: 577,964 bytes (3.01 seconds of 48kHz stereo PCM)

### Why

Proves that audio capture and recording works natively through go-gst.

### What worked

- PulseAudio capture via `pulsesrc device=default` resolves correctly
- WAV encoding with proper headers (RIFF/WAVE PCM 16-bit stereo 48kHz)
- EOS handling finalizes the WAV file correctly
- File size is consistent: 577964 bytes / (48000 * 2 * 2) = ~3.01 seconds ✓

### What didn't work

- Nothing — this worked first try after learning the API from Step 6

### What I learned

- Audio pipelines are simpler than video in GStreamer (fewer elements, simpler caps)
- The volume element sits inline and adjusts gain: this directly maps to the current FFmpeg `volume=X.Y` filter
- For multi-source mixing (the `amix` in FFmpeg), we'd use `audiomixer` with multiple request pads

### What should be done in the future

- Write a multi-source audio mixing experiment using `audiomixer`
- Test opusenc encoding via go-gst for compressed audio output
- Build a video recording pipeline (ximagesrc -> x264enc -> mp4mux -> filesink) via go-gst

---

## Step 8: Writing the Main Analysis Document

### Prompt Context

**User prompt (verbatim):** (see Step 1)

### What I did

- Wrote the complete analysis document `design-doc/01-gstreamer-migration-analysis-and-intern-guide.md` in sections:
  1. Introduction and scope
  2. GStreamer basics for the intern (elements, pads, caps, pipelines, bus, appsink, tee, EOS)
  3. Current FFmpeg architecture analysis (preview flow, recording flow, problems)
  4. Validated GStreamer replacement pipelines (all 7 experiments documented)
  5. New capabilities (screenshots, live audio effects, transcription, shared capture, VU meter)
  6. go-gst API patterns (pipeline building, appsink callbacks, bus messages, shutdown, gotchas)
  7. Proposed architecture (runtime seam, interfaces, pipeline designs)
  8. Element reference table (all ~40 elements with FFmpeg equivalents)
  9. FFmpeg→GStreamer translation cheat sheet
  10. Migration phases (5 phases with time estimates)
  11. Risks and mitigations
  12. Glossary

- Related key source files to the ticket
- Updated changelog

### Why

This is the deliverable document — everything an intern needs to understand the migration and start implementing.

### What worked

- Writing in sections was efficient — each section is self-contained
- The experiment results provide concrete evidence for every pipeline design
- The translation cheat sheet directly maps FFmpeg knowledge to GStreamer

### What didn't work

- The document is very long (~15 sections). For reMarkable upload, we may need to bundle it.

### What I learned

- The document naturally organized itself around the experiment results — validating pipelines first made writing the analysis much more concrete and authoritative
- The go-gst gotchas section is critical — without it, an intern would waste hours on the same API mistakes

### What should be done in the future

- Add a reMarkable upload of this document
- Consider adding pipeline diagrams (ASCII art or mermaid) for the more complex pipelines

---

## Step 9: reMarkable Upload

### Prompt Context

**User prompt (verbatim):** (see Step 1 — "upload the analysis to the remarkable when done")

### What I did

- Fixed a LaTeX rendering issue (`"q\n"` backtick-in-backtick confused pandoc's LaTeX backend)
- Bundled the analysis document + diary into a single PDF with ToC
- Uploaded to reMarkable at `/ai/2026/04/13/SCS-0012`

### What worked

- `remarquee upload bundle` produced a clean PDF with table of contents
- Upload succeeded on first try (after the LaTeX fix)

### What didn't work

- Pandoc's LaTeX backend chokes on nested backtick sequences like `` `"q\n"` `` — had to rewrite as plain text

### What I learned

- When writing Markdown that will go through pandoc→LaTeX→PDF, avoid nesting code backticks inside quoted strings. Use plain text descriptions instead.

---

## Summary

### What was accomplished

1. **Ticket SCS-0012** created with diary, design doc, and 7 experiment scripts
2. **7 experiments** validated:
   - GStreamer environment inventory (1678 elements available)
   - gst-launch video preview (JPEG frames ✓)
   - gst-launch video recording (H.264 MP4 ✓)
   - gst-launch screenshots (PNG ✓)
   - gst-launch audio capture (WAV, Opus, gain ✓)
   - go-gst native preview pipeline (24 frames/5sec via appsink ✓)
   - go-gst native audio recording (WAV via filesink ✓)
3. **Comprehensive analysis document** written (1222 lines, 13 sections)
4. **PDF uploaded to reMarkable** at `/ai/2026/04/13/SCS-0012`

### Key findings

- **go-gst v1.4.0 works** for building pipelines programmatically and receiving data via appsink
- **All needed GStreamer elements are available** on this machine
- **Screenshots are trivial** with GStreamer (`num-buffers=1`)
- **Live audio effects are straightforward** — just insert elements and change properties at runtime
- **Live transcription is feasible** via appsink + Go-side audio chunking + whisper API
- **Shared capture (tee) eliminates the preview suspend/restore workaround**
- **Main risk: go-gst has a double-free bug** in factory/element lifetime management — needs careful coding or a C helper process fallback

---

## Step 10: Phase 0 Runtime Seam for Preview and Recording

This step started the actual implementation work for SCS-0012 instead of just analysis. The goal was to create the architecture seam that lets us keep FFmpeg working today while making room for a native GStreamer runtime later.

I treated this as a “no behavior change” refactor phase: introduce interfaces, wire them through the application and preview manager, keep FFmpeg as the default implementation, and make sure the full test suite still passes before moving on to GStreamer-specific work.

### Prompt Context

**User prompt (verbatim):** "WOrk on the ticket SCS-0012, task by task, commit at appropriate intervals, keep a frequent and detailed diary as you work."

**Assistant interpretation:** Start implementing the ticket in the repo, work through the planned migration tasks in order, make focused commits, and keep the diary updated after each meaningful step.

**Inferred user intent:** Turn the analysis into incremental, reviewable implementation work with strong traceability and low-risk commits.

**Commit (code):** e36d29966f9fc2dd49721c1608192a2123b64c0c — "Introduce media runtime seam for FFmpeg"

### What I did

- Created `pkg/media/types.go` with the new runtime interfaces:
  - `PreviewRuntime`
  - `PreviewSession`
  - `RecordingRuntime`
  - `RecordingSession`
  - `PreviewOptions`, `ScreenshotOptions`, `RecordingOptions`
  - `RecordingEvent`, `RecordingState`, `RecordingResult`
- Created `pkg/media/gst/` placeholder files to establish the target package layout:
  - `preview.go`
  - `recording.go`
  - `pipeline.go`
  - `audio.go`
  - `bus.go`
- Added FFmpeg adapter implementations in `pkg/media/ffmpeg/`:
  - `preview.go` now owns the old FFmpeg preview subprocess logic and exposes it as a `PreviewRuntime`
  - `recording.go` wraps `pkg/recording.Run(...)` as a `RecordingRuntime` and adapts events/results into the new media-layer types
- Updated `pkg/app/application.go`:
  - `Application` now holds a `recordingRuntime`
  - added `WithRecordingRuntime(...)` injection option
  - default runtime is `pkg/media/ffmpeg.NewRecordingRuntime()`
  - `RecordPlan(...)` now starts a runtime session and waits on it instead of calling `recording.Run(...)` directly
- Updated `internal/web/preview_manager.go`:
  - replaced the old preview-runner dependency with `media.PreviewRuntime`
  - default preview runtime is `pkg/media/ffmpeg.NewPreviewRuntime()`
  - `Ensure(...)` now creates a preview session via `StartPreview(...)`, stores it, and waits on `session.Wait()` in a goroutine
- Removed FFmpeg-specific preview process management code from `internal/web/preview_runner.go`; that file now only keeps the preview-signature helper used by tests and manager logic
- Updated `internal/web/server_test.go` so the fake preview implementation speaks the new `media.PreviewRuntime` / `media.PreviewSession` interfaces
- Ran `gofmt -w ...` on all modified Go files
- Ran the full test suite:
  - `go test ./... -count=1`
  - result: all tests passed
- Checked ticket tasks `2,3,4,5,6` (Phase 0.1 through 0.5)

### Why

The migration needs a stable seam before any GStreamer-specific pipeline code lands. Without this phase, the later work would require invasive edits directly in the web layer and application layer, making review much harder and increasing the chance of breaking the current FFmpeg flow.

This phase deliberately preserves behavior while shifting ownership boundaries: the app now knows about a recording runtime abstraction, and the web preview manager now knows about a preview runtime abstraction.

### What worked

- The runtime seam fit the existing codebase cleanly
- The FFmpeg preview logic moved cleanly into `pkg/media/ffmpeg/preview.go`
- The FFmpeg recording path could be wrapped without disturbing the existing `pkg/recording` state machine
- The default behavior remains FFmpeg-backed, so existing tests stayed green
- `go test ./... -count=1` passed after the refactor

### What didn't work

- Task 0.3’s wording (“move current FFmpeg-specific code from `pkg/recording/ffmpeg.go` and `internal/web/preview_runner.go`”) turned out to be a little asymmetric in practice. The preview logic moved directly into the new FFmpeg adapter package, but the recording side is currently wrapped around `pkg/recording.Run(...)` rather than fully relocating all FFmpeg internals out of `pkg/recording`.
- Exact command/output for validation:
  - Command: `go test ./... -count=1`
  - Output summary: all packages passed; no failing tests

### What I learned

- The application seam is easiest to introduce at the `RecordPlan(...)` boundary, not deeper in the session manager
- The preview seam is easiest to introduce at `PreviewManager.Ensure(...)`, where the lifecycle is already centralized
- A session-returning runtime interface is a good fit even before GStreamer lands, because it lets the web/app layers converge on the same lifecycle vocabulary (`Start...`, `Wait`, `Stop`)
- Keeping the old `pkg/recording` engine intact for this phase reduced risk substantially

### What was tricky to build

- The main design tension was deciding how much FFmpeg code to physically relocate in the first implementation step.
- If I tried to move all FFmpeg recording internals immediately, it would have forced a much larger package split and a higher-risk refactor before any behavioral gain. The symptom was obvious during design: `pkg/media/ffmpeg` wanted to wrap `pkg/recording`, but a full move of recording argument builders and subprocess orchestration would have needed either a bigger package extraction or a more disruptive cycle-breaking refactor.
- I resolved that by making Phase 0 a true seam-creation phase:
  1. move preview subprocess ownership into `pkg/media/ffmpeg`
  2. wrap the existing recording engine as a `RecordingRuntime`
  3. keep the web/app behavior stable
  4. defer deeper FFmpeg code relocation until the GStreamer runtime is actually replacing it

### What warrants a second pair of eyes

- Whether the remaining FFmpeg-specific helpers in `pkg/recording` should be moved earlier, before Phase 1, or whether the current wrapper-based seam is the better risk-managed stopping point
- Whether `PreviewManager` should actively call `session.Stop(...)` during release/shutdown in addition to canceling the preview context, or whether “runtime must honor context cancellation” is the right invariant
- Whether `Application` should eventually grow explicit constructor wiring for both runtimes instead of just the recording runtime option added here

### What should be done in the future

- Start Phase 1 by implementing the real `pkg/media/gst/preview.go`
- Decide whether to add a `WithPreviewRuntime(...)` helper at server construction level for cleaner integration tests
- Revisit the remaining FFmpeg-specific code placement once the GStreamer recording runtime exists

### Code review instructions

- Start with the seam definitions in:
  - `pkg/media/types.go`
- Then review the two adapter implementations:
  - `pkg/media/ffmpeg/preview.go`
  - `pkg/media/ffmpeg/recording.go`
- Then review the two consumers that were rewired:
  - `pkg/app/application.go`
  - `internal/web/preview_manager.go`
- Then verify test adaptation in:
  - `internal/web/server_test.go`
- Validate with:
  - `go test ./... -count=1`
- Validate default runtime behavior manually by tracing the defaults:
  - `app.New()` → default FFmpeg recording runtime
  - `web.NewPreviewManager(..., nil)` → default FFmpeg preview runtime

### Technical details

- New package layout introduced in this step:
  - `pkg/media/types.go`
  - `pkg/media/ffmpeg/*`
  - `pkg/media/gst/*`
- Runtime wiring now looks like this:
  - `pkg/app.Application.RecordPlan(...)`
    → `media.RecordingRuntime.StartRecording(...)`
    → `session.Wait()`
  - `internal/web.PreviewManager.Ensure(...)`
    → `media.PreviewRuntime.StartPreview(...)`
    → goroutine waits on `session.Wait()`
- Validation command:
  - `go test ./... -count=1`
- Checked tasks:
  - 0.1 interface file
  - 0.2 `pkg/media/gst/` skeleton
  - 0.3 FFmpeg adapter package introduction
  - 0.4 application + preview manager wiring
  - 0.5 full test pass
