# Changelog

## 2026-04-13

Created ticket **SCS-0014** to capture preview regressions observed during live manual testing of the running screencast-studio web UI. Wrote four detailed bug reports covering:

1. webcam preview removal/re-add instability and device-node churn,
2. webcam preview quality/format regression,
3. second desktop capture collapsing onto the root X11 display source,
4. preview-limit races and stale preview recovery failures.

Each bug report includes observed behavior, evidence, probable root cause, a fixing analysis, and an implementation plan so the ticket can move directly from reproduction into engineering work.

Validated the ticket with `docmgr doctor --ticket SCS-0014 --stale-after 30` and uploaded the bundle to reMarkable as **SCS-0014 Preview Regression Bug Reports**, verified under `/ai/2026/04/13/SCS-0014`.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/discovery/service.go — camera enumeration behavior implicated in duplicate camera choices and device-node instability; later extended with root-geometry discovery for region/window cropping
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go — shared source signatures and preview JPEG generation implicated in multiple preview bugs
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/preview.go — region/window preview capture path moved from ximagesrc coordinate cropping toward full-root capture plus videocrop
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go — limit accounting and release/ensure timing implicated in preview-limit races
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/conversion.ts — source picker duplication behavior implicated in repeated camera-source creation

Implemented and validated several fixes after the initial ticket write-up:

1. camera discovery deduplication and duplicate-camera add prevention,
2. source-aware preview quality tuning,
3. preview-limit race hardening,
4. full postmortem and runtime investigation for the window/region full-display bug.

The window/region investigation established that standalone `ximagesrc` region-coordinate capture could still visually show the full desktop while matching the requested output dimensions, whereas full-root capture plus `videocrop` produced a true crop. The active runtime fix is therefore shifting region/window capture toward explicit cropping rather than trusting `ximagesrc` coordinate semantics.

Later in the same ticket, I added standalone recording-performance harnesses and saved the raw results under `scripts/` so the recording CPU spike could be measured without modifying the main server path. The saved results show that for the real `2880x960` region shape, pure `gst-launch-1.0` recording with the current x264 settings already costs about `86.50%` average CPU, a faster x264 preset drops that to about `49.83%`, and the current shared-source Go bridge path is much more expensive still (`139.57%` avg CPU recorder-only, `151.62%` avg CPU preview+recorder). That makes the encoder configuration and current shared raw-consumer/appsrc bridge the main performance suspects rather than idle server behavior.

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/06-gst-recording-performance-matrix/run.sh — standalone pure-GStreamer benchmark harness for capture, preview-like JPEG, and direct x264 recording
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/run.sh — standalone benchmark harness for the current shared-source Go bridge path
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/main.go — scenario runner for preview-only, recorder-only, and preview+recorder shared-bridge measurements
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/08-recording-performance-measurements-summary.md — human-readable summary of the saved benchmark results

After that first measurement pass, I added a second staged benchmark suite under `scripts/09-go-bridge-overhead-matrix/` to isolate the cost of the path that looks like `normalized raw source -> appsink -> Go callback/copy/queue -> appsrc -> downstream sink`. That benchmark saved its raw results under `scripts/09-go-bridge-overhead-matrix/results/20260413-232943/` and its higher-level summary under `scripts/10-go-bridge-overhead-measurements-summary.md`.

The staged results were surprisingly useful: the normalized baseline was about `24.50%` average CPU, `appsink` discard was about `25.33%`, `appsink + buffer.Copy()` discard about `28.33%`, `appsink + copy + async queue + appsrc -> fakesink` about `24.33%`, and only when `x264` was added did the CPU jump sharply again (`77.83%` avg CPU). That suggests the bridge machinery alone is much cheaper than the encoder for this tested `2880x960 @ 24 fps` region shape, even though the earlier full shared-runtime benchmark remains higher and still needs reconciliation.

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/run.sh — staged standalone benchmark driver for raw baseline, appsink, Go copy, async queue, appsrc, and x264 scenarios
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/main.go — scenario runner implementing the staged bridge-overhead pipelines
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/10-go-bridge-overhead-measurements-summary.md — human-readable interpretation of the staged bridge-overhead benchmark

To reconcile the earlier disagreement between benchmark families, I then added a same-session wrapper suite under `scripts/11-go-shared-vs-bridge-reconciliation-matrix/`. That suite reran the direct GStreamer benchmark (`06`), the shared-runtime benchmark (`07`), and the staged bridge-overhead benchmark (`09`) back-to-back under the same `2880x960 @ 24 fps` region shape and recorded the unified result under `scripts/11-go-shared-vs-bridge-reconciliation-matrix/results/20260413-233847/01-summary.md`.

The key reconciliation result is that the earlier extreme mismatch did **not** reproduce. In the same-session run:

- `06` direct current x264 record measured about `94.33%` average CPU,
- `07` shared-runtime recorder-only measured about `94.00%`,
- `09` staged `appsink -> Go -> appsrc -> x264` measured about `91.50%`,
- while `07` preview + recorder together remained substantially higher at about `131.00%` average CPU.

That changes the engineering interpretation. The bridge alone no longer looks like the dominant recorder-only cost center. The more reliable current conclusion is that recorder-only cost is broadly aligned across direct encode, the shared runtime, and the staged bridge+x264 case, while the more clearly expensive combined case is **preview + recorder together**.

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/reference/02-gstreamer-setup-performance-and-region-debugging-project-report.md — copied ticket-local version of the longer Obsidian project report so the knowledge also lives inside the ticket workspace
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/11-go-shared-vs-bridge-reconciliation-matrix/run.sh — wrapper benchmark that reruns `06`, `07`, and `09` in one same-session matrix
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/11-go-shared-vs-bridge-reconciliation-matrix/results/20260413-233847/01-summary.md — reconciled same-session summary across direct, shared-runtime, and staged bridge benchmark families

After the reconciliation run, I followed the new leading clue and added `scripts/12-go-preview-recorder-interplay-matrix/` plus the higher-level note `scripts/13-preview-recorder-interplay-summary.md`. This new standalone benchmark keeps the same shared source shape but measures preview-only, recorder-only, preview+recorder with current preview settings, and preview+recorder with a cheaper preview profile.

That benchmark produced the clearest current performance result in the ticket so far:

- preview-only: about `12.17%` average CPU,
- recorder-only: about `94.00%`,
- current preview + recorder: about `188.43%`,
- cheap preview + recorder: about `170.00%`.

So the remaining large cost spike is now strongly localized to the **combined preview + recorder case**. Cheaper preview settings help somewhat, but they do not eliminate the spike. That suggests the main remaining performance problem is the interaction of the live preview branch with the recorder branch on the same shared source, not recorder-only bridge overhead.

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/main.go — standalone shared-source benchmark for preview-only, recorder-only, and combined preview+recorder scenarios
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/run.sh — runner for the preview/recorder interplay benchmark
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/results/20260414-070646/01-summary.md — raw saved summary for the interplay benchmark run
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/13-preview-recorder-interplay-summary.md — human-readable interpretation of the interplay benchmark

The next standalone step was a more targeted ablation suite under `scripts/14-go-preview-branch-ablation-matrix/` plus the summary note `scripts/15-preview-branch-ablation-summary.md`. That benchmark kept the recorder path fixed and varied only the preview branch to separate:

- second active branch cost (`preview-fakesink-plus-recorder`),
- JPEG cost without Go frame copying (`preview-jpeg-discard-plus-recorder`),
- raw frame-copy cost without JPEG (`preview-raw-copy-plus-recorder`),
- the current real preview path (`preview-current-plus-recorder`),
- and a cheap preview profile (`preview-cheap-plus-recorder`).

This run reinforced the current direction: within that benchmark, every non-trivial preview branch was more expensive than recorder-only, and the current real preview path was more expensive than the stripped-down preview variants. A very cheap preview profile reduced total CPU materially. The absolute recorder-only baseline in that run was higher than the earlier reconciled recorder-only baseline, so the safest reading is relative within-run rather than replacing the earlier absolute baseline. Still, the ablation result supports the idea that the next most practical optimization target is **degrading preview work while recording** before attempting more invasive recorder-architecture changes.

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/main.go — standalone benchmark that isolates second-branch overhead, JPEG, raw frame-copy, and current preview-path cost while recording
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/run.sh — runner for the preview-branch ablation benchmark
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/results/20260414-081748/01-summary.md — raw saved summary for the ablation benchmark run
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/15-preview-branch-ablation-summary.md — human-readable interpretation of the ablation benchmark
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/reference/03-research-brief-preview-and-recording-performance-investigation-handoff.md — copied ticket-local version of the detailed researcher handoff brief from the Obsidian vault so the research package also lives inside the ticket workspace
