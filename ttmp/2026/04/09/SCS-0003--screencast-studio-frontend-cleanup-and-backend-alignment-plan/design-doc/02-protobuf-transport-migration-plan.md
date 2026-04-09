---
Title: Protobuf Transport Migration Plan
Ticket: SCS-0003
Status: active
Topics:
    - frontend
    - backend
    - ui
    - architecture
    - dsl
    - video
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/web/api_types.go
      Note: Current handwritten Go REST and websocket transport structs that should be replaced by generated protobuf types
    - Path: internal/web/handlers_api.go
      Note: Current REST handlers that will switch to protojson request and response helpers
    - Path: internal/web/handlers_preview.go
      Note: Preview REST handlers that will move to generated protobuf messages
    - Path: internal/web/handlers_ws.go
      Note: Websocket bootstrap and event streaming code that will move to generated protobuf event messages
    - Path: ui/src/api/types.ts
      Note: Current handwritten frontend transport declarations that should stop being the source of truth
    - Path: ui/src/api/recordingApi.ts
      Note: RTK Query layer that will decode generated protobuf JSON instead of handwritten interfaces
    - Path: ui/src/api/setupApi.ts
      Note: Setup transport layer that will move to generated protobuf messages
    - Path: ui/src/api/previewsApi.ts
      Note: Preview transport layer that will move to generated protobuf messages
    - Path: ui/src/features/session/wsClient.ts
      Note: Frontend websocket decoder and dispatch path that should stop relying on local schema logic
ExternalSources: []
Summary: Detailed analysis and implementation plan for replacing handwritten shared REST and websocket transport types with protobuf-generated Go and TypeScript code.
LastUpdated: 2026-04-09T18:02:00-04:00
WhatFor: Explain why the transport contract should move to protobuf and give an intern a concrete migration sequence for REST and websocket payloads.
WhenToUse: Read before implementing the protobuf migration, reviewing transport PRs, or deciding whether to add more handwritten contract code.
---

# Protobuf Transport Migration Plan

## Executive Summary

The current transport layer is declared by hand in multiple places:

- Go structs in `internal/web/api_types.go`
- TypeScript interfaces in `ui/src/api/types.ts`
- ad hoc frontend runtime validation whenever the TypeScript declarations are not trusted enough for websocket input

That is the opposite of the cleanup direction this ticket wants. The project is still small enough that we should consolidate the contract now instead of adding a better-looking version of the same duplication.

The proposed solution is to define the shared REST and websocket contract once in protobuf, generate Go and TypeScript code from that schema, and keep JSON as the actual wire format. Go will use `protojson` for request decode, response encode, and websocket event emission. The frontend will use generated protobuf schemas plus `@bufbuild/protobuf` JSON decode helpers instead of handwritten interface declarations and validators.

This plan intentionally does not switch the app to binary protobuf, gRPC, or Connect. It is a contract-consolidation step, not a transport-stack rewrite.

## Implementation Update

The migration described in this document is now the active transport path in the repo.

The schema lives in `proto/screencast/studio/v1/web.proto`. Generated outputs are committed for both languages:

- Go: `gen/go/proto/screencast/studio/v1/web.pb.go`
- TypeScript: `ui/src/gen/proto/screencast/studio/v1/web_pb.ts`

The backend now uses explicit boundary helpers in `internal/web/`:

- `pb_mapping.go` maps domain values into generated transport messages
- `protojson.go` provides shared `protojson` request and response helpers
- `handlers_api.go`, `handlers_preview.go`, and `handlers_ws.go` emit generated protobuf JSON instead of handwritten transport structs

The frontend now uses generated schemas as the transport source of truth:

- `ui/src/api/proto.ts` provides the small `fromJson(...)` / `toJson(...)` adapter boundary
- `ui/src/api/*.ts` decodes RTK Query responses through generated schemas
- `ui/src/features/session/wsClient.ts` decodes websocket frames as generated `ServerEvent` messages

The old handwritten shared transport file `internal/web/api_types.go` has been removed. `ui/src/api/types.ts` remains only as a re-export convenience layer around generated message types plus the local API error envelope.

## Problem Statement

The frontend cleanup effort has reached a point where transport duplication is the main architectural drag. The visual and composition work in `ui/` is now good enough that the biggest remaining confusion is not “which button should exist,” but “which transport contract is real.”

Three specific problems follow from the current arrangement.

### Problem 1: The Same Model Is Declared More Than Once

The same entities exist in both the Go backend and the TypeScript frontend:

- discovery descriptors
- normalize and compile responses
- recording session state
- process logs
- preview descriptors
- websocket events

Each time we edit a field, the same change has to be made by hand in Go and TypeScript. That is manageable for one field, but it becomes expensive and brittle once the app starts moving faster.

### Problem 2: Websocket Runtime Safety Pushes Us Toward Handwritten Validators

TypeScript interfaces are compile-time only. Once websocket frames arrive, the frontend must either trust raw JSON or add runtime validators. A validator path like this:

```ts
const isPreviewDescriptor = (value: unknown): value is PreviewDescriptor =>
  isObject(value) &&
  typeof value.id === 'string' &&
  typeof value.source_id === 'string';
```

is a second schema language. If we keep going in that direction, we will end up with:

- Go schema
- TS interface schema
- TS validator schema

That is exactly the wrong simplification.

### Problem 3: REST And Websocket Payloads Are One Shared API Surface

The same domain-shaped values appear across REST and websocket boundaries. If websocket goes schema-first but REST remains handwritten, we still keep one large source of drift alive. The right move is to migrate both together while the transport surface is still relatively small.

## Proposed Solution

Introduce a shared protobuf schema under `proto/` for the web transport contract, generate Go and TypeScript code from it, and convert both REST and websocket codepaths to those generated types.

### Scope

The protobuf migration should cover the backend/frontend transport surface used by `internal/web/` and `ui/`:

- health response
- discovery response
- setup normalize request and response
- setup compile response
- recording start request
- session envelope and session state
- preview ensure and release requests
- preview descriptor, preview envelope, and preview list response
- websocket server events

### Non-goals

- switching to binary protobuf transport
- rewriting `pkg/` domain models around protobuf
- moving internal recorder state to generated types
- introducing backwards compatibility aliases for the current snake_case JSON contract

### Target Layout

```text
proto/
  screencast/studio/v1/web.proto

buf.yaml
buf.gen.yaml

gen/go/proto/
  proto/screencast/studio/v1/web.pb.go

ui/src/gen/proto/
  proto/screencast/studio/v1/web_pb.ts
```

### Wire Format Strategy

The wire format remains JSON for this phase.

- Go decode: `protojson.Unmarshal`
- Go encode: `protojson.Marshal`
- websocket emit: `protojson.Marshal`
- TS decode: `fromJson(...)`
- TS encode: `toJson(...)`

This keeps the app easy to debug with browser devtools and test fixtures while removing the duplicate handwritten schema.

### Naming Strategy

Use protobuf’s default JSON naming, which means lowerCamelCase JSON keys. That means the app will intentionally move from snake_case request and response fields to camelCase.

Examples:

- `session_id` becomes `sessionId`
- `source_id` becomes `sourceId`
- `preview_id` becomes `previewId`
- `has_frame` becomes `hasFrame`

No compatibility layer is needed. The frontend and backend will be migrated together.

### Websocket Event Strategy

Use a protobuf `oneof` instead of a handwritten event string plus free-form payload.

Pseudocode:

```proto
message ServerEvent {
  google.protobuf.Timestamp timestamp = 1;

  oneof kind {
    RecordingSession session_state = 10;
    ProcessLog session_log = 11;
    PreviewListResponse preview_list = 12;
    PreviewDescriptor preview_state = 13;
    ProcessLog preview_log = 14;
  }
}
```

This makes the event type part of the generated schema and removes the need for frontend string-dispatch plus local shape validation.

### Go Boundary Strategy

The protobuf layer should live at the web boundary, not inside the core domain packages.

Good boundary:

```text
pkg/* domain types
  -> internal/web mapping helpers
  -> generated protobuf messages
  -> protojson
```

Bad boundary:

- importing generated protobuf messages directly into `pkg/discovery`
- rewriting domain structs in `pkg/dsl`
- scattering transport-message construction across every handler

### TypeScript Boundary Strategy

The frontend should stop treating `ui/src/api/types.ts` as the source of truth for shared transport messages. Instead:

- generated protobuf schemas live under `ui/src/gen/proto/`
- small adapter helpers live under `ui/src/api/`
- RTK Query endpoints decode with generated schemas
- websocket decode uses the generated event schema

This still allows small local view-model transforms for presentation, but it removes handwritten shared message declarations.

## Design Decisions

### Decision 1: Use Protobuf For Shared Schema, Not For A New RPC Framework

Rationale:

- we already have a working REST plus websocket transport
- the problem is contract duplication, not lack of an RPC stack
- this keeps the migration focused and reviewable

### Decision 2: Migrate REST And Websocket Together

Rationale:

- they share the same entities
- websocket-only migration leaves REST duplication in place
- REST-only migration leaves the worst runtime-decode pressure unresolved

### Decision 3: Keep JSON On The Wire For Now

Rationale:

- easier debugging
- easier incremental migration
- enough to eliminate duplicate schema code

### Decision 4: Use LowerCamelCase JSON And Do Not Preserve Snake Case

Rationale:

- protobuf JSON defaults are simpler than fighting them
- generated TS code and JSON names align naturally
- the ticket explicitly says no compatibility layer is needed

### Decision 5: Keep Transport Mapping Explicit At The Web Layer

Rationale:

- clearer ownership
- less leakage of transport concerns into domain packages
- easier review of API changes

### Decision 6: Use Generated Runtime Decode In The Frontend

Rationale:

- avoids growing another handwritten validation layer
- gives runtime confidence and static typing from the same schema
- keeps websocket handling small and explicit

## Alternatives Considered

### Alternative 1: Keep Handwritten TypeScript Types And Add Manual Validators

Rejected because it adds a third schema layer and does not solve the REST duplication problem.

### Alternative 2: Use JSON Schema

Rejected for now because this migration needs code generation for both Go and TS as much as it needs validation. Protobuf already has a clean toolchain available locally.

### Alternative 3: Migrate Only Websocket Events To Protobuf

Rejected because the same payloads appear in REST and websocket flows. Splitting the contract strategy would reduce clarity.

### Alternative 4: Move Internal Domain Packages To Generated Messages

Rejected because it would make the transport layer infect the internal model. The better boundary is explicit mapping at `internal/web/`.

### Alternative 5: Switch To gRPC-Web Or Connect Immediately

Rejected because that is a much larger transport rewrite than this ticket needs. The immediate problem is duplicate schema code, not lack of a new protocol stack.

## Implementation Plan

### Phase 1: Author The Shared Schema

Create `proto/screencast/studio/v1/web.proto`.

Messages to include:

- `HealthResponse`
- `DiscoveryResponse` and its nested descriptors
- `DslRequest`
- `NormalizeResponse`
- `CompileResponse`
- `RecordingStartRequest`
- `RecordingSession`
- `SessionEnvelope`
- `PreviewEnsureRequest`
- `PreviewReleaseRequest`
- `PreviewDescriptor`
- `PreviewEnvelope`
- `PreviewListResponse`
- `ProcessLog`
- `ServerEvent`

Intern note:

- mirror the current semantics first
- do not redesign the product model while writing the schema

### Phase 2: Add Generation Tooling

Add:

- `buf.yaml`
- `buf.gen.yaml`
- generated output directories for Go and TS
- one reproducible generation entrypoint

Dependencies:

- Go: `google.golang.org/protobuf`
- UI: `@bufbuild/protobuf`

Intern note:

- use Buf v2 remote plugins
- keep generation config checked in so CI and local dev can run the same command

### Phase 3: Migrate Go REST Handlers

Introduce helper functions such as:

```go
func decodeProtoJSON(r *http.Request, msg proto.Message) error
func writeProtoJSON(w http.ResponseWriter, status int, msg proto.Message)
```

Then migrate:

- `internal/web/handlers_api.go`
- `internal/web/handlers_preview.go`
- health response handling in `internal/web/server.go`

Also add focused mapping helpers, for example:

```go
func mapPreviewSnapshot(snapshot previewSnapshot) *studio_webv1.PreviewDescriptor
```

### Phase 4: Migrate Go Websocket Events

Replace the handwritten websocket event emission with generated `ServerEvent` messages.

Important invariant:

- initial websocket bootstrap messages and live event fan-out should use the same generated event schema

### Phase 5: Migrate Frontend REST Decode And Encode

Replace `ui/src/api/types.ts` as the source of truth.

The desired shape is:

- generated schemas under `ui/src/gen/proto/`
- thin transport helpers under `ui/src/api/`
- RTK Query endpoints that use generated decode and encode helpers

Example flow:

```text
request body object
  -> create generated message
  -> toJson(...)
  -> fetch
  -> fromJson(...)
  -> app view model
```

### Phase 6: Migrate Frontend Websocket Decode

Desired flow:

```text
message.data
  -> JSON.parse
  -> fromJson(ServerEventSchema, parsed)
  -> switch on event.kind.case
  -> dispatch Redux actions
```

This should replace handwritten event interfaces and ad hoc websocket payload guards.

### Phase 7: Delete The Handwritten Shared Contract

Delete or shrink:

- shared transport declarations in `ui/src/api/types.ts`
- frontend websocket validation helpers that only exist to reconstruct the schema manually
- redundant shared Go structs in `internal/web/api_types.go`

Keep only:

- non-shared UI view models
- error helpers if they are intentionally not moved to protobuf yet
- mapping functions between domain and transport

## Intern Review Guide

Review the migration in this order:

1. `proto/` schema
2. Buf config
3. generated outputs
4. Go mapping helpers
5. Go REST handlers
6. Go websocket event emission
7. TS transport helpers
8. RTK Query consumers
9. websocket client

This review order makes drift obvious quickly.

## Testing Checklist

- `buf generate`
- `go test ./...`
- `go build ./...`
- `pnpm --dir ui lint`
- `pnpm --dir ui build`
- `pnpm --dir ui build-storybook`
- manual smoke test against the real server:
  - load discovery
  - normalize DSL
  - compile DSL
  - start recording
  - stop recording
  - observe websocket bootstrap events
  - ensure and release a preview

## Open Questions

- Should API errors also move to protobuf now, or remain handwritten JSON for the first migration slice? Recommendation: keeping errors handwritten for the first pass is acceptable if it reduces risk.
- Should we wire generation into `go generate ./...` immediately or first land `buf generate` plus documentation? Recommendation: use at least one committed reproducible generation command in this ticket.
- Should websocket events expose only the `oneof` case, or also carry a duplicate event-name string? Recommendation: `oneof` alone is sufficient.

## References

- `internal/web/api_types.go`
- `internal/web/handlers_api.go`
- `internal/web/handlers_preview.go`
- `internal/web/handlers_ws.go`
- `ui/src/api/types.ts`
- `ui/src/api/setupApi.ts`
- `ui/src/api/recordingApi.ts`
- `ui/src/api/previewsApi.ts`
- `/home/manuel/.codex/skills/protobuf-go-ts-schema-exchange/SKILL.md`
- `/home/manuel/.codex/skills/protobuf-go-ts-schema-exchange/references/templates.md`
