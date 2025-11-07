.PHONY: build test lint fmt clean install help run

# Build the application
build:
	@echo "Building tasks..."
	@go build -o bin/tasks cmd/tasks/main.go

# Run tests with race detection and coverage
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	@gofumpt -w .
	@go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out

# Install the application
install: build
	@echo "Installing tasks..."
	@cp bin/tasks $(GOPATH)/bin/tasks

# Run the application
run: build
	@./bin/tasks

# Display help
help:
	@echo "Available targets:"
	@echo "  build    - Build the application"
	@echo "  test     - Run tests with race detection and coverage"
	@echo "  lint     - Run golangci-lint"
	@echo "  fmt      - Format code with gofumpt"
	@echo "  clean    - Remove build artifacts"
	@echo "  install  - Install the binary to GOPATH/bin"
	@echo "  run      - Build and run the application"
	@echo "  help     - Show this help message"
