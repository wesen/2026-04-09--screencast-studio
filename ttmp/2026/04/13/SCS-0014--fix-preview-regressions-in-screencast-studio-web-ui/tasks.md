# Tasks

## Ticket setup and delivery

- [x] Create ticket SCS-0014 for the observed preview regressions
- [x] Gather evidence from the running `tmux` server logs
- [x] Gather evidence from current discovery output and relevant code paths
- [x] Write a detailed bug-report document for webcam re-add/device churn
- [x] Write a detailed bug-report document for webcam preview quality/compression
- [x] Write a detailed bug-report document for second desktop capture collapsing onto the root X11 source
- [x] Write a detailed bug-report document for preview-limit races and stale preview recovery
- [x] Add fixing analysis and implementation plan sections to each bug report
- [x] Validate the ticket with `docmgr doctor --ticket SCS-0014 --stale-after 30`
- [x] Upload the bug report bundle to reMarkable under `/ai/2026/04/13/SCS-0014`

## Reproducibility and debugging assets

- [x] Save the retroactive debugging scripts and helper workflows used during SCS-0014 investigation into `scripts/` with numerical prefixes
- [x] Add a restart script for the live `scs-web-ui` tmux server
- [x] Add standalone GStreamer repro scripts for `ximagesrc` coordinate capture vs `videocrop`
- [x] Add API-level preview/recording inspection scripts for screenshot, ffprobe, and preview-freeze polling

## Standalone recording performance investigation

- [x] Add a pure `gst-launch-1.0` recording-performance matrix harness under `scripts/`
- [x] Add a standalone Go shared-bridge recording-performance harness under `scripts/`
- [x] Save raw `pidstat`, `ffprobe`, stdout, and stderr results under the ticket-local `scripts/` folder
- [x] Compare pure GStreamer recording CPU against the current shared-source bridge path for the real `2880x960` region capture shape
- [x] Add a staged standalone benchmark that isolates `appsink`, Go buffer copy, async queueing, `appsrc`, and `x264`
- [x] Save the staged bridge-overhead results under the ticket `scripts/` folder and summarize the findings
- [x] Add a standalone shared-source benchmark that measures preview-only, recorder-only, and preview+recorder together
- [x] Record whether cheaper preview settings materially reduce the combined preview+recorder CPU spike

## Bug-fix themes captured by this ticket

- [x] Camera discovery should stop treating every `/dev/video*` node as a separate end-user camera choice
- [x] Source creation should avoid duplicate logical sources for the same physical camera when that would create confusing UX or output collisions
- [x] Preview rendering quality should be improved or at least made configurable instead of hard-clamped to low-quality JPEG previews
- [x] Per-display source modeling needs to stop collapsing all display previews onto the same root X11 capture source
- [x] Preview release/ensure flows should not temporarily exceed the preview limit during common UI reconfiguration sequences
