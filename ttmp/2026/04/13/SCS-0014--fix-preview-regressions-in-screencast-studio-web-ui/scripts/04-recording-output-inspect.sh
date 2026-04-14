#!/usr/bin/env bash
set -euo pipefail

TARGET_DIR="${1:-/home/manuel/code/wesen/2026-04-09--screencast-studio/recordings/demo-2}"
FRAME_BASENAME="${FRAME_BASENAME:-}"
OUT_DIR="${OUT_DIR:-/tmp/scs-recording-inspect}"
mkdir -p "$OUT_DIR"

echo "Inspecting $TARGET_DIR"
shopt -s nullglob
for f in "$TARGET_DIR"/*.mov "$TARGET_DIR"/*.wav; do
  echo "=== $f"
  ffprobe -v error -show_entries format=duration,size -of default=noprint_wrappers=1:nokey=0 "$f"
  ffprobe -v error -count_frames -select_streams v:0 -show_entries stream=codec_name,width,height,avg_frame_rate,nb_read_frames,duration -of default=noprint_wrappers=1:nokey=0 "$f" 2>/dev/null || true
  echo
  if [[ -n "$FRAME_BASENAME" && "$f" == *"$FRAME_BASENAME"* && "$f" == *.mov ]]; then
    ffmpeg -y -loglevel error -ss 0.0 -i "$f" -frames:v 1 "$OUT_DIR/${FRAME_BASENAME}-start.jpg"
    ffmpeg -y -loglevel error -ss 2.5 -i "$f" -frames:v 1 "$OUT_DIR/${FRAME_BASENAME}-mid.jpg"
    ffmpeg -y -loglevel error -ss 4.8 -i "$f" -frames:v 1 "$OUT_DIR/${FRAME_BASENAME}-end.jpg"
    sha256sum "$OUT_DIR/${FRAME_BASENAME}-start.jpg" "$OUT_DIR/${FRAME_BASENAME}-mid.jpg" "$OUT_DIR/${FRAME_BASENAME}-end.jpg"
    file "$OUT_DIR/${FRAME_BASENAME}-start.jpg" "$OUT_DIR/${FRAME_BASENAME}-mid.jpg" "$OUT_DIR/${FRAME_BASENAME}-end.jpg"
  fi
done
