// Package cli implements the ready and overdue Issue queries.
package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/list"
)

// newReadyCmd builds `mt ready`, which lists open Issues that are available
// now in the vault's established priority order.
func newReadyCmd() *cobra.Command {
	return newIssueQueryCmd("ready", "List open Issues available now", func(item list.Item, now time.Time) bool {
		return list.Ready(item, now)
	})
}

// newOverdueCmd builds `mt overdue`, which lists non-done Issues whose
// informational Deadline has passed.
func newOverdueCmd() *cobra.Command {
	return newIssueQueryCmd("overdue", "List non-done Issues with a passed Deadline", func(item list.Item, now time.Time) bool {
		return list.Overdue(item, now)
	})
}

// newIssueQueryCmd builds a read-only query command over a vault's Issues.
// All queries keep list's priority order and line format, but apply their own
// eligibility rule. An empty result is a successful, empty output.
func newIssueQueryCmd(use, short string, matches func(list.Item, time.Time) bool) *cobra.Command {
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
func runIssueQuery(cmd *cobra.Command, matches func(list.Item, time.Time) bool) error {
	vaultDir, err := resolveVault(cmd)
	if err != nil {
		return err
	}
	items, err := loadItems(vaultDir)
	if err != nil {
		return err
	}
	list.Sort(items)

	now := time.Now()
	for _, item := range items {
		if matches(item, now) {
			fm := item.Issue.Frontmatter
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s  %s\n", list.Glyph(fm.Status), item.ID, fm.Title)
		}
	}
	return nil
}
