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
- avg_cpu: 29.17%
- max_cpu: 30.00%
- counters:
  scenario=normalized-fakesink
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233741/normalized-fakesink.mp4
  samples_pulled=0
  buffers_copied=0
  enqueued=0
  dropped=0
  worker_handled=0
  appsrc_pushed=0

## appsink-discard
- scenario: appsink-discard
- avg_cpu: 29.33%
- max_cpu: 31.00%
- counters:
  scenario=appsink-discard
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233741/appsink-discard.mp4
  samples_pulled=145
  buffers_copied=0
  enqueued=0
  dropped=0
  worker_handled=0
  appsrc_pushed=0

## appsink-copy-discard
- scenario: appsink-copy-discard
- avg_cpu: 28.33%
- max_cpu: 30.00%
- counters:
  scenario=appsink-copy-discard
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233741/appsink-copy-discard.mp4
  samples_pulled=145
  buffers_copied=145
  enqueued=0
  dropped=0
  worker_handled=0
  appsrc_pushed=0

## appsink-copy-async-discard
- scenario: appsink-copy-async-discard
- avg_cpu: 29.33%
- max_cpu: 33.00%
- counters:
  scenario=appsink-copy-async-discard
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233741/appsink-copy-async-discard.mp4
  samples_pulled=145
  buffers_copied=145
  enqueued=145
  dropped=0
  worker_handled=145
  appsrc_pushed=0

## appsink-copy-async-appsrc-fakesink
- scenario: appsink-copy-async-appsrc-fakesink
- avg_cpu: 31.00%
- max_cpu: 33.00%
- counters:
  scenario=appsink-copy-async-appsrc-fakesink
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233741/appsink-copy-async-appsrc-fakesink.mp4
  samples_pulled=144
  buffers_copied=144
  enqueued=144
  dropped=0
  worker_handled=144
  appsrc_pushed=144

## appsink-copy-async-appsrc-x264
- scenario: appsink-copy-async-appsrc-x264
- avg_cpu: 89.67%
- max_cpu: 117.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233741/appsink-copy-async-appsrc-x264.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.041667
  size=881628
- counters:
  scenario=appsink-copy-async-appsrc-x264
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233741/appsink-copy-async-appsrc-x264.mp4
  samples_pulled=145
  buffers_copied=145
  enqueued=145
  dropped=0
  worker_handled=145
  appsrc_pushed=145

