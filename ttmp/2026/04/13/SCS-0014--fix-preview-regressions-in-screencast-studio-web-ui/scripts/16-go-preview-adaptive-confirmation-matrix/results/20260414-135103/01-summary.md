---
Title: 16 preview adaptive confirmation matrix run summary
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
Summary: Saved per-run summary for the standalone benchmark that confirms preview-profile and preview-ordering mitigations independently of the main runtime.
LastUpdated: 2026-04-14T14:05:00-04:00
WhatFor: Preserve the CPU and counter summary for adaptive preview mitigation scenarios in an isolated harness.
WhenToUse: Read when deciding whether recording-time preview degradation and rate-first preview ordering are worth implementing in the real runtime.
---

# 16 preview adaptive confirmation matrix

## recorder-only
- scenario: recorder-only
- avg_cpu: 86.33%
- max_cpu: 117.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/16-go-preview-adaptive-confirmation-matrix/results/20260414-135103/recorder-only.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.041667
  size=763269
- counters:
  scenario=recorder-only
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  preview_enabled=false
  preview_label=none
  preview_order=disabled
  preview_width=0
  preview_height=0
  preview_fps=0
  preview_quality=0
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/16-go-preview-adaptive-confirmation-matrix/results/20260414-135103/recorder-only.mp4
  preview_samples_pulled=0
  preview_frames_copied=0
  preview_bytes_copied=0
  recorder_samples_pulled=145
  recorder_buffers_copied=145
  recorder_enqueued=145
  recorder_dropped=0
  recorder_worker_handled=145
  recorder_appsrc_pushed=145

## preview-scale-first-current-plus-recorder
- scenario: preview-scale-first-current-plus-recorder
- avg_cpu: 91.83%
- max_cpu: 108.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/16-go-preview-adaptive-confirmation-matrix/results/20260414-135103/preview-scale-first-current-plus-recorder.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.000000
  size=657645
- counters:
  scenario=preview-scale-first-current-plus-recorder
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  preview_enabled=true
  preview_label=current
  preview_order=scale-first
  preview_width=1280
  preview_height=427
  preview_fps=10
  preview_quality=80
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/16-go-preview-adaptive-confirmation-matrix/results/20260414-135103/preview-scale-first-current-plus-recorder.mp4
  preview_samples_pulled=61
  preview_frames_copied=61
  preview_bytes_copied=16712742
  recorder_samples_pulled=144
  recorder_buffers_copied=144
  recorder_enqueued=144
  recorder_dropped=0
  recorder_worker_handled=144
  recorder_appsrc_pushed=144

## preview-rate-first-current-plus-recorder
- scenario: preview-rate-first-current-plus-recorder
- avg_cpu: 117.33%
- max_cpu: 183.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/16-go-preview-adaptive-confirmation-matrix/results/20260414-135103/preview-rate-first-current-plus-recorder.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.000000
  size=594990
- counters:
  scenario=preview-rate-first-current-plus-recorder
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  preview_enabled=true
  preview_label=current
  preview_order=rate-first
  preview_width=1280
  preview_height=427
  preview_fps=10
  preview_quality=80
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/16-go-preview-adaptive-confirmation-matrix/results/20260414-135103/preview-rate-first-current-plus-recorder.mp4
  preview_samples_pulled=61
  preview_frames_copied=61
  preview_bytes_copied=16687611
  recorder_samples_pulled=144
  recorder_buffers_copied=144
  recorder_enqueued=144
  recorder_dropped=0
  recorder_worker_handled=144
  recorder_appsrc_pushed=144

## preview-scale-first-constrained-plus-recorder
- scenario: preview-scale-first-constrained-plus-recorder
- avg_cpu: 102.83%
- max_cpu: 167.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/16-go-preview-adaptive-confirmation-matrix/results/20260414-135103/preview-scale-first-constrained-plus-recorder.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.000000
  size=643998
- counters:
  scenario=preview-scale-first-constrained-plus-recorder
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  preview_enabled=true
  preview_label=constrained
  preview_order=scale-first
  preview_width=640
  preview_height=213
  preview_fps=4
  preview_quality=50
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/16-go-preview-adaptive-confirmation-matrix/results/20260414-135103/preview-scale-first-constrained-plus-recorder.mp4
  preview_samples_pulled=25
  preview_frames_copied=25
  preview_bytes_copied=1340416
  recorder_samples_pulled=144
  recorder_buffers_copied=144
  recorder_enqueued=144
  recorder_dropped=0
  recorder_worker_handled=144
  recorder_appsrc_pushed=144

## preview-rate-first-constrained-plus-recorder
- scenario: preview-rate-first-constrained-plus-recorder
- avg_cpu: 90.50%
- max_cpu: 121.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/16-go-preview-adaptive-confirmation-matrix/results/20260414-135103/preview-rate-first-constrained-plus-recorder.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.041667
  size=647103
- counters:
  scenario=preview-rate-first-constrained-plus-recorder
  display=:0
  root=2880x1920
  region=0,960,2880,960
  fps=24
  duration=6s
  preview_enabled=true
  preview_label=constrained
  preview_order=rate-first
  preview_width=640
  preview_height=213
  preview_fps=4
  preview_quality=50
  output=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/16-go-preview-adaptive-confirmation-matrix/results/20260414-135103/preview-rate-first-constrained-plus-recorder.mp4
  preview_samples_pulled=25
  preview_frames_copied=25
  preview_bytes_copied=1339901
  recorder_samples_pulled=145
  recorder_buffers_copied=145
  recorder_enqueued=145
  recorder_dropped=0
  recorder_worker_handled=145
  recorder_appsrc_pushed=145

