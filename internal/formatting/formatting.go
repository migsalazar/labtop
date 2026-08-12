// Package formatting provides deterministic terminal-independent value formatting.
package formatting

import (
	"fmt"
	"math"
	"time"

	"github.com/migsalazar/labtop/internal/model"
)

const unavailable = "—"

// Percentage formats a percentage as a rounded whole number.
func Percentage(value model.Optional[float64]) string {
	if !validNumber(value) {
		return unavailable
	}
	return fmt.Sprintf("%.0f%%", value.Value)
}

// Temperature formats Celsius as a rounded whole number.
func Temperature(value model.Optional[float64]) string {
	if !validNumber(value) {
		return unavailable
	}
	return fmt.Sprintf("%.0f°C", value.Value)
}

// Throughput formats non-negative bytes per second using adaptive IEC units.
func Throughput(value model.Optional[float64]) string {
	if !validNumber(value) || value.Value < 0 {
		return unavailable
	}
	units := [...]string{"B/s", "KiB/s", "MiB/s", "GiB/s", "TiB/s", "PiB/s"}
	scaled := value.Value
	unit := 0
	for scaled >= 1024 && unit < len(units)-1 {
		scaled /= 1024
		unit++
	}
	if unit == 0 {
		return fmt.Sprintf("%.0f %s", scaled, units[unit])
	}
	return fmt.Sprintf("%.1f %s", scaled, units[unit])
}

// Duration formats a non-negative duration in compact, stable units.
func Duration(value model.Optional[time.Duration]) string {
	if !value.Valid || value.Value < 0 {
		return unavailable
	}
	duration := value.Value
	switch {
	case duration < time.Minute:
		return fmt.Sprintf("%ds", duration/time.Second)
	case duration < time.Hour:
		return fmt.Sprintf("%dm", duration/time.Minute)
	case duration < 24*time.Hour:
		return fmt.Sprintf("%dh %dm", duration/time.Hour, duration%time.Hour/time.Minute)
	default:
		return fmt.Sprintf("%dd %dh", duration/(24*time.Hour), duration%(24*time.Hour)/time.Hour)
	}
}

// ProbeLatency formats non-negative probe latency at millisecond precision.
func ProbeLatency(value model.Optional[time.Duration]) string {
	if !value.Valid || value.Value < 0 {
		return unavailable
	}
	if value.Value < time.Millisecond {
		return "<1 ms"
	}
	return fmt.Sprintf("%d ms", value.Value/time.Millisecond)
}

// LastSeen formats a timestamp relative to an explicitly supplied current time.
func LastSeen(value model.Optional[time.Time], now time.Time) string {
	if !value.Valid {
		return "never"
	}
	age := now.Sub(value.Value)
	if age <= 0 || age < time.Second {
		return "now"
	}
	if age < time.Minute {
		return fmt.Sprintf("%ds", age/time.Second)
	}
	if age < time.Hour {
		return fmt.Sprintf("%dm", age/time.Minute)
	}
	if age < 24*time.Hour {
		return fmt.Sprintf("%dh", age/time.Hour)
	}
	return fmt.Sprintf("%dd", age/(24*time.Hour))
}

func validNumber(value model.Optional[float64]) bool {
	return value.Valid && !math.IsNaN(value.Value) && !math.IsInf(value.Value, 0)
}
