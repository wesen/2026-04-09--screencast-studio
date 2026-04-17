---
Title: 26 python parse launch direct full desktop harness readme
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
Summary: Ticket-local Python control that uses Gst.parse_launch to run the matched direct full-desktop recording graph for SCS-0016.
LastUpdated: 2026-04-15T03:05:00-04:00
WhatFor: Preserve the parse-launch Python control used to compare Python-hosted GStreamer against gst-launch and Go-hosted controls.
WhenToUse: Read or run this when reproducing the Python parse_launch control experiment in SCS-0016.
---

# 26 python parse launch direct full desktop harness

This directory contains a Python-hosted direct full-desktop GStreamer control built with `Gst.parse_launch(...)`.

Files:

- `main.py` — creates and runs the pipeline
- `run.sh` — captures CPU, ffprobe output, and dot dumps into a timestamped results directory
