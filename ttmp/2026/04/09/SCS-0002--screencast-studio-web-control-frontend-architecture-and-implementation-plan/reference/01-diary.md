---
Title: Diary
Ticket: SCS-0002
Status: active
Topics:
    - backend
    - frontend
    - video
    - audio
    - dsl
    - cli
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/screencast-studio/main.go
      Note: Existing binary entrypoint inspected for future server integration
    - Path: jank-prototype/main.go
      Note: Prototype HTTP handler surface inspected as the legacy baseline
    - Path: jank-prototype/web/app.js
      Note: |-
        Prototype browser logic inspected for current behavior and limitations
        Prototype browser flow inspected during ticket creation
    - Path: jank-prototype/web/index.html
      Note: Prototype UI shape inspected for current browser scope
    - Path: pkg/app/application.go
      Note: |-
        Current application boundary that the new web transport should wrap
        Application boundary inspected for the new ticket
    - Path: pkg/cli/root.go
      Note: Existing command tree that will eventually coexist with a serve command
    - Path: pkg/dsl/types.go
      Note: Current setup/plan data structures inspected for frontend editing boundaries
    - Path: pkg/recording/run.go
      Note: |-
        Existing stateful runtime inspected for web-session observation needs
        Runtime session behavior inspected for web observation needs
    - Path: ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/sources/local/screencast-studio-v2.jsx.jsx
      Note: Imported mock used as the visual and product target for the web ticket
ExternalSources: []
Summary: Chronological record of how the second ticket for the web control frontend was created and documented.
LastUpdated: 2026-04-09T17:03:00-04:00
WhatFor: Track how the web-control frontend ticket was assembled, what evidence was used, and how to review the resulting design deliverables.
WhenToUse: Read when continuing the frontend ticket, reviewing design provenance, or checking the exact repo evidence behind the recommendations.
---


# Diary

## Goal

Capture how the second ticket was created for the web control frontend, what evidence informed the architecture, and how to review the resulting deliverables.

## Step 1: Create The Web Frontend Ticket And Detailed Design Guide

This step created the follow-up ticket for the deferred browser control surface and turned the current backend, prototype web files, imported JSX mock, and first-ticket architecture work into a new intern-facing implementation guide.

The critical decision in this step was to treat the browser as a client of the existing domain layer rather than as a new recorder implementation. That means the new ticket focuses on HTTP/WebSocket transport, frontend state modeling, preview lifecycle management, and Go-plus-SPA packaging, while preserving the discovery, DSL, compile, and runtime decisions from the first ticket.

### Prompt Context

**User prompt (verbatim):** "Let's commit the mds, and open up the second web ticket. Create a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Commit the staged prototype markdown files, create a second ticket specifically for the web control frontend, write a long-form design/implementation guide for that ticket, then validate and upload the ticket bundle to reMarkable.

**Inferred user intent:** Preserve the CLI-first backend milestone as the foundation, then start a distinct web-ticket track with its own architecture and onboarding-quality guidance.

**Commit (code):** `76b3582` — "docs: add prototype dsl and research notes"

### What I did

- Committed the staged prototype markdown files:
  - `jank-prototype/dsl.md`
  - `jank-prototype/research.md`
- Created ticket `SCS-0002`.
- Added:
  - the primary design document
  - the diary document
- Inspected:
  - `jank-prototype/main.go`
  - `jank-prototype/web/app.js`
  - `jank-prototype/web/index.html`
  - `pkg/app/application.go`
  - `pkg/cli/root.go`
  - `pkg/dsl/types.go`
  - `pkg/recording/run.go`
  - the imported UI mock from the first ticket
- Wrote a detailed design guide focused on:
  - Go + SPA topology
  - frontend state ownership
  - REST and WebSocket contracts
  - preview lifecycle
  - phased implementation for an intern

### Why

- The web work is large enough that it deserves its own ticket and deliverable bundle.
- The frontend architecture should build on the first ticket’s clean backend boundaries rather than being mixed into that ticket retroactively.
- A new engineer needs a document that explains not just the browser UI, but also how the browser should relate to the existing Go domain packages.

### What worked

- The first ticket already contained the right backend architecture references, which made it easy to frame the web ticket as a follow-up rather than as a redesign.
- The imported JSX mock still served as an excellent product and visual target.
- The current backend package structure is small and explicit enough to reference directly in an onboarding doc.

### What didn't work

- The newly created design doc and diary started from empty templates, so they had to be replaced entirely rather than incrementally edited.

### What I learned

- The most important frontend design decision is not the UI library or styling approach. It is the state-ownership boundary between browser, transport, and recorder runtime.
- The prototype web layer is still useful, but mostly as a list of anti-patterns to avoid in the second implementation.

### What was tricky to build

- The subtle part was writing a web ticket that stays very detailed without duplicating the first ticket’s entire backend design. The guide had to assume the CLI-first architecture exists, then explain exactly how the frontend should sit on top of it.

### What warrants a second pair of eyes

- Whether YAML editing should be a first-class visible mode in the web UI, or only a debug/advanced tab.
- Whether preview management should use an explicit ensure/release API in version 1 or a simpler session-scoped default.

### What should be done in the future

- Validate the new ticket with `docmgr doctor`.
- Upload the ticket bundle to reMarkable.
- Optionally commit the ticket docs after review.

### Code review instructions

- Start with:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0002--screencast-studio-web-control-frontend-architecture-and-implementation-plan/design-doc/01-screencast-studio-web-control-frontend-system-design.md`
- Then compare the main claims against:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/types.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/app.js`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/index.html`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/sources/local/screencast-studio-v2.jsx.jsx`

### Technical details

Commands used in this step:

```bash
git commit -m "docs: add prototype dsl and research notes" -- jank-prototype/dsl.md jank-prototype/research.md
docmgr status --summary-only
docmgr ticket create-ticket --ticket SCS-0002 --title "Screencast Studio Web Control Frontend Architecture and Implementation Plan" --topics backend,frontend,video,audio,dsl,cli
docmgr doc add --ticket SCS-0002 --doc-type design-doc --title "Screencast Studio Web Control Frontend System Design"
docmgr doc add --ticket SCS-0002 --doc-type reference --title "Diary"
```

## Step 2: Expand The Ticket Tasks Into An Intern Execution Plan

This step turned the placeholder task list into a detailed implementation checklist for the intern who will build the web ticket. The goal was to remove ambiguity about sequencing and to make the work executable in phases instead of leaving the ticket with only broad “start implementation” notes.

The new task plan deliberately starts with backend transport and session ownership before frontend screens. That sequencing matters because the frontend should consume stable server primitives rather than inventing temporary browser-only behavior that later has to be undone.

### Prompt Context

**User prompt (verbatim):** "add detailed tasks to the new ticket for an intern to implement"

**Assistant interpretation:** Expand `SCS-0002` so the ticket contains a practical, phased, intern-friendly implementation plan with concrete milestones, acceptance criteria, and file targets.

### What I did

- Replaced the placeholder “Next” section in `tasks.md` with a phased implementation plan.
- Added explicit phases for:
  - backend server shell
  - HTTP APIs
  - recording transport
  - WebSocket events
  - preview lifecycle
  - React frontend scaffold
  - operator UI
  - production packaging
  - testing and validation
- Added acceptance criteria to each phase so implementation can be reviewed against concrete outcomes.
- Added guardrails to reduce architecture drift during implementation.
- Added suggested milestone commit boundaries.
- Updated the changelog to record that the task plan was expanded.

### Why

- A strong design document is not enough for an intern if the execution order is still ambiguous.
- The web ticket spans backend, transport, runtime observation, and frontend implementation. Without a phased checklist, it is too easy to work in the wrong order.
- Acceptance criteria make future code review much easier.

### What worked

- The design document already had enough structure to derive an implementation sequence from it.
- The backend-first ordering became much clearer once each phase was forced to name its success condition.

### What didn't work

- The original task file was too small to act as an execution plan. It needed a full rewrite rather than incremental edits.

### What I learned

- The most useful task lists for this kind of ticket are not generic issue bullets. They are phase-based and explicit about state ownership, transport boundaries, and validation commands.

### What warrants a second pair of eyes

- Whether the intern should build the YAML editor surface in the same milestone as the main operator screen or leave it as a later debugging feature.
- Whether preview lifecycle management should be completed before any browser UI work or can safely proceed in parallel once the API contract is frozen.

### What should be done in the future

- Keep the task list synchronized with actual implementation progress.
- Mark phase completion in the ticket as commits land.
- Update the design doc if implementation forces different API contracts.

### Technical details

Files edited in this step:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0002--screencast-studio-web-control-frontend-architecture-and-implementation-plan/tasks.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0002--screencast-studio-web-control-frontend-architecture-and-implementation-plan/changelog.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0002--screencast-studio-web-control-frontend-architecture-and-implementation-plan/reference/01-diary.md`

## Step 3: Implement Phase 1 With A Real Serve Command And Server Shell

This step created the actual Go server shell for the web ticket. The goal was not to add business logic yet. It was to establish the long-lived HTTP process, graceful shutdown behavior, root command wiring, and the minimal route skeleton that later phases can extend safely.

The key architectural choice was to put the transport boundary in `internal/web` rather than embedding handlers into `pkg/app` or the CLI package. That keeps `pkg/app` as the domain-facing application boundary while letting the web layer own HTTP concerns such as routing, response serialization, and server lifecycle.

### What I did

- Added `internal/web/server.go` with:
  - server config
  - `NewServer(...)`
  - `ListenAndServe(...)`
  - graceful shutdown driven by context cancellation
- Added `internal/web/routes.go` with:
  - `/api/healthz`
  - `/ws` placeholder
  - `/` placeholder
- Added a glazed `serve` command in `pkg/cli/serve.go`.
- Wired `serve` into the root command in `pkg/cli/root.go`.
- Added small duration conversion support in `pkg/cli/common.go`.
- Added handler tests in `internal/web/server_test.go`.
- Verified:
  - `go test ./internal/web ./pkg/cli`
  - `go build ./...`

### Why

- The later API and preview phases need a stable server package first.
- Graceful shutdown needs to be correct before long-lived recording or preview workers exist.
- The route skeleton makes it possible to layer handlers incrementally rather than writing one large, difficult-to-review patch.

### What worked

- The existing glazed command structure made it straightforward to add `serve` without disturbing the current CLI commands.
- The app boundary was already thin enough that the server could depend on `pkg/app` immediately.
- A focused route test provided a quick sanity check before any complex transport logic was added.

### What didn't work

- The first version of `serve` inherited the caller context directly, which would not have given clean Ctrl-C shutdown behavior. That had to be corrected by wrapping the command in `signal.NotifyContext(...)`, mirroring the existing `record` command.

### What I learned

- The cleanest first web milestone is not “add discovery handlers.” It is “make the server lifecycle boring and correct.” That reduces risk when later phases introduce long-running goroutines.

### What warrants a second pair of eyes

- Whether the placeholder `/` behavior should stay as a simple HTML response until the frontend exists or whether it should become a stricter 404 in API-only mode.
- Whether server config should gain explicit static-mode enums once the embedded frontend phase starts.

### Technical details

Files added in this step:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/routes.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server_test.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/serve.go`

Files updated in this step:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/root.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/common.go`

## Step 4: Refactor Managed Process Ownership Before Adding Web Session APIs

This step interrupted the API implementation on purpose. While building phases 2 through 4, a runtime concern became obvious: the existing `startManagedProcess(...)` helper was still doing too much hidden work. It launched FFmpeg, created background drainers, and hid process ownership behind a constructor-style function. That shape would have made the upcoming web session manager harder to reason about, especially once HTTP-triggered stop and live session observation were added.

The key change was to invert that ownership. `ManagedProcess` is now a configured process object, and `ManagedProcess.Run(ctx, ...)` is the blocking owner of the subprocess lifecycle. The outer recording session runner remains responsible for concurrency with its own `errgroup`, while the process object owns:

- starting FFmpeg
- draining stdout and stderr
- waiting for process completion
- exposing a single `Stop(...)` path for graceful termination

### Prompt Context

**User prompt (verbatim):** "Every go routine should be able to be  cancelled / passed a ctx, and its ownership should be accounted for. (i'm looking at startManagedProcess), maybe there are others."

**Follow-up prompt (verbatim):** "can we maybe reorganize this more elegantly? make ManagedProcess.Run(ctx, ...) (blocking) and such?"

**Assistant interpretation:** Stop adding new web-layer behavior until the subprocess concurrency model is explicit and defensible. Refactor `ManagedProcess` so the process object has a clean blocking execution method and the session runner owns concurrency transparently.

### What I did

- Audited the current `go` statements in the codebase.
- Confirmed the main hidden-goroutine cluster was in `pkg/recording/run.go`.
- Added structured runtime events in `pkg/recording/events.go`.
- Refactored `ManagedProcess` in `pkg/recording/run.go` so:
  - process specs are created first
  - `ManagedProcess.Run(ctx, ...)` starts and blocks on FFmpeg
  - the outer recording runtime owns worker concurrency
  - `Stop(...)` remains the explicit graceful-shutdown hook
- Expanded `pkg/app/application.go` with helpers needed by the web layer:
  - discovery snapshot
  - normalize DSL
  - compile DSL
  - plan-aware record execution with runtime event forwarding
- Verified:
  - `go test ./pkg/recording ./pkg/app`
  - `go build ./...`

### Why

- The original constructor-style process startup made goroutine ownership implicit.
- The web ticket needs session coordination and live observation, which depend on a runtime that is explicit about who owns what.
- Refactoring before adding more HTTP endpoints is cheaper than retrofitting cancellation semantics after the transport layer exists.

### What worked

- Moving the subprocess lifecycle into a blocking `Run(...)` method made the outer recording runtime easier to read immediately.
- The event hook added in the same step sets up a cleaner path for WebSocket publication later.

### What didn't work

- The first attempt to patch the runtime in place drifted against the current file and had to be re-applied in smaller steps.
- `pkg/app/application.go` still needed a follow-up cleanup so `RecordPlan(...)` preserved useful session information even when the runtime returned an error.

### What I learned

- The right concurrency boundary here is not “constructor starts work.” It is “runner owns concurrency, process object owns subprocess lifecycle.”
- That boundary is substantially better for HTTP-triggered recording, because the session manager can now supervise process workers instead of relying on hidden background state.

### What warrants a second pair of eyes

- Whether `ManagedProcess.Run(ctx, ...)` should eventually observe a dedicated internal lifecycle context separate from the outer session context, or whether the current session-owned stop behavior is already the right final shape.
- Whether scanner drain behavior should eventually surface pipe read errors more explicitly in runtime events.

### Technical details

Files added in this step:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/events.go`

Files updated in this step:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go`

## Step 5: Implement Phases 2 Through 4 With Stable APIs And A Web Recording Coordinator

This step added the first real business endpoints to the new server shell. The work covered three linked concerns:

1. expose discovery and current-session state as stable HTTP payloads,
2. expose normalize and compile operations for browser-driven DSL editing,
3. add a web-owned recording session coordinator that can start, stop, and report on one active recording at a time.

The architectural goal was to keep the transport layer thin while still making the session lifecycle observable. The cleanest way to do that was:

- add a web-layer application interface so handlers stay testable,
- add explicit API payload types instead of returning internal structs directly,
- add a `RecordingManager` that owns the currently active web-triggered recording session,
- reuse the runtime event hook added in the previous step to keep session state and logs synchronized.

### What I did

- Added `internal/web/application.go` to define the web-layer application boundary.
- Added `internal/web/api_types.go` for stable HTTP payloads and mapping helpers.
- Added `internal/web/event_hub.go` as the server-side event publication backbone.
- Added `internal/web/session_manager.go` as a single-session recording coordinator.
- Added `internal/web/handlers_api.go` for:
  - `GET /api/discovery`
  - `GET /api/session`
  - `POST /api/setup/normalize`
  - `POST /api/setup/compile`
  - `POST /api/recordings/start`
  - `POST /api/recordings/stop`
  - `GET /api/recordings/current`
- Updated route registration in `internal/web/routes.go`.
- Replaced the basic server test with fake-application endpoint and lifecycle tests.
- Verified:
  - `go test ./internal/web ./pkg/recording ./pkg/app`
  - `go build ./...`

### Why

- The browser needs a stable read model before it can render real UI state.
- The setup editor flow depends on normalize and compile APIs long before the UI exists.
- Recording control from HTTP requires a long-lived server-side owner for the active session; it cannot be implemented as a simple request-scoped call.

### What worked

- Splitting stable transport payloads into their own file made the handlers much easier to read.
- The fake application interface made it possible to test session start/stop/current flows without touching real system discovery or FFmpeg.
- The event hub gives phase 5 a direct bridge into WebSocket publication without forcing the handler layer to know about subscribers yet.

### What didn't work

- The first build failed because the web layer accidentally referenced a CLI-only duration helper. That was corrected by moving the conversion into `internal/web`.
- The testable recording manager currently uses the event hub for publication but does not yet expose `/ws`; that remains phase 5 work.

### What I learned

- The most important HTTP abstraction here is not “request handler.” It is “session owner.” Once that exists, the handlers become very small.
- The runtime event sink added during the process refactor pays off immediately: it keeps the web session model honest without duplicating state transitions in the transport layer.

### What warrants a second pair of eyes

- Whether `RecordingManager` should store explicit ownership metadata for its background `errgroup` beyond the current cancel function and active snapshot.
- Whether the current session payload should eventually separate “runtime logs” from “operator-facing warnings” more aggressively once the UI consumes both.

### Technical details

Files added in this step:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/application.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/api_types.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/event_hub.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/session_manager.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go`

Files updated in this step:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/routes.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server_test.go`

## Step 6: Implement Phases 5 And 6 With WebSocket Delivery And Preview Workers

This step completed the backend-only portion of the web ticket. It added the missing live event transport and preview runtime needed before the React frontend can start consuming the server in earnest.

The key design decision was to keep previews separate from recordings, but still make them look structurally similar:

- recordings are owned by `RecordingManager`
- previews are owned by `PreviewManager`
- both publish lifecycle changes into the shared server event hub
- `/ws` is only an event transport, not a second session owner

For previews, version 1 stayed with MJPEG rather than RTC/WebRTC. That was deliberate. The preview panels are small operator panes in a local control interface, so MJPEG keeps the backend simpler and much easier to debug while still delivering the product behavior we need right now.

### What I did

- Added `pkg/recording.BuildPreviewArgs(...)` so preview FFmpeg argv construction lives with the other media builders.
- Added `internal/web/preview_runner.go` with:
  - a preview runner interface
  - the FFmpeg-backed runner
  - JPEG frame extraction helpers
- Added `internal/web/preview_manager.go` with:
  - preview ensure logic
  - release logic
  - preview leasing/reference counting
  - per-preview state and latest-frame tracking
  - event publication for preview lifecycle and logs
- Added `internal/web/handlers_preview.go` with:
  - `POST /api/previews/ensure`
  - `POST /api/previews/release`
  - `GET /api/previews`
  - `GET /api/previews/:id/mjpeg`
- Added `internal/web/handlers_ws.go` with:
  - WebSocket upgrade
  - initial session and preview-list events
  - event fan-out from the shared event hub
- Updated `internal/web/server.go` so the server owns a preview manager in addition to the recording manager.
- Expanded `internal/web/server_test.go` with:
  - preview lifecycle endpoint tests
  - MJPEG streaming test using a fake preview runner
  - WebSocket handshake and initial-event test
- Verified:
  - `go test ./internal/web ./pkg/recording ./pkg/app`
  - `go build ./...`

### Why

- The frontend cannot be built meaningfully until the server can push state changes and serve preview frames.
- A preview runner abstraction keeps phase 6 testable without real FFmpeg or real capture devices.
- Separating preview workers from recording workers avoids the exact ownership ambiguity we were already cleaning up in the runtime.

### What worked

- The event hub turned out to be the right seam: once it existed, `/ws` was mostly a transport concern.
- The preview runner abstraction made the MJPEG endpoint testable with a fake frame source instead of an actual media process.
- Reusing the recording package for preview argv generation kept the FFmpeg-specific logic out of the HTTP package.

### What didn't work

- The preview and MJPEG step added more moving parts than the earlier API phases, so it only became safe after the managed-process refactor was already done.
- The handler layer needed a second round of route and payload expansion once preview endpoints existed.

### What I learned

- The preview manager is the media equivalent of the recording manager: the hardest problem is not “stream bytes,” it is “own lifecycle and state transitions cleanly.”
- For the local control-surface use case, MJPEG remains a pragmatic first transport. It is not elegant, but it is observable and easy to wire into browser `<img>` tags.

### What warrants a second pair of eyes

- Whether preview managers should retain failed previews longer for debugging, rather than deleting finished previews once leases reach zero.
- Whether the current preview event shapes are already sufficient for the frontend or should grow explicit `preview.created` and `preview.released` event types beyond the state updates.

### Technical details

Files added in this step:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_runner.go`

Files updated in this step:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/routes.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/api_types.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/ffmpeg.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server_test.go`
