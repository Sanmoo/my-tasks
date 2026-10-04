// Package issuefiles_test holds the black-box tests of the Issue file
// store (Seam 2): the file policy of a Vault — what counts as an Issue,
// what is skipped, what is refused — against t.TempDir() vaults. Only
// the exported interface is touched.
package issuefiles_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// validIssue returns a well-formed Issue file body with the given title.
func validIssue(title string) string {
	return fmt.Sprintf("---\ntitle: %s\nstatus: open\nlabels: []\ncreated_at: 2026-01-01T10:00\n---\n\n## Description\n## Notes\n## Comments\n", title)
}

// newIssue returns a minimal well-formed Issue value for Create.
func newIssue(title string) issue.Issue {
	return issue.Issue{
		Frontmatter: issue.Frontmatter{
			Title:     title,
			Status:    "open",
			Labels:    []string{},
			CreatedAt: "2026-01-01T10:00",
		},
		Body: issue.DefaultBody,
	}
}

// newVault returns a Vault directory with an empty issues/ directory.
func newVault(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "issues"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// writeIssue writes issues/<id>.md with content.
func writeIssue(t *testing.T, dir, id, content string) {
	t.Helper()
	if err := os.WriteFile(issuePath(dir, id), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// readIssueFile reads issues/<id>.md.
func readIssueFile(t *testing.T, dir, id string) string {
	t.Helper()
	data, err := os.ReadFile(issuePath(dir, id))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// issuePath is the on-disk path of id inside a Vault directory.
func issuePath(dir, id string) string {
	return filepath.Join(dir, "issues", id+".md")
}

// link creates a symlink at path pointing to target.
func link(t *testing.T, target, path string) {
	t.Helper()
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
}

// ids extracts the IDs of a list of items.
func ids(items []issue.Item) []string {
	out := make([]string, len(items))
	for i, item := range items {
		out[i] = item.ID
	}
	return out
}
