package config

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/migsalazar/labtop/internal/modules"
)

const supportedColumns = 3

var supportedAttentionEvents = map[string]struct{}{
	"machine.offline":            {},
	"system.temperature_warning": {},
	"system.disk_warning":        {},
}

func validate(raw rawConfig) (Config, error) {
	if raw.Console == nil || raw.Layout == nil || raw.Attention == nil || raw.Network == nil {
		return Config{}, fmt.Errorf("internal error: defaults are incomplete")
	}

	title := strings.TrimSpace(value(raw.Console.Title))
	if title == "" {
		return Config{}, fmt.Errorf("console.title must not be empty")
	}
	clockFormat := value(raw.Console.ClockFormat)
	if strings.TrimSpace(clockFormat) == "" {
		return Config{}, fmt.Errorf("console.clock_format must not be empty")
	}
	if value(raw.Layout.Columns) != supportedColumns {
		return Config{}, fmt.Errorf("layout.columns must be %d", supportedColumns)
	}

	returnHome, err := positiveDuration("attention.return_home_seconds", value(raw.Attention.ReturnHomeSeconds))
	if err != nil {
		return Config{}, err
	}
	eventTypes, err := validateEventTypes(raw.Attention.EventTypes)
	if err != nil {
		return Config{}, err
	}
	externalInterval, err := positiveDuration("network.external_probe_seconds", value(raw.Network.ExternalProbeSeconds))
	if err != nil {
		return Config{}, err
	}
	externalTimeout, err := positiveDuration("network.external_probe_timeout_seconds", value(raw.Network.ExternalProbeTimeoutSeconds))
	if err != nil {
		return Config{}, err
	}
	targets, err := validateExternalTargets(raw.Network.ExternalTargets)
	if err != nil {
		return Config{}, err
	}
	configuredModules, err := validateModules(value(raw.Modules), supportedColumns)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Console: Console{Title: title, ClockFormat: clockFormat},
		Layout:  Layout{Columns: supportedColumns},
		Attention: Attention{
			Enabled:    value(raw.Attention.Enabled),
			ReturnHome: returnHome,
			EventTypes: eventTypes,
		},
		Network: Network{
			ExternalProbeInterval: externalInterval,
			ExternalProbeTimeout:  externalTimeout,
			ExternalTargets:       targets,
		},
		Modules: configuredModules,
	}, nil
}

func validateEventTypes(raw []string) ([]string, error) {
	result := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for index, eventType := range raw {
		if _, ok := supportedAttentionEvents[eventType]; !ok {
			return nil, fmt.Errorf("attention.event_types[%d]: unsupported event type %q", index, eventType)
		}
		if _, ok := seen[eventType]; ok {
			return nil, fmt.Errorf("attention.event_types[%d]: duplicate event type %q", index, eventType)
		}
		seen[eventType] = struct{}{}
		result = append(result, eventType)
	}
	return result, nil
}

func validateExternalTargets(raw []rawExternalTarget) ([]ExternalTarget, error) {
	result := make([]ExternalTarget, 0, len(raw))
	for index, target := range raw {
		prefix := fmt.Sprintf("network.external_targets[%d]", index)
		if strings.TrimSpace(target.Label) == "" {
			return nil, fmt.Errorf("%s.label must not be empty", prefix)
		}
		if strings.TrimSpace(target.Host) == "" {
			return nil, fmt.Errorf("%s.host must not be empty", prefix)
		}
		port, err := validPort(prefix+".port", target.Port)
		if err != nil {
			return nil, err
		}
		result = append(result, ExternalTarget{Label: target.Label, Host: target.Host, Port: port})
	}
	return result, nil
}

func validateModules(raw []rawModule, columns int) ([]Module, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("modules must contain at least one module")
	}

	result := make([]Module, 0, len(raw))
	ids := make(map[string]int, len(raw))
	types := make(map[modules.Type]int, len(raw))
	for index, rawModule := range raw {
		prefix := fmt.Sprintf("modules[%d]", index)
		if strings.TrimSpace(rawModule.ID) == "" {
			return nil, fmt.Errorf("%s.id must not be empty", prefix)
		}
		if previous, ok := ids[rawModule.ID]; ok {
			return nil, fmt.Errorf("%s.id duplicates modules[%d].id %q", prefix, previous, rawModule.ID)
		}
		ids[rawModule.ID] = index
		if strings.TrimSpace(rawModule.Title) == "" {
			return nil, fmt.Errorf("%s.title must not be empty", prefix)
		}

		definition, ok := modules.Lookup(rawModule.Type)
		if !ok {
			return nil, fmt.Errorf("%s.type: unsupported module type %q", prefix, rawModule.Type)
		}
		if definition.Singleton {
			if previous, ok := types[rawModule.Type]; ok {
				return nil, fmt.Errorf("%s.type duplicates singleton type in modules[%d]", prefix, previous)
			}
			types[rawModule.Type] = index
		}
		if !definition.RefreshAllowed && rawModule.RefreshSeconds != nil {
			return nil, fmt.Errorf("%s.refresh_seconds is not supported for type %q", prefix, rawModule.Type)
		}
		if err := validateSettingsBlock(prefix, rawModule, definition.Settings); err != nil {
			return nil, err
		}

		column, err := nonNegative(prefix+".column", rawModule.Column)
		if err != nil {
			return nil, err
		}
		row, err := nonNegative(prefix+".row", rawModule.Row)
		if err != nil {
			return nil, err
		}
		columnSpan, err := positiveInt(prefix+".column_span", rawModule.ColumnSpan)
		if err != nil {
			return nil, err
		}
		rowSpan, err := positiveInt(prefix+".row_span", rawModule.RowSpan)
		if err != nil {
			return nil, err
		}
		if column > columns-columnSpan {
			return nil, fmt.Errorf("%s exceeds the %d-column layout", prefix, columns)
		}
		if row > math.MaxInt-rowSpan {
			return nil, fmt.Errorf("%s.row and row_span overflow", prefix)
		}

		configured := Module{
			ID: rawModule.ID, Type: rawModule.Type, Title: rawModule.Title,
			Column: column, Row: row, ColumnSpan: columnSpan, RowSpan: rowSpan,
		}
		if definition.RefreshAllowed {
			configured.Refresh, err = positiveDuration(prefix+".refresh_seconds", value(rawModule.RefreshSeconds))
			if err != nil {
				return nil, err
			}
		}
		if rawModule.System != nil {
			configured.System, err = validateSystem(prefix+".system", *rawModule.System)
		}
		if rawModule.Machines != nil {
			configured.Machines, err = validateMachines(prefix+".machines", *rawModule.Machines)
		}
		if rawModule.Agents != nil {
			configured.Agents, err = validateAgents(prefix+".agents", *rawModule.Agents)
		}
		if err != nil {
			return nil, err
		}
		result = append(result, configured)
	}

	for left := 0; left < len(result); left++ {
		for right := left + 1; right < len(result); right++ {
			if overlaps(result[left], result[right]) {
				return nil, fmt.Errorf("module %q overlaps module %q", result[right].ID, result[left].ID)
			}
		}
	}
	sort.SliceStable(result, func(left, right int) bool {
		if result[left].Row != result[right].Row {
			return result[left].Row < result[right].Row
		}
		return result[left].Column < result[right].Column
	})
	return result, nil
}

func validateSettingsBlock(prefix string, raw rawModule, expected modules.SettingsKind) error {
	present := make([]modules.SettingsKind, 0, 3)
	if raw.System != nil {
		present = append(present, modules.SettingsSystem)
	}
	if raw.Machines != nil {
		present = append(present, modules.SettingsMachines)
	}
	if raw.Agents != nil {
		present = append(present, modules.SettingsAgents)
	}
	if expected == modules.SettingsNone {
		if len(present) != 0 {
			return fmt.Errorf("%s: type %q does not accept a type-specific settings block", prefix, raw.Type)
		}
		return nil
	}
	if len(present) == 0 {
		return fmt.Errorf("%s: type %q requires a %s settings block", prefix, raw.Type, expected)
	}
	if len(present) != 1 || present[0] != expected {
		return fmt.Errorf("%s: type %q requires only a %s settings block", prefix, raw.Type, expected)
	}
	return nil
}

func validateSystem(prefix string, raw rawSystemSettings) (*SystemSettings, error) {
	if strings.TrimSpace(value(raw.LocalLabel)) == "" {
		return nil, fmt.Errorf("%s.local_label must not be empty", prefix)
	}
	slowRefresh, err := positiveDuration(prefix+".slow_refresh_seconds", value(raw.SlowRefreshSeconds))
	if err != nil {
		return nil, err
	}
	minutes := value(raw.SparklineMinutes)
	if minutes <= 0 {
		return nil, fmt.Errorf("%s.sparkline_minutes must be positive", prefix)
	}
	sample, err := positiveDuration(prefix+".sparkline_sample_seconds", value(raw.SparklineSampleSeconds))
	if err != nil {
		return nil, err
	}
	if int64(minutes) > int64(math.MaxInt64)/int64(time.Minute) {
		return nil, fmt.Errorf("%s.sparkline_minutes is too large", prefix)
	}
	duration := time.Duration(minutes) * time.Minute
	if duration/sample <= 0 || duration%sample != 0 {
		return nil, fmt.Errorf("%s: sparkline duration must be an integral number of samples", prefix)
	}
	capacity64 := int64(duration / sample)
	if capacity64 > int64(math.MaxInt) {
		return nil, fmt.Errorf("%s: sparkline capacity is too large", prefix)
	}
	temperature := value(raw.TemperatureWarningCelsius)
	if !finite(temperature) {
		return nil, fmt.Errorf("%s.temperature_warning_celsius must be finite", prefix)
	}
	disk := value(raw.DiskWarningPercent)
	if !finite(disk) || disk <= 0 || disk > 100 {
		return nil, fmt.Errorf("%s.disk_warning_percent must be greater than 0 and at most 100", prefix)
	}
	return &SystemSettings{
		LocalLabel: value(raw.LocalLabel), SlowRefresh: slowRefresh,
		NetworkInterface: value(raw.NetworkInterface), SparklineDuration: duration,
		SparklineSampleInterval: sample, SparklineCapacity: int(capacity64),
		TemperatureWarningCelsius: temperature, DiskWarningPercent: disk,
	}, nil
}

func validateMachines(prefix string, raw rawMachinesSettings) (*MachinesSettings, error) {
	probeTimeout, err := positiveDuration(prefix+".probe_timeout_seconds", value(raw.ProbeTimeoutSeconds))
	if err != nil {
		return nil, err
	}
	if len(raw.Items) > 2 {
		return nil, fmt.Errorf("%s.items supports at most two machines", prefix)
	}
	items := make([]Machine, 0, len(raw.Items))
	ids := make(map[string]int, len(raw.Items))
	for index, item := range raw.Items {
		itemPrefix := fmt.Sprintf("%s.items[%d]", prefix, index)
		if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Label) == "" || strings.TrimSpace(item.Host) == "" {
			return nil, fmt.Errorf("%s requires non-empty id, label, and host", itemPrefix)
		}
		if previous, ok := ids[item.ID]; ok {
			return nil, fmt.Errorf("%s.id duplicates items[%d].id %q", itemPrefix, previous, item.ID)
		}
		ids[item.ID] = index
		port, err := validPort(itemPrefix+".port", item.Port)
		if err != nil {
			return nil, err
		}
		items = append(items, Machine{ID: item.ID, Label: item.Label, Host: item.Host, Port: port})
	}
	return &MachinesSettings{ProbeTimeout: probeTimeout, Items: items}, nil
}

func validateAgents(prefix string, raw rawAgentsSettings) (*AgentsSettings, error) {
	provider := value(raw.Provider)
	if provider != "" {
		return nil, fmt.Errorf("%s.provider: no agent provider is implemented", prefix)
	}
	return &AgentsSettings{}, nil
}

func positiveDuration(name string, seconds float64) (time.Duration, error) {
	if !finite(seconds) || seconds <= 0 {
		return 0, fmt.Errorf("%s must be a finite positive number", name)
	}
	if seconds > float64(math.MaxInt64)/float64(time.Second) {
		return 0, fmt.Errorf("%s is too large", name)
	}
	duration := time.Duration(seconds * float64(time.Second))
	if duration <= 0 {
		return 0, fmt.Errorf("%s must be at least one nanosecond", name)
	}
	return duration, nil
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func validPort(name string, port int) (uint16, error) {
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("%s must be from 1 through 65535", name)
	}
	return uint16(port), nil
}

func nonNegative(name string, number *int) (int, error) {
	if number == nil {
		return 0, fmt.Errorf("%s is required", name)
	}
	if *number < 0 {
		return 0, fmt.Errorf("%s must not be negative", name)
	}
	return *number, nil
}

func positiveInt(name string, number *int) (int, error) {
	if number == nil {
		return 0, fmt.Errorf("%s is required", name)
	}
	if *number <= 0 {
		return 0, fmt.Errorf("%s must be positive", name)
	}
	return *number, nil
}

func overlaps(left, right Module) bool {
	return left.Column < right.Column+right.ColumnSpan && right.Column < left.Column+left.ColumnSpan &&
		left.Row < right.Row+right.RowSpan && right.Row < left.Row+left.RowSpan
}

func value[T any](pointer *T) T {
	if pointer == nil {
		var zero T
		return zero
	}
	return *pointer
}
