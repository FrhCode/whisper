# Spec: Native History GUI

## Objective
Build native Windows history window for Whispr so user can inspect recent history entries, see newest items first, and copy cleaned transcript manually without opening Notepad.

User story:
- When user clicks tray menu Open History, app opens GUI window instead of Notepad.
- Window shows 50 newest valid history entries.
- Each entry shows timestamp, raw text, and cleaned text.
- Selecting an entry shows full cleaned text in detail panel.
- User can copy cleaned text from selected entry with button or double-click.
- User can refresh history without closing the window.

Success looks like:
- No Notepad launch for history viewing.
- History entries are readable in dedicated GUI.
- Copy action copies cleaned text only.
- Refresh action reloads latest history without reopening window.
- Invalid JSONL lines are skipped, with status feedback.

## Assumptions
- History opens as separate standard Windows window.
- History supports single selected row only.
- History window loads 50 newest valid entries only.
- Copy action copies cleaned text only.
- Copy success feedback is small status text.
- UI has Refresh button and cleaned-text detail panel.
- Double-clicking row copies cleaned text.
- UI uses native Win32 controls, no GUI framework.

## Tech Stack
- Go 1.22
- Native Win32 UI via `golang.org/x/sys/windows`
- Windows common controls: ListView, Button, Static
- Existing systray app and Windows-only build tags
- Existing JSONL history file at `history.jsonl`

## Commands
Build:
```powershell
.\scripts\build-windows.ps1
```

Run:
```powershell
.\dist\whispr.exe
```

Test:
```powershell
go test ./...
```

Targeted package test:
```powershell
go test ./internal/history
```

## Project Structure
- `cmd/whispr/main.go` → tray menu, history action, window launch wiring
- `internal/history/history.go` → history entry model and file handling
- `internal/historyui/` → new native Win32 history window
- `history.jsonl` → runtime history source
- `tasks/spec.md` → this spec
- `tasks/plan.md` → implementation plan after approval
- `tasks/todo.md` → task breakdown after plan approval

## Code Style
Keep code small, Windows-specific, and close to current style.

Example:
```go
func openHistory() error {
	return launchHistoryWindow("history.jsonl")
}
```

Conventions:
- Prefer one responsibility per function.
- Keep Windows code behind `//go:build windows`.
- Use stdlib and `golang.org/x/sys/windows` before adding dependencies.
- Load 50 newest valid history entries and show newest-first.
- Skip invalid JSONL lines and report skipped count in status.
- Copy button and double-click copy cleaned field from selected row.

## Testing Strategy
- Add focused unit test for history loading/sorting if logic is isolated.
- Add targeted test for JSONL parsing if needed.
- Manual Windows verification for tray click, window open, refresh, selection, detail panel, double-click copy, and copy button.
- Keep tests near affected package.

## Boundaries
- Always: validate history file parsing, keep newest-first order, load only 50 newest valid entries, preserve existing history file format, use native Windows behavior, handle empty or missing history file safely, skip invalid JSONL lines safely.
- Ask first: add dependencies, change history JSON format, change tray menu wording, alter clipboard restore behavior, add search/pagination/delete/export.
- Never: reintroduce Notepad for history access, remove existing history fields, break non-Windows build stubs, commit secrets.

## Success Criteria
- Tray menu Open History opens GUI window.
- Window lists 50 newest valid history entries newest-first.
- Each row shows timestamp, raw, and cleaned text.
- Selecting a row shows full cleaned text in detail panel.
- Copy button copies cleaned text from selected entry.
- Double-clicking a row copies cleaned text from that entry.
- Refresh button reloads history from `history.jsonl`.
- Missing or empty history file opens empty window with “No history yet” status, not crash.
- Invalid JSONL lines are skipped and counted in status.
- Windows build still succeeds.

## Open Questions
- None. User approved single row, standard window, small copy feedback, 50 newest entries, refresh button, detail panel, double-click copy, and invalid-line skip.
