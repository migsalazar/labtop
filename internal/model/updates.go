package model

import "time"

// ModuleUpdate is the closed set of value messages collectors may publish.
type ModuleUpdate interface {
	moduleUpdate()
}

// SystemUpdate publishes one complete system snapshot and its transitions.
type SystemUpdate struct {
	ModuleID string
	Snapshot SystemSnapshot
	Events   []ConsoleEvent
}

func (SystemUpdate) moduleUpdate() {}

// SystemSampleUpdate publishes values for bounded CPU and memory history.
type SystemSampleUpdate struct {
	ModuleID      string
	SampledAt     time.Time
	CPUPercent    Optional[float64]
	MemoryPercent Optional[float64]
}

func (SystemSampleUpdate) moduleUpdate() {}

// MachinesUpdate publishes complete machine states and their transitions.
type MachinesUpdate struct {
	ModuleID string
	Machines []MachineState
	Events   []ConsoleEvent
}

func (MachinesUpdate) moduleUpdate() {}

// LocalInterfaceUpdate publishes local-interface selection evidence.
type LocalInterfaceUpdate struct {
	Snapshot LocalInterfaceSnapshot
}

func (LocalInterfaceUpdate) moduleUpdate() {}

// ExternalReachabilityUpdate publishes configured external-target evidence.
type ExternalReachabilityUpdate struct {
	Snapshot ExternalReachabilitySnapshot
}

func (ExternalReachabilityUpdate) moduleUpdate() {}
