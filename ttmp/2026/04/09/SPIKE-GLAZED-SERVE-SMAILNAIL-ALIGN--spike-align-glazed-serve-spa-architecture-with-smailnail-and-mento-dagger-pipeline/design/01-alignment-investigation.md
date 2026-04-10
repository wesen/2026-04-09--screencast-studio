---
title: "Alignment Investigation: Glazed Serve vs Smailnail vs Mento Dagger Pipeline"
ticket: SPIKE-GLAZED-SERVE-SMAILNAIL-ALIGN
doc-type: analysis
topics:
  - architecture
  - web-frontend
  - go-embedding
  - dagger
  - spa
status: draft
created: 2026-04-10
related:
  - "[[/home/manuel/code/wesen/obsidian-vault/Projects/2026/04/08/PROJ - Glazed Serve - Help Browser, Embedded Docs, and SPA]]"
  - "[[/home/manuel/code/mento/go-go-mento/docs/infra/dagger-build-pipeline.md]]"
---

# Alignment Investigation: Glazed Serve vs Smailnail vs Mento Dagger Pipeline

## Executive Summary

This investigation compares three implementations of the "Go binary with embedded SPA" pattern:

1. **Glazed Serve** (`/home/manuel/code/wesen/corporate-headquarters/glazed`) — A help browser HTTP server built into the `glaze` CLI
2. **Smailnail** (`/home/manuel/code/wesen/corporate-headquarters/smailnail`) — An email annotation/management tool with embedded UI
3. **Mento Dagger Pipeline** (`/home/manuel/code/mento/go-go-mento/docs/infra/dagger-build-pipeline.md`) — A documented pattern for Dagger-based web builds

The goal is to determine how well Glazed Serve aligns with the established Smailnail implementation and the documented Mento pipeline, and to identify concrete steps for alignment.

## Source Systems

### 1. Glazed Serve

**Location:** `/home/manuel/code/wesen/corporate-headquarters/glazed`

**Purpose:** Browser-facing help system for the Glazed CLI. The `glaze serve` command starts an HTTP server that:
- Exposes a REST API backed by SQLite help store
- Renders documentation through an embedded React SPA
- Supports explicit path loading for user-supplied markdown docs

**Key Architecture:**

```
web/ (Vite + React)
  -> cmd/build-web (Dagger-based)
  -> pkg/web/dist (built assets)
  -> pkg/web/static.go (//go:embed)
  -> glaze serve (embedded SPA handler)
```

**SPA Handler:** `pkg/web/static.go`
```go
//go:embed dist
var FS embed.FS

func NewSPAHandler() (http.Handler, error) {
    sub, err := fs.Sub(FS, "dist")
    // ... serves index.html for unknown paths (SPA fallback)
}
```

**Build Web:** `cmd/build-web/main.go`
- Dagger-based build inside `node:22` container
- Falls back to local `pnpm` if Dagger unavailable
- Exports to `pkg/web/dist`

### 2. Smailnail

**Location:** `/home/manuel/code/wesen/corporate-headquarters/smailnail`

**Purpose:** Email annotation and management tool with a React-based UI embedded in the Go binary.

**Key Architecture:**

```
ui/ (Vite + React + TypeScript)
  -> cmd/build-web (Dagger-based)
  -> pkg/smailnaild/web/embed/public (built assets)
  -> pkg/smailnaild/web/embed.go (//go:embed embed/public)
  -> Smailnail HTTP server
```

**SPA Handler:** `pkg/smailnaild/web/spa.go`
```go
func RegisterSPA(mux *http.ServeMux, publicFS fs.FS, opts SPAOptions) {
    // API prefix handling + SPA fallback
}
```

**Embedding:** `pkg/smailnaild/web/embed.go`
```go
//go:build embed
package web

//go:embed embed/public
var embeddedFS embed.FS

var PublicFS, _ = fs.Sub(embeddedFS, "embed/public")
```

**Build Web:** `cmd/build-web/main.go`
- Uses `node:22-bookworm` as base
- Dagger with cache volumes for pnpm store
- Exports to `pkg/smailnaild/web/embed/public`

### 3. Mento Dagger Pipeline

**Location:** `/home/manuel/code/mento/go-go-mento/docs/infra/dagger-build-pipeline.md`

**Purpose:** Documented pattern for reproducible Dagger-based web builds with:

- **PNPM Builder Base (PBB):** Pre-activated pnpm on `node:22`, published by digest
- **Dagger Builder Image (DBI):** Contains precompiled `build-web` generator
- **Composite Actions:** `prepare-dagger-builder` and `run-web-with-dagger`
- **Digest-first pattern:** Immutable references for reproducibility
- **Offline pnpm:** Host-mounted cache for `pnpm store`

## Alignment Analysis

### Category 1: SPA Embedding Strategy

| Aspect | Glazed | Smailnail | Mento Pipeline |
|--------|--------|-----------|----------------|
| Embed path | `pkg/web/dist` | `pkg/smailnaild/web/embed/public` | `go/cmd/frontend/dist` |
| Embed directive | `//go:embed dist` | `//go:embed embed/public` | `//go:embed dist` |
| Build output dir | `pkg/web/dist` | `pkg/smailnaild/web/embed/public` | `go/cmd/frontend/dist` |
| SPA handler location | `pkg/web/static.go` | `pkg/smailnaild/web/spa.go` | `go/cmd/frontend/main.go` |
| API prefix handling | Composed handler | `RegisterSPA` with `APIPrefix` option | Not specified |
| Index fallback | Yes (manual) | Yes (via RegisterSPA) | Yes |

**Assessment:** ✅ Both Glazed and Smailnail follow the same pattern. Smailnail has a slightly cleaner separation with `RegisterSPA` that accepts API prefix configuration.

**Gap:** Glazed's `NewSPAHandler` doesn't have explicit API prefix handling — it composes the API handler separately via `NewServeHandler`.

### Category 2: Build Pipeline

| Aspect | Glazed | Smailnail | Mento Pipeline |
|--------|--------|-----------|----------------|
| Base image | `node:22` | `node:22-bookworm` | Separate PBB + DBI |
| Dagger usage | Direct SDK call | Direct SDK call | Composite actions |
| pnpm version | `10.15.0` (env) | `10.15.0` (env) | From `package.json` |
| Cache strategy | None (Dagger managed) | `CacheVolume` for pnpm store | Host-mounted `/pnpm/store` |
| Local fallback | Yes (pnpm) | No | Yes (via generator) |
| go:generate | Yes (`pkg/web/gen.go`) | Yes (`pkg/smailnaild/web/generate.go`) | Yes |

**Assessment:** ⚠️ Smailnail and Mento have more sophisticated caching. Glazed's `cmd/build-web` is simpler but lacks cache persistence across runs.

**Gap:** Glazed should consider:
1. Adding `CacheVolume` for pnpm store (like Smailnail)
2. Using a tagged pnpm version from `package.json` (like Mento)

### Category 3: File Organization

| Aspect | Glazed | Smailnail | Mento Pipeline |
|--------|--------|-----------|----------------|
| Frontend source | `web/` | `ui/` | `web/` (per repo) |
| Go package | `pkg/web` | `pkg/smailnaild/web` | `go/cmd/frontend` |
| Build cmd | `cmd/build-web` | `cmd/build-web` | `go/cmd/dagger/build-web` |
| Embedding package | `pkg/web` | `pkg/smailnaild/web/embed` | `go/cmd/frontend` |

**Assessment:** ✅ Smailnail has the cleanest separation with `embed/` as a subpackage. Glazed co-locates everything in `pkg/web`.

**Gap:** Glazed could consider separating `pkg/web/embed/` from `pkg/web/static.go`, but this is a minor organizational preference.

### Category 4: HTTP Server Integration

| Aspect | Glazed | Smailnail |
|--------|--------|-----------|
| Server architecture | Standalone `serve` command | Registered on `http.ServeMux` |
| API route handling | Composed handler with path check | API prefix option in `RegisterSPA` |
| CORS | Applied internally by `NewHandler` | Not shown (may be elsewhere) |
| Mount prefix | `MountPrefix()` helper function | Not shown |

**Assessment:** ⚠️ Different integration styles. Glazed is designed to run as a standalone server; Smailnail registers handlers on an existing mux.

**Gap:** If Glazed ever needs to be embedded in a larger server, it would benefit from a `RegisterSPA`-style API similar to Smailnail.

## Specific File Comparisons

### `cmd/build-web/main.go`

**Glazed:**
```go
func buildWithDagger(webPath, outPath, pnpmVersion, builderImage string) error {
    // Uses dagger.Connect + container.From + WithMountedDirectory
    // No persistent cache
}
```

**Smailnail:**
```go
// Uses dagger.Connect with LogOutput
// Has CacheVolume for pnpm store
pnpmStore := client.CacheVolume("smailnail-ui-pnpm-store")
// ...
WithMountedCache("/pnpm/store", pnpmStore)
```

**Mento:**
```go
// Composite actions for prepare + run
// Immutable digest references
// Host-mounted pnpm cache
```

**Recommendation:** Glazed should add `CacheVolume` support like Smailnail for faster rebuilds.

### `pkg/web/static.go` vs `pkg/smailnaild/web/spa.go`

**Glazed (static.go):**
```go
func NewSPAHandler() (http.Handler, error) {
    // Returns handler directly, API is composed elsewhere
}
```

**Smailnail (spa.go):**
```go
func RegisterSPA(mux *http.ServeMux, publicFS fs.FS, opts SPAOptions) {
    // Registers on mux, has APIPrefix option
}
```

**Recommendation:** Glazed could add an `APIPrefix` parameter to `NewSPAHandler` or create a `RegisterSPA` variant.

## Alignment Gaps Summary

| Priority | Gap | Effort |
|----------|-----|--------|
| Medium | Add `CacheVolume` for pnpm store in Glazed's `build-web` | Low |
| Low | Extract pnpm version from `package.json` (like Mento) | Low |
| Low | Add `APIPrefix` option to Glazed's SPA handler | Low |
| Low | Consider separating `embed/` subpackage (like Smailnail) | Medium |

## Recommendations

### 1. Add Persistent pnpm Cache to Glazed

Replace the current `buildWithDagger` with cache volume support:

```go
// Add to buildWithDagger:
pnpmStore := client.CacheVolume("glazed-ui-pnpm-store")
ctr = ctr.WithMountedCache("/pnpm/store", pnpmStore)
```

This aligns Glazed with Smailnail's approach and will significantly speed up repeated builds.

### 2. Consider Extract pnpm Version from package.json

The Mento pipeline reads pnpm version from `packageManager` field:
```json
"packageManager": "pnpm@10.15.0"
```

Glazed's `web/package.json` doesn't have this field. Adding it would enable Mento-style "prepare once" builder images.

### 3. Keep Current Embedding Structure

The `//go:embed dist` approach in Glazed is correct and matches the Mento pattern. No structural changes needed.

### 4. Document the Pattern

Both Glazed and Smailnail would benefit from having their patterns documented similarly to the Mento pipeline doc. This creates institutional knowledge and makes the pattern transferable to other projects.

## Conclusion

**Glazed Serve is well-aligned with Smailnail and the Mento Dagger pipeline.** All three implementations follow the same fundamental pattern:

1. Build frontend with Dagger (or local fallback)
2. Export to a specific directory in the Go repo
3. Embed with `//go:embed`
4. Serve via HTTP handler with SPA fallback

The main opportunities for alignment are:
- **Caching:** Add `CacheVolume` for pnpm store (easy)
- **Configuration:** Read pnpm version from `package.json` (easy)
- **API:** Consider `RegisterSPA`-style API for mux integration (if needed)

None of these are blockers — the current architecture is sound and proven across three independent implementations.

---

## Related Files

| File | Purpose |
|------|---------|
| `/home/manuel/code/wesen/corporate-headquarters/glazed/cmd/build-web/main.go` | Glazed Dagger builder |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/web/static.go` | Glazed SPA handler |
| `/home/manuel/code/wesen/corporate-headquarters/smailnail/cmd/build-web/main.go` | Smailnail Dagger builder |
| `/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/smailnaild/web/spa.go` | Smailnail SPA handler |
| `/home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/smailnaild/web/embed.go` | Smailnail embed directive |
| `/home/manuel/code/mento/go-go-mento/docs/infra/dagger-build-pipeline.md` | Mento pipeline documentation |
