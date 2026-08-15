package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// parsedIssueFile is the shared process-level view of one Issue file. The
// raw bytes let check validate schema details without making list/prioritize
// duplicate directory readers, while the parsed Issue serves their normal
// projections.
type parsedIssueFile struct {
	ID    string
	Data  []byte
	Issue issue.Issue
}

func readIssueFiles(vaultDir string) ([]parsedIssueFile, error) {
	dir := filepath.Join(vaultDir, "issues")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading issues directory: %w", err)
	}
	files := make([]parsedIssueFile, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}
		id := strings.TrimSuffix(name, ".md")
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("reading issue %s: %w", id, err)
		}
		i, err := issue.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("parsing issue %s: %w", id, err)
		}
		files = append(files, parsedIssueFile{ID: id, Data: data, Issue: i})
	}
	return files, nil
}
