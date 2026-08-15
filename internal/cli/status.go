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
		RunE: func(cmd *cobra.Command, args []string) error {
			vaultDir, err := resolveVault(cmd)
			if err != nil {
				return fmt.Errorf("resolving vault: %w", err)
			}
			vcfg, err := vault.LoadVault(vaultDir)
			if err != nil {
				return fmt.Errorf("loading vault: %w", err)
			}
			if !vcfg.IsStatus(args[1]) {
				return fmt.Errorf("status %q is not in the vault's status list (valid: %s)",
					args[1], strings.Join(vcfg.StatusList(), ", "))
			}
			return applyMutation(cmd, vaultDir, args[0], func(i issue.Issue) issue.Issue {
				return i.SetStatus(args[1])
			})
		},
	}
}

// runMutation is the shared body of done and reopen: resolve the vault,
// then apply the mutation to the Issue.
func runMutation(cmd *cobra.Command, id string, mutate func(issue.Issue) issue.Issue) error {
	vaultDir, err := resolveVault(cmd)
	if err != nil {
		return fmt.Errorf("resolving vault: %w", err)
	}
	if err := applyMutation(cmd, vaultDir, id, mutate); err != nil {
		return fmt.Errorf("mutating issue: %w", err)
	}
	return nil
}

// applyMutation applies and persists a mutation, then prints the new
// status. It is the shared tail of done, reopen and status.
func applyMutation(cmd *cobra.Command, vaultDir, id string, mutate func(issue.Issue) issue.Issue) error {
	i, err := mutateIssue(vaultDir, id, mutate)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s is now %s\n", id, i.Frontmatter.Status)
	return nil
}

// mutateIssue loads the Issue for id, applies mutate and persists the
// result. Callers own any command-specific confirmation output.
func mutateIssue(vaultDir, id string, mutate func(issue.Issue) issue.Issue) (issue.Issue, error) {
	if err := checkID(id); err != nil {
		return issue.Issue{}, err
	}
	i, err := readIssue(vaultDir, id)
	if err != nil {
		return issue.Issue{}, err
	}
	i = mutate(i)
	if err := writeIssueFile(vaultDir, id, i); err != nil {
		return issue.Issue{}, err
	}
	return i, nil
}

// readIssue loads and parses the Issue file for id inside vaultDir.
// O_NOFOLLOW keeps a symlink in the issues directory from redirecting the
// read outside the Vault, including if the path changes after discovery.
func readIssue(vaultDir, id string) (issue.Issue, error) {
	data, err := readIssueData(vaultDir, id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return issue.Issue{}, fmt.Errorf("issue %s not found", id)
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
