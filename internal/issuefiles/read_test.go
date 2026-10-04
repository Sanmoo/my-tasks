package issuefiles_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issuefiles"
)

func TestReadReturnsTheParsedIssue(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-aaa", validIssue("the title"))
	writeIssue(t, dir, "pkm-bbb", validIssue("other"))

	item, err := issuefiles.Open(dir).Read("pkm-aaa")
	if err != nil {
		t.Fatal(err)
	}
	if item.ID != "pkm-aaa" {
		t.Errorf("Read() ID = %q, want %q", item.ID, "pkm-aaa")
	}
	if got, want := item.Issue.Frontmatter.Title, "the title"; got != want {
		t.Errorf("Read() title = %q, want %q", got, want)
	}
}

func TestReadMissingIssueWrapsNotExist(t *testing.T) {
	dir := newVault(t)
	_, err := issuefiles.Open(dir).Read("pkm-nope")
	if err == nil {
		t.Fatal("Read(missing) = nil error, want failure")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Read(missing) error = %v, want it to wrap os.ErrNotExist", err)
	}
}

// TestReadRefusesSymlink points the link at a perfectly valid Issue: if
// Read followed it, the call would succeed. The refusal itself is the
// assertion that the target was never read.
func TestReadRefusesSymlink(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-real", validIssue("real"))
	link(t, issuePath(dir, "pkm-real"), issuePath(dir, "pkm-link"))

	_, err := issuefiles.Open(dir).Read("pkm-link")
	var notRegular *issuefiles.NotRegularError
	if !errors.As(err, &notRegular) {
		t.Fatalf("Read(symlink) error = %v, want NotRegularError", err)
	}
	if got, want := notRegular.ID, "pkm-link"; got != want {
		t.Errorf("NotRegularError.ID = %q, want %q", got, want)
	}
	if !strings.Contains(err.Error(), "not a regular file") {
		t.Errorf("Read(symlink) error = %v, want the not-a-regular-file refusal", err)
	}
}

func TestReadRefusesDanglingSymlink(t *testing.T) {
	dir := newVault(t)
	link(t, filepath.Join(t.TempDir(), "nowhere", "target.md"), issuePath(dir, "pkm-link"))

	_, err := issuefiles.Open(dir).Read("pkm-link")
	var notRegular *issuefiles.NotRegularError
	if !errors.As(err, &notRegular) {
		t.Fatalf("Read(dangling symlink) error = %v, want NotRegularError", err)
	}
	if errors.Is(err, os.ErrNotExist) {
		t.Error("Read(dangling symlink) reported absence; a symlink exists, it is just not an Issue")
	}
}

func TestReadRejectsIDsThatWouldEscapeTheDirectory(t *testing.T) {
	for _, id := range []string{"", "a/b", `a\b`, "../escape"} {
		t.Run(id, func(t *testing.T) {
			// No issues/ directory at all: an invalid ID must be refused
			// before any filesystem access, so the error is not absence.
			_, err := issuefiles.Open(t.TempDir()).Read(id)
			var invalid *issuefiles.InvalidIDError
			if !errors.As(err, &invalid) {
				t.Fatalf("Read(%q) error = %v, want InvalidIDError", id, err)
			}
			if errors.Is(err, os.ErrNotExist) {
				t.Errorf("Read(%q) touched the filesystem before rejecting the ID", id)
			}
			if !strings.Contains(err.Error(), fmt.Sprintf("invalid issue ID %q", id)) {
				t.Errorf("Read(%q) error = %q, want the invalid-ID reason", id, err)
			}
		})
	}
}

func TestReadReportsMalformedFrontmatter(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-bad", "---\ntitle: [unclosed\n---\n")

	_, err := issuefiles.Open(dir).Read("pkm-bad")
	if err == nil {
		t.Fatal("Read(malformed) = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "parsing issue pkm-bad") {
		t.Errorf("Read(malformed) error = %v, want it to name the parse step", err)
	}
}

func TestReadReportsAnUnreadableIssueFile(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-secret", validIssue("secret"))
	unreadable(t, issuePath(dir, "pkm-secret"))

	_, err := issuefiles.Open(dir).Read("pkm-secret")
	if err == nil {
		t.Fatal("Read(unreadable) = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "reading issue pkm-secret: opening issue pkm-secret") {
		t.Errorf("Read(unreadable) error = %v, want the read failure to name the file", err)
	}
}
