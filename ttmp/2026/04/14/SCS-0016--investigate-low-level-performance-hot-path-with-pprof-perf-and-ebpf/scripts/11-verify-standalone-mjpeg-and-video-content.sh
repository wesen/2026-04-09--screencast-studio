#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
RESULTS_ROOT="$SCRIPT_DIR/results/11-verify-standalone-mjpeg-and-video-content"
STAMP="$(date +%Y%m%d-%H%M%S)"
RUN_DIR="$RESULTS_ROOT/$STAMP"
mkdir -p "$RUN_DIR"

HTTP_ADDR="${HTTP_ADDR:-127.0.0.1:7793}"
WARMUP_SECONDS="${WARMUP_SECONDS:-2}"
RECORD_SECONDS="${RECORD_SECONDS:-8}"
FPS="${FPS:-24}"
PREVIEW_SIZE="${PREVIEW_SIZE:-960x640}"
SOURCE_TYPE="${SOURCE_TYPE:-display}"
DISPLAY_NAME="${DISPLAY_NAME:-${DISPLAY:-}}"
HARNESS_BIN="$RUN_DIR/standalone-harness-bin"
OUTPUT_PATH="$RUN_DIR/output.mp4"
MJPEG_URL="http://$HTTP_ADDR/mjpeg"
HEALTH_URL="http://$HTTP_ADDR/healthz"

cd "$REPO_ROOT"
go build -o "$HARNESS_BIN" ./ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/10-standalone-shared-preview-record-mjpeg-harness

"$HARNESS_BIN" \
  --mode harness \
  --http-addr "$HTTP_ADDR" \
  --source-type "$SOURCE_TYPE" \
  --display-name "$DISPLAY_NAME" \
  --fps "$FPS" \
  --preview-size "$PREVIEW_SIZE" \
  --warmup-seconds "$WARMUP_SECONDS" \
  --record-seconds "$RECORD_SECONDS" \
  --spawn-client=false \
  --output-path "$OUTPUT_PATH" \
  > "$RUN_DIR/harness.stdout.json" \
  2> "$RUN_DIR/harness.stderr.log" &
HARNESS_PID=$!
echo "$HARNESS_PID" > "$RUN_DIR/harness.pid"

for _ in $(seq 1 50); do
  if curl -fsS "$HEALTH_URL" > "$RUN_DIR/healthz.json" 2>/dev/null; then
    break
  fi
  sleep 0.2
done

python3 - <<'PY' "$MJPEG_URL" "$RUN_DIR"
import json
import sys
import time
import urllib.request
from pathlib import Path

url = sys.argv[1]
out_dir = Path(sys.argv[2])
out_dir.mkdir(parents=True, exist_ok=True)

req = urllib.request.Request(url, method='GET')
resp = urllib.request.urlopen(req, timeout=20)
frames = []
start = time.time()
try:
    while len(frames) < 20 and time.time() - start < 8:
        line = resp.readline()
        if not line:
            break
        if not line.startswith(b'--frame'):
            continue
        headers = {}
        while True:
            line = resp.readline()
            if not line:
                break
            if line in (b'\r\n', b'\n'):
                break
            key, _, value = line.decode('utf-8', 'replace').partition(':')
            headers[key.strip().lower()] = value.strip()
        length = int(headers.get('content-length', '0'))
        if length <= 0:
            continue
        payload = resp.read(length)
        _ = resp.readline()
        frames.append(payload)
finally:
    resp.close()

meta = {
    'captured_frames': len(frames),
    'saved': []
}
if frames:
    first = out_dir / 'mjpeg-frame-01.jpg'
    first.write_bytes(frames[0])
    meta['saved'].append(first.name)
if len(frames) >= 10:
    mid = out_dir / 'mjpeg-frame-10.jpg'
    mid.write_bytes(frames[9])
    meta['saved'].append(mid.name)
if len(frames) >= 20:
    later = out_dir / 'mjpeg-frame-20.jpg'
    later.write_bytes(frames[19])
    meta['saved'].append(later.name)
(out_dir / 'mjpeg-capture.json').write_text(json.dumps(meta, indent=2) + '\n')
PY

wait "$HARNESS_PID"

ffprobe -v error -show_entries stream=codec_name,width,height,avg_frame_rate -show_entries format=duration,size -of default=noprint_wrappers=1 "$OUTPUT_PATH" > "$RUN_DIR/ffprobe.txt"
ffmpeg -y -ss 0.5 -i "$OUTPUT_PATH" -frames:v 1 "$RUN_DIR/video-frame-01.png" > "$RUN_DIR/ffmpeg-extract-01.stdout.log" 2> "$RUN_DIR/ffmpeg-extract-01.stderr.log"
ffmpeg -y -ss 4.0 -i "$OUTPUT_PATH" -frames:v 1 "$RUN_DIR/video-frame-02.png" > "$RUN_DIR/ffmpeg-extract-02.stdout.log" 2> "$RUN_DIR/ffmpeg-extract-02.stderr.log"

cat > "$RUN_DIR/01-summary.md" <<EOF
---
Title: 11 verify standalone mjpeg and video content summary
Ticket: SCS-0016
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - backend
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Saved standalone harness JPEG frames and extracted MP4 frames for content verification.
LastUpdated: $(date --iso-8601=seconds)
WhatFor: Preserve image artifacts that let us verify the standalone harness is capturing visually correct content.
WhenToUse: Read when checking whether standalone MJPEG frames and recorded video frames look correct.
---

# 11 verify standalone mjpeg and video content summary

- http_addr: $HTTP_ADDR
- mjpeg_url: $MJPEG_URL
- output_path: $OUTPUT_PATH
- warmup_seconds: $WARMUP_SECONDS
- record_seconds: $RECORD_SECONDS

## Files

- harness.stdout.json
- harness.stderr.log
- healthz.json
- mjpeg-capture.json
- mjpeg-frame-01.jpg
- mjpeg-frame-10.jpg
- mjpeg-frame-20.jpg
- output.mp4
- ffprobe.txt
- video-frame-01.png
- video-frame-02.png
EOF

echo "$RUN_DIR"
