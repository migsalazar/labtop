package system

import (
	"context"
	"fmt"
	"net"

	gocpu "github.com/shirou/gopsutil/v4/cpu"
	godisk "github.com/shirou/gopsutil/v4/disk"
	gohost "github.com/shirou/gopsutil/v4/host"
	gomem "github.com/shirou/gopsutil/v4/mem"
	gonet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"
)

type cpuTimes struct {
	User, System, Idle, Nice, IOWait, IRQ, SoftIRQ, Steal float64
}

type netCounters struct {
	Name      string
	BytesSent uint64
	BytesRecv uint64
}

type temperature struct {
	Key   string
	Value float64
}

type interfaceInfo struct {
	Name  string
	Flags net.Flags
	Addrs []net.Addr
}

type sources struct {
	cpuTimes     func(context.Context) (cpuTimes, error)
	memory       func(context.Context) (float64, error)
	disk         func(context.Context) (float64, error)
	temperatures func(context.Context) ([]temperature, error)
	uptime       func(context.Context) (uint64, error)
	netCounters  func(context.Context) ([]netCounters, error)
	interfaces   func() ([]interfaceInfo, error)
	defaultRoute func() (string, error)
}

func systemSources() sources {
	return sources{
		cpuTimes: func(ctx context.Context) (cpuTimes, error) {
			stats, err := gocpu.TimesWithContext(ctx, false)
			if err != nil {
				return cpuTimes{}, err
			}
			if len(stats) == 0 {
				return cpuTimes{}, fmt.Errorf("CPU times are unavailable")
			}
			stat := stats[0]
			return cpuTimes{
				User: stat.User, System: stat.System, Idle: stat.Idle,
				Nice: stat.Nice, IOWait: stat.Iowait, IRQ: stat.Irq,
				SoftIRQ: stat.Softirq, Steal: stat.Steal,
			}, nil
		},
		memory: func(ctx context.Context) (float64, error) {
			stat, err := gomem.VirtualMemoryWithContext(ctx)
			if err != nil {
				return 0, err
			}
			return stat.UsedPercent, nil
		},
		disk: func(ctx context.Context) (float64, error) {
			stat, err := godisk.UsageWithContext(ctx, "/")
			if err != nil {
				return 0, err
			}
			return stat.UsedPercent, nil
		},
		temperatures: func(ctx context.Context) ([]temperature, error) {
			stats, err := sensors.TemperaturesWithContext(ctx)
			result := make([]temperature, 0, len(stats))
			for _, stat := range stats {
				result = append(result, temperature{Key: stat.SensorKey, Value: stat.Temperature})
			}
			return result, err
		},
		uptime: gohost.UptimeWithContext,
		netCounters: func(ctx context.Context) ([]netCounters, error) {
			stats, err := gonet.IOCountersWithContext(ctx, true)
			result := make([]netCounters, 0, len(stats))
			for _, stat := range stats {
				result = append(result, netCounters{Name: stat.Name, BytesSent: stat.BytesSent, BytesRecv: stat.BytesRecv})
			}
			return result, err
		},
		interfaces:   listInterfaces,
		defaultRoute: defaultRouteInterface,
	}
}

func listInterfaces() ([]interfaceInfo, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	result := make([]interfaceInfo, 0, len(interfaces))
	for _, item := range interfaces {
		addresses, addressErr := item.Addrs()
		if addressErr != nil {
			addresses = nil
		}
		result = append(result, interfaceInfo{Name: item.Name, Flags: item.Flags, Addrs: addresses})
	}
	return result, nil
}
