// Package cli — the mt list command. It owns the process concerns of
// listing (reading the issue files, filtering, rendering, the duplicate
// rank warning); the ordering and glyph decisions live in
// internal/list.
package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/list"
)

// newListCmd builds `mt list`: issues in priority order (rank → backlog
// by created_at → id), one glyph per status, done and future-deferred
// hidden by default (--all reveals them, future-deferred marked with a
// [defer ...] suffix), filterable by --status and --label.
func newListCmd() *cobra.Command {
	var all bool
	var statusFilter string
	var labelFilters []string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues in priority order",
		Long:  listLong,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return exitcode.Usage(fmt.Errorf("list takes no arguments"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, all, statusFilter, labelFilters)
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "show done and future-deferred issues (deferred marked with a suffix)")
	cmd.Flags().StringVar(&statusFilter, "status", "", "only issues with this status")
	cmd.Flags().StringArrayVar(&labelFilters, "label", nil, "only issues with this label; repeatable")
	return cmd
}

// runList loads the vault's issues, sorts them, and prints them with the
// given filters. The duplicate-rank warning goes to stderr and is
// computed over the whole vault, before any filter, so a filtered view
// still reports vault integrity.
func runList(cmd *cobra.Command, all bool, statusFilter string, labelFilters []string) error {
	vaultDir, err := resolveVault(cmd)
	if err != nil {
		return err
	}
	items, err := loadSortedItems(vaultDir)
	if err != nil {
		return err
	}

	if dups := list.DuplicateRanks(items); len(dups) > 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), duplicateRanksWarning(dups))
	}

	now := time.Now()
	out := cmd.OutOrStdout()
	opts := list.Options{All: all, Status: statusFilter, Labels: labelFilters}
	for _, it := range items {
		if !list.Visible(it, opts, now) {
			continue
		}
		line := formatListLine(it)
		if all {
			if suffix := list.DeferSuffix(it.Issue.Frontmatter.DeferredUntil, now); suffix != "" {
				line += " " + suffix
			}
		}
		fmt.Fprintln(out, line)
	}
	return nil
}

// loadSortedItems reads every Issue in the vault and orders the result
// according to the shared list order.
func loadSortedItems(vaultDir string) ([]list.Item, error) {
	items, err := loadItems(vaultDir)
	if err != nil {
		return nil, err
	}
	list.Sort(items)
	return items, nil
}

// formatListLine renders the standard one-line Issue representation used by
// list and its focused query views.
func formatListLine(it list.Item) string {
	fm := it.Issue.Frontmatter
	return fmt.Sprintf("%s %s  %s", list.Glyph(fm.Status), it.ID, fm.Title)
}

// loadItems reads every *.md file in the vault's issues/ directory and
// parses it into a list item. The file name (minus .md) is the ID; a
// malformed file fails the whole list with the offending ID named.
func loadItems(vaultDir string) ([]list.Item, error) {
	files, err := readIssueFiles(vaultDir)
	if err != nil {
		return nil, err
	}
	items := make([]list.Item, 0, len(files))
	for _, file := range files {
		items = append(items, list.Item{ID: file.ID, Issue: file.Issue})
	}
	return items, nil
}

// duplicateRanksWarning renders the duplicate-rank warning line.
func duplicateRanksWarning(dups []int) string {
	strs := make([]string, len(dups))
	for i, r := range dups {
		strs[i] = strconv.Itoa(r)
	}
	label := "rank"
	if len(dups) > 1 {
		label = "ranks"
	}
	return fmt.Sprintf("Warning: duplicate %s: %s", label, strings.Join(strs, ", "))
}

const listLong = `list prints the vault's issues in priority order:

  ranked issues first, lowest rank first; then the Backlog (issues
  without a rank) ordered by created_at; then ID as the final tiebreak.

Each line is a status glyph (○ open, ◐ in_progress, ● done, ? for a
custom status), the ID, and the title. done issues and issues deferred
to the future are hidden by default; --all shows them, marking
future-deferred issues with a [defer MM-DD HH:MM] suffix. Use --status
and --label to narrow the view.`
