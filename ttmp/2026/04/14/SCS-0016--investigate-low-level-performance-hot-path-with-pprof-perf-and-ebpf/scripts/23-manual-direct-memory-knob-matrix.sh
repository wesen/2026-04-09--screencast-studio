#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
RESULTS_ROOT="$SCRIPT_DIR/results/23-manual-direct-memory-knob-matrix"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

DISPLAY_NAME="${DISPLAY_NAME:-${DISPLAY:-:0}}"
FPS="${FPS:-24}"
BITRATE="${BITRATE:-6920}"
CONTAINER="${CONTAINER:-mov}"
DURATION_SECONDS="${DURATION_SECONDS:-8}"
WARMUP_SECONDS="${WARMUP_SECONDS:-1}"
PERF_BIN="${PERF_BIN:-$(command -v perf || true)}"
PYTHON_BIN="${PYTHON_BIN:-$(command -v python3 || true)}"
HELPER="$SCRIPT_DIR/22-prctl-disable-thp-and-exec.py"

if [[ -z "$PERF_BIN" ]]; then
  echo "perf not found in PATH" >&2
  exit 1
fi
if [[ -z "$PYTHON_BIN" ]]; then
  echo "python3 not found in PATH" >&2
  exit 1
fi

cd "$REPO_ROOT"
go build -o "$RUN_DIR/manual-direct-harness-bin" ./ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/17-go-manual-direct-full-desktop-harness

save_thp_context() {
  local out_file="$1"
  {
    echo "enabled=$(cat /sys/kernel/mm/transparent_hugepage/enabled 2>/dev/null || echo unavailable)"
    echo "defrag=$(cat /sys/kernel/mm/transparent_hugepage/defrag 2>/dev/null || echo unavailable)"
    echo "shmem_enabled=$(cat /sys/kernel/mm/transparent_hugepage/shmem_enabled 2>/dev/null || echo unavailable)"
    echo "use_zero_page=$(cat /sys/kernel/mm/transparent_hugepage/use_zero_page 2>/dev/null || echo unavailable)"
  } > "$out_file"
}

sample_proc() {
  local pid="$1"
  local out_dir="$2"
  mkdir -p "$out_dir/proc-samples"
  while kill -0 "$pid" 2>/dev/null; do
    local epoch sample_dir
    epoch="$(date +%s)"
    sample_dir="$out_dir/proc-samples/$epoch"
    mkdir -p "$sample_dir"
    cat "/proc/$pid/status" > "$sample_dir/status.txt" 2>/dev/null || true
    cat "/proc/$pid/smaps_rollup" > "$sample_dir/smaps_rollup.txt" 2>/dev/null || true
    cat "/proc/$pid/stat" > "$sample_dir/stat.txt" 2>/dev/null || true
    cat "/proc/$pid/statm" > "$sample_dir/statm.txt" 2>/dev/null || true
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
  "$PERF_BIN" stat -x, -e page-faults,minor-faults,major-faults,cycles,instructions,context-switches,cpu-migrations -t "$tids" -- sleep "$((DURATION_SECONDS + 1))" \
    > "$out_dir/perf-stat.stdout.log" \
    2> "$out_dir/perf-stat.csv" || true
}

summarize_one() {
  local out_dir="$1"
  local duration_seconds="$2"
  python - "$out_dir" "$duration_seconds" <<'PY'
import pathlib, sys
out_dir = pathlib.Path(sys.argv[1])
duration_seconds = int(sys.argv[2])
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
pidstat_file = out_dir / 'pidstat.log'
avg_cpu = ''
max_cpu = ''
if pidstat_file.exists():
    vals = []
    for line in pidstat_file.read_text(errors='replace').splitlines():
        parts = line.split()
        if len(parts) == 11 and parts[2].isdigit() and parts[10].startswith('manual-direct'):
            try:
                vals.append(float(parts[8]))
            except ValueError:
                pass
    if vals:
        use = vals[:duration_seconds] if len(vals) >= duration_seconds else vals
        avg_cpu = f"{sum(use) / len(use):.2f}"
        max_cpu = f"{max(use):.2f}"
summary = out_dir / '01-summary.md'
lines = []
lines.append(f'# {out_dir.name} manual direct memory knob summary')
lines.append('')
if avg_cpu:
    lines.append(f'- avg_cpu: {avg_cpu}%')
if max_cpu:
    lines.append(f'- max_cpu: {max_cpu}%')
if (out_dir / 'tids.txt').exists():
    tids = [t for t in (out_dir / 'tids.txt').read_text().strip().split(',') if t]
    lines.append(f'- attached_tids: {len(tids)}')
lines.append('')
lines.append('| metric | value | unit |')
lines.append('|---|---:|---|')
for key in ['page-faults', 'minor-faults', 'major-faults', 'cycles', 'instructions', 'context-switches', 'cpu-migrations']:
    if key in metrics:
        value, unit = metrics[key]
        lines.append(f'| {key} | {value} | {unit} |')
summary.write_text('\n'.join(lines) + '\n')
PY
}

run_scenario() {
  local name="$1"
  shift
  local out_dir="$RUN_DIR/$name"
  mkdir -p "$out_dir"
  save_thp_context "$out_dir/thp-context-before.txt"
  {
    echo "scenario=$name"
    echo "display=$DISPLAY_NAME"
    echo "fps=$FPS"
    echo "bitrate=$BITRATE"
    echo "container=$CONTAINER"
    echo "duration_seconds=$DURATION_SECONDS"
    printf 'command='; printf '%q ' "$@"; echo
    echo "started_at=$(date --iso-8601=seconds)"
  } > "$out_dir/context.txt"

  "$@" \
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
  sample_proc "$pid" "$out_dir" &
  local sample_pid=$!
  pidstat -u -r -p "$pid" 1 "$((DURATION_SECONDS + 3))" > "$out_dir/pidstat.log" &
  local pidstat_pid=$!
  run_threadgroup_perf_stat "$pid" "$out_dir"
  wait "$pid" || true
  wait "$sample_pid" || true
  wait "$pidstat_pid" || true
  save_thp_context "$out_dir/thp-context-after.txt"
  ffprobe -hide_banner -loglevel error \
    -show_entries format=duration,size:stream=codec_name,width,height,avg_frame_rate \
    -of default=noprint_wrappers=1 "$out_dir/direct-full-desktop.${CONTAINER}" > "$out_dir/ffprobe.txt"
  summarize_one "$out_dir" "$DURATION_SECONDS"
}

run_scenario baseline "$RUN_DIR/manual-direct-harness-bin"
run_scenario thp_disable_prctl "$PYTHON_BIN" "$HELPER" "$RUN_DIR/manual-direct-harness-bin"
run_scenario malloc_arena_max_1 env MALLOC_ARENA_MAX=1 "$RUN_DIR/manual-direct-harness-bin"

python - "$RUN_DIR" <<'PY'
import pathlib, sys
run_dir = pathlib.Path(sys.argv[1])
lines = ['# 23 manual direct memory knob matrix', '']
lines.append('| scenario | avg_cpu | max_cpu | page-faults | minor-faults | major-faults | attached_tids |')
lines.append('|---|---:|---:|---:|---:|---:|---:|')
for label in ['baseline', 'thp_disable_prctl', 'malloc_arena_max_1']:
    summary = (run_dir / label / '01-summary.md').read_text().splitlines()
    kv = {}
    table = {}
    for line in summary:
        if line.startswith('- ') and ': ' in line:
            k, v = line[2:].split(': ', 1)
            kv[k] = v
        if line.startswith('| ') and not line.startswith('| metric') and not line.startswith('|---'):
            parts = [p.strip() for p in line.strip('|').split('|')]
            if len(parts) >= 3:
                table[parts[0]] = parts[1]
    lines.append(f"| {label} | {kv.get('avg_cpu','?')} | {kv.get('max_cpu','?')} | {table.get('page-faults','?')} | {table.get('minor-faults','?')} | {table.get('major-faults','?')} | {kv.get('attached_tids','?')} |")
(run_dir / '01-summary.md').write_text('\n'.join(lines) + '\n')
PY

echo "$RUN_DIR"
