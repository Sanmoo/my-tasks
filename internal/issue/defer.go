package issue

// Defer returns i with DeferredUntil set to until. Status and every
// other field are untouched: deferral is data — a datetime — not a
// state. The Issue stays open and simply becomes unavailable until the
// time arrives; there is no "undefer" operation, availability is
// computed from the clock (see internal/list.IsFutureDeferred).
func (i Issue) Defer(until string) Issue {
	i.Frontmatter.DeferredUntil = until
	return i
}
