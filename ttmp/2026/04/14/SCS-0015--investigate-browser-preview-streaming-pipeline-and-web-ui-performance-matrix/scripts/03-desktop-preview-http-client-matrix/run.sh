#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

REPO="${REPO:-$(cd -- "$SCRIPT_DIR/../../../../../../.." && pwd)}"
PORT_BASE="${PORT_BASE:-7790}"
SERVER_BASE="http://127.0.0.1"
DISPLAY_NAME="${DISPLAY:-:0}"
DURATION="${DURATION:-6}"
INTERVAL="${INTERVAL:-1}"
SCENARIOS="${SCENARIOS:-no-client:0 one-client:1 two-clients:2}"
PIDSTAT_COUNT="$((DURATION + 2))"

cleanup() {
  if [[ -n "${SERVER_PID:-}" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT

mkdir -p "$RUN_DIR"

echo "repo=$REPO" > "$RUN_DIR/context.txt"
echo "display=$DISPLAY_NAME" >> "$RUN_DIR/context.txt"
echo "duration=$DURATION" >> "$RUN_DIR/context.txt"
echo "scenarios=$SCENARIOS" >> "$RUN_DIR/context.txt"

build_server() {
  (
    cd "$REPO"
    go build -o "$RUN_DIR/server-bin" ./cmd/screencast-studio
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
  cat > "$out" <<YAML
schema: "recorder.config/v1"
session_id: "browser-preview-stream-matrix"
destination_templates:
  per_source: "recordings/{session_id}/{source_name}.{ext}"
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
  local index="$3"
  local port="$((PORT_BASE + index))"
  local server="$SERVER_BASE:$port"
  local scenario_dir="$RUN_DIR/$name"
  mkdir -p "$scenario_dir"

  make_measure_yaml "$scenario_dir/measure.yaml"

  "$RUN_DIR/server-bin" serve --addr ":$port" > "$scenario_dir/server.stdout.log" 2> "$scenario_dir/server.stderr.log" &
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

  sleep "$DURATION"

  for pid in "${client_pids[@]:-}"; do
    if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
      wait "$pid" 2>/dev/null || true
    fi
  done
  wait "$metrics_pid" || true
  wait "$pidstat_pid" || true

  curl -fsS "$server/api/previews" > "$scenario_dir/previews-after.json"

  local avg_cpu max_cpu
  avg_cpu=$(parse_avg "$scenario_dir/server.pidstat.log")
  max_cpu=$(parse_max "$scenario_dir/server.pidstat.log")

  cat > "$scenario_dir/01-summary.md" <<EOF
---
Title: 03 desktop preview http client matrix summary
Ticket: SCS-0015
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - performance
    - preview
    - browser
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: One desktop-preview HTTP client matrix scenario result for SCS-0015.
LastUpdated: 2026-04-14T16:10:00-04:00
WhatFor: Preserve one scenario result from the desktop preview HTTP client baseline matrix.
WhenToUse: Read when comparing 0/1/2 preview-stream clients against the same server-side desktop preview workload.
---

# 03 desktop preview http client matrix summary

- scenario: ${name}
- client_count: ${client_count}
- server: ${server}
- avg_cpu: ${avg_cpu}%
- max_cpu: ${max_cpu}%
- preview_id: ${preview_id}
- metrics_dir: $(cat "$scenario_dir/metrics-dir.txt")

## Files

- healthz.json
- preview-resp.json
- previews-after.json
- server.pidstat.log
- metrics/
EOF

  kill "$SERVER_PID" 2>/dev/null || true
  wait "$SERVER_PID" 2>/dev/null || true
  SERVER_PID=""
}

build_server

i=0
for scenario in $SCENARIOS; do
  name="${scenario%%:*}"
  clients="${scenario##*:}"
  run_scenario "$name" "$clients" "$i"
  i=$((i + 1))
done

cat > "$RUN_DIR/01-summary.md" <<EOF
---
Title: 03 desktop preview http client matrix run summary
Ticket: SCS-0015
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - performance
    - preview
    - browser
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Aggregate summary for one desktop preview HTTP client baseline matrix run.
LastUpdated: 2026-04-14T16:10:00-04:00
WhatFor: Preserve the scenario directories and quick CPU summary for a desktop preview MJPEG-client baseline run.
WhenToUse: Read before drilling into individual scenario result directories.
---

# 03 desktop preview http client matrix run summary

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
