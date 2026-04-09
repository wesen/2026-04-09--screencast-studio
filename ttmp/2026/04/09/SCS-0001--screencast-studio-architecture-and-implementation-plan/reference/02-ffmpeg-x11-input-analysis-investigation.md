---
Title: FFmpeg X11 Input Analysis Investigation
Ticket: SCS-0001
Status: active
Topics:
    - backend
    - video
    - cli
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/recording/ffmpeg.go
      Note: Runtime FFmpeg argv construction for bounded capture jobs
    - Path: pkg/recording/run.go
      Note: Runtime process supervision and FFmpeg stdout/stderr capture
    - Path: ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/direct-ffmpeg-region-smoke.sh
      Note: Direct FFmpeg baseline used to isolate runtime from host behavior
    - Path: ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/smoke-record-region.sh
      Note: CLI smoke repro for the bounded recording runtime
ExternalSources: []
Summary: Reference writeup capturing why short X11 screen captures on this host take roughly five seconds of wall-clock time, how that was diagnosed, and what FFmpeg knobs are relevant if startup latency becomes important later.
LastUpdated: 2026-04-09T14:59:13.801934705-04:00
WhatFor: Preserve the investigation around FFmpeg input analysis on x11grab so future runtime work does not have to rediscover why short captures can appear slow even when the recorder lifecycle is correct.
WhenToUse: Read this when debugging slow startup for screen capture, tuning FFmpeg input options for x11grab, or deciding whether reduced probing should be part of the production runtime.
---

# FFmpeg X11 Input Analysis Investigation

## Goal

Capture the exact finding behind the "why does a one-second X11 screen capture take about five seconds" debugging session, including the evidence that this delay is primarily FFmpeg input analysis behavior rather than a bug in the screencast-studio state machine.

## Context

While implementing the CLI-first `record` runtime, bounded test captures appeared to hang for roughly five to six seconds even when the runtime passed `-t 1` to FFmpeg. The first suspicion was that the new recorder coordination logic was still mishandling bounded completion or graceful stop.

That hypothesis became weaker once the same timing behavior reproduced in a direct FFmpeg invocation outside the runtime. At that point the investigation shifted from "is our state machine wrong?" to "what is FFmpeg doing during startup for an x11grab input at low frame rate?"

The relevant runtime and repro commands are stored in:

- [smoke-record-region.sh](/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/smoke-record-region.sh)
- [direct-ffmpeg-region-smoke.sh](/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/direct-ffmpeg-region-smoke.sh)

The bounded runtime command currently emits FFmpeg argv from [pkg/recording/run.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go), and the argument shape for X11 screen capture is built in [pkg/recording/ffmpeg.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/ffmpeg.go).

## Quick Reference

### Working theory

For `x11grab`, FFmpeg spends up to its default analysis window probing the live input stream before proceeding. On this host that analysis phase takes about five seconds, which dominates the wall-clock duration of short captures.

The delay is most likely coming from stream probing and frame-rate inference, not from X11 connection setup and not from our recording state machine.

### Key evidence

Observed FFmpeg diagnostic output:

```text
Opening an input file: :0.0+0,0.
[x11grab @ 0x60e7bb6e43c0] max_analyze_duration 5000000 reached at 5000000 microseconds st:0
[x11grab @ 0x60e7bb6e43c0] rfps: 1.916667 0.011577
[x11grab @ 0x60e7bb6e43c0] rfps: 2.000000 0.000000
[x11grab @ 0x60e7bb6e43c0] rfps: 2.083333 0.011571
[x11grab @ 0x60e7bb6e43c0] rfps: 3.916667 0.011580
```

Interpretation:

- `max_analyze_duration 5000000` means FFmpeg spent the default 5,000,000 microseconds, or five seconds, in stream analysis.
- `rfps` logging shows FFmpeg trying to infer frame timing characteristics from a live source.
- With `-framerate 2`, five seconds only yields about ten frames, so the analysis window is a very large fraction of the whole capture.

### Runtime FFmpeg argv that reproduced the behavior

```bash
ffmpeg -hide_banner -loglevel error -y -t 1 -f x11grab -framerate 2 -video_size 320x240 -draw_mouse 0 -i :0.0+0,0 -c:v libx264 -preset veryfast -crf 24 -pix_fmt yuv420p ttmp/smoke-output/smoke_region.mkv
```

### Direct FFmpeg baseline used to isolate the runtime

```bash
TIMEFMT='%E'; time ./ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/direct-ffmpeg-region-smoke.sh /tmp/direct-smoke-runtime-check.mkv
```

Observed result:

```text
5.57s
```

That result is important because it shows the five-second wall-clock delay persists even when the screencast-studio runtime is removed from the equation.

### Why this matters for our recorder

- The formal session state machine did improve correctness. It eliminated the zero-byte output problem that happened when Go tried to stop FFmpeg exactly when bounded duration elapsed.
- The remaining five-second delay is not evidence that the state machine is wrong.
- If startup latency matters later, the likely tuning surface is FFmpeg probing behavior, not session-state transitions.

### FFmpeg knobs worth testing later

These are the likely input-analysis knobs for x11grab if we want to reduce startup latency:

```bash
-analyzeduration 0
-probesize 32
-fpsprobesize 2
```

Likely policy for v1:

- `x11grab`: aggressive reduction of analysis and probing
- `pulse`: likely small probing as well, since sample rate and channel count are explicitly set
- `v4l2`: probably reduced probing too, but only after validating that specific cameras still initialize reliably

### Practical conclusion

For screen capture, full FFmpeg input analysis probably provides little value because we already know the essential input shape from our compiled plan:

- input format is explicit
- frame rate is explicit
- region size is explicit
- there is no container demux ambiguity

That makes default live-input analysis mostly overhead for this product.

## Usage Examples

### Example 1: Reproduce the behavior through the runtime

```bash
./ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/smoke-record-region.sh
```

Use this when validating the bounded `record` command end-to-end.

### Example 2: Reproduce the behavior without the runtime

```bash
TIMEFMT='%E'; time ./ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/direct-ffmpeg-region-smoke.sh /tmp/direct-smoke-runtime-check.mkv
```

Use this when you need to answer the question "is this runtime coordination, or is this just FFmpeg/x11grab on this host?"

### Example 3: Future tuning experiment

This exact command was not yet committed into the runtime, but it is the recommended next manual experiment if startup latency becomes important:

```bash
ffmpeg -hide_banner -loglevel debug -y \
  -t 5 \
  -analyzeduration 0 \
  -probesize 32 \
  -fpsprobesize 2 \
  -f x11grab \
  -framerate 2 \
  -video_size 320x240 \
  -draw_mouse 0 \
  -i :0.0+0,0 \
  -c:v libx264 \
  -preset veryfast \
  -crf 24 \
  -pix_fmt yuv420p \
  /tmp/x11grab-fast-probe-test.mkv
```

If this materially improves startup time, the next implementation step would be to add a live-capture input policy in the recorder that disables or sharply reduces probing for screen capture sources.

## Related

- [01-screencast-studio-system-design.md](/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/design-doc/01-screencast-studio-system-design.md)
- [01-diary.md](/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/reference/01-diary.md)
- [scripts/README.md](/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/README.md)
