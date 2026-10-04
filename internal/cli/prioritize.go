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
	"github.com/Sanmoo/my-tasks2/internal/issuefiles"
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
	items, err := loadItems(vaultDir)
	if err != nil {
		return err
	}
	prioritizable := make([]issue.Item, 0, len(items))
	for _, it := range items {
		if priority.Prioritizable(it.Issue.Frontmatter.Status) {
			prioritizable = append(prioritizable, it)
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
	changes, err := priority.Plan(entries, items)
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

// applyRankChanges applies each rank change in-process, without spawning a
// subprocess per issue.
func applyRankChanges(vaultDir string, changes []issue.Change) error {
	for _, ch := range changes {
		if err := writeRank(vaultDir, ch.ID, ch.Rank); err != nil {
			return err
		}
	}
	return nil
}

// writeRank is the single Rank write path of the CLI: it validates the ID —
// a malformed ID is a usage error (exit 2), never a write outside the Vault —
// and persists rank (nil = Backlog) through the store's Mutate door,
// preserving every other field. mt prioritize, mt check --fix, the quick-order
// commands (rank/top/bottom/unrank), mt place and create/q's --top/--bottom
// all funnel through it, so the ID validation cannot be forgotten in one of
// them. A missing Issue produces the same not-found error the rank paths
// always produced.
func writeRank(vaultDir, id string, rank *int) error {
	if err := checkID(id); err != nil {
		return err
	}
	_, err := issuefiles.Open(vaultDir).Mutate(id, func(i issue.Issue) (issue.Issue, error) {
		i.Frontmatter.Rank = rank
		return i, nil
	})
	return vaultTarget{dir: vaultDir}.mapNotFound(err, id)
}

const prioritizeLong = `prioritize opens $EDITOR on a buffer of the vault's open and in_progress
issues, one line each:

  [P] pkm-055  <title>   — prioritized (in the queue)
  [ ] pkm-5qa8  <title>   — Backlog (not prioritized)

Reorder the [P] lines to change their order; toggle [ ]↔[P] to move an
issue between the queue and the Backlog. Save and close to apply: ranks
are renormalized 1..N and only the files whose rank changed are rewritten.
Ranked done/custom-status issues are not in the buffer and keep their ranks
right after the queue (N+1..M), so the vault never holds a duplicate rank.

An invalid buffer (unknown or duplicated issue ID) is rejected and
nothing is applied.`
