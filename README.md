# go-enphase

[![CI](https://github.com/sebrandon1/go-enphase/actions/workflows/pre-main.yaml/badge.svg)](https://github.com/sebrandon1/go-enphase/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/sebrandon1/go-enphase.svg)](https://pkg.go.dev/github.com/sebrandon1/go-enphase)

Go CLI and library for the Enphase cloud API (v4) and local Envoy gateway.

## Installation

```bash
go install github.com/sebrandon1/go-enphase@latest
```

Or build from source:

```bash
git clone https://github.com/sebrandon1/go-enphase.git
cd go-enphase
make build
```

## Configuration

Credentials can be provided via CLI flags, environment variables, or a config file.

### Config file

Default location: `~/.enphase/config` (KEY=VALUE format, one per line, `#` comments supported).

| Key | Env var | CLI flag | Description |
|---|---|---|---|
| `ENPHASE_API_KEY` | `ENPHASE_API_KEY` | `--api-key` | Enphase Developer API key |
| `ENPHASE_ACCESS_TOKEN` | `ENPHASE_ACCESS_TOKEN` | `--access-token` | OAuth2 access token |
| `ENPHASE_REFRESH_TOKEN` | — | `--refresh-token` | OAuth2 refresh token |
| `ENPHASE_CLIENT_ID` | — | `--client-id` | OAuth2 client ID |
| `ENPHASE_CLIENT_SECRET` | — | `--client-secret` | OAuth2 client secret |
| `ENPHASE_SYSTEM_ID` | — | `--system-id` | Enphase system ID |
| `ENPHASE_RATE_PER_KWH` | — | — | Electricity rate ($/kWh) for reports |
| `ENPHASE_REDIRECT_URI` | — | — | OAuth2 redirect URI |

**Precedence:** CLI flags > config file values.

Use a custom config file with `--config /path/to/file`.

### Cloud API

Obtain an API key and OAuth2 access token from the [Enphase Developer Portal](https://developer-v4.enphase.com/).

```bash
# ~/.enphase/config
ENPHASE_API_KEY=your-api-key
ENPHASE_ACCESS_TOKEN=your-access-token
ENPHASE_SYSTEM_ID=12345
```

### Local Envoy

Set `ENPHASE_ENVOY_IP` to your Envoy's IP address. A JWT is required for most endpoints:

```bash
go-enphase auth envoy-token --email user@example.com --password mypass --envoy-serial ABC123
```

## CLI Usage

### Cloud API commands

```bash
# List all systems
go-enphase get systems

# Get system summary
go-enphase get summary --system-id 12345

# List devices
go-enphase get devices --system-id 12345

# Get production meter readings
go-enphase get production --system-id 12345

# Get lifetime energy production
go-enphase get energy-lifetime --system-id 12345 --start-date 2024-01-01 --end-date 2024-12-31

# Get lifetime consumption
go-enphase get consumption --system-id 12345

# Get battery status
go-enphase get battery --system-id 12345
```

### Authentication commands

```bash
# Check token status
go-enphase auth status

# Refresh access token
go-enphase auth refresh --refresh-token $REFRESH --client-id $CID --client-secret $CS

# Get Envoy JWT token
go-enphase auth envoy-token --email user@example.com --password mypass --envoy-serial ABC123
```

### Report Commands

```bash
# Today's production summary
go-enphase report today --system-id 12345

# Daily report: today's live status + last 7 days production vs consumption
go-enphase report daily --system-id 12345

# Daily report with custom trailing days
go-enphase report daily --system-id 12345 --days 14

# Compare two months of production
go-enphase report compare --system-id 12345 2025-01 2025-02

# Export full production/consumption history to JSON
go-enphase report history --system-id 12345 --output ~/solar/history.json
```

All report commands support `--rate 0.13` to include dollar estimates (also configurable via `ENPHASE_RATE_PER_KWH` in the config file).

### Local Envoy Commands

```bash
# Get production/consumption
go-enphase envoy status --envoy-ip 192.168.1.100

# Get sensor readings
go-enphase envoy sensors --envoy-ip 192.168.1.100 --envoy-token $JWT
```

## Library Usage

### Basic cloud access

```go
client, err := lib.NewClient("api-key", "access-token")
systems, err := client.ListSystems()
summary, err := client.GetSystemSummary("12345")
```

### Context-aware calls

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

summary, err := client.GetSystemSummaryCtx(ctx, "12345")
inverters, err := envoy.GetInverterReadingsCtx(ctx)
```

### Functional options

```go
client, err := lib.NewClientWithOptions("api-key", "access-token",
    lib.WithTimeout(15*time.Second),
    lib.WithInsecureSkipVerify(false),
)
```

### Local Envoy

```go
envoy, err := lib.NewEnvoyClient("192.168.1.100", jwtToken)

// Basic production
prod, err := envoy.GetEnvoyProduction()

// Per-inverter readings
inverters, err := envoy.GetInverterReadings()

// Revenue-grade meters
meters, err := envoy.GetMeterConfig()
readings, err := envoy.GetMeterReadings()
```

### SSE real-time streaming

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

envoy.StreamMeter(ctx, func(ev *lib.StreamMeterEvent, err error) {
    if err != nil {
        log.Printf("stream error: %v", err)
        return
    }
    fmt.Printf("EID %d: %.1f W at %d\n", ev.EID, ev.ActPower, ev.Timestamp)
})
```

### Envoy JWT caching

```go
// EnsureEnvoyToken fetches or reuses a cached JWT, refreshing 5 min before expiry.
err := client.EnsureEnvoyToken(ctx, "user@example.com", "password", "SERIAL123")
```

## API Coverage

### Cloud API (v4)

| Endpoint | Method |
|---|---|
| `GET /api/v4/systems` | `ListSystems[Ctx]()` |
| `GET /api/v4/systems/{id}` | `GetSystem[Ctx]()` |
| `GET /api/v4/systems/{id}/summary` | `GetSystemSummary[Ctx]()` |
| `GET /api/v4/systems/{id}/devices` | `ListDevices[Ctx]()` |
| `GET /api/v4/systems/{id}/production_meter_readings` | `GetProductionMeterReadings[Ctx]()` |
| `GET /api/v4/systems/{id}/energy_lifetime` | `GetEnergyLifetime[Ctx]()` |
| `GET /api/v4/systems/{id}/consumption_lifetime` | `GetConsumptionLifetime[Ctx]()` |
| `GET /api/v4/systems/{id}/battery_lifetime` | `GetBatteryStatus[Ctx]()` |

### Local Envoy

| Endpoint | Method |
|---|---|
| `GET /production.json` | `GetEnvoyProduction[Ctx]()` |
| `GET /ivp/sensors/readings_object` | `GetEnvoySensors[Ctx]()` |
| `GET /api/v1/production` | `GetEnvoySimpleProduction[Ctx]()` |
| `GET /api/v1/production/inverters` | `GetInverterReadings[Ctx]()` |
| `GET /ivp/meters` | `GetMeterConfig[Ctx]()` |
| `GET /ivp/meters/readings` | `GetMeterReadings[Ctx]()` |
| `GET /stream/meter` (SSE) | `StreamMeter(ctx, handler)` |

### Authentication

| Flow | Method |
|---|---|
| Token refresh | `RefreshAccessToken[Ctx]()` |
| Auth code exchange | `ExchangeAuthCode[Ctx]()` |
| Envoy JWT acquisition | `GetEnvoyToken[Ctx]()` |
| Envoy JWT caching | `EnsureEnvoyToken(ctx, ...)` |

## ha-exporter

`examples/ha-exporter/` is a standalone binary that:

- Polls the Enphase cloud and local Envoy APIs on a configurable interval
- Exposes a Prometheus `/metrics` endpoint
- Publishes Home Assistant MQTT auto-discovery config and state values

### Config file (JSON)

```json
{
  "api_key": "your-api-key",
  "access_token": "your-access-token",
  "system_id": "12345",
  "envoy_ip": "192.168.1.100",
  "envoy_token": "your-envoy-jwt",
  "envoy_serial": "SERIAL123",
  "poll_interval": "30s",
  "metrics_addr": ":9090",
  "mqtt_broker": "tcp://192.168.1.10:1883",
  "mqtt_username": "mqtt_user",
  "mqtt_password": "mqtt_pass",
  "mqtt_topic_prefix": "homeassistant"
}
```

### Running

```bash
cd examples/ha-exporter
go build .
./ha-exporter --config config.json
./ha-exporter --config config.json --dry-run   # no API calls or MQTT publishes
```

### Prometheus metrics

| Metric | Type | Description |
|---|---|---|
| `enphase_current_power_watts` | gauge | Current production (W) |
| `enphase_energy_today_wh` | gauge | Energy produced today (Wh) |
| `enphase_energy_lifetime_wh` | counter | Lifetime energy (Wh) |
| `enphase_net_power_watts` | gauge | Net power: production − consumption (W) |
| `enphase_inverter_watts{serial="..."}` | gauge | Per-inverter production (W) |

### Home Assistant integration

The exporter publishes MQTT auto-discovery messages compatible with the [MQTT integration](https://www.home-assistant.io/integrations/mqtt/). Ensure your HA instance has MQTT configured and the broker address matches your config.

### Prometheus scrape config

```yaml
scrape_configs:
  - job_name: enphase
    static_configs:
      - targets: ['localhost:9090']
```

## Development

```bash
make build    # Build binary
make test     # Run tests (verbose)
make lint     # Run golangci-lint
make vet      # Run go vet
make clean    # Remove binary
```

## License

Apache License 2.0
