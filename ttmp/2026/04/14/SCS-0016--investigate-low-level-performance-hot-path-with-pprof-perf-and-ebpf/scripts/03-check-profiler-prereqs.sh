#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results/03-check-profiler-prereqs"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

LISTEN_PID="$(lsof -tiTCP:7777 -sTCP:LISTEN | head -n1 || true)"
PERF_PATH="$(command -v perf || true)"
BPFTRACE_PATH="$(command -v bpftrace || true)"
BPFTOOL_PATH="$(command -v bpftool || true)"
PERF_PARANOID="$(cat /proc/sys/kernel/perf_event_paranoid 2>/dev/null || echo unavailable)"

{
  echo "listen_pid=${LISTEN_PID}"
  echo "perf_path=${PERF_PATH}"
  echo "bpftrace_path=${BPFTRACE_PATH}"
  echo "bpftool_path=${BPFTOOL_PATH}"
  echo "perf_event_paranoid=${PERF_PARANOID}"
} > "$RUN_DIR/context.txt"

{
  echo "## perf"
  if [[ -n "$PERF_PATH" ]]; then
    perf --version || true
    if [[ -n "$LISTEN_PID" ]]; then
      perf stat -p "$LISTEN_PID" sleep 1
    else
      echo "no_listening_pid_on_7777"
    fi
  else
    echo "perf_not_found"
  fi
} > "$RUN_DIR/perf-check.txt" 2>&1 || true

{
  echo "## bpftrace"
  if [[ -n "$BPFTRACE_PATH" ]]; then
    bpftrace --version || true
    bpftrace -e 'BEGIN { printf("ok\\n"); exit(); }'
  else
    echo "bpftrace_not_found"
  fi
} > "$RUN_DIR/bpftrace-check.txt" 2>&1 || true

{
  echo "## bpftool"
  if [[ -n "$BPFTOOL_PATH" ]]; then
    bpftool version || true
  else
    echo "bpftool_not_found"
  fi
} > "$RUN_DIR/bpftool-check.txt" 2>&1 || true

cat > "$RUN_DIR/01-summary.md" <<EOF
---
Title: 03 check profiler prereqs summary
Ticket: SCS-0016
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - backend
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Checks local availability and permissions for pprof-adjacent lower-level profiling tools.
LastUpdated: $(date --iso-8601=seconds)
WhatFor: Preserve profiler availability and permission checks before attempting perf or eBPF captures.
WhenToUse: Read before running perf or bpftrace scripts on this machine.
---

# 03 check profiler prereqs summary

- listen_pid: ${LISTEN_PID:-none}
- perf_path: ${PERF_PATH:-missing}
- bpftrace_path: ${BPFTRACE_PATH:-missing}
- bpftool_path: ${BPFTOOL_PATH:-missing}
- perf_event_paranoid: ${PERF_PARANOID}

## Files

- context.txt
- perf-check.txt
- bpftrace-check.txt
- bpftool-check.txt
EOF

echo "$RUN_DIR"
