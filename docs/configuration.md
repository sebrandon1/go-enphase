# Configuration

Credentials can be provided via CLI flags, environment variables, or a config file.

## Config file

Default location: `~/.enphase/config` (KEY=VALUE format, one per line, `#` comments supported).

| Key | Env var | CLI flag | Description |
|---|---|---|---|
| `ENPHASE_API_KEY` | `ENPHASE_API_KEY` | `--api-key` | Enphase Developer API key |
| `ENPHASE_ACCESS_TOKEN` | `ENPHASE_ACCESS_TOKEN` | `--access-token` | OAuth2 access token |
| `ENPHASE_REFRESH_TOKEN` | — | `--refresh-token` | OAuth2 refresh token |
| `ENPHASE_CLIENT_ID` | — | `--client-id` | OAuth2 client ID |
| `ENPHASE_CLIENT_SECRET` | — | `--client-secret` | OAuth2 client secret |
| `ENPHASE_SYSTEM_ID` | — | `--system-id` | Enphase system ID |
| `ENPHASE_RATE_PER_KWH` | — | — | Electricity rate ($/kWh) for reports |
| `ENPHASE_REDIRECT_URI` | — | — | OAuth2 redirect URI |

**Precedence:** CLI flags > config file values.

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
