---
Title: Small graph hosting ladder debugging plan
Ticket: SCS-0016
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - backend
    - analysis
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/recording.go
      Note: Current production direct recording graph that motivated the smaller-graph ladder
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/15-go-direct-record-full-desktop-harness/main.go
      Note: Existing app-like direct Go control that reproduces the hot path on the full graph
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/17-go-manual-direct-full-desktop-harness/main.go
      Note: Existing manual Go direct control that is structurally simpler but still hotter than gst-launch
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/16-gst-launch-direct-full-desktop-match-app.sh
      Note: Existing cooler gst-launch full-graph control used as the baseline for the hosting gap
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/28-python-direct-hosting-control-matrix.sh
      Note: Existing Python controls that already showed hosted Python is much closer to gst-launch than Go on the full graph
ExternalSources: []
Summary: Stage-by-stage plan to determine the first graph segment where Go-hosted behavior diverges from Python and gst-launch, and to decide whether the remaining issue is graph correctness, cgo usage, or host-process memory behavior.
LastUpdated: 2026-04-15T03:25:00-04:00
WhatFor: Turn the current Go-vs-Python-vs-gst-launch hosting suspicion into a disciplined smaller-graph experiment ladder with clear stop conditions and interpretation rules.
WhenToUse: Read before implementing or interpreting the small-graph ladder so the experiments stay stage-based and question-driven.
---

# Small graph hosting ladder debugging plan

## Executive Summary

The current full-graph evidence says the remaining gap is probably not explained by graph shape alone. We now have five strong comparison points on the full desktop direct-recording path:

- Go app-like direct control
- Go manual direct control
- Python `Gst.parse_launch(...)` control
- Python manual-graph control
- `gst-launch-1.0` control

The most important finding from the newest comparison is that **Python-hosted GStreamer is much closer to `gst-launch` than to Go**. That sharply lowers the probability that the gap is caused by a generic “hosted language vs `gst-launch`” penalty. It also lowers the probability that the graph is fundamentally malformed in every embedded control. Instead, the remaining suspicion is increasingly specific: **Go-hosted process behavior**, especially around memory behavior, faults, allocator/runtime interaction, or some other Go-specific hosting effect.

The next disciplined move is therefore not more full-graph surgery. It is a **small-graph ladder**. We will start from `ximagesrc ! fakesink` and add one graph segment at a time across three hosts — Go manual, Python manual, and `gst-launch` — while capturing CPU plus page-fault counters at every stage. The goal is to identify the **first stage where Go diverges materially**. Once that first divergence point is found, we can decide whether to keep looking at graph construction, switch to tighter memory/fault instrumentation, or test cgo/build flags with much stronger evidence.

## Problem Statement

The direct full-desktop graph is already narrow enough to be useful, but it is still too large to localize the first real divergence point. The current full-graph evidence leaves three broad interpretations alive:

1. **Graph correctness / graph realization problem**
   - Example: Go realizes a materially different graph than the cooler controls.
   - Current evidence against this: the normalized Go graphs now match; Python manual also uses an explicit graph and stays comparatively cool.

2. **Generic embedded-control penalty**
   - Example: any host-language GStreamer control would be much hotter than `gst-launch`.
   - Current evidence against this: Python controls are much closer to `gst-launch` than to Go.

3. **Go-host-process-specific penalty**
   - Example: Go runtime / cgo / allocator / page-fault / thread / memory-pool interaction changes the steady-state native cost.
   - Current evidence for this: Go shows a huge page-fault delta relative to `gst-launch`; Python looks much closer to the cooler side of the matrix.

The problem is that the full graph still combines multiple stages:

- raw X11 capture
- colorspace conversion
- frame pacing / raw caps shaping
- H.264 encoding
- bitstream parsing
- muxing / file output

If the divergence starts very early, then the encoder is not the whole story. If it starts only when encoding begins, the most likely explanation shifts toward encoder input buffers / allocator / memory behavior. Without a stage ladder, the current suspicion remains too coarse.

## Proposed Solution

### Core idea

Build a ladder of smaller graphs and run each stage across three host/control families:

- **Go manual graph**
- **Python manual graph**
- **`gst-launch` shell control**

All three controls should use the same desktop source and the same stage-local graph. For each stage, capture:

- average CPU (`pidstat`)
- max CPU (`pidstat`)
- `perf stat` counters for:
  - `page-faults`
  - `minor-faults`
  - `major-faults`
  - `cycles`
  - `instructions`
  - `context-switches`
  - `cpu-migrations`
- valid output proof for the mux stage (`ffprobe`)
- optional `.dot` dumps for Go/Python stages for graph sanity

### Stages

#### Stage A — capture only

```text
ximagesrc ! fakesink
```

Question answered:
- Does Go already diverge at raw capture / source scheduling / memory-touch time?

Interpretation:
- If Go is already much hotter here, the problem is upstream of conversion and encoding.
- If the hosts are still close here, the later stages become more interesting.

#### Stage B — add conversion

```text
ximagesrc ! videoconvert ! fakesink
```

Question answered:
- Does the divergence begin when raw frames must be transformed / reformatted?

Interpretation:
- A first jump here would point toward raw buffer negotiation / allocation / conversion / pool behavior.

#### Stage C — add rate + raw caps

```text
ximagesrc ! videoconvert ! videorate ! capsfilter(video/x-raw,format=I420,framerate=24/1,pixel-aspect-ratio=1/1) ! fakesink
```

Question answered:
- Does the divergence begin when we force the same steady-state raw frame contract used by the recorder path?

Interpretation:
- A first jump here would make the shaped raw-video stage itself the important boundary, even before encoding.

#### Stage D — add encoder only

```text
ximagesrc ! videoconvert ! videorate ! capsfilter(...) ! x264enc ! fakesink
```

Question answered:
- Is the first major gap introduced by handing shaped raw frames into `x264enc`?

Interpretation:
- A first jump here would strongly support an encoder-input / memory / threading explanation.

#### Stage E — add H.264 parser

```text
... ! x264enc ! h264parse ! fakesink
```

Question answered:
- Is the parser stage materially changing cost or just passing the same native workload through?

Interpretation:
- This stage is mostly a small control boundary, not the likely main culprit, but it helps distinguish encode-only from encode+parse.

#### Stage F — add mux + file output

```text
... ! x264enc ! h264parse ! qtmux|mp4mux ! filesink
```

Question answered:
- Does the final recording completion/output path add a meaningful host-specific penalty?

Interpretation:
- If the gap only widens here, mux/filesink/finalization deserves more attention.
- If the big divergence already happened earlier, file output is mostly a tail detail.

## Design Decisions

### 1. Use manual Go and manual Python controls, not only parse-launch

Rationale:
- `gst-launch` already provides the parse-launch/control baseline.
- Python manual and Go manual keep graph construction explicit and parallel.
- If both manual embedded graphs are cool, then Go app-specific higher layers deserve renewed scrutiny.
- If Python manual stays cool while Go manual is hot, that points more specifically at Go hosting rather than parse-launch differences.

### 2. Keep the stage ladder desktop-only and full-desktop sized

Rationale:
- The earlier confusion around `960x640` vs `2880x1920` already taught us that resolution mismatches can distort conclusions.
- The ladder should stay on the same full-desktop workload class as the full-gap investigation.

### 3. Capture page-fault counters at every stage, not just CPU

Rationale:
- Earlier trusted `perf stat` comparison showed page faults as the strongest Go-vs-`gst-launch` differentiator on the full graph.
- If the fault delta first appears at a specific stage, that is likely the most important stage boundary in the whole ladder.

### 4. Treat the first divergence point as the main outcome, not the final full-graph ranking

Rationale:
- The point of this ladder is localization, not just generating more averages.
- A small early divergence is more interesting than a large late aggregate number because it tells us where to instrument next.

## Alternatives Considered

### Alternative A: keep reasoning only from the full direct graph

Rejected because:
- it leaves too many stages coupled together,
- and it makes it too easy to blame the encoder without proving where the divergence first appears.

### Alternative B: immediately test cgo compiler optimization flags on the full graph

Deferred, not rejected.

Rationale:
- it is a useful later control,
- but it is more convincing after the small-graph ladder identifies the first divergence stage.
- If divergence starts before heavy native work, build flags matter less. If divergence starts exactly at or before the encoder boundary, cgo/build controls become more interesting.

### Alternative C: immediately switch to eBPF for per-thread attribution

Deferred.

Rationale:
- the ladder is cheaper and easier to replay,
- and it will likely give a much cleaner next question for any later eBPF work.

## Implementation Plan

### Slice 1: write the debugging plan and define interpretation rules

- add this design doc
- state the exact stage list
- define the metrics captured per stage
- define what counts as a meaningful divergence

### Slice 2: implement ticket-local stage controls

Add ticket-local scripts under `scripts/`:

- `29-go-manual-stage-ladder-harness/`
- `30-python-manual-stage-ladder-harness/`
- `31-gst-launch-stage-ladder.sh`
- `32-small-graph-hosting-ladder-matrix.sh`

Each should be stage-parameterized and save results under ticket-local `scripts/results/`.

### Slice 3: validate the stage controls before the full run

- `gofmt` the Go harness
- `bash -n` all shell wrappers
- run one stage by hand in each host to catch quoting / EOS / output bugs

### Slice 4: run the full ladder matrix

For each host and each stage:

- run for a fixed duration
- sample CPU with `pidstat`
- collect `perf stat` threadgroup counters
- save summaries and manifest rows

### Slice 5: interpret the first divergence point

Primary questions:

- At what stage does Go first diverge materially from Python and `gst-launch`?
- Does the page-fault delta appear at the same stage as the CPU delta?
- Is the first big divergence before encoding, at encoding, or only at mux/file output?

### Slice 6: decide the next targeted follow-up based on the ladder

Possible next actions:

- If divergence starts at `capture` or `convert`: inspect source/conversion pool behavior and memory reuse earlier in the graph.
- If divergence starts at `rate-caps`: look at shaped raw-frame negotiation and allocator behavior around the I420 pacing boundary.
- If divergence starts at `encode`: prioritize x264-input memory behavior, buffer-pool differences, and cgo/build-flag controls.
- If divergence starts only at `mux-file`: investigate output/finalization path and muxer behavior more closely.

## Meaningful Divergence Rules

To keep interpretation honest, use these heuristics:

- A stage is a **candidate first divergence point** if Go avg CPU is materially above both Python and `gst-launch` while earlier stages remain comparatively close.
- A stage becomes a **strong first divergence point** if:
  - Go CPU jumps meaningfully relative to the previous stage,
  - and Go page faults also jump materially relative to Python and `gst-launch`,
  - and that pattern persists for the remaining later stages.
- Do not over-interpret one noisy sample if the next later stage reverses the ranking unexpectedly. In that case, rerun the suspicious stage pair before drawing a conclusion.

## Open Questions

- Will Go already diverge at raw capture, or only when conversion/encoding begins?
- Will page faults track the same stage as CPU, or appear one stage earlier?
- Does Python stay close to `gst-launch` all the way down the ladder, or only on the full graph?
- Will the Go-vs-Python difference be large enough on smaller stages to justify additional cgo/build-flag controls immediately?

## Expected Outcomes

The ladder should give one of four useful outcomes:

1. **Early divergence (`capture`/`convert`)**
   - pushes the investigation toward source/conversion/memory-pool behavior.
2. **Mid divergence (`rate-caps`)**
   - points at shaped raw-frame pacing / negotiation / allocator effects.
3. **Encoder divergence (`encode`)**
   - strongly supports x264-input memory behavior as the key boundary.
4. **Late divergence (`mux-file`)**
   - shifts attention toward mux/output/finalization behavior.

Any of those outcomes is more actionable than the current full-graph suspicion alone.
