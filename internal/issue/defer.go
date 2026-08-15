package issue

// Defer returns i deferred until until. A deferred Issue is always open:
// deferral is data — a datetime — not a separate status. Every field
// other than Status and DeferredUntil is untouched; availability is
// computed from the clock (see internal/list.IsFutureDeferred), so there
// is no "undefer" operation.
func (i Issue) Defer(until string) Issue {
	i.Frontmatter.Status = "open"
	i.Frontmatter.DeferredUntil = until
	return i
}
