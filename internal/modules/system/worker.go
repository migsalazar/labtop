package system

import (
	"context"
	"fmt"
	"time"

	"github.com/migsalazar/labtop/internal/config"
	"github.com/migsalazar/labtop/internal/model"
	moduletypes "github.com/migsalazar/labtop/internal/modules"
)

type ticker interface {
	C() <-chan time.Time
	Stop()
}

type realTicker struct{ *time.Ticker }

func (value realTicker) C() <-chan time.Time { return value.Ticker.C }

type tickerFactory func(time.Duration) ticker

// Worker collects local system evidence without owning presentation state.
type Worker struct {
	moduleID  string
	settings  config.SystemSettings
	fast      time.Duration
	sources   sources
	now       func() time.Time
	newTicker tickerFactory

	cpu         cpuCalculator
	network     networkCalculator
	latest      model.SystemSnapshot
	sequence    uint64
	temperature thresholdCondition
	disk        thresholdCondition
}

var _ moduletypes.Worker = (*Worker)(nil)

// NewWorker constructs a system worker from validated module configuration.
func NewWorker(module config.Module) (*Worker, error) {
	if module.Type != moduletypes.TypeSystem {
		return nil, fmt.Errorf("module %q has type %q, want %q", module.ID, module.Type, moduletypes.TypeSystem)
	}
	if module.System == nil {
		return nil, fmt.Errorf("system module %q has no system settings", module.ID)
	}
	if module.Refresh <= 0 || module.System.SlowRefresh <= 0 || module.System.SparklineSampleInterval <= 0 {
		return nil, fmt.Errorf("system module %q has invalid collection intervals", module.ID)
	}
	return newWorker(module, systemSources(), time.Now, func(interval time.Duration) ticker {
		return realTicker{time.NewTicker(interval)}
	}), nil
}

func newWorker(module config.Module, source sources, now func() time.Time, tickers tickerFactory) *Worker {
	return &Worker{
		moduleID: module.ID, settings: *module.System, fast: module.Refresh,
		sources: source, now: now, newTicker: tickers,
		latest: model.SystemSnapshot{Status: model.ModuleUnavailable},
		temperature: thresholdCondition{
			eventType: "system.temperature_warning", recoveryType: "system.temperature_recovered",
			sourceID: "temperature", warningTitle: "TEMPERATURE WARNING", recoveryTitle: "TEMPERATURE RECOVERED",
			unit: "°C", threshold: module.System.TemperatureWarningCelsius,
		},
		disk: thresholdCondition{
			eventType: "system.disk_warning", recoveryType: "system.disk_recovered",
			sourceID: "root-disk", warningTitle: "ROOT DISK WARNING", recoveryTitle: "ROOT DISK RECOVERED",
			unit: "%", threshold: module.System.DiskWarningPercent,
		},
	}
}

// Run performs immediate collection, then owns all periodic scheduling until cancellation.
func (worker *Worker) Run(ctx context.Context, updates chan<- model.ModuleUpdate) {
	fastTicker := worker.newTicker(worker.fast)
	slowTicker := worker.newTicker(worker.settings.SlowRefresh)
	sampleTicker := worker.newTicker(worker.settings.SparklineSampleInterval)
	defer fastTicker.Stop()
	defer slowTicker.Stop()
	defer sampleTicker.Stop()

	if !worker.collectAndSend(ctx, updates, true) {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-fastTicker.C():
			if !worker.collectAndSend(ctx, updates, false) {
				return
			}
		case <-slowTicker.C():
			if !worker.collectSlowAndSend(ctx, updates) {
				return
			}
		case sampledAt := <-sampleTicker.C():
			if !send(ctx, updates, model.SystemSampleUpdate{
				ModuleID: worker.moduleID, SampledAt: sampledAt,
				CPUPercent: worker.latest.CPUPercent, MemoryPercent: worker.latest.MemoryPercent,
			}) {
				return
			}
		}
	}
}

func (worker *Worker) collectAndSend(ctx context.Context, updates chan<- model.ModuleUpdate, includeSlow bool) bool {
	collectedAt := worker.now()
	interfaceSnapshot := worker.collectFast(ctx, collectedAt)
	events := []model.ConsoleEvent(nil)
	if includeSlow {
		events = worker.collectSlow(ctx, collectedAt)
	}
	if !send(ctx, updates, model.LocalInterfaceUpdate{Snapshot: interfaceSnapshot}) {
		return false
	}
	return send(ctx, updates, model.SystemUpdate{ModuleID: worker.moduleID, Snapshot: worker.latest, Events: events})
}

func (worker *Worker) collectSlowAndSend(ctx context.Context, updates chan<- model.ModuleUpdate) bool {
	collectedAt := worker.now()
	events := worker.collectSlow(ctx, collectedAt)
	worker.latest.CollectedAt = collectedAt
	worker.updateStatus()
	return send(ctx, updates, model.SystemUpdate{ModuleID: worker.moduleID, Snapshot: worker.latest, Events: events})
}

func (worker *Worker) collectFast(ctx context.Context, collectedAt time.Time) model.LocalInterfaceSnapshot {
	times, cpuErr := worker.sources.cpuTimes(ctx)
	worker.latest.CPUPercent = worker.cpu.update(times, cpuErr)
	memory, memoryErr := worker.sources.memory(ctx)
	worker.latest.MemoryPercent = validPercentage(memory, memoryErr)
	uptime, uptimeErr := worker.sources.uptime(ctx)
	worker.latest.Uptime = validUptime(uptime, uptimeErr)
	interfaceSnapshot := worker.collectNetwork(ctx, collectedAt)
	worker.latest.CollectedAt = collectedAt
	worker.updateStatus()
	return interfaceSnapshot
}

func (worker *Worker) collectNetwork(ctx context.Context, collectedAt time.Time) model.LocalInterfaceSnapshot {
	interfaces, err := worker.sources.interfaces()
	defaultRoute := ""
	if worker.settings.NetworkInterface == "" {
		defaultRoute, _ = worker.sources.defaultRoute()
	}
	selected, available := selectInterface(worker.settings.NetworkInterface, interfaces, defaultRoute)
	interfaceSnapshot := model.LocalInterfaceSnapshot{CheckedAt: collectedAt, Status: model.LocalInterfaceUnavailable}
	if err == nil && available {
		interfaceSnapshot.Status = model.LocalInterfaceAvailable
		interfaceSnapshot.InterfaceName = selected.Name
		counters, countersErr := worker.sources.netCounters(ctx)
		if counter, ok := findCounters(selected.Name, counters); countersErr == nil && ok {
			worker.latest.NetworkReceiveBytesPerSecond, worker.latest.NetworkTransmitBytesPerSecond = worker.network.update(counter, collectedAt)
		} else {
			worker.network.primed = false
			worker.latest.NetworkReceiveBytesPerSecond = model.None[float64]()
			worker.latest.NetworkTransmitBytesPerSecond = model.None[float64]()
		}
	} else {
		worker.network.primed = false
		worker.latest.NetworkReceiveBytesPerSecond = model.None[float64]()
		worker.latest.NetworkTransmitBytesPerSecond = model.None[float64]()
	}
	return interfaceSnapshot
}

func (worker *Worker) collectSlow(ctx context.Context, collectedAt time.Time) []model.ConsoleEvent {
	disk, diskErr := worker.sources.disk(ctx)
	worker.latest.DiskPercent = validPercentage(disk, diskErr)
	temperatures, temperatureErr := worker.sources.temperatures(ctx)
	if temperatureErr != nil {
		worker.latest.TemperatureCelsius = model.None[float64]()
	} else {
		worker.latest.TemperatureCelsius = selectTemperature(temperatures)
	}
	events := worker.temperature.update(worker.moduleID, worker.latest.TemperatureCelsius, collectedAt, worker.nextEventID)
	events = append(events, worker.disk.update(worker.moduleID, worker.latest.DiskPercent, collectedAt, worker.nextEventID)...)
	worker.latest.CollectedAt = collectedAt
	worker.updateStatus()
	return events
}

func (worker *Worker) updateStatus() {
	if worker.latest.CPUPercent.Valid || worker.latest.MemoryPercent.Valid || worker.latest.DiskPercent.Valid ||
		worker.latest.TemperatureCelsius.Valid || worker.latest.NetworkReceiveBytesPerSecond.Valid ||
		worker.latest.NetworkTransmitBytesPerSecond.Valid || worker.latest.Uptime.Valid {
		worker.latest.Status = model.ModuleReady
		return
	}
	worker.latest.Status = model.ModuleUnavailable
}

func (worker *Worker) nextEventID(eventType string) string {
	worker.sequence++
	return fmt.Sprintf("%s:%s:%d", worker.moduleID, eventType, worker.sequence)
}

func send(ctx context.Context, updates chan<- model.ModuleUpdate, update model.ModuleUpdate) bool {
	select {
	case updates <- update:
		return true
	case <-ctx.Done():
		return false
	}
}
