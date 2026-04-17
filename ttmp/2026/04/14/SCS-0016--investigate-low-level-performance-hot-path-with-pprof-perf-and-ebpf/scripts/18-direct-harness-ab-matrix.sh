#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
RESULTS_ROOT="$SCRIPT_DIR/results/18-direct-harness-ab-matrix"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

REPEATS="${REPEATS:-3}"
DURATION_SECONDS="${DURATION_SECONDS:-8}"
DISPLAY_NAME="${DISPLAY_NAME:-${DISPLAY:-:0}}"

HARNESS_15="ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/15-go-direct-record-full-desktop-harness/run.sh"
HARNESS_17="ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/17-go-manual-direct-full-desktop-harness/run.sh"

cd "$REPO_ROOT"
gofmt -w \
  pkg/media/gst/recording.go \
  ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/15-go-direct-record-full-desktop-harness/main.go \
  ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/17-go-manual-direct-full-desktop-harness/main.go

printf 'harness\trepeat\tresult_dir\tavg_cpu\tmax_cpu\tcodec\twidth\theight\tduration\n' > "$RUN_DIR/01-manifest.tsv"

run_one() {
  local harness_name="$1"
  local script_path="$2"
  local repeat="$3"

  echo "running $harness_name repeat $repeat" >&2
  local result_dir
  result_dir=$(DISPLAY_NAME="$DISPLAY_NAME" DURATION_SECONDS="$DURATION_SECONDS" bash "$script_path")

  local summary_file="$result_dir/01-summary.md"
  local ffprobe_file="$result_dir/ffprobe.txt"
  local avg_cpu max_cpu codec width height duration
  avg_cpu=$(awk -F': ' '/avg_cpu:/ {gsub(/%/, "", $2); print $2; exit}' "$summary_file")
  max_cpu=$(awk -F': ' '/max_cpu:/ {gsub(/%/, "", $2); print $2; exit}' "$summary_file")
  codec=$(awk -F'=' '/^codec_name=/ {print $2; exit}' "$ffprobe_file")
  width=$(awk -F'=' '/^width=/ {print $2; exit}' "$ffprobe_file")
  height=$(awk -F'=' '/^height=/ {print $2; exit}' "$ffprobe_file")
  duration=$(awk -F'=' '/^duration=/ {print $2; exit}' "$ffprobe_file")

  printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
    "$harness_name" "$repeat" "$result_dir" "$avg_cpu" "$max_cpu" "$codec" "$width" "$height" "$duration" \
    >> "$RUN_DIR/01-manifest.tsv"
}

for repeat in $(seq 1 "$REPEATS"); do
  run_one "15-copied-direct" "$HARNESS_15" "$repeat"
done

for repeat in $(seq 1 "$REPEATS"); do
  run_one "17-manual-direct" "$HARNESS_17" "$repeat"
done

awk -F'\t' -v duration_seconds="$DURATION_SECONDS" '
  NR == 1 { next }
  {
    name = $1
    avg = $4 + 0
    peak = $5 + 0
    count[name]++
    sum_avg[name] += avg
    if (!(name in min_avg) || avg < min_avg[name]) min_avg[name] = avg
    if (!(name in max_avg) || avg > max_avg[name]) max_avg[name] = avg
    if (!(name in peak_max) || peak > peak_max[name]) peak_max[name] = peak
  }
  END {
    print "# 18 direct harness ab matrix"
    print ""
    printf("- repeats: %d copied, %d manual\n", count["15-copied-direct"], count["17-manual-direct"])
    printf("- duration_seconds: %s\n", duration_seconds)
    print ""
    print "| harness | runs | avg_cpu_mean | avg_cpu_min | avg_cpu_max | max_cpu_peak |"
    print "|---|---:|---:|---:|---:|---:|"
    order[1] = "15-copied-direct"
    order[2] = "17-manual-direct"
    for (i = 1; i <= 2; i++) {
      name = order[i]
      if (!(name in count)) continue
      printf("| %s | %d | %.2f | %.2f | %.2f | %.2f |\n", name, count[name], sum_avg[name] / count[name], min_avg[name], max_avg[name], peak_max[name])
    }
  }
' "$RUN_DIR/01-manifest.tsv" > "$RUN_DIR/02-summary.md"

echo "$RUN_DIR"
