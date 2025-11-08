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
	repo, err := storage.NewMarkdownStorage(
		config.ProjectFiles,
		config.ProjectAliases,
	)
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

	// Try to load YAML configuration
	configPath := filepath.Join(homeDir, ".config", "tasks", "tasks.conf.yaml")
	yamlConfig, err := loadYAMLConfig(configPath)
	if err != nil {
		return nil, err
	}

	// YAML config loaded successfully
	projectAliases := yamlConfig.Project.Aliases
	projectFiles := make([]string, len(yamlConfig.Project.Files))

	// Expand environment variables in file paths
	for i, file := range yamlConfig.Project.Files {
		projectFiles[i] = os.ExpandEnv(file)
	}

	return &Config{
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

func NewApp() (*App, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize storage with the specified project file
	repo, err := storage.NewMarkdownStorage(config.ProjectFiles, config.ProjectAliases)
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
