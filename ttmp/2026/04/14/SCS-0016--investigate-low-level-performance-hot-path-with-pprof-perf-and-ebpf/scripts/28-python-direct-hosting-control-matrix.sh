#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results/28-python-direct-hosting-control-matrix"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

DISPLAY_NAME="${DISPLAY_NAME:-${DISPLAY:-:0}}"
FPS="${FPS:-24}"
BITRATE="${BITRATE:-6920}"
CONTAINER="${CONTAINER:-mov}"
DURATION_SECONDS="${DURATION_SECONDS:-8}"

PARSE_RUN=$(DISPLAY_NAME="$DISPLAY_NAME" FPS="$FPS" BITRATE="$BITRATE" CONTAINER="$CONTAINER" DURATION_SECONDS="$DURATION_SECONDS" bash "$SCRIPT_DIR/26-python-parse-launch-direct-full-desktop-harness/run.sh")
MANUAL_RUN=$(DISPLAY_NAME="$DISPLAY_NAME" FPS="$FPS" BITRATE="$BITRATE" CONTAINER="$CONTAINER" DURATION_SECONDS="$DURATION_SECONDS" bash "$SCRIPT_DIR/27-python-manual-direct-full-desktop-harness/run.sh")

cat > "$RUN_DIR/01-summary.md" <<EOF
---
Title: 28 python direct hosting control matrix summary
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
Summary: Comparison summary for the Python parse-launch and Python manual direct full-desktop controls.
LastUpdated: 2026-04-15T04:10:00-04:00
WhatFor: Preserve the Python-only hosting comparison used before the smaller stage ladder work.
WhenToUse: Read this when comparing Python-hosted direct recording variants.
---

# 28 python direct hosting control matrix

- parse_launch_run: $PARSE_RUN
- manual_run: $MANUAL_RUN

| scenario | avg_cpu | max_cpu | duration |
|---|---:|---:|---:|
| python-parse-launch | $(awk -F': ' '/avg_cpu:/ {print $2; exit}' "$PARSE_RUN/01-summary.md") | $(awk -F': ' '/max_cpu:/ {print $2; exit}' "$PARSE_RUN/01-summary.md") | $(awk -F'=' '/^duration=/ {print $2; exit}' "$PARSE_RUN/ffprobe.txt") |
| python-manual | $(awk -F': ' '/avg_cpu:/ {print $2; exit}' "$MANUAL_RUN/01-summary.md") | $(awk -F': ' '/max_cpu:/ {print $2; exit}' "$MANUAL_RUN/01-summary.md") | $(awk -F'=' '/^duration=/ {print $2; exit}' "$MANUAL_RUN/ffprobe.txt") |
EOF

echo "$RUN_DIR"
