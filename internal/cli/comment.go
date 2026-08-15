// Package cli — the comment command. It owns process concerns (files,
// randomness, stdio); the append-only comment logic lives in
// internal/issue.
package cli

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// newCommentCmd builds `mt comment <id> <text>`: appends a comment to the
// Issue's Comments section — a ### timestamp heading, the text, and a
// stable <!-- comment: … --> anchor. The existing body is preserved
// byte-for-byte.
func newCommentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "comment <id> <text>",
		Short: "Append a comment to an Issue",
		Long: `comment appends a comment to the Issue's Comments section: a ###
timestamp heading, the text, and a stable <!-- comment: … --> anchor.
The existing body is preserved byte-for-byte (append-only).`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return exitcode.Usage(fmt.Errorf("comment needs an issue ID and a comment text"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			vaultDir, err := resolveVault(cmd)
			if err != nil {
				return err
			}
			id := args[0]
			if err := checkID(id); err != nil {
				return err
			}
			text := strings.Join(args[1:], " ")
			return appendComment(vaultDir, id, text)
		},
	}
}

// appendComment loads the Issue for id, appends a timestamped comment with
// a fresh stable anchor and writes it back.
func appendComment(vaultDir, id, text string) error {
	i, err := readIssue(vaultDir, id)
	if err != nil {
		return err
	}
	anchor, err := issue.NewAnchor(rand.Reader)
	if err != nil {
		return err
	}
	i.Body = issue.AppendComment(i.Body, time.Now().Format(issue.NaiveLayout), text, anchor)
	return writeIssueFile(vaultDir, id, i)
}
