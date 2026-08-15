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

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/issue"
	"github.com/Sanmoo/my-tasks2/internal/vault"
)

// newCreateCmd builds `mt create <título>`: writes a new Issue file with
// the spec schema. The title is the remaining positional args joined
// with spaces, so it needs no shell quoting.
func newCreateCmd() *cobra.Command {
	var labels []string
	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new Issue",
		Long: `create writes a new Issue file (issues/<id>.md) with the spec schema:
title, status (open), labels and created_at in the frontmatter, and the
empty Description/Notes/Comments body. The ID is the vault prefix plus a
short random suffix; created_at is stamped automatically.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return exitcode.Usage(fmt.Errorf("create needs a title"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, strings.Join(args, " "), labels, false)
		},
	}
	cmd.Flags().StringArrayVar(&labels, "label", nil, "label; repeatable (free-form)")
	return cmd
}

// newQCmd builds `mt q <título>`: like create, but prints only the ID —
// for capturing ideas without leaving the flow.
func newQCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "q <title>",
		Short: "Create an Issue and print only its ID",
		Long:  "q is the quiet create: it writes the same Issue file as create, but prints only the new ID.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return exitcode.Usage(fmt.Errorf("q needs a title"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, strings.Join(args, " "), nil, true)
		},
	}
}

// runCreate writes a new Issue with title and labels and prints its ID
// (just the ID when quiet, a confirmation line otherwise).
func runCreate(cmd *cobra.Command, title string, labels []string, quiet bool) error {
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
	data, err := issue.Render(i)
	if err != nil {
		return err
	}
	if err := os.WriteFile(issuePath(vaultDir, id), data, 0o644); err != nil {
		return fmt.Errorf("writing issue %s: %w", id, err)
	}
	if quiet {
		fmt.Fprintln(cmd.OutOrStdout(), id)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", id)
	}
	return nil
}

// newShowCmd builds `mt show <id>`: prints the Issue file — frontmatter
// and body — exactly as stored.
func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show an Issue (frontmatter + body)",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return exitcode.Usage(fmt.Errorf("show needs exactly one issue ID"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			vaultDir, err := resolveVault(cmd)
			if err != nil {
				return err
			}
			if err := checkID(args[0]); err != nil {
				return err
			}
			data, err := os.ReadFile(issuePath(vaultDir, args[0]))
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("issue %s not found", args[0])
				}
				return fmt.Errorf("reading issue %s: %w", args[0], err)
			}
			_, err = cmd.OutOrStdout().Write(data)
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
		RunE: func(cmd *cobra.Command, args []string) error {
			vaultDir, err := resolveVault(cmd)
			if err != nil {
				return err
			}
			if err := checkID(args[0]); err != nil {
				return err
			}
			path := issuePath(vaultDir, args[0])
			if _, err := os.Stat(path); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("issue %s not found", args[0])
				}
				return fmt.Errorf("checking issue %s: %w", args[0], err)
			}
			return editFile(path)
		},
	}
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
	entries, err := os.ReadDir(filepath.Join(vaultDir, "issues"))
	if err != nil {
		return "", fmt.Errorf("reading issues directory: %w", err)
	}
	taken := make(map[string]bool, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".md") {
			taken[strings.TrimSuffix(name, ".md")] = true
		}
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
