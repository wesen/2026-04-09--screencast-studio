---
Title: 12 preview recorder interplay matrix run summary
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
Summary: Saved per-run summary for the standalone preview-plus-recorder interplay benchmark.
LastUpdated: 2026-04-14T00:05:00-04:00
WhatFor: Preserve the CPU and counter summary for preview-only, recorder-only, and combined preview+recorder scenarios.
WhenToUse: Read when investigating why preview+record together costs more than recorder-only.
---

# 12 preview recorder interplay matrix

## preview-current-only
- scenario: preview-current-only
- avg_cpu: 12.17%
- max_cpu: 14.00%
- counters:
  scenario=preview-current-only
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/results/20260414-070646/preview-current-only.mp4
  preview_frames=60
  preview_bytes_copied=17314800
  recorder_samples_pulled=0
  recorder_buffers_copied=0
  recorder_enqueued=0
  recorder_dropped=0
  recorder_worker_handled=0
  recorder_appsrc_pushed=0

## recorder-current-only
- scenario: recorder-current-only
- avg_cpu: 94.00%
- max_cpu: 120.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/results/20260414-070646/recorder-current-only.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.000000
  size=998173
- counters:
  scenario=recorder-current-only
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/results/20260414-070646/recorder-current-only.mp4
  preview_frames=0
  preview_bytes_copied=0
  recorder_samples_pulled=144
  recorder_buffers_copied=144
  recorder_enqueued=144
  recorder_dropped=0
  recorder_worker_handled=144
  recorder_appsrc_pushed=144

## preview-current-plus-recorder
- scenario: preview-current-plus-recorder
- avg_cpu: 188.43%
- max_cpu: 492.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/results/20260414-070646/preview-current-plus-recorder.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.000000
  size=983893
- counters:
  scenario=preview-current-plus-recorder
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/results/20260414-070646/preview-current-plus-recorder.mp4
  preview_frames=61
  preview_bytes_copied=17603380
  recorder_samples_pulled=144
  recorder_buffers_copied=144
  recorder_enqueued=144
  recorder_dropped=0
  recorder_worker_handled=144
  recorder_appsrc_pushed=144

## preview-cheap-plus-recorder
- scenario: preview-cheap-plus-recorder
- avg_cpu: 170.00%
- max_cpu: 427.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/results/20260414-070646/preview-cheap-plus-recorder.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.041667
  size=994280
- counters:
  scenario=preview-cheap-plus-recorder
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/results/20260414-070646/preview-cheap-plus-recorder.mp4
  preview_frames=31
  preview_bytes_copied=1637529
  recorder_samples_pulled=145
  recorder_buffers_copied=145
  recorder_enqueued=145
  recorder_dropped=0
  recorder_worker_handled=145
  recorder_appsrc_pushed=145

