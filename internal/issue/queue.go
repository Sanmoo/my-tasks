package issue

import (
	"cmp"
	"slices"
)

// Compare orders two items under the Fila order: lower Rank first (Items
// without a Rank form the Backlog and come last, ordered by created_at),
// then ID as the final tiebreak everywhere. It returns a negative value
// when a sorts before b, zero when they are equal, and a positive value
// otherwise — the cmp.Compare convention, so it plugs into
// slices.SortFunc.
//
// created_at is compared as a string, which equals chronological order
// for the canonical zero-padded, fixed-width stamp (NaiveLayout); a
// hand-edited stamp that drifts from that layout is mt check's to flag.
func Compare(a, b Item) int {
	ar, br := a.Issue.Frontmatter.Rank, b.Issue.Frontmatter.Rank
	if ar != nil && br != nil {
		if c := cmp.Compare(*ar, *br); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	}
	if ar != nil {
		return -1 // a is ranked, b is Backlog → a first
	}
	if br != nil {
		return 1 // a is Backlog, b is ranked → b first
	}
	// Both Backlog: oldest created_at first, then ID.
	if c := cmp.Compare(a.Issue.Frontmatter.CreatedAt, b.Issue.Frontmatter.CreatedAt); c != 0 {
		return c
	}
	return cmp.Compare(a.ID, b.ID)
}

// Change is a computed target Rank for one Item. Rank == nil means the
// Item returns to the Backlog (its Rank is removed).
type Change struct {
	ID   string
	Rank *int
}

// RenormalizeRanks computes the minimal changes needed to make every
// ranked Item's Rank contiguous from 1 through N. It includes every
// status because Rank is a vault-wide invariant; Backlog Items remain
// unranked. The existing Fila order is preserved — Rank, then ID as the
// tiebreak when duplicate Ranks need one. An Item already at its target
// Rank yields no change, so the caller rewrites only what moved.
func RenormalizeRanks(items []Item) []Change {
	ranked := make([]Item, 0, len(items))
	for _, item := range items {
		if item.Issue.Frontmatter.Rank != nil {
			ranked = append(ranked, item)
		}
	}
	slices.SortFunc(ranked, Compare)
	changes := make([]Change, 0, len(ranked))
	for index, item := range ranked {
		target := index + 1
		if *item.Issue.Frontmatter.Rank == target {
			continue
		}
		changes = append(changes, Change{ID: item.ID, Rank: &target})
	}
	return changes
}
