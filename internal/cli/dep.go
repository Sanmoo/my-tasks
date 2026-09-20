// Package cli — the mt dep commands. They own the process concerns of
// editing blocked_by (resolving the vault, reading/writing the Issue
// file, stdio); the field mutation lives in internal/issue and the
// blocked computation in internal/list.
package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/issue"
	"github.com/Sanmoo/my-tasks2/internal/vault"
)

// newDepCmd builds `mt dep`: the parent of add and rm. A bare `mt dep`
// prints the group's help.
func newDepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dep",
		Short: "Manage Issue dependencies (blocked_by)",
		Long:  depLong,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newDepAddCmd(), newDepRmCmd())
	return cmd
}

// newDepAddCmd builds `mt dep add <id> <blocker>`: records that the
// Issue id is blocked by the Issue blocker, of the same Vault.
func newDepAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <id> <blocker>",
		Short: "Record that an Issue is blocked by another",
		Long: `add appends the blocker ID to the Issue's blocked_by: the Issue is
blocked — hidden from ready and pick-next, marked [blocked] in list —
until the blocker is done. The blocker must be an existing Issue of the
same Vault, and may not be the Issue itself.`,
		Args: depArgs("dep add"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDepAdd(cmd, args[0], args[1])
		},
	}
}

// newDepRmCmd builds `mt dep rm <id> <blocker>`: removes the blocker
// from the Issue's blocked_by.
func newDepRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <id> <blocker>",
		Short: "Remove a blocker from an Issue",
		Long: `rm removes the blocker ID from the Issue's blocked_by. The edit is
idempotent: removing a blocker that does not block the Issue leaves it
untouched, so stale references (e.g. to a deleted Issue) can be cleaned
up.`,
		Args: depArgs("dep rm"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDepRm(cmd, args[0], args[1])
		},
	}
}

// depArgs validates the two positional arguments of dep add and dep rm:
// exactly one subject ID and one blocker ID, each a single file name
// component. A malformed invocation is a usage error (exit 2).
func depArgs(use string) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if len(args) != 2 {
			return exitcode.Usage(fmt.Errorf("%s needs an issue ID and a blocker ID", use))
		}
		if err := checkID(args[0]); err != nil {
			return err
		}
		return checkID(args[1])
	}
}

// runDepAdd records blocker in id's blocked_by. The subject's vault is
// resolved by the key's ID prefix like every key command; the blocker's
// prefix is consulted only by the same-vault guard — never to address
// the blocker — so the intra-vault rule of blocked_by holds under
// prefix resolution. The subject and the blocker must both exist in
// the subject's Vault, and a blocker may not be the Issue itself —
// well-formed invocations that fail against the current state are
// user errors (exit 1), like any issue-not-found.
func runDepAdd(cmd *cobra.Command, id, blocker string) error {
	t, err := resolveVaultForKey(cmd, id)
	if err != nil {
		return err
	}
	if err := checkSameVaultBlocker(t.dir, blocker); err != nil {
		return err
	}
	if _, err := readIssue(t, id); err != nil {
		return err
	}
	if _, err := readIssue(t, blocker); err != nil {
		return err
	}
	if blocker == id {
		return fmt.Errorf("issue %s cannot block itself", id)
	}
	if _, err := mutateIssue(t, id, func(i issue.Issue) issue.Issue {
		return i.AddBlocker(blocker)
	}); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s is now blocked by %s\n", id, blocker)
	return nil
}

// checkSameVaultBlocker enforces the intra-vault rule of blocked_by: a
// blocker whose ID prefix matches a bookmarked vault other than dir is
// rejected, naming the vault it belongs to. A prefix matching only the
// subject's vault (or nothing) leaves the blocker to be looked up in
// the subject's vault, exactly as before prefix resolution.
func checkSameVaultBlocker(dir, blocker string) error {
	global, home, err := loadGlobalConfig()
	if err != nil {
		return err
	}
	for _, m := range vault.MatchByPrefix(blocker, global, home) {
		if filepath.Clean(m.Path) == filepath.Clean(dir) {
			continue
		}
		return fmt.Errorf("blocker %s belongs to @%s — blocked_by requires the same vault", blocker, m.Bookmark)
	}
	return nil
}

// runDepRm removes blocker from id's blocked_by. Only the subject
// resolves by its ID prefix; the blocker is a bare name to remove, as
// today — cleaning a stale reference stays idempotent even when the
// blocker belongs to another vault. The blocker need not exist in the
// Vault: removing a stale reference is a legitimate cleanup, and the
// edit is idempotent when the blocker is not listed.
func runDepRm(cmd *cobra.Command, id, blocker string) error {
	t, err := resolveVaultForKey(cmd, id)
	if err != nil {
		return err
	}
	if _, err := mutateIssue(t, id, func(i issue.Issue) issue.Issue {
		return i.RemoveBlocker(blocker)
	}); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s is no longer blocked by %s\n", id, blocker)
	return nil
}

const depLong = `dep manages Issue dependencies: the blocked_by field of an Issue lists
the IDs of the Issues that block it, all in the same Vault. An Issue is
blocked while any of its blockers is not done — computed state, not a
status: closing the blocker unblocks it on its own.

  mt dep add <id> <blocker>   record that <blocker> blocks <id>
  mt dep rm <id> <blocker>    remove <blocker> from <id>'s blocked_by

Blocked Issues are marked [blocked] in list and skipped by ready and
pick-next. mt check validates the references: existence, no self-block,
no cycles.`
