# Configuration

Credentials can be provided via CLI flags, environment variables, or a config file.

## Config file

Default location: `~/.enphase/config` (KEY=VALUE format, one per line, `#` comments supported).

| Key | Env var | CLI flag | Description |
|---|---|---|---|
| `ENPHASE_API_KEY` | `ENPHASE_API_KEY` | `--api-key` | Enphase Developer API key |
| `ENPHASE_ACCESS_TOKEN` | `ENPHASE_ACCESS_TOKEN` | `--access-token` | OAuth2 access token |
| `ENPHASE_REFRESH_TOKEN` | `ENPHASE_REFRESH_TOKEN` | `--refresh-token` | OAuth2 refresh token |
| `ENPHASE_CLIENT_ID` | `ENPHASE_CLIENT_ID` | `--client-id` | OAuth2 client ID |
| `ENPHASE_CLIENT_SECRET` | `ENPHASE_CLIENT_SECRET` | `--client-secret` | OAuth2 client secret |
| `ENPHASE_SYSTEM_ID` | `ENPHASE_SYSTEM_ID` | `--system-id` | Enphase system ID |
| `ENPHASE_RATE_PER_KWH` | `ENPHASE_RATE_PER_KWH` | `--rate` | Electricity rate ($/kWh) for reports |
| `ENPHASE_TIMEZONE` | `ENPHASE_TIMEZONE` | `--timezone` | IANA time zone for report dates (e.g. `America/Chicago`) |
| `ENPHASE_REDIRECT_URI` | — | — | OAuth2 redirect URI (config file only) |
| `ENPHASE_ENVOY_IP` | `ENPHASE_ENVOY_IP` | `--envoy-ip` | Local Envoy gateway IP address |
| `ENPHASE_ENVOY_TOKEN` | `ENPHASE_ENVOY_TOKEN` | `--envoy-token` | Local Envoy JWT token |
| `ENPHASE_ENVOY_SERIAL` | `ENPHASE_ENVOY_SERIAL` | `--envoy-serial` | Envoy serial number |

**Precedence:** CLI flags > env vars > config file values.

The config file also accepts the shorter keys `ENVOY_IP`, `ENVOY_TOKEN`, and `ENVOY_SERIAL` (without the `ENPHASE_` prefix) as aliases for the Envoy settings.

Use a custom config file with `--config /path/to/file`.

## Cloud API

Obtain an API key and OAuth2 access token from the [Enphase Developer Portal](https://developer-v4.enphase.com/).

```bash
# ~/.enphase/config
ENPHASE_API_KEY=your-api-key
ENPHASE_ACCESS_TOKEN=your-access-token
ENPHASE_SYSTEM_ID=12345
```

## Local Envoy

Set `ENPHASE_ENVOY_IP` to your Envoy's IP address. A JWT is required for most endpoints:

```bash
go-enphase auth envoy-token --email user@example.com --password mypass --envoy-serial ABC123
```
