// Package priority holds the pure logic of queue ordering: building the
// `mt prioritize` $EDITOR buffer, parsing it back, and planning rank changes
// (renormalization 1..N with minimal rewrite) for both editor and quick
// commands. It is decision-dense, so
// it lives at Seam 2: black-box unit tested, with the coverage and
// mutation gates. Reading and writing the issue files themselves lives
// in internal/issuefiles.
//
// Every mechanism here speaks the domain's single Issue line (issue.Item)
// and orders it by the domain's single Fila rule (issue.Compare): this
// package decides *when* to ask and *which* positions move, never what the
// order is. Its only repository dependency is internal/issue — the
// deliberate descent recorded in docs/adr/0009.
package priority

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// Prioritizable reports whether an issue of status s participates in the
// prioritize buffer: open and in_progress do; done and any custom status
// do not.
func Prioritizable(status string) bool {
	return status == "open" || status == "in_progress"
}

// bufferHeader is the instruction block at the top of the editor buffer.
const bufferHeader = `# Edit ranking for this vault
# Reorder lines. Use [P] for prioritized, [ ] for backlog.
# Do not edit issue IDs. Save and close to continue.
`

// Buffer builds the $EDITOR buffer for items: the instruction header, a
// blank line, then one line per item in buffer order — ranked items
// first (lowest rank first), then the Backlog ordered by created_at, then
// ID as the final tiebreak. That is the domain's Fila order
// (issue.Compare): the buffer renders it, it does not own it. A ranked
// issue is written "[P] <id>  <title>" and a Backlog issue
// "[ ] <id>  <title>". It does not mutate items.
func Buffer(items []issue.Item) string {
	ordered := slices.Clone(items)
	slices.SortFunc(ordered, issue.Compare)
	var b strings.Builder
	b.WriteString(bufferHeader)
	b.WriteByte('\n')
	for _, it := range ordered {
		marker := "[ ]"
		if it.Issue.Frontmatter.Rank != nil {
			marker = "[P]"
		}
		fmt.Fprintf(&b, "%s %s  %s\n", marker, it.ID, it.Issue.Frontmatter.Title)
	}
	return b.String()
}

// Entry is one data line of the buffer: whether the line is prioritized
// and the issue ID it names.
type Entry struct {
	Prioritized bool
	ID          string
}

// Parse turns the saved buffer text back into ordered entries, preserving
// line order. Comment lines (a leading #, possibly indented) and blank
// lines are skipped; every other line must be a data line of the form
// "[P] <id>  <title>" or "[ ] <id>  <title>". The title is not validated
// and is ignored — the ID alone drives the plan. Parse rejects malformed
// lines and duplicate IDs.
func Parse(text string) ([]Entry, error) {
	entries := make([]Entry, 0)
	seen := make(map[string]bool)
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSuffix(raw, "\r") // tolerate CRLF from editors
		if isBlank(line) || isComment(line) {
			continue
		}
		e, err := parseLine(line)
		if err != nil {
			return nil, err
		}
		if seen[e.ID] {
			return nil, fmt.Errorf("duplicate issue ID %q in the buffer", e.ID)
		}
		seen[e.ID] = true
		entries = append(entries, e)
	}
	return entries, nil
}

// isBlank reports whether line is empty or only whitespace.
func isBlank(line string) bool { return strings.TrimSpace(line) == "" }

// isComment reports whether line is a comment: a leading #, allowing
// leading whitespace so an indented comment is still a comment.
func isComment(line string) bool {
	return strings.HasPrefix(strings.TrimLeft(line, " \t"), "#")
}

// parseLine parses one data line into an Entry. The line must start with
// the "[P] " or "[ ] " marker, followed by the issue ID as the first
// whitespace-delimited token (the rest, the title, is ignored).
func parseLine(line string) (Entry, error) {
	var prioritized bool
	var rest string
	switch {
	case strings.HasPrefix(line, "[P] "):
		prioritized = true
		rest = line[len("[P] "):]
	case strings.HasPrefix(line, "[ ] "):
		prioritized = false
		rest = line[len("[ ] "):]
	default:
		return Entry{}, fmt.Errorf("invalid line %q: expected \"[P] <id>\" or \"[ ] <id>\"", line)
	}
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return Entry{}, fmt.Errorf("invalid line %q: missing issue ID", line)
	}
	return Entry{Prioritized: prioritized, ID: fields[0]}, nil
}

// Plan validates entries against the vault's Items and computes the rank
// changes to apply. [P] entries get ranks 1..N in buffer order; [ ]
// entries go to the Backlog. Ranked issues that are not prioritizable
// (done or a custom status) are not in the buffer, but Rank is a
// vault-wide invariant, so they keep their ranks renumbered after the
// queue (N+1..M) in their existing rank order. It returns an issue.Change
// only for issues whose rank actually differs from their current rank —
// unchanged issues yield no change, so the caller rewrites only what
// moved (zero churn).
//
// Plan fails — returning no plan — when any entry names an unknown ID,
// a non-prioritizable issue (done or a custom status), a duplicate ID, or
// when a prioritizable issue is missing from the buffer. Nothing is
// applied on failure.
func Plan(entries []Entry, items []issue.Item) ([]issue.Change, error) {
	byID := make(map[string]issue.Item, len(items))
	for _, it := range items {
		byID[it.ID] = it
	}
	seen := make(map[string]bool, len(entries))
	for _, e := range entries {
		if seen[e.ID] {
			return nil, fmt.Errorf("duplicate issue ID %q in the buffer", e.ID)
		}
		seen[e.ID] = true
		it, ok := byID[e.ID]
		if !ok {
			return nil, fmt.Errorf("unknown issue ID %q", e.ID)
		}
		if status := it.Issue.Frontmatter.Status; !Prioritizable(status) {
			return nil, fmt.Errorf("issue %s is %s and cannot be prioritized", e.ID, status)
		}
	}
	for _, it := range items {
		if Prioritizable(it.Issue.Frontmatter.Status) && !seen[it.ID] {
			return nil, fmt.Errorf("issue %s is missing from the buffer", it.ID)
		}
	}
	changes := make([]issue.Change, 0, len(entries))
	rank := 0
	for _, e := range entries {
		var target *int
		if e.Prioritized {
			rank++
			r := rank
			target = &r
		}
		if !rankEqual(byID[e.ID].Issue.Frontmatter.Rank, target) {
			changes = append(changes, issue.Change{ID: e.ID, Rank: target})
		}
	}
	// Ranked issues outside the buffer (done, custom statuses) follow the
	// queue at N+1..M in their existing rank order, so the vault keeps
	// unique contiguous ranks — a duplicate would make mt check fail and
	// mt pick-next refuse to select.
	rest := make([]issue.Item, 0)
	for _, it := range items {
		if it.Issue.Frontmatter.Rank != nil && !Prioritizable(it.Issue.Frontmatter.Status) {
			rest = append(rest, it)
		}
	}
	slices.SortFunc(rest, issue.Compare)
	for _, it := range rest {
		rank++
		r := rank
		if !rankEqual(it.Issue.Frontmatter.Rank, &r) {
			changes = append(changes, issue.Change{ID: it.ID, Rank: &r})
		}
	}
	return changes, nil
}

// rankEqual reports whether two ranks are the same value, treating nil
// (Backlog) as a value equal only to nil.
func rankEqual(a, b *int) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
