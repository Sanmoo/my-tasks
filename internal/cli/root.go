// Package cli wires the cobra command tree of mt. It owns process
// concerns — args, stdio, exit codes — while domain decisions live in
// other packages under internal/. Command behavior is covered by the
// e2e suite against the compiled binary, not by unit coverage.
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/vault"
)

// bookmark is the @bookmark extracted from the command line for this
// process, before cobra sees the args. Vault-requiring commands address
// the vault through it (see resolveVault). Run is single-shot per
// process, so a package var is safe.
var bookmark string

// Execute runs mt with the process's args and streams, returning the
// process exit code under the project convention.
func Execute() int {
	return Run(os.Args[1:], os.Stdout, os.Stderr)
}

// Run runs mt with the given args, writing to stdout and stderr, and
// returns the exit code under the project convention.
func Run(args []string, stdout, stderr io.Writer) int {
	// The @bookmark token may appear anywhere among the args; strip it
	// before cobra parses, so command argument validators never see it.
	var err error
	bookmark, args, err = vault.BookmarkFromArgs(args)
	if err != nil {
		return report(stderr, exitcode.Usage(err))
	}
	// A command line that is entirely a @bookmark (e.g. bare `mt @bjd`)
	// leaves args as a nil slice; cobra's SetArgs treats nil as "unspecified"
	// and falls back to os.Args[1:], re-introducing the @bookmark. Normalize
	// so cobra always runs exactly the stripped args.
	if args == nil {
		args = []string{}
	}

	cmd := NewRootCmd()
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	// Cobra returns plain errors for usage problems such as unknown
	// commands; classify them up front so the convention holds.
	if _, _, err := cmd.Find(args); err != nil {
		return report(stderr, exitcode.Usage(err))
	}
	// SetArgs makes Execute run the passed args instead of os.Args[1:],
	// so Run behaves identically for the real process and for
	// in-process callers.
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		return report(stderr, err)
	}
	return exitcode.Success
}

// report prints err to stderr and returns its exit code, with a hint
// for usage errors.
func report(stderr io.Writer, err error) int {
	code := exitcode.For(err)
	fmt.Fprintln(stderr, "Error:", err)
	if code == exitcode.UsageError {
		fmt.Fprintln(stderr, "Run 'mt --help' for usage.")
	}
	return code
}

// NewRootCmd builds the mt command tree.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mt",
		Short: "Personal issue tracker — one Markdown file per Issue",
		Long:  rootLong,
		// No positional args at all: an unknown word is an unknown
		// command, not an argument — a usage error.
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.NoArgs(cmd, args); err != nil {
				return exitcode.Usage(err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Bare `mt` (no command after extracting @bookmark) lists the
			// resolved vault's in_progress Issues — strictly `mt list
			// --status in_progress` (see rootLong for the full behavior).
			return runList(cmd, false, statusInProgress, nil)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return exitcode.Usage(err)
	})
	// Cobra's built-in help command shows the root help with exit 0 for
	// an unknown topic; the convention says an unknown command word is a
	// usage error, so help topics go through the same classification.
	cmd.SetHelpCommand(newHelpCmd())
	cmd.PersistentFlags().String("vault", "", "vault path (takes precedence over the default bookmark)")
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newQCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newEditCmd())
	cmd.AddCommand(newDoneCmd())
	cmd.AddCommand(newReopenCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newDeferCmd())
	cmd.AddCommand(newUndeferCmd())
	cmd.AddCommand(newDepCmd())
	cmd.AddCommand(newPickNextCmd())
	cmd.AddCommand(newPrioritizeCmd())
	cmd.AddCommand(newTopCmd())
	cmd.AddCommand(newBottomCmd())
	cmd.AddCommand(newRankCmd())
	cmd.AddCommand(newUnrankCmd())
	cmd.AddCommand(newCommentCmd())
	cmd.AddCommand(newBookmarkCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newCheckCmd())
	cmd.AddCommand(newReadyCmd())
	cmd.AddCommand(newOverdueCmd())
	return cmd
}

// newHelpCmd builds the help command with the project's error handling:
// `mt help` and `mt help <known command>` print help and exit 0; an
// unknown topic is a usage error (exit 2), matching `mt <unknown word>`.
func newHelpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Args:  cobra.ArbitraryArgs, // topics are validated in RunE
		RunE: func(cmd *cobra.Command, args []string) error {
			target, rest, err := cmd.Root().Find(args)
			if err != nil {
				return unknownHelpTopic(args)
			}
			// An unknown word resolves to the root command with the word
			// left over as a trailing argument — that is an unknown
			// topic, not a command path; so is a stray extra argument
			// after a known command (`mt help list extra`).
			if len(rest) > 0 {
				return unknownHelpTopic(rest)
			}
			return target.Help()
		},
	}
}

// unknownHelpTopic wraps an unknown help topic as a usage error (exit 2),
// matching the convention that an unknown command word is a usage error.
func unknownHelpTopic(words []string) error {
	return exitcode.Usage(fmt.Errorf("unknown help topic %q", strings.Join(words, " ")))
}

// resolveVault resolves the vault path for a vault-requiring command
// under the @bookmark > --vault > default precedence. The global config
// is auto-detected (XDG). It is the entry point of the commands that
// take no issue key; key commands resolve through resolveVaultForKey.
func resolveVault(cmd *cobra.Command) (string, error) {
	vaultFlag, err := cmd.Flags().GetString("vault")
	if err != nil {
		return "", fmt.Errorf("reading --vault flag: %w", err)
	}
	global, home, err := loadGlobalConfig()
	if err != nil {
		return "", err
	}
	resolved, err := vault.Resolve(bookmark, vaultFlag, global, home)
	if err != nil {
		return "", fmt.Errorf("resolving vault: %w", err)
	}
	return resolved, nil
}

// loadGlobalConfig loads the auto-detected global config (XDG) and
// returns it with the home directory its paths expand against.
func loadGlobalConfig() (vault.Global, string, error) {
	path, home, err := globalConfigPath()
	if err != nil {
		return vault.Global{}, "", fmt.Errorf("locating global config: %w", err)
	}
	global, err := vault.LoadGlobal(path)
	if err != nil {
		return vault.Global{}, "", fmt.Errorf("loading global config: %w", err)
	}
	return global, home, nil
}

// vaultTarget is the vault a key command runs against, plus the
// message context its not-found errors need: how the vault was picked
// decides whether an error names the vault, suggests another, or stays
// exactly as before.
type vaultTarget struct {
	// dir is the vault directory the command operates on.
	dir string
	// byPrefix is the bookmark whose vault the key's ID prefix picked, ""
	// when the selection was explicit, defaulted or absent.
	byPrefix string
	// explicit is the "@name" or --vault path the user picked
	// explicitly, "" when the vault came from the prefix or default.
	explicit string
	// hint is the "hint: ..." line for an explicit selection whose key
	// prefix matches a different bookmarked vault, "" when none.
	hint string
}

// resolveVaultForKey resolves the vault of a key command under the
// @bookmark > --vault > ID prefix > default precedence: an explicit
// selection behaves exactly like resolveVault — the key's prefix never
// overrides it — and the prefix (the part of the key before the first
// '-') only replaces the default step. A prefix matching several
// bookmarks is ambiguous and errors listing them; with no match the
// resolution falls back to the default bookmark, messages included.
func resolveVaultForKey(cmd *cobra.Command, key string) (vaultTarget, error) {
	vaultFlag, err := cmd.Flags().GetString("vault")
	if err != nil {
		return vaultTarget{}, fmt.Errorf("reading --vault flag: %w", err)
	}
	global, home, err := loadGlobalConfig()
	if err != nil {
		return vaultTarget{}, err
	}
	if bookmark != "" || vaultFlag != "" {
		dir, err := vault.Resolve(bookmark, vaultFlag, global, home)
		if err != nil {
			return vaultTarget{}, fmt.Errorf("resolving vault: %w", err)
		}
		t := vaultTarget{dir: dir}
		if bookmark != "" {
			t.explicit = "@" + bookmark
		} else {
			t.explicit = dir
		}
		t.hint = prefixHint(global, home, key, bookmark, dir)
		return t, nil
	}
	matches := vault.MatchByPrefix(key, global, home)
	switch len(matches) {
	case 1:
		return vaultTarget{dir: matches[0].Path, byPrefix: matches[0].Bookmark}, nil
	case 0:
		dir, err := vault.Resolve("", "", global, home)
		if err != nil {
			return vaultTarget{}, fmt.Errorf("resolving vault: %w", err)
		}
		return vaultTarget{dir: dir}, nil
	default:
		return vaultTarget{}, ambiguousPrefixError(cmd, key, matches)
	}
}

// prefixHint returns the "hint: the prefix … suggests vault @…" line
// for an explicit selection, when the key's prefix matches a bookmark
// the user did not already pick. An explicit selection is never
// second-guessed: a hint only appears when it points somewhere
// different. Empty when the key has no prefix or no such bookmark
// exists.
func prefixHint(global vault.Global, home, key, explicitBookmark, explicitDir string) string {
	for _, m := range vault.MatchByPrefix(key, global, home) {
		if m.Bookmark == explicitBookmark || filepath.Clean(m.Path) == filepath.Clean(explicitDir) {
			continue
		}
		prefix, _ := vault.PrefixOf(key)
		return fmt.Sprintf("hint: the prefix %q suggests vault @%s", prefix, m.Bookmark)
	}
	return ""
}

// ambiguousPrefixError builds the error for a key whose prefix matches
// several bookmarks: it lists them — never choosing silently — and
// shows the explicit invocation that disambiguates.
func ambiguousPrefixError(cmd *cobra.Command, key string, matches []vault.KeyVault) error {
	prefix, _ := vault.PrefixOf(key)
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		names = append(names, "@"+m.Bookmark)
	}
	suggestion := fmt.Sprintf("%s @%s %s", cmd.CommandPath(), matches[0].Bookmark, key)
	return fmt.Errorf("key %s is ambiguous: bookmarks %s both use prefix %q — pick one explicitly: %s",
		key, strings.Join(names, " and "), prefix, suggestion)
}

// notFoundError builds the "issue <id> not found" error of a key
// command: it names the vault the key's prefix selected — so the user
// knows resolution jumped there — or the explicit selection, appending
// the hint when one applies. The fallback case (no prefix match) keeps
// today's message exactly.
func (t vaultTarget) notFoundError(id string) error {
	switch {
	case t.byPrefix != "":
		return fmt.Errorf("issue %s not found in @%s", id, t.byPrefix)
	case t.hint != "":
		return fmt.Errorf("issue %s not found in %s\n%s", id, t.explicit, t.hint)
	default:
		return fmt.Errorf("issue %s not found", id)
	}
}

const rootLong = `mt is a personal, git-friendly issue tracker: one Markdown file per Issue,
one Vault per domain, everything versioned in Git.

Vaults are addressed by @bookmark, --vault <path>, the ID prefix of
an issue key (e.g. dom-xyz picks the vault whose prefix is dom), or
the default bookmark in the global config. An explicit @bookmark or
--vault always wins; with none of them, vault-requiring commands
fail with instructions.

Bare mt (no command) shows the resolved vault's in_progress Issues,
like 'mt list --status in_progress'. 'mt help' and 'mt --help' show
this help.

Exit codes:
  0  success
  1  user error — the command is well-formed but failed against the
     current state (vault undefined, issue not found, nothing available,
     duplicate rank, invalid edit)
  2  usage error — the invocation is malformed (unknown command or flag,
     wrong argument count, malformed argument)

Errors always go to stderr; command results go to stdout.`
