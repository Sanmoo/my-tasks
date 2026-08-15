// Package cli — the mt prioritize command. It owns the process concerns
// of the $EDITOR flow (the temp buffer file, running the editor, reading
// the issue files, applying the rank changes in-process); the buffer
// format, parsing and the rank renormalization plan live in
// internal/priority.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/issue"
	"github.com/Sanmoo/my-tasks2/internal/list"
	"github.com/Sanmoo/my-tasks2/internal/priority"
)

// newPrioritizeCmd builds `mt prioritize`: opens $EDITOR on a buffer of
// the vault's open and in_progress issues ([P]/[ ] lines), then applies
// the saved order — reordered [P] lines change the ranks, [ ]↔[P] toggles
// move an issue between the queue and the Backlog. The ranks are
// renormalized 1..N and only the files whose rank changed are rewritten.
func newPrioritizeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prioritize",
		Short: "Prioritize Issues in $EDITOR",
		Long:  prioritizeLong,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return exitcode.Usage(fmt.Errorf("prioritize takes no arguments"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrioritize(cmd)
		},
	}
}

// runPrioritize is the prioritize flow: load the vault's issues, build
// the buffer, run $EDITOR, parse the saved buffer, plan the rank changes
// and apply them — all in-process, one file rewrite per changed issue.
func runPrioritize(cmd *cobra.Command) error {
	vaultDir, err := resolveVault(cmd)
	if err != nil {
		return err
	}
	issues, err := loadPriorityIssues(vaultDir)
	if err != nil {
		return err
	}
	prioritizable := make([]priority.Issue, 0, len(issues))
	for _, is := range issues {
		if priority.Prioritizable(is.Status) {
			prioritizable = append(prioritizable, is)
		}
	}

	path, err := writeBuffer(priority.Buffer(prioritizable))
	if err != nil {
		return err
	}
	defer os.Remove(path)

	if err := editFile(path); err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading prioritize buffer: %w", err)
	}
	entries, err := priority.Parse(string(data))
	if err != nil {
		return err
	}
	changes, err := priority.Plan(entries, issues)
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

// writeBuffer writes the editor buffer to a fresh temp file and returns
// its path. The temp file is the $EDITOR target: the editor overwrites it
// and the saved contents are read back.
func writeBuffer(buffer string) (string, error) {
	f, err := os.CreateTemp("", "mt-prioritize-*.md")
	if err != nil {
		return "", fmt.Errorf("creating prioritize buffer: %w", err)
	}
	path := f.Name()
	_, writeErr := f.WriteString(buffer)
	closeErr := f.Close()
	switch {
	case writeErr != nil:
		os.Remove(path)
		return "", fmt.Errorf("writing prioritize buffer: %w", writeErr)
	case closeErr != nil:
		os.Remove(path)
		return "", fmt.Errorf("closing prioritize buffer: %w", closeErr)
	}
	return path, nil
}

// loadPriorityIssues reuses the CLI's shared issue-file loader and
// projects each parsed Issue into the minimal priority.Issue view.
func loadPriorityIssues(vaultDir string) ([]priority.Issue, error) {
	items, err := loadItems(vaultDir)
	if err != nil {
		return nil, err
	}
	return priorityIssuesFromItems(items), nil
}

func priorityIssueFrom(id string, i issue.Issue) priority.Issue {
	fm := i.Frontmatter
	return priority.Issue{
		ID:        id,
		Title:     fm.Title,
		Status:    fm.Status,
		Rank:      fm.Rank,
		CreatedAt: fm.CreatedAt,
	}
}

func priorityIssuesFromItems(items []list.Item) []priority.Issue {
	issues := make([]priority.Issue, 0, len(items))
	for _, item := range items {
		issues = append(issues, priorityIssueFrom(item.ID, item.Issue))
	}
	return issues
}

// applyRankChanges applies each rank change in-process, without spawning a
// subprocess per issue.
func applyRankChanges(vaultDir string, changes []priority.Change) error {
	for _, ch := range changes {
		if err := applyRankChange(vaultDir, ch); err != nil {
			return err
		}
	}
	return nil
}

// applyRankChange writes the new rank (nil = Backlog) into the issue file,
// preserving everything else.
func applyRankChange(vaultDir string, ch priority.Change) error {
	if err := checkID(ch.ID); err != nil {
		return err
	}
	i, err := readIssue(vaultDir, ch.ID)
	if err != nil {
		return err
	}
	i.Frontmatter.Rank = ch.Rank
	return writeIssueFile(vaultDir, ch.ID, i)
}

const prioritizeLong = `prioritize opens $EDITOR on a buffer of the vault's open and in_progress
issues, one line each:

  [P] pkm-055  <title>   — prioritized (in the queue)
  [ ] pkm-5qa8  <title>   — Backlog (not prioritized)

Reorder the [P] lines to change their order; toggle [ ]↔[P] to move an
issue between the queue and the Backlog. Save and close to apply: ranks
are renormalized 1..N and only the files whose rank changed are rewritten.

An invalid buffer (unknown or duplicated issue ID) is rejected and
nothing is applied.`
