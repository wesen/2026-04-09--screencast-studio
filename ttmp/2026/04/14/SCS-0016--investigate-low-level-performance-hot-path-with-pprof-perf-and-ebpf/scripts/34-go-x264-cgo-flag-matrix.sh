#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results/34-go-x264-cgo-flag-matrix"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

DISPLAY_NAME="${DISPLAY_NAME:-${DISPLAY:-:0}}"
FPS="${FPS:-24}"
BITRATE="${BITRATE:-6920}"
DURATION_SECONDS="${DURATION_SECONDS:-6}"
HARNESS="$SCRIPT_DIR/29-go-manual-stage-ladder-harness/run.sh"

printf 'scenario\tcgo_cflags\tresult_dir\tavg_cpu\tmax_cpu\tpage_faults\tminor_faults\tmajor_faults\tresult\terror\n' > "$RUN_DIR/01-manifest.tsv"

run_one() {
  local scenario="$1"
  local cgo_cflags="$2"
  local result_dir=""
  if [[ -n "$cgo_cflags" ]]; then
    result_dir=$(DISPLAY_NAME="$DISPLAY_NAME" FPS="$FPS" BITRATE="$BITRATE" DURATION_SECONDS="$DURATION_SECONDS" STAGE=encode ENCODER=x264enc CGO_CFLAGS="$cgo_cflags" bash "$HARNESS")
  else
    result_dir=$(DISPLAY_NAME="$DISPLAY_NAME" FPS="$FPS" BITRATE="$BITRATE" DURATION_SECONDS="$DURATION_SECONDS" STAGE=encode ENCODER=x264enc bash "$HARNESS")
  fi

  local summary="$result_dir/01-summary.md"
  local perf_csv="$result_dir/perf-stat.csv"
  local avg_cpu max_cpu page_faults minor_faults major_faults result error
  avg_cpu=$(awk -F': ' '/avg_cpu:/ {gsub(/%/, "", $2); print $2; exit}' "$summary")
  max_cpu=$(awk -F': ' '/max_cpu:/ {gsub(/%/, "", $2); print $2; exit}' "$summary")
  page_faults=$(awk -F',' '$3=="page-faults" {print $1; exit}' "$perf_csv" 2>/dev/null || true)
  minor_faults=$(awk -F',' '$3=="minor-faults" {print $1; exit}' "$perf_csv" 2>/dev/null || true)
  major_faults=$(awk -F',' '$3=="major-faults" {print $1; exit}' "$perf_csv" 2>/dev/null || true)
  result=$(python3 - <<'PY' "$result_dir/harness.stdout.json"
import json,sys
try:
    print(json.load(open(sys.argv[1])).get('result',''))
except Exception:
    print('')
PY
)
  error=$(python3 - <<'PY' "$result_dir/harness.stdout.json"
import json,sys
try:
    print(str(json.load(open(sys.argv[1])).get('error','')).replace('\n',' ')[:180])
except Exception:
    print('')
PY
)

  printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$scenario" "$cgo_cflags" "$result_dir" "${avg_cpu:-}" "${max_cpu:-}" "${page_faults:-}" "${minor_faults:-}" "${major_faults:-}" "${result:-}" "${error:-}" >> "$RUN_DIR/01-manifest.tsv"
}

run_one default ""
run_one cgo_o2 "-O2"
run_one cgo_o3 "-O3"

python3 - "$RUN_DIR" <<'PY'
import csv, pathlib, sys
run_dir = pathlib.Path(sys.argv[1])
rows = list(csv.DictReader((run_dir / '01-manifest.tsv').open(), delimiter='\t'))
out = [
    '---',
    'Title: 34 Go x264 CGO flag matrix summary',
    'Ticket: SCS-0016',
    'Status: active',
    'Topics:',
    '    - screencast-studio',
    '    - gstreamer',
    '    - backend',
    '    - analysis',
    'DocType: reference',
    'Intent: long-term',
    'Owners: []',
    'RelatedFiles: []',
    'ExternalSources: []',
    'Summary: Focused Go-only x264enc encode-stage comparison across different CGO_CFLAGS settings.',
    'LastUpdated: 2026-04-15T04:35:00-04:00',
    'WhatFor: Preserve the build-flag control used to test whether the Go x264enc anomaly is sensitive to cgo C optimization levels.',
    'WhenToUse: Read this when evaluating whether the x264enc anomaly is likely in thin cgo wrapper compilation or elsewhere in the Go-hosted process behavior.',
    '---',
    '',
    '# 34 Go x264 CGO flag matrix',
    '',
    '| scenario | cgo_cflags | avg_cpu | max_cpu | page-faults | minor-faults | major-faults | result | error |',
    '|---|---|---:|---:|---:|---:|---:|---|---|',
]
for scenario in ['default', 'cgo_o2', 'cgo_o3']:
    r = next((x for x in rows if x['scenario'] == scenario), {})
    out.append(f"| {scenario} | {r.get('cgo_cflags','')} | {r.get('avg_cpu','?')} | {r.get('max_cpu','?')} | {r.get('page_faults','?')} | {r.get('minor_faults','?')} | {r.get('major_faults','?')} | {r.get('result','')} | {r.get('error','')} |")
(run_dir / '02-summary.md').write_text('\n'.join(out) + '\n')
PY

echo "$RUN_DIR"
