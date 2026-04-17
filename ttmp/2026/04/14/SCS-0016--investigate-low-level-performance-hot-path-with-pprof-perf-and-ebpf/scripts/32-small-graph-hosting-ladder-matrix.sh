#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results/32-small-graph-hosting-ladder-matrix"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

DISPLAY_NAME="${DISPLAY_NAME:-${DISPLAY:-:0}}"
FPS="${FPS:-24}"
BITRATE="${BITRATE:-6920}"
CONTAINER="${CONTAINER:-mov}"
DURATION_SECONDS="${DURATION_SECONDS:-8}"
STAGES=(capture convert rate-caps encode parse mux-file)

declare -A HOST_SCRIPT=(
  [go]="$SCRIPT_DIR/29-go-manual-stage-ladder-harness/run.sh"
  [python]="$SCRIPT_DIR/30-python-manual-stage-ladder-harness/run.sh"
  [gst-launch]="$SCRIPT_DIR/31-gst-launch-stage-ladder.sh"
)

printf 'host\tstage\tresult_dir\tavg_cpu\tmax_cpu\tpage_faults\tminor_faults\tmajor_faults\n' > "$RUN_DIR/01-manifest.tsv"

for stage in "${STAGES[@]}"; do
  for host in go python gst-launch; do
    echo "running host=$host stage=$stage" >&2
    result_dir=$(DISPLAY_NAME="$DISPLAY_NAME" FPS="$FPS" BITRATE="$BITRATE" CONTAINER="$CONTAINER" DURATION_SECONDS="$DURATION_SECONDS" STAGE="$stage" bash "${HOST_SCRIPT[$host]}")
    summary="$result_dir/01-summary.md"
    perf_csv="$result_dir/perf-stat.csv"
    avg_cpu=$(awk -F': ' '/avg_cpu:/ {gsub(/%/, "", $2); print $2; exit}' "$summary")
    max_cpu=$(awk -F': ' '/max_cpu:/ {gsub(/%/, "", $2); print $2; exit}' "$summary")
    page_faults=$(awk -F',' '$3=="page-faults" {print $1; exit}' "$perf_csv" 2>/dev/null || true)
    minor_faults=$(awk -F',' '$3=="minor-faults" {print $1; exit}' "$perf_csv" 2>/dev/null || true)
    major_faults=$(awk -F',' '$3=="major-faults" {print $1; exit}' "$perf_csv" 2>/dev/null || true)
    printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$host" "$stage" "$result_dir" "${avg_cpu:-}" "${max_cpu:-}" "${page_faults:-}" "${minor_faults:-}" "${major_faults:-}" >> "$RUN_DIR/01-manifest.tsv"
  done
 done

python - "$RUN_DIR" <<'PY'
import csv, pathlib, sys
run_dir = pathlib.Path(sys.argv[1])
rows = list(csv.DictReader((run_dir / '01-manifest.tsv').open(), delimiter='\t'))
out = ['# 32 small graph hosting ladder matrix', '']
out.append('| stage | go avg_cpu | python avg_cpu | gst-launch avg_cpu | go page-faults | python page-faults | gst-launch page-faults |')
out.append('|---|---:|---:|---:|---:|---:|---:|')
stages = ['capture', 'convert', 'rate-caps', 'encode', 'parse', 'mux-file']
for stage in stages:
    bucket = {r['host']: r for r in rows if r['stage'] == stage}
    out.append(f"| {stage} | {bucket.get('go',{}).get('avg_cpu','?')} | {bucket.get('python',{}).get('avg_cpu','?')} | {bucket.get('gst-launch',{}).get('avg_cpu','?')} | {bucket.get('go',{}).get('page_faults','?')} | {bucket.get('python',{}).get('page_faults','?')} | {bucket.get('gst-launch',{}).get('page_faults','?')} |")
(run_dir / '02-summary.md').write_text('\n'.join(out) + '\n')
PY

echo "$RUN_DIR"
