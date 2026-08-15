package vault

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// bookmarkRe matches a vault bookmark token: @ followed by a short
// alphanumeric name. Anything else that merely starts with @ (a bare
// "@", a title with spaces) is left alone.
var bookmarkRe = regexp.MustCompile(`^@[A-Za-z0-9_-]+$`)

// BookmarkFromArgs extracts the optional @bookmark token from a command
// line: at most one argument of the form @name, anywhere among args.
// The remaining args keep their original order. A second @token is an
// error: the invocation is ambiguous.
func BookmarkFromArgs(args []string) (bookmark string, rest []string, err error) {
	for _, arg := range args {
		if !bookmarkRe.MatchString(arg) {
			rest = append(rest, arg)
			continue
		}
		if bookmark != "" {
			return "", nil, fmt.Errorf("multiple @bookmark arguments: %q and %q", bookmark, arg[1:])
		}
		bookmark = arg[1:]
	}
	return bookmark, rest, nil
}

// ExpandHome expands a leading ~ in p to home. Paths without a leading
// tilde come back unchanged.
func ExpandHome(p, home string) string {
	if p == "~" {
		return home
	}
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(home, p[2:])
	}
	return p
}

// Resolve determines the vault path for a command invocation.
//
// Precedence: explicit @bookmark > --vault flag > default bookmark.
// With none of them, the error carries the instructions the user needs.
// Bookmark paths and --vault paths get ~ expansion against home.
func Resolve(bookmark, vaultFlag string, g Global, home string) (string, error) {
	if bookmark != "" {
		path, ok := g.Bookmarks[bookmark]
		if !ok {
			return "", fmt.Errorf("bookmark @%s not found — add it with 'mt bookmark add %s <path>'", bookmark, bookmark)
		}
		return ExpandHome(path, home), nil
	}
	if vaultFlag != "" {
		return ExpandHome(vaultFlag, home), nil
	}
	if g.Default != "" {
		path, ok := g.Bookmarks[g.Default]
		if !ok {
			return "", fmt.Errorf("default bookmark %q is not defined in the global config", g.Default)
		}
		return ExpandHome(path, home), nil
	}
	return "", errors.New("no vault: use @bookmark, --vault <path>, or set a default bookmark in the global config")
}
