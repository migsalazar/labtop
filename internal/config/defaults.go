package config

import "github.com/migsalazar/labtop/internal/modules"

const (
	defaultTitle                       = "LABTOP // CONSOLE"
	defaultClockFormat                 = "15:04:05"
	defaultColumns                     = 3
	defaultReturnHomeSeconds           = 15.0
	defaultExternalProbeSeconds        = 15.0
	defaultExternalProbeTimeoutSeconds = 2.0
	defaultSystemRefreshSeconds        = 1.0
	defaultSystemSlowRefreshSeconds    = 5.0
	defaultSparklineMinutes            = 15
	defaultSparklineSampleSeconds      = 5.0
	defaultTemperatureWarningCelsius   = 80.0
	defaultDiskWarningPercent          = 90.0
	defaultMachinesRefreshSeconds      = 5.0
	defaultMachineProbeTimeoutSeconds  = 1.5
)

var defaultAttentionEventTypes = []string{
	"machine.offline",
	"system.temperature_warning",
	"system.disk_warning",
}

func defaultRawConfig() rawConfig {
	return rawConfig{
		Console: &rawConsole{
			Title:       stringPointer(defaultTitle),
			ClockFormat: stringPointer(defaultClockFormat),
		},
		Layout: &rawLayout{Columns: intPointer(defaultColumns)},
		Attention: &rawAttention{
			Enabled:           boolPointer(true),
			ReturnHomeSeconds: floatPointer(defaultReturnHomeSeconds),
			EventTypes:        append([]string(nil), defaultAttentionEventTypes...),
		},
		Network: &rawNetwork{
			ExternalProbeSeconds:        floatPointer(defaultExternalProbeSeconds),
			ExternalProbeTimeoutSeconds: floatPointer(defaultExternalProbeTimeoutSeconds),
		},
		Modules: moduleSlicePointer([]rawModule{
			{
				ID:             "local-system",
				Type:           modules.TypeSystem,
				Title:          "SYSTEM",
				Column:         intPointer(0),
				Row:            intPointer(0),
				ColumnSpan:     intPointer(1),
				RowSpan:        intPointer(1),
				RefreshSeconds: floatPointer(defaultSystemRefreshSeconds),
				System: &rawSystemSettings{
					LocalLabel:                stringPointer("LOCAL NODE"),
					SlowRefreshSeconds:        floatPointer(defaultSystemSlowRefreshSeconds),
					NetworkInterface:          stringPointer(""),
					SparklineMinutes:          intPointer(defaultSparklineMinutes),
					SparklineSampleSeconds:    floatPointer(defaultSparklineSampleSeconds),
					TemperatureWarningCelsius: floatPointer(defaultTemperatureWarningCelsius),
					DiskWarningPercent:        floatPointer(defaultDiskWarningPercent),
				},
			},
			{
				ID:         "active-agents",
				Type:       modules.TypeAgents,
				Title:      "AGENTS",
				Column:     intPointer(1),
				Row:        intPointer(0),
				ColumnSpan: intPointer(2),
				RowSpan:    intPointer(1),
				Agents:     &rawAgentsSettings{Provider: stringPointer("")},
			},
			{
				ID:             "machines",
				Type:           modules.TypeMachines,
				Title:          "MACHINES",
				Column:         intPointer(0),
				Row:            intPointer(1),
				ColumnSpan:     intPointer(2),
				RowSpan:        intPointer(1),
				RefreshSeconds: floatPointer(defaultMachinesRefreshSeconds),
				Machines: &rawMachinesSettings{
					ProbeTimeoutSeconds: floatPointer(defaultMachineProbeTimeoutSeconds),
				},
			},
			{
				ID:         "recent-events",
				Type:       modules.TypeEvents,
				Title:      "EVENTS",
				Column:     intPointer(2),
				Row:        intPointer(1),
				ColumnSpan: intPointer(1),
				RowSpan:    intPointer(1),
			},
		}),
	}
}

func moduleSlicePointer(value []rawModule) *[]rawModule { return &value }
func stringPointer(value string) *string                { return &value }
func boolPointer(value bool) *bool                      { return &value }
func intPointer(value int) *int                         { return &value }
func floatPointer(value float64) *float64               { return &value }
