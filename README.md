# Tracker CLI

A Go-based command-line interface and Terminal UI (TUI) client for `tracker-server` to manage time tracking, weekly schedules, deficit catch-ups, and productivity planning.

## Build and Install

```shell
# Build binary
go build -o tracker ./cmd/app/main.go

# Install globally (optional)
sudo mv tracker /usr/local/bin/tracker
```

## Run

```shell
# Quick run during development
go run ./cmd/app/main.go [command]

# Run installed binary
tracker [command]
```

## Backend Configuration

The CLI is stateless and requires an accessible `tracker-server` backend. Configure the backend base URL in `config/config.go` via `TrackerDomain` (toggled between remote production and local dev).

## Available Commands

### Core & TUI
- `tracker dashboard` (aliases: `tui`, `dash`) - Live full-screen TUI dashboard displaying current task, schedule, rollover deficit, rest balance, and warm-up ramp ladder.
- `tracker menu [-t min] [-p percent]` - Interactive task picker table, then starts timer.
- `tracker task -n NAME [-t min] [-p percent] [-s source-day] [--previous-days]` - Start authoritative task timer.
- `tracker evening [-c category] [-t sprint-min] [-s skip-task] [-C [-d combo-min]]` - Evening Catch-Up sprint targeting weekly gaps; `-C` chains the top-3 deficit tasks sequentially.
- `tracker session [duration]` (alias: `batch`) - Run a schedule-aware percent batch session (default 30m).

### Planning & Schedules
- `tracker plan percent run|schedule` - Execute next task in percent plan queue (`--delay`, `-r rest-limit`, `-b batch`).
- `tracker plan percent set --role ROLE --values v1,v2,...` - Update percent distribution for a role.
- `tracker plan backlog` (aliases: `catchup`, `game`) - Sequence through weekly rollover/deficit tasks (`--delay`, `-r rest-limit`, `-b batch`).
- `tracker schedule adjust <task> <delta-min> [-d day]` - Adjust scheduled duration for a task.
- `tracker schedule set <task> <target-min> [-d day]` - Set scheduled target duration for a task.
- `tracker schedule rollover` - View deficit tasks carried over from prior weekdays.
- `tracker ramp [status|reset|set-cap <minutes>]` - Manage warm-up ramp ladder (status, reset to 1m, or adjust cap).

### Tasks & Rest
- `tracker taskadd -n NAME -r ROLE [-t min] [-P priority]` - Add a new task under a role.
- `tracker tasklist` - Display all tasks in a table.
- `tracker statistic` - Display today's statistics, completed tasks, and role totals.
- `tracker rest-spend -d MINUTES` - Record rest minutes spent.
- `tracker rest reset` - Reset daily rest balance.
- `tracker config [-n TASK -t MIN -p PRIORITY]` - Configure task parameters or global scheduler time.

### Maintenance
- `tracker timer-list-set -c COUNT` - Seed backend timer slots.
- `tracker role-recheck` - Recalculate backend role statistics.
- `tracker clean` - Trigger backend data cleanup.

Use `tracker --help` or `tracker [command] --help` for detailed flag usage.

## Technologies

- **Go 1.23.0** - Core language
- **Cobra** - CLI command routing and flags
- **Bubble Tea & Lipgloss** - Terminal UI framework and styling
- **Gorilla WebSocket** - Real-time timer event synchronization
- **slog & tint** - Structured logging

## Developer Guidelines

Detailed architecture notes and development conventions can be found in `CLAUDE.md`, `GEMINI.md`, and `AGENTS.md`.