#!/bin/bash

# Setup script for Tasks CLI
# This script initializes the Go module and downloads dependencies

set -e

echo "Tasks CLI - Setup Script"
echo "========================"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Please install Go 1.21 or higher from https://go.dev/dl/"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "Go version: $GO_VERSION"
echo ""

# Download dependencies
echo "Downloading dependencies..."
go mod download
echo ""

# Verify dependencies
echo "Verifying dependencies..."
go mod verify
echo ""

# Build the application
echo "Building application..."
make build
echo ""

echo "Setup complete!"
echo ""
echo "You can now run the application with:"
echo "  ./bin/tasks"
echo ""
echo "Or install it globally with:"
echo "  make install"
echo ""
echo "For more information, see README.md"
