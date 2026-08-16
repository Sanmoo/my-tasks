// Package cli — the mt undefer command. It owns the process concerns of
// undeferral (resolving the vault, reading/writing Issue files, stdio);
// the expired-deferral predicate lives in internal/list and the field
// write in internal/issue.
package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/issue"
	"github.com/Sanmoo/my-tasks2/internal/list"
)

// newUndeferCmd builds `mt undefer [id]`: without an ID, it clears
// deferred_until from every expired deferral in the vault (the
// "archive the reminder" step of the daily overdue flow); with an ID,
// it clears that one Issue, even a still-future deferral — the user
// changed their mind. Only the deferred_until field is touched: Status
// and Rank stay exactly as they are.
func newUndeferCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "undefer [id]",
		Short: "Clear deferred_until (all expired deferrals, or one Issue)",
		Long:  undeferLong,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 1 {
				return exitcode.Usage(fmt.Errorf("undefer takes at most one issue ID"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return runUndeferOne(cmd, args[0])
			}
			return runUndeferAll(cmd)
		},
	}
}

// runUndeferOne clears deferred_until from one specific Issue, even
// when the deferral is still in the future. An Issue without the field
// is a user error (exit 1): the target is wrong.
func runUndeferOne(cmd *cobra.Command, id string) error {
	vaultDir, err := resolveVault(cmd)
	if err != nil {
		return err
	}
	if err := checkID(id); err != nil {
		return err
	}
	i, err := readIssue(vaultDir, id)
	if err != nil {
		return err
	}
	if i.Frontmatter.DeferredUntil == "" {
		return fmt.Errorf("issue %s has no deferred_until to undefer", id)
	}
	was := i.Frontmatter.DeferredUntil
	if _, err := mutateIssue(vaultDir, id, func(i issue.Issue) issue.Issue {
		return i.Undefer()
	}); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Undeferred %s (was %s)\n", id, was)
	return nil
}

// runUndeferAll sweeps the vault's Issues in their priority order,
// clearing deferred_until from every expired deferral and printing one
// Undeferred line per Issue. An expired deferred_until has no remaining
// functional effect (availability already returned — see ADR-0002), so
// there is no confirmation prompt; an empty sweep prints nothing and
// succeeds, keeping the daily habit runnable blind.
func runUndeferAll(cmd *cobra.Command) error {
	vaultDir, err := resolveVault(cmd)
	if err != nil {
		return err
	}
	items, err := loadSortedItems(vaultDir)
	if err != nil {
		return err
	}
	now := time.Now()
	out := cmd.OutOrStdout()
	for _, it := range items {
		until := it.Issue.Frontmatter.DeferredUntil
		if !list.DeferralExpired(until, now) {
			continue
		}
		if _, err := mutateIssue(vaultDir, it.ID, func(i issue.Issue) issue.Issue {
			return i.Undefer()
		}); err != nil {
			return err
		}
		fmt.Fprintf(out, "Undeferred %s (was %s)\n", it.ID, until)
	}
	return nil
}

const undeferLong = `undefer clears an Issue's deferred_until. Without an
ID, it sweeps the vault: every expired deferral is cleared, printing one
"Undeferred <id> (was <datetime>)" line per Issue — with nothing
expired, it prints nothing and succeeds. With an ID, it clears that one
Issue even when its deferral is still in the future (you changed your
mind). Only the deferred_until field is touched: Status and Rank stay as
they are, and availability (mt ready, mt pick-next) is unaffected.`
