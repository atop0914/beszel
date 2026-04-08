# Beszel — Lightweight Server Monitoring

A self-hosted server monitoring agent and dashboard built in Go. Single binary, SQLite storage, zero external dependencies.

## Features

- **System metrics**: CPU, memory, disk, network I/O
- **Docker stats**: Container-level CPU and memory monitoring
- **Historical data**: 24h and 7d charts via Chart.js
- **SQLite storage**: No external database required
- **Single binary**: Compiles to a standalone executable

## Quick Start

```bash
# Build
go build -o beszel ./cmd/beszel

# Start the server (web dashboard + background collector)
./beszel serve

# Open browser
open http://localhost:8080
```

## Configuration

Edit `beszel.yaml`:

```yaml
port: 8080
db_path: ./beszel.db
collection_interval: 30s
docker_enabled: true
alerts:
  cpu_threshold: 80
  memory_threshold: 85
  disk_threshold: 90
  webhook_url: ""
```

## Commands

```bash
./beszel serve        # Start web server with background collection
./beszel collect-once # Collect metrics once and exit
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Dashboard HTML |
| GET | `/api/metrics` | Latest metrics (JSON) |
| GET | `/api/metrics/history?hours=24` | Historical metrics (JSON) |
| GET | `/api/containers` | Docker container stats (JSON) |
| GET | `/health` | Health check |

## Architecture

```
cmd/beszel/main.go        — CLI entry point
internal/
  collector/system.go    — CPU, memory, disk, network collection
  collector/docker.go    — Docker container stats
  store/sqlite.go        — SQLite persistence
  web/server.go          — HTTP server
  web/templates.go       — Embedded dashboard HTML
  config/config.go       — YAML config loading
```

## License

MIT
