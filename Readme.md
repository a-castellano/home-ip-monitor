# Home IP Monitor

[![pipeline status](https://git.windmaker.net/a-castellano/home-ip-monitor/badges/master/pipeline.svg)](https://git.windmaker.net/a-castellano/home-ip-monitor/pipelines)[![coverage report](https://git.windmaker.net/a-castellano/home-ip-monitor/badges/master/coverage.svg)](https://a-castellano.gitpages.windmaker.net/home-ip-monitor/coverage.html)[![Quality Gate Status](https://sonarqube.windmaker.net/api/project_badges/measure?project=a-castellano_home-ip-monitor_a0d9946c-4181-4181-af10-e5dac69d0658&metric=alert_status&token=sqb_991ee37d1ea08ee63db5ea610f2a2d9e49fe1430)](https://sonarqube.windmaker.net/dashboard?id=a-castellano_home-ip-monitor_a0d9946c-4181-4181-af10-e5dac69d0658)

A Go-based service that monitors your home's public IP address and notifies when changes occur. It's designed to run as a systemd service with automatic execution every 2 minutes.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Configuration](#configuration)
- [Testing](#testing)
- [License](#license)

## Overview

Home IP Monitor is a lightweight service that:

1. **Fetches your public IP** from [ipinfo.io](https://ipinfo.io/)
2. **Validates ISP consistency** to ensure you're still with your expected provider
3. **Checks for changes** by comparing with the previously stored IP, cross-checking against the domain's live DNS record when storage looks unchanged
4. **Sends notifications** via RabbitMQ when changes are detected, persisting the new IP only after the notifications succeed

## Features

- **Real-time IP monitoring** with configurable check intervals
- **ISP validation** to detect unexpected provider changes
- **Redis-based storage** for persistent IP tracking
- **RabbitMQ integration** for reliable message delivery
- **Systemd service** with automatic startup and timer

## Architecture

The project follows a Clean Architecture layout: an inner `domain` layer that
depends on nothing, an `app` layer that holds the use case, and an `infra` layer
with the adapters that talk to the outside world (HTTP, DNS, Redis, RabbitMQ).
Dependencies always point inwards — the use case only knows about domain ports
(interfaces), never about concrete infrastructure.

```
                         cmd/home-ip-monitor (composition root)
                                       │ wires adapters into the use case
                                       ▼
┌──────────────────────────────────────────────────────────────────────┐
│ app: Monitor use case                                                  │
│   reads IP → validates ISP → compares stored/DNS IP → notifies/persists│
└──────────────────────────────────────────────────────────────────────┘
        │ depends only on domain ports (interfaces)
        ▼
┌──────────────────────────────────────────────────────────────────────┐
│ domain: IPInfo, IPInfoProvider, DNSResolver, IPStore, Notifier         │
└──────────────────────────────────────────────────────────────────────┘
        ▲ implemented by infra adapters
        │
┌───────────────┬───────────────┬───────────────┬───────────────────────┐
│ ipinfodata    │ nslookup      │ storage       │ notify                │
│ → ipinfo.io   │ → DNS server  │ → Redis/Valkey│ → RabbitMQ            │
└───────────────┴───────────────┴───────────────┴───────────────────────┘
```

### Layers and components

- **`internal/domain`**: pure business types and ports (interfaces) with no
  external dependencies — `IPInfo` (+ `BelongsToISP`) and the
  `IPInfoProvider`, `DNSResolver`, `IPStore` and `Notifier` ports.
- **`internal/app`**: the `Monitor` use case. It receives the domain ports via
  `NewMonitor` and an `app.Settings` value object, so it has zero knowledge of
  HTTP, Redis or RabbitMQ.
- **`internal/infra/ipinfodata`**: HTTP adapter that fetches the public IP from
  [ipinfo.io](https://ipinfo.io/) and maps it to `domain.IPInfo`.
- **`internal/infra/nslookup`**: DNS adapter that resolves the configured domain
  through an external DNS server.
- **`internal/infra/storage`**: Redis/Valkey-backed adapter (via go-services
  `memorydatabase`) for persistent IP tracking.
- **`internal/infra/notify`**: RabbitMQ adapter (via go-services
  `messagebroker`) for delivering messages.
- **`internal/infra/config`**: environment-based configuration loading.
- **`cmd/home-ip-monitor`**: the composition root (`main`) that builds every
  adapter, maps `config.Config` to `app.Settings` and runs the use case.

## Installation

### Prerequisites

- **Arch Linux** or compatible distribution
- **Redis** or **Valkey** server for data storage
- **RabbitMQ** server for message queuing
- **Systemd** for service management (Linux)

### Package Installation (Recommended)

The project provides pre-built packages for Arch Linux. Download the latest package from the CI artifacts:

1. **Download the package** from the latest successful pipeline:

   - Go to [GitLab CI/CD Pipelines](https://git.windmaker.net/a-castellano/home-ip-monitor/pipelines)
   - Find the latest successful pipeline
   - Download the `arch_package` artifact

2. **Install the package**:

   ```bash
   # Install the downloaded package
   sudo pacman -U windmaker-home-ip-monitor-*.pkg.tar.zst
   ```

3. **Configure the service**:

   ```bash
   # Copy the sample file and edit your configuration
   sudo cp /etc/default/windmaker-home-ip-monitor-example /etc/default/windmaker-home-ip-monitor
   sudo vim /etc/default/windmaker-home-ip-monitor
   ```

4. **Enable and start the service**:
   ```bash
   # Enable the timer (runs every 2 minutes)
   sudo systemctl enable windmaker-home-ip-monitor.timer
   sudo systemctl start windmaker-home-ip-monitor.timer
   ```

## Configuration

### Environment Variables

All configuration is done through environment variables:

#### Required Variables

| Variable      | Description                     | Example              |
| ------------- | ------------------------------- | -------------------- |
| `DOMAIN_NAME` | Domain to verify IP against     | `"home.example.com"` |
| `ISP_NAME`    | Expected ISP provider name      | `"DIGI"`             |
| `DNS_SERVER`  | External DNS server for lookups | `"8.8.8.8:53"`       |

#### Optional Variables

| Variable            | Description                     | Default                           |
| ------------------- | ------------------------------- | --------------------------------- |
| `UPDATE_QUEUE_NAME` | Queue for IP update messages    | `"home-ip-monitor-updates"`       |
| `NOTIFY_QUEUE_NAME` | Queue for notification messages | `"home-ip-monitor-notifications"` |

#### Application and Logging

Logging is handled through [go-types `slog`](https://git.windmaker.net/a-castellano/go-types/-/tree/master/slog). `APP_NAME` is required by that type; the rest fall back to sane defaults.

| Variable          | Description                                            | Default          |
| ----------------- | ----------------------------------------------------- | ---------------- |
| `APP_NAME`        | Application name attached to every log entry          | _(required)_     |
| `SLOG_LEVEL`      | Log level: `Debug`, `Info`, `Warn` or `Error`         | `Info`           |
| `SLOG_FORMAT`     | Log format: `JSON` or `plain`                         | `JSON`           |
| `SLOG_ADD_SOURCE` | Whether to add `file:line` to log entries             | `true`           |

#### Redis Configuration

See [go-types Redis documentation](https://git.windmaker.net/a-castellano/go-types/-/tree/master/redis) for complete Redis configuration options.

| Variable         | Description                   | Default       |
| ---------------- | ----------------------------- | ------------- |
| `REDIS_HOST`     | Redis server hostname         | `"127.0.0.1"` |
| `REDIS_PORT`     | Redis server port             | `6379`        |
| `REDIS_PASSWORD` | Redis authentication password | `""`          |
| `REDIS_DATABASE` | Redis database number         | `10`          |

#### RabbitMQ Configuration

See [go-types RabbitMQ documentation](https://git.windmaker.net/a-castellano/go-types/-/tree/master/rabbitmq) for complete RabbitMQ configuration options.

| Variable            | Description              | Default       |
| ------------------- | ------------------------ | ------------- |
| `RABBITMQ_HOST`     | RabbitMQ server hostname | `"localhost"` |
| `RABBITMQ_PORT`     | RabbitMQ server port     | `5672`        |
| `RABBITMQ_USER`     | RabbitMQ username        | `"guest"`     |
| `RABBITMQ_PASSWORD` | RabbitMQ password        | `"guest"`     |

### Configuration File

The package installs a sample file at `/etc/default/windmaker-home-ip-monitor-example`. Copy it to `/etc/default/windmaker-home-ip-monitor` (the path read by the systemd unit) and edit it to configure the service:

```bash
sudo cp /etc/default/windmaker-home-ip-monitor-example /etc/default/windmaker-home-ip-monitor
sudo vim /etc/default/windmaker-home-ip-monitor
```

```bash
# Application and logging
APP_NAME="home-ip-monitor"
SLOG_LEVEL="Info"
SLOG_FORMAT="JSON"

# Required configuration
DOMAIN_NAME="your-domain.com"
ISP_NAME="DIGI"
DNS_SERVER="8.8.8.8:53"

# Queue configuration
UPDATE_QUEUE_NAME="home-ip-monitor-updates"
NOTIFY_QUEUE_NAME="home-ip-monitor-notifications"

# Redis configuration
REDIS_HOST="127.0.0.1"
REDIS_PORT=6379
REDIS_PASSWORD=""
REDIS_DATABASE=10

# RabbitMQ configuration
RABBITMQ_HOST="localhost"
RABBITMQ_PORT=5672
RABBITMQ_USER="guest"
RABBITMQ_PASSWORD="guest"
```

## Usage

### Service Management

```bash
# Enable and start the timer (runs every 2 minutes)
sudo systemctl enable windmaker-home-ip-monitor.timer
sudo systemctl start windmaker-home-ip-monitor.timer

# Check service status
sudo systemctl status windmaker-home-ip-monitor.service
sudo systemctl status windmaker-home-ip-monitor.timer

# View logs
sudo journalctl -u windmaker-home-ip-monitor.service -f

# Manual execution
sudo systemctl start windmaker-home-ip-monitor.service
```

### Message Queues

The service sends two types of messages to RabbitMQ:

#### Update Messages (`UPDATE_QUEUE_NAME`)

Contains the new IP address as plain text:

```
192.168.1.100
```

#### Notification Messages (`NOTIFY_QUEUE_NAME`)

Contains human-readable notifications:

```
Home IP has changed to 192.168.1.100.
Read IP 192.168.1.100 belongs to DIGI ISP, it seems that home is not using main ISP ORANGE.
```

### Monitoring and Logging

The service uses structured logging through [`log/slog`](https://pkg.go.dev/log/slog) (via go-types `slog`). Output goes to standard streams, which systemd captures into the journal. The format (`JSON` or `plain`) and verbosity are controlled by `SLOG_FORMAT` and `SLOG_LEVEL`.

- **DEBUG**: configuration loading, IP retrieval, comparison and decision steps (most of the flow is logged at this level)
- **INFO**: high-level milestones such as service startup
- **ERROR**: connection failures, configuration errors, processing failures

Each entry carries an `operation` attribute (e.g. `NewConfig`, `Monitor.Run`) plus structured fields. Example output with `SLOG_FORMAT="JSON"` and `SLOG_LEVEL="Debug"`:

```json
{"time":"2026-06-24T10:00:00Z","level":"DEBUG","msg":"Loading config"}
{"time":"2026-06-24T10:00:00Z","level":"DEBUG","msg":"Domain name has been set","operation":"NewConfig","domain":"home.example.com"}
{"time":"2026-06-24T10:00:00Z","level":"INFO","msg":"Initiating required services"}
{"time":"2026-06-24T10:00:00Z","level":"DEBUG","msg":"Validating that ipinfo provider is the expected provider","operation":"Monitor.Run","currentProvider":"DIGI","expectedProvider":"DIGI","currentIP":"192.168.1.100"}
{"time":"2026-06-24T10:00:00Z","level":"DEBUG","msg":"IPs differ, stored IP must be updated","operation":"Monitor.Run","currentIP":"192.168.1.100","storedIP":"192.168.1.99"}
{"time":"2026-06-24T10:00:00Z","level":"DEBUG","msg":"Updating stored IP","operation":"Monitor.Run","currentIP":"192.168.1.100"}
```

## Development

### Prerequisites

- Go 1.26+
- Docker (or Podman) and Docker (or Podman) Compose
- Make

### Setup Development Environment

```bash
# Start development services (Go container + Valkey + RabbitMQ)
docker-compose -f development/docker-compose.yml up -d

# Install dependencies
go mod download

# Run tests
make test
make test_integration

# Generate coverage report
make coverage
make coverhtml
```

To run the binary against the development services in a production-like setup,
build it and exec into the dev container, then load the sample environment:

```bash
make build
docker-compose -f development/docker-compose.yml exec golang /bin/bash

# Inside the container
source development/env_variables
./home-ip-monitor
```

`development/env_variables` points `REDIS_HOST`/`RABBITMQ_HOST` at the Valkey and
RabbitMQ containers and sets `APP_NAME` and the `SLOG_*` logging variables, so the
service runs in an environment similar to production.

### Project Structure

```
home-ip-monitor/
├── cmd/
│   └── home-ip-monitor/    # main package: composition root / wiring
├── internal/
│   ├── domain/             # business types and ports (no external deps)
│   ├── app/                # Monitor use case (depends only on domain)
│   └── infra/              # adapters that implement the domain ports
│       ├── config/         # environment-based configuration
│       ├── ipinfodata/     # ipinfo.io HTTP client (+ generated mocks)
│       ├── nslookup/       # DNS resolution
│       ├── storage/        # Redis/Valkey persistence
│       └── notify/         # RabbitMQ notifications
├── development/            # Docker/Podman dev setup and coverage script
└── packaging/              # nfpm spec, systemd units and defaults
```

## Testing

### Test Types

- **Unit Tests**: Individual component testing
- **Integration Tests**: Full system testing with Redis/RabbitMQ

### Running Tests

```bash
# Unit tests only
make test

# Integration tests
make test_integration

# All tests with coverage
make coverage

# HTML coverage report
make coverhtml
```

Per-module targets are also available for focused runs, e.g. `make test_app`,
`make test_config`, `make test_storage` (and their `*_unit` variants for unit
tests only). Run `make help` to list every target.

### Mocks

HTTP-dependent code is unit-tested by mocking the `http.RoundTripper` instead of
hitting the network. Mocks are generated with
[Uber's `mockgen`](https://github.com/uber-go/mock), registered as a Go tool in
`go.mod` via the `tool` directive (requires Go 1.24+), so no global install is
needed — `go test`/`go generate` resolve it automatically.

The generator is declared with a `//go:generate` directive in
`internal/infra/ipinfodata/ipinfodata.go`:

```go
//go:generate go tool mockgen -destination mocks/http.go -package mock net/http RoundTripper
```

Regenerate the mocks after changing a mocked interface:

```bash
go generate ./...
```

> The directive must live in a file **without** build tags (e.g. `ipinfodata.go`).
> If it sits in a `_test.go` file guarded by `//go:build`, `go generate ./...` skips
> it unless you pass the matching `-tags`.

This produces `internal/infra/ipinfodata/mocks/http.go` (package `mock`) exposing
`MockRoundTripper`. Tests build an `http.Client` with the mock transport and set
expectations on `RoundTrip`:

```go
ctrl := gomock.NewController(t)
transport := mock.NewMockRoundTripper(ctrl)
transport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
    StatusCode: 200,
    Body:       io.NopCloser(bytes.NewBufferString(`{"ip":"1.2.3.4","org":"AS1 EXAMPLE"}`)),
}, nil)

requester := IPInfoRequester{httpClient: &http.Client{Transport: transport}}
info, err := requester.GetIPInfo(context.Background())
```

Generated mocks live under `mocks/` directories and are excluded from the coverage
report by `development/coverage.sh` (the `PKG_LIST` filter drops `/mocks` packages).

### Test Coverage

The project maintains high test coverage across all components:

- Configuration management
- IP information retrieval
- DNS resolution
- Storage operations
- Notification system
- Core monitoring logic

## License

This project is licensed under the GPLv3 License - see the LICENSE file for details.
