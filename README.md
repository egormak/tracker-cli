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
- `tracker dashboard` (aliases: `tui`, `dash`) - Live full-screen TUI dashboard displaying current task, schedule, rollover deficit, debt line and week strip, and warm-up ramp ladder.
- `tracker menu [-t min] [-p percent]` - Interactive task picker table, then starts timer.
- `tracker task -n NAME [-t min] [-p percent] [-s source-day] [--previous-days] [--debt P]` - Start authoritative task timer; `--debt P` sends P% of the session to this task's previous-days debt on stop. When the server splits a session, a line like `Записано: сегодня 15 · Ср +5` is printed.
- `tracker evening [-c category] [-t sprint-min] [-s skip-task] [-C [-d combo-min]]` - Evening Catch-Up sprint for the task most behind to date (debt, then weekly gap); `-C` chains the top-3 candidates sequentially.
- `tracker evening --batch [-a anchor] [-t 15] [--debt P] [--tasks 3|4]` - Evening batch: the anchor plus the next candidates by debt, one sprint each (default 15m).
- `tracker debt` - Show this week's debt: summary line, week strip, and items.
- `tracker session [duration]` (alias: `batch`) - Run a schedule-aware percent batch session (default 30m).

### Planning & Schedules
- `tracker plan percent run|schedule` - Execute next task in percent plan queue (`--delay`, `-b batch`); without `-b` one task runs.
- `tracker plan percent set --role ROLE --values v1,v2,...` - Update percent distribution for a role.
- `tracker plan backlog` (aliases: `catchup`, `game`) - Sequence through weekly rollover/deficit tasks (`--delay`, `-b batch`); without `-b` runs until no backlog is left.
- `tracker schedule adjust <task> <delta-min> [-d day]` - Adjust scheduled duration for a task.
- `tracker schedule set <task> <target-min> [-d day]` - Set scheduled target duration for a task.
- `tracker schedule rollover` - View deficit tasks carried over from prior weekdays.
- `tracker ramp [status|reset|set-cap <minutes>]` - Manage warm-up ramp ladder (status, reset to 1m, or adjust cap).

### Tasks
- `tracker taskadd -n NAME -r ROLE [-t min] [-P priority]` - Add a new task under a role.
- `tracker tasklist` - Display all tasks in a table.
- `tracker statistic` - Display today's statistics, completed tasks, and role totals.
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