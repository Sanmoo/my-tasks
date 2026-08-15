package priority

import (
	"fmt"
	"slices"
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
// ignored by the other actions. Non-prioritizable issues are not part of the
// queue, just as they are not part of the prioritize buffer.
func QuickPlan(issues []Issue, id string, action QuickAction, position int) ([]Change, error) {
	ordered := prioritizableIssues(issues)
	var target Issue
	found := false
	for _, is := range ordered {
		if is.ID == id {
			target = is
			found = true
			break
		}
	}
	if !found {
		for _, is := range issues {
			if is.ID == id {
				return nil, fmt.Errorf("issue %s is %s and cannot be prioritized", id, is.Status)
			}
		}
		return nil, fmt.Errorf("unknown issue ID %q", id)
	}

	queue := make([]Issue, 0, len(ordered))
	backlog := make([]Issue, 0, len(ordered))
	for _, is := range ordered {
		if is.ID == id {
			continue
		}
		if is.Rank == nil {
			backlog = append(backlog, is)
		} else {
			queue = append(queue, is)
		}
	}

	switch action {
	case MoveTop:
		queue = append([]Issue{target}, queue...)
	case MoveBottom:
		queue = append(queue, target)
	case MoveToRank:
		finalLength := len(queue) + 1
		if position < 1 || position > finalLength {
			return nil, fmt.Errorf("rank position must be between 1 and %d", finalLength)
		}
		queue = append(queue, Issue{})
		copy(queue[position:], queue[position-1:])
		queue[position-1] = target
	case RemoveRank:
		backlog = append(backlog, target)
	default:
		return nil, fmt.Errorf("unsupported quick ordering action %d", action)
	}
	return planOrdered(queue, backlog, issues)
}

func planOrdered(queue, backlog []Issue, all []Issue) ([]Change, error) {
	entries := make([]Entry, 0, len(queue)+len(backlog))
	for _, is := range queue {
		entries = append(entries, Entry{Prioritized: true, ID: is.ID})
	}
	for _, is := range backlog {
		entries = append(entries, Entry{ID: is.ID})
	}
	return Plan(entries, all)
}

func prioritizableIssues(issues []Issue) []Issue {
	ordered := make([]Issue, 0, len(issues))
	for _, is := range issues {
		if Prioritizable(is.Status) {
			ordered = append(ordered, is)
		}
	}
	slices.SortFunc(ordered, Compare)
	return ordered
}
