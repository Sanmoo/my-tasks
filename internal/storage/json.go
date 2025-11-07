package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/Sanmoo/my-tasks/internal/task"
)

// JSONStorage implements task.Repository using a JSON file
type JSONStorage struct {
	filepath string
	mu       sync.RWMutex
}

// NewJSONStorage creates a new JSON storage
func NewJSONStorage(filePath string) (*JSONStorage, error) {
	// Create parent directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	storage := &JSONStorage{
		filepath: filePath,
	}

	// Initialize file if it doesn't exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := storage.save(map[string]*task.Task{}); err != nil {
			return nil, err
		}
	}

	return storage, nil
}

// Create creates a new task
func (s *JSONStorage) Create(ctx context.Context, t *task.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.load()
	if err != nil {
		return err
	}

	tasks[t.ID] = t
	return s.save(tasks)
}

// GetByID retrieves a task by its ID
func (s *JSONStorage) GetByID(ctx context.Context, id string) (*task.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks, err := s.load()
	if err != nil {
		return nil, err
	}

	t, exists := tasks[id]
	if !exists {
		return nil, task.ErrTaskNotFound
	}

	return t, nil
}

// List retrieves all tasks with optional filtering
func (s *JSONStorage) List(ctx context.Context, filter task.Filter) ([]*task.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks, err := s.load()
	if err != nil {
		return nil, err
	}

	var result []*task.Task
	for _, t := range tasks {
		if s.matchesFilter(t, filter) {
			result = append(result, t)
		}
	}

	return result, nil
}

// Update updates an existing task
func (s *JSONStorage) Update(ctx context.Context, t *task.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.load()
	if err != nil {
		return err
	}

	if _, exists := tasks[t.ID]; !exists {
		return task.ErrTaskNotFound
	}

	tasks[t.ID] = t
	return s.save(tasks)
}

// Delete deletes a task by its ID
func (s *JSONStorage) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.load()
	if err != nil {
		return err
	}

	if _, exists := tasks[id]; !exists {
		return task.ErrTaskNotFound
	}

	delete(tasks, id)
	return s.save(tasks)
}

// load reads tasks from the JSON file
func (s *JSONStorage) load() (map[string]*task.Task, error) {
	data, err := os.ReadFile(s.filepath)
	if err != nil {
		return nil, err
	}

	var tasks map[string]*task.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}

	if tasks == nil {
		tasks = make(map[string]*task.Task)
	}

	return tasks, nil
}

// save writes tasks to the JSON file
func (s *JSONStorage) save(tasks map[string]*task.Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filepath, data, 0644)
}

// matchesFilter checks if a task matches the given filter
func (s *JSONStorage) matchesFilter(t *task.Task, filter task.Filter) bool {
	if filter.Status != nil && t.Status != *filter.Status {
		return false
	}

	if filter.Priority != nil && t.Priority != *filter.Priority {
		return false
	}

	if len(filter.Tags) > 0 {
		taskTags := make(map[string]bool)
		for _, tag := range t.Tags {
			taskTags[tag] = true
		}

		for _, filterTag := range filter.Tags {
			if !taskTags[filterTag] {
				return false
			}
		}
	}

	return true
}
