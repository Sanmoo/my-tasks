package vault_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/vault"
)

// writeVault writes an mt.yaml with the given prefix into dir,
// returning dir. It is the temp-dir setup of the prefix match tests.
func writeVault(t *testing.T, dir, prefix string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mt.yaml"), []byte("prefix: "+prefix+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func matchPaths(matches []vault.KeyVault) []string {
	paths := make([]string, 0, len(matches))
	for _, m := range matches {
		paths = append(paths, m.Path)
	}
	return paths
}

func TestMatchByPrefixExactMatch(t *testing.T) {
	home := t.TempDir()
	dom := writeVault(t, filepath.Join(home, "dom"), "dom")
	g := vault.Global{Bookmarks: map[string]string{"dom": "~/dom", "pkm": "~/pkm"}}
	matches := vault.MatchByPrefix("dom-xyz", g, home)
	if len(matches) != 1 {
		t.Fatalf("MatchByPrefix(dom-xyz) = %v, want 1 match", matches)
	}
	if matches[0].Bookmark != "dom" {
		t.Errorf("Bookmark = %q, want %q", matches[0].Bookmark, "dom")
	}
	if matches[0].Path != dom {
		t.Errorf("Path = %q, want %q", matches[0].Path, dom)
	}
}

func TestMatchByPrefixOnlyMatchingBookmarkIsReturned(t *testing.T) {
	home := t.TempDir()
	writeVault(t, filepath.Join(home, "dom"), "dom")
	writeVault(t, filepath.Join(home, "pkm"), "pkm")
	g := vault.Global{Bookmarks: map[string]string{"dom": "~/dom", "pkm": "~/pkm"}}
	matches := vault.MatchByPrefix("dom-1", g, home)
	if len(matches) != 1 || matches[0].Bookmark != "dom" {
		t.Fatalf("MatchByPrefix(dom-1) = %v, want only the dom bookmark", matches)
	}
}

func TestMatchByPrefixCaseSensitive(t *testing.T) {
	home := t.TempDir()
	writeVault(t, filepath.Join(home, "dom"), "DOM")
	g := vault.Global{Bookmarks: map[string]string{"dom": "~/dom"}}
	if matches := vault.MatchByPrefix("dom-xyz", g, home); len(matches) != 0 {
		t.Errorf("MatchByPrefix(dom-xyz) = %v against prefix DOM, want no match (case-sensitive)", matches)
	}
	if matches := vault.MatchByPrefix("DOM-xyz", g, home); len(matches) != 1 {
		t.Errorf("MatchByPrefix(DOM-xyz) = %v, want 1 match", matches)
	}
}

func TestMatchByPrefixFirstDashWins(t *testing.T) {
	home := t.TempDir()
	writeVault(t, filepath.Join(home, "dom"), "dom")
	g := vault.Global{Bookmarks: map[string]string{"dom": "~/dom"}}
	if matches := vault.MatchByPrefix("dom-x-y", g, home); len(matches) != 1 {
		t.Errorf("MatchByPrefix(dom-x-y) = %v, want 1 match (prefix dom, not dom-x)", matches)
	}
}

func TestMatchByPrefixNoDash(t *testing.T) {
	home := t.TempDir()
	writeVault(t, filepath.Join(home, "dom"), "dom")
	g := vault.Global{Bookmarks: map[string]string{"dom": "~/dom"}}
	for _, key := range []string{"dom", "legacy", ""} {
		if matches := vault.MatchByPrefix(key, g, home); len(matches) != 0 {
			t.Errorf("MatchByPrefix(%q) = %v, want no match (no '-')", key, matches)
		}
	}
}

func TestMatchByPrefixEmptyPrefix(t *testing.T) {
	home := t.TempDir()
	writeVault(t, filepath.Join(home, "dom"), "dom")
	g := vault.Global{Bookmarks: map[string]string{"dom": "~/dom"}}
	for _, key := range []string{"-xyz", "-"} {
		if matches := vault.MatchByPrefix(key, g, home); len(matches) != 0 {
			t.Errorf("MatchByPrefix(%q) = %v, want no match (empty prefix)", key, matches)
		}
	}
}

func TestMatchByPrefixNoMatchFallsBackEmpty(t *testing.T) {
	home := t.TempDir()
	writeVault(t, filepath.Join(home, "pkm"), "pkm")
	g := vault.Global{Bookmarks: map[string]string{"pkm": "~/pkm"}}
	if matches := vault.MatchByPrefix("dom-xyz", g, home); len(matches) != 0 {
		t.Errorf("MatchByPrefix(dom-xyz) = %v, want no match", matches)
	}
}

func TestMatchByPrefixAmbiguity(t *testing.T) {
	home := t.TempDir()
	writeVault(t, filepath.Join(home, "a"), "dom")
	writeVault(t, filepath.Join(home, "b"), "dom")
	g := vault.Global{Bookmarks: map[string]string{"a": "~/a", "b": "~/b"}}
	matches := vault.MatchByPrefix("dom-xyz", g, home)
	if len(matches) != 2 {
		t.Fatalf("MatchByPrefix(dom-xyz) = %v, want 2 matches", matches)
	}
	// Candidates come in sorted bookmark order, so the message the CLI
	// builds from them is stable.
	if matches[0].Bookmark != "a" || matches[1].Bookmark != "b" {
		t.Errorf("candidates = %+v, want sorted [a b]", matches)
	}
}

func TestMatchByPrefixAliasesDedupeByPath(t *testing.T) {
	home := t.TempDir()
	dom := writeVault(t, filepath.Join(home, "dom"), "dom")
	// Two bookmarks for the same vault (an alias): one candidate, not
	// an ambiguity (story 16).
	g := vault.Global{Bookmarks: map[string]string{"dom": "~/dom", "website": "~/dom"}}
	matches := vault.MatchByPrefix("dom-xyz", g, home)
	if len(matches) != 1 {
		t.Fatalf("MatchByPrefix(dom-xyz) = %v, want 1 candidate (aliases dedupe)", matches)
	}
	if matches[0].Bookmark != "dom" || matches[0].Path != dom {
		t.Errorf("candidate = %+v, want the dom bookmark", matches[0])
	}
}

func TestMatchByPrefixDedupeIsByPathNotByName(t *testing.T) {
	home := t.TempDir()
	writeVault(t, filepath.Join(home, "dom"), "dom")
	writeVault(t, filepath.Join(home, "website"), "website")
	g := vault.Global{Bookmarks: map[string]string{"dom": "~/dom", "website": "~/website"}}
	if matches := vault.MatchByPrefix("dom-xyz", g, home); len(matches) != 1 {
		t.Errorf("MatchByPrefix(dom-xyz) = %v, want only the dom vault (distinct paths keep both)", matches)
	}
}

func TestMatchByPrefixSkippedUnreadableConfigs(t *testing.T) {
	home := t.TempDir()
	// A vault whose mt.yaml cannot be parsed.
	broken := filepath.Join(home, "broken")
	if err := os.MkdirAll(broken, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(broken, "mt.yaml"), []byte("prefix: [unclosed"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A directory that is not a vault at all (no mt.yaml).
	notVault := filepath.Join(home, "novault")
	if err := os.MkdirAll(notVault, 0o755); err != nil {
		t.Fatal(err)
	}
	// A bookmark pointing at a directory that does not exist.
	writeVault(t, filepath.Join(home, "dom"), "dom")
	g := vault.Global{Bookmarks: map[string]string{
		"broken":  "~/broken",
		"novault": "~/novault",
		"missing": "~/missing",
		"dom":     "~/dom",
	}}
	if matches := vault.MatchByPrefix("dom-xyz", g, home); len(matches) != 1 || matches[0].Bookmark != "dom" {
		t.Fatalf("MatchByPrefix(dom-xyz) = %v, want only the readable dom vault (unreadable configs skipped)", matches)
	}
}

func TestMatchByPrefixEmptyAndNilConfig(t *testing.T) {
	home := t.TempDir()
	if matches := vault.MatchByPrefix("dom-xyz", vault.Global{}, home); len(matches) != 0 {
		t.Errorf("MatchByPrefix(empty global) = %v, want no match", matches)
	}
	g := vault.Global{Default: "dom"} // no bookmarks at all
	if matches := vault.MatchByPrefix("dom-xyz", g, home); len(matches) != 0 {
		t.Errorf("MatchByPrefix(nil bookmarks) = %v, want no match", matches)
	}
}

func TestMatchByPrefixAbsoluteAndTildePaths(t *testing.T) {
	home := t.TempDir()
	abs := writeVault(t, filepath.Join(home, "abs"), "abs")
	tilde := writeVault(t, filepath.Join(home, "work", "pkm"), "pkm")
	g := vault.Global{Bookmarks: map[string]string{"abs": abs, "pkm": "~/work/pkm"}}
	matched := vault.MatchByPrefix("abs-1", g, home)
	if len(matched) != 1 || matched[0].Path != abs {
		t.Errorf("MatchByPrefix(abs-1) = %v, want the absolute-path vault", matched)
	}
	matched = vault.MatchByPrefix("pkm-1", g, home)
	if len(matched) != 1 || matched[0].Path != tilde {
		t.Errorf("MatchByPrefix(pkm-1) = %v, want the tilde-expanded vault (got Path %q)", matched, matchPaths(matched))
	}
}

func TestPrefixOf(t *testing.T) {
	tests := []struct {
		key    string
		prefix string
		ok     bool
	}{
		{"dom-xyz", "dom", true},
		{"dom-x-y", "dom", true},
		{"dom", "", false},
		{"-xyz", "", false},
		{"-", "", false},
		{"", "", false},
		{"a-b", "a", true},
	}
	for _, tt := range tests {
		got, ok := vault.PrefixOf(tt.key)
		if got != tt.prefix || ok != tt.ok {
			t.Errorf("PrefixOf(%q) = (%q, %v), want (%q, %v)", tt.key, got, ok, tt.prefix, tt.ok)
		}
	}
}
