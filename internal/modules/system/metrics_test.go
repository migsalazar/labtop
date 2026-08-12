package system

import (
	"errors"
	"math"
	"net"
	"testing"
	"time"
)

func TestCPUCalculator(t *testing.T) {
	t.Parallel()

	calculator := cpuCalculator{}
	baseline := cpuTimes{User: 10, System: 10, Idle: 80}
	if got := calculator.update(baseline, nil); got.Valid {
		t.Fatalf("first CPU reading = %#v, want unavailable", got)
	}
	got := calculator.update(cpuTimes{User: 20, System: 20, Idle: 160}, nil)
	if !got.Valid || math.Abs(got.Value-20) > 0.0001 {
		t.Fatalf("CPU percentage = %#v, want 20", got)
	}
	idle := calculator.update(cpuTimes{User: 20, System: 20, Idle: 260}, nil)
	if !idle.Valid || idle.Value != 0 {
		t.Fatalf("idle CPU percentage = %#v, want 0", idle)
	}
	if reset := calculator.update(baseline, nil); reset.Valid {
		t.Fatalf("reset CPU percentage = %#v, want unavailable", reset)
	}
	if afterReset := calculator.update(cpuTimes{User: 11, System: 11, Idle: 88}, nil); !afterReset.Valid {
		t.Fatalf("CPU did not recover after reset: %#v", afterReset)
	}
	if failed := calculator.update(cpuTimes{}, errors.New("failed")); failed.Valid {
		t.Fatalf("failed CPU = %#v, want unavailable", failed)
	}
	if reprime := calculator.update(baseline, nil); reprime.Valid {
		t.Fatalf("CPU after failure = %#v, want new baseline", reprime)
	}
}

func TestCPUCalculatorRejectsInvalidCounters(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]cpuTimes{
		"negative": {User: -1},
		"NaN":      {User: math.NaN()},
		"infinity": {User: math.Inf(1)},
	} {
		t.Run(name, func(t *testing.T) {
			calculator := cpuCalculator{}
			if got := calculator.update(value, nil); got.Valid {
				t.Fatalf("CPU = %#v, want unavailable", got)
			}
		})
	}
}

func TestNetworkCalculator(t *testing.T) {
	t.Parallel()

	calculator := networkCalculator{}
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	receive, transmit := calculator.update(netCounters{Name: "eth0", BytesRecv: 100, BytesSent: 200}, now)
	if !receive.Valid || receive.Value != 0 || transmit.Value != 0 {
		t.Fatalf("first rates = %#v %#v, want zero", receive, transmit)
	}
	receive, transmit = calculator.update(netCounters{Name: "eth0", BytesRecv: 2148, BytesSent: 1224}, now.Add(2*time.Second))
	if receive.Value != 1024 || transmit.Value != 512 {
		t.Fatalf("rates = %v %v, want 1024 512", receive.Value, transmit.Value)
	}
	for name, current := range map[string]struct {
		counter netCounters
		at      time.Time
	}{
		"counter reset":       {counter: netCounters{Name: "eth0", BytesRecv: 1, BytesSent: 1}, at: now.Add(3 * time.Second)},
		"interface change":    {counter: netCounters{Name: "wlan0", BytesRecv: 500, BytesSent: 500}, at: now.Add(4 * time.Second)},
		"nonpositive elapsed": {counter: netCounters{Name: "wlan0", BytesRecv: 600, BytesSent: 600}, at: now.Add(4 * time.Second)},
	} {
		t.Run(name, func(t *testing.T) {
			receive, transmit := calculator.update(current.counter, current.at)
			if !receive.Valid || receive.Value != 0 || transmit.Value != 0 {
				t.Fatalf("reset rates = %#v %#v, want zero", receive, transmit)
			}
		})
	}
}

func TestSelectInterface(t *testing.T) {
	t.Parallel()

	address := &net.IPNet{IP: net.ParseIP("192.0.2.10"), Mask: net.CIDRMask(24, 32)}
	interfaces := []interfaceInfo{
		{Name: "zeta", Flags: net.FlagUp, Addrs: []net.Addr{address}},
		{Name: "alpha", Flags: net.FlagUp, Addrs: []net.Addr{address}},
		{Name: "lo", Flags: net.FlagUp | net.FlagLoopback, Addrs: []net.Addr{&net.IPNet{IP: net.ParseIP("127.0.0.1"), Mask: net.CIDRMask(8, 32)}}},
		{Name: "down", Addrs: []net.Addr{address}},
	}

	if selected, ok := selectInterface("zeta", interfaces, "alpha"); !ok || selected.Name != "zeta" {
		t.Fatalf("configured selection = %#v %t", selected, ok)
	}
	if _, ok := selectInterface("missing", interfaces, "alpha"); ok {
		t.Fatal("missing configured interface fell back")
	}
	if selected, ok := selectInterface("", interfaces, "zeta"); !ok || selected.Name != "zeta" {
		t.Fatalf("default-route selection = %#v %t", selected, ok)
	}
	if selected, ok := selectInterface("", interfaces, "missing"); !ok || selected.Name != "alpha" {
		t.Fatalf("deterministic fallback = %#v %t", selected, ok)
	}
	if _, ok := selectInterface("", interfaces[2:], "missing"); ok {
		t.Fatal("invalid interfaces produced a selection")
	}
}

func TestSelectTemperature(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values []temperature
		want   float64
		valid  bool
	}{
		{name: "cpu thermal priority", values: []temperature{{Key: "CPU package", Value: 60}, {Key: "cpu_thermal", Value: 50}}, want: 50, valid: true},
		{name: "coretemp priority", values: []temperature{{Key: "CPU package", Value: 60}, {Key: "CORETEMP", Value: 55}}, want: 55, valid: true},
		{name: "CPU fallback", values: []temperature{{Key: "gpu", Value: 40}, {Key: "CPU package", Value: 60}}, want: 60, valid: true},
		{name: "nonfinite ignored", values: []temperature{{Key: "cpu_thermal", Value: math.NaN()}}, valid: false},
		{name: "unrelated", values: []temperature{{Key: "gpu", Value: 40}}, valid: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := selectTemperature(test.values)
			if got.Valid != test.valid || got.Value != test.want {
				t.Fatalf("temperature = %#v, want valid=%t value=%v", got, test.valid, test.want)
			}
		})
	}
}

func TestValueValidation(t *testing.T) {
	t.Parallel()

	if got := validPercentage(50, nil); !got.Valid || got.Value != 50 {
		t.Fatalf("percentage = %#v", got)
	}
	for _, value := range []float64{-1, 101, math.NaN(), math.Inf(1)} {
		if got := validPercentage(value, nil); got.Valid {
			t.Fatalf("validPercentage(%v) = %#v, want unavailable", value, got)
		}
	}
	if got := validUptime(42, nil); !got.Valid || got.Value != 42*time.Second {
		t.Fatalf("uptime = %#v", got)
	}
	if got := validUptime(math.MaxUint64, nil); got.Valid {
		t.Fatalf("overflow uptime = %#v, want unavailable", got)
	}
}
