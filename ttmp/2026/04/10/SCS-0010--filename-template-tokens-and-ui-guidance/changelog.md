# Changelog

## 2026-04-10

- Created ticket `SCS-0010` for filename-template tokens and mounted-UI guidance.
- Added a small implementation guide for backend `{index}` rendering, builder-managed filename suffixes, and inline output naming help.
- Implemented filename-suffix editing in structured mode and backend `{index}` rendering in commit `bab0885` (`output: add filename template tokens`).
- Validated the change with `go test ./... -count=1`, `pnpm --dir ui build`, and `go generate ./internal/web`.
- Extended token support to the `Name` field and clarified in the mounted UI that tokens apply to `Name`, `Save to`, and `Filename` in commit `06be6c1` (`output: expand name tokens`).
