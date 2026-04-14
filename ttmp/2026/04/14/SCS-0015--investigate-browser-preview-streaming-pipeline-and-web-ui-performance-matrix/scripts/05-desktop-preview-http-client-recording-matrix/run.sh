#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

REPO="${REPO:-$(cd -- "$SCRIPT_DIR/../../../../../../.." && pwd)}"
PORT_BASE="${PORT_BASE:-7810}"
SERVER_BASE="http://127.0.0.1"
DISPLAY_NAME="${DISPLAY:-:0}"
DURATION="${DURATION:-6}"
INTERVAL="${INTERVAL:-1}"
SCENARIOS="${SCENARIOS:-preview-no-client:0:0 preview-one-client:1:0 preview-two-clients:2:0 record-no-client:0:1 record-one-client:1:1 record-two-clients:2:1}"
PIDSTAT_COUNT="$((DURATION + 3))"
SERVER_BIN="$(mktemp /tmp/scs-0015-http-recording-matrix-XXXXXX)"
chmod +x "$SERVER_BIN"

cleanup() {
  if [[ -n "${SERVER_PID:-}" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
  rm -f "$SERVER_BIN"
}
trap cleanup EXIT

echo "repo=$REPO" > "$RUN_DIR/context.txt"
echo "display=$DISPLAY_NAME" >> "$RUN_DIR/context.txt"
echo "duration=$DURATION" >> "$RUN_DIR/context.txt"
echo "scenarios=$SCENARIOS" >> "$RUN_DIR/context.txt"
echo "server_bin=$SERVER_BIN" >> "$RUN_DIR/context.txt"

build_server() {
  (
    cd "$REPO"
    go build -o "$SERVER_BIN" ./cmd/screencast-studio
  )
}

wait_for_health() {
  local server="$1"
  local out="$2"
  for _ in $(seq 1 60); do
    if curl -fsS "$server/api/healthz" > "$out" 2>/dev/null; then
      return 0
    fi
    sleep 1
  done
  return 1
}

make_measure_yaml() {
  local out="$1"
  local output_root="$2"
  cat > "$out" <<YAML
schema: "recorder.config/v1"
session_id: "browser-preview-stream-recording-matrix"
destination_templates:
  per_source: "${output_root}/{source_name}.{ext}"
video_sources:
  - id: "desktop-1"
    name: "Full Desktop"
    type: "display"
    enabled: true
    target:
      display: "${DISPLAY_NAME}"
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
YAML
}

ensure_preview() {
  local server="$1"
  local scenario_dir="$2"
  python3 - <<'PY' "$scenario_dir"
import json, pathlib, sys
wd = pathlib.Path(sys.argv[1])
text = (wd / 'measure.yaml').read_text()
(wd / 'preview.json').write_text(json.dumps({'dsl': text, 'sourceId': 'desktop-1'}))
(wd / 'record.json').write_text(json.dumps({'dsl': text, 'maxDurationSeconds': 6}))
PY
  curl -fsS -X POST "$server/api/previews/ensure" -H 'content-type: application/json' --data-binary @"$scenario_dir/preview.json" > "$scenario_dir/preview-resp.json"
  python3 - <<'PY' "$scenario_dir"
import json, pathlib, sys
wd = pathlib.Path(sys.argv[1])
print(json.loads((wd / 'preview-resp.json').read_text())['preview']['id'])
PY
}

wait_for_screenshot() {
  local server="$1"
  local preview_id="$2"
  local out="$3"
  for _ in $(seq 1 30); do
    if curl -fsS "$server/api/previews/$preview_id/screenshot" -o "$out" 2>/dev/null; then
      return 0
    fi
    sleep 0.25
  done
  return 1
}

wait_for_recording_finish() {
  local server="$1"
  local out="$2"
  python3 - <<'PY' "$server" "$out"
import json, sys, time, urllib.request
server, out = sys.argv[1:3]
deadline = time.time() + 40
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
}

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

run_scenario() {
  local name="$1"
  local client_count="$2"
  local recording_enabled="$3"
  local index="$4"
  local port="$((PORT_BASE + index))"
  local server="$SERVER_BASE:$port"
  local scenario_dir="$RUN_DIR/$name"
  local output_root="$scenario_dir/output"
  mkdir -p "$scenario_dir" "$output_root"

  make_measure_yaml "$scenario_dir/measure.yaml" "$output_root"

  "$SERVER_BIN" serve --addr ":$port" > "$scenario_dir/server.stdout.log" 2> "$scenario_dir/server.stderr.log" &
  SERVER_PID=$!
  echo "$SERVER_PID" > "$scenario_dir/server.pid"

  if ! wait_for_health "$server" "$scenario_dir/healthz.json"; then
    echo "server did not become healthy for $name" >&2
    return 1
  fi

  local preview_id
  preview_id=$(ensure_preview "$server" "$scenario_dir")
  echo "$preview_id" > "$scenario_dir/preview.id"
  wait_for_screenshot "$server" "$preview_id" "$scenario_dir/pre.jpg" || true

  pidstat -u -p "$SERVER_PID" 1 "$PIDSTAT_COUNT" > "$scenario_dir/server.pidstat.log" &
  local pidstat_pid=$!
  DURATION_SECONDS="$DURATION" INTERVAL_SECONDS="$INTERVAL" METRICS_URL="$server/metrics" OUTPUT_DIR="$scenario_dir/metrics" \
    bash "$SCRIPT_DIR/../02-sample-preview-metrics.sh" > "$scenario_dir/metrics-dir.txt" &
  local metrics_pid=$!

  local client_pids=()
  for client_idx in $(seq 1 "$client_count"); do
    curl -fsS "$server/api/previews/$preview_id/mjpeg" > /dev/null 2> "$scenario_dir/client-${client_idx}.stderr.log" &
    client_pids+=("$!")
  done

  if [[ "$recording_enabled" == "1" ]]; then
    curl -fsS -X POST "$server/api/recordings/start" -H 'content-type: application/json' --data-binary @"$scenario_dir/record.json" > "$scenario_dir/record-start.json"
    wait_for_recording_finish "$server" "$scenario_dir/session-finish.json"
  else
    sleep "$DURATION"
  fi

  for pid in "${client_pids[@]:-}"; do
    if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
      wait "$pid" 2>/dev/null || true
    fi
  done
  wait "$metrics_pid" || true
  wait "$pidstat_pid" || true

  wait_for_screenshot "$server" "$preview_id" "$scenario_dir/post.jpg" || true
  curl -fsS "$server/api/previews" > "$scenario_dir/previews-after.json"

  local avg_cpu max_cpu video_path video_ffprobe
  avg_cpu=$(parse_avg "$scenario_dir/server.pidstat.log")
  max_cpu=$(parse_max "$scenario_dir/server.pidstat.log")
  video_path=$(find "$output_root" -maxdepth 1 -type f \( -name '*.mov' -o -name '*.mp4' \) | head -n1 || true)
  if [[ -n "$video_path" ]]; then
    ffprobe -hide_banner -loglevel error -show_entries format=duration,size:stream=codec_name,width,height,avg_frame_rate -of default=noprint_wrappers=1 "$video_path" > "$scenario_dir/video.ffprobe.txt" || true
  fi

  cat > "$scenario_dir/01-summary.md" <<EOF
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

- scenario: ${name}
- client_count: ${client_count}
- recording_enabled: ${recording_enabled}
- server: ${server}
- avg_cpu: ${avg_cpu}%
- max_cpu: ${max_cpu}%
- preview_id: ${preview_id}
- metrics_dir: $(cat "$scenario_dir/metrics-dir.txt")
- video_path: ${video_path}

## Files

- healthz.json
- preview-resp.json
- previews-after.json
- server.pidstat.log
- metrics/
EOF
if [[ -f "$scenario_dir/video.ffprobe.txt" ]]; then
cat >> "$scenario_dir/01-summary.md" <<'EOF'

## Video ffprobe

```text
EOF
cat "$scenario_dir/video.ffprobe.txt" >> "$scenario_dir/01-summary.md"
cat >> "$scenario_dir/01-summary.md" <<'EOF'
```
EOF
fi
if [[ -f "$scenario_dir/session-finish.json" ]]; then
cat >> "$scenario_dir/01-summary.md" <<'EOF'

## Session finish payload

```json
EOF
cat "$scenario_dir/session-finish.json" >> "$scenario_dir/01-summary.md"
cat >> "$scenario_dir/01-summary.md" <<'EOF'
```
EOF
fi

  kill "$SERVER_PID" 2>/dev/null || true
  wait "$SERVER_PID" 2>/dev/null || true
  SERVER_PID=""
}

build_server

i=0
for scenario in $SCENARIOS; do
  name="${scenario%%:*}"
  rest="${scenario#*:}"
  clients="${rest%%:*}"
  recording="${rest##*:}"
  run_scenario "$name" "$clients" "$recording" "$i"
  i=$((i + 1))
done

cat > "$RUN_DIR/01-summary.md" <<EOF
---
Title: 05 desktop preview http client recording matrix run summary
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
Summary: Aggregate summary for one desktop preview HTTP-client plus recording matrix run.
LastUpdated: 2026-04-14T16:40:00-04:00
WhatFor: Preserve the scenario directories and quick CPU summary for the desktop preview MJPEG-client plus recording matrix.
WhenToUse: Read before drilling into individual scenario result directories.
---

# 05 desktop preview http client recording matrix run summary

- run_dir: $RUN_DIR
- scenarios: $SCENARIOS
- duration: $DURATION
- display: $DISPLAY_NAME

## Scenario directories

EOF
for scenario in $SCENARIOS; do
  name="${scenario%%:*}"
  echo "- $name -> $RUN_DIR/$name" >> "$RUN_DIR/01-summary.md"
done

echo "$RUN_DIR"
