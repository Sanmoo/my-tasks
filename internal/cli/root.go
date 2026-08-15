// Package cli wires the cobra command tree of mt. It owns process
// concerns — args, stdio, exit codes — while domain decisions live in
// other packages under internal/. Command behavior is covered by the
// e2e suite against the compiled binary, not by unit coverage.
package cli

import (
	"fmt"
	"io"
	"os"

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
			// Bare `mt` shows help, like `mt --help`.
			return cmd.Help()
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return exitcode.Usage(err)
	})
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
	cmd.AddCommand(newPrioritizeCmd())
	cmd.AddCommand(newCommentCmd())
	cmd.AddCommand(newBookmarkCmd())
	cmd.AddCommand(newListCmd())
	return cmd
}

// resolveVault resolves the vault path for a vault-requiring command
// under the @bookmark > --vault > default precedence. The global config
// is auto-detected (XDG). Vault-requiring commands land in later
// tickets (create/show/edit, list, ...); this helper is their shared
// entry point.
func resolveVault(cmd *cobra.Command) (string, error) {
	vaultFlag, err := cmd.Flags().GetString("vault")
	if err != nil {
		return "", err
	}
	path, home, err := globalConfigPath()
	if err != nil {
		return "", err
	}
	global, err := vault.LoadGlobal(path)
	if err != nil {
		return "", err
	}
	return vault.Resolve(bookmark, vaultFlag, global, home)
}

const rootLong = `mt is a personal, git-friendly issue tracker: one Markdown file per Issue,
one Vault per domain, everything versioned in Git.

Vaults are addressed by @bookmark, --vault <path>, or the default
bookmark in the global config. With none of them, vault-requiring
commands fail with instructions.

Exit codes:
  0  success
  1  user error (vault undefined, nothing available, duplicate rank, invalid edit)
  2  usage error (unknown command, unknown flag, wrong arguments)`
