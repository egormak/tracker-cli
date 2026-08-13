# tracker_cli

A Go-based command-line interface and Terminal UI (TUI) client for `tracker-server`. Uses Cobra for CLI commands and Bubble Tea + Lipgloss for interactive components.

## Project Overview

- **Technologies**: Go 1.22+, Cobra, Bubble Tea, Lipgloss.
- **Architecture**:
  - `cmd/app/main.go`: Entry point.
  - `cmd/command/`: Individual Cobra CLI command files (`task.go`, `plan.go`, `menu.go`, etc.).
  - `internal/service/`: Business logic and Bubble Tea TUI models (e.g. `internal/service/task/task_timer.go`, `internal/service/menu/menu.go`).
  - `internal/repository/api/`: REST API client layer (`api.go` with `sendRequest()`).
  - `internal/domain/entity/`: DTO entities shared between API layer and services.
  - `internal/pkg/restutil/`: Rest unit boundary converters (`units = minutes * 100`).
  - `config/config.go`: Contains hardcoded `TrackerDomain` base URL (toggle local `:3000` vs remote).

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

## Key Commands

- `tracker menu`: Interactive Bubble Tea task picker table, then starts timer for selected task.
- `tracker task -n "Name" [-t minutes] [-p percent] [-s source-day]`: Start task timer.
- `tracker plan percent run`: Execute next task in percent plan queue.
- `tracker plan percent schedule`: Execute schedule-aware percent plan task.
- `tracker plan backlog`: Run sequence of weekly deficit/rollover tasks.
- `tracker taskadd -n "Name" -r "Role"`: Add a new task under a role.
- `tracker tasklist`: Display list of tasks.
- `tracker statistic`: Display today's statistics and role totals.
- `tracker rest-spend -d [duration]`: Record rest minutes spent.
- `tracker timer-recheck`: Trigger backend timer recheck.

## Development Conventions

- **Server-Authoritative Timer**: `teaTimerModel` polls `GET /api/v1/timer/run/status` every 1.5s while running a 1s local tick for visual smoothing between polls.
- **Universal Task Duration Logic**: `calculateDuration` is universal across all tasks. Explicit `-t <minutes>` is honored for manual soft-scheduling; omitted `-t` calculates `timeLeft = (params.Time * percent)/100 - done`.
- **Rest-Time Conversion**: Always use `internal/pkg/restutil` (`MinutesFromUnits` / `UnitsFromMinutes`) when interacting with API rest values (`units = minutes * 100`).
- **API Communication**: Prefer adding new endpoints to `internal/repository/api/` using `sendRequest()`.
- **No Local Persistence**: The CLI is completely stateless; state lives behind `tracker-server`.
