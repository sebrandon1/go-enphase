# CLAUDE.md

## Project Overview

go-enphase is a Go CLI tool and library for interacting with the Enphase cloud API (v4) and local Envoy gateway. It provides commands for querying solar system data, managing authentication (OAuth2 token refresh with automatic token persistence, Envoy JWT), generating reports (daily summaries, day-over-day comparisons, month stats, month-over-month comparisons, week summaries, history export), and SSE real-time meter streaming from the local Envoy.

## Go Version

Go 1.26 (specified in `go.mod`, toolchain go1.26.4)

## Dependencies

- `github.com/spf13/cobra` - CLI framework

## Build, Test, and Lint

```bash
make build    # Build binary (outputs ./go-enphase), embeds version via ldflags
make install  # Install to $GOPATH/bin or $HOME/go/bin
make test     # Run tests with verbose output
make lint     # Run golangci-lint
make vet      # Run go vet
make clean    # Remove built binary
```

Always run `make lint` before committing and fix any issues.

Run a specific test:
```bash
go test -v ./lib -run TestRefreshAccessToken
go test -v ./cmd -run TestRootCommand
```

## Project Structure

```
main.go              # Entry point, calls cmd.Execute()
cmd/                 # CLI commands (cobra)
  root.go            # Root command, flag definitions, client constructors, config loading
  constants.go       # Command name constants
  auth.go            # auth subcommand (status, refresh, envoy-token)
  system.go          # get systems/system/summary/devices subcommands
  production.go      # get production/energy-lifetime/consumption/battery subcommands
  envoy.go           # envoy status/sensors/inverters/meters/meter-readings/stream subcommands
  report.go          # report today/summary/daily/week/month/compare/history subcommands
  helpers.go         # Shared CLI helpers (JSON output, unit conversion)
  helpers_test.go    # Tests for helpers
  root_test.go       # Tests for root command
lib/                 # Library (API client, types, business logic)
  client.go          # HTTP client (cloud + envoy), request helpers, functional options
  auth.go            # OAuth2 token refresh, auth code exchange, Envoy JWT
  system.go          # Cloud API: list systems, get system, get summary, list devices
  production.go      # Cloud API: production meter readings, energy lifetime, consumption
  battery.go         # Cloud API: battery status
  envoy.go           # Local Envoy: production, sensor readings, inverters, meters
  config.go          # Config file loading/saving (~/.enphase/config), atomic writes
  format.go          # Human-readable table/summary formatters for all data types
  report.go          # Report formatting and statistics (daily, week, month, compare, history)
  retry.go           # HTTP retry transport with exponential backoff for GET requests
  stream.go          # SSE real-time meter streaming from local Envoy with auto-reconnect
  structs.go         # API response types and data structures
  *_test.go          # Unit tests for each module
docs/                # Detailed documentation
  cli-reference.md   # All CLI commands with examples
  configuration.md   # Config file, env vars, CLI flags, cloud and Envoy setup
  library-usage.md   # Go library examples, context-aware calls, SSE streaming
  api-coverage.md    # Endpoint-to-method mapping for cloud, Envoy, and auth APIs
  ha-exporter.md     # Prometheus + Home Assistant MQTT exporter
examples/
  ha-exporter/       # Standalone Prometheus + Home Assistant MQTT exporter
```

## CLI Command Tree

```
enphase
  get
    systems              # List all systems
    system               # Get details for a specific system
    summary              # Get system summary
    devices              # List devices in a system
    production           # Get production meter readings
    energy-lifetime      # Get lifetime energy production history
    consumption          # Get lifetime consumption history
    battery              # Get battery status
  auth
    status               # Show token status (no secrets displayed)
    refresh              # Refresh access token (--save to persist)
    envoy-token          # Get Envoy JWT via Enlighten login
  envoy
    status               # Get local Envoy production and consumption
    sensors              # Get local Envoy sensor readings
    inverters            # Get per-inverter production data
    meters               # Get meter configuration
    meter-readings       # Get latest meter readings
    stream               # Stream live meter events (SSE, Ctrl+C to stop)
  report
    today                # Today's production snapshot
    summary              # Day-over-day comparison + current month stats
    daily                # Daily report with trailing history table (--days N)
    week                 # Last N complete days with totals/averages (--days N)
    month [YYYY-MM]      # Aggregate stats for a calendar month
    compare [M1] [M2]    # Side-by-side month comparison
    history              # Export full production/consumption history to JSON
  version                # Print version
```

Global flags: `--output json` (`-o json`), `--system-id`, `--api-key`, `--access-token`, `--refresh-token`, `--client-id`, `--client-secret`, `--envoy-ip`, `--envoy-token`, `--envoy-serial`, `--config`, `--rate` (on report commands).

## Configuration

Credentials can be provided via CLI flags, environment variables, or a config file at `~/.enphase/config` (KEY=VALUE format). Precedence: CLI flags > config file > environment variables.

All environment variables / config keys:
- `ENPHASE_API_KEY` - Enphase Developer API key
- `ENPHASE_ACCESS_TOKEN` - OAuth2 access token
- `ENPHASE_REFRESH_TOKEN` - OAuth2 refresh token (for automatic token refresh)
- `ENPHASE_CLIENT_ID` - OAuth2 client ID
- `ENPHASE_CLIENT_SECRET` - OAuth2 client secret
- `ENPHASE_SYSTEM_ID` - Default system ID (auto-detected from first system if omitted)
- `ENPHASE_REDIRECT_URI` - OAuth2 redirect URI
- `ENPHASE_RATE_PER_KWH` - Electricity rate for dollar estimates in reports
- `ENPHASE_ENVOY_IP` - Local Envoy gateway IP address (also accepts `ENVOY_IP`)
- `ENPHASE_ENVOY_TOKEN` - Local Envoy JWT token (also accepts `ENVOY_TOKEN`)
- `ENPHASE_ENVOY_SERIAL` - Envoy serial number (also accepts `ENVOY_SERIAL`)

Token auto-refresh: when a refresh token, client ID, and client secret are configured, the client automatically refreshes expired access tokens and persists them back to the config file.

## ha-exporter (examples/)

Standalone binary that polls Enphase APIs and exports metrics to Prometheus and Home Assistant via MQTT.

```bash
cd examples/ha-exporter
go build .
./ha-exporter --config config.json
```

See `examples/ha-exporter/README.md` for full configuration and MQTT details.

## CI

GitHub Actions workflows run on push/PR to main:
- **Lint** - golangci-lint with goconst, gocritic, gocyclo, misspell, unparam enabled
- **Test** - `go test -v -race -coverprofile` with Codecov upload
- **Build** - Compile binary and verify `--help` runs
- **macOS** - Separate macOS build/test workflow
- **Release** - Cross-compile linux/darwin (amd64/arm64) binaries on tag publish

## Linter Configuration

golangci-lint v2 config in `.golangci.yml`. Enabled linters: goconst, gocritic, gocyclo (min complexity 15), misspell (US locale), unparam. Formatters: gofmt, goimports. errcheck and goconst are excluded in test files.

## Git Conventions

- Do not add "Co-Authored-By" lines to commit messages.
- Run `make lint` before committing and fix any issues.
