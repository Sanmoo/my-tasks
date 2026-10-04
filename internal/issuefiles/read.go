package issuefiles

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// IDs lists the IDs of the Vault's Issues: the regular *.md files,
// sorted by name. Symlinks, directories and other non-regular entries
// are omitted — completion must never offer a name the commands
// refuse. It is what shell completion uses.
func (s Store) IDs() ([]string, error) {
	entries, err := s.readDir()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		id, ok := issueID(entry)
		if !ok {
			continue
		}
		// A symlink or directory is not an Issue: never offered.
		if entry.Type()&os.ModeSymlink != 0 || entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("checking issue %s: %w", id, err)
		}
		if !info.Mode().IsRegular() {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// List returns the Vault's Issues in directory order (sorted by file
// name). It is the view path: list, ready, overdue, pick-next, place,
// prioritize and undefer all read the Vault through it.
//
// A *.md symlink and a *.md directory are skipped — a stray link never
// brings the listing down, an accidental directory never becomes a
// ghost Issue — while a non-regular entry that is neither fails loud.
func (s Store) List() ([]issue.Item, error) {
	files, err := s.ListFiles()
	if err != nil {
		return nil, err
	}
	items := make([]issue.Item, 0, len(files))
	for _, file := range files {
		items = append(items, file.Item)
	}
	return items, nil
}

// ListFiles is List plus the raw bytes of every file. mt check is the
// single consumer: frontmatter validation needs the original YAML, not
// the parsed round-trip.
func (s Store) ListFiles() ([]File, error) {
	entries, err := s.readDir()
	if err != nil {
		return nil, err
	}
	files := make([]File, 0, len(entries))
	for _, entry := range entries {
		id, ok := issueID(entry)
		if !ok {
			continue
		}
		// A symlink is not an Issue: skip it without touching its target.
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		// A directory named *.md is not an Issue either.
		if entry.IsDir() {
			continue
		}
		if err := s.checkRegular(id); err != nil {
			return nil, err
		}
		data, err := s.readBytes(id)
		if err != nil {
			return nil, fmt.Errorf("reading issue %s: %w", id, err)
		}
		i, err := issue.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("parsing issue %s: %w", id, err)
		}
		files = append(files, File{Item: issue.Item{ID: id, Issue: i}, Data: data})
	}
	return files, nil
}

// readBytes reads an Issue file. The path is validated as a regular
// file first, then opened with O_NOFOLLOW, so neither a symlink parsed
// from the directory nor one swapped in between can redirect the read
// outside the Vault.
func (s Store) readBytes(id string) ([]byte, error) {
	if err := s.checkRegular(id); err != nil {
		return nil, err
	}
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

// Read returns the Issue with the given ID. The ID must be a single
// path component and the file a regular one: a symlink is refused with
// NotRegularError, never followed. Absence is a wrapped os.ErrNotExist
// so the command layer can keep rendering its own not-found message
// with the Vault context.
func (s Store) Read(id string) (issue.Item, error) {
	if err := checkID(id); err != nil {
		return issue.Item{}, err
	}
	data, err := s.readBytes(id)
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
