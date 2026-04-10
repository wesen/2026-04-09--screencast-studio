---
Title: Source Management UX and Setup Builder System Design
Ticket: SCS-0005
Status: active
Topics:
    - frontend
    - backend
    - ui
    - dsl
    - video
    - product
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/discovery/service.go
      Note: |-
        Backend discovery model for displays, windows, cameras, and audio devices
        Discovery model that supplies candidate sources
    - Path: pkg/dsl/types.go
      Note: |-
        DSL source model that the structured setup builder must round-trip to correctly
        DSL source model that the structured builder must round-trip to
    - Path: ui/src/api/discoveryApi.ts
      Note: |-
        Discovery API that should feed the setup builder
        Discovery transport that should drive the source picker
    - Path: ui/src/api/setupApi.ts
      Note: |-
        Normalize and compile APIs that should validate structured edits through the backend
        Normalize and compile transport that should validate structured edits
    - Path: ui/src/components/source-card/SourceCard.tsx
      Note: |-
        Existing source-card component that will likely host or trigger source-edit actions
        Current source-card rendering that needs real source-management actions
    - Path: ui/src/components/studio/SourceGrid.tsx
      Note: |-
        Existing source-grid shell that should become interactive instead of read-only
        Current source-grid shell that should become interactive
    - Path: ui/src/pages/StudioPage.tsx
      Note: |-
        Mounted page that currently derives sources from normalized DSL and should become the setup-builder owner
        Mounted page that currently derives sources from normalized DSL
ExternalSources: []
Summary: Detailed design and implementation guide for building a structured source-management and setup-builder UI on top of discovery and DSL normalization.
LastUpdated: 2026-04-09T21:10:00-04:00
WhatFor: Guide an intern through implementing a real source-management workflow without drifting from the backend DSL model.
WhenToUse: Read before implementing source add/edit/remove flows, structured setup drafts, or DSL synchronization behavior.
---


# Source Management UX and Setup Builder System Design

## Executive Summary

The current mounted UI can show real sources and real previews, but it still treats source definition as something that mostly comes from raw DSL. That is acceptable for engineering validation, but it is not enough for a finished product. A user should be able to construct a setup through the UI: choose discovered displays or windows, add a camera, define a region, rename a source, disable or remove it, and understand the setup they are about to record without having to write YAML.

This ticket introduces a structured setup builder layered on top of the existing DSL and normalization model. The goal is not to delete raw DSL. The goal is to make raw DSL the advanced surface and give the product a real UI for ordinary source-management work.

## Problem Statement

### Current State

- `StudioPage` derives sources from normalized DSL and renders them.
- `SourceGrid` and `SourceCard` look product-like.
- Previews now work for displayed sources.
- The mounted app does not let the user add or remove sources meaningfully.
- The mounted app does not expose discovery-driven source creation.
- The mounted app does not give the user a structured source editor.

### Why This Matters

The capture setup is the heart of the application. If source management lives only in raw DSL, the app remains an advanced prototype. The UI needs a structured path for the common operations:

- add a display
- add a window
- add a camera
- add a region
- rename sources
- enable or disable sources
- reorder sources
- remove sources

### Risk If Left Unfixed

- Raw DSL remains the only real setup editor.
- The source grid remains a viewer instead of a setup tool.
- Future product work gets funneled back into YAML editing rather than structured flows.

## Proposed Solution

Introduce a structured setup-draft model in the frontend that can:

- be created from discovery results
- be edited through UI controls
- be serialized to canonical DSL
- be validated by the backend normalize and compile APIs

The structured editor and raw DSL tab should coexist, but the structured editor should be the default product path.

## Detailed Design

## 1. Source Types And Editable Fields

The builder must explicitly understand source kinds.

### Display Source

Editable fields:

- source id
- display target
- source name
- enabled flag
- capture defaults overrides if exposed

### Window Source

Editable fields:

- source id
- target display if needed
- target window id
- source name
- enabled flag

### Region Source

Editable fields:

- source id
- target display
- region rectangle
- source name
- enabled flag

### Camera Source

Editable fields:

- source id
- device
- source name
- enabled flag
- camera capture overrides if exposed

## 2. Structured Draft Model

The UI should not directly manipulate protobuf messages or normalized backend config as its editable state. It needs a dedicated draft model.

### Recommended Shape

```ts
type SetupDraft = {
  sessionId: string;
  destinationTemplates: Record<string, string>;
  videoSources: SetupDraftVideoSource[];
  audioSources: SetupDraftAudioSource[];
};
```

Each source should carry exactly the fields needed for editing plus any derived UI metadata required for display.

### Important Rule

The draft model must round-trip to DSL without silent data loss for the supported v1 feature set.

## 3. Discovery-Driven Creation

Source creation should start from real backend discovery data.

Flow diagram:

```text
user clicks Add Source
  -> choose source kind
  -> show discovered candidates
  -> choose candidate
  -> create setup draft source
  -> serialize to DSL
  -> normalize
  -> render source card + preview
```

### Region Sources

Regions are the awkward case because they are not discovered as a native object.

Recommended v1:

- pick a display
- choose preset region or enter numeric rectangle
- create a region source draft object

Do not block the whole builder on a future visual region-drag selector.

## 4. Source Editing

The source grid can remain the visual surface, but editing probably needs either:

- expandable inline controls
- a side inspector panel
- or a modal editor

For an intern-friendly implementation, a side inspector or modal is safer than trying to cram every field into the card body.

### Recommended Interaction Pattern

```text
SourceGrid shows cards
  -> click Edit on one card
  -> open SourceInspector
       -> edit fields relevant to that kind
       -> save
  -> serialize draft to DSL
  -> normalize
  -> update previews
```

## 5. Raw DSL Synchronization

The raw DSL tab should remain, but it should not be the only authoritative place a user can edit the setup.

### Recommended Rule

- structured editor writes canonical DSL text
- raw editor can still edit DSL directly
- if raw DSL enters a shape unsupported by the structured editor, the structured editor should become partially read-only and explain why

That is better than pretending the structured editor understands every possible advanced DSL variation.

## 6. Preview Integration

Preview ownership must stay coherent while the setup changes.

This means:

- removing a source should release its preview
- renaming a source should not break the preview if the source id remains stable
- changing source ids is dangerous and must be handled explicitly

### Design Rule

Treat source id stability as important. Prefer keeping source IDs stable across ordinary edits.

## 7. Backend Alignment

The structured editor should validate its output by sending DSL back through normalize and compile, not by trusting local frontend rules alone.

Pseudocode:

```ts
const nextDsl = serializeSetupDraft(draft);
const normalized = await normalizeSetup({ dsl: nextDsl }).unwrap();
dispatch(normalizeSucceeded(normalized));
```

This ensures the UI remains grounded in backend truth.

## API References

Relevant current files:

- discovery transport: `ui/src/api/discoveryApi.ts`
- setup validation: `ui/src/api/setupApi.ts`
- mounted page: `ui/src/pages/StudioPage.tsx`
- source grid: `ui/src/components/studio/SourceGrid.tsx`
- source card: `ui/src/components/source-card/SourceCard.tsx`
- DSL model: `pkg/dsl/types.go`
- discovery model: `pkg/discovery/service.go`

## Design Decisions

### Decision 1: Build A Structured Draft Model Rather Than Editing Normalized Backend Data Directly

Rationale:

- normalized data is a validation/result model, not an editing model
- explicit draft state is easier to reason about
- it makes DSL serialization clearer

### Decision 2: Keep Raw DSL As Advanced Mode

Rationale:

- engineering users still need it
- the DSL already exists and matters
- removing it would reduce flexibility

### Decision 3: Discovery Must Drive Source Creation

Rationale:

- avoids invented IDs and device names
- aligns source setup with actual runtime resources
- reduces invalid configuration churn

### Decision 4: Stable Source IDs Matter

Rationale:

- previews, outputs, and logs often key off source ids
- careless id churn will make the app harder to reason about

## Alternatives Considered

### Alternative 1: Keep Raw DSL As The Only Source Editor

Rejected because:

- it leaves the product incomplete
- it limits usability to technical users

### Alternative 2: Build A Completely Separate UI Model And Only Export DSL At Record Time

Rejected because:

- it risks drifting from backend validation
- it hides errors too late in the workflow

### Alternative 3: Only Support Display Sources In The Structured Builder First

Rejected because:

- the app already visibly suggests broader source support
- partial structured support would still push users back to DSL for ordinary flows

## Implementation Plan

### Step 1: Model The Draft

- define structured setup draft types
- implement serialize and hydrate helpers

### Step 2: Build Source Creation

- source picker
- discovery-backed add flow
- region creation flow

### Step 3: Build Source Editing

- rename
- retarget
- enable/disable
- remove
- reorder

### Step 4: Synchronize With Raw DSL

- canonical serialization
- unsupported-shape handling
- validation feedback

### Step 5: Integrate Previews

- re-ensure/release behavior across edits
- stable id rules

### Step 6: Validate

- tests
- stories
- smoke walkthrough

## Open Questions

- Do we want a side inspector or modal editor for sources in v1?
- How much capture/output override editing belongs in the structured builder versus advanced raw DSL?
- Should audio sources be edited in the same builder surface or a dedicated audio section?

## Intern Notes

- Do not invent source data in the frontend when discovery can provide it.
- Keep raw DSL and structured mode honest with each other.
- Prefer stable source IDs.
- Re-run normalize after structured edits instead of trusting local assumptions.
