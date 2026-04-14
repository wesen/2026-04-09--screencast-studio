---
Title: Preview limit race and stale preview recovery bug report
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - video
    - backend
    - ui
    - analysis
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/preview_manager.go
      Note: |-
        Preview limit check runs before asynchronous release cleanup has removed old previews from the manager maps
        Preview limit is checked before async cleanup of a released preview has completed
    - Path: ui/src/pages/StudioPage.tsx
      Note: |-
        The UI releases and re-ensures previews asynchronously while reacting to rapid source changes
        UI preview replacement logic can overlap release and ensure operations during rapid edits
ExternalSources: []
Summary: Bug report for transient preview-limit failures and stale recovery behavior when the user rapidly changes sources, causing ensure requests to outrun async preview cleanup.
LastUpdated: 2026-04-13T22:00:00-04:00
WhatFor: Explain the observed preview-limit race and propose a safer release/ensure strategy.
WhenToUse: Read when fixing transient preview failures during rapid UI editing or source replacement.
---


# Preview limit race and stale preview recovery bug report

## Executive Summary

During the live session, rapid source changes triggered a transient `preview limit exceeded` warning even though a preview release for the replaced source was already in progress. This is a classic race between asynchronous cleanup and synchronous admission control. The preview manager counts previews using `len(m.byID)` before the soon-to-be-released preview has finished exiting and been removed from its maps, while the UI is already attempting to ensure a replacement preview.

The recommended fix is to make preview replacement more transactional: either reserve capacity for a replacement during release, or make the UI wait for release completion before counting the slot as available.

## Problem Statement

Observed behavior:

- the user changed sources quickly,
- a preview release started,
- a replacement preview ensure followed immediately,
- the backend returned `preview limit exceeded`,
- and the UI wound up in a confusing/stale state.

This is especially damaging in the setup UI because preview replacement is a normal workflow, not an edge case.

## Evidence

### Runtime log evidence

The most revealing sequence from the live logs is:

```text
2026-04-13T21:34:34.693028320-04:00 INF preview release triggered cancellation ...
2026-04-13T21:34:34.693515169-04:00 WRN preview limit exceeded event=preview.ensure.limit_exceeded limit=4 source_id=laptop-camera-...-3
2026-04-13T21:34:34.693792562-04:00 INF detached shared preview consumer ...
2026-04-13T21:34:34.693851990-04:00 INF preview finished ...
2026-04-13T21:34:34.693882355-04:00 INF removing preview from manager maps ...
2026-04-13T21:34:34.707579723-04:00 INF preview ensure requested ...
2026-04-13T21:34:34.707897276-04:00 INF created preview worker ...
```

This shows the new ensure attempt arriving while the old preview still counted against the limit.

### Backend code evidence

In `internal/web/preview_manager.go:123-145`, the ensure path checks:

```go
if len(m.byID) >= m.limit {
    return ErrPreviewLimitExceeded
}
```

This check happens before the asynchronously-cancelled preview has necessarily reached `finishPreview(...)` and been removed from the maps.

### UI behavior evidence

The UI preview sync logic in `ui/src/pages/StudioPage.tsx` releases and ensures previews asynchronously while reacting to desired-source changes. That is not inherently wrong, but it means rapid user edits can easily expose any timing gap between backend release bookkeeping and replacement admission.

## Likely Root Cause

This is a lifecycle-accounting race.

### Backend side

A preview slot is only freed when the old preview has actually finished and been removed from `byID`/`bySignature`. Cancellation is initiated first, cleanup happens later.

### UI side

The UI treats preview replacement as a fast asynchronous process and can ask for the new preview before the old one has completed teardown.

### Result

Both sides are individually reasonable, but the contract between them is incomplete.

## Fixing Analysis

There are several possible fixes.

### Option A: backend reserves replacement capacity during release

This is the strongest backend-side fix.

The backend would understand “replacement of one preview by another” as a single logical transition rather than two unrelated operations.

Pros:

- most robust against future UI timing changes
- keeps admission policy authoritative in one place

Cons:

- more state-machine complexity in the preview manager

### Option B: UI waits for release completion before ensure

This is a reasonable first fix if done carefully.

Pros:

- smaller code change
- easier to reason about initially

Cons:

- still depends on UI discipline
- easier to regress if other code paths ensure previews differently later

### Option C: backend counts only non-stopping previews against the limit

This is tempting, but risky. A stopping preview can still consume resources until it is actually gone.

Recommendation: do not use this as the only fix.

### Recommendation

Use a **hybrid fix**:

1. tighten the UI replacement sequence so it does not immediately stampede the backend, and
2. make the backend replacement path more explicit so limit checks behave predictably during normal source replacement.

## Implementation Plan

### Phase 1: Make the UI replacement flow serialize release before re-ensure

Target file:

- `ui/src/pages/StudioPage.tsx`

Plan:

- when a source is being replaced or detached, wait for the release acknowledgement before ensuring a new preview for the same slot in the common path
- keep the current stale-preview cleanup support, but reduce opportunistic overlap

### Phase 2: Add backend-side replacement-aware admission logic

Target file:

- `internal/web/preview_manager.go`

Plan:

- add an explicit concept of a preview in replacement/teardown transition, or
- add an API path that can release-and-replace atomically for a source mutation

The goal is that ordinary source replacement should never trip a hard preview-limit failure just because cleanup is a few milliseconds behind.

### Phase 3: Add a focused regression test

Target files:

- `internal/web/manager_shutdown_test.go`
- or a new preview-manager test file
- optionally a ticket-local browser/API harness in `ttmp/.../scripts/`

Test scenario:

1. fill preview slots to the limit
2. trigger release of one preview
3. immediately request a replacement preview
4. assert no transient `ErrPreviewLimitExceeded` in the supported replacement path

## Alternatives Considered

### Alternative: increase the preview limit

Rejected. That would mask the race rather than fix it.

### Alternative: ignore the race as a UI-only edge case

Rejected. Rapid source editing is a normal user action in this application.

## Validation Checklist

- replacing one preview with another near the limit does not produce transient `preview limit exceeded`
- stale preview IDs are cleaned up deterministically
- the UI no longer gets stuck with a missing preview after a fast replace operation
- limit enforcement still works for genuine over-capacity requests

## References

- `internal/web/preview_manager.go:123-145`
- `ui/src/pages/StudioPage.tsx` preview ensure/release sync logic
- live logs captured from the running `scs-web-ui` tmux session on 2026-04-13
