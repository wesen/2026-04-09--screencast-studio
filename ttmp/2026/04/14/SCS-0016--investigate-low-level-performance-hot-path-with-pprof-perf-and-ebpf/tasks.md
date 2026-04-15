# Tasks

## Ticket setup and analysis scaffolding

- [x] Create ticket SCS-0016 for the lower-level profiling work
- [x] Write a detailed profiling plan document
- [x] Start a chronological investigation diary
- [ ] Keep the diary and changelog updated after each profiling slice

## Phase 0: profiling boundaries and reproducibility

- [ ] Write down the exact high-signal repro that all low-level tools should target
- [ ] Define what counts as a successful profile capture for pprof, perf, and eBPF
- [ ] Add ticket-local scripts for restarting the server in a profiling-friendly way
- [ ] Add ticket-local scripts for capturing and storing profiling artifacts under `scripts/`

## Phase 1: Go pprof first

- [ ] Add an optional local pprof enablement path to the serve runtime
- [ ] Ensure pprof is disabled by default and only enabled deliberately during local investigation
- [ ] Add a script to capture a CPU profile during the real browser one-tab desktop preview + recording hot phase
- [ ] Add a script to capture heap and goroutine profiles during the same scenario if useful
- [ ] Save raw pprof artifacts plus human-readable summaries in the ticket-local `scripts/` tree
- [ ] Decide whether pprof gives a sufficiently explanatory answer or mostly points at CGO / runtime boundaries

## Phase 2: perf if pprof is not enough

- [ ] Add a reproducible `perf record` capture script for the same high-signal scenario
- [ ] Save `perf.data`, `perf report` text output, and any stack-collapse / flamegraph artifacts under the ticket-local `scripts/` tree
- [ ] Verify symbol quality is good enough to separate Go, CGO, libc, GStreamer, and kernel stacks
- [ ] Summarize whether the dominant hot path is in Go, CGO, GStreamer, libc, syscalls, or scheduler behavior

## Phase 3: eBPF only if still needed

- [ ] Decide which eBPF questions remain unanswered after pprof/perf
- [ ] If needed, add targeted eBPF scripts for on-CPU, off-CPU, scheduler, syscall, or socket behavior
- [ ] Save raw outputs and concise summaries under the ticket-local `scripts/` tree
- [ ] Keep the scope narrow and question-driven rather than running generic eBPF tools blindly

## Reporting and conclusions

- [ ] Produce a lower-level findings report with concrete evidence and caveats
- [ ] Connect the lower-level findings back to the SCS-0015 browser-path hypothesis
- [ ] Recommend the next code-change target based on profiler evidence, not speculation
- [ ] Validate the ticket with `docmgr doctor --ticket SCS-0016 --stale-after 30`
