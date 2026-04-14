#!/usr/bin/env bash
# 04-gst-launch-audio-capture.sh
# Experiment: reproduce the FFmpeg audio capture + mix pipeline using gst-launch-1.0.
#
# Current FFmpeg audio pipeline:
#   ffmpeg -f pulse -sample_rate 48000 -channels 2 -i default \
#     -filter_complex "[0:a]volume=1.0[a0];[a0]anull[aout]" \
#     -map "[aout]" -ar 48000 -ac 2 -c:a pcm_s16le output.wav
set -euo pipefail

OUTDIR="/tmp/gst-experiment-audio"
mkdir -p "$OUTDIR"

echo "=== PulseAudio sources ==="
pactl list short sources 2>/dev/null || echo "pactl not available"

echo ""
echo "--- Test 1: Record 3 seconds of WAV from PulseAudio ---"
timeout 4 gst-launch-1.0 -e -v \
  pulsesrc device=default \
  ! "audio/x-raw,rate=48000,channels=2" \
  ! audioconvert \
  ! wavenc \
  ! filesink location="$OUTDIR/test-audio.wav" 2>&1 | tail -8

echo ""
ls -la "$OUTDIR/test-audio.wav" 2>/dev/null && file "$OUTDIR/test-audio.wav"

echo ""
echo "--- Test 2: Record 3 seconds of Opus audio ---"
timeout 4 gst-launch-1.0 -e -v \
  pulsesrc device=default \
  ! "audio/x-raw,rate=48000,channels=2" \
  ! audioconvert \
  ! audioresample \
  ! opusenc bitrate=160000 \
  ! oggmux \
  ! filesink location="$OUTDIR/test-audio.ogg" 2>&1 | tail -8

echo ""
ls -la "$OUTDIR/test-audio.ogg" 2>/dev/null && file "$OUTDIR/test-audio.ogg"

echo ""
echo "--- Test 3: Audio with volume adjustment (gain=1.5) ---"
timeout 4 gst-launch-1.0 -e -v \
  pulsesrc device=default \
  ! "audio/x-raw,rate=48000,channels=2" \
  ! audioconvert \
  ! volume volume=1.5 \
  ! wavenc \
  ! filesink location="$OUTDIR/test-audio-gain.wav" 2>&1 | tail -8

echo ""
ls -la "$OUTDIR/test-audio-gain.wav" 2>/dev/null && file "$OUTDIR/test-audio-gain.wav"
