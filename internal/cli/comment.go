// Package cli — the comment feature: the `mt comment` command and the
// comment a status transition may carry. It owns process concerns (files,
// randomness, stdio); the append-only comment logic lives in
// internal/issue.
package cli

import (
	"crypto/rand"
	"errors"
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
The existing body is preserved byte-for-byte (append-only). The status
transitions accept the same text as a trailing argument and write it the
same way.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return exitcode.Usage(errors.New("comment needs an issue ID and a comment text"))
			}
			return nil
		},
		ValidArgsFunction: completeIssueID,
		RunE: func(cmd *cobra.Command, args []string) error {
			text, err := commentText(args[1:])
			if err != nil {
				return err
			}
			t, err := resolveVaultForKey(cmd, args[0])
			if err != nil {
				return err
			}
			id := args[0]
			if err := checkID(id); err != nil {
				return err
			}
			return appendComment(t, id, text)
		},
	}
}

// commentText joins the trailing comment arguments into one comment text,
// one space between arguments — the rule `mt comment` and the status
// transitions share. No arguments means no comment at all (""). A text that
// is present but blank is a usage error: an empty comment would write a
// heading with no body into the append-only Comments section for good.
func commentText(args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}
	text := strings.Join(args, " ")
	if strings.TrimSpace(text) == "" {
		return "", exitcode.Usage(errors.New("comment text is blank"))
	}
	return text, nil
}

// appendComment loads the Issue for id, appends a timestamped comment with
// a fresh stable anchor and writes it back through the single mutation path.
func appendComment(t vaultTarget, id, text string) error {
	_, err := mutateIssue(t, id, func(i issue.Issue) (issue.Issue, error) {
		return addComment(i, text, time.Now().Format(issue.NaiveLayout))
	})
	return err
}

// addComment returns i with text appended as a new Comment: a heading
// carrying now, the text, and a freshly allocated stable anchor. It is the
// body half shared by `mt comment` and by the comment a status transition
// may carry — the caller owns the write. Comments are append-only, so the
// existing body is extended, never rewritten.
func addComment(i issue.Issue, text, now string) (issue.Issue, error) {
	anchor, err := issue.NextAnchor(rand.Reader, i.Body)
	if err != nil {
		return issue.Issue{}, fmt.Errorf("allocating comment anchor: %w", err)
	}
	i.Body = issue.AppendComment(i.Body, now, text, anchor)
	return i, nil
}
