# Tasks

## Goal

Prevent preview subprocesses from competing with recording subprocesses for the same camera device by suspending previews before recording starts and restoring them after the recording session finishes.

## Documentation

- [x] Create a new ticket workspace for the preview handoff bug.
- [x] Write a small implementation guide describing the current failure mode and the handoff approach.
- [x] Keep an implementation diary with the final code changes, tests, and commit hashes.

## Implementation

- [x] Add a preview-manager API to suspend currently active previews and wait for them to exit.
- [x] Add a preview-manager API to restore a suspended set of previews from the recording DSL.
- [x] Wire recording start to suspend previews before the recording manager starts the session.
- [x] Wire recording completion to restore the suspended previews after the session fully finishes.
- [x] Ensure failed recording starts restore any previews that were suspended before the start attempt.
- [x] Add structured logs for preview suspend/restore and recording handoff lifecycle.

## Validation

- [x] Add focused backend tests for preview suspend/restore behavior.
- [x] Add a recording lifecycle test that proves previews are removed before recording and restored after recording stops.
- [x] Run Go tests for the touched packages.

## Completion Definition

- [x] Starting a recording while previews are active no longer leaves preview workers holding camera devices.
- [x] Preview workers are restored automatically after the recording session finishes.
- [x] The ticket docs and diary match the final implementation.
