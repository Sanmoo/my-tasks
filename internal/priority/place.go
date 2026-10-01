package priority

import (
	"fmt"
	"slices"
)

// This file holds the pure logic of the Encaixe: discovering the rank of one
// Issue by asking the user a sequence of Comparações (a-vs-b) and halving the
// interval of possible positions each time. The session is the optimal
// ⌈log₂(N+1)⌉ Comparações for inserting into an ordered list of N items — the
// information-theoretic lower bound — because it assumes the current queue is
// a valid order and only searches the insertion point. It is decision-dense,
// so it lives at Seam 2 with the rest of the ordering logic: black-box unit
// tested, with the coverage and mutation gates. Asking and reading answers is
// a process concern and stays in internal/cli.

// Candidates returns the reference list for placing id: the vault's ranked,
// prioritizable Issues in Compare order (rank, then ID), with the Issue being
// placed left out — a session never compares an Issue with itself. Backlog
// Issues (no rank) are not candidates: the Backlog is not ordered by
// priority, so a Comparação against one would ask about a position the queue
// never uses. Non-prioritizable Issues (done, custom status) are not
// candidates either: they hold ranks after the queue (N+1..M) and stay there.
func Candidates(issues []Issue, id string) []Issue {
	candidates := make([]Issue, 0, len(issues))
	for _, is := range issues {
		if is.ID == id || is.Rank == nil || !Prioritizable(is.Status) {
			continue
		}
		candidates = append(candidates, is)
	}
	slices.SortFunc(candidates, Compare)
	return candidates
}

// PlacementTarget returns the Issue that id names among issues for an
// Encaixe: it has to exist and be prioritizable (open or in_progress). A
// session validates its target before its first Comparação, so an unknown or
// finished Issue fails without asking anything.
func PlacementTarget(issues []Issue, id string) (Issue, error) {
	return findPrioritizable(issues, id)
}

// findPrioritizable returns the Issue id names among issues, or an error: an
// unknown ID, or a known Issue that is not prioritizable (done or a custom
// status). It is the eligibility rule shared by the ordering plans and the
// Encaixe session.
func findPrioritizable(issues []Issue, id string) (Issue, error) {
	for _, is := range issues {
		if is.ID != id {
			continue
		}
		if !Prioritizable(is.Status) {
			return Issue{}, fmt.Errorf("issue %s is %s and cannot be prioritized", id, is.Status)
		}
		return is, nil
	}
	return Issue{}, fmt.Errorf("unknown issue ID %q", id)
}

// Answer is the reply to one Comparação.
type Answer uint8

const (
	// TargetBefore is "the Issue being placed is more prioritized than the
	// Candidate" — the placed Issue takes the earlier position.
	TargetBefore Answer = iota
	// TargetAfter is "the Candidate is more prioritized" — the placed Issue
	// takes the later position.
	TargetAfter
	// Indifferent leaves the decision to the order already in the vault: the
	// placed Issue goes immediately after the Candidate, so an Issue that was
	// already in the queue never loses a position to a newcomer out of
	// indecision. It ends the session — a Comparison is only worth asking
	// when the answer moves the search.
	Indifferent
)

// Place is the state of one Encaixe session: the ordered Candidates and the
// interval of queue positions still possible. The zero value is not a valid
// session; use NewPlace.
type Place struct {
	candidates []Issue
	// lo and hi bound the remaining interval of slots: slot i means "before
	// Candidate i", so the N Candidates offer N+1 slots (0..N). The session
	// is over when a single slot is left.
	lo, hi int
	asked  int
}

// NewPlace starts a session over candidates, which must be in Compare order
// and must not contain the Issue being placed (see Candidates). The slice is
// copied, so the caller's order is not shared with the session.
func NewPlace(candidates []Issue) *Place {
	copied := slices.Clone(candidates)
	// The N Candidates offer N+1 slots, so the interval starts at N+1 — not
	// at the candidate count, which would leave the last slot unsearchable.
	return &Place{candidates: copied, lo: 0, hi: len(copied) + 1}
}

// Question returns the Candidate in the middle of the remaining interval —
// the one the current Comparação asks about. It reports false when the
// session is over (see Finished), which is the case from the start for an
// empty queue.
func (p *Place) Question() (Issue, bool) {
	if p.Finished() {
		return Issue{}, false
	}
	return p.candidates[p.middle()], true
}

// Answer records the reply to the current Question and narrows the interval
// to the half the reply allows. It is a no-op once the session is over.
func (p *Place) Answer(a Answer) {
	if p.Finished() {
		return
	}
	mid := p.middle()
	p.asked++
	switch a {
	case TargetBefore:
		// The placed Issue comes first, so its slot is at most mid — slot
		// mid itself means "immediately before Candidate mid".
		p.hi = mid + 1
	case TargetAfter:
		p.lo = mid + 1
	case Indifferent:
		// Immediately after the Candidate: slot mid+1, which the interval
		// already guarantees to be in range. The interval collapses onto that
		// single slot, so any later Question is over.
		p.lo = mid + 1
		p.hi = p.lo
	}
}

// Finished reports whether the session has a position: the interval has
// collapsed to a single slot. An empty queue is finished from the start —
// the placed Issue takes rank 1 with no Comparação at all.
func (p *Place) Finished() bool { return p.hi-p.lo <= 1 }

// Rank is the one-based queue position the placed Issue takes: the slot the
// search converged to. Before the session ends it is the best position the
// answers so far allow.
func (p *Place) Rank() int { return p.lo + 1 }

// Asked is the number of Comparações already answered.
func (p *Place) Asked() int { return p.asked }

// Total is the number of Comparações the session asks at most: the optimal
// ⌈log₂(N+1)⌉ for N Candidates, and zero for an empty queue. It is the
// progress denominator; Indifferent can end a session earlier.
func (p *Place) Total() int {
	// The number of slots, and the smallest q with 2^q >= slots.
	slots := len(p.candidates) + 1
	total := 0
	for size := 1; size < slots; size *= 2 {
		total++
	}
	return total
}

// middle is the Candidate index the current Comparação asks about: the lower
// middle of the interval, so the "before" half never grows past the "after"
// one and the ⌈log₂⌉ bound holds.
func (p *Place) middle() int { return p.lo + (p.hi-p.lo-1)/2 }
