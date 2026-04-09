# Changelog

## 2026-04-09

- Initial workspace created


## 2026-04-09

Created the screencast studio architecture ticket, imported the JSX control-surface mock, and wrote the detailed system design and diary.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/design-doc/01-screencast-studio-system-design.md — Primary design deliverable
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/reference/01-diary.md — Chronological documentation record
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/sources/local/screencast-studio-v2.jsx.jsx — Imported local UI reference artifact


## 2026-04-09

Validated the ticket with docmgr doctor, dry-ran the reMarkable bundle upload, uploaded the final PDF bundle, and verified the remote listing.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/tasks.md — Task checklist now reflects completed validation and delivery


## 2026-04-09

Re-scoped the ticket to a CLI-first milestone centered on discover, compile, and record; deferred the web frontend to a follow-up ticket.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/design-doc/01-screencast-studio-system-design.md — Implementation order and command surface updated for the CLI milestone
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/reference/01-diary.md — Diary now records the scope shift and execution order
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/tasks.md — Detailed CLI-first milestone checklist added


## 2026-04-09

Added the root Go module and CLI skeleton with discovery, setup compile, setup validate, and record commands (commit 047d61c).

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/go.mod — Top-level module established for the new implementation
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/record.go — Record verb introduced in the CLI surface
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/root.go — Root command tree introduced


## 2026-04-09

Implemented real platform discovery for displays, windows, cameras, and audio inputs behind the discovery CLI (commit cd94620).

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/discovery/service.go — Command-backed discovery implementation
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/discovery/types.go — Discovery descriptor definitions


## 2026-04-09

Extracted the setup DSL into `pkg/dsl` and replaced the compile stub with a real planning pipeline that emits concrete output rows from a setup file.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/normalize.go — DSL normalization and validation rules
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/compile.go — Compiled output manifest generation
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/setup/compile.go — Structured CLI rendering of compiled outputs


## 2026-04-09

Introduced a formal recording session state machine, added FFmpeg stdout/stderr capture, and stored ticket-local smoke repro scripts for the recording runtime investigation.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/session.go — Explicit session states and transition helpers
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go — Event-driven runtime coordination around the session state machine
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/smoke-record-region.sh — Ticket-local CLI smoke repro
