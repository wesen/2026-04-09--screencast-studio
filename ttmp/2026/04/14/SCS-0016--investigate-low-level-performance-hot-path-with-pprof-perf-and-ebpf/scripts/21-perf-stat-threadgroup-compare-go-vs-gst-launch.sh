#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
RESULTS_ROOT="$SCRIPT_DIR/results/21-perf-stat-threadgroup-compare-go-vs-gst-launch"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR/go-manual" "$RUN_DIR/gst-launch"

DISPLAY_NAME="${DISPLAY_NAME:-${DISPLAY:-:0}}"
FPS="${FPS:-24}"
QUALITY="${QUALITY:-75}"
BITRATE="${BITRATE:-$((1000 + (QUALITY - 1) * 80))}"
CONTAINER="${CONTAINER:-mov}"
DURATION_SECONDS="${DURATION_SECONDS:-8}"
WARMUP_SECONDS="${WARMUP_SECONDS:-1}"
PERF_BIN="${PERF_BIN:-$(command -v perf || true)}"

if [[ -z "$PERF_BIN" ]]; then
  echo "perf not found in PATH" >&2
  exit 1
fi

ROOT_GEOM="$( (DISPLAY="$DISPLAY_NAME" xwininfo -root 2>/dev/null | awk '/-geometry/ {print $2; exit}') || true )"
if [[ -z "$ROOT_GEOM" ]]; then
  ROOT_GEOM="$(DISPLAY="$DISPLAY_NAME" xdpyinfo 2>/dev/null | awk '/dimensions:/ {print $2; exit}')"
fi
if [[ -z "$ROOT_GEOM" ]]; then
  echo "failed to detect root geometry for DISPLAY=$DISPLAY_NAME" >&2
  exit 1
fi

MUX="qtmux"
case "${CONTAINER,,}" in
  ""|mov|qt) MUX="qtmux" ;;
  mp4) MUX="mp4mux" ;;
  *) echo "unsupported CONTAINER=$CONTAINER" >&2; exit 1 ;;
esac

cd "$REPO_ROOT"
go build -o "$RUN_DIR/go-manual/manual-direct-harness-bin" ./ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/17-go-manual-direct-full-desktop-harness

sample_threads() {
  local pid="$1"
  local out_dir="$2"
  printf 'epoch\tthreads_status\tthreads_ps\n' > "$out_dir/thread-samples.tsv"
  while kill -0 "$pid" 2>/dev/null; do
    local epoch threads_status threads_ps
    epoch="$(date +%s)"
    threads_status="$(awk '/^Threads:/ {print $2; exit}' "/proc/$pid/status" 2>/dev/null || echo '')"
    threads_ps="$(ps -L -p "$pid" --no-headers 2>/dev/null | wc -l | tr -d ' ')"
    printf '%s\t%s\t%s\n' "$epoch" "${threads_status:-}" "${threads_ps:-}" >> "$out_dir/thread-samples.tsv"
    sleep 1
  done
}

run_threadgroup_perf_stat() {
  local pid="$1"
  local out_dir="$2"
  sleep "$WARMUP_SECONDS"
  if ! kill -0 "$pid" 2>/dev/null; then
    echo "process exited before warmup" > "$out_dir/perf-stat.error.txt"
    return 1
  fi
  local tids
  tids="$(ls "/proc/$pid/task" | paste -sd, -)"
  echo "$tids" > "$out_dir/tids.txt"
  "$PERF_BIN" stat -x, -d -d -d -t "$tids" -- sleep "$((DURATION_SECONDS + 1))" \
    > "$out_dir/perf-stat.stdout.log" \
    2> "$out_dir/perf-stat.csv" || true
}

summarize_one() {
  local out_dir="$1"
  python - "$out_dir" <<'PY'
import csv, pathlib, sys
out_dir = pathlib.Path(sys.argv[1])
metrics = {}
perf_file = out_dir / 'perf-stat.csv'
if perf_file.exists():
    for line in perf_file.read_text(errors='replace').splitlines():
        if not line or line.startswith('#'):
            continue
        parts = line.split(',')
        if len(parts) < 3:
            continue
        value = parts[0].strip()
        unit = parts[1].strip()
        event = parts[2].strip()
        if event not in metrics:
            metrics[event] = (value, unit)
thread_file = out_dir / 'thread-samples.tsv'
thread_rows = []
if thread_file.exists():
    reader = csv.DictReader(thread_file.open(), delimiter='\t')
    for row in reader:
        try:
            ts = int(row.get('threads_status') or 0)
            tp = int(row.get('threads_ps') or 0)
        except ValueError:
            continue
        thread_rows.append((ts, tp))
summary = out_dir / '01-summary.md'
lines = []
lines.append(f'# {out_dir.name} perf stat threadgroup summary')
lines.append('')
if thread_rows:
    status_vals = [r[0] for r in thread_rows]
    ps_vals = [r[1] for r in thread_rows]
    lines.append(f'- thread_samples: {len(thread_rows)}')
    lines.append(f'- threads_status_min_max: {min(status_vals)}..{max(status_vals)}')
    lines.append(f'- threads_ps_min_max: {min(ps_vals)}..{max(ps_vals)}')
else:
    lines.append('- thread_samples: 0')
if (out_dir / 'tids.txt').exists():
    tids = (out_dir / 'tids.txt').read_text().strip().split(',')
    lines.append(f'- attached_tids: {len([t for t in tids if t])}')
lines.append('')
lines.append('| metric | value | unit |')
lines.append('|---|---:|---|')
for key in ['task-clock', 'context-switches', 'cpu-migrations', 'page-faults', 'cycles', 'instructions', 'branches', 'branch-misses', 'cache-misses', 'cache-references']:
    if key in metrics:
        value, unit = metrics[key]
        lines.append(f'| {key} | {value} | {unit} |')
summary.write_text('\n'.join(lines) + '\n')
PY
}

run_go_manual() {
  local out_dir="$RUN_DIR/go-manual"
  "$RUN_DIR/go-manual/manual-direct-harness-bin" \
    --display-name "$DISPLAY_NAME" \
    --fps "$FPS" \
    --bitrate "$BITRATE" \
    --container "$CONTAINER" \
    --duration-seconds "$DURATION_SECONDS" \
    --output-path "$out_dir/direct-full-desktop.${CONTAINER}" \
    --dot-dir "$out_dir/dot" \
    > "$out_dir/harness.stdout.json" \
    2> "$out_dir/harness.stderr.log" &
  local pid=$!
  echo "$pid" > "$out_dir/pid"
  sample_threads "$pid" "$out_dir" &
  local sampler_pid=$!
  run_threadgroup_perf_stat "$pid" "$out_dir"
  wait "$pid" || true
  wait "$sampler_pid" || true
  ffprobe -hide_banner -loglevel error \
    -show_entries format=duration,size:stream=codec_name,width,height,avg_frame_rate \
    -of default=noprint_wrappers=1 "$out_dir/direct-full-desktop.${CONTAINER}" > "$out_dir/ffprobe.txt"
  summarize_one "$out_dir"
}

run_gst_launch() {
  local out_dir="$RUN_DIR/gst-launch"
  gst-launch-1.0 -e \
    ximagesrc display-name="$DISPLAY_NAME" use-damage=false show-pointer=true ! \
    videoconvert ! \
    videorate ! \
    "video/x-raw,format=I420,framerate=${FPS}/1,pixel-aspect-ratio=1/1" ! \
    x264enc bitrate="$BITRATE" bframes=0 tune=zerolatency speed-preset=3 ! \
    h264parse ! \
    "$MUX" ! \
    filesink location="$out_dir/direct-full-desktop.${CONTAINER}" \
    > "$out_dir/gst.stdout.log" \
    2> "$out_dir/gst.stderr.log" &
  local pid=$!
  echo "$pid" > "$out_dir/pid"
  sample_threads "$pid" "$out_dir" &
  local sampler_pid=$!
  (
    sleep "$DURATION_SECONDS"
    kill -INT "$pid" 2>/dev/null || true
  ) &
  local stopper_pid=$!
  run_threadgroup_perf_stat "$pid" "$out_dir"
  wait "$pid" || true
  wait "$sampler_pid" || true
  wait "$stopper_pid" || true
  ffprobe -hide_banner -loglevel error \
    -show_entries format=duration,size:stream=codec_name,width,height,avg_frame_rate \
    -of default=noprint_wrappers=1 "$out_dir/direct-full-desktop.${CONTAINER}" > "$out_dir/ffprobe.txt"
  summarize_one "$out_dir"
}

run_go_manual
run_gst_launch

python - "$RUN_DIR" <<'PY'
import pathlib, sys
run_dir = pathlib.Path(sys.argv[1])
lines = ['# 21 perf stat threadgroup compare go vs gst-launch', '']
lines.append('| scenario | codec | width | height | fps | duration | threads_status_min_max | attached_tids | task-clock | context-switches | cpu-migrations | page-faults | cycles | instructions |')
lines.append('|---|---|---:|---:|---|---:|---|---:|---:|---:|---:|---:|---:|---:|')
for label in ['go-manual', 'gst-launch']:
    ffprobe = {}
    for line in (run_dir / label / 'ffprobe.txt').read_text().splitlines():
        if '=' in line:
            k, v = line.split('=', 1)
            ffprobe[k] = v
    summary_lines = (run_dir / label / '01-summary.md').read_text().splitlines()
    kv = {}
    for line in summary_lines:
        if line.startswith('- ') and ': ' in line:
            k, v = line[2:].split(': ', 1)
            kv[k] = v
    table = {}
    for line in summary_lines:
        if line.startswith('| ') and not line.startswith('| metric') and not line.startswith('|---'):
            parts = [p.strip() for p in line.strip('|').split('|')]
            if len(parts) >= 3:
                table[parts[0]] = parts[1]
    lines.append(f"| {label} | {ffprobe.get('codec_name','?')} | {ffprobe.get('width','?')} | {ffprobe.get('height','?')} | {ffprobe.get('avg_frame_rate','?')} | {ffprobe.get('duration','?')} | {kv.get('threads_status_min_max','?')} | {kv.get('attached_tids','?')} | {table.get('task-clock','?')} | {table.get('context-switches','?')} | {table.get('cpu-migrations','?')} | {table.get('page-faults','?')} | {table.get('cycles','?')} | {table.get('instructions','?')} |")
(run_dir / '01-summary.md').write_text('\n'.join(lines) + '\n')
PY

echo "$RUN_DIR"
