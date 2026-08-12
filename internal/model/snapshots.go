package model

import "time"

// Optional represents a value that may be unavailable without assigning
// meaning to the value type's zero value.
type Optional[T any] struct {
	Value T
	Valid bool
}

// Some returns a present optional value.
func Some[T any](value T) Optional[T] {
	return Optional[T]{Value: value, Valid: true}
}

// None returns an unavailable optional value.
func None[T any]() Optional[T] {
	return Optional[T]{}
}

// SystemSnapshot is a complete value produced by one local collection update.
type SystemSnapshot struct {
	CollectedAt                   time.Time
	Status                        ModuleStatus
	CPUPercent                    Optional[float64]
	MemoryPercent                 Optional[float64]
	DiskPercent                   Optional[float64]
	TemperatureCelsius            Optional[float64]
	NetworkReceiveBytesPerSecond  Optional[float64]
	NetworkTransmitBytesPerSecond Optional[float64]
	Uptime                        Optional[time.Duration]
}

// LocalInterfaceSnapshot reports local-interface selection evidence.
type LocalInterfaceSnapshot struct {
	CheckedAt     time.Time
	Status        LocalInterfaceStatus
	InterfaceName string
	Error         string
}

// ExternalReachabilitySnapshot reports configured external-target evidence.
type ExternalReachabilitySnapshot struct {
	CheckedAt time.Time
	Status    ExternalReachabilityStatus
	Latency   Optional[time.Duration]
	Error     string
}

// MachineState stores the latest public reachability state for one machine.
type MachineState struct {
	ID                  string
	Status              ReachabilityStatus
	Latency             Optional[time.Duration]
	LastSeen            Optional[time.Time]
	ConsecutiveFailures int
}

// AgentSnapshot is reserved for a future truthful provider implementation.
type AgentSnapshot struct {
	ID          string
	Label       string
	Status      AgentStatus
	Task        Optional[string]
	Model       Optional[string]
	Runtime     Optional[time.Duration]
	CPUPercent  Optional[float64]
	MemoryBytes Optional[uint64]
	ChildCount  Optional[int]
}
