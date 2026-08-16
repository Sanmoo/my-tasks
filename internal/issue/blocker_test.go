package issue_test

import (
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

func TestAddBlockerAppendsAndLeavesRestUntouched(t *testing.T) {
	i := populated()
	got := i.AddBlocker("pkm-055")

	if len(got.Frontmatter.BlockedBy) != 1 || got.Frontmatter.BlockedBy[0] != "pkm-055" {
		t.Errorf("BlockedBy = %v, want [pkm-055]", got.Frontmatter.BlockedBy)
	}
	// Every field other than BlockedBy is untouched.
	if got.Frontmatter.Title != "t" || got.Frontmatter.Status != "in_progress" ||
		got.Frontmatter.Rank == nil || *got.Frontmatter.Rank != 2 ||
		got.Frontmatter.Deadline != "2026-08-22T18:00" || got.Frontmatter.StartedAt != "2026-08-21T09:00" ||
		got.Frontmatter.CompletedAt != "2026-08-22T10:00" {
		t.Errorf("AddBlocker changed unrelated fields: %+v", got.Frontmatter)
	}
	// The receiver is untouched (value receiver).
	if len(i.Frontmatter.BlockedBy) != 0 {
		t.Errorf("AddBlocker mutated the receiver: %+v", i.Frontmatter)
	}
}

func TestAddBlockerPreservesExistingBlockers(t *testing.T) {
	i := populated().AddBlocker("pkm-001")
	got := i.AddBlocker("pkm-002")
	want := []string{"pkm-001", "pkm-002"}
	if len(got.Frontmatter.BlockedBy) != 2 ||
		got.Frontmatter.BlockedBy[0] != want[0] || got.Frontmatter.BlockedBy[1] != want[1] {
		t.Errorf("BlockedBy = %v, want %v", got.Frontmatter.BlockedBy, want)
	}
}

func TestAddBlockerIsIdempotent(t *testing.T) {
	i := populated().AddBlocker("pkm-055")
	got := i.AddBlocker("pkm-055")
	if len(got.Frontmatter.BlockedBy) != 1 || got.Frontmatter.BlockedBy[0] != "pkm-055" {
		t.Errorf("BlockedBy = %v, want [pkm-055] listed once", got.Frontmatter.BlockedBy)
	}
}

func TestRemoveBlockerRemovesOnlyIt(t *testing.T) {
	i := populated().AddBlocker("pkm-001").AddBlocker("pkm-002")
	got := i.RemoveBlocker("pkm-001")
	if len(got.Frontmatter.BlockedBy) != 1 || got.Frontmatter.BlockedBy[0] != "pkm-002" {
		t.Errorf("BlockedBy = %v, want [pkm-002]", got.Frontmatter.BlockedBy)
	}
	// The receiver is untouched (value receiver).
	if len(i.Frontmatter.BlockedBy) != 2 {
		t.Errorf("RemoveBlocker mutated the receiver: %+v", i.Frontmatter)
	}
}

func TestRemoveBlockerLastOneLeavesEmptyList(t *testing.T) {
	got := populated().AddBlocker("pkm-055").RemoveBlocker("pkm-055")
	if len(got.Frontmatter.BlockedBy) != 0 {
		t.Errorf("BlockedBy = %v, want empty", got.Frontmatter.BlockedBy)
	}
}

func TestRemoveBlockerIsIdempotentWhenNotListed(t *testing.T) {
	i := populated()
	got := i.RemoveBlocker("pkm-055")
	if len(got.Frontmatter.BlockedBy) != 0 {
		t.Errorf("BlockedBy = %v, want empty", got.Frontmatter.BlockedBy)
	}
	// Removing a blocker from an Issue that never listed it leaves the
	// Issue byte-identical on disk.
	a, err := issue.Render(i)
	if err != nil {
		t.Fatal(err)
	}
	b, err := issue.Render(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(a) != string(b) {
		t.Errorf("RemoveBlocker changed an Issue that did not list the blocker")
	}
}

func TestBlockerRoundTripThroughRenderAndParse(t *testing.T) {
	got, err := issue.Render(populated().AddBlocker("pkm-001").AddBlocker("pkm-002"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	if !strings.Contains(s, "blocked_by: [pkm-001, pkm-002]") {
		t.Fatalf("rendered issue misses blocked_by flow list:\n%s", s)
	}
	parsed, err := issue.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.Frontmatter.BlockedBy) != 2 ||
		parsed.Frontmatter.BlockedBy[0] != "pkm-001" || parsed.Frontmatter.BlockedBy[1] != "pkm-002" {
		t.Errorf("Parse round-trip BlockedBy = %v, want [pkm-001 pkm-002]", parsed.Frontmatter.BlockedBy)
	}
}

func TestRenderOmitsBlockedByWhenEmpty(t *testing.T) {
	got, err := issue.Render(populated())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "blocked_by") {
		t.Errorf("Render = %q contains blocked_by, want it omitted", got)
	}
}
