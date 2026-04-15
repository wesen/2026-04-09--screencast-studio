#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results/04-capture-perf-cpu-profile"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

PORT="${PORT:-7777}"
SERVER_URL="${SERVER_URL:-http://127.0.0.1:7777}"
SERVER_PID="${SERVER_PID:-$(lsof -tiTCP:${PORT} -sTCP:LISTEN | head -n1 || true)}"
EXE_PATH="${EXE_PATH:-}"
PERF_SECONDS="${PERF_SECONDS:-20}"
PERF_FREQ="${PERF_FREQ:-99}"
CALL_GRAPH="${CALL_GRAPH:-dwarf,16384}"
PERF_DATA="$RUN_DIR/perf.data"
PERF_BIN="${PERF_BIN:-$(command -v perf || true)}"
LABEL="${LABEL:-desktop-preview-recording-one-browser-tab}"

if [[ -z "$PERF_BIN" ]]; then
  echo "perf not found in PATH" >&2
  exit 1
fi

if [[ -z "$SERVER_PID" ]]; then
  echo "failed to detect server pid on port ${PORT}" >&2
  exit 1
fi

if ! curl -fsS "$SERVER_URL/api/healthz" > "$RUN_DIR/healthz.json" 2>/dev/null; then
  echo "server health check failed at $SERVER_URL/api/healthz" >&2
  exit 1
fi

if [[ -z "$EXE_PATH" ]] && [[ -e "/proc/$SERVER_PID/exe" ]]; then
  EXE_PATH="$(readlink -f "/proc/$SERVER_PID/exe" || true)"
fi

{
  echo "label=${LABEL}"
  echo "server_url=${SERVER_URL}"
  echo "port=${PORT}"
  echo "server_pid=${SERVER_PID}"
  echo "exe_path=${EXE_PATH:-unknown}"
  echo "perf_bin=${PERF_BIN}"
  echo "perf_seconds=${PERF_SECONDS}"
  echo "perf_freq=${PERF_FREQ}"
  echo "call_graph=${CALL_GRAPH}"
  echo "perf_event_paranoid=$(cat /proc/sys/kernel/perf_event_paranoid 2>/dev/null || echo unavailable)"
  echo "uname=$(uname -a)"
  echo "started_at=$(date --iso-8601=seconds)"
} > "$RUN_DIR/context.txt"

"$PERF_BIN" --version > "$RUN_DIR/perf-version.txt"

cat > "$RUN_DIR/README.txt" <<'EOF'
Run this while the real browser high-signal repro is active:
- desktop preview
- recording active
- one real browser tab open on the Studio page

This script only captures perf artifacts from the live server PID.
It does not drive the browser scenario by itself.
EOF

"$PERF_BIN" record \
  -F "$PERF_FREQ" \
  --call-graph "$CALL_GRAPH" \
  -o "$PERF_DATA" \
  -p "$SERVER_PID" \
  -- sleep "$PERF_SECONDS" \
  > "$RUN_DIR/perf-record.stdout.log" \
  2> "$RUN_DIR/perf-record.stderr.log"

"$PERF_BIN" report --stdio -i "$PERF_DATA" \
  > "$RUN_DIR/perf-report.txt" \
  2> "$RUN_DIR/perf-report.stderr.log"

"$PERF_BIN" report --stdio --sort dso,symbol -i "$PERF_DATA" \
  > "$RUN_DIR/perf-report-dso-symbol.txt" \
  2> "$RUN_DIR/perf-report-dso-symbol.stderr.log"

"$PERF_BIN" script -i "$PERF_DATA" \
  > "$RUN_DIR/perf-script.txt" \
  2> "$RUN_DIR/perf-script.stderr.log"

cat > "$RUN_DIR/01-summary.md" <<EOF
---
Title: 04 capture perf cpu profile summary
Ticket: SCS-0016
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - backend
    - analysis
    - perf
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Saved mixed-stack perf profile capture from the live server PID while the browser high-signal repro is active.
LastUpdated: $(date --iso-8601=seconds)
WhatFor: Preserve one perf capture and its text reports for comparison against pprof and later reruns.
WhenToUse: Read when comparing mixed-stack perf captures across the same high-signal repro.
---

# 04 capture perf cpu profile summary

- label: $LABEL
- server_url: $SERVER_URL
- server_pid: $SERVER_PID
- exe_path: ${EXE_PATH:-unknown}
- perf_seconds: $PERF_SECONDS
- perf_freq: $PERF_FREQ
- call_graph: $CALL_GRAPH
- perf_data: $PERF_DATA

## Files

- context.txt
- healthz.json
- perf-version.txt
- perf.data
- perf-record.stdout.log
- perf-record.stderr.log
- perf-report.txt
- perf-report-dso-symbol.txt
- perf-script.txt
- go-top-addresses.txt (generated later via `08-resolve-perf-go-addresses.sh`)
- go-addr2line.txt (generated later via `08-resolve-perf-go-addresses.sh`)
EOF

echo "$RUN_DIR"
