package issue_test

import (
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// populated returns an Issue with every optional field set, so the
// transition tests can assert what is preserved, changed and cleared.
func populated() issue.Issue {
	return issue.Issue{
		Frontmatter: issue.Frontmatter{
			Title:         "t",
			Status:        "in_progress",
			Labels:        []string{"a"},
			CreatedAt:     "2026-08-15T09:30",
			Rank:          intPtr(2),
			DeferredUntil: "2026-08-20T08:00",
			Deadline:      "2026-08-22T18:00",
			StartedAt:     "2026-08-21T09:00",
			CompletedAt:   "2026-08-22T10:00",
		},
		Body: issue.DefaultBody,
	}
}

func TestDoneStampsCompletedAtAndSetsStatus(t *testing.T) {
	i := populated()
	got := i.Done("2026-08-22T12:00")

	if got.Frontmatter.Status != "done" {
		t.Errorf("Status = %q, want %q", got.Frontmatter.Status, "done")
	}
	if got.Frontmatter.CompletedAt != "2026-08-22T12:00" {
		t.Errorf("CompletedAt = %q, want the stamped time", got.Frontmatter.CompletedAt)
	}
	// done records completion, not a fresh start: started_at survives.
	if got.Frontmatter.StartedAt != "2026-08-21T09:00" {
		t.Errorf("StartedAt = %q, want it preserved", got.Frontmatter.StartedAt)
	}
	if got.Frontmatter.Title != "t" || got.Frontmatter.Labels[0] != "a" || got.Frontmatter.Rank == nil || *got.Frontmatter.Rank != 2 {
		t.Errorf("Done changed unrelated fields: %+v", got.Frontmatter)
	}
	// The original is untouched (value receiver).
	if i.Frontmatter.Status != "in_progress" || i.Frontmatter.CompletedAt != "2026-08-22T10:00" {
		t.Errorf("Done mutated the receiver: %+v", i.Frontmatter)
	}
}

func TestDoneRendersCompletedAt(t *testing.T) {
	got, err := issue.Render(populated().Done("2026-08-22T12:00"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "completed_at: 2026-08-22T12:00") {
		t.Errorf("rendered issue misses completed_at:\n%s", got)
	}
}

func TestReopenClearsTimestampsAndSetsOpen(t *testing.T) {
	i := populated()
	got := i.Reopen()

	if got.Frontmatter.Status != "open" {
		t.Errorf("Status = %q, want %q", got.Frontmatter.Status, "open")
	}
	if got.Frontmatter.CompletedAt != "" {
		t.Errorf("CompletedAt = %q, want cleared", got.Frontmatter.CompletedAt)
	}
	if got.Frontmatter.StartedAt != "" {
		t.Errorf("StartedAt = %q, want cleared", got.Frontmatter.StartedAt)
	}
	if got.Frontmatter.Title != "t" || got.Frontmatter.Rank == nil || *got.Frontmatter.Rank != 2 || got.Frontmatter.DeferredUntil != "2026-08-20T08:00" {
		t.Errorf("Reopen changed unrelated fields: %+v", got.Frontmatter)
	}
	if i.Frontmatter.Status != "in_progress" || i.Frontmatter.CompletedAt != "2026-08-22T10:00" {
		t.Errorf("Reopen mutated the receiver: %+v", i.Frontmatter)
	}
}

func TestReopenDropsTimestampsFromDisk(t *testing.T) {
	got, err := issue.Render(populated().Reopen())
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, absent := range []string{"completed_at:", "started_at:"} {
		if strings.Contains(s, absent) {
			t.Errorf("rendered reopened issue contains %q, want it omitted:\n%s", absent, s)
		}
	}
}

func TestStartStampsStartedAtAndSetsInProgress(t *testing.T) {
	i := populated()
	got := i.Start("2026-08-22T12:00")

	if got.Frontmatter.Status != "in_progress" {
		t.Errorf("Status = %q, want %q", got.Frontmatter.Status, "in_progress")
	}
	if got.Frontmatter.StartedAt != "2026-08-22T12:00" {
		t.Errorf("StartedAt = %q, want the stamped time", got.Frontmatter.StartedAt)
	}
	if got.Frontmatter.CompletedAt != "2026-08-22T10:00" || got.Frontmatter.Rank == nil || *got.Frontmatter.Rank != 2 {
		t.Errorf("Start changed unrelated fields: %+v", got.Frontmatter)
	}
	if i.Frontmatter.Status != "in_progress" || i.Frontmatter.StartedAt != "2026-08-21T09:00" {
		t.Errorf("Start mutated the receiver: %+v", i.Frontmatter)
	}
}

func TestStartRendersStartedAt(t *testing.T) {
	got, err := issue.Render(populated().Start("2026-08-22T12:00"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "started_at: 2026-08-22T12:00") {
		t.Errorf("rendered issue misses started_at:\n%s", got)
	}
}

func TestSetStatusChangesOnlyStatus(t *testing.T) {
	i := populated()
	got := i.SetStatus("blocked")

	if got.Frontmatter.Status != "blocked" {
		t.Errorf("Status = %q, want %q", got.Frontmatter.Status, "blocked")
	}
	// Free transition: no timestamps touched.
	if got.Frontmatter.CompletedAt != "2026-08-22T10:00" || got.Frontmatter.StartedAt != "2026-08-21T09:00" {
		t.Errorf("SetStatus touched timestamps: completed=%q started=%q", got.Frontmatter.CompletedAt, got.Frontmatter.StartedAt)
	}
	if got.Frontmatter.Title != "t" || got.Frontmatter.Rank == nil || *got.Frontmatter.Rank != 2 {
		t.Errorf("SetStatus changed unrelated fields: %+v", got.Frontmatter)
	}
	if i.Frontmatter.Status != "in_progress" {
		t.Errorf("SetStatus mutated the receiver: %+v", i.Frontmatter)
	}
}
