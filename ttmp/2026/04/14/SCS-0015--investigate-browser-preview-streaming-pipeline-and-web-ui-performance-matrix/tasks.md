# Tasks

## Ticket setup and analysis scaffolding

- [x] Create ticket SCS-0015 for the browser preview streaming investigation
- [x] Write a primary analysis/plan document for the browser preview streaming path
- [x] Start a chronological investigation diary
- [x] Add a report scaffold for final matrix findings

## Current-state architecture mapping

- [x] Map the browser-facing preview transport from GStreamer preview branch to HTTP MJPEG to frontend `<img>` rendering
- [x] Map frontend preview ensure/release ownership logic in the Studio page
- [x] Map the current metrics boundary and identify missing preview-serving observability

## Preview-serving observability

- [x] Add metrics for active MJPEG preview clients
- [x] Add metrics for preview HTTP stream starts and finishes
- [x] Add metrics for preview frames served and preview bytes served
- [x] Add metrics for preview flushes and/or write attempts
- [x] Add metrics for preview ensure/release events in the server path if needed
- [x] Add MJPEG handler loop/idle counters for real-browser timing analysis
- [x] Add cumulative MJPEG handler write/flush timing metrics for real-browser timing analysis
- [x] Keep label cardinality under control and document the chosen label set
- [x] Validate `/metrics` output and add focused tests for the new metric families

## Browser-driven performance matrix

- [x] Create new ticket-local scripts for browser-driven preview/recording measurements under `scripts/`
- [x] Add a clean restart script for the local `scs-web-ui` runtime under this ticket
- [x] Add a metrics sampler script that snapshots `/metrics` over time during runs
- [x] Add a desktop preview HTTP-client baseline harness for the server-side MJPEG streaming path
- [x] Add a desktop-only browser preview matrix harness
- [ ] Add a camera-only browser preview matrix harness
- [x] Add a desktop-plus-camera browser preview matrix harness
- [x] Add a multi-tab browser preview matrix harness
- [x] Save raw `pidstat`, metrics snapshots, network summaries, preview snapshots, stdout, and stderr under the ticket-local `scripts/` directory

## Required measurement scenarios

- [x] Measure desktop preview only with no browser attached, one browser tab, and two browser tabs
- [x] Measure desktop preview + recording with no browser attached, one browser tab, and two browser tabs
- [ ] Measure camera preview only with one browser tab
- [ ] Measure camera preview + recording with one browser tab
- [x] Measure desktop + camera preview only with one browser tab
- [x] Measure desktop + camera preview + recording with one browser tab
- [x] Record whether duplicate browser tabs or stale listeners materially amplify server CPU
- [x] Compare browser-driven measurements against earlier API-only / backend-only results from SCS-0014

## Reporting and conclusions

- [ ] Write the final browser preview streaming performance report
- [x] Create and maintain an ongoing lab report document that backfills the current experiments in detail
- [x] Summarize the raw result directories and matrix outcomes in a human-readable report note
- [x] Decide whether the dominant web-UI-specific cost is upstream preview generation, MJPEG serving fan-out, frontend lifecycle behavior, or a combination
- [x] Propose concrete optimization options ranked by impact and implementation risk
- [x] Validate the ticket with `docmgr doctor --ticket SCS-0015 --stale-after 30`
- [x] Rerun the high-signal desktop preview + recording + one real browser tab scenario with the new MJPEG timing metrics enabled
- [ ] Instrument further upstream of the final MJPEG write path, especially frame-copy/publication work before the HTTP write
