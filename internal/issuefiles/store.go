// Package issuefiles is the single owner of the Issue files of a Vault:
// listing, reading, writing, editing and creating them. It is Seam 2 —
// pure logic, black-box tested against t.TempDir() vaults, with the
// coverage and mutation gates.
//
// The file policy is written here once, and its central rule is that a
// symlink inside issues/ is not an Issue. Symlinks and *.md directories
// are skipped everywhere — a stray link never fails a listing, an
// accidental directory never becomes a ghost Issue — so the paths
// differ only in who errors on a non-regular entry that is not a
// directory (a FIFO, a device):
//
//   - IDs — the completion path — omits every non-regular entry: it
//     must never offer a name the commands refuse;
//   - the sound listing and read paths (List, ListFiles, Read) fail
//     loud on such an entry, so a broken Vault never passes silently;
//   - every targeted operation (Read, Mutate, Edit) refuses a symlink,
//     and Create never allocates a name a symlink occupies, so no
//     command reads, writes, edits or creates outside the Vault.
//
// Everything else about the on-disk shape is preserved: an unreadable
// issues/ directory produces a clear error, and Open never loads
// mt.yaml — the module decides nothing about what a Vault is, it
// receives the directory.
//
// Errors carry the operation and the ID ("reading issue x: ...") so the
// command layer can render them as it always has; absence is always a
// wrapped os.ErrNotExist, an invalid ID and a non-regular file are the
// exported typed errors below, and the command layer keeps owning the
// exit code and the final message.
package issuefiles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// File is one Issue file: the Item line (ID + parsed Issue) plus the
// raw bytes of the file. Only mt check needs the bytes — frontmatter
// validation works on the original YAML — so only ListFiles carries
// them; no mutating call ever stashes stale bytes in a result.
type File struct {
	issue.Item
	Data []byte
}

// Store addresses the Issue files of one Vault. Open it with Open; its
// zero value is not useful.
type Store struct {
	dir string
}

// Open returns the Store of the Vault rooted at dir. It is pure: it
// touches no file and never reads mt.yaml, so the commands that run
// without the Vault config keep running without it. dir is the Vault
// directory; the Issues live in dir/issues.
func Open(dir string) Store {
	return Store{dir: dir}
}

// InvalidIDError reports an ID that is not a single path component and
// therefore names no Issue inside the Vault. It is the boundary
// invariant shared with the CLI's argument check — both use
// issue.ValidID — because an ID with a separator would make
// issues/<id>.md resolve outside the Vault.
type InvalidIDError struct {
	ID string
}

func (e *InvalidIDError) Error() string {
	return fmt.Sprintf("invalid issue ID %q", e.ID)
}

// NotRegularError reports that the path of an Issue is not a regular
// file: a symlink, a FIFO, a device or any other non-regular entry. It
// is the one refusal every command renders for such a path, so a
// symlink planted in issues/ can never redirect an operation outside
// the Vault.
type NotRegularError struct {
	ID string
}

func (e *NotRegularError) Error() string {
	return fmt.Sprintf("issue %s is not a regular file", e.ID)
}

// checkID applies the single-component invariant of issue IDs.
func checkID(id string) error {
	if !issue.ValidID(id) {
		return &InvalidIDError{ID: id}
	}
	return nil
}

// path is the on-disk path of id inside the Vault.
func (s Store) path(id string) string {
	return filepath.Join(s.dir, "issues", id+".md")
}

// readDir reads the Vault's issues directory. ReadDir sorts by name, so
// every listing built on it is deterministic.
func (s Store) readDir() ([]os.DirEntry, error) {
	entries, err := os.ReadDir(filepath.Join(s.dir, "issues"))
	if err != nil {
		return nil, fmt.Errorf("reading issues directory: %w", err)
	}
	return entries, nil
}

// issueID returns the ID an entry names when it is an Issue file: a
// *.md name with the suffix stripped. Every other extension is not an
// Issue and reports false.
func issueID(entry os.DirEntry) (string, bool) {
	name := entry.Name()
	if !strings.HasSuffix(name, ".md") {
		return "", false
	}
	return strings.TrimSuffix(name, ".md"), true
}

// entryKind is what a *.md entry of issues/ is on disk. It comes from
// the directory listing alone — os.DirEntry.Type resolves the kind
// without following a symlink — so one ReadDir pass classifies every
// entry and no path is Lstat'ed for the walk itself.
type entryKind uint8

const (
	// kindRegular is a regular file: the only kind that is an Issue.
	kindRegular entryKind = iota
	// kindSymlink is a symbolic link: not an Issue anywhere.
	kindSymlink
	// kindDirectory is a directory named *.md: not an Issue anywhere.
	kindDirectory
	// kindOther is any other non-regular entry (a FIFO, a device): the
	// sound paths fail loud on it.
	kindOther
)

// issueEntry is one *.md entry of the issues/ directory: the ID its
// file name names and its kind.
type issueEntry struct {
	id   string
	kind entryKind
}

// scanIssues reads the issues/ directory once and classifies every *.md
// entry; an entry with another extension is not an Issue and is
// omitted. It is the Store's only directory walk, and each caller then
// applies its own policy in terms of the classification: a name that is
// taken, an ID to list, a file to read.
func (s Store) scanIssues() ([]issueEntry, error) {
	entries, err := s.readDir()
	if err != nil {
		return nil, err
	}
	issues := make([]issueEntry, 0, len(entries))
	for _, entry := range entries {
		id, ok := issueID(entry)
		if !ok {
			continue
		}
		issues = append(issues, issueEntry{id: id, kind: kindOf(entry)})
	}
	return issues, nil
}

// kindOf classifies one directory entry by its kind, never by following
// a symlink.
func kindOf(entry os.DirEntry) entryKind {
	mode := entry.Type()
	if mode&os.ModeSymlink != 0 {
		return kindSymlink
	}
	if mode.IsDir() {
		return kindDirectory
	}
	if mode.IsRegular() {
		return kindRegular
	}
	return kindOther
}

// checkRegular validates that the path of id is a regular file without
// following symlinks: Lstat sees a symlink for what it is, and anything
// that is not a regular file is refused with NotRegularError.
func (s Store) checkRegular(id string) error {
	info, err := os.Lstat(s.path(id))
	if err != nil {
		return fmt.Errorf("checking issue %s: %w", id, err)
	}
	if !info.Mode().IsRegular() {
		return &NotRegularError{ID: id}
	}
	return nil
}
