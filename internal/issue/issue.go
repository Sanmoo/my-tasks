// Package issue holds the pure logic of Issues: the YAML frontmatter
// (stable field order, optional fields only-when-set) and issue ID
// generation. It is decision-dense, so it lives at Seam 2: black-box
// unit tested, with the coverage and mutation gates.
package issue

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// NaiveLayout is the canonical datetime layout of every Issue datetime
// field (created_at, deferred_until, deadline, started_at, completed_at):
// local time, naive (no timezone), no seconds — YYYY-MM-DDTHH:MM.
const NaiveLayout = "2006-01-02T15:04"

// Frontmatter is the YAML header of an Issue file. The field order here
// is the canonical order on disk; keep it in sync with the spec schema.
//
// Always present: title, status, labels, created_at. Present only when
// they have a value: rank, deferred_until, deadline, started_at,
// completed_at. There is no id (the file name is the authority) and no
// updated_at (Git and mtime track that).
type Frontmatter struct {
	Title     string   `yaml:"title"`
	Status    string   `yaml:"status"`
	Labels    []string `yaml:"labels,flow"`
	CreatedAt string   `yaml:"created_at"`

	Rank          *int   `yaml:"rank,omitempty"`
	DeferredUntil string `yaml:"deferred_until,omitempty"`
	Deadline      string `yaml:"deadline,omitempty"`
	StartedAt     string `yaml:"started_at,omitempty"`
	CompletedAt   string `yaml:"completed_at,omitempty"`
}

// Issue is one unit of work: the frontmatter plus the Markdown body
// (the Description/Notes/Comments sections), verbatim after the closing
// delimiter.
type Issue struct {
	Frontmatter Frontmatter
	Body        string
}

// DefaultBody is the body of a freshly created Issue: a blank line after
// the closing delimiter, then the three sections of the spec schema, all
// empty, with a trailing newline.
const DefaultBody = "\n## Description\n## Notes\n## Comments\n"

// Render serializes an Issue to the on-disk form: a --- delimiter, the
// frontmatter mapping, a closing --- delimiter, then the body verbatim.
func Render(i Issue) ([]byte, error) {
	fm, err := yaml.Marshal(i.Frontmatter)
	if err != nil {
		return nil, fmt.Errorf("encoding frontmatter: %w", err)
	}
	var b strings.Builder
	b.WriteString("---\n")
	b.Write(fm)
	b.WriteString("---\n")
	b.WriteString(i.Body)
	return []byte(b.String()), nil
}

// Parse reads an Issue from its on-disk form: a leading --- line, the
// frontmatter mapping, a closing --- line, then the body (returned
// verbatim, including any trailing newline).
func Parse(data []byte) (Issue, error) {
	// strings.Split always yields at least one element (an empty string
	// for empty input), so lines[0] is safe and covers the empty case.
	lines := strings.Split(string(data), "\n")
	if lines[0] != "---" {
		return Issue{}, errors.New("issue file must start with a --- frontmatter delimiter")
	}
	// Find the closing --- delimiter. ok disambiguates "found at index
	// end" from "not found": a valid end index is always >= 1, so a
	// numeric sentinel would leave an equivalent boundary mutant.
	var end int
	ok := false
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			end = i
			ok = true
			break
		}
	}
	if !ok {
		return Issue{}, errors.New("frontmatter is not closed with a --- delimiter")
	}
	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(strings.Join(lines[1:end], "\n")), &fm); err != nil {
		return Issue{}, fmt.Errorf("parsing frontmatter: %w", err)
	}
	return Issue{Frontmatter: fm, Body: strings.Join(lines[end+1:], "\n")}, nil
}
