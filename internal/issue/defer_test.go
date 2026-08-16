package issue_test

import (
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

func TestDeferSetsDeferredUntilAndOpensIssue(t *testing.T) {
	i := populated()
	got := i.Defer("2026-08-20T08:00")

	if got.Frontmatter.DeferredUntil != "2026-08-20T08:00" {
		t.Errorf("DeferredUntil = %q, want the deferred-until value", got.Frontmatter.DeferredUntil)
	}
	// A deferred Issue remains open; deferral is not a separate status.
	if got.Frontmatter.Status != "open" {
		t.Errorf("Status = %q, want open", got.Frontmatter.Status)
	}
	// Every field other than Status and DeferredUntil is untouched.
	if got.Frontmatter.Title != "t" || got.Frontmatter.Rank == nil || *got.Frontmatter.Rank != 2 ||
		got.Frontmatter.Deadline != "2026-08-22T18:00" || got.Frontmatter.StartedAt != "2026-08-21T09:00" ||
		got.Frontmatter.CompletedAt != "2026-08-22T10:00" {
		t.Errorf("Defer changed unrelated fields: %+v", got.Frontmatter)
	}
	// The receiver is untouched (value receiver).
	if i.Frontmatter.Status != "in_progress" || i.Frontmatter.DeferredUntil != "2026-08-20T08:00" {
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

func TestUndeferClearsOnlyDeferredUntil(t *testing.T) {
	i := populated()
	got := i.Undefer()

	if got.Frontmatter.DeferredUntil != "" {
		t.Errorf("DeferredUntil = %q, want cleared", got.Frontmatter.DeferredUntil)
	}

	// Undefer has no opinion on status or rank — every other field is
	// untouched, including a non-open Status.
	if got.Frontmatter.Status != "in_progress" || got.Frontmatter.Rank == nil || *got.Frontmatter.Rank != 2 ||
		got.Frontmatter.Deadline != "2026-08-22T18:00" || got.Frontmatter.StartedAt != "2026-08-21T09:00" ||
		got.Frontmatter.CompletedAt != "2026-08-22T10:00" {
		t.Errorf("Undefer changed unrelated fields: %+v", got.Frontmatter)
	}
	// The receiver is untouched (value receiver).
	if i.Frontmatter.DeferredUntil != "2026-08-20T08:00" {
		t.Errorf("Undefer mutated the receiver: %+v", i.Frontmatter)
	}
}

func TestUndeferRendersWithoutDeferredUntil(t *testing.T) {
	got, err := issue.Render(populated().Undefer())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "deferred_until") {
		t.Errorf("rendered issue still carries deferred_until:\n%s", got)
	}
}
