#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
RESULTS_ROOT="$SCRIPT_DIR/results"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

DISPLAY_NAME="${DISPLAY_NAME:-${DISPLAY:-:0}}"
FPS="${FPS:-24}"
BITRATE="${BITRATE:-6920}"
CONTAINER="${CONTAINER:-mov}"
DURATION_SECONDS="${DURATION_SECONDS:-8}"
OUTPUT_PATH="${OUTPUT_PATH:-$RUN_DIR/manual-direct-full-desktop.${CONTAINER}}"
DOT_DIR="$RUN_DIR/dot"
BIN_PATH="$RUN_DIR/manual-direct-full-desktop-harness-bin"

cd "$REPO_ROOT"
go build -o "$BIN_PATH" ./ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/17-go-manual-direct-full-desktop-harness

"$BIN_PATH" \
  --display-name "$DISPLAY_NAME" \
  --fps "$FPS" \
  --bitrate "$BITRATE" \
  --container "$CONTAINER" \
  --duration-seconds "$DURATION_SECONDS" \
  --output-path "$OUTPUT_PATH" \
  --dot-dir "$DOT_DIR" \
  > "$RUN_DIR/harness.stdout.json" \
  2> "$RUN_DIR/harness.stderr.log" &
HARNESS_PID=$!
echo "$HARNESS_PID" > "$RUN_DIR/harness.pid"
pidstat -u -p "$HARNESS_PID" 1 "$((DURATION_SECONDS + 3))" > "$RUN_DIR/harness.pidstat.log" || true
wait "$HARNESS_PID" || true

ROOT_GEOM=$( (DISPLAY="$DISPLAY_NAME" xwininfo -root 2>/dev/null | awk '/-geometry/ {print $2; exit}') || true )
if [[ -z "$ROOT_GEOM" ]]; then
  ROOT_GEOM=$(DISPLAY="$DISPLAY_NAME" xdpyinfo 2>/dev/null | awk '/dimensions:/ {print $2; exit}')
fi

ffprobe -hide_banner -loglevel error \
  -show_entries format=duration,size:stream=codec_name,width,height,avg_frame_rate \
  -of default=noprint_wrappers=1 "$OUTPUT_PATH" > "$RUN_DIR/ffprobe.txt"

AVG_CPU=$(awk '$1=="Average:" {print $8; found=1} END {if (!found) print ""}' "$RUN_DIR/harness.pidstat.log" | tail -n1)
if [[ -z "$AVG_CPU" ]]; then
  AVG_CPU=$(awk 'NF >= 11 && $3 ~ /^[0-9]+$/ && $9 ~ /^[0-9.]+$/ {sum += $9; n++} END {if (n) printf "%.2f", sum / n}' "$RUN_DIR/harness.pidstat.log")
fi
MAX_CPU=$(awk 'NF >= 11 && $3 ~ /^[0-9]+$/ && $9 ~ /^[0-9.]+$/ {if ($9 + 0 > max) max = $9} END {if (max == "") max = 0; print max}' "$RUN_DIR/harness.pidstat.log")

cat > "$RUN_DIR/01-summary.md" <<EOF
# 17 go manual direct full desktop harness

- display: $DISPLAY_NAME
- root: $ROOT_GEOM
- fps: $FPS
- bitrate: $BITRATE
- container: $CONTAINER
- duration_seconds: $DURATION_SECONDS
- avg_cpu: ${AVG_CPU}%
- max_cpu: ${MAX_CPU}%
- output: $OUTPUT_PATH
- dot_dir: $DOT_DIR

## ffprobe

~~~text
$(cat "$RUN_DIR/ffprobe.txt")
~~~
EOF

echo "$RUN_DIR"
