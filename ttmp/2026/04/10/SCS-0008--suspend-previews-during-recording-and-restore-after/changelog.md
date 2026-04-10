# Changelog

## 2026-04-10

- Created ticket `SCS-0008` for the preview handoff bug.
- Added a small implementation guide for suspending previews before recording and restoring them after.
- Started the implementation diary for the backend coordination work.
- Implemented preview suspend/restore handoff in commit `4634a75` (`recording: hand off previews around sessions`).
- Added focused preview-manager and recording lifecycle tests covering suspend, restore, and failed start recovery.
