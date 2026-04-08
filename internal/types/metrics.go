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
}

type ContainerMetrics struct {
	Timestamp     time.Time
	ContainerID   string
	ContainerName string
	CPUPercent    float64
	MemPercent    float64
	MemUsage      uint64
}
