---
Title: Serve cancellation, subprocess shutdown, and observability implementation guide
Ticket: SCS-0007
Status: active
Topics:
    - screencast-studio
    - backend
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/handlers_api.go
      Note: HTTP entrypoints into compile
    - Path: internal/web/preview_manager.go
      Note: Preview lifecycle and detached preview contexts
    - Path: internal/web/preview_runner.go
      Note: Preview ffmpeg subprocess execution
    - Path: internal/web/server.go
      Note: Top-level server lifecycle
    - Path: internal/web/session_manager.go
      Note: Recording manager start/stop and current detached background context usage
    - Path: internal/web/telemetry_manager.go
      Note: Telemetry loop and parec subprocess lifecycle
    - Path: pkg/app/application.go
      Note: Boundary between transport layer and recording runtime
    - Path: pkg/cli/serve.go
      Note: Serve command signal hookup and server construction
    - Path: pkg/recording/run.go
      Note: Recording subprocess lifecycle
    - Path: pkg/recording/session.go
      Note: Recording state machine and transition rules
ExternalSources: []
Summary: Evidence-based design guide for fixing Ctrl-C/serve shutdown, eliminating orphaned ffmpeg/parec processes, and adding structured lifecycle logging across the screencast-studio runtime.
LastUpdated: 2026-04-10T09:31:48.0843538-04:00
WhatFor: Orient a new engineer, explain the current runtime and cancellation model, and provide a concrete implementation plan for robust shutdown and observability.
WhenToUse: Use when implementing or reviewing serve shutdown, recording lifecycle cancellation, preview cleanup, telemetry subprocess management, or subprocess logging.
---


# Serve cancellation, subprocess shutdown, and observability implementation guide

## Executive Summary

This ticket exists because the current `serve` runtime does not yet have a single, explicit ownership model for everything it starts. The CLI receives `Ctrl-C`, but the work started behind the HTTP server is split across several independent subsystems: the HTTP listener, the telemetry loops, preview workers, and recording workers. Some of those workers are tied to the server context, but others are created from `context.Background()` and therefore outlive the serving command unless they are stopped through a separate code path. That is the most important architectural fact for a new engineer to understand.

The practical symptom reported by the user is that interrupting the app can leave `ffmpeg` running in the background. The code supports some graceful-stop behavior already, especially in the recording runner, but the shutdown contract is not centralized. There is no single component that can answer the question “what subprocesses did this server start, and have all of them exited yet?” The current code also lacks sufficiently rich structured logs to reconstruct cancellation races after the fact.

The recommended direction is to treat runtime cancellation as a first-class subsystem rather than an incidental feature. The server should own all background managers. Each manager should expose a `Shutdown(ctx)` method. Each subprocess launch should flow through a shared helper that logs start/stop/kill/wait events with PIDs, labels, deadlines, and exit outcomes. On `Ctrl-C`, the system should stop accepting new work, cancel all background managers, wait for their subprocesses to exit, and then escalate from graceful stop to force-kill only when the deadline expires.

## Problem Statement and Scope

The user asked to step back from ad hoc cancellation changes and instead create a proper ticket with a detailed design and implementation guide. The immediate problem is not just “make `Ctrl-C` work.” The real problem is broader:

1. The runtime has multiple cancellation domains.
2. Some long-lived workers are detached from the `serve` command.
3. The runtime has incomplete visibility into subprocess lifecycle events.
4. Existing tests cover API behavior more than real shutdown behavior.

This document covers:

- how `serve` is wired today,
- how recording, preview, and telemetry workers are started,
- where ownership and cancellation boundaries are currently inconsistent,
- why orphaned `ffmpeg` is plausible in the current design,
- a recommended target architecture,
- a detailed implementation plan with file references,
- a logging plan intended to make race conditions diagnosable.

This document does **not** propose a large functional rewrite of discovery, DSL normalization, or the frontend. It is focused on runtime lifecycle management and observability for the existing local control server.

## Reader Orientation: What This System Is

Before changing shutdown behavior, a new intern needs a mental model of the product.

### High-level purpose

`screencast-studio` is a CLI-first application that can:

- inspect capture sources,
- parse/normalize a DSL that describes a recording setup,
- compile that setup into concrete jobs,
- run recordings via `ffmpeg`,
- serve a local HTTP API + frontend for controlling recordings and previews.

### Major layers

1. **CLI layer**
   - entrypoint: `cmd/screencast-studio/main.go`
   - root command wiring: `pkg/cli/root.go`
   - top-level user commands include `record` and `serve`

2. **Application layer**
   - `pkg/app/application.go`
   - translates HTTP/CLI requests into discovery, DSL normalization, plan compilation, and recording execution

3. **Web server layer**
   - `internal/web/server.go`
   - registers HTTP routes and owns web-oriented runtime managers

4. **Runtime managers**
   - recording state: `internal/web/session_manager.go`
   - preview state: `internal/web/preview_manager.go`
   - telemetry state: `internal/web/telemetry_manager.go`

5. **Subprocess execution layer**
   - recording processes: `pkg/recording/run.go`
   - preview processes: `internal/web/preview_runner.go`
   - telemetry audio meter subprocess: `internal/web/telemetry_manager.go`

### Current control surface

The HTTP server registers the API routes in `internal/web/routes.go:9-24`, including:

- `/api/setup/compile`
- `/api/recordings/start`
- `/api/recordings/stop`
- `/api/previews/ensure`
- `/api/previews/release`
- `/ws`

That means the `serve` process is not just an HTTP listener. It is effectively a runtime host for multiple independent worker types.

## Current-State Architecture (Evidence-Based)

### 1. CLI entrypoint and signal setup

The `serve` command decodes CLI flags, constructs a `web.Server`, then wraps the command context with `signal.NotifyContext` for `SIGINT`/`SIGTERM` before calling `ListenAndServe`.

Evidence:

- `pkg/cli/serve.go:54-70`
- default bind address flag: `pkg/cli/serve.go:42-46`
- `signal.NotifyContext`: `pkg/cli/serve.go:67-70`

Relevant snippet:

```go
serverCtx, stopSignals := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
defer stopSignals()
return server.ListenAndServe(serverCtx)
```

This means `serve` does receive a cancellable context. The problem is what the rest of the system does with it.

### 2. Server composition

`NewServer` constructs three manager objects and stores them inside the web server object.

Evidence:

- server fields: `internal/web/server.go:25-33`
- manager construction: `internal/web/server.go:46-57`

That is important because it means the web server is already the natural place for top-level runtime ownership. The current architecture is close to having a supervisor, but it is not fully acting like one yet.

### 3. Current serve lifecycle wiring

`ListenAndServe` currently creates a child context, starts the HTTP server in one errgroup goroutine, starts telemetry in another, and starts a third goroutine that waits for the parent context to end and then calls `httpServer.Shutdown(...)`.

Evidence:

- `internal/web/server.go:64-123`
- child server context: `internal/web/server.go:71-75`
- HTTP server goroutine: `internal/web/server.go:77-93`
- telemetry goroutine: `internal/web/server.go:95-98`
- shutdown goroutine: `internal/web/server.go:100-116`

Observed properties:

- The HTTP server itself is cancellable.
- Telemetry is at least partially tied to the server context.
- There is **no explicit shutdown of `recordings` or `previews`** from this method.

That last point is the central architectural gap.

### 4. Application layer responsibilities

The application layer is intentionally thin. It exposes discovery, DSL normalization, DSL compilation, and recording execution.

Evidence:

- `pkg/app/application.go:15-217`
- `RecordPlan` delegates to `recording.Run(...)`: `pkg/app/application.go:168-191`

This layer is not where runtime ownership should live. It is more of a service facade.

### 5. Recording flow inside the web server

The recording API handler calls `s.recordings.Start(...)` and `s.recordings.Stop()`.

Evidence:

- start handler: `internal/web/handlers_api.go:74-103`
- stop handler: `internal/web/handlers_api.go:106-113`

The recording manager compiles the plan, then starts a goroutine that calls `m.app.RecordPlan(...)`.

Evidence:

- `internal/web/session_manager.go:72-119`

However, the recording manager creates its runtime context using `context.WithCancel(context.Background())` rather than a server-owned parent context.

Evidence:

- `internal/web/session_manager.go:85-88`

That means:

- the recording session is stoppable through its own manager,
- but it is **not** naturally cancelled when the `serve` process context ends,
- unless some other code explicitly calls `recordings.Stop()` during shutdown.

At the moment, `ListenAndServe` does not do that.

### 6. Recording subprocess model

The low-level recording engine in `pkg/recording/run.go` creates one `ManagedProcess` per video/audio job, then drives a session state machine.

Evidence:

- session creation and process registration: `pkg/recording/run.go:69-110`
- event loop: `pkg/recording/run.go:112-287`
- stop helper: `pkg/recording/run.go:289-302`
- `ManagedProcess` implementation: `pkg/recording/run.go:36-47`, `313-430`
- session states: `pkg/recording/session.go:9-27`

Important behavior details:

- `ManagedProcess.Run(...)` launches `ffmpeg` with `exec.Command("ffmpeg", p.Args...)` in `pkg/recording/run.go:323`.
- That is **not** `exec.CommandContext`.
- Cancellation is handled indirectly through the session event loop and `ManagedProcess.Stop(...)`.
- `Stop(...)` attempts graceful termination by writing `q\n` to ffmpeg stdin, then waits for `done`, and only after a timeout does it call `Process.Kill()`.
  - evidence: `pkg/recording/run.go:404-430`

This design can work, but it depends on several assumptions:

1. `ffmpeg` is still responsive to stdin.
2. The process is still the one holding the relevant resources.
3. Killing the direct process is sufficient to kill any child processes.
4. The session event loop actually reaches `stopProcesses(...)` before the parent runtime exits.

Those assumptions may be invalid during abrupt shutdown races.

### 7. Preview flow

The preview manager normalizes DSL, identifies a source, then creates a new preview context with `context.WithCancel(context.Background())`.

Evidence:

- preview start: `internal/web/preview_manager.go:79-143`
- detached context creation: `internal/web/preview_manager.go:104-110`
- release path: `internal/web/preview_manager.go:146-166`

The underlying preview runner launches `ffmpeg` with `exec.CommandContext(ctx, "ffmpeg", args...)`.

Evidence:

- `internal/web/preview_runner.go:27-44`

This is better than the recording path because the OS process is tied to the preview context. However, the preview context is **not** tied to the server context; it is rooted in `context.Background()`. That means a preview can remain alive unless someone explicitly calls `preview.cancel()` through `Release(...)` or another future shutdown hook.

### 8. Telemetry flow

The telemetry manager is the only major runtime manager already wired into `Server.ListenAndServe(...)`.

Evidence:

- telemetry goroutine: `internal/web/server.go:95-98`
- manager run loop: `internal/web/telemetry_manager.go:74-79`

The audio meter path starts `parec` with `exec.CommandContext(ctx, "parec", ...)` and explicitly waits for the process in a goroutine.

Evidence:

- `internal/web/telemetry_manager.go:235-268`

The current implementation also waits for runner shutdown when device selection changes or when the loop exits.

Evidence:

- defer shutdown: `internal/web/telemetry_manager.go:176-183`
- runner refresh replacement: `internal/web/telemetry_manager.go:188-221`

Telemetry is therefore the closest thing to the target model, but it still needs stronger logging and a clearer contract for “what happens if shutdown is racing with a blocked `stdout.Read(...)` or `cmd.Wait()`?”

### 9. Existing observability

Current logs exist, but they are not sufficient for race diagnosis.

Observed logging today:

- web server startup and HTTP requests: `internal/web/server.go:79-83`, `157-166`
- recording stop requested: `internal/web/session_manager.go:131-133`
- recording subprocess stdout/stderr lines: `pkg/recording/run.go:321`, `361-386`

Missing or incomplete logging today:

- no guaranteed PID logging for spawned processes,
- no parent-child ownership logging,
- no explicit “shutdown phase start/end” logs for every manager,
- no log entries for signal sent vs force kill vs wait completion,
- no single summary at server exit that says all subprocesses are gone.

### 10. Test coverage shape

The existing tests are helpful for API behavior, but there is no evidence of a dedicated `ListenAndServe` cancellation integration test with real subprocess lifecycle assertions.

Evidence:

- route and handler-oriented tests in `internal/web/server_test.go`
- search results show stop-handler tests but not serve-process interrupt tests

This gap matters because cancellation bugs usually only appear when multiple goroutines and OS processes are involved.

## Why Orphaned `ffmpeg` Is Plausible in the Current Design

A new engineer should understand that the orphan problem does not require a single obvious bug. It can emerge from the combination of several individually reasonable choices.

### Primary causes

#### A. Detached manager contexts

Both recording and preview sessions are currently rooted in `context.Background()` instead of a server-owned parent.

Evidence:

- recording: `internal/web/session_manager.go:85-88`
- preview: `internal/web/preview_manager.go:104-110`

Implication:

- when `serve` stops, those managers do not automatically stop their work.

#### B. Server shutdown does not explicitly drain all managers

`ListenAndServe` currently shuts down the HTTP server and telemetry, but it does not call anything like:

- `recordings.Shutdown(...)`
- `previews.Shutdown(...)`

Evidence:

- `internal/web/server.go:64-123`

Implication:

- the top-level process can exit its HTTP/telemetry paths without first proving all preview and recording workers are dead.

#### C. Recording subprocesses are not created with `CommandContext`

`ManagedProcess.Run(...)` launches `ffmpeg` using plain `exec.Command(...)`.

Evidence:

- `pkg/recording/run.go:323`

Implication:

- context cancellation does not automatically propagate to the OS process.
- cleanup relies on higher-level stop orchestration reaching `ManagedProcess.Stop(...)`.

#### D. Stop semantics are asymmetric across subsystems

- recording uses explicit stdin + timeout + kill
- preview uses `CommandContext`
- telemetry uses `CommandContext` plus explicit wait/kill handling

Implication:

- shutdown reasoning is harder because each subsystem follows a different model.

#### E. No unified subprocess registry

There is no shared in-memory registry that can answer:

- what is currently running,
- which manager owns it,
- what PID/PGID it has,
- whether it has acknowledged graceful stop,
- whether it needed force-kill,
- whether it has fully reaped.

Without that registry, diagnosing race conditions depends on scattered logs and user reports.

## Gap Analysis

### Gap 1: Ownership tree is missing

**Observed:** server owns managers structurally, but not lifecycle-wise.

**Desired:** server should be the root owner of all background work started while serving.

### Gap 2: Shutdown API is incomplete

**Observed:** managers expose operational methods (`Start`, `Stop`, `Ensure`, `Release`) but not top-level `Shutdown(ctx)` methods.

**Desired:** every manager that can own background work should implement a deterministic shutdown contract.

### Gap 3: Subprocess launch policy is inconsistent

**Observed:** recording, preview, and telemetry launch processes differently.

**Desired:** a shared process wrapper should enforce common logging, wait semantics, and escalation strategy.

### Gap 4: Logging is not rich enough for concurrency debugging

**Observed:** we can see some app logs and process stderr, but not a trustworthy lifecycle trace.

**Desired:** every start/stop/cancel/kill/wait event should be visible in structured logs.

### Gap 5: Tests do not prove no child processes are leaked

**Observed:** current tests mostly validate handlers and in-memory transitions.

**Desired:** add integration tests that launch real helper processes and assert shutdown completion and process reaping.

## Proposed Solution

The solution has four main parts:

1. introduce explicit runtime ownership,
2. give every manager a shutdown contract,
3. standardize subprocess lifecycle handling,
4. add structured logging and integration tests.

### Proposed runtime ownership model

The server should become the definitive runtime supervisor for serve-mode work.

#### Proposed responsibilities of `web.Server`

`Server` should:

- own a root runtime context for serve mode,
- own manager instances as it does today,
- expose a `Shutdown(ctx)` helper or equivalent internal sequence,
- stop accepting new requests before cancelling workers,
- wait for manager shutdown completion before returning from `ListenAndServe`.

#### Proposed responsibilities of managers

- `RecordingManager`
  - owns zero or one live recording session
  - must expose `Shutdown(ctx)`
  - must cancel current recording and wait for completion

- `PreviewManager`
  - owns zero or more live previews
  - must expose `Shutdown(ctx)`
  - must cancel all preview contexts and wait for all `done` channels

- `TelemetryManager`
  - should retain its current `Run(ctx)` model
  - may optionally expose a no-op `Shutdown(ctx)` if symmetry helps

### Proposed context hierarchy

Current code uses `context.Background()` in key places. Replace that with parent contexts owned by the server.

#### Target model

```text
serve CLI signal context
└── server runtime context
    ├── http server lifecycle
    ├── telemetry manager run loop
    ├── recording manager parent context
    │   └── recording session context
    │       └── ffmpeg managed processes
    └── preview manager parent context
        └── preview context(s)
            └── preview ffmpeg process(es)
```

### Proposed API changes

#### `internal/web/server.go`

Add an explicit orderly shutdown flow.

Pseudo-API:

```go
type shutdownable interface {
    Shutdown(ctx context.Context) error
}
```

Potential internal helper:

```go
func (s *Server) shutdownRuntime(ctx context.Context) error
```

Responsibilities:

1. log shutdown phase start,
2. stop accepting new HTTP requests,
3. trigger manager shutdown in defined order,
4. wait for managers,
5. log a shutdown summary.

#### `internal/web/session_manager.go`

Recommended additions:

```go
type RecordingManager struct {
    parentCtx context.Context
    ...
}

func NewRecordingManager(parentCtx context.Context, application ApplicationService, publish func(ServerEvent)) *RecordingManager
func (m *RecordingManager) Shutdown(ctx context.Context) error
```

Implementation notes:

- recording sessions should derive from `m.parentCtx`, not `context.Background()`.
- `Shutdown(ctx)` should:
  - detect active recording,
  - call `cancel()`,
  - wait on `done`,
  - respect `ctx.Done()`.

#### `internal/web/preview_manager.go`

Recommended additions:

```go
type PreviewManager struct {
    parentCtx context.Context
    ...
}

func NewPreviewManager(parentCtx context.Context, application ApplicationService, publish func(ServerEvent), limit int, runner PreviewRunner) *PreviewManager
func (m *PreviewManager) Shutdown(ctx context.Context) error
```

Implementation notes:

- preview contexts should derive from `m.parentCtx`, not `context.Background()`.
- `Shutdown(ctx)` should:
  - snapshot all active previews,
  - cancel them all,
  - wait for each `done` channel,
  - aggregate errors/timeouts.

#### `pkg/recording/run.go`

Introduce a subprocess helper rather than open-coding every launch pattern.

A practical API sketch:

```go
type ProcessHandle interface {
    PID() int
    Label() string
    Wait() error
    Stop(ctx context.Context) error
}

type ProcessRunner struct {
    Logger LifecycleLogger
}

func (r *ProcessRunner) StartFFmpeg(ctx context.Context, label string, args []string, opts ProcessOptions) (*ManagedProcess, error)
```

This does not need to be a huge abstraction. The point is to centralize:

- PID/PGID capture,
- structured start logging,
- graceful stop vs force kill,
- `Wait()` logging,
- exit status mapping.

### Proposed subprocess logging model

This is the part the user explicitly asked for: “copious logging.”

Every managed subprocess event should log at least:

- `component` (`recording`, `preview`, `telemetry`, `server`)
- `manager` or `owner`
- `session_id` and/or `preview_id`
- `process_label`
- `pid`
- `pgid` when available
- `argv`
- `shutdown_reason`
- `signal` (if one is sent)
- `grace_timeout`
- `exit_code`
- `wait_error`
- `elapsed_ms`

#### Recommended log events

1. `runtime.shutdown.begin`
2. `runtime.shutdown.http.begin`
3. `runtime.shutdown.http.done`
4. `runtime.shutdown.recordings.begin`
5. `runtime.shutdown.recordings.done`
6. `runtime.shutdown.previews.begin`
7. `runtime.shutdown.previews.done`
8. `runtime.shutdown.telemetry.done`
9. `process.start`
10. `process.stop.requested`
11. `process.stop.stdin_q_sent`
12. `process.stop.sigterm_sent`
13. `process.stop.kill_sent`
14. `process.wait.begin`
15. `process.wait.done`
16. `runtime.shutdown.summary`

#### Example structured log payload

```json
{
  "event": "process.stop.requested",
  "component": "recording",
  "session_id": "session-123",
  "process_label": "Display 1",
  "pid": 412233,
  "pgid": 412233,
  "grace_timeout_ms": 5000,
  "reason": "server shutdown after SIGINT"
}
```

### Proposed shutdown order

The order matters. A clear sequence reduces race conditions.

#### Recommended order

1. receive `SIGINT` / `SIGTERM`
2. log `runtime.shutdown.begin`
3. stop new HTTP traffic (`httpServer.Shutdown`)
4. cancel server root context
5. call `recordings.Shutdown(ctx)`
6. call `previews.Shutdown(ctx)`
7. wait for telemetry loop to exit
8. verify subprocess registry is empty
9. log shutdown summary
10. return from `ListenAndServe`

Why this order?

- Step 3 prevents new API requests from creating new work while we are shutting down.
- Step 4 informs background loops they must stop.
- Steps 5 and 6 explicitly drain detached-style work.
- Step 8 creates a concrete postcondition.

### Process-group recommendation

If the environment is Linux/macOS, consider using process groups so a force-kill can kill the entire subtree, not just the direct `ffmpeg` process.

Pseudo-approach for Unix-like systems:

```go
cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
```

Then on escalation:

```go
syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
```

This is especially valuable if `ffmpeg` indirectly spawns helpers or if the shell/process tree is more complex in future integrations.

This should be introduced carefully and tested explicitly, but it is a strong candidate for hardening.

## Diagrams

### Current serve-mode ownership (observed)

```text
CLI serve command
└── signal.NotifyContext
    └── Server.ListenAndServe(ctx)
        ├── HTTP listener (owned by server ctx)
        ├── TelemetryManager.Run(groupCtx)
        │   └── parec process (mostly tied to ctx)
        ├── RecordingManager (constructed by server)
        │   └── Start() uses context.Background()
        │       └── recording.Run(...)
        │           └── ffmpeg process(es)
        └── PreviewManager (constructed by server)
            └── Ensure() uses context.Background()
                └── exec.CommandContext(ctx, "ffmpeg", ...)
```

The visual issue is that recording and preview workers are created from detached roots even though they conceptually belong to `serve`.

### Target ownership model

```text
SIGINT/SIGTERM
└── serve root context cancelled
    └── Server runtime supervisor
        ├── stop HTTP accept loop
        ├── cancel manager parent contexts
        ├── RecordingManager.Shutdown()
        │   └── cancel + wait recording session(s)
        ├── PreviewManager.Shutdown()
        │   └── cancel + wait preview session(s)
        ├── TelemetryManager exits
        └── subprocess registry empty => return cleanly
```

### Recording stop sequence (target)

```text
server shutdown begins
→ recording manager sees active session
→ logs process.stop.requested for each ffmpeg worker
→ writes "q\n" to stdin (graceful)
→ waits until deadline
→ if still alive: SIGTERM / SIGKILL escalation
→ logs wait result and exit code
→ session transitions stopping → finished/failed
→ manager reports shutdown done
```

## Detailed Implementation Plan

### Phase 0: Do not ship more ad hoc fixes before instrumentation

Before changing more cancellation behavior, add the logging needed to understand the race.

Why:

- The system already has partially modified cancellation code.
- Without lifecycle logs, it is too easy to “fix” one path while masking another orphan source.

Deliverables:

- structured logs for subprocess start/stop/wait in recording, preview, and telemetry,
- shutdown-phase logs in `Server.ListenAndServe`,
- PIDs in logs,
- a final shutdown summary log.

Primary files:

- `internal/web/server.go`
- `internal/web/session_manager.go`
- `internal/web/preview_manager.go`
- `internal/web/preview_runner.go`
- `internal/web/telemetry_manager.go`
- `pkg/recording/run.go`

### Phase 1: Make manager ownership explicit

Refactor constructors so managers receive a parent context owned by the server.

Suggested changes:

- change `NewRecordingManager(...)`
- change `NewPreviewManager(...)`
- optionally pass parent context into telemetry manager if symmetry helps
- update `NewServer(...)` accordingly

Primary objective:

- remove `context.Background()` from serve-owned long-lived tasks.

Evidence motivating this phase:

- detached recording context: `internal/web/session_manager.go:85-88`
- detached preview context: `internal/web/preview_manager.go:104-110`

### Phase 2: Add `Shutdown(ctx)` to each manager

#### Recording manager

Behavior:

- if no active recording, return quickly
- if active recording exists:
  - log shutdown begin
  - call session cancel
  - wait on `done`
  - return timeout error if the passed shutdown context expires first

#### Preview manager

Behavior:

- gather active preview IDs under lock
- cancel every preview
- wait for each `done`
- aggregate timeout/failure info

Pseudo-code:

```go
func (m *PreviewManager) Shutdown(ctx context.Context) error {
    previews := snapshotActivePreviews()
    for _, p := range previews {
        logCancel(p)
        p.cancel()
    }
    for _, p := range previews {
        select {
        case <-p.done:
            logDone(p)
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    return nil
}
```

### Phase 3: Unify subprocess lifecycle handling

Refactor subprocess launches around a shared pattern.

At minimum, standardize these elements:

- capture PID immediately after `cmd.Start()`
- log `process.start`
- expose a single helper for graceful stop + escalation
- expose a single helper for wait logging
- normalize errors returned on cancellation vs real failures

The recording path is the highest priority because it currently uses `exec.Command(...)` and relies on higher-level cleanup.

Potential minimal refactor:

- keep `ManagedProcess`
- add explicit fields for `PID`, `startedAt`, `stopRequestedAt`
- add structured logging helpers
- add optional Unix process-group support

### Phase 4: Make `ListenAndServe` perform a full runtime shutdown

Refactor `internal/web/server.go` so shutdown is explicit and staged.

Pseudo-code:

```go
func (s *Server) ListenAndServe(ctx context.Context) error {
    runtimeCtx, cancel := context.WithCancel(ctx)
    defer cancel()

    startHTTP()
    startTelemetry(runtimeCtx)

    <-ctx.Done()
    log("runtime.shutdown.begin")

    httpServer.Shutdown(deadlineCtx)
    cancel()

    if err := s.recordings.Shutdown(deadlineCtx); err != nil { ... }
    if err := s.previews.Shutdown(deadlineCtx); err != nil { ... }

    waitForTelemetry()
    logSummary()
    return nil
}
```

Implementation detail:

- the HTTP listener should stop new traffic before managers are drained,
- but the manager shutdown should not be skipped even if `httpServer.Shutdown(...)` returns quickly.

### Phase 5: Add tests that prove shutdown correctness

#### Unit tests

Add focused tests for:

- `RecordingManager.Shutdown(ctx)` with fake `RecordPlan` worker
- `PreviewManager.Shutdown(ctx)` with fake preview runner
- process stop escalation helper behavior

#### Integration tests

Add at least one serve-mode integration test that:

1. launches the real binary or a close equivalent,
2. starts a recording or preview using helper subprocesses,
3. sends `SIGINT`,
4. asserts the server exits within deadline,
5. asserts helper processes are gone.

Possible helper strategy:

- create test helper commands that emulate `ffmpeg`-like behavior:
  - ignore stdin for a while,
  - optionally trap `SIGTERM`,
  - print PIDs to stdout/stderr,
  - exit only after receiving a signal.

This is more deterministic than relying on the real `ffmpeg` binary in every test.

#### Logging assertions

Tests should also assert key lifecycle logs exist. The goal is not just “it exited,” but “the system explained what it did.”

## API and File Reference Guide for a New Intern

This section is intentionally redundant. It exists so a new engineer can orient quickly without re-deriving the system map.

### Start reading here

1. `pkg/cli/root.go`
   - shows how commands are registered
2. `pkg/cli/serve.go`
   - shows signal hookup and server construction
3. `internal/web/server.go`
   - shows serve lifecycle and manager ownership
4. `internal/web/routes.go`
   - shows API surface
5. `internal/web/handlers_api.go`
   - shows which handlers invoke recording and compilation flows
6. `internal/web/session_manager.go`
   - shows recording runtime state under the server
7. `pkg/recording/run.go`
   - shows actual ffmpeg orchestration
8. `internal/web/preview_manager.go`
   - shows preview ownership and release model
9. `internal/web/preview_runner.go`
   - shows preview ffmpeg execution
10. `internal/web/telemetry_manager.go`
   - shows `parec` execution and looping behavior
11. `pkg/recording/session.go`
   - shows legal state transitions
12. `pkg/app/application.go`
   - shows the boundary between transport layer and runtime layer

### Key symbols to know

- `serveCommand.RunIntoGlazeProcessor`
- `web.NewServer`
- `(*Server).ListenAndServe`
- `(*RecordingManager).Start`
- `(*RecordingManager).Stop`
- `recording.Run`
- `(*ManagedProcess).Run`
- `(*ManagedProcess).Stop`
- `(*PreviewManager).Ensure`
- `FFmpegPreviewRunner.Run`
- `(*TelemetryManager).Run`
- `(*TelemetryManager).streamAudioMeter`

## Risks, Alternatives, and Tradeoffs

### Risk 1: Adding shutdown hooks can deadlock if locks are held while waiting

Mitigation:

- take state snapshots under lock,
- release the lock before blocking on `done` channels,
- keep shutdown waits outside critical sections.

### Risk 2: Overusing `CommandContext` without explicit graceful stop may truncate outputs

Tradeoff:

- pure context cancellation is simple,
- but ffmpeg often benefits from receiving `q` and flushing outputs.

Recommendation:

- keep graceful ffmpeg stop semantics,
- but wrap them in a better supervisor with escalation and logging.

### Risk 3: Process-group killing is platform-sensitive

Tradeoff:

- process groups improve reliability on Unix-like systems,
- but they need platform-conditional handling.

Recommendation:

- design the abstraction now,
- implement Unix support first if needed,
- keep Windows-compatible fallbacks explicit.

### Alternative considered: only add more logging and leave ownership model alone

Rejected because:

- logging alone does not prevent leaks,
- detached manager contexts remain a correctness problem.

### Alternative considered: only tie everything to `CommandContext`

Rejected because:

- recording already encodes a graceful stdin-based stop policy that is useful,
- the real issue is not only process launch but also manager ownership and wait discipline.

## Open Questions

1. Should recording continue to use stdin `q` as the primary graceful-stop path, or should `SIGTERM` become first-class as well?
2. Does `ffmpeg` ever create child processes in the current deployment environment, making process-group kill necessary immediately?
3. Should serve shutdown block until all preview MJPEG clients disconnect, or should it hard-cut those streams once previews are cancelled?
4. Do we want a reusable `internal/process` or `pkg/process` package now, or should we first refactor locally and extract later?
5. Should the subprocess registry also be exposed via `/api/healthz` or a debug endpoint during development?

## Implementation Checklist

- [ ] Add structured shutdown logs to `internal/web/server.go`
- [ ] Add structured subprocess lifecycle logs to recording, preview, telemetry paths
- [ ] Pass server-owned parent contexts into `RecordingManager` and `PreviewManager`
- [ ] Remove serve-owned `context.Background()` roots from recording and preview paths
- [ ] Add `Shutdown(ctx)` to `RecordingManager`
- [ ] Add `Shutdown(ctx)` to `PreviewManager`
- [ ] Wire manager shutdown into `Server.ListenAndServe`
- [ ] Add unit tests for manager shutdown behavior
- [ ] Add integration test for `SIGINT` with subprocess cleanup assertions
- [ ] Add final runtime shutdown summary log

## References

### Primary code references

- `pkg/cli/serve.go:32-70` — serve command construction and signal context hookup
- `internal/web/server.go:25-123` — server composition and serve lifecycle
- `internal/web/routes.go:9-24` — route registration
- `internal/web/handlers_api.go:53-112` — compile, start recording, stop recording handlers
- `internal/web/session_manager.go:72-133` — recording start/stop manager API
- `internal/web/preview_manager.go:79-166` — preview ensure/release lifecycle
- `internal/web/preview_runner.go:27-76` — preview ffmpeg lifecycle
- `internal/web/telemetry_manager.go:74-303` — telemetry loops and `parec` lifecycle
- `pkg/app/application.go:168-191` — app-level delegation into recording runtime
- `pkg/recording/run.go:49-430` — recording session event loop and `ManagedProcess`
- `pkg/recording/session.go:9-87` — session states and transitions
- `pkg/recording/events.go:5-28` — run event contract

### Secondary references

- `internal/web/server_test.go` — existing API-focused web tests
- `pkg/recording/session_test.go` — session transition coverage

### Ticket companion doc

- `reference/01-investigation-diary.md` — chronological diary for this investigation
