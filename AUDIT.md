# go-enphase Codebase Audit

## Cloud API Endpoint Coverage

| Endpoint | Library Method | Status |
|---|---|---|
| `GET /api/v4/systems` | `ListSystems()` / `ListSystemsCtx()` | Implemented |
| `GET /api/v4/systems/{id}` | `GetSystem()` / `GetSystemCtx()` | Implemented |
| `GET /api/v4/systems/{id}/summary` | `GetSystemSummary()` / `GetSystemSummaryCtx()` | Implemented |
| `GET /api/v4/systems/{id}/devices` | `ListDevices()` / `ListDevicesCtx()` | Implemented |
| `GET /api/v4/systems/{id}/production_meter_readings` | `GetProductionMeterReadings()` / `GetProductionMeterReadingsCtx()` | Implemented |
| `GET /api/v4/systems/{id}/energy_lifetime` | `GetEnergyLifetime()` / `GetEnergyLifetimeCtx()` | Implemented |
| `GET /api/v4/systems/{id}/consumption_lifetime` | `GetConsumptionLifetime()` / `GetConsumptionLifetimeCtx()` | Implemented |
| `GET /api/v4/systems/{id}/battery_lifetime` | `GetBatteryStatus()` / `GetBatteryStatusCtx()` | Implemented |

## Local Envoy Endpoint Coverage

| Endpoint | Library Method | Status |
|---|---|---|
| `GET /production.json` | `GetEnvoyProduction()` / `GetEnvoyProductionCtx()` | Implemented |
| `GET /ivp/sensors/readings_object` | `GetEnvoySensors()` / `GetEnvoySensorsCtx()` | Implemented |
| `GET /api/v1/production` | `GetEnvoySimpleProduction()` / `GetEnvoySimpleProductionCtx()` | Implemented |
| `GET /api/v1/production/inverters` | `GetInverterReadings()` / `GetInverterReadingsCtx()` | Implemented |
| `GET /ivp/meters` | `GetMeterConfig()` / `GetMeterConfigCtx()` | Implemented |
| `GET /ivp/meters/readings` | `GetMeterReadings()` / `GetMeterReadingsCtx()` | Implemented |
| `GET /stream/meter` (SSE) | `StreamMeter()` | Implemented |

## Authentication Flow Coverage

| Flow | Method | Status |
|---|---|---|
| OAuth2 token refresh | `RefreshAccessToken()` / `RefreshAccessTokenCtx()` | Implemented |
| Authorization code exchange | `ExchangeAuthCode()` / `ExchangeAuthCodeCtx()` | Implemented |
| Envoy JWT acquisition | `GetEnvoyToken()` / `GetEnvoyTokenCtx()` | Implemented |
| Envoy JWT caching + auto-refresh | `EnsureEnvoyToken()` | Implemented |

## Issues Identified and Resolved

| Issue | Location | Resolution |
|---|---|---|
| No `context.Context` support | All `lib/*.go` HTTP methods | Added `*Ctx` variants; old methods delegate to `context.Background()` |
| No functional options | `lib/client.go` | Added `ClientOption`, `WithTimeout`, `WithHTTPClient`, `WithInsecureSkipVerify`, `NewClientWithOptions` |
| No retry logic | `lib/client.go` | Added `retryTransport` in `lib/retry.go` wrapping all HTTP clients |
| No JWT caching | `lib/auth.go` | Added `EnsureEnvoyToken` with mutex + expiry check; `parseJWTExpiry` without external deps |
| `GetEnvoyToken` bare-string bug | `lib/auth.go` | Entrez returns bare JWT (not JSON); changed from `decodeJSON` to `io.ReadAll` |
| `devicesResponse` silently drops non-microinverter devices | `lib/structs.go` | Documented; only `micro_inverters` field is populated — other device types are not returned by this endpoint |
| No response body draining | `lib/client.go` | Added `drainAndClose` used by all ctx helpers; enables HTTP keep-alive reuse |
| Missing Envoy endpoints | `lib/envoy.go` | Added 4 new endpoint pairs + SSE stream |
| README incorrectly states "No config file required" | `README.md` | Fixed with full config documentation |
