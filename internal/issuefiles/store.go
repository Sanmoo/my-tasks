// Package issuefiles is the single owner of the Issue files of a Vault:
// listing, reading, writing, editing and creating them. It is Seam 2 —
// pure logic, black-box tested against t.TempDir() vaults, with the
// coverage and mutation gates.
//
// The file policy is written here once, and its central rule is that a
// symlink inside issues/ is not an Issue:
//
//   - every listing (IDs, List, ListFiles) skips a *.md symlink instead
//     of failing the whole Vault;
//   - every targeted operation (Read, Mutate, Edit, Create) refuses one,
//     so no command reads, writes or edits outside the Vault;
//   - Create never allocates a name a symlink occupies.
//
// Everything else about the on-disk shape is preserved: a *.md entry
// that is a directory is skipped, a non-regular entry that is not a
// directory (a FIFO, a device) fails loud, an unreadable issues/
// directory produces a clear error, and Open never loads mt.yaml — the
// module decides nothing about what a Vault is, it receives the
// directory.
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
