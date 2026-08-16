// Package list holds the pure logic of `mt list` and `mt pick-next`: the
// Rank ordering of issues (Rank → Backlog by created_at → ID), the
// per-status glyphs, the visibility rules (only done hides by default;
// status/label filters), the deferred-until availability/suffix rules,
// the computed blocked state (an Issue is blocked while any ID in its
// blocked_by is not done), and duplicate-rank detection. It is
// decision-dense, so it lives at Seam 2: black-box unit tested, with the
// coverage and mutation gates. Reading the issue files themselves is a
// process concern and stays in internal/cli.
package list

import (
	"cmp"
	"errors"
	"fmt"
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

// Ready reports whether item is an open Issue that is temporally
// available at now: its deferral, if any, has passed. Blocked state is
// computed separately against the whole Vault (see Blocked); callers
// that mean "available" combine both. An empty or malformed
// deferred_until does not prevent availability; mt check owns
// validation of persisted datetime fields.
func Ready(item Item, now time.Time) bool {
	return item.Issue.Frontmatter.Status == "open" && !IsFutureDeferred(item.Issue.Frontmatter.DeferredUntil, now)
}

// Overdue reports whether item has a Deadline before now and is not done.
// Deadline is informational, so a future deferral does not affect this result.
// An empty or malformed deadline is not overdue; mt check owns validation.
func Overdue(item Item, now time.Time) bool {
	deadline, ok := parseNaive(item.Issue.Frontmatter.Deadline)
	return ok && deadline.Before(now) && item.Issue.Frontmatter.Status != "done"
}

// DeferralExpired reports whether deferredUntil is a datetime whose time
// has arrived: now >= deferred_until. It is the mirror of IsFutureDeferred.
// An empty or malformed value is not an expired deferral (format
// validation is mt check's).
func DeferralExpired(deferredUntil string, now time.Time) bool {
	t, ok := parseNaive(deferredUntil)
	return ok && !t.After(now)
}

// ExpiredSuffix returns the "[expirada MM-DD]" marker for a
// deferred_until datetime whose time has arrived relative to now. It
// returns "" when deferredUntil is empty, still in the future, or
// malformed.
func ExpiredSuffix(deferredUntil string, now time.Time) string {
	t, ok := parseNaive(deferredUntil)
	if !ok || t.After(now) {
		return ""
	}
	return "[expirada " + t.Format("01-02") + "]"
}

// DeadlineSuffix returns the "[deadline MM-DD]" marker for a deadline
// that has passed relative to now. It returns "" when deadline is empty,
// not yet passed, or malformed.
func DeadlineSuffix(deadline string, now time.Time) string {
	t, ok := parseNaive(deadline)
	if !ok || !t.Before(now) {
		return ""
	}
	return "[deadline " + t.Format("01-02") + "]"
}

// OverdueGroups partitions items into the two groups of the overdue
// temporal-attention view: items whose deferral has expired first, then
// items whose deadline has passed. Only non-done Items participate; an
// Item with both signals appears once, in the expired group. The input
// order (the vault's Rank order) is preserved within each group.
func OverdueGroups(items []Item, now time.Time) (expired, late []Item) {
	for _, it := range items {
		if it.Issue.Frontmatter.Status == "done" {
			continue
		}
		if DeferralExpired(it.Issue.Frontmatter.DeferredUntil, now) {
			expired = append(expired, it)
			continue
		}
		if Overdue(it, now) {
			late = append(late, it)
		}
	}
	return expired, late
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

// StatusByID indexes items by their ID for the blocked lookup: the map
// holds each item's status, and an ID with no issue is absent. The
// caller builds it once per vault view and passes it to Blocked.
func StatusByID(items []Item) map[string]string {
	byID := make(map[string]string, len(items))
	for _, item := range items {
		byID[item.ID] = item.Issue.Frontmatter.Status
	}
	return byID
}

// Blocked reports whether an Issue with the given blocked_by list is
// blocked: at least one referenced ID is not done. An ID absent from
// statusByID counts as not done — a dangling reference can never be
// done, so the Issue stays blocked; mt check owns flagging the missing
// reference. An empty list is never blocked.
func Blocked(blockedBy []string, statusByID map[string]string) bool {
	for _, id := range blockedBy {
		if statusByID[id] != "done" {
			return true
		}
	}
	return false
}

// Options selects the issues a list view shows.
//
// All reveals done issues; without it, done issues are hidden. A future
// deferral never hides an issue: the [defer ...] suffix marks it.
type Options struct {
	// All reveals done issues (future-deferred issues are always
	// shown, marked with DeferSuffix).
	All bool
	// Status narrows to a single status when non-empty. An explicit
	// Status overrides the default hiding of done: --status done shows
	// done issues even without All.
	Status string
	// Labels narrows, when non-empty, to issues carrying at least one
	// of these labels.
	Labels []string
}

// Visible reports whether item appears in a list view under opts. done
// issues are hidden unless opts.All or an explicit opts.Status asks for
// them; a future deferral does not hide an issue (the [defer ...]
// suffix signals its unavailability); opts.Labels narrow the view to
// issues carrying at least one of the labels.
func Visible(item Item, opts Options) bool {
	fm := item.Issue.Frontmatter
	if opts.Status != "" {
		if fm.Status != opts.Status {
			return false
		}
	} else if !opts.All && fm.Status == "done" {
		return false
	}
	if len(opts.Labels) > 0 && !hasAnyLabel(fm.Labels, opts.Labels) {
		return false
	}
	return true
}

// PickNext returns the first available open Issue under the Rank order:
// the lowest rank wins; when no ranked candidate is
// available, the oldest Backlog Issue wins, with ID as the final tie-break.
// Future-deferred and blocked Issues are unavailable, while a deferred_until
// exactly at now is available. Duplicate ranks anywhere in the vault are rejected
// before candidate selection, including ranks on non-open Issues.
func PickNext(items []Item, now time.Time) (Item, error) {
	if dups := DuplicateRanks(items); len(dups) > 0 {
		return Item{}, fmt.Errorf("duplicate rank: %d", dups[0])
	}

	statusByID := StatusByID(items)
	candidates := make([]Item, 0, len(items))
	for _, item := range items {
		if item.Issue.Frontmatter.Status != "open" {
			continue
		}
		if IsFutureDeferred(item.Issue.Frontmatter.DeferredUntil, now) {
			continue
		}
		if Blocked(item.Issue.Frontmatter.BlockedBy, statusByID) {
			continue
		}
		candidates = append(candidates, item)
	}
	if len(candidates) == 0 {
		return Item{}, errors.New("no available open issues")
	}
	Sort(candidates)
	return candidates[0], nil
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
