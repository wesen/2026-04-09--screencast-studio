---
title: "Implementation Guide: Align Glazed Serve with Smailnail and Mento"
ticket: SPIKE-GLAZED-SERVE-SMAILNAIL-ALIGN
doc-type: implementation-guide
topics:
  - implementation
  - dagger
  - web-frontend
  - go-embedding
status: draft
created: 2026-04-10
prerequisites:
  - Go 1.21+
  - Dagger installed (`go install dagger.io/dagger/cmd/dagger@latest`)
  - pnpm installed (for local fallback)
---

# Implementation Guide: Align Glazed Serve with Smailnail and Mento

This guide provides step-by-step instructions for implementing the alignment improvements identified in the [Alignment Investigation](./01-alignment-investigation.md).

## Overview of Changes

| Step | Change | Effort | Priority |
|------|--------|--------|----------|
| 1 | Add `packageManager` field to `web/package.json` | 5 min | High |
| 2 | Add `CacheVolume` for pnpm store in `cmd/build-web/main.go` | 15 min | High |
| 3 | Extract pnpm version from `package.json` | 10 min | Medium |
| 4 | Test the changes | 10 min | High |

---

## Step 1: Add `packageManager` Field to `web/package.json`

**Purpose:** Enable Mento-style "prepare once" builder images and version pinning.

### 1.1 Read the current `web/package.json`

```bash
cat /home/manuel/code/wesen/corporate-headquarters/glazed/web/package.json
```

### 1.2 Add the `packageManager` field

Add the following field to the top level of the JSON object:

```json
"packageManager": "pnpm@10.15.0"
```

### 1.3 Verify the change

```bash
cat /home/manuel/code/wesen/corporate-headquarters/glazed/web/package.json
```

The file should now include the `packageManager` field at the same level as `name`, `version`, etc.

---

## Step 2: Add `CacheVolume` for pnpm Store in `cmd/build-web/main.go`

**Purpose:** Speed up repeated builds by caching the pnpm store across runs (aligns with Smailnail).

### 2.1 Read the current implementation

```bash
cat /home/manuel/code/wesen/corporate-headquarters/glazed/cmd/build-web/main.go
```

### 2.2 Identify the `buildWithDagger` function

The current implementation at lines ~60-80:

```go
func buildWithDagger(webPath, outPath, pnpmVersion, builderImage string) error {
    ctx := context.Background()
    client, err := dagger.Connect(ctx)
    // ...
    ctr := base.
        WithWorkdir("/src").
        WithMountedDirectory("/src", webDir).
        WithEnvVariable("PNPM_HOME", "/pnpm")
    // ...
}
```

### 2.3 Add `dagger.WithLogOutput(os.Stdout)` to the Connect call

Change line ~63 from:

```go
client, err := dagger.Connect(ctx)
```

To:

```go
client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
```

This provides better visibility into Dagger operations.

### 2.4 Add CacheVolume for pnpm store

After the container is created (after `base := client.Container().From(builderImage)`), add:

```go
// Add persistent pnpm store cache
pnpmStore := client.CacheVolume("glazed-ui-pnpm-store")
ctr := ctr.WithMountedCache("/pnpm/store", pnpmStore)
```

The complete `buildWithDagger` function should look like:

```go
func buildWithDagger(webPath, outPath, pnpmVersion, builderImage string) error {
    ctx := context.Background()
    client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
    if err != nil {
        return fmt.Errorf("connect dagger: %v", err)
    }
    defer func() { _ = client.Close() }()

    base := client.Container().From(builderImage)

    // Add persistent pnpm store cache
    pnpmStore := client.CacheVolume("glazed-ui-pnpm-store")
    ctr := base.
        WithWorkdir("/src").
        WithMountedDirectory("/src", client.Host().Directory(webPath)).
        WithEnvVariable("PNPM_HOME", "/pnpm").
        WithMountedCache("/pnpm/store", pnpmStore)

    ctr = ctr.WithExec([]string{
        "sh", "-lc",
        fmt.Sprintf("corepack enable && corepack prepare pnpm@%s --activate", pnpmVersion),
    })

    ctr = ctr.
        WithExec([]string{"sh", "-lc", "pnpm --version"}).
        WithExec([]string{"sh", "-lc", "pnpm install --prefer-offline"}).
        WithExec([]string{"sh", "-lc", "pnpm build"})

    dist := ctr.Directory("/src/dist")
    if _, err := dist.Export(ctx, outPath); err != nil {
        return fmt.Errorf("export dist to %s: %v", outPath, err)
    }
    return nil
}
```

### 2.5 Update import for `os`

Ensure `os` is imported (it likely already is):

```go
import (
    "context"
    "fmt"
    "log"
    "os"              // Already present
    "os/exec"         // Already present
    "path/filepath"

    "dagger.io/dagger"
)
```

---

## Step 3: Extract pnpm Version from `package.json`

**Purpose:** Use the `packageManager` field we added in Step 1 instead of hardcoding.

### 3.1 Add a helper function to read pnpm version

Add this function to `cmd/build-web/main.go`:

```go
// getPnpmVersionFromPackageJSON reads the pnpm version from package.json's packageManager field.
func getPnpmVersionFromPackageJSON(webPath string) (string, error) {
    pkgJSONPath := filepath.Join(webPath, "package.json")
    data, err := os.ReadFile(pkgJSONPath)
    if err != nil {
        return "", fmt.Errorf("read package.json: %w", err)
    }

    // Simple regex extraction to avoid adding json unmarshal dependency
    re := regexp.MustCompile(`"packageManager"\s*:\s*"pnpm@([^"]+)"`)
    matches := re.FindSubmatch(data)
    if len(matches) < 2 {
        return "", fmt.Errorf("packageManager field not found in package.json")
    }
    return string(matches[1]), nil
}
```

### 3.2 Add `regexp` to imports

```go
import (
    // ...
    "os/exec"
    "path/filepath"
    "regexp"           // Add this
    // ...
)
```

### 3.3 Update `main()` to use the helper

Change the pnpm version resolution in `main()` from:

```go
func main() {
    pnpmVersion := getenv("WEB_PNPM_VERSION", "10.15.0")
    builderImage := getenv("WEB_BUILDER_IMAGE", "node:22")
    // ...
}
```

To:

```go
func main() {
    // Try to get pnpm version from package.json first
    pnpmVersion := getenv("WEB_PNPM_VERSION", "")
    if pnpmVersion == "" {
        wd, _ := os.Getwd()
        repoRoot, _ := findRepoRoot(wd)
        webPath := filepath.Join(repoRoot, "web")
        if v, err := getPnpmVersionFromPackageJSON(webPath); err == nil {
            pnpmVersion = v
        } else {
            log.Printf("Could not read pnpm version from package.json: %v, using default", err)
            pnpmVersion = "10.15.0"
        }
    }

    builderImage := getenv("WEB_BUILDER_IMAGE", "node:22")
    // ...
}
```

---

## Step 4: Test the Changes

### 4.1 Build the web assets

```bash
cd /home/manuel/code/wesen/corporate-headquarters/glazed
go run ./cmd/build-web
```

Expected output:
- Dagger should connect and show logs (due to `dagger.WithLogOutput`)
- Build should complete successfully
- `pkg/web/dist` should be populated

### 4.2 Run the local fallback (no Dagger)

Kill the Dagger engine if running, then:

```bash
cd /home/manuel/code/wesen/corporate-headquarters/glazed
WEB_BUILDER_IMAGE="" go run ./cmd/build-web
```

Expected: Falls back to local pnpm.

### 4.3 Verify the binary embeds correctly

```bash
cd /home/manuel/code/wesen/corporate-headquarters/glazed
go build -tags sqlite_fts5 ./cmd/glaze
./glaze serve
```

Open http://localhost:8088 and verify:
- The SPA loads
- Navigation works
- API endpoints respond (`/api/health`, `/api/sections`)

### 4.4 Run the full Makefile build

```bash
cd /home/manuel/code/wesen/corporate-headquarters/glazed
make build
```

Verify `glaze` binary is produced with embedded assets.

---

## Verification Checklist

| Check | Command | Expected Result |
|-------|---------|-----------------|
| Build web via Dagger | `go run ./cmd/build-web` | Dagger logs visible, success |
| Build web via local pnpm | `WEB_BUILDER_IMAGE="" go run ./cmd/build-web` | Falls back to local pnpm |
| SPA serves correctly | `./glaze serve` + curl http://localhost:8088 | HTML page returned |
| API responds | `curl http://localhost:8088/api/health` | `{"status":"ok"}` |
| Full build | `make build` | Binary produced |

---

## Rollback Instructions

If something goes wrong:

### Revert `cmd/build-web/main.go`

```bash
cd /home/manuel/code/wesen/corporate-headquarters/glazed
git checkout cmd/build-web/main.go
```

### Revert `web/package.json`

```bash
cd /home/manuel/code/wesen/corporate-headquarters/glazed
git checkout web/package.json
```

### Clean build artifacts

```bash
cd /home/manuel/code/wesen/corporate-headquarters/glazed
rm -rf pkg/web/dist
go generate ./pkg/web
```

---

## Expected File Changes Summary

| File | Change |
|------|--------|
| `cmd/build-web/main.go` | Added CacheVolume, LogOutput, pnpm version extraction |
| `web/package.json` | Added `packageManager` field |

---

## Related Implementation References

| File | Purpose |
|------|---------|
| `/home/manuel/code/wesen/corporate-headquarters/smailnail/cmd/build-web/main.go` | Reference implementation with CacheVolume |
| `/home/manuel/code/mento/go-go-mento/docs/infra/dagger-build-pipeline.md` | Mento pipeline documentation |
