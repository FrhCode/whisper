# Tasks: Native History GUI

- [ ] Task: Add history loader for GUI
  - Acceptance: History package can read `history.jsonl`, keep 50 newest valid entries, skip invalid lines, and return newest-first data plus skipped count without changing file format.
  - Verify: `go test ./internal/history`
  - Files: `internal/history/history.go`, `internal/history/history_test.go`

- [ ] Task: Add native history window
  - Acceptance: Windows-only history window opens with ListView, detail panel, Refresh button, Copy Cleaned button, status text, empty state, and double-click copy behavior.
  - Verify: Windows build compiles; manual open/refresh/select/copy flow works.
  - Files: `internal/historyui/*.go`

- [ ] Task: Wire tray menu to history window
  - Acceptance: Open History tray action launches GUI window instead of Notepad and preserves existing tray flow.
  - Verify: Manual tray click opens GUI, not Notepad.
  - Files: `cmd/whispr/main.go`

- [ ] Task: Add clipboard copy and status feedback
  - Acceptance: Copy button and double-click copy selected cleaned text to clipboard and show small success or empty/error status.
  - Verify: Manual copy then paste into target app returns cleaned text.
  - Files: `internal/historyui/*.go`

- [ ] Task: Verify final Windows build
  - Acceptance: Full Windows build succeeds with history GUI changes included.
  - Verify: `./scripts/build-windows.ps1`
  - Files: none
