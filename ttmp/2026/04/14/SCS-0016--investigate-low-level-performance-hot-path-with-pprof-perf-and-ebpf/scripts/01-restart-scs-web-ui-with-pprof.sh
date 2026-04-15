#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT=${REPO_ROOT:-/home/manuel/code/wesen/2026-04-09--screencast-studio}
SESSION=${SESSION:-scs-web-ui-pprof}
ADDR=${ADDR:-:7777}
PPROF_ADDR=${PPROF_ADDR:-127.0.0.1:6060}
HEALTH_URL=${HEALTH_URL:-http://127.0.0.1:7777/api/healthz}
PPROF_URL=${PPROF_URL:-http://127.0.0.1:6060/debug/pprof/}
WAIT_SECONDS=${WAIT_SECONDS:-90}

if ! command -v tmux >/dev/null 2>&1; then
  echo "tmux is required" >&2
  exit 1
fi

cd "$REPO_ROOT"

if tmux has-session -t "$SESSION" 2>/dev/null; then
  tmux kill-session -t "$SESSION"
fi

tmux new-session -d -s "$SESSION" "cd '$REPO_ROOT' && exec bash"
tmux send-keys -t "$SESSION" "cd '$REPO_ROOT'" C-m
tmux send-keys -t "$SESSION" "go run ./cmd/screencast-studio serve --addr $ADDR --pprof-addr $PPROF_ADDR" C-m

health_ok=""
pprof_ok=""
for _ in $(seq 1 "$WAIT_SECONDS"); do
  if [[ -z "$health_ok" ]] && curl -fsS "$HEALTH_URL" >/tmp/scs-pprof-healthz.json 2>/dev/null; then
    health_ok=1
  fi
  if [[ -z "$pprof_ok" ]] && curl -fsS "$PPROF_URL" >/tmp/scs-pprof-index.html 2>/dev/null; then
    pprof_ok=1
  fi
  if [[ -n "$health_ok" && -n "$pprof_ok" ]]; then
    break
  fi
  sleep 1
done

if [[ ! -s /tmp/scs-pprof-healthz.json ]]; then
  echo "server did not become healthy at $HEALTH_URL" >&2
  tmux capture-pane -pt "$SESSION" -S -120 >&2 || true
  exit 1
fi

if [[ ! -s /tmp/scs-pprof-index.html ]]; then
  echo "pprof server did not become healthy at $PPROF_URL" >&2
  tmux capture-pane -pt "$SESSION" -S -120 >&2 || true
  exit 1
fi

cat /tmp/scs-pprof-healthz.json
echo
echo "--- pprof ---"
head -n 5 /tmp/scs-pprof-index.html || true
echo
echo "--- tmux pane ---"
tmux capture-pane -pt "$SESSION" -S -120
