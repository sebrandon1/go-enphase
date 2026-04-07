package lib

import (
	"fmt"
	"strings"
	"time"
)

// FormatSystems formats a list of systems as a human-readable table.
func FormatSystems(systems []System) string {
	var b strings.Builder

	fmt.Fprintf(&b, "=== Systems (%d) ===\n\n", len(systems))

	if len(systems) == 0 {
		b.WriteString("  No systems found.\n")
		return b.String()
	}

	fmt.Fprintf(&b, "  %-12s %-20s %-10s %8s  %s\n", "SYSTEM ID", "NAME", "STATUS", "MODULES", "LOCATION")
	for _, s := range systems {
		location := s.City
		if s.State != "" {
			if location != "" {
				location += ", "
			}
			location += s.State
		}
		fmt.Fprintf(&b, "  %-12d %-20s %-10s %8d  %s\n", s.SystemID, s.Name, s.Status, s.Modules, location)
	}

	return b.String()
}

// FormatSystemSummary formats a system summary as human-readable key-value pairs.
func FormatSystemSummary(s *SystemSummary) string {
	var b strings.Builder

	fmt.Fprintf(&b, "=== System Summary (%d) ===\n\n", s.SystemID)
	fmt.Fprintf(&b, "  Status:          %s\n", s.Status)
	fmt.Fprintf(&b, "  Current Power:   %d W\n", s.CurrentPower)
	fmt.Fprintf(&b, "  Modules:         %d\n", s.Modules)
	fmt.Fprintf(&b, "  Size:            %d W\n", s.SizeW)
	fmt.Fprintf(&b, "  Energy Today:    %.2f kWh\n", float64(s.EnergyToday)/1000.0)
	fmt.Fprintf(&b, "  Energy Lifetime: %.1f kWh\n", float64(s.EnergyLifetime)/1000.0)

	if s.LastReportAt > 0 {
		t := time.Unix(s.LastReportAt, 0)
		fmt.Fprintf(&b, "  Last Report:     %s\n", t.Format("2006-01-02 15:04:05"))
	}

	return b.String()
}

// FormatDevices formats a list of devices as a human-readable table.
func FormatDevices(devices []Device) string {
	var b strings.Builder

	fmt.Fprintf(&b, "=== Devices (%d) ===\n\n", len(devices))

	if len(devices) == 0 {
		b.WriteString("  No devices found.\n")
		return b.String()
	}

	fmt.Fprintf(&b, "  %-10s %-14s %-12s %-10s %s\n", "ID", "SERIAL", "MODEL", "STATUS", "LAST REPORT")
	for _, d := range devices {
		lastReport := ""
		if d.LastReportAt > 0 {
			lastReport = time.Unix(d.LastReportAt, 0).Format("2006-01-02 15:04:05")
		}
		fmt.Fprintf(&b, "  %-10d %-14s %-12s %-10s %s\n", d.ID, d.SerialNumber, d.Model, d.Status, lastReport)
	}

	return b.String()
}

// FormatMeterReadings formats production meter readings as a human-readable table.
func FormatMeterReadings(readings []MeterReading) string {
	var b strings.Builder

	fmt.Fprintf(&b, "=== Production Meter Readings (%d) ===\n\n", len(readings))

	if len(readings) == 0 {
		b.WriteString("  No readings found.\n")
		return b.String()
	}

	fmt.Fprintf(&b, "  %-14s %12s  %s\n", "SERIAL", "VALUE (Wh)", "READ AT")
	for _, r := range readings {
		readAt := ""
		if r.ReadAt > 0 {
			readAt = time.Unix(r.ReadAt, 0).Format("2006-01-02 15:04:05")
		}
		fmt.Fprintf(&b, "  %-14s %12d  %s\n", r.SerialNumber, r.Value, readAt)
	}

	return b.String()
}

// FormatEnergyLifetime formats energy lifetime data as a summary with recent days.
func FormatEnergyLifetime(e *EnergyLifetime) string {
	var b strings.Builder

	fmt.Fprintf(&b, "=== Energy Lifetime (System %d) ===\n\n", e.SystemID)
	fmt.Fprintf(&b, "  Start Date:  %s\n", e.StartDate)
	fmt.Fprintf(&b, "  Days:        %d\n", len(e.Production))

	if len(e.Production) == 0 {
		return b.String()
	}

	var total int
	for _, v := range e.Production {
		total += v
	}
	totalKWh := float64(total) / 1000.0
	avgKWh := totalKWh / float64(len(e.Production))

	fmt.Fprintf(&b, "  Total:       %.1f kWh\n", totalKWh)
	fmt.Fprintf(&b, "  Daily Avg:   %.1f kWh\n", avgKWh)

	// Show last 7 days (or fewer if less data available).
	showDays := 7
	if len(e.Production) < showDays {
		showDays = len(e.Production)
	}

	b.WriteString("\n  Last Days:\n")
	fmt.Fprintf(&b, "  %-12s %10s\n", "DATE", "PRODUCTION")

	startTime, err := time.Parse("2006-01-02", e.StartDate)
	if err != nil {
		return b.String()
	}

	offset := len(e.Production) - showDays
	for i := offset; i < len(e.Production); i++ {
		date := startTime.AddDate(0, 0, i).Format("2006-01-02")
		fmt.Fprintf(&b, "  %-12s %8.1f kWh\n", date, float64(e.Production[i])/1000.0)
	}

	return b.String()
}

// FormatConsumptionLifetime formats consumption lifetime data as a summary with recent days.
func FormatConsumptionLifetime(c *ConsumptionLifetime) string {
	var b strings.Builder

	fmt.Fprintf(&b, "=== Consumption Lifetime (System %d) ===\n\n", c.SystemID)
	fmt.Fprintf(&b, "  Start Date:  %s\n", c.StartDate)
	fmt.Fprintf(&b, "  Days:        %d\n", len(c.Consumption))

	if len(c.Consumption) == 0 {
		return b.String()
	}

	var total int
	for _, v := range c.Consumption {
		total += v
	}
	totalKWh := float64(total) / 1000.0
	avgKWh := totalKWh / float64(len(c.Consumption))

	fmt.Fprintf(&b, "  Total:       %.1f kWh\n", totalKWh)
	fmt.Fprintf(&b, "  Daily Avg:   %.1f kWh\n", avgKWh)

	showDays := 7
	if len(c.Consumption) < showDays {
		showDays = len(c.Consumption)
	}

	b.WriteString("\n  Last Days:\n")
	fmt.Fprintf(&b, "  %-12s %12s\n", "DATE", "CONSUMPTION")

	startTime, err := time.Parse("2006-01-02", c.StartDate)
	if err != nil {
		return b.String()
	}

	offset := len(c.Consumption) - showDays
	for i := offset; i < len(c.Consumption); i++ {
		date := startTime.AddDate(0, 0, i).Format("2006-01-02")
		fmt.Fprintf(&b, "  %-12s %10.1f kWh\n", date, float64(c.Consumption[i])/1000.0)
	}

	return b.String()
}

// FormatBatteryStatus formats battery status as human-readable key-value pairs.
func FormatBatteryStatus(b2 *BatteryStatus) string {
	var b strings.Builder

	b.WriteString("=== Battery Status ===\n\n")
	fmt.Fprintf(&b, "  Status:          %s\n", b2.Status)
	fmt.Fprintf(&b, "  Battery Count:   %d\n", b2.BatteryCount)
	fmt.Fprintf(&b, "  Charge:          %.1f%%\n", b2.ChargePercent)
	fmt.Fprintf(&b, "  Stored Lifetime: %.1f kWh\n", float64(b2.EnergyStoredLifetime)/1000.0)
	fmt.Fprintf(&b, "  Used Lifetime:   %.1f kWh\n", float64(b2.EnergyConsumedLifetime)/1000.0)

	return b.String()
}
