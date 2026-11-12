package main

import (
	"context"
	"sync"
	"time"

	"github.com/sebrandon1/go-enphase/lib"
)

// Snapshot holds a point-in-time view of all collected metrics.
type Snapshot struct {
	CurrentPowerW    float64
	EnergyTodayWh    float64
	EnergyLifetimeWh float64
	NetPowerW        float64
	InverterWatts    map[string]float64
	UpdatedAt        time.Time
}

// collectResult holds the values polled from each source plus per-source success flags.
type collectResult struct {
	cloudPower     float64
	energyToday    float64
	energyLifetime float64
	envoyPower     float64
	netPower       float64
	inverters      map[string]float64
	cloudOK        bool
	envoyOK        bool
	invertersOK    bool
}

// mergeSnapshot merges res into prev, copying each source's values only when that source
// succeeded. UpdatedAt advances only when at least one source succeeded.
func mergeSnapshot(prev Snapshot, res collectResult, now time.Time) Snapshot {
	next := prev
	if res.cloudOK {
		next.EnergyTodayWh = res.energyToday
		next.EnergyLifetimeWh = res.energyLifetime
		// Use cloud power only when Envoy data was not collected this round.
		if !res.envoyOK {
			next.CurrentPowerW = res.cloudPower
		}
	}
	if res.envoyOK {
		next.CurrentPowerW = res.envoyPower
		next.NetPowerW = res.netPower
	}
	if res.invertersOK {
		next.InverterWatts = res.inverters
	}
	if res.cloudOK || res.envoyOK || res.invertersOK {
		next.UpdatedAt = now
	}
	return next
}

// Collector polls Enphase cloud and local Envoy APIs and maintains a Snapshot.
type Collector struct {
	cfg    *Config
	client *lib.Client
	envoy  *lib.Client
	mu     sync.RWMutex
	snap   Snapshot
}

// NewCollector creates a Collector from the given config.
func NewCollector(cfg *Config, configPath string) (*Collector, error) {
	var cloud *lib.Client
	var err error
	if cfg.APIKey != "" && cfg.AccessToken != "" {
		if cfg.RefreshToken != "" && cfg.ClientID != "" && cfg.ClientSecret != "" {
			cloud, err = lib.NewClientWithRefresh(cfg.APIKey, cfg.AccessToken, cfg.RefreshToken, cfg.ClientID, cfg.ClientSecret)
			if err != nil {
				return nil, err
			}
			cloud.OnTokenRefresh = func(accessToken, refreshToken string) {
				if saveErr := cfg.SaveTokens(configPath, accessToken, refreshToken); saveErr != nil {
					Warn("failed to persist refreshed tokens: %v", saveErr)
				}
			}
		} else {
			cloud, err = lib.NewClient(cfg.APIKey, cfg.AccessToken)
			if err != nil {
				return nil, err
			}
		}
	}

	var envoy *lib.Client
	if cfg.EnvoyIP != "" {
		envoy, err = lib.NewEnvoyClient(cfg.EnvoyIP, cfg.EnvoyToken)
		if err != nil {
			return nil, err
		}
	}

	if cloud == nil {
		Warn("cloud client not configured: set api_key and access_token to enable cloud polling")
	} else if cfg.SystemID == "" {
		Warn("cloud client configured but system_id is not set: cloud summary metrics will be skipped")
	}
	if envoy == nil {
		Warn("Envoy client not configured: set envoy_ip to enable local Envoy polling")
	}

	return &Collector{cfg: cfg, client: cloud, envoy: envoy}, nil
}

// Snapshot returns the most recently collected metrics.
func (c *Collector) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.snap
}

// Run starts the polling loop, blocking until ctx is canceled.
func (c *Collector) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	c.collect(ctx)
	for {
		select {
		case <-ticker.C:
			c.collect(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (c *Collector) collect(ctx context.Context) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var res collectResult

	if c.client != nil && c.cfg.SystemID != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, err := c.client.GetSystemSummaryCtx(ctx, c.cfg.SystemID)
			if err != nil {
				Warn("cloud summary error: %v", err)
				return
			}
			mu.Lock()
			res.cloudPower = float64(summary.CurrentPower)
			res.energyToday = float64(summary.EnergyToday)
			res.energyLifetime = float64(summary.EnergyLifetime)
			res.cloudOK = true
			mu.Unlock()
		}()
	}

	if c.envoy != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			prod, err := c.envoy.GetEnvoyProductionCtx(ctx)
			if err != nil {
				Warn("envoy production error: %v", err)
				return
			}
			var envoyPower, netPower float64
			for _, p := range prod.Production {
				if p.Type == "inverters" {
					envoyPower = p.WNow
				}
			}
			for _, cons := range prod.Consumption {
				if cons.MeasurementType == "total-consumption" {
					netPower = envoyPower - cons.WNow
				}
			}
			mu.Lock()
			res.envoyPower = envoyPower
			res.netPower = netPower
			res.envoyOK = true
			mu.Unlock()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			readings, err := c.envoy.GetInverterReadingsCtx(ctx)
			if err != nil {
				Warn("inverter readings error: %v", err)
				return
			}
			m := make(map[string]float64, len(readings))
			for _, inv := range readings {
				m[inv.SerialNumber] = float64(inv.LastReportWatts)
			}
			mu.Lock()
			res.inverters = m
			res.invertersOK = true
			mu.Unlock()
		}()
	}

	wg.Wait()

	c.mu.Lock()
	prev := c.snap
	c.snap = mergeSnapshot(prev, res, time.Now())
	c.mu.Unlock()

	if res.cloudOK || res.envoyOK || res.invertersOK {
		Info("collected: %.0f W current, %.0f Wh today", c.snap.CurrentPowerW, c.snap.EnergyTodayWh)
	} else {
		Warn("all sources failed; retaining last known values")
	}
}
