```yaml
schema: recorder.config/v1

destination_templates:
  per_source: "{date}/{source_name}.{ext}"
  camera: "{date}/camera/{source_name}.{ext}"
  mixed_audio: "{date}/audio/mixed.{ext}"

screen_capture_defaults:
  capture:
    fps: 24
    cursor: true
    follow_resize: false
  output:
    container: mov
    video_codec: h264
    quality: 75

camera_capture_defaults:
  capture:
    fps: 24
    device_mode: auto
    mirror: false
  output:
    container: mov
    video_codec: h264
    quality: 75

audio_defaults:
  output:
    codec: pcm_s16le
    sample_rate_hz: 48000
    channels: stereo

video_sources:
  - name: Display 1
    type: display
    target:
      display: 1
    destination:
      template: per_source

  - name: Camera 2
    type: camera
    target:
      device: built_in_camera
    settings:
      capture:
        mirror: false
      output:
        quality: 80
    destination:
      template: camera

  - name: Region 9
    type: region
    target:
      display: 1
      rect:
        x: 0
        y: 0
        w: 1920
        h: 540
    settings:
      capture:
        cursor: false
    destination:
      template: per_source

  - name: Display 14
    type: display
    target:
      display: 14
    destination:
      template: per_source

audio_sources:
  - name: Built-in Mic
    device: built_in_mic
    settings:
      gain: 1.0
      noise_gate: false
    destination:
      template: mixed_audio
```

```yaml
schema: recorder.config/v1

destination_templates:
  recording: "/Recordings/{session_id}/{source_name}.{ext}"

screen_capture_defaults:
  capture:
    fps: 30
    cursor: true
  output:
    container: mp4
    video_codec: h264
    quality: 70

camera_capture_defaults:
  capture:
    fps: 30
    mirror: true
  output:
    container: mp4
    video_codec: h264
    quality: 85

audio_defaults:
  output:
    codec: aac
    sample_rate_hz: 48000
    channels: stereo

video_sources:
  - name: Main Window
    type: window
    target:
      window:
        match_title: "Firefox"
    destination:
      template: recording

  - name: Webcam
    type: camera
    target:
      device: usb_camera_1
    destination:
      template: recording

audio_sources:
  - name: Mic
    device: default
    settings:
      gain: 1.2
    destination:
      template: recording
```

```yaml
schema: recorder.config/v1

destination_templates:
  default: "{source_type}/{source_name}-{time}.{ext}"

screen_capture_defaults:
  capture: { fps: 24, cursor: true }
  output: { container: mov, video_codec: h264, quality: 75 }

camera_capture_defaults:
  capture: { fps: 24, mirror: false }
  output: { container: mov, video_codec: h264, quality: 75 }

audio_defaults:
  output: { codec: pcm_s16le, sample_rate_hz: 48000, channels: stereo }

video_sources:
  - name: Display 1
    type: display
    target: { display: 1 }
    destination: { template: default }

  - name: Top Half
    type: region
    target:
      display: 1
      preset: top_half
    destination: { template: default }

  - name: Cam
    type: camera
    target: { device: built_in }
    settings:
      capture: { mirror: true }
    destination: { template: default }

audio_sources:
  - name: Built-in Mic
    device: built_in_mic
    settings: { gain: 1.0 }
    destination: { template: default }
```

