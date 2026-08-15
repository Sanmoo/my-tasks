// Package cli — the mt defer command. It owns the process concerns of
// deferral (resolving the vault, reading/writing the Issue file, stdio);
// the time-argument parsing lives in internal/deferral and the field
// write in internal/issue.
package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/deferral"
	"github.com/Sanmoo/my-tasks2/internal/exitcode"
)

// newDeferCmd builds `mt defer <id> <when>`: sets deferred_until on the
// Issue, keeping its status unchanged. The Issue simply becomes
// unavailable until the time arrives — there is no "undefer".
func newDeferCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "defer <id> <when>",
		Short: "Defer an Issue until a datetime",
		Long:  deferLong,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return exitcode.Usage(fmt.Errorf("defer needs an issue ID and a time (YY-MM-DD HH:MM or +2d/+1w/+3h)"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// The time may be spelled "26-08-20 08:00" as one quoted
			// argument or two; join the remainder so both forms parse.
			return runDefer(cmd, args[0], strings.Join(args[1:], " "))
		},
	}
}

// runDefer resolves the vault, parses the time argument into the
// canonical deferred_until value, and writes it onto the Issue.
func runDefer(cmd *cobra.Command, id, when string) error {
	vaultDir, err := resolveVault(cmd)
	if err != nil {
		return err
	}
	if err := checkID(id); err != nil {
		return err
	}
	until, err := deferral.Parse(when, time.Now())
	if err != nil {
		return err
	}
	i, err := readIssue(vaultDir, id)
	if err != nil {
		return err
	}
	if err := writeIssueFile(vaultDir, id, i.Defer(until)); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s deferred until %s\n", id, until)
	return nil
}

const deferLong = `defer sets an Issue's deferred_until, keeping its status: the
Issue stays open and simply becomes unavailable until the moment
arrives (now >= deferred_until), then reappears on its own — there is
no "undefer".

The time is either an absolute local datetime in YY-MM-DD HH:MM form
(e.g. 26-08-20 08:00 — the hour is kept) or a relative duration from
now (+2d, +1w, +3h).`
