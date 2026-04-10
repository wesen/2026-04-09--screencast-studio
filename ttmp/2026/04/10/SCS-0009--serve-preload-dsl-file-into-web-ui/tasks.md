# Tasks

## Goal

Allow `screencast-studio serve` to accept a DSL file and preload that setup into the web UI from first load.

## Documentation

- [x] Create a ticket workspace for serve DSL preload support.
- [x] Write a small implementation guide for CLI preload, API bootstrap, and frontend hydration.
- [ ] Keep an implementation diary with the final code changes, tests, and commit hashes.

## Implementation

- [ ] Add `--file` to `screencast-studio serve`.
- [ ] Load and validate the DSL file at startup so invalid preload files fail fast.
- [ ] Extend the web bootstrap API payload with optional initial DSL metadata.
- [ ] Expose the preloaded DSL from the server configuration.
- [ ] Update the frontend bootstrap path so preload is applied before the first normalize/compile cycle.
- [ ] Ensure the setup draft hydrates from the preloaded DSL instead of from the built-in default.

## Validation

- [ ] Add backend tests for preload bootstrap payloads.
- [ ] Add frontend or store-level tests for preload application if appropriate.
- [ ] Run Go tests and relevant frontend tests/build validation.

## Completion Definition

- [ ] `screencast-studio serve --file path/to/setup.yaml` starts successfully when the file is valid.
- [ ] The first visible DSL/setup state in the web UI comes from the provided file.
- [ ] Invalid preload files cause startup to fail clearly instead of serving a broken preload.
- [ ] The ticket docs and diary match the implemented behavior.
