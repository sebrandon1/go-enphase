# go-enphase

Go CLI and library for the Enphase cloud API and local Envoy gateway.

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

All configuration is via CLI flags or environment variables. No config file required.

### Cloud API

Requires an API key and OAuth2 access token from the [Enphase Developer Portal](https://developer-v4.enphase.com/).

```bash
export ENPHASE_API_KEY="your-api-key"
export ENPHASE_ACCESS_TOKEN="your-access-token"
```

### Local Envoy

Requires the IP address of your local Envoy gateway. Optionally provide a JWT token for authenticated endpoints.

```bash
export ENPHASE_ENVOY_IP="192.168.1.100"
```

## CLI Usage

### Cloud API Commands

```bash
# List all systems
go-enphase get systems --api-key $KEY --access-token $TOKEN

# Get system summary
go-enphase get summary --api-key $KEY --access-token $TOKEN --system-id 12345

# List devices
go-enphase get devices --api-key $KEY --access-token $TOKEN --system-id 12345

# Get production meter readings
go-enphase get production --api-key $KEY --access-token $TOKEN --system-id 12345

# Get lifetime energy production
go-enphase get energy-lifetime --api-key $KEY --access-token $TOKEN --system-id 12345 \
  --start-date 2024-01-01 --end-date 2024-12-31

# Get lifetime consumption
go-enphase get consumption --api-key $KEY --access-token $TOKEN --system-id 12345

# Get battery status
go-enphase get battery --api-key $KEY --access-token $TOKEN --system-id 12345
```

### Authentication Commands

```bash
# Check token status (no secrets displayed)
go-enphase auth status --api-key $KEY --access-token $TOKEN

# Refresh access token
go-enphase auth refresh --api-key $KEY --access-token $TOKEN \
  --refresh-token $REFRESH --client-id $CID --client-secret $CS

# Get Envoy JWT token
go-enphase auth envoy-token --api-key $KEY --access-token $TOKEN \
  --email user@example.com --password mypass --envoy-serial ABC123
```

### Local Envoy Commands

```bash
# Get local production/consumption
go-enphase envoy status --envoy-ip 192.168.1.100

# Get sensor readings
go-enphase envoy sensors --envoy-ip 192.168.1.100 --envoy-token $JWT
```

## Library Usage

```go
package main

import (
    "fmt"
    "github.com/sebrandon1/go-enphase/lib"
)

func main() {
    // Cloud API
    client, _ := lib.NewClient("api-key", "access-token")
    systems, _ := client.ListSystems()
    fmt.Printf("Found %d systems\n", len(systems))

    summary, _ := client.GetSystemSummary("12345")
    fmt.Printf("Current power: %d W\n", summary.CurrentPower)

    // Local Envoy
    envoy, _ := lib.NewEnvoyClient("192.168.1.100", "")
    production, _ := envoy.GetEnvoyProduction()
    for _, p := range production.Production {
        fmt.Printf("%s: %.0f W\n", p.Type, p.WNow)
    }
}
```

## API Coverage

### Cloud API (v4)

| Endpoint | Method | Status |
|----------|--------|--------|
| `GET /api/v4/systems` | `ListSystems()` | Implemented |
| `GET /api/v4/systems/{id}` | `GetSystem()` | Implemented |
| `GET /api/v4/systems/{id}/summary` | `GetSystemSummary()` | Implemented |
| `GET /api/v4/systems/{id}/devices` | `ListDevices()` | Implemented |
| `GET /api/v4/systems/{id}/production_meter_readings` | `GetProductionMeterReadings()` | Implemented |
| `GET /api/v4/systems/{id}/energy_lifetime` | `GetEnergyLifetime()` | Implemented |
| `GET /api/v4/systems/{id}/consumption_lifetime` | `GetConsumptionLifetime()` | Implemented |
| `GET /api/v4/systems/{id}/battery_lifetime` | `GetBatteryStatus()` | Implemented |

### Local Envoy

| Endpoint | Method | Status |
|----------|--------|--------|
| `GET /production.json` | `GetEnvoyProduction()` | Implemented |
| `GET /ivp/sensors/readings_object` | `GetEnvoySensors()` | Implemented |

### Authentication

| Flow | Method | Status |
|------|--------|--------|
| Token refresh | `RefreshAccessToken()` | Implemented |
| Auth code exchange | `ExchangeAuthCode()` | Implemented |
| Envoy JWT | `GetEnvoyToken()` | Implemented |

## Development

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run golangci-lint
make vet      # Run go vet
make clean    # Remove binary
```

## License

Apache License 2.0
