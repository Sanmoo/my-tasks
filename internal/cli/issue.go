// Package cli — Issue commands: create, q, show and edit. These own
// process concerns (files, the editor, stdio); the Issue schema itself
// lives in internal/issue.
package cli

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/issue"
	"github.com/Sanmoo/my-tasks2/internal/priority"
	"github.com/Sanmoo/my-tasks2/internal/show"
	"github.com/Sanmoo/my-tasks2/internal/vault"
)

// newCreateCmd builds `mt create <título>`: writes a new Issue file with
// the spec schema. The title is the remaining positional args joined
// with spaces, so it needs no shell quoting.
func newCreateCmd() *cobra.Command {
	var labels []string
	var top, bottom bool
	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new Issue",
		Long: `create writes a new Issue file (issues/<id>.md) with the spec schema:
title, status (open), labels and created_at in the frontmatter, and the
empty Description/Notes/Comments body. The ID is the vault prefix plus a
short random suffix; created_at is stamped automatically.

With --top or --bottom the Issue is created already in the priority
queue — at position 1, or at the end — instead of the Backlog, with the
same order semantics as ` + "`mt top`/`mt bottom`" + `; the output then reports
its rank.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := checkPlacementFlags(top, bottom); err != nil {
				return err
			}
			if len(args) < 1 {
				return exitcode.Usage(fmt.Errorf("create needs a title"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, strings.Join(args, " "), labels, false, placementAction(top, bottom))
		},
	}
	cmd.Flags().StringArrayVar(&labels, "label", nil, "label; repeatable (free-form)")
	addPlacementFlags(cmd, &top, &bottom)
	return cmd
}

// newQCmd builds `mt q <título>`: like create, but prints only the ID —
// for capturing ideas without leaving the flow. It accepts the same
// --top/--bottom placement flags, still printing just the ID.
func newQCmd() *cobra.Command {
	var top, bottom bool
	cmd := &cobra.Command{
		Use:   "q <title>",
		Short: "Create an Issue and print only its ID",
		Long:  "q is the quiet create: it writes the same Issue file as create, but prints only the new ID. Like create, it accepts --top/--bottom to place the Issue in the queue; the output stays just the ID.",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := checkPlacementFlags(top, bottom); err != nil {
				return err
			}
			if len(args) < 1 {
				return exitcode.Usage(fmt.Errorf("q needs a title"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, strings.Join(args, " "), nil, true, placementAction(top, bottom))
		},
	}
	addPlacementFlags(cmd, &top, &bottom)
	return cmd
}

// addPlacementFlags registers the queue-placement flags shared by create
// and q: -t/--top puts the new Issue at position 1, -b/--bottom at the
// end. The pair is mutually exclusive (checkPlacementFlags).
func addPlacementFlags(cmd *cobra.Command, top, bottom *bool) {
	cmd.Flags().BoolVarP(top, "top", "t", false, "create the Issue at the top of the queue (rank 1)")
	cmd.Flags().BoolVarP(bottom, "bottom", "b", false, "create the Issue at the end of the queue")
}

// checkPlacementFlags rejects --top and --bottom together: the two
// placements contradict each other, so the invocation is malformed
// (a usage error under the exit-code convention).
func checkPlacementFlags(top, bottom bool) error {
	if top && bottom {
		return exitcode.Usage(fmt.Errorf("--top and --bottom are mutually exclusive"))
	}
	return nil
}

// placementAction maps the --top/--bottom flags onto the quick-order
// action the new Issue is planned with. nil keeps the default: the
// Backlog.
func placementAction(top, bottom bool) *priority.QuickAction {
	switch {
	case top:
		a := priority.MoveTop
		return &a
	case bottom:
		a := priority.MoveBottom
		return &a
	default:
		return nil
	}
}

// runCreate writes a new Issue with title and labels and prints its ID
// (just the ID when quiet, a confirmation line otherwise). With a
// placement action the Issue is created already in the queue: the
// quick-order plan — the same one behind `mt top`/`mt bottom` — computes
// its rank and the shifts of the Issues it displaces.
func runCreate(cmd *cobra.Command, title string, labels []string, quiet bool, placement *priority.QuickAction) error {
	vaultDir, err := resolveVault(cmd)
	if err != nil {
		return err
	}
	vcfg, err := vault.LoadVault(vaultDir)
	if err != nil {
		return err
	}
	if vcfg.Prefix == "" {
		return fmt.Errorf("vault %s has no ID prefix in its config — set prefix in mt.yaml", vaultDir)
	}
	id, err := newIssueID(vcfg.Prefix, vaultDir)
	if err != nil {
		return err
	}
	i := issue.Issue{
		Frontmatter: issue.Frontmatter{
			Title:     title,
			Status:    "open",
			Labels:    labels,
			CreatedAt: time.Now().Format(issue.NaiveLayout),
		},
		Body: issue.DefaultBody,
	}
	if placement != nil {
		rank, others, err := planCreatePlacement(vaultDir, priorityIssueFrom(id, i), *placement)
		if err != nil {
			return err
		}
		// Rewrite the displaced Issues before the new file lands, so the
		// rank the new Issue takes (e.g. 1) is never duplicated on disk
		// by the Issue it displaces.
		if err := applyRankChanges(vaultDir, others); err != nil {
			return err
		}
		i.Frontmatter.Rank = &rank
	}
	data, err := issue.Render(i)
	if err != nil {
		return err
	}
	if err := os.WriteFile(issuePath(vaultDir, id), data, 0o644); err != nil {
		return fmt.Errorf("writing issue %s: %w", id, err)
	}
	switch {
	case quiet:
		fmt.Fprintln(cmd.OutOrStdout(), id)
	case placement != nil:
		fmt.Fprintf(cmd.OutOrStdout(), "Created %s (rank %d)\n", id, *i.Frontmatter.Rank)
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", id)
	}
	return nil
}

// planCreatePlacement computes where a new Issue enters the queue when
// created with --top/--bottom: it reuses the quick-order plan (the same
// single source of truth as `mt top`/`mt bottom`) by planning the new
// Issue as if it already existed unranked next to the vault's Issues.
// It returns the new Issue's rank and the rank changes of the Issues it
// displaces; the new Issue's own file is written by the caller.
func planCreatePlacement(vaultDir string, newIssue priority.Issue, action priority.QuickAction) (int, []priority.Change, error) {
	issues, err := loadPriorityIssues(vaultDir)
	if err != nil {
		return 0, nil, err
	}
	changes, err := priority.QuickPlan(append(issues, newIssue), newIssue.ID, action, 0)
	if err != nil {
		return 0, nil, err
	}
	rank, others := 0, make([]priority.Change, 0, len(changes))
	found := false
	for _, ch := range changes {
		if ch.ID == newIssue.ID && ch.Rank != nil {
			rank, found = *ch.Rank, true
			continue
		}
		others = append(others, ch)
	}
	if !found {
		return 0, nil, fmt.Errorf("planning the queue placement of %s: no rank was planned", newIssue.ID)
	}
	return rank, others, nil
}

// newShowCmd builds `mt show <id>`: renders the structured, colored
// Issue view (header, metadata, Markdown body) in the style of nd show.
// Colors follow the standard convention (NO_COLOR, CLICOLOR, TTY); a
// piped stdout renders the same view without ANSI codes.
func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show an Issue (rendered view)",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return exitcode.Usage(fmt.Errorf("show needs exactly one issue ID"))
			}
			return nil
		},
		ValidArgsFunction: completeIssueID,
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := resolveVaultForKey(cmd, args[0])
			if err != nil {
				return err
			}
			if err := checkID(args[0]); err != nil {
				return err
			}
			data, err := os.ReadFile(issuePath(t.dir, args[0]))
			if err != nil {
				if os.IsNotExist(err) {
					return t.notFoundError(args[0])
				}
				return fmt.Errorf("reading issue %s: %w", args[0], err)
			}
			i, err := issue.Parse(data)
			if err != nil {
				return fmt.Errorf("parsing issue %s: %w", args[0], err)
			}
			// TTY detection reads the real stdout, not the injected
			// writer: the decision is about the terminal the process
			// is attached to.
			_, err = fmt.Fprint(cmd.OutOrStdout(), show.Render(i, args[0], show.Options{
				Color: show.ShouldUseColor(term.IsTerminal(int(os.Stdout.Fd()))),
				Width: termWidth(),
			}))
			return err
		},
	}
}

// newEditCmd builds `mt edit <id>`: opens the Issue file in $EDITOR. The
// file is edited in place, so anything untouched survives untouched.
func newEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <id>",
		Short: "Open an Issue in $EDITOR",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return exitcode.Usage(fmt.Errorf("edit needs exactly one issue ID"))
			}
			return nil
		},
		ValidArgsFunction: completeIssueID,
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := resolveVaultForKey(cmd, args[0])
			if err != nil {
				return err
			}
			if err := checkID(args[0]); err != nil {
				return err
			}
			path := issuePath(t.dir, args[0])
			if _, err := os.Stat(path); err != nil {
				if os.IsNotExist(err) {
					return t.notFoundError(args[0])
				}
				return fmt.Errorf("checking issue %s: %w", args[0], err)
			}
			return editFile(path)
		},
	}
}

// termWidth returns the terminal width in columns, 0 when unknown.
func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 0
	}
	return w
}

// issuePath returns the Issue file path for an ID inside a vault.
func issuePath(vaultDir, id string) string {
	return filepath.Join(vaultDir, "issues", id+".md")
}

// checkID guards against an ID that would escape the issues directory.
// A real issue ID is a single file name component; anything with a path
// separator cannot name an issue — the invocation is malformed, so the
// error is a usage error (exit 2) under the exit-code convention.
func checkID(id string) error {
	if id == "" || strings.ContainsAny(id, `/\\`) {
		return exitcode.Usage(fmt.Errorf("invalid issue ID %q", id))
	}
	return nil
}

// newIssueID allocates an ID for a new Issue: the vault prefix plus a
// random suffix that does not collide with any existing issue file.
func newIssueID(prefix, vaultDir string) (string, error) {
	ids, err := issueIDs(vaultDir)
	if err != nil {
		return "", err
	}
	taken := make(map[string]bool, len(ids))
	for _, id := range ids {
		taken[id] = true
	}
	return issue.NextID(prefix, taken, rand.Reader)
}

// editFile opens path in the user's $EDITOR and waits for it to finish.
// The editor value is split on whitespace, so editors with arguments
// (e.g. "code --wait") work; the path is appended as a separate argv
// entry and is never interpreted by a shell.
func editFile(path string) error {
	editor := os.Getenv("EDITOR")
	args := strings.Fields(editor)
	if len(args) == 0 {
		return fmt.Errorf("no $EDITOR set — set EDITOR to your editor to use this command")
	}
	args = append(args, path)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}
	return nil
}
