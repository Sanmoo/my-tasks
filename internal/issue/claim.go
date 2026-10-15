package issue

// The claim rules of an Issue. Claim is data — the claimed_by field —
// not a status and not a transition: it records who assumed the Issue
// before any work started. Availability (internal/list) treats a set
// claim as unavailable, and the lifecycle is: reopen clears the claim,
// done preserves it.

// Claim returns i claimed by who: the claimed_by field set, every other
// field untouched — claim is not a transition, so the Status stays
// exactly as it is. A claim already set is overwritten: possession can
// be transferred explicitly.
func (i Issue) Claim(who string) Issue {
	i.Frontmatter.ClaimedBy = who
	return i
}

// Unclaim returns i with its claimed_by cleared, the pair of Claim.
// Only ClaimedBy is touched — Status and every other field stay exactly
// as they are. On a never-claimed Issue the edit is a noop.
func (i Issue) Unclaim() Issue {
	i.Frontmatter.ClaimedBy = ""
	return i
}
