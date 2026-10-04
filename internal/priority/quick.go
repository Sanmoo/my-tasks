package priority

import (
	"fmt"
	"slices"

	"github.com/Sanmoo/my-tasks2/internal/issue"
)

// QuickAction identifies the immediate order change requested by a quick
// ordering command.
type QuickAction uint8

const (
	// MoveTop promotes an issue to the first position in the queue.
	MoveTop QuickAction = iota
	// MoveBottom moves an issue to the last position in the queue.
	MoveBottom
	// MoveToRank inserts an issue at the requested one-based position.
	MoveToRank
	// RemoveRank returns an issue to the Backlog.
	RemoveRank
)

// QuickPlan computes the minimal rank changes for a quick ordering action.
// The position is the one-based final queue position for MoveToRank and is
// ignored by the other actions. Non-prioritizable Issues are not part of the
// queue, just as they are not part of the prioritize buffer.
func QuickPlan(items []issue.Item, id string, action QuickAction, position int) ([]issue.Change, error) {
	target, err := findPrioritizable(items, id)
	if err != nil {
		return nil, err
	}

	ordered := prioritizableItems(items)
	queue := make([]issue.Item, 0, len(ordered))
	backlog := make([]issue.Item, 0, len(ordered))
	for _, it := range ordered {
		if it.ID == id {
			continue
		}
		if it.Issue.Frontmatter.Rank == nil {
			backlog = append(backlog, it)
		} else {
			queue = append(queue, it)
		}
	}

	switch action {
	case MoveTop:
		queue = append([]issue.Item{target}, queue...)
	case MoveBottom:
		queue = append(queue, target)
	case MoveToRank:
		finalLength := len(queue) + 1
		if position < 1 || position > finalLength {
			return nil, fmt.Errorf("rank position must be between 1 and %d", finalLength)
		}
		queue = append(queue, issue.Item{})
		copy(queue[position:], queue[position-1:])
		queue[position-1] = target
	case RemoveRank:
		backlog = append(backlog, target)
	default:
		return nil, fmt.Errorf("unsupported quick ordering action %d", action)
	}
	return planOrdered(queue, backlog, items)
}

func planOrdered(queue, backlog []issue.Item, all []issue.Item) ([]issue.Change, error) {
	entries := make([]Entry, 0, len(queue)+len(backlog))
	for _, it := range queue {
		entries = append(entries, Entry{Prioritized: true, ID: it.ID})
	}
	for _, it := range backlog {
		entries = append(entries, Entry{ID: it.ID})
	}
	return Plan(entries, all)
}

func prioritizableItems(items []issue.Item) []issue.Item {
	ordered := make([]issue.Item, 0, len(items))
	for _, it := range items {
		if Prioritizable(it.Issue.Frontmatter.Status) {
			ordered = append(ordered, it)
		}
	}
	slices.SortFunc(ordered, issue.Compare)
	return ordered
}
