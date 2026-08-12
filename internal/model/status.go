// Package model defines shared immutable values passed between Labtop components.
package model

// ReachabilityStatus describes evidence from a configured machine TCP probe.
type ReachabilityStatus string

const (
	ReachabilityNotConfigured ReachabilityStatus = "not_configured"
	ReachabilityChecking      ReachabilityStatus = "checking"
	ReachabilityOnline        ReachabilityStatus = "online"
	ReachabilityOffline       ReachabilityStatus = "offline"
)

// LocalInterfaceStatus describes whether a qualifying local interface exists.
type LocalInterfaceStatus string

const (
	LocalInterfaceChecking    LocalInterfaceStatus = "checking"
	LocalInterfaceAvailable   LocalInterfaceStatus = "available"
	LocalInterfaceUnavailable LocalInterfaceStatus = "unavailable"
)

// ExternalReachabilityStatus describes configured external-target evidence.
type ExternalReachabilityStatus string

const (
	ExternalNotConfigured ExternalReachabilityStatus = "not_configured"
	ExternalChecking      ExternalReachabilityStatus = "checking"
	ExternalReachable     ExternalReachabilityStatus = "reachable"
	ExternalUnreachable   ExternalReachabilityStatus = "unreachable"
)

// ModuleStatus describes whether a module can currently provide data.
type ModuleStatus string

const (
	ModuleNotConfigured ModuleStatus = "not_configured"
	ModuleReady         ModuleStatus = "ready"
	ModuleUnavailable   ModuleStatus = "unavailable"
)

// AgentStatus is reserved for truthful snapshots from a future agent provider.
type AgentStatus string

const (
	AgentUnknown   AgentStatus = "unknown"
	AgentWorking   AgentStatus = "working"
	AgentWaiting   AgentStatus = "waiting"
	AgentCompleted AgentStatus = "completed"
	AgentFailed    AgentStatus = "failed"
)

// EventSeverity controls deterministic warning priority.
type EventSeverity int

const (
	SeverityInfo     EventSeverity = 10
	SeverityWarning  EventSeverity = 20
	SeverityCritical EventSeverity = 30
)

// DisplayMode identifies the ambient display controller state.
type DisplayMode string

const (
	DisplayHome      DisplayMode = "home"
	DisplayAttention DisplayMode = "attention"
)
