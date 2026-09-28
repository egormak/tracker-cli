# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Tracker CLI is a Go command-line client for a separate Tracker backend service (time/task tracking with three roles: Work, Learn, Rest). It has no local persistence or database of its own — every command reads or writes state by calling the backend's REST API (plus a WebSocket for live timer events). Cobra provides the command tree; Bubble Tea (+ Lipgloss) provides the interactive TUI parts (task timer, dashboard, task-selection menu, evening picker).

The backend's HTTP contract is documented in `openapi.yml` at the repo root — check it when a command's request/response shape is unclear.

## Build, Test, Run

```shell
go build -o tracker ./cmd/app/main.go
sudo mv tracker /usr/local/bin/tracker   # optional, to install globally
go run ./cmd/app/main.go [command]       # quick dev run

go build ./... && go vet ./...
go test ./...
go test ./internal/service/plan -run TestRunPercentBatch   # single test / subset
```

Pushing a git tag triggers `.github/workflows/github-actions-demo.yml`, which builds the binary and publishes it as a GitHub release artifact. Note the workflow pins Go `1.22` while `go.mod` declares `go 1.23.0` — bump the workflow if a release build fails on the toolchain version.

### Testing patterns

Tests never hit a live backend. Two seams are used:
- **Package-level function variables** that tests swap and restore — e.g. `percentTaskSelector`, `percentTimerRunner`, `telegramSender` in `internal/service/plan/percent.go`, `backlogTimerRunner` in `backlog.go`. Follow this pattern when adding logic that loops over timer runs.
- **`api.SetClientTransport(rt)`** in `internal/repository/api/api.go` replaces the shared `http.Client`'s transport with a fake `RoundTripper` (returns a cleanup func). This only affects calls that go through `sendRequest` — packages with their own inline `http.Client` (see below) can't be faked this way.

### Backend

The backend base URL is a single hardcoded const in `config/config.go` (`TrackerDomain`), switched by commenting/uncommenting a line — it is **not** an env var or flag. It currently points at production (`http://tracker.makegorka.com:8080`); the local dev URL (`http://127.0.0.1:3000`) is commented out. Flip these when working against a local backend. The WebSocket URL is derived from the same const (`http→ws`, `https→wss`).

## Architecture

- **`cmd/command/`** — one file per Cobra command. Each file defines a `*cobra.Command` and registers itself onto `rootCmd` (or a parent like `planCmd`) from an `init()` function, so adding a command means creating a new file here rather than editing a central registry.
- **`internal/service/<feature>/`** — business logic per feature. Bubble Tea models live alongside their feature's service code (`task/task_timer.go`, `dashboard/dashboard.go`, `menu/menu.go`, `evening/evening.go`). There is also a legacy root package `internal/service` (`statistic.go`, `timer.go`) that is still imported by `cmd/command/statistic.go` and `service/plan` — don't confuse it with `service/statistic` / `service/timer`.
- **`internal/repository/api/`** — the REST client layer, built around a single `sendRequest(method, path, body)` helper in `api.go` (shared 15s-timeout client, JSON headers, non-200 → error).
- **`internal/repository/ws/`** — reconnecting WebSocket client for `/api/v1/timer/ws`, emitting typed events (`TASK_STARTED/PAUSED/RESUMED/STOPPED/ADJUSTED`, `STATE_SYNC`, …).
- **`internal/domain/entity/`** — DTOs shared between the API layer and services.
- **`internal/ui/theme/`** — shared Lipgloss palette/styles and role badges; use it for new TUI output rather than defining new colors.
- **`internal/pkg/`** — `restutil` (rest-unit conversion), `notifier` (terminal bell + macOS `osascript` / Linux `notify-send` desktop notifications), `day_method`.

### Inconsistent API-call pattern — read before adding backend calls

Not all HTTP calls go through `sendRequest`. These still build their own `http.Client` inline with a duplicated timeout/header block: `internal/service/rest/rest.go`, `internal/service/role/role.go`, `internal/service/telegram/telegram.go`, `internal/service/task_params/task_params.go`, `internal/service/statistic/tasklist.go`, and `GetRolloverTasks` in `internal/repository/api/schedule.go` (the rest of that file uses `sendRequest`). When adding a new backend call, put it in `internal/repository/api/` using `sendRequest` + a typed response struct (see `running_task.go`, `ramp.go`, `timer.go`), and have the service layer call that.

### Server-synchronized task timer

The backend owns running-task state; the local process only renders it. The flow, spread across `internal/service/task/task.go`, `task_timer.go`, and `internal/repository/api/running_task.go`:

1. `task.CreateTaskTimer` computes the duration to run (from task params, requested time/percent, and time already done) but does not contact the server yet.
2. `TaskTimer.Run()` calls `POST /api/v1/timer/run/start`, which returns the authoritative `entity.RunningTask`, and opens a `ws.Client`.
3. The Bubble Tea model (`teaTimerModel`) stays in sync three ways: WebSocket events (another client paused/resumed/stopped/adjusted the task), a `GET /api/v1/timer/run/status` poll every 1.5s as a fallback, and a `POST .../heartbeat` every 20s. A separate 1s local tick only smooths the on-screen display.
4. Pause/resume (`p`) → `.../pause` / `.../resume`; duration changes → `.../adjust`; stop (`enter`/`q`) or abort (`ctrl+c`) → `.../stop`. `ctrl+c` additionally sets `abortPlan`, which surfaces as `task.ErrTaskAborted` so looping callers (`plan`, `session`, `evening --combo`) stop chaining tasks. Bubble Tea's own signal handling is disabled; a custom `SIGINT` forwarder sends `interruptMsg` into the program.
5. On a completed run, the timer sends a percent-plan-change + rest-balance Telegram message (`procent.ChangeGroupPlanPercent`, `telegram`), fires a desktop notification, then prints stats and rest.

`tracker dashboard` (`internal/service/dashboard`) is a separate full-screen TUI that polls the running task, task list, rollover, rest, schedule, evening focus, and ramp status, and can pause/resume/adjust(±5)/stop the running task and reset the ramp.

### Planning loops

`internal/service/plan` holds both scheduling strategies; each ultimately runs `task.CreateTaskTimer(...).Run()` in a loop:

- **Percent** (`RunPercent`, `RunPercentSchedule`, `RunPercentBatch`) — pulls the next task from `GET /api/v1/task/plan-percent` (or the schedule-aware variant).
- **Backlog** (`RunBacklog`, `RunBacklogBatch`) — works through rollover/deficit tasks from `GET /api/v1/schedule/active/rollover`.

Both support `--rest-limit` (keep looping while banked rest stays under N minutes; negative disables) and `--batch` (a wall-clock budget: each task's duration is clamped to the remaining batch time). `tracker session [duration]` is a shortcut for a schedule-aware percent batch (default 30m).

### Rest-time units

The backend stores/returns rest time as integer "units" equal to `minutes * 100`. Always convert with `internal/pkg/restutil` (`MinutesFromUnits` / `UnitsFromMinutes`) — especially anywhere a `--rest-limit` minute value is compared against a value fetched from the API.

## Commands

- `tracker task -n NAME [-t min] [-p percent] [-s source-day] [--previous-days]` — run a task timer
- `tracker menu [-t min] [-p percent]` — interactive task picker, then runs the timer
- `tracker dashboard` (`tui`, `dash`) — live TUI dashboard
- `tracker evening [-c category] [-t sprint-min] [-s skip-task] [-C [-d combo-min]]` — "Evening Catch-Up" picker for the biggest weekly-gap task; `-C` chains the top-3 candidates
- `tracker session [duration]` (`batch`) — schedule-aware percent batch session
- `tracker plan backlog` (`catchup`, `game`) / `plan percent run|schedule` — planning loops (`--delay`, `-r rest-limit`, `-b batch`)
- `tracker plan percent set --role R --values v1,v2,...` — update a role's percent distribution
- `tracker schedule adjust <task> <delta-min> [-d day]` / `schedule set <task> <target-min> [-d day]` / `schedule rollover` — edit this week's schedule targets or view rollover
- `tracker ramp [status|reset|set-cap <minutes>]` — warm-up ramp ladder
- `tracker taskadd -n NAME -r ROLE [-t min] [-P priority]`, `tracker tasklist`, `tracker statistic`
- `tracker rest-spend -d MINUTES`, `tracker rest reset`
- `tracker config [-n TASK -t MIN -p PRIORITY]` — with `-n`, sets that task's params; without it, `-t` sets the global scheduler time
- `tracker timer-list-set -c COUNT`, `tracker role-recheck`, `tracker clean` — backend maintenance endpoints

## Other agent instruction files

`GEMINI.md`, `AGENTS.md`, and `.github/instructions/copilot.instructions.md` are synchronized with this file. Key conventions across all guides: run `go fmt ./...`, use `slog` for runtime logging rather than `fmt.Printf`, keep tests next to the code they exercise, and `test/main.go` is a throwaway Bubble Tea demo, not a test.
