// Package storage exposes ...
package storage

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Sanmoo/my-tasks/internal/task"
	"github.com/google/uuid"
)

// MarkdownStorage implements task.Repository using markdown files
type MarkdownStorage struct {
	filepath string
	mu       sync.RWMutex
}

// NewMarkdownStorage creates a new markdown storage
func NewMarkdownStorage(filePath string) (*MarkdownStorage, error) {
	// Create parent directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	storage := &MarkdownStorage{
		filepath: filePath,
	}

	// Initialize file if it doesn't exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := storage.save([]*task.Task{}); err != nil {
			return nil, err
		}
	}

	return storage, nil
}

// Create creates a new task
func (s *MarkdownStorage) Create(ctx context.Context, t *task.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.load()
	if err != nil {
		return err
	}

	// Generate ID if not set
	if t.ID == "" {
		t.ID = uuid.New().String()
	}

	tasks = append(tasks, t)
	return s.save(tasks)
}

// GetByID retrieves a task by its ID
func (s *MarkdownStorage) GetByID(ctx context.Context, id string) (*task.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks, err := s.load()
	if err != nil {
		return nil, err
	}

	for _, t := range tasks {
		if t.ID == id || strings.HasPrefix(t.ID, id) {
			return t, nil
		}
	}

	return nil, task.ErrTaskNotFound
}

// List retrieves all tasks with optional filtering
func (s *MarkdownStorage) List(ctx context.Context, filter task.Filter) ([]*task.Task, error) {
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
func (s *MarkdownStorage) Update(ctx context.Context, t *task.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.load()
	if err != nil {
		return err
	}

	found := false
	for i, existing := range tasks {
		if existing.ID == t.ID {
			tasks[i] = t
			found = true
			break
		}
	}

	if !found {
		return task.ErrTaskNotFound
	}

	return s.save(tasks)
}

// Delete deletes a task by its ID
func (s *MarkdownStorage) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.load()
	if err != nil {
		return err
	}

	found := false
	newTasks := make([]*task.Task, 0, len(tasks))
	for _, t := range tasks {
		if t.ID != id && !strings.HasPrefix(t.ID, id) {
			newTasks = append(newTasks, t)
		} else {
			found = true
		}
	}

	if !found {
		return task.ErrTaskNotFound
	}

	return s.save(newTasks)
}

// load reads tasks from the markdown file
func (s *MarkdownStorage) load() ([]*task.Task, error) {
	file, err := os.Open(s.filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tasks []*task.Task
	var currentProject string
	var currentPhase string
	var currentTask *task.Task

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Project (# header)
		if project, found := strings.CutPrefix(line, "# "); found {
			currentProject = project
			currentPhase = ""
			currentTask = nil
			continue
		}

		// Phase (## header)
		if phase, found := strings.CutPrefix(line, "## "); found {
			currentPhase = phase
			currentTask = nil
			continue
		}

		// Task (bullet point starting with *)
		trimmedLine := strings.TrimSpace(line)
		if title, found := strings.CutPrefix(trimmedLine, "* "); found && !strings.HasPrefix(trimmedLine, "  *") {
			// Save previous task if any
			if currentTask != nil {
				tasks = append(tasks, currentTask)
			}

			currentTask = &task.Task{
				ID:        uuid.New().String(),
				Title:     title,
				Project:   currentProject,
				Phase:     currentPhase,
				Status:    task.StatusPending,
				CreatedAt: time.Now(),
				Priority:  3,
			}
			continue
		}

		// Sub-bullet (directive or comment)
		if currentTask != nil {
			if subLine, found := strings.CutPrefix(trimmedLine, "* "); found {
				s.parseSubBullet(currentTask, subLine)
			}
		}
	}

	// Add last task
	if currentTask != nil {
		tasks = append(tasks, currentTask)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// parseSubBullet parses a sub-bullet line and updates the task
func (s *MarkdownStorage) parseSubBullet(t *task.Task, line string) {
	// @remind or @lembrar directive
	if strings.HasPrefix(line, "@remind ") || strings.HasPrefix(line, "@lembrar ") {
		dateStr := extractDateFromParenthesis(line)
		if date, err := parseDate(dateStr); err == nil {
			t.Reminders = append(t.Reminders, task.Reminder{
				Time:         date,
				Acknowledged: false,
			})
		}
		return
	}

	// @reminded directive
	if strings.HasPrefix(line, "@reminded ") {
		dateStr := extractDateFromParenthesis(line)
		if date, err := parseDate(dateStr); err == nil {
			t.Reminders = append(t.Reminders, task.Reminder{
				Time:         date,
				Acknowledged: true,
			})
		}
		return
	}

	// @tags directive
	if tagsStr, found := strings.CutPrefix(line, "@tags "); found {
		tags := strings.Fields(tagsStr)
		t.Tags = tags
		return
	}

	// @due directive
	if strings.HasPrefix(line, "@due ") {
		dateStr := extractDateFromParenthesis(line)
		if date, err := parseDate(dateStr); err == nil {
			t.DueDate = &date
		}
		return
	}

	// Regular comment
	t.Comments = append(t.Comments, line)
}

// save writes tasks to the markdown file
func (s *MarkdownStorage) save(tasks []*task.Task) error {
	// Group tasks by project and phase
	type projectPhase struct {
		project string
		phase   string
	}
	grouped := make(map[projectPhase][]*task.Task)

	for _, t := range tasks {
		key := projectPhase{project: t.Project, phase: t.Phase}
		grouped[key] = append(grouped[key], t)
	}

	// Build markdown content
	var content strings.Builder

	// Group by project first
	projectMap := make(map[string]map[string][]*task.Task)
	for key, taskList := range grouped {
		if projectMap[key.project] == nil {
			projectMap[key.project] = make(map[string][]*task.Task)
		}
		projectMap[key.project][key.phase] = taskList
	}

	// Write projects
	for project, phases := range projectMap {
		content.WriteString("# ")
		content.WriteString(project)
		content.WriteString("\n\n")

		for phase, taskList := range phases {
			content.WriteString("## ")
			content.WriteString(phase)
			content.WriteString("\n\n")

			for _, t := range taskList {
				s.writeTask(&content, t)
			}
		}
	}

	return os.WriteFile(s.filepath, []byte(content.String()), 0o600)
}

// writeTask writes a single task to the content builder
func (s *MarkdownStorage) writeTask(content *strings.Builder, t *task.Task) {
	// Write task title
	content.WriteString("* ")
	content.WriteString(t.Title)
	content.WriteString("\n")

	// Write comments
	for _, comment := range t.Comments {
		content.WriteString("  * ")
		content.WriteString(comment)
		content.WriteString("\n")
	}

	// Write reminders
	for _, reminder := range t.Reminders {
		if reminder.Acknowledged {
			content.WriteString("  * @reminded ")
		} else {
			content.WriteString("  * @remind ")
		}
		content.WriteString("(")
		content.WriteString(formatDate(reminder.Time))
		content.WriteString(")")
		content.WriteString("\n")
	}

	// Write tags
	if len(t.Tags) > 0 {
		content.WriteString("  * @tags ")
		content.WriteString(strings.Join(t.Tags, " "))
		content.WriteString("\n")
	}

	// Write due date
	if t.DueDate != nil {
		content.WriteString("  * @due ")
		content.WriteString("(")
		content.WriteString(formatDate(*t.DueDate))
		content.WriteString(")")
		content.WriteString("\n")
	}
}

// matchesFilter checks if a task matches the given filter
func (s *MarkdownStorage) matchesFilter(t *task.Task, filter task.Filter) bool {
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

// extractDateFromParenthesis extracts date string from parenthesis
func extractDateFromParenthesis(line string) string {
	re := regexp.MustCompile(`\(([^)]+)\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// parseDate parses date from various formats
func parseDate(dateStr string) (time.Time, error) {
	// Try format: YY-MM-DD HH:MM:SS
	if t, err := time.Parse("06-01-02 15:04:05", dateStr); err == nil {
		return t, nil
	}

	// Try format: YY-MM-DD HH:MM
	if t, err := time.Parse("06-01-02 15:04", dateStr); err == nil {
		return t, nil
	}

	// Try format: YY-MM-DD
	if t, err := time.Parse("06-01-02", dateStr); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
}

// formatDate formats a time to the markdown format
func formatDate(t time.Time) string {
	if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
		return t.Format("06-01-02")
	}
	return t.Format("06-01-02 15:04:05")
}
