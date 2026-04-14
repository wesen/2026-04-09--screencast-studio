#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/../../../../../../.." && pwd)"
SOURCE_DIR="$REPO_ROOT/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/16-go-preview-adaptive-confirmation-matrix"
RESULTS_ROOT="$SCRIPT_DIR/results"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

DISPLAY_NAME="${DISPLAY:-:0}"
DURATION="${DURATION:-6s}"
FPS="${FPS:-24}"
ROUNDS="${ROUNDS:-3}"
DURATION_SECS="${DURATION%s}"
if [[ "$DURATION_SECS" == "$DURATION" ]]; then
  echo "DURATION must be specified in whole seconds, e.g. 6s" >&2
  exit 1
fi
PIDSTAT_COUNT="$((DURATION_SECS + 2))"

root_geometry() {
  local geom
  if geom=$(DISPLAY="$DISPLAY_NAME" xwininfo -root 2>/dev/null | awk '/-geometry/ {print $2; exit}'); then
    if [[ -n "$geom" ]]; then
      geom="${geom%%+*}"
      if [[ "$geom" == *x* ]]; then echo "$geom"; return 0; fi
    fi
  fi
  geom=$(DISPLAY="$DISPLAY_NAME" xdpyinfo 2>/dev/null | awk '/dimensions:/ {print $2; exit}')
  if [[ -z "$geom" || "$geom" != *x* ]]; then
    echo "failed to detect root geometry for DISPLAY=$DISPLAY_NAME" >&2
    exit 1
  fi
  echo "$geom"
}
parse_geom() { local geom="$1"; echo "${geom%x*} ${geom#*x}"; }
ROOT_GEOM="$(root_geometry)"
read -r ROOT_W ROOT_H <<<"$(parse_geom "$ROOT_GEOM")"
DEFAULT_REGION="0,$((ROOT_H/2)),$ROOT_W,$((ROOT_H - ROOT_H/2))"
REGION="${REGION:-$DEFAULT_REGION}"

printf 'display=%s\nroot=%sx%s\nregion=%s\nfps=%s\nduration=%s\nrounds=%s\n' \
  "$DISPLAY_NAME" "$ROOT_W" "$ROOT_H" "$REGION" "$FPS" "$DURATION" "$ROUNDS" \
  | tee "$RUN_DIR/context.txt"

cd "$REPO_ROOT"
gofmt -w "$SOURCE_DIR/main.go" >/dev/null 2>&1 || true
go build -o "$RUN_DIR/bench" "$SOURCE_DIR"

parse_avg() {
  local file="$1"
  local avg
  avg=$(awk '$1=="Average:" {print $8; found=1} END {if (!found) print ""}' "$file" | tail -n1)
  if [[ -z "$avg" ]]; then
    avg=$(awk 'NF >= 11 && $3 ~ /^[0-9]+$/ && $9 ~ /^[0-9.]+$/ {sum += $9; n++} END {if (n) printf "%.2f", sum / n}' "$file")
  fi
  printf '%s' "$avg"
}
parse_max() {
  local file="$1"
  awk 'NF >= 11 && $3 ~ /^[0-9]+$/ && $9 ~ /^[0-9.]+$/ {if ($9 + 0 > max) max = $9} END {if (max == "") max = 0; print max}' "$file"
}

run_case() {
  local round="$1"
  local name="$2"
  local scenario="$3"
  local output_path="$RUN_DIR/${round}-${name}.mp4"
  local log_prefix="$RUN_DIR/${round}-${name}"

  SCENARIO="$scenario" DISPLAY="$DISPLAY_NAME" ROOT_WIDTH="$ROOT_W" ROOT_HEIGHT="$ROOT_H" REGION="$REGION" FPS="$FPS" DURATION="$DURATION" OUTPUT_PATH="$output_path" \
    "$RUN_DIR/bench" > "$log_prefix.stdout.log" 2> "$log_prefix.stderr.log" &
  local pid=$!
  echo "$pid" > "$log_prefix.pid"
  pidstat -u -p "$pid" 1 "$PIDSTAT_COUNT" > "$log_prefix.pidstat.log" || true
  wait "$pid"

  local avg_cpu max_cpu
  avg_cpu=$(parse_avg "$log_prefix.pidstat.log")
  max_cpu=$(parse_max "$log_prefix.pidstat.log")
  printf '%s,%s,%s,%s\n' "$round" "$name" "$avg_cpu" "$max_cpu" >> "$RUN_DIR/results.csv"
}

cat > "$RUN_DIR/results.csv" <<'CSVEOF'
round,scenario,avg_cpu,max_cpu
CSVEOF

run_round() {
  local round="$1"
  local parity="$2"
  run_case "$round" recorder-only recorder-only
  if [[ "$parity" == "odd" ]]; then
    run_case "$round" preview-scale-first-current-plus-recorder preview-scale-first-current-plus-recorder
    run_case "$round" preview-rate-first-current-plus-recorder preview-rate-first-current-plus-recorder
    run_case "$round" preview-scale-first-constrained-plus-recorder preview-scale-first-constrained-plus-recorder
    run_case "$round" preview-rate-first-constrained-plus-recorder preview-rate-first-constrained-plus-recorder
  else
    run_case "$round" preview-rate-first-current-plus-recorder preview-rate-first-current-plus-recorder
    run_case "$round" preview-scale-first-current-plus-recorder preview-scale-first-current-plus-recorder
    run_case "$round" preview-rate-first-constrained-plus-recorder preview-rate-first-constrained-plus-recorder
    run_case "$round" preview-scale-first-constrained-plus-recorder preview-scale-first-constrained-plus-recorder
  fi
}

for round in $(seq 1 "$ROUNDS"); do
  if (( round % 2 == 1 )); then
    run_round "round${round}" odd
  else
    run_round "round${round}" even
  fi
done

python3 - "$RUN_DIR/results.csv" "$RUN_DIR/01-summary.md" <<'PY'
import csv
import statistics
import sys
from collections import defaultdict
csv_path, out_path = sys.argv[1], sys.argv[2]
rows = list(csv.DictReader(open(csv_path)))
by = defaultdict(list)
for row in rows:
    row['avg_cpu'] = float(row['avg_cpu'])
    row['max_cpu'] = float(row['max_cpu'])
    by[row['scenario']].append(row)

order = [
    'recorder-only',
    'preview-scale-first-current-plus-recorder',
    'preview-rate-first-current-plus-recorder',
    'preview-scale-first-constrained-plus-recorder',
    'preview-rate-first-constrained-plus-recorder',
]

def mean(vals):
    return sum(vals)/len(vals)

with open(out_path, 'w') as f:
    f.write('---\n')
    f.write('Title: 17 preview adaptive repeatability matrix run summary\n')
    f.write('Ticket: SCS-0014\n')
    f.write('Status: active\n')
    f.write('Topics:\n')
    f.write('    - screencast-studio\n')
    f.write('    - gstreamer\n')
    f.write('    - video\n')
    f.write('    - performance\n')
    f.write('    - analysis\n')
    f.write('DocType: reference\n')
    f.write('Intent: long-term\n')
    f.write('Owners: []\n')
    f.write('RelatedFiles: []\n')
    f.write('ExternalSources: []\n')
    f.write('Summary: Repeated standalone confirmation runs for adaptive preview mitigation scenarios to reduce single-run noise.\n')
    f.write('LastUpdated: 2026-04-14T14:15:00-04:00\n')
    f.write('WhatFor: Preserve repeated CPU measurements for preview-profile and preview-ordering mitigation scenarios independent of the main runtime.\n')
    f.write('WhenToUse: Read when deciding whether the adaptive preview mitigation direction remains plausible after repeatability checks.\n')
    f.write('---\n\n')
    f.write('# 17 preview adaptive repeatability matrix\n\n')
    f.write('## Per-scenario aggregates\n\n')
    for scenario in order:
        vals = by.get(scenario, [])
        avgs = [v['avg_cpu'] for v in vals]
        maxs = [v['max_cpu'] for v in vals]
        f.write(f'## {scenario}\n')
        f.write(f'- runs: {len(vals)}\n')
        f.write(f'- avg_cpu_mean: {mean(avgs):.2f}%\n')
        f.write(f'- avg_cpu_min: {min(avgs):.2f}%\n')
        f.write(f'- avg_cpu_max: {max(avgs):.2f}%\n')
        if len(avgs) > 1:
            f.write(f'- avg_cpu_stdev: {statistics.pstdev(avgs):.2f}\n')
        f.write(f'- max_cpu_mean: {mean(maxs):.2f}%\n')
        f.write('- run_details:\n')
        for v in vals:
            f.write(f"  - {v['round']}: avg={v['avg_cpu']:.2f}% max={v['max_cpu']:.2f}%\n")
        f.write('\n')

    f.write('## Raw CSV\n\n```csv\n')
    with open(csv_path) as csv_f:
        f.write(csv_f.read())
    f.write('```\n')
PY

echo "Results written to $RUN_DIR"
