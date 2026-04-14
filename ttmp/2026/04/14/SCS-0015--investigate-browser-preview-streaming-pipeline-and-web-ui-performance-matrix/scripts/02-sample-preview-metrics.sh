#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
TICKET_DIR=$(cd -- "$SCRIPT_DIR/.." && pwd)
DEFAULT_RESULTS_DIR="$SCRIPT_DIR/results"
mkdir -p "$DEFAULT_RESULTS_DIR"

DURATION_SECONDS=${DURATION_SECONDS:-10}
INTERVAL_SECONDS=${INTERVAL_SECONDS:-1}
METRICS_URL=${METRICS_URL:-http://127.0.0.1:7777/metrics}
OUTPUT_DIR=${OUTPUT_DIR:-$DEFAULT_RESULTS_DIR/$(date +%Y%m%d-%H%M%S)}

mkdir -p "$OUTPUT_DIR/raw"
MANIFEST="$OUTPUT_DIR/01-manifest.tsv"
SUMMARY="$OUTPUT_DIR/02-summary.txt"

printf "sample_index\ttimestamp_epoch\ttimestamp_iso\tpath\n" > "$MANIFEST"

if ! curl -fsS "$METRICS_URL" >/dev/null 2>&1; then
  echo "metrics endpoint not reachable at $METRICS_URL" >&2
  exit 1
fi

sample_count=$(python3 - <<'PY' "$DURATION_SECONDS" "$INTERVAL_SECONDS"
import math, sys

duration = float(sys.argv[1])
interval = float(sys.argv[2])
if interval <= 0:
    raise SystemExit("interval must be > 0")
print(max(1, int(math.ceil(duration / interval))))
PY
)

for i in $(seq 1 "$sample_count"); do
  ts_epoch=$(date +%s)
  ts_iso=$(date --iso-8601=seconds)
  out_file=$(printf "%s/raw/%03d-%s.prom" "$OUTPUT_DIR" "$i" "$ts_epoch")
  curl -fsS "$METRICS_URL" -o "$out_file"
  printf "%s\t%s\t%s\t%s\n" "$i" "$ts_epoch" "$ts_iso" "$out_file" >> "$MANIFEST"
  if [[ "$i" -lt "$sample_count" ]]; then
    sleep "$INTERVAL_SECONDS"
  fi
done

{
  echo "metrics_url: $METRICS_URL"
  echo "duration_seconds: $DURATION_SECONDS"
  echo "interval_seconds: $INTERVAL_SECONDS"
  echo "sample_count: $sample_count"
  echo "output_dir: $OUTPUT_DIR"
} > "$SUMMARY"

echo "$OUTPUT_DIR"
