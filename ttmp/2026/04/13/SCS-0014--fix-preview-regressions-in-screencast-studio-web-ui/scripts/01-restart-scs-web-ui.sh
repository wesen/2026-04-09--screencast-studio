#!/usr/bin/env bash
set -euo pipefail

SESSION="${SESSION:-scs-web-ui}"
REPO="${REPO:-/home/manuel/code/wesen/2026-04-09--screencast-studio}"
ADDR="${ADDR:-:7777}"

if tmux has-session -t "$SESSION" 2>/dev/null; then
  tmux kill-session -t "$SESSION"
fi

cd "$REPO"
tmux new-session -d -s "$SESSION" "cd '$REPO' && go run ./cmd/screencast-studio serve --addr $ADDR"

for i in $(seq 1 40); do
  if curl -fsS "http://127.0.0.1${ADDR/:/}:7777/api/healthz" >/tmp/scs-healthz.json 2>/dev/null; then
    cat /tmp/scs-healthz.json
    exit 0
  fi
  if curl -fsS "http://127.0.0.1:7777/api/healthz" >/tmp/scs-healthz.json 2>/dev/null; then
    cat /tmp/scs-healthz.json
    exit 0
  fi
  sleep 1
done

echo "server did not become healthy in time" >&2
TMUX_PANE=$(tmux list-panes -t "$SESSION" -F '#{pane_id}' | head -n1)
tmux capture-pane -pt "$TMUX_PANE" || true
exit 1
