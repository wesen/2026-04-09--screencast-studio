---
Title: 14 preview branch ablation matrix run summary
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
Summary: Saved per-run summary for the standalone preview-branch ablation benchmark.
LastUpdated: 2026-04-14T00:20:00-04:00
WhatFor: Preserve the CPU and counter summary for isolated preview-branch variants while recording.
WhenToUse: Read when investigating which part of the preview branch amplifies recording cost the most.
---

# 14 preview branch ablation matrix

## recorder-only
- scenario: recorder-only
- avg_cpu: 125.33%
- max_cpu: 171.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/results/20260414-081748/recorder-only.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.000000
  size=514728
- counters:
  scenario=recorder-only
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  preview_mode=none
  preview_width=0
  preview_height=0
  preview_fps=0
  preview_quality=0
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/results/20260414-081748/recorder-only.mp4
  preview_samples_pulled=0
  preview_frames_copied=0
  preview_bytes_copied=0
  recorder_samples_pulled=144
  recorder_buffers_copied=144
  recorder_enqueued=144
  recorder_dropped=0
  recorder_worker_handled=144
  recorder_appsrc_pushed=144

## preview-fakesink-plus-recorder
- scenario: preview-fakesink-plus-recorder
- avg_cpu: 134.00%
- max_cpu: 176.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/results/20260414-081748/preview-fakesink-plus-recorder.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.000000
  size=508227
- counters:
  scenario=preview-fakesink-plus-recorder
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  preview_mode=fakesink
  preview_width=1280
  preview_height=427
  preview_fps=10
  preview_quality=80
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/results/20260414-081748/preview-fakesink-plus-recorder.mp4
  preview_samples_pulled=0
  preview_frames_copied=0
  preview_bytes_copied=0
  recorder_samples_pulled=144
  recorder_buffers_copied=144
  recorder_enqueued=144
  recorder_dropped=0
  recorder_worker_handled=144
  recorder_appsrc_pushed=144

## preview-jpeg-discard-plus-recorder
- scenario: preview-jpeg-discard-plus-recorder
- avg_cpu: 137.83%
- max_cpu: 182.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/results/20260414-081748/preview-jpeg-discard-plus-recorder.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.000000
  size=503496
- counters:
  scenario=preview-jpeg-discard-plus-recorder
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  preview_mode=jpeg-discard
  preview_width=1280
  preview_height=427
  preview_fps=10
  preview_quality=80
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/results/20260414-081748/preview-jpeg-discard-plus-recorder.mp4
  preview_samples_pulled=61
  preview_frames_copied=0
  preview_bytes_copied=0
  recorder_samples_pulled=144
  recorder_buffers_copied=144
  recorder_enqueued=144
  recorder_dropped=0
  recorder_worker_handled=144
  recorder_appsrc_pushed=144

## preview-raw-copy-plus-recorder
- scenario: preview-raw-copy-plus-recorder
- avg_cpu: 141.33%
- max_cpu: 197.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/results/20260414-081748/preview-raw-copy-plus-recorder.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.000000
  size=504788
- counters:
  scenario=preview-raw-copy-plus-recorder
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  preview_mode=raw-copy
  preview_width=1280
  preview_height=427
  preview_fps=10
  preview_quality=80
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/results/20260414-081748/preview-raw-copy-plus-recorder.mp4
  preview_samples_pulled=61
  preview_frames_copied=61
  preview_bytes_copied=133360640
  recorder_samples_pulled=144
  recorder_buffers_copied=144
  recorder_enqueued=144
  recorder_dropped=0
  recorder_worker_handled=144
  recorder_appsrc_pushed=144

## preview-current-plus-recorder
- scenario: preview-current-plus-recorder
- avg_cpu: 152.33%
- max_cpu: 204.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/results/20260414-081748/preview-current-plus-recorder.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.000000
  size=503548
- counters:
  scenario=preview-current-plus-recorder
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  preview_mode=current
  preview_width=1280
  preview_height=427
  preview_fps=10
  preview_quality=80
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/results/20260414-081748/preview-current-plus-recorder.mp4
  preview_samples_pulled=61
  preview_frames_copied=61
  preview_bytes_copied=12987154
  recorder_samples_pulled=144
  recorder_buffers_copied=144
  recorder_enqueued=144
  recorder_dropped=0
  recorder_worker_handled=144
  recorder_appsrc_pushed=144

## preview-cheap-plus-recorder
- scenario: preview-cheap-plus-recorder
- avg_cpu: 112.00%
- max_cpu: 143.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/results/20260414-081748/preview-cheap-plus-recorder.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.041667
  size=514334
- counters:
  scenario=preview-cheap-plus-recorder
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  preview_mode=cheap
  preview_width=640
  preview_height=213
  preview_fps=5
  preview_quality=50
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/results/20260414-081748/preview-cheap-plus-recorder.mp4
  preview_samples_pulled=31
  preview_frames_copied=31
  preview_bytes_copied=1364682
  recorder_samples_pulled=145
  recorder_buffers_copied=145
  recorder_enqueued=145
  recorder_dropped=0
  recorder_worker_handled=145
  recorder_appsrc_pushed=145

