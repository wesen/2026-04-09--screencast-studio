#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

DISPLAY_NAME="${DISPLAY:-:0}"
DURATION="${DURATION:-6}"
RECORD_DURATION="$((DURATION + 1))"
FPS="${FPS:-24}"
PREVIEW_FPS="${PREVIEW_FPS:-10}"
PREVIEW_WIDTH="${PREVIEW_WIDTH:-1280}"
PREVIEW_JPEG_QUALITY="${PREVIEW_JPEG_QUALITY:-80}"

root_geometry() {
  local geom
  if geom=$(DISPLAY="$DISPLAY_NAME" xwininfo -root 2>/dev/null | awk '/-geometry/ {print $2; exit}'); then
    if [[ -n "$geom" && "$geom" == *x* ]]; then
      echo "$geom"
      return 0
    fi
  fi
  geom=$(DISPLAY="$DISPLAY_NAME" xdpyinfo 2>/dev/null | awk '/dimensions:/ {print $2; exit}')
  if [[ -z "$geom" || "$geom" != *x* ]]; then
    echo "failed to detect root geometry for DISPLAY=$DISPLAY_NAME" >&2
    exit 1
  fi
  echo "$geom"
}

parse_geom() {
  local geom="$1"
  local w="${geom%x*}"
  local h="${geom#*x}"
  echo "$w $h"
}

ROOT_GEOM="$(root_geometry)"
read -r ROOT_W ROOT_H <<<"$(parse_geom "$ROOT_GEOM")"
DEFAULT_REGION="0,$((ROOT_H/2)),$ROOT_W,$((ROOT_H - ROOT_H/2))"
REGION="${REGION:-$DEFAULT_REGION}"
IFS=',' read -r REGION_X REGION_Y REGION_W REGION_H <<<"$REGION"
CROP_LEFT="$REGION_X"
CROP_TOP="$REGION_Y"
CROP_RIGHT="$((ROOT_W - (REGION_X + REGION_W)))"
CROP_BOTTOM="$((ROOT_H - (REGION_Y + REGION_H)))"
if (( CROP_RIGHT < 0 || CROP_BOTTOM < 0 )); then
  echo "invalid REGION=$REGION for root $ROOT_GEOM" >&2
  exit 1
fi
PREVIEW_HEIGHT="$(( (REGION_H * PREVIEW_WIDTH + REGION_W/2) / REGION_W ))"
if (( PREVIEW_HEIGHT < 1 )); then PREVIEW_HEIGHT=1; fi

printf 'display=%s\nroot=%sx%s\nregion=%s\nfps=%s\npreview_fps=%s\npreview_width=%s\npreview_height=%s\n' \
  "$DISPLAY_NAME" "$ROOT_W" "$ROOT_H" "$REGION" "$FPS" "$PREVIEW_FPS" "$PREVIEW_WIDTH" "$PREVIEW_HEIGHT" \
  | tee "$RUN_DIR/context.txt"

run_case() {
  local name="$1"
  local pipeline="$2"
  local output_path="${3:-}"
  local log_prefix="$RUN_DIR/$name"

  echo "## $name" | tee -a "$RUN_DIR/01-summary.md"
  echo "pipeline: $pipeline" > "$log_prefix.pipeline.txt"

  bash -lc "exec gst-launch-1.0 -e $pipeline" >"$log_prefix.stdout.log" 2>"$log_prefix.stderr.log" &
  local pid=$!
  echo "$pid" > "$log_prefix.pid"

  pidstat -u -p "$pid" 1 "$DURATION" > "$log_prefix.pidstat.log" || true

  if kill -0 "$pid" 2>/dev/null; then
    kill -INT "$pid" 2>/dev/null || true
  fi
  wait "$pid" || true

  local avg_cpu max_cpu
  avg_cpu=$(awk '
    $1=="Average:" {print $8; found=1}
    END {if (!found) print ""}
  ' "$log_prefix.pidstat.log" | tail -n1)
  if [[ -z "$avg_cpu" ]]; then
    avg_cpu=$(awk '
      NF >= 11 && $3 ~ /^[0-9]+$/ && $9 ~ /^[0-9.]+$/ {sum += $9; n++}
      END {if (n) printf "%.2f", sum / n}
    ' "$log_prefix.pidstat.log")
  fi
  max_cpu=$(awk '
    NF >= 11 && $3 ~ /^[0-9]+$/ && $9 ~ /^[0-9.]+$/ {if ($9 + 0 > max) max = $9}
    END {if (max == "") max = 0; print max}
  ' "$log_prefix.pidstat.log")
  {
    echo "- avg_cpu: ${avg_cpu}%"
    echo "- max_cpu: ${max_cpu}%"
  } | tee -a "$RUN_DIR/01-summary.md"

  if [[ -n "$output_path" && -f "$output_path" ]]; then
    ffprobe -hide_banner -loglevel error \
      -show_entries format=duration,size:stream=codec_name,width,height,avg_frame_rate \
      -of default=noprint_wrappers=1 "$output_path" \
      > "$log_prefix.ffprobe.txt" || true
    {
      echo "- output: $output_path"
      sed 's/^/  /' "$log_prefix.ffprobe.txt"
    } | tee -a "$RUN_DIR/01-summary.md"
  fi
  echo >> "$RUN_DIR/01-summary.md"
}

CAPTURE_PIPE="ximagesrc display-name=$DISPLAY_NAME use-damage=false show-pointer=true ! videocrop left=$CROP_LEFT top=$CROP_TOP right=$CROP_RIGHT bottom=$CROP_BOTTOM ! videoconvert ! videoscale ! videorate ! video/x-raw,format=I420,width=$REGION_W,height=$REGION_H,framerate=$FPS/1,pixel-aspect-ratio=1/1 ! fakesink sync=false"
PREVIEW_PIPE="ximagesrc display-name=$DISPLAY_NAME use-damage=false show-pointer=true ! videocrop left=$CROP_LEFT top=$CROP_TOP right=$CROP_RIGHT bottom=$CROP_BOTTOM ! videoscale ! video/x-raw,width=$PREVIEW_WIDTH,height=$PREVIEW_HEIGHT,pixel-aspect-ratio=1/1 ! videorate ! video/x-raw,framerate=$PREVIEW_FPS/1 ! jpegenc quality=$PREVIEW_JPEG_QUALITY ! fakesink sync=false"
OUT_CURRENT="$RUN_DIR/direct-record-current.mp4"
OUT_FAST="$RUN_DIR/direct-record-ultrafast.mp4"
CURRENT_PIPE="ximagesrc display-name=$DISPLAY_NAME use-damage=false show-pointer=true ! videocrop left=$CROP_LEFT top=$CROP_TOP right=$CROP_RIGHT bottom=$CROP_BOTTOM ! videoconvert ! videorate ! video/x-raw,framerate=$FPS/1 ! x264enc bitrate=2500 bframes=0 tune=zerolatency speed-preset=3 ! h264parse ! mp4mux ! filesink location=$OUT_CURRENT"
FAST_PIPE="ximagesrc display-name=$DISPLAY_NAME use-damage=false show-pointer=true ! videocrop left=$CROP_LEFT top=$CROP_TOP right=$CROP_RIGHT bottom=$CROP_BOTTOM ! videoconvert ! videorate ! video/x-raw,framerate=$FPS/1 ! x264enc bitrate=2500 bframes=0 tune=zerolatency speed-preset=1 ! h264parse ! mp4mux ! filesink location=$OUT_FAST"

cat > "$RUN_DIR/01-summary.md" <<'SUMMARYEOF'
---
Title: 06 gst-launch recording performance matrix run summary
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - video
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Saved per-run summary for the standalone pure-GStreamer recording performance matrix.
LastUpdated: 2026-04-13T23:12:00-04:00
WhatFor: Preserve the CPU and ffprobe summary for one standalone gst-launch performance run.
WhenToUse: Read when reviewing the raw results under this run directory.
---

# 06 gst-launch recording performance matrix

SUMMARYEOF
run_case capture-to-fakesink "$CAPTURE_PIPE"
run_case preview-like-jpeg "$PREVIEW_PIPE"
run_case direct-record-current "$CURRENT_PIPE" "$OUT_CURRENT"
run_case direct-record-ultrafast "$FAST_PIPE" "$OUT_FAST"

echo "Results written to $RUN_DIR"
