# API Coverage

## Cloud API (v4)

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

## Local Envoy

| Endpoint | Method |
|---|---|
| `GET /production.json` | `GetEnvoyProduction[Ctx]()` |
| `GET /ivp/sensors/readings_object` | `GetEnvoySensors[Ctx]()` |
| `GET /api/v1/production` | `GetEnvoySimpleProduction[Ctx]()` |
| `GET /api/v1/production/inverters` | `GetInverterReadings[Ctx]()` |
| `GET /ivp/meters` | `GetMeterConfig[Ctx]()` |
| `GET /ivp/meters/readings` | `GetMeterReadings[Ctx]()` |
| `GET /stream/meter` (SSE) | `StreamMeter(ctx, handler)` |

## Authentication

| Flow | Method |
|---|---|
| Token refresh | `RefreshAccessToken[Ctx]()` |
| Auth code exchange | `ExchangeAuthCode[Ctx]()` |
| Envoy JWT acquisition | `GetEnvoyToken[Ctx]()` |
| Envoy JWT caching | `EnsureEnvoyToken(ctx, ...)` |
