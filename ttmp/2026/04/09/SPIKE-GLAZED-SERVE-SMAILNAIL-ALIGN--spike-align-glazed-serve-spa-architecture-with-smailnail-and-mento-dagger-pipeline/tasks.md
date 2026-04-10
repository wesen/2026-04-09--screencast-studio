# Tasks

## Implementation

- [x] Add `packageManager` to `ui/package.json`.
- [x] Add `cmd/build-web/main.go` for Dagger-based SPA builds.
- [x] Add `dagger.WithLogOutput(os.Stdout)` for build visibility.
- [x] Add a Dagger `CacheVolume` for the pnpm store.
- [x] Extract the pnpm version from `ui/package.json`.
- [x] Use `pnpm install --prefer-offline --frozen-lockfile` in the Dagger build.
- [x] Add a local pnpm fallback if Dagger is unavailable.
- [x] Export build output into `internal/web/dist`.
- [x] Wire `internal/web/generate.go` so `go generate ./...` builds the UI before protobuf generation.
- [x] Add a `build-web` Make target.

## Serving Alignment

- [x] Add an internal web FS that serves generated assets from disk in normal dev builds.
- [x] Add an embed-tagged internal web FS for embedded SPA builds.
- [x] Change the root handler so it serves the built SPA before falling back to the placeholder page.
- [x] Preserve `--static-dir` as an explicit development override.
- [ ] Consider a `RegisterSPA`-style mux helper API if the web layer grows beyond the current single-server shape.

## Validation

- [x] Run `go run ./cmd/build-web` and verify Dagger logs plus export into `internal/web/dist`.
- [x] Run `go test ./...`.
- [x] Run `go build ./...`.
- [x] Run `go build -tags embed ./cmd/screencast-studio`.
- [x] Run `go run ./cmd/screencast-studio serve --addr :18081` without `--static-dir` and verify `/` serves the built SPA.

## Notes

- The repo now matches the intended Smailnail-style shape closely enough for normal development:
  - Dagger builds `ui/`
  - build output lands in `internal/web/dist`
  - normal builds can serve generated assets from disk
  - embed builds can serve the SPA from `//go:embed`
