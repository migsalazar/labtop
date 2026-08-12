package config

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/migsalazar/labtop/internal/modules"
	"github.com/pelletier/go-toml/v2"
)

type rawConfig struct {
	Console   *rawConsole   `toml:"console"`
	Layout    *rawLayout    `toml:"layout"`
	Attention *rawAttention `toml:"attention"`
	Network   *rawNetwork   `toml:"network"`
	Modules   *[]rawModule  `toml:"modules"`
}

type rawConsole struct {
	Title       *string `toml:"title"`
	ClockFormat *string `toml:"clock_format"`
}

type rawLayout struct {
	Columns *int `toml:"columns"`
}

type rawAttention struct {
	Enabled           *bool    `toml:"enabled"`
	ReturnHomeSeconds *float64 `toml:"return_home_seconds"`
	EventTypes        []string `toml:"event_types"`
}

type rawNetwork struct {
	ExternalProbeSeconds        *float64            `toml:"external_probe_seconds"`
	ExternalProbeTimeoutSeconds *float64            `toml:"external_probe_timeout_seconds"`
	ExternalTargets             []rawExternalTarget `toml:"external_targets"`
}

type rawExternalTarget struct {
	Label string `toml:"label"`
	Host  string `toml:"host"`
	Port  int    `toml:"port"`
}

type rawModule struct {
	ID             string               `toml:"id"`
	Type           modules.Type         `toml:"type"`
	Title          string               `toml:"title"`
	Column         *int                 `toml:"column"`
	Row            *int                 `toml:"row"`
	ColumnSpan     *int                 `toml:"column_span"`
	RowSpan        *int                 `toml:"row_span"`
	RefreshSeconds *float64             `toml:"refresh_seconds"`
	System         *rawSystemSettings   `toml:"system"`
	Machines       *rawMachinesSettings `toml:"machines"`
	Agents         *rawAgentsSettings   `toml:"agents"`
}

type rawSystemSettings struct {
	LocalLabel                *string  `toml:"local_label"`
	SlowRefreshSeconds        *float64 `toml:"slow_refresh_seconds"`
	NetworkInterface          *string  `toml:"network_interface"`
	SparklineMinutes          *int     `toml:"sparkline_minutes"`
	SparklineSampleSeconds    *float64 `toml:"sparkline_sample_seconds"`
	TemperatureWarningCelsius *float64 `toml:"temperature_warning_celsius"`
	DiskWarningPercent        *float64 `toml:"disk_warning_percent"`
}

type rawMachinesSettings struct {
	ProbeTimeoutSeconds *float64     `toml:"probe_timeout_seconds"`
	Items               []rawMachine `toml:"items"`
}

type rawMachine struct {
	ID    string `toml:"id"`
	Label string `toml:"label"`
	Host  string `toml:"host"`
	Port  int    `toml:"port"`
}

type rawAgentsSettings struct {
	Provider *string `toml:"provider"`
}

// Load loads an explicit path, or the local default path when path is empty.
// A missing implicit default file uses the built-in generic configuration.
func Load(path string) (Config, error) {
	if path == "" {
		file, err := os.Open(DefaultPath)
		if errors.Is(err, os.ErrNotExist) {
			return validate(defaultRawConfig())
		}
		if err != nil {
			return Config{}, fmt.Errorf("read %s: %w", DefaultPath, err)
		}
		defer file.Close()
		return decode(file)
	}

	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}
	defer file.Close()
	return decode(file)
}

func decode(reader io.Reader) (Config, error) {
	var raw rawConfig
	decoder := toml.NewDecoder(reader).DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return Config{}, fmt.Errorf("decode TOML: %w", err)
	}
	applyDefaults(&raw)
	return validate(raw)
}

func applyDefaults(raw *rawConfig) {
	defaults := defaultRawConfig()
	if raw.Console == nil {
		raw.Console = defaults.Console
	} else {
		if raw.Console.Title == nil {
			raw.Console.Title = defaults.Console.Title
		}
		if raw.Console.ClockFormat == nil {
			raw.Console.ClockFormat = defaults.Console.ClockFormat
		}
	}
	if raw.Layout == nil {
		raw.Layout = defaults.Layout
	} else if raw.Layout.Columns == nil {
		raw.Layout.Columns = defaults.Layout.Columns
	}
	if raw.Attention == nil {
		raw.Attention = defaults.Attention
	} else {
		if raw.Attention.Enabled == nil {
			raw.Attention.Enabled = defaults.Attention.Enabled
		}
		if raw.Attention.ReturnHomeSeconds == nil {
			raw.Attention.ReturnHomeSeconds = defaults.Attention.ReturnHomeSeconds
		}
		if raw.Attention.EventTypes == nil {
			raw.Attention.EventTypes = append([]string(nil), defaults.Attention.EventTypes...)
		}
	}
	if raw.Network == nil {
		raw.Network = defaults.Network
	} else {
		if raw.Network.ExternalProbeSeconds == nil {
			raw.Network.ExternalProbeSeconds = defaults.Network.ExternalProbeSeconds
		}
		if raw.Network.ExternalProbeTimeoutSeconds == nil {
			raw.Network.ExternalProbeTimeoutSeconds = defaults.Network.ExternalProbeTimeoutSeconds
		}
	}
	if raw.Modules == nil {
		raw.Modules = defaults.Modules
	}

	for index := range *raw.Modules {
		module := &(*raw.Modules)[index]
		switch module.Type {
		case modules.TypeSystem:
			if module.RefreshSeconds == nil {
				module.RefreshSeconds = floatPointer(defaultSystemRefreshSeconds)
			}
			if module.System != nil {
				applySystemDefaults(module.System)
			}
		case modules.TypeMachines:
			if module.RefreshSeconds == nil {
				module.RefreshSeconds = floatPointer(defaultMachinesRefreshSeconds)
			}
			if module.Machines != nil && module.Machines.ProbeTimeoutSeconds == nil {
				module.Machines.ProbeTimeoutSeconds = floatPointer(defaultMachineProbeTimeoutSeconds)
			}
		case modules.TypeAgents:
			if module.Agents != nil && module.Agents.Provider == nil {
				module.Agents.Provider = stringPointer("")
			}
		}
	}
}

func applySystemDefaults(settings *rawSystemSettings) {
	if settings.LocalLabel == nil {
		settings.LocalLabel = stringPointer("LOCAL NODE")
	}
	if settings.SlowRefreshSeconds == nil {
		settings.SlowRefreshSeconds = floatPointer(defaultSystemSlowRefreshSeconds)
	}
	if settings.NetworkInterface == nil {
		settings.NetworkInterface = stringPointer("")
	}
	if settings.SparklineMinutes == nil {
		settings.SparklineMinutes = intPointer(defaultSparklineMinutes)
	}
	if settings.SparklineSampleSeconds == nil {
		settings.SparklineSampleSeconds = floatPointer(defaultSparklineSampleSeconds)
	}
	if settings.TemperatureWarningCelsius == nil {
		settings.TemperatureWarningCelsius = floatPointer(defaultTemperatureWarningCelsius)
	}
	if settings.DiskWarningPercent == nil {
		settings.DiskWarningPercent = floatPointer(defaultDiskWarningPercent)
	}
}
