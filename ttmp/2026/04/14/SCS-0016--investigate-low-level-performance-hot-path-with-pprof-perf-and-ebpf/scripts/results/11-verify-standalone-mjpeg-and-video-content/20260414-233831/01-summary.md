---
Title: 11 verify standalone mjpeg and video content summary
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
Summary: Saved standalone harness JPEG frames and extracted MP4 frames for content verification.
LastUpdated: 2026-04-14T23:38:44-04:00
WhatFor: Preserve image artifacts that let us verify the standalone harness is capturing visually correct content.
WhenToUse: Read when checking whether standalone MJPEG frames and recorded video frames look correct.
---

# 11 verify standalone mjpeg and video content summary

- http_addr: 127.0.0.1:7793
- mjpeg_url: http://127.0.0.1:7793/mjpeg
- output_path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/11-verify-standalone-mjpeg-and-video-content/20260414-233831/output.mp4
- warmup_seconds: 2
- record_seconds: 8

## Files

- harness.stdout.json
- harness.stderr.log
- healthz.json
- mjpeg-capture.json
- mjpeg-frame-01.jpg
- mjpeg-frame-10.jpg
- mjpeg-frame-20.jpg
- output.mp4
- ffprobe.txt
- video-frame-01.png
- video-frame-02.png
