#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results/04-capture-perf-cpu-profile"
RUN_DIR="${1:-}"
TOP_N="${TOP_N:-20}"

if [[ -z "$RUN_DIR" ]]; then
  RUN_DIR="$(find "$RESULTS_ROOT" -mindepth 1 -maxdepth 1 -type d | sort | tail -n1)"
fi

if [[ -z "$RUN_DIR" || ! -d "$RUN_DIR" ]]; then
  echo "run dir not found" >&2
  exit 1
fi

REPORT="$RUN_DIR/perf-report-dso-symbol.txt"
CONTEXT="$RUN_DIR/context.txt"
ADDRS_OUT="$RUN_DIR/go-top-addresses.txt"
RESOLVED_OUT="$RUN_DIR/go-addr2line.txt"
EXE_PATH="${EXE_PATH:-}"

if [[ ! -f "$REPORT" ]]; then
  echo "missing perf report: $REPORT" >&2
  exit 1
fi

if [[ -z "$EXE_PATH" && -f "$CONTEXT" ]]; then
  EXE_PATH="$(awk -F= '$1=="exe_path" {print substr($0, index($0, "=")+1)}' "$CONTEXT" | tail -n1)"
fi

if [[ -z "$EXE_PATH" || "$EXE_PATH" == "unknown" ]]; then
  SERVER_PID="$(awk -F= '$1=="server_pid" {print substr($0, index($0, "=")+1)}' "$CONTEXT" | tail -n1)"
  if [[ -n "$SERVER_PID" && -e "/proc/$SERVER_PID/exe" ]]; then
    EXE_PATH="$(readlink -f "/proc/$SERVER_PID/exe" || true)"
  fi
fi

if [[ -z "$EXE_PATH" || ! -f "$EXE_PATH" ]]; then
  echo "failed to resolve executable path for addr2line" >&2
  exit 1
fi

python3 - <<'PY' "$REPORT" "$TOP_N" > "$ADDRS_OUT"
from pathlib import Path
import sys
report = Path(sys.argv[1])
limit = int(sys.argv[2])
addrs = []
for line in report.read_text().splitlines():
    if 'screencast-studio' not in line or '[.] 0x' not in line:
        continue
    for tok in line.split():
        if tok.startswith('0x0000000000'):
            addrs.append(tok)
            break
    if len(addrs) >= limit:
        break
import sys
sys.stdout.write('\n'.join(addrs))
PY

if [[ ! -s "$ADDRS_OUT" ]]; then
  cat > "$ADDRS_OUT" <<EOF
No address-only screencast-studio frames were extracted from:
$REPORT

This usually means perf already symbolized the main Go binary directly in perf-report-dso-symbol.txt.
EOF
  cat > "$RESOLVED_OUT" <<EOF
No address-only screencast-studio frames were extracted from:
$REPORT

This usually means perf already symbolized the main Go binary directly in perf-report-dso-symbol.txt.
Review that report first instead of addr2line fallback output.
EOF
  echo "$RESOLVED_OUT"
  exit 0
fi

go tool addr2line "$EXE_PATH" < "$ADDRS_OUT" > "$RESOLVED_OUT"

echo "$RESOLVED_OUT"
