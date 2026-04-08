package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/atop0914/beszel/internal/types"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		hostname TEXT NOT NULL,
		cpu_percent REAL NOT NULL,
		memory_percent REAL NOT NULL,
		memory_used INTEGER NOT NULL,
		memory_total INTEGER NOT NULL,
		disk_percent REAL NOT NULL,
		disk_used INTEGER NOT NULL,
		disk_total INTEGER NOT NULL,
		network_rx INTEGER NOT NULL,
		network_tx INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp);
	CREATE TABLE IF NOT EXISTS containers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		container_id TEXT NOT NULL,
		container_name TEXT NOT NULL,
		cpu_percent REAL NOT NULL,
		memory_percent REAL NOT NULL,
		memory_usage INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_containers_timestamp ON containers(timestamp);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) InsertMetrics(m *types.Metrics) error {
	_, err := s.db.Exec(`
		INSERT INTO metrics (timestamp, hostname, cpu_percent, memory_percent,
			memory_used, memory_total, disk_percent, disk_used, disk_total,
			network_rx, network_tx)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Timestamp, m.Hostname, m.CPUPercent, m.MemPercent,
		m.MemUsed, m.MemTotal, m.DiskPercent, m.DiskUsed, m.DiskTotal,
		m.NetworkRX, m.NetworkTX,
	)
	return err
}

func (s *Store) LatestMetrics() (*types.Metrics, error) {
	row := s.db.QueryRow(`
		SELECT timestamp, hostname, cpu_percent, memory_percent,
			memory_used, memory_total, disk_percent, disk_used, disk_total,
			network_rx, network_tx
		FROM metrics ORDER BY timestamp DESC LIMIT 1`)
	var m types.Metrics
	err := row.Scan(&m.Timestamp, &m.Hostname, &m.CPUPercent, &m.MemPercent,
		&m.MemUsed, &m.MemTotal, &m.DiskPercent, &m.DiskUsed, &m.DiskTotal,
		&m.NetworkRX, &m.NetworkTX)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Store) MetricsHistory(hours int) ([]*types.Metrics, error) {
	rows, err := s.db.Query(`
		SELECT timestamp, hostname, cpu_percent, memory_percent,
			memory_used, memory_total, disk_percent, disk_used, disk_total,
			network_rx, network_tx
		FROM metrics
		WHERE timestamp > datetime('now', '-' || ? || ' hours')
		ORDER BY timestamp ASC`, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []*types.Metrics
	for rows.Next() {
		var m types.Metrics
		err := rows.Scan(&m.Timestamp, &m.Hostname, &m.CPUPercent, &m.MemPercent,
			&m.MemUsed, &m.MemTotal, &m.DiskPercent, &m.DiskUsed, &m.DiskTotal,
			&m.NetworkRX, &m.NetworkTX)
		if err != nil {
			return nil, err
		}
		results = append(results, &m)
	}
	return results, nil
}

func (s *Store) InsertContainerMetrics(m *types.ContainerMetrics) error {
	_, err := s.db.Exec(`
		INSERT INTO containers (timestamp, container_id, container_name,
			cpu_percent, memory_percent, memory_usage)
		VALUES (?, ?, ?, ?, ?, ?)`,
		m.Timestamp, m.ContainerID, m.ContainerName,
		m.CPUPercent, m.MemPercent, m.MemUsage,
	)
	return err
}

func (s *Store) LatestContainers() ([]*types.ContainerMetrics, error) {
	rows, err := s.db.Query(`
		SELECT timestamp, container_id, container_name, cpu_percent,
			memory_percent, memory_usage
		FROM containers
		WHERE (container_id, timestamp) IN (
			SELECT container_id, MAX(timestamp) FROM containers GROUP BY container_id
		)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []*types.ContainerMetrics
	for rows.Next() {
		var m types.ContainerMetrics
		err := rows.Scan(&m.Timestamp, &m.ContainerID, &m.ContainerName,
			&m.CPUPercent, &m.MemPercent, &m.MemUsage)
		if err != nil {
			return nil, err
		}
		results = append(results, &m)
	}
	return results, nil
}

func (s *Store) Prune(olderThan time.Duration) error {
	hours := int(olderThan.Hours())
	_, err := s.db.Exec(fmt.Sprintf("DELETE FROM metrics WHERE timestamp < datetime('now', '-%d hours')", hours))
	if err != nil {
		return err
	}
	_, err = s.db.Exec(fmt.Sprintf("DELETE FROM containers WHERE timestamp < datetime('now', '-%d hours')", hours))
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}
