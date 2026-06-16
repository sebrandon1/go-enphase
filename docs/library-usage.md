# Library Usage

## Basic cloud access

```go
client, err := lib.NewClient("api-key", "access-token")
systems, err := client.ListSystems()
summary, err := client.GetSystemSummary("12345")
```

## Context-aware calls

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

summary, err := client.GetSystemSummaryCtx(ctx, "12345")
inverters, err := envoy.GetInverterReadingsCtx(ctx)
```

## Functional options

```go
client, err := lib.NewClientWithOptions("api-key", "access-token",
    lib.WithTimeout(15*time.Second),
    lib.WithInsecureSkipVerify(false),
)
```

## Local Envoy

```go
envoy, err := lib.NewEnvoyClient("192.168.1.100", jwtToken)

// Basic production
prod, err := envoy.GetEnvoyProduction()

// Per-inverter readings
inverters, err := envoy.GetInverterReadings()

// Revenue-grade meters
meters, err := envoy.GetMeterConfig()
readings, err := envoy.GetMeterReadings()
```

## SSE real-time streaming

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

envoy.StreamMeter(ctx, func(ev *lib.StreamMeterEvent, err error) {
    if err != nil {
        log.Printf("stream error: %v", err)
        return
    }
    fmt.Printf("EID %d: %.1f W at %d\n", ev.EID, ev.ActPower, ev.Timestamp)
})
```

## Envoy JWT caching

```go
// EnsureEnvoyToken fetches or reuses a cached JWT, refreshing 5 min before expiry.
err := client.EnsureEnvoyToken(ctx, "user@example.com", "password", "SERIAL123")
```
