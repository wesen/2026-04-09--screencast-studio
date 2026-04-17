---
Title: 27 python manual direct full desktop harness readme
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
Summary: Ticket-local Python control that manually constructs the matched direct full-desktop recording graph for SCS-0016.
LastUpdated: 2026-04-15T03:05:00-04:00
WhatFor: Preserve the manual Python direct-recording control used to compare Python-hosted and Go-hosted GStreamer behavior.
WhenToUse: Read or run this when reproducing the Python manual direct-control experiment in SCS-0016.
---

# 27 python manual direct full desktop harness

This directory contains a Python-hosted direct full-desktop GStreamer control built by explicit element creation and linking.

Files:

- `main.py` — creates and runs the manual pipeline
- `run.sh` — captures CPU, ffprobe output, and dot dumps into a timestamped results directory
