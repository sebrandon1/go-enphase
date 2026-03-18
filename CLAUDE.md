# CLAUDE.md

## Project Overview

go-enphase is a Go CLI tool and library for interacting with the Enphase cloud API (v4) and local Envoy gateway. It provides commands for querying solar system data, managing authentication (OAuth2 token refresh, Envoy JWT), and generating reports (daily summaries, month-over-month comparisons).

## Go Version

Go 1.24 (specified in `go.mod`)

## Dependencies

- `github.com/spf13/cobra` - CLI framework

## Build, Test, and Lint

```bash
make build    # Build binary (outputs ./go-enphase)
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
  root.go            # Root command, flag definitions, client constructors
  auth.go            # auth subcommand (status, refresh, envoy-token)
  system.go          # get systems/summary/devices subcommands
  production.go      # get production/energy-lifetime/consumption subcommands
  envoy.go           # envoy status/sensors subcommands
  report.go          # report today/daily/compare/history subcommands
  helpers.go         # Shared CLI helpers (formatting, output)
  helpers_test.go    # Tests for helpers
  root_test.go       # Tests for root command
lib/                 # Library (API client, types, business logic)
  client.go          # HTTP client (cloud + envoy), request helpers
  auth.go            # OAuth2 token refresh, auth code exchange, Envoy JWT
  system.go          # Cloud API: list systems, get summary, list devices
  production.go      # Cloud API: production meter readings, energy lifetime, consumption
  battery.go         # Cloud API: battery status
  envoy.go           # Local Envoy: production, sensor readings
  config.go          # Config file loading/saving (~/.enphase/config)
  report.go          # Report formatting and statistics (daily summary, daily report, month comparison, history export)
  structs.go         # API response types and data structures
  *_test.go          # Unit tests for each module
```

## Configuration

Credentials can be provided via CLI flags, environment variables, or a config file at `~/.enphase/config` (KEY=VALUE format). CLI flags take precedence over config file values.

Key environment variables:
- `ENPHASE_API_KEY` - Enphase Developer API key
- `ENPHASE_ACCESS_TOKEN` - OAuth2 access token
- `ENPHASE_ENVOY_IP` - Local Envoy gateway IP address

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

## Linter Configuration

golangci-lint v2 config in `.golangci.yml`. Enabled linters: goconst, gocritic, gocyclo (min complexity 15), misspell (US locale), unparam. Formatters: gofmt, goimports. errcheck and goconst are excluded in test files.

## Git Conventions

- Do not add "Co-Authored-By" lines to commit messages.
- Run `make lint` before committing and fix any issues.
