package priority_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issue"
	"github.com/Sanmoo/my-tasks2/internal/priority"
)

// ptr returns a pointer to n, for building Rank values in tests.
func ptr(n int) *int { return &n }

// item builds the domain's single Issue line with the fields the ordering
// mechanisms read: status, Rank and created_at.
func item(id, status string, rank *int, createdAt string) issue.Item {
	return issue.Item{ID: id, Issue: issue.Issue{Frontmatter: issue.Frontmatter{
		Status:    status,
		Rank:      rank,
		CreatedAt: createdAt,
	}}}
}

// titled is item plus the Title the prioritize buffer prints.
func titled(id, title, status string, rank *int, createdAt string) issue.Item {
	it := item(id, status, rank, createdAt)
	it.Issue.Frontmatter.Title = title
	return it
}

func TestPrioritizable(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"open", true},
		{"in_progress", true},
		{"done", false},
		{"blocked", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := priority.Prioritizable(tt.status); got != tt.want {
			t.Errorf("Prioritizable(%q) = %v, want %v", tt.status, got, tt.want)
		}
	}
}

func TestBufferOrderAndFormat(t *testing.T) {
	items := []issue.Item{
		titled("b", "Backlog B", "open", nil, "2026-08-16T10:00"),
		titled("r2", "Rank 2", "in_progress", ptr(2), "2026-08-14T10:00"),
		titled("a", "Backlog A", "open", nil, "2026-08-15T10:00"),
		titled("r1", "Rank 1", "open", ptr(1), "2026-08-13T10:00"),
	}
	got := priority.Buffer(items)

	wantLines := []string{
		"# Edit ranking for this vault",
		"# Reorder lines. Use [P] for prioritized, [ ] for backlog.",
		"# Do not edit issue IDs. Save and close to continue.",
		"",
		"[P] r1  Rank 1",
		"[P] r2  Rank 2",
		"[ ] a  Backlog A",
		"[ ] b  Backlog B",
	}
	if got != strings.Join(wantLines, "\n")+"\n" {
		t.Errorf("Buffer() output mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, strings.Join(wantLines, "\n")+"\n")
	}
}

func TestBufferDoesNotMutateInput(t *testing.T) {
	items := []issue.Item{
		item("a", "open", ptr(2), "2026-08-15T10:00"),
		item("b", "open", ptr(1), "2026-08-16T10:00"),
	}
	priority.Buffer(items)
	if items[0].ID != "a" || items[1].ID != "b" {
		t.Errorf("Buffer() mutated its argument: %+v", items)
	}
}

func TestBufferBacklogTiebreakByID(t *testing.T) {
	items := []issue.Item{
		titled("z", "z", "open", nil, "2026-08-15T10:00"),
		titled("a", "a", "open", nil, "2026-08-15T10:00"),
	}
	got := priority.Buffer(items)
	if !strings.Contains(got, "[ ] a  a\n[ ] z  z\n") {
		t.Errorf("backlog tiebreak by ID failed:\n%s", got)
	}
}

func TestParse(t *testing.T) {
	text := "# a comment\n" +
		"\n" +
		"  # an indented comment\n" +
		"[P] r1  Rank 1\n" +
		"[ ] b1  Backlog 1\n" +
		"[P] r2\n"
	entries, err := priority.Parse(text)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	want := []priority.Entry{
		{Prioritized: true, ID: "r1"},
		{Prioritized: false, ID: "b1"},
		{Prioritized: true, ID: "r2"},
	}
	if len(entries) != len(want) {
		t.Fatalf("Parse() = %d entries, want %d: %+v", len(entries), len(want), entries)
	}
	for i := range want {
		if entries[i] != want[i] {
			t.Errorf("entries[%d] = %+v, want %+v", i, entries[i], want[i])
		}
	}
}

func TestParseToleratesCRLF(t *testing.T) {
	entries, err := priority.Parse("[P] a  title\r\n[ ] b  title\r\n")
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(entries) != 2 || entries[0].ID != "a" || entries[1].ID != "b" {
		t.Errorf("Parse() = %+v", entries)
	}
}

func TestParseIgnoresTitle(t *testing.T) {
	// The title may contain anything (spaces, brackets, #); only the ID
	// matters, and it is the first whitespace-delimited token.
	entries, err := priority.Parse("[P] pkm-055  some [P] title # not a comment\n")
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(entries) != 1 || entries[0].ID != "pkm-055" || !entries[0].Prioritized {
		t.Errorf("Parse() = %+v", entries)
	}
}

func TestParseRejectsMalformedLine(t *testing.T) {
	for _, line := range []string{
		"[Q] a  title",
		"[P]a  title",
		"a  title",
		"[P]",
		"[ ]",
		"[P] ",
	} {
		if _, err := priority.Parse(line + "\n"); err == nil {
			t.Errorf("Parse(%q) = no error, want error", line)
		}
	}
}

func TestParseRejectsDuplicateID(t *testing.T) {
	_, err := priority.Parse("[P] a  one\n[ ] a  two\n")
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("Parse() error = %v, want duplicate error", err)
	}
}

func TestPlanReorderAndRenormalize(t *testing.T) {
	items := []issue.Item{
		item("a", "open", ptr(1), "2026-08-15T10:00"),
		item("b", "open", ptr(2), "2026-08-16T10:00"),
		item("c", "open", nil, "2026-08-17T10:00"), // Backlog
	}
	entries := []priority.Entry{
		{Prioritized: true, ID: "b"},
		{Prioritized: true, ID: "a"},
		{Prioritized: false, ID: "c"},
	}
	changes, err := priority.Plan(entries, items)
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	// b: 2 → 1, a: 1 → 2, c: Backlog → Backlog (no change).
	want := []issue.Change{
		{ID: "b", Rank: ptr(1)},
		{ID: "a", Rank: ptr(2)},
	}
	if len(changes) != len(want) {
		t.Fatalf("Plan() = %d changes, want %d: %+v", len(changes), len(want), changes)
	}
	for i := range want {
		if changes[i].ID != want[i].ID || !rankEqual(changes[i].Rank, want[i].Rank) {
			t.Errorf("changes[%d] = %+v, want %+v", i, changes[i], want[i])
		}
	}
}

func TestPlanPromote(t *testing.T) {
	items := []issue.Item{
		item("a", "open", ptr(1), "2026-08-15T10:00"),
		item("b", "in_progress", nil, "2026-08-16T10:00"), // Backlog
	}
	entries := []priority.Entry{
		{Prioritized: true, ID: "a"},
		{Prioritized: true, ID: "b"},
	}
	changes, err := priority.Plan(entries, items)
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	// a stays rank 1 (no change); b is promoted to rank 2.
	if len(changes) != 1 || changes[0].ID != "b" || !rankEqual(changes[0].Rank, ptr(2)) {
		t.Errorf("Plan() = %+v, want only b→2", changes)
	}
}

func TestPlanDemote(t *testing.T) {
	items := []issue.Item{
		item("a", "open", ptr(1), "2026-08-15T10:00"),
		item("b", "open", ptr(2), "2026-08-16T10:00"),
	}
	entries := []priority.Entry{
		{Prioritized: true, ID: "a"},
		{Prioritized: false, ID: "b"},
	}
	changes, err := priority.Plan(entries, items)
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	// a stays rank 1; b is demoted to Backlog (nil rank).
	if len(changes) != 1 || changes[0].ID != "b" || changes[0].Rank != nil {
		t.Errorf("Plan() = %+v, want only b→nil", changes)
	}
}

func TestPlanZeroChurn(t *testing.T) {
	items := []issue.Item{
		item("a", "open", ptr(1), "2026-08-15T10:00"),
		item("b", "open", ptr(2), "2026-08-16T10:00"),
	}
	entries := []priority.Entry{
		{Prioritized: true, ID: "a"},
		{Prioritized: true, ID: "b"},
	}
	changes, err := priority.Plan(entries, items)
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	if len(changes) != 0 {
		t.Errorf("Plan() = %+v, want no changes (zero churn)", changes)
	}
}

func TestPlanRenormalizesGaps(t *testing.T) {
	items := []issue.Item{
		item("a", "open", ptr(5), "2026-08-15T10:00"),
		item("b", "open", ptr(9), "2026-08-16T10:00"),
	}
	entries := []priority.Entry{
		{Prioritized: true, ID: "a"},
		{Prioritized: true, ID: "b"},
	}
	changes, err := priority.Plan(entries, items)
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("Plan() = %d changes, want 2 (both renormalized)", len(changes))
	}
	if changes[0].ID != "a" || !rankEqual(changes[0].Rank, ptr(1)) {
		t.Errorf("changes[0] = %+v, want a→1", changes[0])
	}
	if changes[1].ID != "b" || !rankEqual(changes[1].Rank, ptr(2)) {
		t.Errorf("changes[1] = %+v, want b→2", changes[1])
	}
}

func TestPlanRanksIgnoreBacklogLines(t *testing.T) {
	// A [ ] line between [P] lines does not consume a rank: ranks are
	// the positions among [P] entries only, contiguous 1..N.
	items := []issue.Item{
		item("a", "open", nil, "2026-08-15T10:00"),
		item("b", "open", nil, "2026-08-16T10:00"),
		item("c", "open", nil, "2026-08-17T10:00"),
	}
	entries := []priority.Entry{
		{Prioritized: true, ID: "a"},
		{Prioritized: false, ID: "b"},
		{Prioritized: true, ID: "c"},
	}
	changes, err := priority.Plan(entries, items)
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	// a→1, b→Backlog (no change, was Backlog), c→2.
	want := []issue.Change{
		{ID: "a", Rank: ptr(1)},
		{ID: "c", Rank: ptr(2)},
	}
	if len(changes) != len(want) {
		t.Fatalf("Plan() = %d changes, want %d: %+v", len(changes), len(want), changes)
	}
	for i := range want {
		if changes[i].ID != want[i].ID || !rankEqual(changes[i].Rank, want[i].Rank) {
			t.Errorf("changes[%d] = %+v, want %+v", i, changes[i], want[i])
		}
	}
}

func TestPlanRejectsUnknownID(t *testing.T) {
	items := []issue.Item{item("a", "open", nil, "2026-08-15T10:00")}
	entries := []priority.Entry{{Prioritized: true, ID: "ghost"}}
	if _, err := priority.Plan(entries, items); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Errorf("Plan() error = %v, want unknown ID error", err)
	}
}

func TestPlanRejectsDoneIssue(t *testing.T) {
	items := []issue.Item{
		item("a", "done", ptr(1), "2026-08-15T10:00"),
	}
	entries := []priority.Entry{{Prioritized: true, ID: "a"}}
	if _, err := priority.Plan(entries, items); err == nil || !strings.Contains(err.Error(), "cannot be prioritized") {
		t.Errorf("Plan() error = %v, want cannot-be-prioritized error", err)
	}
}

func TestPlanRejectsMissingIssue(t *testing.T) {
	items := []issue.Item{
		item("a", "open", ptr(1), "2026-08-15T10:00"),
		item("b", "open", nil, "2026-08-16T10:00"),
	}
	// b (open) is missing from the buffer.
	entries := []priority.Entry{{Prioritized: true, ID: "a"}}
	if _, err := priority.Plan(entries, items); err == nil || !strings.Contains(err.Error(), "missing") {
		t.Errorf("Plan() error = %v, want missing-issue error", err)
	}
}

func TestPlanIgnoresNonPrioritizableInItems(t *testing.T) {
	// done and custom-status issues are not in the buffer and must not
	// trigger the missing-issue check.
	items := []issue.Item{
		item("a", "open", ptr(1), "2026-08-15T10:00"),
		item("d", "done", ptr(2), "2026-08-14T10:00"),
		item("x", "blocked", nil, "2026-08-13T10:00"),
	}
	entries := []priority.Entry{{Prioritized: true, ID: "a"}}
	changes, err := priority.Plan(entries, items)
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	if len(changes) != 0 {
		t.Errorf("Plan() = %+v, want no changes", changes)
	}
}

func TestPlanRejectsDuplicateDirectly(t *testing.T) {
	// Plan defends against duplicates even when called without Parse.
	items := []issue.Item{item("a", "open", nil, "2026-08-15T10:00")}
	entries := []priority.Entry{
		{Prioritized: true, ID: "a"},
		{Prioritized: true, ID: "a"},
	}
	if _, err := priority.Plan(entries, items); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("Plan() error = %v, want duplicate error", err)
	}
}

func TestQuickPlanTopPromotesAndShifts(t *testing.T) {
	items := []issue.Item{
		item("a", "open", ptr(1), "2026-08-15T10:00"),
		item("b", "open", ptr(2), "2026-08-16T10:00"),
		item("c", "open", nil, "2026-08-17T10:00"),
	}
	changes, err := priority.QuickPlan(items, "c", priority.MoveTop, 0)
	if err != nil {
		t.Fatalf("QuickPlan() error: %v", err)
	}
	want := []issue.Change{
		{ID: "c", Rank: ptr(1)},
		{ID: "a", Rank: ptr(2)},
		{ID: "b", Rank: ptr(3)},
	}
	if len(changes) != len(want) {
		t.Fatalf("QuickPlan() = %d changes, want %d: %+v", len(changes), len(want), changes)
	}
	for i := range want {
		if changes[i].ID != want[i].ID || !rankEqual(changes[i].Rank, want[i].Rank) {
			t.Errorf("changes[%d] = %+v, want %+v", i, changes[i], want[i])
		}
	}
}

func TestQuickPlanBottomReordersQueue(t *testing.T) {
	items := []issue.Item{
		item("a", "open", ptr(1), "2026-08-15T10:00"),
		item("b", "open", ptr(2), "2026-08-16T10:00"),
		item("c", "open", nil, "2026-08-17T10:00"),
	}
	changes, err := priority.QuickPlan(items, "a", priority.MoveBottom, 0)
	if err != nil {
		t.Fatalf("QuickPlan() error: %v", err)
	}
	want := []issue.Change{
		{ID: "b", Rank: ptr(1)},
		{ID: "a", Rank: ptr(2)},
	}
	if len(changes) != len(want) {
		t.Fatalf("QuickPlan() = %d changes, want %d: %+v", len(changes), len(want), changes)
	}
	for i := range want {
		if changes[i].ID != want[i].ID || !rankEqual(changes[i].Rank, want[i].Rank) {
			t.Errorf("changes[%d] = %+v, want %+v", i, changes[i], want[i])
		}
	}
}

func TestQuickPlanRankInsertsBacklogItem(t *testing.T) {
	items := []issue.Item{
		item("a", "open", ptr(1), "2026-08-15T10:00"),
		item("b", "open", ptr(2), "2026-08-16T10:00"),
		item("c", "open", ptr(3), "2026-08-17T10:00"),
		item("d", "open", nil, "2026-08-18T10:00"),
	}
	changes, err := priority.QuickPlan(items, "d", priority.MoveToRank, 2)
	if err != nil {
		t.Fatalf("QuickPlan() error: %v", err)
	}
	want := []issue.Change{
		{ID: "d", Rank: ptr(2)},
		{ID: "b", Rank: ptr(3)},
		{ID: "c", Rank: ptr(4)},
	}
	if len(changes) != len(want) {
		t.Fatalf("QuickPlan() = %d changes, want %d: %+v", len(changes), len(want), changes)
	}
	for i := range want {
		if changes[i].ID != want[i].ID || !rankEqual(changes[i].Rank, want[i].Rank) {
			t.Errorf("changes[%d] = %+v, want %+v", i, changes[i], want[i])
		}
	}
}

func TestQuickPlanUnrankShiftsRemainingQueue(t *testing.T) {
	items := []issue.Item{
		item("a", "open", ptr(1), "2026-08-15T10:00"),
		item("b", "open", ptr(2), "2026-08-16T10:00"),
		item("c", "open", ptr(3), "2026-08-17T10:00"),
		item("d", "open", nil, "2026-08-18T10:00"),
	}
	changes, err := priority.QuickPlan(items, "b", priority.RemoveRank, 0)
	if err != nil {
		t.Fatalf("QuickPlan() error: %v", err)
	}
	want := []issue.Change{
		{ID: "c", Rank: ptr(2)},
		{ID: "b", Rank: nil},
	}
	if len(changes) != len(want) {
		t.Fatalf("QuickPlan() = %d changes, want %d: %+v", len(changes), len(want), changes)
	}
	for i := range want {
		if changes[i].ID != want[i].ID || !rankEqual(changes[i].Rank, want[i].Rank) {
			t.Errorf("changes[%d] = %+v, want %+v", i, changes[i], want[i])
		}
	}
}

func TestQuickPlanRejectsInvalidRequest(t *testing.T) {
	items := []issue.Item{
		item("a", "open", ptr(1), "2026-08-15T10:00"),
		item("done", "done", ptr(2), "2026-08-16T10:00"),
	}
	tests := []struct {
		name   string
		id     string
		action priority.QuickAction
		pos    int
		want   string
	}{
		{name: "unknown issue", id: "ghost", action: priority.MoveTop, want: "unknown issue ID"},
		{name: "done issue", id: "done", action: priority.MoveTop, want: "cannot be prioritized"},
		{name: "zero rank", id: "a", action: priority.MoveToRank, pos: 0, want: "between 1 and 1"},
		{name: "rank beyond queue", id: "a", action: priority.MoveToRank, pos: 2, want: "between 1 and 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := priority.QuickPlan(items, tt.id, tt.action, tt.pos); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("QuickPlan() error = %v, want %q", err, tt.want)
			}
		})
	}
}

// rankEqual compares two *int ranks, treating nil as a value equal only
// to nil. It mirrors the semantics under test.
func rankEqual(a, b *int) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func TestPlanQueueFirstThenNonPrioritizableRanks(t *testing.T) {
	// A done or custom-status Issue holding a rank inside the queue range
	// must not collide with the queue renormalization: the queue keeps
	// 1..N and the non-prioritizable ranked Issues follow at N+1..M in
	// their existing rank order.
	items := []issue.Item{
		item("d", "done", ptr(1), "2026-08-14T10:00"),
		item("x", "blocked", ptr(2), "2026-08-13T10:00"),
		item("a", "open", ptr(3), "2026-08-15T10:00"),
		item("b", "open", ptr(4), "2026-08-16T10:00"),
	}
	entries := []priority.Entry{
		{Prioritized: true, ID: "a"},
		{Prioritized: true, ID: "b"},
	}
	changes, err := priority.Plan(entries, items)
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	want := []issue.Change{
		{ID: "a", Rank: ptr(1)},
		{ID: "b", Rank: ptr(2)},
		{ID: "d", Rank: ptr(3)},
		{ID: "x", Rank: ptr(4)},
	}
	if len(changes) != len(want) {
		t.Fatalf("Plan() = %d changes, want %d: %+v", len(changes), len(want), changes)
	}
	for i := range want {
		if changes[i].ID != want[i].ID || !rankEqual(changes[i].Rank, want[i].Rank) {
			t.Errorf("changes[%d] = %+v, want %+v", i, changes[i], want[i])
		}
	}
	// Applying the changes must leave every rank unique across the vault.
	if dups := duplicateRanksAfter(items, changes); len(dups) > 0 {
		t.Errorf("Plan() changes produce duplicate ranks %v: %+v", dups, changes)
	}
}

func TestPlanZeroChurnWithNonPrioritizableRankAfterQueue(t *testing.T) {
	// done holding rank 3 already sits after the 2-issue queue: nothing
	// may be rewritten, the done file stays untouched.
	items := []issue.Item{
		item("a", "open", ptr(1), "2026-08-15T10:00"),
		item("b", "in_progress", ptr(2), "2026-08-16T10:00"),
		item("d", "done", ptr(3), "2026-08-14T10:00"),
	}
	entries := []priority.Entry{
		{Prioritized: true, ID: "a"},
		{Prioritized: true, ID: "b"},
	}
	changes, err := priority.Plan(entries, items)
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}
	if len(changes) != 0 {
		t.Errorf("Plan() = %+v, want no changes (zero churn)", changes)
	}
}

// duplicateRanksAfter applies the changes to a copy of items and returns
// the rank values held by more than one Issue — the vault-wide uniqueness
// check mirroring internal/list.DuplicateRanks.
func duplicateRanksAfter(items []issue.Item, changes []issue.Change) []int {
	final := make(map[string]*int, len(items))
	for _, it := range items {
		final[it.ID] = it.Issue.Frontmatter.Rank
	}
	for _, ch := range changes {
		final[ch.ID] = ch.Rank
	}
	counts := make(map[int]int)
	for _, r := range final {
		if r != nil {
			counts[*r]++
		}
	}
	dups := make([]int, 0)
	for r, n := range counts {
		if n > 1 {
			dups = append(dups, r)
		}
	}
	slices.Sort(dups)
	return dups
}
