#!/usr/bin/env bash
set -euo pipefail

ps -ef | rg 'screencast-studio|ffmpeg|screencast-smoke'
