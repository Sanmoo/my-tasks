// Command mt is a personal, git-friendly issue tracker: one Markdown
// file per Issue, one Vault per domain.
package main

import (
	"os"

	"github.com/Sanmoo/my-tasks2/internal/cli"
)

func main() {
	// cli.Execute maps outcomes onto the project exit code convention:
	// 0 success, 1 user error, 2 usage error.
	os.Exit(cli.Execute())
}
