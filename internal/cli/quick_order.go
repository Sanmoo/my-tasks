package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/priority"
)

// newTopCmd builds `mt top <id>`: promotes an Issue to the first position
// in the queue without opening an editor.
func newTopCmd() *cobra.Command {
	return newQuickOrderCmd("top", "Move an Issue to the top of the queue", priority.MoveTop)
}

// newBottomCmd builds `mt bottom <id>`: moves an Issue to the last position
// in the queue without opening an editor.
func newBottomCmd() *cobra.Command {
	return newQuickOrderCmd("bottom", "Move an Issue to the bottom of the queue", priority.MoveBottom)
}

func newQuickOrderCmd(name, short string, action priority.QuickAction) *cobra.Command {
	return &cobra.Command{
		Use:   name + " <id>",
		Short: short,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return exitcode.Usage(fmt.Errorf("%s needs exactly one issue ID", name))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuickOrder(cmd, args[0], action, 0)
		},
	}
}

// newUnrankCmd builds `mt unrank <id>`: returns an Issue to the Backlog.
func newUnrankCmd() *cobra.Command {
	return newQuickOrderCmd("unrank", "Return an Issue to the Backlog", priority.RemoveRank)
}

// newRankCmd builds `mt rank <id> <n>`: inserts an Issue at the requested
// one-based position in the queue without opening an editor.
func newRankCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rank <id> <n>",
		Short: "Insert an Issue at a queue position",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return exitcode.Usage(fmt.Errorf("rank needs an issue ID and a position"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			position, err := strconv.Atoi(args[1])
			if err != nil || position < 1 {
				return exitcode.Usage(fmt.Errorf("rank position must be a positive integer"))
			}
			return runQuickOrder(cmd, args[0], priority.MoveToRank, position)
		},
	}
}

// runQuickOrder applies a pure quick-order plan and rewrites only the Issues
// whose ranks changed, just like `mt prioritize`.
func runQuickOrder(cmd *cobra.Command, id string, action priority.QuickAction, position int) error {
	vaultDir, err := resolveVault(cmd)
	if err != nil {
		return err
	}
	issues, err := loadPriorityIssues(vaultDir)
	if err != nil {
		return err
	}
	changes, err := priority.QuickPlan(issues, id, action, position)
	if err != nil {
		return err
	}
	if err := applyRankChanges(vaultDir, changes); err != nil {
		return err
	}
	word := "issues"
	if len(changes) == 1 {
		word = "issue"
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Updated %d %s\n", len(changes), word)
	return nil
}
