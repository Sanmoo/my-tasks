package issue

import "slices"

// AddBlocker returns i with blocker recorded in blocked_by, unless it is
// already listed. The edit is idempotent: a blocker that already blocks
// i stays listed exactly once, and every other field is untouched.
// Blocked state is computed from the blockers' statuses (see
// internal/list.Blocked), so this only records the reference — there is
// no "block" operation.
func (i Issue) AddBlocker(blocker string) Issue {
	if slices.Contains(i.Frontmatter.BlockedBy, blocker) {
		return i
	}
	i.Frontmatter.BlockedBy = append(i.Frontmatter.BlockedBy, blocker)
	return i
}

// RemoveBlocker returns i with blocker removed from blocked_by. The edit
// is idempotent: a blocker that does not block i leaves i untouched,
// and every other field is untouched. When the last blocker is removed,
// blocked_by becomes empty and Render omits it.
func (i Issue) RemoveBlocker(blocker string) Issue {
	blockedBy := i.Frontmatter.BlockedBy
	out := make([]string, 0, len(blockedBy))
	for _, id := range blockedBy {
		if id != blocker {
			out = append(out, id)
		}
	}
	i.Frontmatter.BlockedBy = out
	return i
}
