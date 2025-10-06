---
applyTo: '**'
---

# Tracker CLI - Project Instructions

## Project Overview

This is a **time tracking CLI application** written in Go that helps users track tasks, manage timers, and view statistics. The application follows a **clean architecture** pattern with clear separation of concerns between domain, application, and infrastructure layers.

## Architecture

### Directory Structure
```
cmd/
├── app/                    # Application entry point (main.go)
└── command/                # Cobra CLI command definitions
    ├── root.go            # Root command and Execute()
    ├── task.go            # Task execution command
    ├── taskadd.go         # Add task command
    ├── tasklist.go        # List tasks command
    ├── manager.go         # Clean/manager command
    ├── statistic.go       # Statistics display command
    ├── rest.go            # Rest time management
    ├── plan.go            # Plan parent command
    └── plan_percent.go    # Percentage-based planning

internal/
├── domain/
│   ├── entity/            # Core business entities
│   │   ├── task.go       # Task*, TaskList, TaskParams structs
│   │   ├── role.go       # RoleAnswer struct
│   │   ├── statistic.go  # TaskTimeDurationResponse
│   │   ├── timers.go     # Timer-related entities
│   │   └── general.go    # General/shared entities
│   └── repository/       # Repository interfaces (contracts)
├── repository/
│   └── api/              # External API communication implementations
│       ├── api.go        # sendRequest() - centralized HTTP client
│       ├── task.go       # GetTaskParams, AddTaskRecord, GetTaskRecords
│       ├── statistic.go  # StatisticTaskGet
│       ├── role.go       # TaskRoleGet
│       ├── timer.go      # Timer-related API calls
│       ├── plan_percent.go # Percentage planning API
│       └── clean_data.go # Data cleanup operations
├── service/              # Business logic layer
│   ├── task/            # Task management services
│   │   ├── task.go      # TaskRun, CreateTaskTimer, duration calculation
│   │   ├── task_timer.go # TaskTimer.Run(), Start(), Stop()
│   │   ├── task_structure.go # TaskTimer struct definition
│   │   └── task_methods.go # Task helper methods
│   ├── timer/           # Timer management
│   ├── telegram/        # Telegram notification integration
│   ├── statistic/       # Statistics calculation and display
│   ├── rest/            # Rest time tracking
│   ├── menu/            # Interactive Bubble Tea menu
│   │   ├── menu.go     # RunMenu() - interactive task selector
│   │   └── const_values.go # Menu constants
│   ├── plan/            # Planning features
│   ├── procent/         # Percentage calculations
│   ├── role/            # Role management
│   └── manager/         # Cleanup and management tasks
├── application/         # Application-level services
│   ├── service/        # Application services
│   └── errors/         # Custom error types
├── infrastructure/      # Infrastructure adapters
│   ├── api/           # API infrastructure
│   └── repository/    # Repository implementations
├── interface/
│   └── cli/           # CLI interface utilities
│       └── cli.go
└── pkg/               # Shared utilities
    └── day_method/    # Day-related utility functions

config/
└── config.go          # TrackerDomain constant configuration

test/
└── main.go           # Bubble Tea UI experiments/demos
```

### Key Technologies
- **Go 1.22** - Primary language with toolchain
- **Cobra** - CLI framework for commands and flags (`github.com/spf13/cobra`)
- **Bubble Tea** - Terminal UI framework (`github.com/charmbracelet/bubbletea`)
- **Bubbles** - Terminal UI components (`github.com/charmbracelet/bubbles`)
- **Lipgloss** - Terminal styling (`github.com/charmbracelet/lipgloss`)
- **slog** - Structured logging (stdlib)
- **tint** - Colored slog output (`github.com/lmittmann/tint`)
- **HTTP Client** - Standard library net/http for REST API communication

### Application Flow
1. **Entry**: `cmd/app/main.go` initializes slog with tint handler and calls `command.Execute()`
2. **Routing**: `cmd/command/root.go` defines the root Cobra command tree
3. **Execution**: Command handlers in `cmd/command/*.go` delegate to services in `internal/service/`
4. **Business Logic**: Services in `internal/service/*` implement core functionality
5. **Data Access**: Services call `internal/repository/api/*` for external API communication
6. **API Layer**: `internal/repository/api/api.go` provides centralized `sendRequest()` function

## Coding Standards

### Go Conventions
- Follow standard Go naming conventions (camelCase for private, PascalCase for public)
- Use meaningful package names that reflect their purpose
- Implement proper error handling with structured logging
- Use Go modules for dependency management
- Run `go fmt ./...` before committing
- Keep packages focused on single responsibilities

### Error Handling
- Use structured logging with `slog` for consistent error reporting
- Include context in error messages: `slog.Error("operation failed", "error", err, "context", value)`
- **Critical Operations**: Exit with `os.Exit(1)` on unrecoverable errors (API failures, JSON decode errors)
- **Command Handlers**: Return errors from RunE handlers for Cobra to handle
- **Service Layer**: Return wrapped errors with context: `fmt.Errorf("description: %w", err)`
- Define custom error variables for business logic: `var ErrTaskCompleted = fmt.Errorf("no time remaining for this task")`
- Use `errors.Is()` to check for specific error types

### HTTP Communication
- All external API calls go through the `internal/repository/api/` package
- Use the centralized `sendRequest` function in `internal/repository/api/api.go`
- Set appropriate timeouts (15 seconds default)
- Handle HTTP status codes properly:
  - 200: Success
  - 404: Not Found
  - 500: Internal Server Error
- Use proper JSON marshaling/unmarshaling with struct tags
- Always defer `resp.Body.Close()` when reading response bodies
- Set headers: `Content-Type: application/json` and `Accept: application/json`

### CLI Commands
- Each command should be in its own file in `cmd/command/`
- Use Cobra's flag system for parameters with short and long forms
- Mark required flags: `cmd.MarkFlagRequired("flag")`
- Provide clear help text in `Use` and `Short` fields
- Use `RunE` for commands that can return errors, `Run` for commands that cannot
- Initialize commands in `init()` functions and register with `rootCmd.AddCommand(cmd)`
- Follow the pattern: `var cmdName = &cobra.Command{...}`

### Entity Design
- Keep entities in `internal/domain/entity/` 
- Use proper JSON tags for API communication: `json:"field_name"`
- Separate request/response structures when needed (e.g., `TaskRecorcRequest`)
- Include time.Time fields for temporal data
- Keep entities focused on data structure, not behavior
- Use descriptive struct names that indicate purpose (e.g., `TaskTimeDurationResponse`)

### Service Layer
- Business logic goes in `internal/service/` subdirectories organized by domain
- Each service package should focus on a single domain area (task, timer, statistics, menu, etc.)
- Services coordinate between repositories and implement business rules
- Keep service functions cohesive and focused
- Extract helper functions for calculations (e.g., `calculateDuration`, `calculateTimeLeft`)
- Services should accept Cobra command context when used as command handlers
- Keep complex logic in services, not in command files

### Repository Layer
- All repository implementations in `internal/repository/api/`
- Use the centralized `sendRequest()` function from `api.go`
- Handle JSON encoding/decoding within repository functions
- Return domain entities, not raw HTTP responses
- Log errors with context before returning or exiting
- Keep repository functions focused on single API operations

## Specific Guidelines

### Task Management
- **Task Structure**: Tasks have name, role, duration, time begin/end, time done, percent, and Telegram message ID
- **Task Types**: 
  - `TaskManager`: Full task with timing and tracking
  - `TaskParams`: Task planning parameters (name, time, priority)
  - `TaskList`: Task list view with statistics
  - `TaskRecorcRequest`: Recording task completion
- **Duration Calculation**: 
  - Support default duration from API or explicit duration flag
  - Calculate time left based on plan duration, percentage, and time done
  - Return `ErrTaskCompleted` when no time remains
  - Prefer time left over requested duration when less time remains
- **Task Execution Flow**:
  1. Get task parameters from API
  2. Get time already spent on task
  3. Calculate duration for this session
  4. Create TaskTimer and run
  5. Send Telegram start notification
  6. Run timer loop (minute by minute)
  7. Handle interruption signals gracefully
  8. Record time done to API
  9. Show statistics
  10. Send Telegram completion notification

### Timer Functionality
- **Signal Handling**: Implement graceful shutdown for SIGTERM and SIGINT
- **Timer Loop**: Sleep one minute per iteration, log progress
- **Timer Lifecycle**:
  - `Start()`: Sends Telegram start message, runs timer loop
  - `Stop()`: Calculates time done, records to API, shows stats, sends Telegram stop message
  - `Run()`: Sets up signal handling and coordinates Start/Stop
- **Background Process**: Use goroutine with channel communication for signal handling
- **Cleanup**: Always call `timer.TimeDurationDel()` after completion
- **Telegram Integration**: Track message ID to update Telegram notifications

### Interactive Menu System
- **Bubble Tea**: Use for interactive task selection in `internal/service/menu/menu.go`
- **Table Component**: Display tasks with columns: Name, Role, Priority, Duration, Done, Left
- **Keyboard Navigation**:
  - Arrow keys: Navigate table rows
  - Enter: Select task and return name
  - Q/Ctrl+C: Quit without selection
- **Styling**: Use Lipgloss for borders, colors, and highlighting
- **Data Source**: Fetch task list from `/api/v1/tasklist` endpoint
- **Sorting**: Sort tasks by priority (descending)
- **Return Value**: Return selected task name or empty string if cancelled

### Statistics & Planning
- **Statistics Display**: Show task-specific and full day statistics
- **Rest Tracking**: Track and display rest time separately
- **Percentage Planning**: Support percentage-based task completion planning
- **API Endpoints**:
  - `/api/v1/record/task-day?task_name=X`: Get time spent on specific task today
  - `/api/v1/task/params?task_name=X`: Get task planning parameters
  - `/api/v1/record`: POST to record completed time
  - `/api/v1/records`: GET all task records

### Telegram Notifications
- **Start Notification**: Send task name, receive message ID
- **Stop Notification**: Update message with task name, time done, end time
- **API Endpoints**:
  - `/api/v1/manage/telegram/start`: POST with task_name
  - `/api/v1/manage/telegram/stop`: POST with task_name, msg_id, time_done, time_end
- **Time Format**: Use "2 January 2006 15:04" for end time display

### Configuration
- Centralize configuration in `config/config.go`
- Use const for TrackerDomain: `http://tracker.makegorka.com:8080`
- Comment out alternative endpoints (e.g., localhost for development)
- No environment variables or config files currently used

### Testing
- Write tests for business logic in service layer
- Mock external dependencies (HTTP calls)
- Test error conditions and edge cases (e.g., task completed, no time left)
- Use table-driven tests for multiple scenarios
- Test files go in `test/` directory or alongside code as `*_test.go`

### Documentation
- Include clear comments for exported functions
- Document complex business logic (especially duration calculations)
- Provide usage examples in CLI help text
- Maintain README with build and usage instructions
- Document API endpoints and request/response formats

## Common Patterns

### HTTP Request Pattern
```go
func SomeAPICall(param string) (entity.Entity, error) {
    var result entity.Entity
    
    responseBody, err := sendRequest("GET", fmt.Sprintf("/api/v1/path?param=%s", param), nil)
    if err != nil {
        slog.Error("request error", "error", err)
        os.Exit(1)
    }
    
    err = json.NewDecoder(responseBody).Decode(&result)
    if err != nil {
        slog.Error("failed to decode response", "error", err)
        os.Exit(1)
    }
    
    return result, nil
}
```

### HTTP POST Pattern
```go
func SomeAPIPost(data entity.Request) entity.Response {
    var result entity.Response
    
    jsonData, err := json.Marshal(&data)
    if err != nil {
        slog.Error("can't marshal JSON", "error", err)
        os.Exit(1)
    }
    
    responseBody, err := sendRequest("POST", "/api/v1/path", bytes.NewBuffer(jsonData))
    if err != nil {
        slog.Error("request error", "error", err)
        os.Exit(1)
    }
    
    err = json.NewDecoder(responseBody).Decode(&result)
    if err != nil {
        slog.Error("failed to decode response", "error", err)
        os.Exit(1)
    }
    
    return result
}
```

### CLI Command Pattern
```go
var someCmd = &cobra.Command{
    Use:   "command-name",
    Short: "Brief description",
    RunE:  handlerFunction,
}

func init() {
    someCmd.Flags().StringP("name", "n", "", "Parameter description")
    someCmd.Flags().IntP("time", "t", 0, "Time duration")
    someCmd.MarkFlagRequired("name")
    rootCmd.AddCommand(someCmd)
}

func handlerFunction(cmd *cobra.Command, args []string) error {
    name, err := cmd.Flags().GetString("name")
    if err != nil {
        return fmt.Errorf("read name flag: %w", err)
    }
    // ... implementation
    return nil
}
```

### Logging Pattern
```go
slog.Info("operation started", "param", value)
slog.Info(fmt.Sprintf("Progress: %d/%d", current, total))
slog.Error("operation failed", "error", err, "context", additionalInfo)
```

### Signal Handling Pattern
```go
func (t *TaskTimer) Run() error {
    exitCh := make(chan os.Signal)
    signal.Notify(exitCh, syscall.SIGTERM, syscall.SIGINT)
    
    go func() {
        defer func() {
            exitCh <- syscall.SIGTERM
        }()
        t.Start()
    }()
    
    <-exitCh
    t.Stop()
    return nil
}
```

## Development Workflow

1. **Add New Entity**: Define struct in `internal/domain/entity/*.go` with JSON tags
2. **Implement Repository**: Add API function in `internal/repository/api/*.go`
3. **Create Service Logic**: Implement business logic in `internal/service/*/`
4. **Add CLI Command**: Create command file in `cmd/command/*.go`
5. **Register Command**: Add to root command in `init()` function
6. **Test**: Run `go test ./...` and manual testing with `go run ./cmd/app/main.go`
7. **Build**: Run `go build -o tracker ./cmd/app/main.go`
8. **Install**: Move binary to `/usr/local/bin/tracker` for global access

## Build and Run

### Build Binary
```bash
go build -o tracker ./cmd/app/main.go
sudo mv tracker /usr/local/bin/tracker
```

### Development Run
```bash
go run ./cmd/app/main.go --help
go run ./cmd/app/main.go task --name "Task Name" --time 25
```

### Testing
```bash
go test ./...
go test ./internal/service/task/...
```

### MongoDB (for testing)
```bash
docker run -it --rm -p 27017:27017 -v /home/egorka/Downloads/test_mongo:/data/db mongo:5.0.6
```

## Available Commands

- `tracker task -n "name" [-t time] [-p percent]` - Run a task timer
- `tracker taskadd` - Add a new task (interactive or with flags)
- `tracker tasklist` - List all tasks
- `tracker statistic` - Show statistics for the day
- `tracker rest-spend -d duration` - Record rest time
- `tracker plan` - Parent command for planning features
- `tracker clean` - Clean/manage data
- Use `--help` on any command for detailed usage

## External Dependencies

### Primary Tracker Service
- **Domain**: `tracker.makegorka.com:8080`
- **Protocol**: HTTP (not HTTPS)
- **Base URL**: Configured in `config/config.go` as `TrackerDomain`
- **Alternative**: Can switch to `http://127.0.0.1:3000` for local development

### API Endpoints (Examples)
- `GET /api/v1/task/params?task_name=X` - Get task parameters
- `GET /api/v1/record/task-day?task_name=X` - Get today's time for task
- `POST /api/v1/record` - Record completed task time
- `GET /api/v1/tasklist` - List all tasks with statistics
- `GET /api/v1/timer/get` - Get default timer duration
- `POST /api/v1/timer/set` - Set timer count
- `POST /api/v1/timer/del` - Delete timer count
- `POST /api/v1/manage/telegram/start` - Start Telegram notification
- `POST /api/v1/manage/telegram/stop` - Stop Telegram notification
- `GET /api/v1/manage/timer/recheck` - Recheck timer state

All API interactions must go through the repository layer to maintain separation of concerns and enable easy testing/mocking.