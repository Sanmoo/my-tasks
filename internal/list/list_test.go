// Package list_test holds the black-box unit tests of the mt list pure
// logic (Seam 2): the priority ordering comparator, the per-status
// glyphs, the deferred-until rules, and duplicate-rank detection.
package list_test

import (
	"slices"
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
