# Tasks

## Research and Delivery

- [x] Create ticket SCS-0013 for integrating the transcription-go pipeline into screencast-studio
- [x] Study the current screencast-studio recording/websocket/UI seams relevant to live transcription
- [x] Study the transcription-go prototype repository, especially the ASR service, chunk transport, websocket transport, accumulator, and output sinks
- [x] Write a detailed intern-facing analysis / design / implementation guide in the ticket
- [x] Write and maintain the investigation diary for this ticket
- [x] Validate docmgr metadata and vocabulary with `docmgr doctor --ticket SCS-0013 --stale-after 30`
- [x] Upload the final document bundle to reMarkable under `/ai/2026/04/13/SCS-0013`

## Recommended Implementation Phases

- [ ] Phase 0: Add a transcription seam inside screencast-studio (`pkg/transcription`) with backend/session/update types and a fake backend for tests
- [ ] Phase 0a: Extend `proto/screencast/studio/v1/web.proto` with transcription event messages and regenerate web protobuf outputs
- [ ] Phase 0b: Extend the frontend websocket client and session slice to parse and store transcription updates
- [ ] Phase 1: Add a server-side transcription state holder (recommended: `TranscriptionManager`) and include transcript bootstrap state on `/ws`
- [ ] Phase 2: Add a GStreamer transcription branch to the recording audio graph using a tee + appsink on normalized mono PCM
- [ ] Phase 2a: Keep existing recording file output and audio meter behavior intact while the transcription branch is attached
- [ ] Phase 2b: Add a focused GStreamer smoke script proving that the transcription branch receives 16kHz mono PCM without breaking recording finalization
- [ ] Phase 3: Implement a transcription backend adapter that speaks the transcription-go service protocol
- [ ] Phase 3a: Support websocket transport (`/transcribe/stream`) as the primary live mode
- [ ] Phase 3b: Optionally support chunk HTTP transport (`/transcribe/chunk`) as a debug fallback
- [ ] Phase 4: Feed backend transcript updates into screencast-studio server events and publish them over the existing `/ws` transport
- [ ] Phase 4a: Distinguish pending/partial words from committed/final words in server-side state
- [ ] Phase 4b: Handle browser reconnect/bootstrap so current transcript state is not lost on refresh
- [ ] Phase 5: Add a minimal live transcript UI to the studio page
- [ ] Phase 5a: Show committed transcript text
- [ ] Phase 5b: Optionally show a muted pending/partial tail
- [ ] Phase 6: Persist transcript sidecar outputs (TXT/SRT/VTT/SQLite) at recording end
- [ ] Phase 7: Add fake/integration tests and operator docs for the end-to-end transcription path

## Key Design Decisions Captured in This Ticket

- [x] Prefer protocol-level integration with transcription-go over direct imports from its `internal/...` Go packages
- [x] Prefer a GStreamer appsink PCM branch over temp-file polling inside screencast-studio
- [x] Prefer server-mediated transcript events to the browser over direct browser access to the ASR backend
- [x] Prefer websocket streaming transport as the target architecture, with chunk HTTP as a fallback/debug mode
- [x] Prefer transcript sidecar outputs first, and only later consider first-class DSL/compiler support if product needs it
