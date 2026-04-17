# Changelog

## 2026-04-14

Created ticket **SCS-0016** to isolate the lower-level profiling work from SCS-0015.

The reason for splitting the work is that SCS-0015 already narrowed several plausible browser-path suspects with app-level instrumentation:

- final MJPEG HTTP write/flush time looks too small to explain the hot phase,
- PreviewManager cached-frame copy/store/publication time looks too small,
- EventHub publish time also looks too small.

That means the remaining unexplained cost likely lives lower in the stack, closer to CGO, GStreamer buffer handoff, appsink callbacks, memcpy-heavy transitions, runtime scheduling, or some combination of those.

Started the new ticket documents:

- `design-doc/01-low-level-profiling-plan.md`
- `reference/01-investigation-diary.md`

The initial plan for this ticket is intentionally staged:

1. **Go pprof first** to answer whether the hot phase is still largely visible in Go userland.
2. **perf second** if pprof mainly points at `runtime.cgocall` or otherwise fails to explain the hot phase.
3. **eBPF third** only for narrowly targeted unanswered questions such as off-CPU, scheduler, syscall, or socket behavior.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go — upstream ticket already showed the final MJPEG write path is not the dominant explanation
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go — upstream ticket already measured PreviewManager frame-store/copy/publication timing
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/event_hub.go — upstream ticket already measured EventHub publish timing
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go — likely next low-level boundary around GStreamer appsink callbacks and frame handoff into Go
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/serve.go — likely place to add optional local profiling enablement
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/reference/02-browser-preview-streaming-lab-report.md — parent evidence trail that justified this new lower-level profiling ticket

Implemented the first real SCS-0016 code slice in commit `d06957a0318d1294609b35e8673ca307fad4bbab` (`Add optional pprof profiling support`).

This slice added:

- optional `--pprof-addr` support on `screencast-studio serve`
- a separate pprof HTTP server that stays disabled when the flag is empty
- a focused `internal/web/pprof.go` mux and `internal/web/pprof_test.go`
- ticket-local helper scripts:
  - `scripts/01-restart-scs-web-ui-with-pprof.sh`
  - `scripts/02-capture-pprof-cpu-profile.sh`

The first live pprof-enabled restart worked and exposed:

- app health at `http://127.0.0.1:7777/api/healthz`
- pprof index at `http://127.0.0.1:6060/debug/pprof/`

I then used the same high-signal repro — desktop preview + recording + one real browser tab — to capture the first CPU profile artifact under:

- `scripts/results/02-capture-pprof-cpu-profile/20260414-204800/`

That first pprof pass is already informative. The top output is dominated by external/native code rather than clear Go-userland hotspots:

- `<unknown>` → `67.74%`
- `[libgstvideo-1.0.so.0.2402.0]` → `15.77%`
- `[orcexec.Mmdcyh]` → `7.66%`
- `[libc.so.6]` → `5.21%`
- `[libjpeg.so.8.2.2]` → `1.31%`
- `runtime._ExternalCode` → `98.65%` cumulative

That means pprof did its first job well: it strongly suggests the hot phase is not mainly sitting in ordinary Go userland, which in turn justifies moving toward `perf` rather than continuing to add only Go-level counters.

Two additional practical findings came out of this first profiling slice:

1. The first version of the pprof capture script accidentally used Bash’s special `SECONDS` variable as the capture parameter name. That caused the summary to report `seconds: 41` even though I launched it intending a shorter capture window. This is a script bug to keep in mind for follow-up cleanup.
2. Lower-level native tools are installed but currently privilege-blocked for the current user:
   - `perf` is installed but blocked by `perf_event_paranoid=4`
   - `bpftrace` is installed but requires root

While waiting on `perf` access, I also wrote two continuation-friendly reports for this ticket:

- `reference/02-performance-investigation-approaches-and-tricks-report.md`
- `reference/03-prometheus-metrics-architecture-and-field-guide.md`

These reports summarize both the practical investigation playbook and the current metrics architecture so the profiling phase is easier to hand off and continue.

I then added a reproducible profiler-prereq check script:

- `scripts/03-check-profiler-prereqs.sh`

and saved the first prereq result under:

- `scripts/results/03-check-profiler-prereqs/20260414-211300/`

That follow-up mattered because the machine state had changed since the first quick manual check. The saved result shows:

- `perf_event_paranoid=1`
- `perf stat -p <server-pid> sleep 1` now works for the current user
- `bpftrace` is still installed but still fails without root

This means the ticket is no longer waiting on hypothetical `perf` access. The next practical step can be a real `perf record` capture for the same high-signal browser repro.

I also fixed the pprof capture script bug where Bash’s special `SECONDS` variable had been used as the capture parameter name. The script now uses `PROFILE_SECONDS`, which keeps future summaries honest.

With `perf` now usable, I added the first mixed-stack capture harness:

- `scripts/04-capture-perf-cpu-profile.sh`

For reproducibility, I also backfilled the exact browser-driving helpers and the Go-address resolver used during this slice into the ticket-local scripts directory:

- `scripts/05-open-studio-and-wait-desktop.js`
- `scripts/06-start-recording.js`
- `scripts/07-stop-recording.js`
- `scripts/08-resolve-perf-go-addresses.sh`

I then used that flow against the same high-signal repro — desktop preview + recording + one real browser tab — and saved the first `perf` capture under:

- `scripts/results/04-capture-perf-cpu-profile/20260414-224952/`

The first `perf` result is more concrete than the earlier pprof output. It points strongly at native recording work in `libx264`, with visible GStreamer push-path frames leading into the encoder:

- `libx264.so.164 x264_8_trellis_coefn` ≈ `45.63%` children / `44.49%` self
- additional unresolved native frames under `[unknown]` ≈ `27.84%`, likely still part of the same native-heavy region
- `libgstreamer-1.0.so.0.2402.0 gst_pad_push` and nearby pad-push frames are visible in the call paths into `x264_encoder_encode`

That is already useful because it shifts the working explanation from a vague “external/native” bucket toward a more specific picture: the dominant hot path in this repro is heavily in **native encoder + pipeline push work**, not primarily in the final MJPEG HTTP write path or the previously measured Go-side web/event bookkeeping.

The mixed-stack symbol quality is only partially solved, though. Native libraries are visible enough to be actionable, but the main Go binary still shows many address-only frames because the process is running from a temporary `go run` executable. I used the new resolver helper to map the top `screencast-studio` addresses anyway, and that backfilled evidence shows the visible Go-side frames cluster around runtime CGO callback machinery and some `zerolog` / `encoding/json` paths rather than a single clean application function.

Per the current instruction, the raw `perf.data` artifact is being kept as a local saved result, but it does not need to be committed. The lighter text artifacts and ticket docs are the important continuation surface for source control.

I then improved the reproducibility and symbol-quality story one more step by adding:

- `scripts/09-restart-scs-web-ui-with-built-binary-and-pprof.sh`

This helper builds a stable `screencast-studio` binary under the ticket-local `scripts/bin/` directory and restarts the app from that path instead of `go run`. I used it to produce a second perf capture under:

- `scripts/results/04-capture-perf-cpu-profile/20260414-230415/`

That rerun improved the main-binary symbolization enough that `perf-report-dso-symbol.txt` now shows named Go/runtime and websocket/server functions directly instead of mostly address-only `screencast-studio` frames. The rerun also reduced perf writer wakeups dramatically:

- first run (`go run`): `206800` wakeups, `6429` samples, about `101.983 MB`
- built-binary rerun: `669` wakeups, `10893` samples, about `172.843 MB`

The main interpretation did **not** change in the way that would rescue the old Go-side hypothesis. The stable-binary rerun still points primarily at native media work:

- `libx264.so.164 x264_8_trellis_coefn` ≈ `41.74%` children / `40.93%` self
- `[unknown]` native frames ≈ `25.75%`
- `libgstreamer-1.0.so.0.2402.0 gst_pad_push` ≈ `13.68%`
- `libgstreamer` / `gst_buffer_copy_into` and libc `__memcpy_evex_unaligned_erms` are visible in the hot path into `x264_encoder_encode`

The newly visible `screencast-studio` symbols are real but small by comparison. They include runtime scheduling/CGO callback paths and websocket write paths, which is useful for confidence, but they do not overturn the main conclusion that the dominant cost is still overwhelmingly in the native encoder/pipeline region rather than ordinary Go web code.

I also fixed a small helper bug uncovered by this rerun: `scripts/08-resolve-perf-go-addresses.sh` originally assumed the report would contain address-only `screencast-studio` frames. With better symbolization that assumption breaks, so the helper now emits a clear note when addr2line fallback is unnecessary because the perf report already contains direct symbols.

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/pprof.go — separate pprof mux for the optional debug server
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/pprof_test.go — focused coverage that the pprof mux serves a valid index
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/serve.go — `serve` now accepts the optional `--pprof-addr` flag
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go — runtime now starts and stops the separate pprof server when configured
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/01-restart-scs-web-ui-with-pprof.sh — restarts the app with pprof enabled and validates both health and pprof index reachability
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/02-capture-pprof-cpu-profile.sh — saves raw pprof CPU profiles and top summaries
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/03-check-profiler-prereqs.sh — saves the current local availability and permission state for perf, bpftrace, and bpftool
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/02-capture-pprof-cpu-profile/20260414-204800/pprof-top.txt — first saved pprof top output showing the dominance of native/external code
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/03-check-profiler-prereqs/20260414-211300/01-summary.md — saved prereq result showing working perf access for the current user and root-only bpftrace
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/04-capture-perf-cpu-profile.sh — mixed-stack perf capture helper for the live server PID
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/05-open-studio-and-wait-desktop.js — ticket-local browser helper used to establish the desktop preview repro
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/06-start-recording.js — ticket-local browser helper used to enter the recording hot phase
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/07-stop-recording.js — ticket-local browser helper used to end the hot-phase repro cleanly
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/08-resolve-perf-go-addresses.sh — helper that extracts top Go-binary addresses from the perf report and resolves them with `go tool addr2line`, now also handling the already-symbolized case cleanly
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/09-restart-scs-web-ui-with-built-binary-and-pprof.sh — stable-binary restart helper used to improve main-binary symbolization for the second perf rerun
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/04-capture-perf-cpu-profile/20260414-224952/01-summary.md — first mixed-stack perf capture summary for the browser one-tab recording repro
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/04-capture-perf-cpu-profile/20260414-224952/go-addr2line.txt — first saved address-resolution artifact for the top `screencast-studio` frames in the perf report
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/04-capture-perf-cpu-profile/20260414-230415/01-summary.md — second perf summary from the stable built binary rerun
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/04-capture-perf-cpu-profile/20260414-230415/perf-report-dso-symbol.txt — second perf report with much better direct `screencast-studio` symbolization
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/04-capture-perf-cpu-profile/20260414-230415/go-addr2line.txt — fallback note artifact showing direct perf symbolization made addr2line unnecessary in the stable-binary rerun
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/reference/02-performance-investigation-approaches-and-tricks-report.md — project report on the investigation playbook used so far
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/reference/03-prometheus-metrics-architecture-and-field-guide.md — project report on the current metrics architecture

## 2026-04-15

Wrote the direct-recording hosting-gap investigation report and the online research query packet after narrowing the remaining gap to a likely native hosting / memory-fault problem rather than a graph-construction bug; also saved matched gst-launch, Go-harness A/B, mixed-stack perf, and threadgroup perf-stat comparison evidence.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/reference/04-direct-recording-hosting-gap-investigation-report.md — Main findings report for the current hosting-gap interpretation
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/reference/05-online-research-query-packet-for-go-hosted-gstreamer-performance.md — Copy/paste-ready external research packet
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/21-perf-stat-threadgroup-compare-go-vs-gst-launch/20260415-012510/01-summary.md — Current strongest evidence for the large Go-hosted page-fault delta

Added a new stage-by-stage debugging plan and executed the first full small-graph hosting ladder to localize the first real Go-hosted divergence point.

New planning and control files:

- `design-doc/02-small-graph-hosting-ladder-debugging-plan.md`
- `scripts/29-go-manual-stage-ladder-harness/main.go`
- `scripts/29-go-manual-stage-ladder-harness/run.sh`
- `scripts/30-python-manual-stage-ladder-harness/main.py`
- `scripts/30-python-manual-stage-ladder-harness/run.sh`
- `scripts/31-gst-launch-stage-ladder.sh`
- `scripts/32-small-graph-hosting-ladder-matrix.sh`

Saved the first full ladder matrix under:

- `scripts/results/32-small-graph-hosting-ladder-matrix/20260415-033745/`

The matrix compared six stages across three hosts:

- `capture`
- `convert`
- `rate-caps`
- `encode`
- `parse`
- `mux-file`

with host families:

- Go manual
- Python manual
- `gst-launch`

The result is much sharper than the earlier full-graph suspicion. Go stays essentially aligned with Python and `gst-launch` through the pre-encode ladder:

- `capture`: Go `1.00%`, Python `2.16%`, `gst-launch` `1.00%`
- `convert`: Go `1.00%`, Python `2.17%`, `gst-launch` `1.00%`
- `rate-caps`: Go `37.83%`, Python `36.83%`, `gst-launch` `36.67%`

The first strong divergence appears at the encoder boundary:

- `encode`: Go `199.97%`, Python `125.17%`, `gst-launch` `129.00%`
- `parse`: Go `170.50%`, Python `126.50%`, `gst-launch` `128.44%`
- `mux-file`: Go `197.17%`, Python `131.17%`, `gst-launch` `131.11%`

The page-fault signal lines up with the same stage boundary. The pre-encode stages stay tiny for all three hosts, but Go faults explode as soon as `x264enc` is present:

- `rate-caps` page faults:
  - Go `34`
  - Python `51`
  - `gst-launch` `0`
- `encode` page faults:
  - Go `312438`
  - Python `62`
  - `gst-launch` `29`
- `parse` page faults:
  - Go `275291`
  - Python `163`
  - `gst-launch` `4`
- `mux-file` page faults:
  - Go `288738`
  - Python `209`
  - `gst-launch` `233`

This is the cleanest localization result in the ticket so far. It means the next code-change target should not be the raw capture path, conversion path, or shaped raw-video pacing path. The best current next target is the **encoder-input / Go-hosted memory-behavior boundary around `x264enc`**.

That does not yet prove the exact mechanism, but it does narrow the search sharply:

- more graph-shape surgery is now lower priority,
- more early raw-stage instrumentation is also lower priority,
- and cgo/build-flag or allocator/THP-style controls are now best interpreted specifically as tests of the `x264enc` boundary rather than the whole pipeline.

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/design-doc/02-small-graph-hosting-ladder-debugging-plan.md — Detailed plan for the stage-by-stage ladder and its interpretation rules
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/29-go-manual-stage-ladder-harness/main.go — Go manual small-graph control used for the ladder
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/30-python-manual-stage-ladder-harness/main.py — Python manual small-graph control used for the ladder
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/31-gst-launch-stage-ladder.sh — `gst-launch` stage ladder used as the cooler control family
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/32-small-graph-hosting-ladder-matrix.sh — Matrix runner that executed the first 6-stage x 3-host comparison
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/32-small-graph-hosting-ladder-matrix/20260415-033745/02-summary.md — Topline ladder result showing the first strong divergence at `x264enc`

Added a focused encode-stage encoder-contrast slice after the ladder localized the first strong divergence to the encoder boundary.

New/updated control files:

- `scripts/29-go-manual-stage-ladder-harness/main.go` — now accepts `--encoder`
- `scripts/29-go-manual-stage-ladder-harness/run.sh` — now passes encoder choice through to the Go harness
- `scripts/30-python-manual-stage-ladder-harness/main.py` — now accepts `--encoder`
- `scripts/30-python-manual-stage-ladder-harness/run.sh` — now passes encoder choice through to the Python harness
- `scripts/31-gst-launch-stage-ladder.sh` — now accepts `ENCODER=` for stage-level encoder swaps
- `scripts/33-encode-stage-encoder-contrast-matrix.sh` — new focused matrix runner for the encode stage only

The first full encoder-contrast attempt suggested that `vaapih264enc` is not a good default comparison point on this machine right now because the run did not complete cleanly and appeared to wedge the matrix. Rather than letting the hardware path block the software comparison, I reran the focused matrix with the two software encoders that completed normally:

- `x264enc`
- `openh264enc`

Saved result:

- `scripts/results/33-encode-stage-encoder-contrast-matrix/20260415-035541/02-summary.md`

This result is highly informative because it shows the Go anomaly is **not a generic encode-stage problem across all encoders**.

For `x264enc` the old pattern still holds:

- Go: `168.83%` avg CPU, `276940` page faults
- Python: `127.17%`, `61` page faults
- `gst-launch`: `129.67%`, `2` page faults

But for `openh264enc` the Go path no longer shows the same blow-up:

- Go: `58.00%` avg CPU, `53` page faults
- Python: `87.50%`, `110` page faults
- `gst-launch`: `88.11%`, `53` page faults

That is a major narrowing step. The remaining anomaly now looks much more specifically tied to the **Go-hosted `x264enc` path** (or something tightly coupled to it) rather than to the whole generic idea of “Go-hosted encoding.”

So the next useful target is no longer just “encoder boundary” in general. It is more precisely:

- Go-hosted interaction with `x264enc`
- `libx264`-specific memory/fault behavior
- or some `x264enc`-specific buffer-pool / allocation behavior that does not reproduce the same way with `openh264enc`

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/33-encode-stage-encoder-contrast-matrix.sh — Focused encode-stage contrast runner for software encoder swaps
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/33-encode-stage-encoder-contrast-matrix/20260415-035541/02-summary.md — Main software-encoder comparison showing the anomaly persists for `x264enc` but not `openh264enc`

Added a focused Go-only `x264enc` build-flag control to test the narrower “thin cgo wrapper compilation is unoptimized” suspicion.

New script:

- `scripts/34-go-x264-cgo-flag-matrix.sh`

Saved result:

- `scripts/results/34-go-x264-cgo-flag-matrix/20260415-040159/02-summary.md`

This first pass compared the same Go manual encode-stage `x264enc` harness under three build conditions:

- default `CGO_CFLAGS`
- `CGO_CFLAGS=-O2`
- `CGO_CFLAGS=-O3`

The result does **not** support a simple “the cgo wrapper was unoptimized, so that explains the anomaly” story. In this single run set:

- default: `254.33%` avg CPU, `284806` page faults
- `-O2`: `349.47%`, `262226` page faults
- `-O3`: `272.02%`, `283258` page faults

The exact CPU values are noisy enough that this should not be oversold as a precise ranking, but the important negative result is still clear: neither `-O2` nor `-O3` made the `x264enc` anomaly go away. If anything, the `-O2` run was worse in this pass.

So the current best interpretation remains:

- the issue is unlikely to be explained primarily by optimization level of the thin cgo glue itself,
- and the stronger explanations are still around the Go-hosted `x264enc` / `libx264` interaction, buffer allocation, memory behavior, or some other host-process effect that is not rescued by simple cgo C optimization flags.

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/34-go-x264-cgo-flag-matrix.sh — Focused Go-only build-flag control for the encode-stage `x264enc` path
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/34-go-x264-cgo-flag-matrix/20260415-040159/02-summary.md — First build-flag result showing no rescue from `-O2` or `-O3`

Added a focused `x264enc` property-ablation slice to test whether the Go anomaly is really just tied to expensive x264 algorithm settings such as `zerolatency`, `trellis`, or the `veryfast` preset.

New script:

- `scripts/35-x264-property-ablation-matrix.sh`

Saved result:

- `scripts/results/35-x264-property-ablation-matrix/20260415-042332/02-summary.md`

The matrix compared four `x264enc` variants at the encode stage across Go, Python, and `gst-launch`:

- `baseline` — `speed-preset=3`, `tune=4`, `bframes=0`, `trellis=true`
- `no_tune` — `speed-preset=3`, `tune=0`, `bframes=0`, `trellis=true`
- `no_trellis` — `speed-preset=3`, `tune=4`, `bframes=0`, `trellis=false`
- `ultrafast` — `speed-preset=1`, `tune=4`, `bframes=0`, `trellis=false`

This result is another strong **negative** control against the easy explanations.

Baseline still shows the familiar pattern:

- Go: `165.88%` avg CPU, `233962` page faults
- Python: `145.50%`, `63` page faults
- `gst-launch`: `144.11%`, `2` page faults

But the more important finding is what happens when the x264 settings are simplified.

For Python and `gst-launch`, the `ultrafast` variant behaves the way you would expect a cheaper encoder setting to behave:

- Python drops to `83.17%`
- `gst-launch` drops to `79.33%`

The Go path does **not** follow that pattern:

- Go `ultrafast` is still `332.00%` with `136321` page faults

Likewise, disabling `trellis` did **not** rescue the Go path even though the earlier perf symbol `x264_8_trellis_coefn` had made trellis look suspicious:

- Go `no_trellis`: `343.56%`, `179717` page faults
- Python `no_trellis`: `139.33%`, `62` page faults
- `gst-launch` `no_trellis`: `137.44%`, `0` page faults

And removing `zerolatency` did not help either:

- Go `no_tune`: `301.67%`, `187211` page faults
- Python `no_tune`: `146.50%`, `10607` page faults
- `gst-launch` `no_tune`: `128.11%`, `3534` page faults

So the new conclusion is sharper than before. The issue is not just “Go happens to pick an expensive x264 setting.” Even when the `x264enc` configuration is simplified, the Go-hosted path remains abnormally hot while the other host families behave much more reasonably.

That shifts the next suspicion away from individual `x264enc` quality knobs and more toward:

- Go-hosted interaction with the `x264enc` element lifecycle or buffer handling,
- `x264enc` / `libx264` buffer-pool or memory-behavior differences under the Go-hosted process,
- or some other host-process interaction that survives across x264 property changes.

### Additional Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/35-x264-property-ablation-matrix.sh — Focused x264-property comparison across Go, Python, and gst-launch hosts
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/35-x264-property-ablation-matrix/20260415-042332/02-summary.md — Main x264-property ablation result showing no rescue for the Go path from cheaper/faster x264 settings

