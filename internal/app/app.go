package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sanmoo/my-tasks/internal/storage"
	"github.com/Sanmoo/my-tasks/internal/task"
)

// App holds the application configuration and dependencies
type App struct {
	TaskService *task.Service
	Config      *Config
}

// Config holds application configuration
type Config struct {
	DataDir  string
	DataFile string
}

// New creates and initializes a new application instance
func New() (*App, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize storage
	repo, err := storage.NewMarkdownStorage(config.DataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize service
	taskService := task.NewService(repo)

	return &App{
		TaskService: taskService,
		Config:      config,
	}, nil
}

// loadConfig loads application configuration
func loadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dataDir := filepath.Join(homeDir, ".tasks", "kanbans")
	dataFile := filepath.Join(dataDir, "tasks.md")

	// Check for custom data directory from environment
	if customDir := os.Getenv("TASKS_DATA_DIR"); customDir != "" {
		dataDir = customDir
		dataFile = filepath.Join(dataDir, "tasks.md")
	}

	// Check for custom data file from environment
	if customFile := os.Getenv("TASKS_DATA_FILE"); customFile != "" {
		dataFile = customFile
		dataDir = filepath.Dir(dataFile)
	}

	return &Config{
		DataDir:  dataDir,
		DataFile: dataFile,
	}, nil
}
