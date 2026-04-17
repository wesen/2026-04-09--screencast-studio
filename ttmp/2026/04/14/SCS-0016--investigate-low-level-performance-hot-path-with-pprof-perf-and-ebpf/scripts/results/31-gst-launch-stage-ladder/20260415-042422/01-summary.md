---
Title: 31 gst-launch stage ladder summary
Ticket: SCS-0016
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - backend
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Saved result summary for one gst-launch small-graph ladder run.
LastUpdated: 2026-04-15T04:10:00-04:00
WhatFor: Preserve CPU and output measurements for one gst-launch ladder stage.
WhenToUse: Read this when comparing one gst-launch ladder stage against Go and Python controls.
---

# 31 gst-launch stage ladder

- display: :0
- root: 2880x1920+0+0
- stage: encode
- fps: 24
- bitrate: 6920
- encoder: x264enc
- x264_speed_preset: 3
- x264_tune: 0
- x264_bframes: 0
- x264_trellis: true
- container: mov
- duration_seconds: 6
- avg_cpu: 128.11%
- max_cpu: 156.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/31-gst-launch-stage-ladder/20260415-042422/output.mov
- pipeline: ximagesrc display-name=:0 use-damage=false show-pointer=true ! videoconvert ! videorate ! video/x-raw,format=I420,framerate=24/1,pixel-aspect-ratio=1/1 ! x264enc bitrate=6920 bframes=0 tune=0 speed-preset=3 trellis=true ! fakesink sync=false
