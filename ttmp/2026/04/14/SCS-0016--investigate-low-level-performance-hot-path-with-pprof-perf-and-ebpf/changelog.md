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
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/08-resolve-perf-go-addresses.sh — helper that extracts top Go-binary addresses from the perf report and resolves them with `go tool addr2line`
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/04-capture-perf-cpu-profile/20260414-224952/01-summary.md — first mixed-stack perf capture summary for the browser one-tab recording repro
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/04-capture-perf-cpu-profile/20260414-224952/go-addr2line.txt — first saved address-resolution artifact for the top `screencast-studio` frames in the perf report
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/reference/02-performance-investigation-approaches-and-tricks-report.md — project report on the investigation playbook used so far
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/reference/03-prometheus-metrics-architecture-and-field-guide.md — project report on the current metrics architecture
