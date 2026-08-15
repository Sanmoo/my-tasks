// Package issue_test holds the black-box unit tests of the Issue
// frontmatter round-trip (Seam 2): stable field order, optional fields
// only-when-set, and byte-exact round-trip through Render/Parse.
package issue_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

func intPtr(v int) *int { return &v }

// specExample is the Issue exactly as the spec schema example renders it,
// minus the comment section (comments are a separate feature). It is the
// canonical fixture for round-trip tests.
const specExample = `---
title: '[Niver Edu] comprar material de Assaí conforme doc compartilhado'
status: open
labels: [compras, familia]
created_at: 2026-08-15T09:30
rank: 2
deferred_until: 2026-08-20T08:00
deadline: 2026-08-22T18:00
---

## Description
## Notes
## Comments
`

func TestRenderMatchesSpecExample(t *testing.T) {
	i := issue.Issue{
		Frontmatter: issue.Frontmatter{
			Title:         "[Niver Edu] comprar material de Assaí conforme doc compartilhado",
			Status:        "open",
			Labels:        []string{"compras", "familia"},
			CreatedAt:     "2026-08-15T09:30",
			Rank:          intPtr(2),
			DeferredUntil: "2026-08-20T08:00",
			Deadline:      "2026-08-22T18:00",
		},
		Body: issue.DefaultBody,
	}
	got, err := issue.Render(i)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != specExample {
		t.Errorf("Render = %q, want the spec schema exactly %q", got, specExample)
	}
}

func TestRenderOmitsOptionalFieldsWhenUnset(t *testing.T) {
	i := issue.Issue{
		Frontmatter: issue.Frontmatter{
			Title:     "Buy milk",
			Status:    "open",
			Labels:    []string{"compras"},
			CreatedAt: "2026-08-15T09:30",
		},
		Body: issue.DefaultBody,
	}
	got, err := issue.Render(i)
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	for _, absent := range []string{"rank:", "deferred_until:", "deadline:", "started_at:", "completed_at:", "id:", "updated_at:"} {
		if strings.Contains(s, absent) {
			t.Errorf("Render = %q contains %q, want it omitted", s, absent)
		}
	}
}

func TestRenderFieldOrderIsStable(t *testing.T) {
	r := 2
	i := issue.Issue{
		Frontmatter: issue.Frontmatter{
			Title:         "t",
			Status:        "open",
			Labels:        []string{"a"},
			CreatedAt:     "2026-08-15T09:30",
			Rank:          &r,
			DeferredUntil: "2026-08-20T08:00",
			Deadline:      "2026-08-22T18:00",
			StartedAt:     "2026-08-21T09:00",
			CompletedAt:   "2026-08-22T10:00",
		},
		Body: issue.DefaultBody,
	}
	got, err := issue.Render(i)
	if err != nil {
		t.Fatal(err)
	}
	order := []string{"title:", "status:", "labels:", "created_at:", "rank:", "deferred_until:", "deadline:", "started_at:", "completed_at:"}
	s := string(got)
	prev := -1
	for _, key := range order {
		idx := strings.Index(s, key)
		if idx < 0 {
			t.Fatalf("Render = %q misses field %q", s, key)
		}
		if idx < prev {
			t.Errorf("field %q out of order in %q", key, s)
		}
		prev = idx
	}
}

func TestRenderEmptyLabelsIsPresentEmptyFlow(t *testing.T) {
	i := issue.Issue{
		Frontmatter: issue.Frontmatter{Title: "x", Status: "open", CreatedAt: "2026-08-15T09:30"},
		Body:        issue.DefaultBody,
	}
	got, err := issue.Render(i)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "labels: []") {
		t.Errorf("Render = %q, want labels always present as []", got)
	}
}

func TestRoundTripIsByteExact(t *testing.T) {
	in := []byte(specExample)
	i, err := issue.Parse(in)
	if err != nil {
		t.Fatal(err)
	}
	got, err := issue.Render(i)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, in) {
		t.Errorf("round-trip changed bytes:\n got: %q\nwant: %q", got, in)
	}
}

func TestParseReturnsFieldsAndBody(t *testing.T) {
	i, err := issue.Parse([]byte(specExample))
	if err != nil {
		t.Fatal(err)
	}
	fm := i.Frontmatter
	if fm.Title != "[Niver Edu] comprar material de Assaí conforme doc compartilhado" {
		t.Errorf("Title = %q", fm.Title)
	}
	if fm.Status != "open" {
		t.Errorf("Status = %q", fm.Status)
	}
	if len(fm.Labels) != 2 || fm.Labels[0] != "compras" || fm.Labels[1] != "familia" {
		t.Errorf("Labels = %v", fm.Labels)
	}
	if fm.CreatedAt != "2026-08-15T09:30" {
		t.Errorf("CreatedAt = %q", fm.CreatedAt)
	}
	if fm.Rank == nil || *fm.Rank != 2 {
		t.Errorf("Rank = %v, want 2", fm.Rank)
	}
	if fm.DeferredUntil != "2026-08-20T08:00" {
		t.Errorf("DeferredUntil = %q", fm.DeferredUntil)
	}
	if fm.Deadline != "2026-08-22T18:00" {
		t.Errorf("Deadline = %q", fm.Deadline)
	}
	if fm.StartedAt != "" || fm.CompletedAt != "" {
		t.Errorf("started/completed = %q/%q, want empty", fm.StartedAt, fm.CompletedAt)
	}
	if i.Body != issue.DefaultBody {
		t.Errorf("Body = %q, want %q", i.Body, issue.DefaultBody)
	}
}

func TestParsePreservesBodyWithoutTrailingNewline(t *testing.T) {
	in := "---\ntitle: t\nstatus: open\nlabels: []\ncreated_at: 2026-08-15T09:30\n---\n## Description\nno trailing newline"
	i, err := issue.Parse([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	if i.Body != "## Description\nno trailing newline" {
		t.Errorf("Body = %q", i.Body)
	}
	got, err := issue.Render(i)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != in {
		t.Errorf("round-trip without trailing newline = %q, want %q", got, in)
	}
}

func TestParseMissingOpeningDelimiterFails(t *testing.T) {
	if _, err := issue.Parse([]byte("title: t\n---\nbody\n")); err == nil {
		t.Fatal("Parse without leading --- = nil error, want failure")
	}
}

func TestParseEmptyFails(t *testing.T) {
	if _, err := issue.Parse(nil); err == nil {
		t.Fatal("Parse(nil) = nil error, want failure")
	}
}

func TestParseUnclosedFrontmatterFails(t *testing.T) {
	if _, err := issue.Parse([]byte("---\ntitle: t\n")); err == nil {
		t.Fatal("Parse(unclosed frontmatter) = nil error, want failure")
	}
}

func TestParseMalformedYAMLFails(t *testing.T) {
	if _, err := issue.Parse([]byte("---\nlabels: [unclosed\n---\nbody\n")); err == nil {
		t.Fatal("Parse(malformed yaml) = nil error, want failure")
	}
}

func TestParseEmptyBody(t *testing.T) {
	i, err := issue.Parse([]byte("---\ntitle: t\nstatus: open\nlabels: []\ncreated_at: 2026-08-15T09:30\n---\n"))
	if err != nil {
		t.Fatal(err)
	}
	if i.Body != "" {
		t.Errorf("Body = %q, want empty", i.Body)
	}
}
