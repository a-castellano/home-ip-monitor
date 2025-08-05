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
2. **Checks for changes** by comparing with the previously stored IP
3. **Validates ISP consistency** to ensure you're still with your expected provider
4. **Sends notifications** via RabbitMQ when changes are detected

## Features

- ✅ **Real-time IP monitoring** with configurable check intervals
- ✅ **ISP validation** to detect unexpected provider changes
- ✅ **Redis-based storage** for persistent IP tracking
- ✅ **RabbitMQ integration** for reliable message delivery
- ✅ **Systemd service** with automatic startup and timer

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Systemd Timer │───▶│  Home IP Monitor│───▶│   ipinfo.io     │
│   (every 2min)  │    │   (Go Service)  │    │   (IP Detection)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     Redis       │◀───│   Storage Layer │───▶│   RabbitMQ      │
│  (IP History)   │    │   (IP Compare)  │    │  (Notifications)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Components

- **Monitor**: Core logic for IP checking and change detection
- **IPInfo**: Client for fetching public IP information
- **Storage**: Redis-based persistence layer
- **Notify**: RabbitMQ message broker integration
- **NSLookup**: DNS resolution for domain verification
- **Config**: Environment-based configuration management

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
   # Edit the configuration file
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

The package installation creates `/etc/default/windmaker-home-ip-monitor`. Edit this file to configure the service:

```bash
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
Readed IP 192.168.1.100 belongs to DIGI ISP, it seems than home is not using main ISP ORANGE.
```

### Monitoring and Logging

The service logs to syslog with the following message types:

- **INFO**: Configuration loading, IP retrieval, status updates
- **ERROR**: Connection failures, configuration errors, processing failures

Example log output:

```
Loading config
Domain name has been set to "home.example.com"
ISP name has been set to "DIGI"
DNS Server has been set to "8.8.8.8:53"
Creating Redis client
Creating RabbitMQ client
Retrieving IP info
Retrieved IP is "192.168.1.100"
Retrieved OrgName is "DIGI"
Checking IP info in storage.
IP update required: true
Home IP has changed to 192.168.1.100.
Updating IP in storage
Execution finished
```

## Development

### Prerequisites

- Go 1.24+
- Docker (or Podman) and Docker (or Podman) Compose
- Make

### Setup Development Environment

```bash
# Start development services
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

### Project Structure

```
home-ip-monitor/
├── app/                    # Application orchestration
├── config/                 # Configuration management
├── ipinfo/                 # IP information client
├── monitor/                # Core monitoring logic
├── nslookup/               # DNS resolution
├── notify/                 # Message notification
├── storage/                # Data persistence
├── development/            # Docker development setup
├── packaging/              # Systemd and packaging files
└── scripts/                # Build and utility scripts
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
