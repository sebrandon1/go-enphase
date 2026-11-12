package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

var reportRoutes = routes{
	systemPath + "/summary":              `{"system_id":123,"status":"normal","modules":20,"current_power":1500,"energy_today":9000,"energy_lifetime":1000000}`,
	systemPath + "/consumption_lifetime": `{"start_date":"2026-01-01","system_id":123,"consumption":[7000]}`,
}

func TestReportToday(t *testing.T) {
	useCloud(t, reportRoutes, http.StatusOK)
	out, err := runCmd(t, reportTodayCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected text output, got none")
	}
}

func TestReportTodayJSON(t *testing.T) {
	useCloud(t, reportRoutes, http.StatusOK)
	outputFormat = outputJSON
	out, err := runCmd(t, reportTodayCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got todayJSON
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON %q: %v", out, err)
	}
	if got.SystemID != 123 || got.ProductionKWh != 9.0 || got.ConsumptionKWh != 7.0 {
		t.Errorf("unexpected values: %+v", got)
	}
}

func TestReportTodayRateAddsValue(t *testing.T) {
	useCloud(t, reportRoutes, http.StatusOK)
	outputFormat = outputJSON
	orig := ratePerKWh
	ratePerKWh = 0.5
	t.Cleanup(func() { ratePerKWh = orig })
	out, err := runCmd(t, reportTodayCmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got todayJSON
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON %q: %v", out, err)
	}
	if got.ValueUSD != 4.5 {
		t.Errorf("ValueUSD = %v, want 4.5", got.ValueUSD)
	}
}

func TestReportTodaySummaryError(t *testing.T) {
	useCloud(t, reportRoutes, http.StatusBadRequest)
	_, err := runCmd(t, reportTodayCmd)
	if err == nil || !strings.Contains(err.Error(), "error getting summary") {
		t.Errorf("expected wrapped summary error, got %v", err)
	}
}

func TestReportTodayConsumptionFailureIsNonFatal(t *testing.T) {
	r := routes{systemPath + "/summary": reportRoutes[systemPath+"/summary"]}
	useCloud(t, r, http.StatusOK) // consumption path is 404
	outputFormat = outputJSON
	out, err := runCmd(t, reportTodayCmd)
	if err != nil {
		t.Fatalf("consumption failure should not fail the report: %v", err)
	}
	var got todayJSON
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON %q: %v", out, err)
	}
	if got.ConsumptionWh != 0 {
		t.Errorf("ConsumptionWh = %d, want 0", got.ConsumptionWh)
	}
}
