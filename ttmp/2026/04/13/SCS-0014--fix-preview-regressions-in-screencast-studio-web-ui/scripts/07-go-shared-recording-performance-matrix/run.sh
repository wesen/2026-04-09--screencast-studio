#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/../../../../../../.." && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

DISPLAY_NAME="${DISPLAY:-:0}"
DURATION="${DURATION:-6s}"
FPS="${FPS:-24}"
DURATION_SECS="${DURATION%s}"
if [[ "$DURATION_SECS" == "$DURATION" ]]; then
  echo "DURATION must be specified in whole seconds, e.g. 6s" >&2
  exit 1
fi
PIDSTAT_COUNT="$((DURATION_SECS + 2))"

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

printf 'display=%s\nroot=%sx%s\nregion=%s\nfps=%s\nduration=%s\n' \
  "$DISPLAY_NAME" "$ROOT_W" "$ROOT_H" "$REGION" "$FPS" "$DURATION" \
  | tee "$RUN_DIR/context.txt"

cd "$REPO_ROOT"
go build -o "$RUN_DIR/bench" "$SCRIPT_DIR"

run_case() {
  local name="$1"
  local scenario="$2"
  local output_path="$RUN_DIR/$name.mp4"
  local log_prefix="$RUN_DIR/$name"

  SCENARIO="$scenario" DISPLAY="$DISPLAY_NAME" REGION="$REGION" FPS="$FPS" DURATION="$DURATION" OUTPUT_PATH="$output_path" \
    "$RUN_DIR/bench" > "$log_prefix.stdout.log" 2> "$log_prefix.stderr.log" &
  local pid=$!
  echo "$pid" > "$log_prefix.pid"
  pidstat -u -p "$pid" 1 "$PIDSTAT_COUNT" > "$log_prefix.pidstat.log" || true
  wait "$pid"

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

  echo "## $name" >> "$RUN_DIR/01-summary.md"
  echo "- scenario: $scenario" >> "$RUN_DIR/01-summary.md"
  echo "- avg_cpu: ${avg_cpu}%" >> "$RUN_DIR/01-summary.md"
  echo "- max_cpu: ${max_cpu}%" >> "$RUN_DIR/01-summary.md"
  if [[ -f "$output_path" ]]; then
    ffprobe -hide_banner -loglevel error \
      -show_entries format=duration,size:stream=codec_name,width,height,avg_frame_rate \
      -of default=noprint_wrappers=1 "$output_path" \
      > "$log_prefix.ffprobe.txt" || true
    echo "- output: $output_path" >> "$RUN_DIR/01-summary.md"
    sed 's/^/  /' "$log_prefix.ffprobe.txt" >> "$RUN_DIR/01-summary.md"
  fi
  echo >> "$RUN_DIR/01-summary.md"
}

cat > "$RUN_DIR/01-summary.md" <<'SUMMARYEOF'
---
Title: 07 go shared recording performance matrix run summary
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
Summary: Saved per-run summary for the standalone Go shared-source recording performance matrix.
LastUpdated: 2026-04-13T23:12:00-04:00
WhatFor: Preserve the CPU and ffprobe summary for one standalone shared-bridge performance run.
WhenToUse: Read when reviewing the raw results under this run directory.
---

# 07 go shared recording performance matrix

SUMMARYEOF
run_case preview-only preview-only
run_case recorder-only recorder-only
run_case preview-plus-recorder preview-plus-recorder

echo "Results written to $RUN_DIR"
