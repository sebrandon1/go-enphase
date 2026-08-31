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

func TestFormatSystem(t *testing.T) {
	s := &System{SystemID: 123, Name: "Home Solar", Status: "normal", Modules: 20, City: "Rochester", State: "MN", TimeZone: "America/Chicago"}

	out := FormatSystem(s)

	if !strings.Contains(out, "System 123") {
		t.Error("Expected system ID in header")
	}
	if !strings.Contains(out, "Home Solar") {
		t.Error("Expected name")
	}
	if !strings.Contains(out, "Rochester, MN") {
		t.Error("Expected location")
	}
	if !strings.Contains(out, "America/Chicago") {
		t.Error("Expected time zone")
	}
	if !strings.Contains(out, "20") {
		t.Error("Expected module count")
	}
}

func TestFormatSystemNoLocation(t *testing.T) {
	s := &System{SystemID: 99, Name: "Unnamed", Status: "normal", Modules: 5}

	out := FormatSystem(s)

	if !strings.Contains(out, "System 99") {
		t.Error("Expected system ID in header")
	}
	if strings.Contains(out, "Location:") {
		t.Error("Should not show Location when empty")
	}
}

func TestFormatInverterReadings(t *testing.T) {
	readings := []InverterReading{
		{SerialNumber: "SN001", LastReportDate: 1700000000, LastReportWatts: 250, MaxReportWatts: 300},
		{SerialNumber: "SN002", LastReportDate: 0, LastReportWatts: 0, MaxReportWatts: 0},
	}

	out := FormatInverterReadings(readings)

	if !strings.Contains(out, "Inverter Readings (2)") {
		t.Error("Expected count in header")
	}
	if !strings.Contains(out, "SN001") {
		t.Error("Expected serial SN001")
	}
	if !strings.Contains(out, "250") {
		t.Error("Expected watts")
	}
	if !strings.Contains(out, "SERIAL") {
		t.Error("Expected table header")
	}
}

func TestFormatInverterReadingsEmpty(t *testing.T) {
	out := FormatInverterReadings(nil)

	if !strings.Contains(out, "No readings found") {
		t.Error("Expected empty message")
	}
}

func TestFormatMeterConfig(t *testing.T) {
	configs := []MeterConfig{
		{EID: 704643328, State: "enabled", MeasurementType: "production", PhaseMode: "three", PhaseCount: 3},
	}

	out := FormatMeterConfig(configs)

	if !strings.Contains(out, "Meter Configuration (1)") {
		t.Error("Expected count in header")
	}
	if !strings.Contains(out, "enabled") {
		t.Error("Expected state")
	}
	if !strings.Contains(out, "production") {
		t.Error("Expected measurement type")
	}
	if !strings.Contains(out, "three") {
		t.Error("Expected phase mode")
	}
}

func TestFormatMeterConfigEmpty(t *testing.T) {
	out := FormatMeterConfig(nil)

	if !strings.Contains(out, "No meters found") {
		t.Error("Expected empty message")
	}
}

func TestFormatMeterData(t *testing.T) {
	data := []MeterData{
		{EID: 704643328, Timestamp: 1700000000, ActPower: 4000.5, ApprntPwr: 4100.0, RmsVoltage: 240.0, RmsCurrent: 16.5},
	}

	out := FormatMeterData(data)

	if !strings.Contains(out, "Meter Readings (1)") {
		t.Error("Expected count in header")
	}
	if !strings.Contains(out, "4000.5") {
		t.Error("Expected active power")
	}
	if !strings.Contains(out, "240.0") {
		t.Error("Expected voltage")
	}
	if !strings.Contains(out, "EID") {
		t.Error("Expected table header")
	}
}

func TestFormatMeterDataEmpty(t *testing.T) {
	out := FormatMeterData(nil)

	if !strings.Contains(out, "No readings found") {
		t.Error("Expected empty message")
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

func TestFormatEnvoyProduction(t *testing.T) {
	p := &EnvoyProduction{
		Production: []EnvoyProductionEntry{
			{Type: "inverters", ActiveCount: 12, WNow: 1500.5, WhToday: 8200.0, WhLifetime: 123456.0},
		},
		Consumption: []EnvoyConsumptionEntry{
			{Type: "eim", MeasurementType: "total-consumption", ActiveCount: 1, WNow: 900.0, WhToday: 5000.0},
		},
	}
	out := FormatEnvoyProduction(p)

	if !strings.Contains(out, "=== Envoy Production") {
		t.Error("Expected production header")
	}
	if !strings.Contains(out, "inverters") {
		t.Error("Expected inverters type")
	}
	if !strings.Contains(out, "=== Envoy Consumption") {
		t.Error("Expected consumption header")
	}
	if !strings.Contains(out, "total-consumption") {
		t.Error("Expected consumption measurement type")
	}
}

func TestFormatEnvoyProductionNoConsumption(t *testing.T) {
	p := &EnvoyProduction{
		Production: []EnvoyProductionEntry{
			{Type: "inverters", ActiveCount: 8, WNow: 600.0, WhToday: 3000.0, WhLifetime: 50000.0},
		},
	}
	out := FormatEnvoyProduction(p)

	if !strings.Contains(out, "=== Envoy Production") {
		t.Error("Expected production header")
	}
	if strings.Contains(out, "=== Envoy Consumption") {
		t.Error("Should not have consumption section when consumption is empty")
	}
}

func TestFormatEnvoySensors(t *testing.T) {
	readings := []SensorReading{
		{MeasurementType: "total-consumption", ActivePower: 850.0, RmsVoltage: 240.1, RmsCurrent: 3.54, Frequency: 60.0, PowerFactor: 0.995},
	}
	out := FormatEnvoySensors(readings)

	if !strings.Contains(out, "=== Envoy Sensor Readings") {
		t.Error("Expected sensor readings header")
	}
	if !strings.Contains(out, "total-consumption") {
		t.Error("Expected measurement type in output")
	}
	if !strings.Contains(out, "850.0") {
		t.Error("Expected active power in output")
	}
}

func TestFormatEnvoySensorsEmpty(t *testing.T) {
	out := FormatEnvoySensors(nil)

	if !strings.Contains(out, "=== Envoy Sensor Readings") {
		t.Error("Expected sensor readings header even when empty")
	}
	if !strings.Contains(out, "No sensor readings found") {
		t.Error("Expected empty message for nil readings")
	}
}

func TestFormatEnvoySensorsEmptySlice(t *testing.T) {
	out := FormatEnvoySensors([]SensorReading{})

	if !strings.Contains(out, "No sensor readings found") {
		t.Error("Expected empty message for empty slice")
	}
}
