# Changelog

## 2026-04-09

- Initial workspace created


## 2026-04-09

Created SPIKE-GLAZED-SERVE-SMAILNAIL-ALIGN ticket with alignment investigation comparing Glazed Serve, Smailnail, and Mento Dagger pipeline

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/ttmp/2026/04/09/SPIKE-GLAZED-SERVE-SMAILNAIL-ALIGN--spike-align-glazed-serve-spa-architecture-with-smailnail-and-mento-dagger-pipeline/design/01-alignment-investigation.md — Created detailed alignment analysis


## 2026-04-09

Added implementation guide (design/02-implementation-guide.md) with step-by-step instructions for adding CacheVolume, pnpm version extraction, and testing

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/ttmp/2026/04/09/SPIKE-GLAZED-SERVE-SMAILNAIL-ALIGN--spike-align-glazed-serve-spa-architecture-with-smailnail-and-mento-dagger-pipeline/design/02-implementation-guide.md — Implementation guide with 9 detailed steps

## 2026-04-09

Implemented the aligned web build and serve pipeline in this repo:

- added a Dagger-based `cmd/build-web` command with pnpm cache volume, package-manager version detection, and local fallback
- changed `go generate ./...` in `internal/web` to build the SPA before protobuf generation
- added disk and embed-backed frontend serving inside `internal/web`
- changed the root handler to serve the built SPA when available instead of the placeholder page
- added validation coverage for SPA fallback routing and verified the real server serves the built UI without `--static-dir`

### Related Files

- /home/manuel/code/wesen/2026-04-09--screencast-studio/cmd/build-web/main.go — New Dagger-based frontend build command
- /home/manuel/code/wesen/2026-04-09--screencast-studio/ui/package.json — Added `packageManager` pin for pnpm version discovery
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/generate.go — `go:generate` now builds the SPA before `buf generate`
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/embed.go — Embed-tagged SPA filesystem
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/embed_none.go — Disk-backed generated-asset filesystem for normal builds
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/static.go — SPA asset and index fallback helpers
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server.go — Root handler now serves static assets from disk or embed before placeholder fallback
- /home/manuel/code/wesen/2026-04-09--screencast-studio/internal/web/server_test.go — Added SPA-serving tests
- /home/manuel/code/wesen/2026-04-09--screencast-studio/Makefile — Added `build-web` target
