# tracker_cli

A Go-based command-line interface and Terminal UI (TUI) client for `tracker-server`. Uses Cobra for CLI commands and Bubble Tea + Lipgloss for interactive components.

## Project Overview

- **Technologies**: Go 1.23.0, Cobra, Bubble Tea, Bubbles, Lipgloss, Gorilla WebSocket, slog with tint.
- **Architecture**:
  - `cmd/app/main.go`: Application entry point; initializes `slog` with `tint` and invokes Cobra root command.
  - `cmd/command/`: Individual Cobra CLI command definitions (`task.go`, `menu.go`, `dashboard.go`, `evening.go`, `session.go`, `plan_percent.go`, `plan_backlog.go`, `schedule.go`, `ramp.go`, `debt.go`, `statistic.go`, etc.).
  - `internal/service/`: Feature-specific business logic and interactive TUI models (`task`, `dashboard`, `debt`, `digest`, `evening`, `menu`, `plan`, `procent`, `role`, `statistic`, `task_params`, `telegram`, `timer`).
  - `internal/repository/api/`: REST API client layer (`api.go` with centralized `sendRequest()`, 15-second default timeout, and mockable transport).
  - `internal/repository/ws/`: Reconnecting WebSocket client (`/api/v1/timer/ws`) that receives real-time timer lifecycle events.
  - `internal/domain/entity/`: Shared DTO entities exchanged between API client and service logic.
  - `internal/ui/theme/`: Unified Lipgloss styling, theme colors, and role badge formatting.
  - `internal/pkg/`:
    - `notifier`: Cross-platform notifications (terminal bell, macOS `osascript`, Linux `notify-send`).
    - `day_method`: Weekday calculation utilities.
  - `config/config.go`: Contains `TrackerDomain` base URL const (toggled between remote production `:8080` and local dev `:3000`).

## Building and Running

### Build
```bash
go build -o tracker ./cmd/app/main.go
```

### Install Globally
```bash
sudo mv tracker /usr/local/bin/tracker
```

### Run
```bash
go run ./cmd/app/main.go [command]
```

### Test
```bash
go test ./...
go test ./internal/service/plan -run TestRunPercentBatch
```

## Key Commands

- `tracker task -n "Name" [-t min] [-p percent] [-s source-day] [--previous-days] [--debt P]`: Start task timer with authoritative backend synchronization; `--debt` = % of the session paid into this task's previous-days debt.
- `tracker menu [-t min] [-p percent]`: Interactive Bubble Tea task picker table, then runs timer for chosen task.
- `tracker dashboard` (aliases: `tui`, `dash`): Full-screen live TUI monitoring active timer, daily progress, schedule, rollover, debt (line + week strip), evening focus, and ramp ladder.
- `tracker evening [-c category] [-t sprint-min] [-s skip-task] [-C [-d combo-min]]`: Evening Catch-Up sprint for the task most behind to date (debt, then weekly gap); `-C` chains the top-3 candidates sequentially.
- `tracker evening --batch [-a anchor] [-t 15] [--debt P] [--tasks N]`: Evening batch — anchor + next candidates by debt, one sprint each (`BuildBatch` / `RunBatch` in `internal/service/evening/combo.go`).
- `tracker debt`: This week's debt line, week strip, and items (`internal/service/debt`).
- `tracker session [duration]` (alias: `batch`): Run a schedule-aware percent batch session (defaults to 30m) clamping tasks to the remaining time budget.
- `tracker plan percent run|schedule`: Run next task in percent plan queue (`--delay`, `-b batch`); without `-b` one task runs.
- `tracker plan percent set --role R --values v1,v2,...`: Update percent distribution for a role.
- `tracker plan backlog` (aliases: `catchup`, `game`): Run sequence of rollover/deficit tasks (`--delay`, `-b batch`); without `-b` runs until no backlog is left.
- `tracker schedule adjust <task> <delta-min> [-d day]`: Adjust scheduled minutes for a task atomically.
- `tracker schedule set <task> <target-min> [-d day]`: Set scheduled target minutes for a task.
- `tracker schedule rollover`: View rollover deficit tasks carried over from prior weekdays.
- `tracker ramp [status|reset|set-cap <minutes>]`: Warm-up ramp ladder status, step reset (to 1m), or cap adjustment (min 5m).
- `tracker taskadd -n "Name" -r "Role" [-t min] [-P priority]`: Register a new task under a role.
- `tracker tasklist`: Display full task table with priorities, times, and completion.
- `tracker statistic`: Display today's statistics, completed tasks, and role totals.
- `tracker config [-n task -t min -p priority]`: Configure task parameters or global scheduler time.
- `tracker timer-list-set -c [count]`: Seed backend timer slots.
- `tracker role-recheck`: Recalculate backend role statistics.
- `tracker clean`: Trigger backend record cleanup.

## Development Conventions

- **Server-Authoritative Timer**: `TaskTimer.Run()` initializes the task on the backend via `POST /api/v1/timer/run/start` and attaches a `ws.Client`. `teaTimerModel` stays synchronized via WebSocket events, a 1.5s `GET /api/v1/timer/run/status` poll fallback, and sends a `POST /api/v1/timer/run/heartbeat` every 20s. Local 1s ticks provide smooth UI countdowns between server updates. Pause/resume (`p`) calls `.../pause` / `.../resume`, adjustments call `.../adjust`, and stop/abort calls `.../stop`. Default Bubble Tea signal handling is disabled; a custom SIGINT forwarder sends `interruptMsg` so caller loops detect `task.ErrTaskAborted`. `origin` and `debt_share` go on start (`api.StartRunningTask(..., origin, debtShare)`); after a stop the server's split is printed via `task.FormatSplit` (e.g. `Записано: сегодня 15 · Ср +5`).
- **Universal Task Duration Logic**: `calculateDuration` is universal across all tasks. Explicit `-t <minutes>` is honored for manual soft-scheduling; omitted `-t` calculates `timeLeft = (params.Time * percent)/100 - done`. If `timeLeft <= 0`, `ErrTaskCompleted` is returned.
- **API Communication**: Prefer adding new endpoints to `internal/repository/api/` using `sendRequest()`. Avoid inline `http.Client` instances in service packages.
- **Testing Seams**: Unit tests must never contact a live backend:
  - For HTTP calls: Use `api.SetClientTransport(mockTransport)` in `internal/repository/api/api.go`.
  - For planning loops: Use package-level function variables (e.g. `percentTaskSelector`, `percentTimerRunner`, `backlogTimerRunner`, `comboSprintCreator`).
- **No Local Persistence**: The CLI is completely stateless; all application state lives behind `tracker-server`.
