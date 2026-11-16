# ha-exporter

`examples/ha-exporter/` is a standalone binary that:

- Polls the Enphase cloud and local Envoy APIs on a configurable interval
- Exposes a Prometheus `/metrics` endpoint
- Publishes Home Assistant MQTT auto-discovery config and state values

## Config file (JSON)

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

## Running

```bash
cd examples/ha-exporter
go build .
./ha-exporter --config config.json
./ha-exporter --config config.json --dry-run   # no API calls or MQTT publishes
```

## Prometheus metrics

| Metric | Type | Description |
|---|---|---|
| `enphase_current_power_watts` | gauge | Current production (W) |
| `enphase_energy_today_wh` | gauge | Energy produced today (Wh) |
| `enphase_energy_lifetime_wh` | counter | Lifetime energy (Wh) |
| `enphase_net_power_watts` | gauge | Net power: production - consumption (W) |
| `enphase_inverter_watts{serial="..."}` | gauge | Per-inverter production (W) |

## Home Assistant integration

The exporter publishes MQTT auto-discovery messages compatible with the [MQTT integration](https://www.home-assistant.io/integrations/mqtt/). Ensure your HA instance has MQTT configured and the broker address matches your config.

**Broker resilience:** The MQTT connection is non-blocking at startup — the exporter serves `/metrics` immediately even if the broker is down. It retries the connection every 30 s. Discovery messages are published automatically on every successful (re)connection, so a broker restart or one that loses retained messages will trigger a fresh discovery publish without restarting the exporter.

If MQTT is not connected during a state publish cycle, one warning is logged per outage and the publish is skipped until the connection recovers.

## Health check

`GET /healthz` returns `200 OK` while metrics are fresh and `503 Service Unavailable` with body `stale: no successful collection within 2x poll interval` once they age out.

When a poll round fails for one or more sources, the exporter **retains the last known values** for those sources rather than zeroing them. `UpdatedAt` advances only when at least one source succeeds, so a full outage (all sources fail) will eventually trigger the `/healthz` staleness check.

## Prometheus scrape config

```yaml
scrape_configs:
  - job_name: enphase
    static_configs:
      - targets: ['localhost:9090']
```
