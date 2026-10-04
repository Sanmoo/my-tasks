package issuefiles

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// Mutate is the only write door of a Store: it reads the Issue, applies
// the mutation, renders the result and writes it back. A mutation that
// fails — a Comment anchor that cannot be allocated, say — writes
// nothing at all, so the file on disk is either the old one or the
// fully mutated one. No raw bytes are stashed anywhere: the caller
// receives the Item it would have read back.
func (s Store) Mutate(id string, mutate func(issue.Issue) (issue.Issue, error)) (issue.Item, error) {
	item, err := s.Read(id)
	if err != nil {
		return issue.Item{}, err
	}
	mutated, err := mutate(item.Issue)
	if err != nil {
		return issue.Item{}, err
	}
	data, err := issue.Render(mutated)
	if err != nil {
		return issue.Item{}, err
	}
	if err := s.writeBytes(id, data); err != nil {
		return issue.Item{}, err
	}
	return issue.Item{ID: id, Issue: mutated}, nil
}

// Edit validates the Issue (it exists, is a regular file and its ID is
// a single path component) BEFORE the path leaves the Store, then hands
// the absolute path to edit — the $EDITOR invocation of mt edit. A
// symlink is refused here, so the editor is never pointed outside the
// Vault.
func (s Store) Edit(id string, edit func(path string) error) error {
	if err := checkID(id); err != nil {
		return err
	}
	if err := s.checkRegular(id); err != nil {
		return err
	}
	path, err := filepath.Abs(s.path(id))
	if err != nil {
		return fmt.Errorf("resolving issue %s: %w", id, err)
	}
	return edit(path)
}

// Create allocates a fresh ID for prefix, renders i and writes it as a
// new file. Every *.md entry that is not a directory counts as
// occupied — symlinks included — so a new Issue never takes a name the
// Vault already uses and never writes through a link; O_CREATE|O_EXCL
// plus O_NOFOLLOW on the open is the belt to that suspenders. The
// caller supplies the randomness (rng) and the prefix (from the Vault
// config it already loaded).
func (s Store) Create(prefix string, i issue.Issue, rng io.Reader) (string, error) {
	taken, err := s.takenIDs()
	if err != nil {
		return "", err
	}
	id, err := issue.NextID(prefix, taken, rng)
	if err != nil {
		return "", err
	}
	if err := checkID(id); err != nil {
		return "", err
	}
	data, err := issue.Render(i)
	if err != nil {
		return "", err
	}
	if err := s.createBytes(id, data); err != nil {
		return "", err
	}
	return id, nil
}

// takenIDs maps every Issue ID that Create must not reuse: every *.md
// entry that is not a directory, symlinks included. This preserves the
// long-standing allocation semantics of the CLI — a name occupied by a
// link is never chosen — and it is what keeps Create away from a
// symlink.
func (s Store) takenIDs() (map[string]bool, error) {
	entries, err := s.readDir()
	if err != nil {
		return nil, err
	}
	taken := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if id, ok := issueID(entry); ok {
			taken[id] = true
		}
	}
	return taken, nil
}

// writeBytes writes data over an existing Issue file. The path is
// validated as a regular file and opened with O_NOFOLLOW, so neither a
// symlink read from the directory nor one swapped in between can
// redirect the write outside the Vault.
func (s Store) writeBytes(id string, data []byte) error {
	if err := s.checkRegular(id); err != nil {
		return fmt.Errorf("writing issue %s: %w", id, err)
	}
	f, err := os.OpenFile(s.path(id), os.O_WRONLY|os.O_TRUNC|issueOpenNoFollow, 0)
	if err != nil {
		return fmt.Errorf("writing issue %s: %w", id, fmt.Errorf("opening issue %s: %w", id, err))
	}
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr != nil {
		return fmt.Errorf("writing issue %s: %w", id, writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("closing issue %s: %w", id, closeErr)
	}
	return nil
}

// createBytes writes a brand-new Issue file. O_EXCL never truncates an
// existing file and never follows a symlink, so the worst case of a
// name slipped in after the scan is a loud failure, not a write outside
// the Vault.
func (s Store) createBytes(id string, data []byte) error {
	f, err := os.OpenFile(s.path(id), os.O_WRONLY|os.O_CREATE|os.O_EXCL|issueOpenNoFollow, 0o644)
	if err != nil {
		return fmt.Errorf("writing issue %s: %w", id, err)
	}
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr != nil {
		return fmt.Errorf("writing issue %s: %w", id, writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("closing issue %s: %w", id, closeErr)
	}
	return nil
}
