# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- **Functional options** — `ClientOption` type and `NewClientWithOptions` constructor with `WithTimeout`, `WithHTTPClient`, and `WithInsecureSkipVerify` options (`lib/client.go`).
- **Context propagation** — `*Ctx` variants for every public HTTP-calling method in `lib/` (e.g. `ListSystemsCtx`, `GetBatteryStatusCtx`). Existing methods are backward-compatible one-line delegates to `context.Background()`.
- **Retry logic** — `retryTransport` in `lib/retry.go` wraps all HTTP clients. Retries GET requests on 5xx responses and transient network timeouts with exponential backoff (base 500 ms, cap 30 s, max 3 retries). Context cancellation interrupts the backoff immediately.
- **Response body draining** — `drainAndClose` helper in `lib/client.go` properly drains and closes response bodies in all context-aware request helpers, enabling HTTP keep-alive connection reuse.
- **Envoy JWT caching** — `EnsureEnvoyToken(ctx, email, password, serial)` on `Client`; acquires a mutex, checks expiry (with 5-minute buffer), and refreshes only when needed. `parseJWTExpiry` decodes the JWT `exp` claim without external dependencies.
- **New Envoy endpoints** — four new endpoint pairs (plain + `Ctx`): `GetEnvoySimpleProduction`, `GetInverterReadings`, `GetMeterConfig`, `GetMeterReadings`.
- **SSE streaming** — `StreamMeter(ctx, handler)` in `lib/stream.go` connects to `/stream/meter` on the local Envoy, calls the handler for each `StreamMeterEvent`, and automatically reconnects with exponential backoff on error.
- **New struct types** — `EnvoySimpleProduction`, `InverterReading`, `MeterConfig`, `MeterData`, `StreamMeterEvent` in `lib/structs.go`.
- **ha-exporter example** — `examples/ha-exporter/` sub-module: Prometheus metrics server (`/metrics`), Home Assistant MQTT auto-discovery, parallel cloud + Envoy polling with `sync.WaitGroup`, graceful shutdown on SIGINT/SIGTERM, `--dry-run` mode.
- **Test coverage** — retry transport tests, functional options tests, JWT caching tests, new Envoy endpoint tests, SSE stream tests. `lib/` coverage increased to 87%+.
- **AUDIT.md** — documents all API endpoint coverage, auth flows, and issues resolved.

### Fixed

- `GetEnvoyToken`: Entrez `/tokens` endpoint returns a bare JWT string, not JSON-wrapped. Fixed by using `io.ReadAll` instead of `decodeJSON` in the second step of the Enlighten/Entrez flow.
- **README**: corrected incorrect claim "No config file required" — the config file at `~/.enphase/config` has been supported since an earlier release. Updated with full key/env-var/flag precedence table.

### Changed

- `lib/system.go`, `lib/production.go`, `lib/battery.go`, `lib/envoy.go`, `lib/auth.go` — existing public methods now delegate to their `*Ctx` counterparts for DRY implementation.
- `newHTTPClientWithTLS` now wraps the transport with `retryTransport` automatically; all three existing constructors and `NewClientWithOptions` inherit retry behaviour.
