# Changelog

## 2026-04-09

- Initial workspace created


## 2026-04-09

Created the second ticket for the web control frontend and wrote the detailed system design guide for the browser transport, API, preview, and UI architecture.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0002--screencast-studio-web-control-frontend-architecture-and-implementation-plan/design-doc/01-screencast-studio-web-control-frontend-system-design.md — Primary design deliverable for the web ticket
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0002--screencast-studio-web-control-frontend-architecture-and-implementation-plan/reference/01-diary.md — Chronological documentation record for the second ticket

## 2026-04-09

Expanded the ticket task list into a detailed intern-facing implementation plan with phased backend, transport, preview, frontend, packaging, and validation work items.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0002--screencast-studio-web-control-frontend-architecture-and-implementation-plan/tasks.md — Detailed execution checklist for implementing the web ticket
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0002--screencast-studio-web-control-frontend-architecture-and-implementation-plan/reference/01-diary.md — Diary updated to explain why the task expansion was added

## 2026-04-09

Implemented phase 1 of the web ticket: added the Go web server shell, a glazed `serve` command, a minimal health endpoint, a placeholder WebSocket route, and initial handler tests.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go — New web server shell and graceful shutdown handling
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/routes.go — Initial HTTP route registration and JSON helper
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server_test.go — Health and placeholder transport tests
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/serve.go — New `serve` glazed command
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/cli/root.go — Root command now wires in `serve`

## 2026-04-09

Refactored the recording runtime so `ManagedProcess.Run(ctx, ...)` is the blocking owner of the subprocess lifecycle and its drainers, reducing hidden goroutine ownership before the web session manager is layered on top.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/run.go — Managed process execution now uses an explicit blocking run method
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/events.go — Structured runtime events for session state and process log forwarding
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/app/application.go — Application boundary now exposes normalization, compile, and event-aware record helpers for web transport

## 2026-04-09

Implemented phases 2 through 4 of the web ticket: added read-only discovery/session APIs, setup normalize and compile endpoints, and a web recording session manager with start/stop/current handlers and fake-app tests.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/application.go — Web-layer application interface for testable transport code
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/api_types.go — Stable transport payloads and response mappers
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/event_hub.go — Event publication backbone for later WebSocket delivery
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/session_manager.go — Single-session recording coordinator for the web server
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_api.go — Discovery, setup, and recording HTTP handlers
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server_test.go — Handler and lifecycle tests using a fake application

## 2026-04-09

Implemented phases 5 and 6 of the backend web milestone: added `/ws` event delivery, preview lifecycle management with ensure/release/list semantics, MJPEG streaming endpoints, a preview runner abstraction, and fake-runner tests.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go — WebSocket upgrade and event fan-out
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_preview.go — Preview ensure/release/list/MJPEG handlers
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_manager.go — Preview leasing, worker ownership, and state publication
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/preview_runner.go — FFmpeg-backed preview runner plus JPEG frame parsing
- /home/manuel/code/wesen/2026-04-09--screencast-studio/pkg/recording/ffmpeg.go — Exported preview argument builder reused by the web preview runtime
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server_test.go — Preview lifecycle, MJPEG, and websocket tests

## 2026-04-09

Implemented Phase 7 (scaffold the React frontend): created ui/ workspace with Vite, TypeScript, React 18, Redux Toolkit, RTK Query, Storybook, and MSW. Implemented all base primitives (Btn, Radio, Sel, Slider, Win, WinBar), composite components (FakeScreen, MicMeter, Waveform), source cards, and studio panels matching the screencast-studio-v2.jsx.jsx visual language. Added RTK Query API layer and MSW mock handlers.

### Commit

`981641b` — "ui: scaffold React frontend with RTK Query, Storybook, and MSW"

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/ — New frontend workspace
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/api/types.ts — TypeScript types matching Go DSL
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/session/sessionSlice.ts — Session state management
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/studio-draft/studioDraftSlice.ts — Draft state management
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/styles/tokens.css — CSS design tokens

## 2026-04-09

Phase 8: Built main operator screen with source cards, preview panes, and recording controls (commit 86576d9)


## 2026-04-09

Phase 10: Added WebSocket client with reconnection and PreviewStream component (commit 0e101c4)


## 2026-04-09

Reviewed the current `ui/` frontend against the implemented Go web backend, documented the major frontend/backend contract drift, identified the reusable strengths in the component and styling layers, and wrote a detailed intern-facing cleanup and next-steps guide.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0002--screencast-studio-web-control-frontend-architecture-and-implementation-plan/design-doc/02-frontend-assessment-and-improvement-guide.md — New frontend assessment, review, and improvement guide
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0002--screencast-studio-web-control-frontend-architecture-and-implementation-plan/reference/01-diary.md — Diary updated with the evidence and validation steps behind the frontend review
