package cli

import (
	"fmt"
	"io"
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

// openIssueFile validates and opens an Issue as a regular file. On Unix,
// issueOpenNoFollow also closes the validation/open race for symlink paths.
func openIssueFile(vaultDir, id string, flags int) (*os.File, error) {
	path := issuePath(vaultDir, id)
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("checking issue %s: %w", id, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("issue %s is not a regular file", id)
	}
	f, err := os.OpenFile(path, flags|issueOpenNoFollow, 0)
	if err != nil {
		return nil, fmt.Errorf("opening issue %s: %w", id, err)
	}
	return f, nil
}

func readIssueData(vaultDir, id string) ([]byte, error) {
	f, err := openIssueFile(vaultDir, id, os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(f)
	closeErr := f.Close()
	if readErr != nil {
		return nil, fmt.Errorf("reading issue %s: %w", id, readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("closing issue %s: %w", id, closeErr)
	}
	return data, nil
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
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		id := strings.TrimSuffix(name, ".md")
		if entry.Type()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("issue %s is a symbolic link", id)
		}
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("checking issue %s: %w", id, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("issue %s is not a regular file", id)
		}
		data, err := readIssueData(vaultDir, id)
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
