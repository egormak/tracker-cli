## Project Overview

This project is a command-line interface (CLI) tool named `tracker_cli` for time and task tracking. It is written in Go and uses the Cobra library for command management and Bubble Tea for an interactive text-based user interface (TUI).

The application acts as a client for a backend service. All data is persisted by making REST API calls to the backend.

## Building and Running

### Build

To build the application, run the following command from the root of the project:

```shell
go build -o tracker ./cmd/app/main.go
```

### Installation

After building, you can move the binary to a directory in your system's PATH to make it accessible from anywhere:

```shell
sudo mv tracker /usr/local/bin/tracker
```

### Running

To run the application, you need to have the backend service running. 

**[TODO: Add instructions on how to run the backend service]**

Once the backend service is running, you can use the application's commands:

```shell
tracker [command]
```

### Commands

-   `tracker clean`: Run Clean.
-   `tracker rest-spend -d [duration]`: Set how much time you spent on rest.
-   `tracker timer-recheck`: Recheck timers and refresh the timer list.
-   `tracker statistic`: Show statistic.
-   `tracker task -n [task-name]`: Run Task.
-   `tracker taskadd -n [task-name] -r [task-role]`: Add a new task with a role.
-   `tracker tasklist`: Show the list of tasks.

## Development Conventions

The project follows standard Go project conventions and a clean architecture. It uses Go modules for dependency management.

### Project Structure

The code is structured into the following main packages:

-   `cmd`: Contains the CLI command definitions using the Cobra library. Each command is in its own file in the `cmd/command` directory.
-   `internal/domain`: Defines the core data structures (entities) of the application.
-   `internal/service`: Implements the business logic for each command.
-   `internal/repository`: Handles data access. In this project, it's responsible for making API calls to the backend service.
-   `config`: Handles application configuration.

### TUI

The application uses the Bubble Tea library for its TUI, which suggests that the user interface is built around a model-view-update (MVU) architecture.
