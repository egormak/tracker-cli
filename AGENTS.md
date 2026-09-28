# Repository Guidelines

## Project Overview
- **Tracker CLI** is a Go-based command-line interface and Terminal UI (TUI) client for `tracker-server`.
- **Technologies**: Go 1.23.0, Cobra (`cmd/command`), Bubble Tea & Lipgloss (`internal/ui/theme`, TUIs), gorilla/websocket, slog + tint.
- **Stateless Architecture**: The CLI has zero local persistence or database. All state lives behind `tracker-server` accessed via REST API (`/api/v1/...`) and live WebSocket (`/api/v1/timer/ws`). Do not add local SQLite/Mongo storage to the CLI.

## Project Structure & Module Organization
- `cmd/app/main.go`: Entry point; sets up `slog` default logger with `tint` handler and runs `command.Execute()`.
- `cmd/command/`: Individual Cobra CLI command files (`task.go`, `menu.go`, `dashboard.go`, `evening.go`, `session.go`, `plan_percent.go`, `plan_backlog.go`, `schedule.go`, `ramp.go`, `rest.go`, `statistic.go`, etc.). Each command registers itself onto `rootCmd` (or parent) via its `init()` function.
- `internal/service/<feature>/`: Feature-scoped business logic and Bubble Tea models (`task`, `dashboard`, `evening`, `menu`, `plan`, `procent`, `rest`, `role`, `statistic`, `task_params`, `telegram`, `timer`).
- `internal/repository/api/`: REST API client layer (`sendRequest` with 15s timeout, JSON headers).
- `internal/repository/ws/`: Reconnecting WebSocket client for `/api/v1/timer/ws`, broadcasting typed timer events (`TASK_STARTED`, `TASK_PAUSED`, `TASK_RESUMED`, `TASK_STOPPED`, `TASK_ADJUSTED`, `HEARTBEAT_ACK`, `STATE_SYNC`).
- `internal/domain/entity/`: Shared DTOs for requests and responses between API and service layers.
- `internal/ui/theme/`: Shared Lipgloss styling, color palette, and role badges. Use this for new TUI elements instead of hardcoding colors.
- `internal/pkg/`: Shared utility packages:
  - `restutil`: Rest unit boundary converters (`units = minutes * 100`).
  - `notifier`: Cross-platform notifications (terminal bell, macOS `osascript`, Linux `notify-send`).
  - `day_method`: Weekday calculation helpers.
- `config/config.go`: Holds `TrackerDomain` constant (toggle production `http://tracker.makegorka.com:8080` vs local dev `http://127.0.0.1:3000`).
- `test/main.go`: Scratchpad for Bubble Tea UI experiments/demos, not an automated test.

## Build, Test, and Development Commands
- `go build -o tracker ./cmd/app/main.go` builds the binary.
- `go run ./cmd/app/main.go [command]` runs commands directly during development.
- `go test ./...` runs all unit tests.
- `go test ./internal/service/plan -run TestRunPercentBatch` runs a specific test or package.
- `go vet ./...` and `go fmt ./...` should pass before committing.

## Key Development Conventions

### 1. Server-Authoritative Task Timer
- `task.CreateTaskTimer` determines session duration, then `TaskTimer.Run()` starts server tracking via `POST /api/v1/timer/run/start` and opens `ws.Client`.
- State synchronization operates three ways:
  1. Live WebSocket events (external pause/resume/stop/adjust).
  2. Fallback polling: `GET /api/v1/timer/run/status` every 1.5s.
  3. Liveness heartbeat: `POST /api/v1/timer/run/heartbeat` every 20s.
  4. Local 1s tick smooths UI display between server updates.
- Controls: `p` toggles pause/resume; duration changes call `/adjust`; `enter`/`q` stops cleanly; `ctrl+c` aborts. Bubble Tea default signal handling is disabled in favor of a custom SIGINT forwarder sending `interruptMsg` so loops can catch `task.ErrTaskAborted`.
- On completion: sends Telegram notification (`procent.ChangeGroupPlanPercent`), fires desktop notification, and displays updated statistics and rest.

### 2. Universal Task Duration Logic
- Duration calculations MUST be universal across all tasks. Never hardcode specific task names in calculation logic.
- Explicit `-t <minutes>`: Honors requested duration for manual soft-scheduling.
- Omitted `-t`: Calculates `timeLeft = (params.Time * percent)/100 - done`. If `timeLeft <= 0`, returns `ErrTaskCompleted`. Otherwise session duration is clamped to `min(defaultDuration, timeLeft)`.

### 3. Rest-Time Units
- The backend stores and returns rest time as integer units (`units = minutes * 100`).
- Always use `internal/pkg/restutil` (`MinutesFromUnits` / `UnitsFromMinutes`) when sending or displaying rest values.

### 4. API Client Consistency
- When adding backend calls, implement them in `internal/repository/api/` using `sendRequest` and typed structs, rather than creating inline `http.Client`s in service packages.

### 5. Testing Guidelines
- Tests must never hit a live backend:
  - Use `api.SetClientTransport(rt)` in `internal/repository/api` to swap the shared HTTP transport with a mock `http.RoundTripper`.
  - Use package-level function variables (e.g. `percentTaskSelector`, `percentTimerRunner`, `backlogTimerRunner`) to swap runner logic in loop tests.
- Keep tests next to code in `*_test.go` files; table-driven tests are preferred.

## Key CLI Commands
- `tracker task -n NAME [-t min] [-p percent] [-s source-day] [--previous-days]`: Run task timer.
- `tracker menu [-t min] [-p percent]`: Interactive Bubble Tea task picker table, then starts timer.
- `tracker dashboard` (aliases: `tui`, `dash`): Live full-screen TUI dashboard.
- `tracker evening [-c category] [-t sprint-min] [-s skip-task] [-C [-d combo-min]]`: Evening Catch-Up mode for biggest weekly-gap task; `-C` chains top-3 candidates.
- `tracker session [duration]` (alias: `batch`): Schedule-aware percent batch session (default 30m).
- `tracker plan backlog` (aliases: `catchup`, `game`): Sequence through deficit/rollover tasks (`--delay`, `-r rest-limit`, `-b batch`).
- `tracker plan percent run|schedule`: Start next task from percent plan (`--delay`, `-r rest-limit`, `-b batch`).
- `tracker plan percent set --role R --values v1,v2,...`: Update role percent distribution.
- `tracker schedule adjust <task> <delta-min> [-d day]` / `schedule set <task> <target-min> [-d day]` / `schedule rollover`: Schedule targets and rollover management.
- `tracker ramp [status|reset|set-cap <minutes>]`: Warm-up ramp ladder controls.
- `tracker taskadd -n NAME -r ROLE [-t min] [-P priority]`: Add a new task under a role.
- `tracker tasklist`: List all tasks in table format.
- `tracker statistic`: Display today's statistics and role totals.
- `tracker rest-spend -d MINUTES`: Record spent rest minutes.
- `tracker rest reset`: Reset daily rest balance.
- `tracker config [-n TASK -t MIN -p PRIORITY]`: Configure task parameters or global scheduler time.
- `tracker timer-list-set`, `tracker role-recheck`, `tracker clean`: Backend maintenance commands.
