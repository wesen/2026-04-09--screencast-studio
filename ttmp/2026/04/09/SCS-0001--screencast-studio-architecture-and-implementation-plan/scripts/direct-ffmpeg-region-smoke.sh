#!/usr/bin/env bash
set -euo pipefail

outfile="${1:-/tmp/direct-smoke.mkv}"

rm -f "$outfile"

ffmpeg \
  -hide_banner \
  -loglevel error \
  -y \
  -f x11grab \
  -framerate 2 \
  -video_size 320x240 \
  -draw_mouse 0 \
  -t 1 \
  -i :0.0+0,0 \
  -c:v libx264 \
  -preset veryfast \
  -crf 23 \
  -pix_fmt yuv420p \
  "$outfile"

ls -l "$outfile"
