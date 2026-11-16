package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// captureStderr runs f and returns anything written to os.Stderr.
func captureStderr(f func()) string {
	orig := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f()

	w.Close()
	os.Stderr = orig

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestMergeSnapshot(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Minute)

	prev := Snapshot{
		CurrentPowerW:    1000,
		EnergyTodayWh:    5000,
		EnergyLifetimeWh: 100000,
		NetPowerW:        800,
		UpdatedAt:        t0,
	}

	tests := []struct {
		name        string
		res         collectResult
		wantPower   float64
		wantToday   float64
		wantUpdated bool // true = UpdatedAt should advance to t1
	}{
		{
			name:        "all sources fail — values and UpdatedAt unchanged",
			res:         collectResult{},
			wantPower:   1000,
			wantToday:   5000,
			wantUpdated: false,
		},
		{
			name: "cloud succeeds — energy values update, UpdatedAt advances",
			res: collectResult{
				cloudPower: 1200, energyToday: 6000, energyLifetime: 110000,
				cloudOK: true,
			},
			wantPower:   1200, // cloud power used because envoyOK=false
			wantToday:   6000,
			wantUpdated: true,
		},
		{
			name: "envoy succeeds, cloud fails — power updates, energy retained",
			res: collectResult{
				envoyPower: 1500, netPower: 300,
				envoyOK: true,
			},
			wantPower:   1500,
			wantToday:   5000, // cloud failed, retain previous
			wantUpdated: true,
		},
		{
			name: "both succeed — envoy power wins",
			res: collectResult{
				cloudPower: 1200, energyToday: 6000, energyLifetime: 110000, cloudOK: true,
				envoyPower: 1500, netPower: 300, envoyOK: true,
			},
			wantPower:   1500, // envoy wins
			wantToday:   6000,
			wantUpdated: true,
		},
		{
			name: "envoy power is zero (night) — still updates",
			res: collectResult{
				envoyPower: 0, netPower: -200,
				envoyOK: true,
			},
			wantPower:   0,
			wantToday:   5000,
			wantUpdated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeSnapshot(prev, tt.res, t1)
			if got.CurrentPowerW != tt.wantPower {
				t.Errorf("CurrentPowerW = %.0f, want %.0f", got.CurrentPowerW, tt.wantPower)
			}
			if got.EnergyTodayWh != tt.wantToday {
				t.Errorf("EnergyTodayWh = %.0f, want %.0f", got.EnergyTodayWh, tt.wantToday)
			}
			if tt.wantUpdated && !got.UpdatedAt.Equal(t1) {
				t.Errorf("UpdatedAt = %v, want %v", got.UpdatedAt, t1)
			}
			if !tt.wantUpdated && !got.UpdatedAt.Equal(t0) {
				t.Errorf("UpdatedAt should not advance: got %v, want %v", got.UpdatedAt, t0)
			}
		})
	}
}

func TestNewCollectorWarnsWhenCloudUnconfigured(t *testing.T) {
	cfg := &Config{} // no APIKey or AccessToken

	out := captureStderr(func() {
		_, _ = NewCollector(cfg, "")
	})

	if !strings.Contains(out, "cloud client not configured") {
		t.Errorf("expected cloud-not-configured warning, got: %s", out)
	}
}

func TestNewCollectorWarnsWhenEnvoyUnconfigured(t *testing.T) {
	cfg := &Config{
		APIKey:      "key",
		AccessToken: "token",
		SystemID:    "12345",
		// no EnvoyIP
	}

	out := captureStderr(func() {
		_, _ = NewCollector(cfg, "")
	})

	if !strings.Contains(out, "Envoy client not configured") {
		t.Errorf("expected envoy-not-configured warning, got: %s", out)
	}
}

func TestNewCollectorWarnsWhenSystemIDMissing(t *testing.T) {
	cfg := &Config{
		APIKey:      "key",
		AccessToken: "token",
		// no SystemID, no EnvoyIP
	}

	out := captureStderr(func() {
		_, _ = NewCollector(cfg, "")
	})

	if !strings.Contains(out, "system_id is not set") {
		t.Errorf("expected system_id warning, got: %s", out)
	}
}
