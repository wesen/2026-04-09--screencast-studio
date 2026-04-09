---
Title: Diary
Ticket: SCS-0001
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
    - Path: go.mod
      Note: Root module added in Step 3
    - Path: jank-prototype/dsl.md
      Note: Primary DSL evidence inspected during documentation
    - Path: jank-prototype/main.go
      Note: Primary backend evidence inspected during documentation
    - Path: jank-prototype/research.md
      Note: Platform research inspected during documentation
    - Path: pkg/app/application.go
      Note: Placeholder application boundary added in Step 3
    - Path: pkg/cli/discovery/list.go
      Note: Discovery list command skeleton added in Step 3
    - Path: pkg/cli/setup/compile.go
      Note: Compile command now emits a real execution plan in Step 5
    - Path: pkg/cli/root.go
      Note: CLI skeleton added in Step 3
    - Path: pkg/discovery/service.go
      Note: Discovery implementation added in Step 4
    - Path: pkg/discovery/types.go
      Note: Discovery descriptor types added in Step 4
    - Path: pkg/dsl/compile.go
      Note: DSL-to-plan compilation added in Step 5
    - Path: pkg/dsl/normalize.go
      Note: DSL normalization and validation added in Step 5
    - Path: pkg/recording/ffmpeg.go
      Note: FFmpeg builders and stdout/stderr capture added in Step 6
    - Path: pkg/recording/run.go
      Note: Recording runtime refactored around an explicit session state machine in Step 6
    - Path: pkg/recording/session.go
      Note: Formal recording session states and transition helpers added in Step 6
    - Path: ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/sources/local/screencast-studio-v2.jsx.jsx
      Note: Imported UI mock inspected during documentation
ExternalSources: []
Summary: Chronological record of how the screencast studio architecture ticket, supporting docs, and delivery bundle were created.
LastUpdated: 2026-04-09T14:53:43.000000000-04:00
WhatFor: Chronological record of how the screencast studio architecture ticket was assembled and what evidence shaped the design.
WhenToUse: Read when reviewing this ticket, continuing the documentation work, or checking what commands and sources were used.
---




# Diary

## Goal

Capture how the initial architecture ticket was created, what evidence informed the design, and how to review the resulting documents and imported artifacts.

## Step 1: Create The Architecture Ticket And Long-Form Design Guide

This step created the first ticket workspace for the screencast studio and turned the prototype plus imported JSX mock into a concrete architecture guide. The work focused on gathering evidence first, then writing one detailed design document that a new intern can use as an onboarding and implementation map.

The critical decision in this step was to center the design around a normalized setup DSL and a compiled runtime plan instead of extending the prototype handlers directly. That keeps the browser, backend, and OS adapters decoupled and gives the next implementation phase a much cleaner shape.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr tickt to create a screencast studio with go backend (using glazed commands for the flags) and streaming video to a control web frontend (docmgr import file --file ~/Downloads/screencast-studio-v2.jsx), where we use an intermediate DSL data structure to represent the setup (see jank-prototype/dsl.md and the research in jank-prototype/research.md ).

The software allows selecting display/region/window/camera video streams, and audio, and view previews, and then record the streams to separate files according to templates.

Let's sketch out the arhitecture in detail.

Create a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable.

[Image #1]"

**Assistant interpretation:** Create a new docmgr ticket, import the provided JSX mock, write a detailed architecture and implementation guide grounded in the prototype DSL and research notes, then validate and upload the ticket bundle to reMarkable.

**Inferred user intent:** Establish a serious implementation-ready plan for turning the prototype into a real screencast studio application, with enough structure that a new engineer can start building confidently.

**Commit (code):** N/A — documentation-only turn

### What I did

- Read the `docmgr`, `ticket-research-docmgr-remarkable`, `remarkable-upload`, and `diary` skill instructions.
- Inspected the prototype files:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/app.js`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/index.html`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/dsl.md`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/research.md`
- Inspected the imported UI reference from `/home/manuel/Downloads/screencast-studio-v2.jsx`.
- Added missing docmgr vocabulary topics: `frontend`, `video`, `audio`, `dsl`, `cli`.
- Created ticket `SCS-0001`.
- Added the primary design doc and diary doc.
- Imported the JSX mock into the ticket workspace with `docmgr import file`.
- Wrote the index, design doc, and task list.

### Why

- The prototype already contains useful behavior and file-backed evidence, so the architecture needed to stay anchored to that code.
- The imported JSX mock makes the product model clearer than the current HTML UI and therefore had to be included in the ticket.
- The ticket needed to be self-contained so the reMarkable bundle would still make sense without local shell context.

### What worked

- `docmgr` was already initialized for this repo.
- The prototype files were small enough to inspect directly and cite concretely.
- The JSX mock lined up well with the DSL-based product concept.
- The ticket workspace and import flow both worked cleanly on the first try.

### What didn't work

- The imported file ended up named `screencast-studio-v2.jsx.jsx` inside `sources/local/` because the provided `--name` included the `.jsx` extension and `docmgr import file` preserved its own suffixing behavior. This is not blocking, but it is worth noting when referencing the imported artifact.

### What I learned

- The current prototype is already very close to a compiler pipeline conceptually; it just does not name that layer explicitly.
- The biggest missing piece is not media support but architecture hygiene: state separation, discovery boundaries, and a proper frontend model.
- The imported mock strongly suggests the app should be framed around a studio setup editor, not a raw DSL textarea.

### What was tricky to build

- The main challenge was deciding how opinionated the version 1 preview transport should be. The current prototype uses MJPEG, while a richer app could justify WebRTC. I chose to recommend MJPEG for source-card previews plus WebSocket for control events because it keeps the version 1 system understandable for an intern while preserving a clean upgrade path later.
- Another subtle point was platform scope. The research note is X11-specific, and the prototype already uses X11, V4L2, and PulseAudio. The document therefore treats Linux and X11 as the version 1 target rather than pretending the design is cross-platform already.

### What warrants a second pair of eyes

- The preview transport decision for source cards.
- Whether the setup DSL should stay close to the current YAML shape or be renamed more aggressively around `StudioSpec`.
- Whether the first production target is explicitly Linux/X11-only.

### What should be done in the future

- Implement the code according to the phases in the design doc.
- Add a playbook document once the first runnable backend exists.
- Revisit preview transport if the UI needs higher frame rates or more simultaneous previews.

### Code review instructions

- Start with `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/design-doc/01-screencast-studio-system-design.md`.
- Then compare the architecture claims against:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/main.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/web/app.js`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/dsl.md`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/jank-prototype/research.md`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/sources/local/screencast-studio-v2.jsx.jsx`
- Validate with `docmgr doctor --ticket SCS-0001 --stale-after 30`.
- Validate the reMarkable bundle via `remarquee cloud ls /ai/2026/04/09/SCS-0001 --long --non-interactive`.

### Technical details

Commands run:

```bash
docmgr status --summary-only
docmgr ticket create-ticket --ticket SCS-0001 --title "Screencast Studio Architecture and Implementation Plan" --topics backend,frontend,video,audio,dsl,cli
docmgr doc add --ticket SCS-0001 --doc-type design-doc --title "Screencast Studio System Design"
docmgr doc add --ticket SCS-0001 --doc-type reference --title "Diary"
docmgr import file --ticket SCS-0001 --file ~/Downloads/screencast-studio-v2.jsx --name "screencast-studio-v2.jsx"
```

Primary deliverables:

- `ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/index.md`
- `ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/design-doc/01-screencast-studio-system-design.md`
- `ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/reference/01-diary.md`

## Step 2: Re-Scope The Ticket To A CLI-First Milestone

This step changed the execution order of the project. Instead of treating the web control surface as part of the first implementation milestone, the ticket now treats `discover`, `compile`, and `record` as the first deliverable and explicitly defers the browser UI to a second ticket.

That change matters because it gives the implementation a much better debugging surface. If the DSL, discovery, and recording primitives work cleanly from the terminal, the later web ticket can reuse those packages instead of inventing a separate path through the system.

### Prompt Context

**User prompt (verbatim):** "Update to also have a \"record\" CLI verb that compiles a given DSL and then executes it. Maybe we should actually build that part first before tackling the web part, to make sure the primitives are working correctly. Update the design doc for that, then add detailed tasks for that first step of having a working discover + compile + CLI cpature functionality. We'll do web in a second ticket. 

Then, work on these tasks one by one, committing at appropriate intervals, keeping a detailed diary as you work."

**Assistant interpretation:** Re-scope the design to prioritize a CLI-first milestone, add the `record` verb explicitly, define the detailed task list for `discover` plus `compile` plus CLI capture, then implement that milestone in small committed steps.

**Inferred user intent:** Reduce project risk by proving the underlying capture primitives from the terminal before investing effort in the control frontend.

**Commit (code):** `9e194aa` — "docs: refocus screencast studio on cli milestone"

### What I did

- Rewrote the architecture document to make the current milestone CLI-first.
- Added `record` as a first-class command in the command inventory and implementation plan.
- Removed the web API and frontend phases from the active milestone and marked them as deferred follow-up work.
- Expanded the task list into a concrete checklist for the CLI milestone.
- Identified a structural prerequisite: the repo needs a top-level `go.mod`, because only `jank-prototype/go.mod` exists today.

### Why

- The browser is a poor place to debug unstable discovery and recording primitives.
- A terminal-first milestone is easier to test, easier to script, and easier to review incrementally.
- The new order gives the eventual web ticket a much cleaner contract: it can call into existing compiler and runtime packages.

### What worked

- The original design doc already had a clear compiler/runtime direction, so re-scoping it to CLI-first was mostly a matter of changing sequencing and command boundaries.
- The imported JSX mock still remains useful as a product reference even though it is now deferred.

### What didn't work

- My first large patch to the design doc failed because the context no longer matched exactly. I re-read the file with line numbers and reapplied the edits in smaller chunks.

### What I learned

- The absence of a top-level `go.mod` is not just housekeeping. It is the first real implementation task because the project guidance expects the active module to live at the repo root.
- The right first code slice is now obvious: module plus root CLI, then DSL, then discovery, then `record`.

### What was tricky to build

- The subtle part of the re-scope was preserving the long-term architecture while changing the active milestone. I wanted the design doc to stay useful for the future web ticket without letting web concerns dominate the immediate implementation plan.

### What warrants a second pair of eyes

- Whether the CLI milestone should include pause and resume semantics or stop at start and stop.
- Whether the root module should absorb the prototype code or merely use it as reference during reimplementation.

### What should be done in the future

- Implement the CLI milestone in focused slices.
- Commit after each slice.
- Open the follow-up web ticket only after the CLI milestone is stable.

### Code review instructions

- Re-read the updated task list in `ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/tasks.md`.
- Re-read the updated implementation phases in `ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/design-doc/01-screencast-studio-system-design.md`.
- Confirm that the immediate work is now `discover`, `compile`, and `record`, not web delivery.

### Technical details

Key edits were made to:

- `ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/design-doc/01-screencast-studio-system-design.md`
- `ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/tasks.md`
- `ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/reference/01-diary.md`

## Step 3: Establish The Root Module And CLI Skeleton

This step created the first executable code for the new CLI-first direction. The goal was not to implement real discovery or recording yet, but to create the top-level module, root command tree, and Glazed wiring so the later implementation slices have a stable place to land.

The result is a new root binary with `discovery`, `setup`, and `record` commands, plus a small application container that exposes the future service boundaries. At this point the commands compile and render help correctly, but the underlying discovery and compile methods still return `not implemented`.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Build the first implementation slice of the CLI milestone by establishing the root module and command plumbing before filling in real runtime behavior.

**Inferred user intent:** Make the project executable from the repo root and reduce architectural thrash by proving the command structure first.

**Commit (code):** `047d61c` — "feat: add cli-first screencast studio skeleton"

### What I did

- Created the root `go.mod` with Glazed, Cobra, and `pkg/errors`.
- Added `cmd/screencast-studio/main.go`.
- Added `pkg/app/application.go` as the first application boundary.
- Added the root CLI and Glazed command helpers under `pkg/cli/`.
- Added grouped commands for:
  - `discovery list`
  - `setup compile`
  - `setup validate`
  - `record`
- Ran:
  - `go mod tidy`
  - `go mod download github.com/pkg/errors github.com/spf13/cobra github.com/go-go-golems/glazed`
  - `go build ./...`
  - `go run ./cmd/screencast-studio --help`
  - `go run ./cmd/screencast-studio discovery list --help`
  - `go run ./cmd/screencast-studio setup compile --help`
  - `go run ./cmd/screencast-studio record --help`

### Why

- The repo previously had no top-level module, only `jank-prototype/go.mod`.
- The command tree needed to exist before the deeper discovery and DSL work could be implemented cleanly.
- Verifying help output early proves the Glazed wiring and flag layout without mixing in media-runtime debugging yet.

### What worked

- The local Glazed patterns from nearby repos transferred cleanly.
- The root command now exposes the intended CLI surface.
- `go build ./...` succeeds from the repo root.

### What didn't work

- The first `go build ./...` complained about missing `go.sum` entries even though `go mod tidy` had already run. Running `go mod download` for the direct dependencies resolved that.
- `go run ./cmd/screencast-studio discovery list` still exits with `not implemented`, which is expected for this slice but confirms that the next task is real discovery logic.

### What I learned

- The CLI wiring is now stable enough that the next steps can focus on discovery and DSL behavior rather than Cobra or Glazed mechanics.
- A root module plus command tree is a meaningful milestone in this repo because it moves active development out of the prototype submodule.

### What was tricky to build

- The main subtlety was committing only my files while leaving the user's staged prototype research files untouched. I handled that by committing only the new root-module and CLI paths explicitly.

### What warrants a second pair of eyes

- The initial module path choice: `github.com/wesen/2026-04-09--screencast-studio`.
- Whether the temporary `pkg/app` placeholder should stay minimal or be replaced quickly by more domain-specific packages in the next slice.

### What should be done in the future

- Implement real structured discovery output next.
- Replace the `not implemented` application methods with actual services.
- Start extracting the DSL package immediately after discovery plumbing is in place.

### Code review instructions

- Start with:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/go.mod`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/cmd/screencast-studio/main.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/root.go`
- Then inspect the command implementations under:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/discovery/`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/setup/`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/record.go`
- Validate with:
  - `go build ./...`
  - `go run ./cmd/screencast-studio --help`

### Technical details

The current command status is:

```text
screencast-studio --help            -> works
screencast-studio discovery list    -> wired, returns not implemented
screencast-studio setup compile     -> wired
screencast-studio record            -> wired
```

## Step 4: Implement Real Platform Discovery

This step replaced the placeholder discovery method with actual source enumeration backed by the local machine. The new implementation shells out to the platform tools that are already available in the environment and turns their output into structured rows for displays, windows, cameras, and audio inputs.

This is the first slice where the CLI starts returning real user-facing data instead of just plumbing. That matters because discovery is one half of the compile problem: if the runtime cannot enumerate concrete sources reliably, there is nothing useful to compile a setup against.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Keep working through the CLI-first task list and implement discovery as a concrete, usable command.

**Inferred user intent:** Prove the platform primitives on the real machine before moving on to compile and record.

**Commit (code):** `cd94620` — "feat: implement platform discovery commands"

### What I did

- Added `pkg/discovery/types.go`.
- Added `pkg/discovery/service.go`.
- Implemented display discovery using `xrandr --listmonitors`.
- Implemented window discovery using `xprop -root _NET_CLIENT_LIST`, `xprop -id`, and `xwininfo -id`.
- Implemented camera discovery using `v4l2-ctl --list-devices`.
- Implemented audio-input discovery using `pactl list short sources`.
- Updated `pkg/app/application.go` so `discovery list` returns structured rows by kind.
- Verified:
  - `go build ./...`
  - `go run ./cmd/screencast-studio discovery list --kind display --output json`
  - `go run ./cmd/screencast-studio discovery list --kind audio --output json`

### Why

- Discovery had to become real before `setup compile` could do anything meaningful.
- Using the existing platform utilities is the fastest way to validate the runtime model before deciding whether any direct X11 bindings are worth adding later.
- Filtering by kind now avoids making audio discovery depend on X11, which is the right command behavior.

### What worked

- The environment already had `DISPLAY=:0`, `xrandr`, `xprop`, `xwininfo`, `v4l2-ctl`, `pactl`, and `ffmpeg`.
- The display, window, camera, and audio enumeration commands all returned real data.
- The command now emits useful JSON rows that can feed later compile work.

### What didn't work

- The first `xrandr` parser was wrong because it treated the monitor connector token as the geometry token. The failure was:

```text
Error: unexpected monitor geometry token "eDP-1"
```

- I fixed that by parsing the monitor line as fields and extracting geometry from the dedicated geometry token instead of from the connector token.

### What I learned

- The CLI-first sequence is already paying off. The parser bug was easy to spot and fix from a direct command invocation, and it would have been noisier to diagnose through a future web layer.
- For discovery, command-level isolation matters. `--kind audio` should succeed even if the X11 side is unavailable.

### What was tricky to build

- The tricky part was not discovering the existence of the tools. It was designing the parsing logic so that it is strict enough to catch malformed output but not so brittle that one token-layout change causes the entire command to fail unexpectedly.

### What warrants a second pair of eyes

- Whether command-backed discovery is sufficient for version 1 or whether some parts should move to native bindings later.
- Whether the window titles and geometries should include frame extents in addition to client bounds.

### What should be done in the future

- Add parser-focused tests for the discovery package.
- Start the DSL extraction immediately so `setup compile` can resolve against these discovery results.
- Consider fallback behavior when one platform tool is missing but others are present.

### Code review instructions

- Start with:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/discovery/service.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/discovery/types.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go`
- Validate with:
  - `go build ./...`
  - `go run ./cmd/screencast-studio discovery list --kind display --output json`
  - `go run ./cmd/screencast-studio discovery list --kind audio --output json`

### Technical details

Observed platform primitives during implementation:

```text
DISPLAY=:0
/usr/bin/xrandr
/usr/bin/xwininfo
/usr/bin/xprop
/usr/bin/v4l2-ctl
/usr/bin/pactl
/usr/bin/ffmpeg
```

## Step 5: Extract The DSL Package And Make `setup compile` Real

This step turned the setup DSL from architecture intent into executable code. The repo now has a dedicated `pkg/dsl` package that can load a YAML or JSON setup file, normalize it into an effective configuration, validate the source definitions, and compile the result into a concrete output manifest for the session.

This is the second major milestone in the CLI-first plan because it creates the stable boundary that `record` will execute next. Discovery gives us visibility into the machine. Compilation gives us a deterministic runtime plan that can be inspected and tested before any FFmpeg process is launched.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Continue the CLI-first implementation by extracting the DSL into its own package and replacing the placeholder compile command with a real planning pipeline.

**Inferred user intent:** Make the setup compiler concrete and reviewable before attempting recording execution.

**Commit (code):** Pending at the time of this diary entry

### What I did

- Added `pkg/dsl/types.go` with the configuration, normalized, and compiled-plan structs.
- Added `pkg/dsl/load.go` so setup files can be read and decoded as JSON or YAML.
- Added `pkg/dsl/normalize.go` to validate source definitions, merge defaults, and emit warnings for unsupported-but-known fields.
- Added `pkg/dsl/templates.go` and `pkg/dsl/helpers.go` for destination rendering, slug generation, and path handling.
- Added package tests:
  - `pkg/dsl/normalize_test.go`
  - `pkg/dsl/templates_test.go`
- Updated `pkg/app/application.go` so `CompileFile` now performs:
  - load file
  - normalize DSL
  - build compiled plan
- Updated `pkg/cli/setup/compile.go` so the command now emits one structured row per planned output instead of a placeholder session id.
- Updated `pkg/cli/record.go` dry-run output so it exposes output and warning counts from the real compile step.
- Ran:
  - `go mod tidy`
  - `go test ./pkg/dsl`
  - `go build ./...`
  - `go run ./cmd/screencast-studio setup compile --file ./jank-prototype/examples/example.yaml --output json`

### Why

- `record` should execute a reviewed plan, not parse ad hoc config at launch time.
- Keeping normalization in `pkg/dsl` makes the same logic reusable from the future web ticket without duplicating policy in handlers.
- Package-level tests are the right place to lock down DSL rules, template rendering, and default-merging behavior before FFmpeg orchestration enters the picture.

### What worked

- The sample config in `jank-prototype/examples/example.yaml` mapped cleanly into the new DSL package.
- The compile command now returns five concrete outputs from the sample setup: four video files and one mixed-audio file.
- `go build ./...` and `go test ./pkg/dsl` both passed after the module metadata was updated.

### What didn't work

- The first test run failed with:

```text
go: updates to go.mod needed; to update it:
    go mod tidy
```

- That was expected once `gopkg.in/yaml.v3` and `github.com/stretchr/testify/require` were introduced. Running `go mod tidy` resolved it.
- I briefly called a nonexistent `dsl.Now()` helper while wiring the application layer. I replaced that with `time.Now()` immediately instead of inventing a premature abstraction.

### What I learned

- The prototype DSL is already close to a valid compiler boundary. The main work was separating configuration loading, normalization, and plan-building into explicit functions with tests.
- The compiled plan is already useful as a standalone artifact. It makes `setup compile` a meaningful review command instead of just a validation pass.

### What was tricky to build

- The most delicate part was deciding which unsupported fields should be hard validation errors and which should be preserved as warnings. For now, obviously incomplete features like `follow_resize`, `noise_gate`, and `denoise` remain visible to users as warnings so the runtime contract stays honest.
- Template rendering also needed to be strict enough to avoid silently generating nonsense paths while still remaining easy to author in YAML.

### What warrants a second pair of eyes

- Whether the current `CompiledPlan` should soon grow explicit FFmpeg input/output argument slices, or whether that should wait for the runtime layer.
- Whether mixed audio should always compile to one output or eventually support multiple named mixes in the DSL.

### What should be done in the future

- Implement `record --file <path>` by translating `CompiledPlan` entries into supervised FFmpeg jobs.
- Add command-level smoke tests that exercise `setup compile` and `record --dry-run`.
- Open the second ticket for the web control surface only after the recording runtime is stable.

### Code review instructions

- Start with:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/normalize.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/compile.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/setup/compile.go`
- Validate with:
  - `go test ./pkg/dsl`
  - `go build ./...`
  - `go run ./cmd/screencast-studio setup compile --file ./jank-prototype/examples/example.yaml --output json`

### Technical details

Observed compile output from the sample config:

```text
recordings/demo/Full Desktop.mov
recordings/demo/Browser Window.mov
recordings/demo/Top Left Region.mov
recordings/demo/Webcam.mov
recordings/demo/audio-mix.wav
```

## Step 6: Introduce A Formal Recording Session State Machine

This step replaced the implicit runtime lifecycle with an explicit recording session state machine and used that refactor to make the `record` command much easier to reason about. The recording package now owns concrete session states, event handling, and shutdown ownership instead of spreading those decisions across context branches and worker waiters.

The practical result is better than just cleaner code. The runtime no longer truncates bounded runs to zero-byte outputs by sending a stop request at the exact same moment FFmpeg is supposed to end naturally. Bounded runs now let FFmpeg own the nominal duration through `-t`, while Go only uses a later hard-timeout as a fallback if the worker fails to exit.

### Prompt Context

**User prompt (verbatim):** "ok, make a formal state machine. Make an implementation / architecture guide in the ticket to clearly document it, then implement it and see if that makes handling the situation better / more robust / more clear."

**Assistant interpretation:** Document the recording lifecycle as a formal state machine in the ticket, then refactor the runtime to use explicit states and validate whether that clarifies and improves shutdown behavior.

**Inferred user intent:** Replace hand-wavy runtime coordination with a precise, reviewable lifecycle model that makes future debugging and frontend integration safer.

**Commit (code):** Pending at the time of this diary entry

### What I did

- Extended the ticket design doc with a formal recording session state machine:
  - explicit states
  - explicit events
  - legal transition table
  - ownership rules
  - session pseudocode
- Added `pkg/recording/session.go` with:
  - `SessionState`
  - session event types
  - transition helper
  - process tracking helpers
- Refactored `pkg/recording/run.go` so the runtime now:
  - creates a `Session`
  - transitions from `starting` to `running`
  - reacts to worker exits, cancellation, and hard timeouts as events
  - transitions into `stopping`, `finished`, or `failed`
- Changed bounded-run policy so FFmpeg owns the nominal duration via `-t`, and Go only triggers a later hard-timeout fallback.
- Added stdout capture alongside stderr capture for FFmpeg worker processes.
- Added a startup log line with the exact FFmpeg argv used for each worker.
- Added ticket-local repro scripts under:
  - `ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/`
- Validated with:
  - `go test ./...`
  - `go build ./...`
  - `./ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/smoke-record-region.sh`
  - `TIMEFMT='%E'; time ./ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/direct-ffmpeg-region-smoke.sh /tmp/direct-smoke-runtime-check.mkv`

### Why

- The previous runtime had the right components but the wrong abstraction level. It coordinated contexts, worker goroutines, and graceful stop logic without naming the actual lifecycle states.
- A formal state machine is a better fit for a recorder because the system needs to distinguish:
  - natural bounded completion
  - unexpected worker exit
  - user interrupt
  - hard-timeout fallback
- The explicit state model will also be the right shape for the eventual web control surface because a frontend can subscribe to session state changes much more naturally than to opaque goroutine behavior.

### What worked

- The new state machine made the runtime flow much easier to read and reason about.
- The zero-byte bounded-recording failure was eliminated once the runtime stopped sending `q` exactly when the nominal duration elapsed.
- The `record` command now returns a `final_state` and `state_reason`.
- Capturing both stdout and stderr confirmed that stderr is the relevant stream for FFmpeg diagnostics on this path.

### What didn't work

- Even after the runtime semantics were corrected, the one-second X11 smoke capture still took about five to six seconds of wall-clock time to complete.
- The key validation result is that this timing behavior reproduced in a direct FFmpeg call outside the runtime:

```text
TIMEFMT='%E'; time ./.../direct-ffmpeg-region-smoke.sh /tmp/direct-smoke-runtime-check.mkv
...
5.57s
```

- That means the remaining wall-clock delay is not a state-machine bug in the recorder. It appears to be an FFmpeg/X11 behavior on this host.

### What I learned

- The formal state machine absolutely improved robustness and clarity even though it did not change the host-level FFmpeg timing.
- The most important runtime insight was separating:
  - bounded natural completion
  - Go-side hard-timeout fallback
- Once those were treated as different events, the shutdown logic stopped fighting FFmpeg.

### What was tricky to build

- The hardest part was making the state machine own policy without forcing `errgroup` to become the state machine itself.
- Another subtle issue was that one producer goroutine originally waited only on the parent context, which kept the runtime from shutting down cleanly after local completion. That had to be taught to also observe local producer cancellation.

### What warrants a second pair of eyes

- Whether the bounded-run hard-timeout should transition to `failed` immediately or still attempt graceful stop first and then fail.
- Whether the session should eventually keep a ring buffer of recent FFmpeg log lines and attach them to failures.

### What should be done in the future

- Add explicit session-state tests beyond the current builder tests.
- Add command-level smoke validation into a reproducible test harness where practical.
- Investigate why short X11 captures take roughly five seconds of wall-clock time on this host even with `-t 1`.

### Code review instructions

- Start with:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/session.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/ffmpeg.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/record.go`
- Then compare the runtime behavior with the ticket scripts:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/smoke-record-region.sh`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0001--screencast-studio-architecture-and-implementation-plan/scripts/direct-ffmpeg-region-smoke.sh`

### Technical details

Observed bounded-run argv after the refactor:

```text
ffmpeg -hide_banner -loglevel error -y -t 1 -f x11grab -framerate 2 -video_size 320x240 -draw_mouse 0 -i :0.0+0,0 -c:v libx264 -preset veryfast -crf 24 -pix_fmt yuv420p ttmp/smoke-output/smoke_region.mkv
```

Observed result from the runtime smoke script:

```text
final_state: finished
state_reason: all bounded workers exited cleanly
output file: ttmp/smoke-output/smoke_region.mkv
```
