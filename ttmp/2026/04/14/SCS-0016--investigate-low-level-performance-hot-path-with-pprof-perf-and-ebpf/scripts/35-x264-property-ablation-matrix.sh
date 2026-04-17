#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_ROOT="$SCRIPT_DIR/results/35-x264-property-ablation-matrix"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

DISPLAY_NAME="${DISPLAY_NAME:-${DISPLAY:-:0}}"
FPS="${FPS:-24}"
BITRATE="${BITRATE:-6920}"
DURATION_SECONDS="${DURATION_SECONDS:-6}"
RUN_TIMEOUT_SECONDS="${RUN_TIMEOUT_SECONDS:-30}"

VARIANTS=(baseline no_tune no_trellis ultrafast)

declare -A HOST_SCRIPT=(
  [go]="$SCRIPT_DIR/29-go-manual-stage-ladder-harness/run.sh"
  [python]="$SCRIPT_DIR/30-python-manual-stage-ladder-harness/run.sh"
  [gst-launch]="$SCRIPT_DIR/31-gst-launch-stage-ladder.sh"
)

variant_env() {
  local variant="$1"
  case "$variant" in
    baseline)
      echo "X264_SPEED_PRESET=3 X264_TUNE=4 X264_BFRAMES=0 X264_TRELLIS=true"
      ;;
    no_tune)
      echo "X264_SPEED_PRESET=3 X264_TUNE=0 X264_BFRAMES=0 X264_TRELLIS=true"
      ;;
    no_trellis)
      echo "X264_SPEED_PRESET=3 X264_TUNE=4 X264_BFRAMES=0 X264_TRELLIS=false"
      ;;
    ultrafast)
      echo "X264_SPEED_PRESET=1 X264_TUNE=4 X264_BFRAMES=0 X264_TRELLIS=false"
      ;;
    *)
      echo "unknown variant $variant" >&2
      return 1
      ;;
  esac
}

printf 'host\tvariant\tx264_speed_preset\tx264_tune\tx264_bframes\tx264_trellis\tresult_dir\tavg_cpu\tmax_cpu\tpage_faults\tminor_faults\tmajor_faults\tresult\terror\n' > "$RUN_DIR/01-manifest.tsv"

for variant in "${VARIANTS[@]}"; do
  env_spec=$(variant_env "$variant")
  for host in go python gst-launch; do
    echo "running host=$host variant=$variant" >&2
    result_dir=""
    avg_cpu=""
    max_cpu=""
    page_faults=""
    minor_faults=""
    major_faults=""
    result=""
    error=""
    if result_dir=$(eval "$env_spec DISPLAY_NAME=\"$DISPLAY_NAME\" FPS=\"$FPS\" BITRATE=\"$BITRATE\" DURATION_SECONDS=\"$DURATION_SECONDS\" STAGE=encode ENCODER=x264enc timeout --signal=INT \"${RUN_TIMEOUT_SECONDS}s\" bash \"${HOST_SCRIPT[$host]}\""); then
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
    elif [[ -n "$result_dir" && -f "$result_dir/stderr.log" ]]; then
      result="gst-launch"
      if rg -q 'ERROR|error|Failed|failed' "$result_dir/stderr.log"; then
        error=$(tr '\n' ' ' < "$result_dir/stderr.log" | sed 's/[[:space:]]\+/ /g' | cut -c1-180)
      fi
    fi
    eval "$env_spec"
    printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$host" "$variant" "$X264_SPEED_PRESET" "$X264_TUNE" "$X264_BFRAMES" "$X264_TRELLIS" "$result_dir" "${avg_cpu:-}" "${max_cpu:-}" "${page_faults:-}" "${minor_faults:-}" "${major_faults:-}" "${result:-}" "${error:-}" >> "$RUN_DIR/01-manifest.tsv"
  done
 done

python3 - "$RUN_DIR" <<'PY'
import csv, pathlib, sys
run_dir = pathlib.Path(sys.argv[1])
rows = list(csv.DictReader((run_dir / '01-manifest.tsv').open(), delimiter='\t'))
out = [
    '---',
    'Title: 35 x264 property ablation matrix summary',
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
    'Summary: Focused encode-stage x264enc ablation across tune, trellis, and speed-preset for Go, Python, and gst-launch hosts.',
    'LastUpdated: 2026-04-15T05:00:00-04:00',
    'WhatFor: Preserve the x264enc-property matrix used to narrow which x264 settings keep or reduce the Go-hosted anomaly.',
    'WhenToUse: Read this when continuing the x264-specific branch of the hosting-gap investigation.',
    '---',
    '',
    '# 35 x264 property ablation matrix',
    '',
]
variants = ['baseline', 'no_tune', 'no_trellis', 'ultrafast']
hosts = ['go', 'python', 'gst-launch']
for variant in variants:
    bucket = [r for r in rows if r['variant'] == variant]
    if bucket:
        meta = bucket[0]
        out.append(f"## {variant} (speed-preset={meta['x264_speed_preset']}, tune={meta['x264_tune']}, bframes={meta['x264_bframes']}, trellis={meta['x264_trellis']})")
        out.append('')
        out.append('| host | avg_cpu | max_cpu | page-faults | minor-faults | major-faults | result | error |')
        out.append('|---|---:|---:|---:|---:|---:|---|---|')
        for host in hosts:
            r = next((x for x in bucket if x['host'] == host), {})
            out.append(f"| {host} | {r.get('avg_cpu','?')} | {r.get('max_cpu','?')} | {r.get('page_faults','?')} | {r.get('minor_faults','?')} | {r.get('major_faults','?')} | {r.get('result','')} | {r.get('error','')} |")
        out.append('')
(run_dir / '02-summary.md').write_text('\n'.join(out) + '\n')
PY

echo "$RUN_DIR"
