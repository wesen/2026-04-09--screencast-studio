#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PARENT_DIR="$(cd -- "$SCRIPT_DIR/.." && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

DISPLAY_NAME="${DISPLAY:-:0}"
DURATION="${DURATION:-6s}"
FPS="${FPS:-24}"
PREVIEW_FPS="${PREVIEW_FPS:-10}"
PREVIEW_WIDTH="${PREVIEW_WIDTH:-1280}"
PREVIEW_JPEG_QUALITY="${PREVIEW_JPEG_QUALITY:-80}"

root_geometry() {
  local geom
  if geom=$(DISPLAY="$DISPLAY_NAME" xwininfo -root 2>/dev/null | awk '/-geometry/ {print $2; exit}'); then
    if [[ -n "$geom" ]]; then
      geom="${geom%%+*}"
      if [[ "$geom" == *x* ]]; then
        echo "$geom"
        return 0
      fi
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
DURATION_FOR_06="${DURATION%s}"
if [[ "$DURATION_FOR_06" == "$DURATION" ]]; then
  echo "DURATION must be specified like 6s" >&2
  exit 1
fi

printf 'display=%s\nroot=%sx%s\nregion=%s\nfps=%s\nduration=%s\npreview_fps=%s\npreview_width=%s\npreview_jpeg_quality=%s\n' \
  "$DISPLAY_NAME" "$ROOT_W" "$ROOT_H" "$REGION" "$FPS" "$DURATION" "$PREVIEW_FPS" "$PREVIEW_WIDTH" "$PREVIEW_JPEG_QUALITY" \
  | tee "$RUN_DIR/context.txt"

run_child() {
  local name="$1"
  shift
  local log="$RUN_DIR/$name.run.log"
  ( "$@" ) 2>&1 | tee "$log" >&2
  awk '/Results written to /{print $NF}' "$log" | tail -n1
}

RUN06=$(run_child run06 env DISPLAY="$DISPLAY_NAME" REGION="$REGION" FPS="$FPS" DURATION="$DURATION_FOR_06" PREVIEW_FPS="$PREVIEW_FPS" PREVIEW_WIDTH="$PREVIEW_WIDTH" PREVIEW_JPEG_QUALITY="$PREVIEW_JPEG_QUALITY" bash "$PARENT_DIR/06-gst-recording-performance-matrix/run.sh")
RUN07=$(run_child run07 env DISPLAY="$DISPLAY_NAME" REGION="$REGION" FPS="$FPS" DURATION="$DURATION" bash "$PARENT_DIR/07-go-shared-recording-performance-matrix/run.sh")
RUN09=$(run_child run09 env DISPLAY="$DISPLAY_NAME" REGION="$REGION" FPS="$FPS" DURATION="$DURATION" bash "$PARENT_DIR/09-go-bridge-overhead-matrix/run.sh")

cat > "$RUN_DIR/01-summary.md" <<SUMMARYEOF
---
Title: 11 shared vs bridge reconciliation matrix summary
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - video
    - performance
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Unified same-session rerun of the direct GStreamer, shared-runtime, and staged bridge-overhead benchmark suites for reconciliation.
LastUpdated: 2026-04-13T23:50:00-04:00
WhatFor: Compare the earlier benchmark suites in one same-session run so their differences can be interpreted more confidently.
WhenToUse: Read when reconciling the full shared-runtime CPU results against the staged bridge-overhead results.
---

# 11 shared vs bridge reconciliation matrix

## Context

- display: \
  \
  \
  $DISPLAY_NAME
- root: $ROOT_W x $ROOT_H
- region: $REGION
- fps: $FPS
- duration: $DURATION

## Child result directories

- 06 direct/pure GStreamer: $RUN06
- 07 shared runtime: $RUN07
- 09 staged bridge overhead: $RUN09

## 06 summary

SUMMARYEOF
cat "$RUN06/01-summary.md" >> "$RUN_DIR/01-summary.md"
cat >> "$RUN_DIR/01-summary.md" <<'SUMMARYEOF'

## 07 summary

SUMMARYEOF
cat "$RUN07/01-summary.md" >> "$RUN_DIR/01-summary.md"
cat >> "$RUN_DIR/01-summary.md" <<'SUMMARYEOF'

## 09 summary

SUMMARYEOF
cat "$RUN09/01-summary.md" >> "$RUN_DIR/01-summary.md"

python3 - <<'PY' "$RUN06/01-summary.md" "$RUN07/01-summary.md" "$RUN09/01-summary.md" >> "$RUN_DIR/01-summary.md"
import re, sys
from pathlib import Path
run06, run07, run09 = [Path(p).read_text() for p in sys.argv[1:4]]

def grab(text, name):
    m = re.search(rf"## {re.escape(name)}\n- .*?avg_cpu: ([0-9.]+)%", text, re.S)
    return m.group(1) if m else "?"

print("""
## Reconciliation highlights

The purpose of this run is not to create a new theory yet, but to rerun the previously disagreeing benchmark families in one same-session matrix.

### Key numbers from this run
""")
print(f"- 06 direct-record-current avg CPU: **{grab(run06, 'direct-record-current')}%**")
print(f"- 06 direct-record-ultrafast avg CPU: **{grab(run06, 'direct-record-ultrafast')}%**")
print(f"- 07 recorder-only avg CPU: **{grab(run07, 'recorder-only')}%**")
print(f"- 07 preview-plus-recorder avg CPU: **{grab(run07, 'preview-plus-recorder')}%**")
print(f"- 09 appsink-copy-async-appsrc-x264 avg CPU: **{grab(run09, 'appsink-copy-async-appsrc-x264')}%**")
print(f"- 09 normalized-fakesink avg CPU: **{grab(run09, 'normalized-fakesink')}%**")
print("""
### Why this comparison matters

- 06 tells us the cost of direct GStreamer capture and encode without the app's shared runtime.
- 07 tells us the cost of the current real shared-runtime recording path.
- 09 tells us the cost of staged bridge components in isolation.

If 07 stays much higher than both 06 direct encode and 09 staged bridge+x264, then the remaining discrepancy is likely coming from the full shared-runtime path rather than from x264 alone or from the minimal appsink/appsrc bridge alone.
""")
PY

echo "Results written to $RUN_DIR"
