#!/usr/bin/env bash
# 03-gst-launch-preview-x11.sh
# Experiment: reproduce the current FFmpeg preview pipeline using gst-launch-1.0.
#
# Current FFmpeg preview does:
#   ffmpeg -f x11grab -framerate 5 -i :0.0 \
#     -vf "fps=5,scale=640:-1:force_original_aspect_ratio=decrease" \
#     -q:v 7 -f image2pipe -vcodec mjpeg pipe:1
#
# GStreamer equivalent: ximagesrc -> videoconvert -> videoscale -> capsfilter -> jpegenc -> multifdsink (stdout)
#
# NOTE: gst-launch-1.0 can't write MJPEG frames to stdout cleanly like ffmpeg.
# Instead, we demonstrate the pipeline and save to a file to prove it works.
set -euo pipefail

DISPLAY="${DISPLAY:-:0}"
FPS=5
OUTDIR="/tmp/gst-experiment-preview"
mkdir -p "$OUTDIR"

# Check ximagesrc properties
echo "=== ximagesrc properties ==="
gst-inspect-1.0 ximagesrc 2>&1 | grep -A1 "display\|startx\|starty\|endx\|endy\|use-damage\|show-cursor"


# Simplest: capture 2 seconds of low-fps JPEG frames to files
echo ""
echo "--- Test 1: Save JPEG frames for 3 seconds ---"
timeout 4 gst-launch-1.0 -v \
  ximagesrc startx=0 starty=0 endx=639 endy=479 use-damage=false \
  ! videoconvert \
  ! videoscale \
  ! "video/x-raw,width=640" \
  ! videorate \
  ! "video/x-raw,framerate=${FPS}/1" \
  ! jpegenc quality=50 \
  ! multifilesink location="$OUTDIR/frame-%05d.jpg" 2>&1 | tail -15

echo ""
echo "Frames captured:"
ls -la "$OUTDIR/" | head -20

echo ""
echo "--- Test 2: Encode 3 seconds to a test MP4 (h264) ---"
timeout 4 gst-launch-1.0 -v \
  ximagesrc startx=0 starty=0 endx=639 endy=479 use-damage=false \
  ! videoconvert \
  ! videoscale \
  ! "video/x-raw,width=640,framerate=10/1" \
  ! x264enc tune=zerolatency bitrate=1000 speed-preset=veryfast \
  ! mp4mux \
  ! filesink location="$OUTDIR/test-preview.mp4" 2>&1 | tail -15

echo ""
ls -la "$OUTDIR/test-preview.mp4"

echo ""
echo "--- Test 3: Screenshot via gdkpixbufsink ---"
timeout 3 gst-launch-1.0 -v \
  ximagesrc startx=0 starty=0 endx=639 endy=479 num-buffers=1 \
  ! videoconvert \
  ! gdkpixbufsink location="$OUTDIR/screenshot.png" 2>&1 | tail -10

echo ""
ls -la "$OUTDIR/screenshot.png" 2>/dev/null || echo "Screenshot not created"
