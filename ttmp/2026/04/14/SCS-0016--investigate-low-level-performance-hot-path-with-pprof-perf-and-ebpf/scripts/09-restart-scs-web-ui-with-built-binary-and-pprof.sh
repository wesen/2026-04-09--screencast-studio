#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT=${REPO_ROOT:-/home/manuel/code/wesen/2026-04-09--screencast-studio}
SESSION=${SESSION:-scs-web-ui-pprof}
ADDR=${ADDR:-:7777}
PPROF_ADDR=${PPROF_ADDR:-127.0.0.1:6060}
HEALTH_URL=${HEALTH_URL:-http://127.0.0.1:7777/api/healthz}
PPROF_URL=${PPROF_URL:-http://127.0.0.1:6060/debug/pprof/}
WAIT_SECONDS=${WAIT_SECONDS:-90}
BUILD_DIR=${BUILD_DIR:-$REPO_ROOT/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/bin}
BUILD_PATH=${BUILD_PATH:-$BUILD_DIR/screencast-studio}

if ! command -v tmux >/dev/null 2>&1; then
  echo "tmux is required" >&2
  exit 1
fi

mkdir -p "$BUILD_DIR"
cd "$REPO_ROOT"

go build -o "$BUILD_PATH" ./cmd/screencast-studio

if tmux has-session -t "$SESSION" 2>/dev/null; then
  tmux kill-session -t "$SESSION"
fi

tmux new-session -d -s "$SESSION" "cd '$REPO_ROOT' && exec bash"
tmux send-keys -t "$SESSION" "cd '$REPO_ROOT'" C-m
tmux send-keys -t "$SESSION" "'$BUILD_PATH' serve --addr $ADDR --pprof-addr $PPROF_ADDR" C-m

health_ok=""
pprof_ok=""
for _ in $(seq 1 "$WAIT_SECONDS"); do
  if [[ -z "$health_ok" ]] && curl -fsS "$HEALTH_URL" >/tmp/scs-built-pprof-healthz.json 2>/dev/null; then
    health_ok=1
  fi
  if [[ -z "$pprof_ok" ]] && curl -fsS "$PPROF_URL" >/tmp/scs-built-pprof-index.html 2>/dev/null; then
    pprof_ok=1
  fi
  if [[ -n "$health_ok" && -n "$pprof_ok" ]]; then
    break
  fi
  sleep 1
done

if [[ ! -s /tmp/scs-built-pprof-healthz.json ]]; then
  echo "server did not become healthy at $HEALTH_URL" >&2
  tmux capture-pane -pt "$SESSION" -S -120 >&2 || true
  exit 1
fi

if [[ ! -s /tmp/scs-built-pprof-index.html ]]; then
  echo "pprof server did not become healthy at $PPROF_URL" >&2
  tmux capture-pane -pt "$SESSION" -S -120 >&2 || true
  exit 1
fi

SERVER_PID="$(lsof -tiTCP:${ADDR#:} -sTCP:LISTEN | head -n1 || true)"
EXE_PATH=""
if [[ -n "$SERVER_PID" && -e "/proc/$SERVER_PID/exe" ]]; then
  EXE_PATH="$(readlink -f "/proc/$SERVER_PID/exe" || true)"
fi

cat /tmp/scs-built-pprof-healthz.json
echo
echo "--- pprof ---"
head -n 5 /tmp/scs-built-pprof-index.html || true
echo
echo "--- binary ---"
echo "build_path=$BUILD_PATH"
echo "server_pid=${SERVER_PID:-unknown}"
echo "exe_path=${EXE_PATH:-unknown}"
echo
echo "--- tmux pane ---"
tmux capture-pane -pt "$SESSION" -S -120
