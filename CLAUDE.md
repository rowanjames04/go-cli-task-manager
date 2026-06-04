# CLAUDE.md - Project Guidelines

## Project Overview

This is a CLI task manager written in Go with both command-line and interactive TUI interfaces.

## Build & Run

```bash
# Build
go build -o taskmanager .

# Run TUI
./taskmanager tui

# Run CLI commands
./taskmanager list
./taskmanager add "My task" -p 2 -d 2026-12-31
```

## Architecture

- `core.go` - Data models (`Task`, `Store`) and persistence logic
- `cli.go` - Cobra-based CLI commands
- `tui.go` - Interactive TUI using tview/tcell
- `tasks.json` - JSON data store

## Key Conventions

- **Priority levels**: 1=Low, 2=Medium, 3=High
- **Task IDs**: Auto-incrementing integers starting at 1
- **Date format**: YYYY-MM-DD (Go layout: `2006-01-02`)
- **Error handling**: Return errors up, print user-facing messages at CLI boundary

## TUI Development

The TUI uses `tview` v0.42.0 and `tcell/v2`:

- Forms for modals use `*tview.Form` with explicit variable declaration for closure capture
- `DropDown.GetCurrentOption()` returns `(int, string)` - index and label
- Key bindings use `Ctrl+Letter` pattern to avoid conflicts with navigation
- Modal pages are added/removed from `*tview.Pages` stack

## Testing

```bash
go test -v
```

See `core_test.go` for existing test patterns.

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/rivo/tview` - TUI framework
- `github.com/gdamore/tcell/v2` - Terminal cells (tview dependency)
