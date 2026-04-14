---
Title: 05 desktop preview http client recording matrix summary
Ticket: SCS-0015
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - performance
    - preview
    - browser
    - recording
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: One desktop-preview HTTP-client recording matrix scenario result for SCS-0015.
LastUpdated: 2026-04-14T16:40:00-04:00
WhatFor: Preserve one scenario result from the desktop preview HTTP-client plus recording matrix.
WhenToUse: Read when comparing client fan-out and recording combinations before the real browser-tab matrix.
---

# 05 desktop preview http client recording matrix summary

- scenario: record-no-client
- client_count: 0
- recording_enabled: 1
- server: http://127.0.0.1:7813
- avg_cpu: 162.22%
- max_cpu: 463.00%
- preview_id: preview-909758a3cc65
- metrics_dir: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/05-desktop-preview-http-client-recording-matrix/results/20260414-163154/record-no-client/metrics
- video_path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/05-desktop-preview-http-client-recording-matrix/results/20260414-163154/record-no-client/output/Full Desktop.mov

## Files

- healthz.json
- preview-resp.json
- previews-after.json
- server.pidstat.log
- metrics/

## Video ffprobe

```text
codec_name=h264
width=2880
height=1920
avg_frame_rate=24/1
duration=6.291667
size=847011
```

## Session finish payload

```json
{
  "session": {
    "active": false,
    "sessionId": "browser-preview-stream-recording-matrix",
    "state": "finished",
    "reason": "max duration reached after 6s",
    "startedAt": "2026-04-14T16:32:29-04:00",
    "finishedAt": "2026-04-14T16:32:36-04:00",
    "warnings": [],
    "outputs": [
      {
        "kind": "video",
        "sourceId": "desktop-1",
        "name": "Full Desktop",
        "path": "/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/05-desktop-preview-http-client-recording-matrix/results/20260414-163154/record-no-client/output/Full Desktop.mov"
      }
    ],
    "logs": [
      {
        "timestamp": "2026-04-14T16:32:29-04:00",
        "processLabel": "Full Desktop",
        "stream": "system",
        "message": "gstreamer pipeline started"
      },
      {
        "timestamp": "2026-04-14T16:32:35-04:00",
        "processLabel": "Full Desktop",
        "stream": "system",
        "message": "stopping gstreamer pipeline via EOS"
      }
    ]
  }
}```
