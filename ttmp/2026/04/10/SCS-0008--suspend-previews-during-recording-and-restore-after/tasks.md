# Tasks

## Goal

Prevent preview subprocesses from competing with recording subprocesses for the same camera device by suspending previews before recording starts and restoring them after the recording session finishes.

## Documentation

- [x] Create a new ticket workspace for the preview handoff bug.
- [x] Write a small implementation guide describing the current failure mode and the handoff approach.
- [ ] Keep an implementation diary with the final code changes, tests, and commit hashes.

## Implementation

- [ ] Add a preview-manager API to suspend currently active previews and wait for them to exit.
- [ ] Add a preview-manager API to restore a suspended set of previews from the recording DSL.
- [ ] Wire recording start to suspend previews before the recording manager starts the session.
- [ ] Wire recording completion to restore the suspended previews after the session fully finishes.
- [ ] Ensure failed recording starts restore any previews that were suspended before the start attempt.
- [ ] Add structured logs for preview suspend/restore and recording handoff lifecycle.

## Validation

- [ ] Add focused backend tests for preview suspend/restore behavior.
- [ ] Add a recording lifecycle test that proves previews are removed before recording and restored after recording stops.
- [ ] Run Go tests for the touched packages.

## Completion Definition

- [ ] Starting a recording while previews are active no longer leaves preview workers holding camera devices.
- [ ] Preview workers are restored automatically after the recording session finishes.
- [ ] The ticket docs and diary match the final implementation.
