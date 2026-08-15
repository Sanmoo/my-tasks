package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/vault"
)

// newInitCmd builds `mt init [dir]`: creates a usable Vault at dir
// (default: the current directory) — the issues/ directory plus the
// vault config mt.yaml with the ID prefix and status list.
func newInitCmd() *cobra.Command {
	var prefix string
	var statuses []string
	cmd := &cobra.Command{
		Use:   "init [dir]",
		Short: "Create a new Vault",
		Long: `init creates a usable Vault at dir (default: the current directory):
the issues/ directory plus the vault config mt.yaml with the ID prefix
and status list.

The prefix defaults to a derivation of the directory name (lowercased,
non-alphanumeric stripped); the statuses default to
[open, in_progress, done]. init refuses to overwrite an existing
vault config.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return exitcode.Usage(err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if bookmark != "" {
				return exitcode.Usage(fmt.Errorf("init does not take a @bookmark: it creates a vault, it does not address one"))
			}
			vaultFlag, err := cmd.Flags().GetString("vault")
			if err != nil {
				return err
			}
			if vaultFlag != "" {
				return exitcode.Usage(fmt.Errorf("init does not take --vault: it creates a vault, it does not address one"))
			}
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("locating home directory: %w", err)
			}
			dir = vault.ExpandHome(dir, home)
			// A bare `mt init` targets the current directory: resolve it
			// to a real path so the prefix derives from the cwd name.
			if dir == "." {
				dir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("locating the current directory: %w", err)
				}
			}
			if prefix == "" {
				prefix = vault.PrefixFor(dir)
				if prefix == "" {
					return fmt.Errorf("no ID prefix: could not derive one from %q — pass --prefix", dir)
				}
			}
			if err := (vault.Vault{Prefix: prefix, Status: statuses}).Save(dir); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Vault ready at %s\n", dir)
			return nil
		},
	}
	cmd.Flags().StringVar(&prefix, "prefix", "", "ID prefix for issues (default: derived from the directory name)")
	cmd.Flags().StringArrayVar(&statuses, "status", nil, "status of the vault; repeatable (default: open, in_progress, done)")
	return cmd
}
