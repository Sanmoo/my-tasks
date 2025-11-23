# Kanban Task Manager

A powerful CLI tool for managing tasks organized in markdown-based kanban boards with built-in reminders and due dates.

## Features

- **Markdown-based Storage**: Organize tasks in simple markdown files with intuitive syntax
- **Kanban Board Views**: Beautiful terminal output showing tasks organized by phases/columns
- **Smart Reminders**: Built-in reminder system with overdue task notifications
- **Project Management**: Support for multiple projects with aliases and configuration
- **Due Date Tracking**: Track task deadlines with automatic overdue detection
- **Tag Support**: Organize tasks with custom tags for better categorization
- **Clean Architecture**: Well-structured Go codebase with comprehensive testing

## Installation

### Prerequisites

- Go 1.23 or higher
- (Optional) golangci-lint for code quality checks
- (Optional) gofumpt for strict formatting

### Build from Source

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

### Project Structure

Tasks are organized in markdown files within the `kanbans/` directory. Each file represents a project with kanban-style columns:

```markdown
# Project Name

## Phase Name

* Task title
* Another task
  * @tags tag1,tag2
  * @remind(25-10-09 23:00) reminder message
  * @due(25-10-23) due date
  * Additional comments as sub-bullets
```

### List Tasks from Projects

```bash
# List tasks from specific projects (comma-separated)
tasks list project1,project2

# List tasks with status filtering
tasks list project1 --status pending,running

# Use project aliases (configured in config.yaml)
tasks list alias1,alias2
```

### Check Overdue Reminders

```bash
# Show all overdue reminders across all projects
tasks remind
```

## Configuration

### Project Configuration

Create a `config.yaml` file in your project directory:

```yaml
project_aliases:
  alias1: "Full Project Name"
  alias2: "Another Project Name"

default_project: "alias1"
default_timezone: "Europe/Lisbon"
```

### Markdown File Format

Tasks are organized in markdown files with the following structure:

```markdown
# My Project

## Backlog

* Task in backlog
* Another backlog task
  * @tags important,backend

## 🏃 Doing

* Current active task
  * Working on API integration
  * @remind(25-10-10 18:00) check progress

## 🗓️ Scheduled

* Upcoming task
  * @due(25-10-15) Deadline for completion
  * @remind(25-10-14 09:00) start working on this

## ✅ Done

* Completed task
```

### Directives

- `@tags(tag1,tag2)`: Add tags to tasks
- `@remind(YYYY-MM-DD HH:MM) message`: Set reminders with messages
- `@due(YYYY-MM-DD)`: Set due dates for tasks

## Project Structure

```
my-tasks/
├── cmd/
│   └── tasks/
│       └── main.go              # Application entrypoint
├── internal/
│   ├── app/
│   │   ├── app.go              # Application configuration
│   │   └── app_test.go         # Configuration tests
│   ├── storage/
│   │   ├── markdown.go         # Markdown storage implementation
│   │   └── markdown_test.go    # Storage tests
│   └── task/
│       ├── model.go            # Domain models (Project, Phase, Task)
│       ├── model_test.go       # Model tests
│       ├── repository.go       # Storage interface
│       ├── service.go          # Business logic
│       └── service_test.go     # Service tests
├── kanbans/                    # Markdown task files
│   ├── sample.md               # Example project file
│   └── tasks.md                # Your task files
├── pkg/
│   ├── cli/
│   │   ├── root.go             # Root command
│   │   ├── list.go             # List command
│   │   ├── remind.go           # Remind command
│   │   └── *.test.go           # CLI tests
│   └── views/
│       ├── kanban.go           # Kanban board rendering
│       ├── reminder-panel.go   # Reminder panel rendering
│       └── *.test.go           # View tests
├── .golangci.yml               # Linter configuration
├── .markdownlint.jsonc         # Markdown linting rules
├── Makefile                    # Build automation
├── go.mod                      # Go module definition
└── README.md                   # This file
```

## Development

### Prerequisites for Development

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install gofumpt
go install mvdan.cc/gofumpt@latest
```

### Available Make Targets

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
- **Testing**: Comprehensive test coverage with race detection
- **Architecture**: Clean architecture with clear separation of concerns

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...
```

## Architecture

The project follows clean architecture principles:

1. **Domain Layer** (`internal/task`): Core business logic and domain models (Project, Phase, Task)
2. **Storage Layer** (`internal/storage`): Markdown file parsing and data persistence
3. **Application Layer** (`internal/app`): Application setup, configuration, and service orchestration
4. **Presentation Layer** (`pkg/cli` and `pkg/views`): CLI commands and terminal rendering

### Key Design Patterns

- **Repository Pattern**: Abstract storage implementation
- **Dependency Injection**: Services receive dependencies via constructors
- **Interface Segregation**: Small, focused interfaces
- **Single Responsibility**: Each package has one clear purpose

## Dependencies

### Core Dependencies

- [cobra](https://github.com/spf13/cobra) - CLI framework
- [pterm](https://github.com/pterm/pterm) - Beautiful terminal output
- [go-snaps](https://github.com/gkampitakis/go-snaps) - Snapshot testing

### Development Dependencies

- [golangci-lint](https://github.com/golangci/golangci-lint) - Linter aggregator
- [gofumpt](https://github.com/mvdan/gofumpt) - Stricter formatter
- [testify](https://github.com/stretchr/testify) - Testing utilities

## Contributing

Contributions are welcome! Please ensure:

1. All tests pass: `make test`
2. Code is properly formatted: `make fmt`
3. No linting errors: `make lint`
4. Add tests for new functionality
5. Update documentation accordingly

## License

MIT License - feel free to use this project as you wish.

## Future Enhancements

Potential features for future versions:

- [ ] Interactive task management (add, edit, complete tasks)
- [ ] Multiple storage backends (JSON, SQLite, etc.)
- [ ] Web interface for task management
- [ ] Task search and filtering
- [ ] Export tasks to different formats
- [ ] Task templates and recurring tasks
- [ ] Integration with calendar systems
- [ ] Mobile companion app

## Support

For issues, questions, or contributions, please open an issue on GitHub.