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
LastUpdated: 2026-04-09T16:02:00-04:00
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
