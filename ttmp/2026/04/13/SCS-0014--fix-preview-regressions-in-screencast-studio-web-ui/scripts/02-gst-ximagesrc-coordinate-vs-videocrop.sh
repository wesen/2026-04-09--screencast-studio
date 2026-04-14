#!/usr/bin/env bash
set -euo pipefail

OUT_DIR="${OUT_DIR:-/tmp/scs-gst-crop-debug}"
mkdir -p "$OUT_DIR"

# Default geometries from the live debugging session on this machine.
ROOT_W="${ROOT_W:-2880}"
ROOT_H="${ROOT_H:-1920}"
BOTTOM_Y="${BOTTOM_Y:-960}"
WINDOW_X="${WINDOW_X:-6}"
WINDOW_Y="${WINDOW_Y:-30}"
WINDOW_W="${WINDOW_W:-1431}"
WINDOW_H="${WINDOW_H:-1884}"

REGION_DIRECT="$OUT_DIR/region-direct.jpg"
REGION_CROP="$OUT_DIR/region-videocrop.jpg"
WINDOW_DIRECT="$OUT_DIR/window-direct.jpg"

echo "Writing outputs to $OUT_DIR"

# Direct coordinate capture that was observed to be unreliable visually.
gst-launch-1.0 -q \
  ximagesrc use-damage=false startx=0 starty="$BOTTOM_Y" endx="$((ROOT_W-1))" endy="$((ROOT_H-1))" num-buffers=1 \
  ! videoconvert ! jpegenc ! multifilesink location="$REGION_DIRECT"

# Full-root capture plus explicit videocrop, which produced a true crop during debugging.
gst-launch-1.0 -q \
  ximagesrc use-damage=false num-buffers=1 \
  ! videocrop top="$BOTTOM_Y" bottom=0 left=0 right=0 \
  ! videoconvert ! jpegenc ! multifilesink location="$REGION_CROP"

# Direct coordinate capture for the selected window geometry.
gst-launch-1.0 -q \
  ximagesrc use-damage=false startx="$WINDOW_X" starty="$WINDOW_Y" endx="$((WINDOW_X+WINDOW_W-1))" endy="$((WINDOW_Y+WINDOW_H-1))" num-buffers=1 \
  ! videoconvert ! jpegenc ! multifilesink location="$WINDOW_DIRECT"

file "$REGION_DIRECT" "$REGION_CROP" "$WINDOW_DIRECT"
sha256sum "$REGION_DIRECT" "$REGION_CROP" "$WINDOW_DIRECT"

echo
echo "Use an image viewer or understand_image on:"
echo "  $REGION_DIRECT"
echo "  $REGION_CROP"
echo "  $WINDOW_DIRECT"
