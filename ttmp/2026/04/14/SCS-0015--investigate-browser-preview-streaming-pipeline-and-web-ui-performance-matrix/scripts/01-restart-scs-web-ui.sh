#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT=${REPO_ROOT:-/home/manuel/code/wesen/2026-04-09--screencast-studio}
SESSION=${SESSION:-scs-web-ui}
ADDR=${ADDR:-:7777}
HEALTH_URL=${HEALTH_URL:-http://127.0.0.1:7777/api/healthz}
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
tmux send-keys -t "$SESSION" "go run ./cmd/screencast-studio serve --addr $ADDR" C-m

for _ in $(seq 1 "$WAIT_SECONDS"); do
  if curl -fsS "$HEALTH_URL" >/tmp/scs-healthz.json 2>/dev/null; then
    break
  fi
  sleep 1
done

if [[ ! -s /tmp/scs-healthz.json ]]; then
  echo "server did not become healthy at $HEALTH_URL" >&2
  tmux capture-pane -pt "$SESSION" -S -120 >&2 || true
  exit 1
fi

cat /tmp/scs-healthz.json
echo
echo "--- tmux pane ---"
tmux capture-pane -pt "$SESSION" -S -120
