// Package cli — the mt prioritize command. It owns the process concerns
// of the $EDITOR flow (the temp buffer file, running the editor, reading
// the issue files, applying the rank changes in-process); the buffer
// format, parsing and the rank renormalization plan live in
// internal/priority.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/issue"
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
	for _, ch := range changes {
		if err := applyRankChange(vaultDir, ch); err != nil {
			return err
		}
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

// loadPriorityIssues reads every *.md file in the vault's issues/
// directory into the minimal priority.Issue view. A malformed file fails
// the whole command with the offending ID named.
func loadPriorityIssues(vaultDir string) ([]priority.Issue, error) {
	dir := filepath.Join(vaultDir, "issues")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading issues directory: %w", err)
	}
	issues := make([]priority.Issue, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}
		id := strings.TrimSuffix(name, ".md")
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("reading issue %s: %w", id, err)
		}
		i, err := issue.Parse(data)
		if err != nil {
			return nil, fmt.Errorf("parsing issue %s: %w", id, err)
		}
		issues = append(issues, priority.Issue{
			ID:        id,
			Title:     i.Frontmatter.Title,
			Status:    i.Frontmatter.Status,
			Rank:      i.Frontmatter.Rank,
			CreatedAt: i.Frontmatter.CreatedAt,
		})
	}
	return issues, nil
}

// applyRankChange writes the new rank (nil = Backlog) into the issue
// file, preserving everything else. It is in-process: no subprocess is
// spawned per issue.
func applyRankChange(vaultDir string, ch priority.Change) error {
	if err := checkID(ch.ID); err != nil {
		return err
	}
	data, err := os.ReadFile(issuePath(vaultDir, ch.ID))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("issue %s not found", ch.ID)
		}
		return fmt.Errorf("reading issue %s: %w", ch.ID, err)
	}
	i, err := issue.Parse(data)
	if err != nil {
		return fmt.Errorf("parsing issue %s: %w", ch.ID, err)
	}
	i.Frontmatter.Rank = ch.Rank
	out, err := issue.Render(i)
	if err != nil {
		return err
	}
	if err := os.WriteFile(issuePath(vaultDir, ch.ID), out, 0o644); err != nil {
		return fmt.Errorf("writing issue %s: %w", ch.ID, err)
	}
	return nil
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
