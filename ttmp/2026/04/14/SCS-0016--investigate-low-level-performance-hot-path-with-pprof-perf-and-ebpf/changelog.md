# Changelog

## 2026-04-14

Created ticket **SCS-0016** to isolate the lower-level profiling work from SCS-0015.

The reason for splitting the work is that SCS-0015 already narrowed several plausible browser-path suspects with app-level instrumentation:

- final MJPEG HTTP write/flush time looks too small to explain the hot phase,
- PreviewManager cached-frame copy/store/publication time looks too small,
- EventHub publish time also looks too small.

That means the remaining unexplained cost likely lives lower in the stack, closer to CGO, GStreamer buffer handoff, appsink callbacks, memcpy-heavy transitions, runtime scheduling, or some combination of those.

Started the new ticket documents:

- `design-doc/01-low-level-profiling-plan.md`
- `reference/01-investigation-diary.md`

The initial plan for this ticket is intentionally staged:

1. **Go pprof first** to answer whether the hot phase is still largely visible in Go userland.
2. **perf second** if pprof mainly points at `runtime.cgocall` or otherwise fails to explain the hot phase.
3. **eBPF third** only for narrowly targeted unanswered questions such as off-CPU, scheduler, syscall, or socket behavior.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go — upstream ticket already showed the final MJPEG write path is not the dominant explanation
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go — upstream ticket already measured PreviewManager frame-store/copy/publication timing
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/event_hub.go — upstream ticket already measured EventHub publish timing
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/media/gst/shared_video.go — likely next low-level boundary around GStreamer appsink callbacks and frame handoff into Go
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/serve.go — likely place to add optional local profiling enablement
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0015--investigate-browser-preview-streaming-pipeline-and-web-ui-performance-matrix/reference/02-browser-preview-streaming-lab-report.md — parent evidence trail that justified this new lower-level profiling ticket
