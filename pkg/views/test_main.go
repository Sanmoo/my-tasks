package views

import (
	"os"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMain(m *testing.M) {
	v := m.Run()

	// After all tests have run, clean up obsolete snapshots
	if _, err := snaps.Clean(m); err != nil {
		panic(err)
	}

	os.Exit(v)
}
