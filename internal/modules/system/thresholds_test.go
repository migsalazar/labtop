package system

import (
	"testing"
	"time"

	"github.com/migsalazar/labtop/internal/model"
)

func TestThresholdConditionTransitions(t *testing.T) {
	t.Parallel()

	condition := thresholdCondition{
		eventType: "warning", recoveryType: "recovered", sourceID: "metric",
		warningTitle: "WARNING", recoveryTitle: "RECOVERED", unit: "%", threshold: 90,
	}
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	sequence := 0
	nextID := func(eventType string) string {
		sequence++
		return eventType
	}

	if events := condition.update("system", model.Some(89.9), now, nextID); len(events) != 0 {
		t.Fatalf("below threshold events = %#v", events)
	}
	warning := condition.update("system", model.Some(90.0), now, nextID)
	if len(warning) != 1 || warning[0].Type != "warning" || warning[0].Severity != model.SeverityWarning || warning[0].SourceID != "metric" {
		t.Fatalf("warning = %#v", warning)
	}
	if events := condition.update("system", model.Some(99.0), now, nextID); len(events) != 0 {
		t.Fatalf("duplicate warning = %#v", events)
	}
	if events := condition.update("system", model.None[float64](), now, nextID); len(events) != 0 || !condition.active {
		t.Fatalf("missing active metric events=%#v active=%t", events, condition.active)
	}
	if events := condition.update("system", model.Some(85.1), now, nextID); len(events) != 0 {
		t.Fatalf("above recovery boundary events = %#v", events)
	}
	recovery := condition.update("system", model.Some(85.0), now, nextID)
	if len(recovery) != 1 || recovery[0].Type != "recovered" || recovery[0].Severity != model.SeverityInfo || condition.active {
		t.Fatalf("recovery = %#v active=%t", recovery, condition.active)
	}
	if recurrence := condition.update("system", model.Some(91.0), now, nextID); len(recurrence) != 1 || recurrence[0].Type != "warning" {
		t.Fatalf("recurrence = %#v", recurrence)
	}
}
