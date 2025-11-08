package storage

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Sanmoo/my-tasks/internal/task"
)

type subbulletLine struct {
	filePath    string
	line        string
	projectName string
	lineNumber  int
}

type MarkdownStorage struct {
	projectAliases map[string]string
	filepaths      []string
}

// NewMarkdownStorage creates a new markdown storage
func NewMarkdownStorage(filePaths []string, projectAliases map[string]string) (*MarkdownStorage, error) {
	storage := &MarkdownStorage{
		filepaths:      filePaths,
		projectAliases: projectAliases,
	}

	return storage, nil
}

func (s *MarkdownStorage) GetProject(ctx context.Context, nameOrAlias string) (*task.Project, error) {
	name := nameOrAlias

	for k, v := range s.projectAliases {
		if k == nameOrAlias {
			name = v
			break
		}
	}

	project, err := s.load(name)
	if err != nil {
		return nil, err
	}

	return project, nil
}

// load reads tasks from the markdown file
func (s *MarkdownStorage) load(projectName string) (*task.Project, error) {
	for _, fp := range s.filepaths {
		file, err := os.Open(fp)
		if err != nil {
			return nil, err
		}

		var currentProject *task.Project
		var currentPhase *task.Phase
		var currentTask *task.Task

		scanner := bufio.NewScanner(file)
		currentLineNumber := 0
		for scanner.Scan() {
			currentLineNumber++
			line := scanner.Text()

			// empty line
			if strings.TrimSpace(line) == "" {
				continue
			}

			// Project (# header)
			if project, found := strings.CutPrefix(line, "# "); found {
				if currentProject != nil && currentProject.GetName() == projectName {
					return currentProject, nil
				}
				project = strings.TrimSpace(project)
				currentProject, err = task.NewProject(project)
				if err != nil {
					return nil, err
				}
				currentPhase = nil
				currentTask = nil
				continue
			}

			if currentProject != nil && currentProject.GetName() != projectName {
				// we are not looking for data of this project, so
				continue
			}

			// Phase (## header)
			if phase, found := strings.CutPrefix(line, "## "); found {
				currentPhase, err = task.NewPhase(phase)
				if err != nil {
					return nil, err
				}
				if currentProject == nil {
					return nil, fmt.Errorf("error found in file %s: found a invalid phase declaration at line %d (named %s). all phases should be declared under a project", fp, currentLineNumber, phase)
				}
				currentProject.AddPhase(currentPhase)
				currentTask = nil
				continue
			}

			// Task (bullet point starting with '*')
			if title, found := strings.CutPrefix(line, "* "); found {
				currentTask, err = task.NewTask(title)
				if err != nil {
					return nil, err
				}
				if currentPhase == nil {
					return nil, fmt.Errorf("error found in file %s: found a invalid task declaration at line %d (title %s). all tasks should be declared under a phase", fp, currentLineNumber, title)
				}
				currentPhase.AddTask(currentTask)

				continue
			}

			// Task item (sub bullet): directive or comment
			if subLine, found := strings.CutPrefix(line, "  * "); found {
				if currentTask == nil {
					return nil, fmt.Errorf("error found in file %s: found a invalid task item declaration at line %d (line \"%s\"). all task items should be declared under a task", fp, currentLineNumber, subLine)
				}
				err = s.parseSubBullet(currentTask, subbulletLine{
					line:        subLine,
					filePath:    fp,
					lineNumber:  currentLineNumber,
					projectName: currentProject.GetName(),
				})
				if err != nil {
					return nil, err
				}
				continue
			}
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}

		if currentProject != nil && currentProject.GetName() == projectName {
			return currentProject, nil
		}

		file.Close()
	}

	return nil, fmt.Errorf("no project named '%s' could be found in the configured file paths", projectName)
}

// parseSubBullet parses a sub-bullet line and updates the task
func (s *MarkdownStorage) parseSubBullet(t *task.Task, line subbulletLine) error {
	// @remind directive
	if strings.HasPrefix(line.line, "@remind") {
		dateStr := extractDateFromParenthesis(line.line)
		date, err := parseDate(dateStr)
		if err != nil {
			return fmt.Errorf(
				"error found in file %s, project %s: remider declared at line %d has not a parseable date, please check",
				line.filePath,
				line.projectName,
				line.lineNumber,
			)
		}

		t.AddReminder(&task.Reminder{
			Time:         date,
			Acknowledged: false,
		})

		return nil
	}

	// @reminded directive
	if strings.HasPrefix(line.line, "@reminded") {
		dateStr := extractDateFromParenthesis(line.line)
		date, err := parseDate(dateStr)
		if err != nil {
			return fmt.Errorf(
				"error found in file %s, project %s: remider declared at line %d has not a parseable date, please check",
				line.filePath,
				line.projectName,
				line.lineNumber,
			)
		}

		t.AddReminder(&task.Reminder{
			Time:         date,
			Acknowledged: true,
		})
	}

	// @tags directive
	if tagsStr, found := strings.CutPrefix(line.line, "@tags "); found {
		tags := strings.Fields(tagsStr)
		for _, tag := range tags {
			err := t.AddTag(tag)
			if err != nil {
				return err
			}
		}
	}

	// @due directive
	if strings.HasPrefix(line.line, "@due ") {
		dateStr := extractDateFromParenthesis(line.line)
		date, err := parseDate(dateStr)
		if err != nil {
			return fmt.Errorf(
				"error found in file %s, project %s: due directive declared at line %d has not a parseable date, please check",
				line.filePath,
				line.projectName,
				line.lineNumber,
			)
		}
		if err := t.SetDueDate(&date); err != nil {
			return err
		}
	}

	// Regular comment
	if err := t.AddComment(line.line); err != nil {
		return err
	}

	return nil
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
