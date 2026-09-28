---
applyTo: '**'
---

# Tracker CLI - Project Instructions

## Project Overview

This is a **time tracking CLI and Terminal UI (TUI) application** written in Go that acts as a client for `tracker-server`. It has **no local database or state**; all business state lives on the backend and is managed via REST API calls and real-time WebSocket events.

## Architecture

### Directory Structure
```
cmd/
├── app/                    # Application entry point (main.go)
└── command/                # Cobra CLI command definitions
    ├── root.go             # Root command and Execute()
    ├── task.go             # Task timer command
    ├── taskadd.go          # Add task command
    ├── tasklist.go         # List tasks table
    ├── menu.go             # Interactive task picker
    ├── dashboard.go        # Live full-screen TUI dashboard
    ├── evening.go          # Evening Catch-Up sprint & combo
    ├── session.go          # Schedule-aware percent batch session
    ├── plan.go             # Plan parent command
    ├── plan_percent.go     # Percent-based planning commands (run, schedule, set)
    ├── plan_backlog.go     # Backlog/rollover planning commands
    ├── schedule.go         # Weekly schedule adjustment & rollover commands
    ├── ramp.go             # Linear warm-up ramp ladder commands
    ├── rest.go             # Rest management commands (spend, reset)
    ├── statistic.go        # Statistics display command
    ├── config.go           # Global timer and task parameter configuration
    ├── role_recheck.go     # Recalculate role statistics
    ├── timer_list_set.go   # Seed backend timer slots
    └── manager.go          # Backend data cleanup

internal/
├── domain/
│   └── entity/             # Shared DTOs for API requests/responses
│       ├── task.go         # Task*, TaskList, TaskParams, TaskRecord
│       ├── role.go         # RoleAnswer struct
│       ├── statistic.go    # TaskTimeDurationResponse
│       ├── timers.go       # RunningTask and timer entities
│       ├── evening.go      # EveningFocusCandidate entities
│       ├── ramp.go         # RampStatus and ramp requests
│       ├── schedule.go     # Schedule entities
│       └── general.go      # Shared general entities
├── repository/
│   ├── api/                # External REST API client implementations
│   │   ├── api.go          # Centralized sendRequest() & SetClientTransport()
│   │   ├── running_task.go # Start, status, pause, resume, stop, adjust, heartbeat
│   │   ├── task.go         # GetTaskParams, AddTaskRecord, GetTaskRecords
│   │   ├── plan_percent.go # Percentage planning endpoints
│   │   ├── procents.go     # Role percentage distributions
│   │   ├── schedule.go     # Schedule adjust, set, rollover
│   │   ├── evening.go      # Evening focus and skip endpoints
│   │   ├── ramp.go         # Warm-up ramp API
│   │   ├── rest.go         # Rest balance endpoints
│   │   ├── statistic.go    # Completion stats
│   │   ├── role.go         # Task role endpoints
│   │   ├── timer.go        # Timer count endpoints
│   │   └── clean_data.go   # Data cleanup operations
│   └── ws/                 # WebSocket client
│       └── client.go       # Reconnecting client for /api/v1/timer/ws
├── service/                # Business logic and interactive TUI models
│   ├── task/               # Task timer execution and Bubble Tea model
│   ├── dashboard/          # Full-screen live dashboard TUI
│   ├── evening/            # Evening Catch-Up TUI and combo chain
│   ├── menu/               # Interactive task picker table
│   ├── plan/               # Percent and backlog planning loops with batching
│   ├── procent/            # Percentage management and distribution
│   ├── task_params/        # Task parameter configuration
│   ├── statistic/          # Statistics formatting and display
│   ├── rest/               # Rest tracking and conversion
│   ├── role/               # Role management
│   ├── timer/              # Timer management
│   ├── telegram/           # Telegram notifications
│   └── manager/            # Backend cleanup service
├── ui/
│   └── theme/              # Shared Lipgloss colors, styles, and role badges
└── pkg/
    ├── restutil/           # Rest unit conversion (units = minutes * 100)
    ├── notifier/           # Desktop notifications and terminal bell
    └── day_method/         # Day and date utility functions

config/
└── config.go               # TrackerDomain constant configuration

test/
└── main.go                 # Bubble Tea UI experiments/demos (not automated tests)
```

### Key Technologies
- **Go 1.23.0** - Primary language
- **Cobra** - CLI command framework (`github.com/spf13/cobra`)
- **Bubble Tea & Lipgloss** - Terminal UI framework and styling (`github.com/charmbracelet/bubbletea`, `lipgloss`)
- **Gorilla WebSocket** - Real-time event streaming (`github.com/gorilla/websocket`)
- **slog & tint** - Structured colorized logging (stdlib `log/slog` + `github.com/lmittmann/tint`)
- **HTTP Client** - Standard library `net/http` configured with 15s timeout

## Coding Standards

### Go Conventions
- Follow standard Go naming conventions (camelCase for unexported, PascalCase for exported).
- Run `go fmt ./...` before committing.
- Prefer structured logging via `slog` over `fmt.Printf` in runtime paths.
- Return wrapped errors with context: `fmt.Errorf("description: %w", err)`.

### HTTP & WebSocket Communication
- Centralize API calls in `internal/repository/api/` using `sendRequest()` in `api.go`.
- Avoid creating inline `http.Client`s in service packages.
- Always handle JSON encoding/decoding within the repository layer and return domain entities.
- WebSocket events (`internal/repository/ws/client.go`) handle real-time timer updates across clients.

### Rest-Time Units
- The backend stores and returns rest time as integer "units" where `units = minutes * 100`.
- Always convert using `internal/pkg/restutil` (`MinutesFromUnits` / `UnitsFromMinutes`).

### Universal Task Duration Logic
- Duration calculations MUST be universal across all tasks. Never hardcode specific task names.
- Explicit `-t <minutes>`: Respect requested duration for manual scheduling.
- Omitted `-t`: Calculate `timeLeft = (params.Time * percent) / 100 - done`. If `timeLeft <= 0`, return `task.ErrTaskCompleted`. Otherwise session duration is `min(defaultDuration, timeLeft)`.

### Server-Authoritative Timer Lifecycle
1. `TaskTimer.Run()` initiates tracking on the backend via `POST /api/v1/timer/run/start`.
2. Connects to `ws.Client` for live events (`TASK_STARTED`, `TASK_PAUSED`, `TASK_RESUMED`, `TASK_STOPPED`, `TASK_ADJUSTED`).
3. Fallback polling queries `GET /api/v1/timer/run/status` every 1.5 seconds.
4. Liveness heartbeat calls `POST /api/v1/timer/run/heartbeat` every 20 seconds.
5. Local 1s tick smoothly decrements display countdown between server events.
6. Actions:
   - `p`: Toggle pause/resume (`POST /api/v1/timer/run/pause` / `.../resume`).
   - Duration adjustment: `POST /api/v1/timer/run/adjust`.
   - `enter` / `q`: Stop task cleanly (`POST /api/v1/timer/run/stop`).
   - `ctrl+c`: Abort task (`POST /api/v1/timer/run/stop` + returns `task.ErrTaskAborted`).
7. Completion: Sends Telegram notification, fires desktop notification (`notifier.Send`), and prints stats/rest.

### Testing Guidelines
- Unit tests must be deterministic and never call a live backend:
  - For HTTP API testing, use `api.SetClientTransport(mockTransport)` in `internal/repository/api/api.go`.
  - For planning loops, use package-level function variables (e.g. `percentTaskSelector`, `percentTimerRunner`, `backlogTimerRunner`).
- Place tests next to code in `*_test.go` files using table-driven tests.

## Available CLI Commands

- `tracker task -n NAME [-t min] [-p percent] [-s source-day] [--previous-days]`: Run a task timer.
- `tracker menu [-t min] [-p percent]`: Interactive Bubble Tea task picker table, then starts timer.
- `tracker dashboard` (aliases: `tui`, `dash`): Live full-screen TUI dashboard.
- `tracker evening [-c category] [-t sprint-min] [-s skip-task] [-C [-d combo-min]]`: Evening Catch-Up sprint targeting weekly gaps; `-C` chains the top-3 deficit tasks.
- `tracker session [duration]` (alias: `batch`): Run a schedule-aware percent batch session (default 30m).
- `tracker plan percent run|schedule`: Start next task from percent plan (`--delay`, `-r rest-limit`, `-b batch`).
- `tracker plan percent set --role R --values v1,v2,...`: Update percent distribution for a role.
- `tracker plan backlog` (aliases: `catchup`, `game`): Sequence through deficit/rollover tasks (`--delay`, `-r rest-limit`, `-b batch`).
- `tracker schedule adjust <task> <delta-min> [-d day]`: Adjust scheduled minutes for a task.
- `tracker schedule set <task> <target-min> [-d day]`: Set scheduled target minutes for a task.
- `tracker schedule rollover`: View rollover deficit tasks.
- `tracker ramp [status|reset|set-cap <minutes>]`: Warm-up ramp ladder management.
- `tracker taskadd -n NAME -r ROLE [-t min] [-P priority]`: Add a new task under a role.
- `tracker tasklist`: Display full task table.
- `tracker statistic`: Display today's statistics, completed tasks, and role totals.
- `tracker rest-spend -d MINUTES`: Record spent rest minutes.
- `tracker rest reset`: Reset daily rest balance.
- `tracker config [-n TASK -t MIN -p PRIORITY]`: Configure task parameters or global scheduler time.
- `tracker timer-list-set -c COUNT`: Seed backend timer slots.
- `tracker role-recheck`: Recalculate backend role statistics.
- `tracker clean`: Trigger backend record cleanup.