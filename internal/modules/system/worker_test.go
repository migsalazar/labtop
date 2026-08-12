package system

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/migsalazar/labtop/internal/config"
	"github.com/migsalazar/labtop/internal/model"
	moduletypes "github.com/migsalazar/labtop/internal/modules"
)

type fakeTicker struct {
	channel chan time.Time
	stopped atomic.Bool
}

func (ticker *fakeTicker) C() <-chan time.Time { return ticker.channel }
func (ticker *fakeTicker) Stop()               { ticker.stopped.Store(true) }

func TestWorkerImmediateCollectionAndSampling(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	cpuCall := 0
	counterCall := 0
	source := successfulSources()
	source.cpuTimes = func(context.Context) (cpuTimes, error) {
		cpuCall++
		if cpuCall == 1 {
			return cpuTimes{User: 10, System: 10, Idle: 80}, nil
		}
		return cpuTimes{User: 20, System: 20, Idle: 160}, nil
	}
	source.netCounters = func(context.Context) ([]netCounters, error) {
		counterCall++
		return []netCounters{{Name: "eth0", BytesRecv: uint64(counterCall * 1024), BytesSent: uint64(counterCall * 512)}}, nil
	}

	tickers := []*fakeTicker{
		{channel: make(chan time.Time, 1)},
		{channel: make(chan time.Time, 1)},
		{channel: make(chan time.Time, 1)},
	}
	nextTicker := 0
	worker := newWorker(systemModule(), source, func() time.Time {
		result := now
		now = now.Add(time.Second)
		return result
	}, func(time.Duration) ticker {
		result := tickers[nextTicker]
		nextTicker++
		return result
	})

	ctx, cancel := context.WithCancel(context.Background())
	updates := make(chan model.ModuleUpdate, 8)
	done := make(chan struct{})
	go func() {
		worker.Run(ctx, updates)
		close(done)
	}()

	interfaceUpdate := receiveUpdate[model.LocalInterfaceUpdate](t, updates)
	if interfaceUpdate.Snapshot.Status != model.LocalInterfaceAvailable || interfaceUpdate.Snapshot.InterfaceName != "eth0" {
		t.Fatalf("interface update = %#v", interfaceUpdate)
	}
	initial := receiveUpdate[model.SystemUpdate](t, updates)
	if initial.ModuleID != "system" || initial.Snapshot.Status != model.ModuleReady || initial.Snapshot.CPUPercent.Valid {
		t.Fatalf("initial update = %#v", initial)
	}
	if initial.Snapshot.MemoryPercent.Value != 25 || initial.Snapshot.DiskPercent.Value != 40 || initial.Snapshot.TemperatureCelsius.Value != 50 || initial.Snapshot.Uptime.Value != time.Hour {
		t.Fatalf("initial snapshot = %#v", initial.Snapshot)
	}

	tickers[0].channel <- now
	receiveUpdate[model.LocalInterfaceUpdate](t, updates)
	fast := receiveUpdate[model.SystemUpdate](t, updates)
	if !fast.Snapshot.CPUPercent.Valid || fast.Snapshot.CPUPercent.Value != 20 {
		t.Fatalf("fast CPU = %#v, want 20", fast.Snapshot.CPUPercent)
	}
	if fast.Snapshot.NetworkReceiveBytesPerSecond.Value != 1024 || fast.Snapshot.NetworkTransmitBytesPerSecond.Value != 512 {
		t.Fatalf("network rates = %#v %#v", fast.Snapshot.NetworkReceiveBytesPerSecond, fast.Snapshot.NetworkTransmitBytesPerSecond)
	}

	sampledAt := now.Add(time.Minute)
	tickers[2].channel <- sampledAt
	sample := receiveUpdate[model.SystemSampleUpdate](t, updates)
	if sample.ModuleID != "system" || !sample.CPUPercent.Valid || sample.SampledAt != sampledAt {
		t.Fatalf("sample update = %#v", sample)
	}

	cancel()
	waitDone(t, done)
	for index, ticker := range tickers {
		if !ticker.stopped.Load() {
			t.Fatalf("ticker %d was not stopped", index)
		}
	}
	updates <- model.SystemUpdate{}
}

func TestWorkerPartialFailuresRemainReady(t *testing.T) {
	t.Parallel()

	source := failingSources()
	source.memory = func(context.Context) (float64, error) { return 25, nil }
	worker := workerWithIdleTickers(source)
	ctx, cancel := context.WithCancel(context.Background())
	updates := make(chan model.ModuleUpdate, 2)
	done := make(chan struct{})
	go func() {
		worker.Run(ctx, updates)
		close(done)
	}()

	interfaceUpdate := receiveUpdate[model.LocalInterfaceUpdate](t, updates)
	if interfaceUpdate.Snapshot.Status != model.LocalInterfaceUnavailable {
		t.Fatalf("interface status = %q, want unavailable", interfaceUpdate.Snapshot.Status)
	}
	update := receiveUpdate[model.SystemUpdate](t, updates)
	if update.Snapshot.Status != model.ModuleReady || !update.Snapshot.MemoryPercent.Valid {
		t.Fatalf("partial snapshot = %#v", update.Snapshot)
	}
	if update.Snapshot.CPUPercent.Valid || update.Snapshot.DiskPercent.Valid || update.Snapshot.TemperatureCelsius.Valid || update.Snapshot.Uptime.Valid {
		t.Fatalf("failed metrics became available: %#v", update.Snapshot)
	}
	cancel()
	waitDone(t, done)
}

func TestWorkerAllFailuresAreUnavailable(t *testing.T) {
	t.Parallel()

	worker := workerWithIdleTickers(failingSources())
	ctx, cancel := context.WithCancel(context.Background())
	updates := make(chan model.ModuleUpdate, 2)
	done := make(chan struct{})
	go func() {
		worker.Run(ctx, updates)
		close(done)
	}()
	receiveUpdate[model.LocalInterfaceUpdate](t, updates)
	update := receiveUpdate[model.SystemUpdate](t, updates)
	if update.Snapshot.Status != model.ModuleUnavailable {
		t.Fatalf("status = %q, want unavailable", update.Snapshot.Status)
	}
	cancel()
	waitDone(t, done)
}

func TestWorkerCancellationUnblocksFullChannel(t *testing.T) {
	t.Parallel()

	tickers := []*fakeTicker{
		{channel: make(chan time.Time)},
		{channel: make(chan time.Time)},
		{channel: make(chan time.Time)},
	}
	index := 0
	worker := newWorker(systemModule(), successfulSources(), time.Now, func(time.Duration) ticker {
		result := tickers[index]
		index++
		return result
	})
	ctx, cancel := context.WithCancel(context.Background())
	updates := make(chan model.ModuleUpdate)
	done := make(chan struct{})
	go func() {
		worker.Run(ctx, updates)
		close(done)
	}()

	cancel()
	waitDone(t, done)
	for tickerIndex, ticker := range tickers {
		if !ticker.stopped.Load() {
			t.Fatalf("ticker %d was not stopped", tickerIndex)
		}
	}
}

func TestNewWorkerRejectsMissingSettingsAndIntervals(t *testing.T) {
	t.Parallel()

	wrongType := systemModule()
	if _, err := NewWorker(wrongType); err == nil {
		t.Fatal("NewWorker accepted a non-system module")
	}
	missingSettings := systemModule()
	missingSettings.Type = moduletypes.TypeSystem
	missingSettings.System = nil
	if _, err := NewWorker(missingSettings); err == nil {
		t.Fatal("NewWorker accepted missing settings")
	}
	module := systemModule()
	module.Type = moduletypes.TypeSystem
	module.Refresh = 0
	if _, err := NewWorker(module); err == nil {
		t.Fatal("NewWorker accepted invalid intervals")
	}
}

func successfulSources() sources {
	address := &net.IPNet{IP: net.ParseIP("192.0.2.10"), Mask: net.CIDRMask(24, 32)}
	return sources{
		cpuTimes: func(context.Context) (cpuTimes, error) { return cpuTimes{User: 10, System: 10, Idle: 80}, nil },
		memory:   func(context.Context) (float64, error) { return 25, nil },
		disk:     func(context.Context) (float64, error) { return 40, nil },
		temperatures: func(context.Context) ([]temperature, error) {
			return []temperature{{Key: "cpu_thermal", Value: 50}}, nil
		},
		uptime: func(context.Context) (uint64, error) { return 3600, nil },
		netCounters: func(context.Context) ([]netCounters, error) {
			return []netCounters{{Name: "eth0", BytesRecv: 1024, BytesSent: 512}}, nil
		},
		interfaces: func() ([]interfaceInfo, error) {
			return []interfaceInfo{{Name: "eth0", Flags: net.FlagUp, Addrs: []net.Addr{address}}}, nil
		},
		defaultRoute: func() (string, error) { return "eth0", nil },
	}
}

func failingSources() sources {
	failure := errors.New("unavailable")
	return sources{
		cpuTimes:     func(context.Context) (cpuTimes, error) { return cpuTimes{}, failure },
		memory:       func(context.Context) (float64, error) { return 0, failure },
		disk:         func(context.Context) (float64, error) { return 0, failure },
		temperatures: func(context.Context) ([]temperature, error) { return nil, failure },
		uptime:       func(context.Context) (uint64, error) { return 0, failure },
		netCounters:  func(context.Context) ([]netCounters, error) { return nil, failure },
		interfaces:   func() ([]interfaceInfo, error) { return nil, failure },
		defaultRoute: func() (string, error) { return "", failure },
	}
}

func systemModule() config.Module {
	return config.Module{
		ID: "system", Refresh: time.Second,
		System: &config.SystemSettings{
			SlowRefresh: time.Second, SparklineSampleInterval: time.Second,
			TemperatureWarningCelsius: 80, DiskWarningPercent: 90,
		},
	}
}

func workerWithIdleTickers(source sources) *Worker {
	return newWorker(systemModule(), source, time.Now, func(time.Duration) ticker {
		return &fakeTicker{channel: make(chan time.Time)}
	})
}

func receiveUpdate[T model.ModuleUpdate](t *testing.T, updates <-chan model.ModuleUpdate) T {
	t.Helper()
	select {
	case update := <-updates:
		result, ok := update.(T)
		if !ok {
			t.Fatalf("update type = %T", update)
		}
		return result
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for update")
		var zero T
		return zero
	}
}

func waitDone(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
}
