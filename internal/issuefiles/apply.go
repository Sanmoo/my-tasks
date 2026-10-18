package issuefiles

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// recoveryDirName is the directory, inside the Vault and beside issues/,
// that keeps the originals of a rank operation until it is known to be
// complete. It is never listed: every listing reads issues/ only.
const recoveryDirName = ".mt-recovery"

// RankPlan is one planned Rank operation: the Rank changes of existing
// Issues and, optionally, the Issue the operation creates. The Changes
// are the caller's computed intent (issue.Change); this package decides
// nothing about order or position — it persists them as one operation.
type RankPlan struct {
	// Changes are applied in order. An Issue that appears in no change is
	// never written.
	Changes []issue.Change
	// Create, when set, is the Issue written after the Changes. Its ID is
	// allocated while the plan is prepared.
	Create *NewIssue
}

// NewIssue describes the Issue a RankPlan creates. Prefix and Rand feed
// issue.NextID; Issue is written as it is, with its final Rank.
type NewIssue struct {
	Prefix string
	Issue  issue.Issue
	Rand   io.Reader
}

// ApplyRankPlan persists plan as one recoverable operation and returns the
// ID of the Issue it created ("" when it creates none).
//
// The plan is prepared in full before the first write: every target is
// validated as a regular file, its original bytes and permission bits
// are read, and the bytes it intends to write are rendered. A failure
// there leaves every Issue untouched. When the plan changes existing
// Issues, their originals are then copied to <Vault>/.mt-recovery/<op>/.
//
// The Issues are written in order. When a write fails, every Issue whose
// write was attempted — the one that failed included, since a failed
// write may already have changed it — is restored to its original bytes
// and permissions, and the Issue created by the plan, if this call
// created it, is removed. When that restoration completes, nothing of
// the operation remains and the error says the previous state was
// restored. When it does not complete, the error is a *RecoveryError: it
// names the Issues still pending and the absolute path of the originals
// kept for manual recovery. Nothing is recovered automatically later.
//
// A plan with no change and no creation writes nothing and creates no
// recovery artefact. The contract assumes no other writer of the Vault
// while the operation runs; it does not survive an abrupt termination.
func (s Store) ApplyRankPlan(plan RankPlan) (string, error) {
	writes, err := s.prepareWrites(plan.Changes)
	if err != nil {
		return "", err
	}
	var create *createTarget
	if plan.Create != nil {
		if create, err = s.prepareCreate(*plan.Create); err != nil {
			return "", err
		}
	}
	dir := ""
	if len(writes) > 0 {
		if dir, err = s.keepOriginals(writes); err != nil {
			return "", err
		}
	}
	if applyErr := s.apply(writes, create); applyErr != nil {
		return "", s.rollback(applyErr, writes, create, dir)
	}
	id := ""
	if create != nil {
		id = create.id
	}
	if dir != "" {
		if err := inject(stageCleanup, ""); err != nil {
			return id, &CleanupError{Dir: dir, Err: err}
		}
		if err := removeRecovery(dir); err != nil {
			return id, &CleanupError{Dir: dir, Err: err}
		}
	}
	return id, nil
}

// rankWrite is one prepared Rank change: the Issue, the bytes it had, its
// permission bits, and the bytes the plan writes. attempted records that
// the write was started, which is when the file may have changed.
type rankWrite struct {
	id        string
	perm      fs.FileMode
	original  []byte
	next      []byte
	attempted bool
}

// createTarget is the prepared Issue of a plan: its allocated ID and the
// bytes to write. created is set only when this operation made the file.
type createTarget struct {
	id      string
	data    []byte
	created bool
}

// MissingIssueError reports a plan whose target Issue is not in the Vault.
// Absence is still a wrapped os.ErrNotExist, so a caller can keep testing
// errors.Is; the typed error carries the ID for the not-found message the
// command layer renders.
type MissingIssueError struct {
	ID  string
	Err error
}

func (e *MissingIssueError) Error() string { return e.Err.Error() }

// Unwrap returns the absence error (a wrapped os.ErrNotExist).
func (e *MissingIssueError) Unwrap() error { return e.Err }

// prepareWrites validates and renders every Rank change without touching
// the disk beyond reads. It refuses a malformed or duplicated ID, a
// missing or non-regular target and an unparsable Issue — so each such
// failure happens before the first write.
func (s Store) prepareWrites(changes []issue.Change) ([]*rankWrite, error) {
	writes := make([]*rankWrite, 0, len(changes))
	planned := make(map[string]bool, len(changes))
	for _, ch := range changes {
		if err := checkID(ch.ID); err != nil {
			return nil, err
		}
		if planned[ch.ID] {
			return nil, fmt.Errorf("issue %s is planned more than once", ch.ID)
		}
		planned[ch.ID] = true
		item, original, perm, err := s.load(ch.ID)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil, &MissingIssueError{ID: ch.ID, Err: err}
			}
			return nil, err
		}
		if sameRank(item.Issue.Frontmatter.Rank, ch.Rank) {
			// An Issue whose Rank does not change is never rewritten.
			continue
		}
		item.Issue.Frontmatter.Rank = ch.Rank
		next, err := issue.Render(item.Issue)
		if err != nil {
			return nil, err
		}
		writes = append(writes, &rankWrite{id: ch.ID, perm: perm, original: original, next: next})
	}
	return writes, nil
}

// prepareCreate allocates the ID of the Issue the plan creates and renders
// it. The ID avoids every name the Vault uses, symlinks included.
func (s Store) prepareCreate(n NewIssue) (*createTarget, error) {
	taken, err := s.takenIDs()
	if err != nil {
		return nil, err
	}
	id, err := issue.NextID(n.Prefix, taken, n.Rand)
	if err != nil {
		return nil, err
	}
	if err := checkID(id); err != nil {
		return nil, err
	}
	data, err := issue.Render(n.Issue)
	if err != nil {
		return nil, err
	}
	return &createTarget{id: id, data: data}, nil
}

// keepOriginals creates the operation's recovery directory and copies the
// original of every Issue the plan will write into it. If that fails, the
// directory is removed again: no Issue has been touched yet.
func (s Store) keepOriginals(writes []*rankWrite) (string, error) {
	dir, err := s.newRecoveryDir()
	if err != nil {
		return "", err
	}
	for _, w := range writes {
		if err := inject(stageKeep, w.id); err != nil {
			return "", discardRecovery(dir, err)
		}
		if err := writeOriginal(filepath.Join(dir, w.id+".md"), w); err != nil {
			return "", discardRecovery(dir, err)
		}
	}
	return dir, nil
}

// newRecoveryDir creates <Vault>/.mt-recovery/<op>/ and returns its
// absolute path. The name carries a UTC instant and random bytes so two
// operations never share a directory. An existing .mt-recovery that is
// not a real directory (a file or a symlink) is refused, never followed.
func (s Store) newRecoveryDir() (string, error) {
	base, err := filepath.Abs(s.dir)
	if err != nil {
		return "", fmt.Errorf("resolving the Vault directory: %w", err)
	}
	parent := filepath.Join(base, recoveryDirName)
	info, err := os.Lstat(parent)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		if err := os.Mkdir(parent, 0o755); err != nil {
			return "", fmt.Errorf("creating recovery directory: %w", err)
		}
	case err != nil:
		return "", fmt.Errorf("checking recovery directory: %w", err)
	case !info.IsDir():
		return "", fmt.Errorf("recovery path %s is not a directory", parent)
	}
	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("naming recovery directory: %w", err)
	}
	name := time.Now().UTC().Format("20060102T150405.000000000Z") + "-" + hex.EncodeToString(random)
	dir := filepath.Join(parent, name)
	if err := os.Mkdir(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating recovery directory: %w", err)
	}
	return dir, nil
}

// writeOriginal stores the original bytes of one Issue at path, with the
// Issue's permission bits. O_EXCL keeps an existing copy from being
// overwritten.
func writeOriginal(path string, w *rankWrite) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("keeping original of issue %s: %w", w.id, err)
	}
	_, writeErr := f.Write(w.original)
	chmodErr := f.Chmod(w.perm)
	closeErr := f.Close()
	if err := firstErr(writeErr, chmodErr, closeErr); err != nil {
		return fmt.Errorf("keeping original of issue %s: %w", w.id, err)
	}
	return nil
}

// apply writes the prepared plan in order: the Rank changes first, then
// the created Issue. Each write is attempted before it starts, so a
// failure at any point leaves every touched file in the attempted set.
// The inject calls are the package's fault seam (see fault.go).
func (s Store) apply(writes []*rankWrite, create *createTarget) error {
	for _, w := range writes {
		w.attempted = true
		if err := inject(stageBeforeWrite, w.id); err != nil {
			return fmt.Errorf("writing issue %s: %w", w.id, err)
		}
		if err := s.writeBytes(w.id, w.next); err != nil {
			return err
		}
		if err := inject(stageAfterWrite, w.id); err != nil {
			return fmt.Errorf("writing issue %s: %w", w.id, err)
		}
	}
	if create != nil {
		if err := inject(stageBeforeCreate, create.id); err != nil {
			return fmt.Errorf("writing issue %s: %w", create.id, err)
		}
		created, err := s.createBytes(create.id, create.data)
		create.created = created
		if err != nil {
			return err
		}
		if err := inject(stageCreated, create.id); err != nil {
			return fmt.Errorf("writing issue %s: %w", create.id, err)
		}
	}
	return nil
}

// rollback restores the state the operation found after applyErr. It
// tries every attempted Issue and the created one, even when some fail,
// and reports what it could not restore. Originals are kept whenever
// something is pending; a complete restoration removes the recovery
// directory.
func (s Store) rollback(applyErr error, writes []*rankWrite, create *createTarget, dir string) error {
	pending := s.restore(writes, create, dir)
	if len(pending) == 0 {
		if dir != "" {
			if err := inject(stageCleanup, ""); err != nil {
				return &RecoveryError{Cause: applyErr, Cleanup: err, Dir: dir}
			}
			if err := removeRecovery(dir); err != nil {
				return &RecoveryError{Cause: applyErr, Cleanup: err, Dir: dir}
			}
		}
		if !touched(writes, create) {
			return applyErr
		}
		return fmt.Errorf("%w; the previous state was restored", applyErr)
	}
	keep := ""
	for _, p := range pending {
		if p.Original != "" {
			keep = dir
		}
	}
	if keep == "" && dir != "" {
		// Only a removal is pending: no original is left to keep.
		_ = removeRecovery(dir)
	}
	return &RecoveryError{Cause: applyErr, Dir: keep, Pending: pending}
}

// touched reports whether the operation attempted any write or created a
// file: only then does a failure have a state to restore.
func touched(writes []*rankWrite, create *createTarget) bool {
	for _, w := range writes {
		if w.attempted {
			return true
		}
	}
	return create != nil && create.created
}

// sameRank reports whether two Ranks are equal, Backlog (nil) included.
func sameRank(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// restore brings back every attempted Issue from its original bytes and
// removes the created Issue when this operation made it. It returns what
// it could not restore. The copy of an Issue restored successfully is
// removed on the way; copies of pending Issues stay in dir.
func (s Store) restore(writes []*rankWrite, create *createTarget, dir string) []Pending {
	var pending []Pending
	for _, w := range writes {
		if !w.attempted {
			continue
		}
		original := ""
		if dir != "" {
			original = filepath.Join(dir, w.id+".md")
		}
		if err := s.restoreFile(w); err != nil {
			pending = append(pending, Pending{ID: w.id, Original: original, Err: err})
			continue
		}
		if original != "" {
			// Best effort: a copy that survives is still a correct
			// original, and a complete restoration removes the directory.
			_ = os.Remove(original)
		}
	}
	if create != nil && create.created {
		if err := os.Remove(s.path(create.id)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			pending = append(pending, Pending{ID: create.id, Created: true, Err: err})
		}
	}
	return pending
}

// restoreFile writes the original bytes back into an Issue and sets its
// permission bits. The file must still be a regular file, and the open
// refuses a symlink. A read-only Issue is opened for the owner's write
// access only for the restoration, and its mode is set back afterwards.
func (s Store) restoreFile(w *rankWrite) error {
	if err := inject(stageRestore, w.id); err != nil {
		return err
	}
	if _, err := s.regularPerm(w.id); err != nil {
		return err
	}
	path := s.path(w.id)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|issueOpenNoFollow, 0)
	if errors.Is(err, fs.ErrPermission) {
		if err := os.Chmod(path, w.perm|0o200); err != nil {
			return err
		}
		f, err = os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|issueOpenNoFollow, 0)
		if err != nil {
			_ = os.Chmod(path, w.perm)
		}
	}
	if err != nil {
		return err
	}
	_, writeErr := f.Write(w.original)
	chmodErr := f.Chmod(w.perm)
	closeErr := f.Close()
	return firstErr(writeErr, chmodErr, closeErr)
}

// discardRecovery removes a recovery directory whose preparation failed,
// keeping the preparation error as the cause.
func discardRecovery(dir string, cause error) error {
	if err := removeRecovery(dir); err != nil {
		return fmt.Errorf("%w; removing the recovery directory %s failed: %v", cause, dir, err)
	}
	return cause
}

// removeRecovery deletes an operation's recovery directory and, when it
// was the last one, the .mt-recovery directory beside it. A parent that
// still holds another operation is expected, not a failure.
func removeRecovery(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	err := os.Remove(filepath.Dir(dir))
	if err != nil && !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, syscall.ENOTEMPTY) && !errors.Is(err, syscall.EEXIST) {
		return err
	}
	return nil
}

// firstErr returns the first non-nil error, or nil.
func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// Pending is one Issue a restoration could not bring back. A restored Issue
// is named by its original bytes (Original, an absolute path); an Issue
// the operation created and could not remove has Created set and no
// original, because there was none.
type Pending struct {
	ID       string
	Created  bool
	Original string
	Err      error
}

// RecoveryError reports a rank operation whose application failed and
// whose restoration did not complete. Cause is the error that stopped the
// application. Pending lists the Issues the restoration could not bring
// back, and Dir is the absolute path of the directory that keeps their
// originals. The Vault may hold an intermediate state; the operation has
// not succeeded and is not recovered. Cleanup, when set, reports a
// restoration that completed while its recovery directory could not be
// removed. Exit code 1 — it is an operational failure, not a usage one.
type RecoveryError struct {
	Cause   error
	Dir     string
	Pending []Pending
	Cleanup error
}

func (e *RecoveryError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%v", e.Cause)
	if len(e.Pending) == 0 {
		fmt.Fprintf(&b, "; the issues were restored, but the recovery directory %s could not be removed: %v", e.Dir, e.Cleanup)
		return b.String()
	}
	b.WriteString("; recovery incomplete: the Vault may hold an intermediate state")
	for _, p := range e.Pending {
		if p.Created {
			fmt.Fprintf(&b, "\n  issue %s was created by this operation and could not be removed: %v", p.ID, p.Err)
			continue
		}
		fmt.Fprintf(&b, "\n  issue %s was not restored: %v; its original is kept at %s", p.ID, p.Err, p.Original)
	}
	if e.Dir != "" {
		fmt.Fprintf(&b, "\nto restore an issue by hand, copy its original over issues/<id>.md (bytes and permissions are kept); originals are in %s", e.Dir)
	}
	return b.String()
}

// Unwrap returns the error that stopped the application.
func (e *RecoveryError) Unwrap() error { return e.Cause }

// CleanupError reports an operation that was applied — the Issues carry the
// new state — whose recovery directory could not be removed.
type CleanupError struct {
	Dir string
	Err error
}

func (e *CleanupError) Error() string {
	return fmt.Sprintf("rank change applied, but the recovery directory %s could not be removed: %v", e.Dir, e.Err)
}

// Unwrap returns the removal error.
func (e *CleanupError) Unwrap() error { return e.Err }
