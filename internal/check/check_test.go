package check_test

import (
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/check"
	"github.com/Sanmoo/my-tasks2/internal/issue"
)

func intPtr(value int) *int { return &value }

func item(id, status string, rank *int, createdAt string) check.Item {
	return check.Item{
		ID: id,
		Issue: issue.Issue{Frontmatter: issue.Frontmatter{
			Title:     "title",
			Status:    status,
			Labels:    []string{},
			CreatedAt: createdAt,
			Rank:      rank,
		}},
	}
}

func TestRankGapRanges(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	tests := []struct {
		name  string
		items []check.Item
		want  []check.RankGap
	}{
		{"empty", nil, nil},
		{"contiguous", []check.Item{item("a", "open", intPtr(1), ""), item("b", "open", intPtr(2), "")}, nil},
		{"missing first", []check.Item{item("b", "open", intPtr(2), "")}, []check.RankGap{{Start: 1, End: 1}}},
		{"missing middle", []check.Item{item("a", "open", intPtr(1), ""), item("c", "open", intPtr(3), "")}, []check.RankGap{{Start: 2, End: 2}}},
		{"multiple ranges", []check.Item{item("a", "open", intPtr(1), ""), item("c", "open", intPtr(3), ""), item("e", "open", intPtr(5), "")}, []check.RankGap{{Start: 2, End: 2}, {Start: 4, End: 4}}},
		{"duplicate and backlog", []check.Item{item("a", "open", intPtr(1), ""), item("b", "open", intPtr(1), ""), item("c", "open", nil, "")}, nil},
		{"non-positive values ignored", []check.Item{item("a", "open", intPtr(-1), ""), item("b", "open", intPtr(0), ""), item("c", "open", intPtr(1), "")}, nil},
		{"huge range stays compact", []check.Item{item("a", "open", intPtr(1), ""), item("z", "open", &maxInt, "")}, []check.RankGap{{Start: 2, End: maxInt - 1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := check.RankGapRanges(tt.items)
			if len(got) != len(tt.want) {
				t.Fatalf("RankGapRanges() = %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("RankGapRanges()[%d] = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestNonPositiveRanks(t *testing.T) {
	items := []check.Item{
		item("zero", "open", intPtr(0), ""),
		item("negative", "open", intPtr(-1), ""),
		item("zero-again", "open", intPtr(0), ""),
		item("positive", "open", intPtr(1), ""),
	}
	got := check.NonPositiveRanks(items)
	want := []int{-1, 0}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("NonPositiveRanks() = %v, want %v", got, want)
	}
	if got := check.NonPositiveRanks(nil); len(got) != 0 {
		t.Errorf("NonPositiveRanks(nil) = %v, want empty", got)
	}
}

func TestDuplicateRanks(t *testing.T) {
	items := []check.Item{
		item("r3a", "open", intPtr(3), ""),
		item("r1a", "open", intPtr(1), ""),
		item("backlog", "open", nil, ""),
		item("r1b", "open", intPtr(1), ""),
		item("r3b", "open", intPtr(3), ""),
	}
	got := check.DuplicateRanks(items)
	want := []int{1, 3}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("DuplicateRanks() = %v, want %v", got, want)
	}
	if got := check.DuplicateRanks(nil); len(got) != 0 {
		t.Errorf("DuplicateRanks(nil) = %v, want empty", got)
	}
}

func validFile() []byte {
	return []byte("---\ntitle: title\nstatus: open\nlabels: []\ncreated_at: 2026-01-01T10:00\nrank: 2\n---\nbody\n")
}

func TestValidateFrontmatter(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"valid", validFile(), ""},
		{"valid with blocked_by", []byte("---\ntitle: title\nstatus: open\nlabels: []\ncreated_at: 2026-01-01T10:00\nblocked_by: [pkm-001]\n---\nbody\n"), ""},
		{"empty blocked_by scalar is an empty optional field", []byte("---\ntitle: title\nstatus: open\nlabels: []\ncreated_at: 2026-01-01T10:00\nblocked_by:\n---\nbody\n"), "empty optional field"},
		{"missing opening delimiter", []byte("title: title\n---\n"), "start with"},
		{"missing closing delimiter", []byte("---\ntitle: title\n"), "not closed"},
		{"malformed YAML", []byte("---\nlabels: [unclosed\n---\n"), "malformed frontmatter"},
		{"non-mapping YAML", []byte("---\n- title\n---\n"), "YAML mapping"},
		{"unknown field", []byte("---\ntitle: title\nstatus: open\nlabels: []\ncreated_at: 2026-01-01T10:00\nid: wrong\n---\n"), "unknown field"},
		{"empty optional field", []byte("---\ntitle: title\nstatus: open\nlabels: []\ncreated_at: 2026-01-01T10:00\nrank:\n---\n"), "empty optional field"},
		{"fractional rank", []byte("---\ntitle: title\nstatus: open\nlabels: []\ncreated_at: 2026-01-01T10:00\nrank: 1.2\n---\n"), "rank must be an integer"},
		{"trailing YAML after document marker", []byte("---\ntitle: title\nstatus: open\nlabels: []\ncreated_at: 2026-01-01T10:00\n...\nother: value\n---\n"), "malformed frontmatter"},
		{"malformed trailing YAML", []byte("---\ntitle: title\nstatus: open\nlabels: []\ncreated_at: 2026-01-01T10:00\n...\nother: [unclosed\n---\n"), "malformed frontmatter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := check.ValidateFrontmatter(tt.data, "pkm-001")
			if tt.want == "" {
				if err != nil {
					t.Fatalf("ValidateFrontmatter() = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("ValidateFrontmatter() error = %v, want substring %q", err, tt.want)
			}
		})
	}
}

func TestValidateItem(t *testing.T) {
	statuses := []string{"open", "in_progress", "done"}
	base := item("pkm-001", "open", nil, "2026-01-01T10:00")
	base.Issue.Frontmatter.DeferredUntil = "2026-01-02T10:00"
	base.Issue.Frontmatter.Deadline = "2026-01-03T10:00"
	base.Issue.Frontmatter.StartedAt = "2026-01-04T10:00"
	base.Issue.Frontmatter.CompletedAt = "2026-01-05T10:00"
	if err := check.ValidateItem(base, statuses); err != nil {
		t.Fatalf("ValidateItem(valid) = %v, want nil", err)
	}

	tests := []struct {
		name string
		item check.Item
		want string
	}{
		{"missing title", func() check.Item { x := base; x.Issue.Frontmatter.Title = ""; return x }(), "missing title"},
		{"missing status", func() check.Item { x := base; x.Issue.Frontmatter.Status = ""; return x }(), "missing status"},
		{"missing labels", func() check.Item { x := base; x.Issue.Frontmatter.Labels = nil; return x }(), "missing labels"},
		{"missing created_at", func() check.Item { x := base; x.Issue.Frontmatter.CreatedAt = ""; return x }(), "missing created_at"},
		{"invalid status", func() check.Item { x := base; x.Issue.Frontmatter.Status = "blocked"; return x }(), "not configured"},
		{"invalid created_at", func() check.Item { x := base; x.Issue.Frontmatter.CreatedAt = "2026-01-01"; return x }(), "created_at"},
		{"invalid deferred_until", func() check.Item { x := base; x.Issue.Frontmatter.DeferredUntil = "bad"; return x }(), "deferred_until"},
		{"invalid deadline", func() check.Item { x := base; x.Issue.Frontmatter.Deadline = "bad"; return x }(), "deadline"},
		{"invalid started_at", func() check.Item { x := base; x.Issue.Frontmatter.StartedAt = "bad"; return x }(), "started_at"},
		{"invalid completed_at", func() check.Item { x := base; x.Issue.Frontmatter.CompletedAt = "bad"; return x }(), "completed_at"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := check.ValidateItem(tt.item, statuses); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("ValidateItem() error = %v, want substring %q", err, tt.want)
			}
		})
	}
	if err := check.ValidateItem(base, nil); err == nil || !strings.Contains(err.Error(), "valid:") {
		t.Errorf("ValidateItem(empty statuses) = %v, want status-list error", err)
	}
}

// blockedByItem builds an Item with a blocked_by list.
func blockedByItem(id string, blockedBy []string) check.Item {
	it := item(id, "open", nil, "2026-01-01T10:00")
	it.Issue.Frontmatter.BlockedBy = blockedBy
	return it
}

func TestValidateBlockedBy(t *testing.T) {
	t.Run("valid vault", func(t *testing.T) {
		items := []check.Item{
			blockedByItem("pkm-001", nil),
			blockedByItem("pkm-002", []string{"pkm-001"}),
		}
		if err := check.ValidateBlockedBy(items); err != nil {
			t.Errorf("ValidateBlockedBy() = %v, want nil", err)
		}
	})

	t.Run("reference to a done Issue is valid", func(t *testing.T) {
		items := []check.Item{
			blockedByItem("pkm-001", nil),
			blockedByItem("pkm-002", []string{"pkm-001"}),
		}
		items[0].Issue.Frontmatter.Status = "done"
		if err := check.ValidateBlockedBy(items); err != nil {
			t.Errorf("ValidateBlockedBy() = %v, want nil", err)
		}
	})

	t.Run("unknown reference", func(t *testing.T) {
		items := []check.Item{
			blockedByItem("pkm-001", nil),
			blockedByItem("pkm-002", []string{"pkm-999"}),
		}
		err := check.ValidateBlockedBy(items)
		if err == nil || !strings.Contains(err.Error(), "pkm-002") || !strings.Contains(err.Error(), "pkm-999") {
			t.Errorf("ValidateBlockedBy() error = %v, want unknown-reference error naming both IDs", err)
		}
	})

	t.Run("self-block", func(t *testing.T) {
		items := []check.Item{
			blockedByItem("pkm-001", []string{"pkm-001"}),
		}
		err := check.ValidateBlockedBy(items)
		if err == nil || !strings.Contains(err.Error(), "pkm-001") || !strings.Contains(err.Error(), "itself") {
			t.Errorf("ValidateBlockedBy() error = %v, want self-block error", err)
		}
	})

	t.Run("two-cycle", func(t *testing.T) {
		items := []check.Item{
			blockedByItem("pkm-001", []string{"pkm-002"}),
			blockedByItem("pkm-002", []string{"pkm-001"}),
		}
		err := check.ValidateBlockedBy(items)
		if err == nil || !strings.Contains(err.Error(), "cycle") ||
			!strings.Contains(err.Error(), "pkm-001 -> pkm-002 -> pkm-001") {
			t.Errorf("ValidateBlockedBy() error = %v, want the two-cycle path", err)
		}
	})

	t.Run("three-cycle", func(t *testing.T) {
		items := []check.Item{
			blockedByItem("pkm-001", []string{"pkm-002"}),
			blockedByItem("pkm-002", []string{"pkm-003"}),
			blockedByItem("pkm-003", []string{"pkm-001"}),
		}
		err := check.ValidateBlockedBy(items)
		if err == nil || !strings.Contains(err.Error(), "cycle") ||
			!strings.Contains(err.Error(), "pkm-001 -> pkm-002 -> pkm-003 -> pkm-001") {
			t.Errorf("ValidateBlockedBy() error = %v, want the three-cycle path", err)
		}
	})

	t.Run("cycle reachable only through a non-cycle start", func(t *testing.T) {
		items := []check.Item{
			blockedByItem("pkm-001", []string{"pkm-002"}),
			blockedByItem("pkm-002", []string{"pkm-003"}),
			blockedByItem("pkm-003", []string{"pkm-004"}),
			blockedByItem("pkm-004", []string{"pkm-003"}),
		}
		err := check.ValidateBlockedBy(items)
		if err == nil || !strings.Contains(err.Error(), "cycle") ||
			!strings.Contains(err.Error(), "pkm-003 -> pkm-004 -> pkm-003") {
			t.Errorf("ValidateBlockedBy() error = %v, want the inner cycle", err)
		}
	})

	t.Run("cycle in a later Issue after an acyclic start", func(t *testing.T) {
		// The first Item is acyclic, so a walk that stops at the first
		// visited Item would miss the cycle between pkm-002 and pkm-003.
		items := []check.Item{
			blockedByItem("pkm-001", nil),
			blockedByItem("pkm-002", []string{"pkm-003"}),
			blockedByItem("pkm-003", []string{"pkm-002"}),
		}
		err := check.ValidateBlockedBy(items)
		if err == nil || !strings.Contains(err.Error(), "cycle") ||
			!strings.Contains(err.Error(), "pkm-002 -> pkm-003 -> pkm-002") {
			t.Errorf("ValidateBlockedBy() error = %v, want the later cycle", err)
		}
	})

	t.Run("acyclic chain plus a reference back stays valid", func(t *testing.T) {
		// pkm-001 -> pkm-002 -> pkm-003 is a plain chain and pkm-004
		// references pkm-001; nothing is cyclic. A walk that abandons
		// the chain after its first acyclic subtree would report a
		// false cycle through the abandoned path.
		items := []check.Item{
			blockedByItem("pkm-001", []string{"pkm-002"}),
			blockedByItem("pkm-002", []string{"pkm-003"}),
			blockedByItem("pkm-003", nil),
			blockedByItem("pkm-004", []string{"pkm-001"}),
		}
		if err := check.ValidateBlockedBy(items); err != nil {
			t.Errorf("ValidateBlockedBy() = %v, want nil", err)
		}
	})

	t.Run("unknown reference is reported before a cycle", func(t *testing.T) {
		items := []check.Item{
			blockedByItem("pkm-001", []string{"pkm-002"}),
			blockedByItem("pkm-002", []string{"pkm-001"}),
			blockedByItem("pkm-003", []string{"pkm-999"}),
		}
		err := check.ValidateBlockedBy(items)
		if err == nil || !strings.Contains(err.Error(), "unknown issue pkm-999") {
			t.Errorf("ValidateBlockedBy() error = %v, want the unknown reference first", err)
		}
	})
}
