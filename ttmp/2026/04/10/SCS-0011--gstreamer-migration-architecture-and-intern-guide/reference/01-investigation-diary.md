---
Title: Investigation diary
Ticket: SCS-0011
Status: active
Topics:
    - screencast-studio
    - backend
    - frontend
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/preview_manager.go
      Note: Evidence source for preview lifecycle and suspend/restore behavior
    - Path: internal/web/session_manager.go
      Note: Evidence source for recording state management during the investigation
    - Path: pkg/app/application.go
      Note: Top-level application service used during architecture mapping
    - Path: pkg/recording/run.go
      Note: Evidence source for the current recording subprocess model
    - Path: proto/screencast/studio/v1/web.proto
      Note: Evidence source for current browser/server contract assumptions
ExternalSources: []
Summary: Chronological record of the investigation, documentation, and delivery work for the GStreamer migration planning ticket.
LastUpdated: 2026-04-10T13:07:13.411757966-04:00
WhatFor: Record the investigation process, evidence gathering, documentation work, and delivery steps for the GStreamer migration ticket.
WhenToUse: Use when reviewing how the migration guide was produced or resuming ticket SCS-0011.
---


# Investigation diary

## Goal

Capture the investigation and documentation work for planning a migration from the current FFmpeg subprocess runtime to a GStreamer-based runtime.

## Step 1: Create the ticket and map the current architecture

I started by creating a dedicated ticket workspace and then reading the current runtime from the top down. The first important observation was that the codebase already has a strong planning boundary. The DSL normalization and compile stages are separate from the runtime execution layer, which means the migration does not need to redesign the product model in order to replace the media engine.

The second important observation was that there is no single media-backend seam yet. FFmpeg-specific behavior is split across the recording runner and the preview runner, while the server/runtime layer has accumulated handoff logic to paper over device conflicts. That is exactly the kind of architecture that produces “brittle and complex” behavior in live capture systems.

### Prompt Context

**User prompt (verbatim):** "Ok, great. Now create a new ticket because I want to port this over to use gstreamer, the ffmpeg process juggling is quite brittle and complex. 

Create a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a new ticket, investigate the current runtime architecture in depth, write a detailed intern-facing migration guide for moving to GStreamer, and upload the resulting document bundle to reMarkable.

**Inferred user intent:** Reduce the long-term maintenance cost of the media backend by planning a thoughtful migration rather than continuing to patch increasingly brittle FFmpeg orchestration.

**Commit (code):** N/A

### What I did
- Ran `docmgr status --summary-only`.
- Created ticket `SCS-0011`.
- Added the primary design doc and investigation diary via `docmgr doc add`.
- Read and line-anchored the main files across:
  - DSL normalization and compile
  - application service boundary
  - FFmpeg arg building
  - recording session/process supervision
  - preview runner and preview manager
  - recording manager
  - web server and handlers
  - protobuf API
  - React studio page

### Why
- The migration guide needed to be evidence-based rather than aspirational.
- A new intern will need both conceptual orientation and concrete file entry points.

### What worked
- The codebase has a relatively clean planning boundary.
- The web/runtime layers already expose stable session and preview state models that can likely survive the backend migration.

### What didn't work
- N/A

### What I learned
- The core problem is not the DSL, the browser, or the HTTP API. The core problem is duplicated media backend control paths and subprocess-oriented lifecycle management.

### What was tricky to build
- The hard part of the investigation was separating “valuable orchestration code that should stay” from “backend-specific execution code that should be replaced.” For example, `PreviewManager` and `RecordingManager` look complex, but much of that complexity is product state management, not FFmpeg coupling. The migration guide therefore had to be explicit about which parts should be preserved and which parts should move behind a runtime seam.

### What warrants a second pair of eyes
- Whether the proposed runtime seam is the right size.
- Whether the recommendation against `gst-launch-1.0` subprocess migration is practical given repository constraints and available Go bindings.

### What should be done in the future
- Validate the GStreamer binding strategy before implementation starts.
- Decide whether the first GStreamer milestone should use one pipeline per output or invest immediately in shared graph fanout.

### Code review instructions
- Start with the design doc and compare each major recommendation against the referenced files.
- Confirm that each “keep stable” recommendation actually matches current API/browser expectations.

### Technical details
- Commands run:

```bash
docmgr status --summary-only
docmgr ticket create-ticket --ticket SCS-0011 --title "GStreamer migration architecture and intern guide" --topics screencast-studio,backend,frontend
docmgr doc add --ticket SCS-0011 --doc-type design-doc --title "GStreamer migration analysis and implementation guide"
docmgr doc add --ticket SCS-0011 --doc-type reference --title "Investigation diary"
rg -n "ffmpeg|preview|session manager|recording run|BuildPreviewArgs|buildVideoRecordArgs|buildAudioMixArgs|CompileDSL|destination_templates" pkg internal ui/src proto
```

## Step 2: Write the migration guide and prepare it for delivery

After the evidence pass, I wrote the primary design document as an intern-facing migration guide rather than a short RFC. The document explains the current architecture first, because that is the only way a new engineer can make sense of the migration scope. It then moves into gap analysis, runtime invariants, recommended runtime interfaces, GStreamer design direction, file-by-file guidance, phased implementation, testing strategy, risks, and alternatives.

I intentionally kept the browser and DSL stable in the proposed plan. That is the most important architectural simplification in the document. The guide argues that the migration should replace the media runtime while preserving the current compile model and browser-facing contracts wherever possible.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Turn the investigation into a detailed, self-contained migration document and deliver it through the ticket workflow and reMarkable.

**Inferred user intent:** Produce a document that an intern can execute from, not just read for context.

**Commit (code):** N/A

### What I did
- Replaced the generated placeholder docs with:
  - a detailed ticket index,
  - a concrete task list,
  - a long-form design/migration guide,
  - this diary.
- Updated the ticket changelog with the investigation and document-writing progress.

### Why
- The default ticket skeleton was too generic to serve as an implementation guide.
- The reMarkable bundle needs complete, readable documents rather than placeholders.

### What worked
- The evidence gathered from the codebase was enough to support a concrete phased migration plan.
- The existing file structure maps cleanly onto the proposed “planning layer above runtime layer” explanation.

### What didn't work
- N/A

### What I learned
- The migration narrative becomes much clearer when framed as “keep planning and web contracts stable; replace media execution.”

### What was tricky to build
- The main writing challenge was balancing exhaustiveness with navigability. A new intern needs a lot of detail, but they also need a stable reading order. I used a fixed progression: current state, gaps, invariants, proposed design, phases, testing, risks, file-level references.

## Step 3: Validate the docs and upload the bundle to reMarkable

After the guide and diary were written, I turned the ticket into a deliverable instead of leaving it as a draft. That meant linking the docs back to the relevant code, running the doc health checks, performing a dry-run for the reMarkable bundle generation, and then uploading the final PDF bundle to the expected remote path.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Finish the ticket end to end, including validation and document delivery to reMarkable.

**Inferred user intent:** Produce a durable planning artifact that is both locally searchable and easy to read away from the codebase.

**Commit (code):** N/A

### What I did
- Added `RelatedFiles` entries to the design doc and diary with `docmgr doc relate`.
- Ran `docmgr doctor --ticket SCS-0011 --stale-after 30`.
- Ran a dry-run bundle upload with `remarquee upload bundle --dry-run`.
- Uploaded the real bundle to reMarkable.
- Verified the remote listing with `remarquee cloud ls /ai/2026/04/10/SCS-0011 --long --non-interactive`.

### Why
- The ticket needed to be continuation-friendly inside the repository.
- The reMarkable upload needed explicit verification so the final report could state the artifact path confidently.

### What worked
- `docmgr doctor` passed cleanly.
- The bundle upload succeeded without requiring retries.
- The uploaded document was visible at the expected remote path.

### What didn't work
- N/A

### What I learned
- The combination of `docmgr doc relate` plus a single bundled reMarkable PDF works well for this kind of intern-facing architecture packet. The local docs stay richly linked, while the remote copy stays easy to read in sequence.

### What was tricky to build
- The main detail here was choosing a bundle order that reads well as a packet rather than as independent markdown files. I kept the order as index, design guide, diary, tasks, and changelog so the narrative moves from orientation to deep design to process record to tracking metadata.

### What warrants a second pair of eyes
- Whether future research tickets should bundle only the design guide and diary, or continue bundling tasks/changelog for operational completeness.

### What should be done in the future
- If implementation begins, create a follow-on execution ticket rather than continuing to grow this research ticket.
- If the GStreamer binding choice becomes clear, append a focused addendum documenting the exact library/runtime decision.

### Technical details
- Commands run:

```bash
docmgr doc relate --doc ttmp/2026/04/10/SCS-0011--gstreamer-migration-architecture-and-intern-guide/design-doc/01-gstreamer-migration-analysis-and-implementation-guide.md ...
docmgr doc relate --doc ttmp/2026/04/10/SCS-0011--gstreamer-migration-architecture-and-intern-guide/reference/01-investigation-diary.md ...
docmgr doctor --ticket SCS-0011 --stale-after 30
remarquee upload bundle ttmp/2026/04/10/SCS-0011--gstreamer-migration-architecture-and-intern-guide/index.md \
  ttmp/2026/04/10/SCS-0011--gstreamer-migration-architecture-and-intern-guide/design-doc/01-gstreamer-migration-analysis-and-implementation-guide.md \
  ttmp/2026/04/10/SCS-0011--gstreamer-migration-architecture-and-intern-guide/reference/01-investigation-diary.md \
  ttmp/2026/04/10/SCS-0011--gstreamer-migration-architecture-and-intern-guide/tasks.md \
  ttmp/2026/04/10/SCS-0011--gstreamer-migration-architecture-and-intern-guide/changelog.md \
  --dry-run \
  --name "SCS-0011 GStreamer Migration Guide" \
  --remote-dir /ai/2026/04/10/SCS-0011 \
  --toc-depth 2 \
  --layout editor \
  --non-interactive
remarquee upload bundle ttmp/2026/04/10/SCS-0011--gstreamer-migration-architecture-and-intern-guide/index.md \
  ttmp/2026/04/10/SCS-0011--gstreamer-migration-architecture-and-intern-guide/design-doc/01-gstreamer-migration-analysis-and-implementation-guide.md \
  ttmp/2026/04/10/SCS-0011--gstreamer-migration-architecture-and-intern-guide/reference/01-investigation-diary.md \
  ttmp/2026/04/10/SCS-0011--gstreamer-migration-architecture-and-intern-guide/tasks.md \
  ttmp/2026/04/10/SCS-0011--gstreamer-migration-architecture-and-intern-guide/changelog.md \
  --name "SCS-0011 GStreamer Migration Guide" \
  --remote-dir /ai/2026/04/10/SCS-0011 \
  --toc-depth 2 \
  --layout editor \
  --non-interactive
remarquee cloud ls /ai/2026/04/10/SCS-0011 --long --non-interactive
```

### What warrants a second pair of eyes
- Whether the document is detailed enough about the capture graph options for preview/recording sharing.
- Whether the proposed package layout aligns with the project’s preferred naming and layering conventions.

### What should be done in the future
- Once the GStreamer binding choice is made, add a companion implementation playbook to this ticket.

### Code review instructions
- Review the design doc from top to bottom once.
- Then revisit the “Detailed File-Level Analysis” and “Proposed Implementation Phases” sections while cross-checking the referenced files.

### Technical details
- Primary output:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/10/SCS-0011--gstreamer-migration-architecture-and-intern-guide/design-doc/01-gstreamer-migration-analysis-and-implementation-guide.md`

## Quick Reference

### Current architecture in one view

```text
Browser UI
  -> HTTP API + websocket
  -> internal/web/*
  -> ApplicationService
  -> DSL normalize/compile
  -> FFmpeg recording subprocesses
  -> FFmpeg preview subprocesses
```

### Proposed architecture in one view

```text
Browser UI
  -> HTTP API + websocket
  -> internal/web/*
  -> ApplicationService
  -> DSL normalize/compile
  -> media runtime seam
      -> ffmpeg implementation (migration phase)
      -> gstreamer implementation (target)
```

### Key files

- [application.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go)
- [compile.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/dsl/compile.go)
- [ffmpeg.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/ffmpeg.go)
- [run.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go)
- [preview_runner.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_runner.go)
- [preview_manager.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go)
- [session_manager.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/session_manager.go)
- [server.go](/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go)
- [web.proto](/home/manuel/code/wesen/2026-04-09--screencast-studio/proto/screencast/studio/v1/web.proto)

## Usage Examples

### When planning implementation

Start in the design doc’s “Intern Implementation Guide” section, then read the files in the listed order.

### When reviewing proposed changes

Use the design doc’s “Migration Goals And Non-Negotiable Invariants” section as a checklist to decide whether a patch preserves the correct browser and DSL contracts.

## Related

- [01-gstreamer-migration-analysis-and-implementation-guide.md](../design-doc/01-gstreamer-migration-analysis-and-implementation-guide.md)
