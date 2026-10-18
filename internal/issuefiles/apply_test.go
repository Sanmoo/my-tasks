// The recoverable application of a RankPlan, black-box tested against
// t.TempDir() vaults. The fault seam (export_test.go) forces failures at
// the package's own stages; every assertion is on observable results:
// bytes and permissions on disk, the returned error, and the recovery
// artefacts left behind.
package issuefiles_test

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issue"
	"github.com/Sanmoo/my-tasks2/internal/issuefiles"
)

// rankedIssue is a well-formed Issue file with a rank field.
func rankedIssue(title string, rank int) string {
	return fmt.Sprintf("---\ntitle: %s\nstatus: open\nlabels: []\ncreated_at: 2026-01-01T10:00\nrank: %d\n---\n\n## Description\n## Notes\n## Comments\n", title, rank)
}

// backlogIssue is a well-formed Issue file without a rank field.
func backlogIssue(title string) string {
	return strings.Replace(validIssue(title), "status: open", "status: open", 1)
}

// withRank returns i with its Rank set to rank.
func withRank(i issue.Issue, rank int) issue.Issue {
	r := rank
	i.Frontmatter.Rank = &r
	return i
}

// rankedItem returns the Change that sets id's Rank to rank (nil for the
// Backlog).
func rankedItem(id string, rank int) issue.Change {
	r := rank
	return issue.Change{ID: id, Rank: &r}
}

// changeRank builds a Change with rank (nil = Backlog).
func changeRank(id string, rank int) issue.Change {
	return rankedItem(id, rank)
}

// readPerm returns the permission bits of path.
func readPerm(t *testing.T, path string) fs.FileMode {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Mode().Perm()
}

// recoveryRoot is the Vault's recovery directory.
func recoveryRoot(dir string) string {
	return filepath.Join(dir, ".mt-recovery")
}

// recoveryOps lists the operation directories under .mt-recovery.
func recoveryOps(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(recoveryRoot(dir))
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// assertNoRecoveryArtifacts fails the test when the Vault holds any
// recovery artefact.
func assertNoRecoveryArtifacts(t *testing.T, dir string) {
	t.Helper()
	if _, err := os.Lstat(recoveryRoot(dir)); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf(".mt-recovery exists after the operation: want no recovery artefacts (Lstat err = %v)", err)
	}
}

// hookAt returns a fault hook that fails exactly once at stage/id with
// boom, and a function reporting whether it fired.
func hookAt(stage issuefiles.FaultStage, id string, boom error) (func(issuefiles.FaultStage, string) error, func() bool) {
	fired := false
	hook := func(s issuefiles.FaultStage, i string) error {
		if s == stage && i == id {
			fired = true
			return boom
		}
		return nil
	}
	return hook, func() bool { return fired }
}

// TestApplyRankPlanAppliesTheWholePlan is the happy path: every change
// lands, the returned ID names the created Issue, unchanged Issues are
// not rewritten and no recovery artefact remains.
func TestApplyRankPlanAppliesTheWholePlan(t *testing.T) {
	dir := newVault(t)
	a := rankedIssue("a", 1)
	b := rankedIssue("b", 2)
	c := backlogIssue("c")
	writeIssue(t, dir, "pkm-aaa", a)
	writeIssue(t, dir, "pkm-bbb", b)
	writeIssue(t, dir, "pkm-ccc", c)

	id, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-aaa", 2), changeRank("pkm-ccc", 1)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if id != "" {
		t.Errorf("ApplyRankPlan without a creation returned id %q, want \"\"", id)
	}
	if got := readIssueFile(t, dir, "pkm-aaa"); got != rankedIssue("a", 2) {
		t.Errorf("pkm-aaa after the plan = %q, want rank 2", got)
	}
	if got := readIssueFile(t, dir, "pkm-ccc"); got != rankedIssue("c", 1) {
		t.Errorf("pkm-ccc after the plan = %q, want rank 1", got)
	}
	// An Issue whose Rank does not change is never rewritten.
	if got := readIssueFile(t, dir, "pkm-bbb"); got != b {
		t.Errorf("pkm-bbb was rewritten by the plan: %q", got)
	}
	assertNoRecoveryArtifacts(t, dir)
}

// TestApplyRankPlanPreparationFailureWritesNothing: a plan whose later
// target is missing (or non-regular) leaves every earlier Issue
// untouched and creates no recovery artefact.
func TestApplyRankPlanPreparationFailureWritesNothing(t *testing.T) {
	for name, target := range map[string]func(t *testing.T, dir string){
		"missing": func(t *testing.T, dir string) {
			// pkm-zzz does not exist.
		},
		"symlink": func(t *testing.T, dir string) {
			link(t, filepath.Join(t.TempDir(), "outside.md"), issuePath(dir, "pkm-zzz"))
		},
	} {
		t.Run(name, func(t *testing.T) {
			dir := newVault(t)
			before := rankedIssue("a", 1)
			writeIssue(t, dir, "pkm-aaa", before)
			target(t, dir)

			_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
				Changes: []issue.Change{changeRank("pkm-aaa", 2), changeRank("pkm-zzz", 1)},
			})
			if err == nil {
				t.Fatal("ApplyRankPlan(preparation failure) = nil, want an error")
			}
			if got := readIssueFile(t, dir, "pkm-aaa"); got != before {
				t.Errorf("pkm-aaa was changed by a plan that failed its preparation: %q", got)
			}
			assertNoRecoveryArtifacts(t, dir)
		})
	}
}

// TestApplyRankPlanPreparationFailureIsTyped: a missing target is a
// MissingIssueError that still wraps os.ErrNotExist, so the command
// layer keeps rendering its not-found message.
func TestApplyRankPlanPreparationFailureIsTyped(t *testing.T) {
	dir := newVault(t)
	_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-zzz", 1)},
	})
	var missing *issuefiles.MissingIssueError
	if !errors.As(err, &missing) {
		t.Fatalf("ApplyRankPlan(missing target) error = %v, want MissingIssueError", err)
	}
	if missing.ID != "pkm-zzz" {
		t.Errorf("MissingIssueError.ID = %q, want pkm-zzz", missing.ID)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("MissingIssueError does not wrap ErrNotExist: %v", err)
	}
}

// TestApplyRankPlanRestoresAfterAFailedWrite: a failure after a
// successful write restores every attempted Issue to its original bytes
// and permissions — the Issue whose write failed included — and leaves
// no recovery artefact.
func TestApplyRankPlanRestoresAfterAFailedWrite(t *testing.T) {
	for name, stage := range map[string]issuefiles.FaultStage{
		"failure before the write of pkm-bbb": issuefiles.StageBeforeWrite,
		"failure after the write of pkm-bbb":  issuefiles.StageAfterWrite,
	} {
		t.Run(name, func(t *testing.T) {
			dir := newVault(t)
			a, b := rankedIssue("a", 1), rankedIssue("b", 2)
			writeIssue(t, dir, "pkm-aaa", a)
			writeIssue(t, dir, "pkm-bbb", b)
			if err := os.Chmod(issuePath(dir, "pkm-bbb"), 0o600); err != nil {
				t.Fatal(err)
			}
			bPerm := readPerm(t, issuePath(dir, "pkm-bbb"))
			boom := errors.New("disk on fire")

			hook, fired := hookAt(stage, "pkm-bbb", boom)
			cleanup := issuefiles.SetFaultHook(hook)
			defer cleanup()

			_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
				Changes: []issue.Change{changeRank("pkm-aaa", 2), changeRank("pkm-bbb", 1)},
			})
			if !fired() {
				t.Fatal("the fault hook did not fire")
			}
			if err == nil {
				t.Fatal("ApplyRankPlan(failing write) = nil, want an error")
			}
			if !errors.Is(err, boom) {
				t.Errorf("the error does not name the write failure: %v", err)
			}
			if got := readIssueFile(t, dir, "pkm-aaa"); got != a {
				t.Errorf("pkm-aaa after the failed operation = %q, want its original bytes", got)
			}
			if got := readIssueFile(t, dir, "pkm-bbb"); got != b {
				t.Errorf("pkm-bbb after the failed operation = %q, want its original bytes", got)
			}
			if got := readPerm(t, issuePath(dir, "pkm-bbb")); got != bPerm {
				t.Errorf("pkm-bbb permissions after the failed operation = %v, want %v", got, bPerm)
			}
			assertNoRecoveryArtifacts(t, dir)
		})
	}
}

// TestApplyRankPlanRestoresASwappedPair: the Rank 1/2 exchange fails at
// the second write — the disk holds an intermediate state, and the
// operation may not call it recovered until both Issues are back.
func TestApplyRankPlanRestoresASwappedPair(t *testing.T) {
	dir := newVault(t)
	a, b := rankedIssue("a", 1), rankedIssue("b", 2)
	writeIssue(t, dir, "pkm-aaa", a)
	writeIssue(t, dir, "pkm-bbb", b)
	boom := errors.New("second write fails")

	// Swap: a→2, b→1. The hook fails the second write after pkm-aaa was
	// already rewritten, so the disk briefly holds ranks 2 and 2.
	hook, fired := hookAt(issuefiles.StageAfterWrite, "pkm-bbb", boom)
	cleanup := issuefiles.SetFaultHook(hook)
	defer cleanup()

	_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-aaa", 2), changeRank("pkm-bbb", 1)},
	})
	if !fired() || err == nil {
		t.Fatalf("hook fired = %v, err = %v", fired(), err)
	}
	if strings.Contains(err.Error(), "restored") {
		// The message may only say restored when the restoration ended.
		if got := readIssueFile(t, dir, "pkm-aaa"); got != a || readIssueFile(t, dir, "pkm-bbb") != b {
			t.Errorf("error claims restoration but the disk holds: a=%q b=%q", readIssueFile(t, dir, "pkm-aaa"), readIssueFile(t, dir, "pkm-bbb"))
		}
	}
	if got := readIssueFile(t, dir, "pkm-aaa"); got != a {
		t.Errorf("pkm-aaa after the failed swap = %q, want its original bytes", got)
	}
	if got := readIssueFile(t, dir, "pkm-bbb"); got != b {
		t.Errorf("pkm-bbb after the failed swap = %q, want its original bytes", got)
	}
	assertNoRecoveryArtifacts(t, dir)
}

// TestApplyRankPlanRestoresReadOnlyPermissions: a restoration brings
// back the original permission bits, including a read-only Issue.
func TestApplyRankPlanRestoresReadOnlyPermissions(t *testing.T) {
	dir := newVault(t)
	a, b := rankedIssue("a", 1), rankedIssue("b", 2)
	writeIssue(t, dir, "pkm-aaa", a)
	writeIssue(t, dir, "pkm-bbb", b)
	if err := os.Chmod(issuePath(dir, "pkm-aaa"), 0o400); err != nil {
		t.Fatal(err)
	}
	boom := errors.New("write fails after the first")

	// The second change fails after its write, so the restoration has to
	// write the read-only pkm-aaa back — it opens it for the owner's
	// write access only for the restoration and sets the mode back.
	hook, fired := hookAt(issuefiles.StageAfterWrite, "pkm-bbb", boom)
	cleanup := issuefiles.SetFaultHook(hook)
	defer cleanup()

	_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-bbb", 3), changeRank("pkm-aaa", 2)},
	})
	if !fired() {
		t.Fatal("the fault hook did not fire")
	}
	if err == nil {
		t.Fatal("ApplyRankPlan(failing write) = nil, want an error")
	}
	if got, err := os.ReadFile(issuePath(dir, "pkm-aaa")); err != nil || string(got) != a {
		t.Errorf("pkm-aaa after the failed operation = %q (%v), want its original bytes", string(got), err)
	}
	if got := readPerm(t, issuePath(dir, "pkm-aaa")); got != 0o400 {
		t.Errorf("pkm-aaa permissions after the failed operation = %v, want 0400", got)
	}
	assertNoRecoveryArtifacts(t, dir)
}

// TestApplyRankPlanWithoutChangesLeavesNoArtifacts: a plan whose every
// change is already in place writes nothing and creates no recovery
// directory at all.
func TestApplyRankPlanWithoutChangesLeavesNoArtifacts(t *testing.T) {
	dir := newVault(t)
	a := rankedIssue("a", 1)
	writeIssue(t, dir, "pkm-aaa", a)

	if _, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-aaa", 1)},
	}); err != nil {
		t.Fatal(err)
	}
	if got := readIssueFile(t, dir, "pkm-aaa"); got != a {
		t.Errorf("pkm-aaa was rewritten by an empty plan: %q", got)
	}
	assertNoRecoveryArtifacts(t, dir)
}

// TestApplyRankPlanKeepsFieldsThatAreNotRank: the planned write renders
// the whole Issue — body and every frontmatter field survive.
func TestApplyRankPlanKeepsFieldsThatAreNotRank(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-aaa", "---\ntitle: a\nstatus: open\nlabels: [x, y]\ncreated_at: 2026-01-01T10:00\nblocked_by: [pkm-bbb]\nrank: 2\n---\n\n## Description\n\nbody line\n\n## Notes\n## Comments\n")

	if _, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-aaa", 1)},
	}); err != nil {
		t.Fatal(err)
	}
	got := readIssueFile(t, dir, "pkm-aaa")
	for _, want := range []string{"title: a", "labels: [x, \"y\"]", "created_at: 2026-01-01T10:00", "blocked_by: [pkm-bbb]", "body line"} {
		if !strings.Contains(got, want) {
			t.Errorf("pkm-aaa after the plan misses %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "rank: 2") {
		t.Errorf("pkm-aaa kept the old rank:\n%s", got)
	}
}

// TestApplyRankPlanKeepsTheOriginalWhenTheRestorationIsIncomplete: one
// Issue cannot be restored — the error names it and the absolute path of
// its original, the others are restored, and the original kept on disk
// carries the original bytes and permissions.
func TestApplyRankPlanKeepsTheOriginalWhenTheRestorationIsIncomplete(t *testing.T) {
	dir := newVault(t)
	a, b := rankedIssue("a", 1), rankedIssue("b", 2)
	writeIssue(t, dir, "pkm-aaa", a)
	writeIssue(t, dir, "pkm-bbb", b)
	if err := os.Chmod(issuePath(dir, "pkm-aaa"), 0o600); err != nil {
		t.Fatal(err)
	}
	aPerm := readPerm(t, issuePath(dir, "pkm-aaa"))
	boom := errors.New("write fails")

	// The operation fails after pkm-bbb was written, then the hook also
	// refuses pkm-aaa's restoration: the recovery stays incomplete.
	var calls int
	hook := func(stage issuefiles.FaultStage, id string) error {
		switch {
		case stage == issuefiles.StageAfterWrite && id == "pkm-bbb":
			return boom
		case stage == issuefiles.StageRestore && id == "pkm-aaa":
			calls++
			return errors.New("restoration fails")
		}
		return nil
	}
	cleanup := issuefiles.SetFaultHook(hook)
	defer cleanup()

	_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-aaa", 2), changeRank("pkm-bbb", 1)},
	})
	var recovery *issuefiles.RecoveryError
	if !errors.As(err, &recovery) {
		t.Fatalf("ApplyRankPlan(incomplete recovery) error = %v, want RecoveryError", err)
	}
	if !errors.Is(err, boom) {
		t.Errorf("RecoveryError does not carry the original failure: %v", err)
	}
	if len(recovery.Pending) != 1 || recovery.Pending[0].ID != "pkm-aaa" || recovery.Pending[0].Created {
		t.Fatalf("Pending = %+v, want exactly pkm-aaa pending", recovery.Pending)
	}
	if calls == 0 {
		t.Error("the restoration of pkm-aaa was not attempted")
	}
	// pkm-bbb was restored; pkm-aaa may hold anything the failed
	// restoration left — the error says so instead of claiming success.
	if got := readIssueFile(t, dir, "pkm-bbb"); got != b {
		t.Errorf("pkm-bbb after the incomplete recovery = %q, want its original bytes", got)
	}
	// The original of the pending Issue is kept beside the Vault, with
	// its bytes and permissions, outside the Issue listing.
	original := recovery.Pending[0].Original
	if !filepath.IsAbs(original) {
		t.Errorf("Pending original path %q is not absolute", original)
	}
	if !strings.Contains(original, ".mt-recovery") {
		t.Errorf("Pending original path %q is not under .mt-recovery", original)
	}
	if data, err := os.ReadFile(original); err != nil || string(data) != a {
		t.Errorf("kept original = %q (%v), want pkm-aaa's original bytes", string(data), err)
	}
	if got := readPerm(t, original); got != aPerm {
		t.Errorf("kept original permissions = %v, want %v", got, aPerm)
	}
	items, err := issuefiles.Open(dir).List()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range items {
		if it.ID == "a" || filepath.Base(it.ID) != it.ID {
			t.Errorf("the recovery area leaked into the Issue listing: %v", ids(items))
		}
	}
	// A recovery that was incomplete keeps the originals: nothing is
	// removed automatically afterwards.
	ops := recoveryOps(t, dir)
	if len(ops) != 1 {
		t.Errorf("recovery operations on disk = %v, want the pending one kept", ops)
	}
}

// TestApplyRankPlanCreatesJointly: the shifts and the newcomer are one
// operation — the shifts land, the newcomer is created with its rank,
// and no recovery artefact remains.
func TestApplyRankPlanCreatesJointly(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-aaa", rankedIssue("a", 1))
	writeIssue(t, dir, "pkm-bbb", rankedIssue("b", 2))

	id, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-aaa", 2), changeRank("pkm-bbb", 3)},
		Create:  &issuefiles.NewIssue{Prefix: "pkm", Issue: withRank(newIssue("new"), 1), Rand: zeroReader{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("ApplyRankPlan with a creation returned an empty id")
	}
	if got := readIssueFile(t, dir, "pkm-aaa"); got != rankedIssue("a", 2) {
		t.Errorf("pkm-aaa after the operation = %q, want rank 2", got)
	}
	data, err := os.ReadFile(issuePath(dir, id))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "title: new") {
		t.Errorf("the created Issue %s = %q, want the newcomer", id, string(data))
	}
	assertNoRecoveryArtifacts(t, dir)
}

// TestApplyRankPlanCreateFailureRestoresTheNeighbors: a failure after
// the shifts were written restores them and removes the Issue this
// operation created.
func TestApplyRankPlanCreateFailureRestoresTheNeighbors(t *testing.T) {
	dir := newVault(t)
	a := rankedIssue("a", 1)
	writeIssue(t, dir, "pkm-aaa", a)
	boom := errors.New("write fails")

	hook, fired := hookAt(issuefiles.StageAfterWrite, "pkm-aaa", boom)
	cleanup := issuefiles.SetFaultHook(hook)
	defer cleanup()

	_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-aaa", 2)},
		Create:  &issuefiles.NewIssue{Prefix: "pkm", Issue: newIssue("new"), Rand: zeroReader{}},
	})
	if !fired() || err == nil {
		t.Fatalf("hook fired = %v, err = %v", fired(), err)
	}
	if got := readIssueFile(t, dir, "pkm-aaa"); got != a {
		t.Errorf("pkm-aaa after the failed creation = %q, want its original bytes", got)
	}
	entries, err := os.ReadDir(filepath.Join(dir, "issues"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("issues/ holds %d entries after the failed creation, want only pkm-aaa", len(entries))
	}
	assertNoRecoveryArtifacts(t, dir)
}

// TestApplyRankPlanCreateFailureRemovesOnlyTheCreatedIssue: when the
// created Issue could not be removed, the error reports the pending
// removal — never a pending restoration, because there was no original.
func TestApplyRankPlanCreateFailureRemovesOnlyTheCreatedIssue(t *testing.T) {
	dir := newVault(t)
	a, b := rankedIssue("a", 1), rankedIssue("b", 2)
	writeIssue(t, dir, "pkm-aaa", a)
	writeIssue(t, dir, "pkm-bbb", b)
	boom := errors.New("write fails")
	createdIssueChmodded := false

	// The hook fails the operation right after the creation landed, and
	// then makes issues/ read-only before the restores: the writes come
	// back (a write into an existing file needs no directory write), but
	// the removal of the created Issue fails — issues/ refuses it.
	hook := func(stage issuefiles.FaultStage, id string) error {
		switch {
		case stage == issuefiles.StageCreated && strings.HasPrefix(id, "pkm-0"):
			return boom
		case stage == issuefiles.StageRestore && !createdIssueChmodded:
			createdIssueChmodded = true
			if err := os.Chmod(filepath.Join(dir, "issues"), 0o555); err != nil {
				t.Fatal(err)
			}
			return nil
		}
		return nil
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(dir, "issues"), 0o755) })
	cleanup := issuefiles.SetFaultHook(hook)
	defer cleanup()

	_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-aaa", 2), changeRank("pkm-bbb", 3)},
		Create:  &issuefiles.NewIssue{Prefix: "pkm", Issue: newIssue("new"), Rand: zeroReader{}},
	})
	var recovery *issuefiles.RecoveryError
	if !errors.As(err, &recovery) {
		t.Fatalf("ApplyRankPlan(pending removal) error = %v, want RecoveryError", err)
	}
	if !errors.Is(err, boom) {
		t.Errorf("RecoveryError does not carry the original failure: %v", err)
	}
	if len(recovery.Pending) != 1 {
		t.Fatalf("Pending = %+v, want exactly the pending removal", recovery.Pending)
	}
	p := recovery.Pending[0]
	if !p.Created || p.Original != "" {
		t.Errorf("Pending = %+v, want a pending removal (Created) without an original", p)
	}
	if got := readIssueFile(t, dir, "pkm-aaa"); got != a {
		t.Errorf("pkm-aaa after the recovery = %q, want its original bytes", got)
	}
	if got := readIssueFile(t, dir, "pkm-bbb"); got != b {
		t.Errorf("pkm-bbb after the recovery = %q, want its original bytes", got)
	}
	// No original was kept: a pending removal has nothing to restore.
	if recovery.Dir != "" {
		t.Errorf("RecoveryError.Dir = %q, want empty — a pending removal keeps no original", recovery.Dir)
	}
}

// TestApplyRankPlanNeverRemovesAPreexistingEntry: a name that slipped in
// between the scan and the open fails the creation, and the rollback
// must leave that entry alone — it was never this operation's.
func TestApplyRankPlanNeverRemovesAPreexistingEntry(t *testing.T) {
	dir := newVault(t)
	a := rankedIssue("a", 1)
	writeIssue(t, dir, "pkm-aaa", a)
	occupied := rankedIssue("occupant", 1)
	boom := errors.New("open fails")

	// The hook plants the occupant at the ID the plan is about to
	// create — the name the scan could not have seen — and fails the
	// open, as an EEXIST would.
	hook := func(stage issuefiles.FaultStage, id string) error {
		if stage == issuefiles.StageBeforeCreate {
			if err := os.WriteFile(issuePath(dir, id), []byte(occupied), 0o644); err != nil {
				t.Fatal(err)
			}
			return boom
		}
		return nil
	}
	cleanup := issuefiles.SetFaultHook(hook)
	defer cleanup()

	if _, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-aaa", 2)},
		Create:  &issuefiles.NewIssue{Prefix: "pkm", Issue: newIssue("new"), Rand: zeroReader{}},
	}); !errors.Is(err, boom) {
		t.Fatalf("ApplyRankPlan(occupied name) error = %v, want the creation failure", err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, "issues"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Errorf("issues/ holds %d entries after the failed creation, want pkm-aaa and the occupant", len(entries))
	}
	for _, e := range entries {
		if e.Name() == "pkm-aaa.md" {
			continue
		}
		if got, err := os.ReadFile(issuePath(dir, strings.TrimSuffix(e.Name(), ".md"))); err != nil || string(got) != occupied {
			t.Errorf("the preexisting entry %s was touched by the rollback: %q (%v)", e.Name(), string(got), err)
		}
	}
	if got := readIssueFile(t, dir, "pkm-aaa"); got != a {
		t.Errorf("pkm-aaa after the failed creation = %q, want its original bytes", got)
	}
	assertNoRecoveryArtifacts(t, dir)
}

// TestApplyRankPlanReportsACleanupFailure: an applied operation whose
// recovery directory cannot be removed is an error that says so — the
// directory stays and the success is not hidden.
func TestApplyRankPlanReportsACleanupFailure(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-aaa", rankedIssue("a", 1))
	boom := errors.New("cleanup fails")

	hook, fired := hookAt(issuefiles.StageCleanup, "", boom)
	cleanup := issuefiles.SetFaultHook(hook)
	defer cleanup()

	id, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-aaa", 2)},
	})
	if !fired() {
		t.Fatal("the fault hook did not fire")
	}
	var cleanupErr *issuefiles.CleanupError
	if !errors.As(err, &cleanupErr) {
		t.Fatalf("ApplyRankPlan(failing cleanup) error = %v, want CleanupError", err)
	}
	if !errors.Is(err, boom) {
		t.Errorf("CleanupError does not carry the removal failure: %v", err)
	}
	if got := readIssueFile(t, dir, "pkm-aaa"); got != rankedIssue("a", 2) {
		t.Errorf("pkm-aaa after the applied operation = %q, want the new rank", got)
	}
	if id != "" {
		t.Errorf("id = %q, want \"\"", id)
	}
	ops := recoveryOps(t, dir)
	if len(ops) != 1 {
		t.Errorf("recovery operations on disk = %v, want the failed cleanup's kept", ops)
	}
}

// TestApplyRankPlanRejectsAMalformedOrDuplicatedTarget: the preparation
// refuses a plan that names an ID that cannot exist or names one Issue
// twice — before any write and without artefacts.
func TestApplyRankPlanRejectsAMalformedOrDuplicatedTarget(t *testing.T) {
	for name, changes := range map[string][]issue.Change{
		"malformed ID": {{ID: "../escape", Rank: nil}},
		"duplicated":   {changeRank("pkm-aaa", 2), changeRank("pkm-aaa", 3)},
	} {
		t.Run(name, func(t *testing.T) {
			dir := newVault(t)
			before := rankedIssue("a", 1)
			writeIssue(t, dir, "pkm-aaa", before)

			if _, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{Changes: changes}); err == nil {
				t.Fatal("ApplyRankPlan(broken plan) = nil, want an error")
			}
			if got := readIssueFile(t, dir, "pkm-aaa"); got != before {
				t.Errorf("pkm-aaa was changed by a broken plan: %q", got)
			}
			assertNoRecoveryArtifacts(t, dir)
		})
	}
}

// TestApplyRankPlanPreparationFailureNamesTheCause: a preparation failure
// says which Issue it was preparing and why, before anything is written.
func TestApplyRankPlanPreparationFailureNamesTheCause(t *testing.T) {
	dir := newVault(t)
	writeIssue(t, dir, "pkm-aaa", rankedIssue("a", 1))
	writeIssue(t, dir, "pkm-bbb", rankedIssue("b", 2))
	boom := errors.New("keeping originals fails")

	hook, fired := hookAt(issuefiles.StageKeep, "pkm-aaa", boom)
	cleanup := issuefiles.SetFaultHook(hook)
	defer cleanup()

	_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-aaa", 2), changeRank("pkm-bbb", 1)},
	})
	if !fired() {
		t.Fatal("the fault hook did not fire")
	}
	if !errors.Is(err, boom) {
		t.Fatalf("ApplyRankPlan(keep failure) error = %v, want the preparation failure", err)
	}
	if got := readIssueFile(t, dir, "pkm-aaa"); got != rankedIssue("a", 1) {
		t.Errorf("pkm-aaa was changed by a failed preparation: %q", got)
	}
	assertNoRecoveryArtifacts(t, dir)
}

// TestApplyRankPlanPreparationFailureNamesTheMissingIssue: the typed
// error renders the target the plan could not find.
func TestApplyRankPlanPreparationFailureNamesTheMissingIssue(t *testing.T) {
	dir := newVault(t)
	_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-zzz", 1)},
	})
	var missing *issuefiles.MissingIssueError
	if !errors.As(err, &missing) {
		t.Fatalf("error = %v, want MissingIssueError", err)
	}
	if !strings.Contains(missing.Error(), "pkm-zzz") {
		t.Errorf("MissingIssueError.Error() = %q, want the missing ID", missing.Error())
	}
}

// TestApplyRankPlanRefusesACreationThatCannotBeNamed: the creation's
// randomness failing is a preparation failure — nothing is written.
func TestApplyRankPlanRefusesACreationThatCannotBeNamed(t *testing.T) {
	dir := newVault(t)
	before := rankedIssue("a", 1)
	writeIssue(t, dir, "pkm-aaa", before)

	_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Create: &issuefiles.NewIssue{Prefix: "pkm", Issue: newIssue("new"), Rand: failingReader{}},
	})
	if err == nil {
		t.Fatal("ApplyRankPlan(failing randomness) = nil, want an error")
	}
	if got := readIssueFile(t, dir, "pkm-aaa"); got != before {
		t.Errorf("pkm-aaa was changed by a failed preparation: %q", got)
	}
	assertNoRecoveryArtifacts(t, dir)
}

type failingReader struct{}

func (failingReader) Read(p []byte) (int, error) {
	return 0, errors.New("no randomness")
}

// TestApplyRankPlanRefusesARecoveryPathThatIsNotADirectory: an entry
// planted at .mt-recovery — a file or a symlink — refuses the operation
// before any Issue is written; the link is never followed.
func TestApplyRankPlanRefusesARecoveryPathThatIsNotADirectory(t *testing.T) {
	for name, plant := range map[string]func(t *testing.T, dir string){
		"file": func(t *testing.T, dir string) {
			if err := os.WriteFile(recoveryRoot(dir), []byte("not a dir"), 0o644); err != nil {
				t.Fatal(err)
			}
		},
		"symlink": func(t *testing.T, dir string) {
			link(t, t.TempDir(), recoveryRoot(dir))
		},
	} {
		t.Run(name, func(t *testing.T) {
			dir := newVault(t)
			before := rankedIssue("a", 1)
			writeIssue(t, dir, "pkm-aaa", before)
			plant(t, dir)

			_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
				Changes: []issue.Change{changeRank("pkm-aaa", 2)},
			})
			if err == nil {
				t.Fatal("ApplyRankPlan(broken recovery path) = nil, want an error")
			}
			if got := readIssueFile(t, dir, "pkm-aaa"); got != before {
				t.Errorf("pkm-aaa was changed: %q", got)
			}
		})
	}
}

// TestApplyRankPlanReportsAnUnwritableRecoveryArea: a recovery area the
// process cannot write fails the preparation — no Issue is touched.
func TestApplyRankPlanReportsAnUnwritableRecoveryArea(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permissions do not stop root")
	}
	dir := newVault(t)
	before := rankedIssue("a", 1)
	writeIssue(t, dir, "pkm-aaa", before)
	if err := os.Mkdir(recoveryRoot(dir), 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(recoveryRoot(dir), 0o755) })

	_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-aaa", 2)},
	})
	if err == nil {
		t.Fatal("ApplyRankPlan(read-only recovery area) = nil, want an error")
	}
	if got := readIssueFile(t, dir, "pkm-aaa"); got != before {
		t.Errorf("pkm-aaa was changed: %q", got)
	}
}

// TestApplyRankPlanSiblingsKeepTheirOriginals: an operation that ends
// with pending originals leaves them in place, and a later operation of
// the same Vault removes only its own directory — the pending one stays,
// untouched, for the manual recovery.
func TestApplyRankPlanSiblingsKeepTheirOriginals(t *testing.T) {
	dir := newVault(t)
	a, b, c := rankedIssue("a", 1), rankedIssue("b", 2), backlogIssue("c")
	writeIssue(t, dir, "pkm-aaa", a)
	writeIssue(t, dir, "pkm-bbb", b)
	writeIssue(t, dir, "pkm-ccc", c)
	boom := errors.New("write fails")

	// Operation 1: fails, and its restoration of pkm-aaa also fails —
	// its original stays in .mt-recovery.
	hook := func(stage issuefiles.FaultStage, id string) error {
		switch {
		case stage == issuefiles.StageAfterWrite && id == "pkm-bbb":
			return boom
		case stage == issuefiles.StageRestore && id == "pkm-aaa":
			return errors.New("restoration fails")
		}
		return nil
	}
	cleanup := issuefiles.SetFaultHook(hook)
	defer cleanup()
	if _, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-aaa", 2), changeRank("pkm-bbb", 1)},
	}); !errors.Is(err, boom) {
		t.Fatalf("operation 1 error = %v, want the write failure", err)
	}
	kept, err := os.ReadDir(recoveryRoot(dir))
	if err != nil {
		t.Fatal(err)
	}
	if len(kept) != 1 {
		t.Fatalf("recovery area holds %d operations, want 1", len(kept))
	}
	pendingDir := filepath.Join(recoveryRoot(dir), kept[0].Name())

	// Operation 2: fails before its write, restores cleanly — and its
	// cleanup leaves operation 1's pending originals alone.
	hook2, fired := hookAt(issuefiles.StageBeforeWrite, "pkm-ccc", boom)
	cleanup2 := issuefiles.SetFaultHook(hook2)
	defer cleanup2()

	_, err = issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
		Changes: []issue.Change{changeRank("pkm-ccc", 1)},
	})
	if !fired() {
		t.Error("the second operation's fault hook did not fire")
	}
	var recovery *issuefiles.RecoveryError
	if errors.As(err, &recovery) {
		t.Errorf("operation 2 reported an incomplete recovery: %v", err)
	}
	if err == nil {
		t.Fatal("operation 2 = nil, want the write failure")
	}
	if !strings.Contains(err.Error(), "restored") {
		t.Errorf("operation 2 error = %v, want a complete restoration", err)
	}
	// The pending original of operation 1 is exactly where it was.
	kept2, err := os.ReadDir(pendingDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(kept2) != 1 || kept2[0].Name() != "pkm-aaa.md" {
		t.Errorf("pending originals = %v, want pkm-aaa.md untouched", kept2)
	}
}

// TestApplyRankPlanMessagesDescribeTheState: the errors say what the
// user has on disk — restored, pending, or applied-with-artefacts — and
// never claim success over a state the Vault does not hold.
func TestApplyRankPlanMessagesDescribeTheState(t *testing.T) {
	t.Run("complete restoration", func(t *testing.T) {
		dir := newVault(t)
		writeIssue(t, dir, "pkm-aaa", rankedIssue("a", 1))
		hook, _ := hookAt(issuefiles.StageBeforeWrite, "pkm-aaa", errors.New("boom"))
		cleanup := issuefiles.SetFaultHook(hook)
		defer cleanup()

		_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
			Changes: []issue.Change{changeRank("pkm-aaa", 2)},
		})
		if err == nil || !strings.Contains(err.Error(), "the previous state was restored") {
			t.Errorf("error = %v, want a complete-restoration message", err)
		}
	})
	t.Run("incomplete restoration", func(t *testing.T) {
		dir := newVault(t)
		writeIssue(t, dir, "pkm-aaa", rankedIssue("a", 1))
		writeIssue(t, dir, "pkm-bbb", rankedIssue("b", 2))
		hook := func(stage issuefiles.FaultStage, id string) error {
			switch {
			case stage == issuefiles.StageAfterWrite && id == "pkm-bbb":
				return errors.New("boom")
			case stage == issuefiles.StageRestore && id == "pkm-aaa":
				return errors.New("restore fails")
			}
			return nil
		}
		cleanup := issuefiles.SetFaultHook(hook)
		defer cleanup()

		_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
			Changes: []issue.Change{changeRank("pkm-aaa", 2), changeRank("pkm-bbb", 1)},
		})
		if err == nil {
			t.Fatal("error = nil, want an incomplete recovery")
		}
		for _, want := range []string{"recovery incomplete", "pkm-aaa", ".mt-recovery"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error message misses %q: %v", want, err)
			}
		}
	})
	t.Run("pending removal", func(t *testing.T) {
		dir := newVault(t)
		writeIssue(t, dir, "pkm-aaa", rankedIssue("a", 1))
		// The hook fails the operation right after the creation landed,
		// making issues/ read-only on the way: the removal of the created
		// Issue is then refused, and the recovery stays incomplete with a
		// pending removal — not a pending restoration.
		hook := func(stage issuefiles.FaultStage, id string) error {
			if stage == issuefiles.StageCreated && strings.HasPrefix(id, "pkm-0") {
				if err := os.Chmod(filepath.Join(dir, "issues"), 0o555); err != nil {
					t.Fatal(err)
				}
				return errors.New("boom")
			}
			return nil
		}
		cleanupHook := issuefiles.SetFaultHook(hook)
		defer cleanupHook()
		t.Cleanup(func() { _ = os.Chmod(filepath.Join(dir, "issues"), 0o755) })

		_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
			Create: &issuefiles.NewIssue{Prefix: "pkm", Issue: newIssue("new"), Rand: zeroReader{}},
		})
		var recovery *issuefiles.RecoveryError
		if !errors.As(err, &recovery) {
			t.Fatalf("error = %v, want RecoveryError", err)
		}
		if !strings.Contains(err.Error(), "could not be removed") {
			t.Errorf("error message misses the pending removal: %v", err)
		}
	})
	t.Run("applied but not cleaned", func(t *testing.T) {
		dir := newVault(t)
		writeIssue(t, dir, "pkm-aaa", rankedIssue("a", 1))
		hook, _ := hookAt(issuefiles.StageCleanup, "", errors.New("cleanup fails"))
		cleanup := issuefiles.SetFaultHook(hook)
		defer cleanup()

		_, err := issuefiles.Open(dir).ApplyRankPlan(issuefiles.RankPlan{
			Changes: []issue.Change{changeRank("pkm-aaa", 2)},
		})
		if err == nil || !strings.Contains(err.Error(), "applied") {
			t.Errorf("error = %v, want an applied-but-not-cleaned message", err)
		}
	})
}
