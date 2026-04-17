---
Title: 30 python manual stage ladder harness readme
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
Summary: Python-hosted manual small-graph ladder control for SCS-0016.
LastUpdated: 2026-04-15T03:20:00-04:00
WhatFor: Preserve the Python manual ladder harness used to compare stage-by-stage host overhead against Go and gst-launch.
WhenToUse: Read or run this when reproducing the Python side of the small-graph hosting ladder in SCS-0016.
---

# 30 python manual stage ladder harness

Stages:

- `capture`
- `convert`
- `rate-caps`
- `encode`
- `parse`
- `mux-file`
