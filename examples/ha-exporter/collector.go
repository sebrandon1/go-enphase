package main

import (
	"context"
	"sync"
	"time"

	"github.com/sebrandon1/go-enphase/lib"
)

// Snapshot holds a point-in-time view of all collected metrics.
type Snapshot struct {
	CurrentPowerW   float64
	EnergyTodayWh   float64
	EnergyLifetimeWh float64
	NetPowerW        float64
	InverterWatts    map[string]float64
	UpdatedAt        time.Time
}

// Collector polls Enphase cloud and local Envoy APIs and maintains a Snapshot.
type Collector struct {
	cfg     *Config
	client  *lib.Client
	envoy   *lib.Client
	mu      sync.RWMutex
	snap    Snapshot
}

// NewCollector creates a Collector from the given config.
func NewCollector(cfg *Config) (*Collector, error) {
	var cloud *lib.Client
	var err error
	if cfg.APIKey != "" && cfg.AccessToken != "" {
		cloud, err = lib.NewClient(cfg.APIKey, cfg.AccessToken)
		if err != nil {
			return nil, err
		}
	}

	var envoy *lib.Client
	if cfg.EnvoyIP != "" {
		envoy, err = lib.NewEnvoyClient(cfg.EnvoyIP, cfg.EnvoyToken)
		if err != nil {
			return nil, err
		}
	}

	return &Collector{cfg: cfg, client: cloud, envoy: envoy}, nil
}

// Snapshot returns the most recently collected metrics.
func (c *Collector) Snapshot() Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.snap
}

// Run starts the polling loop, blocking until ctx is cancelled.
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
	var (
		currentPower   float64
		energyToday    float64
		energyLifetime float64
		netPower       float64
		inverters      map[string]float64
	)

	if c.client != nil && c.cfg.SystemID != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, err := c.client.GetSystemSummaryCtx(ctx, c.cfg.SystemID)
			if err != nil {
				Warn("cloud summary error: %v", err)
				return
			}
			currentPower = float64(summary.CurrentPower)
			energyToday = float64(summary.EnergyToday)
			energyLifetime = float64(summary.EnergyLifetime)
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
			for _, p := range prod.Production {
				if p.Type == "inverters" {
					currentPower = p.WNow
				}
			}
			for _, cons := range prod.Consumption {
				if cons.MeasurementType == "total-consumption" {
					netPower = currentPower - cons.WNow
				}
			}
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
			inverters = m
		}()
	}

	wg.Wait()

	c.mu.Lock()
	c.snap = Snapshot{
		CurrentPowerW:    currentPower,
		EnergyTodayWh:    energyToday,
		EnergyLifetimeWh: energyLifetime,
		NetPowerW:        netPower,
		InverterWatts:    inverters,
		UpdatedAt:        time.Now(),
	}
	c.mu.Unlock()

	Info("collected: %.0f W current, %.0f Wh today", currentPower, energyToday)
}
