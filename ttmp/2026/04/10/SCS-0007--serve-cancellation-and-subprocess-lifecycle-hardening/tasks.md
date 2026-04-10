# Tasks

## Goal

Make `screencast-studio serve` shut down cleanly and predictably on `Ctrl-C`, ensure that serve-owned recording/preview/telemetry subprocesses do not leak, and add enough structured lifecycle logging that future race conditions can be diagnosed from logs instead of guesswork.

## Completed Research And Documentation

- [x] Map current serve/record/preview/telemetry cancellation paths and subprocess ownership.
- [x] Write an intern-oriented design and implementation guide for proper cancellation.
- [x] Write an investigation diary with exact commands, failures, and next steps.
- [x] Relate key code files to the design and diary docs.
- [x] Validate the ticket with `docmgr doctor` and upload the bundle to reMarkable.

## Phase 0: Add Lifecycle Observability Before More Refactors

- [ ] Add structured shutdown logs to `internal/web/server.go` for:
  - signal received
  - HTTP shutdown begin/end
  - recording shutdown begin/end
  - preview shutdown begin/end
  - telemetry shutdown begin/end
  - final runtime shutdown summary
- [ ] Add structured subprocess start logs for every serve-owned subprocess with:
  - component
  - owner/manager
  - session or preview identifier
  - process label
  - pid
  - argv
  - start timestamp
- [ ] Add structured subprocess stop logs for:
  - graceful stop requested
  - stdin `q` written when applicable
  - `SIGTERM` sent
  - `SIGKILL` sent
  - wait begin/end
  - exit result
- [ ] Ensure preview subprocess logging includes preview ID and source ID.
- [ ] Ensure telemetry subprocess logging includes device ID and current telemetry target.
- [ ] Ensure recording subprocess logging includes session ID, output path, and process label.
- [ ] Add a final log record that states whether any managed subprocesses are still believed to be alive when shutdown completes.

Acceptance criteria:

- A developer can reconstruct the full shutdown sequence from logs alone.
- Each started subprocess has a matching stop/wait trail in logs.
- Race-condition reports can be tied to concrete lifecycle events rather than inferred behavior.

## Phase 1: Make Runtime Ownership Explicit

- [ ] Audit every remaining `context.Background()` use in serve-owned runtime paths.
- [x] Pass a server-owned parent context into `RecordingManager`.
- [x] Pass a server-owned parent context into `PreviewManager`.
- [ ] Decide whether `TelemetryManager` also needs explicit constructor-level parent ownership for symmetry.
- [x] Refactor manager constructors and call sites in `internal/web/server.go` accordingly.
- [ ] Document the new runtime ownership tree in comments near the server/manager wiring.

Acceptance criteria:

- Serve-owned long-lived workers are no longer rooted in detached background contexts.
- An intern can trace ownership from `serve` → `server` → `manager` → `subprocess` without ambiguity.

## Phase 2: Add Explicit Manager Shutdown APIs

- [x] Add `Shutdown(ctx)` to `RecordingManager`.
- [x] Implement `RecordingManager.Shutdown(ctx)` so it:
  - detects whether a recording is active
  - logs shutdown intent
  - cancels the active session
  - waits for the session `done` channel
  - respects the shutdown deadline
- [x] Add `Shutdown(ctx)` to `PreviewManager`.
- [x] Implement `PreviewManager.Shutdown(ctx)` so it:
  - snapshots active previews under lock
  - cancels all active preview contexts
  - waits for all preview `done` channels outside the lock
  - aggregates timeout/failure information
- [ ] Decide whether `TelemetryManager` needs a formal `Shutdown(ctx)` method or whether `Run(ctx)` is sufficient once ownership is corrected.
- [x] Ensure manager shutdown paths never block while holding mutable state locks.

Acceptance criteria:

- Each manager has a deterministic shutdown contract.
- Serve shutdown can explicitly drain managers instead of assuming context cancellation is enough.

## Phase 3: Standardize Subprocess Lifecycle Handling

- [ ] Design a shared subprocess lifecycle helper or shared `ManagedProcess` conventions.
- [ ] Capture PID immediately after every successful `cmd.Start()`.
- [ ] Decide whether Unix process-group support should be added now (`Setpgid`) for stronger subtree cleanup.
- [ ] Normalize graceful-stop behavior across subprocess types where appropriate.
- [ ] Keep ffmpeg-friendly graceful shutdown for recordings (stdin `q`) while still supporting escalation.
- [ ] Normalize cancellation-vs-failure result mapping so logs and state transitions do not treat expected shutdown as an unexplained error.
- [ ] Ensure every process is explicitly reaped with `Wait()` and that wait results are logged.

Acceptance criteria:

- Recording, preview, and telemetry subprocesses follow a shared lifecycle model.
- The code has one clear answer to “how do we stop and reap a managed subprocess?”

## Phase 4: Refactor `Server.ListenAndServe` Into A Full Runtime Supervisor

- [x] Refactor `internal/web/server.go` so shutdown is a staged runtime operation rather than only HTTP shutdown.
- [x] Stop accepting new HTTP traffic before draining background managers.
- [x] Trigger manager shutdown explicitly during serve shutdown.
- [x] Wait for telemetry exit as part of the same shutdown sequence.
- [x] Aggregate shutdown errors from HTTP server, recording manager, preview manager, and telemetry manager.
- [ ] Decide on the final shutdown order and encode it clearly in code comments.
- [x] Add a final success/failure summary log before `ListenAndServe` returns.

Acceptance criteria:

- `Ctrl-C` causes a bounded, explainable shutdown sequence.
- `ListenAndServe` does not return until serve-owned background work is either drained or timed out explicitly.

## Phase 5: Strengthen Recording Path Cancellation

- [ ] Review whether `pkg/recording/run.go` should keep plain `exec.Command(...)` or move closer to `CommandContext` semantics plus explicit graceful stop.
- [ ] Add more detailed lifecycle fields to `ManagedProcess` if needed:
  - pid
  - startedAt
  - stopRequestedAt
  - forcedKillAt
  - exitAt
- [ ] Ensure `ManagedProcess.Stop(timeout)` logs each escalation step.
- [ ] Verify that `stopProcesses(...)` continues to stop all processes even when one stop fails.
- [ ] Review whether process killing should target the process group instead of only the direct process on Unix.
- [ ] Add comments explaining why recording uses graceful ffmpeg stop semantics instead of only context cancellation.

Acceptance criteria:

- The recording path is no longer the likely source of silent ffmpeg leaks.
- The stop logic is understandable to a new engineer without reverse-engineering the session event loop.

## Phase 6: Strengthen Preview And Telemetry Cleanup

- [ ] Tie preview contexts to server ownership rather than detached `context.Background()` roots.
- [ ] Add explicit preview shutdown logging for preview creation, cancel, wait, and completion.
- [ ] Verify that preview cleanup removes entries from manager maps even during shutdown races.
- [ ] Review telemetry audio-runner replacement logic for blocking waits during device changes.
- [ ] Add explicit telemetry logs for runner start, cancel, wait, and restart.
- [ ] Confirm that telemetry shutdown cannot deadlock on `runnerDone` waits.

Acceptance criteria:

- Preview and telemetry workers behave consistently during normal stop and server shutdown.
- Shutdown races in these managers are diagnosable from logs.

## Phase 7: Add Focused Tests For Cancellation And Cleanup

- [ ] Add unit tests for `RecordingManager.Shutdown(ctx)`.
- [ ] Add unit tests for `PreviewManager.Shutdown(ctx)`.
- [ ] Add tests for subprocess escalation behavior when graceful shutdown times out.
- [ ] Add tests for log/event behavior around cancellation and final-state mapping.
- [ ] Add a serve-level integration test that starts work, sends cancellation, and asserts bounded shutdown.
- [ ] Add helper subprocess fixtures if needed instead of relying only on real `ffmpeg`/`parec` binaries.
- [ ] Add assertions that no managed child processes remain after the shutdown test completes.

Acceptance criteria:

- The repo contains repeatable tests that prove serve shutdown is bounded and cleans up subprocesses.
- At least one integration test would fail if a child process were leaked.

## Phase 8: Manual Validation And Runbook

- [ ] Write a manual validation recipe in the ticket or a playbook doc.
- [ ] Manually test:
  - idle `serve` + `Ctrl-C`
  - active recording + `Ctrl-C`
  - active preview + `Ctrl-C`
  - active telemetry target + `Ctrl-C`
  - mixed preview + recording + `Ctrl-C`
- [ ] Capture exact logs from a successful clean shutdown.
- [ ] Capture exact logs from at least one intentionally forced kill path.
- [ ] Record any remaining edge cases in the diary and design doc.

Acceptance criteria:

- A reviewer can reproduce the shutdown scenarios manually.
- The ticket contains a human-readable runbook for future regressions.

## Suggested Commit Boundaries

- [x] Commit 1: lifecycle logging only
- [ ] Commit 2: manager parent-context ownership refactor
- [x] Commit 3: recording/preview manager `Shutdown(ctx)` implementations
- [x] Commit 4: server supervisor shutdown sequence refactor
- [ ] Commit 5: subprocess lifecycle helper/process-group hardening
- [ ] Commit 6: tests and manual validation docs

## Completion Definition

- [ ] `screencast-studio serve` exits cleanly on `Ctrl-C` in the documented scenarios.
- [ ] Serve-owned `ffmpeg`/`parec` processes are reaped or explicitly force-killed before exit.
- [ ] Logs clearly show what happened during shutdown.
- [ ] Tests cover the main cancellation paths.
- [ ] The ticket docs and diary are updated to match the implemented behavior.
