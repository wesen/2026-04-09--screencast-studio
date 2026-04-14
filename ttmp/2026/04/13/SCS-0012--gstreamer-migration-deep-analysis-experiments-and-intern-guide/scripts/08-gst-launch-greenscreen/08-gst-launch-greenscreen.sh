#!/usr/bin/env bash
# 08-gst-launch-greenscreen.sh
# Experiment: greenscreen (chroma key) compositing in GStreamer.
#
# Pipeline:
#   Background layer: videotestsrc (snow/noise) → compositor sink_0
#   Foreground layer: videotestsrc (ball on green bg) → alpha(method=green) → compositor sink_1
#   Output: JPEG frames of the composited result
#
# Key elements:
#   - alpha: chroma-keying (green, blue, or custom color)
#   - compositor: multi-stream compositing with per-stream position, size, z-order
#
# This proves GStreamer can do real-time greenscreen compositing.
set -euo pipefail

OUTDIR="/tmp/gst-greenscreen"
mkdir -p "$OUTDIR"

echo "=== Test 1: Ball on green background, chroma-keyed over snow ==="
timeout 4 gst-launch-1.0 -e \
  compositor name=mix \
    sink_0::zorder=0 \
    sink_1::zorder=1 \
    sink_1::xpos=160 \
    sink_1::ypos=120 \
    sink_1::width=320 \
    sink_1::height=240 \
  ! "video/x-raw,width=640,height=480,framerate=5/1" \
  ! videoconvert \
  ! jpegenc quality=50 \
  ! multifilesink location="$OUTDIR/composite-%05d.jpg" \
  videotestsrc pattern=snow num-buffers=10 \
  ! "video/x-raw,width=640,height=480,framerate=5/1" \
  ! videoconvert \
  ! mix.sink_0 \
  videotestsrc pattern=ball background-color=4278255360 num-buffers=10 \
  ! "video/x-raw,width=320,height=240,framerate=5/1" \
  ! alpha method=green angle=20 noise-level=4 \
  ! videoconvert \
  ! mix.sink_1 \
  2>&1 | tail -5

echo ""
frames=$(ls "$OUTDIR"/composite-*.jpg 2>/dev/null | wc -l)
echo "Composite frames created: $frames"

echo ""
echo "=== Test 2: Custom chroma key (blue screen, RGB 0,0,255) ==="
timeout 4 gst-launch-1.0 -e \
  compositor name=mix \
    sink_0::zorder=0 \
    sink_1::zorder=1 \
  ! "video/x-raw,width=640,height=480,framerate=5/1" \
  ! videoconvert \
  ! jpegenc quality=50 \
  ! multifilesink location="$OUTDIR/bluescreen-%05d.jpg" \
  videotestsrc pattern=smpte num-buffers=5 \
  ! "video/x-raw,width=640,height=480,framerate=5/1" \
  ! videoconvert \
  ! mix.sink_0 \
  videotestsrc pattern=ball background-color=4278190335 num-buffers=5 \
  ! "video/x-raw,width=320,height=240,framerate=5/1" \
  ! alpha method=custom target-r=0 target-g=0 target-b=255 angle=20 \
  ! videoconvert \
  ! mix.sink_1 \
  2>&1 | tail -5

echo ""
frames2=$(ls "$OUTDIR"/bluescreen-*.jpg 2>/dev/null | wc -l)
echo "Bluescreen composite frames created: $frames2"

echo ""
echo "=== Available alpha methods ==="
echo "  green     - built-in green screen keying"
echo "  blue      - built-in blue screen keying"  
echo "  custom    - specify target-r, target-g, target-b for any color"
echo ""
echo "=== Key alpha properties ==="
echo "  angle N          - color tolerance (0-90, higher = more aggressive key)"
echo "  noise-level N    - noise suppression (0-64)"
echo "  black-sensitivity - how to handle dark areas (0-128)"
echo "  white-sensitivity - how to handle bright areas (0-128)"
