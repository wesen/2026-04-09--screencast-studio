---
Title: Investigation diary
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - video
    - backend
    - ui
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/preview_manager.go
      Note: Investigated for preview-limit race behavior
    - Path: pkg/discovery/service.go
      Note: |-
        Investigated for duplicate camera discovery behavior
        Investigated for camera duplication and /dev/video node handling
    - Path: pkg/media/gst/shared_video.go
      Note: |-
        Investigated for shared-source signatures and preview JPEG settings
        Investigated for preview quality settings and display/camera shared-source signatures
ExternalSources: []
Summary: Diary of the live bug-triage and documentation work for the preview regressions ticket.
LastUpdated: 2026-04-13T21:55:00-04:00
WhatFor: Record how the preview bug reports were derived from a real running server session and code inspection.
WhenToUse: Read when continuing SCS-0014 and you need the chronology of observations, commands, and conclusions.
---


# Diary

## Goal

Capture the investigation behind the preview regressions observed while manually testing the running screencast-studio web UI.

## Step 1: Converted the live testing session into a bug-fix ticket

The user reported that the webcam stopped coming back after add/remove operations, that the webcam format looked wrong and compressed, and that other things stopped working later in the session. I first checked the live `tmux` session logs instead of guessing. That immediately showed a useful pattern: the failures were not all one bug. There were at least four distinct bug classes in the same session.

The critical decision in this step was to split the observed failures into separate bug reports instead of writing one vague “previews are broken” note. That makes the eventual fix work much more reviewable and prevents one regression from masking another.

### Prompt Context

**User prompt (verbatim):** "ok create a new bug fixing ticket. then add a detailed bug-report document for each bug to the ticket and upload to remarkable."

**Assistant interpretation:** Create a new docmgr ticket for the preview regressions uncovered during manual testing, and write one detailed document per bug rather than one blended summary.

**Inferred user intent:** Turn an ad hoc broken live session into a structured bug-fix backlog with enough detail that each bug can be fixed deliberately and not lost after the interactive debugging session ends.

### What I did

- Captured the `tmux` pane from the running `scs-web-ui` session.
- Counted the retained lines so I knew how much live evidence I had.
- Inspected the screenshot for UI-visible symptoms.
- Queried `/api/discovery` to see the current camera and display snapshot.
- Inspected the code paths for:
  - camera discovery,
  - source creation in the UI,
  - shared preview source signatures,
  - preview JPEG conversion,
  - preview limit accounting,
  - display-source UI messaging.
- Created ticket `SCS-0014` and wrote one bug report per bug.

### What worked

- The live logs were detailed enough to show camera device-node churn (`/dev/video0` to `/dev/video1`), display-source collapse, and preview-limit races.
- The code inspection lined up cleanly with the observed symptoms.
- The screenshot confirmed that the duplicate camera sources were visible to the user, not just a backend detail.

### What didn't work

There was no durable logfile on disk for this run; I only had the current tmux pane buffer. That was enough, but it means this ticket documents a live incident from transient server output rather than from a persisted log file.

### What I learned

- `ListCameras()` currently exposes every `/dev/video*` node directly to the UI, which is too low-level for end users.
- Preview JPEG generation is intentionally low quality today, which explains the compressed webcam look.
- Per-display sources are still knowingly mapped onto the root X11 display path.
- Preview release and preview ensure are asynchronous enough that a fast replacement action can temporarily trip the preview limit.

### What was tricky to build

The tricky part was deciding how to separate symptoms that were happening close together in time:

- webcam disappearing after remove/re-add,
- duplicate logical camera sources,
- output path collisions,
- compressed preview quality,
- display-source collapse,
- and preview-limit failures.

Several of these interact, but they are not the same bug. Splitting them into separate reports made the causal chains much clearer.

### What warrants a second pair of eyes

- Whether camera deduplication should happen in discovery, in the UI source picker, or both.
- Whether preview-quality fixes should be treated as a fast UX patch or as part of a broader preview-configuration redesign.
- Whether per-display capture should be fixed by rect-based normalization or by extending the source model itself.

### What should be done in the future

- Validate the ticket and upload the bug bundle to reMarkable.
- Once code work starts, preserve one commit per bug family where practical.
- Consider adding an automated regression harness for preview replacement/re-add scenarios because these failures are easy to reintroduce.

### Code review instructions

Start with the four design docs under `design-doc/`, then review these code anchors:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/discovery/service.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/conversion.ts`

### Technical details

Useful commands during the investigation included:

```bash
tmux capture-pane -pt scs-web-ui -S -2000
curl -fsS http://127.0.0.1:7777/api/discovery | python3 -m json.tool
rg -n "ListCameras|preview.ensure.limit_exceeded|computeSharedVideoSourceSignature|jpegenc.Set\(\"quality\"|root X11 target" pkg internal ui/src -S
```

## Step 2: Validated the Bug Ticket and Uploaded It to reMarkable

After the four bug reports were written, I treated delivery as part of the ticket rather than an optional extra. The user had asked for the bug reports to be uploaded to reMarkable, so the ticket was only truly finished once the docs validated cleanly and the final bundle was visible in the expected remote folder.

This step therefore closed the packaging loop: I validated the ticket with `docmgr`, ran the dry-run upload, uploaded the real bundle, and confirmed the remote listing. That makes the ticket usable both in the repo and on the reading device.

### Prompt Context

**User prompt (verbatim):** (same as Step 1, plus the follow-up request that each bug report include a fixing analysis and implementation plan section)

**Assistant interpretation:** Finish the bug-fix ticket deliverable by validating the new workspace and delivering the bug reports bundle to reMarkable.

**Inferred user intent:** Ensure the bug reports are durable, reviewable, and easy to consume outside the terminal session where the bugs were originally observed.

### What I did

- Ran `docmgr doctor --ticket SCS-0014 --stale-after 30`
- Confirmed it passed cleanly
- Ran the reMarkable workflow:
  - `remarquee status`
  - `remarquee cloud account --non-interactive`
  - `remarquee upload bundle --dry-run ...`
  - `remarquee upload bundle ...`
  - `remarquee cloud ls /ai/2026/04/13/SCS-0014 --long --non-interactive`
- Updated `tasks.md`, `index.md`, and `changelog.md` to record validation and upload completion

### What worked

- `docmgr doctor` passed with no findings beyond the final all-clear
- the dry-run bundle upload succeeded
- the real upload succeeded
- the remote verification listing showed the uploaded PDF under the requested folder

### What didn't work

Nothing materially failed in this delivery step after the ticket docs were written.

### What I learned

- The bundle format works well for multi-doc bug tickets because it keeps all bug reports in one navigable PDF.
- For bug-fix tickets, including the index plus all bug reports plus the diary creates a useful review package rather than a pile of disconnected markdown files.

### What was tricky to build

The main subtlety was deciding what to bundle. Because this ticket is organized as one document per bug, uploading only one or two docs would have made the deliverable incomplete. Bundling the full set preserves the intended structure: overview first, then bug 1–4, then the diary.

### What warrants a second pair of eyes

- Whether future bug bundles should also include `tasks.md` explicitly, or whether `index.md + design docs + diary` is the better default reading package.

### What should be done in the future

- If implementation begins, update the corresponding bug report documents instead of scattering separate notes elsewhere.
- If one fix changes the root cause analysis for another bug, keep the ticket docs synchronized so the bundle remains trustworthy.

### Code review instructions

Re-run these commands to verify ticket health and delivery:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
docmgr doctor --ticket SCS-0014 --stale-after 30
remarquee cloud ls /ai/2026/04/13/SCS-0014 --long --non-interactive
```

### Technical details

Verified remote artifact:

```text
SCS-0014 Preview Regression Bug Reports
```

Remote folder:

```text
/ai/2026/04/13/SCS-0014
```

## Step 3: Chose the First Fix Slice — Camera Discovery Deduplication + UI Duplicate Prevention

Before editing code, I narrowed the first fix to the smallest slice that addresses the most confusing webcam behavior without dragging in the other preview bugs. The first target is the camera identity problem, not the preview quality bug or the per-display capture bug. That means the immediate implementation work should reduce duplicate webcam choices and stop the UI from generating repeated logical camera sources for the same physical device.

This is the right first slice because it attacks the deepest source of confusion in the session: the user thought they were operating on one laptop camera, while the app exposed multiple raw `/dev/videoN` entries and multiple suffixed logical sources for that same physical hardware.

### Prompt Context

**User prompt (verbatim):** "go with 2. keep a drequent diary as you work."

**Assistant interpretation:** Start implementing the first bug fix now rather than just documenting it, and keep the ticket diary updated as the work proceeds.

**Inferred user intent:** Move from analysis into real remediation while preserving a useful written trail of decisions and intermediate findings.

### What I did

- Re-read the live bug reports and picked a narrow first implementation slice.
- Confirmed the relevant files for that slice:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/discovery/service.go`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/conversion.ts`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`
- Confirmed there were no existing targeted tests for discovery deduplication or setup-draft duplicate-camera behavior, so I would likely need to introduce at least a small new testable helper on the Go side.

### Why

Trying to fix all four preview bugs at once would make it too hard to tell what actually improved the webcam flow. Starting with discovery dedupe and UI duplicate prevention gives a clean, reviewable change that should immediately make the product less confusing even before deeper backend identity fixes are implemented.

### What worked

- The code inspection aligned very cleanly with the bug report: discovery duplicates the raw camera nodes, and the UI freely creates duplicate logical sources.
- The first fix slice stayed small enough to be realistic for one iteration.

### What didn't work

- There were no pre-existing tests in the obvious discovery/setup-draft paths, so I could not just extend an existing test file.

### What should be done next

- Implement camera discovery deduplication in `pkg/discovery/service.go`.
- Add a small Go test for the deduplication helper if possible.
- Prevent duplicate camera source additions in `StudioPage.tsx` using the already-normalized `deviceId`.
- Re-run tests and manually verify the running UI behavior.

## Step 4: Implemented the First Fix Slice and Rebuilt the Running App

I implemented the first bug-fix slice exactly as planned: backend camera discovery now deduplicates raw `/dev/videoN` nodes into one logical camera entry per physical card/label, and the web UI now refuses to add a duplicate camera source if a camera source with the same device or name already exists in the structured setup draft.

This step does not solve every webcam problem from the session, but it does address the most confusing identity issue first. The discovery API now reports one logical laptop camera instead of both `/dev/video0` and `/dev/video1`, and the rebuilt UI contains a guard against repeatedly adding the same camera as `...-2`, `...-3`, and so on.

### Prompt Context

**User prompt (verbatim):** (same as Step 3)

**Assistant interpretation:** Implement the first actual bug fix now, keep the diary current, and validate that the running app picks up the change.

**Inferred user intent:** Make visible progress on the documented bug ticket, not just more analysis, while keeping the work trail understandable.

### What I did

Changed backend discovery:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/discovery/service.go`
  - `ListCameras()` now returns `dedupeCameras(cameras)`
  - added:
    - `dedupeCameras(...)`
    - `cameraIdentityKey(...)`
    - `preferredCameraNode(...)`
    - `videoDeviceIndex(...)`

Added Go test:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/discovery/service_test.go`
  - verifies deduplication prefers the lowest `/dev/videoN` node per physical camera group
  - verifies device index parsing

Changed UI duplicate prevention:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`
  - updated `applyAddedSource(...)`
  - if the new source is a camera and a camera source already exists with the same device or name, the add is skipped and a UI log line is emitted instead of silently creating a duplicate logical source

Validation steps run:

```bash
gofmt -w pkg/discovery/service.go pkg/discovery/service_test.go
go test ./pkg/discovery ./internal/web ./pkg/app ./pkg/media/gst -count=1
go test ./... -count=1
cd ui && pnpm build
cd .. && go run ./cmd/build-web
```

Then I restarted the running tmux app and queried discovery again:

```bash
curl -fsS http://127.0.0.1:7777/api/discovery
```

Observed result:

```text
camera_count= 1
/dev/video0 | Laptop Camera: Laptop Camera (usb-0000:00:14.0-7) | /dev/video0
```

### Why

This implementation was the smallest credible fix that directly changed the user-visible situation:

- the source picker should not offer the same physical laptop camera twice just because V4L2 exposes multiple nodes,
- and the structured editor should not create multiple logical webcam sources by accident for the same camera.

### What worked

- the new Go dedupe helper passed its tests
- the broader Go test suite still passed
- the frontend TypeScript/Vite build passed
- the embedded web assets were rebuilt successfully
- after restarting the app, `/api/discovery` now reports only one camera entry instead of two

### What didn't work

Nothing failed in this implementation slice after the code changes landed. The main limitation is that this is still a first-pass fix: it reduces duplicate webcam identity confusion, but it does not yet address preview quality, per-display capture collapse, or the preview-limit race.

### What I learned

- fixing discovery alone would not have been enough; the UI also needed a duplicate-source guard because users can still repeatedly add the same logical camera from a single discovery entry
- rebuilding the frontend bundle into `internal/web/dist` is necessary for UI changes to show up in the running `go run ./cmd/screencast-studio serve` app
- the backend change is immediately observable through `/api/discovery`, which makes this fix slice nicely testable

### What was tricky to build

The main subtlety was choosing what should count as a duplicate camera in the UI. Device ID alone is not enough if the underlying node can churn between `/dev/video0` and `/dev/video1`, so the guard currently checks both device ID and camera name. That is a pragmatic first pass. It may be slightly too strict for hypothetical setups with two different cameras that present identical labels, but it is the right tradeoff for reducing the current laptop-camera confusion quickly.

### What warrants a second pair of eyes

- Whether the UI duplicate guard should eventually use a stronger physical-camera identity than `(deviceId || name)`.
- Whether deduplication should later use more detailed V4L2 capability filtering rather than just card/label grouping plus lowest-node preference.
- Whether the UI should do more than log a skipped duplicate add (for example, focus the existing source card visually).

### What should be done in the future

- Manually test the updated UI to confirm the duplicate add guard behaves the way we expect in the browser, not just in code.
- Then move to the next bug in SCS-0014, likely the preview-quality regression or the preview-limit race.

### Code review instructions

Start here:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/discovery/service.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/discovery/service_test.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx`

Then validate with:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
go test ./... -count=1
curl -fsS http://127.0.0.1:7777/api/discovery | python3 -m json.tool
```

## Step 5: Chose a Source-Aware Preview Quality Fix Instead of a Blind Global Bump

For the second implementation slice, I focused on the webcam preview quality regression. Before changing code, I re-read the shared preview path and noticed that the quality problem is coming from one very specific place: every preview consumer branch is currently forced through the same low-fidelity settings regardless of source type.

The existing shared preview branch uses:

- `video/x-raw,width=640`
- `video/x-raw,framerate=5/1`
- `jpegenc quality=50`

That explains the pixelated webcam preview almost by itself. A laptop webcam captured at `1280x720` and then previewed through a `640px` JPEG at quality `50` will look degraded even if the underlying capture is fine.

### Prompt Context

**User prompt (verbatim):** "continue"

**Assistant interpretation:** Move on directly to the next bug fix slice after the camera identity fix was confirmed.

**Inferred user intent:** Keep fixing the most obvious real regressions in sequence, while maintaining the diary.

### What I did

- Re-read `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go`
- Confirmed the preview branch hard-codes low fidelity for all source types
- Re-checked the MJPEG HTTP stream path in `/home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go`
- Noticed the HTTP loop already emits frames at about `10 fps` because it ticks every `100ms`
- Re-checked UI defaults in `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/conversion.ts`
  - camera sources default to `1280x720` and `30 fps`

### Decision

Instead of simply cranking every preview to maximum settings, I chose a source-aware preview profile strategy:

- **camera previews** should preserve substantially more detail than `640px`/`quality 50`
- **display/window/region previews** should also improve, but not at an obviously excessive CPU/network cost
- preview frame rate should not exceed the HTTP streaming cadence unnecessarily
- preview scaling should avoid pointless upscaling of smaller sources/windows

### Why

A blind global bump (for example, setting every preview to `1920px`, `30 fps`, `quality 95`) would risk hiding the original bug under a new performance problem. The user-reported regression is specifically about webcam clarity, so the fix should improve clarity intentionally while staying compatible with the existing shared preview architecture.

### What should be done next

- Add a small preview-profile helper in `pkg/media/gst/shared_video.go`
- Make camera previews use a higher width and JPEG quality than generic previews
- Clamp preview frame rate to the effective stream cadence rather than wasting work beyond what the MJPEG endpoint can deliver
- Add a small unit test for the helper if feasible
- Rebuild the web bundle and manually re-test the preview

## Step 6: Implemented a Source-Aware Preview Quality Profile and Validated It Against the Current Shared-Capture Architecture

I implemented the second bug-fix slice in the GStreamer preview branch. The goal was to improve webcam preview clarity without accidentally overdriving every preview path in the system. Instead of a global “make everything huge” change, I introduced a small source-aware preview profile helper that chooses width, frame rate, and JPEG quality based on the source kind and the source’s natural dimensions.

### What changed

Changed preview runtime internals:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go`
  - added `sharedPreviewProfile`
  - added `previewProfileForSource(...)`
  - added `previewFPS(...)`
  - added `previewSourceWidth(...)`
  - changed `buildSharedPreviewConsumer(...)` to use the computed profile instead of hard-coded `640 / 5 fps / quality 50`

Added tests:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_test.go`
  - verifies camera previews use a higher-fidelity profile
  - verifies smaller window previews do not upscale unnecessarily
  - verifies FPS clamping behavior

### New preview behavior

The old shared preview branch forced:

- width `640`
- frame rate `5 fps`
- JPEG quality `50`

The new logic is more intentional:

- **camera**: target width up to `1280`, JPEG quality `85`, FPS clamped to `10`
- **window/region**: target width up to `1280`, JPEG quality `80`, FPS clamped to `10`
- **display**: target width up to `960`, JPEG quality `75`, FPS clamped to `10`
- if the effective source is smaller than the profile width, keep the smaller width instead of upscaling

### Why the FPS clamp is correct

While investigating, I noticed `/api/previews/{id}/mjpeg` only emits frames on a `100ms` ticker, so the transport path is already effectively `10 fps`. That means pushing the preview branch above `10 fps` would mostly waste work. The quality fix therefore focuses on detail preservation (width + JPEG quality) instead of chasing meaningless extra frame rate.

### Validation

I ran:

```bash
gofmt -w pkg/media/gst/shared_video.go pkg/media/gst/shared_video_test.go
go test ./pkg/media/gst ./internal/web ./pkg/discovery -count=1
bash ./ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/16-web-gst-default-runtime-e2e.sh
```

I also tried the older preview harness:

```bash
bash ./ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/11-web-gst-preview-e2e.sh
```

That script failed, but the failure was *expected* once I looked closely at the output: it still asserts the legacy “preview suspended during recording” model and is therefore stale against the current shared-capture architecture. The newer default-runtime harness passed and is the correct validation path now.

Finally, I restarted the real tmux app so the running UI now uses the updated backend preview branch:

```bash
tmux kill-session -t scs-web-ui
tmux new-session -d -s scs-web-ui "cd /home/manuel/code/wesen/2026-04-09--screencast-studio && go run ./cmd/screencast-studio serve --addr :7777"
```

### What worked

- new unit tests passed
- preview-related Go tests stayed green
- the current no-suspend default-runtime E2E harness passed
- the live tmux app restarted successfully and served healthz again

### What didn't work

- the older `scripts/11-web-gst-preview-e2e.sh` harness is no longer architecture-correct because it expects preview suspension during recording

### What I learned

- the biggest webcam regression factor was compression/detail loss, not frame rate alone
- an old harness can become misleading after architecture changes, even if the product behavior is correct
- source-aware preview settings are a better long-term pattern than one shared low-quality profile for every source kind

### What should be done next

- Manually verify the webcam preview visually in the live UI.
- Then move to the next product bug: the preview-limit race / stale preview recovery behavior.

## Step 7: Reproduced the Window/Region Bug and Narrowed It to Preview Scaling, Not Target Resolution

The user clarified the next real problem: selecting a **window** ends up looking like full-screen, and the same is true for **predefined regions**. I switched immediately from the per-monitor investigation to this higher-priority bug because it affects everyday use right now.

### What I did

I reproduced the issue in the real running UI with Playwright:

- added a window source for `π - 2026-04-09--screencast-studio`
- added a region source for `eDP-1` → `Top Half`
- inspected the Raw DSL tab
- captured tmux server logs while previews were ensured
- fetched screenshots directly from `/api/previews/{id}/screenshot`
- compared dimensions and hashes of the returned JPEGs

### Key findings

The Raw DSL was correct for the window source:

- `type: "window"`
- `window_id: "0x5e0000e"`

The backend logs were also correct:

- the window preview resolved to a concrete rect:
  - `rect=6,30,1431,1884`
- the region preview resolved to a concrete rect:
  - `rect=0,0,2880,960`

So the backend is **not** simply throwing away `window_id` or `rect` and falling back to a root full-display capture target.

However, the direct preview screenshots revealed something wrong in the preview output stage:

```text
full desktop preview: 960x1920
window preview:      1280x1920
top-half region:     1280x1920
```

That is a red flag. The selected region `2880x960` should not become `1280x1920`, and the resolved window `1431x1884` should not become `1280x1920` either. This means the shared preview branch is distorting or renegotiating dimensions incorrectly after capture.

### Conclusion

The bug is now narrowed to the **shared preview scaling / caps negotiation path**, not the earlier suspected areas:

- not a bad source picker kind mapping
- not bad window/region DSL generation
- not a missing backend rect for window/region resolution

### Next action

Patch the preview branch so that when the source dimensions are known (window rect, region rect, camera size), it computes an explicit width **and height** preserving aspect ratio instead of sending only a width cap and letting GStreamer negotiate a distorted result.

## Step 8: Fixed Window/Region Preview Scaling and Verified the Outputs Directly

I patched the shared preview branch so that when the source dimensions are known, the preview caps now include an explicit width **and height** preserving aspect ratio, instead of only setting width and letting the rest of negotiation distort the frame. This was the concrete fix for the user-reported symptom that window and region selection both seemed to behave like full-screen.

### What changed

Changed preview scaling internals again:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go`
  - `sharedPreviewProfile.Width` became `sharedPreviewProfile.MaxWidth`
  - added `previewScaleCaps(...)`
  - added `previewTargetDimensions(...)`
  - added `previewSourceSize(...)`
  - `buildSharedPreviewConsumer(...)` now computes explicit preview width+height and includes `pixel-aspect-ratio=1/1` in the scale caps

Updated tests:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_test.go`
  - updated camera/window preview tests to check target dimensions
  - added a region scaling/aspect-ratio test

### Validation

I ran:

```bash
gofmt -w pkg/media/gst/shared_video.go pkg/media/gst/shared_video_test.go
go test ./pkg/media/gst ./internal/web ./pkg/discovery -count=1
go test ./... -count=1
```

Then I restarted the live app and re-created previews in the real UI.

Most importantly, I fetched the preview screenshots directly from the backend after the patch and checked the resulting image sizes:

```text
preview-18daabfb09b6 (desktop) -> 960x640
preview-4bc0e147ff33 (window)  -> 1280x1685
preview-fb9e27865c89 (region)  -> 1280x427
```

Those results are the key evidence that the bug was fixed at the preview-output stage:

- **desktop** no longer came out as the earlier distorted portrait-shaped preview
- **window** no longer collapsed to the same shape/output as the desktop preview
- **top-half region** now has the expected shallow wide aspect ratio instead of the previous bogus `1280x1920`

### Why this matters

Earlier, the logs already proved that the backend was resolving correct targets for both window and region sources. What was broken was the representation of those captures in the preview branch. After this patch, the preview output dimensions now match the selected source geometry much more closely, which should make the UI behavior line up with what the user selected.

### What worked

- tests passed cleanly
- the running app restarted successfully
- direct preview screenshot dimensions now differ appropriately by source type/geometry
- the region preview in particular now clearly reflects the top-half crop shape instead of masquerading as a portrait full-screen capture

### What should be done next

- Ask the user to re-check window and region previews visually in the live app.
- If they confirm this fixes the problem, the next major remaining bug is the per-monitor display-target model collapse.

## Step 9: Wrote a Full Postmortem and Pivoted the Runtime Fix from ximagesrc Coordinates to Full-Root Capture Plus videocrop

After the user confirmed that the preview still looked like the full display and explicitly asked whether standalone `ximagesrc` experiments had been run, I stopped treating the issue as a possible UI illusion and wrote up the full debugging story as a postmortem. This was the right moment to record the bug properly because the investigation had already passed through several false leads and produced a clear root-cause shift.

The postmortem captures the critical conclusion: the problem was not primarily the source picker, the DSL, or the preview card. The key failure was that `ximagesrc` region-coordinate capture on this machine could produce a frame with the requested output dimensions while still visually showing the full desktop content. The standalone `gst-launch-1.0` experiments made that clear.

### Prompt Context

**User prompt (verbatim):** "ok, write a full postmortem on the bug, and then continue testing the videocrop"

**Assistant interpretation:** Document the full investigation and root cause in ticket docs, then keep working the runtime fix and validation for the `videocrop` approach.

**Inferred user intent:** Preserve the debugging knowledge so the bug is understandable later, while also continuing toward a real fix rather than stopping at documentation.

### What I did

- Wrote a new postmortem document:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/design-doc/05-window-and-region-preview-full-display-postmortem.md`
- Captured in that doc:
  - the initial hypotheses,
  - the misleading intermediate evidence,
  - the preview-scaling side bug,
  - the standalone `ximagesrc` and `videocrop` experiments,
  - the final root-cause statement,
  - and the recommended fix direction.
- Implemented the next runtime step:
  - `pkg/discovery/service.go`
    - added root-geometry discovery support and display-aware command environment handling
  - `pkg/media/gst/preview.go`
    - changed region/window capture construction away from `ximagesrc startx/starty/endx/endy`
    - added `videocrop` based on root geometry and resolved rects
- Hit and fixed one build error while wiring the display environment helper:

```text
pkg/discovery/service.go:367:28: cannot use appendNonDisplayEnv() (value of type []string) as string value in argument to append
```

- Fixed that by building the environment slice first, then appending the `DISPLAY=...` entry.
- Re-ran formatting and tests:

```bash
gofmt -w pkg/discovery/service.go pkg/media/gst/preview.go
go test ./pkg/discovery ./pkg/media/gst ./internal/web -count=1
go test ./... -count=1
```

- Restarted the live tmux app again so the new runtime path is now active.

### Why

At this point the bug was no longer a quick “tune one parameter” issue. It had already crossed the threshold where false assumptions needed to be written down explicitly. The postmortem matters because several layers looked correct before the standalone experiments proved the capture primitive itself was untrustworthy for this use case.

The runtime fix also had to change shape: after the standalone proof, continuing to trust `ximagesrc` region coordinates would have been irresponsible.

### What worked

- the new postmortem consolidates the investigation into one place instead of scattering it across logs and ad hoc notes
- the build error was small and quickly fixed
- the focused and full repo test suites passed after the helper fix
- the running app was restarted successfully with the new `videocrop`-based runtime path

### What didn't work

- the first attempt at the display-aware environment helper was wrong and failed compilation
- that exact error was:

```text
pkg/discovery/service.go:367:28: cannot use appendNonDisplayEnv() (value of type []string) as string value in argument to append
```

### What I learned

- the fastest way to settle “is this a UI bug or a capture bug?” is to step outside the app and test the source element directly
- once a source element proves unreliable in standalone reproduction, the fix strategy should change immediately instead of trying to massage higher layers
- a good postmortem is not just nice documentation; it prevents the team from re-litigating the same dead-end hypotheses later

### What was tricky to build

The trickiest part was that the evidence evolved in stages and some of it initially pointed in the wrong direction. For example, preview IDs were distinct, rects were correct in logs, and the preview dimensions were eventually corrected, yet the content-selection bug still remained. That combination makes it easy to misdiagnose the issue as a rendering quirk rather than a source-element semantics problem.

The second tricky part was converting from a desired absolute region/window rectangle to `videocrop` margins. That requires reliable root-window geometry and careful handling of left/top/right/bottom crop values rather than just width/height.

### What warrants a second pair of eyes

- The new `RootGeometry(...)` helper and display-environment handling in `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/discovery/service.go`
- The `videocrop` construction logic in `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/preview.go`
- Any performance implications of always capturing the full root for region/window sources before cropping

### What should be done in the future

- Finish validating that the live app now shows genuinely different content for full desktop vs. bottom-half region under the new runtime path.
- If `videocrop` still fails in edge cases, the next fallback should be a focused prototype for custom X11 capture feeding GStreamer/appsrc.

### Code review instructions

Read in this order:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/design-doc/05-window-and-region-preview-full-display-postmortem.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/discovery/service.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/preview.go`

Then validate with:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
go test ./... -count=1
```

And finally re-check the live app for the specific reproduction:

- `Full Desktop`
- `eDP-1 bottom half`

### Technical details

The decisive standalone experiments behind this change were:

```bash
# Produced the requested dimensions, but still looked like full desktop content squeezed into the region
GST_DEBUG=1 gst-launch-1.0 -q ximagesrc use-damage=false startx=0 starty=960 endx=2879 endy=1919 num-buffers=1 ! videoconvert ! jpegenc ! multifilesink location=/tmp/gst-region-bottom-half.jpg

# Produced a true bottom-half crop
GST_DEBUG=1 gst-launch-1.0 -q ximagesrc use-damage=false num-buffers=1 ! videocrop top=960 bottom=0 left=0 right=0 ! videoconvert ! jpegenc ! multifilesink location=/tmp/gst-region-bottom-half-videocrop.jpg
```

## Step 10: Confirmed the Remaining “Same Format” Problem Is a UI Preview-Frame Bug, Not a Backend Crop Bug

After the backend `videocrop` path started producing true bottom-half screenshots, the user reported a more precise remaining issue: even when the cropped content is right, the **window/region preview size still appears in the same format**. This was an important clarification because it separated “wrong pixels” from “wrong presentation.”

At that point, the quickest way to validate the user's claim was to compare the screenshot against the actual preview card rendering code. That comparison immediately revealed a frontend bug that had become visible only after the backend crop bug was corrected.

### Prompt Context

**User prompt (verbatim):** "when cropping, the window preview size is still the same format, even if say the bottom half would be 2880x960 or the right half 1440x1920 /tmp/pi-clipboard-010aae0c-542c-4711-942c-ed6ac0c8b28d.png"

**Assistant interpretation:** The user is saying that even though the selected crop dimensions differ significantly, the preview cards still render with the same frame shape in the UI.

**Inferred user intent:** Make the visual presentation of preview cards reflect the actual aspect ratio of the selected source instead of flattening everything into one generic box.

### What I did

- Analyzed the new screenshot and confirmed the visible symptom:
  - `eDP-1 bottom half` shows `2880x960`
  - `eDP-1 right half` shows `1440x1920`
  - but all cards still use the same preview frame shape
- Re-read the preview component:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/preview/PreviewStream.tsx`
- Confirmed the root cause in code:
  - both the loading/error placeholder and the live image container hard-code `aspectRatio: '4/3'`
  - the image uses `objectFit: 'cover'`

### Conclusion

This remaining issue is now a **frontend presentation bug**:

- the backend crop output is different and correct,
- but the card renderer forces every preview into a fixed `4:3` frame,
- which hides the real aspect-ratio differences between full desktop, bottom-half, right-half, window, and camera sources.

### What should be done next

- update the preview card to use a source-aware or image-aware aspect ratio,
- stop forcing everything into `4:3`,
- and prefer `contain`-style presentation so the user can see the full cropped preview shape rather than a cover-style fill.

## Step 11: Verified That `recordings/demo-2` Is Not a Single-Frame Recording

After the user reported that recording into `recordings/demo-2` looked like it only captured the first frame, I checked the server logs and the actual output files immediately instead of assuming this was another preview bug. That distinction mattered because by this point the runtime had already undergone several targeted changes, and it was important to determine whether the regression was in the real recording outputs or only in how the result was being observed.

### Prompt Context

**User prompt (verbatim):** "after pressing record (into recordings/demo-2) it looks like nothing happens after the first frame, as if it only captured the first frame"

**Assistant interpretation:** Investigate whether the latest recording produced a broken one-frame output or only appeared frozen in the UI/player.

**Inferred user intent:** Confirm whether the new runtime changes introduced a real recording regression before moving on.

### What I did

- Listed the generated files under `recordings/demo-2`
- Inspected the live `tmux` logs around the recording
- Ran `ffprobe` on the resulting files
- Extracted multiple frames from `recordings/demo-2/eDP-1 bottom half.mov`
- Compared hashes of early/mid/late extracted frames

Commands used:

```bash
ffprobe -v error -show_entries format=duration,size -of default=noprint_wrappers=1:nokey=0 recordings/demo-2/*.mov recordings/demo-2/*.wav
ffprobe -v error -count_frames -select_streams v:0 -show_entries stream=codec_name,width,height,avg_frame_rate,nb_read_frames,duration -of default=noprint_wrappers=1:nokey=0 recordings/demo-2/*.mov
ffmpeg -y -loglevel error -ss 0.0 -i 'recordings/demo-2/eDP-1 bottom half.mov' -frames:v 1 /tmp/demo2-frames/bottom-start.jpg
ffmpeg -y -loglevel error -ss 2.5 -i 'recordings/demo-2/eDP-1 bottom half.mov' -frames:v 1 /tmp/demo2-frames/bottom-mid.jpg
ffmpeg -y -loglevel error -ss 4.8 -i 'recordings/demo-2/eDP-1 bottom half.mov' -frames:v 1 /tmp/demo2-frames/bottom-end.jpg
sha256sum /tmp/demo2-frames/bottom-start.jpg /tmp/demo2-frames/bottom-mid.jpg /tmp/demo2-frames/bottom-end.jpg
```

### What worked

The outputs clearly showed this was **not** a one-frame recording problem:

- `recordings/demo-2/eDP-1 bottom half.mov`
  - duration `5.208333`
  - size `797647`
  - `125` read frames
  - `2880x960`
- `recordings/demo-2/eDP-1 right half.mov`
  - duration `4.958333`
  - `119` read frames
  - `1440x1920`
- `recordings/demo-2/Full Desktop.mov`
  - duration `3.666667`
  - `88` read frames
  - `2880x1920`

The tmux logs also showed ongoing recorder buffer pushes during the session, for example:

```text
buffer_count=100
buffer_count=110
buffer_count=120
```

The extracted frame hashes from the bottom-half recording were different:

```text
580bd36b29fdb87e5e037fa7d1fee055351ebda2777a7458165936d631e54278  bottom-start.jpg
b5cd9e383e3b50edff4af3da0fa48d4f30c54e1d26ec6462fe29a2a8b7ca21bb  bottom-mid.jpg
0b06f86a411ad09f1201e770e2bc7cbdbda415c2300d4701406caebbe40ec196  bottom-end.jpg
```

That is strong evidence that the file contents change over time and are not frozen on frame 1.

### What didn't work

My first `ffmpeg` extraction command failed because I typed the MOV filename incorrectly and ignored the spaces in the actual path. The exact failure was:

```text
Error opening input file recordings/demo-2/eDP-1.bottom.half.mov.
```

I fixed that by quoting the real filename.

### What I learned

- The latest recording regression report does **not** currently point to a backend “single frame only” failure.
- The backend recording path is still writing multi-frame outputs with plausible durations and frame counts.
- The remaining issue is more likely in one of these buckets:
  - the live preview appearing frozen during recording,
  - the player or thumbnail the user used to inspect the MOV,
  - or a mismatch between what source the user expected to see moving and what actually changed on screen.

### What warrants a second pair of eyes

- The user should clarify *where* the freeze was observed:
  - in the live app preview during recording,
  - in a file manager thumbnail,
  - or inside a video player during playback.

### What should be done in the future

- If the user means the **live preview during recording**, inspect preview frame updates while recording is active.
- If the user means **playback**, test the generated MOV in the same player they used.

### Code review instructions

Review the evidence rather than code first:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
ffprobe -v error -count_frames -select_streams v:0 -show_entries stream=codec_name,width,height,avg_frame_rate,nb_read_frames,duration -of default=noprint_wrappers=1:nokey=0 recordings/demo-2/*.mov
```

Then inspect the live logs around the recording session from the `scs-web-ui` tmux buffer.

## Step 12: Saved the Investigation Scripts Retroactively and Fixed the Preview Freeze During Recording

The user asked for all debugging scripts and helper workflows to be saved under the ticket `scripts/` folder with numerical prefixes, including retroactively. I paused the live debugging long enough to turn the shell-history-style experiments into actual rerunnable scripts, then restarted the killed web server and resumed the freeze investigation using those new scripts instead of ad hoc commands.

This ended up being useful immediately, because the newly saved freeze-poll script reproduced the preview-freeze-on-recording regression cleanly and repeatedly. That let me validate a stronger runtime fix and confirm the regression was gone using the same scripted repro.

### Prompt Context

**User prompt (verbatim):** "save all scripts and helper files you use for debugging in the scripts/ folder of the ticket with a numerical prefix so that we can retrace your steps in the future. do this retroactively too, and then continue"

**Assistant interpretation:** Convert the live debugging workflow into durable ticket-local scripts first, then resume the technical investigation using those scripts.

**Inferred user intent:** Make the debugging process reproducible for future review instead of relying on terminal history or memory.

### What I did

Added ticket-local scripts under:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/01-restart-scs-web-ui.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/02-gst-ximagesrc-coordinate-vs-videocrop.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/03-api-region-preview-smoke.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/04-recording-output-inspect.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/05-preview-freeze-poll.sh`

Restarted the live server after the user noted it had been killed:

```bash
bash ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/01-restart-scs-web-ui.sh
```

Then I used the saved freeze script to reproduce the active bug:

```bash
bash ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/05-preview-freeze-poll.sh
```

Initial repro showed:

- preview hashes changing before recording,
- preview hashes changing once at recording start,
- then `lastFrameAt` freezing during recording,
- and resuming only after stop.

That proved the preview really was freezing during recording.

I first tried a lighter mitigation by making the raw recording consumer queue leaky and its appsink dropping, but the freeze persisted.

I then implemented the stronger fix in:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_recording_bridge.go`

Key change:

- the shared raw appsink callback no longer pushes directly into the recorder pipeline synchronously on the streaming thread
- instead it copies buffers into an async queue and a separate sample-pump goroutine feeds the bridge recorder

This prevents the recording bridge from stalling the shared source and starving the preview consumer.

After that fix, I re-ran the same saved freeze script.

### What worked

The second run of `05-preview-freeze-poll.sh` showed the bug was cleared:

- preview hashes changed before recording
- preview hashes continued changing during recording
- `lastFrameAt` kept advancing during recording instead of freezing
- preview activity also continued after recording stopped

Representative evidence from the passing run:

```text
pre 1 ...
pre 2 ...
pre 3 ...
during 1 ...
during 2 ...
during 3 ...
during 4 ...
post 1 ...
post 2 ...
post 3 ...
```

and

```text
pre 1 running 2026-04-13T22:51:06-04:00 True
pre 2 running 2026-04-13T22:51:07-04:00 True
pre 3 running 2026-04-13T22:51:09-04:00 True
during 1 running 2026-04-13T22:51:10-04:00 True
during 2 running 2026-04-13T22:51:11-04:00 True
during 3 running 2026-04-13T22:51:13-04:00 True
during 4 running 2026-04-13T22:51:15-04:00 True
post 1 running 2026-04-13T22:51:18-04:00 True
post 2 running 2026-04-13T22:51:20-04:00 True
post 3 running 2026-04-13T22:51:21-04:00 True
```

That is exactly the signal I wanted: the preview no longer stalls during recording.

### What didn't work

Two intermediate attempts failed or needed refinement:

1. The first saved version of `05-preview-freeze-poll.sh` was too eager and hit a transient screenshot error:

```text
curl: (22) The requested URL returned error: 400
```

I fixed that by adding screenshot retries to the script.

2. The first runtime mitigation (leaky raw queue + dropping raw appsink) did **not** fix the freeze by itself.

### What I learned

- Saving the debugging flow as scripts paid off immediately; the scripted repro made it much easier to validate fixes honestly.
- The preview freeze was not just “the user maybe looked at a static region.” The `lastFrameAt` evidence proved the preview really stalled during recording.
- The root cause was stronger than a minor buffering issue: the recording bridge needed to be decoupled from the streaming callback path so preview and recording could progress independently.

### What was tricky to build

The async bridge fix had concurrency traps. In particular, once I introduced a sample-pump goroutine, I had to avoid deadlocking the recorder shutdown path if the pump itself triggered an error. I also had to guard against enqueueing into a channel that might already be closed during shutdown.

The final fix required:

- an async buffer queue,
- a dedicated pump goroutine,
- safe enqueue logic,
- and shutdown ordering that does not wait on the current goroutine to finish itself.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_recording_bridge.go`
  - especially the async queue / pump / shutdown ordering
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/05-preview-freeze-poll.sh`
  - this is now the key regression reproducer for the freeze bug

### What should be done in the future

- Re-test the exact bug in the browser manually after the user refreshes.
- If the user still perceives the saved MOVs as frozen, inspect the specific playback path or player they are using, because the backend preview-freeze bug now has a passing scripted repro.

### Code review instructions

Start with:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/05-preview-freeze-poll.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_recording_bridge.go`

Then validate with:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
bash ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/01-restart-scs-web-ui.sh
bash ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/05-preview-freeze-poll.sh
```

## Step 13: Measured Recording CPU as Standalone Benchmarks and Saved All Results Under the Ticket Scripts Folder

The user noticed that the app becomes very CPU-heavy specifically when a recording starts, and asked whether the encoder or the current GStreamer setup might be the real cause. Instead of changing the main server immediately, I treated that as a measurement problem first and built two standalone benchmark harnesses under the ticket `scripts/` folder so the behavior could be reproduced and compared outside the live web server.

This step was useful because it separated “idle server confusion” from actual recording cost. The saved measurements show that recording CPU really does jump sharply, that the x264 encode path is already expensive even in a pure `gst-launch-1.0` pipeline, and that the current shared raw-consumer/appsrc bridge path in the app is even more expensive than direct GStreamer recording.

### Prompt Context

**User prompt (verbatim):** "its when starting a recording that it jumps up real high. Maybe it's the encoder that is not well adapted? Maybe you can do some performance measurements as standalone tests, so that we don't have to modify the main server, see if the issues is with our gsteamer setup?"

**Assistant interpretation:** Build standalone performance benchmarks for the recording path instead of changing the main server immediately, and use those measurements to determine whether the encoder or the current GStreamer topology is the main CPU culprit.

**Inferred user intent:** Get an honest, reproducible answer about recording CPU cost using isolated tests, not speculation or server-only instrumentation.

### What I did

Added benchmark harnesses and saved result directories under:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/06-gst-recording-performance-matrix/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/main.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/08-recording-performance-measurements-summary.md`

Saved raw results under:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/06-gst-recording-performance-matrix/results/20260413-230721/`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/results/20260413-230755/`

The pure-GStreamer harness measured four cases on the real `2880x960` bottom-half region shape:

1. capture to `fakesink`
2. preview-like JPEG path
3. direct x264 recording with the current preset (`speed-preset=3`)
4. direct x264 recording with a faster preset (`speed-preset=1`)

The Go shared-bridge harness measured three cases using the current shared-source runtime topology:

1. preview only
2. recorder only
3. preview + recorder together

Both harnesses save:

- raw `pidstat` CPU logs,
- stdout/stderr logs,
- `ffprobe` output for recorded files,
- and a human summary markdown file.

### Why

At this point the active question was no longer just “is the preview frozen?” or “is the server idle?” It was specifically whether starting a recording is inherently expensive because of x264, or whether our shared GStreamer topology is adding disproportionate overhead.

A standalone measurement harness answers that much more cleanly than ad hoc `top` observation inside the main web server.

### What worked

The standalone results were decisive.

Pure `gst-launch-1.0` results (`06-gst-recording-performance-matrix`):

- capture to `fakesink`: `23.17%` avg CPU, `31.00%` max CPU
- preview-like JPEG pipeline: `8.67%` avg CPU, `10.00%` max CPU
- direct record, current x264 preset: `86.50%` avg CPU, `91.00%` max CPU
- direct record, faster x264 preset: `49.83%` avg CPU, `52.00%` max CPU

Shared Go bridge results (`07-go-shared-recording-performance-matrix`):

- preview only: `10.67%` avg CPU, `14.00%` max CPU
- recorder only: `139.57%` avg CPU, `237.00%` max CPU
- preview + recorder: `151.62%` avg CPU, `265.00%` max CPU

These numbers strongly suggest:

- x264 is already the dominant CPU consumer in the pure pipeline,
- the current x264 preset matters a lot,
- and the current shared raw-consumer/appsrc bridge path is significantly more expensive than the pure direct GStreamer encode path.

### What didn't work

I hit one small tooling bug while summarizing results: the first `pidstat` parser treated the sample `%CPU` column as field 8, but the sample lines include an `AM/PM` token, so `%CPU` is actually field 9 there. That caused:

- `max_cpu` to show `0%`, and
- two `avg_cpu` summaries to come out blank when `pidstat` exited without an `Average:` line.

I fixed the saved parsers to:

- prefer the `Average:` line when present,
- otherwise compute the mean from the sampled `%CPU` values,
- and compute max from the sampled `%CPU` values correctly.

I regenerated the summary markdown files from the raw saved logs instead of re-running everything just to repair the summary formatting.

### What I learned

- The user’s suspicion was well-founded: the recording CPU spike is not imaginary and is not mainly about “the server being alive.”
- For the tested `2880x960 @ 24 fps` region shape, encoding dominates CPU cost.
- The current shared-source bridge topology adds substantial extra cost beyond direct GStreamer recording, likely because it goes through a raw `appsink -> Go -> appsrc -> x264enc` path.

### What was tricky to build

The tricky part was making the benchmark harnesses both useful and durable. They needed to:

- use the real display and region shape from this machine,
- save all raw logs under the ticket folder,
- verify output files with `ffprobe`,
- and produce summaries someone else can inspect later.

The parser bug was a good reminder that measurement tooling can be wrong even when the raw data is fine. Saving the raw `pidstat` logs made it easy to correct the summary afterward without losing trust in the run.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/06-gst-recording-performance-matrix/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/main.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_recording_bridge.go`

The bridge benchmark numbers are high enough that the shared recording topology deserves careful review before we tune around it blindly.

### What should be done in the future

- Add a follow-up benchmark matrix for multiple x264 presets and bitrates inside the shared bridge path.
- Measure smaller region sizes and lower FPS values so we can understand scaling behavior rather than just one large-region case.
- Investigate whether the `appsink -> Go -> appsrc` bridge can be replaced or reduced for lower-overhead shared recording.

### Code review instructions

Start with the human summary:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/08-recording-performance-measurements-summary.md`

Then inspect the saved raw result directories:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/06-gst-recording-performance-matrix/results/20260413-230721/`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/results/20260413-230755/`

Re-run with:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
bash ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/06-gst-recording-performance-matrix/run.sh
bash ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/run.sh
```

### Technical details

The current measured region and capture shape were:

```text
DISPLAY=:0
root=2880x1920
region=0,960,2880,960
fps=24
```

The most important saved result snippets are:

```text
pure gst-launch direct record current preset: 86.50% avg CPU
pure gst-launch direct record faster preset: 49.83% avg CPU
shared bridge recorder only: 139.57% avg CPU
shared bridge preview + recorder: 151.62% avg CPU
```

## Step 14: Copied the Vault Report into the Ticket and Added a Staged Bridge-Overhead Benchmark

After writing the long Obsidian project report, the user asked to copy it into the active ticket and then continue the performance investigation by going through all the bridge-related scenarios in a standalone test and recording the results. I treated that as two separate deliverables: first make the project report available inside the ticket itself, then add a new benchmark that isolates the `appsink -> Go -> appsrc` path in stages instead of inferring its cost only from the earlier coarse benchmark suites.

This was a good next step because the earlier performance measurements strongly suggested that recording was expensive, but they did not tell me whether the bridge itself was the primary cost center or whether the encoder was still doing most of the damage. The new staged benchmark answered that more directly.

### Prompt Context

**User prompt (verbatim):** "no, let's `cp` the project report into our ticket, and let's resume our performance investigation by going through all the scenarios in a standalone test and recording the results."

**Assistant interpretation:** Copy the new Obsidian project report into the ticket workspace, then continue the performance investigation with a more fine-grained standalone scenario matrix.

**Inferred user intent:** Keep the long-form knowledge captured in both the vault and the ticket, and make the performance investigation more systematic by measuring the bridge stages directly rather than speculating about them.

### What I did

Copied the vault report into the ticket:

- source: `/home/manuel/code/wesen/obsidian-vault/Projects/2026/04/13/PROJ - Screencast Studio - GStreamer Setup, Performance, and Region Debugging Report.md`
- destination: `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/reference/02-gstreamer-setup-performance-and-region-debugging-project-report.md`

Then added a new staged benchmark suite under:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/main.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/10-go-bridge-overhead-measurements-summary.md`

Saved raw run results under:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-232943/`

The staged scenarios were:

1. `normalized-fakesink`
2. `appsink-discard`
3. `appsink-copy-discard`
4. `appsink-copy-async-discard`
5. `appsink-copy-async-appsrc-fakesink`
6. `appsink-copy-async-appsrc-x264`

### Why

The earlier benchmark suite compared large buckets:

- pure direct GStreamer encoding,
- full shared Go bridge recording,
- and preview+recording together.

That was useful, but it still left a hard question open:

> Is the expensive part really the `appsink -> Go -> appsrc` bridge machinery, or is x264 still the main CPU hog even after frames cross the Go boundary?

A staged benchmark is the right way to answer that.

### What worked

The staged benchmark produced a much clearer breakdown than I expected.

Saved results from `09-go-bridge-overhead-matrix/results/20260413-232943/01-summary.md`:

- `normalized-fakesink`: `24.50%` avg CPU, `26.00%` max CPU
- `appsink-discard`: `25.33%` avg CPU, `30.00%` max CPU
- `appsink-copy-discard`: `28.33%` avg CPU, `32.00%` max CPU
- `appsink-copy-async-discard`: `25.17%` avg CPU, `28.00%` max CPU
- `appsink-copy-async-appsrc-fakesink`: `24.33%` avg CPU, `29.00%` max CPU
- `appsink-copy-async-appsrc-x264`: `77.83%` avg CPU, `99.00%` max CPU

Those numbers suggest that in this standalone staged benchmark:

- `appsink` callback overhead is small,
- `buffer.Copy()` adds some cost but not an enormous amount,
- the async queue and `appsrc -> fakesink` path are not the main problem,
- the large jump reappears only once `x264` is added.

### What didn't work

The most important thing that did **not** line up cleanly was this: the staged benchmark does not match the earlier full shared-runtime benchmark exactly.

The earlier `07-go-shared-recording-performance-matrix` had shown much higher recorder-only CPU (`139.57%` avg CPU), while the new staged `appsrc -> x264` case came in much lower (`77.83%` avg CPU).

That means the investigation is not finished. There is still a discrepancy to reconcile.

### What I learned

The new benchmark shifted my confidence noticeably:

- I no longer think “the Go/appsrc bridge itself must be the whole CPU problem.”
- It now looks more likely that x264 remains the dominant cost center, and that the earlier higher benchmark includes additional overhead from the full shared runtime or from some other difference in measurement shape.

In other words, the bridge may still be costly, but it does not look catastrophically expensive in isolation.

### What was tricky to build

The main tricky part was defining a staged matrix that was still close enough to the real production path to be meaningful.

I wanted the benchmark to use:

- the real display,
- the real `2880x960` bottom-half region shape,
- the same normalized raw caps pattern,
- and the same style of `buffer.Copy()` plus `appsrc` push that production uses.

At the same time, I deliberately stripped away registry lifecycle, preview manager behavior, and unrelated web/runtime code so the benchmark could answer a narrower question.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/main.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_recording_bridge.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/main.go`

The most important review question is now: what does the full shared-runtime benchmark include that this staged benchmark does not?

### What should be done in the future

- Re-run and reconcile the earlier `07-go-shared-recording-performance-matrix` with the new staged `09-go-bridge-overhead-matrix`.
- Add x264 preset variations to the staged benchmark.
- Check whether the production shared runtime is paying extra cost through registry/tee behavior, logging, or other orchestration that the staged benchmark intentionally omits.

### Code review instructions

Start with:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/10-go-bridge-overhead-measurements-summary.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-232943/01-summary.md`

Then compare against:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/08-recording-performance-measurements-summary.md`

Re-run with:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
bash ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/run.sh
```

### Technical details

The most important staged result summary is:

```text
normalized-fakesink: 24.50% avg CPU
appsink-discard: 25.33% avg CPU
appsink-copy-discard: 28.33% avg CPU
appsink-copy-async-appsrc-fakesink: 24.33% avg CPU
appsink-copy-async-appsrc-x264: 77.83% avg CPU
```

That is why my current interpretation changed: the Go/appsrc bridge path looks much cheaper than x264 in isolation, even though the full shared runtime still needs reconciliation.

## Step 15: Reconciled the Earlier Benchmark Mismatch with a Same-Session Matrix

The user asked me to proceed with the reconciliation work rather than just talking about it, so I built a wrapper benchmark that reruns the direct GStreamer suite, the shared-runtime suite, and the staged bridge-overhead suite in one same-session matrix. The goal was simple: find out whether the earlier large gap between the shared-runtime recorder-only benchmark and the staged bridge+x264 benchmark was stable, or whether it was an artifact of different runs and different timing.

This step turned out to be very valuable because it changed the current interpretation of the performance problem. The large recorder-only mismatch did **not** reproduce. In the same-session comparison, recorder-only shared runtime, staged bridge+x264, and direct current-preset x264 all clustered around the same CPU range, while the big remaining cost spike was preview + recorder together.

### Prompt Context

**User prompt (verbatim):** "alright, do it, commit frequently, keep a diary."

**Assistant interpretation:** Proceed with the benchmark reconciliation work now, keep the ticket diary updated, and make focused commits rather than batching everything into one giant final diff.

**Inferred user intent:** Turn the reconciliation idea into real evidence, preserve the chronology, and keep the work reviewable and durable.

### What I did

Added a same-session wrapper suite under:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/11-go-shared-vs-bridge-reconciliation-matrix/run.sh`

That wrapper reran:

- `scripts/06-gst-recording-performance-matrix/run.sh`
- `scripts/07-go-shared-recording-performance-matrix/run.sh`
- `scripts/09-go-bridge-overhead-matrix/run.sh`

under the same environment:

- display `:0`
- root `2880x1920`
- region `0,960,2880,960`
- fps `24`
- duration `6s`

Saved the unified result under:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/11-go-shared-vs-bridge-reconciliation-matrix/results/20260413-233847/01-summary.md`

### Why

The key open question after Step 14 was:

> Why did the earlier full shared-runtime recorder-only run look so much more expensive than the staged bridge+x264 run?

If that difference were real and stable, it would mean the full shared runtime was paying a large hidden cost beyond the bridge and the encoder. If it was not stable, then the next optimization focus would change.

### What worked

The same-session rerun produced a much clearer answer.

Key numbers from the reconciled run:

- `06` direct current x264 record: `94.33%` avg CPU
- `06` direct ultrafast x264 record: `55.33%` avg CPU
- `07` shared-runtime recorder-only: `94.00%` avg CPU
- `07` shared-runtime preview + recorder: `131.00%` avg CPU
- `09` staged `appsink -> Go -> appsrc -> x264`: `91.50%` avg CPU
- `09` normalized baseline: `26.50%` avg CPU

The most important result is that the earlier large recorder-only mismatch disappeared. Recorder-only shared runtime, staged bridge+x264, and direct current-preset x264 are now all in the same rough range.

### What didn't work

I hit one small shell bug while building the reconciliation wrapper. The first version of `run_child()` captured the child suite's entire stdout instead of only the final `Results written to ...` path. That made the result-directory variable contain a giant blob of text and produced an error like:

```text
cat: 'display=:0' ... '01-summary.md': File name too long
```

I fixed that by sending the child suite's `tee` output to stderr inside `run_child()`, so the wrapper only captures the final extracted result path on stdout.

### What I learned

The reconciliation changed the current theory in an important way.

Before this step, it still looked plausible that the full shared runtime was dramatically more expensive than both direct encode and the staged bridge+x264 path. After the same-session rerun, that is no longer the best current interpretation.

The more reliable current view is:

- recorder-only cost is broadly aligned across:
  - direct current-preset x264 encode,
  - the real shared runtime,
  - and the staged bridge+x264 path,
- the more clearly expensive combined case is **preview + recorder together**,
- and x264 preset choice still matters a great deal.

### What was tricky to build

The tricky part was keeping the wrapper simple enough to trust.

I did **not** want a new giant benchmark framework that introduced its own measurement behavior. The wrapper therefore just reruns the already-existing suites in one controlled pass and produces a combined summary pointing to each child result directory.

That made the reconciliation more credible because it compared the same benchmark families rather than silently replacing them with a new implementation.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/11-go-shared-vs-bridge-reconciliation-matrix/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/11-go-shared-vs-bridge-reconciliation-matrix/results/20260413-233847/01-summary.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/08-recording-performance-measurements-summary.md`

The most important review question is whether the new interpretation is stated narrowly enough: the reconciliation result reduces confidence in “bridge-only overhead is the main recorder-only problem,” but it does **not** prove the bridge is free or that no other runtime overhead exists.

### What should be done in the future

- Tune x264 presets/bitrates/FPS next, because the reconciled numbers still show encoder cost dominating recorder-only work.
- Investigate preview + recorder together, because that remains substantially more expensive than recorder-only.
- Keep the staged bridge benchmark around as a sanity check whenever changing recorder architecture, since it is now useful as a decomposition tool.

### Code review instructions

Start with:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/11-go-shared-vs-bridge-reconciliation-matrix/results/20260413-233847/01-summary.md`

Then compare its child run directories:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/06-gst-recording-performance-matrix/results/20260413-233847/`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/07-go-shared-recording-performance-matrix/results/20260413-233911/`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/09-go-bridge-overhead-matrix/results/20260413-233935/`

Re-run with:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
bash ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/11-go-shared-vs-bridge-reconciliation-matrix/run.sh
```

### Technical details

The core reconciled result is:

```text
06 direct current x264: 94.33% avg CPU
07 shared-runtime recorder-only: 94.00% avg CPU
09 staged appsink->Go->appsrc->x264: 91.50% avg CPU
07 preview+recorder: 131.00% avg CPU
```

That is why my current interpretation changed: recorder-only paths now line up, while preview+recorder together remains the most clearly elevated combined case.

## Step 16: Measured the Preview + Recorder Interaction Directly and Confirmed It Is the Remaining Big CPU Spike

After the reconciliation step, the most important unresolved question was no longer “is recorder-only unexpectedly more expensive than the staged bridge?” That question had largely settled. The new leading question was instead: why does the combined preview + recorder case stay so much more expensive than recorder-only? To answer that, I added a new standalone benchmark that keeps a shared source but measures preview-only, recorder-only, preview+recorder with current preview settings, and preview+recorder with a cheaper preview profile.

This step matters because it shifts the optimization target. The benchmark shows that the current remaining performance spike is strongly associated with keeping the preview branch alive while recording, and that cheaper preview settings help only partially.

### Prompt Context

**User prompt (verbatim):** (same as Step 15 — continue the investigation, keep a diary, and proceed with the next useful benchmark work)

**Assistant interpretation:** Follow the new lead from the reconciliation run and measure the combined preview-plus-recorder case directly.

**Inferred user intent:** Keep narrowing the performance problem until the next optimization target is obvious and justified by saved evidence.

### What I did

Added a new benchmark suite under:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/main.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/13-preview-recorder-interplay-summary.md`

Saved raw results under:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/results/20260414-070646/`

The benchmark uses the same real region shape as the earlier runs:

- display `:0`
- root `2880x1920`
- region `0,960,2880,960`
- fps `24`

Measured scenarios:

1. `preview-current-only`
2. `recorder-current-only`
3. `preview-current-plus-recorder`
4. `preview-cheap-plus-recorder`

Importantly, the preview branch in this benchmark does a real JPEG byte copy into Go, to mirror the production preview path more honestly.

### Why

The reconciliation run showed that recorder-only shared runtime, staged bridge+x264, and direct current-preset encode all land in the same rough CPU range. That meant the next obvious question was not recorder-only anymore. It was:

> How much extra cost comes from keeping the preview branch alive during recording, and can cheaper preview settings materially reduce it?

That is exactly what this benchmark was designed to answer.

### What worked

The benchmark produced a strong result.

Saved results from `12-go-preview-recorder-interplay-matrix/results/20260414-070646/01-summary.md`:

- preview-only: `12.17%` avg CPU, `14.00%` max CPU
- recorder-only: `94.00%` avg CPU, `120.00%` max CPU
- current preview + recorder: `188.43%` avg CPU, `492.00%` max CPU
- cheap preview + recorder: `170.00%` avg CPU, `427.00%` max CPU

Additional counters were also useful:

- current preview-only copied about `17.3 MB` of JPEG data in 6 seconds
- current preview + recorder copied about `17.6 MB` while still pushing `144` recorder frames
- cheap preview + recorder reduced preview byte-copy load dramatically to about `1.6 MB`, but total CPU still remained very high

That means the combined cost spike is real and not explained only by the literal volume of copied JPEG bytes.

### What didn't work

Nothing materially failed in the benchmark implementation itself. The important “didn't work” result was the cheap-preview hypothesis: lowering preview width/fps/quality helped, but it did **not** collapse combined CPU anywhere near recorder-only levels.

### What I learned

This step changed the performance story again, but in a much more targeted way.

The strongest current conclusion is now:

- recorder-only cost is not the main remaining mystery,
- the large remaining spike is the **preview + recorder interaction**,
- and even a cheaper preview branch does not make that interaction cheap.

That suggests the shared source with two simultaneously active downstream branches is where the next optimization attention should go.

### What was tricky to build

The tricky part was making the benchmark closer to production reality without simply embedding the whole production runtime again.

In particular, I wanted to include:

- a shared source with a tee,
- a real preview branch that JPEG-encodes and copies bytes into Go,
- a real recorder raw branch that copies buffers into Go and pushes them into an `appsrc -> x264` pipeline,
- and a way to vary preview cost independently.

That gives a better answer than either the pure direct-encode benchmark or the staged bridge-only decomposition.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/main.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_recording_bridge.go`

The key review question is whether this benchmark is now close enough to the real combined production shape to justify optimizations based on it.

### What should be done in the future

- Investigate how to degrade or reshape preview work while recording, not just tune the encoder.
- Measure more preview profiles, not just one “cheap” fallback.
- Investigate whether conversion/scaling work is being duplicated unnecessarily between preview and recorder branches.

### Code review instructions

Start with:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/13-preview-recorder-interplay-summary.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/results/20260414-070646/01-summary.md`

Re-run with:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
bash ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/12-go-preview-recorder-interplay-matrix/run.sh
```

### Technical details

The most important current result snippet is:

```text
preview-only: 12.17% avg CPU
recorder-only: 94.00% avg CPU
preview-current-plus-recorder: 188.43% avg CPU
preview-cheap-plus-recorder: 170.00% avg CPU
```

That is why my current optimization target shifted again: the remaining expensive case is the live preview branch interacting with recording, not recorder-only bridge overhead.

## Step 17: Copied the Researcher Handoff Brief into the Ticket Workspace

After creating the detailed researcher-facing handoff brief in the Obsidian vault, the user asked to copy it into the ticket as well. I treated that as a packaging step rather than new analysis work: the point was to make sure the same handoff package exists both in the long-term vault and in the ticket workspace where the benchmark evidence and diary already live.

### Prompt Context

**User prompt (verbatim):** "yes"

**Assistant interpretation:** Copy the just-created vault research brief into the ticket workspace too.

**Inferred user intent:** Keep the researcher handoff package co-located with the ticket evidence, not only in the vault.

### What I did

Copied:

- source: `/home/manuel/code/wesen/obsidian-vault/Research/Institute/Projects/Screencast Gstreamer Port/Research Brief - Preview and Recording Performance Investigation Handoff.md`
- destination: `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/reference/03-research-brief-preview-and-recording-performance-investigation-handoff.md`

Then re-ran:

```bash
docmgr doctor --ticket SCS-0014 --stale-after 30
```

and confirmed the ticket still passed cleanly.

### What worked

- the copied brief now exists inside the ticket workspace,
- the ticket still validates cleanly,
- and the handoff package now lives in both the vault and the repo-local research context.

### What didn't work

Nothing failed in this step.

### What I learned

For longer investigations like this one, it is useful to store the handoff package in both places:

- the Obsidian vault for long-term knowledge organization,
- and the ticket for repo-local continuity and review.

### What should be done in the future

If the researcher updates the brief materially, consider keeping the ticket copy and the vault copy synchronized rather than letting them drift too far apart.

## Step 18: Isolated the Preview Branch More Aggressively with an Ablation Matrix

After the preview/recorder interplay benchmark made it clear that the combined case was the remaining big spike, I wanted one more decomposition step before touching production code. The next benchmark held the recorder path fixed and varied only the preview branch so I could separate three different ideas that had previously been bundled together: the cost of having a second active branch at all, the cost of JPEG generation, and the cost of copying preview data into Go.

This step did not give a final production tuning answer by itself, but it did make the next likely mitigation path much clearer: a degraded recording-mode preview is now the most practical next thing to test in production.

### Prompt Context

**User prompt (verbatim):** "yes"

**Assistant interpretation:** Proceed with the next proposed experiment, namely the preview-branch ablation benchmark.

**Inferred user intent:** Keep moving the standalone investigation forward rather than stopping at the previous summary.

### What I did

Added the benchmark suite:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/main.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/15-preview-branch-ablation-summary.md`

Saved raw results under:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/results/20260414-081748/`

Measured scenarios:

1. `recorder-only`
2. `preview-fakesink-plus-recorder`
3. `preview-jpeg-discard-plus-recorder`
4. `preview-raw-copy-plus-recorder`
5. `preview-current-plus-recorder`
6. `preview-cheap-plus-recorder`

### Why

The previous interplay benchmark had already told me that preview + recorder together is expensive. But it still bundled together several different sources of preview work:

- second-branch transform work,
- JPEG generation,
- Go-side byte copying,
- and the current combined real preview path.

This ablation matrix was the next clean way to separate those concerns.

### What worked

The benchmark produced a coherent within-run ordering.

Saved results from `14-go-preview-branch-ablation-matrix/results/20260414-081748/01-summary.md`:

- `recorder-only`: `125.33%` avg CPU
- `preview-fakesink-plus-recorder`: `134.00%`
- `preview-jpeg-discard-plus-recorder`: `137.83%`
- `preview-raw-copy-plus-recorder`: `141.33%`
- `preview-current-plus-recorder`: `152.33%`
- `preview-cheap-plus-recorder`: `112.00%`

The counters also helped explain the branch behavior:

- `preview-jpeg-discard-plus-recorder` pulled preview samples but copied no preview bytes into Go
- `preview-raw-copy-plus-recorder` copied about `133 MB` of raw preview bytes in 6 seconds
- `preview-current-plus-recorder` copied about `13 MB` of JPEG preview bytes
- the cheap preview profile reduced copied preview bytes to about `1.36 MB`

### What didn't work

The absolute recorder-only baseline in this run (`125.33%`) was higher than the earlier reconciled recorder-only baseline (~`94%`). That means I should not over-read the absolute numbers from this one benchmark in isolation.

So the safest interpretation of this ablation run is **relative within-run**, not “this new run replaces every earlier recorder-only number.”

### What I learned

The benchmark still taught something useful despite that variability.

Within this run:

- having a second active branch already costs something,
- JPEG adds more cost,
- raw Go-side copying adds more cost,
- the full current preview path is more expensive than the more stripped-down preview variants,
- and a strong cheap-preview profile materially reduces total CPU.

That supports the current practical next step:

> Try degrading preview while recording before attempting deeper architecture changes.

### What was tricky to build

The tricky part here was choosing preview variants that were simple enough to interpret but still meaningful for the real system.

I specifically wanted to separate:

- branch existence without preview sink work,
- JPEG without Go copy,
- Go copy without JPEG,
- the real current preview path,
- and a plausible low-cost recording-mode preview.

That gave the most interpretable matrix so far for the preview side.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/main.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_recording_bridge.go`

The key review question is whether the within-run ordering is enough to justify production experiments with degraded recording-mode preview settings.

### What should be done in the future

- Re-run the ablation matrix to check stability of the relative ordering.
- Try production-side preview degradation while recording.
- Measure more cheap-preview profiles instead of just one.
- Investigate whether some preview branch work can be shared or moved upstream.

### Code review instructions

Start with:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/15-preview-branch-ablation-summary.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/results/20260414-081748/01-summary.md`

Re-run with:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
bash ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/14-go-preview-branch-ablation-matrix/run.sh
```

### Technical details

The key within-run ordering was:

```text
recorder-only: 125.33%
preview-fakesink-plus-recorder: 134.00%
preview-jpeg-discard-plus-recorder: 137.83%
preview-raw-copy-plus-recorder: 141.33%
preview-current-plus-recorder: 152.33%
preview-cheap-plus-recorder: 112.00%
```

That is why the next practical optimization idea is a degraded preview mode while recording.

## Step 19: Confirmed the Imported Adaptive-Preview Idea with New Standalone Benchmarks

After importing the two newly researched markdown notes, I did not want to jump straight into production code. The imported report proposed a very specific mitigation direction — constrained preview while recording plus rate-first preview ordering — and the right next step was to test that idea independently of the main software so the result would be reproducible and attributable to the media-graph change itself rather than to unrelated runtime behavior.

So I built two new standalone experiment families under the ticket. The first was a direct confirmation matrix. The second was a repeatability wrapper because the first single-run matrix showed enough variance that I did not trust it alone as the basis for a production decision.

### Prompt Context

**User prompt (verbatim):** "ok, run experiments to confirm these new findings, independently of the main software, so that we c an repeoduce it and measure things in isolation."

**Assistant interpretation:** Build and run standalone benchmarks that directly test the imported mitigation ideas without modifying the main runtime first.

**Inferred user intent:** Verify the external research by reproducing it locally in controlled, ticket-local experiments before changing production code.

### What I did

Added and ran:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/16-go-preview-adaptive-confirmation-matrix/main.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/16-go-preview-adaptive-confirmation-matrix/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/17-go-preview-adaptive-repeatability-matrix/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/18-preview-adaptive-confirmation-summary.md`

Saved direct confirmation runs under:

- `scripts/16-go-preview-adaptive-confirmation-matrix/results/20260414-135008/`
- `scripts/16-go-preview-adaptive-confirmation-matrix/results/20260414-135103/`

Saved repeated runs under:

- `scripts/17-go-preview-adaptive-repeatability-matrix/results/20260414-135308/`

### Experimental design

The standalone harness deliberately kept the recorder path fixed and varied only two preview-path dimensions:

1. preview ordering:
   - `scale-first`
   - `rate-first`
2. preview profile:
   - `current`
   - imported `constrained`

Scenarios:

- `recorder-only`
- `preview-scale-first-current-plus-recorder`
- `preview-rate-first-current-plus-recorder`
- `preview-scale-first-constrained-plus-recorder`
- `preview-rate-first-constrained-plus-recorder`

The constrained profile matched the imported screen-like recommendation:

- `640 max width`
- `4 fps`
- `jpeg quality 50`

### What worked

The first direct confirmation matrix was useful, but noisy. That was enough to justify the repeatability wrapper.

The repeated results were much more helpful. From:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/17-go-preview-adaptive-repeatability-matrix/results/20260414-135308/01-summary.md`

mean CPU across three rounds came out to:

- `recorder-only`: `91.00%`
- `preview-scale-first-current-plus-recorder`: `108.28%`
- `preview-rate-first-current-plus-recorder`: `106.45%`
- `preview-scale-first-constrained-plus-recorder`: `104.95%`
- `preview-rate-first-constrained-plus-recorder`: `87.61%`

That gives the clearest independent support so far to the **combined** adaptive-preview direction.

### What didn't work

The first direct confirmation matrix by itself was too noisy to support a strong claim. In particular, one run suggested one ordering result while the next run shifted enough that I did not want to over-interpret it. That is exactly why I added the repeatability wrapper instead of pretending the first matrix was already definitive.

### What I learned

The imported notes were useful, but not in the way “somebody else already solved it” would imply. They were useful because they suggested a focused hypothesis that we could then test ourselves.

The repeated standalone result suggests:

- the **combined** mitigation direction is promising,
- rate-first ordering alone is only a modest win in the current-profile case,
- constrained profile alone is not strongly convincing by itself,
- but constrained profile **plus** rate-first ordering is the best preview+recorder variant we measured in repeated standalone runs.

In other words, the imported idea appears more convincing as a **paired change** than as two separately dominant tweaks.

### What was tricky to build

The tricky part was not the media graph itself. The trickiest part was evidence quality. Single-run CPU measurements can be misleading enough that they encourage confident but brittle interpretations. The repeatability wrapper was necessary because I wanted something I could actually trust before recommending the next production-side prototype.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/16-go-preview-adaptive-confirmation-matrix/main.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/17-go-preview-adaptive-repeatability-matrix/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/18-preview-adaptive-confirmation-summary.md`

The main review question is whether the repeated standalone result is strong enough to justify changing the real runtime next.

### What should be done in the future

Prototype the combined adaptive change in the production shared runtime:

1. constrain preview while recording is active,
2. reorder preview to `videorate -> videoscale -> jpegenc`.

Then validate it in the real app path, not only in the standalone harness.

### Code review instructions

Start with:

- `scripts/18-preview-adaptive-confirmation-summary.md`
- `scripts/17-go-preview-adaptive-repeatability-matrix/results/20260414-135308/01-summary.md`

Re-run the direct matrix with:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
bash ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/16-go-preview-adaptive-confirmation-matrix/run.sh
```

Re-run the repeated matrix with:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
bash ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/17-go-preview-adaptive-repeatability-matrix/run.sh
```

### Technical details

The repeated means that currently matter most are:

```text
recorder-only: 91.00%
scale-first current + recorder: 108.28%
rate-first current + recorder: 106.45%
scale-first constrained + recorder: 104.95%
rate-first constrained + recorder: 87.61%
```

That is why the current recommendation is to prototype the **combined adaptive-preview change** next, not just one isolated tweak.

## Step 20: Refactored Shared Preview Construction into an Explicit Policy / Recipe Layer

I started the production-side adaptive-preview work with a deliberate refactor-only slice instead of changing runtime behavior immediately. The reason was simple: the current preview branch ordering was still effectively “whatever order this one function happened to append elements in.” That makes future changes harder to review, harder to test, and easier to get subtly wrong.

So the first slice separated three concerns that had previously been mixed together:

1. **policy** — which preview profile should be used,
2. **layout** — in what order the preview stages should appear,
3. **assembly** — how those typed stages are turned into real GStreamer elements.

The important discipline in this step was to avoid changing the live behavior yet. I kept the default layout on the existing `scale-first` normal-preview behavior so the refactor could be tested on its own merits before the more behavior-changing slice that will actually enable adaptive preview at runtime.

### Prompt Context

**User prompt (verbatim):** "ok, let's add tasks for that"

**Assistant interpretation:** After creating the adaptive-preview task list, start executing the work task by task, commit at reasonable boundaries, and keep the diary updated as implementation proceeds.

**Inferred user intent:** Do not jump straight to a monolithic patch. Build the adaptive-preview prototype incrementally with a reviewable history.

### What I did

Added a new file:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/preview_policy.go`

Updated:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_test.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/tasks.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/changelog.md`

Implementation details of the slice:

- introduced typed preview layouts:
  - `scale-first`
  - `rate-first`
- introduced typed preview stages:
  - `queue`
  - `scale`
  - `scale-caps`
  - `rate`
  - `rate-caps`
  - `jpeg`
  - `sink`
- introduced a typed preview policy with:
  - normal preview profiles,
  - recording preview profiles,
  - layout selection,
  - and recipe generation for a given source + mode
- changed shared preview consumer construction so it now builds from an explicit preview recipe rather than one implicit hard-coded element ordering slice
- kept the current runtime on the default `scale-first` normal-preview behavior after the refactor

### Why

This slice unlocks the next ones.

The next adaptive-preview behavior changes need to be able to say, clearly and testably:

- “use a recording-time preview profile now,” and
- “build the preview branch as `rate-first` now.”

Before this refactor, those choices were buried inside one element-building function. After this refactor, they are explicit policy decisions.

### What worked

The refactor stayed small enough to validate cleanly.

Validation that passed:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
gofmt -w pkg/media/gst/preview_policy.go pkg/media/gst/shared_video.go pkg/media/gst/shared_video_test.go
go test ./pkg/media/gst ./internal/web ./pkg/discovery -count=1
go test ./... -count=1
```

I also added focused tests for:

- recording-time preview profile selection,
- stage ordering for `scale-first`,
- stage ordering for `rate-first`,
- and preview FPS clamping for non-default profile caps.

### What didn't work

Nothing materially failed in this slice. The main thing I had to catch during implementation was one stale reference to `profile.FPS` after switching the builder over to `recipe.Profile.FPS`. That was a small compile-time refactor slip, not a deeper design problem.

### What I learned

This was a good reminder that even when we think we are “just changing pipeline order,” it helps to make the order itself a first-class typed concept. Once the ordering and profiles were explicit, the next adaptive-preview steps became much more straightforward to reason about.

### What was tricky to build

The tricky part was deciding how far to go with abstraction in one step. I did **not** want to invent a giant generic media-graph DSL yet. That would have been too much for the current goal.

Instead I kept the abstraction tight:

- enough structure to make preview layout and preview profile choices explicit,
- but not so much machinery that the refactor itself became the project.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/preview_policy.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_test.go`

The key review question is whether this is the right level of declaration for preview layout/policy work without over-abstracting too early.

### What should be done in the future

The next slice should actually use this structure for behavior change:

- enable recording-time constrained preview profiles in the real runtime,
- and switch the preview branch to `rate-first` ordering.

### Code review instructions

Review the new preview policy types first:

- `pkg/media/gst/preview_policy.go`

Then review how `buildSharedPreviewConsumer(...)` now consumes the recipe in:

- `pkg/media/gst/shared_video.go`

Finally confirm the new test coverage in:

- `pkg/media/gst/shared_video_test.go`

### Technical details

Tasks completed in this slice:

- `Refactor shared preview construction so preview branch ordering is selected explicitly rather than being implicitly hardcoded in one element slice`
- `Add a small typed preview policy/recipe layer for shared preview consumers so layout and profile choices are easier to reason about and test`

## Step 21: Enabled the First Real Adaptive-Preview Behavior in the Production Runtime

After the policy/recipe refactor was in place, the next slice was the first one that actually changed runtime behavior. I switched the default shared preview policy to `rate-first`, added the recording-time constrained preview profiles to the production policy, and wired the shared source so that preview consumers have their profile reapplied when a recorder raw consumer attaches or detaches.

This slice was intentionally still narrower than “final rollout complete.” It changed the production path, but I kept the older `scale-first` layout available in the policy layer so comparisons remain possible during rollout and future tuning.

### Prompt Context

**User prompt (verbatim):** "ok, build task by task, commit at ppropriate intervals, and keep a detailed diary"

**Assistant interpretation:** Continue from the refactor slice into the first real adaptive-preview runtime change, but keep the implementation history reviewable and documented.

**Inferred user intent:** Make concrete progress now, but preserve enough narrative and checkpoints that the behavior change can be reviewed and, if needed, backed out or tuned.

### What I did

Updated:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/preview_policy.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_recording_bridge.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_test.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/tasks.md`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/changelog.md`

Behavior changes in this slice:

- the default preview layout is now `rate-first`,
- shared preview consumers now derive their recipe from the shared source rather than always forcing normal mode,
- shared sources now define both normal and recording-time preview profiles,
- and raw recorder attach/detach now triggers preview-profile rebalance for existing preview consumers.

Initial recording-time constrained profiles encoded in the production policy:

- display/window/region: `640 max width`, `4 fps`, `jpeg quality 50`
- camera: `960 max width`, `6 fps`, `jpeg quality 70`

### Why

The standalone experiments had already narrowed the most promising production change to the **combined adaptive direction**:

- constrained preview while recording,
- plus rate-first preview ordering.

The right next move was therefore to implement that combined direction in the real runtime and check that it still preserved the working shared-capture behavior.

### What worked

Validation passed cleanly.

Commands that succeeded:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
gofmt -w pkg/media/gst/preview_policy.go pkg/media/gst/shared_video.go pkg/media/gst/shared_video_recording_bridge.go pkg/media/gst/shared_video_test.go
go test ./pkg/media/gst ./internal/web ./pkg/discovery -count=1
go test ./... -count=1
bash ./ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/16-web-gst-default-runtime-e2e.sh
```

The most important production-path validation here was the existing real default-runtime harness. It still completed successfully after the adaptive-preview change:

- preview remained active during recording,
- recording finished successfully,
- and valid output media was produced.

### What didn't work

No new hard failure appeared in this slice.

The familiar `recording.gst.audio_level.parse.failed` log noise remained visible during the default-runtime harness, but that is the known pre-existing audio meter parsing caveat, not a new adaptive-preview regression.

### What I learned

The refactor in Step 20 paid off immediately. Once preview layout and preview profile were explicit policy decisions, the production behavior change became a fairly direct implementation instead of another tangled rewrite of element order and numeric constants.

I also learned that the shared source was already the right place to drive preview-mode transitions. Raw consumer attach/detach is the natural event boundary for switching preview policy on the same source.

### What was tricky to build

The tricky part was making sure I changed **profile** dynamically but not **layout** dynamically. The default runtime now uses `rate-first`, but the attach/detach-time adaptation only reapplies profile settings in place. That keeps the runtime behavior change controlled and avoids trying to restructure a live branch graph at the moment recording starts.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/preview_policy.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_recording_bridge.go`

The key review question is whether the attach/detach-time profile rebalance is happening at the right level and with acceptable live-pipeline risk.

### What should be done in the future

The next slice should measure the effect in the live app path rather than just the default-runtime harness. We now need real before/after app-path evidence, not just isolated and synthetic confirmation.

### Code review instructions

Review the production policy defaults in:

- `pkg/media/gst/preview_policy.go`

Then review the attach/detach-time rebalance hooks in:

- `pkg/media/gst/shared_video.go`
- `pkg/media/gst/shared_video_recording_bridge.go`

Finally review the validation evidence by re-running:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
go test ./... -count=1
bash ./ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/16-web-gst-default-runtime-e2e.sh
```

### Technical details

Tasks completed in this slice:

- `Implement a rate-first preview branch option in the real shared runtime`
- `Keep the existing scale-first ordering available long enough to compare against the new runtime behavior during rollout`
- `Add recording-time constrained preview profiles for shared sources`
- `Start with the experimentally-supported constrained profile for screen-like sources`
- `Choose and document an initial constrained camera profile for recording-time preview`
- `Recompute or reapply preview profiles when a recorder raw consumer attaches to a shared source`
- `Restore the normal preview profile automatically when the recorder raw consumer detaches`
- `Add focused unit tests for preview layout selection and preview profile selection`
- `Add focused unit tests for recording-time preview downgrade and post-record restore behavior`
- `Extend the existing validation harnesses or add one focused runtime test that proves preview remains live while recording under the adaptive-preview path`

## Step 22: Measured the Adaptive-Preview Prototype in the Real Server/API Path Before and After

Once the adaptive-preview behavior change was in place, I wanted one result that lived closer to the actual product than the standalone harnesses. The standalone experiments had already been useful for narrowing the direction, but at this point the more important question was whether the real server/API path improved at all.

To answer that cleanly, I built a new ticket-local measurement script that starts a real `screencast-studio serve` process from a specified repo/revision, drives preview + recording through the HTTP API, measures the real server process with `pidstat`, and saves the resulting media plus screenshots and hashes.

I then ran it twice:

- once on the pre-adaptive revision,
- and once on the current adaptive-preview revision.

### Prompt Context

**User prompt (verbatim):** "ok, build task by task, commit at ppropriate intervals, and keep a detailed diary"

**Assistant interpretation:** After landing the adaptive-preview runtime slice, continue into real validation instead of stopping at unit tests and the in-process default-runtime harness.

**Inferred user intent:** Prove whether the prototype helps in the actual app path, not only in isolated media experiments.

### What I did

Added:

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/20-live-app-adaptive-preview-runtime-summary.md`

Then measured:

- before revision: `1554243058be6ecd73651a240fd1b7fc8272e286`
- after revision: `e27ebdfa5bdda660bdd0caa00ee926e7c4c3435b`

Saved runs:

- before: `scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142838/`
- after: `scripts/19-live-app-preview-recording-cpu-measure/results/20260414-142808/`

### What worked

The real app-path comparison produced a meaningful before/after result.

From the saved summaries:

- before avg CPU: `188.27%`
- after avg CPU: `170.82%`

That is an improvement of about:

- `17.45` percentage points absolute
- about `9.3%` relative

This matters because it is not just a synthetic harness result anymore. It is a real server/API-path improvement.

### What didn't work

I initially made a measurement mistake: the first version of the script ran `go run` and measured the wrapper `go` process rather than the long-lived server binary. That yielded nonsense CPU data (`0.00%`).

I corrected the script by building a real binary first and then measuring that binary’s PID instead.

That fix was important. Without it, the “before/after” result would have looked cleaner than it actually deserved.

### What I learned

The adaptive-preview prototype is helping in the real app path, not just in the standalone media experiments.

But it is also clear that:

- CPU is still high,
- the adaptive preview change is a useful mitigation, not a complete fix,
- and the next question is now partly a product question: is the constrained preview still acceptable while recording?

### What was tricky to build

The trickiest part here was measurement validity. It is easy to collect a number from the wrong PID or from the wrong execution path and then talk yourself into believing you validated production behavior. I had to explicitly correct that when I noticed the impossible `0.00%` result from measuring the `go run` wrapper process.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/run.sh`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/20-live-app-adaptive-preview-runtime-summary.md`

The main review question is whether this script is a good enough reproducible app-path comparison harness for future tuning work.

### What should be done in the future

The next step should be to inspect the live preview quality during recording and decide whether the constrained profile is acceptable enough to keep as the default runtime behavior.

### Code review instructions

Re-run the current revision measurement with:

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio
LABEL=after-current PORT=7783 bash ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/19-live-app-preview-recording-cpu-measure/run.sh
```

Re-run the pre-adaptive comparison by creating a detached worktree at:

```text
1554243058be6ecd73651a240fd1b7fc8272e286
```

and then pointing `REPO=` at that worktree when invoking the same measurement script.

### Technical details

The key app-path comparison currently is:

```text
before avg CPU: 188.27%
after  avg CPU: 170.82%
```

That is why the adaptive-preview prototype currently looks worth keeping and iterating on rather than backing out immediately.

## Step 23: Re-ran the Preview-Freeze Check Against the Adaptive-Preview Runtime

After the live CPU before/after comparison came back positive, I wanted one more runtime safety check before calling the adaptive-preview prototype a serious keeper: I needed to make sure the new preview-profile rebalance on raw attach/detach had not accidentally reintroduced the earlier preview-freeze regression.

Instead of inventing a new script, I reused the existing freeze-poll workflow and pointed it at a fresh server started from the current adaptive-preview revision. I saved the artifacts under the ticket so the check is reproducible and reviewable.

### What I did

- built a fresh server binary from the current repo,
- started it on `:7785`,
- ran the existing `scripts/05-preview-freeze-poll.sh` against it,
- and saved the output under:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/21-adaptive-preview-freeze-check/20260414-143226/`

### What worked

The important result was that `lastFrameAt` kept advancing during recording:

```text
during 1 running 2026-04-14T14:32:34-04:00 True
during 2 running 2026-04-14T14:32:35-04:00 True
during 3 running 2026-04-14T14:32:37-04:00 True
during 4 running 2026-04-14T14:32:38-04:00 True
```

That is exactly the evidence I wanted here. It means the adaptive-preview profile rebalance did **not** bring back the old preview-freeze-on-record failure in this targeted runtime check.

### What didn't work

My first one-liner that wrapped the freeze-check had a shell scoping bug around `WORK_DIR` under `set -u`. I fixed that and re-ran successfully.

### What I learned

The adaptive-preview change is looking more credible now because it has passed three different kinds of post-change validation:

1. unit tests and full repo tests,
2. the default-runtime server/app seam harness,
3. and the targeted freeze-poll check.

### What should be done in the future

The remaining open task is more product-facing now: judge whether the constrained live preview during recording is acceptable enough to keep as the default behavior, or whether it should remain configurable.

## Step 24: Saved Reusable Playwright UI Repro Scripts for Camera Add, Preview Verification, and Recording

After using Playwright interactively to verify that the frontend was really receiving desktop and camera previews, I converted those one-off browser actions into saved ticket-local scripts. The goal here was reproducibility: future investigation should not depend on scrolling back through a chat transcript to recover an injected-JS snippet.

I stored the new files under a fresh numbered script directory so they can be rerun later through Playwright MCP filename mode. I also validated the main flows live against the running `scs-web-ui` server: add camera, verify preview streams, start recording, stop recording, and a combined add-camera-and-record smoke flow.

### Prompt Context

**User prompt (verbatim):** "store the JS in the scripts folder of the ticket. do the same for Rec, and potentially other playwright things that are useful to be reproduced in the future"

**Assistant interpretation:** Save the injected Playwright browser automation used during the live debugging session as first-class ticket artifacts, not just ephemeral chat output.

**Inferred user intent:** Make the frontend/browser repro path durable enough that future work can repeat camera-add and recording actions exactly, while avoiding accidental extra preview listeners from stale browser tabs.

### What I did

- Created:
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/22-playwright-ui-repros/00-usage.txt`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/22-playwright-ui-repros/01-add-camera-via-injected-js.js`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/22-playwright-ui-repros/02-start-recording-via-injected-js.js`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/22-playwright-ui-repros/03-stop-recording-via-injected-js.js`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/22-playwright-ui-repros/04-verify-preview-streams.js`
  - `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/22-playwright-ui-repros/05-add-camera-and-start-recording-smoke.js`
- Validated `01-add-camera-via-injected-js.js` live in filename mode; it moved the page from `Sources (1)` to `Sources (2)` and the camera preview reached `1280x720` with `complete=true`.
- Validated `02-start-recording-via-injected-js.js` and `03-stop-recording-via-injected-js.js` live against the running server.
- Validated `04-verify-preview-streams.js` and the combined `05-add-camera-and-start-recording-smoke.js` flow.
- Closed Playwright tabs afterwards and added explicit tab-hygiene guidance to `00-usage.txt` so old MJPEG listeners do not remain attached during later measurements.

### What worked

- Saving the scripts in Playwright MCP `run_code`-compatible form means they can be rerun by filename directly.
- The camera-add script reproduced the exact UI path that had previously been done manually in chat.
- The combined smoke flow proved that the saved browser repro assets can cover a real frontend path rather than just an API-only path.

### What didn't work

- During validation, I found that the first version of the recording-start script should guard against already-active recordings; otherwise a rerun could appear to succeed while attaching to a page that was already mid-recording. I tightened that before treating the script set as done.

### What I learned

The saved benchmark and API harnesses are not enough by themselves when the suspected regression boundary includes the real browser/UI path. Having a durable Playwright repro set inside the ticket makes it easier to separate:

- API-only / backend runtime behavior,
- browser-connected preview behavior,
- and browser-tab hygiene issues such as stray MJPEG listeners.

### What was tricky to build

The subtle part was that Playwright tabs themselves can distort the very thing we are trying to measure. Even when extra tabs do not create a whole new capture pipeline, they can keep additional preview HTTP listeners and browser-side rendering work alive. That means saved UI repro scripts should be paired with explicit tab cleanup guidance, not just with click automation.

### What warrants a second pair of eyes

- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/22-playwright-ui-repros/02-start-recording-via-injected-js.js`
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0014--fix-preview-regressions-in-screencast-studio-web-ui/scripts/22-playwright-ui-repros/05-add-camera-and-start-recording-smoke.js`

The review question is whether these should remain pure injected-JS UI scripts or whether a later follow-up should add a more formal browser-driven measurement harness that also samples `/metrics` over time.

### What should be done in the future

- Add a browser-driven CPU/metrics capture harness if we decide the unexplained web-UI-specific server CPU spike needs to be benchmarked directly.
- Keep closing Playwright tabs before and after frontend repro runs so measurements remain interpretable.

### Code review instructions

Re-run any of the saved scripts through Playwright MCP filename mode while the local server is running on `:7777`. Start with:

```text
.../scripts/22-playwright-ui-repros/01-add-camera-via-injected-js.js
.../scripts/22-playwright-ui-repros/04-verify-preview-streams.js
.../scripts/22-playwright-ui-repros/05-add-camera-and-start-recording-smoke.js
```

### Technical details

The validated live camera-preview result from the saved script was:

```text
preview-dc54cc832d0f/mjpeg
naturalWidth=1280
naturalHeight=720
complete=true
```
