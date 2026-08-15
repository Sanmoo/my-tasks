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
)

// Execute runs mt with the process's args and streams, returning the
// process exit code under the project convention.
func Execute() int {
	return Run(os.Args[1:], os.Stdout, os.Stderr)
}

// Run runs mt with the given args, writing to stdout and stderr, and
// returns the exit code under the project convention.
func Run(args []string, stdout, stderr io.Writer) int {
	cmd := NewRootCmd()
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	// Cobra returns plain errors for usage problems such as unknown
	// commands; classify them up front so the convention holds.
	if _, _, err := cmd.Find(args); err != nil {
		return report(stderr, exitcode.Usage(err))
	}
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
	return cmd
}

const rootLong = `mt is a personal, git-friendly issue tracker: one Markdown file per Issue,
one Vault per domain, everything versioned in Git.

Exit codes:
  0  success
  1  user error (vault undefined, nothing available, duplicate rank, invalid edit)
  2  usage error (unknown command, unknown flag, wrong arguments)`
