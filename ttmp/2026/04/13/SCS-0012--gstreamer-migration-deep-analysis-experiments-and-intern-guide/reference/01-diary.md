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
    - Path: internal/web/server.go
      Note: Step 14 server option seam for preview runtime injection
    - Path: pkg/app/application.go
      Note: |-
        Application runtime seam work recorded in Step 10
        RecordPlan rewired through media runtime seam (commit e36d29966f9fc2dd49721c1608192a2123b64c0c)
    - Path: pkg/discovery/service.go
      Note: Step 13 exported WindowGeometry for preview runtime reuse
    - Path: pkg/media/ffmpeg/preview.go
      Note: |-
        FFmpeg preview adapter introduced in Step 10
        FFmpeg preview runtime adapter (commit e36d29966f9fc2dd49721c1608192a2123b64c0c)
    - Path: pkg/media/ffmpeg/recording.go
      Note: |-
        FFmpeg recording adapter introduced in Step 10
        FFmpeg recording runtime adapter (commit e36d29966f9fc2dd49721c1608192a2123b64c0c)
    - Path: pkg/media/gst/bus.go
      Note: Step 11 GLib bus watch helper (commit 806c14e630a108ac3dd9670af0eb205c4c1072c9)
    - Path: pkg/media/gst/pipeline.go
      Note: Step 11 pipeline assembly helpers (commit 806c14e630a108ac3dd9670af0eb205c4c1072c9)
    - Path: pkg/media/gst/preview.go
      Note: |-
        Step 11 native GStreamer preview runtime (commit 806c14e630a108ac3dd9670af0eb205c4c1072c9)
        Step 13 geometry fallback fix for window preview
    - Path: pkg/media/gst/recording.go
      Note: |-
        Step 15 native GStreamer video recording runtime
        Step 16 added native GStreamer audio mixing/runtime support
    - Path: pkg/media/types.go
      Note: |-
        New media runtime interfaces introduced in Step 10
        Phase 0 runtime seam interfaces (commit e36d29966f9fc2dd49721c1608192a2123b64c0c)
    - Path: ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/09-go-gst-preview-runtime-smoke/main.go
      Note: Step 11 reproducible preview runtime smoke test (commit 806c14e630a108ac3dd9670af0eb205c4c1072c9)
    - Path: ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/10-window-preview-investigation.sh
      Note: Step 13 investigation script for unreliable XID window capture
    - Path: ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/11-web-gst-preview-e2e/main.go
      Note: Step 14 end-to-end HTTP validation harness for GStreamer preview
    - Path: ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/12-go-gst-recording-runtime-smoke/main.go
      Note: Step 15 smoke harness for validating real MP4 output
    - Path: ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/13-go-gst-audio-recording-runtime-smoke/main.go
      Note: Step 16 smoke harness for WAV/Opus/mixed audio validation
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

---

## Step 11: Native GStreamer Preview Runtime in `pkg/media/gst`

This step implemented the first real in-repo GStreamer runtime instead of just prototype scripts. The scope was preview only: build the pipeline programmatically, receive JPEG frames via `appsink`, handle bus messages with a GLib main loop, and validate the runtime against real display, region, and camera sources.

I intentionally kept FFmpeg as the default preview runtime in the web layer for now. That lets us land and review the native GStreamer runtime in isolation before switching production preview traffic over to it.

### Prompt Context

**User prompt (verbatim):** (see Step 10)

**Assistant interpretation:** Continue the ticket implementation task-by-task, make another focused commit, and record exactly what worked and what failed.

**Inferred user intent:** Turn the analyzed migration plan into small, validated implementation increments instead of a risky one-shot rewrite.

**Commit (code):** 806c14e630a108ac3dd9670af0eb205c4c1072c9 — "Add native GStreamer preview runtime"

### What I did

- Added Go GStreamer dependencies to the repo module:
  - `github.com/go-gst/go-gst v1.4.0`
  - `github.com/go-gst/go-glib v1.4.0`
- Implemented `pkg/media/gst/preview.go`:
  - `PreviewRuntime.StartPreview(...)`
  - `previewSession.Wait()`
  - `previewSession.Stop(...)`
  - `LatestFrame()` / `TakeScreenshot(...)`
  - one-time `gst.Init(nil)` initialization
  - source mapping for:
    - `display`
    - `region`
    - `window`
    - `camera`
  - preview pipeline construction:
    - source → `videoconvert`
    - optional `videoflip` for mirrored camera preview
    - `videoscale`
    - `capsfilter(video/x-raw,width=640)`
    - `videorate`
    - `capsfilter(video/x-raw,framerate=5/1)`
    - `jpegenc`
    - `appsink`
- Implemented `pkg/media/gst/bus.go`:
  - GLib main-loop-backed bus watch helper
  - watch cleanup / loop shutdown helper
- Implemented `pkg/media/gst/pipeline.go`:
  - capsfilter helper
  - element-link helper
- Added reproducible validation script in the ticket:
  - `scripts/09-go-gst-preview-runtime-smoke/main.go`
  - `scripts/09-go-gst-preview-runtime-smoke.sh`
- Ran dependency hygiene and tests:
  - `go mod tidy`
  - `go test ./... -count=1`
- Ran live smoke tests against the real runtime:
  - display preview ✓
  - region preview ✓
  - camera preview ✓
  - window preview ✗ (X11 BadMatch / MIT-SHM issue)
- Checked ticket tasks:
  - 1.1 preview runtime implementation
  - 1.2 appsink callback wiring
  - 1.3 bus handling
  - 1.4 preview stop semantics

### Why

The architecture seam from Step 10 is only useful if we can drop a real GStreamer implementation behind it. Preview is the safest first vertical slice because it has a small, well-understood contract: take a video source and deliver JPEG frames into Go.

This step proves the migration path is not theoretical anymore. The runtime now exists inside the main codebase and not just in ticket experiments.

### What worked

- `pkg/media/gst/preview.go` compiled and integrated cleanly into the repo
- `go test ./... -count=1` passed after adding CGO-backed GStreamer dependencies
- The smoke test worked for real display capture:
  - command: `bash ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/09-go-gst-preview-runtime-smoke.sh`
  - result: 15 JPEG frames in ~3 seconds
- Region preview worked:
  - command: `REGION=0,0,640,480 bash ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/09-go-gst-preview-runtime-smoke.sh`
  - result: 15 JPEG frames in ~3 seconds
- Camera preview worked:
  - command: `SOURCE_TYPE=camera DEVICE=/dev/video0 bash ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/09-go-gst-preview-runtime-smoke.sh`
  - result: 13 JPEG frames in ~3 seconds
- Context-driven shutdown worked: the preview session exits cleanly when the timeout cancels the context
- Appsink delivery worked exactly as intended: raw JPEG bytes arrive in Go without stdout parsing

### What didn't work

- **Window preview failed** in the current environment.
- Exact validation command:
  - `SOURCE_TYPE=window WINDOW_ID=0x05c0000e bash ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/09-go-gst-preview-runtime-smoke.sh`
- Exact failure:

```text
X Error of failed request:  BadMatch (invalid parameter attributes)
  Major opcode of failed request:  130 (MIT-SHM)
  Minor opcode of failed request:  4 (X_ShmGetImage)
  Serial number of failed request:  59
  Current serial number in output stream:  59
exit status 1
```

- I also reproduced the same problem with raw `gst-launch-1.0`, so this is not just a Go wrapper bug.
- Because of that, I did **not** check task 1.5 or 1.6 yet.

### What I learned

- The in-process GStreamer preview pipeline is straightforward once the go-gst API quirks are understood
- `appsink` is a clean replacement for the old FFmpeg `pipe:1` MJPEG parsing path
- A tiny GLib bus-watch helper is enough for the preview use case; we do not need a giant framework around bus handling
- The display / region / camera paths are now validated in the real repo, not just in one-off experiments
- Window capture on X11 is the first serious preview-runtime blocker that needs deeper investigation

### What was tricky to build

- The most delicate part was orchestrating three lifecycles at once:
  1. the Go context
  2. the GStreamer pipeline state
  3. the GLib main loop used for bus watches
- The symptom to avoid was a preview session that would keep running after context cancellation or leave the main loop/watch hanging. I handled that by making the session goroutine own shutdown sequencing:
  - on cancellation, send EOS
  - briefly wait for bus EOS/error
  - stop the bus watch
  - set pipeline state to NULL
  - close the session `done` channel
- The other sharp edge was not over-wiring the web layer too early. I chose to keep FFmpeg as the default runtime until the GStreamer preview path has broader validation.

### What warrants a second pair of eyes

- The X11 window-capture failure with `ximagesrc xid=...` — especially whether we should:
  - keep pursuing `xid`,
  - capture windows by resolved geometry instead,
  - or treat window capture as an X11-specific compatibility path that may need a fallback
- The `videoflip` mirror property path for cameras should be reviewed once browser-visible preview switching is wired up
- The GLib main-loop approach should be reviewed if we later run many simultaneous preview sessions

### What should be done in the future

- Investigate/fix the window preview path before checking task 1.5 and 1.6
- Add a web/server-level injection point or config switch so the browser preview can be exercised end-to-end with the GStreamer runtime
- Once preview is sufficiently validated, decide whether to make GStreamer the default preview runtime while recording remains FFmpeg-backed

### Code review instructions

- Start with:
  - `pkg/media/gst/preview.go`
- Then review the supporting helpers:
  - `pkg/media/gst/bus.go`
  - `pkg/media/gst/pipeline.go`
- Then review the reproducible smoke-test entrypoint:
  - `ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/09-go-gst-preview-runtime-smoke/main.go`
  - `ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/09-go-gst-preview-runtime-smoke.sh`
- Validate with:
  - `go test ./... -count=1`
  - `bash .../09-go-gst-preview-runtime-smoke.sh`
  - `REGION=0,0,640,480 bash .../09-go-gst-preview-runtime-smoke.sh`
  - `SOURCE_TYPE=camera DEVICE=/dev/video0 bash .../09-go-gst-preview-runtime-smoke.sh`

### Technical details

- New repo dependencies added to the main module:
  - `github.com/go-gst/go-gst v1.4.0`
  - `github.com/go-gst/go-glib v1.4.0`
- Preview runtime design in this step:

```text
StartPreview(ctx, source, opts)
  -> gst.Init(nil) once
  -> build pipeline for source type
  -> appsink callback copies JPEG bytes to Go
  -> bus watch publishes EOS/error into result channel
  -> session goroutine waits for EOS/error or ctx cancellation
  -> pipeline set to NULL on exit
```

- Validated source types in this step:
  - display ✓
  - region ✓
  - camera ✓
  - window ✗ (BadMatch / MIT-SHM failure)

---

## Step 12: Window Preview Investigation Plan

Before touching the GStreamer preview code again, I want a reproducible explanation for why `window` preview currently fails while `display`, `region`, and `camera` succeed. The failure reproduces outside Go, so the next step is to debug the X11/GStreamer behavior directly rather than guessing inside the runtime implementation.

This step stores the plan before execution so the subsequent investigation stays scoped, reviewable, and reproducible.

### Prompt Context

**User prompt (verbatim):** "store the investigation plan, and then do it, keeping diary as you go"

**Assistant interpretation:** Save a concrete plan into the ticket docs first, then execute the investigation incrementally and document what actually happened.

**Inferred user intent:** Avoid ad-hoc debugging; turn the window-preview issue into a traceable investigation with preserved commands, evidence, and next actions.

### What I plan to do

1. **Minimize the failing case**
   - Reproduce the `ximagesrc xid=...` failure using the smallest possible `gst-launch-1.0` pipeline.
   - Vary relevant properties one at a time (`remote`, `show-pointer`, `use-damage`) to determine whether the failure is specific to one property or to `xid` capture itself.

2. **Inspect the target window with X11 tools**
   - Use `xwininfo -id ...` and `xprop -id ...` to capture geometry, mapping state, depth/visual hints, and root/display relationships.
   - Confirm the XID is valid and the window is actually viewable.

3. **Compare XID capture with geometry capture**
   - Resolve the same window to `(x, y, width, height)`.
   - Run `ximagesrc startx/starty/endx/endy` against the same rectangle.
   - If region capture works while `xid` fails, treat `xid` as the broken path and geometry capture as the likely implementation direction.

4. **Test multiple windows**
   - Try at least a terminal/editor window and one browser/app window.
   - Determine whether the issue is generic to all `xid` capture in this environment or specific to certain windows/compositor states.

5. **Compare with FFmpeg**
   - Run a minimal FFmpeg `-window_id` capture for at least one of the same windows.
   - Determine whether the current FFmpeg path really relies on window-ID capture successfully in this environment.

6. **Only then change code**
   - If geometry capture is clearly the reliable path, update the GStreamer preview runtime so `window` sources are implemented via resolved geometry rather than `xid`.
   - Re-run the preview smoke tests and then check task 1.5 if all source types are covered.

### What warrants a second pair of eyes

- Whether the correct long-term model for “window capture” should be “capture by XID” or “resolve geometry and capture that rectangle”
- Whether the environment-specific MIT-SHM behavior means we should avoid `xid` entirely in X11 mode

### What should be done in the future

- Execute the plan immediately after this step and record each result in the next diary entry

### Code review instructions

- For this step, just review that the plan covers the raw pipeline, X11 inspection, FFmpeg comparison, and code-change decision point before implementation.

### Technical details

- Current failing command from Step 11:

```text
SOURCE_TYPE=window WINDOW_ID=0x05c0000e bash .../09-go-gst-preview-runtime-smoke.sh

X Error of failed request:  BadMatch (invalid parameter attributes)
  Major opcode of failed request:  130 (MIT-SHM)
  Minor opcode of failed request:  4 (X_ShmGetImage)
```

---

## Step 13: Window Preview Investigation Results and Geometry Fallback Fix

This step executed the investigation plan from Step 12 and produced a concrete answer: `ximagesrc xid=...` is not universally broken, but it is unreliable for some windows that show up as `IsUnviewable` under `xwininfo`. For those windows, geometry-based capture of the same on-screen rectangle works reliably, and that is the safer implementation strategy for preview.

The outcome was not just a diagnosis. I used the evidence to update the native GStreamer preview runtime so `window` sources are resolved to geometry first and then captured as a region instead of relying on the fragile XID path.

### Prompt Context

**User prompt (verbatim):** (see Step 12)

**Assistant interpretation:** Execute the stored investigation, keep the evidence reproducible, and apply the fix if the results clearly point to one.

**Inferred user intent:** Use evidence to drive the implementation, not guesses.

**Commit (code):** b247f270b07600df025c5652317023b03a6f347d — "Use geometry fallback for GStreamer window preview"

### What I did

- Wrote a reproducible investigation script:
  - `ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/10-window-preview-investigation.sh`
- The script records, per window:
  - `xwininfo` summary
  - `xprop` summary
  - `gst-launch-1.0` XID capture result
  - `gst-launch-1.0` geometry capture result
  - FFmpeg `-window_id` capture result
  - FFmpeg geometry capture result
- Ran the script against two real windows:
  - `0x03c00005`
  - `0x0200026c`
- Ran the existing preview smoke test again with the previously failing window ID
- Implemented a runtime fix in `pkg/media/gst/preview.go`:
  - before building a `window` preview pipeline, resolve its geometry via `xwininfo`
  - populate `source.Target.Rect`
  - capture the resolved rectangle with `ximagesrc startx/starty/endx/endy`
  - stop using `ximagesrc xid=...` in the preview runtime
- Exported `discovery.WindowGeometry(...)` from `pkg/discovery/service.go` so the preview runtime can reuse the same X11 geometry logic the discovery layer already trusts
- Re-ran validation:
  - `go test ./... -count=1`
  - display preview smoke test
  - region preview smoke test
  - camera preview smoke test
  - previously failing window preview smoke test
- Checked task:
  - 1.5 source-type handling

### Why

The investigation showed a more nuanced reality than the original hypothesis. Some window IDs worked directly with `xid`, some did not, and the dividing line was not “all window capture is broken” but “some windows in this environment are not reliably capturable by XID.”

Because the system already has geometry for windows and rectangle capture is more reliable, geometry fallback is the more robust preview implementation.

### What worked

- The investigation script produced a clear contrast between two windows:
  - **Window `0x03c00005`**
    - `Map state: IsUnviewable`
    - GStreamer XID capture: **failed**
    - GStreamer geometry capture: **succeeded**
    - FFmpeg geometry capture: **succeeded**
  - **Window `0x0200026c`**
    - `Map state: IsViewable`
    - GStreamer XID capture: **succeeded**
    - GStreamer geometry capture: **succeeded**
    - FFmpeg XID capture: **succeeded**
- The old failing preview now works after the geometry fallback:
  - command: `SOURCE_TYPE=window WINDOW_ID=0x03c00005 bash .../09-go-gst-preview-runtime-smoke.sh`
  - result: 15 frames captured successfully
- After the fix, all four preview source types were validated with the smoke test:
  - display ✓
  - region ✓
  - camera ✓
  - window ✓
- `go test ./... -count=1` still passed after the code change

### What didn't work

- The raw XID path remains unreliable for some windows in this environment.
- The most important failure reproduced by the investigation was:

```text
X Error of failed request:  BadMatch (invalid parameter attributes)
  Major opcode of failed request:  130 (MIT-SHM)
  Minor opcode of failed request:  4 (X_ShmGetImage)
```

- This was observed on an `IsUnviewable` window and reproduced outside the Go runtime with direct `gst-launch-1.0`.

### What I learned

- The right conclusion was **not** “window preview is broken in GStreamer.”
- The real conclusion is:
  - direct XID capture works for some windows,
  - fails for others,
  - and geometry capture is consistently safer.
- Discovery already relies on `xwininfo`-based window geometry, so reusing that logic in the preview runtime is consistent with the rest of the app rather than a special-case hack.
- The earlier failed window preview test was misleading because it happened to use an especially bad XID for direct capture.

### What was tricky to build

- The subtle part was resisting the urge to overgeneralize after the first failing window.
- After the first experiments, it looked like `xid` capture might be universally broken. The symptom changed once I tested more than one window: one XID failed hard while another succeeded cleanly. That meant the problem was environmental/window-specific, not a blanket GStreamer bug.
- The solution was to move one level up conceptually:
  1. stop asking “why does XID fail sometimes?” as the only question,
  2. ask “what capture strategy is robust across the windows the app discovers?”
  3. use geometry, because the app already has it and it works for both the good and bad cases.

### What warrants a second pair of eyes

- Whether the preview runtime should always use geometry for `window` sources long-term, or whether there is still value in preserving an optional direct-XID path for special cases
- Whether the recording runtime should make the same change when GStreamer recording support lands
- Whether the FFmpeg path should eventually receive the same window-geometry fallback if we want current window recording/preview behavior to become more robust before FFmpeg is removed

### What should be done in the future

- Finish task 1.6 by wiring the GStreamer preview runtime through the web layer in an end-to-end path and validating browser MJPEG streaming plus suspend/restore behavior
- Consider adding an automated integration test that exercises the window-preview path with a fake/Xvfb-style display if practical
- Revisit FFmpeg `window_id` behavior separately if the current FFmpeg runtime also suffers from the same unreliable XID issue in production use

### Code review instructions

- Start with the investigation script:
  - `ttmp/.../scripts/10-window-preview-investigation.sh`
- Then review the exported geometry hook:
  - `pkg/discovery/service.go` (`WindowGeometry`)
- Then review the runtime behavior change:
  - `pkg/media/gst/preview.go`
- Reproduce with:
  - `bash ttmp/.../scripts/10-window-preview-investigation.sh 0x03c00005 0x0200026c`
  - `SOURCE_TYPE=window WINDOW_ID=0x03c00005 bash ttmp/.../scripts/09-go-gst-preview-runtime-smoke.sh`
  - `go test ./... -count=1`

### Technical details

- Investigation evidence from the script:

```text
Window 0x03c00005
  Map state: IsUnviewable
  gst xid capture: failed
  gst geometry capture: succeeded
  ffmpeg geometry capture: succeeded

Window 0x0200026c
  Map state: IsViewable
  gst xid capture: succeeded
  gst geometry capture: succeeded
  ffmpeg xid capture: succeeded
```

- New runtime flow for `window` preview:

```text
StartPreview(window source)
  -> discovery.WindowGeometry(ctx, windowID)
  -> source.Target.Rect = resolved geometry
  -> ximagesrc startx/starty/endx/endy
  -> videoconvert -> videoscale -> videorate -> jpegenc -> appsink
```

- Validation after the fix:
  - display: 15 frames
  - region: 15 frames
  - camera: 15 frames
  - previously failing window: 15 frames

---

## Step 14: HTTP End-to-End Validation of GStreamer Preview and Suspend/Restore

This step finished Phase 1 by proving the native GStreamer preview runtime works through the actual web handlers, not just through direct runtime smoke tests. I added a small server-construction hook so a test harness can ask the web layer to use the GStreamer preview runtime explicitly, then exercised `/api/previews/ensure`, `/api/previews/{id}/mjpeg`, preview suspension during recording start, and restoration after recording finish.

The important distinction is that this step validates the browser-facing path: preview frames flow through `PreviewManager`, the MJPEG endpoint still streams correctly, and the existing preview handoff logic still works when the preview implementation underneath is GStreamer instead of FFmpeg.

### Prompt Context

**User prompt (verbatim):** "continue."

**Assistant interpretation:** Keep moving through the ticket tasks in order and complete the next unfinished preview task with the same commit/diary discipline.

**Inferred user intent:** Finish the remaining preview-phase work before moving on to recording migration.

**Commit (code):** 7db020ad31048d9b2f4f47adc9f04c8b6742ef6b — "Add web-level GStreamer preview validation hook"

### What I did

- Added server-level preview runtime injection in `internal/web/server.go`:
  - `ServerOption`
  - `WithPreviewRuntime(...)`
  - `NewServerWithOptions(...)`
  - `NewServer(...)` now delegates to `NewServerWithOptions(...)`
- Kept existing call sites compatible by preserving `NewServer(...)`
- Wrote a reproducible HTTP-level validation harness:
  - `ttmp/.../scripts/11-web-gst-preview-e2e/main.go`
  - `ttmp/.../scripts/11-web-gst-preview-e2e.sh`
- The harness:
  - starts an `httptest` server using `web.NewServerWithOptions(..., web.WithPreviewRuntime(gst.NewPreviewRuntime()))`
  - serves a real GStreamer preview runtime through the normal web handlers
  - calls `/api/previews/ensure`
  - reads `/api/previews/{id}/mjpeg` until it sees a frame boundary
  - starts a fake recording through `/api/recordings/start`
  - polls `/api/previews` until preview count becomes `0` (suspended)
  - waits for preview count to become `1` again (restored)
  - reads MJPEG again after restore
  - releases the preview
- Re-ran test suite:
  - `go test ./... -count=1`
- Ran the end-to-end harness for:
  - display preview
  - region preview
  - camera preview
  - window preview (extra validation beyond task scope)
- Checked task:
  - 1.6 end-to-end preview validation

### Why

Phase 1 was not really complete until the preview runtime proved itself through the actual HTTP contract the browser uses. Direct runtime smoke tests were necessary, but not sufficient.

This step proves the current web layer can keep its shape while the preview runtime underneath changes. That is one of the central migration goals.

### What worked

- `go test ./... -count=1` stayed green after adding the server option
- Display preview worked end-to-end through HTTP:
  - ensure endpoint succeeded
  - MJPEG endpoint returned `--frame`
  - preview suspended during recording start
  - preview restored after recording finished
  - MJPEG still worked after restore
- Region preview worked through the same HTTP path
- Camera preview worked through the same HTTP path
- Window preview also worked through the same HTTP path after the geometry fallback from Step 13
- The preview handoff logic in `internal/web/server.go` did not need semantic changes; the GStreamer runtime fit under the existing `PreviewManager`

### What didn't work

- Nothing new failed in this step.
- The only noteworthy caveat was that camera MJPEG reads after restore could take longer than display/region because frame delivery timing is hardware-dependent. The request still succeeded and returned a frame boundary.

### What I learned

- The existing HTTP preview contract is decoupled enough from FFmpeg that a runtime swap is feasible without redesigning the browser API
- The preview suspend/restore lifecycle belongs at the manager/server layer, not inside FFmpeg/GStreamer-specific code
- Adding an explicit server option is much cleaner than manually replacing `server.previews` in every external harness
- The web path now has practical proof for:
  - ensure
  - MJPEG streaming
  - suspend
  - restore
  - release

### What was tricky to build

- The main challenge was getting true end-to-end coverage without changing production defaults prematurely.
- If I had simply switched `NewServer(...)` to use the GStreamer runtime by default, that would have mixed validation work with rollout policy. The symptom would have been a harder-to-review change where every server construction path silently changed behavior at once.
- The solution was to add a narrow constructor seam:
  1. preserve `NewServer(...)` behavior
  2. add `NewServerWithOptions(...)`
  3. inject `WithPreviewRuntime(gst.NewPreviewRuntime())` only in the validation harness

### What warrants a second pair of eyes

- Whether `WithPreviewRuntime(...)` should remain a test/integration-only hook for now or become part of the normal rollout strategy when we switch defaults
- Whether we want a small CLI or env toggle for serving with GStreamer preview in manual developer testing
- Whether the preview ID stability assumption across suspend/restore is something we want to preserve explicitly, or whether the browser should always treat restored previews as potentially new identities

### What should be done in the future

- Decide whether to switch the web server’s preview runtime default from FFmpeg to GStreamer now that Phase 1 is fully validated
- Start Phase 2: native GStreamer recording runtime
- Reuse the same server-option pattern if a staged rollout is needed for recording runtime injection later

### Code review instructions

- Start with the server constructor seam:
  - `internal/web/server.go`
- Then review the end-to-end harness:
  - `ttmp/.../scripts/11-web-gst-preview-e2e/main.go`
  - `ttmp/.../scripts/11-web-gst-preview-e2e.sh`
- Validate with:
  - `go test ./... -count=1`
  - `bash ttmp/.../scripts/11-web-gst-preview-e2e.sh`
  - `REGION=0,0,640,480 bash ttmp/.../scripts/11-web-gst-preview-e2e.sh`
  - `SOURCE_TYPE=camera DEVICE=/dev/video0 bash ttmp/.../scripts/11-web-gst-preview-e2e.sh`
  - optional: `SOURCE_TYPE=window WINDOW_ID=0x03c00005 bash ttmp/.../scripts/11-web-gst-preview-e2e.sh`

### Technical details

- Server API added in this step:

```go
type ServerOption func(*serverOptions)

func WithPreviewRuntime(runtime media.PreviewRuntime) ServerOption
func NewServerWithOptions(parentCtx context.Context, application ApplicationService, cfg Config, opts ...ServerOption) *Server
```

- End-to-end flow validated in this step:

```text
HTTP POST /api/previews/ensure
  -> PreviewManager.Ensure
  -> GStreamer PreviewRuntime.StartPreview
  -> appsink frames stored in manager

HTTP GET /api/previews/{id}/mjpeg
  -> reads manager.LatestFrame()
  -> multipart/x-mixed-replace stream

HTTP POST /api/recordings/start
  -> server suspends previews
  -> fake recording runs
  -> server restores suspended previews on finish
```

- Source types validated through the HTTP path:
  - display ✓
  - region ✓
  - camera ✓
  - window ✓ (extra)

---

## Step 15: Native GStreamer Video Recording Runtime (Phase 2.1)

This step started the recording side of the migration by implementing the first native GStreamer recording runtime slice: **video jobs only**. I kept audio explicitly out of scope for this step so the code and validation could focus on one problem at a time: build a real pipeline, write a real MP4, stop cleanly with EOS, and surface results through the `media.RecordingRuntime` seam already created in Phase 0.

The key success criterion for this step was not just “pipeline starts.” It was “pipeline starts, records, stops on context cancellation, finalizes a valid MP4, and returns a coherent recording result.”

### Prompt Context

**User prompt (verbatim):** (see Step 14)

**Assistant interpretation:** Keep progressing through the ticket tasks after preview completion and move into the next recording-runtime task in the same disciplined way.

**Inferred user intent:** Finish the migration incrementally, with each runtime slice validated before moving deeper.

**Commit (code):** bc6e63e584291432ce857b7137053ed8576213fb — "Add native GStreamer video recording runtime"

### What I did

- Implemented `pkg/media/gst/recording.go`:
  - `RecordingRuntime.StartRecording(...)`
  - `recordingSession.Wait()` / `Stop(...)`
  - per-job recording workers with:
    - GStreamer pipeline creation
    - bus watch for `MessageError` / `MessageEOS`
    - EOS-driven stop handling on context cancellation
  - session-level event emission through `media.RecordingEvent`
- Implemented video pipeline construction for `dsl.VideoJob`:
  - source elements from the same source-mapping logic used by preview
  - window sources resolved to geometry first
  - `videoconvert`
  - optional `videoflip` for mirrored camera recording
  - `videorate`
  - caps filter for target FPS
  - `x264enc`
  - `mp4mux` / `qtmux`
  - `filesink`
- Added a focused smoke-test harness:
  - `ttmp/.../scripts/12-go-gst-recording-runtime-smoke/main.go`
  - `ttmp/.../scripts/12-go-gst-recording-runtime-smoke.sh`
- Validated the runtime with real recordings for:
  - display
  - region
  - camera
- Used `file` and `ffprobe` to verify the produced MP4 files
- Re-ran the repo test suite:
  - `go test ./... -count=1`
- Checked task:
  - 2.1 video recording runtime implementation

### Why

Recording is a riskier migration surface than preview because the output must be finalized correctly. A runtime that “mostly records” but leaves corrupt or empty files is not useful.

By isolating this step to video jobs only, I could validate the muxing and EOS path without mixing in audio-mixer complexity yet.

### What worked

- The runtime now records valid MP4 files directly via GStreamer inside the main repo
- Display recording smoke test succeeded:
  - output file: valid MP4
  - `ffprobe` duration: `3.000000`
  - non-zero file size
- Region recording smoke test succeeded:
  - output file: valid MP4
  - `ffprobe` duration: `3.000000`
- Camera recording smoke test succeeded:
  - output file: valid MP4
  - `ffprobe` duration: `2.900000`
- Window-geometry reuse worked naturally for recording because the same source-resolution logic is shared in the GStreamer package
- `go test ./... -count=1` passed after landing the runtime and harness

### What didn't work

- The **first** display-recording attempt failed during shutdown.
- Exact symptom from the first run:

```text
event: type=state_changed state=failed reason="Recording Source failed: timed out waiting for recording EOS"
wait recording: Recording Source failed: timed out waiting for recording EOS
```

- The pipeline started fine, but the initial encoder/stop configuration did not finalize quickly enough when the context timed out.
- I fixed this by:
  - increasing the EOS wait budget from 3s to 10s
  - setting lower-latency x264 options:
    - `tune=zerolatency`
    - `speed-preset=veryfast`
    - `bframes=0`

After those changes, the recording finalized cleanly and produced valid MP4 files.

### What I learned

- The recording-runtime correctness hinges on **EOS finalization behavior**, not just pipeline startup
- Preview-style cancellation timing is too optimistic for muxed recording outputs; recording needs a larger shutdown budget
- x264 latency settings materially affect whether the pipeline drains/finalizes quickly enough for a clean stop
- A real smoke harness with `ffprobe` is essential here; file existence alone would have been too weak a validation signal

### What was tricky to build

- The sharpest edge was shutdown ownership.
- In the first draft, there was a concurrency risk where the session-level cancellation path and the worker-level stop path could both try to consume the same worker result channel. That would have made failures intermittent and very hard to reason about.
- I fixed that by making the worker’s `wait(ctx)` own the EOS-on-cancel behavior. The session loop observes worker results and context cancellation, but it no longer races the worker by trying to stop pipelines separately.
- The second tricky part was encoder latency: the pipeline could start, run, and still fail the step if it didn’t flush quickly enough on EOS.

### What warrants a second pair of eyes

- Whether the current session result semantics should distinguish user-requested stop from generic `context deadline exceeded` more explicitly
- Whether the current bitrate heuristic from `quality` is the right long-term mapping for x264 in this app
- Whether we want additional container support beyond `mp4` / `mov` before wiring this runtime into higher-level app code

### What should be done in the future

- Implement Phase 2.2: native GStreamer audio recording/mixing runtime
- Then implement Phase 2.3 / 2.4 so the recording stop/event model matches the existing web session manager expectations more closely
- Add end-to-end app/web validation for recording after audio is present

### Code review instructions

- Start with:
  - `pkg/media/gst/recording.go`
- Then review the smoke harness:
  - `ttmp/.../scripts/12-go-gst-recording-runtime-smoke/main.go`
  - `ttmp/.../scripts/12-go-gst-recording-runtime-smoke.sh`
- Validate with:
  - `go test ./... -count=1`
  - `OUT_PATH=/tmp/scs-gst-recording-display.mp4 bash ttmp/.../scripts/12-go-gst-recording-runtime-smoke.sh`
  - `OUT_PATH=/tmp/scs-gst-recording-region.mp4 REGION=0,0,640,480 bash ttmp/.../scripts/12-go-gst-recording-runtime-smoke.sh`
  - `OUT_PATH=/tmp/scs-gst-recording-camera.mp4 SOURCE_TYPE=camera DEVICE=/dev/video0 bash ttmp/.../scripts/12-go-gst-recording-runtime-smoke.sh`

### Technical details

- Video pipeline shape in this step:

```text
[source]
  -> videoconvert
  -> optional videoflip (camera mirror)
  -> videorate
  -> capsfilter(video/x-raw,framerate=N/1)
  -> x264enc(bitrate, bframes=0, tune=zerolatency, speed-preset=veryfast)
  -> mp4mux / qtmux
  -> filesink
```

- Current limitation in this step:
  - `plan.AudioJobs` are rejected with `gstreamer audio recording is not implemented yet`
- Validated outputs from this step:
  - display MP4 ✓
  - region MP4 ✓
  - camera MP4 ✓

---

## Step 16: Native GStreamer Audio Recording and Mixing Runtime (Phase 2.2)

This step completed the second big recording-runtime slice by adding native GStreamer support for `AudioJobs`. The goal was to match the current app’s audio model closely enough to be useful: one or more PulseAudio/PipeWire sources, per-source gain, mixing through `audiomixer`, and file output in at least the formats already validated during earlier experiments.

I kept the validation practical rather than theoretical. The output had to be real, non-empty, and recognizable by `file`/`ffprobe`, and the multi-input graph had to work with real request-pad linking instead of a hard-coded single-source shortcut.

### Prompt Context

**User prompt (verbatim):** (see Step 15)

**Assistant interpretation:** Continue with the next planned recording-runtime task and validate it the same way as the earlier runtime slices.

**Inferred user intent:** Keep progressing through the migration checklist, one tested subsystem at a time.

**Commit (code):** ec7136d5168f7e911ef209e369110157225e5e52 — "Add GStreamer audio mixing runtime"

### What I did

- Extended `pkg/media/gst/recording.go` to support `plan.AudioJobs`
- Added native audio pipeline construction with:
  - one `pulsesrc` per source
  - capsfilter for shared raw-audio format
  - `audioconvert`
  - `audioresample`
  - per-source `volume`
  - `audiomixer`
  - post-mix `audioconvert` / `audioresample`
  - encoder/mux branch:
    - `wavenc -> filesink` for WAV/PCM
    - `opusenc -> oggmux -> filesink` for Opus/Ogg
- Used request pads on `audiomixer` (`sink_%u`) and linked source-branch `volume` src pads directly into the mixer
- Added audio caps helpers so raw/mixed audio formats are normalized before encoding
- Added a dedicated runtime smoke harness:
  - `ttmp/.../scripts/13-go-gst-audio-recording-runtime-smoke/main.go`
  - `ttmp/.../scripts/13-go-gst-audio-recording-runtime-smoke.sh`
- Validated three concrete cases:
  1. single-source WAV
  2. single-source Opus/Ogg
  3. two-branch mixed WAV (using the same `default` Pulse source twice with different gains to validate the graph shape)
- Re-ran the full Go test suite:
  - `go test ./... -count=1`
- Checked task:
  - 2.2 audio recording/mixing runtime

### Why

The video runtime from Step 15 was only half the recording story. The current app’s recording plans can also contain audio mix jobs, and the migration isn’t credible until both sides exist natively under the new runtime seam.

This step aimed for the minimal useful implementation: real mix graph, real outputs, real encoder branches, and real runtime supervision.

### What worked

- Single-source WAV recording worked:
  - output recognized as `RIFF ... WAVE audio, Microsoft PCM, 16 bit, stereo 48000 Hz`
  - `ffprobe` duration: `2.970000`
- Single-source Opus/Ogg recording worked:
  - output recognized as `Ogg data, Opus audio`
  - `ffprobe` duration: `3.016500`
- Two-branch mixed WAV recording worked:
  - same `default` source wired twice with gains `1.0` and `0.5`
  - output recognized as valid WAV
  - confirms the `audiomixer` request-pad graph is functioning, not just the single-source happy path
- The runtime session model from Step 15 handled audio workers cleanly without needing a separate orchestration model
- `go test ./... -count=1` still passed after the audio extension landed

### What didn't work

- Nothing failed at runtime once the graph was wired up.
- The main caveat is validation realism for the mixed case: I mixed the same Pulse `default` device twice rather than two genuinely different microphones/sources, because this machine setup makes that the fastest way to verify the mixer graph without inventing extra infrastructure.

### What I learned

- The go-gst request-pad APIs are workable enough for `audiomixer`:
  - `GetRequestPad("sink_%u")`
  - `GetStaticPad("src")`
  - `Pad.Link(...)`
- The runtime seam from Phase 0 is holding up well: the same worker/session model now supervises both video and audio pipelines
- Opus needs a 48kHz raw-audio encoder input cap, so it was worth making encoder caps explicit instead of assuming the pre-mix raw caps could be reused blindly
- Multi-input audio validation does not require a special fake source to prove the pipeline shape; duplicating the same live device is enough to verify the graph mechanics

### What was tricky to build

- The trickiest part was deciding where to normalize audio caps.
- There are two distinct needs:
  1. make individual source branches mixer-compatible
  2. make post-mix audio encoder-compatible
- If I reused a single caps choice everywhere, Opus would have been awkward because its encoder wants 48kHz S16LE input regardless of the general output defaults. The solution was to split the concerns into two helpers:
  - `audioRawCaps(...)` for branch/mixer compatibility
  - `audioEncoderCaps(...)` for codec-specific post-mix expectations
- The other subtle part was avoiding overclaiming the mixed-source test. The graph is validated, but the “two different live sources” scenario still deserves real-world testing later.

### What warrants a second pair of eyes

- Whether the current codec support matrix for the GStreamer audio runtime should stay intentionally smaller than FFmpeg for now (`wav` / `opus`) or be expanded immediately
- Whether `audiomixer` + per-source `volume` is enough for parity right now, or whether we should bring in optional normalization/latency tuning sooner
- Whether the session result reason should be more semantic than `context deadline exceeded` when a bounded smoke test intentionally stops the runtime via timeout

### What should be done in the future

- Implement Phase 2.3: recording stop semantics cleanup / graceful EOS handling review across both video and audio workers
- Implement Phase 2.4: align runtime-emitted recording states/events even more closely with the existing web session manager expectations
- Add an end-to-end recording validation harness that exercises mixed video+audio plans through the app/web layer

### Code review instructions

- Start with:
  - `pkg/media/gst/recording.go`
- Then review the audio-specific harness:
  - `ttmp/.../scripts/13-go-gst-audio-recording-runtime-smoke/main.go`
  - `ttmp/.../scripts/13-go-gst-audio-recording-runtime-smoke.sh`
- Validate with:
  - `go test ./... -count=1`
  - `OUT_PATH=/tmp/scs-gst-audio-runtime.wav bash ttmp/.../scripts/13-go-gst-audio-recording-runtime-smoke.sh`
  - `CODEC=opus OUT_PATH=/tmp/scs-gst-audio-runtime.ogg bash ttmp/.../scripts/13-go-gst-audio-recording-runtime-smoke.sh`
  - `OUT_PATH=/tmp/scs-gst-audio-runtime-mix.wav DEVICES=default,default GAINS=1.0,0.5 bash ttmp/.../scripts/13-go-gst-audio-recording-runtime-smoke.sh`

### Technical details

- Audio graph in this step:

```text
for each source:
  pulsesrc(device=...)
    -> capsfilter(audio/x-raw,format=S16LE,rate=...,channels=...)
    -> audioconvert
    -> audioresample
    -> volume(volume=gain)
    -> audiomixer request pad

mixer output:
  audioconvert
    -> audioresample
    -> capsfilter(codec-specific caps)
    -> (wavenc -> filesink) OR (opusenc -> oggmux -> filesink)
```

- Validated outputs:
  - WAV ✓
  - Opus/Ogg ✓
  - Multi-branch mix graph ✓
- Current supported GStreamer audio codecs in this runtime slice:
  - `wav` / `pcm_s16le`
  - `opus`
