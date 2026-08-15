package vault

import (
	"fmt"
	"regexp"
	"sort"
)

// bookmarkNameRe matches a valid bookmark name: the part after '@' in a
// bookmark token. It shares the name grammar (letters, digits, '-' and
// '_') with bookmarkRe (resolve.go), so a name added here is always
// addressable as @name.
var bookmarkNameRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// IsValidBookmarkName reports whether name is usable as a bookmark name:
// non-empty, made of letters, digits, '-' and '_' — the exact grammar of
// the @name token minus the leading '@'.
func IsValidBookmarkName(name string) bool {
	return bookmarkNameRe.MatchString(name)
}

// cloneBookmarks returns a shallow copy of m (nil stays nil), so mutations
// of the copy never alias the receiver's map.
func cloneBookmarks(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// AddBookmark returns a copy of g with the bookmark name set to path,
// upserting an existing name. The receiver is not mutated. It fails when
// name is not a valid bookmark name.
func (g Global) AddBookmark(name, path string) (Global, error) {
	if !IsValidBookmarkName(name) {
		return Global{}, fmt.Errorf("invalid bookmark name %q: use letters, digits, '-' or '_'", name)
	}
	bookmarks := cloneBookmarks(g.Bookmarks)
	bookmarks[name] = path
	return Global{Default: g.Default, Bookmarks: bookmarks}, nil
}

// RemoveBookmark returns a copy of g without the bookmark name. When name
// is the default bookmark, the default is cleared too, so the config never
// points at a removed bookmark. The receiver is not mutated. It fails when
// the bookmark does not exist.
func (g Global) RemoveBookmark(name string) (Global, error) {
	if _, ok := g.Bookmarks[name]; !ok {
		return Global{}, fmt.Errorf("bookmark @%s not found", name)
	}
	bookmarks := cloneBookmarks(g.Bookmarks)
	delete(bookmarks, name)
	def := g.Default
	if def == name {
		def = ""
	}
	return Global{Default: def, Bookmarks: bookmarks}, nil
}

// Names returns the bookmark names in sorted order, for stable listing.
func (g Global) Names() []string {
	names := make([]string, 0, len(g.Bookmarks))
	for name := range g.Bookmarks {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
