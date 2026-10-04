package priority_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issue"
	"github.com/Sanmoo/my-tasks2/internal/priority"
)

// placeCandidates builds n queued Items already in Fila order (ranks
// 1..n), which is what a Place session expects as its reference list.
func placeCandidates(n int) []issue.Item {
	candidates := make([]issue.Item, 0, n)
	for i := range n {
		candidates = append(candidates, titled(
			fmt.Sprintf("c%02d", i+1),
			fmt.Sprintf("Candidate %d", i+1),
			"open",
			ptr(i+1),
			fmt.Sprintf("2026-01-%02dT10:00", i+1),
		))
	}
	return candidates
}

// candidateIndex is the position of the Issue the session asked about in the
// reference list — the mid the oracle needs to answer consistently.
func candidateIndex(t *testing.T, candidates []issue.Item, id string) int {
	t.Helper()
	for i, it := range candidates {
		if it.ID == id {
			return i
		}
	}
	t.Fatalf("session asked about %q, which is not a candidate", id)
	return -1
}

// slotAnswer is the reply of a user whose mind is perfectly consistent: the
// placed Issue belongs in slot (it is more prioritized than the Candidate at
// or after that slot, and less prioritized than the ones before it).
func slotAnswer(slot, mid int) priority.Answer {
	if slot <= mid {
		return priority.TargetBefore
	}
	return priority.TargetAfter
}

// ceilLog2 is the question count the session promises, computed independently
// of the implementation.
func ceilLog2(n int) int {
	total := 0
	for size := 1; size < n; size *= 2 {
		total++
	}
	return total
}

func TestCandidatesKeepsOnlyQueuedPrioritizableItems(t *testing.T) {
	items := []issue.Item{
		titled("b1", "Backlog", "open", nil, "2026-01-01T10:00"),
		titled("r2", "Rank 2", "open", ptr(2), "2026-01-02T10:00"),
		titled("target", "Placed", "open", ptr(3), "2026-01-03T10:00"),
		titled("r1", "Rank 1", "in_progress", ptr(1), "2026-01-04T10:00"),
		titled("r9", "Finished", "done", ptr(9), "2026-01-05T10:00"),
		titled("rc", "Review", "review", ptr(4), "2026-01-06T10:00"),
	}

	got := priority.Candidates(items, "target")

	ids := make([]string, 0, len(got))
	for _, it := range got {
		ids = append(ids, it.ID)
	}
	// Rank order, without the Issue being placed, the Backlog, done and the
	// custom status.
	if want := []string{"r1", "r2"}; !slices.Equal(ids, want) {
		t.Errorf("Candidates() = %v, want %v", ids, want)
	}
	if len(items) != 6 {
		t.Errorf("Candidates() mutated its input: len = %d, want 6", len(items))
	}
}

func TestCandidatesIsEmptyWithoutAQueue(t *testing.T) {
	items := []issue.Item{
		titled("b1", "Backlog", "open", nil, "2026-01-01T10:00"),
		titled("d1", "Finished", "done", ptr(1), "2026-01-02T10:00"),
	}
	if got := priority.Candidates(items, "b1"); len(got) != 0 {
		t.Errorf("Candidates() = %v, want empty", got)
	}
}

func TestPlacementTarget(t *testing.T) {
	items := []issue.Item{
		titled("open", "Open", "open", nil, ""),
		titled("doing", "Doing", "in_progress", nil, ""),
		titled("finished", "Finished", "done", nil, ""),
		titled("review", "Review", "review", nil, ""),
	}
	for _, id := range []string{"open", "doing"} {
		target, err := priority.PlacementTarget(items, id)
		if err != nil {
			t.Errorf("PlacementTarget(%q) error = %v, want none", id, err)
		}
		if target.ID != id {
			t.Errorf("PlacementTarget(%q).ID = %q, want %q", id, target.ID, id)
		}
	}
	tests := []struct {
		id      string
		wantErr string
	}{
		{"finished", "finished is done and cannot be prioritized"},
		{"review", "review is review and cannot be prioritized"},
		{"nope", `unknown issue ID "nope"`},
	}
	for _, tt := range tests {
		_, err := priority.PlacementTarget(items, tt.id)
		if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
			t.Errorf("PlacementTarget(%q) error = %v, want %q", tt.id, err, tt.wantErr)
		}
	}
}

func TestPlaceTotalIsTheOptimalQuestionCount(t *testing.T) {
	tests := []struct {
		candidates int
		want       int
	}{
		{0, 0}, {1, 1}, {2, 2}, {3, 2}, {4, 3}, {5, 3}, {6, 3}, {7, 3}, {8, 4}, {21, 5},
	}
	for _, tt := range tests {
		place := priority.NewPlace(placeCandidates(tt.candidates))
		if got := place.Total(); got != tt.want {
			t.Errorf("Total() with %d candidates = %d, want %d", tt.candidates, got, tt.want)
		}
	}
}

func TestPlaceEmptyQueuePlacesAtRankOneWithoutAsking(t *testing.T) {
	place := priority.NewPlace(nil)
	if !place.Finished() {
		t.Fatal("Finished() = false with an empty queue, want true")
	}
	if _, ok := place.Question(); ok {
		t.Error("Question() asked about an empty queue")
	}
	if got := place.Rank(); got != 1 {
		t.Errorf("Rank() = %d, want 1", got)
	}
	if got := place.Asked(); got != 0 {
		t.Errorf("Asked() = %d, want 0", got)
	}
}

func TestPlaceAnswersNarrowTheInterval(t *testing.T) {
	candidates := placeCandidates(4)
	tests := []struct {
		name string
		// mid answers: one letter per Comparação, in order.
		answers string
		want    int
	}{
		{"all before", "aaa", 1},
		{"all after", "bb", 5},
		{"mid candidate first", "ab", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			place := priority.NewPlace(candidates)
			for _, key := range tt.answers {
				question, ok := place.Question()
				if !ok {
					t.Fatalf("session ended after fewer than %d answers with Rank() = %d", len(tt.answers), place.Rank())
				}
				if key == 'a' {
					place.Answer(priority.TargetBefore)
				} else {
					place.Answer(priority.TargetAfter)
				}
				if question.ID == "" {
					t.Error("Question() returned an empty Candidate")
				}
			}
			if !place.Finished() {
				t.Fatalf("Finished() = false after %d answers", len(tt.answers))
			}
			if got := place.Rank(); got != tt.want {
				t.Errorf("Rank() = %d, want %d", got, tt.want)
			}
			if got := place.Asked(); got != len(tt.answers) {
				t.Errorf("Asked() = %d, want %d", got, len(tt.answers))
			}
		})
	}
}

func TestPlaceQuestionIsTheMiddleCandidate(t *testing.T) {
	candidates := placeCandidates(4)
	place := priority.NewPlace(candidates)
	question, ok := place.Question()
	if !ok {
		t.Fatal("Question() reported a finished session, want a question")
	}
	if question.ID != "c03" {
		t.Errorf("first Question() = %s, want the middle candidate c03", question.ID)
	}
	// "Before" keeps slot 0..2 in play, so the next question is the middle of
	// that half.
	place.Answer(priority.TargetBefore)
	question, _ = place.Question()
	if question.ID != "c02" {
		t.Errorf("second Question() = %s, want c02", question.ID)
	}
}

func TestPlaceIndifferentKeepsTheCandidateAhead(t *testing.T) {
	candidates := placeCandidates(6)
	place := priority.NewPlace(candidates)
	question, ok := place.Question()
	if !ok {
		t.Fatal("Question() reported a finished session")
	}
	mid := candidateIndex(t, candidates, question.ID)
	before := place.Rank()

	place.Answer(priority.Indifferent)

	if !place.Finished() {
		t.Fatal("Finished() = false after Indifferent, want the session over")
	}
	if got, want := place.Rank(), mid+2; got != want {
		t.Errorf("Rank() = %d, want %d (immediately after %s)", got, want, question.ID)
	}
	if got := place.Rank(); got <= before {
		t.Errorf("Rank() = %d did not move past the pre-answer position %d", got, before)
	}
	if got := place.Asked(); got != 1 {
		t.Errorf("Asked() = %d, want 1", got)
	}
}

func TestPlaceAnswerAfterTheEndIsIgnored(t *testing.T) {
	place := priority.NewPlace(placeCandidates(1))
	place.Answer(priority.TargetAfter)
	if !place.Finished() {
		t.Fatal("Finished() = false after one answer, want the session over")
	}
	rank, asked := place.Rank(), place.Asked()

	place.Answer(priority.TargetBefore)
	place.Answer(priority.Indifferent)

	if got := place.Rank(); got != rank {
		t.Errorf("Rank() = %d after answers past the end, want %d", got, rank)
	}
	if got := place.Asked(); got != asked {
		t.Errorf("Asked() = %d after answers past the end, want %d", got, asked)
	}
}

func TestPlaceCopiesTheCandidates(t *testing.T) {
	candidates := placeCandidates(1)
	place := priority.NewPlace(candidates)
	candidates[0].Issue.Frontmatter.Title = "mutated"
	candidates[0].Issue.Frontmatter.Status = "done"

	question, ok := place.Question()
	if !ok {
		t.Fatal("Question() reported a finished session")
	}
	fm := question.Issue.Frontmatter
	if fm.Title != "Candidate 1" || fm.Status != "open" {
		t.Errorf("Question() = %+v, want the copy taken at NewPlace", question)
	}
}

// TestPlaceFindsEverySlotInTheOptimalNumberOfQuestions is the property behind
// the feature: whatever slot the user's mind is in, the session converges to
// that slot, never asks more than Total(), and does reach Total() for the
// slot that needs it — so Total() is both an upper bound and tight.
func TestPlaceFindsEverySlotInTheOptimalNumberOfQuestions(t *testing.T) {
	for n := range 13 {
		candidates := placeCandidates(n)
		wantTotal := ceilLog2(n + 1)
		worst := 0
		for slot := range n + 1 {
			place := priority.NewPlace(candidates)
			for {
				question, ok := place.Question()
				if !ok {
					break
				}
				place.Answer(slotAnswer(slot, candidateIndex(t, candidates, question.ID)))
				if place.Asked() > wantTotal {
					t.Fatalf("n = %d, slot = %d: asked %d questions, want at most %d", n, slot, place.Asked(), wantTotal)
				}
			}
			if !place.Finished() {
				t.Fatalf("n = %d, slot = %d: session did not finish", n, slot)
			}
			if got, want := place.Rank(), slot+1; got != want {
				t.Errorf("n = %d, slot = %d: Rank() = %d, want %d", n, slot, got, want)
			}
			worst = max(worst, place.Asked())
		}
		if worst != wantTotal {
			t.Errorf("n = %d: worst case asked %d questions, Total() = %d", n, worst, wantTotal)
		}
	}
}

// TestPlaceRankDrivesTheQueueOrder closes the loop between the session and the
// plan the commands apply: feeding the session's Rank to the quick-order plan
// puts the placed Issue exactly at the slot the answers described. It covers
// both entry points — an Issue already in the queue and one still in the
// Backlog — for every slot.
func TestPlaceRankDrivesTheQueueOrder(t *testing.T) {
	const queue = 5
	for _, ranked := range []bool{true, false} {
		for slot := range queue + 1 {
			items := placeCandidates(queue)
			target := titled("target", "Nova", "open", ptr(3), "2026-02-01T10:00")
			if !ranked {
				target.Issue.Frontmatter.Rank = nil
			}
			items = append(items, target)

			candidates := priority.Candidates(items, "target")
			place := priority.NewPlace(candidates)
			for {
				question, ok := place.Question()
				if !ok {
					break
				}
				place.Answer(slotAnswer(slot, candidateIndex(t, candidates, question.ID)))
			}

			changes, err := priority.QuickPlan(items, "target", priority.MoveToRank, place.Rank())
			if err != nil {
				t.Fatalf("ranked = %v, slot = %d: QuickPlan error = %v", ranked, slot, err)
			}
			updated := slices.Clone(items)
			for _, ch := range changes {
				for i := range updated {
					if updated[i].ID == ch.ID {
						updated[i].Issue.Frontmatter.Rank = ch.Rank
					}
				}
			}
			slices.SortFunc(updated, issue.Compare)
			placed := 0
			for i, it := range updated {
				if it.ID == "target" {
					placed = i
				}
			}
			if placed != slot {
				t.Errorf("ranked = %v, slot = %d: the plan put the Issue at %d", ranked, slot, placed)
			}
		}
	}
}
