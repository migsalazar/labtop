package system

import (
	"math"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/migsalazar/labtop/internal/model"
)

type cpuCalculator struct {
	previous cpuTimes
	primed   bool
}

func (calculator *cpuCalculator) update(current cpuTimes, err error) model.Optional[float64] {
	if err != nil || !validCPUTimes(current) {
		calculator.primed = false
		return model.None[float64]()
	}
	if !calculator.primed {
		calculator.previous = current
		calculator.primed = true
		return model.None[float64]()
	}
	previous := calculator.previous
	calculator.previous = current
	if cpuDecreased(current, previous) {
		return model.None[float64]()
	}
	totalDelta := cpuTotal(current) - cpuTotal(previous)
	idleDelta := current.Idle + current.IOWait - previous.Idle - previous.IOWait
	if !finite(totalDelta) || !finite(idleDelta) || totalDelta <= 0 || idleDelta < 0 {
		return model.None[float64]()
	}
	percentage := (totalDelta - idleDelta) / totalDelta * 100
	if !finite(percentage) {
		return model.None[float64]()
	}
	return model.Some(max(0, min(100, percentage)))
}

func cpuTotal(value cpuTimes) float64 {
	return value.User + value.System + value.Idle + value.Nice + value.IOWait + value.IRQ + value.SoftIRQ + value.Steal
}

func validCPUTimes(value cpuTimes) bool {
	values := [...]float64{value.User, value.System, value.Idle, value.Nice, value.IOWait, value.IRQ, value.SoftIRQ, value.Steal}
	for _, number := range values {
		if !finite(number) || number < 0 {
			return false
		}
	}
	return true
}

func cpuDecreased(current, previous cpuTimes) bool {
	return current.User < previous.User || current.System < previous.System || current.Idle < previous.Idle ||
		current.Nice < previous.Nice || current.IOWait < previous.IOWait || current.IRQ < previous.IRQ ||
		current.SoftIRQ < previous.SoftIRQ || current.Steal < previous.Steal
}

type networkCalculator struct {
	previous  netCounters
	collected time.Time
	primed    bool
}

func (calculator *networkCalculator) update(current netCounters, collected time.Time) (model.Optional[float64], model.Optional[float64]) {
	if !calculator.primed || calculator.previous.Name != current.Name || !collected.After(calculator.collected) ||
		current.BytesRecv < calculator.previous.BytesRecv || current.BytesSent < calculator.previous.BytesSent {
		calculator.previous = current
		calculator.collected = collected
		calculator.primed = true
		return model.Some(0.0), model.Some(0.0)
	}
	seconds := collected.Sub(calculator.collected).Seconds()
	receive := float64(current.BytesRecv-calculator.previous.BytesRecv) / seconds
	transmit := float64(current.BytesSent-calculator.previous.BytesSent) / seconds
	calculator.previous = current
	calculator.collected = collected
	if !finite(receive) || !finite(transmit) || receive < 0 || transmit < 0 {
		return model.Some(0.0), model.Some(0.0)
	}
	return model.Some(receive), model.Some(transmit)
}

func selectInterface(configured string, interfaces []interfaceInfo, defaultRoute string) (interfaceInfo, bool) {
	qualifying := make(map[string]interfaceInfo, len(interfaces))
	for _, item := range interfaces {
		if qualifies(item) {
			qualifying[item.Name] = item
		}
	}
	if configured != "" {
		item, ok := qualifying[configured]
		return item, ok
	}
	if item, ok := qualifying[defaultRoute]; ok {
		return item, true
	}
	names := make([]string, 0, len(qualifying))
	for name := range qualifying {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return interfaceInfo{}, false
	}
	return qualifying[names[0]], true
}

func qualifies(item interfaceInfo) bool {
	if item.Name == "" || item.Flags&net.FlagUp == 0 || item.Flags&net.FlagLoopback != 0 {
		return false
	}
	for _, address := range item.Addrs {
		var ip net.IP
		switch value := address.(type) {
		case *net.IPNet:
			ip = value.IP
		case *net.IPAddr:
			ip = value.IP
		default:
			text := address.String()
			if host, _, err := net.SplitHostPort(text); err == nil {
				text = host
			} else if slash := strings.IndexByte(text, '/'); slash >= 0 {
				text = text[:slash]
			}
			ip = net.ParseIP(text)
		}
		if ip != nil && !ip.IsLoopback() && !ip.IsUnspecified() {
			return true
		}
	}
	return false
}

func findCounters(interfaceName string, counters []netCounters) (netCounters, bool) {
	for _, counter := range counters {
		if counter.Name == interfaceName {
			return counter, true
		}
	}
	return netCounters{}, false
}

func selectTemperature(values []temperature) model.Optional[float64] {
	for _, preferred := range []string{"cpu_thermal", "coretemp"} {
		for _, value := range values {
			if strings.EqualFold(value.Key, preferred) && finite(value.Value) {
				return model.Some(value.Value)
			}
		}
	}
	for _, value := range values {
		if strings.Contains(strings.ToLower(value.Key), "cpu") && finite(value.Value) {
			return model.Some(value.Value)
		}
	}
	return model.None[float64]()
}

func validPercentage(value float64, err error) model.Optional[float64] {
	if err != nil || !finite(value) || value < 0 || value > 100 {
		return model.None[float64]()
	}
	return model.Some(value)
}

func validUptime(seconds uint64, err error) model.Optional[time.Duration] {
	if err != nil || seconds > uint64(math.MaxInt64/int64(time.Second)) {
		return model.None[time.Duration]()
	}
	return model.Some(time.Duration(seconds) * time.Second)
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
