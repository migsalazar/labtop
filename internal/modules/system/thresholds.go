package system

import (
	"fmt"
	"time"

	"github.com/migsalazar/labtop/internal/model"
)

const hysteresis = 5.0

type thresholdCondition struct {
	eventType     string
	recoveryType  string
	sourceID      string
	warningTitle  string
	recoveryTitle string
	unit          string
	threshold     float64
	active        bool
}

func (condition *thresholdCondition) update(
	moduleID string,
	value model.Optional[float64],
	occurredAt time.Time,
	nextID func(string) string,
) []model.ConsoleEvent {
	if !value.Valid || !finite(value.Value) {
		return nil
	}
	if !condition.active && value.Value >= condition.threshold {
		condition.active = true
		return []model.ConsoleEvent{{
			ID: nextID(condition.eventType), Type: condition.eventType,
			SourceModuleID: moduleID, SourceID: condition.sourceID,
			Severity: model.SeverityWarning, OccurredAt: occurredAt,
			Title:  condition.warningTitle,
			Detail: fmt.Sprintf("Observed %.1f%s; warning threshold %.1f%s.", value.Value, condition.unit, condition.threshold, condition.unit),
		}}
	}
	if condition.active && value.Value <= condition.threshold-hysteresis {
		condition.active = false
		return []model.ConsoleEvent{{
			ID: nextID(condition.recoveryType), Type: condition.recoveryType,
			SourceModuleID: moduleID, SourceID: condition.sourceID,
			Severity: model.SeverityInfo, OccurredAt: occurredAt,
			Title:  condition.recoveryTitle,
			Detail: fmt.Sprintf("Observed %.1f%s; warning condition resolved.", value.Value, condition.unit),
		}}
	}
	return nil
}
