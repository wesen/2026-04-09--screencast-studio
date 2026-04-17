---
Title: Investigation diary
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
RelatedFiles:
    - Path: pkg/cli/serve.go
      Note: Candidate place to add optional local profiling enablement
    - Path: pkg/media/gst/shared_video.go
      Note: Likely next low-level boundary around appsink callbacks and Go buffer handoff
    - Path: ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/reference/02-browser-preview-streaming-lab-report.md
      Note: |-
        Parent evidence trail that justified creating this lower-level profiling ticket
        Parent evidence trail that justified this ticket
    - Path: ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/design-doc/01-low-level-profiling-plan.md
      Note: |-
        Primary plan document for this ticket
        Primary plan document for Step 1
ExternalSources: []
Summary: Chronological diary for the lower-level profiling investigation into the browser-connected hot path.
LastUpdated: 2026-04-15T05:05:00-04:00
WhatFor: Record the implementation and profiling workflow step by step, including failures, commands, and interpretation changes.
WhenToUse: Read when continuing SCS-0016 and you need the exact reasoning and commands used so far.
---


# Diary

## Goal

Capture the lower-level profiling investigation that begins where SCS-0015 left off: app-level metrics have already ruled out several plausible browser-path suspects, so this ticket focuses on pprof, perf, and eBPF evidence for the remaining hot path.

## Step 1: Created a Separate Ticket for Lower-Level Profiling and Wrote the Initial Plan

SCS-0015 had reached a useful boundary. The real browser-connected hot path was still real, but the latest app-level timing counters strongly suggested that the final MJPEG write loop, PreviewManager cached-frame operations, and EventHub publish costs were all too small to explain the `~380–410%` hot phase. At that point, continuing to pile lower-level profiler work into SCS-0015 would have made that ticket much harder to read and continue.

So I created a new ticket specifically for the lower-level work and used the first step to write down a disciplined plan. The main decision is to proceed in stages: Go pprof first, then perf only if pprof mainly points at CGO or otherwise fails to explain the hot phase, and eBPF only for narrowly targeted unanswered questions after that.

### Prompt Context

**User prompt (verbatim):** "alright , create a new ticket to do this lower level performance measurement work, so we don't over burden the current one. Add a detailed plan, and tasks, and then get to work. Commit at appropriate intervals, keep a frequent and detailed diary"

**Assistant interpretation:** Split the lower-level profiling work into a dedicated ticket, create proper planning and task documents, and then begin implementation in small committed slices while maintaining a detailed diary.

**Inferred user intent:** Keep SCS-0015 focused on browser-path measurement and reasoning, while giving the new lower-level profiling track enough structure to be continued and reviewed independently.

### What I did

- Created a new ticket:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
docmgr ticket create-ticket --ticket SCS-0016 --title "Investigate low-level performance hot path with pprof, perf, and eBPF" --topics screencast-studio,gstreamer,performance,profiling,backend,analysis
```

- Added two initial docs:

```bash
docmgr doc add --ticket SCS-0016 --doc-type design-doc --title "Low-level profiling plan"
docmgr doc add --ticket SCS-0016 --doc-type reference --title "Investigation diary"
```

- Wrote and updated:
  - `index.md`
  - `tasks.md`
  - `changelog.md`
  - `design-doc/01-low-level-profiling-plan.md`
  - `reference/01-investigation-diary.md`

### Why

The main reason for splitting the work is conceptual hygiene. SCS-0015 already has a lot of value as a browser-path measurement ticket. Lower-level profiling introduces different questions, tools, artifacts, and interpretation rules. Keeping the work separate makes it easier to review and easier for a future investigator to pick up the thread without mixing browser-matrix bookkeeping with pprof/perf/eBPF details.

### What worked

- The new ticket workspace was created cleanly.
- `docmgr doc add` worked for the design doc and diary.
- The first plan document was straightforward to write because the preceding SCS-0015 evidence had already narrowed the problem significantly.

### What didn't work

Nothing failed materially in this step.

### What I learned

The project is now at a good profiler boundary. Earlier, lower-level tools would have been premature because the app-level suspects were still too broad. Now the remaining unexplained cost is narrow enough that pprof/perf/eBPF should be substantially more informative.

### What was tricky to build

The tricky part was deciding how much of the old ticket’s findings to repeat. I wanted the new ticket to be continuation-friendly without duplicating all of SCS-0015. The compromise was to summarize only the key narrowing facts in the new plan and point back to the SCS-0015 lab report for the full evidence trail.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/design-doc/01-low-level-profiling-plan.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/tasks.md`

The main review question is whether the staged tool order is right: pprof first, perf second only if needed, eBPF third only for targeted unanswered questions.

### What should be done in the future

- Commit the new ticket docs as the first focused SCS-0016 docs commit.
- Start the first implementation slice for optional local pprof enablement and reproducible capture scripts.

### Code review instructions

Start with:

- `ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/index.md`
- `ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/tasks.md`
- `ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/design-doc/01-low-level-profiling-plan.md`

Then compare the stated rationale against the parent evidence in:

- `ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/reference/02-browser-preview-streaming-lab-report.md`

### Technical details

The current leading repro for all low-level tools remains:

```text
desktop preview + recording + one real browser tab
```

That repro is still the highest-value target because SCS-0015 showed it was already enough to demonstrate the unexplained hot phase.

## Step 2: Added Optional pprof Support and Captured the First CPU Profile

The first real implementation step for SCS-0016 was to add the smallest lower-level profiler that could answer an important question: does the hot phase still show up clearly in Go userland, or does it immediately collapse into CGO/runtime/external code? The answer to that question determines whether it is worth staying in pprof longer or whether the investigation should move to `perf` as quickly as possible.

I added optional `--pprof-addr` support to `screencast-studio serve`, exposed a separate pprof debug server rather than overloading the main mux, and wrote ticket-local helper scripts to restart the app with pprof enabled and capture a CPU profile. I then used the same high-signal browser repro — desktop preview + recording + one real browser tab — to capture the first profile. That first result was already highly informative: it is dominated by native/external code rather than clean Go-userland hotspots.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** After creating the lower-level profiling ticket, begin implementing the first profiling slice immediately and keep the diary current as code and evidence land.

**Inferred user intent:** Make the new ticket operational, not just well documented, and use it to get real evidence as soon as possible.

**Commit (code):** `d06957a0318d1294609b35e8673ca307fad4bbab` — `Add optional pprof profiling support`

### What I did

- Added:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/pprof.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/pprof_test.go`
- Updated:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/serve.go`
- Added ticket-local helper scripts:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/01-restart-scs-web-ui-with-pprof.sh`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/02-capture-pprof-cpu-profile.sh`
- Validated with:

```bash
gofmt -w internal/web/pprof.go internal/web/pprof_test.go internal/web/server.go pkg/cli/serve.go
bash -n ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/01-restart-scs-web-ui-with-pprof.sh
bash -n ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/02-capture-pprof-cpu-profile.sh
go test ./internal/web ./pkg/cli -count=1
go test ./... -count=1
```

- Restarted the app with pprof enabled and confirmed:
  - `http://127.0.0.1:7777/api/healthz`
  - `http://127.0.0.1:6060/debug/pprof/`
- Captured the first CPU profile under:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/02-capture-pprof-cpu-profile/20260414-204800/`

### Why

This was the smallest lower-level profiling slice that could answer a strategic question with very little ambiguity. If pprof had immediately shown clear Go-userland hotspots, the investigation could have stayed at that layer. If it mostly showed opaque external/native time, that would justify moving quickly to `perf`.

### What worked

- The separate pprof server started cleanly when `--pprof-addr` was set and stayed absent when not configured.
- The first CPU profile capture succeeded during the real browser-connected recording repro.
- The profile was informative enough to guide the next tool choice.

### What didn't work

Two concrete issues showed up in this slice.

First, the first version of the pprof capture script used Bash’s special `SECONDS` variable as the capture parameter name. That caused the generated summary to report `seconds: 41` even though I launched it intending a shorter window. The raw profile is still useful, but the script needs cleanup so the capture duration is controlled explicitly.

Second, the next-likely native tools are currently blocked for the current user on this machine. I confirmed:

```text
perf version 6.8.12
perf_event_paranoid = 4
```

and `perf stat -p <pid> sleep 1` failed with the exact permission message about restricted performance monitoring. I also confirmed that `bpftrace` is installed but currently says it only supports running as root.

### What I learned

The first pprof result did exactly what it needed to do. It strongly suggests the hot phase is not mainly in normal Go userland. The saved top output is dominated by:

- `<unknown>`
- `libgstvideo-1.0.so.0.2402.0`
- `libc.so.6`
- `libjpeg.so.8.2.2`
- and `runtime._ExternalCode`

That is enough to say pprof is probably not the end of the story and that `perf` is the right next profiler if permissions can be obtained.

### What was tricky to build

The main tricky part was keeping the pprof path opt-in and separate. I did not want to quietly expose `/debug/pprof` on the main app mux or make profiling a default part of normal local usage. The separate debug address keeps the profiling flow explicit and makes the restart/capture scripts easier to reason about.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/serve.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/pprof.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/02-capture-pprof-cpu-profile.sh`

The key review questions are:

- whether the separate pprof server boundary is the right default shape,
- and whether the next move should indeed be `perf` rather than additional Go-level profiling.

### What should be done in the future

- Fix the capture-script duration variable so it does not use Bash’s special `SECONDS` variable.
- Record the current profiler permission blockers in a reproducible prereq script.
- Move to `perf` once permissions are available, because pprof already suggests the dominant cost is external/native.

### Code review instructions

Start with the runtime integration files:

- `internal/web/server.go`
- `pkg/cli/serve.go`
- `internal/web/pprof.go`

Then review the first captured artifacts:

- `scripts/results/02-capture-pprof-cpu-profile/20260414-204800/pprof-top.txt`
- `scripts/results/02-capture-pprof-cpu-profile/20260414-204800/pprof-top-cum.txt`

### Technical details

The first profile summary reported:

```text
<unknown>                        67.74%
[libgstvideo-1.0.so.0.2402.0]    15.77%
[orcexec.Mmdcyh]                  7.66%
[libc.so.6]                       5.21%
[libjpeg.so.8.2.2]                1.31%
runtime._ExternalCode            98.65% cumulative
```

That is the strongest current reason to escalate from pprof to mixed-stack native profiling.

## Step 3: Wrote Two Project Reports While Waiting on Lower-Level Tool Permissions

While waiting on `perf` permissions, I used the pause productively by writing two continuation-oriented reports. The first captures the investigation playbook itself: which approaches and tricks actually helped narrow the performance mystery so far. The second is specifically about the project’s Prometheus-style metrics architecture: what metric families exist, how they are implemented, and how the investigation has used them.

This mattered because a lot of the recent work is no longer just “what result did we get?” It is also “what methods have now proven effective enough to reuse?” and “how exactly should a future investigator interpret the current `/metrics` surface?” Those are the kinds of answers that are easy to lose if they are only scattered across tickets and commit diffs.

### Prompt Context

**User prompt (verbatim):** "ok while we are waiting, let's write a project report that details all the tricks and approaches we used to figure out the performance issues so far, and a separate project report specifically on the metrics for prometheus and how they work and what we collect and for what"

**Assistant interpretation:** Use the waiting period to write two substantial project reports: one about the overall investigation methodology and one about the metrics architecture and field usage.

**Inferred user intent:** Preserve the investigation knowledge itself, not just the raw results, so the project is easier to continue and easier to hand off.

### What I did

- Added two new reports under SCS-0016:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/reference/02-performance-investigation-approaches-and-tricks-report.md`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/reference/03-prometheus-metrics-architecture-and-field-guide.md`
- Updated the ticket index so these documents are treated as primary SCS-0016 references.

### Why

The investigation now spans multiple tickets and several layers of evidence. A new person joining at this point would benefit from two kinds of synthesis:

1. a practical playbook of the investigation techniques that actually worked,
2. a field guide to the metrics surface that the project built along the way.

Writing those explicitly now is cheaper than reconstructing them later from many tickets and run directories.

### What worked

- The reports could be grounded in real ticket-local artifacts rather than abstract advice.
- The metrics guide benefited from the fact that the project’s metrics system is still small enough to document completely.

### What didn't work

Nothing failed materially in this documentation slice.

### What I learned

The project now has a reusable performance-investigation playbook. The most useful pieces are not glamorous:

- isolate scenarios carefully,
- separate fresh-server and real-browser runs,
- compare deltas rather than only averages,
- treat negative findings as progress,
- and only escalate to lower-level profilers after app-level narrowing has done its job.

### What was tricky to build

The tricky part was balancing “report” and “reference.” I wanted both documents to be readable as narrative reports, but also useful as field guides. That is why both documents keep a strong quick-reference structure while still explaining the rationale behind the choices.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/reference/02-performance-investigation-approaches-and-tricks-report.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/reference/03-prometheus-metrics-architecture-and-field-guide.md`

The main review question is whether these reports are continuation-friendly enough for a new investigator who did not live through SCS-0014 and SCS-0015.

### What should be done in the future

- Keep both reports updated as lower-level profiling adds new techniques or new metric families.
- Once `perf` runs are possible, extend the “approaches and tricks” report with what actually worked at the mixed-stack profiler level.

### Code review instructions

Read the two new reports first, then compare them against the underlying implementation and ledger docs they summarize:

- `reference/02-performance-investigation-approaches-and-tricks-report.md`
- `reference/03-prometheus-metrics-architecture-and-field-guide.md`
- `ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/reference/02-browser-preview-streaming-lab-report.md`
- `pkg/metrics/metrics.go`
- `internal/web/preview_metrics.go`
- `internal/web/event_metrics.go`

### Technical details

The metrics report currently documents the main metric families in:

- `pkg/metrics/metrics.go`
- `internal/web/preview_metrics.go`
- `internal/web/event_metrics.go`
- `pkg/media/gst/recording.go`
- `pkg/media/gst/shared_video_recording_bridge.go`

and explains how the ticket-local samplers turned cumulative counters into per-run deltas by comparing the first and last saved `.prom` snapshots.

## Step 4: Rechecked Profiler Prerequisites, Confirmed perf Access, and Fixed the pprof Script Duration Bug

After the first pprof pass, the next blocker question was practical rather than analytical: can I actually run mixed-stack tooling as the current user on this machine, or do I need to stop and ask for privilege changes first? Earlier quick checks had shown `perf_event_paranoid=4` and a failing `perf stat`, so I wrote a ticket-local prereq script to capture the real machine state in a reproducible way before spending more time designing the `perf` slice.

I also used this step to fix the small but important pprof helper bug where the script had mistakenly used Bash’s special `SECONDS` variable as the capture-duration parameter. That bug did not invalidate the first capture, but it did make the saved summary misleading. Renaming the input variable to `PROFILE_SECONDS` makes future pprof summaries much easier to trust.

### Prompt Context

**Relevant prior context:** Earlier investigation notes had captured a failing `perf` permission state, but the broader project context also said the user may temporarily adjust perf permissions or run `perf` manually if needed.

**Assistant interpretation:** Save a reproducible check of the current local profiling-tool state instead of relying on memory from earlier ad hoc shell checks, and bring the ticket docs in line with the result.

**Inferred user intent:** Avoid hand-wavy profiler assumptions; capture the actual machine state and keep the continuation docs honest.

### What I did

- Fixed the pprof helper script:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/02-capture-pprof-cpu-profile.sh`
- Added the profiler prereq script:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/03-check-profiler-prereqs.sh`
- Ran it and saved the result under:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/03-check-profiler-prereqs/20260414-211300/`

### Commands

```bash
bash ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/03-check-profiler-prereqs.sh
```

### Why

This step turns profiler availability from a fuzzy remembered fact into a saved artifact. That matters because machine state can change between sessions, and the next ticket action — writing and running a `perf record` harness — depends directly on whether `perf` is genuinely usable now.

### What worked

- The prereq script captured the current local state cleanly.
- The result showed that `perf` is now available to the current user in a useful way.
- The script also preserved the still-important limitation that `bpftrace` remains root-only.
- The pprof helper bug fix is small but important for future evidence quality.

### What didn't work

Nothing failed materially in this step.

### What I learned

The current machine state is better than the initial quick check suggested. The saved result shows:

- `perf_event_paranoid=1`
- `perf stat -p <listening-pid> sleep 1` succeeds
- `bpftrace` still fails with the message that it only supports running as root

That means the next likely step is no longer “wait for perf permissions.” It is “write the actual `perf record` capture script and use it on the established high-signal browser repro.”

### What was tricky to build

The tricky part here was mostly epistemic: earlier manual checks were already in the diary, and it would have been easy to treat them as still true. Saving a fresh prereq artifact forced the ticket back into an evidence-first state and avoided carrying an obsolete blocker forward.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/03-check-profiler-prereqs.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/03-check-profiler-prereqs/20260414-211300/perf-check.txt`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/02-capture-pprof-cpu-profile.sh`

The main review question is whether the prereq script captures enough environment truth for continuation work, or whether it should also snapshot more kernel/perf settings before the first `perf record` run.

### What should be done in the future

- Add a ticket-local `perf record` capture script for the same high-signal repro.
- Save raw `perf.data` and a text `perf report` in the ticket tree.
- Compare the mixed-stack view against the current pprof conclusion that most of the hot phase lives in external/native code.

### Code review instructions

Review the prereq artifact first:

- `scripts/results/03-check-profiler-prereqs/20260414-211300/01-summary.md`
- `scripts/results/03-check-profiler-prereqs/20260414-211300/perf-check.txt`
- `scripts/results/03-check-profiler-prereqs/20260414-211300/bpftrace-check.txt`

Then confirm that the pprof helper now uses `PROFILE_SECONDS` rather than Bash’s special `SECONDS` variable:

- `scripts/02-capture-pprof-cpu-profile.sh`

### Technical details

The saved prereq result shows:

```text
listen_pid=716961
perf_path=/usr/bin/perf
bpftrace_path=/usr/bin/bpftrace
bpftool_path=/usr/sbin/bpftool
perf_event_paranoid=1
```

and `perf-check.txt` now includes a successful `perf stat` run instead of the earlier permission failure.

## Step 5: Backfilled the Exact Repro Scripts, Captured the First Mixed-Stack perf Profile, and Resolved the Top Go Addresses

Once `perf` availability was confirmed, the next practical job was to make the actual repro flow retraceable from inside the ticket. I had already been using a combination of browser helpers and ad hoc shell commands to reproduce the one-tab desktop preview + recording hot phase, but the user explicitly asked that all scripts I was running be stored in the ticket with increasing numeric prefixes so the analysis can be replayed later. I therefore backfilled the exact browser-driving helpers I used into SCS-0016 before treating the first `perf` result as a stable slice.

The backfilled scripts for this step are:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/04-capture-perf-cpu-profile.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/05-open-studio-and-wait-desktop.js`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/06-start-recording.js`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/07-stop-recording.js`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/08-resolve-perf-go-addresses.sh`

I then used that flow against the already-established high-signal repro:

```text
desktop preview + recording + one real browser tab
```

and saved the first mixed-stack capture under:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/04-capture-perf-cpu-profile/20260414-224952/`

### Prompt Context

**User prompt (verbatim):** "store all the scripts you are running with a numerical increasing prefix in the scripts/ folder of the ticket so that we can retrace your analysis. backfill."

**Assistant interpretation:** Stop relying on scattered or cross-ticket helper locations, copy the exact scripts used for this profiling slice into SCS-0016 with ordered prefixes, and keep the capture/review flow reproducible from the ticket alone.

**Inferred user intent:** The ticket itself should be a faithful replay surface for the analysis rather than a prose summary that still depends on hidden ad hoc commands.

### What I did

- Backfilled the exact browser-driving helpers into SCS-0016 with ordered numeric prefixes.
- Added the ticket-local `perf` capture script and the ticket-local Go-address resolver helper.
- Ran the browser repro in the live Studio page.
- Captured the first `perf` profile while recording was active.
- Resolved the top `screencast-studio` addresses from the `perf` report into a saved `go-addr2line.txt` artifact.

### Commands

Browser helpers used through the browser tool:

- `scripts/05-open-studio-and-wait-desktop.js`
- `scripts/06-start-recording.js`
- `scripts/07-stop-recording.js`

Shell commands used:

```bash
bash ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/04-capture-perf-cpu-profile.sh
bash ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/08-resolve-perf-go-addresses.sh \
  ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/04-capture-perf-cpu-profile/20260414-224952
```

### Why

This step was necessary for two reasons.

First, the ticket needed honest replayability. The user specifically asked for the script surface to be backfilled into the ticket, and that request is correct: without that, the investigation would still depend on memory about which helpers lived in SCS-0015 and which commands were only one-off shell invocations.

Second, the project needed the first real mixed-stack answer, not just another profiler setup step. pprof had already told us that the hot path was mostly external/native. `perf` was the first tool that could reasonably distinguish encoder work, GStreamer pad-push work, Go runtime callback work, and kernel/syscall traces in one view.

### What worked

- The browser repro could be re-established cleanly with the ticket-local scripts.
- `perf record` succeeded against the live server PID with no lost samples.
- The first report was immediately more specific than pprof.
- The Go-address resolver helper worked and saved a reproducible address-to-symbol artifact instead of leaving the resolution as a one-off shell trick.

### What didn't work

One limitation remains: native libraries are symbolized much better than the main Go binary. The `perf` report can clearly show `libx264`, `libgstreamer`, `gst_pad_push`, `x264_encoder_encode`, libc, and kernel frames, but many top `screencast-studio` frames still show up as raw addresses in the `perf` report itself.

I mitigated that limitation with `08-resolve-perf-go-addresses.sh`, but the result is still weaker than ideal because the live process came from a temporary `go run` executable under `/tmp/go-build.../exe/screencast-studio` rather than from a deliberately built and preserved binary path.

### What I learned

The first `perf` result moves the explanation forward in an important way.

The strongest visible hot path in this capture is native encoder work:

- `libx264.so.164 x264_8_trellis_coefn` accounts for about `45.63%` children / `44.49%` self.
- Another `~27.84%` sits in unresolved `[unknown]` native frames that are very likely still part of the same native-heavy region.
- GStreamer pad-push frames are visible in the call chain into `x264_encoder_encode`, including `gst_pad_push` and nearby `libgstreamer` / `libgstbase` frames.

That means the current best explanation is no longer just “something native below Go.” It is more specific: during this browser-connected recording repro, a large fraction of the heat is in **native x264 encoder work plus the surrounding GStreamer push path**.

The resolved Go-side addresses add a second, smaller insight. The top visible `screencast-studio` addresses map to things like:

- `runtime.cgocallbackg1`
- `runtime.cgoCheckArg`
- `github.com/rs/zerolog.(*Event).Dict`
- `github.com/rs/zerolog.ConsoleWriter.Write`
- `encoding/json.(*Decoder).readValue`

I do not yet want to over-interpret those names, because the address resolution is still somewhat brittle in this `go run` setup. But it does suggest that the visible Go-side slice is more about **CGO callback/runtime boundary plus some logging/JSON work** than about a clean, dominant application-level business function.

### What was tricky to build

The trickiest part here was not the `perf` command itself. It was keeping the whole repro chain honest and replayable:

- live app on `:7777`
- one real browser tab
- preview active
- recording active
- `perf` attached to the right listening PID
- top Go addresses resolved against the right executable path

If any one of those pieces stays implicit, the next person has to reverse-engineer the analysis procedure from scratch.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/04-capture-perf-cpu-profile.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/08-resolve-perf-go-addresses.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/04-capture-perf-cpu-profile/20260414-224952/perf-report.txt`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/04-capture-perf-cpu-profile/20260414-224952/go-addr2line.txt`

The main review questions are:

- whether the current symbol quality is already good enough to start targeting code changes,
- or whether the next best move is to rerun the same capture from a stable `go build` binary to improve Go symbolization before drawing tighter conclusions.

### What should be done in the future

- Keep the raw `perf.data` artifact locally saved, but avoid treating it as required source-control material.
- Consider rerunning the same scenario from a stable built binary path instead of `go run` to improve Go-symbol resolution.
- Compare the first mixed-stack result against the earlier pprof output and decide whether the next code-change target should be encoder settings, pipeline structure, or a smaller Go-side callback/logging reduction.

### Code review instructions

Review the capture and summary in this order:

- `scripts/results/04-capture-perf-cpu-profile/20260414-224952/01-summary.md`
- `scripts/results/04-capture-perf-cpu-profile/20260414-224952/perf-report.txt`
- `scripts/results/04-capture-perf-cpu-profile/20260414-224952/perf-report-dso-symbol.txt`
- `scripts/results/04-capture-perf-cpu-profile/20260414-224952/go-addr2line.txt`

Then compare that mixed-stack result against the earlier pprof evidence in:

- `scripts/results/02-capture-pprof-cpu-profile/20260414-204800/pprof-top.txt`

### Technical details

The saved `perf` run reported:

```text
[ perf record: Woken up 206800 times to write data ]
[ perf record: Captured and wrote 101.983 MB ... (6429 samples) ]
```

The raw `perf.data` file is intentionally being kept as a local saved artifact and does not need to be committed.

## Step 6: Reran perf from a Stable Built Binary to Improve Go Symbolization

The first `perf` slice was already useful, but it still had one annoying limitation: the process was running from a temporary `go run` executable under `/tmp/go-build...`, so the main binary showed too many address-only `screencast-studio` frames. That was enough to trust the native `libx264` / GStreamer story, but not enough to be fully satisfied about the Go-side slice.

So the next step was not to broaden scenarios. It was to hold the same repro constant and improve the capture quality. I added a ticket-local restart helper that builds a stable `screencast-studio` binary under the SCS-0016 `scripts/bin/` directory, restarts the app from that binary path, and keeps pprof enabled on the same separate debug address. Then I reran the exact same browser one-tab recording repro and captured a second `perf` sample against the stable-binary server PID.

### Prompt Context

**Relevant prior context:** The first `perf` run showed strong native hot-path evidence, but the main-binary frames were still weaker than ideal because the server was running from `go run` rather than a preserved built binary.

**Assistant interpretation:** Improve profiler evidence quality without changing the repro itself: build a stable binary, restart from that path, rerun the same perf capture, and compare symbolization quality.

**Inferred user intent:** Keep the investigation incremental and evidence-driven; if the current profiler result is almost good enough, tighten the symbol story before deciding the next code-change target.

### What I did

- Added:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/09-restart-scs-web-ui-with-built-binary-and-pprof.sh`
- Restarted the app from:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/bin/screencast-studio`
- Reused the same browser helpers:
  - `scripts/05-open-studio-and-wait-desktop.js`
  - `scripts/06-start-recording.js`
  - `scripts/07-stop-recording.js`
- Captured a second perf run under:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/04-capture-perf-cpu-profile/20260414-230415/`
- Fixed two small script bugs uncovered along the way:
  1. `scripts/04-capture-perf-cpu-profile.sh` summary generation had markdown backticks that accidentally triggered shell command substitution inside an unquoted heredoc.
  2. `scripts/08-resolve-perf-go-addresses.sh` assumed address-only Go frames would always exist; it now emits a clear note when the perf report already contains direct symbols.

### Commands

```bash
bash ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/09-restart-scs-web-ui-with-built-binary-and-pprof.sh
bash ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/04-capture-perf-cpu-profile.sh
bash ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/08-resolve-perf-go-addresses.sh \
  ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/04-capture-perf-cpu-profile/20260414-230415
```

### Why

This step is a good example of “measure first, then improve the measurement surface.” The repro itself was already strong enough. What was weak was the symbol story for the main Go binary. Restarting from a stable built binary let the investigation answer a more refined question:

> once Go symbolization is no longer the limiting factor, does the main conclusion change?

### What worked

- The built-binary restart succeeded cleanly and exposed the expected health and pprof endpoints.
- The second `perf` capture succeeded with no lost samples.
- Main-binary symbolization improved enough that the report now contains named `screencast-studio` runtime and websocket/server functions directly.
- The main conclusion stayed stable instead of collapsing.

### What didn't work

Two small helper bugs surfaced during this step, but both were fixed immediately:

- the heredoc/backtick bug in `04-capture-perf-cpu-profile.sh`
- the empty-address newline bug in `08-resolve-perf-go-addresses.sh`

Those were not investigation blockers, but they were exactly the kind of script quality issues worth documenting because the user explicitly wants the analysis flow to be replayable from the ticket.

### What I learned

The stable-binary rerun improved confidence in the interpretation substantially.

The dominant hot path is still not ordinary Go web code. The second run shows:

- `libx264.so.164 x264_8_trellis_coefn` ≈ `41.74%` children / `40.93%` self
- `[unknown]` native frames ≈ `25.75%`
- `libgstreamer-1.0.so.0.2402.0 gst_pad_push` ≈ `13.68%`
- visible `gst_buffer_copy_into` and libc `__memcpy_evex_unaligned_erms` in the call path into `x264_encoder_encode`

The newly visible `screencast-studio` frames are informative mostly because they are **small**. They include runtime scheduler/GC/CGO callback functions and websocket write paths such as:

- `runtime.cgocallbackg1`
- `runtime.asmcgocall.abi0`
- `internal/web.(*Server).handleWebsocket.func2`
- `internal/web.writeWebsocketServerEvent`
- `github.com/gorilla/websocket.(*Conn).WriteMessage`

Those names are useful because they confirm the Go-side browser/event path is present in the sample. But the percentages are tiny compared with the native encoder + pipeline region, so they do not overturn the earlier conclusion.

A second interesting side observation is that the perf writer wakeup behavior changed a lot between the two runs:

- first run: `206800` wakeups for about `101.983 MB` / `6429` samples
- built-binary rerun: `669` wakeups for about `172.843 MB` / `10893` samples

That does not by itself explain product CPU, but it does suggest the stable-binary capture path is a better-quality measurement surface and probably the right baseline for future perf reruns.

### What was tricky to build

The trickiest part here was preserving the exact same runtime behavior while only changing the executable path. If I had changed the repro, the new symbolization could have been confused with a workload shift. Keeping the same one-tab desktop preview + recording scenario made the comparison much more trustworthy.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/09-restart-scs-web-ui-with-built-binary-and-pprof.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/04-capture-perf-cpu-profile/20260414-230415/perf-report-dso-symbol.txt`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/04-capture-perf-cpu-profile/20260414-230415/go-addr2line.txt`

The main review question is whether the symbol quality is now good enough to move from profiling into a code-change hypothesis, or whether one more perf pass with slightly different call-graph settings is still worthwhile.

### What should be done in the future

- Keep using the stable built binary rather than `go run` for future mixed-stack perf captures.
- Write a short lower-level findings report once the current perf interpretation is stable enough.
- Decide whether the next code-change target should focus on native encoder cost, shared preview branch copy behavior, or a combination.

### Code review instructions

Compare these two directories directly:

- `scripts/results/04-capture-perf-cpu-profile/20260414-224952/`
- `scripts/results/04-capture-perf-cpu-profile/20260414-230415/`

The main thing to look for is not whether the exact percentages match perfectly, but whether the better-symbolized rerun changes the diagnosis. Right now it does not; it mostly strengthens it.

### Technical details

The second run used:

```text
exe_path=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/bin/screencast-studio
perf_event_paranoid=1
```

and `perf-record.stderr.log` reports:

```text
[ perf record: Woken up 669 times to write data ]
[ perf record: Captured and wrote 172.843 MB ... (10893 samples) ]
```

## Step 7: Narrowed the Remaining Direct-Recording Gap to a Native Hosting / Memory-Fault Problem and Wrote the Main Findings Report

After the stable-binary perf rerun, the next job was to stop accumulating isolated facts and turn them into a tighter explanation. I built a sequence of direct controls around the same full-desktop workload: a matched `gst-launch` control, a copied/app-like Go direct harness with dot dumps, a fully manual Go direct harness with dot dumps, a small A/B matrix, a side-by-side mixed-stack perf compare, and finally a threadgroup-oriented `perf stat` compare. That work made the current boundary much sharper.

The most important result is that the remaining Go-hosted vs `gst-launch` gap is no longer best explained by graph shape or by obvious per-frame Go userland work. The direct-path `Y444` negotiation bug was real and worth fixing, but once the direct caps were pinned to `I420`, the manual and copied Go-hosted pipelines converged to the same realized graph. Mixed-stack perf then showed both cases dominated by native `libx264` and GStreamer pad-push work, while visible Go/cgo frames were tiny. The strongest remaining differentiator is now the huge page-fault delta in the Go-hosted case.

### Prompt Context

**User prompt (verbatim):** "yes,write a really detailed investigation report and a set of research queries to be given to our internet to look them up online. Store in the ticket and then upload to remarkable."

**Assistant interpretation:** Consolidate the recent SCS-0016 direct-recording evidence into a continuation-friendly report, write a separate copy/paste-ready online research query packet, keep the ticket bookkeeping current, validate the ticket, and prepare the documents for reMarkable delivery.

**Inferred user intent:** Turn the current lower-level profiling findings into durable project documentation that can be handed to both internal engineers and an internet-enabled researcher without losing the evidence trail.

### What I did

- Added the matched `gst-launch` direct control script:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/16-gst-launch-direct-full-desktop-match-app.sh`
- Added the fully manual Go direct harness:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/17-go-manual-direct-full-desktop-harness/main.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/17-go-manual-direct-full-desktop-harness/run.sh`
- Added comparison helpers:
  - `scripts/18-direct-harness-ab-matrix.sh`
  - `scripts/19-perf-compare-go-manual-vs-gst-launch.sh`
  - `scripts/20-perf-stat-and-thread-compare-go-vs-gst-launch.sh`
  - `scripts/21-perf-stat-threadgroup-compare-go-vs-gst-launch.sh`
- Fixed the direct-path raw caps in:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/recording.go`
- Wrote two new ticket docs:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/reference/04-direct-recording-hosting-gap-investigation-report.md`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/reference/05-online-research-query-packet-for-go-hosted-gstreamer-performance.md`
- Updated ticket bookkeeping:
  - related the new docs to the exact code/result files with `docmgr doc relate`
  - checked tasks `5,6,19,20,24,25` because the new report and query pack now cover the high-signal repro, capture criteria, dominant-hot-path summary, eBPF-unanswered-question decision, and the main findings report itself
  - updated `index.md` so the new docs and current interpretation appear on the ticket landing page

### Commands

```bash
bash ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/16-gst-launch-direct-full-desktop-match-app.sh
bash ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/17-go-manual-direct-full-desktop-harness/run.sh
DURATION_SECONDS=8 REPEATS=3 bash ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/18-direct-harness-ab-matrix.sh
DURATION_SECONDS=8 bash ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/19-perf-compare-go-manual-vs-gst-launch.sh
DURATION_SECONDS=8 bash ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/20-perf-stat-and-thread-compare-go-vs-gst-launch.sh
DURATION_SECONDS=8 bash ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/21-perf-stat-threadgroup-compare-go-vs-gst-launch.sh

docmgr doc add --ticket SCS-0016 --doc-type reference --title "Direct recording hosting-gap investigation report"
docmgr doc add --ticket SCS-0016 --doc-type reference --title "Online research query packet for Go-hosted GStreamer performance"
docmgr doc relate --doc /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/reference/04-direct-recording-hosting-gap-investigation-report.md --file-note "/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/recording.go:Direct recording builder and I420 fix"
docmgr doc relate --doc /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/reference/05-online-research-query-packet-for-go-hosted-gstreamer-performance.md --file-note "/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/21-perf-stat-threadgroup-compare-go-vs-gst-launch/20260415-012510/01-summary.md:Page-fault and thread-count comparison used to scope the queries"
docmgr task check --ticket SCS-0016 --id 5,6,19,20,24,25
```

### Why

The investigation had accumulated enough strong artifacts that continuing to answer questions only in chat would have been wasteful. The user asked for two continuation-oriented deliverables:

1. one detailed internal report that explains the current hosting-gap boundary honestly,
2. one external-facing query pack that tells an internet-enabled researcher exactly what to search for online.

That is the right split because the report is evidence-heavy and local-repo-specific, while the query packet is optimized for web research and triage.

### What worked

- The matched `gst-launch` control was much cooler than the Go-hosted direct controls, which forced the investigation away from the earlier “full-desktop direct x264 is just inherently 500–600%” hypothesis.
- The dot dumps showed the pre-fix app-like direct path negotiating `Y444`, which was a real graph bug worth fixing.
- The post-fix dot dumps showed the copied/app-like and manual Go-hosted graphs converging on `I420`, which ruled out the “DSL built a different graph” theory.
- The A/B matrix still showed both Go-hosted variants much hotter than the matched `gst-launch` control, which sharpened the problem statement further.
- The mixed-stack perf compare showed both cases dominated by native `libx264` + `gst_pad_push` work, while visible Go/cgo frames remained tiny.
- The improved threadgroup `perf stat` compare showed similar thread counts but a massive page-fault delta, which is now the strongest current clue.
- The two new docs were written and related into the ticket successfully.

### What didn't work

Two recent workflow issues were worth recording exactly.

First, the first attempt at the side-by-side perf compare script ran long enough that the outer command timed out with:

```text
Command exited with code 124
```

The run itself had mostly completed; I recovered the partially written result directory, finished the missing report generation manually, and kept the saved artifacts instead of rerunning blindly.

Second, the first `perf stat` comparison method in `scripts/20-perf-stat-and-thread-compare-go-vs-gst-launch.sh` attached too early to just the process PID and badly undercounted the Go-hosted case. The most obviously wrong number was:

```text
| task-clock | 154.16 | msec |
```

for an 8-second workload that had clearly done real work and produced valid output. I treated that run as methodologically useful for thread sampling but not trustworthy for the whole-thread-set cost comparison, then wrote `scripts/21-perf-stat-threadgroup-compare-go-vs-gst-launch.sh` to attach after warmup to the entire current thread set instead.

### What I learned

The current direct-recording problem statement is now much sharper.

1. The direct path in the app is built imperatively in Go, but that by itself is not the current main explanation.
2. The direct-path `Y444` bug was real, but fixing it did not erase the gap.
3. Once the direct path is corrected to `I420`, the remaining gap is no longer best explained by graph shape.
4. Mixed-stack perf says the direct running pipeline is mostly native `libx264` + GStreamer work, not ordinary Go application code.
5. The page-fault delta in the warmed-up threadgroup compare is now the strongest current signal:
   - Go-hosted manual harness: `134832` page faults
   - matched `gst-launch`: `232` page faults

That shifts the next investigation target away from graph construction and toward allocator / memory-layout / first-touch / compaction / THP style questions.

### What was tricky to build

The trickiest conceptual part was separating “a real bug we found” from “the final explanation.” The `Y444` discovery was a satisfying graph-level bug, but it would have been a mistake to stop there. The post-fix A/B matrix, perf compare, and threadgroup perf-stat compare were all necessary to prove that the graph bug was only one part of the story.

The trickiest practical part was measurement quality. Several of the recent scripts needed one more iteration before their outputs were trustworthy. The timeout-recovery path in the perf compare and the undercount in the first PID-only `perf stat` method were good reminders that replayable investigations need not just automation, but automation with the right attachment semantics.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/recording.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/bus.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/reference/04-direct-recording-hosting-gap-investigation-report.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/reference/05-online-research-query-packet-for-go-hosted-gstreamer-performance.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/21-perf-stat-threadgroup-compare-go-vs-gst-launch/20260415-012510/01-summary.md`

The main review questions are:

- is the current “native hosting / memory-fault effect” interpretation the most conservative reading of the evidence?
- and does the new online query packet target the right upstream questions rather than wasting time on already-ruled-out graph-shape theories?

### What should be done in the future

- Validate the updated ticket with `docmgr doctor --ticket SCS-0016 --stale-after 30`.
- Upload the new report + query packet bundle to reMarkable.
- Follow the query packet with internet-enabled research focused on allocator, page-fault, THP/compaction, GLib hosting, and generic embedded-app-vs-`gst-launch` differences.
- Add one more local measurement slice that breaks the current page-fault delta into more specific categories if the online research suggests a promising direction.

### Code review instructions

Start with the new report:

- `reference/04-direct-recording-hosting-gap-investigation-report.md`

Then review the exact raw evidence it cites:

- `scripts/results/16-gst-launch-direct-full-desktop-match-app/20260415-004852/01-summary.md`
- `scripts/15-go-direct-record-full-desktop-harness/results/20260415-005501/dot/0.00.01.015137962-direct-full-desktop-playing.dot`
- `scripts/15-go-direct-record-full-desktop-harness/results/20260415-010455/dot/0.00.01.020657234-direct-full-desktop-playing.dot`
- `scripts/results/18-direct-harness-ab-matrix/20260415-010905/02-summary.md`
- `scripts/results/19-perf-compare-go-manual-vs-gst-launch/20260415-011757/01-summary.md`
- `scripts/results/21-perf-stat-threadgroup-compare-go-vs-gst-launch/20260415-012510/01-summary.md`

Finally, review the internet handoff packet:

- `reference/05-online-research-query-packet-for-go-hosted-gstreamer-performance.md`

### Technical details

The two numbers that currently matter most are:

```text
matched gst-launch avg_cpu: 128.00%
Go-hosted threadgroup page-faults: 134832
matched gst-launch threadgroup page-faults: 232
```

And the most important raw-profile sanity check is still that the mixed-stack Go-hosted profile is dominated by native media work, not ordinary Go userland:

```text
42.48% / 38.23%  libx264.so.164                x264_8_trellis_coefn
24.31%           libc.so.6                     clone3
12.91%           libgstreamer-1.0.so.0.2402.0  gst_pad_push
0.37%            manual-direct-harness-bin     runtime.asmcgocall.abi0
```

## Step 8: Added the Small-Graph Hosting Ladder and Localized the First Strong Divergence to `x264enc`

At this point the investigation had a good full-graph story but still lacked a clean first-divergence point. We knew the full direct graph was much hotter in Go than in Python or `gst-launch`, and we knew the large page-fault delta made a memory-behavior explanation plausible. But without shrinking the graph, it was still too easy to keep arguing vaguely about “the whole pipeline” instead of identifying the first stage where Go actually peels away from the cooler controls.

So I wrote a dedicated debugging plan for a smaller-graph ladder and then implemented the ladder as ticket-local controls for Go manual, Python manual, and `gst-launch`. The key outcome is strong: Go stays aligned with Python and `gst-launch` through the pre-encode stages, and the first real break happens exactly when `x264enc` enters the graph. The page-fault story lines up with the same boundary, which makes this the cleanest localization result in the whole ticket so far.

### Prompt Context

**User prompt (verbatim):** "alright, add a detailed debugging plan, and then go through all these steps"

**Assistant interpretation:** Write a concrete stage-by-stage plan for the smaller-graph investigation and then actually execute that ladder rather than stopping at planning.

**Inferred user intent:** Replace broad suspicion with a disciplined experiment sequence that identifies where the Go-hosted path first becomes abnormal.

### What I did

- Added a new plan doc:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/design-doc/02-small-graph-hosting-ladder-debugging-plan.md`
- Added new ladder controls:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/29-go-manual-stage-ladder-harness/main.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/29-go-manual-stage-ladder-harness/run.sh`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/30-python-manual-stage-ladder-harness/main.py`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/30-python-manual-stage-ladder-harness/run.sh`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/31-gst-launch-stage-ladder.sh`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/32-small-graph-hosting-ladder-matrix.sh`
- Validated the new controls:

```bash
gofmt -w ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/29-go-manual-stage-ladder-harness/main.go
bash -n ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/29-go-manual-stage-ladder-harness/run.sh
python3 -m py_compile ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/30-python-manual-stage-ladder-harness/main.py
bash -n ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/30-python-manual-stage-ladder-harness/run.sh
bash -n ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/31-gst-launch-stage-ladder.sh
bash -n ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/32-small-graph-hosting-ladder-matrix.sh
go build ./ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/29-go-manual-stage-ladder-harness
```

- Ran capture-stage smoke tests for all three host families to catch parser/attachment bugs.
- Ran the full matrix with:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
DURATION_SECONDS=6 bash ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/32-small-graph-hosting-ladder-matrix.sh
```

- Saved the matrix under:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/32-small-graph-hosting-ladder-matrix/20260415-033745/`

### Why

The smaller-graph ladder was the fastest honest way to answer the question that the full graph could not: *where does Go first become weird?* Once that question is answered, later controls like cgo optimization flags, allocator knobs, or deeper fault instrumentation become much easier to interpret.

### What worked

The ladder gave a very clean result.

The pre-encode stages stay aligned across all three hosts:

```text
capture:
  Go         1.00%   page-faults 28
  Python     2.16%   page-faults 61
  gst-launch 1.00%   page-faults 0

convert:
  Go         1.00%   page-faults 28
  Python     2.17%   page-faults 59
  gst-launch 1.00%   page-faults 0

rate-caps:
  Go         37.83%  page-faults 34
  Python     36.83%  page-faults 51
  gst-launch 36.67%  page-faults 0
```

The first strong divergence appears at the encoder boundary:

```text
encode:
  Go         199.97% page-faults 312438
  Python     125.17% page-faults 62
  gst-launch 129.00% page-faults 29

parse:
  Go         170.50% page-faults 275291
  Python     126.50% page-faults 163
  gst-launch 128.44% page-faults 4

mux-file:
  Go         197.17% page-faults 288738
  Python     131.17% page-faults 209
  gst-launch 131.11% page-faults 233
```

This is the strongest evidence yet that the best next target is the `x264enc` input / memory-behavior boundary rather than raw capture or early graph structure.

### What didn't work

Two small but real harness bugs showed up and needed correction before the matrix could be trusted.

First, the first version of the ladder wrappers parsed the wrong `pidstat` column for Go and Python, which produced fake `0.00%` CPU summaries for the smoke runs. The underlying logs were fine; the parser was wrong because it assumed a fixed field position instead of using the `%CPU` field near the end of the line.

Second, the first wrapper versions launched `pidstat` synchronously and only attempted `perf stat` afterward, which meant the process had often already exited by the time `perf` tried to attach. I fixed that by running `pidstat` in the background, attaching `perf stat` while the process was still alive, and then waiting for both samplers.

### What I learned

The current hosting-gap explanation is now much sharper:

1. The raw capture path is not where the Go path first diverges.
2. `videoconvert` alone is not where the divergence begins.
3. The shaped raw-video pacing stage (`videorate + I420 caps`) is also not the first problem.
4. The first strong divergence appears exactly when `x264enc` enters the graph.
5. The page-fault spike appears at the same stage boundary, which strengthens the memory-behavior interpretation over a pure graph-construction theory.

### What was tricky to build

The tricky part was making the ladder small enough to localize a boundary without accidentally changing the workload class. Earlier in the ticket, resolution mismatches confused some comparisons, so for this ladder I kept the full desktop workload class and only removed graph segments one stage at a time.

The other tricky part was measurement discipline. It would have been easy to declare success from the first manifest, but the early `pidstat` parsing bug would have made that dishonest. The smoke reruns were necessary to ensure the ladder summaries were actually reading the same data the raw logs contained.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/design-doc/02-small-graph-hosting-ladder-debugging-plan.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/29-go-manual-stage-ladder-harness/main.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/30-python-manual-stage-ladder-harness/main.py`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/32-small-graph-hosting-ladder-matrix.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/32-small-graph-hosting-ladder-matrix/20260415-033745/02-summary.md`

The review questions are:

- does the ladder look fairly matched across the three host families?
- and is `x264enc` the conservative first-divergence call, or does any earlier stage still deserve more suspicion than the matrix suggests?

### What should be done in the future

- Add a focused next control around the encoder boundary, such as rebuilding the Go harness with explicit `CGO_CFLAGS=-O2`, to see whether cgo wrapper optimization materially changes the `encode` stage.
- Add a focused fault/memory capture around the Go `encode` stage specifically, rather than only the full direct graph.
- If needed, add one more intermediate ladder variant that isolates `x264enc` input behavior even more tightly, but only if the next encoder-boundary controls remain ambiguous.

### Code review instructions

Start with the topline result:

- `scripts/results/32-small-graph-hosting-ladder-matrix/20260415-033745/02-summary.md`

Then inspect the exact stage rows in:

- `scripts/results/32-small-graph-hosting-ladder-matrix/20260415-033745/01-manifest.tsv`

Then review the three control families:

- `scripts/29-go-manual-stage-ladder-harness/main.go`
- `scripts/30-python-manual-stage-ladder-harness/main.py`
- `scripts/31-gst-launch-stage-ladder.sh`

### Technical details

The single most important matrix boundary is:

```text
rate-caps:
  Go 37.83% / faults 34
encode:
  Go 199.97% / faults 312438
```

That boundary is why the next code-change target should now be framed as:

```text
Go-hosted encoder-input / memory-behavior around x264enc
```

not simply “the whole Go-hosted GStreamer pipeline.”

## Step 9: Tested Software Encoder Contrast and Then Tested CGO Build Flags on the Go `x264enc` Path

Once the small-graph ladder localized the first strong divergence to the encode boundary, the next useful question was whether that anomaly belonged to encoding in general or to a narrower encoder family. I answered that by making the stage ladder accept an explicit encoder and then running a focused encode-stage matrix across `x264enc` and `openh264enc` for Go, Python, and `gst-launch`. I also tried `vaapih264enc`, but that path did not complete cleanly on this machine and is currently being treated as a separate hardware/runtime caveat rather than part of the main comparison.

The software-encoder result was the important narrowing step: the anomaly reproduces strongly with `x264enc`, but not with `openh264enc`. After that, I tested the narrower suspicion that maybe the Go path was only bad because the thin cgo wrapper code was being compiled without useful optimization flags. I added a small Go-only matrix for the `x264enc` encode stage under default, `CGO_CFLAGS=-O2`, and `CGO_CFLAGS=-O3`. That first pass did not rescue the anomaly, which makes the simple “unoptimized cgo glue” explanation much less convincing.

### Prompt Context

**User prompt (verbatim):** "go ahead"

**Assistant interpretation:** Proceed with the encoder-contrast and follow-up controls rather than stopping at the old-codec/available-encoders discussion.

**Inferred user intent:** Use the newly narrowed encoder boundary to run more discriminating controls and keep reducing the space of plausible explanations.

### What I did

- Extended the stage-ladder controls so they now accept encoder choice:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/29-go-manual-stage-ladder-harness/main.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/30-python-manual-stage-ladder-harness/main.py`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/31-gst-launch-stage-ladder.sh`
- Added the focused encoder matrix:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/33-encode-stage-encoder-contrast-matrix.sh`
- Ran the software-encoder matrix and saved:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/33-encode-stage-encoder-contrast-matrix/20260415-035541/`
- Added a Go-only build-flag control:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/34-go-x264-cgo-flag-matrix.sh`
- Ran the build-flag matrix and saved:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/34-go-x264-cgo-flag-matrix/20260415-040159/`

### Why

The encoder-contrast matrix answers whether the anomaly is attached to the generic concept of encoding or to a much narrower encoder implementation. The cgo build-flag matrix answers a different question: whether the remaining `x264enc` anomaly could plausibly be blamed on low-optimization compilation of the thin cgo wrapper code instead of the deeper hosted `libx264` path.

### What worked

The software-encoder matrix was very informative.

For `x264enc` the anomaly still reproduced clearly:

```text
Go         168.83% avg CPU   page-faults 276940
Python     127.17%           page-faults 61
gst-launch 129.67%           page-faults 2
```

For `openh264enc` it did not:

```text
Go         58.00% avg CPU    page-faults 53
Python     87.50%            page-faults 110
gst-launch 88.11%            page-faults 53
```

That means the current anomaly is not well described as “Go-hosted encoding is bad.” It is much better described as “the Go-hosted `x264enc` / `libx264` path is bad.”

The build-flag control also answered its question, even though the run-to-run CPU values were noisy:

```text
default   254.33% avg CPU   page-faults 284806
-O2       349.47%           page-faults 262226
-O3       272.02%           page-faults 283258
```

The important negative result is that neither `-O2` nor `-O3` made the anomaly disappear.

### What didn't work

The initial `vaapih264enc` attempt was not usable as part of the main comparison. The matrix appeared to wedge on that path, which matches the user’s observation that it looked hung. I treated that as a machine-specific hardware/runtime issue and reran the comparison with just the two software encoders that completed cleanly.

The exact CPU numbers in the cgo build-flag matrix were also noisier than I would like, especially compared with some earlier `x264enc` encode-stage runs. That means the cgo matrix should be read as a **negative rescue test**, not as a precise ranking of `default` vs `-O2` vs `-O3`.

### What I learned

Three things are now much clearer.

1. The anomaly is narrower than “Go-hosted encoding.”
2. It is strongly associated with the `x264enc` path specifically.
3. It is not plausibly explained just by low optimization of the thin cgo wrapper code.

That shifts the best next target again. The current best question is no longer “what is wrong with Go hosting in general?” It is more like:

- what is different about the Go-hosted `x264enc` / `libx264` memory behavior,
- and why does that difference not reproduce the same way with `openh264enc`?

### What was tricky to build

The trickiest part here was keeping the comparison fair while changing only one thing at a time. The encoder-contrast matrix needed to stay on the same encode stage and same host families while only swapping the encoder. The cgo matrix needed to stay on the same Go `x264enc` path while only changing build flags.

The other tricky part was not over-trusting the VAAPI path. It would have been easy to spend a long time trying to “fix” the hardware path mid-investigation, but for the current question that would mostly have been a distraction. The software-encoder comparison already answered the important narrowing question.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/33-encode-stage-encoder-contrast-matrix.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/33-encode-stage-encoder-contrast-matrix/20260415-035541/02-summary.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/34-go-x264-cgo-flag-matrix.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/34-go-x264-cgo-flag-matrix/20260415-040159/02-summary.md`

The review questions are:

- does the software-encoder contrast justify focusing much more specifically on `x264enc` / `libx264`?
- and is the cgo build-flag result strong enough to deprioritize the simple wrapper-optimization theory?

### What should be done in the future

- Add a tighter `x264enc`-specific investigation slice, not just generic encoder work.
- Prefer `x264enc`-specific memory/fault/pool comparisons over more broad graph surgery.
- If another contrast encoder is needed, use it only if it helps isolate `libx264`-specific behavior further.
- Treat VAAPI as a separate machine/runtime investigation if it becomes important later, not as a blocker for the current software-encoder conclusion.

### Code review instructions

Start with:

- `scripts/results/33-encode-stage-encoder-contrast-matrix/20260415-035541/02-summary.md`
- `scripts/results/34-go-x264-cgo-flag-matrix/20260415-040159/02-summary.md`

Then inspect the harness surfaces that changed:

- `scripts/29-go-manual-stage-ladder-harness/main.go`
- `scripts/30-python-manual-stage-ladder-harness/main.py`
- `scripts/31-gst-launch-stage-ladder.sh`
- `scripts/33-encode-stage-encoder-contrast-matrix.sh`
- `scripts/34-go-x264-cgo-flag-matrix.sh`

### Technical details

The two most important conclusions from this step are captured by these two comparisons:

```text
x264enc:
  Go 168.83% / 276940 faults
  Python 127.17% / 61 faults
  gst-launch 129.67% / 2 faults
```

versus

```text
openh264enc:
  Go 58.00% / 53 faults
  Python 87.50% / 110 faults
  gst-launch 88.11% / 53 faults
```

and the negative rescue result:

```text
CGO_CFLAGS=-O2 / -O3 did not make the Go x264enc anomaly go away
```

## Step 10: Ran a Focused `x264enc` Property Ablation and Ruled Out the Easy “Expensive x264 Settings” Explanation

After the software-encoder contrast and the cgo build-flag control, one easy explanation was still hanging around: maybe the Go-hosted `x264enc` path was only bad because our chosen x264 configuration happened to be unusually expensive. The most tempting suspect was `trellis`, because the earlier perf traces had shown `x264_8_trellis_coefn` prominently. A second suspect was simply the `veryfast + zerolatency` preset/tune combination.

So I added explicit x264 property controls to the stage harnesses and ran a focused ablation across Go, Python, and `gst-launch`. The result was another important negative control. Cheaper/faster x264 settings improved or stabilized Python and `gst-launch`, but they did **not** rescue the Go path. That means the current anomaly is not well explained as “Go is just spending more time in expensive x264 settings.” The remaining suspicion is now more about the Go-hosted interaction with `x264enc` itself than about the exact x264 quality knob values.

### Prompt Context

**User prompt (verbatim):** "go ahead"

**Assistant interpretation:** Continue the next focused x264-specific experiment rather than stopping at the previous CGO build-flag result.

**Inferred user intent:** Keep narrowing the x264-specific branch until the remaining explanations are very concrete.

### What I did

- Extended the stage harnesses to accept x264-specific properties:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/29-go-manual-stage-ladder-harness/main.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/29-go-manual-stage-ladder-harness/run.sh`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/30-python-manual-stage-ladder-harness/main.py`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/30-python-manual-stage-ladder-harness/run.sh`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/31-gst-launch-stage-ladder.sh`
- Added a focused matrix runner:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/35-x264-property-ablation-matrix.sh`
- Ran the matrix and saved:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/35-x264-property-ablation-matrix/20260415-042332/`

### Why

This was the right next negative control after the encoder-family and build-flag comparisons. If the Go anomaly disappeared when x264 was made simpler or faster, that would point toward an algorithmic-complexity explanation. If it stayed, the remaining explanation would have to be more about hosting interaction than about specific x264 quality knobs.

### What worked

The matrix was highly informative.

Baseline still showed the familiar anomaly:

```text
baseline (speed=3 tune=4 trellis=true)
Go         165.88%  faults 233962
Python     145.50%  faults 63
gst-launch 144.11%  faults 2
```

Removing `zerolatency` did **not** help Go:

```text
no_tune (speed=3 tune=0 trellis=true)
Go         301.67%  faults 187211
Python     146.50%  faults 10607
gst-launch 128.11%  faults 3534
```

Disabling `trellis` did **not** help Go either, even though the earlier perf symbol had made it look suspicious:

```text
no_trellis (speed=3 tune=4 trellis=false)
Go         343.56%  faults 179717
Python     139.33%  faults 62
gst-launch 137.44%  faults 0
```

And the strongest contrast is `ultrafast`:

```text
ultrafast (speed=1 tune=4 trellis=false)
Go         332.00%  faults 136321
Python      83.17%  faults 95
gst-launch  79.33%  faults 23
```

That is the most important pattern in this whole step. The cheaper x264 setting behaves as expected for Python and `gst-launch`, but not for Go.

### What didn't work

Nothing failed materially in the ablation itself. The results are noisy in absolute CPU terms, but the directional conclusion is consistent enough to be useful.

The only caveat is interpretive: I do **not** want to claim that each specific variant ranking is stable to the last percent. What is stable is the broader negative result that the Go path remains abnormally hot across all four tested x264 variants.

### What I learned

This step rules out another easy explanation.

The anomaly is not well explained by:

- `zerolatency` tune alone,
- `trellis` alone,
- or simply using a slower x264 preset.

If anything, the `ultrafast` result makes the negative case stronger: Python and `gst-launch` become much cheaper, while Go remains very hot.

That means the remaining explanation is more likely about:

- Go-hosted `x264enc` element interaction,
- `x264enc` / `libx264` buffer-pool or allocation behavior under the Go-hosted process,
- or another host-process effect that survives even when the x264 configuration itself is made simpler.

### What was tricky to build

The tricky part was making the ablation comparable across all three host families without accidentally changing anything except the intended x264 properties. That required carrying the same x264 knobs through the Go manual harness, Python manual harness, and `gst-launch` ladder in parallel.

The other tricky part was resisting the urge to overfit to the `trellis` symbol from the earlier perf run. It was a good suspect, but this matrix shows that simply turning `trellis` off does not solve the Go anomaly.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/35-x264-property-ablation-matrix.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/35-x264-property-ablation-matrix/20260415-042332/02-summary.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/29-go-manual-stage-ladder-harness/main.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/30-python-manual-stage-ladder-harness/main.py`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/31-gst-launch-stage-ladder.sh`

The key review question is whether this matrix is now strong enough to stop blaming individual x264 settings and move on to deeper `x264enc`/buffer/memory interaction hypotheses.

### What should be done in the future

- Prefer x264-specific memory / pool / allocator investigation over more x264 property tweaking.
- If another x264-focused experiment is needed, it should probably inspect buffer-pool / memory behavior rather than more preset/tune combinations.
- Consider whether a lower-level native allocator/fault comparison should now be done specifically for the Go `x264enc` path and not for the whole generic direct graph.

### Code review instructions

Start with:

- `scripts/results/35-x264-property-ablation-matrix/20260415-042332/02-summary.md`

Then inspect the detailed manifest:

- `scripts/results/35-x264-property-ablation-matrix/20260415-042332/01-manifest.tsv`

Then review the x264-property plumbing in:

- `scripts/29-go-manual-stage-ladder-harness/main.go`
- `scripts/30-python-manual-stage-ladder-harness/main.py`
- `scripts/31-gst-launch-stage-ladder.sh`

### Technical details

The most important single comparison from this step is:

```text
ultrafast:
  Go         332.00% / 136321 faults
  Python      83.17% / 95 faults
  gst-launch  79.33% / 23 faults
```

That is why the current best summary is now:

```text
The Go anomaly persists across x264 property simplifications; it is not just an expensive-x264-settings problem.
```
