// Package priority holds the pure logic of `mt prioritize`: building the
// $EDITOR buffer, parsing it back, and planning the rank changes
// (renormalization 1..N with minimal rewrite). It is decision-dense, so
// it lives at Seam 2: black-box unit tested, with the coverage and
// mutation gates. Reading and writing the issue files themselves is a
// process concern and stays in internal/cli.
package priority

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

// Issue is the minimal view of an Issue the prioritize flow needs: the ID
// (the file name authority), title, status, rank (nil = Backlog) and
// created_at (the Backlog ordering key).
type Issue struct {
	ID        string
	Title     string
	Status    string
	Rank      *int
	CreatedAt string
}

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

// Buffer builds the $EDITOR buffer for issues: the instruction header, a
// blank line, then one line per issue in buffer order — ranked issues
// first (lowest rank first), then the Backlog ordered by created_at, then
// ID as the final tiebreak. A ranked issue is written "[P] <id>  <title>"
// and a Backlog issue "[ ] <id>  <title>". It does not mutate issues.
func Buffer(issues []Issue) string {
	ordered := slices.Clone(issues)
	slices.SortFunc(ordered, Compare)
	var b strings.Builder
	b.WriteString(bufferHeader)
	b.WriteByte('\n')
	for _, is := range ordered {
		marker := "[ ]"
		if is.Rank != nil {
			marker = "[P]"
		}
		fmt.Fprintf(&b, "%s %s  %s\n", marker, is.ID, is.Title)
	}
	return b.String()
}

// Compare orders two issues for the prioritize buffer: lower rank first
// (issues without a rank form the Backlog and come last, ordered by
// created_at), then ID as the final tiebreak. It returns a negative value
// when a sorts before b, zero when they are equal, and a positive value
// otherwise — the cmp.Compare convention, so it plugs into slices.SortFunc.
func Compare(a, b Issue) int {
	if a.Rank != nil && b.Rank != nil {
		if c := cmp.Compare(*a.Rank, *b.Rank); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	}
	if a.Rank != nil {
		return -1
	}
	if b.Rank != nil {
		return 1
	}
	if c := cmp.Compare(a.CreatedAt, b.CreatedAt); c != 0 {
		return c
	}
	return cmp.Compare(a.ID, b.ID)
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

// Change is a computed target rank for one issue. Rank == nil means the
// issue returns to the Backlog (its rank is removed).
type Change struct {
	ID   string
	Rank *int
}

// Plan validates entries against the vault's issues and computes the rank
// changes to apply. [P] entries get ranks 1..N in buffer order; [ ]
// entries go to the Backlog. It returns a Change only for issues whose
// rank actually differs from their current rank — unchanged issues yield
// no change, so the caller rewrites only what moved (zero churn).
//
// Plan fails — returning no plan — when any entry names an unknown ID,
// a non-prioritizable issue (done or a custom status), a duplicate ID, or
// when a prioritizable issue is missing from the buffer. Nothing is
// applied on failure.
func Plan(entries []Entry, issues []Issue) ([]Change, error) {
	byID := make(map[string]Issue, len(issues))
	for _, is := range issues {
		byID[is.ID] = is
	}
	seen := make(map[string]bool, len(entries))
	for _, e := range entries {
		if seen[e.ID] {
			return nil, fmt.Errorf("duplicate issue ID %q in the buffer", e.ID)
		}
		seen[e.ID] = true
		is, ok := byID[e.ID]
		if !ok {
			return nil, fmt.Errorf("unknown issue ID %q", e.ID)
		}
		if !Prioritizable(is.Status) {
			return nil, fmt.Errorf("issue %s is %s and cannot be prioritized", e.ID, is.Status)
		}
	}
	for _, is := range issues {
		if Prioritizable(is.Status) && !seen[is.ID] {
			return nil, fmt.Errorf("issue %s is missing from the buffer", is.ID)
		}
	}
	changes := make([]Change, 0, len(entries))
	rank := 0
	for _, e := range entries {
		var target *int
		if e.Prioritized {
			rank++
			r := rank
			target = &r
		}
		if !rankEqual(byID[e.ID].Rank, target) {
			changes = append(changes, Change{ID: e.ID, Rank: target})
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
