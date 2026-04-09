# recorderd

Minimal Go + HTML/JS runtime for the simplified capture DSL.

## What it does

- Accepts a YAML DSL over HTTP.
- Validates and normalizes the config.
- Exposes one browser preview stream per enabled video source.
- Starts and stops background capture processes.
- Records one video file per enabled video source.
- Records one mixed audio file from all enabled audio sources.

## Notes

- Preview is MJPEG over HTTP.
- Each preview request starts its own `ffmpeg` preview process.
- Capture start launches one `ffmpeg` process per video source and one for mixed audio.
- `follow_resize`, `noise_gate`, and `denoise` are accepted but only surfaced as warnings in this minimal implementation.
- Browser preview is video-only. Audio is recorded, not previewed.

## Build

```bash
go build -o recorderd .
```

## Run

```bash
./recorderd
```

Then open `http://localhost:8080/`.

## Requirements

- `ffmpeg` in `PATH`
- Python 3 with `PyYAML` for YAML parsing
- X11 for `display`, `window`, and `region`
- V4L2 devices for `camera`
- PulseAudio sources for `audio_sources`
