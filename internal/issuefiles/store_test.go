package issuefiles_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issuefiles"
)

// TestOpenIsPure: Open must not touch the filesystem (it never loads
// mt.yaml either), so it succeeds on a directory that does not exist;
// the failure shows up only when the store reads.
func TestOpenIsPure(t *testing.T) {
	store := issuefiles.Open(filepath.Join(t.TempDir(), "nope"))
	if _, err := store.List(); err == nil {
		t.Error("List() on a missing Vault = nil error, want failure")
	}
}

func TestListReadsIssuesInNameOrder(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-bbb", validIssue("second"))
	writeIssue(t, dir, "pkm-aaa", validIssue("first"))

	items, err := issuefiles.Open(dir).List()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := ids(items), []string{"pkm-aaa", "pkm-bbb"}; !reflect.DeepEqual(got, want) {
		t.Errorf("List() IDs = %v, want %v", got, want)
	}
	if got, want := items[0].Issue.Frontmatter.Title, "first"; got != want {
		t.Errorf("items[0] title = %q, want %q", got, want)
	}
}

func TestListSkipsSymlinksDirectoriesAndOtherExtensions(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-real", validIssue("real"))
	link(t, issuePath(dir, "pkm-real"), issuePath(dir, "pkm-link"))
	if err := os.Mkdir(issuePath(dir, "pkm-dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "issues", "notes.txt"), []byte("ignored"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A symlink with another extension is ignored twice over.
	link(t, issuePath(dir, "pkm-real"), filepath.Join(dir, "issues", "pkm-alias.yaml"))

	items, err := issuefiles.Open(dir).List()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := ids(items), []string{"pkm-real"}; !reflect.DeepEqual(got, want) {
		t.Errorf("List() IDs = %v, want %v (symlinks, directories and other extensions skipped)", got, want)
	}
}

func TestListFilesCarriesTheRawBytes(t *testing.T) {
	dir := newVault(t)
	content := validIssue("with bytes")
	writeIssue(t, dir, "pkm-aaa", content)

	files, err := issuefiles.Open(dir).ListFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("ListFiles() = %d files, want 1", len(files))
	}
	if files[0].ID != "pkm-aaa" {
		t.Errorf("file ID = %q, want %q", files[0].ID, "pkm-aaa")
	}
	if got := string(files[0].Data); got != content {
		t.Errorf("file bytes = %q, want the exact on-disk content %q", got, content)
	}
	if got, want := files[0].Issue.Frontmatter.Title, "with bytes"; got != want {
		t.Errorf("file title = %q, want %q", got, want)
	}
}

func TestListFilesSkipsSymlinks(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-real", validIssue("real"))
	link(t, issuePath(dir, "pkm-real"), issuePath(dir, "pkm-link"))

	files, err := issuefiles.Open(dir).ListFiles()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(files), 1; got != want {
		t.Fatalf("ListFiles() = %d files, want %d (symlink skipped)", got, want)
	}
	if files[0].ID != "pkm-real" {
		t.Errorf("file ID = %q, want %q", files[0].ID, "pkm-real")
	}
}

func TestListReportsMalformedFrontmatter(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-bad", "---\ntitle: [unclosed\n---\n")

	_, err := issuefiles.Open(dir).List()
	if err == nil {
		t.Fatal("List() on malformed frontmatter = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "parsing issue pkm-bad") {
		t.Errorf("List() error = %v, want it to name pkm-bad and the parse step", err)
	}
}

func TestListFailsWhenTheIssuesDirectoryIsMissing(t *testing.T) {
	dir := t.TempDir() // no issues/ at all
	_, err := issuefiles.Open(dir).List()
	if err == nil {
		t.Fatal("List() without issues/ = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "reading issues directory") {
		t.Errorf("List() error = %v, want it to name the issues directory", err)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("List() error = %v, want it to wrap os.ErrNotExist", err)
	}
}

// unreadable chmods path so reading it fails. Root bypasses permission
// bits entirely, so such tests skip there — the equivalent always-on
// cases (issues/ as a file, a directory named *.md) carry the branch.
func unreadable(t *testing.T, path string) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("running as root: permission bits do not deny access")
	}
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatal(err)
	}
}

func TestListFilesReportsAnUnreadableIssueFile(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-secret", validIssue("secret"))
	unreadable(t, issuePath(dir, "pkm-secret"))

	_, err := issuefiles.Open(dir).ListFiles()
	if err == nil {
		t.Fatal("ListFiles() with an unreadable file = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "reading issue pkm-secret: opening issue pkm-secret") {
		t.Errorf("ListFiles() error = %v, want the read failure to name the file", err)
	}
}

// TestListFailsWhenIssuesIsNotADirectory is the root-proof stand-in for
// an unreadable issues/: a regular file named "issues" makes ReadDir
// fail the same way a permission denial does.
func TestListFailsWhenIssuesIsNotADirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "issues"), []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := issuefiles.Open(dir).List()
	if err == nil {
		t.Fatal("List() with issues as a file = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "reading issues directory") {
		t.Errorf("List() error = %v, want it to name the issues directory", err)
	}
}

func TestListReportsAnUnreadableIssuesDirectory(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-aaa", validIssue("a"))
	issues := filepath.Join(dir, "issues")
	unreadable(t, issues)
	// Restore before t.TempDir's cleanup, which has to read the directory
	// to remove it.
	defer func() {
		if err := os.Chmod(issues, 0o755); err != nil {
			t.Errorf("restoring issues/ permissions: %v", err)
		}
	}()

	_, err := issuefiles.Open(dir).List()
	if err == nil {
		t.Fatal("List() with an unreadable issues/ = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "reading issues directory") {
		t.Errorf("List() error = %v, want it to name the issues directory", err)
	}
	if !errors.Is(err, os.ErrPermission) {
		t.Errorf("List() error = %v, want it to wrap os.ErrPermission", err)
	}
}

func TestIDsReturnsOnlyRegularIssueFiles(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-aaa", validIssue("regular"))
	writeIssue(t, dir, "pkm-ccc", validIssue("other regular"))
	link(t, issuePath(dir, "pkm-aaa"), issuePath(dir, "pkm-bbb"))
	if err := os.Mkdir(issuePath(dir, "pkm-dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "issues", "pkm-note.txt"), []byte("ignored"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := issuefiles.Open(dir).IDs()
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"pkm-aaa", "pkm-ccc"}; !reflect.DeepEqual(got, want) {
		t.Errorf("IDs() = %v, want %v", got, want)
	}
}

func TestIDsIsEmptyForAnEmptyVault(t *testing.T) {
	got, err := issuefiles.Open(newVault(t)).IDs()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("IDs() = %v, want none", got)
	}
}

func TestIDsFailsWithoutTheIssuesDirectory(t *testing.T) {
	_, err := issuefiles.Open(t.TempDir()).IDs()
	if err == nil {
		t.Fatal("IDs() without issues/ = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "reading issues directory") {
		t.Errorf("IDs() error = %v, want it to name the issues directory", err)
	}
}
