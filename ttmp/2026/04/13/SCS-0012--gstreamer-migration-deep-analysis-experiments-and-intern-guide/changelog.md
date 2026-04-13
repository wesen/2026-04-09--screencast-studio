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

