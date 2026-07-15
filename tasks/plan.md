# Plan: Native History GUI

## Goal
Replace history tray action that opens Notepad with native Windows GUI for browsing 50 newest valid history entries and copying cleaned text.

## Implementation Order
1. Add history loading helper that reads `history.jsonl`, keeps 50 newest valid entries, skips invalid JSONL lines, and returns newest-first entries plus skipped count.
2. Add unit test for valid parsing, invalid-line skip, 50-entry cap, and newest-first order.
3. Build Windows-only history window package with native controls.
4. Wire tray menu Open History to launch history window instead of Notepad.
5. Add refresh, detail panel, copy-cleaned button, double-click copy, empty state, and status feedback.
6. Verify build and manual UI flow on Windows.

## Components
### `internal/history`
- Keep `Entry` format unchanged.
- Add load helper for GUI consumption:
  - reads JSONL line-by-line
  - keeps 50 latest valid entries
  - reverses to newest-first
  - skips malformed lines
  - returns skipped count
- Add tests for parsing, newest-first ordering, cap behavior, empty file, and invalid lines.

### `internal/historyui`
- New Windows-only package.
- Use Win32 window class + message loop.
- Render controls:
  - ListView with columns `Time`, `Raw`, `Cleaned`
  - read-only cleaned text detail panel
  - `Copy Cleaned` button
  - `Refresh` button
  - status text for copied/empty/skipped feedback
- Load history on open and populate ListView.
- Refresh reloads same file path and repopulates ListView.
- Selection updates detail panel.
- Copy button and double-click row copy selected entry cleaned text to clipboard.

### `cmd/whispr/main.go`
- Replace Notepad launch with `historyui.Open("history.jsonl")`.
- Keep tray command flow otherwise unchanged.

## Risks
- Win32 control plumbing can be verbose and error-prone.
- Clipboard ownership must be handled carefully.
- ListView population may need explicit row index management.
- Windows-only code must not break Linux build stubs.
- Large `history.jsonl` could still cost disk scan time, because newest entries are at file end but JSONL has no index.

## Mitigation
- Isolate Windows code in one package with build tags.
- Keep history parsing separate from UI for testability.
- Reuse existing app style: minimal, direct, no framework.
- Keep only 50 entries in memory while scanning.
- Add one small test file for history load behavior.
- Avoid search/pagination/delete/export in v1.

## Verification Checkpoints
- After step 1-2: `go test ./internal/history`
- After step 3: compile Windows package via Windows build environment.
- After step 4: manual tray click opens GUI instead of Notepad.
- After step 5: manual verify refresh, selection detail, copy button, double-click copy, empty state, and skipped-invalid status.
- Final: Windows build via `./scripts/build-windows.ps1`

## Deferred
- Search.
- Pagination or load-more.
- Delete history.
- Edit history.
- Export history.
- Auto-refresh file watcher.
- Fancy styling.

## Notes
- No history file format changes.
- No new dependency unless Win32 binding gap appears in existing stdlib/x/sys support.
- Keep UI lightweight and maintainable rather than framework-based.
