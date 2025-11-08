package cli

import (
	"fmt"

	"github.com/Sanmoo/my-tasks/internal/task"
	"github.com/pterm/pterm"
)

func RenderWarningsIfAny(projects []*task.Project) {
	warnings := []string{}
	for _, p := range projects {
		warnings = append(warnings, p.GetWarnings()...)
	}

	if len(warnings) == 0 {
		return
	}

	fmt.Print("\n\n")
	pterm.Warning.Printf("==================================== WARNINGS ====================================\n\n")

	for _, warning := range warnings {
		pterm.Warning.Printf(warning + "\n")
	}

	pterm.Warning.Printf("==================================================================================\n\n\n\n")
	fmt.Print("\n\n")
}
