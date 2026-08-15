// Package list holds the pure logic of `mt list`: the priority ordering
// of issues (Rank → Backlog by created_at → ID), the per-status glyphs,
// the visibility rules (done/future-deferred hiding, status/label
// filters), the deferred-until availability/suffix rules, and
// duplicate-rank detection. It is decision-dense, so it lives at Seam 2:
// black-box unit tested, with the coverage and mutation gates. Reading
// the issue files themselves is a process concern and stays in
// internal/cli.
package list

import (
	"cmp"
	"slices"
	"sort"
	"time"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// Item is one Issue in a list view: the file name ID (the authority,
// no id field in the frontmatter) plus the parsed Issue.
type Item struct {
	ID    string
	Issue issue.Issue
}

// glyphs maps the three built-in statuses to their list glyph. Any
// other (custom) status falls back to a distinct marker.
const (
	glyphOpen       = "○"
	glyphInProgress = "◐"
	glyphDone       = "●"
	glyphOther      = "?"
)

// Glyph returns the list glyph for a status: ○ for open, ◐ for
// in_progress, ● for done, and ? for any other (custom) status.
func Glyph(status string) string {
	switch status {
	case "open":
		return glyphOpen
	case "in_progress":
		return glyphInProgress
	case "done":
		return glyphDone
	default:
		return glyphOther
	}
}

// Compare orders two items under the list order: lower rank first
// (issues without a rank form the Backlog and come last, ordered by
// created_at), then ID as the final tiebreak everywhere. created_at is
// compared as a string, which equals chronological order for the
// canonical zero-padded, fixed-width stamp (issue.NaiveLayout); a
// hand-edited stamp that drifts from that layout is mt check's to flag.
func Compare(a, b Item) int {
	ar, br := a.Issue.Frontmatter.Rank, b.Issue.Frontmatter.Rank
	if ar != nil && br != nil {
		if c := cmp.Compare(*ar, *br); c != 0 {
			return c
		}
		return compareID(a.ID, b.ID)
	}
	if ar != nil {
		return -1 // a is ranked, b is Backlog → a first
	}
	if br != nil {
		return 1 // a is Backlog, b is ranked → b first
	}
	// Both Backlog: oldest created_at first (lexicographic, see Compare),
	// then ID.
	if c := cmp.Compare(a.Issue.Frontmatter.CreatedAt, b.Issue.Frontmatter.CreatedAt); c != 0 {
		return c
	}
	return compareID(a.ID, b.ID)
}

// compareID orders two IDs ascending. A plain string comparison is the
// order of last resort: IDs are unique per vault, so it makes the list
// order total and deterministic even under a duplicate rank.
func compareID(a, b string) int { return cmp.Compare(a, b) }

// Sort orders items in place by Compare.
func Sort(items []Item) {
	sort.SliceStable(items, func(i, j int) bool {
		return Compare(items[i], items[j]) < 0
	})
}

// parseNaive parses a stored naive datetime in the local zone. ok is
// false when s is empty or malformed; callers then treat the field as
// absent (list is lenient — mt check owns format validation).
func parseNaive(s string) (time.Time, bool) {
	t, err := time.ParseInLocation(issue.NaiveLayout, s, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// IsFutureDeferred reports whether deferredUntil is a datetime after
// now. An empty or malformed value is not deferred.
func IsFutureDeferred(deferredUntil string, now time.Time) bool {
	t, ok := parseNaive(deferredUntil)
	return ok && t.After(now)
}

// DeferSuffix returns the "[defer MM-DD HH:MM]" marker for a
// deferred_until datetime in the future relative to now. It returns ""
// when deferredUntil is empty, not in the future, or malformed.
func DeferSuffix(deferredUntil string, now time.Time) string {
	t, ok := parseNaive(deferredUntil)
	if !ok || !t.After(now) {
		return ""
	}
	return "[defer " + t.Format("01-02 15:04") + "]"
}

// Options selects the issues a list view shows.
type Options struct {
	// All reveals done issues and future-deferred issues (the latter
	// marked with DeferSuffix).
	All bool
	// Status narrows to a single status when non-empty. An explicit
	// Status overrides the default hiding of done: --status done shows
	// done issues even without All.
	Status string
	// Labels narrows, when non-empty, to issues carrying at least one
	// of these labels.
	Labels []string
}

// Visible reports whether item appears in a list view under opts at time
// now. done and future-deferred issues are hidden unless opts.All or an
// explicit opts.Status asks for them; opts.Labels narrow the view to
// issues carrying at least one of the labels.
func Visible(item Item, opts Options, now time.Time) bool {
	fm := item.Issue.Frontmatter
	if opts.Status != "" {
		if fm.Status != opts.Status {
			return false
		}
	} else if !opts.All && fm.Status == "done" {
		return false
	}
	if !opts.All && IsFutureDeferred(fm.DeferredUntil, now) {
		return false
	}
	if len(opts.Labels) > 0 && !hasAnyLabel(fm.Labels, opts.Labels) {
		return false
	}
	return true
}

// hasAnyLabel reports whether labels contains at least one of filters.
func hasAnyLabel(labels, filters []string) bool {
	for _, f := range filters {
		if slices.Contains(labels, f) {
			return true
		}
	}
	return false
}

// DuplicateRanks returns the rank values that appear in more than one
// ranked item, sorted ascending. A nil rank (Backlog) is not a rank
// value and never counts as a duplicate.
func DuplicateRanks(items []Item) []int {
	counts := make(map[int]int)
	for _, it := range items {
		if r := it.Issue.Frontmatter.Rank; r != nil {
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
