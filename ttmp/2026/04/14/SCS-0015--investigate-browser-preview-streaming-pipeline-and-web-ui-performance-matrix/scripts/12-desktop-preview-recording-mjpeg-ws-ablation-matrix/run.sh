#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

REPO="${REPO:-$(cd -- "$SCRIPT_DIR/../../../../../../.." && pwd)}"
PORT_BASE="${PORT_BASE:-7830}"
SERVER_BASE="http://127.0.0.1"
DISPLAY_NAME="${DISPLAY:-:0}"
DURATION="${DURATION:-6}"
INTERVAL="${INTERVAL:-1}"
SCENARIOS="${SCENARIOS:-mjpeg-only:0 mjpeg-plus-ws:1}"
PIDSTAT_COUNT="$((DURATION + 3))"
SERVER_BIN="$(mktemp /tmp/scs-0015-mjpeg-ws-ablation-XXXXXX)"
chmod +x "$SERVER_BIN"

cleanup() {
  if [[ -n "${WS_PID:-}" ]] && kill -0 "$WS_PID" 2>/dev/null; then
    kill "$WS_PID" 2>/dev/null || true
    wait "$WS_PID" 2>/dev/null || true
  fi
  if [[ -n "${MJPEG_PID:-}" ]] && kill -0 "$MJPEG_PID" 2>/dev/null; then
    kill "$MJPEG_PID" 2>/dev/null || true
    wait "$MJPEG_PID" 2>/dev/null || true
  fi
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
session_id: "browser-preview-mjpeg-ws-ablation"
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

write_metric_deltas() {
  local raw_dir="$1"
  local out="$2"
  python3 - <<'PY' "$raw_dir" "$out"
from pathlib import Path
import sys
raw_dir = Path(sys.argv[1])
out_path = Path(sys.argv[2])
files = sorted(raw_dir.glob('*.prom'))
if len(files) < 2:
    out_path.write_text('not enough metric snapshots for delta calculation\n')
    raise SystemExit(0)

def parse(path):
    vals = {}
    for line in path.read_text().splitlines():
        if not line or line.startswith('#'):
            continue
        try:
            key, value = line.rsplit(' ', 1)
            vals[key] = float(value)
        except ValueError:
            pass
    return vals

first = parse(files[0])
last = parse(files[-1])
prefixes = (
    'screencast_studio_preview_http_',
    'screencast_studio_preview_frame_updates_total',
    'screencast_studio_eventhub_',
    'screencast_studio_websocket_',
)
lines = []
for key in sorted(k for k in last if k.startswith(prefixes)):
    if key in ('screencast_studio_eventhub_subscribers', 'screencast_studio_websocket_connections') or 'clients{' in key:
        lines.append(f'{key}\tlast={last[key]:.0f}')
    else:
        delta = last.get(key, 0.0) - first.get(key, 0.0)
        lines.append(f'{key}\tdelta={delta:.0f}')
out_path.write_text('\n'.join(lines) + ('\n' if lines else ''))
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
  local with_ws="$2"
  local index="$3"
  local port="$((PORT_BASE + index))"
  local server="$SERVER_BASE:$port"
  local ws_url="ws://127.0.0.1:$port/ws"
  local scenario_dir="$RUN_DIR/$name"
  local output_root="$scenario_dir/output"
  mkdir -p "$scenario_dir" "$output_root"
  MJPEG_PID=""
  WS_PID=""

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
  wait_for_screenshot "$server" "$preview_id" "$scenario_dir/pre.jpg"

  pidstat -u -p "$SERVER_PID" 1 "$PIDSTAT_COUNT" > "$scenario_dir/server.pidstat.log" &
  local pidstat_pid=$!
  DURATION_SECONDS="$DURATION" INTERVAL_SECONDS="$INTERVAL" METRICS_URL="$server/metrics" OUTPUT_DIR="$scenario_dir/metrics" \
    bash "$SCRIPT_DIR/../02-sample-preview-metrics.sh" > "$scenario_dir/metrics-dir.txt" &
  local metrics_pid=$!

  curl -fsS "$server/api/previews/$preview_id/mjpeg" > /dev/null 2> "$scenario_dir/mjpeg.stderr.log" &
  MJPEG_PID=$!
  sleep 1

  if [[ "$with_ws" == "1" ]]; then
    (
      cd "$REPO"
      go run ./ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/ws_client \
        --url "$ws_url" \
        --output "$scenario_dir/ws-client-summary.json"
    ) > "$scenario_dir/ws-client.stdout.log" 2> "$scenario_dir/ws-client.stderr.log" &
    WS_PID=$!
    sleep 1
  fi

  curl -fsS -X POST "$server/api/recordings/start" -H 'content-type: application/json' --data-binary @"$scenario_dir/record.json" > "$scenario_dir/record-start.json"
  wait_for_recording_finish "$server" "$scenario_dir/session-finish.json"

  if [[ -n "$WS_PID" ]] && kill -0 "$WS_PID" 2>/dev/null; then
    kill "$WS_PID" 2>/dev/null || true
    wait "$WS_PID" 2>/dev/null || true
  fi
  if [[ -n "$MJPEG_PID" ]] && kill -0 "$MJPEG_PID" 2>/dev/null; then
    kill "$MJPEG_PID" 2>/dev/null || true
    wait "$MJPEG_PID" 2>/dev/null || true
  fi
  wait "$metrics_pid" || true
  wait "$pidstat_pid" || true

  curl -fsS "$server/api/previews" > "$scenario_dir/previews-after.json"
  write_metric_deltas "$scenario_dir/metrics/raw" "$scenario_dir/metric-deltas.txt"

  local avg_cpu max_cpu video_path
  avg_cpu=$(parse_avg "$scenario_dir/server.pidstat.log")
  max_cpu=$(parse_max "$scenario_dir/server.pidstat.log")
  video_path=$(find "$output_root" -maxdepth 1 -type f \( -name '*.mov' -o -name '*.mp4' \) | head -n1 || true)
  if [[ -n "$video_path" ]]; then
    ffprobe -hide_banner -loglevel error -show_entries format=duration,size:stream=codec_name,width,height,avg_frame_rate -of default=noprint_wrappers=1 "$video_path" > "$scenario_dir/video.ffprobe.txt" || true
  fi

  cat > "$scenario_dir/01-summary.md" <<EOF
---
Title: 12 desktop preview recording mjpeg websocket ablation summary
Ticket: SCS-0015
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - performance
    - preview
    - browser
    - websocket
    - recording
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: One focused desktop preview+recording ablation scenario comparing MJPEG-only versus MJPEG-plus-websocket server load.
LastUpdated: 2026-04-14T17:40:00-04:00
WhatFor: Preserve a focused server-side ablation run for the desktop preview+recording browser hypothesis.
WhenToUse: Read when comparing the likely websocket/event contribution against the earlier browser-path findings.
---

# 12 desktop preview recording mjpeg websocket ablation summary

- scenario: ${name}
- with_ws: ${with_ws}
- server: ${server}
- avg_cpu: ${avg_cpu}%
- max_cpu: ${max_cpu}%
- preview_id: ${preview_id}
- metrics_dir: $(cat "$scenario_dir/metrics-dir.txt")
- video_path: ${video_path}

## Metric deltas

~~~text
EOF
  cat "$scenario_dir/metric-deltas.txt" >> "$scenario_dir/01-summary.md" || true
  cat >> "$scenario_dir/01-summary.md" <<'EOF'
~~~
EOF
  if [[ -f "$scenario_dir/ws-client-summary.json" ]]; then
    cat >> "$scenario_dir/01-summary.md" <<'EOF'

## WS client summary

```json
EOF
    cat "$scenario_dir/ws-client-summary.json" >> "$scenario_dir/01-summary.md"
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
  with_ws="${scenario##*:}"
  run_scenario "$name" "$with_ws" "$i"
  i=$((i + 1))
done

cat > "$RUN_DIR/01-summary.md" <<EOF
---
Title: 12 desktop preview recording mjpeg websocket ablation run summary
Ticket: SCS-0015
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - performance
    - preview
    - browser
    - websocket
    - recording
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Aggregate summary for the focused desktop preview+recording MJPEG-vs-websocket ablation matrix.
LastUpdated: 2026-04-14T17:40:00-04:00
WhatFor: Preserve the scenario directories for the focused server-side websocket ablation pass.
WhenToUse: Read before drilling into the per-scenario result directories.
---

# 12 desktop preview recording mjpeg websocket ablation run summary

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
