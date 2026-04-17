#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results/16-gst-launch-direct-full-desktop-match-app"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

DISPLAY_NAME="${DISPLAY_NAME:-${DISPLAY:-:0}}"
FPS="${FPS:-24}"
QUALITY="${QUALITY:-75}"
BITRATE="${BITRATE:-$((1000 + (QUALITY - 1) * 80))}"
CONTAINER="${CONTAINER:-mov}"
DURATION_SECONDS="${DURATION_SECONDS:-8}"
EOS_WAIT_SECONDS="${EOS_WAIT_SECONDS:-20}"
OUTPUT_PATH="${OUTPUT_PATH:-$RUN_DIR/direct-full-desktop.${CONTAINER}}"

ROOT_GEOM="$( (DISPLAY="$DISPLAY_NAME" xwininfo -root 2>/dev/null | awk '/-geometry/ {print $2; exit}') || true )"
if [[ -z "$ROOT_GEOM" ]]; then
  ROOT_GEOM="$(DISPLAY="$DISPLAY_NAME" xdpyinfo 2>/dev/null | awk '/dimensions:/ {print $2; exit}')"
fi
if [[ -z "$ROOT_GEOM" ]]; then
  echo "failed to detect root geometry for DISPLAY=$DISPLAY_NAME" >&2
  exit 1
fi

MUX="qtmux"
case "${CONTAINER,,}" in
  ""|mov|qt) MUX="qtmux" ;;
  mp4) MUX="mp4mux" ;;
  *) echo "unsupported CONTAINER=$CONTAINER" >&2; exit 1 ;;
esac

PIPELINE="ximagesrc display-name=$DISPLAY_NAME use-damage=false show-pointer=true ! videoconvert ! videorate ! video/x-raw,format=I420,framerate=${FPS}/1,pixel-aspect-ratio=1/1 ! x264enc bitrate=${BITRATE} bframes=0 tune=zerolatency speed-preset=3 ! h264parse ! ${MUX} ! filesink location=${OUTPUT_PATH}"

cat > "$RUN_DIR/context.txt" <<EOF
DISPLAY_NAME=$DISPLAY_NAME
ROOT_GEOM=$ROOT_GEOM
FPS=$FPS
QUALITY=$QUALITY
BITRATE=$BITRATE
CONTAINER=$CONTAINER
DURATION_SECONDS=$DURATION_SECONDS
EOS_WAIT_SECONDS=$EOS_WAIT_SECONDS
OUTPUT_PATH=$OUTPUT_PATH
PIPELINE=$PIPELINE
EOF

bash -lc "exec gst-launch-1.0 -e $PIPELINE" > "$RUN_DIR/stdout.log" 2> "$RUN_DIR/stderr.log" &
GST_PID=$!
echo "$GST_PID" > "$RUN_DIR/gst.pid"

pidstat -u -p "$GST_PID" 1 "$((DURATION_SECONDS + 3))" > "$RUN_DIR/pidstat.log" &
PIDSTAT_PID=$!

sleep "$DURATION_SECONDS"
if kill -0 "$GST_PID" 2>/dev/null; then
  kill -INT "$GST_PID" 2>/dev/null || true
fi

WAIT_OK=0
for _ in $(seq 1 "$EOS_WAIT_SECONDS"); do
  if ! kill -0 "$GST_PID" 2>/dev/null; then
    WAIT_OK=1
    break
  fi
  sleep 1
done

if [[ "$WAIT_OK" -ne 1 ]] && kill -0 "$GST_PID" 2>/dev/null; then
  kill -TERM "$GST_PID" 2>/dev/null || true
  sleep 1
fi
if kill -0 "$GST_PID" 2>/dev/null; then
  kill -KILL "$GST_PID" 2>/dev/null || true
fi

wait "$GST_PID" || true
wait "$PIDSTAT_PID" || true

ffprobe -hide_banner -loglevel error \
  -show_entries format=duration,size:stream=codec_name,width,height,avg_frame_rate \
  -of default=noprint_wrappers=1 "$OUTPUT_PATH" > "$RUN_DIR/ffprobe.txt"

AVG_CPU=$(awk '$1=="Average:" {print $8; found=1} END {if (!found) print ""}' "$RUN_DIR/pidstat.log" | tail -n1)
if [[ -z "$AVG_CPU" ]]; then
  AVG_CPU=$(awk 'NF >= 11 && $3 ~ /^[0-9]+$/ && $9 ~ /^[0-9.]+$/ {sum += $9; n++} END {if (n) printf "%.2f", sum / n}' "$RUN_DIR/pidstat.log")
fi
MAX_CPU=$(awk 'NF >= 11 && $3 ~ /^[0-9]+$/ && $9 ~ /^[0-9.]+$/ {if ($9 + 0 > max) max = $9} END {if (max == "") max = 0; print max}' "$RUN_DIR/pidstat.log")

cat > "$RUN_DIR/01-summary.md" <<EOF
# 16 gst-launch direct full desktop match app

- display: $DISPLAY_NAME
- root: $ROOT_GEOM
- fps: $FPS
- quality: $QUALITY
- bitrate: $BITRATE
- container: $CONTAINER
- duration_seconds: $DURATION_SECONDS
- avg_cpu: ${AVG_CPU}%
- max_cpu: ${MAX_CPU}%
- output: $OUTPUT_PATH

## ffprobe

~~~text
$(cat "$RUN_DIR/ffprobe.txt")
~~~
EOF

echo "$RUN_DIR"
