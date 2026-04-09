#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../../../../../.."

rm -rf ./ttmp/smoke-output

tmpfile="$(mktemp /tmp/screencast-smoke-XXXXXX.yaml)"
trap 'rm -f "$tmpfile"' EXIT

cat >"$tmpfile" <<'EOF'
schema: recorder.config/v1
session_id: smoke-test

destination_templates:
  per_source: "./ttmp/smoke-output/{source_id}.{ext}"

screen_capture_defaults:
  capture:
    fps: 2
    cursor: false
  output:
    container: mkv
    video_codec: h264
    quality: 60

video_sources:
  - id: smoke_region
    name: Smoke Region
    type: region
    target:
      display: ":0.0"
      rect:
        x: 0
        y: 0
        w: 320
        h: 240
    destination_template: per_source

audio_sources: []
EOF

echo "DSL file: $tmpfile"
go run ./cmd/screencast-studio record --file "$tmpfile" --duration 1 --output json
ls -l ./ttmp/smoke-output/smoke_region.mkv
