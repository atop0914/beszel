package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port                int           `yaml:"port"`
	DBPath              string        `yaml:"db_path"`
	CollectionInterval  time.Duration `yaml:"collection_interval"`
	DockerEnabled       bool          `yaml:"docker_enabled"`
	Alerts              AlertsConfig  `yaml:"alerts"`
}

type AlertsConfig struct {
	CPUThreshold    float64 `yaml:"cpu_threshold"`
	MemoryThreshold float64 `yaml:"memory_threshold"`
	DiskThreshold   float64 `yaml:"disk_threshold"`
	WebhookURL      string  `yaml:"webhook_url"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if cfg.CollectionInterval == 0 {
		cfg.CollectionInterval = 30 * time.Second
	}
	if cfg.DBPath == "" {
		cfg.DBPath = "./beszel.db"
	}
	return &cfg, nil
}

func Default() *Config {
	return &Config{
		Port:               8080,
		DBPath:             "./beszel.db",
		CollectionInterval: 30 * time.Second,
		DockerEnabled:      true,
		Alerts: AlertsConfig{
			CPUThreshold:    80,
			MemoryThreshold: 85,
			DiskThreshold:   90,
		},
	}
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}
	if c.DBPath == "" {
		return fmt.Errorf("db_path is required")
	}
	if c.CollectionInterval < time.Second {
		return fmt.Errorf("collection_interval must be at least 1 second")
	}
	if c.Alerts.CPUThreshold < 0 || c.Alerts.CPUThreshold > 100 {
		return fmt.Errorf("alerts.cpu_threshold must be between 0 and 100")
	}
	if c.Alerts.MemoryThreshold < 0 || c.Alerts.MemoryThreshold > 100 {
		return fmt.Errorf("alerts.memory_threshold must be between 0 and 100")
	}
	if c.Alerts.DiskThreshold < 0 || c.Alerts.DiskThreshold > 100 {
		return fmt.Errorf("alerts.disk_threshold must be between 0 and 100")
	}
	return nil
}
