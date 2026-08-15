// Package vault_test holds the black-box unit tests of the vault and
// global configs (Seam 2): load/save round-trips, defaults, and the
// derived ID prefix.
package vault_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/vault"
)

func TestLoadGlobalMissingFileIsEmptyConfig(t *testing.T) {
	g, err := vault.LoadGlobal(filepath.Join(t.TempDir(), "nope", "config.yaml"))
	if err != nil {
		t.Fatalf("LoadGlobal(missing) = %v, want no error", err)
	}
	if g.Default != "" || len(g.Bookmarks) != 0 {
		t.Errorf("LoadGlobal(missing) = %+v, want empty config", g)
	}
}

func TestLoadGlobalParsesBookmarksAndDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := "default: bjd\nbookmarks:\n  bjd: ~/dev/bjd/.vault\n  dom: /v/dom\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	g, err := vault.LoadGlobal(path)
	if err != nil {
		t.Fatal(err)
	}
	if g.Default != "bjd" {
		t.Errorf("Default = %q, want %q", g.Default, "bjd")
	}
	if got := g.Bookmarks["bjd"]; got != "~/dev/bjd/.vault" {
		t.Errorf("Bookmarks[bjd] = %q, want %q", got, "~/dev/bjd/.vault")
	}
	if got := g.Bookmarks["dom"]; got != "/v/dom" {
		t.Errorf("Bookmarks[dom] = %q, want %q", got, "/v/dom")
	}
}

func TestLoadGlobalMalformedYAMLFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("bookmarks: [unclosed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := vault.LoadGlobal(path); err == nil {
		t.Fatal("LoadGlobal(malformed) = nil error, want failure")
	}
}

func TestLoadGlobalUnreadablePathFails(t *testing.T) {
	// A directory is readable by os.ReadFile? No — reading a directory
	// fails with a non-NotExist error, exercising the wrap path.
	if _, err := vault.LoadGlobal(t.TempDir()); err == nil {
		t.Fatal("LoadGlobal(directory) = nil error, want failure")
	}
}

func TestGlobalConfigPath(t *testing.T) {
	if got, want := vault.GlobalConfigPath("/xdg", "/home/u"), "/xdg/mt/config.yaml"; got != want {
		t.Errorf("GlobalConfigPath(xdg) = %q, want %q", got, want)
	}
	if got, want := vault.GlobalConfigPath("", "/home/u"), "/home/u/.config/mt/config.yaml"; got != want {
		t.Errorf("GlobalConfigPath(no xdg) = %q, want %q", got, want)
	}
}

func TestSaveCreatesUsableVault(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "pkm")
	v := vault.Vault{Prefix: "pkm", Status: []string{"open", "in_progress", "done", "blocked"}}
	if err := v.Save(dir); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(dir, "issues"))
	if err != nil {
		t.Fatalf("issues/ not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("issues/ is not a directory")
	}
	got, err := vault.LoadVault(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Prefix != "pkm" {
		t.Errorf("round-trip Prefix = %q, want %q", got.Prefix, "pkm")
	}
	if len(got.Status) != 4 || got.Status[3] != "blocked" {
		t.Errorf("round-trip Status = %v, want [open in_progress done blocked]", got.Status)
	}
}

func TestSaveWritesSpecShape(t *testing.T) {
	dir := t.TempDir()
	v := vault.Vault{Prefix: "PKM", Status: []string{"open", "in_progress", "done"}}
	if err := v.Save(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "mt.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"prefix: PKM", "status: [open, in_progress, done]"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("mt.yaml does not contain %q:\n%s", want, data)
		}
	}
}

func TestSaveFillsDefaultStatusWhenNoneConfigured(t *testing.T) {
	dir := t.TempDir()
	v := vault.Vault{Prefix: "pkm"}
	if err := v.Save(dir); err != nil {
		t.Fatal(err)
	}
	got, err := vault.LoadVault(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Prefix != "pkm" {
		t.Errorf("Prefix = %q, want %q", got.Prefix, "pkm")
	}
	if len(got.Status) != 3 || got.Status[0] != "open" || got.Status[2] != "done" {
		t.Errorf("Status = %v, want the defaults [open in_progress done]", got.Status)
	}
}

func TestSaveRefusesExistingVault(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "pkm")
	v := vault.Vault{Prefix: "pkm"}
	if err := v.Save(dir); err != nil {
		t.Fatal(err)
	}
	err := v.Save(dir)
	if err == nil {
		t.Fatal("second Save = nil error, want refusal")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("refusal error %q does not say 'already exists'", err)
	}
}

func TestSaveRefusesPreExistingConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mt.yaml"), []byte("prefix: other\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := (vault.Vault{Prefix: "pkm"}).Save(dir); err == nil {
		t.Fatal("Save over an existing mt.yaml = nil error, want refusal")
	}
}

func TestSaveFailsWhenTargetIsAFile(t *testing.T) {
	base := t.TempDir()
	blocked := filepath.Join(base, "blocked")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := (vault.Vault{Prefix: "pkm"}).Save(blocked); err == nil {
		t.Fatal("Save onto a file path = nil error, want failure")
	}
}

func TestSaveFailsWhenConfigPathIsUnreachable(t *testing.T) {
	// The mt.yaml path sits under a regular file, so the existence
	// check itself fails with a non-NotExist error.
	base := t.TempDir()
	blocked := filepath.Join(base, "blocked")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := (vault.Vault{Prefix: "pkm"}).Save(filepath.Join(blocked, "sub")); err == nil {
		t.Fatal("Save with unreachable config path = nil error, want failure")
	}
}

func TestSaveFailsWhenConfigCannotBeWritten(t *testing.T) {
	// A dangling symlink at the mt.yaml path: the existence check
	// reports not-exists (the stat follows the link), but the write
	// cannot create the target.
	base := t.TempDir()
	dir := filepath.Join(base, "pkm")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(base, "nowhere", "mt.yaml"), filepath.Join(dir, "mt.yaml")); err != nil {
		t.Fatal(err)
	}
	if err := (vault.Vault{Prefix: "pkm"}).Save(dir); err == nil {
		t.Fatal("Save with unwritable mt.yaml = nil error, want failure")
	}
}

func TestLoadVaultMissingConfigFails(t *testing.T) {
	_, err := vault.LoadVault(t.TempDir())
	if err == nil {
		t.Fatal("LoadVault(no mt.yaml) = nil error, want failure")
	}
	for _, want := range []string{"not a vault", "mt init"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

func TestLoadVaultMalformedYAMLFails(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mt.yaml"), []byte("prefix: [unclosed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := vault.LoadVault(dir); err == nil {
		t.Fatal("LoadVault(malformed) = nil error, want failure")
	}
}

func TestLoadVaultUnreadablePathFails(t *testing.T) {
	// mt.yaml under a path where the parent is a regular file: the
	// read fails with a non-NotExist error, exercising the wrap path.
	base := t.TempDir()
	blocked := filepath.Join(base, "blocked")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := vault.LoadVault(filepath.Join(blocked, "sub")); err == nil {
		t.Fatal("LoadVault(unreadable) = nil error, want failure")
	}
}

func TestStatusList(t *testing.T) {
	if got := (vault.Vault{}).StatusList(); len(got) != 3 || got[0] != "open" || got[2] != "done" {
		t.Errorf("StatusList() empty = %v, want the defaults", got)
	}
	if got := (vault.Vault{}).StatusList(); len(got) != len(vault.DefaultStatus) {
		t.Errorf("StatusList() empty = %v, want len %d", got, len(vault.DefaultStatus))
	}
	custom := []string{"todo", "doing", "done"}
	if got := (vault.Vault{Status: custom}).StatusList(); len(got) != 3 || got[0] != "todo" {
		t.Errorf("StatusList() custom = %v, want the configured list", got)
	}
}

func TestStatusListDoesNotLeakMutations(t *testing.T) {
	// The empty case must return a fresh slice: mutating it must not
	// corrupt the package default for later calls.
	got := (vault.Vault{}).StatusList()
	got[0] = "mutated"
	if again := (vault.Vault{}).StatusList(); again[0] != "open" {
		t.Errorf("StatusList() after mutation = %v, want [open in_progress done]", again)
	}
}

func TestPrefixFor(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		want string
	}{
		{"simple name", "pkm", "pkm"},
		{"dot directory", ".vault", "vault"},
		{"dash stripped", "my-tasks2", "mytasks2"},
		{"underscore stripped", "a_b", "ab"},
		{"dot stripped", "a.b", "ab"},
		{"uppercase lowered", "PKM", "pkm"},
		{"digits kept", "dom2", "dom2"},
		{"long capped", "toolongname", "toolongn"},
		{"unicode stripped", "café", "caf"},
		{"spaces stripped", "x y", "xy"},
		{"full path uses base", "/tmp/pkm", "pkm"},
		{"trailing slash", "pkm/", "pkm"},
		{"root has no prefix", "/", ""},
		{"dot has no prefix", ".", ""},
		{"empty has no prefix", "", ""},
		// Boundary characters: the character-range conditions must keep
		// 'a', 'z', '0' and '9' and drop everything just outside.
		{"lower boundary a", "a", "a"},
		{"upper boundary z", "z", "z"},
		{"digit boundary zero", "0", "0"},
		{"digit boundary nine", "9", "9"},
		{"symbols only", "@!", ""},
		{"uppercase boundaries", "AZ09", "az09"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := vault.PrefixFor(tt.dir); got != tt.want {
				t.Errorf("PrefixFor(%q) = %q, want %q", tt.dir, got, tt.want)
			}
		})
	}
}
