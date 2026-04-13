#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../../../../../.."

echo "Running web GStreamer preview E2E..."
echo "Override SOURCE_TYPE, DISPLAY_NAME, DEVICE, WINDOW_ID, REGION as needed."
echo

go run ./ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/11-web-gst-preview-e2e
