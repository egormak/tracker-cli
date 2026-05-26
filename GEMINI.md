# tracker_cli

A Go-based command-line interface for the tracker system. Uses Cobra for command management and Bubble Tea for an interactive TUI.

## Project Overview

- **Technologies**: Go 1.22+, Cobra, Bubble Tea, Lipgloss.
- **Architecture**:
  - `cmd/app/main.go`: Entry point, sets up logging and delegates to commands.
  - `cmd/command/`: Individual Cobra command implementations.
  - `internal/service/`: Business logic for features (task, timer, stats, etc.).
  - `internal/repository/api/`: Handles REST API calls to `tracker-server`.
  - `internal/domain/entity/`: Core data structures.
  - `config/config.go`: Configuration defaults, including `TrackerDomain`.

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
./tracker [command]
```

## Key Commands

- `tracker menu`: Interactive menu to select and start tasks (TUI).
- `tracker task -n "Name"`: Start a task timer with TUI.
- `tracker taskadd -n "Name" -r "Role"`: Add a new task.
- `tracker tasklist`: Show the list of tasks.
- `tracker statistic`: Show today's statistics.
- `tracker rest-spend -d [duration]`: Track rest time.
- `tracker timer-recheck`: Recheck and refresh timers.

## Development Conventions

- **API Communication**: All requests go through `internal/repository/api/api.go` via `sendRequest()`.
- **TUI**: Use Bubble Tea for interactive components. Follow the MVU (Model-View-Update) pattern.
- **Logging**: Use `slog` for structured logging.
- **Clean Architecture**: Maintain separation between command handlers, services, and repository layers.
- **Testing**: Add `_test.go` files in the same package. Stub repository interfaces for deterministic tests.

## Configuration
The backend URL is configured in `config/config.go`. Default is `http://127.0.0.1:3000`.
