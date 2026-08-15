// Package exitcode defines the process exit code convention of the mt CLI:
//
//	0  success
//	1  user error — the invocation was fine, but the requested action could
//	   not be completed (vault undefined, nothing available, duplicate rank,
//	   invalid edit)
//	2  usage error — mt was invoked wrong (unknown command, unknown flag,
//	   wrong arguments)
//
// and maps typed errors onto that convention.
package exitcode

import "errors"

const (
	// Success: the command did what was asked.
	Success = 0
	// UserError: the requested action could not be completed.
	UserError = 1
	// UsageError: mt was invoked wrong.
	UsageError = 2
)

// For returns the exit code for err under the convention: nil → Success,
// a UsageProblem → UsageError, anything else → UserError.
func For(err error) int {
	if err == nil {
		return Success
	}
	var usage *UsageProblem
	if errors.As(err, &usage) {
		return UsageError
	}
	return UserError
}

// UsageProblem marks an error as a usage error (exit code 2).
type UsageProblem struct {
	problem
}

// UserProblem marks an error as a user error (exit code 1).
type UserProblem struct {
	problem
}

// problem carries the message and unwrapping shared by both kinds.
type problem struct {
	err error
}

func (p problem) Error() string { return p.err.Error() }
func (p problem) Unwrap() error { return p.err }

// Usage marks err as a usage error. It returns nil when err is nil.
func Usage(err error) error {
	if err == nil {
		return nil
	}
	return &UsageProblem{problem{err}}
}

// User marks err as a user error. It returns nil when err is nil.
func User(err error) error {
	if err == nil {
		return nil
	}
	return &UserProblem{problem{err}}
}
