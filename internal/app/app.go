package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sanmoo/my-tasks/internal/storage"
	"github.com/Sanmoo/my-tasks/internal/task"
	"gopkg.in/yaml.v3"
)

// App holds the application configuration and dependencies
type App struct {
	TaskService *task.Service
	Config      *Config
}

// Config holds application configuration
type Config struct {
	DataDir        string
	DataFile       string
	ProjectAliases map[string]string // Maps alias to full project name
	ProjectFiles   []string          // List of markdown files to search for projects
}

// YAMLConfig represents the structure of the YAML configuration file
type YAMLConfig struct {
	Project struct {
		Aliases map[string]string `yaml:"aliases"`
		Files   []string          `yaml:"files"`
	} `yaml:"project"`
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

	// Default configuration
	dataDir := filepath.Join(homeDir, ".tasks", "kanbans")
	dataFile := filepath.Join(dataDir, "tasks.md")
	projectAliases := make(map[string]string)
	projectFiles := []string{}

	// Try to load YAML configuration
	configPath := filepath.Join(homeDir, ".config", "tasks", "tasks.conf.yaml")
	yamlConfig, err := loadYAMLConfig(configPath)
	if err == nil {
		// YAML config loaded successfully
		projectAliases = yamlConfig.Project.Aliases
		projectFiles = make([]string, len(yamlConfig.Project.Files))

		// Expand environment variables in file paths
		for i, file := range yamlConfig.Project.Files {
			projectFiles[i] = os.ExpandEnv(file)
		}
	}
	// If YAML config doesn't exist or fails to load, continue with defaults

	// Check for custom data directory from environment (overrides YAML)
	if customDir := os.Getenv("TASKS_DATA_DIR"); customDir != "" {
		dataDir = customDir
		dataFile = filepath.Join(dataDir, "tasks.md")
	}

	// Check for custom data file from environment (overrides YAML)
	if customFile := os.Getenv("TASKS_DATA_FILE"); customFile != "" {
		dataFile = customFile
		dataDir = filepath.Dir(dataFile)
	}

	return &Config{
		DataDir:        dataDir,
		DataFile:       dataFile,
		ProjectAliases: projectAliases,
		ProjectFiles:   projectFiles,
	}, nil
}

// loadYAMLConfig loads configuration from a YAML file
func loadYAMLConfig(path string) (*YAMLConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config YAMLConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return &config, nil
}

// ResolveProjectName resolves a project alias to its full name
// If the name is not an alias, returns the name as-is
func (c *Config) ResolveProjectName(nameOrAlias string) string {
	if fullName, exists := c.ProjectAliases[nameOrAlias]; exists {
		return fullName
	}
	return nameOrAlias
}

// GetProjectFile finds the markdown file that contains the specified project
// It searches through all configured project files
func (a *App) GetProjectFile(projectName string) (string, error) {
	// Resolve alias to full project name
	fullProjectName := a.Config.ResolveProjectName(projectName)

	// If no project files are configured, use the default data file
	if len(a.Config.ProjectFiles) == 0 {
		return a.Config.DataFile, nil
	}

	// Search through all project files
	for _, file := range a.Config.ProjectFiles {
		// Try to load the storage and check if the project exists
		repo, err := storage.NewMarkdownStorage(file)
		if err != nil {
			continue // Skip files that can't be loaded
		}

		// Check if this file contains the project by listing tasks
		tasks, err := repo.List(nil, task.Filter{})
		if err != nil {
			continue
		}

		// Check if any task belongs to this project
		for _, t := range tasks {
			if t.Project == fullProjectName {
				return file, nil
			}
		}
	}

	// If project not found in any file, return error
	return "", fmt.Errorf("project '%s' not found in any configured file", fullProjectName)
}

// NewWithProjectFile creates an App instance using a specific project file
func NewWithProjectFile(projectFile string) (*App, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize storage with the specified project file
	repo, err := storage.NewMarkdownStorage(projectFile)
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
