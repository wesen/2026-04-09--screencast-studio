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
LastUpdated: 2026-04-14T20:46:00-04:00
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
