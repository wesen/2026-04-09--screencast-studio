#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/../../../../../.." && pwd)"

cd "$REPO_ROOT"
go run ./ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/20-go-gst-shared-bridge-recorder-smoke
