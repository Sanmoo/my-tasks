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

// Undefer returns i with its deferred_until cleared, archiving the
// reminder. It is the pair of Defer: the Issue is back in normal
// circulation immediately (availability is computed from the clock, see
// internal/list). Only DeferredUntil is touched — Status and Rank stay
// exactly as they are; undefer has no opinion on priority or state.
func (i Issue) Undefer() Issue {
	i.Frontmatter.DeferredUntil = ""
	return i
}
