// Package cli — the mt pick-next command. It owns the process concerns
// of choosing and starting the next Issue; Rank, availability and
// duplicate-rank decisions live in internal/list.
package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/issue"
	"github.com/Sanmoo/my-tasks2/internal/list"
)

// newPickNextCmd builds `mt pick-next`: starts the available open Issue with
// the lowest Rank, or the oldest available Backlog Issue when no ranked
// candidate exists. It allows multiple Issues to remain in_progress simultaneously.
func newPickNextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pick-next",
		Short: "Start the next available Issue",
		Long:  pickNextLong,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return exitcode.Usage(fmt.Errorf("pick-next takes no arguments"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPickNext(cmd)
		},
	}
}

// runPickNext resolves the Vault, validates its ranks, selects an available
// open Issue and starts it with one timestamp shared by selection and write.
func runPickNext(cmd *cobra.Command) error {
	vaultDir, err := resolveVault(cmd)
	if err != nil {
		return fmt.Errorf("resolving vault: %w", err)
	}
	items, err := loadItems(vaultDir)
	if err != nil {
		return fmt.Errorf("loading issues: %w", err)
	}
	now := time.Now()
	next, err := list.PickNext(items, now)
	if err != nil {
		return fmt.Errorf("selecting next issue: %w", err)
	}
	if err := applyMutation(cmd, vaultDir, next.ID, func(i issue.Issue) issue.Issue {
		return i.Start(now.Format(issue.NaiveLayout))
	}); err != nil {
		return fmt.Errorf("starting issue: %w", err)
	}
	return nil
}

const pickNextLong = `pick-next starts the next available open Issue:

  ranked Issues are chosen by the lowest Rank; when no ranked Issue is
  available, the oldest available Backlog Issue is chosen, with ID as the
  final tie-break. Issues deferred into the future are skipped.

The chosen Issue becomes in_progress and receives a started_at timestamp.
Duplicate ranks are rejected, and multiple Issues may be in_progress at once.`
