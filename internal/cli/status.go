// Package cli — Status commands: done (with close as alias), reopen and
// status. These own process concerns (files, stdio); the transition
// rules themselves live in internal/issue, and status validation in
// internal/vault.
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
	"github.com/Sanmoo/my-tasks2/internal/vault"
)

// newDoneCmd builds `mt done <id>` (alias `close`): closes the Issue —
// status done, completed_at stamped now.
func newDoneCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "done <id>",
		Aliases: []string{"close"},
		Short:   "Close an Issue (stamp completed_at)",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return exitcode.Usage(fmt.Errorf("done needs exactly one issue ID"))
			}
			return nil
		},
		ValidArgsFunction: completeIssueID,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMutation(cmd, args[0], func(i issue.Issue) issue.Issue {
				return i.Done(time.Now().Format(issue.NaiveLayout))
			})
		},
	}
}

// newReopenCmd builds `mt reopen <id>`: back to open, clearing
// completed_at and started_at.
func newReopenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reopen <id>",
		Short: "Reopen an Issue (clear completed_at and started_at)",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return exitcode.Usage(fmt.Errorf("reopen needs exactly one issue ID"))
			}
			return nil
		},
		ValidArgsFunction: completeIssueID,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMutation(cmd, args[0], func(i issue.Issue) issue.Issue {
				return i.Reopen()
			})
		},
	}
}

// newStatusCmd builds `mt status <id> <status>`: the free transition,
// validated against the vault's configured status list.
func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <id> <status>",
		Short: "Set an Issue's status (free transition)",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return exitcode.Usage(fmt.Errorf("status needs an issue ID and a status"))
			}
			return nil
		},
		ValidArgsFunction: completeIssueID,
		RunE: func(cmd *cobra.Command, args []string) error {
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
			return applyMutation(cmd, t, args[0], func(i issue.Issue) issue.Issue {
				return i.SetStatus(args[1])
			})
		},
	}
}

// runMutation is the shared body of done and reopen: resolve the vault
// from the key, then apply the mutation to the Issue. Resolution errors
// already name the failing step, so they propagate unwrapped.
func runMutation(cmd *cobra.Command, id string, mutate func(issue.Issue) issue.Issue) error {
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
func applyMutation(cmd *cobra.Command, t vaultTarget, id string, mutate func(issue.Issue) issue.Issue) error {
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
// result. Callers own any command-specific confirmation output. When
// the Issue is missing, the error carries the target's vault context
// (prefix-picked vault named, or hint for explicit selections).
func mutateIssue(t vaultTarget, id string, mutate func(issue.Issue) issue.Issue) (issue.Issue, error) {
	if err := checkID(id); err != nil {
		return issue.Issue{}, err
	}
	i, err := readIssue(t, id)
	if err != nil {
		return issue.Issue{}, err
	}
	i = mutate(i)
	if err := writeIssueFile(t.dir, id, i); err != nil {
		return issue.Issue{}, err
	}
	return i, nil
}

// readIssue loads and parses the Issue file for id in the target
// vault; a missing Issue produces the target's not-found error, which
// names the vault the key's prefix picked and appends the hint for
// explicit selections. O_NOFOLLOW keeps a symlink in the issues
// directory from redirecting the read outside the Vault, including if
// the path changes after discovery.
func readIssue(t vaultTarget, id string) (issue.Issue, error) {
	data, err := readIssueData(t.dir, id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return issue.Issue{}, t.notFoundError(id)
		}
		return issue.Issue{}, fmt.Errorf("reading issue %s: %w", id, err)
	}
	i, err := issue.Parse(data)
	if err != nil {
		return issue.Issue{}, fmt.Errorf("parsing issue %s: %w", id, err)
	}
	return i, nil
}

// writeIssueFile renders i and writes it back to its file in the vault.
// It is the shared render-and-persist tail of the mutating commands; the
// O_NOFOLLOW flag prevents a symlink from redirecting the write outside the
// Vault. The confirmation line is the caller's concern.
func writeIssueFile(vaultDir, id string, i issue.Issue) error {
	data, err := issue.Render(i)
	if err != nil {
		return err
	}
	f, err := openIssueFile(vaultDir, id, os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("writing issue %s: %w", id, err)
	}
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr != nil {
		return fmt.Errorf("writing issue %s: %w", id, writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("closing issue %s: %w", id, closeErr)
	}
	return nil
}
