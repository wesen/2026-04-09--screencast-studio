---
Title: 19 live app preview recording cpu measure summary
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - performance
    - preview
    - recording
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: One live app-path CPU measurement run for preview plus recording against a real screencast-studio server process.
LastUpdated: 2026-04-14T14:45:00-04:00
WhatFor: Preserve live server CPU and output evidence for a specific repo/revision under a fixed preview-plus-recording scenario.
WhenToUse: Read when comparing before/after app-path behavior across revisions.
---

# 19 live app preview recording cpu measure

- label: after-current
- repo: /home/manuel/code/wesen/2026-04-09--screencast-studio
- server: http://127.0.0.1:7783
- avg_cpu: 170.82%
- max_cpu: 324.00%
- preview_id: preview-e7444805eb17
- video_path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142808/output/Measure Region.mov
- audio_path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142808/output/audio-mix.wav

## Video ffprobe

- codec_name=h264
- width=2880
- height=960
- avg_frame_rate=24/1
- duration=8.208333
- size=620257

## Audio ffprobe

- codec_name=pcm_s16le
- duration=7.960000
- size=1528364

## Screenshot hashes

```text
b41c42e776c8607b904f88a6f25256ecee12f50e6badde9d519b32998df5c7b0  /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142808/pre.jpg
b41c42e776c8607b904f88a6f25256ecee12f50e6badde9d519b32998df5c7b0  /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142808/during.jpg
2f9964dd80b5f41ac8f8fb6aed41c865e9f0c63e42b3ae015928e405a222ca5c  /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142808/post.jpg
```

## Session finish payload

```json
{
  "session": {
    "active": false,
    "sessionId": "adaptive-preview-measure-after-current",
    "state": "finished",
    "reason": "max duration reached after 8s",
    "startedAt": "2026-04-14T14:28:13-04:00",
    "finishedAt": "2026-04-14T14:28:21-04:00",
    "warnings": [],
    "outputs": [
      {
        "kind": "video",
        "sourceId": "measure-region",
        "name": "Measure Region",
        "path": "/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142808/output/Measure Region.mov"
      },
      {
        "kind": "audio",
        "sourceId": "",
        "name": "audio-mix",
        "path": "/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142808/output/audio-mix.wav"
      }
    ],
    "logs": [
      {
        "timestamp": "2026-04-14T14:28:13-04:00",
        "processLabel": "Measure Region",
        "stream": "system",
        "message": "gstreamer pipeline started"
      },
      {
        "timestamp": "2026-04-14T14:28:13-04:00",
        "processLabel": "audio-mix",
        "stream": "system",
        "message": "gstreamer pipeline started"
      },
      {
        "timestamp": "2026-04-14T14:28:21-04:00",
        "processLabel": "Measure Region",
        "stream": "system",
        "message": "stopping gstreamer pipeline via EOS"
      },
      {
        "timestamp": "2026-04-14T14:28:21-04:00",
        "processLabel": "audio-mix",
        "stream": "system",
        "message": "stopping gstreamer pipeline via EOS"
      }
    ]
  }
}```
