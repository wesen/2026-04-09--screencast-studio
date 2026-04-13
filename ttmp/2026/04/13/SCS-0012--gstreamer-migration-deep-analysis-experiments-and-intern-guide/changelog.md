# Changelog

## 2026-04-13

- Initial workspace created


## 2026-04-13

Completed GStreamer environment analysis, gst-launch experiments (video preview, audio capture, screenshots), go-gst binding compilation, and native pipeline experiments for both preview (appsink JPEG) and audio recording (WAV). All experiments successful.


## 2026-04-13

Wrote comprehensive GStreamer migration analysis (13 sections, 1222 lines) covering: GStreamer basics, current FFmpeg architecture, 7 validated replacement pipelines, new capabilities (screenshots/transcription/live effects), go-gst API patterns, architecture design, migration phases, risk analysis, element reference table, and FFmpeg-to-GStreamer cheat sheet. Uploaded bundled PDF (analysis + diary) to reMarkable at /ai/2026/04/13/SCS-0012.


## 2026-04-13

Implemented Phase 0 runtime seam for the GStreamer migration: added pkg/media interfaces, FFmpeg preview/recording adapters, rewired app recording and web preview lifecycles, created gst package skeleton, and kept all tests green (commit e36d29966f9fc2dd49721c1608192a2123b64c0c).

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go — PreviewManager now depends on PreviewRuntime and preview sessions
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go — Application now records through an injected/default runtime instead of calling recording.Run directly
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/ffmpeg/preview.go — Moved FFmpeg preview subprocess lifecycle behind PreviewRuntime
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/ffmpeg/recording.go — Wrapped pkg/recording.Run as a RecordingRuntime session
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/types.go — Introduced preview/recording runtime interfaces and media-layer event/result types


## 2026-04-13

Implemented the first in-repo native GStreamer preview runtime with go-gst: appsink JPEG delivery, GLib bus watch handling, source mapping for display/region/window/camera, and a reproducible smoke test script. Validated display, region, and camera preview; window preview currently fails with X11 BadMatch/MIT-SHM (commit 806c14e630a108ac3dd9670af0eb205c4c1072c9).

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/bus.go — GLib main-loop-backed bus watch helper for preview sessions
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/pipeline.go — Shared capsfilter and link helpers for pipeline assembly
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/preview.go — Native GStreamer preview runtime implementation
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/09-go-gst-preview-runtime-smoke/main.go — Reproducible runtime smoke test for display/region/camera/window preview paths


## 2026-04-13

Investigated unreliable window preview capture and fixed the native GStreamer preview runtime to resolve window geometry first, then capture the rectangle instead of relying on fragile ximagesrc XID capture. Added a reproducible investigation script and validated all preview source types, including the previously failing window case (commit b247f270b07600df025c5652317023b03a6f347d).

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/discovery/service.go — Exported WindowGeometry for reuse by the GStreamer preview runtime
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/preview.go — Window preview now resolves geometry and captures via region-style ximagesrc settings
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/10-window-preview-investigation.sh — Reproducible evidence for XID-vs-geometry window capture behavior


## 2026-04-13

Completed Phase 1 end-to-end validation for native GStreamer preview: added a server-level preview runtime injection hook and a reproducible HTTP harness that validates ensure, MJPEG streaming, preview suspend during recording, and preview restore after recording across display, region, camera, and window sources (commit 7db020ad31048d9b2f4f47adc9f04c8b6742ef6b).

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go — Added NewServerWithOptions and WithPreviewRuntime for staged preview-runtime validation
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/11-web-gst-preview-e2e/main.go — HTTP end-to-end harness for GStreamer preview ensure/MJPEG/suspend/restore


## 2026-04-13

Implemented the first native GStreamer recording runtime slice for video jobs: programmatic video pipeline construction, per-job worker supervision, EOS-driven stop/finalization, and a reproducible smoke harness that validated real MP4 output for display, region, and camera sources. Had to tune x264/stop behavior so EOS finalization completed cleanly (commit bc6e63e584291432ce857b7137053ed8576213fb).

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/recording.go — Native GStreamer video recording runtime with x264/mp4mux/qtmux pipelines and EOS stop handling
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/12-go-gst-recording-runtime-smoke/main.go — Reproducible validation harness for display/region/camera recording output


## 2026-04-13

Extended the native GStreamer recording runtime to support audio jobs: one pulsesrc branch per source, per-source gain, audiomixer request-pad linking, WAV and Opus/Ogg output branches, and a dedicated smoke harness that validated single-source WAV, single-source Opus, and a two-branch mixed WAV output (commit ec7136d5168f7e911ef209e369110157225e5e52).

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/recording.go — Added native GStreamer audio branch/mixer/output support to the recording runtime
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/13-go-gst-audio-recording-runtime-smoke/main.go — Smoke harness for WAV


## 2026-04-13

Finished Phase 2 of the GStreamer migration: refined recording lifecycle semantics (starting/running/stopping/finished/failed), added synthetic process_log parity, implemented internal max-duration shutdown via cancel causes + EOS, extended the runtime smoke harness for max-duration testing, and added a browser/API-level recording E2E harness covering explicit stop, max-duration timeout, parent cancel, active preview suspension, preview restore, valid MP4 output, and valid WAV output (commit 8adcd278fbee22f6199a0fff596ac99d7ded6bad).

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/recording.go — Phase 2 lifecycle semantics
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/12-go-gst-recording-runtime-smoke/main.go — Extended smoke harness for internal max-duration recording validation
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/14-web-gst-recording-e2e/main.go — Browser/API-level recording validation for stop


## 2026-04-13

Implemented Phase 3 feature plumbing for the GStreamer runtime: preview screenshots via latest-frame retrieval, a new screenshot HTTP endpoint, live audio effect controls via runtime-settable volume/audiodynamic elements, and websocket audio-meter publication from the recording graph. Added a Phase 3 web-level harness that validates JPEG screenshots, live effect updates during recording, and end-to-end audio-meter event flow. Note: exact level RMS decoding is still blocked by a go-gst binding limitation exposing the field as an unsafe pointer on this machine, so the meter path currently uses an availability fallback while keeping the graph/message plumbing real (commit a2f89f78bcd9bd502976727ed8088a3e0d22f1e6).

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go — Added POST /api/audio/effects for live gain/compressor updates
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go — Added screenshot endpoint for live preview sessions
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/recording.go — Added live audio-control hooks
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/15-web-gst-phase3-e2e/main.go — Phase 3 web-level validation harness for screenshot


## 2026-04-13

Switched the application and preview-manager defaults to the native GStreamer runtimes while preserving the older preview suspend/restore workaround for stability. A real-defaults end-to-end harness showed that removing suspend/restore without a true shared capture graph is still unsafe: preview can remain visually active and screenshots can work during recording, but the combined video+audio recording path can still fail finalization with a recording EOS timeout. This establishes the true Phase 4 requirement: a tee-based shared capture graph, not just deletion of the server-side handoff code.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go — Default recording runtime now points at the GStreamer runtime
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go — Default preview runtime now points at the GStreamer runtime
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/16-web-gst-default-runtime-e2e/main.go — Real-defaults harness documenting why shared capture is still required before Phase 4 can be closed



## 2026-04-13

Added focused Phase 4 media-runtime experiments to answer the shared-capture question honestly. A shared-tee experiment showed that the obvious EOS points which finalize MP4 also poison the whole shared pipeline, while the branch-local EOS points that keep preview alive fail to finalize MP4. A second shared-source appsink→appsrc bridge experiment showed that preview continuity is achievable without duplicate capture, but the recording side still has unresolved appsrc segment/timestamp handling before the architecture is production-ready.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/17-go-gst-shared-video-tee-experiment/main.go — Reproducible shared-tee branch-stop experiment across multiple EOS targets
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/18-go-gst-shared-source-appsink-appsrc-bridge/main.go — Reproducible shared-source bridge experiment isolating appsrc segment/timestamp issues


## 2026-04-13

Added a new long-form intern-facing Phase 4 architecture guide focused on shared source capture, preview/recording lifecycle semantics, and the experiment evidence from scripts 16–18. Uploaded the guide to reMarkable at `/ai/2026/04/13/SCS-0012` as `SCS-0012 Phase 4 Shared Capture Intern Guide`.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/design-doc/02-phase-4-shared-capture-architecture-and-intern-implementation-guide.md — New detailed intern guide for the remaining shared capture architecture problem


## 2026-04-13

Completed the first production Phase 4 shared-capture slice: added a shared GStreamer video source registry and migrated preview sessions to attach tee-backed preview branches instead of creating standalone source pipelines. Added a focused shared-preview smoke harness proving attach/detach, continued preview after one consumer stops, last-consumer shutdown, and clean source recreation. Re-ran the existing preview runtime and web preview end-to-end harnesses; both still passed under the new shared-preview implementation (commit 5fea3b6485af7cc9701bdcaee6475fb32ef7b3a8).

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go — New shared video source registry and tee-backed preview consumer primitives
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/preview.go — Preview runtime now acquires shared sources from the registry instead of building standalone source pipelines
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/19-go-gst-shared-preview-runtime-smoke/main.go — Focused validation harness for shared-preview branch lifetime behavior


## 2026-04-13

Implemented an isolated shared-source recording bridge prototype on top of the new shared video registry: a raw shared-source consumer plus an experimental `appsink -> appsrc -> x264enc -> mp4mux -> filesink` recorder. Added a focused bridge smoke harness that proved preview continuity during and after bridge recording without touching the stable production recording runtime. The remaining failure is recorder-side only: the bridge still emits `gst_segment_to_running_time` assertions and produces a tiny output file, so this is a bridge implementation milestone, not yet the Phase 4 recording-validation milestone.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video_recording_bridge.go — Experimental shared-source recording bridge implementation kept isolated from the stable runtime
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go — Extended shared source lifecycle to support raw recording consumers as well as preview consumers
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/SCS-0012--gstreamer-migration-deep-analysis-experiments-and-intern-guide/scripts/20-go-gst-shared-bridge-recorder-smoke/main.go — Focused validation harness for preview continuity plus isolated bridge recorder stop behavior
