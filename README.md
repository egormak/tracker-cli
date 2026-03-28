# Tracker CLI

A command-line time tracking application written in Go.

## Build

```shell
go build -o tracker ./cmd/app/main.go
sudo mv tracker /usr/local/bin/tracker
```

## Run

```shell
# Quick run during development
go run ./cmd/app/main.go [command]

# Run installed binary
tracker [command]
```

## Backend Requirements

The application requires a backend service running. Configure the backend URL in `config/config.go`.

### MongoDB (for local backend development)
```shell
docker run -it --rm -p 27017:27017 -v /home/egorka/Downloads/test_mongo:/data/db mongo:5.0.6
```

## Available Commands

- `tracker menu` - Interactive menu to select and start tasks
- `tracker task -n "name" [-t time] [-p percent]` - Run a task timer
- `tracker task -n "name" --previous-days` - Run a task from previous days with schedule awareness (searches Monday to today)
- `tracker taskadd` - Add a new task
- `tracker tasklist` - List all tasks
- `tracker statistic` - Show statistics for the day
- `tracker rest-spend -d [duration]` - Record rest time
- `tracker plan` - Planning features
- `tracker plan-percent` - Work with planning percentages
- `tracker timer-recheck` - Recheck timers
- `tracker timer-list-set` - Set timer list
- `tracker config` - Configure application settings
- `tracker role-recheck` - Recheck role settings
- `tracker clean` - Clean/manage data

Use `tracker --help` or `tracker [command] --help` for detailed usage information.

## Technologies

- **Go 1.22** - Core language
- **Cobra** - CLI framework
- **Bubble Tea** - Terminal UI framework
- **Lipgloss** - Terminal styling
- **slog** - Structured logging

## Documentation

See `CLAUDE.md` for detailed project documentation and development guidelines.