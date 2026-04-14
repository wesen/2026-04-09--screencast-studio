# Changelog

## 2026-04-13

Created ticket **SCS-0013** to study how the separate `transcription-go` prototype should be integrated into `screencast-studio`. Wrote a detailed intern-facing design and implementation guide that maps the current GStreamer recording graph, the current server websocket/event path, and the transcription-go service/session protocol into a phased integration plan.

The main architectural conclusion is that screencast-studio should integrate with transcription-go at the **protocol level**, not by trying to import its `internal/...` Go packages directly. The recommended design adds a new transcription seam inside screencast-studio, branches normalized PCM out of the current GStreamer audio graph with a tee + appsink, and forwards transcript updates to the browser over the app's existing `/ws` websocket transport.

Validated the ticket docs with `docmgr doctor --ticket SCS-0013 --stale-after 30` after adding missing topic vocabulary entries (`go`, `gstreamer`, `transcription`). Uploaded the final document bundle to reMarkable as **SCS-0013 Live Transcription Integration Intern Guide** and verified it under `/ai/2026/04/13/SCS-0013`.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/recording.go — Existing audio graph and event seam that transcription must extend safely
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go — Existing browser websocket path that should carry transcript updates
- /home/manuel/code/wesen/2026-04-09--screencast-studio/proto/screencast/studio/v1/web.proto — Existing server event schema that needs transcription message additions
- /home/manuel/code/wesen/2026-04-13--transcription-go/server/server.py — Prototype ASR service API, including websocket streaming endpoint
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/runner.go — Prototype live transport orchestration and state model that informed the plan
