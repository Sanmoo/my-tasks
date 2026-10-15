package issue_test

import (
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

func TestClaimRecordsWhoAndTouchesNothingElse(t *testing.T) {
	i := populated()
	got := i.Claim("pi-sessao-7")

	if got.Frontmatter.ClaimedBy != "pi-sessao-7" {
		t.Errorf("ClaimedBy = %q, want the claimer", got.Frontmatter.ClaimedBy)
	}
	// Claim is a field write, not a transition: the status is untouched.
	if got.Frontmatter.Status != "in_progress" {
		t.Errorf("Status = %q, want untouched (%q)", got.Frontmatter.Status, "in_progress")
	}
	if got.Frontmatter.Title != "t" || got.Frontmatter.Rank == nil || *got.Frontmatter.Rank != 2 ||
		got.Frontmatter.DeferredUntil != "2026-08-20T08:00" {
		t.Errorf("Claim changed unrelated fields: %+v", got.Frontmatter)
	}
	// The receiver is untouched (value receiver).
	if i.Frontmatter.ClaimedBy != "" {
		t.Errorf("Claim mutated the receiver: %+v", i.Frontmatter)
	}
}

func TestClaimOverwritesAnExistingClaim(t *testing.T) {
	got := populated().Claim("a").Claim("b")
	if got.Frontmatter.ClaimedBy != "b" {
		t.Errorf("ClaimedBy = %q, want %q (posse pode ser transferida)", got.Frontmatter.ClaimedBy, "b")
	}
}

func TestUnclaimClearsTheFieldAndRendersNothing(t *testing.T) {
	i := populated().Claim("pi-sessao-7")
	got := i.Unclaim()

	if got.Frontmatter.ClaimedBy != "" {
		t.Errorf("ClaimedBy = %q, want empty", got.Frontmatter.ClaimedBy)
	}
	rendered, err := issue.Render(got)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(rendered), "claimed_by") {
		t.Errorf("unclaimed issue renders claimed_by:\n%s", rendered)
	}
}

func TestUnclaimOnNeverClaimedIsNoop(t *testing.T) {
	i := populated()
	got := i.Unclaim()
	if got.Frontmatter.ClaimedBy != "" {
		t.Errorf("ClaimedBy = %q, want empty", got.Frontmatter.ClaimedBy)
	}
	if got.Frontmatter.Status != "in_progress" {
		t.Errorf("Status = %q, want untouched", got.Frontmatter.Status)
	}
}

func TestUnclaimKeepsEverythingElse(t *testing.T) {
	i := populated().Claim("a")
	got := i.Unclaim()
	if got.Frontmatter.Rank == nil || *got.Frontmatter.Rank != 2 ||
		got.Frontmatter.CompletedAt != "2026-08-22T10:00" {
		t.Errorf("Unclaim changed unrelated fields: %+v", got.Frontmatter)
	}
	// The receiver keeps its claim (value receiver).
	if i.Frontmatter.ClaimedBy != "a" {
		t.Errorf("Unclaim mutated the receiver: %+v", i.Frontmatter)
	}
}

func TestReopenClearsTheClaim(t *testing.T) {
	i := populated().Claim("pi-sessao-7")
	got := i.Reopen()

	if got.Frontmatter.ClaimedBy != "" {
		t.Errorf("ClaimedBy = %q, want empty after reopen", got.Frontmatter.ClaimedBy)
	}
	if got.Frontmatter.Status != "open" || got.Frontmatter.CompletedAt != "" || got.Frontmatter.StartedAt != "" {
		t.Errorf("Reopen is broken: %+v", got.Frontmatter)
	}
}

func TestDonePreservesTheClaim(t *testing.T) {
	i := populated().Claim("pi-sessao-7")
	got := i.Done("2026-08-22T12:00")

	if got.Frontmatter.ClaimedBy != "pi-sessao-7" {
		t.Errorf("ClaimedBy = %q, want preserved by done", got.Frontmatter.ClaimedBy)
	}
	if got.Frontmatter.Status != "done" {
		t.Errorf("Status = %q, want done", got.Frontmatter.Status)
	}
}

func TestClaimRoundTripsThroughDisk(t *testing.T) {
	source := populated().Claim("pi-sessao-7")
	data, err := issue.Render(source)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := issue.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Frontmatter.ClaimedBy != "pi-sessao-7" {
		t.Errorf("ClaimedBy after round-trip = %q, want pi-sessao-7", parsed.Frontmatter.ClaimedBy)
	}
	// The field sits in the canonical order: before blocked_by, after
	// completed_at.
	text := string(data)
	completed := strings.Index(text, "completed_at:")
	claimed := strings.Index(text, "claimed_by:")
	if claimed < 0 || completed < 0 || claimed < completed {
		t.Errorf("claimed_by out of canonical order:\n%s", text)
	}
}
