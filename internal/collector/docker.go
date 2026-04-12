package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/atop0914/beszel/internal/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerCollector struct {
	client *client.Client
}

func NewDockerCollector() (*DockerCollector, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}
	return &DockerCollector{client: cli}, nil
}

// CollectContainerStats gathers metrics for all running containers.
func (d *DockerCollector) CollectContainerStats() ([]*types.ContainerMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	containers, err := d.client.ContainerList(ctx, container.ListOptions{All: false})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	var results []*types.ContainerMetrics

	for _, c := range containers {
		stats, err := d.client.ContainerStats(ctx, c.ID, false)
		if err != nil {
			log.Printf("container stats for %s: %v", c.ID[:12], err)
			continue
		}

		// Decode the stats body (StatsResponse embeds Stats)
		var sr container.StatsResponse
		if err := json.NewDecoder(stats.Body).Decode(&sr); err != nil {
			stats.Body.Close()
			// Fallback: use zero values
			results = append(results, &types.ContainerMetrics{
				Timestamp:     time.Now(),
				ContainerID:   c.ID,
				ContainerName: c.Names[0],
				CPUPercent:    0,
				MemPercent:    0,
				MemUsage:      0,
			})
			continue
		}
		stats.Body.Close()

		cpuPercent := calculateCPUPercent(&sr.Stats)
		memPercent := 0.0
		memUsage := uint64(0)
		if sr.MemoryStats.Limit > 0 {
			memPercent = float64(sr.MemoryStats.Usage) / float64(sr.MemoryStats.Limit) * 100
			memUsage = sr.MemoryStats.Usage
		}

		results = append(results, &types.ContainerMetrics{
			Timestamp:     time.Now(),
			ContainerID:   c.ID,
			ContainerName: c.Names[0],
			CPUPercent:    cpuPercent,
			MemPercent:    memPercent,
			MemUsage:      memUsage,
		})
	}

	return results, nil
}

// calculateCPUPercent calculates CPU usage percentage from container stats.
func calculateCPUPercent(stats *container.Stats) float64 {
	// CPU percentage = (delta cpu usage / delta system cpu) * number of CPUs * 100
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)

	if systemDelta > 0 && cpuDelta > 0 {
		numCPUs := float64(stats.CPUStats.OnlineCPUs)
		if numCPUs == 0 {
			numCPUs = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
		}
		if numCPUs == 0 {
			numCPUs = 1
		}
		return (cpuDelta / systemDelta) * numCPUs * 100
	}
	return 0
}

// Close closes the Docker client connection.
func (d *DockerCollector) Close() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}
