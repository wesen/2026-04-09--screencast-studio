---
Title: GStreamer Migration — Full Analysis and Intern Guide
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
    - Path: pkg/recording/ffmpeg.go
      Note: Current FFmpeg argument builders for preview, video recording, audio mixing
    - Path: pkg/recording/run.go
      Note: Current subprocess supervisor with stdin-based stop semantics
    - Path: internal/web/preview_runner.go
      Note: Current FFmpeg preview runner with MJPEG stdout parsing
    - Path: internal/web/preview_manager.go
      Note: Preview lifecycle manager (leases, suspend/restore) — to be preserved
    - Path: internal/web/session_manager.go
      Note: Recording session state — to be preserved
    - Path: pkg/dsl/types.go
      Note: DSL type model — stable, media-backend-agnostic
    - Path: pkg/dsl/compile.go
      Note: Plan compilation — stable, media-backend-agnostic
Summary: Comprehensive analysis of the screencast-studio FFmpeg architecture, validated GStreamer replacement pipelines (tested via both gst-launch-1.0 and native go-gst bindings), and detailed guidance for an intern new to GStreamer and audio/video work. Covers screenshots, live audio transcription, live preview with effects, and the complete migration path.
LastUpdated: 2026-04-13T14:30:00-04:00
WhatFor: Onboard an intern to the GStreamer migration, explain every concept they need, and provide validated pipeline designs for preview, recording, screenshots, audio effects, and transcription.
WhenToUse: Read this before writing any GStreamer code for screencast-studio. Refer back when implementing specific pipeline types.
---


# GStreamer Migration — Full Analysis and Intern Guide

## 1. Introduction: What Is This Document?

You are an intern tasked with migrating a screen recording application called **screencast-studio** from FFmpeg subprocess orchestration to native GStreamer pipeline management. This document explains everything you need to know.

**What screencast-studio does today:**

- Captures video from X11 displays, screen regions, windows, and webcams
- Captures audio from PulseAudio devices (with gain/volume control)
- Mixes multiple audio sources together
- Records video and audio to files (MP4, MOV, WAV, OGG, etc.)
- Shows live MJPEG preview of video sources in a browser UI
- Provides a web UI for configuring sources and controlling recording

**What we want it to do with GStreamer:**

- All of the above, plus:
- Take screenshots during recording or preview
- Apply live audio effects (noise gate, denoise, compressor) while monitoring the result
- Support live audio transcription (speech-to-text) during recording
- Share a single capture source between preview and recording (no more device contention)
- All with a more robust, programmable media engine

**Key insight:** The DSL (Domain-Specific Language), the web UI, and the HTTP/WebSocket API are all **stable and working**. We are only replacing the media runtime layer underneath.

---

## 2. GStreamer Basics for the Intern

### 2.1 What Is GStreamer?

GStreamer is a **pipeline-based multimedia framework**. You build graphs of elements connected by pads, and data flows through them. Think of it like plumbing: water comes in from a source (faucet), flows through filters and valves, and comes out at a sink (drain).

The key concepts, in order of importance:

### 2.2 Elements

An **element** is a single processing unit. Examples:

| Element | What it does | Analogy |
|---------|-------------|---------|
| `ximagesrc` | Captures X11 screen | Camera |
| `pulsesrc` | Captures audio from PulseAudio | Microphone |
| `videoconvert` | Converts between video color formats | Color filter |
| `videoscale` | Resizes video frames | Magnifying glass |
| `videorate` | Drops/duplicates frames to hit target FPS | Metronome |
| `audioconvert` | Converts between audio formats | Audio adapter |
| `volume` | Adjusts audio volume/gain | Volume knob |
| `jpegenc` | Encodes raw video to JPEG | Photo printer |
| `x264enc` | Encodes raw video to H.264 | Video compressor |
| `wavenc` | Wraps raw audio in WAV container | Audio file writer |
| `appsink` | Hands raw data to your application code | Bucket you hold |
| `filesink` | Writes data to a file | File writer |

### 2.3 Pads

Every element has **pads** — the connection points. Think of them as the input and output ports.

- **Source pads (src):** Output of an element (where data comes out)
- **Sink pads (sink):** Input of an element (where data goes in)

You link elements by connecting a source pad to a sink pad:

```
[ximagesrc] --src--> --sink--> [videoconvert] --src--> --sink--> [jpegenc]
```

### 2.4 Caps (Capabilities)

**Caps** describe what format of data a pad can accept or produce. For example:

- `video/x-raw, width=640, framerate=5/1` — raw video, 640 pixels wide, 5 frames per second
- `audio/x-raw, rate=48000, channels=2` — raw audio, 48kHz, stereo
- `image/jpeg` — JPEG encoded image data

When you link two elements, GStreamer automatically **negotiates caps** — it figures out a format that both sides agree on. You can force specific caps using `capsfilter` elements.

### 2.5 Pipelines

A **pipeline** is a top-level container that holds all your elements. It manages the overall state:

- `NULL` — not doing anything
- `READY` — initialized, ready to go
- `PAUSED` — ready to process data but not flowing yet
- `PLAYING` — data is flowing

You typically go directly from `NULL` to `PLAYING`.

### 2.6 The Bus

Every pipeline has a **bus** — a message queue that the pipeline uses to send you events:

- `ERROR` — something broke
- `EOS` (End Of Stream) — the data has finished
- `STATE_CHANGED` — the pipeline changed state
- `ELEMENT` — custom element messages (e.g., a screenshot is ready)

You watch the bus to know what's happening.

### 2.7 Appsink and Appsrc

These are the **bridge elements** between GStreamer and your application code:

- **appsink**: GStreamer pushes data *to you*. You set a callback, GStreamer calls it with each buffer of data. This is how we get JPEG preview frames into Go.
- **appsrc**: *You push data* to GStreamer. Not needed for screencast-studio right now, but useful for things like injecting text overlays.

### 2.8 Tee

A **tee** element splits one stream into multiple copies. This is how we'll share a single capture source between preview and recording:

```
[ximagesrc] → [tee] → branch 1 → [preview pipeline]
                    → branch 2 → [recording pipeline]
```

### 2.9 EOS (End Of Stream)

EOS is a special event that means "no more data." When you send EOS to a recording pipeline, the encoder flushes its buffers, the muxer writes its headers, and the file is properly finalized. **Always send EOS before stopping a recording pipeline** — otherwise your output files may be truncated or corrupt.


---

## 3. Current Architecture: How FFmpeg Is Used Today

### 3.1 The Big Picture

The current architecture has two layers:

```
┌─────────────────────────────────────────────────────────┐
│  Web UI (React)                                         │
│  - Configure sources via DSL editor                      │
│  - See live previews (MJPEG <img> tags)                  │
│  - Start/stop recording                                  │
│  - View logs and output files                            │
└──────────────────────┬──────────────────────────────────┘
                       │ HTTP + WebSocket
                       ▼
┌─────────────────────────────────────────────────────────┐
│  Go Web Server                                          │
│  - Parse/normalize/compile DSL                           │
│  - Manage preview lifecycle (PreviewManager)             │
│  - Manage recording lifecycle (SessionManager)           │
│  - Stream MJPEG frames to browser                       │
│  - Publish events via WebSocket                          │
└──────────────────────┬──────────────────────────────────┘
                       │ exec.Command("ffmpeg", ...)
                       ▼
┌─────────────────────────────────────────────────────────┐
│  FFmpeg Subprocesses (one per output)                    │
│  - Separate process for each video source                │
│  - Separate process for audio mix                        │
│  - Separate process for each preview stream              │
│  - Communicated via stdin (q to stop), stdout, stderr    │
└─────────────────────────────────────────────────────────┘
```

### 3.2 The Files You Need to Understand

Read these in order:

| # | File | What It Does | Migration Impact |
|---|------|-------------|-----------------|
| 1 | `pkg/dsl/types.go` | Defines the DSL data model (sources, outputs, jobs) | **No change** — media-agnostic |
| 2 | `pkg/dsl/compile.go` | Compiles DSL → `CompiledPlan` with `VideoJob` and `AudioMixJob` | **No change** — media-agnostic |
| 3 | `pkg/recording/ffmpeg.go` | Builds FFmpeg command lines from jobs | **Replace entirely** |
| 4 | `pkg/recording/run.go` | Supervises multiple FFmpeg subprocesses | **Replace** with GStreamer pipeline management |
| 5 | `internal/web/preview_runner.go` | Runs FFmpeg for preview, parses MJPEG from stdout | **Replace** with GStreamer appsink-based preview |
| 6 | `internal/web/preview_manager.go` | Manages preview lifecycle (leases, caching, suspend/restore) | **Preserve** — media-agnostic |
| 7 | `internal/web/session_manager.go` | Manages recording session state | **Preserve** — media-agnostic |
| 8 | `internal/web/handlers_preview.go` | HTTP handlers for preview (MJPEG streaming) | **Mostly preserve** — may add screenshot endpoint |
| 9 | `internal/web/server.go` | Server wiring and startup | **Modify** — inject GStreamer runtime |

### 3.3 How Preview Works Today (FFmpeg)

```
1. Browser sends: POST /api/previews/ensure {dsl, sourceId}
2. Server normalizes DSL, finds the video source
3. Server checks if a preview already exists (by source signature)
4. If new:
   a. Build FFmpeg args:
      ffmpeg -hide_banner -loglevel error -nostdin \
        -f x11grab -framerate 5 -draw_mouse 1 -i :0.0 \
        -an -vf "fps=5,scale=640:-1" -q:v 7 \
        -f image2pipe -vcodec mjpeg pipe:1
   b. Start FFmpeg subprocess
   c. Read JPEG frames from stdout (parse MJPEG stream byte-by-byte)
   d. Store latest frame in memory
   e. Publish state change events via WebSocket
5. Browser polls: GET /api/previews/{id}/mjpeg
   → Server streams stored frames as multipart/x-mixed-replace
```

**Key problems with this approach:**

- Each preview is a separate OS process
- Preview and recording cannot share the same capture source
- MJPEG parsing is fragile (byte-scanning for FF D8 / FF D9 markers)
- Stop requires writing `q` + newline to stdin and hoping FFmpeg cooperates

### 3.4 How Recording Works Today (FFmpeg)

```
1. Browser sends: POST /api/recordings/start {dsl}
2. Server compiles DSL → CompiledPlan (VideoJobs + AudioMixJobs)
3. For each VideoJob:
   a. Build FFmpeg args (x11grab or v4l2 input → encoder → file output)
   b. Start ManagedProcess wrapping exec.Command("ffmpeg", args...)
4. For each AudioMixJob:
   a. Build FFmpeg args (multiple pulse inputs → volume filters → amix → encoder → file)
   b. Start ManagedProcess
5. Session supervisor watches all processes:
   - If any process exits unexpectedly → fail the session
   - On cancel → write `q` + newline to each process stdin
   - On hard timeout → kill processes
6. When all processes exit → transition session to finished/failed
```

**Key problems with this approach:**

- One process per output = duplicate capture work
- Stop depends on FFmpeg's stdin protocol (not a real API)
- No way to dynamically adjust audio effects during recording
- No way to take a screenshot mid-recording
- No way to tap into the audio stream for live transcription


---

## 4. Validated GStreamer Replacement Pipelines

Every pipeline in this section has been **tested and verified** on this machine. The experiments are stored in the ticket's `scripts/` directory so you can re-run them.

### 4.1 Environment

| Component | Version |
|-----------|---------|
| GStreamer | 1.24.2 |
| Go | 1.25.5 |
| go-gst bindings | v1.4.0 |
| OS | Ubuntu 24.04 (X11) |
| Audio server | PipeWire (PulseAudio compatible) |
| Available elements | 1,678 |

### 4.2 Video Preview Pipeline (VALIDATED ✓)

**Current FFmpeg:**
```bash
ffmpeg -f x11grab -framerate 5 -draw_mouse 1 -i :0.0 \
  -an -vf "fps=5,scale=640:-1" -q:v 7 \
  -f image2pipe -vcodec mjpeg pipe:1
```

**GStreamer replacement (gst-launch):**
```bash
gst-launch-1.0 -v \
  ximagesrc startx=0 starty=0 endx=639 endy=479 use-damage=false \
  ! videoconvert \
  ! videoscale \
  ! "video/x-raw,width=640" \
  ! videorate \
  ! "video/x-raw,framerate=5/1" \
  ! jpegenc quality=50 \
  ! multifilesink location=/tmp/frame-%05d.jpg
```

**Result:** ✓ 17 JPEG frames in 3 seconds (~75KB each)

**GStreamer replacement (go-gst native):**
```go
// Build the pipeline programmatically
ximagesrc  → videoconvert → videoscale → capsfilter(width=640)
→ videorate → capsfilter(fps=5/1) → jpegenc(quality=50) → appsink
```

**Result:** ✓ 24 frames in 5 seconds at 5fps (~267KB each)

Experiment: `scripts/06-go-gst-preview-pipeline/`

**Key difference:** The go-gst version captures the full 640x480 region and produces larger JPEGs because `quality=50` in GStreamer maps differently than `-q:v 7` in FFmpeg. We can tune the quality parameter down.

### 4.3 Video Recording Pipeline (VALIDATED ✓)

**Current FFmpeg:**
```bash
ffmpeg -f x11grab -framerate 24 -draw_mouse 1 -i :0.0 \
  -c:v libx264 -preset veryfast -crf 17 -pix_fmt yuv420p \
  output.mov
```

**GStreamer replacement:**
```bash
gst-launch-1.0 -e -v \
  ximagesrc startx=0 starty=0 endx=639 endy=479 use-damage=false \
  ! videoconvert \
  ! videoscale \
  ! "video/x-raw,width=640,framerate=10/1" \
  ! x264enc tune=zerolatency bitrate=1000 speed-preset=veryfast \
  ! mp4mux \
  ! filesink location=output.mp4
```

**Result:** ✓ Valid MP4 file (68KB for 3 seconds)

**Important:** Always use the `-e` flag with `gst-launch-1.0` to ensure EOS is sent on Ctrl+C. In Go, send `gst.NewEOSEvent()` before stopping.

### 4.4 Screenshot Pipeline (VALIDATED ✓)

This is a **new capability** — not possible with the current FFmpeg architecture without starting a separate process.

**GStreamer screenshot (single frame capture):**
```bash
gst-launch-1.0 -v \
  ximagesrc startx=0 starty=0 endx=639 endy=479 num-buffers=1 \
  ! videoconvert \
  ! pngenc \
  ! filesink location=screenshot.png
```

**Result:** ✓ Valid PNG (46KB, 640x480)

The magic is `num-buffers=1` on ximagesrc — it captures exactly one frame, then the pipeline finishes with EOS automatically.

**In go-gst, you would:**
1. Build a pipeline with `pngenc` or `jpegenc`
2. Set `num-buffers=1` on the source
3. Set to PLAYING
4. Wait for EOS
5. Done — the file is written

### 4.5 Audio Capture Pipeline (VALIDATED ✓)

**Current FFmpeg:**
```bash
ffmpeg -f pulse -sample_rate 48000 -channels 2 -i default \
  -filter_complex "[0:a]volume=1.0[a0];[a0]anull[aout]" \
  -map "[aout]" -ar 48000 -ac 2 -c:a pcm_s16le output.wav
```

**GStreamer replacement (gst-launch):**
```bash
gst-launch-1.0 -e -v \
  pulsesrc device=default \
  ! "audio/x-raw,rate=48000,channels=2" \
  ! audioconvert \
  ! wavenc \
  ! filesink location=output.wav
```

**Result:** ✓ Valid WAV (739KB for 3 seconds)

**GStreamer replacement (go-gst native):**
```go
pulsesrc(device=default) → capsfilter(48kHz/stereo) → audioconvert
→ volume(1.0) → wavenc → filesink
```

**Result:** ✓ Valid WAV (577KB for 3 seconds)

Experiment: `scripts/07-go-gst-audio-capture/`

### 4.6 Audio with Opus Encoding (VALIDATED ✓)

```bash
gst-launch-1.0 -e -v \
  pulsesrc device=default \
  ! "audio/x-raw,rate=48000,channels=2" \
  ! audioconvert \
  ! audioresample \
  ! opusenc bitrate=160000 \
  ! oggmux \
  ! filesink location=output.ogg
```

**Result:** ✓ Valid Ogg/Opus file (66KB for 3 seconds)

### 4.7 Audio with Gain Adjustment (VALIDATED ✓)

```bash
gst-launch-1.0 -e -v \
  pulsesrc device=default \
  ! "audio/x-raw,rate=48000,channels=2" \
  ! audioconvert \
  ! volume volume=1.5 \
  ! wavenc \
  ! filesink location=output.wav
```

**Result:** ✓ Valid WAV with amplified audio

The `volume` element in GStreamer directly maps to FFmpeg's `volume=X.Y` audio filter.


---

## 5. New Capabilities Enabled by GStreamer

These are things that are difficult or impossible with the current FFmpeg subprocess architecture, but become natural with GStreamer's programmable pipeline model.

### 5.1 Screenshots During Recording or Preview

**Why it's hard with FFmpeg:** Each FFmpeg process captures one output. To take a screenshot, you'd need to either:
- Start a separate FFmpeg process (wasteful, slow)
- Parse the MJPEG preview stream and save a frame (lossy, low-res)

**Why it's easy with GStreamer:** Use a **tee** to branch the capture source:

```
[ximagesrc] → [tee]
    → branch 1: [videoconvert → videoscale → jpegenc → appsink]     ← preview
    → branch 2: [videoconvert → x264enc → mp4mux → filesink]        ← recording
    → branch 3: [videoconvert → pngenc → filesink]                   ← screenshot on demand
```

For on-demand screenshots, you can:

1. **Add a temporary branch** to the tee when a screenshot is requested, capture one buffer, then remove the branch.
2. **Use `gdkpixbufsink`** which posts a pixbuf message on the bus — grab it, save it.
3. **Pull from the appsink** and save the latest JPEG frame directly in Go.

The simplest approach for V1: option 3 — the preview appsink already has JPEG frames in memory. Just take the latest frame and serve it as a screenshot endpoint.

### 5.2 Live Audio Effects During Recording

**The user's request:** "live preview while recording of the audio to change say, audio effects"

**How it works in GStreamer:**

The audio pipeline in GStreamer is just a chain of elements. You can insert effect elements at any point:

```
[pulsesrc] → [audioconvert] → [audioresample]
  → [volume gain=1.5]      ← gain adjustment
  → [audiodynamic]          ← compressor/limiter
  → [audiochebband]         ← noise gate (band-pass filter)
  → [tee]
      → branch 1: [wavenc → filesink]                    ← recording
      → branch 2: [appsink] → Go callback → waveform UI  ← live monitoring
```

**Key audio effect elements available:**

| Element | What it does | Use case |
|---------|-------------|----------|
| `volume` | Gain adjustment | Boost quiet mic |
| `audiodynamic` | Compressor/expander/limiter | Prevent clipping |
| `audioecho` | Echo/delay effect | Post-production |
| `equalizer-nbands` | Multi-band EQ | Tone adjustment |
| `audioiirfilter` | Custom IIR filter | Noise reduction |
| `audiofirfilter` | Custom FIR filter | Noise reduction |
| `level` | Reports audio level via bus messages | VU meter in UI |
| `spectrum` | Reports frequency spectrum | Spectrum analyzer |

**Live parameter changes:** GStreamer elements support live property changes while the pipeline is running. In go-gst:

```go
// Change gain while recording is active
volumeElement.Set("volume", 2.0)  // This takes effect immediately
```

No restart needed. The change applies to the next buffer that passes through.

### 5.3 Live Audio Transcription (Speech-to-Text)

**The user's request:** "live audio transcription"

This is the most ambitious new capability. There are several approaches:

#### Approach A: External Whisper Process via appsink

```
[pulsesrc] → [audioconvert] → [audioresample]
  → [capsfilter("audio/x-raw,rate=16000,channels=1")]   ← Whisper needs 16kHz mono
  → [tee]
      → branch 1: [volume → wavenc → filesink]           ← recording (full quality)
      → branch 2: [appsink] → Go callback →              ← transcription feed
           accumulate 5 seconds of PCM → write to temp WAV
           → invoke whisper CLI or API → get text → publish via WebSocket
```

**Pros:** Simple, works today with `whisper` CLI or cloud API
**Cons:** Latency (5-second chunks), external process

#### Approach B: GStreamer Whisper Plugin

There are community GStreamer plugins that wrap whisper.cpp or use cloud APIs. As of 2026, the landscape is:

- **gst-whisper** — wraps whisper.cpp as a GStreamer element (not packaged for Ubuntu)
- **Vosk** — offline speech recognition with a GStreamer element (available, but lower accuracy)

If a `whisper` element existed, the pipeline would be:
```
[pulsesrc] → [audioconvert] → [audioresample]
  → [capsfilter("audio/x-raw,rate=16000,channels=1")]
  → [whisper] → text output → appsink → Go callback → WebSocket
```

#### Approach C: Go-side audio chunking + API call

This is the most practical approach for now:

1. In the go-gst appsink callback, accumulate PCM buffers
2. Every N seconds (e.g., 3-5 seconds), flush the accumulated audio
3. Send to a transcription service (local whisper binary, cloud API, etc.)
4. Publish the transcription text via WebSocket to the UI

```go
sink.SetCallbacks(&app.SinkCallbacks{
    NewSampleFunc: func(s *app.Sink) gst.FlowReturn {
        sample := s.PullSample()
        buffer := sample.GetBuffer()
        audioData := buffer.Map(gst.MapRead)
        defer buffer.Unmap()

        // Accumulate in a buffer
        transcriptionBuffer.Write(audioData.AsBytes())

        // Every 3 seconds worth of audio
        if transcriptionBuffer.Len() >= 16000*2*3 {  // 16kHz, 16-bit, 3 seconds
            chunk := transcriptionBuffer.Bytes()
            transcriptionBuffer.Reset()
            go transcribeAsync(chunk)  // Call whisper/API in background
        }
        return gst.FlowOK
    },
})
```

### 5.4 Live Audio Level Monitoring (VU Meter)

The `level` element reports audio levels via bus messages. This is how you show a live VU meter or waveform in the UI:

```go
level, _ := gst.NewElement("level")
level.Set("interval", 50000000)   // Report every 50ms (in nanoseconds)
level.Set("post-messages", true)
level.Set("message", true)

// In your bus watch:
case gst.MessageElement:
    // Parse level message to get dB values
    // Send to browser via WebSocket for VU meter rendering
```

### 5.5 Shared Capture Source (No More Suspend/Restore)

**Current problem:** Preview and recording are separate FFmpeg processes. Both try to open the same X11 display. The server has to suspend previews before recording starts.

**GStreamer solution:** Use a **tee** element to fan out a single capture source:

```
                    ┌──→ [preview branch: videoscale → jpegenc → appsink]
                    │
[ximagesrc] → [tee] ┼──→ [recording branch: x264enc → mp4mux → filesink]
                    │
                    └──→ [screenshot branch: pngenc → filesink (on demand)]
```

When recording starts, you just add a new branch to the existing tee. No need to stop the preview. When recording stops, remove the recording branch. The preview keeps running the entire time.

This eliminates the `SuspendAll` / `RestoreSuspended` dance in `preview_manager.go`.


---

## 6. go-gst API Patterns — What the Intern Needs to Know

### 6.1 Building a Pipeline

The go-gst API (v1.4.0) follows this pattern:

```go
gst.Init(nil)

// Create an empty pipeline
pipeline, _ := gst.NewPipeline("")

// Create elements
src, _ := gst.NewElement("ximagesrc")
src.Set("startx", 0)
src.Set("starty", 0)

convert, _ := gst.NewElement("videoconvert")
sink, _ := gst.NewElement("appsink")

// Add elements to the pipeline
pipeline.AddMany(src, convert, sink)

// Link elements (source pad of one → sink pad of the next)
src.Link(convert)
convert.Link(sink)

// Start the pipeline
pipeline.SetState(gst.StatePlaying)
```

### 6.2 Getting Data Out (appsink Pattern)

```go
sink, _ := app.NewAppSink()
sink.SetCaps(gst.NewCapsFromString("image/jpeg"))

sink.SetCallbacks(&app.SinkCallbacks{
    NewSampleFunc: func(s *app.Sink) gst.FlowReturn {
        sample := s.PullSample()
        if sample == nil {
            return gst.FlowEOS
        }

        buffer := sample.GetBuffer()
        mapInfo := buffer.Map(gst.MapRead)
        defer buffer.Unmap()

        // mapInfo.AsBytes() gives you the raw JPEG bytes
        // Pass them to your callback (e.g., storePreviewFrame)
        onFrame(mapInfo.AsBytes())

        return gst.FlowOK
    },
})

// Add sink.Element (not sink itself) to the pipeline
pipeline.AddMany(src, convert, sink.Element)
```

### 6.3 Handling Bus Messages

```go
bus := pipeline.GetPipelineBus()
bus.AddWatch(func(msg *gst.Message) bool {
    switch msg.Type() {
    case gst.MessageError:
        err := msg.ParseError()
        fmt.Println("Error:", err.Error())
        fmt.Println("Debug:", err.DebugString())
        return false  // Stop watching
    case gst.MessageEOS:
        fmt.Println("End of stream")
        return false
    }
    return true  // Continue watching
})
```

**Important:** The bus watch needs a GLib main loop running to dispatch messages:

```go
mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)
go mainLoop.Run()
// ...
mainLoop.Quit()
```

### 6.4 Clean Shutdown

```go
// Send End-Of-Stream to flush encoders and finalize files
pipeline.SendEvent(gst.NewEOSEvent())

// Wait briefly for EOS to be processed
time.Sleep(500 * time.Millisecond)

// Stop the pipeline
pipeline.BlockSetState(gst.StateNull)

// Stop the main loop
mainLoop.Quit()
```

### 6.5 Dynamic Property Changes (Live Effects)

```go
// You can change element properties while the pipeline is running
volumeElement.Set("volume", 2.0)   // Double the gain
// The change applies to the very next audio buffer
```

### 6.6 API Gotchas (Lessons Learned the Hard Way)

| Gotcha | Explanation |
|--------|-------------|
| `gst.NewCapsFromString()` returns `*Caps`, not `(*Caps, error)` | Single return value in go-gst v1.4.0 |
| Use `app.NewAppSink()`, not `gst.NewElement("appsink")` + wrapping | The app.NewAppSink() gives you the typed wrapper directly |
| Add `sink.Element` to the pipeline, not `sink` | The `app.Sink` wrapper has an `.Element` field |
| Method names have no `Get` prefix | `pad.Name()` not `pad.GetName()`, `factory.GetMetadata()` is an exception |
| `mainLoop.Run()` returns nothing | Use `mainLoop.RunError()` if you need the error |
| CGO compilation is slow (~60-90 seconds) | Expected — GStreamer is a C library |
| Double-free possible with factory + element in same scope | Avoid calling GetMetadata on a factory after creating an element from it |


---

## 7. Proposed Architecture

### 7.1 The Runtime Seam

We introduce a media runtime abstraction layer between the application logic and the concrete media backend:

```
pkg/dsl/          ← DSL parsing, normalization, compilation (UNCHANGED)
pkg/media/        ← NEW: Runtime interfaces and types
  types.go          RecordingRuntime, PreviewRuntime, RecordingSession, PreviewSession
  events.go         RecordingEvent, StructuredLog, etc.
pkg/media/gst/    ← NEW: GStreamer implementation
  preview.go        GStreamerPreviewRuntime
  recording.go      GStreamerRecordingRuntime
  pipeline.go       Shared pipeline builder helpers
  audio.go          Audio pipeline construction
  bus.go            Bus message handling utilities
internal/web/     ← Web layer (MOSTLY UNCHANGED)
  preview_runner.go → deleted (replaced by pkg/media/gst/preview.go)
  preview_manager.go → preserved (just changes its runner dependency)
```

### 7.2 Interface Design

```go
// pkg/media/types.go

type PreviewRuntime interface {
    StartPreview(ctx context.Context, source dsl.EffectiveVideoSource, opts PreviewOptions) (PreviewSession, error)
}

type PreviewSession interface {
    Wait() error
    Stop(ctx context.Context) error
    LatestFrame() ([]byte, error)
    TakeScreenshot(ctx context.Context, opts ScreenshotOptions) ([]byte, error)
}

type RecordingRuntime interface {
    StartRecording(ctx context.Context, plan *dsl.CompiledPlan, opts RecordingOptions) (RecordingSession, error)
}

type RecordingSession interface {
    Wait() (*RecordingResult, error)
    Stop(ctx context.Context) error
}

type PreviewOptions struct {
    OnFrame func([]byte)
    OnLog   func(stream, message string)
}

type RecordingOptions struct {
    MaxDuration time.Duration
    EventSink   func(RecordingEvent)
    Logger      func(string, ...any)
}

type RecordingResult struct {
    State      string
    Reason     string
    Outputs    []dsl.PlannedOutput
    StartedAt  time.Time
    FinishedAt time.Time
}
```

### 7.3 Preview Pipeline Design (GStreamer)

```
For "display" type sources:

[ximagesrc startx=X starty=Y endx=W endy=H use-damage=false]
  → [videoconvert]
  → [videoscale]
  → [capsfilter "video/x-raw,width=640"]
  → [videorate]
  → [capsfilter "video/x-raw,framerate=5/1"]
  → [jpegenc quality=30]
  → [appsink caps="image/jpeg"]
      → Go callback: onFrame(jpegBytes)

For "camera" type sources:

[v4l2src device=/dev/videoN]
  → [videoconvert]
  → [videoscale]
  → [capsfilter "video/x-raw,width=640"]
  → [videorate]
  → [capsfilter "video/x-raw,framerate=5/1"]
  → [jpegenc quality=30]
  → [appsink caps="image/jpeg"]
      → Go callback: onFrame(jpegBytes)
```

### 7.4 Recording Pipeline Design (GStreamer)

For each video job:

```
[ximagesrc or v4l2src]
  → [videoconvert]
  → [capsfilter "video/x-raw,width=W,height=H,framerate=F/1"]
  → [x264enc tune=zerolatency speed-preset=veryfast crf=17]
  → [mp4mux or qtmux]
  → [filesink location=output.mp4]
```

For each audio mix job:

```
[pulsesrc device=mic1] → [audioconvert] → [volume volume=G1] ─┐
[pulsesrc device=mic2] → [audioconvert] → [volume volume=G2] ─┤→ [audiomixer]
                                                                 │
[audiomixer] → [audioconvert] → [wavenc or opusenc] → [filesink]
```

### 7.5 Screenshot Design

**Simplest approach:** Extract the latest JPEG frame from the preview appsink:

```go
func (s *GStreamerPreviewSession) TakeScreenshot(ctx context.Context) ([]byte, error) {
    // The preview session already has the latest frame in memory
    // Just return it (it's already a JPEG)
    s.mu.Lock()
    frame := append([]byte(nil), s.latestFrame...)
    s.mu.Unlock()
    return frame, nil
}
```

**Higher-quality approach:** Build a one-shot PNG pipeline:

```go
func takeScreenshotPNG(source dsl.EffectiveVideoSource) ([]byte, error) {
    // Build: ximagesrc(num-buffers=1) → videoconvert → pngenc → appsink
    // Run to completion (takes ~200ms)
    // Return PNG bytes
}
```

### 7.6 Live Audio Effects + Monitoring Design

```
[pulsesrc] → [audioconvert] → [audioresample]
  → [volume volume=1.0]         ← adjustable live
  → [audiodynamic]               ← optional compressor
  → [tee]
      → branch 1: [wavenc → filesink]                ← recording
      → branch 2: [level] → [appsink]                 ← monitoring (level + waveform)
      → branch 3: [capsfilter(16kHz/mono)] → [appsink] ← transcription feed
```

The Go code manages effect parameters via `element.Set()`:

```go
// Called from HTTP API: POST /api/audio/effects
func (s *GStreamerRecordingSession) SetAudioGain(gain float64) {
    s.volumeElement.Set("volume", gain)
}
```

### 7.7 Transcription Integration Design

```
[appsink from transcription tee branch]
  → Go callback accumulates PCM buffers
  → Every 3 seconds:
      1. Encode accumulated PCM as WAV (in-memory)
      2. Send to transcription backend:
         - Option A: local `whisper` CLI subprocess
         - Option B: cloud API (OpenAI, Google, etc.)
         - Option C: Go whisper binding (whisper.cpp)
      3. Publish transcription text via WebSocket
```

This keeps transcription decoupled from GStreamer. GStreamer just provides the audio stream; the transcription is a separate concern.


---

## 8. Element Reference Table

Every GStreamer element needed for screencast-studio, what it replaces in FFmpeg, and its status on this machine:

### 8.1 Video Source Elements

| GStreamer Element | FFmpeg Equivalent | Description | Available |
|---|---|---|---|
| `ximagesrc` | `-f x11grab` | X11 screen capture | ✓ |
| `pipewiresrc` | (no equivalent) | PipeWire screen capture (Wayland) | ✓ |
| `v4l2src` | `-f v4l2` | Webcam capture (Video4Linux2) | ✓ |

### 8.2 Audio Source Elements

| GStreamer Element | FFmpeg Equivalent | Description | Available |
|---|---|---|---|
| `pulsesrc` | `-f pulse` | PulseAudio/PipeWire audio capture | ✓ |
| `alsasrc` | `-f alsa` | Direct ALSA capture | ✓ |

### 8.3 Video Processing Elements

| GStreamer Element | FFmpeg Equivalent | Description | Available |
|---|---|---|---|
| `videoconvert` | auto (codec level) | Color format conversion | ✓ |
| `videoscale` | `scale=W:H` | Resize video | ✓ |
| `videorate` | `fps=N` | Frame rate conversion | ✓ |
| `capsfilter` | (implicit in codec options) | Force specific format | ✓ (built-in) |
| `textoverlay` | `drawtext` | Overlay text on video | ✓ |
| `timeoverlay` | (no equivalent) | Overlay running time | ✓ |
| `clockoverlay` | (no equivalent) | Overlay clock | ✓ |

### 8.4 Audio Processing Elements

| GStreamer Element | FFmpeg Equivalent | Description | Available |
|---|---|---|---|
| `audioconvert` | auto (codec level) | Audio format conversion | ✓ |
| `audioresample` | `-ar RATE` | Audio resampling | ✓ |
| `volume` | `volume=X` filter | Volume/gain | ✓ |
| `audiomixer` | `amix=inputs=N` | Mix multiple audio streams | ✓ |
| `level` | `astats` | Report audio levels (VU meter) | ✓ |
| `audiodynamic` | `acompressor` | Compressor/limiter | ✓ |

### 8.5 Encoder Elements

| GStreamer Element | FFmpeg Equivalent | Description | Available |
|---|---|---|---|
| `jpegenc` | `-vcodec mjpeg` | JPEG image encoder | ✓ |
| `pngenc` | (no equivalent in pipeline) | PNG image encoder | ✓ |
| `x264enc` | `-c:v libx264` | H.264 video encoder | ✓ |
| `opusenc` | `-c:a libopus` | Opus audio encoder | ✓ |
| `wavenc` | `-c:a pcm_s16le` | WAV audio container | ✓ |
| `vorbisenc` | `-c:a libvorbis` | Vorbis audio encoder | ✓ |

### 8.6 Muxer Elements

| GStreamer Element | FFmpeg Equivalent | Description | Available |
|---|---|---|---|
| `mp4mux` | `-f mp4` | MP4 container | ✓ |
| `oggmux` | `-f ogg` | Ogg container | ✓ |
| `qtmux` | `-f mov` | QuickTime/MOV container | ✓ |

### 8.7 Sink Elements

| GStreamer Element | FFmpeg Equivalent | Description | Available |
|---|---|---|---|
| `appsink` | `pipe:1` (stdout) | Pass data to Go code | ✓ |
| `filesink` | output filename | Write to file | ✓ |
| `gdkpixbufsink` | (no equivalent) | Post image to bus (for screenshots) | ✓ |

### 8.8 Utility Elements

| GStreamer Element | FFmpeg Equivalent | Description | Available |
|---|---|---|---|
| `tee` | (no equivalent) | Split stream to multiple branches | ✓ (built-in) |
| `queue` | (no equivalent) | Thread boundary + buffering | ✓ (built-in) |
| `fakesink` | `-f null` | Discard data (testing) | ✓ (built-in) |

---

## 9. FFmpeg → GStreamer Translation Cheat Sheet

### 9.1 Video Capture

| FFmpeg | GStreamer |
|--------|-----------|
| `-f x11grab -i :0.0` | `ximagesrc display-name=":0"` |
| `-f x11grab -i :0.0+100,200` | `ximagesrc startx=100 starty=200` |
| `-f x11grab -video_size 1280x720` | `ximagesrc endx=1279 endy=719` |
| `-framerate 24` | `videorate ! "video/x-raw,framerate=24/1"` |
| `-draw_mouse 1` | `ximagesrc show-cursor=true` |
| `-window_id 0x1234` | Not directly available in ximagesrc |

### 9.2 Audio Capture

| FFmpeg | GStreamer |
|--------|-----------|
| `-f pulse -i default` | `pulsesrc device=default` |
| `-sample_rate 48000` | `capsfilter "audio/x-raw,rate=48000"` |
| `-channels 2` | `capsfilter "audio/x-raw,channels=2"` |
| `volume=1.5` filter | `volume volume=1.5` element |
| `amix=inputs=2` | `audiomixer` with 2 request pads |

### 9.3 Video Encoding

| FFmpeg | GStreamer |
|--------|-----------|
| `-c:v libx264 -preset veryfast` | `x264enc speed-preset=veryfast` |
| `-crf 17` | `x264enc quantizer=17` |
| `-pix_fmt yuv420p` | Handled by `videoconvert` |
| `-c:v ffv1` (lossless) | `avenc_ffv1` (via gst-libav) or `ffv1enc` |
| `-q:v 7` (JPEG quality) | `jpegenc quality=30` (different scale) |

### 9.4 Audio Encoding

| FFmpeg | GStreamer |
|--------|-----------|
| `-c:a pcm_s16le output.wav` | `wavenc ! filesink` |
| `-c:a libopus -b:a 160k` | `opusenc bitrate=160000` |
| `-c:a aac -b:a 192k` | `avenc_aac bitrate=192000` (via gst-libav) |
| `-c:a libmp3lame -b:a 192k` | `lamemp3enc bitrate=192` |

### 9.5 Output

| FFmpeg | GStreamer |
|--------|-----------|
| `output.mp4` | `mp4mux ! filesink location=output.mp4` |
| `output.mov` | `qtmux ! filesink location=output.mov` |
| `output.wav` | `wavenc ! filesink location=output.wav` |
| `pipe:1` (stdout) | `appsink` (in Go) or `fdsrc fd=1` |


---

## 10. Migration Phases

### Phase 0: Introduce the Runtime Seam (1-2 days)

**Goal:** Create interfaces without changing behavior.

1. Create `pkg/media/types.go` with the runtime interfaces
2. Create `pkg/media/gst/` directory
3. Wrap the current FFmpeg code behind the new interfaces (adapter pattern)
4. Wire the new interfaces into `Application` and `PreviewManager`
5. Run all existing tests — nothing should change

**Validation:** All existing tests pass. The web UI works identically.

### Phase 1: GStreamer Preview (3-5 days)

**Goal:** Replace FFmpeg preview with GStreamer.

1. Implement `GStreamerPreviewRuntime` in `pkg/media/gst/preview.go`
2. Build the preview pipeline programmatically (as validated in experiment 06)
3. Wire the appsink callback into `PreviewManager.storePreviewFrame()`
4. Test all source types: display, region, camera
5. Verify preview suspend/restore still works

**Validation:**
- Display preview works in browser
- Camera preview works in browser
- Preview suspend/restore during recording works
- No MJPEG parsing — just raw JPEG bytes from appsink

### Phase 2: GStreamer Recording (5-7 days)

**Goal:** Replace FFmpeg recording with GStreamer.

1. Implement `GStreamerRecordingRuntime` in `pkg/media/gst/recording.go`
2. Build video recording pipelines (one pipeline per output)
3. Build audio recording pipelines (with mixer for multiple sources)
4. Implement graceful stop via EOS
5. Map session states and events to existing `SessionManager` expectations

**Validation:**
- Recording produces valid MP4/MOV files
- Recording produces valid WAV/OGG files
- Stop recording creates properly finalized files
- Cancel recording cleans up correctly
- Max duration timeout works

### Phase 3: Screenshots + Live Audio Effects (3-5 days)

**Goal:** Add new capabilities.

1. Add `TakeScreenshot()` to `PreviewSession`
2. Add HTTP endpoint `GET /api/previews/{id}/screenshot`
3. Add audio effect elements (volume, audiodynamic) to recording pipeline
4. Add HTTP endpoint to change audio effects live: `POST /api/audio/effects`
5. Wire audio level messages from `level` element to WebSocket for VU meter

**Validation:**
- Screenshot endpoint returns a JPEG/PNG image
- Audio gain can be adjusted during recording
- VU meter shows live audio levels in browser

### Phase 4: Shared Capture + Remove FFmpeg (3-5 days)

**Goal:** Eliminate device contention, remove dead code.

1. Introduce capture graph registry (keyed by source signature)
2. Use tee to share capture sources between preview and recording
3. Remove preview suspend/restore workaround
4. Delete `pkg/recording/ffmpeg.go`
5. Delete `internal/web/preview_runner.go`
6. Remove FFmpeg from system dependencies

### Phase 5: Live Transcription (5-7 days)

**Goal:** Add live speech-to-text.

1. Add transcription tee branch to audio pipeline
2. Implement PCM chunking in appsink callback
3. Integrate with transcription backend (whisper CLI or API)
4. Publish transcription text via WebSocket
5. Add transcription UI in browser

---

## 11. Risks and Mitigations

### Risk 1: go-gst Memory Management Bugs

**Observed:** Double-free crash when calling `GetMetadata` on a factory after creating an element from it.

**Mitigation:**
- Avoid mixing factory queries with element creation in the same scope
- Create elements, configure them, add to pipeline — don't introspect factories after that
- If issues persist, consider a thin C helper process that manages GStreamer pipelines and communicates with Go via structured IPC (Unix sockets + JSON/NDJSON)

### Risk 2: Slow CGO Compilation

**Observed:** ~60-90 seconds per `go build` due to GStreamer C headers.

**Mitigation:**
- Use `go install` caching effectively
- Structure code so GStreamer-dependent code is in clearly separated packages
- CI can cache the build cache
- Development iteration: use `go test -run SpecificTest` to avoid rebuilding everything

### Risk 3: Pipeline Latency

**Concern:** Native GStreamer pipelines may introduce latency compared to FFmpeg's direct pipe I/O.

**Mitigation:**
- Use `tune=zerolatency` on x264enc
- Keep queue sizes small (1-2 buffers)
- Monitor latency with GStreamer's built-in latency reporting: `GST_DEBUG=GST_TRACER:7 gst-launch ...`

### Risk 4: Wayland Compatibility

**Current state:** `ximagesrc` only works on X11. Wayland requires `pipewiresrc` + xdg-desktop-portal.

**Mitigation:**
- For now, stick with X11 (the current FFmpeg version also requires X11)
- Future: add `pipewiresrc` as an alternative source type in the DSL
- GStreamer's `pipewiresrc` element is already installed and available

### Risk 5: Audio Sync Issues

**Concern:** Separate video and audio pipelines may drift over time.

**Mitigation:**
- GStreamer's timestamp system is designed to handle this
- Use a single pipeline with both audio and video branches if possible
- If separate pipelines are needed, use the same clock source

---

## 12. Glossary

| Term | Meaning |
|------|---------|
| **Element** | A single processing unit in GStreamer (source, filter, encoder, sink) |
| **Pad** | Input/output connection point on an element |
| **Caps** | Capabilities — the format specification for data on a pad |
| **Pipeline** | A collection of linked elements managed together |
| **Bus** | Message queue for pipeline events (errors, EOS, state changes) |
| **EOS** | End Of Stream — signals no more data |
| **Tee** | Element that copies one input to multiple outputs |
| **Appsink** | Element that passes data from GStreamer to application code |
| **Appsrc** | Element that accepts data from application code into GStreamer |
| **Muxer** | Combines audio/video streams into a container format (MP4, OGG) |
| **Encoder** | Converts raw data to compressed format (H.264, Opus, JPEG) |
| **Caps negotiation** | Process where linked elements agree on a data format |
| **CGO** | Go's mechanism for calling C code — needed for go-gst bindings |
| **go-gst** | Go bindings for GStreamer (github.com/go-gst/go-gst) |
| **DSL** | Domain-Specific Language — the YAML config that describes recording setups |

---

## 13. References

- **go-gst repository:** https://github.com/go-gst/go-gst
- **go-gst appsink example:** https://github.com/go-gst/go-gst/blob/v1.4.0/examples/appsink/main.go
- **GStreamer documentation:** https://gstreamer.freedesktop.org/documentation/
- **GStreamer Pad Probe docs:** https://gstreamer.freedesktop.org/documentation/plugin-development/advanced/pad-probes.html
- **Existing ticket SCS-0011:** `ttmp/2026/04/10/SCS-0011--gstreamer-migration-architecture-and-intern-guide/`
- **Experiment scripts:** `ttmp/2026/04/13/SCS-0012--.../scripts/01-*` through `07-*`

### Experiment Inventory

| Script | What it Tests | Result |
|--------|--------------|--------|
| `01-check-gstreamer-env.sh` | GStreamer version, elements, packages | ✓ 1678 elements, v1.24.2 |
| `02-install-gstreamer-dev-headers.sh` | Dev header installation guide | ✓ Verified pkg-config |
| `03-gst-launch-preview-x11.sh` | Video preview, recording, screenshot via gst-launch | ✓ All three work |
| `04-gst-launch-audio-capture.sh` | Audio capture (WAV, Opus, gain) via gst-launch | ✓ All three work |
| `05-go-gst-basic-test/` | go-gst compilation and element discovery | ✓ All 20 elements found |
| `06-go-gst-preview-pipeline/` | Native Go video preview pipeline | ✓ 24 frames/5sec, JPEG via appsink |
| `07-go-gst-audio-capture/` | Native Go audio recording pipeline | ✓ Valid WAV via filesink |
