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
- avg_cpu: 11.33%
- max_cpu: 12.00%

## recorder-only
- scenario: recorder-only
- avg_cpu: 91.00%
- max_cpu: 116.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/results/20260413-233717/recorder-only.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=5.958333
  size=926118

## preview-plus-recorder
- scenario: preview-plus-recorder
- avg_cpu: 151.00%
- max_cpu: 420.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/results/20260413-233717/preview-plus-recorder.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=6.291667
  size=869904

