package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

// issueIDs lists the existing Issue IDs of a vault: the issues/*.md
// file names with the suffix stripped, skipping directories. ReadDir
// sorts by name, so the list is deterministic. It is the enumeration
// behind ID allocation (newIssueID) — note the difference from
// issuefiles.IDs, which serves completion with regular files only:
// allocation must also treat a symlink's name as taken.
func issueIDs(vaultDir string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(vaultDir, "issues"))
	if err != nil {
		return nil, fmt.Errorf("reading issues directory: %w", err)
	}
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".md") {
			ids = append(ids, strings.TrimSuffix(name, ".md"))
		}
	}
	return ids, nil
}
