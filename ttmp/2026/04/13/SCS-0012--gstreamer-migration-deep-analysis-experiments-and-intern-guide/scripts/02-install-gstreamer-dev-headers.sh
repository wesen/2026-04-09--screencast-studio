#!/usr/bin/env bash
# 02-install-gstreamer-dev-headers.sh
# Install the -dev packages needed for go-gst (cgo) compilation.
set -euo pipefail

echo "Installing GStreamer dev headers..."
sudo apt-get install -y \
  libgstreamer1.0-dev \
  libgstreamer-plugins-base1.0-dev \
  libgstreamer-plugins-good1.0-dev \
  libgstreamer-plugins-bad1.0-dev \
  libgstreamer-plugins-ugly1.0-dev \
  gstreamer1.0-plugins-base-apps \
  pkg-config

echo ""
echo "=== Verifying pkg-config ==="
pkg-config --modversion gstreamer-1.0
pkg-config --modversion gstreamer-app-1.0
pkg-config --modversion gstreamer-controller-1.0 2>/dev/null || echo "gstreamer-controller not found (may need libgstreamer-plugins-base1.0-dev)"
