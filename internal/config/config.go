// Package config loads and validates Labtop's local TOML configuration.
package config

import (
	"time"

	"github.com/migsalazar/labtop/internal/modules"
)

const DefaultPath = "config.toml"

// Config is a validated runtime configuration.
type Config struct {
	Console   Console
	Layout    Layout
	Attention Attention
	Network   Network
	Modules   []Module
}

type Console struct {
	Title       string
	ClockFormat string
}

type Layout struct {
	Columns int
}

type Attention struct {
	Enabled    bool
	ReturnHome time.Duration
	EventTypes []string
}

type Network struct {
	ExternalProbeInterval time.Duration
	ExternalProbeTimeout  time.Duration
	ExternalTargets       []ExternalTarget
}

type ExternalTarget struct {
	Label string
	Host  string
	Port  uint16
}

type Module struct {
	ID         string
	Type       modules.Type
	Title      string
	Column     int
	Row        int
	ColumnSpan int
	RowSpan    int
	Refresh    time.Duration
	System     *SystemSettings
	Machines   *MachinesSettings
	Agents     *AgentsSettings
}

type SystemSettings struct {
	LocalLabel                string
	SlowRefresh               time.Duration
	NetworkInterface          string
	SparklineDuration         time.Duration
	SparklineSampleInterval   time.Duration
	SparklineCapacity         int
	TemperatureWarningCelsius float64
	DiskWarningPercent        float64
}

type MachinesSettings struct {
	ProbeTimeout time.Duration
	Items        []Machine
}

type Machine struct {
	ID    string
	Label string
	Host  string
	Port  uint16
}

type AgentsSettings struct {
	Provider string
}
