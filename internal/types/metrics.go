package types

import "time"

type Metrics struct {
	Timestamp   time.Time
	Hostname    string
	CPUPercent  float64
	MemPercent  float64
	MemUsed     uint64
	MemTotal    uint64
	DiskPercent float64
	DiskUsed    uint64
	DiskTotal   uint64
	NetworkRX   uint64
	NetworkTX   uint64

	// System info
	Uptime   uint64    // seconds
	OS       string    // e.g. "linux"
	Platform string    // e.g. "ubuntu"
	Kernel   string    // e.g. "6.8.0-94-generic"
	Load1    float64   // 1-min load average
	Load5    float64   // 5-min load average
	Load15   float64   // 15-min load average
}

type ContainerMetrics struct {
	Timestamp     time.Time
	ContainerID   string
	ContainerName string
	CPUPercent    float64
	MemPercent    float64
	MemUsage      uint64
}
