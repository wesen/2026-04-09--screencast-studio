#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results/14-live-server-browser-phase-a-sample"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

SERVER_URL="${SERVER_URL:-http://127.0.0.1:7777}"
PORT="${PORT:-7777}"
LABEL="${LABEL:-phase-a-browser-scenario}"
DURATION="${DURATION:-8}"
INTERVAL="${INTERVAL:-1}"
PIDSTAT_COUNT="$((DURATION + 2))"
SERVER_PID="${SERVER_PID:-$(lsof -tiTCP:${PORT} -sTCP:LISTEN | head -n1)}"

if [[ -z "$SERVER_PID" ]]; then
  echo "failed to detect server pid on port ${PORT}" >&2
  exit 1
fi

printf 'label=%s\nserver_url=%s\nport=%s\nserver_pid=%s\nduration=%s\n' \
  "$LABEL" "$SERVER_URL" "$PORT" "$SERVER_PID" "$DURATION" > "$RUN_DIR/context.txt"

pidstat -u -p "$SERVER_PID" 1 "$PIDSTAT_COUNT" > "$RUN_DIR/server.pidstat.log" &
PIDSTAT_PID=$!
DURATION_SECONDS="$DURATION" INTERVAL_SECONDS="$INTERVAL" METRICS_URL="$SERVER_URL/metrics" OUTPUT_DIR="$RUN_DIR/metrics" \
  bash "$SCRIPT_DIR/13-sample-server-metrics.sh" > "$RUN_DIR/metrics-dir.txt" &
METRICS_PID=$!

for i in $(seq 1 "$DURATION"); do
  curl -fsS "$SERVER_URL/api/previews" > "$RUN_DIR/previews-${i}.json" 2>/dev/null || true
  curl -fsS "$SERVER_URL/api/recordings/current" > "$RUN_DIR/recordings-current-${i}.json" 2>/dev/null || true
  sleep "$INTERVAL"
done

wait "$METRICS_PID" || true
wait "$PIDSTAT_PID" || true
curl -fsS "$SERVER_URL/api/previews" > "$RUN_DIR/previews-final.json" 2>/dev/null || true
curl -fsS "$SERVER_URL/api/recordings/current" > "$RUN_DIR/recordings-current-final.json" 2>/dev/null || true

python3 - <<'PY' "$RUN_DIR/metrics/raw" "$RUN_DIR/metric-deltas.txt"
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
lines = []
prefixes = (
    'screencast_studio_preview_http_',
    'screencast_studio_preview_frame_updates_total',
    'screencast_studio_preview_frame_store_nanoseconds_total',
    'screencast_studio_preview_latest_frame_copy_nanoseconds_total',
    'screencast_studio_preview_state_publish_nanoseconds_total',
    'screencast_studio_eventhub_',
    'screencast_studio_websocket_',
)
for key in sorted(k for k in last if k.startswith(prefixes)):
    if key in ('screencast_studio_eventhub_subscribers', 'screencast_studio_websocket_connections') or 'clients{' in key:
        lines.append(f'{key}\tlast={last[key]:.0f}')
    else:
        delta = last.get(key, 0.0) - first.get(key, 0.0)
        lines.append(f'{key}\tdelta={delta:.0f}')
out_path.write_text('\n'.join(lines) + ('\n' if lines else ''))
PY

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
Title: 14 live server browser phase a sample summary
Ticket: SCS-0016
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
Summary: One live browser-driven server-sampling run against the real Studio page while Phase A debug ablation flags are active.
LastUpdated: $(date --iso-8601=seconds)
WhatFor: Preserve pidstat, metrics, and preview/recording API snapshots for the real browser Phase A perturbation run.
WhenToUse: Read when comparing Phase A ablation against earlier unperturbed real browser runs.
---

# 14 live server browser phase a sample summary

- label: ${LABEL}
- server_url: ${SERVER_URL}
- server_pid: ${SERVER_PID}
- avg_cpu: ${AVG_CPU}%
- max_cpu: ${MAX_CPU}%
- metrics_dir: $(cat "$RUN_DIR/metrics-dir.txt")

## Metric deltas

~~~text
EOF
cat "$RUN_DIR/metric-deltas.txt" >> "$RUN_DIR/01-summary.md" || true
cat >> "$RUN_DIR/01-summary.md" <<'EOT'
~~~

## Files

- server.pidstat.log
- metrics/
- metric-deltas.txt
- previews-*.json
- recordings-current-*.json
EOT

echo "$RUN_DIR"
