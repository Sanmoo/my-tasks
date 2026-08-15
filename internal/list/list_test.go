// Package list_test holds the black-box unit tests of the mt list pure
// logic (Seam 2): the priority ordering comparator, the per-status
// glyphs, the deferred-until rules, and duplicate-rank detection.
package list_test

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Sanmoo/my-tasks2/internal/issue"
	"github.com/Sanmoo/my-tasks2/internal/list"
)

func intPtr(v int) *int { return &v }

func item(id, status string, rank *int, createdAt, deferredUntil string) list.Item {
	return list.Item{
		ID: id,
		Issue: issue.Issue{
			Frontmatter: issue.Frontmatter{
				Status:        status,
				Rank:          rank,
				CreatedAt:     createdAt,
				DeferredUntil: deferredUntil,
			},
		},
	}
}

// labeled builds an item with labels (and a canonical created_at) for
// the visibility tests.
func labeled(id, status string, labels []string, deferredUntil string) list.Item {
	return list.Item{
		ID: id,
		Issue: issue.Issue{
			Frontmatter: issue.Frontmatter{
				Status:        status,
				Labels:        labels,
				CreatedAt:     "2026-08-15T09:30",
				DeferredUntil: deferredUntil,
			},
		},
	}
}

func ids(items []list.Item) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.ID
	}
	return out
}

func TestCompare(t *testing.T) {
	rank1 := intPtr(1)
	rank2 := intPtr(2)
	cases := []struct {
		name string
		a, b list.Item
		want int
	}{
		{"lower rank first", item("a", "open", rank1, "", ""), item("b", "open", rank2, "", ""), -1},
		{"higher rank last", item("b", "open", rank2, "", ""), item("a", "open", rank1, "", ""), 1},
		{"ranked before backlog", item("a", "open", rank1, "", ""), item("b", "open", nil, "", ""), -1},
		{"backlog after ranked", item("b", "open", nil, "", ""), item("a", "open", rank1, "", ""), 1},
		{"equal rank tiebreak by id", item("b", "open", rank1, "", ""), item("a", "open", rank1, "", ""), 1},
		{"equal rank equal id is equal", item("a", "open", rank1, "", ""), item("a", "open", rank1, "", ""), 0},
		{"backlog older created_at first", item("a", "open", nil, "2026-08-15T09:30", ""), item("b", "open", nil, "2026-08-16T09:30", ""), -1},
		{"backlog newer created_at last", item("b", "open", nil, "2026-08-16T09:30", ""), item("a", "open", nil, "2026-08-15T09:30", ""), 1},
		{"backlog equal created_at tiebreak by id", item("b", "open", nil, "2026-08-15T09:30", ""), item("a", "open", nil, "2026-08-15T09:30", ""), 1},
		{"backlog equal everything is equal", item("a", "open", nil, "2026-08-15T09:30", ""), item("a", "open", nil, "2026-08-15T09:30", ""), 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := list.Compare(c.a, c.b); got != c.want {
				t.Errorf("Compare(%s, %s) = %d, want %d", c.a.ID, c.b.ID, got, c.want)
			}
		})
	}
}

func TestSortOrdersByRankThenBacklogCreatedAtThenID(t *testing.T) {
	r1, r2 := intPtr(1), intPtr(2)
	items := []list.Item{
		item("pkm-003", "open", nil, "2026-08-16T09:30", ""),
		item("pkm-002", "open", r2, "", ""),
		item("pkm-004", "open", nil, "2026-08-15T09:30", ""),
		item("pkm-001", "open", r1, "", ""),
	}
	list.Sort(items)
	want := []string{"pkm-001", "pkm-002", "pkm-004", "pkm-003"}
	if got := ids(items); !slices.Equal(got, want) {
		t.Errorf("Sort order = %v, want %v", got, want)
	}
}

func TestSortDoesNotReorderAnEmptyOrSingleList(t *testing.T) {
	list.Sort(nil)
	single := []list.Item{item("pkm-001", "open", intPtr(1), "", "")}
	list.Sort(single)
	if got := ids(single); !slices.Equal(got, []string{"pkm-001"}) {
		t.Errorf("Sort(single) = %v, want unchanged", got)
	}
}

func TestSortIsStableForItemsThatCompareEqual(t *testing.T) {
	// Two items that Compare equal (same rank, created_at and ID) must
	// keep their relative order under a stable sort. Their titles differ
	// so the test can tell them apart.
	mk := func(title string) list.Item {
		return list.Item{
			ID: "same-id",
			Issue: issue.Issue{
				Frontmatter: issue.Frontmatter{
					Title:     title,
					Status:    "open",
					CreatedAt: "2026-08-15T09:30",
				},
			},
		}
	}
	items := []list.Item{mk("first"), mk("second")}
	list.Sort(items)
	if items[0].Issue.Frontmatter.Title != "first" || items[1].Issue.Frontmatter.Title != "second" {
		t.Errorf("Sort reordered equal items: %q, %q", items[0].Issue.Frontmatter.Title, items[1].Issue.Frontmatter.Title)
	}
}

func TestGlyph(t *testing.T) {
	cases := []struct {
		status string
		want   string
	}{
		{"open", "○"},
		{"in_progress", "◐"},
		{"done", "●"},
		{"blocked", "?"}, // custom status falls back
		{"", "?"},        // unknown/empty also falls back
		{"OPEN", "?"},    // statuses are case-sensitive
	}
	for _, c := range cases {
		t.Run(c.status+"_"+c.want, func(t *testing.T) {
			if got := list.Glyph(c.status); got != c.want {
				t.Errorf("Glyph(%q) = %q, want %q", c.status, got, c.want)
			}
		})
	}
}

func TestIsFutureDeferred(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.Local)
	cases := []struct {
		name     string
		deferred string
		want     bool
	}{
		{"empty is not deferred", "", false},
		{"malformed is not deferred", "not-a-date", false},
		{"past is not deferred", "2026-08-10T08:00", false},
		{"exactly now is available", "2026-08-15T12:00", false},
		{"future is deferred", "2026-08-20T08:00", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := list.IsFutureDeferred(c.deferred, now); got != c.want {
				t.Errorf("IsFutureDeferred(%q, now) = %t, want %t", c.deferred, got, c.want)
			}
		})
	}
}

func TestReady(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.Local)
	cases := []struct {
		name string
		it   list.Item
		want bool
	}{
		{"open without deferral is ready", item("open", "open", nil, "", ""), true},
		{"open with past deferral is ready", item("past", "open", nil, "", "2026-08-10T08:00"), true},
		{"open at deferred time is ready", item("now", "open", nil, "", "2026-08-15T12:00"), true},
		{"future-deferred open is not ready", item("future", "open", nil, "", "2026-08-20T08:00"), false},
		{"in-progress is not ready", item("progress", "in_progress", nil, "", ""), false},
		{"done is not ready", item("done", "done", nil, "", ""), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := list.Ready(c.it, now); got != c.want {
				t.Errorf("Ready(%s) = %t, want %t", c.name, got, c.want)
			}
		})
	}
}

func TestOverdue(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.Local)
	withDeadline := func(status, deadline string) list.Item {
		it := item("issue", status, nil, "", "")
		it.Issue.Frontmatter.Deadline = deadline
		return it
	}
	cases := []struct {
		name string
		it   list.Item
		want bool
	}{
		{"past deadline is overdue", withDeadline("open", "2026-08-10T08:00"), true},
		{"future deadline is not overdue", withDeadline("open", "2026-08-20T08:00"), false},
		{"deadline exactly now is not overdue", withDeadline("open", "2026-08-15T12:00"), false},
		{"missing deadline is not overdue", withDeadline("open", ""), false},
		{"malformed deadline is not overdue", withDeadline("open", "not-a-date"), false},
		{"done issue is not overdue", withDeadline("done", "2026-08-10T08:00"), false},
		{"deferred issue can be overdue", func() list.Item {
			it := withDeadline("open", "2026-08-10T08:00")
			it.Issue.Frontmatter.DeferredUntil = "2026-08-20T08:00"
			return it
		}(), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := list.Overdue(c.it, now); got != c.want {
				t.Errorf("Overdue(%s) = %t, want %t", c.name, got, c.want)
			}
		})
	}
}

func TestDeferSuffix(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.Local)
	cases := []struct {
		name     string
		deferred string
		want     string
	}{
		{"empty has no suffix", "", ""},
		{"malformed has no suffix", "garbage", ""},
		{"past has no suffix", "2026-08-10T08:00", ""},
		{"exactly now has no suffix", "2026-08-15T12:00", ""},
		{"future gets a defer suffix", "2026-08-20T08:00", "[defer 08-20 08:00]"},
		{"future suffix zero-pads month day hour minute", "2027-03-05T04:06", "[defer 03-05 04:06]"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := list.DeferSuffix(c.deferred, now); got != c.want {
				t.Errorf("DeferSuffix(%q, now) = %q, want %q", c.deferred, got, c.want)
			}
		})
	}
}

func TestVisible(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.Local)

	open := item("open", "open", nil, "", "")
	progress := item("progress", "in_progress", nil, "", "")
	done := item("done", "done", nil, "", "")
	custom := item("custom", "blocked", nil, "", "")
	futureDeferred := item("future", "open", nil, "", "2026-08-20T08:00")
	pastDeferred := item("past", "open", nil, "", "2026-08-10T08:00")

	cases := []struct {
		name string
		it   list.Item
		opts list.Options
		want bool
	}{
		{"default shows open", open, list.Options{}, true},
		{"default shows in_progress", progress, list.Options{}, true},
		{"default shows a custom status", custom, list.Options{}, true},
		{"default hides done", done, list.Options{}, false},
		{"default hides future-deferred", futureDeferred, list.Options{}, false},
		{"default shows past-deferred (available)", pastDeferred, list.Options{}, true},
		{"all shows done", done, list.Options{All: true}, true},
		{"all shows future-deferred", futureDeferred, list.Options{All: true}, true},
		{"status done shows done without all", done, list.Options{Status: "done"}, true},
		{"status in_progress shows only it", progress, list.Options{Status: "in_progress"}, true},
		{"status in_progress hides open", open, list.Options{Status: "in_progress"}, false},
		{"status open still hides future-deferred", futureDeferred, list.Options{Status: "open"}, false},
		{"label matches", labeled("a", "open", []string{"compras"}, ""), list.Options{Labels: []string{"compras"}}, true},
		{"label any-match", labeled("a", "open", []string{"saude"}, ""), list.Options{Labels: []string{"compras", "saude"}}, true},
		{"label mismatch hides", labeled("a", "open", []string{"saude"}, ""), list.Options{Labels: []string{"compras"}}, false},
		{"unlabelled issue vs label filter hides", open, list.Options{Labels: []string{"compras"}}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := list.Visible(c.it, c.opts, now); got != c.want {
				t.Errorf("Visible(%s) = %t, want %t", c.name, got, c.want)
			}
		})
	}
}

func TestPickNextChoosesLowestRankedAvailableOpenIssue(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.Local)
	items := []list.Item{
		item("backlog-old", "open", nil, "2020-01-01T10:00", ""),
		item("rank-two", "open", intPtr(2), "2026-08-14T10:00", ""),
		item("rank-one", "open", intPtr(1), "2026-08-15T10:00", ""),
		item("in-progress", "in_progress", intPtr(-1), "2026-08-13T10:00", ""),
		item("done", "done", intPtr(-2), "2026-08-12T10:00", ""),
		item("deferred", "open", intPtr(-3), "2026-08-11T10:00", "2026-08-20T08:00"),
	}

	got, err := list.PickNext(items, now)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "rank-one" {
		t.Errorf("PickNext() = %s, want rank-one", got.ID)
	}
}

func TestPickNextUsesAvailableBacklogWhenNoRankedCandidate(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.Local)
	items := []list.Item{
		item("future-old", "open", nil, "2020-01-01T10:00", "2026-08-20T08:00"),
		item("backlog-newer", "open", nil, "2026-08-16T10:00", ""),
		item("backlog-id-b", "open", nil, "2026-08-15T10:00", ""),
		item("backlog-id-a", "open", nil, "2026-08-15T10:00", ""),
	}

	got, err := list.PickNext(items, now)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "backlog-id-a" {
		t.Errorf("PickNext() = %s, want backlog-id-a", got.ID)
	}
}

func TestPickNextTreatsDeferredUntilNowAsAvailable(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.Local)
	items := []list.Item{
		item("at-now", "open", intPtr(1), "2026-08-15T10:00", "2026-08-15T12:00"),
	}

	got, err := list.PickNext(items, now)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "at-now" {
		t.Errorf("PickNext() = %s, want at-now", got.ID)
	}
}

func TestPickNextDoesNotReorderInput(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.Local)
	items := []list.Item{
		item("backlog", "open", nil, "2026-08-16T10:00", ""),
		item("ranked", "open", intPtr(1), "2026-08-15T10:00", ""),
	}

	if _, err := list.PickNext(items, now); err != nil {
		t.Fatal(err)
	}
	if got := ids(items); !slices.Equal(got, []string{"backlog", "ranked"}) {
		t.Errorf("PickNext reordered input: %v", got)
	}
}

func TestPickNextRejectsDuplicateRanks(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.Local)
	items := []list.Item{
		item("open", "open", intPtr(1), "2026-08-15T10:00", ""),
		item("in-progress", "in_progress", intPtr(1), "2026-08-15T11:00", ""),
	}

	if _, err := list.PickNext(items, now); err == nil || !strings.Contains(err.Error(), "duplicate rank") {
		t.Fatalf("PickNext() error = %v, want duplicate-rank error", err)
	}
}

func TestPickNextRejectsWhenNothingIsAvailable(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.Local)
	cases := []struct {
		name  string
		items []list.Item
	}{
		{name: "empty vault"},
		{name: "no open issues", items: []list.Item{
			item("progress", "in_progress", nil, "2026-08-15T10:00", ""),
			item("done", "done", nil, "2026-08-15T11:00", ""),
		}},
		{name: "all open issues deferred", items: []list.Item{
			item("future", "open", nil, "2026-08-15T10:00", "2026-08-20T08:00"),
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := list.PickNext(tc.items, now); err == nil || !strings.Contains(err.Error(), "no available") {
				t.Fatalf("PickNext() error = %v, want no-available error", err)
			}
		})
	}
}

func TestDuplicateRanks(t *testing.T) {
	t.Run("no duplicates", func(t *testing.T) {
		got := list.DuplicateRanks([]list.Item{
			item("a", "open", intPtr(1), "", ""),
			item("b", "open", intPtr(2), "", ""),
			item("c", "open", nil, "", ""),
		})
		if len(got) != 0 {
			t.Errorf("DuplicateRanks = %v, want empty", got)
		}
	})

	t.Run("nil rank is never a duplicate", func(t *testing.T) {
		got := list.DuplicateRanks([]list.Item{
			item("a", "open", nil, "", ""),
			item("b", "open", nil, "", ""),
			item("c", "open", nil, "", ""),
		})
		if len(got) != 0 {
			t.Errorf("DuplicateRanks = %v, want empty", got)
		}
	})

	t.Run("multiple duplicates sorted ascending", func(t *testing.T) {
		got := list.DuplicateRanks([]list.Item{
			item("a", "open", intPtr(3), "", ""),
			item("b", "open", intPtr(1), "", ""),
			item("c", "open", nil, "", ""),
			item("d", "open", intPtr(1), "", ""),
			item("e", "open", intPtr(3), "", ""),
		})
		want := []int{1, 3}
		if !slices.Equal(got, want) {
			t.Errorf("DuplicateRanks = %v, want %v", got, want)
		}
	})
}
