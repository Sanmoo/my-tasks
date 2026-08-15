// Package vault_test holds the black-box unit tests of bookmark
// management (Seam 2): name validation, add/remove mutations, sorted
// listing, and the global-config write/read round-trip that makes a
// fresh bookmark resolvable by the next command.
package vault_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/vault"
)

func TestIsValidBookmarkName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"plain name", "pkm", true},
		{"dash and underscore", "a-b_9", true},
		{"uppercase", "A", true},
		{"digits", "0", true},
		{"leading dash", "-a", true},
		{"empty", "", false},
		{"space", "has space", false},
		{"leading at", "@pkm", false},
		{"slash", "a/b", false},
		{"unicode", "café", false},
		{"dot", "a.b", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := vault.IsValidBookmarkName(tt.in); got != tt.want {
				t.Errorf("IsValidBookmarkName(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// TestBookmarkNameGrammarMatchesTokenMatcher pins the shared grammar: a
// name accepted by IsValidBookmarkName must be addressable as @name, and
// a name the @token matcher rejects must also fail validation. This guards
// the two regexes (bookmarkNameRe, bookmarkRe) against drifting apart.
func TestBookmarkNameGrammarMatchesTokenMatcher(t *testing.T) {
	for _, name := range []string{"pkm", "a-b_9", "A", "0"} {
		if !vault.IsValidBookmarkName(name) {
			t.Errorf("IsValidBookmarkName(%q) = false, want true", name)
		}
		bookmark, rest, err := vault.BookmarkFromArgs([]string{"@" + name})
		if err != nil {
			t.Errorf("BookmarkFromArgs(@%s): %v", name, err)
		}
		if bookmark != name || len(rest) != 0 {
			t.Errorf("BookmarkFromArgs(@%s) = (%q, %v), want (%q, nil)", name, bookmark, rest, name)
		}
	}
	for _, bad := range []string{"", "has space", "a/b", "café", "a.b"} {
		if vault.IsValidBookmarkName(bad) {
			t.Errorf("IsValidBookmarkName(%q) = true, want false", bad)
		}
	}
}

func TestAddBookmarkCreatesNewEntry(t *testing.T) {
	g, err := (vault.Global{}).AddBookmark("pkm", "~/dev/pkm")
	if err != nil {
		t.Fatal(err)
	}
	if g.Default != "" {
		t.Errorf("Default = %q, want empty", g.Default)
	}
	if got := g.Bookmarks["pkm"]; got != "~/dev/pkm" {
		t.Errorf("Bookmarks[pkm] = %q, want %q", got, "~/dev/pkm")
	}
}

func TestAddBookmarkUpsertsExistingName(t *testing.T) {
	g := vault.Global{Bookmarks: map[string]string{"pkm": "~/old"}}
	got, err := g.AddBookmark("pkm", "~/new")
	if err != nil {
		t.Fatal(err)
	}
	if got.Bookmarks["pkm"] != "~/new" {
		t.Errorf("Bookmarks[pkm] = %q, want the upserted path %q", got.Bookmarks["pkm"], "~/new")
	}
	if len(got.Bookmarks) != 1 {
		t.Errorf("len(Bookmarks) = %d, want 1 after upsert", len(got.Bookmarks))
	}
}

func TestAddBookmarkPreservesDefaultAndOthers(t *testing.T) {
	g := vault.Global{
		Default:   "bjd",
		Bookmarks: map[string]string{"bjd": "~/dev/bjd"},
	}
	got, err := g.AddBookmark("dom", "/v/dom")
	if err != nil {
		t.Fatal(err)
	}
	if got.Default != "bjd" {
		t.Errorf("Default = %q, want %q", got.Default, "bjd")
	}
	if got.Bookmarks["bjd"] != "~/dev/bjd" {
		t.Errorf("Bookmarks[bjd] = %q, want the original value", got.Bookmarks["bjd"])
	}
	if got.Bookmarks["dom"] != "/v/dom" {
		t.Errorf("Bookmarks[dom] = %q, want %q", got.Bookmarks["dom"], "/v/dom")
	}
}

func TestAddBookmarkDoesNotMutateReceiver(t *testing.T) {
	g := vault.Global{Default: "bjd", Bookmarks: map[string]string{"bjd": "~/dev/bjd"}}
	if _, err := g.AddBookmark("pkm", "~/dev/pkm"); err != nil {
		t.Fatal(err)
	}
	if _, ok := g.Bookmarks["pkm"]; ok {
		t.Error("AddBookmark mutated the receiver's Bookmarks map")
	}
	if len(g.Bookmarks) != 1 {
		t.Errorf("receiver Bookmarks = %v, want the original single entry", g.Bookmarks)
	}
}

func TestAddBookmarkInvalidNameFails(t *testing.T) {
	for _, name := range []string{"", "@pkm", "has space", "a/b", "café"} {
		if _, err := (vault.Global{}).AddBookmark(name, "/v"); err == nil {
			t.Errorf("AddBookmark(%q) = nil error, want failure", name)
		}
	}
}

func TestRemoveBookmarkRemovesAndPreservesOthers(t *testing.T) {
	g := vault.Global{
		Default:   "bjd",
		Bookmarks: map[string]string{"bjd": "~/dev/bjd", "dom": "~/dev/dom"},
	}
	got, err := g.RemoveBookmark("dom")
	if err != nil {
		t.Fatal(err)
	}
	if got.Default != "bjd" {
		t.Errorf("Default = %q, want %q", got.Default, "bjd")
	}
	if _, ok := got.Bookmarks["dom"]; ok {
		t.Error("Bookmarks still contains the removed name dom")
	}
	if got.Bookmarks["bjd"] != "~/dev/bjd" {
		t.Errorf("Bookmarks[bjd] = %q, want the preserved value", got.Bookmarks["bjd"])
	}
}

func TestRemoveBookmarkClearsDefaultWhenRemoved(t *testing.T) {
	g := vault.Global{
		Default:   "bjd",
		Bookmarks: map[string]string{"bjd": "~/dev/bjd", "dom": "~/dev/dom"},
	}
	got, err := g.RemoveBookmark("bjd")
	if err != nil {
		t.Fatal(err)
	}
	if got.Default != "" {
		t.Errorf("Default = %q, want cleared when the default bookmark is removed", got.Default)
	}
	if _, ok := got.Bookmarks["dom"]; !ok {
		t.Error("the non-default bookmark dom was lost")
	}
}

func TestRemoveBookmarkDoesNotMutateReceiver(t *testing.T) {
	g := vault.Global{
		Default:   "bjd",
		Bookmarks: map[string]string{"bjd": "~/dev/bjd", "dom": "~/dev/dom"},
	}
	if _, err := g.RemoveBookmark("dom"); err != nil {
		t.Fatal(err)
	}
	if _, ok := g.Bookmarks["dom"]; !ok {
		t.Error("RemoveBookmark mutated the receiver's Bookmarks map")
	}
	if g.Default != "bjd" {
		t.Errorf("receiver Default = %q, want unchanged %q", g.Default, "bjd")
	}
}

func TestRemoveBookmarkMissingFails(t *testing.T) {
	g := vault.Global{Bookmarks: map[string]string{"bjd": "/v/bjd"}}
	_, err := g.RemoveBookmark("nope")
	if err == nil {
		t.Fatal("RemoveBookmark(missing) = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "@nope") {
		t.Errorf("error %q does not name the missing bookmark", err)
	}
}

func TestRemoveBookmarkFromEmptyConfigFails(t *testing.T) {
	if _, err := (vault.Global{}).RemoveBookmark("nope"); err == nil {
		t.Fatal("RemoveBookmark from empty config = nil error, want failure")
	}
}

func TestNamesSorted(t *testing.T) {
	g := vault.Global{Bookmarks: map[string]string{"dom": "/d", "bjd": "/b", "app": "/a"}}
	got := g.Names()
	want := []string{"app", "bjd", "dom"}
	if len(got) != len(want) {
		t.Fatalf("Names() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Names() = %v, want %v", got, want)
			break
		}
	}
}

func TestNamesEmpty(t *testing.T) {
	if got := (vault.Global{}).Names(); len(got) != 0 {
		t.Errorf("Names() = %v, want empty", got)
	}
}

func TestSaveGlobalWritesSpecShape(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mt", "config.yaml")
	g := vault.Global{
		Default:   "bjd",
		Bookmarks: map[string]string{"bjd": "~/dev/bjd", "dom": "/v/dom"},
	}
	if err := g.SaveGlobal(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"default: bjd", "bjd: ~/dev/bjd", "dom: /v/dom"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("config does not contain %q:\n%s", want, data)
		}
	}
}

func TestSaveGlobalOmitsEmptyDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	g := vault.Global{Bookmarks: map[string]string{"pkm": "~/dev/pkm"}}
	if err := g.SaveGlobal(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "default:") {
		t.Errorf("config should omit the default key when unset:\n%s", data)
	}
}

func TestSaveGlobalCreatesParentDirectories(t *testing.T) {
	path := filepath.Join(t.TempDir(), "deep", "nested", "config.yaml")
	if err := (vault.Global{Bookmarks: map[string]string{"pkm": "/v"}}).SaveGlobal(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config not written: %v", err)
	}
}

func TestSaveGlobalRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	orig := vault.Global{
		Default:   "bjd",
		Bookmarks: map[string]string{"bjd": "~/dev/bjd", "dom": "/v/dom"},
	}
	if err := orig.SaveGlobal(path); err != nil {
		t.Fatal(err)
	}
	got, err := vault.LoadGlobal(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Default != orig.Default {
		t.Errorf("round-trip Default = %q, want %q", got.Default, orig.Default)
	}
	for name, want := range orig.Bookmarks {
		if got.Bookmarks[name] != want {
			t.Errorf("round-trip Bookmarks[%s] = %q, want %q", name, got.Bookmarks[name], want)
		}
	}
}

// TestAddThenResolveBookmark encodes the ticket's "a freshly added
// bookmark resolves via @nome on the next command": add → save → load →
// resolve, with ~ expansion applied at resolution time.
func TestAddThenResolveBookmark(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	g, err := (vault.Global{}).AddBookmark("pkm", "~/dev/pkm")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.SaveGlobal(path); err != nil {
		t.Fatal(err)
	}
	loaded, err := vault.LoadGlobal(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := vault.Resolve("pkm", "", loaded, "/home/u")
	if err != nil {
		t.Fatal(err)
	}
	if want := "/home/u/dev/pkm"; got != want {
		t.Errorf("Resolve(@pkm) after add = %q, want %q", got, want)
	}
}

func TestSaveGlobalFailsWhenConfigPathIsADirectory(t *testing.T) {
	// The config path itself is a directory, so the write fails with a
	// non-NotExist error, exercising the write-error branch.
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := (vault.Global{Bookmarks: map[string]string{"pkm": "/v"}}).SaveGlobal(path); err == nil {
		t.Fatal("SaveGlobal onto a directory = nil error, want failure")
	}
}

func TestSaveGlobalFailsWhenParentIsAFile(t *testing.T) {
	// The parent of the config path is a regular file, so MkdirAll fails.
	base := t.TempDir()
	blocked := filepath.Join(base, "blocked")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := (vault.Global{}).SaveGlobal(filepath.Join(blocked, "config.yaml")); err == nil {
		t.Fatal("SaveGlobal under a file parent = nil error, want failure")
	}
}
