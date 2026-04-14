---
Title: 11 shared vs bridge reconciliation matrix summary
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - video
    - performance
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Unified same-session rerun of the direct GStreamer, shared-runtime, and staged bridge-overhead benchmark suites for reconciliation.
LastUpdated: 2026-04-13T23:50:00-04:00
WhatFor: Compare the earlier benchmark suites in one same-session run so their differences can be interpreted more confidently.
WhenToUse: Read when reconciling the full shared-runtime CPU results against the staged bridge-overhead results.
---

# 11 shared vs bridge reconciliation matrix

## Context

- display:       :0
- root: 2880 x 1920
- region: 0,960,2880,960
- fps: 24
- duration: 6s

## Child result directories

- 06 direct/pure GStreamer: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/06-gst-recording-performance-matrix/results/20260413-233847
- 07 shared runtime: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/results/20260413-233911
- 09 staged bridge overhead: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233935

## 06 summary

---
Title: 06 gst-launch recording performance matrix run summary
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - video
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Saved per-run summary for the standalone pure-GStreamer recording performance matrix.
LastUpdated: 2026-04-13T23:12:00-04:00
WhatFor: Preserve the CPU and ffprobe summary for one standalone gst-launch performance run.
WhenToUse: Read when reviewing the raw results under this run directory.
---

# 06 gst-launch recording performance matrix

## capture-to-fakesink
- avg_cpu: 25.83%
- max_cpu: 27.00%

## preview-like-jpeg
- avg_cpu: 9.83%
- max_cpu: 12.00%

## direct-record-current
- avg_cpu: 94.33%
- max_cpu: 97.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/06-gst-recording-performance-matrix/results/20260413-233847/direct-record-current.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=5.958333
  size=1494239

## direct-record-ultrafast
- avg_cpu: 55.33%
- max_cpu: 59.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/06-gst-recording-performance-matrix/results/20260413-233847/direct-record-ultrafast.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=5.958333
  size=1545553


## 07 summary

---
Title: 07 go shared recording performance matrix run summary
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - video
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Saved per-run summary for the standalone Go shared-source recording performance matrix.
LastUpdated: 2026-04-13T23:12:00-04:00
WhatFor: Preserve the CPU and ffprobe summary for one standalone shared-bridge performance run.
WhenToUse: Read when reviewing the raw results under this run directory.
---

# 07 go shared recording performance matrix

## preview-only
- scenario: preview-only
- avg_cpu: 13.17%
- max_cpu: 16.00%

## recorder-only
- scenario: recorder-only
- avg_cpu: 94.00%
- max_cpu: 125.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/results/20260413-233911/recorder-only.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=5.958333
  size=880200

## preview-plus-recorder
- scenario: preview-plus-recorder
- avg_cpu: 131.00%
- max_cpu: 391.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/results/20260413-233911/preview-plus-recorder.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.291667
  size=859765


## 09 summary

---
Title: 09 go bridge overhead matrix run summary
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - video
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Saved per-run summary for the staged Go bridge overhead benchmark matrix.
LastUpdated: 2026-04-13T23:30:00-04:00
WhatFor: Preserve the CPU and counter summary for one staged bridge-overhead benchmark run.
WhenToUse: Read when reviewing the raw results under this run directory.
---

# 09 go bridge overhead matrix

## normalized-fakesink
- scenario: normalized-fakesink
- avg_cpu: 26.50%
- max_cpu: 30.00%
- counters:
  scenario=normalized-fakesink
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233935/normalized-fakesink.mp4
  samples_pulled=0
  buffers_copied=0
  enqueued=0
  dropped=0
  worker_handled=0
  appsrc_pushed=0

## appsink-discard
- scenario: appsink-discard
- avg_cpu: 27.29%
- max_cpu: 29.00%
- counters:
  scenario=appsink-discard
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233935/appsink-discard.mp4
  samples_pulled=145
  buffers_copied=0
  enqueued=0
  dropped=0
  worker_handled=0
  appsrc_pushed=0

## appsink-copy-discard
- scenario: appsink-copy-discard
- avg_cpu: 32.67%
- max_cpu: 36.00%
- counters:
  scenario=appsink-copy-discard
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233935/appsink-copy-discard.mp4
  samples_pulled=144
  buffers_copied=144
  enqueued=0
  dropped=0
  worker_handled=0
  appsrc_pushed=0

## appsink-copy-async-discard
- scenario: appsink-copy-async-discard
- avg_cpu: 29.50%
- max_cpu: 31.00%
- counters:
  scenario=appsink-copy-async-discard
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233935/appsink-copy-async-discard.mp4
  samples_pulled=145
  buffers_copied=145
  enqueued=145
  dropped=0
  worker_handled=145
  appsrc_pushed=0

## appsink-copy-async-appsrc-fakesink
- scenario: appsink-copy-async-appsrc-fakesink
- avg_cpu: 31.83%
- max_cpu: 34.00%
- counters:
  scenario=appsink-copy-async-appsrc-fakesink
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233935/appsink-copy-async-appsrc-fakesink.mp4
  samples_pulled=144
  buffers_copied=144
  enqueued=144
  dropped=0
  worker_handled=144
  appsrc_pushed=144

## appsink-copy-async-appsrc-x264
- scenario: appsink-copy-async-appsrc-x264
- avg_cpu: 91.50%
- max_cpu: 118.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233935/appsink-copy-async-appsrc-x264.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.041667
  size=858560
- counters:
  scenario=appsink-copy-async-appsrc-x264
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233935/appsink-copy-async-appsrc-x264.mp4
  samples_pulled=145
  buffers_copied=145
  enqueued=145
  dropped=0
  worker_handled=145
  appsrc_pushed=145


## Reconciliation highlights

The purpose of this run is not to create a new theory yet, but to rerun the previously disagreeing benchmark families in one same-session matrix.

### Key numbers from this run

- 06 direct-record-current avg CPU: **94.33%**
- 06 direct-record-ultrafast avg CPU: **55.33%**
- 07 recorder-only avg CPU: **94.00%**
- 07 preview-plus-recorder avg CPU: **131.00%**
- 09 appsink-copy-async-appsrc-x264 avg CPU: **91.50%**
- 09 normalized-fakesink avg CPU: **26.50%**

### Why this comparison matters

- 06 tells us the cost of direct GStreamer capture and encode without the app's shared runtime.
- 07 tells us the cost of the current real shared-runtime recording path.
- 09 tells us the cost of staged bridge components in isolation.

If 07 stays much higher than both 06 direct encode and 09 staged bridge+x264, then the remaining discrepancy is likely coming from the full shared-runtime path rather than from x264 alone or from the minimal appsink/appsrc bridge alone.

