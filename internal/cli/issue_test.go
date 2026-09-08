package cli_test

// Black-box unit tests of the create/q queue-placement flags, run through
// the exported cli.Run seam (the same entry point as the real process) —
// the command wiring itself stays covered by the e2e suite.

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/cli"
	"github.com/Sanmoo/my-tasks2/internal/issue"
)

func intPtr(n int) *int { return &n }

// runMT runs mt in-process and returns the exit code with both streams.
func runMT(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := cli.Run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// newVault creates a temporary Vault with the pkm prefix and returns its path.
func newVault(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "vault")
	code, _, stderr := runMT(t, "init", "--prefix", "pkm", dir)
	if code != 0 {
		t.Fatalf("mt init: exit %d, stderr: %s", code, stderr)
	}
	return dir
}

// seedIssue writes an Issue file directly into the vault, with the given
// rank (nil = Backlog).
func seedIssue(t *testing.T, vaultDir, id, title, createdAt string, rank *int) {
	t.Helper()
	data, err := issue.Render(issue.Issue{
		Frontmatter: issue.Frontmatter{
			Title:     title,
			Status:    "open",
			Labels:    []string{},
			CreatedAt: createdAt,
			Rank:      rank,
		},
		Body: issue.DefaultBody,
	})
	if err != nil {
		t.Fatalf("rendering %s: %v", id, err)
	}
	if err := os.WriteFile(filepath.Join(vaultDir, "issues", id+".md"), data, 0o644); err != nil {
		t.Fatalf("writing %s: %v", id, err)
	}
}

// rankOf reads an Issue file back and returns its rank (nil = Backlog).
func rankOf(t *testing.T, vaultDir, id string) *int {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(vaultDir, "issues", id+".md"))
	if err != nil {
		t.Fatalf("reading %s: %v", id, err)
	}
	i, err := issue.Parse(data)
	if err != nil {
		t.Fatalf("parsing %s: %v", id, err)
	}
	return i.Frontmatter.Rank
}

// wantRank asserts the Issue's rank equals n.
func wantRank(t *testing.T, vaultDir, id string, n int) {
	t.Helper()
	got := rankOf(t, vaultDir, id)
	if got == nil {
		t.Errorf("issue %s has no rank, want %d", id, n)
		return
	}
	if *got != n {
		t.Errorf("issue %s rank = %d, want %d", id, *got, n)
	}
}

// createdID extracts the new Issue ID from a create/q output line
// ("Created <id> (rank <n>)", "Created <id>" or the bare quiet ID).
func createdID(t *testing.T, stdout string) string {
	t.Helper()
	s := strings.TrimSpace(stdout)
	s = strings.TrimPrefix(s, "Created ")
	if i := strings.IndexByte(s, ' '); i >= 0 {
		s = s[:i]
	}
	if s == "" {
		t.Fatalf("no issue ID in output %q", stdout)
	}
	return s
}

func TestCreateBottomPutsNewIssueAtTheQueueEnd(t *testing.T) {
	vaultDir := newVault(t)
	seedIssue(t, vaultDir, "pkm-001", "first", "2026-01-01T10:00", intPtr(1))
	seedIssue(t, vaultDir, "pkm-002", "second", "2026-01-02T10:00", intPtr(2))

	code, stdout, stderr := runMT(t, "create", "--vault", vaultDir, "--bottom", "novo")
	if code != 0 {
		t.Fatalf("create --bottom: exit %d, stderr: %s", code, stderr)
	}
	id := createdID(t, stdout)
	if want := "Created " + id + " (rank 3)\n"; stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	wantRank(t, vaultDir, id, 3)
	// The neighbors keep their ranks: only the new file is rewritten.
	wantRank(t, vaultDir, "pkm-001", 1)
	wantRank(t, vaultDir, "pkm-002", 2)
}

func TestCreateTopPutsNewIssueFirstAndShiftsQueue(t *testing.T) {
	vaultDir := newVault(t)
	seedIssue(t, vaultDir, "pkm-001", "first", "2026-01-01T10:00", intPtr(1))
	seedIssue(t, vaultDir, "pkm-002", "second", "2026-01-02T10:00", intPtr(2))

	code, stdout, stderr := runMT(t, "create", "--vault", vaultDir, "--top", "novo")
	if code != 0 {
		t.Fatalf("create --top: exit %d, stderr: %s", code, stderr)
	}
	id := createdID(t, stdout)
	if want := "Created " + id + " (rank 1)\n"; stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
	wantRank(t, vaultDir, id, 1)
	wantRank(t, vaultDir, "pkm-001", 2)
	wantRank(t, vaultDir, "pkm-002", 3)
}

// TestCreatePlacementEqualsCreateThenQuickOrder pins the single source of
// truth: creating with --top/--bottom must leave the vault exactly where
// creating without the flag and running mt top/bottom would.
func TestCreatePlacementEqualsCreateThenQuickOrder(t *testing.T) {
	for _, tc := range []struct {
		action string
	}{
		{"top"},
		{"bottom"},
	} {
		t.Run(tc.action, func(t *testing.T) {
			withFlag := newVault(t)
			seedIssue(t, withFlag, "pkm-001", "first", "2026-01-01T10:00", intPtr(1))
			seedIssue(t, withFlag, "pkm-002", "second", "2026-01-02T10:00", intPtr(2))

			thenQuick := newVault(t)
			seedIssue(t, thenQuick, "pkm-001", "first", "2026-01-01T10:00", intPtr(1))
			seedIssue(t, thenQuick, "pkm-002", "second", "2026-01-02T10:00", intPtr(2))

			code, stdout, stderr := runMT(t, "create", "--vault", withFlag, "--"+tc.action, "novo")
			if code != 0 {
				t.Fatalf("create --%s: exit %d, stderr: %s", tc.action, code, stderr)
			}

			code, stdout, stderr = runMT(t, "create", "--vault", thenQuick, "novo")
			if code != 0 {
				t.Fatalf("create: exit %d, stderr: %s", code, stderr)
			}
			id := createdID(t, stdout)
			code, _, stderr = runMT(t, tc.action, "--vault", thenQuick, id)
			if code != 0 {
				t.Fatalf("%s: exit %d, stderr: %s", tc.action, code, stderr)
			}

			got := ranksByTitle(t, withFlag)
			want := ranksByTitle(t, thenQuick)
			if len(got) != len(want) {
				t.Fatalf("ranks by title = %v, want %v", got, want)
			}
			for title, rank := range want {
				if got[title] != rank {
					t.Errorf("title %q: create --%s rank = %d, create+%s rank = %d",
						title, tc.action, got[title], tc.action, rank)
				}
			}
		})
	}
}

func TestCreatePlacementOnEmptyQueue(t *testing.T) {
	for _, tc := range []struct {
		flag string
	}{
		{"top"},
		{"bottom"},
	} {
		t.Run(tc.flag, func(t *testing.T) {
			vaultDir := newVault(t)
			// The shorthand form (-t/-b) doubles as shorthand coverage.
			code, stdout, stderr := runMT(t, "create", "--vault", vaultDir, "-"+tc.flag[:1], "novo")
			if code != 0 {
				t.Fatalf("create -%s: exit %d, stderr: %s", tc.flag[:1], code, stderr)
			}
			id := createdID(t, stdout)
			if want := "Created " + id + " (rank 1)\n"; stdout != want {
				t.Errorf("stdout = %q, want %q", stdout, want)
			}
			wantRank(t, vaultDir, id, 1)
		})
	}
}

func TestCreateRejectsTopAndBottomTogether(t *testing.T) {
	for _, cmd := range []string{"create", "q"} {
		t.Run(cmd, func(t *testing.T) {
			vaultDir := newVault(t)
			code, stdout, stderr := runMT(t, cmd, "--vault", vaultDir, "-t", "-b", "x")
			if code != 2 {
				t.Fatalf("%s --top --bottom: exit %d, want 2 (stderr: %s)", cmd, code, stderr)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want empty", stdout)
			}
			if !strings.Contains(stderr, "--top and --bottom") {
				t.Errorf("stderr = %q, want it to name the conflicting flags", stderr)
			}
			entries, err := os.ReadDir(filepath.Join(vaultDir, "issues"))
			if err != nil {
				t.Fatalf("reading issues directory: %v", err)
			}
			if len(entries) != 0 {
				t.Errorf("issues directory has %d files, want 0 — nothing should be created", len(entries))
			}
		})
	}
}

func TestQPlacementKeepsQuietOutput(t *testing.T) {
	t.Run("top", func(t *testing.T) {
		vaultDir := newVault(t)
		seedIssue(t, vaultDir, "pkm-001", "first", "2026-01-01T10:00", intPtr(1))

		code, stdout, stderr := runMT(t, "q", "--vault", vaultDir, "--top", "ideia")
		if code != 0 {
			t.Fatalf("q --top: exit %d, stderr: %s", code, stderr)
		}
		if !regexp.MustCompile(`^pkm-[0-9a-z]{4}\n$`).MatchString(stdout) {
			t.Errorf("stdout = %q, want only the ID", stdout)
		}
		id := createdID(t, stdout)
		wantRank(t, vaultDir, id, 1)
		wantRank(t, vaultDir, "pkm-001", 2)
	})

	t.Run("bottom", func(t *testing.T) {
		vaultDir := newVault(t)
		seedIssue(t, vaultDir, "pkm-001", "first", "2026-01-01T10:00", intPtr(1))

		code, stdout, stderr := runMT(t, "q", "--vault", vaultDir, "--bottom", "ideia")
		if code != 0 {
			t.Fatalf("q --bottom: exit %d, stderr: %s", code, stderr)
		}
		if !regexp.MustCompile(`^pkm-[0-9a-z]{4}\n$`).MatchString(stdout) {
			t.Errorf("stdout = %q, want only the ID", stdout)
		}
		id := createdID(t, stdout)
		wantRank(t, vaultDir, id, 2)
		wantRank(t, vaultDir, "pkm-001", 1)
	})
}

// ranksByTitle maps every ranked Issue's title to its rank in the vault.
func ranksByTitle(t *testing.T, vaultDir string) map[string]int {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(vaultDir, "issues"))
	if err != nil {
		t.Fatalf("reading issues directory: %v", err)
	}
	ranks := make(map[string]int, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(vaultDir, "issues", name))
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		i, err := issue.Parse(data)
		if err != nil {
			t.Fatalf("parsing %s: %v", name, err)
		}
		if i.Frontmatter.Rank != nil {
			ranks[i.Frontmatter.Title] = *i.Frontmatter.Rank
		}
	}
	return ranks
}
