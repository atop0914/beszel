# Beszel — Lightweight Server Monitoring

A self-hosted server monitoring agent and dashboard built in pure Go. Single binary, SQLite storage, zero external runtime dependencies.

![Dashboard](https://img.shields.io/badge/dashboard-live-brightgreen) ![Go](https://img.shields.io/badge/go-1.21+-blue) ![License](https://img.shields.io/badge/license-MIT-blue)

## Features

- **System metrics** — CPU usage, memory, disk I/O, network throughput, load averages, and uptime
- **Docker stats** — Per-container CPU and memory monitoring (when Docker is available)
- **Historical charts** — 24-hour and 7-day trend charts via Chart.js
- **SQLite persistence** — No external database; metrics stored in a single `.db` file with automatic pruning
- **Alert thresholds** — Configurable CPU/memory/disk thresholds with optional webhook notifications
- **Single binary** — Compiles to a standalone executable, no runtime dependencies

## Requirements

- Go 1.21 or later
- Docker (optional, for container stats)
- Linux, macOS, or Windows

## Installation

### Pre-built Binary

Download the latest release for your platform from the [releases page](https://github.com/atop0914/beszel/releases).

```bash
# Linux amd64 example
curl -LO https://github.com/atop0914/beszel/releases/latest/download/beszel-linux-amd64
chmod +x beszel-linux-amd64
./beszel-linux-amd64 serve
```

### Build from Source

```bash
git clone https://github.com/atop0914/beszel.git
cd beszel
go build -o beszel ./cmd/beszel
```

Or use the Makefile:

```bash
make build
```

## Quick Start

```bash
# Start the web dashboard (background collector starts automatically)
./beszel serve

# Open in browser
open http://localhost:8080
```

On first run, `beszel.yaml` is created with default settings if it doesn't exist. The SQLite database (`beszel.db`) is created automatically.

## Configuration

Edit `beszel.yaml` in the project directory or at `~/.beszel/config.yaml`:

```yaml
# HTTP server port (default: 8080)
port: 8080

# Path to SQLite database (default: ./beszel.db)
db_path: ./beszel.db

# How often to collect metrics (default: 30s)
collection_interval: 30s

# Enable Docker container stats collection (default: true)
docker_enabled: true

# Alert thresholds — checked after each collection
alerts:
  # CPU usage percentage threshold (0–100, default: 80)
  cpu_threshold: 80

  # Memory usage percentage threshold (0–100, default: 85)
  memory_threshold: 85

  # Disk usage percentage threshold (0–100, default: 90)
  disk_threshold: 90

  # Optional webhook URL — POST JSON alert body on threshold breach
  # Leave empty to disable webhooks (alerts are logged to console only).
  webhook_url: ""
```

### Configuration Options Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `port` | int | `8080` | HTTP server port |
| `db_path` | string | `./beszel.db` | SQLite database file path |
| `collection_interval` | duration | `30s` | Metrics collection frequency (minimum 1s) |
| `docker_enabled` | bool | `true` | Enable Docker container monitoring |
| `alerts.cpu_threshold` | float | `80` | CPU alert threshold (0–100%) |
| `alerts.memory_threshold` | float | `85` | Memory alert threshold (0–100%) |
| `alerts.disk_threshold` | float | `90` | Disk alert threshold (0–100%) |
| `alerts.webhook_url` | string | `""` | HTTP endpoint for alert POST requests |

## Commands

```bash
./beszel serve          # Start web dashboard with background collector (default)
./beszel collect-once   # Collect a single metrics snapshot and exit
./beszel config         # Validate configuration and print resolved values
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | Dashboard HTML page |
| `GET` | `/api/metrics` | Latest metrics snapshot as JSON |
| `GET` | `/api/metrics/history?hours=24` | Historical metrics for N hours (max 168 = 7d) |
| `GET` | `/api/containers` | Latest Docker container stats as JSON |
| `GET` | `/health` | Health check — returns `ok` |

### `GET /api/metrics` Response

```json
{
  "timestamp": "2026-04-15T20:30:00Z",
  "hostname": "myserver",
  "cpu_percent": 23.5,
  "memory_percent": 61.2,
  "memory_used": 65812922368,
  "memory_total": 107374182400,
  "disk_percent": 45.3,
  "disk_used": 214748364800,
  "disk_total": 474367988736,
  "network_rx": 1234567890,
  "network_tx": 987654321,
  "uptime": 86400,
  "os": "linux",
  "platform": "Ubuntu 22.04",
  "kernel": "5.15.0-generic",
  "load1": 0.42,
  "load5": 0.38,
  "load15": 0.31
}
```

### `GET /api/containers` Response

```json
[
  {
    "timestamp": "2026-04-15T20:30:00Z",
    "container_id": "abc123def456",
    "container_name": "/my-app",
    "cpu_percent": 5.2,
    "memory_percent": 34.1,
    "memory_usage": 1073741824
  }
]
```

### Webhook Alert Payload

When an alert threshold is breached and `webhook_url` is configured, a POST request is sent with:

```json
{
  "timestamp": "2026-04-15T20:30:00Z",
  "hostname": "myserver",
  "metric": "cpu",
  "value": 85.3,
  "threshold": 80.0
}
```

## Architecture

```
beszel/
├── cmd/beszel/main.go           # CLI entry point, command routing
├── internal/
│   ├── collector/
│   │   ├── system.go            # CPU, memory, disk, network, load, uptime
│   │   └── docker.go            # Docker container stats via Docker API
│   ├── store/
│   │   └── sqlite.go            # SQLite operations, schema migration, pruning
│   ├── web/
│   │   ├── server.go            # HTTP routes, collection loop, alert checking
│   │   ├── templates.go         # Embedded dashboard HTML
│   │   └── dashboard.html       # Dark-themed Chart.js dashboard (gauge + charts)
│   ├── config/
│   │   └── config.go            # YAML config loading and validation
│   ├── alerts/
│   │   └── alerts.go            # Threshold checking, console + webhook alerts
│   └── types/
│       └── metrics.go           # Metrics and ContainerMetrics data structures
├── beszel.yaml                  # Configuration file
├── Makefile                     # Build, run, test, lint shortcuts
└── .goreleaser.yaml             # Release build configuration
```

### Data Flow

```
[collection_interval ticker — every 30s by default]
        │
        ▼
[SystemCollector.Collect()] ──► [SQLite: INSERT metrics]
[DockerCollector.Collect()] ──► [SQLite: INSERT containers]
        │
        ▼
[AlertManager.Check(metrics)] ──► [log / webhook POST]
        │
        ▼
[prune ticker — every 1h] ──► [DELETE rows older than 7 days]
        │
        ▼
[HTTP server ──► JSON / HTML dashboard ← browser]
```

## Development

```bash
make build    # Build the binary
make run      # Build and run ./beszel serve
make test     # Run tests
make lint     # Run go vet
make clean    # Remove binary and .db
```

## Release

Releases are built with [goreleaser](https://goreleaser.com/). Configure your GitHub token and run:

```bash
goreleaser release --clean
```

This produces platform-specific binaries and GitHub releases automatically.

## License

MIT
