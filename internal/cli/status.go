// Package cli — Status commands: done (with close as alias), reopen and
// status. These own process concerns (files, stdio); the transition
// rules themselves live in internal/issue, and status validation in
// internal/vault. Each command accepts an optional trailing comment and
// writes it, when given, in the same write as the transition.
package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/issue"
	"github.com/Sanmoo/my-tasks2/internal/issuefiles"
	"github.com/Sanmoo/my-tasks2/internal/vault"
)

// transitionCommentLong documents the trailing-comment rule that done,
// reopen and status share. The rule itself lives in commentText and
// addComment; this is how the tool explains it in `mt help`.
const transitionCommentLong = `A trailing comment (the arguments after the ID, joined with spaces) is
appended to the Issue's Comments section exactly as mt comment writes one —
heading, text and stable anchor — in the same write as the transition, so
the comment and the transition carry one timestamp. A blank comment is a
usage error (exit 2).`

// newDoneCmd builds `mt done <id> [comment]` (alias `close`): closes the
// Issue — status done, completed_at stamped now — appending the trailing
// comment, when given, in the same write.
func newDoneCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "done <id> [comment]",
		Aliases: []string{"close"},
		Short:   "Close an Issue (stamp completed_at)",
		Long:    "done closes the Issue: status done and completed_at stamped now.\n\n" + transitionCommentLong,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return exitcode.Usage(errors.New("done needs an issue ID"))
			}
			return nil
		},
		ValidArgsFunction: completeIssueID,
		RunE: func(cmd *cobra.Command, args []string) error {
			text, err := commentText(args[1:])
			if err != nil {
				return err
			}
			now := time.Now().Format(issue.NaiveLayout)
			return runMutation(cmd, args[0], func(i issue.Issue) (issue.Issue, error) {
				return withTransitionComment(i.Done(now), text, now)
			})
		},
	}
}

// newReopenCmd builds `mt reopen <id> [comment]`: back to open, clearing
// completed_at and started_at, appending the trailing comment, when given,
// in the same write.
func newReopenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reopen <id> [comment]",
		Short: "Reopen an Issue (clear completed_at and started_at)",
		Long:  "reopen returns the Issue to open, clearing completed_at and started_at.\n\n" + transitionCommentLong,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return exitcode.Usage(errors.New("reopen needs an issue ID"))
			}
			return nil
		},
		ValidArgsFunction: completeIssueID,
		RunE: func(cmd *cobra.Command, args []string) error {
			text, err := commentText(args[1:])
			if err != nil {
				return err
			}
			now := time.Now().Format(issue.NaiveLayout)
			return runMutation(cmd, args[0], func(i issue.Issue) (issue.Issue, error) {
				return withTransitionComment(i.Reopen(), text, now)
			})
		},
	}
}

// newStatusCmd builds `mt status <id> <status> [comment]`: the free
// transition, validated against the vault's configured status list,
// appending the trailing comment, when given, in the same write.
func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <id> <status> [comment]",
		Short: "Set an Issue's status (free transition)",
		Long:  "status sets the Issue's status to any status the Vault lists, without touching timestamps.\n\n" + transitionCommentLong,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return exitcode.Usage(errors.New("status needs an issue ID and a status"))
			}
			return nil
		},
		ValidArgsFunction: completeIssueID,
		RunE: func(cmd *cobra.Command, args []string) error {
			text, err := commentText(args[2:])
			if err != nil {
				return err
			}
			t, err := resolveVaultForKey(cmd, args[0])
			if err != nil {
				return err
			}
			vcfg, err := vault.LoadVault(t.dir)
			if err != nil {
				return fmt.Errorf("loading vault: %w", err)
			}
			if !vcfg.IsStatus(args[1]) {
				return fmt.Errorf("status %q is not in the vault's status list (valid: %s)",
					args[1], strings.Join(vcfg.StatusList(), ", "))
			}
			now := time.Now().Format(issue.NaiveLayout)
			return applyMutation(cmd, t, args[0], func(i issue.Issue) (issue.Issue, error) {
				return withTransitionComment(i.SetStatus(args[1]), text, now)
			})
		},
	}
}

// withTransitionComment returns i with text appended as a Comment when the
// transition carries one, or i unchanged when it does not. The heading and
// the transition's own timestamps come from the caller's single now read,
// so a comment and the transition it rides share one instant.
func withTransitionComment(i issue.Issue, text, now string) (issue.Issue, error) {
	if text == "" {
		return i, nil
	}
	return addComment(i, text, now)
}

// runMutation is the shared body of done and reopen: resolve the vault
// from the key, then apply the mutation to the Issue. Resolution errors
// already name the failing step, so they propagate unwrapped.
func runMutation(cmd *cobra.Command, id string, mutate func(issue.Issue) (issue.Issue, error)) error {
	t, err := resolveVaultForKey(cmd, id)
	if err != nil {
		return err
	}
	if err := applyMutation(cmd, t, id, mutate); err != nil {
		return fmt.Errorf("mutating issue: %w", err)
	}
	return nil
}

// applyMutation applies and persists a mutation, then prints the new
// status — with the Issue's title when it has one. It is the shared
// tail of done, reopen, status and pick-next.
func applyMutation(cmd *cobra.Command, t vaultTarget, id string, mutate func(issue.Issue) (issue.Issue, error)) error {
	i, err := mutateIssue(t, id, mutate)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), transitionLine(id, i))
	return nil
}

// transitionLine renders the transition confirmation for id: "id is now
// status", plus the shared optional title suffix.
func transitionLine(id string, i issue.Issue) string {
	return fmt.Sprintf("%s is now %s%s", id, i.Frontmatter.Status, titleSuffix(i.Frontmatter.Title))
}

// titleSuffix returns ": title" when title has content. It collapses every
// whitespace run to one space so confirmations always occupy one line.
func titleSuffix(title string) string {
	if title = strings.Join(strings.Fields(title), " "); title != "" {
		return ": " + title
	}
	return ""
}

// mutateIssue loads the Issue for id, applies mutate and persists the
// result through the store's single write door. A mutation that cannot
// be applied — a comment anchor that cannot be allocated — fails inside
// the store before anything is written, so the Issue file is either the
// old one or the fully mutated one. Callers own any command-specific
// confirmation output. When the Issue is missing, the error carries the
// target's vault context (prefix-picked vault named, or hint for
// explicit selections).
func mutateIssue(t vaultTarget, id string, mutate func(issue.Issue) (issue.Issue, error)) (issue.Issue, error) {
	if err := checkID(id); err != nil {
		return issue.Issue{}, err
	}
	item, err := issuefiles.Open(t.dir).Mutate(id, mutate)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return issue.Issue{}, t.notFoundError(id)
		}
		return issue.Issue{}, err
	}
	return item.Issue, nil
}

// readIssue loads the Issue file for id in the target vault through
// the store, which refuses a symlink (or any non-regular file) before
// reading it. A missing Issue produces the target's not-found error,
// which names the vault the key's prefix picked and appends the hint
// for explicit selections; the store's own errors already name the
// operation and the ID, so they propagate unchanged.
func readIssue(t vaultTarget, id string) (issue.Issue, error) {
	item, err := issuefiles.Open(t.dir).Read(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return issue.Issue{}, t.notFoundError(id)
		}
		return issue.Issue{}, err
	}
	return item.Issue, nil
}
