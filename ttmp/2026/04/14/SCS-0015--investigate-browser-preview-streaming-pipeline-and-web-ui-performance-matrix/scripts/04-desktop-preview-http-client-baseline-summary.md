---
Title: 04 desktop preview http client baseline summary
Ticket: SCS-0015
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - performance
    - preview
    - browser
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Human-readable summary of the first desktop preview HTTP-client baseline matrix for SCS-0015.
LastUpdated: 2026-04-14T16:14:00-04:00
WhatFor: Record the first 0/1/2-client desktop preview baseline and its initial interpretation.
WhenToUse: Read before expanding the matrix to recording and real browser-tab scenarios.
---

# 04 desktop preview http client baseline summary

Run directory:

- `scripts/03-desktop-preview-http-client-matrix/results/20260414-161024/`

Measured scenarios (`DURATION=4`):

- `no-client`
- `one-client`
- `two-clients`

Observed server CPU from the saved per-scenario summaries:

- `no-client` → `11.67%` avg CPU, `15.00%` max CPU
- `one-client` → `11.50%` avg CPU, `13.00%` max CPU
- `two-clients` → `15.50%` avg CPU, `18.00%` max CPU

## Initial interpretation

This first baseline does **not** yet prove the full web-UI problem. It is an MJPEG HTTP-client approximation, not a real browser-tab matrix. But it already gives a useful signal:

1. Keeping the upstream desktop preview alive with **no** MJPEG client already costs real server CPU.
2. Adding **one** MJPEG client did not materially increase CPU in this short baseline run.
3. Adding **two** MJPEG clients did push CPU higher, which supports the idea that preview-client fan-out can matter.

That makes the next experiment direction clearer:

- add recording on top of this same desktop-preview baseline,
- then compare HTTP-client fan-out against the real browser-tab path,
- and finally see whether the browser UI is hotter because of simple client multiplicity, frontend lifecycle behavior, or something more browser-specific.

## Caveats

- This was a short `4s` baseline.
- It used plain MJPEG HTTP clients (`curl`), not Chromium tabs.
- It only measured desktop preview, not camera preview or combined desktop+camera.
