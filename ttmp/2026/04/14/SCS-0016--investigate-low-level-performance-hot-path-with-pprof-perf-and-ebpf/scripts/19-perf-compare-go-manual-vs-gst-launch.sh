#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
RESULTS_ROOT="$SCRIPT_DIR/results/19-perf-compare-go-manual-vs-gst-launch"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

DISPLAY_NAME="${DISPLAY_NAME:-${DISPLAY:-:0}}"
FPS="${FPS:-24}"
QUALITY="${QUALITY:-75}"
BITRATE="${BITRATE:-$((1000 + (QUALITY - 1) * 80))}"
CONTAINER="${CONTAINER:-mov}"
DURATION_SECONDS="${DURATION_SECONDS:-8}"
PERF_FREQ="${PERF_FREQ:-99}"
CALL_GRAPH="${CALL_GRAPH:-dwarf,16384}"
PERF_BIN="${PERF_BIN:-$(command -v perf || true)}"
TIMEOUT_BIN="${TIMEOUT_BIN:-$(command -v timeout || true)}"

if [[ -z "$PERF_BIN" ]]; then
  echo "perf not found in PATH" >&2
  exit 1
fi
if [[ -z "$TIMEOUT_BIN" ]]; then
  echo "timeout not found in PATH" >&2
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

GST_PIPELINE="ximagesrc display-name=$DISPLAY_NAME use-damage=false show-pointer=true ! videoconvert ! videorate ! video/x-raw,format=I420,framerate=${FPS}/1,pixel-aspect-ratio=1/1 ! x264enc bitrate=${BITRATE} bframes=0 tune=zerolatency speed-preset=3 ! h264parse ! ${MUX} ! filesink location=$RUN_DIR/gst-launch/direct-full-desktop.${CONTAINER}"

mkdir -p "$RUN_DIR/go-manual" "$RUN_DIR/gst-launch"

cd "$REPO_ROOT"
go build -o "$RUN_DIR/go-manual/manual-direct-harness-bin" ./ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/17-go-manual-direct-full-desktop-harness

cat > "$RUN_DIR/context.txt" <<EOF
DISPLAY_NAME=$DISPLAY_NAME
ROOT_GEOM=$ROOT_GEOM
FPS=$FPS
QUALITY=$QUALITY
BITRATE=$BITRATE
CONTAINER=$CONTAINER
DURATION_SECONDS=$DURATION_SECONDS
PERF_FREQ=$PERF_FREQ
CALL_GRAPH=$CALL_GRAPH
PERF_BIN=$PERF_BIN
perf_event_paranoid=$(cat /proc/sys/kernel/perf_event_paranoid 2>/dev/null || echo unavailable)
uname=$(uname -a)
GST_PIPELINE=$GST_PIPELINE
started_at=$(date --iso-8601=seconds)
EOF

run_perf() {
  local label="$1"
  shift
  local out_dir="$RUN_DIR/$label"
  local perf_data="$out_dir/perf.data"

  "$PERF_BIN" --version > "$out_dir/perf-version.txt"
  {
    echo "label=$label"
    echo "started_at=$(date --iso-8601=seconds)"
    printf 'command='
    printf '%q ' "$@"
    echo
  } > "$out_dir/run-context.txt"

  "$PERF_BIN" record \
    -F "$PERF_FREQ" \
    --call-graph "$CALL_GRAPH" \
    -o "$perf_data" \
    -- "$@" \
    > "$out_dir/perf-record.stdout.log" \
    2> "$out_dir/perf-record.stderr.log"

  "$PERF_BIN" report --stdio -i "$perf_data" \
    > "$out_dir/perf-report.txt" \
    2> "$out_dir/perf-report.stderr.log"

  "$PERF_BIN" report --stdio --sort dso,symbol -i "$perf_data" \
    > "$out_dir/perf-report-dso-symbol.txt" \
    2> "$out_dir/perf-report-dso-symbol.stderr.log"

  awk 'NF > 0 {print} NR >= 80 {exit}' "$out_dir/perf-report-dso-symbol.txt" > "$out_dir/perf-top-80.txt"
}

run_perf "go-manual" \
  "$RUN_DIR/go-manual/manual-direct-harness-bin" \
  --display-name "$DISPLAY_NAME" \
  --fps "$FPS" \
  --bitrate "$BITRATE" \
  --container "$CONTAINER" \
  --duration-seconds "$DURATION_SECONDS" \
  --output-path "$RUN_DIR/go-manual/direct-full-desktop.${CONTAINER}" \
  --dot-dir "$RUN_DIR/go-manual/dot"

ffprobe -hide_banner -loglevel error \
  -show_entries format=duration,size:stream=codec_name,width,height,avg_frame_rate \
  -of default=noprint_wrappers=1 "$RUN_DIR/go-manual/direct-full-desktop.${CONTAINER}" \
  > "$RUN_DIR/go-manual/ffprobe.txt"

run_perf "gst-launch" \
  "$TIMEOUT_BIN" --signal=INT --kill-after=20s "${DURATION_SECONDS}s" \
  gst-launch-1.0 -e \
  ximagesrc display-name="$DISPLAY_NAME" use-damage=false show-pointer=true ! \
  videoconvert ! \
  videorate ! \
  "video/x-raw,format=I420,framerate=${FPS}/1,pixel-aspect-ratio=1/1" ! \
  x264enc bitrate="$BITRATE" bframes=0 tune=zerolatency speed-preset=3 ! \
  h264parse ! \
  "$MUX" ! \
  filesink location="$RUN_DIR/gst-launch/direct-full-desktop.${CONTAINER}"

ffprobe -hide_banner -loglevel error \
  -show_entries format=duration,size:stream=codec_name,width,height,avg_frame_rate \
  -of default=noprint_wrappers=1 "$RUN_DIR/gst-launch/direct-full-desktop.${CONTAINER}" \
  > "$RUN_DIR/gst-launch/ffprobe.txt"

python - "$RUN_DIR" <<'PY'
import pathlib, re, sys
run_dir = pathlib.Path(sys.argv[1])
rows = []
for label in ["go-manual", "gst-launch"]:
    txt = (run_dir / label / "perf-report-dso-symbol.txt").read_text(errors="replace").splitlines()
    top = []
    for line in txt:
        if '%' not in line:
            continue
        s = line.strip()
        if not s or s.startswith('#'):
            continue
        top.append(s)
        if len(top) >= 12:
            break
    ffprobe = {}
    for line in (run_dir / label / "ffprobe.txt").read_text().splitlines():
        if '=' in line:
            k, v = line.split('=', 1)
            ffprobe[k] = v
    rows.append((label, top, ffprobe))
summary = []
summary.append('# 19 perf compare go manual vs gst-launch')
summary.append('')
summary.append('| scenario | codec | width | height | fps | duration |')
summary.append('|---|---|---:|---:|---|---:|')
for label, top, ffprobe in rows:
    summary.append(f"| {label} | {ffprobe.get('codec_name','?')} | {ffprobe.get('width','?')} | {ffprobe.get('height','?')} | {ffprobe.get('avg_frame_rate','?')} | {ffprobe.get('duration','?')} |")
summary.append('')
for label, top, ffprobe in rows:
    summary.append(f'## {label} top mixed-stack entries')
    summary.append('')
    summary.append('~~~text')
    summary.extend(top)
    summary.append('~~~')
    summary.append('')
(run_dir / '01-summary.md').write_text('\n'.join(summary) + '\n')
PY

echo "$RUN_DIR"
