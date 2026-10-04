package issuefiles

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// IDs lists the IDs of the Vault's Issues: its regular *.md files,
// sorted by name. Every non-regular entry — a symlink, a *.md
// directory, a FIFO, a device — is omitted: completion must never offer
// a name the commands refuse. It is what shell completion uses.
func (s Store) IDs() ([]string, error) {
	entries, err := s.scanIssues()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.kind != kindRegular {
			continue
		}
		ids = append(ids, entry.id)
	}
	return ids, nil
}

// List returns the Vault's Issues in directory order (sorted by file
// name). It is the view path: list, ready, overdue, pick-next, place,
// prioritize and undefer all read the Vault through it.
//
// It shares the Store's single classification and read path with
// ListFiles, but carries no raw bytes: only mt check needs them.
//
// A *.md symlink and a *.md directory are skipped — a stray link never
// brings the listing down, an accidental directory never becomes a
// ghost Issue — while a non-regular entry that is neither fails loud.
func (s Store) List() ([]issue.Item, error) {
	entries, err := s.scanIssues()
	if err != nil {
		return nil, err
	}
	items := make([]issue.Item, 0, len(entries))
	for _, entry := range entries {
		item, _, ok, err := s.readEntry(entry)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

// ListFiles returns the Vault's Issues with the raw bytes of each file,
// in directory order and under the same policy as List. mt check is the
// single consumer of the bytes: frontmatter validation needs the
// original YAML, not the parsed round-trip.
func (s Store) ListFiles() ([]File, error) {
	entries, err := s.scanIssues()
	if err != nil {
		return nil, err
	}
	files := make([]File, 0, len(entries))
	for _, entry := range entries {
		item, data, ok, err := s.readEntry(entry)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		files = append(files, File{Item: item, Data: data})
	}
	return files, nil
}

// readEntry reads and parses one classified entry. It reports whether
// the entry is an Issue at all: a *.md symlink and a *.md directory are
// not — no error, just skipped, so a stray link never fails a listing
// and an accidental directory never becomes a ghost Issue — while a
// non-regular entry that is neither (a FIFO, a device) fails loud with
// NotRegularError. The kind comes from the directory scan and the open
// carries O_NOFOLLOW, so no entry is Lstat'ed twice for the walk and a
// symlink swapped in after the scan still cannot redirect the read.
func (s Store) readEntry(entry issueEntry) (issue.Item, []byte, bool, error) {
	switch entry.kind {
	case kindSymlink, kindDirectory:
		return issue.Item{}, nil, false, nil
	case kindOther:
		return issue.Item{}, nil, false, &NotRegularError{ID: entry.id}
	}
	data, err := s.readBytes(entry.id)
	if err != nil {
		return issue.Item{}, nil, false, fmt.Errorf("reading issue %s: %w", entry.id, err)
	}
	i, err := issue.Parse(data)
	if err != nil {
		return issue.Item{}, nil, false, fmt.Errorf("parsing issue %s: %w", entry.id, err)
	}
	return issue.Item{ID: entry.id, Issue: i}, data, true, nil
}

// readBytes reads an Issue file whose path the caller already validated
// as a regular file — a targeted read validates with checkRegular, a
// listing read relies on the scan's classification. The open carries
// O_NOFOLLOW, so a symlink swapped in between the validation and the
// open still cannot redirect the read outside the Vault.
func (s Store) readBytes(id string) ([]byte, error) {
	f, err := os.OpenFile(s.path(id), os.O_RDONLY|issueOpenNoFollow, 0)
	if err != nil {
		return nil, fmt.Errorf("opening issue %s: %w", id, err)
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

// readValidated reads the Issue file of a targeted operation: it
// validates first that the path is a regular file — Lstat never follows
// a symlink, so a link is refused with NotRegularError instead of read
// — and then reads it.
func (s Store) readValidated(id string) ([]byte, error) {
	if err := s.checkRegular(id); err != nil {
		return nil, err
	}
	return s.readBytes(id)
}

// Read returns the Issue with the given ID. The ID must be a single
// path component and the file a regular one: a symlink is refused with
// NotRegularError, never followed. Absence is a wrapped os.ErrNotExist
// so the command layer can keep rendering its own not-found message
// with the Vault context.
func (s Store) Read(id string) (issue.Item, error) {
	if err := checkID(id); err != nil {
		return issue.Item{}, err
	}
	data, err := s.readValidated(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return issue.Item{}, err
		}
		// The command layer has always reported read failures as
		// "reading issue <id>: ..."; the wrap keeps that shape so the
		// output of a Vault without symlinks does not change.
		return issue.Item{}, fmt.Errorf("reading issue %s: %w", id, err)
	}
	i, err := issue.Parse(data)
	if err != nil {
		return issue.Item{}, fmt.Errorf("parsing issue %s: %w", id, err)
	}
	return issue.Item{ID: id, Issue: i}, nil
}
