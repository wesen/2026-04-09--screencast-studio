#!/usr/bin/env bash
# 01-check-gstreamer-env.sh
# Quick inventory of the GStreamer environment on this machine.
# Run from the repo root.
set -euo pipefail

echo "=== GStreamer version ==="
gst-inspect-1.0 --version

echo ""
echo "=== Installed GStreamer packages ==="
dpkg -l | grep -i gstreamer | awk '{print $2, $3}'

echo ""
echo "=== pkg-config GStreamer libraries ==="
pkg-config --list-all 2>/dev/null | grep -i gst || echo "(none found via pkg-config)"

echo ""
echo "=== Key source elements ==="
for elem in ximagesrc pipewiresrc pulsesrc v4l2src; do
    count=$(gst-inspect-1.0 2>/dev/null | grep -c "$elem" || true)
    echo "  $elem: $count matches"
done

echo ""
echo "=== Key sink/processor elements ==="
for elem in appsink appsrc audioconvert audioresample audiomixer videoconvert videoscale videorate jpegenc pngenc x264enc opusenc wavenc; do
    count=$(gst-inspect-1.0 2>/dev/null | grep -c "$elem" || true)
    echo "  $elem: $count matches"
done

echo ""
echo "=== Special elements (screenshot, overlay, whisper) ==="
for elem in gdkpixbufsink textoverlay timeoverlay clockoverlay snapshot; do
    count=$(gst-inspect-1.0 2>/dev/null | grep -c "$elem" || true)
    echo "  $elem: $count matches"
done

echo ""
echo "=== Total elements available ==="
gst-inspect-1.0 2>/dev/null | wc -l

echo ""
echo "=== Go version ==="
go version
