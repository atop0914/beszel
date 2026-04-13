package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/atop0914/beszel/internal/types"
)

// Alert represents a single alert event.
type Alert struct {
	Timestamp time.Time `json:"timestamp"`
	Hostname  string    `json:"hostname"`
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
}

// Manager checks metrics against thresholds and sends notifications.
type Manager struct {
	cpuThreshold    float64
	memoryThreshold float64
	diskThreshold   float64
	webhookURL      string
}

// NewManager creates an alert manager with the given thresholds.
func NewManager(cpu, memory, disk, webhookURL string) *Manager {
	return &Manager{
		cpuThreshold:    parseFloat(cpu, 80),
		memoryThreshold: parseFloat(memory, 85),
		diskThreshold:   parseFloat(disk, 90),
		webhookURL:      webhookURL,
	}
}

func parseFloat(s string, fallback float64) float64 {
	if s == "" {
		return fallback
	}
	var v float64
	_, err := fmt.Sscanf(s, "%f", &v)
	if err != nil {
		return fallback
	}
	return v
}

// Check evaluates metrics against thresholds and sends alerts.
func (m *Manager) Check(metrics *types.Metrics) {
	var activeAlerts []Alert
	ts := time.Now()

	if metrics.CPUPercent > m.cpuThreshold {
		activeAlerts = append(activeAlerts, Alert{
			Timestamp: ts,
			Hostname:  metrics.Hostname,
			Metric:    "cpu",
			Value:     metrics.CPUPercent,
			Threshold: m.cpuThreshold,
		})
	}
	if metrics.MemPercent > m.memoryThreshold {
		activeAlerts = append(activeAlerts, Alert{
			Timestamp: ts,
			Hostname:  metrics.Hostname,
			Metric:    "memory",
			Value:     metrics.MemPercent,
			Threshold: m.memoryThreshold,
		})
	}
	if metrics.DiskPercent > m.diskThreshold {
		activeAlerts = append(activeAlerts, Alert{
			Timestamp: ts,
			Hostname:  metrics.Hostname,
			Metric:    "disk",
			Value:     metrics.DiskPercent,
			Threshold: m.diskThreshold,
		})
	}

	for _, alert := range activeAlerts {
		m.send(alert)
	}
}

func (m *Manager) send(alert Alert) {
	msg := fmt.Sprintf("[ALERT] %s %s threshold exceeded: %.1f%% (threshold: %.1f%%)",
		alert.Hostname, alert.Metric, alert.Value, alert.Threshold)
	log.Println(msg)

	if m.webhookURL != "" {
		go m.webhook(alert)
	}
}

func (m *Manager) webhook(alert Alert) {
	body, err := json.Marshal(alert)
	if err != nil {
		log.Printf("[ALERT] webhook marshal error: %v", err)
		return
	}

	resp, err := http.Post(m.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[ALERT] webhook post error: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("[ALERT] webhook returned status %d", resp.StatusCode)
	}
}
