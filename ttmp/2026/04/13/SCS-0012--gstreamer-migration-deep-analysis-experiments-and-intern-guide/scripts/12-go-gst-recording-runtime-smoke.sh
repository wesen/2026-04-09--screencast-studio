#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../../../../../.."

echo "Running GStreamer recording runtime smoke test..."
echo "Override SOURCE_TYPE, DISPLAY_NAME, DEVICE, WINDOW_ID, REGION, SIZE, CONTAINER, OUT_PATH as needed."
echo

go run ./ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/12-go-gst-recording-runtime-smoke

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
