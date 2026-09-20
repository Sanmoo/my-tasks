// Package cli — shell completion wiring. One shared ValidArgsFunction
// completes issue IDs on every command that takes a key; `mt completion
// <shell>` (cobra-generated) uses it to drive the shell's <TAB>.
package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

// completeIssueID is the ValidArgsFunction of every command that takes
// an issue key: it lists the vault's issues/*.md IDs (suffix stripped),
// filtered by the case-insensitive prefix of the ID the user typed, and
// forbids the shell's own file completion (ShellCompDirectiveNoFileComp).
// The vault is resolved with the same resolveVaultForKey the command's
// RunE uses — on the partial key in position 0, and on the subject's
// key in the blocker position of dep add/dep rm — so completion and
// execution inherit the identical @bookmark > --vault > key prefix >
// default precedence and its ambiguity/fallback messages. A resolution
// failure (ambiguous prefix, or no vault at all) yields zero candidates:
// TAB stays silent and the command's own error appears on execution,
// never as a silent choice. Positions that are not an issue key — the
// <status> of status, <when>, <n>, <text> — complete nothing.
func completeIssueID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	key := toComplete
	switch len(args) {
	case 0:
		// Position 0: the key itself. The partial prefix already
		// suffices — PrefixOf("dom-ab") = "dom".
	case 1:
		// Position 1 is a key only on dep add/dep rm, where the
		// blocker completes against the subject's vault (args[0]) —
		// never against the blocker's own partial prefix.
		if cmd.Parent() == nil || cmd.Parent().Name() != "dep" {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		key = args[0]
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	t, err := resolveVaultForKey(cmd, key)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	ids, err := issueIDs(t.dir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	prefix := strings.ToLower(toComplete)
	completions := make([]string, 0, len(ids))
	for _, id := range ids {
		// Case-insensitive prefix of the whole ID: DOM-AB and dom-ab
		// behave alike, while resolution stays case-sensitive in the
		// prefix step (it never sees this lowered form).
		if strings.HasPrefix(strings.ToLower(id), prefix) {
			completions = append(completions, id)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}