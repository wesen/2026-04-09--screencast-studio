# Tasks

## TODO

- [ ] Add tasks here

- [ ] Add CacheVolume for pnpm store in Glazed cmd/build-web/main.go (align with Smailnail)
- [ ] Extract pnpm version from package.json packageManager field (align with Mento)
- [ ] Consider RegisterSPA-style API for mux integration (optional, low priority)
- [ ] Add packageManager field to glazed/web/package.json
- [ ] Step 3: Add CacheVolume for pnpm store - Add pnpmStore := client.CacheVolume("glazed-ui-pnpm-store") and ctr.WithMountedCache("/pnpm/store", pnpmStore) in buildWithDagger
- [ ] Step 6: Update main() to use getPnpmVersionFromPackageJSON with WEB_PNPM_VERSION env var as override
- [ ] Step 1: Add packageManager field to web/package.json - Add `"packageManager": "pnpm@10.15.0"` to the top-level JSON object
- [ ] Step 4: Update pnpm install to use --prefer-offline flag to leverage the CacheVolume
- [ ] Step 2: Add dagger.WithLogOutput(os.Stdout) to dagger.Connect call in cmd/build-web/main.go
- [ ] Step 5: Add getPnpmVersionFromPackageJSON helper function to extract version from package.json's packageManager field
- [ ] Step 8: Test - Run ./glaze serve and verify SPA loads at http://localhost:8088
- [ ] Step 9: Test - Run make build and verify binary is produced with embedded assets
- [ ] Step 7: Test - Run go run ./cmd/build-web and verify Dagger logs and successful build
