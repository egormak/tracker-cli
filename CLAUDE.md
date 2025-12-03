# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Tracker CLI is a command-line interface (CLI) tool for time and task tracking written in Go. It uses the Cobra library for command management and Bubble Tea for an interactive text-based user interface (TUI). The application acts as a client for a backend service, with all data persisted by making REST API calls to the backend.

## Build and Run Commands

### Build the Application

```shell
# Build the application
go build -o tracker ./cmd/app/main.go

# Install the binary to make it globally accessible
sudo mv tracker /usr/local/bin/tracker
```

### Run the Application

```shell
# Quick run during development
go run ./cmd/app/main.go [command]

# Run the installed binary
tracker [command]
```

### Backend Requirements

The application requires a backend service running. The backend URL is configured in `config/config.go`. Currently set to `http://127.0.0.1:3000` for local development. Production URL `http://tracker.makegorka.com:8080` is available but commented out. Switch between them as needed.

```shell
# Run MongoDB for local development (if you're running the backend locally)
docker run -it --rm -p 27017:27017 -v /home/egorka/Downloads/test_mongo:/data/db mongo:5.0.6
```

## Code Architecture

### Project Structure

The codebase follows a clean architecture pattern with the following main components:

1. **Command Layer (`cmd/`)**
   - `cmd/app/main.go`: Entry point that sets up logging and delegates to Cobra commands
   - `cmd/command/*.go`: Individual command implementations using Cobra library

2. **Domain Layer (`internal/domain/`)**
   - `entity/`: Core data structures that represent business objects
     - `task.go`: Task-related data structures (TaskManager, TaskParams, etc.)
     - `role.go`: Role-related data structures
     - `statistic.go`: Statistics-related data structures
     - `timers.go`: Timer-related data structures
     - `general.go`: General purpose data structures like Answer

3. **Service Layer (`internal/service/`)**
   - Business logic for each feature area
   - `task/`: Task management functionality including interactive timer
   - `task_params/`: Task parameters management (time, priority)
   - `menu/`: Interactive menu for task selection
   - `statistic/`: Statistics gathering and display
   - `procent/`: Percentage calculation and management
   - `rest/`: Rest time tracking
   - `telegram/`: Telegram notifications
   - `timer/`: Timer management
   - `role/`: Role management
   - `plan/`: Planning features
   - `manager/`: Cleanup and management tasks

4. **Repository Layer (`internal/repository/`)**
   - `api/`: Handles REST API calls to the backend service
   - All API communication goes through the centralized `sendRequest()` function in `api.go`

5. **Domain Layer (`internal/domain/`)**
   - `entity/`: Core business entities (task, role, statistic, timer, general)
   - `repository/`: Repository interfaces (currently empty, implementations in internal/repository/api)

6. **Interface Layer (`internal/interface/`)**
   - `cli/`: CLI interface utilities (currently empty)

7. **Utilities (`internal/pkg/`)**
   - `day_method/`: Date-related utilities
   - `restutil/`: Rest-related utilities

### Key Workflow

1. User runs a command like `tracker task -n "Task Name"`
2. The command handler in `cmd/command/task.go` processes the flags and calls the appropriate service
3. The service (e.g., `internal/service/task/task.go`) implements the business logic
4. The service uses the repository layer (`internal/repository/api`) to interact with the backend
5. For interactive features, Bubble Tea models are used to create TUI interfaces

## Key Commands

- `tracker menu`: Interactive menu to select and start tasks
- `tracker task -n "Task Name" [-t time] [-p percent]`: Start a task timer
- `tracker taskadd -n "Task Name" -r "Role"`: Add a new task
- `tracker tasklist`: Show the list of tasks
- `tracker statistic`: Show statistics
- `tracker rest-spend -d [duration]`: Track rest time
- `tracker timer-recheck`: Recheck timers and refresh timer list
- `tracker config`: Configure application settings
- `tracker role-recheck`: Recheck role settings
- `tracker timer-list-set`: Set timer list
- `tracker plan-percent`: Work with planning percentages

## UI Components

The application uses Bubble Tea for interactive TUI components:

1. **Task Timer UI** (`internal/service/task/task_timer.go`):
   - Displays a running timer for tasks
   - Shows elapsed and remaining time
   - Provides keyboard controls for pause/resume, stop, and quit

2. **Menu UI** (`internal/service/menu/menu.go`):
   - Displays a table of available tasks with details
   - Allows selection of a task to start

## API Communication

All data is stored and retrieved via REST API calls to the backend:

- Base URL is configured in `config/config.go` (current: `http://127.0.0.1:3000`)
- API calls are implemented in `internal/repository/api/*.go`
- Standard HTTP client with JSON serialization/deserialization
- Centralized `sendRequest()` function in `internal/repository/api/api.go` handles all HTTP requests
- 15-second timeout on all requests
- Proper error handling with structured logging

### Key API Endpoints

**Task Management:**
- `GET /api/v1/task/params?task_name=X` - Get task planning parameters
- `POST /api/v1/taskrecord` - Record completed task time
- `GET /api/v1/records` - Get all task records
- `GET /api/v1/tasklist` - List all tasks with statistics
- `GET /api/v1/record/task-day?task_name=X` - Get time spent on task today
- `GET /api/v1/record/params?task_name=X` - Get task parameters
- `POST /api/v1/record/params` - Set task parameters

**Statistics:**
- `GET /api/v1/stats/done/today` - Today's completion statistics
- `GET /api/v1/stats/tasks/today` - Today's tasks (planned vs done)

**Planning:**
- `GET /api/v1/task/plan-percent` - Get next task by percent
- `GET /api/v1/task/plan/percent/schedule` - Get next task with schedule awareness
- `GET /api/v1/task/plan-percent/change` - Change percent plan
- `POST /api/v1/manage/procents` - Manage percents

**Timer Management:**
- `GET /api/v1/timer/get` - Get default timer duration
- `POST /api/v1/timer/set` - Set timer count
- `POST /api/v1/timer/del` - Delete timer count
- `GET /api/v1/manage/timer/recheck` - Recheck timer state

**Role Management:**
- `GET /api/v1/role/get?task_name=X` - Get task role
- `GET /api/v1/roles/records` - Get all roles

**Rest Time:**
- `GET /api/v1/rest-get` - Get rest time

**Telegram Notifications:**
- `POST /api/v1/manage/telegram/start` - Start notification
- `POST /api/v1/manage/telegram/stop` - Stop notification

**Cleanup:**
- `GET /api/v1/records/clean` - Clean records

## Tips for Development

- Use the interactive menu (`tracker menu`) to quickly see available tasks and start timers
- Check the output of `tracker statistic` to see current task progress
- When implementing new commands, follow the pattern in `cmd/command/*.go` and register them in the Cobra command tree
- For interactive features, leverage the Bubble Tea library patterns shown in existing implementations