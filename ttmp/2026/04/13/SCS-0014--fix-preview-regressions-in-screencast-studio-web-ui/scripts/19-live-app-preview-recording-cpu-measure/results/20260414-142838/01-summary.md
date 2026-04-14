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

- label: before-pre-adaptive
- repo: /tmp/scs-pre-adaptive-1554243
- server: http://127.0.0.1:7784
- avg_cpu: 188.27%
- max_cpu: 329.00%
- preview_id: preview-e7444805eb17
- video_path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142838/output/Measure Region.mov
- audio_path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142838/output/audio-mix.wav

## Video ffprobe

- codec_name=h264
- width=2880
- height=960
- avg_frame_rate=24/1
- duration=8.166667
- size=639902

## Audio ffprobe

- codec_name=pcm_s16le
- duration=8.000000
- size=1536044

## Screenshot hashes

```text
04db842d69e9f334aec25381507b24e7dbdcb82ecc2388bb269faa43d4fdee07  /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142838/pre.jpg
04db842d69e9f334aec25381507b24e7dbdcb82ecc2388bb269faa43d4fdee07  /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142838/during.jpg
f1b9bafec89e1688cec2882cc4463af9179921ae5f501a775c404eecc42b4142  /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142838/post.jpg
```

## Session finish payload

```json
{
  "session": {
    "active": false,
    "sessionId": "adaptive-preview-measure-before-pre-adaptive",
    "state": "finished",
    "reason": "max duration reached after 8s",
    "startedAt": "2026-04-14T14:28:45-04:00",
    "finishedAt": "2026-04-14T14:28:54-04:00",
    "warnings": [],
    "outputs": [
      {
        "kind": "video",
        "sourceId": "measure-region",
        "name": "Measure Region",
        "path": "/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142838/output/Measure Region.mov"
      },
      {
        "kind": "audio",
        "sourceId": "",
        "name": "audio-mix",
        "path": "/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142838/output/audio-mix.wav"
      }
    ],
    "logs": [
      {
        "timestamp": "2026-04-14T14:28:45-04:00",
        "processLabel": "Measure Region",
        "stream": "system",
        "message": "gstreamer pipeline started"
      },
      {
        "timestamp": "2026-04-14T14:28:45-04:00",
        "processLabel": "audio-mix",
        "stream": "system",
        "message": "gstreamer pipeline started"
      },
      {
        "timestamp": "2026-04-14T14:28:53-04:00",
        "processLabel": "Measure Region",
        "stream": "system",
        "message": "stopping gstreamer pipeline via EOS"
      },
      {
        "timestamp": "2026-04-14T14:28:53-04:00",
        "processLabel": "audio-mix",
        "stream": "system",
        "message": "stopping gstreamer pipeline via EOS"
      }
    ]
  }
}```
