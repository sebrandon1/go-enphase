package lib

import (
	"strings"
	"testing"
)

func TestFormatSystems(t *testing.T) {
	systems := []System{
		{SystemID: 123, Name: "Home Solar", Status: "normal", Modules: 20, City: "Rochester", State: "MN"},
		{SystemID: 456, Name: "Office", Status: "micro", Modules: 10, City: "Austin", State: "TX"},
	}

	out := FormatSystems(systems)

	if !strings.Contains(out, "Systems (2)") {
		t.Error("Expected system count in header")
	}
	if !strings.Contains(out, "123") {
		t.Error("Expected system ID 123")
	}
	if !strings.Contains(out, "Home Solar") {
		t.Error("Expected system name")
	}
	if !strings.Contains(out, "Rochester, MN") {
		t.Error("Expected location")
	}
	if !strings.Contains(out, "SYSTEM ID") {
		t.Error("Expected table header")
	}
}

func TestFormatSystemsEmpty(t *testing.T) {
	out := FormatSystems(nil)

	if !strings.Contains(out, "No systems found") {
		t.Error("Expected empty message")
	}
}

func TestFormatSystemSummary(t *testing.T) {
	s := &SystemSummary{
		SystemID:       123,
		Status:         "normal",
		CurrentPower:   3500,
		Modules:        20,
		SizeW:          7600,
		EnergyToday:    15000,
		EnergyLifetime: 5000000,
		LastReportAt:   1700000000,
	}

	out := FormatSystemSummary(s)

	if !strings.Contains(out, "System Summary (123)") {
		t.Error("Expected system ID in header")
	}
	if !strings.Contains(out, "3500 W") {
		t.Error("Expected current power")
	}
	if !strings.Contains(out, "15.00 kWh") {
		t.Error("Expected energy today in kWh")
	}
	if !strings.Contains(out, "5000.0 kWh") {
		t.Error("Expected energy lifetime in kWh")
	}
	if !strings.Contains(out, "7600 W") {
		t.Error("Expected system size")
	}
}

func TestFormatDevices(t *testing.T) {
	devices := []Device{
		{ID: 100, SerialNumber: "ABC123", Model: "IQ7+", Status: "normal", LastReportAt: 1700000000},
		{ID: 101, SerialNumber: "DEF456", Model: "IQ8", Status: "normal", LastReportAt: 0},
	}

	out := FormatDevices(devices)

	if !strings.Contains(out, "Devices (2)") {
		t.Error("Expected device count in header")
	}
	if !strings.Contains(out, "ABC123") {
		t.Error("Expected serial number")
	}
	if !strings.Contains(out, "IQ7+") {
		t.Error("Expected model")
	}
	if !strings.Contains(out, "ID") {
		t.Error("Expected table header")
	}
}

func TestFormatDevicesEmpty(t *testing.T) {
	out := FormatDevices(nil)

	if !strings.Contains(out, "No devices found") {
		t.Error("Expected empty message")
	}
}

func TestFormatMeterReadings(t *testing.T) {
	readings := []MeterReading{
		{SerialNumber: "ABC123", Value: 15000, ReadAt: 1700000000},
	}

	out := FormatMeterReadings(readings)

	if !strings.Contains(out, "Production Meter Readings (1)") {
		t.Error("Expected readings count in header")
	}
	if !strings.Contains(out, "ABC123") {
		t.Error("Expected serial number")
	}
	if !strings.Contains(out, "15000") {
		t.Error("Expected value")
	}
}

func TestFormatMeterReadingsEmpty(t *testing.T) {
	out := FormatMeterReadings(nil)

	if !strings.Contains(out, "No readings found") {
		t.Error("Expected empty message")
	}
}

func TestFormatEnergyLifetime(t *testing.T) {
	e := &EnergyLifetime{
		StartDate:  "2025-01-01",
		SystemID:   123,
		Production: []int{10000, 12000, 14000, 11000, 13000, 15000, 9000, 16000, 14500, 13500},
	}

	out := FormatEnergyLifetime(e)

	if !strings.Contains(out, "Energy Lifetime (System 123)") {
		t.Error("Expected system ID in header")
	}
	if !strings.Contains(out, "Days:        10") {
		t.Error("Expected day count")
	}
	if !strings.Contains(out, "Total:") {
		t.Error("Expected total")
	}
	if !strings.Contains(out, "Daily Avg:") {
		t.Error("Expected daily average")
	}
	if !strings.Contains(out, "Last Days:") {
		t.Error("Expected last days section")
	}
	// Should show last 7 days, not all 10
	if !strings.Contains(out, "2025-01-04") {
		t.Error("Expected day 4 in last 7 days")
	}
}

func TestFormatEnergyLifetimeEmpty(t *testing.T) {
	e := &EnergyLifetime{
		StartDate:  "2025-01-01",
		SystemID:   123,
		Production: []int{},
	}

	out := FormatEnergyLifetime(e)

	if !strings.Contains(out, "Days:        0") {
		t.Error("Expected zero days")
	}
	if strings.Contains(out, "Total:") {
		t.Error("Should not have total for empty data")
	}
}

func TestFormatEnergyLifetimeFewDays(t *testing.T) {
	e := &EnergyLifetime{
		StartDate:  "2025-01-01",
		SystemID:   123,
		Production: []int{10000, 12000, 14000},
	}

	out := FormatEnergyLifetime(e)

	if !strings.Contains(out, "Days:        3") {
		t.Error("Expected 3 days")
	}
	// All 3 days should appear since fewer than 7
	if !strings.Contains(out, "2025-01-01") {
		t.Error("Expected first day")
	}
	if !strings.Contains(out, "2025-01-03") {
		t.Error("Expected last day")
	}
}

func TestFormatConsumptionLifetime(t *testing.T) {
	c := &ConsumptionLifetime{
		StartDate:   "2025-01-01",
		SystemID:    123,
		Consumption: []int{20000, 22000, 18000},
	}

	out := FormatConsumptionLifetime(c)

	if !strings.Contains(out, "Consumption Lifetime (System 123)") {
		t.Error("Expected system ID in header")
	}
	if !strings.Contains(out, "Days:        3") {
		t.Error("Expected day count")
	}
	if !strings.Contains(out, "Total:") {
		t.Error("Expected total")
	}
}

func TestFormatConsumptionLifetimeEmpty(t *testing.T) {
	c := &ConsumptionLifetime{
		StartDate:   "2025-01-01",
		SystemID:    123,
		Consumption: []int{},
	}

	out := FormatConsumptionLifetime(c)

	if !strings.Contains(out, "Days:        0") {
		t.Error("Expected zero days")
	}
}

func TestFormatBatteryStatus(t *testing.T) {
	b := &BatteryStatus{
		Status:                 "normal",
		BatteryCount:           2,
		ChargePercent:          85.5,
		EnergyStoredLifetime:   1500000,
		EnergyConsumedLifetime: 1200000,
	}

	out := FormatBatteryStatus(b)

	if !strings.Contains(out, "Battery Status") {
		t.Error("Expected header")
	}
	if !strings.Contains(out, "normal") {
		t.Error("Expected status")
	}
	if !strings.Contains(out, "85.5%") {
		t.Error("Expected charge percent")
	}
	if !strings.Contains(out, "1500.0 kWh") {
		t.Error("Expected stored lifetime")
	}
	if !strings.Contains(out, "1200.0 kWh") {
		t.Error("Expected consumed lifetime")
	}
}

func TestFormatAuthStatus(t *testing.T) {
	out := FormatAuthStatus(true, true, false, false, true)

	if !strings.Contains(out, "Auth Status") {
		t.Error("Expected header")
	}
	if !strings.Contains(out, "API Key:        set") {
		t.Error("Expected API key set")
	}
	if !strings.Contains(out, "Access Token:   set") {
		t.Error("Expected access token set")
	}
	if !strings.Contains(out, "Refresh Token:  not set") {
		t.Error("Expected refresh token not set")
	}
	if !strings.Contains(out, "Client ID:      not set") {
		t.Error("Expected client ID not set")
	}
	if !strings.Contains(out, "Client Secret:  set") {
		t.Error("Expected client secret set")
	}
}

func TestFormatAuthStatusAllUnset(t *testing.T) {
	out := FormatAuthStatus(false, false, false, false, false)

	for _, label := range []string{"API Key", "Access Token", "Refresh Token", "Client ID", "Client Secret"} {
		if !strings.Contains(out, label) {
			t.Errorf("Expected label %q", label)
		}
	}
	if strings.Count(out, "not set") != 5 {
		t.Error("Expected all 5 credentials to be 'not set'")
	}
}

func TestFormatTokenInfo(t *testing.T) {
	token := &TokenInfo{
		TokenType:    "Bearer",
		ExpiresIn:    86400,
		AccessToken:  "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.abcdefg",
		RefreshToken: "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4.hijklmn",
	}

	out := FormatTokenInfo(token)

	if !strings.Contains(out, "Token Refreshed") {
		t.Error("Expected header")
	}
	if !strings.Contains(out, "Bearer") {
		t.Error("Expected token type")
	}
	if !strings.Contains(out, "86400 seconds") {
		t.Error("Expected expiry")
	}
	// Tokens should be truncated
	if !strings.Contains(out, "...") {
		t.Error("Expected truncated tokens")
	}
	// Full token should NOT appear
	if strings.Contains(out, "abcdefg") {
		t.Error("Token should be truncated, not shown in full")
	}
}

func TestFormatTokenInfoShortToken(t *testing.T) {
	token := &TokenInfo{
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		AccessToken:  "short",
		RefreshToken: "also-short",
	}

	out := FormatTokenInfo(token)

	if !strings.Contains(out, "short") {
		t.Error("Short token should appear in full")
	}
	if strings.Contains(out, "...") {
		t.Error("Short tokens should not be truncated")
	}
}

func TestFormatEnvoyProduction(t *testing.T) {
	p := &EnvoyProduction{
		Production: []EnvoyProductionEntry{
			{Type: "inverters", ActiveCount: 20, WNow: 3500, WhToday: 15000, WhLastSevenDays: 105000, WhLifetime: 5000000},
		},
		Consumption: []EnvoyConsumptionEntry{
			{Type: "total-consumption", ActiveCount: 1, WNow: 2100, WhToday: 45000, WhLifetime: 120000000},
		},
	}

	out := FormatEnvoyProduction(p)

	if !strings.Contains(out, "Envoy Status") {
		t.Error("Expected header")
	}
	if !strings.Contains(out, "Production:") {
		t.Error("Expected production section")
	}
	if !strings.Contains(out, "inverters") {
		t.Error("Expected inverters type")
	}
	if !strings.Contains(out, "3500 W") {
		t.Error("Expected power now")
	}
	if !strings.Contains(out, "Consumption:") {
		t.Error("Expected consumption section")
	}
	if !strings.Contains(out, "total-consumption") {
		t.Error("Expected consumption type")
	}
}

func TestFormatEnvoyProductionEmpty(t *testing.T) {
	p := &EnvoyProduction{}

	out := FormatEnvoyProduction(p)

	if !strings.Contains(out, "Envoy Status") {
		t.Error("Expected header")
	}
	if strings.Contains(out, "Production:") {
		t.Error("Should not have production section when empty")
	}
}

func TestFormatSensorReadings(t *testing.T) {
	readings := []SensorReading{
		{MeasurementType: "production", ActivePower: 3500, RmsVoltage: 240.5, RmsCurrent: 14.6, PowerFactor: 0.99, Frequency: 60.0},
		{MeasurementType: "net-consumption", ActivePower: -1400, RmsVoltage: 240.3, RmsCurrent: 5.8, PowerFactor: 0.98, Frequency: 60.0},
	}

	out := FormatSensorReadings(readings)

	if !strings.Contains(out, "Sensor Readings (2)") {
		t.Error("Expected count in header")
	}
	if !strings.Contains(out, "production") {
		t.Error("Expected production type")
	}
	if !strings.Contains(out, "net-consumption") {
		t.Error("Expected net-consumption type")
	}
	if !strings.Contains(out, "240.5") {
		t.Error("Expected voltage")
	}
}

func TestFormatSensorReadingsEmpty(t *testing.T) {
	out := FormatSensorReadings(nil)

	if !strings.Contains(out, "No readings found") {
		t.Error("Expected empty message")
	}
}
