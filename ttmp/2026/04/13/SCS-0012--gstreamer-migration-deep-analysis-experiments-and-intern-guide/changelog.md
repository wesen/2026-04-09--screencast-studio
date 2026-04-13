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

