package issue

// The status-transition rules of an Issue. They encode the only special
// behaviors of the spec: done is terminal (stamps completed_at) and
// reopen clears the work timestamps. Everything else — the free
// transition of `mt status` — is a bare field write with no timestamps
// and no state machine.

// Done returns i closed: status "done" and completed_at stamped with
// now. started_at is preserved — done records completion, not a fresh
// start.
func (i Issue) Done(now string) Issue {
	i.Frontmatter.Status = "done"
	i.Frontmatter.CompletedAt = now
	return i
}

// Reopen returns i reopened: status "open" with completed_at and
// started_at cleared. Every other field is untouched.
func (i Issue) Reopen() Issue {
	i.Frontmatter.Status = "open"
	i.Frontmatter.CompletedAt = ""
	i.Frontmatter.StartedAt = ""
	return i
}

// SetStatus returns i with its status set to s. It is the free
// transition of `mt status`: no timestamps are touched.
func (i Issue) SetStatus(s string) Issue {
	i.Frontmatter.Status = s
	return i
}
