#!/usr/bin/env bash
set -euo pipefail

SERVER="${SERVER:-http://127.0.0.1:7777}"
WORK_DIR="${WORK_DIR:-/tmp/scs-preview-freeze-poll}"
mkdir -p "$WORK_DIR"

cat > "$WORK_DIR/freeze-test.yaml" <<'YAML'
schema: "recorder.config/v1"
session_id: "freeze-test"
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
text = (wd / 'freeze-test.yaml').read_text()
(wd / 'preview.json').write_text(json.dumps({'dsl': text, 'sourceId': 'edp-1-bottom-half'}))
(wd / 'record.json').write_text(json.dumps({'dsl': text}))
PY

curl -fsS -X POST "$SERVER/api/previews/ensure" -H 'content-type: application/json' --data-binary @"$WORK_DIR/preview.json" > "$WORK_DIR/preview-resp.json"
PREVIEW_ID=$(python3 - <<'PY' "$WORK_DIR"
import json, pathlib, sys
wd = pathlib.Path(sys.argv[1])
print(json.loads((wd / 'preview-resp.json').read_text())['preview']['id'])
PY
)

echo "preview_id=$PREVIEW_ID"
rm -f "$WORK_DIR/hashes.txt" "$WORK_DIR/frame-times.txt"

poll_phase() {
  local phase="$1"
  local count="$2"
  for i in $(seq 1 "$count"); do
    local ok=0
    for attempt in $(seq 1 10); do
      if curl -fsS "$SERVER/api/previews/$PREVIEW_ID/screenshot" -o "$WORK_DIR/$phase-shot.jpg" 2>/dev/null; then
        ok=1
        break
      fi
      sleep 0.5
    done
    if [ "$ok" -ne 1 ]; then
      echo "failed to fetch screenshot for $phase $i" >&2
      return 1
    fi
    printf "%s %s " "$phase" "$i" >> "$WORK_DIR/hashes.txt"
    sha256sum "$WORK_DIR/$phase-shot.jpg" >> "$WORK_DIR/hashes.txt"
    python3 - <<'PY' "$phase" "$i" "$PREVIEW_ID" "$SERVER" >> "$WORK_DIR/frame-times.txt"
import json, sys, urllib.request
phase, idx, preview_id, server = sys.argv[1:5]
with urllib.request.urlopen(server + '/api/previews') as r:
    data = json.load(r)
for p in data.get('previews', []):
    if p['id'] == preview_id:
        print(phase, idx, p.get('state'), p.get('lastFrameAt'), p.get('hasFrame'))
        break
PY
    sleep 1
  done
}

poll_phase pre 3
curl -fsS -X POST "$SERVER/api/recordings/start" -H 'content-type: application/json' --data-binary @"$WORK_DIR/record.json" > "$WORK_DIR/record-start.json"
poll_phase during 4
curl -fsS -X POST "$SERVER/api/recordings/stop" -H 'content-type: application/json' -d '{}' > "$WORK_DIR/record-stop.json" || true
sleep 2
poll_phase post 3

cat "$WORK_DIR/hashes.txt"
echo
cat "$WORK_DIR/frame-times.txt"
echo
echo "Artifacts: $WORK_DIR"
