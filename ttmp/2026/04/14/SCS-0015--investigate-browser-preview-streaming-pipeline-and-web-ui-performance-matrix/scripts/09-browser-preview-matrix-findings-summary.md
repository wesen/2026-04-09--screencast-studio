---
Title: 09 browser preview matrix findings summary
Ticket: SCS-0015
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - performance
    - preview
    - browser
    - recording
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: First combined findings note for the browser-preview performance matrix, covering the fresh-server HTTP-client matrix and the live browser-tab measurements.
LastUpdated: 2026-04-14T17:00:00-04:00
WhatFor: Record the first large SCS-0015 performance findings pass in one human-readable place.
WhenToUse: Read before deciding what deeper browser-path instrumentation or optimization direction should come next.
---

# 09 browser preview matrix findings summary

## Result directories used in this note

### Fresh dedicated-server HTTP-client matrix

- `scripts/05-desktop-preview-http-client-recording-matrix/results/20260414-163154/`

### Live browser-tab server sampling

- `scripts/results/20260414-163610/` — desktop, one tab, preview only
- `scripts/results/20260414-163951/` — desktop, one tab, preview + recording
- `scripts/results/20260414-164457/` — desktop, two tabs, preview only
- `scripts/results/20260414-164535/` — desktop, two tabs, preview + recording
- `scripts/results/20260414-164657/` — desktop + camera, one tab, preview only
- `scripts/results/20260414-164720/` — desktop + camera, one tab, preview + recording

## 1. Fresh server-side MJPEG HTTP-client matrix

These runs used fresh dedicated server processes per scenario and plain `curl` MJPEG consumers.

### Preview only

- `preview-no-client` → `15.67%` avg CPU
- `preview-one-client` → `18.11%` avg CPU
- `preview-two-clients` → `19.22%` avg CPU

### Preview + recording

- `record-no-client` → `162.22%` avg CPU
- `record-one-client` → `158.56%` avg CPU
- `record-two-clients` → `165.00%` avg CPU

### Interpretation

This matrix says two important things:

1. The upstream preview itself already costs CPU even with no attached MJPEG client.
2. Simple HTTP-client fan-out alone does **not** explain the user-observed `~400%` CPU recording spike.

The recording cases with `0`, `1`, and `2` plain MJPEG consumers all stayed in roughly the same band: about `159%` to `165%` avg CPU.

## 2. Real browser-tab measurements against the live Studio page

These runs used the real browser-rendered Studio page on `:7777` plus the live browser sampler.

### Desktop only

- one tab, preview only → `14.20%` avg CPU
- one tab, preview + recording → `410.60%` avg CPU
- two tabs, preview only → `12.69%` avg CPU
- two tabs, preview + recording → `432.97%` avg CPU

### Desktop + camera

- one tab, preview only → `20.10%` avg CPU
- one tab, preview + recording → `343.71%` avg CPU

## 3. Per-run metric deltas from the browser-backed scenarios

The older browser result directories were created before `scripts/07-live-server-browser-scenario-sample.sh` wrote `metric-deltas.txt` directly, so the deltas below were computed from the first and last saved `.prom` snapshots in each run directory.

### Desktop, one tab, preview only

- active clients at end: `display=1`
- frames served delta: `71`
- bytes served delta: `23,498,705`

### Desktop, one tab, preview + recording

- active clients at end: `display=1`
- frames served delta: `28`
- bytes served delta: `3,198,438`

### Desktop, two tabs, preview only

- active clients at end: `display=2`
- frames served delta: `142`
- bytes served delta: `46,059,020`

### Desktop, two tabs, preview + recording

- active clients at end: `display=2`
- frames served delta: `57`
- bytes served delta: `6,146,580`

### Desktop + camera, one tab, preview only

- active clients at end: `display=1`, `camera=1`
- display frames served delta: `71`
- display bytes served delta: `16,902,658`
- camera frames served delta: `64`
- camera bytes served delta: `9,204,346`

### Desktop + camera, one tab, preview + recording

- active clients at end: `display=1`, `camera=1`
- display frames served delta: `27`
- display bytes served delta: `2,490,413`
- camera frames served delta: `43`
- camera bytes served delta: `2,544,617`

## 4. Strongest current conclusion

The biggest new SCS-0015 finding is this:

> The **real browser-connected recording path** is dramatically hotter than the fresh-server `curl` MJPEG baseline.

The clearest comparison is:

- fresh server, desktop preview + recording + `1` plain MJPEG client → `158.56%` avg CPU
- live server, desktop preview + recording + `1` real browser tab → `410.60%` avg CPU

That gap is too large to hand-wave away as minor measurement noise.

## 5. Why this matters

The browser-backed recording scenarios did **not** show proportionally huge served-byte deltas. In fact, during recording the per-run served-frame and served-byte deltas were much smaller than in preview-only runs.

That means the new hot path is **not** explained simply by “the server had to send a lot more JPEG bytes.”

Instead, the current evidence points more toward a combination of:

- real browser-tab behavior as the MJPEG consumer,
- the shared preview + recording interaction already known to be expensive upstream,
- and browser-connected lifecycle/streaming behavior that is materially different from a dumb `curl` reader.

## 6. Secondary observations

- The preview did **not** look architecturally dead from the API side: saved `previews-*.json` snapshots showed `lastFrameAt` advancing during the browser recording runs.
- The two-tab desktop recording run (`432.97%`) was somewhat hotter than one-tab desktop recording (`410.60%`), but not by nearly enough to say that duplicate tabs alone explain the whole problem.
- Adding camera changed the shape of the workload: desktop + camera preview-only was hotter than desktop-only preview-only, but desktop + camera one-tab recording (`343.71%`) did not exceed the desktop-only two-tab recording case.

## 7. Current practical interpretation

At this point, the evidence supports a layered conclusion:

1. **Simple MJPEG fan-out matters somewhat**, but it is not the dominant explanation.
2. **The browser-connected recording slice is the missing hot slice** that earlier SCS-0014 API-only and curl-like tests did not fully capture.
3. **Served-byte volume is not the whole story**, because browser recording scenarios were extremely hot even while the saved metric deltas for frames/bytes served were relatively modest.

## 8. Best next steps after this pass

1. Add a true per-run browser-side network/state capture for each sampled scenario, not just server metrics.
2. Add a camera-only one-tab scenario if we want to fully complete the original matrix.
3. Instrument more of the handler path if needed:
   - preview write loop wait time,
   - per-stream frame skip/drop reasons,
   - time spent blocked in MJPEG writes or flushes.
4. Compare browser recording CPU against a headless Chromium-only consumer versus a visible UI tab if we need to separate rendering cost from network/streaming cost.
