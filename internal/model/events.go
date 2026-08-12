package model

import (
	"sort"
	"time"
)

// ConsoleEvent is a normalized state transition emitted by a built-in module.
type ConsoleEvent struct {
	ID             string
	Type           string
	SourceModuleID string
	SourceID       string
	Severity       EventSeverity
	OccurredAt     time.Time
	Title          string
	Detail         string
}

// AttentionState describes one temporary attention presentation.
type AttentionState struct {
	Event    ConsoleEvent
	ShownAt  time.Time
	ReturnAt time.Time
}

// SortEventsByPriority returns a sorted copy with higher severity first, then
// older events first. IDs provide deterministic ordering for exact ties.
func SortEventsByPriority(events []ConsoleEvent) []ConsoleEvent {
	result := append([]ConsoleEvent(nil), events...)
	sort.SliceStable(result, func(left, right int) bool {
		if result[left].Severity != result[right].Severity {
			return result[left].Severity > result[right].Severity
		}
		if !result[left].OccurredAt.Equal(result[right].OccurredAt) {
			return result[left].OccurredAt.Before(result[right].OccurredAt)
		}
		return result[left].ID < result[right].ID
	})
	return result
}
