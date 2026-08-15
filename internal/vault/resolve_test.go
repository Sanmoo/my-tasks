// Package vault_test holds the black-box unit tests of vault resolution
// (Seam 2: exported API of a pure-logic package). The precedence
// @bookmark > --vault > default is the decision-dense core of the CLI's
// vault addressing, so it gets the full matrix here.
package vault_test

import (
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/vault"
)

func TestResolveBookmarkWinsOverFlagAndDefault(t *testing.T) {
	g := vault.Global{
		Default: "bjd",
		Bookmarks: map[string]string{
			"bjd": "~/dev/bjd",
			"pkm": "~/dev/pkm",
		},
	}
	got, err := vault.Resolve("pkm", "~/flag-vault", g, "/home/u")
	if err != nil {
		t.Fatal(err)
	}
	if want := "/home/u/dev/pkm"; got != want {
		t.Errorf("Resolve(@pkm, --vault, default) = %q, want %q", got, want)
	}
}

func TestResolveBookmarkMissingFailsWithInstructions(t *testing.T) {
	g := vault.Global{Bookmarks: map[string]string{"bjd": "/v/bjd"}}
	_, err := vault.Resolve("nope", "", g, "/home/u")
	if err == nil {
		t.Fatal("Resolve(missing bookmark) = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "@nope") {
		t.Errorf("error %q does not name the bookmark", err)
	}
	if !strings.Contains(err.Error(), "add") {
		t.Errorf("error %q does not instruct how to add the bookmark", err)
	}
}

func TestResolveVaultFlagWinsOverDefault(t *testing.T) {
	g := vault.Global{
		Default:   "bjd",
		Bookmarks: map[string]string{"bjd": "/v/bjd"},
	}
	got, err := vault.Resolve("", "/v/flag", g, "/home/u")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/v/flag" {
		t.Errorf("Resolve(--vault /v/flag, default bjd) = %q, want %q", got, "/v/flag")
	}
}

func TestResolveVaultFlagExpandsTilde(t *testing.T) {
	got, err := vault.Resolve("", "~/vaults/pkm", vault.Global{}, "/home/u")
	if err != nil {
		t.Fatal(err)
	}
	if want := "/home/u/vaults/pkm"; got != want {
		t.Errorf("Resolve(--vault ~/vaults/pkm) = %q, want %q", got, want)
	}
}

func TestResolveVaultFlagAbsolutePathIsKept(t *testing.T) {
	got, err := vault.Resolve("", "/abs/path", vault.Global{}, "/home/u")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/abs/path" {
		t.Errorf("Resolve(--vault /abs/path) = %q, want unchanged path", got)
	}
}

func TestResolveVaultFlagRelativePathIsKept(t *testing.T) {
	got, err := vault.Resolve("", "rel/path", vault.Global{}, "/home/u")
	if err != nil {
		t.Fatal(err)
	}
	if got != "rel/path" {
		t.Errorf("Resolve(--vault rel/path) = %q, want unchanged path", got)
	}
}

func TestResolveUsesDefaultWhenNothingElseGiven(t *testing.T) {
	g := vault.Global{
		Default: "bjd",
		Bookmarks: map[string]string{
			"bjd": "~/dev/bjd",
			"dom": "/v/dom",
		},
	}
	got, err := vault.Resolve("", "", g, "/home/u")
	if err != nil {
		t.Fatal(err)
	}
	if want := "/home/u/dev/bjd"; got != want {
		t.Errorf("Resolve(default bjd) = %q, want %q", got, want)
	}
}

func TestResolveDefaultMissingFromBookmarksFails(t *testing.T) {
	g := vault.Global{Default: "ghost"}
	_, err := vault.Resolve("", "", g, "/home/u")
	if err == nil {
		t.Fatal("Resolve(default missing from bookmarks) = nil error, want failure")
	}
	if !strings.Contains(err.Error(), "ghost") {
		t.Errorf("error %q does not name the default bookmark", err)
	}
}

func TestResolveNothingDefinedFailsWithInstructions(t *testing.T) {
	_, err := vault.Resolve("", "", vault.Global{}, "/home/u")
	if err == nil {
		t.Fatal("Resolve(no vault defined) = nil error, want failure")
	}
	for _, want := range []string{"@bookmark", "--vault", "default"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

func TestResolveDefaultWithNilBookmarksFails(t *testing.T) {
	// A global config with a default but no bookmarks section parses to
	// a nil map; lookup must fail cleanly, not panic.
	g := vault.Global{Default: "bjd"}
	_, err := vault.Resolve("", "", g, "/home/u")
	if err == nil {
		t.Fatal("Resolve(default, nil bookmarks) = nil error, want failure")
	}
}

func TestResolveBookmarkWithNilBookmarksFails(t *testing.T) {
	_, err := vault.Resolve("bjd", "", vault.Global{}, "/home/u")
	if err == nil {
		t.Fatal("Resolve(@bjd, nil bookmarks) = nil error, want failure")
	}
}

func TestBookmarkFromArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		bookmark string
		rest     []string
		wantErr  bool
	}{
		{"no args", nil, "", nil, false},
		{"no bookmark", []string{"list"}, "", []string{"list"}, false},
		{"bookmark last", []string{"list", "@pkm"}, "pkm", []string{"list"}, false},
		{"bookmark first", []string{"@pkm", "list"}, "pkm", []string{"list"}, false},
		{"bookmark middle", []string{"create", "@pkm", "titulo"}, "pkm", []string{"create", "titulo"}, false},
		{"rest keeps order", []string{"status", "pkm-01", "@dom", "done"}, "dom", []string{"status", "pkm-01", "done"}, false},
		{"dash and digit in name", []string{"@a-b_9"}, "a-b_9", nil, false},
		{"uppercase name", []string{"@A"}, "A", nil, false},
		{"multiple bookmarks", []string{"@a", "list", "@b"}, "", nil, true},
		{"bare at is not a bookmark", []string{"list", "@"}, "", []string{"list", "@"}, false},
		{"flag-like arg kept", []string{"-x", "@a"}, "a", []string{"-x"}, false},
		{"empty args", []string{}, "", []string{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmark, rest, err := vault.BookmarkFromArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("BookmarkFromArgs = nil error, want failure")
				}
				return
			}
			if err != nil {
				t.Fatalf("BookmarkFromArgs: %v", err)
			}
			if bookmark != tt.bookmark {
				t.Errorf("bookmark = %q, want %q", bookmark, tt.bookmark)
			}
			if len(rest) != len(tt.rest) {
				t.Fatalf("rest = %v, want %v", rest, tt.rest)
			}
			for i := range rest {
				if rest[i] != tt.rest[i] {
					t.Errorf("rest = %v, want %v", rest, tt.rest)
					break
				}
			}
		})
	}
}

func TestExpandHome(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"bare tilde", "~", "/home/u"},
		{"tilde path", "~/dev/pkm", "/home/u/dev/pkm"},
		{"tilde path nested", "~/a/b/c", "/home/u/a/b/c"},
		{"absolute unchanged", "/abs/path", "/abs/path"},
		{"relative unchanged", "rel/path", "rel/path"},
		{"empty unchanged", "", ""},
		{"tilde not at start", "x~/y", "x~/y"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := vault.ExpandHome(tt.in, "/home/u"); got != tt.want {
				t.Errorf("ExpandHome(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
