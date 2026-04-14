#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

REPO="${REPO:-$(cd -- "$SCRIPT_DIR/../../../../../../.." && pwd)}"
LABEL="${LABEL:-current}"
PORT="${PORT:-7781}"
ADDR=":$PORT"
SERVER="http://127.0.0.1:$PORT"
DURATION="${DURATION:-8}"
DISPLAY_NAME="${DISPLAY:-:0}"
PIDSTAT_COUNT="$((DURATION + 3))"

cleanup() {
  if [[ -n "${SERVER_PID:-}" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

root_geometry() {
  local geom
  if geom=$(DISPLAY="$DISPLAY_NAME" xwininfo -root 2>/dev/null | awk '/-geometry/ {print $2; exit}'); then
    if [[ -n "$geom" ]]; then
      geom="${geom%%+*}"
      if [[ "$geom" == *x* ]]; then echo "$geom"; return 0; fi
    fi
  fi
  geom=$(DISPLAY="$DISPLAY_NAME" xdpyinfo 2>/dev/null | awk '/dimensions:/ {print $2; exit}')
  if [[ -z "$geom" || "$geom" != *x* ]]; then
    echo "failed to detect root geometry for DISPLAY=$DISPLAY_NAME" >&2
    exit 1
  fi
  echo "$geom"
}
parse_geom() { local geom="$1"; echo "${geom%x*} ${geom#*x}"; }
ROOT_GEOM="$(root_geometry)"
read -r ROOT_W ROOT_H <<<"$(parse_geom "$ROOT_GEOM")"
REGION="0,$((ROOT_H/2)),$ROOT_W,$((ROOT_H - ROOT_H/2))"
REGION_W="$ROOT_W"
REGION_H="$((ROOT_H - ROOT_H/2))"

printf 'label=%s\nrepo=%s\nport=%s\nserver=%s\ndisplay=%s\nroot=%sx%s\nregion=%s\nduration=%s\n' \
  "$LABEL" "$REPO" "$PORT" "$SERVER" "$DISPLAY_NAME" "$ROOT_W" "$ROOT_H" "$REGION" "$DURATION" \
  | tee "$RUN_DIR/context.txt"

OUTPUT_ROOT="$RUN_DIR/output"
mkdir -p "$OUTPUT_ROOT"
cat > "$RUN_DIR/measure.yaml" <<YAML
schema: "recorder.config/v1"
session_id: "adaptive-preview-measure-${LABEL}"
destination_templates:
  audio_mix: "${OUTPUT_ROOT}/audio-mix.{ext}"
  per_source: "${OUTPUT_ROOT}/{source_name}.{ext}"
audio_defaults:
  output:
    codec: "pcm_s16le"
    sample_rate_hz: 48000
    channels: 2
audio_mix:
  destination_template: "audio_mix"
video_sources:
  - id: "measure-region"
    name: "Measure Region"
    type: "region"
    enabled: true
    target:
      display: "${DISPLAY_NAME}"
      rect:
        x: 0
        y: $((ROOT_H/2))
        w: ${REGION_W}
        h: ${REGION_H}
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

python3 - <<'PY' "$RUN_DIR"
import json, pathlib, sys
wd = pathlib.Path(sys.argv[1])
text = (wd / 'measure.yaml').read_text()
(wd / 'preview.json').write_text(json.dumps({'dsl': text, 'sourceId': 'measure-region'}))
(wd / 'record.json').write_text(json.dumps({'dsl': text, 'maxDurationSeconds': 8}))
PY

(
  cd "$REPO"
  go build -o "$RUN_DIR/server-bin" ./cmd/screencast-studio
)
"$RUN_DIR/server-bin" serve --addr "$ADDR" > "$RUN_DIR/server.stdout.log" 2> "$RUN_DIR/server.stderr.log" &
SERVER_PID=$!
echo "$SERVER_PID" > "$RUN_DIR/server.pid"

for _ in $(seq 1 60); do
  if curl -fsS "$SERVER/api/healthz" > "$RUN_DIR/healthz.json" 2>/dev/null; then
    break
  fi
  sleep 1
done
if [[ ! -s "$RUN_DIR/healthz.json" ]]; then
  echo "server did not become healthy" >&2
  exit 1
fi

curl -fsS -X POST "$SERVER/api/previews/ensure" -H 'content-type: application/json' --data-binary @"$RUN_DIR/preview.json" > "$RUN_DIR/preview-resp.json"
PREVIEW_ID=$(python3 - <<'PY' "$RUN_DIR"
import json, pathlib, sys
wd = pathlib.Path(sys.argv[1])
print(json.loads((wd / 'preview-resp.json').read_text())['preview']['id'])
PY
)
echo "$PREVIEW_ID" > "$RUN_DIR/preview.id"

fetch_shot() {
  local phase="$1"
  for attempt in $(seq 1 20); do
    if curl -fsS "$SERVER/api/previews/$PREVIEW_ID/screenshot" -o "$RUN_DIR/${phase}.jpg" 2>/dev/null; then
      sha256sum "$RUN_DIR/${phase}.jpg" >> "$RUN_DIR/screenshot-hashes.txt"
      return 0
    fi
    sleep 0.25
  done
  return 1
}

fetch_shot pre

curl -fsS -X POST "$SERVER/api/recordings/start" -H 'content-type: application/json' --data-binary @"$RUN_DIR/record.json" > "$RUN_DIR/record-start.json"
pidstat -u -p "$SERVER_PID" 1 "$PIDSTAT_COUNT" > "$RUN_DIR/server.pidstat.log" &
PIDSTAT_PID=$!
fetch_shot during || true

python3 - <<'PY' "$SERVER" "$RUN_DIR/session-finish.json"
import json, sys, time, urllib.request
server, out = sys.argv[1:3]
deadline = time.time() + 25
last = None
while time.time() < deadline:
    with urllib.request.urlopen(server + '/api/recordings/current') as r:
        data = json.load(r)
    last = data
    session = data.get('session')
    if session and not session.get('active') and session.get('sessionId'):
        break
    time.sleep(0.2)
with open(out, 'w') as f:
    json.dump(last, f, indent=2)
if not last or not last.get('session') or last['session'].get('active'):
    raise SystemExit('recording did not finish in time')
PY

wait "$PIDSTAT_PID" || true
fetch_shot post || true
curl -fsS "$SERVER/api/previews" > "$RUN_DIR/previews-after.json"

VIDEO_PATH=$(find "$OUTPUT_ROOT" -maxdepth 1 -type f \( -name '*.mov' -o -name '*.mp4' \) | head -n1)
AUDIO_PATH=$(find "$OUTPUT_ROOT" -maxdepth 1 -type f \( -name '*.wav' -o -name '*.ogg' \) | head -n1)
if [[ -n "$VIDEO_PATH" ]]; then
  ffprobe -hide_banner -loglevel error -show_entries format=duration,size:stream=codec_name,width,height,avg_frame_rate -of default=noprint_wrappers=1 "$VIDEO_PATH" > "$RUN_DIR/video.ffprobe.txt" || true
fi
if [[ -n "$AUDIO_PATH" ]]; then
  ffprobe -hide_banner -loglevel error -show_entries format=duration,size:stream=codec_name -of default=noprint_wrappers=1 "$AUDIO_PATH" > "$RUN_DIR/audio.ffprobe.txt" || true
fi

parse_avg() {
  local file="$1"
  local avg
  avg=$(awk '$1=="Average:" {print $8; found=1} END {if (!found) print ""}' "$file" | tail -n1)
  if [[ -z "$avg" ]]; then
    avg=$(awk 'NF >= 11 && $3 ~ /^[0-9]+$/ && $9 ~ /^[0-9.]+$/ {sum += $9; n++} END {if (n) printf "%.2f", sum / n}' "$file")
  fi
  printf '%s' "$avg"
}
parse_max() {
  local file="$1"
  awk 'NF >= 11 && $3 ~ /^[0-9]+$/ && $9 ~ /^[0-9.]+$/ {if ($9 + 0 > max) max = $9} END {if (max == "") max = 0; print max}' "$file"
}
AVG_CPU=$(parse_avg "$RUN_DIR/server.pidstat.log")
MAX_CPU=$(parse_max "$RUN_DIR/server.pidstat.log")

cat > "$RUN_DIR/01-summary.md" <<EOF
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

- label: ${LABEL}
- repo: ${REPO}
- server: ${SERVER}
- avg_cpu: ${AVG_CPU}%
- max_cpu: ${MAX_CPU}%
- preview_id: ${PREVIEW_ID}
- video_path: ${VIDEO_PATH}
- audio_path: ${AUDIO_PATH}

## Video ffprobe

EOF
if [[ -f "$RUN_DIR/video.ffprobe.txt" ]]; then
  sed 's/^/- /' "$RUN_DIR/video.ffprobe.txt" >> "$RUN_DIR/01-summary.md"
fi
cat >> "$RUN_DIR/01-summary.md" <<'EOF'

## Audio ffprobe

EOF
if [[ -f "$RUN_DIR/audio.ffprobe.txt" ]]; then
  sed 's/^/- /' "$RUN_DIR/audio.ffprobe.txt" >> "$RUN_DIR/01-summary.md"
fi
cat >> "$RUN_DIR/01-summary.md" <<'EOF'

## Screenshot hashes

```text
EOF
cat "$RUN_DIR/screenshot-hashes.txt" >> "$RUN_DIR/01-summary.md" || true
cat >> "$RUN_DIR/01-summary.md" <<'EOF'
```

## Session finish payload

```json
EOF
cat "$RUN_DIR/session-finish.json" >> "$RUN_DIR/01-summary.md"
cat >> "$RUN_DIR/01-summary.md" <<'EOF'
```
EOF

echo "Results written to $RUN_DIR"
