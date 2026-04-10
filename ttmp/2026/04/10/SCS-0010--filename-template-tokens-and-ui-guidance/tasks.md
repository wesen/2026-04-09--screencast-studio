# Tasks

## Goal

Let users customize filename-oriented output templates with date/time tokens and an incremental-number token, and explain the supported tokens directly in the mounted UI.

## Documentation

- [x] Create a ticket workspace for filename-template token support and UI guidance.
- [x] Write a small implementation guide covering backend rendering, builder-managed template shape, and inline UI help.
- [x] Keep an implementation diary with the final code changes, tests, and commit hashes.

## Implementation

- [x] Add backend support for an incremental-number filename token.
- [x] Preserve the existing date/time filename tokens in planned-output rendering.
- [x] Extend the builder-managed template helpers so the mounted UI can edit a filename suffix pattern without leaving structured mode.
- [x] Add a filename-template input to the output panel.
- [x] Show concise token documentation in the UI near the output naming controls.
- [x] Keep planned outputs wired to the rendered template result so the user can see the final file names.

## Validation

- [x] Add backend tests for incremental filename rendering.
- [x] Run Go tests and relevant frontend validation/build commands.

## Completion Definition

- [x] The backend supports `{index}` in destination templates and resolves it to the first free file number.
- [x] Structured mode supports the common filename-template shape for both per-source and mixed-audio outputs.
- [x] The UI explains available filename tokens clearly enough that a user can discover the feature without opening Raw DSL.
- [x] The ticket docs and diary match the implemented behavior.
