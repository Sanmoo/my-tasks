package issuefiles_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issue"
	"github.com/Sanmoo/my-tasks2/internal/issuefiles"
)

func TestMutatePersistsTheMutation(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-aaa", validIssue("before"))

	store := issuefiles.Open(dir)
	item, err := store.Mutate("pkm-aaa", func(i issue.Issue) (issue.Issue, error) {
		i.Frontmatter.Title = "after"
		return i, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := item.Issue.Frontmatter.Title, "after"; got != want {
		t.Errorf("Mutate() title = %q, want %q", got, want)
	}
	if got := readIssueFile(t, dir, "pkm-aaa"); !strings.Contains(got, "title: after") {
		t.Errorf("file after Mutate() = %q, want the mutated title on disk", got)
	}
}

// TestMutateWritesNothingWhenTheMutationFails pins the atomicity
// contract: the file on disk is the old one or the fully mutated one,
// never half of either.
func TestMutateWritesNothingWhenTheMutationFails(t *testing.T) {
	dir := newVault(t)
	before := validIssue("before")
	writeIssue(t, dir, "pkm-aaa", before)
	boom := errors.New("cannot allocate a comment anchor")

	called := false
	_, err := issuefiles.Open(dir).Mutate("pkm-aaa", func(i issue.Issue) (issue.Issue, error) {
		called = true
		i.Frontmatter.Title = "must not land"
		return i, boom
	})
	if !called {
		t.Fatal("the mutation function was not called")
	}
	if !errors.Is(err, boom) {
		t.Fatalf("Mutate(failing mutation) error = %v, want the mutation's own error", err)
	}
	if got := readIssueFile(t, dir, "pkm-aaa"); got != before {
		t.Errorf("file after a failed Mutate() = %q, want it byte-for-byte untouched", got)
	}
}

func TestMutateRefusesSymlink(t *testing.T) {
	dir := newVault(t)
	outside := filepath.Join(t.TempDir(), "target.md")
	outsideBefore := validIssue("outside")
	if err := os.WriteFile(outside, []byte(outsideBefore), 0o644); err != nil {
		t.Fatal(err)
	}
	link(t, outside, issuePath(dir, "pkm-link"))

	called := false
	_, err := issuefiles.Open(dir).Mutate("pkm-link", func(i issue.Issue) (issue.Issue, error) {
		called = true
		return i, nil
	})
	var notRegular *issuefiles.NotRegularError
	if !errors.As(err, &notRegular) {
		t.Fatalf("Mutate(symlink) error = %v, want NotRegularError", err)
	}
	if called {
		t.Error("the mutation function ran on a symlink")
	}
	data, readErr := os.ReadFile(outside)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != outsideBefore {
		t.Error("Mutate() wrote through the symlink, outside the Vault")
	}
}

func TestMutateMissingIssueWrapsNotExist(t *testing.T) {
	_, err := issuefiles.Open(newVault(t)).Mutate("pkm-nope", func(i issue.Issue) (issue.Issue, error) {
		return i, nil
	})
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Mutate(missing) error = %v, want it to wrap os.ErrNotExist", err)
	}
}

// TestMutateRefusesAFileSwappedForASymlinkAfterTheRead pins the race
// guard: the mutation function is the only place a test can swap the
// file, and the write must still refuse the link instead of following
// it outside the Vault.
func TestMutateRefusesAFileSwappedForASymlinkAfterTheRead(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-aaa", validIssue("before"))
	outside := filepath.Join(t.TempDir(), "target.md")
	outsideBefore := validIssue("outside")
	if err := os.WriteFile(outside, []byte(outsideBefore), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := issuefiles.Open(dir).Mutate("pkm-aaa", func(i issue.Issue) (issue.Issue, error) {
		if err := os.Remove(issuePath(dir, "pkm-aaa")); err != nil {
			t.Fatal(err)
		}
		link(t, outside, issuePath(dir, "pkm-aaa"))
		return i, nil
	})
	var notRegular *issuefiles.NotRegularError
	if !errors.As(err, &notRegular) {
		t.Fatalf("Mutate(swapped symlink) error = %v, want NotRegularError", err)
	}
	data, readErr := os.ReadFile(outside)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != outsideBefore {
		t.Error("Mutate() wrote through a symlink swapped in after the read")
	}
}

func TestMutateReportsAnUnwritableIssueFile(t *testing.T) {
	dir := newVault(t)
	before := validIssue("locked")
	writeIssue(t, dir, "pkm-locked", before)
	unreadable(t, issuePath(dir, "pkm-locked"))
	// Restore read permission but keep the file unwritable.
	if os.Geteuid() != 0 {
		if err := os.Chmod(issuePath(dir, "pkm-locked"), 0o444); err != nil {
			t.Fatal(err)
		}
	}

	_, err := issuefiles.Open(dir).Mutate("pkm-locked", func(i issue.Issue) (issue.Issue, error) {
		i.Frontmatter.Title = "must not land"
		return i, nil
	})
	if err == nil {
		t.Fatal("Mutate() on an unwritable file = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "writing issue pkm-locked") {
		t.Errorf("Mutate() error = %v, want it to name the write step", err)
	}
	got, readErr := os.ReadFile(issuePath(dir, "pkm-locked"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != before {
		t.Error("Mutate() changed a file it could not write")
	}
}

func TestEditHandsTheAbsoluteValidatedPathToTheCallback(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-aaa", validIssue("edit me"))

	var gotPath string
	err := issuefiles.Open(dir).Edit("pkm-aaa", func(path string) error {
		gotPath = path
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.Abs(issuePath(dir, "pkm-aaa"))
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != want {
		t.Errorf("Edit() path = %q, want the absolute path %q", gotPath, want)
	}
}

func TestEditRefusesSymlinkBeforeTheCallback(t *testing.T) {
	dir := newVault(t)
	outside := filepath.Join(t.TempDir(), "target.md")
	if err := os.WriteFile(outside, []byte(validIssue("outside")), 0o644); err != nil {
		t.Fatal(err)
	}
	link(t, outside, issuePath(dir, "pkm-link"))

	called := false
	err := issuefiles.Open(dir).Edit("pkm-link", func(string) error {
		called = true
		return nil
	})
	var notRegular *issuefiles.NotRegularError
	if !errors.As(err, &notRegular) {
		t.Fatalf("Edit(symlink) error = %v, want NotRegularError", err)
	}
	if called {
		t.Error("the callback received the path of a symlink")
	}
}

func TestEditRejectsInvalidIDsBeforeTheCallback(t *testing.T) {
	called := false
	err := issuefiles.Open(t.TempDir()).Edit("a/b", func(string) error {
		called = true
		return nil
	})
	var invalid *issuefiles.InvalidIDError
	if !errors.As(err, &invalid) {
		t.Fatalf("Edit(%q) error = %v, want InvalidIDError", "a/b", err)
	}
	if called {
		t.Error("the callback received the path of an invalid ID")
	}
}

func TestEditMissingIssueWrapsNotExist(t *testing.T) {
	err := issuefiles.Open(newVault(t)).Edit("pkm-nope", func(string) error { return nil })
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Edit(missing) error = %v, want it to wrap os.ErrNotExist", err)
	}
}

func TestEditPropagatesTheCallbackError(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-aaa", validIssue("edit me"))
	boom := errors.New("editor failed")

	err := issuefiles.Open(dir).Edit("pkm-aaa", func(string) error { return boom })
	if !errors.Is(err, boom) {
		t.Errorf("Edit() error = %v, want the callback's own error", err)
	}
}

func TestCreateWritesANewIssue(t *testing.T) {
	dir := newVault(t)

	id, err := issuefiles.Open(dir).Create("pkm", newIssue("fresh"), bytes.NewReader([]byte{0, 0, 0, 0}))
	if err != nil {
		t.Fatal(err)
	}
	if id != "pkm-0000" {
		t.Errorf("Create() ID = %q, want %q", id, "pkm-0000")
	}
	content := readIssueFile(t, dir, id)
	if !strings.Contains(content, "title: fresh") {
		t.Errorf("created file = %q, want the rendered title", content)
	}
	info, err := os.Stat(issuePath(dir, id))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o644); got != want {
		t.Errorf("created file mode = %v, want %v", got, want)
	}
}

func TestCreateRetriesOnACollidingIDSuffix(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-0000", validIssue("already here"))

	// A deterministic rng: the first draw is "0000" (taken),
	// the retry draws "1111" (free), so the allocation can only
	// succeed by retrying.
	rng := bytes.NewReader([]byte{0, 0, 0, 0, 1, 1, 1, 1})
	id, err := issuefiles.Open(dir).Create("pkm", newIssue("fresh"), rng)
	if err != nil {
		t.Fatal(err)
	}
	if id != "pkm-1111" {
		t.Errorf("Create() ID = %q, want the retried %q", id, "pkm-1111")
	}
}

// TestCreateNeverReusesASymlinkName pins both halves of the rule: a
// name occupied by a symlink is taken (so the allocation retries), and
// the new file is never written through the link.
func TestCreateNeverReusesASymlinkName(t *testing.T) {
	dir := newVault(t)
	outside := filepath.Join(t.TempDir(), "target.md")
	outsideBefore := validIssue("outside")
	if err := os.WriteFile(outside, []byte(outsideBefore), 0o644); err != nil {
		t.Fatal(err)
	}
	link(t, outside, issuePath(dir, "pkm-0000"))

	rng := bytes.NewReader([]byte{0, 0, 0, 0, 1, 1, 1, 1})
	id, err := issuefiles.Open(dir).Create("pkm", newIssue("fresh"), rng)
	if err != nil {
		t.Fatal(err)
	}
	if id != "pkm-1111" {
		t.Errorf("Create() ID = %q, want %q (the symlink's name is taken)", id, "pkm-1111")
	}
	data, readErr := os.ReadFile(outside)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != outsideBefore {
		t.Error("Create() wrote through the symlink, outside the Vault")
	}
}

func TestCreateRejectsAnIDThatWouldEscapeTheDirectory(t *testing.T) {
	_, err := issuefiles.Open(newVault(t)).Create("a/b", newIssue("fresh"), bytes.NewReader([]byte{0, 0, 0, 0}))
	var invalid *issuefiles.InvalidIDError
	if !errors.As(err, &invalid) {
		t.Fatalf("Create(bad prefix) error = %v, want InvalidIDError", err)
	}
}

// TestCreateIgnoresDirectoriesWhenCollectingTakenNames: a directory
// named *.md is not an Issue, so it does not occupy a name (the CLI's
// allocation has always skipped directories).
func TestCreateIgnoresDirectoriesWhenCollectingTakenNames(t *testing.T) {
	dir := newVault(t)
	if err := os.Mkdir(issuePath(dir, "pkm-dir"), 0o755); err != nil {
		t.Fatal(err)
	}

	id, err := issuefiles.Open(dir).Create("pkm", newIssue("fresh"), bytes.NewReader([]byte{0, 0, 0, 0}))
	if err != nil {
		t.Fatal(err)
	}
	if id != "pkm-0000" {
		t.Errorf("Create() ID = %q, want %q", id, "pkm-0000")
	}
}

// planterReader draws the first suffix and, as it does, plants a symlink
// at the name that suffix encodes — the race between Create's scan and
// its open. O_CREATE|O_EXCL must refuse the name instead of following
// the link.
type plantingReader struct {
	target string
	path   string
}

func (r *plantingReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	if err := os.Symlink(r.target, r.path); err != nil {
		return 0, err
	}
	return len(p), nil
}

func TestCreateRefusesANameTakenBetweenTheScanAndTheOpen(t *testing.T) {
	dir := newVault(t)
	outside := filepath.Join(t.TempDir(), "target.md")
	outsideBefore := validIssue("outside")
	if err := os.WriteFile(outside, []byte(outsideBefore), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := issuefiles.Open(dir).Create("pkm", newIssue("fresh"), &plantingReader{target: outside, path: issuePath(dir, "pkm-0000")})
	if err == nil {
		t.Fatal("Create() with a name taken during allocation = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "writing issue pkm-0000") {
		t.Errorf("Create() error = %v, want the occupied name to fail the write", err)
	}
	data, readErr := os.ReadFile(outside)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(data) != outsideBefore {
		t.Error("Create() wrote through a symlink planted during allocation")
	}
}

func TestCreateFailsWhenNoIDIsFree(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-0000", validIssue("taken"))

	_, err := issuefiles.Open(dir).Create("pkm", newIssue("fresh"), zeroReader{})
	if err == nil {
		t.Fatal("Create() with every suffix taken = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "could not allocate") {
		t.Errorf("Create() error = %v, want the allocation exhaustion error", err)
	}
}

func TestCreateFailsWithoutTheIssuesDirectory(t *testing.T) {
	_, err := issuefiles.Open(t.TempDir()).Create("pkm", newIssue("fresh"), bytes.NewReader([]byte{0, 0, 0, 0}))
	if err == nil {
		t.Fatal("Create() without issues/ = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "reading issues directory") {
		t.Errorf("Create() error = %v, want it to name the issues directory", err)
	}
}

// zeroReader yields zero bytes forever: the rng that always draws the
// same suffix.
type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
