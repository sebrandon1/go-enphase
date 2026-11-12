package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

var productionRoutes = routes{
	systemPath + "/production_meter_readings": `{"meter_readings":[{"serial_number":"M1","value":100,"read_at":1700000000}]}`,
	systemPath + "/energy_lifetime":           `{"start_date":"2026-01-01","system_id":123,"production":[1000,2000]}`,
	systemPath + "/consumption_lifetime":      `{"start_date":"2026-01-01","system_id":123,"consumption":[500,600]}`,
	systemPath + "/battery_lifetime":          `{}`,
}

func TestProductionCommands(t *testing.T) {
	tests := []struct {
		name string
		cmd  func() (string, error)
	}{
		{"production", func() (string, error) { return runCmd(t, productionCmd) }},
		{"energy-lifetime", func() (string, error) { return runCmd(t, energyLifetimeCmd) }},
		{"consumption", func() (string, error) { return runCmd(t, consumptionCmd) }},
	}
	for _, tt := range tests {
		t.Run(tt.name+"/text", func(t *testing.T) {
			useCloud(t, productionRoutes, http.StatusOK)
			out, err := tt.cmd()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.TrimSpace(out) == "" {
				t.Error("expected text output, got none")
			}
		})
		t.Run(tt.name+"/json", func(t *testing.T) {
			useCloud(t, productionRoutes, http.StatusOK)
			outputFormat = outputJSON
			out, err := tt.cmd()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !json.Valid([]byte(out)) {
				t.Errorf("expected valid JSON, got %q", out)
			}
		})
		t.Run(tt.name+"/server error", func(t *testing.T) {
			useCloud(t, productionRoutes, http.StatusBadRequest)
			if _, err := tt.cmd(); err == nil {
				t.Error("expected error on HTTP 400")
			}
		})
	}
}

func TestEnergyLifetimeInvalidDate(t *testing.T) {
	useCloud(t, productionRoutes, http.StatusOK)
	orig := energyStartDate
	energyStartDate = "01/02/2026"
	t.Cleanup(func() { energyStartDate = orig })
	if _, err := runCmd(t, energyLifetimeCmd); err == nil {
		t.Error("expected error for malformed --start-date")
	}
}

func TestConsumptionInvalidDate(t *testing.T) {
	useCloud(t, productionRoutes, http.StatusOK)
	orig := consumptionEndDate
	consumptionEndDate = "tomorrow"
	t.Cleanup(func() { consumptionEndDate = orig })
	if _, err := runCmd(t, consumptionCmd); err == nil {
		t.Error("expected error for malformed --end-date")
	}
}
