---
Title: Preview and recording performance improvement diary
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - performance
    - preview
    - recording
    - diary
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go
      Note: Main implementation site for the preview-branch mitigation and dynamic profile rebalance
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_recording_bridge.go
      Note: Raw consumer lifecycle now triggers preview-profile rebalance
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_test.go
      Note: Added tests for constrained preview behavior while recording
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/main.go
      Note: Benchmark harness extended to model the adaptive mitigation scenario
    - Path: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/run.sh
      Note: Runner updated to execute the added adaptive scenario
ExternalSources: []
Summary: Diary of the analysis and mitigation work for the preview-plus-recording performance spike.
LastUpdated: 2026-04-14T12:00:00-04:00
WhatFor: Record the reasoning, implementation choices, validation attempts, and limitations of this workspace snapshot.
WhenToUse: Read when reviewing or continuing the preview-plus-recording performance work.
---

# Diary

## Goal

Capture the analysis and mitigation work for the shared-source preview-plus-recording CPU spike, including the reasoning behind the chosen fix slice, the exact files changed, and the validation limits of this zip-based workspace.

## Step 1: Reconstructed the current performance story from the handoff and the code

I started by treating the attached research brief as the source of truth for the measurement history instead of trying to infer the problem from code alone. That mattered because the brief already narrowed the real unknown: recorder-only cost had become broadly reconciled, while the combined preview-plus-recorder case remained the strongest open spike.

After that, I read the shared-source runtime and benchmark code side by side. The important outcome of this step was a mental model of where work happens relative to the `tee`, how preview leaves GStreamer through `appsink`, and why “cheap preview” had helped only partially in the earlier experiments.

### Prompt Context

**User prompt (verbatim):** "Analyze and improve the streaming performance of the preview + recording setup.

Make an in-depth report that explains the fundamentals of everything involved, not jus tthe code.

I want an indepth diary of your work (according to the attached skill and diary guidelines)"

**Assistant interpretation:** Analyze the preview-plus-recording performance issue, make a real improvement in the uploaded workspace, and document both the technical explanation and the implementation journey in a diary that follows the supplied structure.

**Inferred user intent:** Get both engineering progress and durable documentation: a reasoned mitigation, a fundamentals-heavy report, and a continuation-friendly diary that preserves decisions, failures, and review instructions.

### What I did

- Read the attached handoff brief and extracted the key benchmark conclusions.
- Read the attached diary guidelines and the attached diary skill file to lock down the required diary format.
- Inspected these implementation files closely:
  - `pkg/media/gst/shared_video.go`
  - `pkg/media/gst/shared_video_recording_bridge.go`
  - `pkg/media/gst/preview.go`
  - `pkg/media/gst/recording.go`
  - `internal/web/preview_manager.go`
- Inspected the standalone interplay benchmark at `scripts/12-go-preview-recorder-interplay-matrix/main.go`.
- Confirmed that the preview branch was ordered as `videoscale -> videorate -> jpegenc` and that preview did not dynamically change profile when recording became active.

### Why

I needed to answer a very specific question before changing code: is the best next move recorder-side, preview-side, or benchmark-side? The handoff evidence strongly pushed the answer toward preview-side mitigation in the combined case.

### What worked

- The handoff brief made the benchmark history clear enough that I did not need to reinvent the investigative baseline.
- The code inspection lined up with the benchmark interpretation: preview really is a second active consumer on the shared source, and it really does perform JPEG work plus Go-facing appsink handling.
- The interplay benchmark mirrored the production shape closely enough to guide a mitigation choice.

### What didn't work

I could not rely on runtime experimentation in this environment at the outset because the uploaded repo is a zip snapshot rather than a checked-out working tree with a known-good local Go toolchain and module cache.

### What I learned

- The most useful distinction is not “preview cost versus recorder cost.” It is “recorder-only cost versus preview-plus-recorder interaction cost.”
- Preview profile changes alone will only go so far if the branch is still doing expensive work before frame dropping.
- The order of `videorate` and `videoscale` is not a cosmetic detail; it determines whether discarded preview frames are scaled first.

### What was tricky to build

The hard part in this step was resisting the temptation to optimize the most visibly expensive thing, which is `x264`. The benchmark history already showed that `x264` is expensive, but it also showed that recorder-only cost had become comparatively well understood. The real target had to be something specific to the combined case.

### What warrants a second pair of eyes

- Whether my reading of the benchmark hierarchy is too preview-centric and underestimates remaining recorder-side opportunities.
- Whether there are GStreamer negotiation subtleties around branch ordering that are not obvious from static code inspection alone.

### What should be done in the future

- Re-run the updated interplay benchmark on the real machine.
- If the adaptive mitigation only modestly helps, move next to shared-pre-tee transforms or recorder-derived preview.

### Code review instructions

Start with the attached handoff brief, then read these files in this order:

1. `pkg/media/gst/shared_video.go`
2. `pkg/media/gst/shared_video_recording_bridge.go`
3. `scripts/12-go-preview-recorder-interplay-matrix/main.go`

### Technical details

Useful inspection commands from this step:

```bash
cd /tmp/screencast
rg -n "buildSharedPreviewConsumer|previewProfileForSource|attachRawConsumer|videorate|videoscale|jpegenc" pkg/media/gst -S
sed -n '430,760p' pkg/media/gst/shared_video.go
sed -n '293,760p' pkg/media/gst/shared_video_recording_bridge.go
sed -n '1,320p' ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/main.go
```

## Step 2: Chose the mitigation slice — dynamic preview constraining plus rate-first preview ordering

Once the performance story was reconstructed, I deliberately chose a mitigation that was narrow enough to be safe but still meaningful. I did not want to redesign the whole capture architecture inside a zip snapshot with limited validation. The selected slice was: keep the architecture, but make preview cheaper while recording is active and ensure preview drops frames before scaling them.

This step was mainly about saying no to bigger but riskier changes. I did not remove the `appsink -> Go -> appsrc` bridge, I did not change the capture correctness path, and I did not turn preview into screenshot-only mode. I chose the strongest improvement that fit inside the already-known constraints.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Make a real improvement now, but keep the result reviewable and continuation-friendly rather than speculative.

**Inferred user intent:** Get a practical mitigation, not a vague research memo.

### What I did

- Compared several candidate fixes:
  - changing encoder preset,
  - dynamically degrading preview during recording,
  - moving preview frame dropping earlier,
  - removing the second preview branch entirely,
  - reducing Go-side preview copies.
- Selected two linked implementation moves:
  - dynamic preview profile rebalance when recorder raw consumers attach or detach,
  - preview branch reorder from `videoscale -> videorate` to `videorate -> videoscale`.
- Defined constrained preview targets:
  - screen-like sources: `640px max width`, `4 fps`, `jpeg quality 50`
  - camera sources: `960px max width`, `6 fps`, `jpeg quality 70`

### Why

This slice attacks two credible sources of waste while preserving the current system contracts:

- preview should remain live while recording, and
- recording should continue to use the bridge path that avoided preview freezes.

### What worked

- The mitigation aligns tightly with the benchmark evidence instead of guessing at unrelated micro-optimizations.
- The change can be localized mostly to `shared_video.go` plus raw consumer lifecycle hooks.
- The benchmark harness could be extended to model the same mitigation strategy.

### What didn't work

A more ambitious “derive preview from recorder frames” design would likely be more powerful, but it would have forced a larger architectural jump than this workspace could responsibly validate.

### What I learned

- Good performance work often starts by moving work, not just deleting work.
- In streaming systems, the order of operations is often as important as the individual operations.
- Dynamic degradation during contention is a product feature as much as a media-pipeline optimization.

### What was tricky to build

The tricky part was choosing numbers for the constrained profile. There is no mathematically perfect answer because the right numbers are partly UX decisions. I therefore chose values that are clearly cheaper than the current profile while still preserving a usable live preview, especially for screen sources where the CPU spike was already measured.

### What warrants a second pair of eyes

- Whether the constrained camera profile should be milder or harsher.
- Whether profile switching should be a product-configurable contract instead of a hardcoded runtime behavior.

### What should be done in the future

- Confirm that the constrained preview remains acceptable in the real web UI during active recording.
- If needed, make the constrained profile configurable per source type.

### Code review instructions

Review the profile rules and the branch ordering in `pkg/media/gst/shared_video.go`, then compare them against the updated benchmark harness in `scripts/12-go-preview-recorder-interplay-matrix/main.go`.

### Technical details

Decision notes from this step:

```text
Do not regress preview-freeze fix.
Do not abandon full-root + videocrop region correctness fix.
Reduce preview work without changing recorder architecture.
```

## Step 3: Implemented dynamic preview profile rebalance and reordered the preview branch

This step is where the actual mitigation landed. I changed the shared preview consumer structure so it can remember and reapply its active profile, added source-level logic to compute the desired preview profile based on whether recording is active, and updated raw consumer lifecycle hooks so previews are rebalanced automatically when recording starts or stops on the same shared source.

I also changed the preview branch build order so rate limiting happens before scaling. That is the most concrete pipeline improvement in the patch because it directly prevents work from being done on frames that preview does not intend to keep.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Apply the chosen mitigation to the uploaded codebase rather than stopping at analysis.

**Inferred user intent:** Leave behind changed code, not only a diagnosis.

**Commit (code):** N/A — the workspace arrived as a zip snapshot without `.git` metadata, so I could not produce a real commit hash.

### What I did

- Modified `pkg/media/gst/shared_video.go` to:
  - add dynamic preview profile selection with a recording-constrained mode,
  - extend `sharedPreviewConsumer` so it can store caps elements and the active profile,
  - add `applyProfile(...)` for in-place profile updates,
  - add `syncPreviewProfilesLocked()` on the shared source,
  - reorder preview branch construction to `queue -> videorate -> caps -> videoscale -> caps -> jpegenc -> appsink`.
- Modified `pkg/media/gst/shared_video_recording_bridge.go` so raw consumer attach and detach trigger preview profile rebalance.
- Modified `pkg/media/gst/shared_video_test.go` to add tests for the new constrained preview rules.

### Why

This is the smallest code change that meaningfully addresses preview waste in the combined case while leaving recorder architecture and capture correctness intact.

### What worked

- The mitigation could be implemented entirely inside the shared-source abstraction.
- Existing preview profile tests still made sense, and new tests fit naturally beside them.
- The raw consumer lifecycle was already the correct place to detect “recording is active on this shared source.”

### What didn't work

I could not prove the runtime improvement empirically in this environment after the code change because the live benchmark requires the real X11/GStreamer capture machine and a working Go module environment.

### What I learned

- The shared-source abstraction was already strong enough to host dynamic preview-mode changes.
- Profile rebalance is much simpler if the preview consumer stores direct references to its caps filters and encoder element.
- Moving `videorate` earlier is a very high-leverage structural improvement because it changes the work graph itself.

### What was tricky to build

The sharp edge in this step was runtime reconfiguration. The preview consumer is attached to a live shared pipeline, so I could not treat profile changes like build-time constants anymore. The solution I chose was to retain references to the preview branch caps filters and `jpegenc`, then update those in place when recorder raw consumers attach or detach.

The secondary sharp edge was making sure the behavior applies both to previews that start before recording and to previews that are attached after recording is already active. I handled that by selecting an initial profile before construction and then still rebalancing after attach in case the shared-source state changed during the attach window.

### What warrants a second pair of eyes

- Whether applying caps and JPEG-quality changes while the shared source mutex is held is the best locking choice.
- Whether any GStreamer renegotiation edge cases appear when profile changes happen on a live pipeline.
- Whether the chosen constrained profile values are the right product tradeoff.

### What should be done in the future

- Add an integration test or live harness assertion that preview profile really switches when recording starts and restores when recording stops.
- Consider making the constrained-profile values a declarative config contract instead of hardcoded defaults.

### Code review instructions

Start in `pkg/media/gst/shared_video.go` and inspect these symbols in order:

- `desiredPreviewProfileLocked`
- `syncPreviewProfilesLocked`
- `buildSharedPreviewConsumer`
- `previewProfileForSourceWhileRecording`
- `applyProfile`

Then inspect `pkg/media/gst/shared_video_recording_bridge.go` for the raw attach/detach hooks that now call preview rebalance.

### Technical details

Files changed in this step:

```text
pkg/media/gst/shared_video.go
pkg/media/gst/shared_video_recording_bridge.go
pkg/media/gst/shared_video_test.go
```

Formatting command used:

```bash
gofmt -w pkg/media/gst/shared_video.go pkg/media/gst/shared_video_recording_bridge.go pkg/media/gst/shared_video_test.go
```

## Step 4: Extended the standalone interplay benchmark to model the mitigation

I did not want the code change to exist without a direct measurement path. Because the earlier handoff already identified `scripts/12-go-preview-recorder-interplay-matrix/` as the most relevant benchmark family for the combined case, I updated that harness instead of creating a disconnected new script family.

The benchmark now has an additional adaptive scenario and can model the rate-first preview branch ordering. That gives the next reviewer a concrete way to compare the old combined behavior, a cheap-preview-only variant, and the new mitigation strategy on the real machine.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Make the improvement reviewable and measurable, not just implemented.

**Inferred user intent:** Leave behind a reproducible path for validating whether the mitigation is worth keeping.

**Commit (code):** N/A — the workspace arrived as a zip snapshot without `.git` metadata.

### What I did

- Updated `scripts/12-go-preview-recorder-interplay-matrix/main.go` to:
  - add `RateFirst` to the preview profile shape,
  - add a new `preview-adaptive-plus-recorder` scenario,
  - let the benchmark build either the old preview order or the new rate-first order.
- Updated `scripts/12-go-preview-recorder-interplay-matrix/run.sh` so the runner executes the adaptive scenario as part of the saved matrix.
- Ran `bash -n` on the updated runner script.

### Why

A performance mitigation without a measurement path is fragile. The interplay benchmark was already the best synthetic approximation of the real combined case, so it was the right place to encode the new mitigation.

### What worked

- The harness could be extended without breaking the historical scenarios.
- The new scenario lines up with the actual mitigation strategy rather than inventing a benchmark-only idea.
- `bash -n` passed for the runner script after the edit.

### What didn't work

I could not execute the benchmark itself here because it depends on the real capture environment and a buildable Go toolchain plus modules.

### What I learned

- The benchmark harness is valuable not only for measurement but for preserving architectural intent.
- Adding new scenarios is better than overwriting the old ones because it keeps historical results interpretable.

### What was tricky to build

The tricky part was avoiding a benchmark drift problem. I did not want to silently redefine the meaning of the existing “current” and “cheap” scenarios because that would make earlier saved result directories misleading. The solution was to add a distinct adaptive scenario and keep the older ones intact.

### What warrants a second pair of eyes

- Whether the adaptive benchmark scenario is a faithful enough mirror of the production mitigation.
- Whether the runner output summary should be extended further to flag which scenarios are historical baselines versus mitigation candidates.

### What should be done in the future

- Run the updated matrix on the real machine and archive the result directory beside the existing saved runs.
- If the adaptive scenario wins clearly, add a new human-written summary note to interpret the result.

### Code review instructions

Review these files together:

- `ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/main.go`
- `ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/run.sh`

Then validate shell syntax with:

```bash
bash -n ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/run.sh
```

### Technical details

New benchmark scenario name:

```text
preview-adaptive-plus-recorder
```

The scenario models both:

```text
- recording-constrained preview dimensions / fps / jpeg quality
- rate-first preview branch ordering
```

## Step 5: Wrote the report, captured validation failures verbatim, and packaged the patch

The last step was documentation and packaging, not as cleanup but as part of the deliverable. The user asked for both a fundamentals-heavy report and an implementation diary, and the attached diary guidance explicitly values exact failures and continuation-friendly review instructions. I therefore captured the environment limitations verbatim instead of smoothing them over.

I also generated a unified patch against the original uploaded zip snapshot so the modified workspace can be reviewed without diffing by hand.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Finish the work in a form that another engineer can review and continue from immediately.

**Inferred user intent:** Receive not only code changes but also a durable explanation of what was done, why it was done, and how to validate it later.

### What I did

- Wrote `reference/04-preview-and-recording-performance-improvement-report.md`.
- Wrote this diary file.
- Generated a unified patch against a freshly re-unzipped copy of the original workspace.
- Captured the `go test` validation failure exactly.
- Confirmed `gofmt` success on the modified Go files and `bash -n` success on the modified shell script.

### Why

The report explains the media-system fundamentals. The diary preserves the work process. The patch gives a compact review artifact. Together, those three outputs satisfy the user’s request and make continuation materially easier.

### What worked

- The patch generated cleanly.
- The report and diary could be integrated into the ticket workspace itself.
- The validation failure was concrete and explainable rather than ambiguous.

### What didn't work

Exact failed validation commands:

```bash
cd /tmp/screencast && go test ./... -count=1
```

Result:

```text
go: downloading go1.25.5 (linux/amd64)
go: download go1.25.5: golang.org/toolchain@v0.0.1-go1.25.5.linux-amd64: Get "https://proxy.golang.org/golang.org/toolchain/@v/v0.0.1-go1.25.5.linux-amd64.zip": dial tcp: lookup proxy.golang.org ... connection refused
```

Then, in a throwaway copy with the version line relaxed only for diagnostics:

```bash
GOTOOLCHAIN=local go test ./pkg/media/gst -run 'TestPreview' -count=1
```

Result:

```text
pkg/media/gst/bus.go:6:2: missing go.sum entry for module providing package github.com/go-gst/go-glib/glib ...
pkg/media/gst/bus.go:7:2: missing go.sum entry for module providing package github.com/go-gst/go-gst/gst ...
pkg/media/gst/shared_video.go:13:2: missing go.sum entry for module providing package github.com/go-gst/go-gst/gst/app ...
pkg/media/gst/preview.go:12:2: missing go.sum entry for module providing package github.com/pkg/errors ...
pkg/media/gst/preview.go:13:2: missing go.sum entry for module providing package github.com/rs/zerolog/log ...
pkg/dsl/load.go:9:2: missing go.sum entry for module providing package gopkg.in/yaml.v3 ...
```

### What I learned

- In a zip-based offline workspace, “validation failed” needs to be broken down carefully into toolchain failure versus dependency-resolution failure.
- `gofmt` and patch generation are still useful partial checks when full build validation is unavailable.

### What was tricky to build

The most important process difficulty here was not technical but evidentiary: I needed to avoid overstating confidence. The code is reasoned and formatted, but it is not fully runtime-validated in this environment. The right move was to preserve the exact failure chain so the next engineer knows the remaining uncertainty is environmental, not hand-waved away.

### What warrants a second pair of eyes

- The entire patch should be reviewed with the assumption that runtime validation still needs to happen on the real machine.
- The report’s performance claims should be treated as “supported by the prior benchmark history plus structural reasoning” until the updated adaptive scenario is actually measured.

### What should be done in the future

- Run the updated benchmark matrix on the target machine.
- If the adaptive scenario is clearly better, update `changelog.md` and add a follow-up summary note with real numbers.
- If the win is weak, move to a larger design change rather than further micro-tuning this branch order.

### Code review instructions

Primary artifacts to review:

- `reference/04-preview-and-recording-performance-improvement-report.md`
- `reference/05-preview-and-recording-performance-improvement-diary.md`
- `/mnt/data/screencast-performance-improvement.patch` or the equivalent updated workspace snapshot

Validation commands to reproduce my partial checks:

```bash
cd /tmp/screencast
gofmt -w pkg/media/gst/shared_video.go pkg/media/gst/shared_video_recording_bridge.go pkg/media/gst/shared_video_test.go ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/main.go
bash -n ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/run.sh
```

### Technical details

Patch generation command used:

```bash
diff -ruN /tmp/screencast_orig /tmp/screencast > /mnt/data/screencast-performance-improvement.patch || true
```
