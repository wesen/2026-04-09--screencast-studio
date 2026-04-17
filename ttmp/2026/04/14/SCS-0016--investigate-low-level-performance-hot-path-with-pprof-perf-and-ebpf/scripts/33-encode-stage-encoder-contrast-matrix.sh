#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results/33-encode-stage-encoder-contrast-matrix"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

DISPLAY_NAME="${DISPLAY_NAME:-${DISPLAY:-:0}}"
FPS="${FPS:-24}"
BITRATE="${BITRATE:-6920}"
DURATION_SECONDS="${DURATION_SECONDS:-6}"
RUN_TIMEOUT_SECONDS="${RUN_TIMEOUT_SECONDS:-30}"
ENCODER_LIST="${ENCODER_LIST:-x264enc openh264enc vaapih264enc}"
IFS=' ' read -r -a ENCODERS <<< "$ENCODER_LIST"

declare -A HOST_SCRIPT=(
  [go]="$SCRIPT_DIR/29-go-manual-stage-ladder-harness/run.sh"
  [python]="$SCRIPT_DIR/30-python-manual-stage-ladder-harness/run.sh"
  [gst-launch]="$SCRIPT_DIR/31-gst-launch-stage-ladder.sh"
)

printf 'host\tencoder\tresult_dir\tavg_cpu\tmax_cpu\tpage_faults\tminor_faults\tmajor_faults\tresult\terror\n' > "$RUN_DIR/01-manifest.tsv"

for encoder in "${ENCODERS[@]}"; do
  for host in go python gst-launch; do
    echo "running host=$host encoder=$encoder" >&2
    result_dir=""
    avg_cpu=""
    max_cpu=""
    page_faults=""
    minor_faults=""
    major_faults=""
    result=""
    error=""
    if result_dir=$(DISPLAY_NAME="$DISPLAY_NAME" FPS="$FPS" BITRATE="$BITRATE" DURATION_SECONDS="$DURATION_SECONDS" STAGE=encode ENCODER="$encoder" timeout --signal=INT "${RUN_TIMEOUT_SECONDS}s" bash "${HOST_SCRIPT[$host]}"); then
      summary="$result_dir/01-summary.md"
      perf_csv="$result_dir/perf-stat.csv"
      avg_cpu=$(awk -F': ' '/avg_cpu:/ {gsub(/%/, "", $2); print $2; exit}' "$summary")
      max_cpu=$(awk -F': ' '/max_cpu:/ {gsub(/%/, "", $2); print $2; exit}' "$summary")
      page_faults=$(awk -F',' '$3=="page-faults" {print $1; exit}' "$perf_csv" 2>/dev/null || true)
      minor_faults=$(awk -F',' '$3=="minor-faults" {print $1; exit}' "$perf_csv" 2>/dev/null || true)
      major_faults=$(awk -F',' '$3=="major-faults" {print $1; exit}' "$perf_csv" 2>/dev/null || true)
    else
      result="timeout"
      error="timed out after ${RUN_TIMEOUT_SECONDS}s"
    fi
    if [[ -n "$result_dir" && -f "$result_dir/harness.stdout.json" ]]; then
      result=$(python3 - <<'PY' "$result_dir/harness.stdout.json"
import json,sys
p=sys.argv[1]
try:
    data=json.load(open(p))
    print(data.get('result',''))
except Exception:
    print('')
PY
)
      error=$(python3 - <<'PY' "$result_dir/harness.stdout.json"
import json,sys
p=sys.argv[1]
try:
    data=json.load(open(p))
    print(str(data.get('error','')).replace('\n',' ')[:180])
except Exception:
    print('')
PY
)
    elif [[ -n "$result_dir" && -f "$result_dir/stderr.log" ]]; then
      if rg -q 'ERROR|error|No VA display|Failed|failed' "$result_dir/stderr.log"; then
        error=$(tr '\n' ' ' < "$result_dir/stderr.log" | sed 's/[[:space:]]\+/ /g' | cut -c1-180)
      fi
      result="gst-launch"
    fi
    printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$host" "$encoder" "$result_dir" "${avg_cpu:-}" "${max_cpu:-}" "${page_faults:-}" "${minor_faults:-}" "${major_faults:-}" "${result:-}" "${error:-}" >> "$RUN_DIR/01-manifest.tsv"
  done
 done

python3 - "$RUN_DIR" <<'PY'
import csv, pathlib, sys
run_dir = pathlib.Path(sys.argv[1])
rows = list(csv.DictReader((run_dir / '01-manifest.tsv').open(), delimiter='\t'))
out = [
    '---',
    'Title: 33 encode stage encoder contrast matrix summary',
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
    'Summary: Focused encode-stage comparison across x264enc, openh264enc, and vaapih264enc for Go, Python, and gst-launch hosts.',
    'LastUpdated: 2026-04-15T04:20:00-04:00',
    'WhatFor: Preserve the encoder-only contrast matrix used to test whether the Go-hosted anomaly is specific to x264enc or broader across encoder implementations.',
    'WhenToUse: Read this when continuing the encoder-boundary investigation after the small-graph ladder localized the first strong divergence to encode.',
    '---',
    '',
    '# 33 encode stage encoder contrast matrix',
    '',
]
encoders = sorted({r['encoder'] for r in rows}, key=lambda x: ['x264enc','openh264enc','vaapih264enc'].index(x) if x in ['x264enc','openh264enc','vaapih264enc'] else 999)
hosts = ['go', 'python', 'gst-launch']
for enc in encoders:
    out.append(f'## {enc}')
    out.append('')
    out.append('| host | avg_cpu | max_cpu | page-faults | minor-faults | major-faults | result | error |')
    out.append('|---|---:|---:|---:|---:|---:|---|---|')
    bucket = [r for r in rows if r['encoder'] == enc]
    for host in hosts:
        r = next((x for x in bucket if x['host'] == host), {})
        out.append(f"| {host} | {r.get('avg_cpu','?')} | {r.get('max_cpu','?')} | {r.get('page_faults','?')} | {r.get('minor_faults','?')} | {r.get('major_faults','?')} | {r.get('result','')} | {r.get('error','')} |")
    out.append('')
(run_dir / '02-summary.md').write_text('\n'.join(out) + '\n')
PY

echo "$RUN_DIR"
