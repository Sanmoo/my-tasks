# Tasks CLI

A simple, powerful, and beautiful command-line task manager written in Go.

## Features

- Create, list, update, and delete tasks
- Mark tasks as completed or reopen them
- Organize tasks with priorities (1-5) and tags
- Beautiful terminal output with colors and tables
- JSON-based storage (no database required)
- Clean architecture with separation of concerns
- High code quality with extensive linting

## Installation

### Prerequisites

- Go 1.25 or higher
- (Optional) golangci-lint for code quality checks
- (Optional) gofumpt for strict formatting

### Build from source

```bash
# Clone the repository
git clone https://github.com/Sanmoo/my-tasks.git
cd my-tasks

# Download dependencies
go mod download

# Build the application
make build

# Install globally (optional)
make install
```

## Usage

### Add a task

```bash
# Simple task
tasks add "Write project documentation"

# Task with description and priority
tasks add "Fix critical bug" -d "Memory leak in parser" -p 5

# Task with tags
tasks add "Review PR #123" -t "review,urgent" -p 4
```

### List tasks

```bash
# List all tasks
tasks list

# Filter by status
tasks list --status pending
tasks list --status completed

# Filter by priority
tasks list --priority 5

# Filter by tags
tasks list --tags urgent,review
```

### View task details

```bash
tasks show <task-id>
```

### Complete a task

```bash
tasks complete <task-id>
```

### Reopen a completed task

```bash
tasks reopen <task-id>
```

### Update a task

```bash
# Update title
tasks update <task-id> --title "New title"

# Update multiple fields
tasks update <task-id> -d "Updated description" -p 4 -t "tag1,tag2"
```

### Delete a task

```bash
# With confirmation prompt
tasks delete <task-id>

# Skip confirmation
tasks delete <task-id> --force
```

## Configuration

Tasks are stored in `~/.tasks/tasks.json` by default.

You can customize the storage location by setting the `TASKS_DATA_DIR` environment variable:

```bash
export TASKS_DATA_DIR=/path/to/your/data
```

## Project Structure

```
my-tasks/
├── cmd/
│   └── tasks/
│       └── main.go           # Application entrypoint
├── internal/
│   ├── app/
│   │   └── app.go           # Application configuration
│   ├── task/
│   │   ├── task.go          # Task domain model
│   │   ├── repository.go    # Storage interface
│   │   └── service.go       # Business logic
│   └── storage/
│       └── json.go          # JSON storage implementation
├── pkg/
│   └── cli/
│       ├── root.go          # Root command
│       ├── add.go           # Add command
│       ├── list.go          # List command
│       ├── complete.go      # Complete command
│       ├── reopen.go        # Reopen command
│       ├── delete.go        # Delete command
│       ├── update.go        # Update command
│       └── show.go          # Show command
├── .golangci.yml            # Linter configuration
├── Makefile                 # Build automation
├── go.mod                   # Go module definition
└── README.md                # This file
```

## Development

### Prerequisites for development

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install gofumpt
go install mvdan.cc/gofumpt@latest
```

### Available make targets

```bash
make build      # Build the application
make test       # Run tests with coverage
make lint       # Run linters
make fmt        # Format code
make clean      # Remove build artifacts
make install    # Install binary to $GOPATH/bin
make run        # Build and run
make help       # Show all available targets
```

### Code Quality

This project follows strict code quality standards:

- **Linting**: golangci-lint with 20+ enabled linters
- **Formatting**: gofumpt (stricter than gofmt)
- **Testing**: Test coverage with race detection
- **Architecture**: Clean architecture with clear separation of concerns

### Running tests

```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...
```

### Code style

- Follow standard Go conventions
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions small and focused
- Handle errors explicitly

## Architecture

The project follows clean architecture principles:

1. **Domain Layer** (`internal/task`): Core business logic and domain models
2. **Storage Layer** (`internal/storage`): Data persistence implementations
3. **Application Layer** (`internal/app`): Application setup and configuration
4. **Presentation Layer** (`pkg/cli`): CLI commands and user interface

### Key Design Patterns

- **Repository Pattern**: Abstract storage implementation
- **Dependency Injection**: Services receive dependencies via constructors
- **Interface Segregation**: Small, focused interfaces
- **Single Responsibility**: Each package has one clear purpose

## Dependencies

### Core Dependencies

- [cobra](https://github.com/spf13/cobra) - CLI framework
- [pterm](https://github.com/pterm/pterm) - Beautiful terminal output
- [uuid](https://github.com/google/uuid) - UUID generation

### Development Dependencies

- [golangci-lint](https://github.com/golangci/golangci-lint) - Linter aggregator
- [gofumpt](https://github.com/mvdan/gofumpt) - Stricter formatter

## Contributing

Contributions are welcome! Please ensure:

1. All tests pass: `make test`
2. Code is properly formatted: `make fmt`
3. No linting errors: `make lint`
4. Add tests for new functionality

## License

MIT License - feel free to use this project as you wish.

## Future Enhancements

Potential features for future versions:

- [ ] SQLite storage backend option
- [ ] Task due dates and reminders
- [ ] Task dependencies
- [ ] Task search functionality
- [ ] Export tasks to different formats (CSV, Markdown)
- [ ] Subtasks support
- [ ] Task templates
- [ ] Color themes customization
- [ ] Shell completion scripts

## Support

For issues, questions, or contributions, please open an issue on GitHub.
