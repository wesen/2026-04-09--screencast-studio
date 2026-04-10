# Changelog

## 2026-04-10

- Initial workspace created
- Added design doc for serve cancellation, subprocess shutdown, and lifecycle observability
- Added investigation diary capturing current findings, failed ad hoc attempts, and follow-up direction
- Added implementation-oriented tasks for the cancellation hardening work

## 2026-04-10

Created the ticket workspace, wrote an intern-oriented design doc for serve cancellation/subprocess cleanup, and recorded the investigation diary with concrete file-backed findings.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go — Top-level serve lifecycle examined in the analysis
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go — Recording subprocess lifecycle examined in the analysis
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/10/SCS-0007--serve-cancellation-and-subprocess-lifecycle-hardening/design-doc/01-serve-cancellation-subprocess-shutdown-and-observability-implementation-guide.md — Primary design deliverable for the ticket
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/10/SCS-0007--serve-cancellation-and-subprocess-lifecycle-hardening/reference/01-investigation-diary.md — Chronological diary of findings and failed ad hoc attempts


## 2026-04-10

Validated the ticket with docmgr doctor, added the missing screencast-studio topic vocabulary entry, and uploaded the full design bundle to reMarkable under /ai/2026/04/10/SCS-0007.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/10/SCS-0007--serve-cancellation-and-subprocess-lifecycle-hardening/design-doc/01-serve-cancellation-subprocess-shutdown-and-observability-implementation-guide.md — Primary design doc included in the uploaded bundle
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/10/SCS-0007--serve-cancellation-and-subprocess-lifecycle-hardening/index.md — Ticket index included in the uploaded bundle
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/10/SCS-0007--serve-cancellation-and-subprocess-lifecycle-hardening/reference/01-investigation-diary.md — Diary included in the uploaded bundle
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/vocabulary.yaml — Added screencast-studio topic so doctor passes cleanly


## 2026-04-10

Expanded the ticket task list into a detailed phased implementation plan covering lifecycle logging, manager ownership, explicit shutdown APIs, subprocess handling, tests, manual validation, and completion criteria.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/10/SCS-0007--serve-cancellation-and-subprocess-lifecycle-hardening/tasks.md — Replaced the placeholder checklist with a detailed execution plan


## 2026-04-10

Implemented the first runtime observability pass in commit e2fce73 (serve: add runtime lifecycle logging), adding structured lifecycle logs across server, recording, preview, telemetry, and ffmpeg/parec subprocess paths while intentionally deferring ownership and shutdown API refactors to later phases.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go — Added preview lifecycle logging
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_runner.go — Added preview ffmpeg process lifecycle logging
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go — Added top-level runtime start/shutdown/browser logging
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/session_manager.go — Added recording session lifecycle logging
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/telemetry_manager.go — Added telemetry loop and parec lifecycle logging
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go — Added recording ffmpeg process lifecycle logging
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/10/SCS-0007--serve-cancellation-and-subprocess-lifecycle-hardening/reference/01-investigation-diary.md — Recorded the step


## 2026-04-10

Refactored Phase 1 to use constructor-time parent-context injection for the server, recording manager, and preview manager. During validation this introduced a lock-order self-deadlock in the manager context accessors; isolated  testing exposed it and the fix was to stop taking the manager lock when reading immutable constructor-owned parent contexts.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go — Preview manager now receives its parent context at construction time
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go — NewServer now receives the runtime context and passes it into managers
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server_test.go — Updated tests to construct servers/managers with explicit parent contexts
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/session_manager.go — Recording manager now receives its parent context at construction time
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/serve.go — Serve runtime context is now created before server construction
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/10/SCS-0007--serve-cancellation-and-subprocess-lifecycle-hardening/reference/01-investigation-diary.md — Recorded the deadlock


## 2026-04-10

Added explicit `Shutdown(ctx)` APIs to the recording and preview managers in commit `34493fb`, with focused success/timeout tests in `internal/web/manager_shutdown_test.go`. This phase keeps shutdown contracts local to the managers and deliberately defers the top-level server orchestration change to a later step.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/manager_shutdown_test.go — Added focused manager shutdown tests covering success and timeout paths
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go — Added PreviewManager.Shutdown(ctx) with per-preview cancellation and timeout reporting
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/session_manager.go — Added RecordingManager.Shutdown(ctx) with bounded wait semantics
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/10/SCS-0007--serve-cancellation-and-subprocess-lifecycle-hardening/reference/01-investigation-diary.md — Recorded the implementation and validation details for Phase 2


## 2026-04-10

Refactored `ListenAndServe` into a staged runtime shutdown in commit `070e6eb`, so serve now stops HTTP intake, explicitly drains recording and preview managers, waits for HTTP/telemetry goroutines to exit, and logs a final runtime component summary before returning.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go — Orchestrates staged runtime shutdown and waits for manager/goroutine completion
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/10/SCS-0007--serve-cancellation-and-subprocess-lifecycle-hardening/reference/01-investigation-diary.md — Recorded the server orchestration phase and validation details
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/10/SCS-0007--serve-cancellation-and-subprocess-lifecycle-hardening/tasks.md — Marked the completed Phase 4 shutdown orchestration items

