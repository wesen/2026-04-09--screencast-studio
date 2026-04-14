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

- 06 direct/pure GStreamer: display=:0
root=2880x1920+0+0
region=0,960,2880,960
fps=24
preview_fps=10
preview_width=1280
preview_height=427
## capture-to-fakesink
- avg_cpu: 23.17%
- max_cpu: 26.00%
## preview-like-jpeg
- avg_cpu: 11.67%
- max_cpu: 13.00%
## direct-record-current
- avg_cpu: 91.35%
- max_cpu: 93.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/06-gst-recording-performance-matrix/results/20260413-233653/direct-record-current.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=5.958333
  size=1471199
## direct-record-ultrafast
- avg_cpu: 61.83%
- max_cpu: 70.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/06-gst-recording-performance-matrix/results/20260413-233653/direct-record-ultrafast.mp4
  codec_name=h264
  width=2880
  height=960
  avg_frame_rate=24/1
  duration=5.958333
  size=1568566
Results written to /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/06-gst-recording-performance-matrix/results/20260413-233653
/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/06-gst-recording-performance-matrix/results/20260413-233653
- 07 shared runtime: display=:0
root=2880x1920+0+0
region=0,960,2880,960
fps=24
duration=6s
Results written to /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/results/20260413-233717
/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/results/20260413-233717
- 09 staged bridge overhead: display=:0
root=2880x1920
region=0,960,2880,960
fps=24
duration=6s
Results written to /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233741
/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233741

## 06 summary

