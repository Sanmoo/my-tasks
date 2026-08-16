// Package cli implements the ready and overdue Issue queries.
package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/list"
)

// newReadyCmd builds `mt ready`, which lists open Issues that are
// available now — not future-deferred and not blocked — in the vault's
// established priority order.
func newReadyCmd() *cobra.Command {
	return newIssueQueryCmd("ready", "List open Issues available now", func(item list.Item, now time.Time, statusByID map[string]string) bool {
		return list.Ready(item, now) && !list.Blocked(item.Issue.Frontmatter.BlockedBy, statusByID)
	})
}

// newOverdueCmd builds `mt overdue`, the vault's temporal-attention
// command: expired deferrals first (marked [expirada MM-DD]), then
// passed deadlines (marked [deadline MM-DD]), each group in the vault's
// priority order. The two-group output needs its own runner, unlike the
// single-predicate queries of newIssueQueryCmd.
func newOverdueCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "overdue",
		Short: "List Issues needing temporal attention (expired deferrals, passed deadlines)",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return exitcode.Usage(fmt.Errorf("overdue takes no arguments"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runOverdue(cmd)
		},
	}
}

// runOverdue loads and orders all Issues, then prints the two temporal
// groups: expired deferrals first, then passed deadlines, each line
// marked with the reason it is there. An Issue with both signals appears
// only in the expired group; done Issues appear in neither. It
// intentionally does not warn about duplicate ranks, like the other
// focused query views.
func runOverdue(cmd *cobra.Command) error {
	vaultDir, err := resolveVault(cmd)
	if err != nil {
		return err
	}
	items, err := loadSortedItems(vaultDir)
	if err != nil {
		return err
	}

	now := time.Now()
	expired, late := list.OverdueGroups(items, now)
	out := cmd.OutOrStdout()
	for _, it := range expired {
		line := formatListLine(it) + " " + list.ExpiredSuffix(it.Issue.Frontmatter.DeferredUntil, now)
		fmt.Fprintln(out, line)
	}
	for _, it := range late {
		line := formatListLine(it) + " " + list.DeadlineSuffix(it.Issue.Frontmatter.Deadline, now)
		fmt.Fprintln(out, line)
	}
	return nil
}

// newIssueQueryCmd builds a read-only query command over a vault's Issues.
// All queries keep list's priority order and line format, but apply their own
// eligibility rule. An empty result is a successful, empty output.
func newIssueQueryCmd(use, short string, matches func(list.Item, time.Time, map[string]string) bool) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return exitcode.Usage(fmt.Errorf("%s takes no arguments", use))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runIssueQuery(cmd, matches)
		},
	}
}

// runIssueQuery loads and orders all Issues, then prints those matched by the
// query. It intentionally does not warn about duplicate ranks: unlike list,
// these focused views do not serve as vault-integrity reporting.
func runIssueQuery(cmd *cobra.Command, matches func(list.Item, time.Time, map[string]string) bool) error {
	vaultDir, err := resolveVault(cmd)
	if err != nil {
		return err
	}
	items, err := loadSortedItems(vaultDir)
	if err != nil {
		return err
	}

	statusByID := list.StatusByID(items)
	now := time.Now()
	for _, item := range items {
		if matches(item, now, statusByID) {
			fmt.Fprintln(cmd.OutOrStdout(), formatListLine(item))
		}
	}
	return nil
}
