// Package cli — the mt claim commands. They own the process concerns of
// recording possession (resolving the vault, reading/writing the Issue
// file, stdio, the open-only guard); the field write lives in
// internal/issue and the unavailability it implies in internal/list.
package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// newClaimCmd builds `mt claim <id> <who>`: records who assumed the
// Issue before any work started. Claim is allowed only on an open Issue
// — claim means a reservation, not a start.
func newClaimCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "claim <id> <who>",
		Short: "Record who claimed an Issue",
		Long: claimLong,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return exitcode.Usage(fmt.Errorf("claim needs an issue ID and the claimer (mt claim <id> <who>)"))
			}
			return nil
		},
		ValidArgsFunction: completeIssueID,
		RunE: func(cmd *cobra.Command, args []string) error {
			// The claimer may be spelled multi-word; join the remainder
			// so both forms parse, like defer does with its time.
			return runClaim(cmd, args[0], strings.Join(args[1:], " "))
		},
	}
}

// newUnclaimCmd builds `mt unclaim <id>`: clears the claimed_by field,
// returning the Issue to availability. Allowed on any status — rare
// manual corrections (a stale claim on a done Issue) stay possible
// without hand-editing the file.
func newUnclaimCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unclaim <id>",
		Short: "Clear an Issue's claim",
		Long:  unclaimLong,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return exitcode.Usage(fmt.Errorf("unclaim needs an issue ID"))
			}
			return nil
		},
		ValidArgsFunction: completeIssueID,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnclaim(cmd, args[0])
		},
	}
}

// runClaim resolves the vault, validates the open-only guard and the
// claimer value, and writes claimed_by onto the Issue.
func runClaim(cmd *cobra.Command, id, rawWho string) error {
	who := strings.Join(strings.Fields(rawWho), " ")
	if who == "" {
		// A claim without a claimer would keep a field whose only job
		// — recording possession — is empty: a malformed invocation.
		return exitcode.Usage(fmt.Errorf("claim needs the claimer (mt claim <id> <who>)"))
	}
	t, err := resolveVaultForKey(cmd, id)
	if err != nil {
		return err
	}
	current, err := readIssue(t, id)
	if err != nil {
		return err
	}
	// Claim means a reservation of work that has not started: an Issue
	// already in_progress, done or in a custom status cannot be claimed.
	if current.Frontmatter.Status != "open" {
		return fmt.Errorf("issue %s is %s — claim requires an open Issue", id, current.Frontmatter.Status)
	}
	if _, err := mutateIssue(t, id, func(i issue.Issue) (issue.Issue, error) {
		return i.Claim(who), nil
	}); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s claimed by %s\n", id, who)
	return nil
}

// runUnclaim clears the Issue's claimed_by. Unclaiming an Issue with no
// claim fails against the current state: a user error, like any other
// operation on state that does not exist.
func runUnclaim(cmd *cobra.Command, id string) error {
	t, err := resolveVaultForKey(cmd, id)
	if err != nil {
		return err
	}
	current, err := readIssue(t, id)
	if err != nil {
		return err
	}
	if current.Frontmatter.ClaimedBy == "" {
		return fmt.Errorf("issue %s has no claim", id)
	}
	if _, err := mutateIssue(t, id, func(i issue.Issue) (issue.Issue, error) {
		return i.Unclaim(), nil
	}); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s unclaimed\n", id)
	return nil
}

const claimLong = `claim records who assumed the Issue before any work started:

  the claimed_by field is written and the claimer identity is required —
  a claim without a claimer is not a claim. Claim is a field, not a
  status: the Issue keeps its status, exactly as it is.

Only an open Issue can be claimed (claim is a reservation, not a
start); the claim removes availability, so ready and pick-next skip
the Issue until mt unclaim clears it, mt reopen wipes it, or mt done
archives it as history. Claiming an already-claimed Issue overwrites
the claim: possession can be transferred explicitly.`

const unclaimLong = `unclaim clears an Issue's claimed_by field: the Issue is back in
normal circulation immediately. Unclaiming an Issue with no claim is a
user error; the command otherwise works on any status, so a stale
claim on a done Issue can be corrected without hand-editing the file.`
