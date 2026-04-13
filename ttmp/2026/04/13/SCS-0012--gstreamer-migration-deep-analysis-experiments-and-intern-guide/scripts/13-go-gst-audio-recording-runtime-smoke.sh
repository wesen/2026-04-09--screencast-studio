#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../../../../../.."

echo "Running GStreamer audio recording runtime smoke test..."
echo "Override CODEC, DEVICE, DEVICES, GAINS, SAMPLE_RATE_HZ, CHANNELS, OUT_PATH as needed."
echo

go run ./ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/13-go-gst-audio-recording-runtime-smoke

if [[ -n "${OUT_PATH:-}" && -f "${OUT_PATH}" ]]; then
  echo
  echo "file(1):"
  file "${OUT_PATH}" || true
  if command -v ffprobe >/dev/null 2>&1; then
    echo
    echo "ffprobe:"
    ffprobe -hide_banner -loglevel error -show_entries format=duration,size -of default=noprint_wrappers=1:nokey=0 "${OUT_PATH}" || true
  fi
fi
