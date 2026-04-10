# Changelog

## 2026-04-09

- Initial workspace created

## 2026-04-09

Created the dedicated recording-configuration and runtime-telemetry productization ticket, wrote the initial system design guide, and added a phased implementation task list covering naming, destination preview, live metering, and disk/runtime telemetry.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0004--screencast-studio-recording-configuration-and-runtime-telemetry-productization-plan/design-doc/01-recording-configuration-and-runtime-telemetry-system-design.md — Main design and implementation guide
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0004--screencast-studio-recording-configuration-and-runtime-telemetry-productization-plan/reference/01-diary.md — Diary explaining why this ticket exists and how it was scoped
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0004--screencast-studio-recording-configuration-and-runtime-telemetry-productization-plan/tasks.md — Detailed task plan for the intern implementation

## 2026-04-09

Replaced the last fake mounted recording-config state with real DSL-backed setup-draft ownership, rendered backend-resolved planned outputs in the output panel, wired the microphone selector to discovered devices, and removed the obsolete `studioDraft` slice.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/pages/StudioPage.tsx — Mounted page now edits structured draft state and renders compile-preview outputs
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup/setupSlice.ts — Stores compiled output preview state alongside normalize state
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/setup-draft/setupDraftSlice.ts — Owns recording name, template edits, and primary audio-source updates
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/OutputPanel.tsx — Shows name, destination, and planned-output preview
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/MicPanel.tsx — Uses discovered device options instead of hardcoded placeholder inputs

## 2026-04-10

Added schema-first runtime telemetry for the mounted studio: protobuf audio-meter and disk-status events, a server-side telemetry manager that follows the latest compiled setup, websocket delivery of the generated messages, and frontend rendering of real disk state plus live mic availability.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/proto/screencast/studio/v1/web.proto — Added audio-meter and disk-status protobuf messages and websocket event variants
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/telemetry_manager.go — New server-owned telemetry manager with cancellable audio and disk loops
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/handlers_ws.go — Sends initial telemetry snapshots to websocket clients
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/features/session/wsClient.ts — Consumes generated telemetry websocket events
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/components/studio/StatusPanel.tsx — Renders real disk percentage and capacity context

## 2026-04-10

Finished the `SCS-0004` hardening pass by asserting websocket telemetry events in backend tests, adding Storybook states for unavailable meter and invalid destination preview, and closing the remaining validation tasks with a successful live browser smoke against the rebuilt server.

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server_test.go — Websocket test now asserts initial audio-meter and disk-status events
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/MicPanel.stories.tsx — Added explicit unavailable-meter story
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/src/stories/OutputPanel.stories.tsx — Added invalid-destination preview story
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/09/SCS-0004--screencast-studio-recording-configuration-and-runtime-telemetry-productization-plan/tasks.md — Ticket tasks now fully checked off
