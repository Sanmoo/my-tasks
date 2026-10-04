//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package issuefiles_test

import (
	"bytes"
	"errors"
	"syscall"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issuefiles"
)

// mkfifo creates a named pipe: the canary non-regular, non-directory
// entry. A Vault with one must fail loud — silent skipping is only for
// symlinks and directories.
func mkfifo(t *testing.T, path string) {
	t.Helper()
	if err := syscall.Mkfifo(path, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestListFailsLoudOnANonRegularIssueFile(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-real", validIssue("real"))
	mkfifo(t, issuePath(dir, "pkm-fifo"))

	_, err := issuefiles.Open(dir).List()
	var notRegular *issuefiles.NotRegularError
	if !errors.As(err, &notRegular) {
		t.Fatalf("List() with a FIFO = %v, want NotRegularError", err)
	}
	if got, want := notRegular.ID, "pkm-fifo"; got != want {
		t.Errorf("NotRegularError.ID = %q, want %q", got, want)
	}
}

func TestReadRefusesANonRegularIssueFile(t *testing.T) {
	dir := newVault(t)
	mkfifo(t, issuePath(dir, "pkm-fifo"))

	_, err := issuefiles.Open(dir).Read("pkm-fifo")
	var notRegular *issuefiles.NotRegularError
	if !errors.As(err, &notRegular) {
		t.Fatalf("Read(FIFO) = %v, want NotRegularError", err)
	}
}

func TestIDsOmitsNonRegularEntries(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-real", validIssue("real"))
	mkfifo(t, issuePath(dir, "pkm-fifo"))

	got, err := issuefiles.Open(dir).IDs()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "pkm-real" {
		t.Errorf("IDs() = %v, want only the regular file", got)
	}
}

func TestCreateTreatsANonRegularEntryAsAnOccupiedName(t *testing.T) {
	dir := newVault(t)
	mkfifo(t, issuePath(dir, "pkm-0000"))

	rng := bytes.NewReader([]byte{0, 0, 0, 0, 1, 1, 1, 1})
	id, err := issuefiles.Open(dir).Create("pkm", newIssue("fresh"), rng)
	if err != nil {
		t.Fatal(err)
	}
	if id != "pkm-1111" {
		t.Errorf("Create() ID = %q, want %q (the FIFO's name is taken)", id, "pkm-1111")
	}
}
