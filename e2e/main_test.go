// Package e2e runs the Gherkin scenarios of mt against the compiled
// binary (Seam 1): every scenario executes ./mt as a real process.
package e2e

import (
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"

	"github.com/Sanmoo/my-tasks2/e2e/steps"
	"github.com/Sanmoo/my-tasks2/e2e/support"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "mt-e2e-bin-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	bin, err := support.BuildBinary(dir)
	if err != nil {
		os.RemoveAll(dir)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	support.SetBinary(bin)

	status := godog.TestSuite{
		Name:                "mt-e2e",
		ScenarioInitializer: steps.InitializeScenario,
		Options: &godog.Options{
			Format: "pretty",
			Paths:  []string{"features"},
		},
	}.Run()

	// os.Exit skips defers — clean up the built binary dir by hand.
	os.RemoveAll(dir)
	os.Exit(status)
}
