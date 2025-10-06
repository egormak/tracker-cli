# Repository Guidelines

## Project Structure & Module Organization
- `cmd/app/main.go` boots the CLI and delegates to Cobra commands in `cmd/command` (task, rest, statistic, manager).
- Business logic sits in `internal/service/*` and `internal/application`, with domain contracts in `internal/domain` and adapters in `internal/infrastructure`.
- Shared utilities live in `internal/pkg`; configuration defaults, including `TrackerDomain`, reside in `config/config.go`.
- Keep exploratory Bubble Tea demos in `test/main.go`; production tests should live next to the code they exercise.

## Build, Test, and Development Commands
- `go build -o tracker ./cmd/app/main.go` builds the binary; install it manually if you need a global executable.
- `go run ./cmd/app/main.go --help` quickly inspects the command tree during development.
- `go test ./...` runs all Go tests; use `-run` to narrow the scope when iterating.
- `docker run -it --rm -p 27017:27017 -v /home/egorka/Downloads/test_mongo:/data/db mongo:5.0.6` provides the MongoDB instance expected by repository code.

## Coding Style & Naming Conventions
- Always run `go fmt ./...`; rely on gofmt tabs, import grouping, and blank-line spacing.
- Exported APIs use PascalCase, unexported helpers stay camelCase, and new Cobra files mirror their CLI verb (e.g., `task_add.go` for `task-add`).
- Prefer structured logging via `slog` instead of `fmt.Printf` in runtime paths.

## Testing Guidelines
- Add `_test.go` files in the same package; table-driven tests work well for service rules.
- Stub interfaces from `internal/domain/repository` to avoid live HTTP or Mongo calls; keep tests deterministic.
- Ensure `go test ./...` passes before sending a review and document any data fixtures in the PR.

## Commit & Pull Request Guidelines
- Match the concise imperative style in history (`Review task command`, `Add task manager`) and keep each commit focused.
- PRs should describe behaviour changes, list manual verification commands, and link tracker issues when available.
- Attach terminal captures or screenshots for CLI UI adjustments and tag the maintainer responsible for the touched module.

## Configuration & Environment
- Update `config.TrackerDomain` when switching environments and mention the target URL in your PR notes.
- Store secrets outside the repo (e.g., local `.env` ignored by git) and document any new ports or Docker services teammates must start.
