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
LastUpdated: 2026-04-14T23:09:52-04:00
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
