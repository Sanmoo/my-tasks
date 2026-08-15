// Package exitcode_test holds the black-box tests of the exit code
// convention (Seam 2: exported API of a pure-logic package).
package exitcode_test

import (
	"errors"
	"testing"

	"github.com/Sanmoo/my-tasks2/internal/exitcode"
)

func TestConventionValues(t *testing.T) {
	tests := []struct {
		name string
		got  int
		want int
	}{
		{"success", exitcode.Success, 0},
		{"user error", exitcode.UserError, 1},
		{"usage error", exitcode.UsageError, 2},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %d, want %d", tt.name, tt.got, tt.want)
		}
	}
}

func TestForNilIsSuccess(t *testing.T) {
	if got := exitcode.For(nil); got != exitcode.Success {
		t.Errorf("For(nil) = %d, want %d", got, exitcode.Success)
	}
}

func TestForPlainErrorIsUserError(t *testing.T) {
	err := errors.New("no issue available")
	if got := exitcode.For(err); got != exitcode.UserError {
		t.Errorf("For(plain error) = %d, want %d", got, exitcode.UserError)
	}
}

func TestForUsageIsUsageError(t *testing.T) {
	err := exitcode.Usage(errors.New("unknown command"))
	if got := exitcode.For(err); got != exitcode.UsageError {
		t.Errorf("For(usage) = %d, want %d", got, exitcode.UsageError)
	}
	if got := err.Error(); got != "unknown command" {
		t.Errorf("message = %q, want %q", got, "unknown command")
	}
}

func TestForWrappedUsageIsUsageError(t *testing.T) {
	base := exitcode.Usage(errors.New("bad flag"))
	err := errors.Join(errors.New("context"), base)
	if got := exitcode.For(err); got != exitcode.UsageError {
		t.Errorf("For(wrapped usage) = %d, want %d", got, exitcode.UsageError)
	}
}

func TestForUserIsUserError(t *testing.T) {
	err := exitcode.User(errors.New("nothing available"))
	if got := exitcode.For(err); got != exitcode.UserError {
		t.Errorf("For(user) = %d, want %d", got, exitcode.UserError)
	}
	if got := err.Error(); got != "nothing available" {
		t.Errorf("message = %q, want %q", got, "nothing available")
	}
}

func TestUsageOfNilIsNil(t *testing.T) {
	if err := exitcode.Usage(nil); err != nil {
		t.Errorf("Usage(nil) = %v, want nil", err)
	}
}

func TestUserOfNilIsNil(t *testing.T) {
	if err := exitcode.User(nil); err != nil {
		t.Errorf("User(nil) = %v, want nil", err)
	}
}

func TestProblemsUnwrap(t *testing.T) {
	base := errors.New("boom")
	for _, err := range []error{exitcode.Usage(base), exitcode.User(base)} {
		if !errors.Is(err, base) {
			t.Errorf("%v does not unwrap to the original error", err)
		}
	}
}
