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

- label: current
- repo: /home/manuel/code/wesen/2026-04-09--screencast-studio
- server: http://127.0.0.1:7781
- avg_cpu: 0.00%
- max_cpu: 0%
- preview_id: preview-e7444805eb17
- video_path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142718/output/Measure Region.mov
- audio_path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142718/output/audio-mix.wav

## Video ffprobe

- codec_name=h264
- width=2880
- height=960
- avg_frame_rate=24/1
- duration=8.125000
- size=1037820

## Audio ffprobe

- codec_name=pcm_s16le
- duration=8.000000
- size=1536044

## Screenshot hashes

```text
3ea0002e070ce47ec3fa7d7c055a5ae06b17830a9da4d9969a8ea87bacbdb51f  /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142718/pre.jpg
3ea0002e070ce47ec3fa7d7c055a5ae06b17830a9da4d9969a8ea87bacbdb51f  /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142718/during.jpg
a7eb39876e94a4bc51d1266359b024b143442862f31d597556bcd6dbd7d4cbe2  /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142718/post.jpg
```

## Session finish payload

```json
{
  "session": {
    "active": false,
    "sessionId": "adaptive-preview-measure-current",
    "state": "finished",
    "reason": "max duration reached after 8s",
    "startedAt": "2026-04-14T14:27:21-04:00",
    "finishedAt": "2026-04-14T14:27:30-04:00",
    "warnings": [],
    "outputs": [
      {
        "kind": "video",
        "sourceId": "measure-region",
        "name": "Measure Region",
        "path": "/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142718/output/Measure Region.mov"
      },
      {
        "kind": "audio",
        "sourceId": "",
        "name": "audio-mix",
        "path": "/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142718/output/audio-mix.wav"
      }
    ],
    "logs": [
      {
        "timestamp": "2026-04-14T14:27:21-04:00",
        "processLabel": "Measure Region",
        "stream": "system",
        "message": "gstreamer pipeline started"
      },
      {
        "timestamp": "2026-04-14T14:27:21-04:00",
        "processLabel": "audio-mix",
        "stream": "system",
        "message": "gstreamer pipeline started"
      },
      {
        "timestamp": "2026-04-14T14:27:29-04:00",
        "processLabel": "Measure Region",
        "stream": "system",
        "message": "stopping gstreamer pipeline via EOS"
      },
      {
        "timestamp": "2026-04-14T14:27:29-04:00",
        "processLabel": "audio-mix",
        "stream": "system",
        "message": "stopping gstreamer pipeline via EOS"
      }
    ]
  }
}```
