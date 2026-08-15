// Package check holds the pure validation and Rank-integrity rules of
// `mt check`. Reading and writing Issue files remains a process concern in
// internal/cli; these exported APIs are the Seam 2 unit-test boundary.
package check

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// Item is an Issue with the file name that identifies it inside a Vault.
type Item struct {
	ID    string
	Issue issue.Issue
}

// RankGap is one contiguous missing range in the ranked queue. Start and End
// are inclusive; a single missing Rank has equal Start and End.
type RankGap struct {
	Start int
	End   int
}

// RankGapRanges returns the positive Rank ranges missing between 1 and the
// highest ranked Issue. Backlog Issues (nil Rank) and non-positive Ranks do
// not form gaps; callers can report those separately. The range form keeps a
// very large manually edited Rank from forcing a billion-element allocation.
func RankGapRanges(items []Item) []RankGap {
	ranks := make([]int, 0, len(items))
	for _, item := range items {
		if rank := item.Issue.Frontmatter.Rank; rank != nil && *rank > 0 {
			ranks = append(ranks, *rank)
		}
	}
	slices.Sort(ranks)
	gaps := make([]RankGap, 0)
	next := 1
	maxInt := int(^uint(0) >> 1)
	for _, rank := range ranks {
		if rank < next {
			continue
		}
		if rank > next {
			gaps = append(gaps, RankGap{Start: next, End: rank - 1})
		}
		if rank == maxInt {
			break
		}
		next = rank + 1
	}
	return gaps
}

// DuplicateRanks returns the Rank values that occur more than once,
// sorted ascending. Backlog Issues (nil Rank) do not participate.
func DuplicateRanks(items []Item) []int {
	counts := make(map[int]int)
	for _, item := range items {
		if rank := item.Issue.Frontmatter.Rank; rank != nil {
			counts[*rank]++
		}
	}
	duplicates := make([]int, 0)
	for rank, count := range counts {
		if count > 1 {
			duplicates = append(duplicates, rank)
		}
	}
	slices.Sort(duplicates)
	return duplicates
}

// NonPositiveRanks returns distinct Rank values that cannot belong to the
// 1..N queue, sorted ascending.
func NonPositiveRanks(items []Item) []int {
	invalid := make(map[int]struct{})
	for _, item := range items {
		if rank := item.Issue.Frontmatter.Rank; rank != nil && *rank <= 0 {
			invalid[*rank] = struct{}{}
		}
	}
	values := make([]int, 0, len(invalid))
	for rank := range invalid {
		values = append(values, rank)
	}
	slices.Sort(values)
	return values
}

var allowedFrontmatterFields = map[string]struct{}{
	"title": {}, "status": {}, "labels": {}, "created_at": {}, "rank": {},
	"deferred_until": {}, "deadline": {}, "started_at": {}, "completed_at": {},
}

var optionalFrontmatterFields = map[string]struct{}{
	"rank": {}, "deferred_until": {}, "deadline": {}, "started_at": {}, "completed_at": {},
}

// ValidateFrontmatter validates the YAML mapping and schema-specific keys in
// an Issue's raw Markdown file. It rejects unknown/forbidden fields, empty
// optional fields, and additional YAML documents inside the frontmatter.
func ValidateFrontmatter(data []byte, id string) error {
	payload, err := frontmatterPayload(data)
	if err != nil {
		return fmt.Errorf("malformed frontmatter for issue %s: %w", id, err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(payload))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return fmt.Errorf("malformed frontmatter for issue %s: %w", id, err)
	}
	if len(document.Content) == 0 || document.Content[0].Kind != yaml.MappingNode {
		return fmt.Errorf("malformed frontmatter for issue %s: expected a YAML mapping", id)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("malformed frontmatter for issue %s: multiple YAML documents", id)
		}
		return fmt.Errorf("malformed frontmatter for issue %s: %w", id, err)
	}
	mapping := document.Content[0]
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		name := mapping.Content[i].Value
		value := mapping.Content[i+1]
		if _, ok := allowedFrontmatterFields[name]; !ok {
			return fmt.Errorf("malformed frontmatter for issue %s: unknown field %q", id, name)
		}
		_, optional := optionalFrontmatterFields[name]
		if optional && value.Kind == yaml.ScalarNode &&
			(value.Value == "" || value.Tag == "!!null") {
			return fmt.Errorf("malformed frontmatter for issue %s: empty optional field %q must be omitted", id, name)
		}
		if name == "rank" && value.Tag != "!!int" {
			return fmt.Errorf("malformed frontmatter for issue %s: rank must be an integer", id)
		}
	}
	return nil
}

// ValidateItem validates the parsed frontmatter values of an Issue against
// the Vault's configured statuses and the canonical naive datetime layout.
func ValidateItem(item Item, statuses []string) error {
	fm := item.Issue.Frontmatter
	switch {
	case fm.Title == "":
		return fmt.Errorf("malformed frontmatter for issue %s: missing title", item.ID)
	case fm.Status == "":
		return fmt.Errorf("malformed frontmatter for issue %s: missing status", item.ID)
	case fm.Labels == nil:
		return fmt.Errorf("malformed frontmatter for issue %s: missing labels", item.ID)
	case fm.CreatedAt == "":
		return fmt.Errorf("malformed frontmatter for issue %s: missing created_at", item.ID)
	}
	if !slices.Contains(statuses, fm.Status) {
		return fmt.Errorf("status %q for issue %s is not configured (valid: %s)",
			fm.Status, item.ID, strings.Join(statuses, ", "))
	}
	for _, field := range []struct {
		name     string
		value    string
		required bool
	}{
		{name: "created_at", value: fm.CreatedAt, required: true},
		{name: "deferred_until", value: fm.DeferredUntil},
		{name: "deadline", value: fm.Deadline},
		{name: "started_at", value: fm.StartedAt},
		{name: "completed_at", value: fm.CompletedAt},
	} {
		if field.value == "" {
			if field.required {
				return fmt.Errorf("invalid datetime for issue %s: %s is empty", item.ID, field.name)
			}
			continue
		}
		parsed, err := time.ParseInLocation(issue.NaiveLayout, field.value, time.Local)
		if err != nil || parsed.Format(issue.NaiveLayout) != field.value {
			return fmt.Errorf("invalid datetime for issue %s: %s=%q (want %s)",
				item.ID, field.name, field.value, issue.NaiveLayout)
		}
	}
	return nil
}

func frontmatterPayload(data []byte) ([]byte, error) {
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return nil, fmt.Errorf("issue file must start with a --- frontmatter delimiter")
	}
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			return []byte(strings.Join(lines[1:i], "\n")), nil
		}
	}
	return nil, fmt.Errorf("frontmatter is not closed with a --- delimiter")
}
