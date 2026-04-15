#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results/02-capture-pprof-cpu-profile"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

PPROF_URL=${PPROF_URL:-http://127.0.0.1:6060/debug/pprof/profile}
PROFILE_SECONDS=${PROFILE_SECONDS:-20}

curl -fsS "$PPROF_URL?seconds=$PROFILE_SECONDS" -o "$RUN_DIR/cpu.pprof"

go tool pprof -top -nodecount=50 "$RUN_DIR/cpu.pprof" > "$RUN_DIR/pprof-top.txt"
go tool pprof -top -cum -nodecount=50 "$RUN_DIR/cpu.pprof" > "$RUN_DIR/pprof-top-cum.txt"

cat > "$RUN_DIR/01-summary.md" <<EOF
---
Title: 02 capture pprof cpu profile summary
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
Summary: Saved CPU profile capture from the opt-in pprof server.
LastUpdated: $(date --iso-8601=seconds)
WhatFor: Preserve one pprof CPU profile and its top summaries for later comparison.
WhenToUse: Read when comparing repeated pprof captures across the same high-signal repro.
---

# 02 capture pprof cpu profile summary

- pprof_url: $PPROF_URL
- seconds: $PROFILE_SECONDS
- run_dir: $RUN_DIR

## Files

- cpu.pprof
- pprof-top.txt
- pprof-top-cum.txt
EOF

echo "$RUN_DIR"
