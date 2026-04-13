#!/usr/bin/env bash
set -euo pipefail

# Investigate why ximagesrc window capture by XID fails while geometry/region capture works.
#
# Usage:
#   ./10-window-preview-investigation.sh [WINDOW_ID...]
#
# If no window IDs are passed, the script uses the first 3 IDs returned by wmctrl -l.
#
# The script records, for each window:
#   - xwininfo summary (title, geometry, map state)
#   - xprop summary (_NET_WM_NAME / WM_CLASS)
#   - gst-launch ximagesrc xid capture result
#   - gst-launch geometry capture result for the same rectangle
#   - ffmpeg -window_id capture result
#   - ffmpeg geometry capture result for the same rectangle
#
# Findings are printed to stdout and saved under /tmp/scs-window-investigation-<timestamp>/

REPORT_ROOT="/tmp/scs-window-investigation-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$REPORT_ROOT"

echo "Report directory: $REPORT_ROOT"

if (($# == 0)); then
  mapfile -t WINDOW_IDS < <((wmctrl -l || true) 2>/dev/null | awk '{print $1}' | head -3)
else
  WINDOW_IDS=("$@")
fi

if ((${#WINDOW_IDS[@]} == 0)); then
  echo "No window IDs available. Pass explicit IDs or ensure wmctrl works." >&2
  exit 1
fi

run_capture() {
  local name="$1"
  shift
  local outfile="$REPORT_ROOT/${name}.log"
  echo "--- $name ---" | tee -a "$REPORT_ROOT/summary.txt"
  echo "+ $*" | tee "$outfile" >/dev/null
  if "$@" >>"$outfile" 2>&1; then
    echo "RESULT: success" | tee -a "$outfile" "$REPORT_ROOT/summary.txt" >/dev/null
    return 0
  else
    local rc=$?
    echo "RESULT: failure (rc=$rc)" | tee -a "$outfile" "$REPORT_ROOT/summary.txt" >/dev/null
    return $rc
  fi
}

extract_field() {
  local file="$1"
  local key="$2"
  awk -F: -v key="$key" '$0 ~ key {sub(/^[[:space:]]+/, "", $2); print $2; exit}' "$file"
}

for raw_id in "${WINDOW_IDS[@]}"; do
  id=$(printf '%s' "$raw_id" | tr 'A-F' 'a-f')
  safe_id=${id//[^0-9a-z]/_}
  echo
  echo "=============================="
  echo "Investigating window $id"
  echo "=============================="

  xwininfo -id "$id" >"$REPORT_ROOT/${safe_id}.xwininfo.txt" 2>&1 || true
  xprop -id "$id" _NET_WM_NAME WM_NAME WM_CLASS >"$REPORT_ROOT/${safe_id}.xprop.txt" 2>&1 || true

  title=$(awk -F'"' '/Window id:/ {print $2; exit}' "$REPORT_ROOT/${safe_id}.xwininfo.txt")
  x=$(extract_field "$REPORT_ROOT/${safe_id}.xwininfo.txt" "Absolute upper-left X")
  y=$(extract_field "$REPORT_ROOT/${safe_id}.xwininfo.txt" "Absolute upper-left Y")
  width=$(extract_field "$REPORT_ROOT/${safe_id}.xwininfo.txt" "Width")
  height=$(extract_field "$REPORT_ROOT/${safe_id}.xwininfo.txt" "Height")
  map_state=$(extract_field "$REPORT_ROOT/${safe_id}.xwininfo.txt" "Map State")

  echo "Title:     ${title:-<unknown>}" | tee -a "$REPORT_ROOT/summary.txt"
  echo "Map state: ${map_state:-<unknown>}" | tee -a "$REPORT_ROOT/summary.txt"
  echo "Geometry:  ${x:-?},${y:-?} ${width:-?}x${height:-?}" | tee -a "$REPORT_ROOT/summary.txt"

  if [[ -n "${x:-}" && -n "${y:-}" && -n "${width:-}" && -n "${height:-}" ]]; then
    endx=$((x + width - 1))
    endy=$((y + height - 1))
  else
    echo "Skipping capture tests for $id because geometry could not be parsed" | tee -a "$REPORT_ROOT/summary.txt"
    continue
  fi

  run_capture "${safe_id}.gst-xid" \
    timeout 6 gst-launch-1.0 -e \
      ximagesrc display-name=:0 xid="$id" use-damage=false show-pointer=true num-buffers=2 \
      ! videoconvert ! jpegenc ! fakesink || true

  run_capture "${safe_id}.gst-xid-remote" \
    timeout 6 gst-launch-1.0 -e \
      ximagesrc display-name=:0 xid="$id" remote=true use-damage=false show-pointer=true num-buffers=2 \
      ! videoconvert ! jpegenc ! fakesink || true

  run_capture "${safe_id}.gst-geometry" \
    timeout 6 gst-launch-1.0 -e \
      ximagesrc display-name=:0 startx="$x" starty="$y" endx="$endx" endy="$endy" use-damage=false num-buffers=2 \
      ! videoconvert ! jpegenc ! fakesink || true

  ffmpeg_xid_out="$REPORT_ROOT/${safe_id}.ffmpeg-window.jpg"
  run_capture "${safe_id}.ffmpeg-xid" \
    ffmpeg -hide_banner -loglevel error -y -f x11grab -window_id "$id" -i :0 -frames:v 1 "$ffmpeg_xid_out" || true

  ffmpeg_geom_out="$REPORT_ROOT/${safe_id}.ffmpeg-geometry.jpg"
  run_capture "${safe_id}.ffmpeg-geometry" \
    ffmpeg -hide_banner -loglevel error -y -f x11grab -video_size "${width}x${height}" -i ":0+${x},${y}" -frames:v 1 "$ffmpeg_geom_out" || true

  if [[ -s "$ffmpeg_xid_out" ]]; then
    echo "ffmpeg xid output:      success ($(stat -c%s "$ffmpeg_xid_out") bytes)" | tee -a "$REPORT_ROOT/summary.txt"
  else
    echo "ffmpeg xid output:      empty/missing" | tee -a "$REPORT_ROOT/summary.txt"
  fi
  if [[ -s "$ffmpeg_geom_out" ]]; then
    echo "ffmpeg geometry output: success ($(stat -c%s "$ffmpeg_geom_out") bytes)" | tee -a "$REPORT_ROOT/summary.txt"
  else
    echo "ffmpeg geometry output: empty/missing" | tee -a "$REPORT_ROOT/summary.txt"
  fi

done

echo
printf 'Summary file: %s\n' "$REPORT_ROOT/summary.txt"
echo
echo "==== Summary ===="
cat "$REPORT_ROOT/summary.txt"
