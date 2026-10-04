// Package issue_test holds the black-box unit tests of the domain rules
// of the Fila (Seam 2): the single order rule and the Rank
// renormalization, both over the single line of an Issue (issue.Item).
package issue_test

import (
	"slices"
	"strconv"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// queueItem builds an Item with the fields the Fila order reads: the ID,
// the Rank (nil means the Backlog) and created_at.
func queueItem(id string, rank *int, createdAt string) issue.Item {
	return issue.Item{
		ID: id,
		Issue: issue.Issue{
			Frontmatter: issue.Frontmatter{
				Title:     id,
				Status:    "open",
				CreatedAt: createdAt,
				Rank:      rank,
			},
		},
	}
}

func TestCompareOrdersTheFilaWithTheCmpCompareConvention(t *testing.T) {
	cases := []struct {
		name string
		a, b issue.Item
		want int
	}{
		{"lower rank first", queueItem("a", intPtr(1), ""), queueItem("b", intPtr(2), ""), -1},
		{"higher rank last", queueItem("b", intPtr(2), ""), queueItem("a", intPtr(1), ""), 1},
		{"rank beats a lower ID", queueItem("z", intPtr(1), ""), queueItem("a", intPtr(2), ""), -1},
		{"equal rank tiebreak by ID", queueItem("a", intPtr(1), ""), queueItem("b", intPtr(1), ""), -1},
		{"equal rank tiebreak by ID, reversed", queueItem("b", intPtr(1), ""), queueItem("a", intPtr(1), ""), 1},
		{"equal rank and ID are equal", queueItem("a", intPtr(1), ""), queueItem("a", intPtr(1), ""), 0},
		{"ranked before the Backlog even when it has a higher rank", queueItem("a", intPtr(9), ""), queueItem("b", nil, "2026-01-01T09:00"), -1},
		{"Backlog after a ranked Item even when it is older", queueItem("a", nil, "2026-01-01T09:00"), queueItem("b", intPtr(9), ""), 1},
		{"Backlog oldest created_at first", queueItem("a", nil, "2026-08-15T09:30"), queueItem("b", nil, "2026-08-16T09:30"), -1},
		{"Backlog newest created_at last", queueItem("b", nil, "2026-08-16T09:30"), queueItem("a", nil, "2026-08-15T09:30"), 1},
		{"Backlog created_at beats a lower ID", queueItem("z", nil, "2026-08-15T09:30"), queueItem("a", nil, "2026-08-16T09:30"), -1},
		{"Backlog equal created_at tiebreak by ID", queueItem("a", nil, "2026-08-15T09:30"), queueItem("b", nil, "2026-08-15T09:30"), -1},
		{"Backlog equal created_at tiebreak by ID, reversed", queueItem("b", nil, "2026-08-15T09:30"), queueItem("a", nil, "2026-08-15T09:30"), 1},
		{"Backlog equal everything is equal", queueItem("a", nil, "2026-08-15T09:30"), queueItem("a", nil, "2026-08-15T09:30"), 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := issue.Compare(c.a, c.b); got != c.want {
				t.Errorf("Compare(%s, %s) = %d, want %d", c.a.ID, c.b.ID, got, c.want)
			}
		})
	}
}

func TestComparePlugsIntoSortFuncAndYieldsTheFila(t *testing.T) {
	items := []issue.Item{
		queueItem("pkm-003", nil, "2026-08-16T09:30"),
		queueItem("pkm-002", intPtr(2), ""),
		queueItem("pkm-004", nil, "2026-08-15T09:30"),
		queueItem("pkm-001", intPtr(1), ""),
		queueItem("pkm-005", intPtr(2), ""),
	}
	slices.SortFunc(items, issue.Compare)
	want := []string{"pkm-001", "pkm-002", "pkm-005", "pkm-004", "pkm-003"}
	got := make([]string, len(items))
	for i, it := range items {
		got[i] = it.ID
	}
	if !slices.Equal(got, want) {
		t.Errorf("Fila = %v, want %v", got, want)
	}
}

// queueChanges asserts the changes against literal expectations, both the
// order of the slice and each target Rank (nil means the Backlog).
func queueChanges(t *testing.T, got, want []issue.Change) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("RenormalizeRanks() = %d changes (%+v), want %d (%+v)", len(got), got, len(want), want)
	}
	for i := range want {
		sameRank := (got[i].Rank == nil) == (want[i].Rank == nil) &&
			(got[i].Rank == nil || *got[i].Rank == *want[i].Rank)
		if got[i].ID != want[i].ID || !sameRank {
			t.Errorf("changes[%d] = {%s, %v}, want {%s, %v}", i, got[i].ID, rankString(got[i].Rank), want[i].ID, rankString(want[i].Rank))
		}
	}
}

func rankString(rank *int) string {
	if rank == nil {
		return "Backlog"
	}
	return strconv.Itoa(*rank)
}

func TestRenormalizeRanksLeavesAContiguousFilaUntouched(t *testing.T) {
	items := []issue.Item{
		queueItem("pkm-003", intPtr(3), ""),
		queueItem("pkm-001", intPtr(1), ""),
		queueItem("pkm-002", intPtr(2), ""),
	}
	queueChanges(t, issue.RenormalizeRanks(items), nil)
}

func TestRenormalizeRanksClosesGapsInFilaOrder(t *testing.T) {
	items := []issue.Item{
		queueItem("pkm-003", intPtr(6), ""),
		queueItem("pkm-001", intPtr(4), ""),
		queueItem("pkm-002", intPtr(5), ""),
	}
	want := []issue.Change{
		{ID: "pkm-001", Rank: intPtr(1)},
		{ID: "pkm-002", Rank: intPtr(2)},
		{ID: "pkm-003", Rank: intPtr(3)},
	}
	queueChanges(t, issue.RenormalizeRanks(items), want)
}

func TestRenormalizeRanksKeepsItemsAlreadyAtTheirTargetRank(t *testing.T) {
	items := []issue.Item{
		queueItem("pkm-001", intPtr(1), ""),
		queueItem("pkm-002", intPtr(3), ""),
		queueItem("pkm-003", intPtr(4), ""),
	}
	want := []issue.Change{
		{ID: "pkm-002", Rank: intPtr(2)},
		{ID: "pkm-003", Rank: intPtr(3)},
	}
	queueChanges(t, issue.RenormalizeRanks(items), want)
}

func TestRenormalizeRanksBreaksDuplicateRanksByID(t *testing.T) {
	items := []issue.Item{
		queueItem("pkm-bbb", intPtr(1), "2026-08-15T09:30"),
		queueItem("pkm-aaa", intPtr(1), "2026-08-16T09:30"),
	}
	want := []issue.Change{{ID: "pkm-bbb", Rank: intPtr(2)}}
	queueChanges(t, issue.RenormalizeRanks(items), want)
}

func TestRenormalizeRanksLeavesTheBacklogAlone(t *testing.T) {
	items := []issue.Item{
		queueItem("pkm-002", nil, "2026-08-15T09:30"),
		queueItem("pkm-003", intPtr(2), ""),
		queueItem("pkm-001", nil, "2026-08-16T09:30"),
	}
	want := []issue.Change{{ID: "pkm-003", Rank: intPtr(1)}}
	queueChanges(t, issue.RenormalizeRanks(items), want)
}

func TestRenormalizeRanksOfAnEmptyOrBacklogOnlyVault(t *testing.T) {
	queueChanges(t, issue.RenormalizeRanks(nil), nil)
	backlog := []issue.Item{queueItem("pkm-001", nil, "2026-08-15T09:30")}
	queueChanges(t, issue.RenormalizeRanks(backlog), nil)
}

func TestRenormalizeRanksIncludesEveryStatus(t *testing.T) {
	done := queueItem("pkm-done", intPtr(9), "")
	done.Issue.Frontmatter.Status = "done"
	custom := queueItem("pkm-custom", intPtr(5), "")
	custom.Issue.Frontmatter.Status = "waiting"
	open := queueItem("pkm-open", intPtr(7), "")
	backlog := queueItem("pkm-backlog", nil, "2026-08-15T09:30")
	items := []issue.Item{backlog, done, custom, open}
	want := []issue.Change{
		{ID: "pkm-custom", Rank: intPtr(1)},
		{ID: "pkm-open", Rank: intPtr(2)},
		{ID: "pkm-done", Rank: intPtr(3)},
	}
	queueChanges(t, issue.RenormalizeRanks(items), want)
}

func TestRenormalizeRanksDoesNotReorderItsInput(t *testing.T) {
	items := []issue.Item{
		queueItem("pkm-003", intPtr(6), ""),
		queueItem("pkm-001", intPtr(4), ""),
		queueItem("pkm-002", nil, "2026-08-15T09:30"),
	}
	issue.RenormalizeRanks(items)
	want := []string{"pkm-003", "pkm-001", "pkm-002"}
	got := make([]string, len(items))
	for i, it := range items {
		got[i] = it.ID
	}
	if !slices.Equal(got, want) {
		t.Errorf("input order after RenormalizeRanks = %v, want %v", got, want)
	}
}
