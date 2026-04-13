#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../../../../../.."

echo "Running GStreamer preview runtime smoke test..."
echo "You can override SOURCE_TYPE, DISPLAY_NAME, DEVICE, WINDOW_ID, REGION."
echo

go run ./ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/09-go-gst-preview-runtime-smoke
