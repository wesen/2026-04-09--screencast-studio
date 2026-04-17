#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

EXTRA_ARGS="--debug-disable-preview-state-events --debug-disable-audio-meter-events --debug-disable-websocket-preview-events" \
  bash "$SCRIPT_DIR/09-restart-scs-web-ui-with-built-binary-and-pprof.sh"
