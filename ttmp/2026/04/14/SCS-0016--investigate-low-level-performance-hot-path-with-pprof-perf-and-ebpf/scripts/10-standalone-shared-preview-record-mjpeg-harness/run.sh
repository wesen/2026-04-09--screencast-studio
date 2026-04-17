#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
RESULTS_ROOT="$SCRIPT_DIR/results"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

HTTP_ADDR="${HTTP_ADDR:-127.0.0.1:7791}"
SOURCE_TYPE="${SOURCE_TYPE:-display}"
DISPLAY_NAME="${DISPLAY_NAME:-${DISPLAY:-}}"
WINDOW_ID="${WINDOW_ID:-}"
DEVICE="${DEVICE:-}"
RECT="${RECT:-}"
FPS="${FPS:-24}"
PREVIEW_SIZE="${PREVIEW_SIZE:-960x640}"
CONTAINER="${CONTAINER:-mp4}"
QUALITY="${QUALITY:-70}"
WARMUP_SECONDS="${WARMUP_SECONDS:-2}"
RECORD_SECONDS="${RECORD_SECONDS:-8}"
SPAWN_CLIENT="${SPAWN_CLIENT:-true}"
HARNESS_OUTPUT_PATH="${HARNESS_OUTPUT_PATH:-$RUN_DIR/output.mp4}"
PIDSTAT_COUNT="$((WARMUP_SECONDS + RECORD_SECONDS + 4))"

cd "$REPO_ROOT"

HARNESS_BIN="$RUN_DIR/standalone-harness-bin"
go build -o "$HARNESS_BIN" ./ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/10-standalone-shared-preview-record-mjpeg-harness

"$HARNESS_BIN" \
  --mode harness \
  --http-addr "$HTTP_ADDR" \
  --source-type "$SOURCE_TYPE" \
  --display-name "$DISPLAY_NAME" \
  --window-id "$WINDOW_ID" \
  --device "$DEVICE" \
  --rect "$RECT" \
  --fps "$FPS" \
  --preview-size "$PREVIEW_SIZE" \
  --container "$CONTAINER" \
  --quality "$QUALITY" \
  --warmup-seconds "$WARMUP_SECONDS" \
  --record-seconds "$RECORD_SECONDS" \
  --spawn-client "$SPAWN_CLIENT" \
  --output-path "$HARNESS_OUTPUT_PATH" \
  > "$RUN_DIR/harness.stdout.json" \
  2> "$RUN_DIR/harness.stderr.log" &
HARNESS_PID=$!

echo "$HARNESS_PID" > "$RUN_DIR/harness.pid"
pidstat -u -p "$HARNESS_PID" 1 "$PIDSTAT_COUNT" > "$RUN_DIR/harness.pidstat.log" &
PIDSTAT_PID=$!

wait "$HARNESS_PID"
HARNESS_STATUS=$?
wait "$PIDSTAT_PID" || true

python3 - <<'PY' "$RUN_DIR/harness.pidstat.log" "$RUN_DIR/harness.stdout.json" "$RUN_DIR/01-summary.md" "$HARNESS_STATUS"
import json
import sys
from pathlib import Path

pidstat = Path(sys.argv[1]).read_text().splitlines()
summary = json.loads(Path(sys.argv[2]).read_text())
out = Path(sys.argv[3])
status = int(sys.argv[4])

samples = []
for line in pidstat:
    parts = line.split()
    if len(parts) >= 9 and parts[2].isdigit():
        try:
            samples.append(float(parts[8]))
        except ValueError:
            pass
avg_cpu = (sum(samples) / len(samples)) if samples else 0.0
max_cpu = max(samples) if samples else 0.0

out.write_text(f"""---
Title: 10 standalone shared preview record mjpeg harness summary
Ticket: SCS-0016
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - backend
    - analysis
    - perf
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Standalone harness run for shared source + preview to MJPEG HTTP + recording without reusing the web server.
LastUpdated: {summary['finished_at']}
WhatFor: Preserve standalone near-web harness evidence for comparing native/shared path cost against the full web server path.
WhenToUse: Read when comparing this standalone harness against server/browser-backed captures.
---

# 10 standalone shared preview record mjpeg harness summary

- exit_status: {status}
- source_type: {summary.get('source_type')}
- http_addr: {summary.get('http_addr')}
- mjpeg_url: {summary.get('mjpeg_url')}
- output_path: {summary.get('output_path')}
- client_enabled: {summary.get('client_enabled')}
- warmup_seconds: {summary.get('warmup_seconds')}
- record_seconds: {summary.get('record_seconds')}
- avg_cpu: {avg_cpu:.2f}%
- max_cpu: {max_cpu:.2f}%
- preview_frames_seen: {summary.get('preview_frames_seen')}
- preview_bytes_seen: {summary.get('preview_bytes_seen')}
- mjpeg_streams: {summary.get('mjpeg_streams')}
- mjpeg_frames_served: {summary.get('mjpeg_frames_served')}
- mjpeg_bytes_served: {summary.get('mjpeg_bytes_served')}
- recording_state: {summary.get('recording_state')}
- recording_reason: {summary.get('recording_reason')}
- error: {summary.get('error')}

## Files

- harness.stdout.json
- harness.stderr.log
- harness.pidstat.log
- output.mp4 (or configured output path)
""")
PY

echo "$RUN_DIR"
