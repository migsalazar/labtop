package model

import (
	"reflect"
	"testing"
	"time"
)

func TestOptionalDistinguishesMissingFromZero(t *testing.T) {
	t.Parallel()

	missing := None[int]()
	zero := Some(0)
	if missing.Valid || missing.Value != 0 {
		t.Fatalf("None[int]() = %#v", missing)
	}
	if !zero.Valid || zero.Value != 0 {
		t.Fatalf("Some(0) = %#v", zero)
	}
}

func TestModuleUpdateClosedValues(t *testing.T) {
	t.Parallel()

	updates := []ModuleUpdate{
		SystemUpdate{},
		SystemSampleUpdate{},
		MachinesUpdate{},
		LocalInterfaceUpdate{},
		ExternalReachabilityUpdate{},
	}
	if len(updates) != 5 {
		t.Fatalf("updates = %d, want 5", len(updates))
	}
}

func TestSortEventsByPriorityReturnsIndependentDeterministicCopy(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	input := []ConsoleEvent{
		{ID: "warning-new", Severity: SeverityWarning, OccurredAt: now.Add(time.Minute)},
		{ID: "info", Severity: SeverityInfo, OccurredAt: now},
		{ID: "warning-old-b", Severity: SeverityWarning, OccurredAt: now},
		{ID: "critical", Severity: SeverityCritical, OccurredAt: now.Add(2 * time.Minute)},
		{ID: "warning-old-a", Severity: SeverityWarning, OccurredAt: now},
	}
	got := SortEventsByPriority(input)
	wantIDs := []string{"critical", "warning-old-a", "warning-old-b", "warning-new", "info"}
	gotIDs := make([]string, len(got))
	for index := range got {
		gotIDs[index] = got[index].ID
	}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("event order = %v, want %v", gotIDs, wantIDs)
	}
	if input[0].ID != "warning-new" {
		t.Fatal("SortEventsByPriority mutated input order")
	}
	got[0].ID = "mutated"
	if input[3].ID != "critical" {
		t.Fatal("sorted result shares event storage with input")
	}
}

func TestUpdateValuesDoNotAliasSourceSlicesAfterCopy(t *testing.T) {
	t.Parallel()

	machines := []MachineState{{ID: "node-a"}}
	events := []ConsoleEvent{{ID: "event-a"}}
	update := MachinesUpdate{
		ModuleID: "machines",
		Machines: append([]MachineState(nil), machines...),
		Events:   append([]ConsoleEvent(nil), events...),
	}
	machines[0].ID = "changed"
	events[0].ID = "changed"
	if update.Machines[0].ID != "node-a" || update.Events[0].ID != "event-a" {
		t.Fatalf("update aliases source values: %#v", update)
	}
}
