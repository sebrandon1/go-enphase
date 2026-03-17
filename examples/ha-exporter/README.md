# ha-exporter

Standalone binary that exports Enphase solar metrics to Prometheus and Home Assistant via MQTT.

## Features

- Prometheus `/metrics` endpoint with 5 metric types
- Home Assistant MQTT auto-discovery
- Parallel cloud + Envoy polling
- Exponential-backoff reconnection for the SSE stream
- `--dry-run` mode for safe testing
- Graceful shutdown on SIGINT / SIGTERM
- ANSI color-coded logs to stderr; respects `NO_COLOR`

## Build

```bash
cd examples/ha-exporter
go build .
```

## Config file

Create a `config.json`:

```json
{
  "api_key": "your-enphase-api-key",
  "access_token": "your-oauth2-access-token",
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

All MQTT fields are optional — omit them to disable MQTT publishing.

## Running

```bash
./ha-exporter --config config.json

# Override metrics address
./ha-exporter --config config.json --metrics-addr :8080

# Dry run (no API calls, no MQTT publishes)
./ha-exporter --config config.json --dry-run
```

## Prometheus scrape config

```yaml
scrape_configs:
  - job_name: enphase
    scrape_interval: 30s
    static_configs:
      - targets: ['localhost:9090']
```

## Prometheus metrics

| Metric | Type | Labels | Description |
|---|---|---|---|
| `enphase_current_power_watts` | gauge | — | Current production in watts |
| `enphase_energy_today_wh` | gauge | — | Energy produced today (Wh) |
| `enphase_energy_lifetime_wh` | counter | — | Lifetime energy production (Wh) |
| `enphase_net_power_watts` | gauge | — | Net power: production minus consumption (W) |
| `enphase_inverter_watts` | gauge | `serial` | Per-inverter current production (W) |

## Home Assistant MQTT integration

The exporter publishes HA MQTT auto-discovery config on startup:

- Config topic: `{prefix}/sensor/solar_{serial}/config`
- State topic: `{prefix}/sensor/solar_{serial}/state`

In Home Assistant, ensure the MQTT integration is configured and the broker matches your config. Sensors will appear automatically in the MQTT integration device list after the first publish.

### Example state payload

```json
{
  "current_power_w": 4250.0,
  "energy_today_wh": 18400.0,
  "energy_lifetime_wh": 9870000.0,
  "net_power_w": 1800.0
}
```
