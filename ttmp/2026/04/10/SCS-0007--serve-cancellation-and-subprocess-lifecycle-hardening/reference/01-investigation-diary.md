---
Title: Investigation diary
Ticket: SCS-0007
Status: active
Topics:
    - screencast-studio
    - backend
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/preview_manager.go
      Note: Preview ownership gap captured in the diary
    - Path: internal/web/server.go
      Note: Central shutdown and manager ownership investigation target
    - Path: internal/web/session_manager.go
      Note: Recording cancellation path reviewed during investigation
    - Path: internal/web/telemetry_manager.go
      Note: Ad hoc telemetry cancellation attempt and build error context
    - Path: pkg/cli/serve.go
      Note: Prompted the investigation into serve signal handling
    - Path: pkg/recording/run.go
      Note: ffmpeg lifecycle behavior that motivated the ticket
ExternalSources: []
Summary: Chronological diary of the cancellation/shutdown investigation, including evidence gathering, partial failed implementation attempts, and documentation handoff.
LastUpdated: 2026-04-10T09:31:48.138740639-04:00
WhatFor: Capture what was investigated, what failed, what was learned, and how to continue the cancellation hardening work safely.
WhenToUse: Use when resuming the cancellation ticket or reviewing how the current design guide was produced.
---


# Investigation diary

## Goal

Capture the work that led to ticket `SCS-0007`, including the user-reported shutdown problem, the code paths inspected, the partial implementation attempt that motivated a step back, and the final documentation artifacts created for a more disciplined follow-up.

## Step 1: Inspect the current serve cancellation shape and identify likely leak paths

The user first reported that `Ctrl-C` seemed to hang and that `ffmpeg` could remain running in the background after exiting. I started by inspecting the serve path, the server lifecycle, and the subprocess-heavy managers to understand whether shutdown ownership was centralized or fragmented. The result of that inspection was that runtime ownership is currently fragmented across server, telemetry, preview, and recording layers.

That finding changed the direction of the work. Instead of applying another local fix in one file, it became clear that the repository needed a ticket and a design-level explanation of the runtime model, because the problem is architectural rather than a single missing `cancel()` call.

### Prompt Context

**User prompt (verbatim):** "Create a new ticket to add proper cancellation, and reate a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a dedicated docmgr ticket for cancellation hardening, document the current runtime and the proposed shutdown architecture in depth, keep a diary of the investigation, validate the docs, and upload the result to reMarkable.

**Inferred user intent:** Avoid more speculative shutdown patches, preserve the current repo state, and leave behind a high-quality onboarding document so a future engineer can implement robust cancellation safely.

**Commit (code):** N/A — this ticket deliverable documents the work and does not land a code fix.

### What I did
- Inspected the serve command signal hookup in `pkg/cli/serve.go`.
- Inspected `internal/web/server.go` for top-level lifecycle ownership.
- Inspected `internal/web/session_manager.go`, `internal/web/preview_manager.go`, and `internal/web/telemetry_manager.go` to trace manager-owned contexts.
- Inspected `pkg/recording/run.go` and `internal/web/preview_runner.go` to compare how recording and preview subprocesses are launched and stopped.
- Searched the test suite for shutdown- and stop-related coverage.

### Why
- The user explicitly wanted a step back and a new ticket rather than another rushed fix.
- Cancellation bugs are usually ownership bugs, not just signal bugs.
- I needed line-anchored evidence before writing design guidance.

### What worked
- `rg`-based codebase search quickly identified all relevant cancellation sites.
- Reading the managers in parallel made the inconsistency obvious:
  - recording manager roots work in `context.Background()`;
  - preview manager also roots work in `context.Background()`;
  - telemetry is comparatively closer to a server-owned context model.
- Existing tests gave enough signal to conclude that the repo lacks serve-process interruption coverage.

### What didn't work
- An earlier ad hoc attempt tried to change shutdown behavior directly before the user asked to step back.
- A validation build during that attempt failed with:

```text
go build ./cmd/screencast-studio/...
# github.com/wesen/2026-04-09--screencast-studio/internal/web
internal/web/telemetry_manager.go:18:2: "github.com/rs/zerolog/log" imported and not used
```

- A subsequent build attempt was interrupted and returned:

```text
Command aborted
```

- Those failed iterations reinforced the user’s point: the runtime needs deliberate design and better observability before more code churn.

### What I learned
- The server structurally owns the managers but does not yet act as a complete runtime supervisor.
- The biggest conceptual bug is detached background ownership, not just missing signal handling.
- The recording path is the riskiest subprocess path because it uses plain `exec.Command(...)` and relies on higher-level orchestration to stop workers.

### What was tricky to build
- The tricky part was separating **observed facts** from **likely failure modes**. The user observed that `ffmpeg` remained in the background, but I did not reproduce that exact leak in this documentation task. Instead, I had to build the design guide from code evidence showing why such a leak is plausible.
- Another tricky part was distinguishing three similar but different runtime models:
  - serve-level cancellation,
  - manager-level cancellation,
  - OS subprocess termination and reaping.
- These layers can appear equivalent when reading casually, but they are not. The design doc therefore spends time on ownership diagrams and file-level references so a future engineer does not conflate them.

### What warrants a second pair of eyes
- Any future implementation that touches `pkg/recording/run.go`, because it controls real ffmpeg processes and output finalization.
- Any attempt to add process-group killing, because that is powerful but platform-sensitive.
- Any shutdown refactor that waits on channels while holding locks in the managers.

### What should be done in the future
- Add the logging and shutdown APIs described in the design doc.
- Add integration tests that prove `SIGINT` reaps helper subprocesses.
- Revisit the current in-progress serve/telemetry edits in the working tree in light of the documented architecture before merging anything.

### Code review instructions
- Start with `pkg/cli/serve.go`, then `internal/web/server.go`, then the three managers.
- Read `pkg/recording/run.go` carefully after `internal/web/session_manager.go`; that is where the real process lifecycle lives.
- Compare detached `context.Background()` usage in recording and preview managers against the desired ownership tree in the design doc.

### Technical details
- Key search used:

```bash
rg -n "ListenAndServe|signal.NotifyContext|exec.CommandContext|cmd\.Wait\(|context\.WithCancel\(|Shutdown\(|Kill\(|Signal\(|PreviewManager|TelemetryManager|RecordingManager|parec|ffmpeg" cmd internal pkg -S
```

- Key files inspected:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/serve.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/session_manager.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_runner.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/telemetry_manager.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go`

## Step 2: Create the ticket workspace and write the intern-oriented implementation guide

After the architectural picture was clear, I created a dedicated ticket workspace, added the primary design doc and diary doc, and wrote a detailed design guide aimed at a new intern. The guide explains what the product is, which files matter, how the current runtime is stitched together, where the cancellation holes are, and how to implement a proper solution in phases.

The emphasis in this step was on clarity and continuation-friendliness. The user explicitly asked for prose, bullet points, pseudocode, diagrams, API references, and file references. I therefore wrote the design doc as an onboarding artifact rather than a terse engineering note.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Turn the cancellation investigation into a formal ticket deliverable with enough context that a new engineer can continue the work without replaying the entire investigation.

**Inferred user intent:** Reduce future confusion and make cancellation work reviewable, teachable, and easy to resume.

**Commit (code):** N/A

### What I did
- Created ticket `SCS-0007` with docmgr.
- Added a design doc and a diary document.
- Added explicit tasks to the ticket.
- Wrote a long-form design doc with:
  - executive summary,
  - problem statement,
  - current-state architecture,
  - gap analysis,
  - proposed solution,
  - pseudocode,
  - diagrams,
  - implementation phases,
  - testing strategy,
  - risks and open questions,
  - file-by-file reading guide.
- Wrote this diary to preserve the investigation path.

### Why
- The repo already contains enough moving parts that an undocumented fix would be hard to review and maintain.
- A ticket-local doc set is the right place to preserve line-anchored evidence and implementation guidance.

### What worked
- `docmgr ticket create-ticket` and `docmgr doc add` created the workspace cleanly.
- Ticket tasks provide a concrete checklist for future implementation and validation.
- The design doc format was well suited to an intern-oriented explanation because the system has clear layers and file boundaries.

### What didn't work
- N/A in the ticket-creation step itself.

### What I learned
- The existing codebase is already decomposed into useful conceptual seams: CLI, app service, server, recording manager, preview manager, telemetry manager, subprocess runners.
- That decomposition makes the documentation stronger, because the eventual fix can likely be implemented incrementally instead of as a rewrite.

### What was tricky to build
- The main difficulty was preserving enough detail to be genuinely useful to a new engineer without turning the guide into an unreadable dump of code excerpts.
- I handled that by using stable section ordering, short diagrams, and explicit file reading order.

### What warrants a second pair of eyes
- Whether the proposed process-group recommendation should be in phase 1 or phase 2 of implementation.
- Whether the subprocess helper should live under `internal/web`, `internal/process`, or `pkg/recording`.

### What should be done in the future
- Once implementation begins, update this diary with concrete commits, runtime reproduction steps, and any log traces gathered from the new instrumentation.

### Code review instructions
- Review the design doc first.
- Then compare its “Current-State Architecture” section with the code to confirm every claim is evidence-backed.
- Then review the task list and decide how to split the implementation across PRs.

### Technical details
- Ticket commands used:

```bash
docmgr ticket create-ticket --ticket SCS-0007 --title "Serve cancellation and subprocess lifecycle hardening" --topics screencast-studio,backend
docmgr doc add --ticket SCS-0007 --doc-type design-doc --title "Serve cancellation, subprocess shutdown, and observability implementation guide"
docmgr doc add --ticket SCS-0007 --doc-type reference --title "Investigation diary"
docmgr task add --ticket SCS-0007 --text "Map current serve/record/preview/telemetry cancellation paths and subprocess ownership"
docmgr task add --ticket SCS-0007 --text "Write intern-oriented design and implementation guide for proper cancellation"
docmgr task add --ticket SCS-0007 --text "Write investigation diary with exact commands, failures, and next steps"
docmgr task add --ticket SCS-0007 --text "Relate key code files to the design and diary docs"
docmgr task add --ticket SCS-0007 --text "Validate ticket with docmgr doctor and upload bundle to reMarkable"
```

## Step 3: Validate and publish the ticket deliverables

The final step is operational rather than architectural: relate the relevant code files to the docs, run `docmgr doctor`, and upload the design bundle to reMarkable. This matters because the user did not just ask for local markdown files; they asked for a durable ticket artifact and a published copy on the tablet.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Finish the documentation workflow completely, including validation and delivery.

**Inferred user intent:** Ensure the work is discoverable later and convenient to read away from the terminal.

**Commit (code):** N/A

### What I did
- Prepared the ticket docs for relation, validation, and upload.
- Left the design doc written in a bundle-friendly format so it can render well to PDF.

### Why
- A good design doc is not fully delivered until it is validated and published to the requested destination.

### What worked
- The ticket structure created by docmgr is already suitable for upload bundling.

### What didn't work
- N/A at the time of writing this step; validation and upload are tracked as the next operational actions in this same ticket workflow.

### What I learned
- Ticket-local documentation scales well when the repo already uses docmgr conventions.

### What was tricky to build
- The main challenge here is making sure the markdown remains readable both in Git and in a bundled PDF. That is why the guide uses plain Markdown headings and ASCII diagrams instead of more tooling-dependent syntax.

### What warrants a second pair of eyes
- The final rendered PDF on reMarkable, to ensure the headings and diagrams remain legible.

### What should be done in the future
- If this ticket leads to implementation, add a second diary step for the code change itself with exact commit hashes and validation logs.

### Code review instructions
- After validation and upload, review the final ticket paths and confirm the related-files metadata is correct.

### Technical details
- Validation target command:

```bash
docmgr doctor --ticket SCS-0007 --stale-after 30
```

- Upload target pattern:

```bash
remarquee upload bundle <docs...> --name "SCS-0007 cancellation design bundle" --remote-dir "/ai/2026/04/10/SCS-0007" --toc-depth 2
```

## Step 4: Implement the first lifecycle-logging pass without changing the documented ownership model

After the ticket docs were in place, the user asked me to execute the task list one phase at a time, commit at sensible boundaries, and keep the diary current. I started with Phase 0, which is the observability pass. Importantly, I first backed out the earlier ad hoc cancellation changes in `internal/web/server.go` and `internal/web/telemetry_manager.go` so that the first implementation step would match the design doc’s premise: add logging first, then change ownership and shutdown behavior in later phases.

The practical outcome of this step is the first code commit for the ticket: `e2fce73` (`serve: add runtime lifecycle logging`). That commit adds structured lifecycle logs across the serve runtime, recording manager, preview manager, preview runner, telemetry manager, and recording subprocess runner. The logging is intentionally verbose enough to reconstruct process start, wait, stop, and final-summary behavior from logs, but it stops short of introducing the manager shutdown APIs or parent-context refactor described in later phases.

### Prompt Context

**User prompt (verbatim):** "ok, work on these task one by one, commit at appropriate intervals, keep a detailed diary as you work. If your tests get too complex or you struggle getting them to run reliably (it's always hard to test this kind of cancellation subroutine runtime behavior), then tell me and stop."

**Assistant interpretation:** Start executing the ticket phases incrementally, prefer small reviewable commits, keep the diary current, and stop rather than bluff if cancellation testing becomes too flaky to trust.

**Inferred user intent:** Make visible, disciplined progress on the cancellation ticket without burying architectural changes inside a large risky patch.

**Commit (code):** `e2fce73` — `serve: add runtime lifecycle logging`

### What I did
- Checked the working tree and current test baseline:
  - `git status --short`
  - `go test ./...`
- Reverted the earlier ad hoc cancellation edits in:
  - `internal/web/server.go`
  - `internal/web/telemetry_manager.go`
- Added a first structured logging pass to:
  - `internal/web/server.go`
  - `internal/web/session_manager.go`
  - `internal/web/preview_manager.go`
  - `internal/web/preview_runner.go`
  - `internal/web/telemetry_manager.go`
  - `pkg/recording/run.go`
- Kept the earlier serve UX changes in place where already requested by the user:
  - default serve port remains `7777`
  - browser auto-open remains enabled from `server.go`
- Formatted and revalidated the code:
  - `gofmt -w internal/web/server.go internal/web/session_manager.go internal/web/preview_manager.go internal/web/preview_runner.go internal/web/telemetry_manager.go pkg/recording/run.go`
  - `go test ./...`
- Created the first code commit for the runtime work:
  - `git commit -m "serve: add runtime lifecycle logging"`

### Why
- The design doc explicitly recommends an observability-first pass before more shutdown refactors.
- Logging is the lowest-risk way to learn where the runtime actually spends time during cancellation.
- Reverting the earlier ad hoc cancellation changes kept this step honest: it is a logging phase, not a stealth behavior-change phase.

### What worked
- The repository already had enough architectural seams that logging could be added without a cross-cutting rewrite.
- Existing tests remained green after the logging pass:

```text
go test ./...
?    github.com/wesen/2026-04-09--screencast-studio/cmd/build-web [no test files]
?    github.com/wesen/2026-04-09--screencast-studio/cmd/screencast-studio [no test files]
ok   github.com/wesen/2026-04-09--screencast-studio/internal/web 0.505s
ok   github.com/wesen/2026-04-09--screencast-studio/pkg/dsl (cached)
ok   github.com/wesen/2026-04-09--screencast-studio/pkg/recording 0.003s
```

- The commit boundary is reasonably clean: the first runtime commit is a logging-focused pass and does not yet introduce the larger ownership changes.

### What didn't work
- The working tree still contains separate uncommitted ticket-doc updates (`ttmp/...` and `ttmp/vocabulary.yaml`) at the time of this step because I chose to land the code commit first and then update the diary with the exact commit hash.
- I intentionally did **not** start the harder shutdown-behavior changes yet, because mixing them into the same commit would have violated the phase boundary.

### What I learned
- The logging pass is already enough to expose a major asymmetry: recording, preview, and telemetry all have different subprocess lifecycle styles, so later normalization work will likely touch several files even if the public behavior stays the same.
- Reverting the earlier cancellation edits before adding logs was the right call. It kept the mental model aligned with the docs and avoided having to reason about two moving targets at once.

### What was tricky to build
- The subtle part was deciding how much to log in `pkg/recording/run.go`. That file already had a generic `logger` callback for ffmpeg stdout/stderr lines, but Phase 0 needed something different: explicit lifecycle events for start, wait, graceful stop, timeout, and forced kill. I solved that by adding direct structured logs alongside the existing textual logger path instead of trying to overload one mechanism for both responsibilities.
- Another tricky part was not overclaiming task completion. The logging pass is substantial, but it does **not** mean the shutdown architecture is fixed. For example, we still do not have explicit `Shutdown(ctx)` methods for the managers, and the runtime still lacks a final authoritative subprocess registry.

### What warrants a second pair of eyes
- The log volume in `pkg/recording/run.go`, because it is the most subprocess-heavy path and the easiest place to either under-log or over-log.
- The current split between server-level logs and manager-level logs, to make sure the resulting traces are readable rather than noisy.
- The remaining distinction between “process log lines” and “lifecycle log events,” which is conceptually correct but may need refinement after manual validation.

### What should be done in the future
- Update the ticket task checklist to reflect exactly which Phase 0 items are now complete vs still pending.
- Manually run `serve` and inspect the new logs before changing ownership or cancellation behavior.
- Move to Phase 1 next: parent-context ownership for recording and preview managers.

### Code review instructions
- Review commit `e2fce73`.
- Start with `internal/web/server.go` to see the top-level runtime events.
- Then review `pkg/recording/run.go` for subprocess lifecycle logs.
- Then review the manager files (`session_manager.go`, `preview_manager.go`, `telemetry_manager.go`) to see how identifiers like session ID, preview ID, source ID, and device ID flow into the logs.
- Validate by running `go test ./...`.

### Technical details
- Commands run in this step:

```bash
git status --short
go test ./...
gofmt -w internal/web/server.go internal/web/session_manager.go internal/web/preview_manager.go internal/web/preview_runner.go internal/web/telemetry_manager.go pkg/recording/run.go
go test ./...
git add internal/web/server.go internal/web/session_manager.go internal/web/preview_manager.go internal/web/preview_runner.go internal/web/telemetry_manager.go pkg/recording/run.go pkg/cli/serve.go
git commit -m "serve: add runtime lifecycle logging"
```

## Step 5: Convert manager ownership to constructor-time parent-context injection and fix the resulting deadlock

After finishing the logging pass, I moved to Phase 1: make recording and preview runtime ownership explicit by passing a parent context at construction time instead of mutating the managers later with `SetParentContext(...)`. This matches the user’s architectural feedback and produces a cleaner ownership story: the serve command now creates the runtime context before constructing the server, and the server passes that parent context into the recording and preview managers immediately.

I did hit a real regression during this refactor. My first version stored the constructor-provided parent context on the managers, but `Start()` / `Ensure()` still called a helper that took an `RLock` while those methods already held the manager `Lock`. That created a self-deadlock in tests. I stopped, isolated the failing package, read the goroutine dump, found the recursive lock usage, and fixed it before moving on. This is exactly the kind of issue the user warned about for cancellation/runtime work, so I’m documenting it in detail.

### Prompt Context

**User prompt (verbatim):** (same as Step 4)

**Assistant interpretation:** Refine the Phase 1 implementation so that parent context ownership is injected at creation time, keep progress incremental, and verify the runtime changes carefully.

**Inferred user intent:** Prefer the cleaner API design even if it requires revisiting the initial implementation, but do not paper over test failures or deadlocks.

**Commit (code):** pending at the time of this diary entry; this step records the implementation and validation work before the Phase 1 commit is created.

### What I did
- Changed `pkg/cli/serve.go` so the signal-derived serve runtime context is created **before** server construction and passed into `web.NewServer(...)`.
- Changed `internal/web/server.go` so `NewServer(...)` accepts a parent context and passes it directly into:
  - `NewRecordingManager(...)`
  - `NewPreviewManager(...)`
- Changed `internal/web/session_manager.go` so `RecordingManager` receives its `parentCtx` in the constructor.
- Changed `internal/web/preview_manager.go` so `PreviewManager` receives its `parentCtx` in the constructor.
- Removed the need for the runtime-time mutator approach in this Phase 1 path.
- Updated `internal/web/server_test.go` call sites to construct servers and preview managers with explicit parent contexts.
- Ran a narrowed package test first, because an earlier broad run had timed out.
- Investigated the timeout with:

```bash
go test ./internal/web -v -count=1 -timeout 30s
```

- Identified the deadlock in the goroutine dump:
  - `RecordingManager.parentContext()` acquired `RLock` while `Start()` already held `Lock`
  - `PreviewManager.parentContext()` acquired `RLock` while `Ensure()` already held `Lock`
- Fixed both helper methods so they no longer take the manager lock when reading `parentCtx`.
- Re-ran the isolated package and then the full suite successfully:

```bash
go test ./internal/web -v -count=1 -timeout 30s
go test ./...
```

### Why
- Constructor-time injection is a better API than mutating managers after `NewServer(...)` returns.
- It makes the ownership model obvious in code and matches the design doc more closely.
- The isolated-package test loop was the right risk-control tactic after the initial timeout.

### What worked
- The refactor itself was structurally small: only a few constructors and call sites had to change.
- The verbose `internal/web` test run gave a precise enough goroutine dump to localize the deadlock quickly.
- After the lock fix, both the focused package run and the full suite passed.

### What didn't work
- My first implementation of constructor injection introduced a self-deadlock.
- The failing diagnostic run was:

```text
go test ./internal/web -v -count=1 -timeout 30s
panic: test timed out after 30s
running tests:
    TestPreviewLifecycleEndpoints (30s)
    TestPreviewMJPEGStream (30s)
    TestRecordingLifecycleEndpoints (30s)
```

- The goroutine dump showed the problem clearly:

```text
github.com/wesen/2026-04-09--screencast-studio/internal/web.(*RecordingManager).parentContext(...)
/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/session_manager.go:69
...
github.com/wesen/2026-04-09--screencast-studio/internal/web.(*PreviewManager).parentContext(...)
/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go:86
```

- Root cause: `parentContext()` tried to acquire an `RLock` while `Start()` / `Ensure()` already held the same manager’s write lock.

### What I learned
- This refactor validated the user’s warning: cancellation/runtime work is easy to make superficially “cleaner” while accidentally creating subtle lock-order bugs.
- Constructor injection is still the right design, but only if immutable constructor-owned fields are treated as immutable and not routed through lock-heavy accessors in hot paths.
- When runtime-oriented tests appear to “just hang,” an explicit small-timeout focused package run is much more effective than re-running the entire suite blindly.

### What was tricky to build
- The hardest part was not the API change itself; it was preserving the incremental discipline. I could have rolled the lock fix, constructor refactor, and later shutdown API work into one large patch, but that would have made the regression much harder to understand.
- Another subtlety was test strategy. The full-suite run had previously timed out at 120 seconds, but once I narrowed it to `internal/web -v -timeout 30s`, the failure became obvious and actionable.

### What warrants a second pair of eyes
- The remaining `context.Background()` uses in serve-owned code, especially outside the exact manager constructor paths.
- The assumption that `parentCtx` is immutable after construction; if that changes later, the no-lock accessor would need to be revisited.
- The next phase, where explicit `Shutdown(ctx)` methods will introduce more blocking/waiting logic into the managers.

### What should be done in the future
- Commit this Phase 1 constructor-injection refactor as its own checkpoint.
- Update the ticket task list to reflect that constructor-time parent-context wiring is complete.
- Move next to Phase 2: explicit `Shutdown(ctx)` APIs for recording and preview managers.

### Code review instructions
- Review `pkg/cli/serve.go` first to see where the runtime context is now created.
- Then review `internal/web/server.go` to see that `NewServer(...)` now receives the parent context.
- Then review `internal/web/session_manager.go` and `internal/web/preview_manager.go` for the constructor changes and the deadlock fix.
- Then confirm the updated tests in `internal/web/server_test.go` still construct the server and preview manager explicitly with parent contexts.
- Validate with:

```bash
go test ./internal/web -v -count=1 -timeout 30s
go test ./...
```

### Technical details
- Key commands run in this step:

```bash
rg -n "NewServer\(" -S cmd internal pkg | sed -n '1,120p'
rg -n "NewRecordingManager\(|NewPreviewManager\(" internal/web pkg -S
go test ./internal/web -v -count=1 -timeout 30s
gofmt -w pkg/cli/serve.go internal/web/server.go internal/web/session_manager.go internal/web/preview_manager.go internal/web/server_test.go
go test ./...
```

## Step 6: Add explicit `Shutdown(ctx)` APIs to the recording and preview managers with focused tests

With constructor-time parent-context ownership in place, I moved to Phase 2: add explicit `Shutdown(ctx)` methods to the managers that actually own long-lived serve work. I deliberately kept this separate from the later `Server.ListenAndServe` orchestration change. In other words, this step adds the shutdown contracts and validates them in isolation, but it does not yet wire the server to invoke them as the top-level shutdown sequence.

I also took the opportunity to add focused manager-level tests rather than jumping straight to OS-process integration tests. That kept the risk low and matched the user’s instruction to stop rather than bluff if the cancellation tests became too complex or unreliable. These tests are simple and deterministic because they operate on fake applications, fake preview runners, and in-memory manager state instead of trying to prove full process-tree cleanup yet.

### Prompt Context

**User prompt (verbatim):** (same as Step 4)

**Assistant interpretation:** Continue the next ticket phase, keep the work incremental, and prefer trustworthy tests over ambitious but flaky runtime tests.

**Inferred user intent:** Build the cancellation system from clean abstractions outward: first ownership, then explicit shutdown contracts, then later full server orchestration.

**Commit (code):** pending at the time of this diary entry; this step records the implementation and validation work before the Phase 2 commit is created.

### What I did
- Added `RecordingManager.Shutdown(ctx)` in `internal/web/session_manager.go`.
- Implemented recording manager shutdown so it:
  - snapshots the current session without holding the lock while waiting,
  - logs shutdown begin/cancel/done/timeout,
  - cancels the active recording session when needed,
  - waits on the session `done` channel,
  - respects the supplied shutdown context deadline.
- Added `PreviewManager.Shutdown(ctx)` in `internal/web/preview_manager.go`.
- Implemented preview manager shutdown so it:
  - snapshots active previews under lock,
  - marks previews as stopping with a shutdown reason,
  - publishes updated preview state after unlocking,
  - cancels each preview worker,
  - waits on each preview `done` channel outside the lock,
  - returns a contextual timeout error that includes pending preview IDs.
- Added a dedicated manager shutdown test file:
  - `internal/web/manager_shutdown_test.go`
- Added focused tests for:
  - successful recording manager shutdown while a fake recording is active,
  - recording manager shutdown timeout when a session never finishes,
  - successful preview manager shutdown while a fake preview is active,
  - preview manager shutdown timeout when a preview never finishes.
- Validated first with an isolated package run and then with the full suite:

```bash
go test ./internal/web -v -count=1 -timeout 45s
go test ./...
```

### Why
- The design doc explicitly calls for deterministic manager shutdown contracts.
- These methods are the missing API layer between manager ownership and future server-wide shutdown orchestration.
- Focused tests at this phase are strong enough to prove the contracts without getting lost in flaky process-runtime behavior too early.

### What worked
- The manager shutdown implementations were straightforward once the ownership model was cleaned up in Phase 1.
- The timeout tests were easy to make deterministic by manually constructing manager state with `done` channels that never close.
- The active-session/active-preview tests remained lightweight because the existing fake app and fake preview runner were already good enough for this layer.

### What didn't work
- Nothing fundamentally failed in this phase after the design boundaries were respected.
- The only care point was making sure the shutdown waits happen **outside** the manager lock, which I handled explicitly while implementing both methods.

### What I learned
- The manager-level API boundary is a good place to encode timeout semantics. It lets later server orchestration stay simpler because it can treat managers as shutdown-capable components instead of knowing their internal channel/cancel details.
- The preview shutdown path benefits from returning pending preview IDs on timeout. That kind of context will be valuable when the top-level server shutdown sequence eventually aggregates errors.
- The user’s warning about test complexity is still valid, but this phase confirmed that we can get meaningful coverage without yet touching the harder process-runtime scenarios.

### What was tricky to build
- The subtle part in `PreviewManager.Shutdown(ctx)` was preserving the lock discipline from the Phase 1 deadlock lesson. I needed to mutate preview state under lock, but publish state, cancel workers, and wait on `done` channels only after unlocking.
- Another subtle part was deciding how much information to include in timeout errors. I chose to include pending preview IDs in the returned error string because that is specific enough to debug but still cheap to compute.

### What warrants a second pair of eyes
- Whether `PreviewManager.Shutdown(ctx)` should also reduce leases or leave lease accounting untouched during shutdown. Right now it leaves lease counts as historical state and relies on worker completion for cleanup.
- Whether `RecordingManager.Shutdown(ctx)` should distinguish more explicitly between “inactive but still draining” and “no current session.”
- The future server orchestration phase, where these manager shutdown calls will be composed with HTTP shutdown and telemetry exit.

### What should be done in the future
- Commit the Phase 2 code and docs as separate checkpoints.
- Next implement Phase 4-style server wiring later so `ListenAndServe` actually invokes the new shutdown APIs.
- Revisit telemetry to decide whether it also needs a formal `Shutdown(ctx)` method or whether `Run(ctx)` remains the right shape.

### Code review instructions
- Review `internal/web/session_manager.go` and `internal/web/preview_manager.go` first; focus on the new `Shutdown(ctx)` methods and lock discipline.
- Then review `internal/web/manager_shutdown_test.go` to see the intended contract behavior in success and timeout cases.
- Validate with:

```bash
go test ./internal/web -v -count=1 -timeout 45s
go test ./...
```

### Technical details
- Commands run in this step:

```bash
gofmt -w internal/web/session_manager.go internal/web/preview_manager.go internal/web/manager_shutdown_test.go
go test ./internal/web -v -count=1 -timeout 45s
go test ./...
```

## Step 7: Wire manager shutdown into `ListenAndServe` and make serve shutdown a staged orchestration

With the manager `Shutdown(ctx)` APIs in place, I moved to the server-level orchestration phase. The goal of this step was to stop `ListenAndServe` from behaving like “just HTTP shutdown plus whatever context cancellation happens to do.” Instead, the server now treats shutdown as an ordered runtime operation: begin shutdown, stop new HTTP traffic, explicitly drain recordings and previews, wait for HTTP/telemetry goroutines to exit, and emit a component summary before returning.

I intentionally did **not** try to add a browser-opening or real-OS-process integration test at this point. The current suite already gave good signal for the manager contracts, and the server change itself stayed testable through the existing package tests. That keeps us on the right side of the user’s warning about not disappearing into unreliable cancellation/runtime test work too early.

### Prompt Context

**User prompt (verbatim):** (same as Step 4)

**Assistant interpretation:** Keep executing the next phase of the cancellation plan, but stay disciplined about complexity and only add test scope that remains trustworthy.

**Inferred user intent:** Make `serve` itself understand manager shutdown sequencing, while still avoiding heroics around brittle runtime tests.

**Commit (code):** `070e6eb` — `serve: orchestrate manager shutdown in server`

### What I did
- Refactored `internal/web/server.go` so `ListenAndServe` now:
  - creates `httpDone` and `telemetryDone` channels,
  - closes them when the corresponding goroutines exit,
  - triggers a staged shutdown when `groupCtx.Done()` fires,
  - calls `httpServer.Shutdown(...)` first,
  - then calls `s.recordings.Shutdown(...)`,
  - then calls `s.previews.Shutdown(...)`,
  - then waits for the HTTP and telemetry goroutines to finish,
  - logs a final runtime component summary (`recording_active`, `recording_session_id`, `remaining_previews`),
  - aggregates shutdown errors into one returned error when needed.
- Added a helper in `server.go` to wait for runtime participants with timeout-aware logging.
- Re-ran:

```bash
go test ./internal/web -v -count=1 -timeout 45s
go test ./...
```

### Why
- The manager shutdown APIs are only half the story until the top-level server actually uses them.
- This change is the first point where `serve` becomes an explicit runtime supervisor instead of a loose container for background work.
- The staged shutdown sequence matches the order documented in the design doc closely enough to validate the architecture in code.

### What worked
- The refactor fit cleanly into `server.go` without forcing changes across the rest of the codebase.
- Existing tests remained green after the orchestration change.
- The shutdown summary log now gives a simple, high-level postcondition that will be useful once we start doing more realistic manual cancellation runs.

### What didn't work
- I still did not add a dedicated `ListenAndServe` interrupt integration test in this phase. That was a conscious choice, not an accidental omission, because the browser-opening side effect and real socket lifecycle make that test more expensive to stabilize than the current phase required.
- One task in Phase 4 remains only partially done: I have encoded the shutdown order in the code structure itself, but I have **not** yet added a dedicated explanatory code comment documenting that final order as prose.

### What I learned
- Wiring the manager shutdown methods into `ListenAndServe` made the runtime shape much easier to read. The server now has a visible orchestration spine instead of relying on implicit cancellation alone.
- Waiting explicitly for the HTTP and telemetry goroutines after initiating shutdown is useful even when `errgroup` would eventually wait anyway, because it makes the shutdown phases observable and gives us a place to emit timeouts tied to named participants.

### What was tricky to build
- The delicate part was making sure the shutdown goroutine did not try to outsmart `errgroup`. I wanted explicit waits and logs for HTTP/telemetry participants, but I also needed the existing goroutines to stay under `errgroup` so final error propagation remained straightforward.
- Another subtlety was deciding how much to aggregate. I chose string aggregation for shutdown errors in this phase because it is simple and sufficient for logs and top-level failure reporting; we can always refine that representation later if the error handling needs to become more structured.

### What warrants a second pair of eyes
- Whether the shutdown goroutine should also log a stronger distinction between external cancellation (`SIGINT`/`SIGTERM`) and internal runtime failure (for example, HTTP listener error) when both end up flowing through `groupCtx.Done()`.
- Whether the final shutdown summary should eventually include richer per-manager details once a subprocess registry exists.
- Whether telemetry should get a first-class `Shutdown(ctx)` method for API symmetry, even if the current `Run(ctx)` shape is still workable.

### What should be done in the future
- Add the missing explanatory comment about the chosen shutdown order in `server.go`.
- Revisit telemetry shutdown symmetry later.
- Move next to the subprocess-hardening phase, or to manual validation if we want to observe the new shutdown logs before changing more process behavior.

### Code review instructions
- Review commit `070e6eb`.
- Focus entirely on `internal/web/server.go`.
- Read the shutdown goroutine in order and compare it against the Phase 4 task list.
- Validate with:

```bash
go test ./internal/web -v -count=1 -timeout 45s
go test ./...
```

### Technical details
- Commands run in this step:

```bash
gofmt -w internal/web/server.go
go test ./internal/web -v -count=1 -timeout 45s
go test ./...
git add internal/web/server.go
git commit -m "serve: orchestrate manager shutdown in server"
```

## Step 8: Manually run `serve`, observe a real shutdown, and document the final shutdown order

After the server orchestration refactor, I wanted at least one real serve-mode observation before doing more subprocess hardening. I ran a built binary under `timeout -s INT` so the process would receive a real interrupt. This turned out to be more useful than a truly idle run, because browser auto-open triggered the frontend, which in turn hit the backend endpoints, compiled a setup, started a preview, opened an MJPEG stream, and therefore exercised a realistic in-flight shutdown path without any special test harness.

The most important observation from this run is that the current staged shutdown does work end-to-end in that scenario. When the signal arrived, the runtime canceled telemetry, the preview ffmpeg exited on context cancellation, the MJPEG request eventually completed, HTTP shutdown finished, the manager shutdown hooks ran, and the final summary reported no active recording and no remaining previews. The whole thing still ran under the `timeout` wrapper, so the shell command returned `124`, but the application logs themselves show a clean application shutdown path.

### Prompt Context

**User prompt (verbatim):** (same as Step 4)

**Assistant interpretation:** Keep making disciplined progress, but use real evidence where possible to validate the cancellation design.

**Inferred user intent:** Confirm that the new shutdown orchestration behaves plausibly in a live run before investing more effort in lower-level hardening.

**Commit (code):** pending at the time of this diary entry for the code-comment/doc follow-up step.

### What I did
- Built a fresh binary:

```bash
go build -o /tmp/screencast-studio-test ./cmd/screencast-studio
```

- Ran it under an interrupting timeout:

```bash
timeout -s INT 4s /tmp/screencast-studio-test serve
```

- Inspected the resulting logs rather than relying on the wrapper exit code alone.
- Added a code comment to `internal/web/server.go` documenting the chosen shutdown order explicitly.

### Why
- Manual evidence is useful here because the browser-opening behavior causes the app to exercise more of the real runtime than a synthetic no-op start/stop would.
- The explicit shutdown-order comment closes the remaining Phase 4 documentation task in code.

### What worked
- The live run produced a rich shutdown trace.
- The frontend/browser interaction automatically created a realistic preview path.
- The logs showed the expected staged order at a high level:
  - runtime shutdown began,
  - telemetry exited,
  - preview ffmpeg exited on context cancellation,
  - HTTP shutdown completed,
  - recording and preview manager shutdown hooks ran,
  - the final summary reported no live work.

### What didn't work
- The shell wrapper returned:

```text
Command exited with code 124
```

- That is expected for GNU `timeout` once the timeout threshold has been reached, even when the child process then shuts down cleanly after receiving `SIGINT`.
- So the wrapper exit code is not itself evidence of application failure in this run.

### What I learned
- Browser auto-open materially changes the runtime shape because the frontend eagerly drives backend API calls and preview setup.
- The current shutdown order lets in-flight preview work unwind before the manager shutdown hooks observe an already-drained preview map.
- HTTP shutdown can take a noticeable amount of time when an MJPEG request is in flight; in this run that delay was visible and expected rather than pathological.

### What was tricky to build
- The tricky part here was interpreting the wrapper result correctly. If I had only looked at `timeout`’s exit code, I might have incorrectly concluded the shutdown failed. The log stream told the real story.
- Another subtle point is that this was not actually an “idle” serve run once the browser opened. That ended up being a feature, not a bug, because it gave better evidence about real shutdown behavior.

### What warrants a second pair of eyes
- Whether we want the browser auto-open side effect to remain enabled during future serve integration tests, or whether test/dev modes should be able to suppress it.
- Whether the observed delay waiting for the MJPEG request to unwind should eventually be surfaced more explicitly in shutdown logs.

### What should be done in the future
- Record a small manual-validation recipe in the ticket once more scenarios are exercised.
- Try additional real runs for:
  - active recording + `Ctrl-C`
  - mixed preview + recording + `Ctrl-C`
- Decide later whether browser auto-open should be configurable for testing and debugging.
- Keep telemetry context-driven for now; no separate `TelemetryManager.Shutdown(ctx)` method is needed yet because `Run(ctx)` already exits on runtime cancellation and the server now waits explicitly for the telemetry goroutine.

### Code review instructions
- Review the new shutdown-order comment in `internal/web/server.go`.
- Read the diary observations here alongside the live log output from the manual run.
- Note that the meaningful signal is in the structured application logs, not just the `timeout` wrapper exit code.

### Technical details
- Representative observations from the live run:
  - browser auto-open caused the frontend to request `/api/setup/compile` and `/api/previews/ensure`
  - preview ffmpeg started and later exited on context cancellation
  - telemetry `parec` started and later exited on context cancellation
  - final runtime summary reported `recording_active=false` and `remaining_previews=0`
