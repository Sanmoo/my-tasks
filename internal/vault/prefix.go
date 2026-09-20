package vault

import (
	"path/filepath"
	"strings"
)

// KeyVault is a bookmarked vault that a key's ID prefix matched: the
// bookmark name (without the @) and the expanded vault directory.
type KeyVault struct {
	// Bookmark is the bookmark name the vault is addressed by.
	Bookmark string
	// Path is the expanded vault directory.
	Path string
}

// PrefixOf returns the ID prefix of a key: the part before the first
// '-'. A key without a '-' has no prefix (false).
func PrefixOf(key string) (string, bool) {
	prefix, _, hasDash := strings.Cut(key, "-")
	if !hasDash || prefix == "" {
		return "", false
	}
	return prefix, true
}

// MatchByPrefix returns the bookmarked vaults whose configured prefix
// equals the ID prefix of key — the part before the first '-',
// case-sensitive against each vault's mt.yaml `prefix:`.
//
// Zero candidates when the key has no '-' or no bookmarked vault
// matches. More than one when multiple bookmarks share the prefix —
// the caller treats that as ambiguity. Bookmarks pointing at the same
// directory are deduplicated (aliases of one vault do not count as
// ambiguity), and a vault whose config cannot be read (no mt.yaml,
// unreadable or malformed) is skipped: it is unusable for the command
// anyway. Everything that can fail is skipped, so the function never
// returns an error. Candidates come in bookmark-name order (sorted).
func MatchByPrefix(key string, g Global, home string) []KeyVault {
	prefix, ok := PrefixOf(key)
	if !ok {
		return nil
	}
	var matches []KeyVault
	seen := make(map[string]bool)
	for _, name := range g.Names() {
		dir := ExpandHome(g.Bookmarks[name], home)
		cfg, err := LoadVault(dir)
		if err != nil || cfg.Prefix != prefix {
			continue
		}
		// Two bookmarks for the same vault count once: the first (in
		// sorted bookmark order) represents the directory.
		key := filepath.Clean(dir)
		if seen[key] {
			continue
		}
		seen[key] = true
		matches = append(matches, KeyVault{Bookmark: name, Path: dir})
	}
	return matches
}
