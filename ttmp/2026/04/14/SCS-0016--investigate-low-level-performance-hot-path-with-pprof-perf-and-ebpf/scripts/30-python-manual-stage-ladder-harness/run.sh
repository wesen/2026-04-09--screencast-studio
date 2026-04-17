#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

DISPLAY_NAME="${DISPLAY_NAME:-${DISPLAY:-:0}}"
STAGE="${STAGE:-capture}"
FPS="${FPS:-24}"
BITRATE="${BITRATE:-6920}"
ENCODER="${ENCODER:-x264enc}"
X264_SPEED_PRESET="${X264_SPEED_PRESET:-3}"
X264_TUNE="${X264_TUNE:-4}"
X264_BFRAMES="${X264_BFRAMES:-0}"
X264_TRELLIS="${X264_TRELLIS:-true}"
CONTAINER="${CONTAINER:-mov}"
DURATION_SECONDS="${DURATION_SECONDS:-8}"
OUTPUT_PATH="${OUTPUT_PATH:-$RUN_DIR/output.${CONTAINER}}"
DOT_DIR="$RUN_DIR/dot"
PYTHON_BIN="${PYTHON_BIN:-$(command -v python3)}"

ROOT_GEOM=$( (DISPLAY="$DISPLAY_NAME" xwininfo -root 2>/dev/null | awk '/-geometry/ {print $2; exit}') || true )
if [[ -z "$ROOT_GEOM" ]]; then
  ROOT_GEOM=$(DISPLAY="$DISPLAY_NAME" xdpyinfo 2>/dev/null | awk '/dimensions:/ {print $2; exit}')
fi

"$PYTHON_BIN" "$SCRIPT_DIR/main.py" \
  --display-name "$DISPLAY_NAME" \
  --stage "$STAGE" \
  --fps "$FPS" \
  --bitrate "$BITRATE" \
  --encoder "$ENCODER" \
  --x264-speed-preset "$X264_SPEED_PRESET" \
  --x264-tune "$X264_TUNE" \
  --x264-bframes "$X264_BFRAMES" \
  --x264-trellis "$X264_TRELLIS" \
  --container "$CONTAINER" \
  --duration-seconds "$DURATION_SECONDS" \
  --output-path "$OUTPUT_PATH" \
  --dot-dir "$DOT_DIR" \
  > "$RUN_DIR/harness.stdout.json" \
  2> "$RUN_DIR/harness.stderr.log" &
PID=$!
echo "$PID" > "$RUN_DIR/harness.pid"
pidstat -u -p "$PID" 1 "$((DURATION_SECONDS + 3))" > "$RUN_DIR/harness.pidstat.log" &
PIDSTAT_PID=$!
sleep 1
if kill -0 "$PID" 2>/dev/null; then
  TIDS=$(ls "/proc/$PID/task" | paste -sd, -)
  echo "$TIDS" > "$RUN_DIR/tids.txt"
  perf stat -x, -e page-faults,minor-faults,major-faults,cycles,instructions,context-switches,cpu-migrations -t "$TIDS" -- sleep "$((DURATION_SECONDS + 1))" > "$RUN_DIR/perf-stat.stdout.log" 2> "$RUN_DIR/perf-stat.csv" || true
fi
wait "$PID" || true
wait "$PIDSTAT_PID" || true

AVG_CPU=$(awk '$1 ~ /^[0-9]/ && $(NF-1) ~ /^-?[0-9]+$/ {sum += $(NF-2); n++} END {if (n) printf "%.2f", sum / n; else printf "0.00"}' "$RUN_DIR/harness.pidstat.log")
MAX_CPU=$(awk '$1 ~ /^[0-9]/ && $(NF-1) ~ /^-?[0-9]+$/ {if ($(NF-2) + 0 > max) max = $(NF-2)} END {if (max == "") max = 0; printf "%.2f", max}' "$RUN_DIR/harness.pidstat.log")

cat > "$RUN_DIR/01-summary.md" <<EOF
---
Title: 30 python manual stage ladder harness summary
Ticket: SCS-0016
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - backend
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Saved result summary for one Python manual stage-ladder run.
LastUpdated: 2026-04-15T04:10:00-04:00
WhatFor: Preserve CPU and output measurements for one Python manual small-graph ladder stage.
WhenToUse: Read this when comparing one Python manual ladder stage against Go and gst-launch controls.
---

# 30 python manual stage ladder harness

- display: $DISPLAY_NAME
- root: $ROOT_GEOM
- stage: $STAGE
- fps: $FPS
- bitrate: $BITRATE
- encoder: $ENCODER
- x264_speed_preset: $X264_SPEED_PRESET
- x264_tune: $X264_TUNE
- x264_bframes: $X264_BFRAMES
- x264_trellis: $X264_TRELLIS
- container: $CONTAINER
- duration_seconds: $DURATION_SECONDS
- avg_cpu: ${AVG_CPU}%
- max_cpu: ${MAX_CPU}%
- output: $OUTPUT_PATH
- dot_dir: $DOT_DIR
EOF

if [[ -f "$OUTPUT_PATH" ]]; then
  ffprobe -hide_banner -loglevel error -show_entries format=duration,size:stream=codec_name,width,height,avg_frame_rate -of default=noprint_wrappers=1 "$OUTPUT_PATH" > "$RUN_DIR/ffprobe.txt" || true
  if [[ -s "$RUN_DIR/ffprobe.txt" ]]; then
    cat >> "$RUN_DIR/01-summary.md" <<EOF

## ffprobe

~~~text
$(cat "$RUN_DIR/ffprobe.txt")
~~~
EOF
  fi
fi

echo "$RUN_DIR"
