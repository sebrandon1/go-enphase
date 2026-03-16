package lib

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatTodaySummary(t *testing.T) {
	summary := &SystemSummary{
		SystemID:       12345,
		Status:         "normal",
		CurrentPower:   3500,
		Modules:        20,
		EnergyToday:    15000,
		EnergyLifetime: 5000000,
		LastReportAt:   1709913600,
	}

	output := FormatTodaySummary(summary, 0, 0.12)

	if !strings.Contains(output, "System 12345") {
		t.Error("Expected system ID in output")
	}
	if !strings.Contains(output, "normal") {
		t.Error("Expected status in output")
	}
	if !strings.Contains(output, "3500 W") {
		t.Error("Expected current power in output")
	}
	if !strings.Contains(output, "15.00 kWh") {
		t.Error("Expected energy today in kWh")
	}
	if !strings.Contains(output, "$1.80") {
		t.Error("Expected dollar value for today")
	}
	if !strings.Contains(output, "5000.0 kWh") {
		t.Error("Expected lifetime energy in kWh")
	}
	if !strings.Contains(output, "$600.00") {
		t.Error("Expected dollar value for lifetime")
	}
}

func TestFormatTodaySummaryWithConsumption(t *testing.T) {
	summary := &SystemSummary{
		SystemID:     12345,
		EnergyToday:  15000,
		CurrentPower: 3500,
		Modules:      20,
	}

	output := FormatTodaySummary(summary, 20000, 0.12)

	if !strings.Contains(output, "Consumption") {
		t.Error("Expected consumption line in output")
	}
	if !strings.Contains(output, "20.00 kWh") {
		t.Error("Expected consumption kWh in output")
	}
	if !strings.Contains(output, "Grid Draw") {
		t.Error("Expected grid draw line when consuming more than producing")
	}
	if !strings.Contains(output, "Solar Offset") {
		t.Error("Expected solar offset percentage")
	}
	if !strings.Contains(output, "75%") {
		t.Error("Expected 75% solar offset (15/20)")
	}
}

func TestFormatTodaySummaryNoRate(t *testing.T) {
	summary := &SystemSummary{
		SystemID:    12345,
		EnergyToday: 15000,
	}

	output := FormatTodaySummary(summary, 0, 0)
	if strings.Contains(output, "$") {
		t.Error("Expected no dollar values when rate is 0")
	}
}

func TestComputeMonthStats(t *testing.T) {
	production := []int{10000, 20000, 5000, 15000, 3000}
	stats := ComputeMonthStats(production)

	if stats.TotalKWh != 53.0 {
		t.Errorf("Expected total 53.0 kWh, got %.1f", stats.TotalKWh)
	}
	if stats.DaysCount != 5 {
		t.Errorf("Expected 5 days, got %d", stats.DaysCount)
	}
	if stats.BestKWh != 20.0 {
		t.Errorf("Expected best 20.0 kWh, got %.1f", stats.BestKWh)
	}
	if stats.WorstKWh != 3.0 {
		t.Errorf("Expected worst 3.0 kWh, got %.1f", stats.WorstKWh)
	}
	if stats.Above15 != 2 {
		t.Errorf("Expected 2 days above 15 kWh, got %d", stats.Above15)
	}
	if stats.Below5 != 1 {
		t.Errorf("Expected 1 day below 5 kWh, got %d", stats.Below5)
	}
}

func TestComputeMonthStatsEmpty(t *testing.T) {
	stats := ComputeMonthStats(nil)
	if stats.DaysCount != 0 {
		t.Errorf("Expected 0 days, got %d", stats.DaysCount)
	}
}

func TestFormatMonthComparison(t *testing.T) {
	prodA := &EnergyLifetime{Production: []int{10000, 20000, 15000}}
	prodB := &EnergyLifetime{Production: []int{12000, 18000, 16000}}

	output := FormatMonthComparison("12345", "2025-01", "2025-02", prodA, prodB, 0.12)

	if !strings.Contains(output, "2025-01") {
		t.Error("Expected month A in output")
	}
	if !strings.Contains(output, "2025-02") {
		t.Error("Expected month B in output")
	}
	if !strings.Contains(output, "Total Production") {
		t.Error("Expected 'Total Production' label")
	}
	if !strings.Contains(output, "Est. Value") {
		t.Error("Expected 'Est. Value' when rate is set")
	}
}

func TestMonthDateRange(t *testing.T) {
	start, end, err := MonthDateRange("2025-01")
	if err != nil {
		t.Fatalf("MonthDateRange failed: %v", err)
	}
	if start != "2025-01-01" {
		t.Errorf("Expected start 2025-01-01, got %s", start)
	}
	if end != "2025-01-31" {
		t.Errorf("Expected end 2025-01-31, got %s", end)
	}
}

func TestMonthDateRangeInvalid(t *testing.T) {
	_, _, err := MonthDateRange("bad")
	if err == nil {
		t.Error("Expected error for invalid month format")
	}
}

func TestLastTwoMonths(t *testing.T) {
	twoAgo, last := LastTwoMonths()
	if len(twoAgo) != 7 || len(last) != 7 {
		t.Errorf("Expected YYYY-MM format, got %q and %q", twoAgo, last)
	}
}

func TestBuildDailyRows(t *testing.T) {
	prod := &EnergyLifetime{
		StartDate:  "2025-03-01",
		Production: []int{10000, 20000, 15000, 8000, 25000},
	}
	cons := &ConsumptionLifetime{
		StartDate:   "2025-03-01",
		Consumption: []int{30000, 40000, 35000, 50000, 45000},
	}

	rows := BuildDailyRows(prod, cons, 3)
	if len(rows) != 3 {
		t.Fatalf("Expected 3 rows, got %d", len(rows))
	}

	// Should be the last 3 days: Mar 3, 4, 5
	if rows[0].Date != "2025-03-03" {
		t.Errorf("Expected first row date 2025-03-03, got %s", rows[0].Date)
	}
	if rows[0].ProductionWh != 15000 {
		t.Errorf("Expected production 15000 Wh, got %d", rows[0].ProductionWh)
	}
	if rows[0].ConsumptionWh != 35000 {
		t.Errorf("Expected consumption 35000 Wh, got %d", rows[0].ConsumptionWh)
	}
	if rows[2].NetKWh != -20.0 {
		t.Errorf("Expected net -20.0 kWh, got %.1f", rows[2].NetKWh)
	}
}

func TestBuildDailyRowsAllDays(t *testing.T) {
	prod := &EnergyLifetime{
		StartDate:  "2025-03-01",
		Production: []int{10000, 20000},
	}

	rows := BuildDailyRows(prod, nil, 0)
	if len(rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(rows))
	}
	if rows[0].ConsumptionWh != 0 {
		t.Errorf("Expected 0 consumption when no consumption data, got %d", rows[0].ConsumptionWh)
	}
}

func TestBuildDailyRowsMismatchedStarts(t *testing.T) {
	prod := &EnergyLifetime{
		StartDate:  "2025-03-01",
		Production: []int{10000, 20000, 15000},
	}
	cons := &ConsumptionLifetime{
		StartDate:   "2025-03-02",
		Consumption: []int{40000, 35000},
	}

	rows := BuildDailyRows(prod, cons, 0)
	if len(rows) != 3 {
		t.Fatalf("Expected 3 rows, got %d", len(rows))
	}
	// Mar 1: has production but no consumption
	if rows[0].ProductionWh != 10000 || rows[0].ConsumptionWh != 0 {
		t.Errorf("Mar 1: expected prod=10000 cons=0, got prod=%d cons=%d", rows[0].ProductionWh, rows[0].ConsumptionWh)
	}
	// Mar 2: has both
	if rows[1].ProductionWh != 20000 || rows[1].ConsumptionWh != 40000 {
		t.Errorf("Mar 2: expected prod=20000 cons=40000, got prod=%d cons=%d", rows[1].ProductionWh, rows[1].ConsumptionWh)
	}
}

func TestFormatDailyReport(t *testing.T) {
	summary := &SystemSummary{
		SystemID:     12345,
		Status:       "normal",
		CurrentPower: 2000,
		Modules:      12,
		EnergyToday:  8000,
		LastReportAt: 1709913600,
	}
	rows := []DailyRow{
		{Date: "2025-03-01", ProductionKWh: 10.0, ConsumptionKWh: 30.0, NetKWh: -20.0},
		{Date: "2025-03-02", ProductionKWh: 20.0, ConsumptionKWh: 40.0, NetKWh: -20.0},
	}

	output := FormatDailyReport(summary, rows, 0.13)

	if !strings.Contains(output, "Solar Daily Report") {
		t.Error("Expected 'Solar Daily Report' header")
	}
	if !strings.Contains(output, "System 12345") {
		t.Error("Expected system ID")
	}
	if !strings.Contains(output, "2000 W") {
		t.Error("Expected current power")
	}
	if !strings.Contains(output, "8.00 kWh") {
		t.Error("Expected today's energy")
	}
	if !strings.Contains(output, "$1.04") {
		t.Error("Expected today's dollar value")
	}
	if !strings.Contains(output, "2025-03-01") {
		t.Error("Expected history date")
	}
	if !strings.Contains(output, "2-Day Avg") {
		t.Error("Expected average row")
	}
	if !strings.Contains(output, "Prod $") {
		t.Error("Expected dollar column header when rate is set")
	}
}

func TestFormatDailyReportNoRate(t *testing.T) {
	summary := &SystemSummary{
		SystemID:    12345,
		EnergyToday: 8000,
	}
	rows := []DailyRow{
		{Date: "2025-03-01", ProductionKWh: 10.0, ConsumptionKWh: 30.0, NetKWh: -20.0},
	}

	output := FormatDailyReport(summary, rows, 0)
	if strings.Contains(output, "$") {
		t.Error("Expected no dollar values when rate is 0")
	}
	if strings.Contains(output, "Prod $") {
		t.Error("Expected no dollar column when rate is 0")
	}
}

func TestFormatDailyReportNoRows(t *testing.T) {
	summary := &SystemSummary{
		SystemID:    12345,
		EnergyToday: 8000,
	}

	output := FormatDailyReport(summary, nil, 0.13)
	if !strings.Contains(output, "8.00 kWh") {
		t.Error("Expected today's energy even with no history rows")
	}
	if strings.Contains(output, "Recent History") {
		t.Error("Expected no history section when no rows")
	}
}

func TestBuildHistoryRecords(t *testing.T) {
	rows := []DailyRow{
		{Date: "2025-03-01", ProductionWh: 10000, ConsumptionWh: 30000, ProductionKWh: 10.0, ConsumptionKWh: 30.0, NetKWh: -20.0},
		{Date: "2025-03-02", ProductionWh: 20500, ConsumptionWh: 40200, ProductionKWh: 20.5, ConsumptionKWh: 40.2, NetKWh: -19.7},
	}

	records := BuildHistoryRecords(rows, 0.13)
	if len(records) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(records))
	}

	r := records[0]
	if r.Date != "2025-03-01" {
		t.Errorf("Expected date 2025-03-01, got %s", r.Date)
	}
	if r.ProductionValueUSD != 1.30 {
		t.Errorf("Expected production value $1.30, got $%.2f", r.ProductionValueUSD)
	}
	if r.NetKWh != -20.0 {
		t.Errorf("Expected net -20.0, got %.1f", r.NetKWh)
	}
}

func TestWriteHistoryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test_history.json")

	records := []HistoryRecord{
		{Date: "2025-03-01", ProductionWh: 10000, ConsumptionWh: 30000, ProductionKWh: 10.0, ConsumptionKWh: 30.0, NetKWh: -20.0, ProductionValueUSD: 1.30},
	}

	err := WriteHistoryFile(path, 12345, 0.13, records)
	if err != nil {
		t.Fatalf("WriteHistoryFile failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	var hf HistoryFile
	if err := json.Unmarshal(data, &hf); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if hf.SystemID != 12345 {
		t.Errorf("Expected system ID 12345, got %d", hf.SystemID)
	}
	if hf.RatePerKWh != 0.13 {
		t.Errorf("Expected rate 0.13, got %f", hf.RatePerKWh)
	}
	if hf.TotalRecords != 1 {
		t.Errorf("Expected 1 record, got %d", hf.TotalRecords)
	}
	if len(hf.Records) != 1 {
		t.Fatalf("Expected 1 record in array, got %d", len(hf.Records))
	}
	if hf.Records[0].Date != "2025-03-01" {
		t.Errorf("Expected date 2025-03-01, got %s", hf.Records[0].Date)
	}
}
