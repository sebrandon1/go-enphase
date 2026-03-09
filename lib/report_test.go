package lib

import (
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

	output := FormatTodaySummary(summary, 0.12)

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

func TestFormatTodaySummaryNoRate(t *testing.T) {
	summary := &SystemSummary{
		SystemID:    12345,
		EnergyToday: 15000,
	}

	output := FormatTodaySummary(summary, 0)
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
