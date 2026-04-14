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

