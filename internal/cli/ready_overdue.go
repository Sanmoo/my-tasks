// Package cli implements the ready and overdue Issue queries.
package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/list"
)

// newReadyCmd builds `mt ready`, which lists the Vault's available
// Issues — list.Available is the single availability predicate, so this
// command applies no rule of its own — in the vault's established
// priority order.
func newReadyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ready",
		Short: "List open Issues available now",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return exitcode.Usage(fmt.Errorf("ready takes no arguments"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runReady(cmd)
		},
	}
}

// newOverdueCmd builds `mt overdue`, the vault's temporal-attention
// command: expired deferrals first (marked [expirada MM-DD]), then
// passed deadlines (marked [deadline MM-DD]), each group in the vault's
// priority order. Temporal attention is not availability (ADR-0006), so
// the two-group output has its own runner.
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

// runReady loads and orders all Issues, then prints the available ones.
// It intentionally does not warn about duplicate ranks: unlike list,
// this focused view does not serve as vault-integrity reporting.
func runReady(cmd *cobra.Command) error {
	vaultDir, err := resolveVault(cmd)
	if err != nil {
		return err
	}
	items, err := loadSortedItems(vaultDir)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	for _, item := range list.Available(items, time.Now()) {
		fmt.Fprintln(out, formatListLine(item))
	}
	return nil
}
