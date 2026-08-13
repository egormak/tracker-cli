# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Tracker CLI is a Go command-line client for a separate Tracker backend service (time/task tracking with three roles: Work, Learn, Rest). It has no local persistence or database of its own — every command reads or writes state by calling the backend's REST API. Cobra provides the command tree; Bubble Tea (+ Lipgloss) provides the interactive TUI parts (task timer, task-selection menu, table rendering).

The backend's HTTP contract is documented in `openapi.yml` at the repo root — check it when a command's request/response shape is unclear.

## Build and Run

```shell
go build -o tracker ./cmd/app/main.go
sudo mv tracker /usr/local/bin/tracker   # optional, to install globally

go run ./cmd/app/main.go [command]       # quick dev run
```

Test coverage is minimal — currently only `internal/service/task/task_test.go` (table-driven tests for `calculateDuration`/`calculateTimeLeft`). Run `go test ./...` (or `go test ./internal/service/task/...` for just that package); use `go build ./...` / `go vet ./...` for everything else.

Pushing a git tag triggers `.github/workflows/github-actions-demo.yml`, which builds the binary and publishes it as a GitHub release artifact.

### Backend

The backend base URL is a single hardcoded const in `config/config.go` (`TrackerDomain`), switched by commenting/uncommenting a line — it is **not** an env var or flag. It currently points at production (`http://tracker.makegorka.com:8080`); the local dev URL (`http://127.0.0.1:3000`) is commented out. Flip these when working against a local backend.

```shell
# Local backend's MongoDB, if running the backend locally too
docker run -it --rm -p 27017:27017 -v /home/egorka/Downloads/test_mongo:/data/db mongo:5.0.6
```

## Architecture

- **`cmd/command/`** — one file per Cobra command. Each file defines a `*cobra.Command` and registers itself onto `rootCmd` from an `init()` function, so adding a command means creating a new file here (following an existing one) rather than editing a central registry.
- **`internal/service/<feature>/`** — business logic per feature (task, task_params, menu, statistic, procent, rest, telegram, timer, role, plan, manager, evening). Bubble Tea models live alongside their feature's service code (e.g. `internal/service/task/task_timer.go`, `internal/service/menu/menu.go`, `internal/service/evening/evening.go`).
- **`internal/repository/api/`** — the intended REST client layer, built around a single `sendRequest(method, path, body)` helper in `api.go` (15s timeout, JSON headers, base URL from `config.TrackerDomain`).
- **`internal/domain/entity/`** — DTOs shared between the API layer and services (`task.go`, `role.go`, `statistic.go`, `timers.go`, `schedule.go`, `general.go`).

### Inconsistent API-call pattern — read before adding backend calls

Not all HTTP calls go through `internal/repository/api`. A number of older packages build their own `http.Client` inline with a duplicated 15s-timeout/JSON-header block instead of calling `sendRequest`: `internal/service/rest/rest.go`, `internal/service/role/role.go`, `internal/service/timer/timer.go`, `internal/service/telegram/telegram.go`, `internal/service/task_params/task_params.go`, `internal/service/statistic/statistic.go`, `internal/service/statistic/tasklist.go`, `internal/service/menu/menu.go`, and `internal/repository/api/schedule.go` (this last one sits in the `api` package but still bypasses `sendRequest`). When adding a new backend call, follow the pattern in `internal/repository/api/task.go` or `running_task.go` (`sendRequest` + a typed response struct), not these older files.

Also note `timer.TimeListSet` posts to `/v1/timer/set`, missing the `/api` prefix every other endpoint uses — verify against `openapi.yml` before assuming it works.

### Server-synchronized task timer

The task timer's state of truth is the backend, not the local process — this replaced an earlier fully-local Bubble Tea timer. The flow, spread across `internal/service/task/task.go`, `task_timer.go`, and `internal/repository/api/running_task.go`:

1. `task.CreateTaskTimer` computes the duration to run (from task params, requested time/percent, and time already done) but does not contact the server yet.
2. `TaskTimer.Run()` calls `POST /api/v1/timer/run/start`, which returns the authoritative `entity.RunningTask` (start time, accumulated minutes, running/paused state).
3. The Bubble Tea model (`teaTimerModel`) polls `GET /api/v1/timer/run/status` every 1.5s to stay in sync with server state (in case another client paused/stopped/completed the task), while a separate 1s local tick just smooths the on-screen elapsed/remaining display between polls.
4. Pause/resume (`p` key) call `POST /api/v1/timer/run/pause` / `.../resume`; stop (`enter`/`q`) or abort (`ctrl+c`) call `POST /api/v1/timer/run/stop`. `ctrl+c` additionally sets `abortPlan`, which propagates up as `task.ErrTaskAborted` so callers like `plan backlog`/`plan percent` know to stop chaining tasks rather than continuing to the next one.
5. On a completed run, `finalizeSession` triggers a percent-plan-change notification (`procent.ChangeGroupPlanPercent`) and a Telegram message, then prints updated statistics/rest via `statistic` and `rest`.

### Percent-based planning and backlog

`tracker plan` has two independent scheduling strategies, both of which ultimately call `task.CreateTaskTimer(...).Run()` in a loop:

- `plan percent run|start` / `plan percent schedule|sched` — pulls the next task from the backend's percent-based plan queue (`GET /api/v1/task/plan-percent` or the schedule-aware variant), optionally looping while rest time stays under `--rest-limit` minutes (negative disables looping).
- `plan backlog` (aliases `catchup`, `game`) — repeatedly fetches all rollover/deficit tasks (`GET /api/v1/schedule/active/rollover`) and works through them by remaining time, also respecting `--rest-limit`.
- `plan percent set --role R --values ...` updates a role's percent distribution via `internal/service/procent` (`POST /api/v1/manage/procents`).

### Rest-time units

The backend stores/returns rest time as integer "units" equal to `minutes * 100`, not minutes directly. Always convert with `internal/pkg/restutil` (`MinutesFromUnits` / `UnitsFromMinutes`) at the boundary — this matters anywhere a `--rest-limit`-style minute value is compared against a value fetched from the API.

## Commands

- `tracker menu` — interactive task picker (Bubble Tea table), then starts the selected task's timer
- `tracker task -n NAME [-t minutes] [-p percent] [-s source-day] [--previous-days]` — run a task timer; `--previous-days` looks up the task via the schedule-aware endpoint across Monday-to-today
- `tracker evening [-c category] [-t sprint-minutes] [-s skip-task]` — Bubble Tea "Evening Catch-Up" picker (`internal/service/evening`) that surfaces the biggest weekly-gap task via `GetEveningFocus`/`SkipEveningTask`, then runs it through the normal `task.CreateTaskTimer(...).Run()` flow at 100%
- `tracker taskadd -n NAME -r ROLE` — add/register a task under a role
- `tracker tasklist` — render the task list as a table
- `tracker statistic` — print today/yesterday/all-time completion stats and role totals
- `tracker rest-spend -d MINUTES` — record rest time spent
- `tracker plan backlog [--delay 15s] [-r rest-limit]` — sequence through deficit/rollover tasks
- `tracker plan percent run|start [--delay] [-r rest-limit]` — start the next percent-plan task
- `tracker plan percent schedule|sched [--delay] [-r rest-limit]` — same, schedule-aware
- `tracker plan percent set --role R --values v1,v2,...` — update a role's percent distribution
- `tracker timer-recheck` — ask the backend to recheck/refresh timers
- `tracker timer-list-set -c COUNT` — seed the backend's timer list with COUNT slots
- `tracker role-recheck` — recalculate role statistics on the backend
- `tracker config [-n TASK -t MINUTES -p PRIORITY]` — with `-n`, sets that task's time/priority params; without `-n`, `-t` alone sets the global scheduler time
- `tracker clean` — trigger the backend's record cleanup endpoint
