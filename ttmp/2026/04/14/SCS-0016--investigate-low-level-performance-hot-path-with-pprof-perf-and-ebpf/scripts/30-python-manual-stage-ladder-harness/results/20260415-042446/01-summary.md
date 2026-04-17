---
Title: 30 python manual stage ladder harness summary
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
Summary: Saved result summary for one Python manual stage-ladder run.
LastUpdated: 2026-04-15T04:10:00-04:00
WhatFor: Preserve CPU and output measurements for one Python manual small-graph ladder stage.
WhenToUse: Read this when comparing one Python manual ladder stage against Go and gst-launch controls.
---

# 30 python manual stage ladder harness

- display: :0
- root: 2880x1920+0+0
- stage: encode
- fps: 24
- bitrate: 6920
- encoder: x264enc
- x264_speed_preset: 3
- x264_tune: 4
- x264_bframes: 0
- x264_trellis: false
- container: mov
- duration_seconds: 6
- avg_cpu: 139.33%
- max_cpu: 154.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/30-python-manual-stage-ladder-harness/results/20260415-042446/output.mov
- dot_dir: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/30-python-manual-stage-ladder-harness/results/20260415-042446/dot
