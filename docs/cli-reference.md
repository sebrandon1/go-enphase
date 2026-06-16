# CLI Reference

## Cloud API commands

```bash
# List all systems
go-enphase get systems

# Get system summary
go-enphase get summary --system-id 12345

# List devices
go-enphase get devices --system-id 12345

# Get production meter readings
go-enphase get production --system-id 12345

# Get lifetime energy production
go-enphase get energy-lifetime --system-id 12345 --start-date 2024-01-01 --end-date 2024-12-31

# Get lifetime consumption
go-enphase get consumption --system-id 12345

# Get battery status
go-enphase get battery --system-id 12345
```

## Authentication commands

```bash
# Check token status
go-enphase auth status

# Refresh access token
go-enphase auth refresh --refresh-token $REFRESH --client-id $CID --client-secret $CS

# Get Envoy JWT token
go-enphase auth envoy-token --email user@example.com --password mypass --envoy-serial ABC123
```

## Report commands

```bash
# Today's production summary
go-enphase report today --system-id 12345

# Daily report: today's live status + last 7 days production vs consumption
go-enphase report daily --system-id 12345

# Daily report with custom trailing days
go-enphase report daily --system-id 12345 --days 14

# Compare two months of production
go-enphase report compare --system-id 12345 2025-01 2025-02

# Export full production/consumption history to JSON
go-enphase report history --system-id 12345 --output ~/solar/history.json
```

All report commands support `--rate 0.13` to include dollar estimates (also configurable via `ENPHASE_RATE_PER_KWH` in the config file).

## Local Envoy commands

```bash
# Get production/consumption
go-enphase envoy status --envoy-ip 192.168.1.100

# Get sensor readings
go-enphase envoy sensors --envoy-ip 192.168.1.100 --envoy-token $JWT
```
