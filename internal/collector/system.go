package collector

import (
	"fmt"
	"time"

	"github.com/atop0914/beszel/internal/types"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

type SystemCollector struct {
	prevNet  map[string]netIOCounters
	hostname string
}

type netIOCounters struct {
	rx uint64
	tx uint64
}

func NewSystemCollector() (*SystemCollector, error) {
	hInfo, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("hostname: %w", err)
	}
	return &SystemCollector{
		prevNet:  make(map[string]netIOCounters),
		hostname: hInfo.Hostname,
	}, nil
}

// Collect gathers system metrics.
func (c *SystemCollector) Collect() (*types.Metrics, error) {
	m := &types.Metrics{Timestamp: time.Now(), Hostname: c.hostname}

	// CPU
	percents, err := cpu.Percent(time.Second, false)
	if err == nil && len(percents) > 0 {
		m.CPUPercent = percents[0]
	}

	// Memory
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		m.MemPercent = memInfo.UsedPercent
		m.MemUsed = memInfo.Used
		m.MemTotal = memInfo.Total
	}

	// Disk
	diskInfo, err := disk.Usage("/")
	if err == nil {
		m.DiskPercent = diskInfo.UsedPercent
		m.DiskUsed = diskInfo.Used
		m.DiskTotal = diskInfo.Total
	}

	// Network delta
	rx, tx := c.networkDelta()
	m.NetworkRX = rx
	m.NetworkTX = tx

	return m, nil
}

func (c *SystemCollector) networkDelta() (rx, tx uint64) {
	counters, err := net.IOCounters(true)
	if err != nil {
		return 0, 0
	}
	for _, cnt := range counters {
		if cnt.Name == "lo" {
			continue
		}
		prev, ok := c.prevNet[cnt.Name]
		if ok {
			rx += cnt.BytesRecv - prev.rx
			tx += cnt.BytesSent - prev.tx
		}
		c.prevNet[cnt.Name] = netIOCounters{rx: cnt.BytesRecv, tx: cnt.BytesSent}
	}
	return rx, tx
}
