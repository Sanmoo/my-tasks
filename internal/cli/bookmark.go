package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
	"github.com/Sanmoo/my-tasks2/internal/vault"
)

// newBookmarkCmd builds `mt bookmark`: management of the global config's
// bookmarks and default bookmark. The name arguments are bare (no @) —
// the @name form is how other commands address a vault.
func newBookmarkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bookmark",
		Short: "Manage bookmarks to Vaults",
		Long:  bookmarkLong,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newBookmarkAddCmd(), newBookmarkListCmd(), newBookmarkRmCmd())
	return cmd
}

// globalConfigPath returns the auto-detected global config path (XDG)
// and the user's home directory, both derived from the environment.
func globalConfigPath() (path, home string, err error) {
	home, err = os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("locating home directory: %w", err)
	}
	return vault.GlobalConfigPath(os.Getenv("XDG_CONFIG_HOME"), home), home, nil
}

// loadGlobal reads the auto-detected global config, returning it with its
// path.
func loadGlobal() (vault.Global, string, error) {
	path, _, err := globalConfigPath()
	if err != nil {
		return vault.Global{}, "", err
	}
	g, err := vault.LoadGlobal(path)
	return g, path, err
}

// noBookmarkArg rejects the @name form for bookmark subcommands: their
// positional arguments are bare bookmark names, not vault addresses. The
// @token is stripped globally before cobra parses, so a leftover non-empty
// bookmark means the user wrote @name where a bare name belongs.
func noBookmarkArg() error {
	if bookmark != "" {
		return exitcode.Usage(fmt.Errorf("bookmark subcommands take a bare name, not @%s", bookmark))
	}
	return nil
}

// newBookmarkAddCmd builds `mt bookmark add <name> <path>`.
func newBookmarkAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <path>",
		Short: "Add a bookmark to a Vault",
		Long: `add records the bookmark name → path in the global config. The name is
bare (no @); afterwards any command addresses the vault as @name. The
path is stored as given, so ~/… keeps expanding at resolution time.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := noBookmarkArg(); err != nil {
				return err
			}
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return exitcode.Usage(err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name, path := args[0], args[1]
			g, configPath, err := loadGlobal()
			if err != nil {
				return err
			}
			updated, err := g.AddBookmark(name, path)
			if err != nil {
				return exitcode.Usage(err)
			}
			if err := updated.SaveGlobal(configPath); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Bookmark %s added.\n", name)
			return nil
		},
	}
}

// newBookmarkListCmd builds `mt bookmark list`.
func newBookmarkListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List bookmarks",
		Long: `list prints the bookmarks of the global config, one per line as
"name -> path", sorted by name. The default bookmark is marked with
"(default)".`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := noBookmarkArg(); err != nil {
				return err
			}
			if err := cobra.NoArgs(cmd, args); err != nil {
				return exitcode.Usage(err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			g, _, err := loadGlobal()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, name := range g.Names() {
				line := fmt.Sprintf("%s -> %s", name, g.Bookmarks[name])
				if name == g.Default {
					line += " (default)"
				}
				fmt.Fprintln(out, line)
			}
			return nil
		},
	}
}

// newBookmarkRmCmd builds `mt bookmark rm <name>`.
func newBookmarkRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove a bookmark",
		Long: `rm removes the bookmark name from the global config, leaving the rest
intact. Removing the default bookmark also clears the default.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := noBookmarkArg(); err != nil {
				return err
			}
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return exitcode.Usage(err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			g, configPath, err := loadGlobal()
			if err != nil {
				return err
			}
			updated, err := g.RemoveBookmark(name)
			if err != nil {
				// A missing bookmark is a user error, not a usage error.
				return err
			}
			if err := updated.SaveGlobal(configPath); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Bookmark %s removed.\n", name)
			return nil
		},
	}
}

const bookmarkLong = `bookmark manages the global config: named bookmarks pointing at Vaults
and the default bookmark used when a command names none.

  mt bookmark add <name> <path>   add (or update) a bookmark
  mt bookmark list                list bookmarks, marking the default
  mt bookmark rm <name>           remove a bookmark

Bookmark names are bare (letters, digits, '-' and '_'): add them as
"pkm", then address the vault as @pkm. The default bookmark is set with
the "default:" key of the global config.`
