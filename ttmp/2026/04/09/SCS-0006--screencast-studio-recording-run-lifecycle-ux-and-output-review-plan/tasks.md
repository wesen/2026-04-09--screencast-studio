# Tasks

## Goal

Turn the recording run lifecycle into a clear finished-product workflow, with strong state transitions, better feedback during start/stop/failure, and rich output review once a run completes.

## Phase 1: Freeze The Current Lifecycle

- [ ] Audit the current recording UX in `StudioPage`, `OutputPanel`, `StatusPanel`, and `LogPanel`.
- [ ] Record the exact current lifecycle states visible to the user.
- [ ] Record what happens today for:
  - start
  - running
  - stop
  - failure
  - completed output review
- [ ] Identify which parts are currently only visible through logs.

Acceptance criteria:

- The intern can describe the current user-visible lifecycle precisely.
- The missing feedback surfaces are documented clearly.

## Phase 2: Define The Product Lifecycle Model

- [ ] Define explicit product-facing run states:
  - idle
  - validating
  - starting
  - recording
  - stopping
  - finished
  - failed
- [ ] Map each state to required UI affordances and messages.
- [ ] Define what summary data the user should see after a run ends.

Acceptance criteria:

- The lifecycle model is explicit and shared between backend and frontend.
- The intern knows exactly what the user should see in each state.

## Phase 3: Extend The Event And Session Model

- [ ] Add protobuf fields or events as needed for richer lifecycle UX.
- [ ] Decide whether output summaries belong inside session state, separate events, or both.
- [ ] Ensure failure reasons and warnings are structured enough for the UI to present clearly.

Acceptance criteria:

- The shared contract can express the product lifecycle, not just low-level logs.

## Phase 4: Backend Lifecycle Mapping

- [ ] Audit `session_manager.go` and runtime events for missing product-facing transitions.
- [ ] Add or refine mapping from runtime events into UI-facing session states.
- [ ] Ensure failure reasons and output summaries are captured consistently.
- [ ] Keep goroutine and state ownership explicit.

Acceptance criteria:

- Backend session state is rich enough to drive the intended lifecycle UI.

## Phase 5: Frontend Run Lifecycle UX

- [ ] Refactor the mounted page to render clearer lifecycle states.
- [ ] Improve the transport controls so start and stop transitions are obvious.
- [ ] Add visible validation and starting states before recording is truly running.
- [ ] Surface failure states prominently without forcing the user into raw logs.
- [ ] Decide whether the Logs tab should auto-focus on failure or remain a manual choice.

Acceptance criteria:

- The app clearly communicates what is happening during a run.
- Failures are understandable without reading raw websocket traffic.

## Phase 6: Output Review UX

- [ ] Add a post-run summary surface for outputs.
- [ ] Show:
  - output name
  - output kind
  - final path
  - maybe file size if available
- [ ] Add actions like:
  - copy path
  - open containing directory if feasible
  - reveal output list from the completed session
- [ ] Keep the output summary visible after finish until the next run starts.

Acceptance criteria:

- A completed run leaves the user with a clear result summary.
- Output files are easy to inspect after recording.

## Phase 7: Log Panel And Diagnostics Refinement

- [ ] Decide how the log panel should support the lifecycle instead of competing with it.
- [ ] Add filtering or highlighting for warnings and errors if needed.
- [ ] Separate user-facing failure summary from raw process logs.

Acceptance criteria:

- Logs become a diagnostic layer, not the only explanation layer.

## Phase 8: Validation And Smoke Tests

- [ ] Add tests covering lifecycle transitions and error mapping.
- [ ] Add frontend tests or Storybook states for:
  - idle
  - starting
  - recording
  - stopping
  - finished with outputs
  - failed with error summary
- [ ] Run a manual smoke test for:
  - successful record
  - stop during active record
  - backend-side failure
  - output summary review

Acceptance criteria:

- The ticket includes a repeatable lifecycle validation recipe.
- The product lifecycle is proven in both success and failure cases.

## Suggested Commit Boundaries

- [ ] Commit 1: lifecycle model and protobuf/event updates
- [ ] Commit 2: backend session-state enrichment
- [ ] Commit 3: frontend lifecycle rendering and transport UX
- [ ] Commit 4: output summary and review affordances
- [ ] Commit 5: log-panel refinement
- [ ] Commit 6: tests, stories, and smoke validation
