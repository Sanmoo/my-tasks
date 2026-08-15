// Package vault holds the pure logic of Vaults: the global config
// (bookmarks + default bookmark), the per-vault config (mt.yaml), and
// vault resolution (@bookmark > --vault > default). It is
// decision-dense, so it lives at Seam 2: black-box unit tested, with
// the coverage and mutation gates.
package vault

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// vaultConfigName is the vault config file at the vault root.
const vaultConfigName = "mt.yaml"

// Global is the user-level config: named bookmarks pointing at vault
// paths, plus the default bookmark used when a command names none.
type Global struct {
	// Default is the bookmark name used when no @bookmark or --vault
	// is given.
	Default string
	// Bookmarks maps bookmark names to vault paths.
	Bookmarks map[string]string
}

// globalFile is the on-disk shape of the global config:
//
//	default: bjd
//	bookmarks:
//	  bjd: ~/dev/github.com/Sanmoo/pkm/.vault
type globalFile struct {
	Default   string            `yaml:"default"`
	Bookmarks map[string]string `yaml:"bookmarks"`
}

// LoadGlobal reads the global config from path. A missing file yields
// an empty Global without error: no bookmarks, no default — resolution
// then fails with instructions.
func LoadGlobal(path string) (Global, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Global{}, nil
	}
	if err != nil {
		return Global{}, fmt.Errorf("reading global config %s: %w", path, err)
	}
	var f globalFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return Global{}, fmt.Errorf("parsing global config %s: %w", path, err)
	}
	return Global{Default: f.Default, Bookmarks: f.Bookmarks}, nil
}

// GlobalConfigPath returns the global config path: $XDG_CONFIG_HOME/mt/
// config.yaml when xdg is set, else $HOME/.config/mt/config.yaml.
func GlobalConfigPath(xdg, home string) string {
	if xdg != "" {
		return filepath.Join(xdg, "mt", "config.yaml")
	}
	return filepath.Join(home, ".config", "mt", "config.yaml")
}

// Vault is the per-vault config (mt.yaml): the ID prefix and the
// configured Status list.
type Vault struct {
	// Prefix is the ID prefix for issues of this vault (ex.: pkm).
	Prefix string
	// Status is the status list of the vault; empty means the defaults.
	Status []string
}

// vaultFile is the on-disk shape of the vault config:
//
//	prefix: pkm
//	status: [open, in_progress, done]
type vaultFile struct {
	Prefix string   `yaml:"prefix"`
	Status []string `yaml:"status,flow"`
}

// DefaultStatus are the statuses that apply when the vault config
// defines none.
var DefaultStatus = []string{"open", "in_progress", "done"}

// StatusList returns the configured statuses, or DefaultStatus when the
// config defines none.
func (v Vault) StatusList() []string {
	if len(v.Status) > 0 {
		return v.Status
	}
	return slices.Clone(DefaultStatus)
}

// LoadVault reads the vault config from dir/mt.yaml.
func LoadVault(dir string) (Vault, error) {
	path := filepath.Join(dir, vaultConfigName)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Vault{}, fmt.Errorf("not a vault: %s missing in %s — run 'mt init'", vaultConfigName, dir)
	}
	if err != nil {
		return Vault{}, fmt.Errorf("reading vault config %s: %w", path, err)
	}
	var f vaultFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return Vault{}, fmt.Errorf("parsing vault config %s: %w", path, err)
	}
	return Vault{Prefix: f.Prefix, Status: f.Status}, nil
}

// Save creates a usable Vault at dir: the issues/ directory plus
// dir/mt.yaml with this config. It refuses to overwrite an existing
// vault config.
func (v Vault) Save(dir string) error {
	configPath := filepath.Join(dir, vaultConfigName)
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("a vault already exists at %s (%s present)", dir, vaultConfigName)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking vault config %s: %w", configPath, err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "issues"), 0o755); err != nil {
		return fmt.Errorf("creating issues directory: %w", err)
	}
	data, err := yaml.Marshal(vaultFile{Prefix: v.Prefix, Status: v.StatusList()})
	if err != nil {
		return fmt.Errorf("encoding vault config: %w", err)
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("writing vault config %s: %w", configPath, err)
	}
	return nil
}

// maxPrefixLen caps derived ID prefixes so issue IDs stay short.
const maxPrefixLen = 8

// PrefixFor derives an ID prefix from a directory name: lowercased,
// stripped of non-alphanumeric characters, capped at maxPrefixLen.
// Empty when nothing usable remains. Ex.: ".vault" → "vault".
func PrefixFor(dir string) string {
	base := filepath.Base(filepath.Clean(dir))
	var b strings.Builder
	for _, r := range strings.ToLower(base) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			if b.Len() == maxPrefixLen {
				break
			}
		}
	}
	return b.String()
}
