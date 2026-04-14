#!/usr/bin/env bash
set -euo pipefail

SERVER="${SERVER:-http://127.0.0.1:7777}"
WORK_DIR="${WORK_DIR:-/tmp/scs-api-region-preview}"
mkdir -p "$WORK_DIR"

cat > "$WORK_DIR/region-test.yaml" <<'YAML'
schema: "recorder.config/v1"
session_id: "region-test"
destination_templates:
  audio_mix: "recordings/{session_id}/audio-mix.{ext}"
  per_source: "recordings/{session_id}/{source_name}.{ext}"
audio_defaults:
  output:
    codec: "pcm_s16le"
    sample_rate_hz: 48000
    channels: 2
audio_mix:
  destination_template: "audio_mix"
video_sources:
  - id: "desktop-1"
    name: "Full Desktop"
    type: "display"
    enabled: true
    target:
      display: ":0.0"
    settings:
      capture:
        fps: 24
        cursor: true
        follow_resize: false
      output:
        container: "mov"
        video_codec: "h264"
        quality: 75
    destination_template: "per_source"
  - id: "edp-1-bottom-half"
    name: "eDP-1 bottom half"
    type: "region"
    enabled: true
    target:
      display: ":0.0"
      rect:
        x: 0
        y: 960
        w: 2880
        h: 960
    settings:
      capture:
        fps: 24
        cursor: true
        follow_resize: false
      output:
        container: "mov"
        video_codec: "h264"
        quality: 75
    destination_template: "per_source"
audio_sources:
  - id: "mic-1"
    name: "Built-in Mic"
    device: "default"
    enabled: true
    settings:
      gain: 1
      noise_gate: false
      denoise: false
YAML

python3 - <<'PY' "$WORK_DIR"
import json, pathlib, sys
wd = pathlib.Path(sys.argv[1])
text = (wd / 'region-test.yaml').read_text()
(wd / 'desktop.json').write_text(json.dumps({'dsl': text, 'sourceId': 'desktop-1'}))
(wd / 'region.json').write_text(json.dumps({'dsl': text, 'sourceId': 'edp-1-bottom-half'}))
PY

curl -fsS -X POST "$SERVER/api/previews/ensure" -H 'content-type: application/json' --data-binary @"$WORK_DIR/desktop.json" > "$WORK_DIR/ensure-desktop.json"
curl -fsS -X POST "$SERVER/api/previews/ensure" -H 'content-type: application/json' --data-binary @"$WORK_DIR/region.json" > "$WORK_DIR/ensure-region.json"

DESKTOP_ID=$(python3 - <<'PY' "$WORK_DIR"
import json, pathlib, sys
wd = pathlib.Path(sys.argv[1])
print(json.loads((wd / 'ensure-desktop.json').read_text())['preview']['id'])
PY
)
REGION_ID=$(python3 - <<'PY' "$WORK_DIR"
import json, pathlib, sys
wd = pathlib.Path(sys.argv[1])
print(json.loads((wd / 'ensure-region.json').read_text())['preview']['id'])
PY
)

echo "desktop preview: $DESKTOP_ID"
echo "region preview:  $REGION_ID"

for id in "$DESKTOP_ID" "$REGION_ID"; do
  for i in $(seq 1 20); do
    if curl -fsS "$SERVER/api/previews/$id/screenshot" -o "$WORK_DIR/$id.jpg" 2>/dev/null; then
      break
    fi
    sleep 0.5
  done
  printf "%s -> " "$id"
  file "$WORK_DIR/$id.jpg" | sed 's#^.*: ##'
  sha256sum "$WORK_DIR/$id.jpg"
done

echo
echo "Artifacts: $WORK_DIR"
