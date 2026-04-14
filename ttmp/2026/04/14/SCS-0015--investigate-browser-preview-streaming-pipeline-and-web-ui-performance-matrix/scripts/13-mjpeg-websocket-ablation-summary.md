---
Title: 13 MJPEG websocket ablation summary
Ticket: SCS-0015
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - performance
    - preview
    - browser
    - websocket
    - recording
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Human-readable summary of the focused desktop preview+recording MJPEG-vs-websocket ablation pass.
LastUpdated: 2026-04-14T17:46:00-04:00
WhatFor: Record the result of the focused websocket ablation in a short continuation-friendly note.
WhenToUse: Read before deciding whether to prioritize websocket fanout changes or deeper browser-specific investigation.
---

# 13 MJPEG websocket ablation summary

Result directory:

- `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173541/`

Scenarios:

- `mjpeg-only`
- `mjpeg-plus-ws`

Measured server CPU:

- `mjpeg-only` → `166.56%` avg CPU, `469.00%` max CPU
- `mjpeg-plus-ws` → `170.48%` avg CPU, `481.00%` max CPU

Saved websocket/event evidence from the `mjpeg-plus-ws` case:

- `preview.state` published: `33`
- `preview.state` delivered: `23`
- websocket `preview.state` writes: `23`
- websocket client total messages observed: `54`

## Main conclusion

A plain websocket consumer is **not** enough to explain the browser-path jump to `~410%` in the real one-tab desktop preview+recording scenario.

It is a real contributor, but it only increased avg CPU by about `3.92` points in this focused fresh-server ablation:

- from `166.56%`
- to `170.48%`

That means the remaining unexplained heat is more likely tied to browser-specific behavior that this synthetic MJPEG-plus-websocket client pair still does not reproduce.

## Important caveat

The first attempt at this ablation produced an invalid `mjpeg-only` comparison because the harness started recording before the preview path had clearly produced an initial frame. That bad first run lives under:

- `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173359/`

The trusted run is the rerun after the harness fix:

- `scripts/12-desktop-preview-recording-mjpeg-ws-ablation-matrix/results/20260414-173541/`
