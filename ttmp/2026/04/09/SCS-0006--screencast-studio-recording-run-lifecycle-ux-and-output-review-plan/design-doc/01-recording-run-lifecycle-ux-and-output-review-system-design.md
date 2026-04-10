---
Title: Recording Run Lifecycle UX and Output Review System Design
Ticket: SCS-0006
Status: active
Topics:
    - frontend
    - backend
    - ui
    - video
    - product
    - recording
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/session_manager.go
      Note: |-
        Backend recording session state source
        Backend session-state source for lifecycle UX
    - Path: pkg/recording/events.go
      Note: |-
        Runtime event model that informs session transitions
        Runtime event model that may require product-facing enrichment
    - Path: proto/screencast/studio/v1/web.proto
      Note: |-
        Shared session and event contract that may require extension for richer lifecycle UX
        Shared contract for lifecycle and output-summary transport
    - Path: ui/src/components/log-panel/LogPanel.tsx
      Note: |-
        Existing log surface that should become diagnostic support instead of the only source of truth
        Log panel that should become diagnostic support rather than primary lifecycle UX
    - Path: ui/src/components/studio/OutputPanel.tsx
      Note: |-
        Current transport control surface that needs better state communication
        Transport controls that need richer run lifecycle UX
    - Path: ui/src/components/studio/StatusPanel.tsx
      Note: |-
        Existing status panel that should become part of a clearer lifecycle display
        Status panel that should better communicate run state
    - Path: ui/src/pages/StudioPage.tsx
      Note: |-
        Current mounted page and orchestration owner for run lifecycle UX
        Mounted page that owns current recording lifecycle rendering
ExternalSources: []
Summary: Detailed design and implementation guide for turning recording run lifecycle and output review into a strong product experience.
LastUpdated: 2026-04-09T21:18:00-04:00
WhatFor: Guide an intern through improving start/stop/failure UX, output summaries, and user-facing run-state communication.
WhenToUse: Read before implementing lifecycle messaging, post-run output review, or richer run-state presentation.
---


# Recording Run Lifecycle UX and Output Review System Design

## Executive Summary

The application can already record, stop, stream logs, and expose current session state. That is necessary, but not sufficient for a finished product. The missing layer is user-facing lifecycle clarity. A user needs to understand when the system is validating, when it is actually recording, when it is stopping, why it failed, and what outputs were produced when it finishes.

This ticket productizes that experience. It treats logs as a diagnostic layer, not the main UX. It adds explicit lifecycle states, richer finished and failed states, and a better post-run output review surface.

## Problem Statement

### Current State

- The Record and Stop controls work.
- Session state exists.
- Logs stream.
- The app does not yet present a polished lifecycle narrative to the user.
- The app does not yet provide a strong post-run output review experience.

### Why This Matters

Recording software is judged heavily on operational trust:

- did recording really start?
- is it still running?
- did it stop cleanly?
- if it failed, why?
- where are the output files?

If those questions are not answered clearly in the UI, the system still feels fragile even when the backend is technically working.

## Proposed Solution

Define an explicit product lifecycle model and make both backend state and frontend rendering support it directly.

The UI should render:

- clear start/stop progress
- visible success and failure summaries
- a stable output summary after completion
- logs as secondary diagnostics

## Detailed Design

## 1. Product Lifecycle States

The product lifecycle should be explicit.

Recommended states:

- `idle`
- `validating`
- `starting`
- `recording`
- `stopping`
- `finished`
- `failed`

These do not have to map one-to-one to every internal runtime state, but they must map clearly enough that the frontend can render them without guessing.

### State Diagram

```text
idle
  -> validating
  -> starting
  -> recording
  -> stopping
  -> finished

recording
  -> failed

starting
  -> failed

stopping
  -> failed
```

## 2. UI Expectations Per State

### Idle

- Record button enabled
- last successful output summary still visible if useful
- no ambiguous “busy” indicators

### Validating

- user sees that configuration is being checked
- record button disabled

### Starting

- explicit “starting recording” message
- user should not have to infer this from the timer alone

### Recording

- timer active
- running state clearly visible
- status and maybe outputs-in-progress visible if useful

### Stopping

- explicit stopping state
- buttons reflect that the transition is in progress

### Finished

- success message
- output summary with paths
- warnings surfaced if relevant

### Failed

- clear failure summary
- recommended next place to look
- raw logs still available

## 3. Output Review

Finished runs should leave behind a structured output review surface.

### Recommended Data

- output kind
- source name
- source id
- final path
- maybe size if available
- maybe warning status if output was partial

### Recommended UI

A dedicated completion panel or output summary card visible after stop:

```text
Recording Complete
  Desktop        /path/to/Desktop.mov
  Audio Mix      /path/to/audio-mix.wav
```

Potential actions:

- copy path
- open containing directory if feasible
- collapse or expand details

## 4. Failure UX

The current logs are useful, but the product should not require users to parse logs to understand basic failure.

### Recommended Failure Model

Show:

- short failure title
- structured reason
- any affected output information if known
- link or affordance to inspect logs

Pseudocode:

```ts
if (session.state === 'failed') {
  renderFailureSummary({
    reason: session.reason,
    warnings: session.warnings,
    logsAvailable: session.logs.length > 0,
  });
}
```

## 5. Relationship To Logs

Logs should remain available, but should not be the primary UX layer for normal lifecycle communication.

### Recommended Rule

- session state explains what happened
- logs explain details

That separation will make the app easier for non-engineering users to trust.

## 6. Backend Event And Session Mapping

The backend already has runtime events and a session manager. This ticket should refine the mapping from low-level runtime behavior into user-facing state.

Important questions:

- Is validation represented clearly?
- Is “starting” visible distinctly from “running”?
- Are failure reasons structured enough?
- Are outputs and warnings stable after completion?

## 7. API References

Relevant files:

- mounted page: `ui/src/pages/StudioPage.tsx`
- output panel: `ui/src/components/studio/OutputPanel.tsx`
- log panel: `ui/src/components/log-panel/LogPanel.tsx`
- status panel: `ui/src/components/studio/StatusPanel.tsx`
- backend session source: `internal/web/session_manager.go`
- runtime events: `pkg/recording/events.go`
- transport schema: `proto/screencast/studio/v1/web.proto`

## Design Decisions

### Decision 1: Logs Are Secondary To Structured Lifecycle State

Rationale:

- better user comprehension
- less dependence on raw process logs for ordinary success/failure understanding

### Decision 2: Finished Runs Need A Persistent Output Summary

Rationale:

- users care about the result, not only the transition
- output discovery after recording should not require reading logs

### Decision 3: Failure Should Be Summarized, Not Only Streamed

Rationale:

- real users do not want to reverse-engineer errors from stderr lines
- clear summaries improve trust and debuggability

## Alternatives Considered

### Alternative 1: Keep The Current Minimal Lifecycle UI

Rejected because:

- it leaves the app feeling unfinished
- output review remains too weak
- failures remain too log-centric

### Alternative 2: Push Everything Into The Log Panel

Rejected because:

- that keeps lifecycle understanding too technical
- it burdens the user with diagnostic material even for ordinary success

## Implementation Plan

### Step 1: Freeze The Lifecycle Model

- define explicit user-facing states
- map current backend state to desired UI states

### Step 2: Extend Shared Contract If Needed

- protobuf updates
- enriched session data or lifecycle events

### Step 3: Backend Session Enrichment

- improve state mapping
- capture stable finished/failed summaries

### Step 4: Frontend Lifecycle Rendering

- better button states
- better state banners/messages
- failure summary

### Step 5: Output Review Surface

- finished-run summary
- output path actions

### Step 6: Validate

- tests
- stories
- smoke walkthroughs for success and failure

## Open Questions

- Should validation be a first-class session state or only a transient local UI state?
- Should output-size reporting be included if the backend can provide it cheaply?
- Should the app retain the last completed run summary after a new run starts, or replace it immediately?

## Intern Notes

- Do not force users into logs for basic lifecycle understanding.
- Keep logs available, but secondary.
- Make success and failure both explicit.
- Make output review a first-class part of the product, not an afterthought.
