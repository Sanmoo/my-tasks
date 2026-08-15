package issue_test

import (
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

func TestDeferSetsOnlyDeferredUntil(t *testing.T) {
	i := populated()
	got := i.Defer("2026-08-20T08:00")

	if got.Frontmatter.DeferredUntil != "2026-08-20T08:00" {
		t.Errorf("DeferredUntil = %q, want the deferred-until value", got.Frontmatter.DeferredUntil)
	}
	// Deferral is data, not a state: status stays exactly as it was.
	if got.Frontmatter.Status != "in_progress" {
		t.Errorf("Status = %q, want it preserved (deferral is not a state)", got.Frontmatter.Status)
	}
	// Every other field is untouched.
	if got.Frontmatter.Title != "t" || got.Frontmatter.Rank == nil || *got.Frontmatter.Rank != 2 ||
		got.Frontmatter.Deadline != "2026-08-22T18:00" || got.Frontmatter.StartedAt != "2026-08-21T09:00" ||
		got.Frontmatter.CompletedAt != "2026-08-22T10:00" {
		t.Errorf("Defer changed unrelated fields: %+v", got.Frontmatter)
	}
	// The receiver is untouched (value receiver).
	if i.Frontmatter.DeferredUntil != "2026-08-20T08:00" {
		t.Errorf("Defer mutated the receiver: %+v", i.Frontmatter)
	}
}

func TestDeferRendersDeferredUntil(t *testing.T) {
	got, err := issue.Render(populated().Defer("2026-08-20T08:00"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "deferred_until: 2026-08-20T08:00") {
		t.Errorf("rendered issue misses deferred_until:\n%s", got)
	}
}
