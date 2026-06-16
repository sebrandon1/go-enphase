# go-enphase

[![CI](https://github.com/sebrandon1/go-enphase/actions/workflows/pre-main.yaml/badge.svg)](https://github.com/sebrandon1/go-enphase/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/sebrandon1/go-enphase.svg)](https://pkg.go.dev/github.com/sebrandon1/go-enphase)

Go CLI and library for the Enphase cloud API (v4) and local Envoy gateway.

## Features

- Query solar systems, production, consumption, and battery data via the cloud API
- Local Envoy gateway access with SSE real-time meter streaming
- OAuth2 token management with auto-refresh and Envoy JWT caching
- Reports: daily summaries, month-over-month comparisons, history export
- ha-exporter: Prometheus metrics + Home Assistant MQTT integration

## Quick start

```bash
go install github.com/sebrandon1/go-enphase@latest
```

Create `~/.enphase/config`:

```
ENPHASE_API_KEY=your-api-key
ENPHASE_ACCESS_TOKEN=your-access-token
ENPHASE_SYSTEM_ID=12345
```

```bash
go-enphase get systems
go-enphase get summary --system-id 12345
go-enphase report daily --system-id 12345
```

## Library

```go
client, err := lib.NewClient("api-key", "access-token")
systems, err := client.ListSystems()
summary, err := client.GetSystemSummary("12345")
```

## Guides

| Guide | Description |
|---|---|
| [Configuration](docs/configuration.md) | Config file, env vars, CLI flags, cloud and Envoy setup |
| [CLI Reference](docs/cli-reference.md) | All CLI commands with examples |
| [Library Usage](docs/library-usage.md) | Go library examples, context-aware calls, SSE streaming |
| [API Coverage](docs/api-coverage.md) | Endpoint-to-method mapping for cloud, Envoy, and auth APIs |
| [ha-exporter](docs/ha-exporter.md) | Prometheus + Home Assistant MQTT exporter |

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
