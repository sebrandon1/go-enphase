package lib

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// FormatTodaySummary formats a SystemSummary as human-readable text,
// matching the output of the old enphase-today.sh script.
func FormatTodaySummary(s *SystemSummary, ratePerKWh float64) string {
	var b strings.Builder

	fmt.Fprintf(&b, "=== Solar Today (System %d) ===\n", s.SystemID)
	b.WriteString("\n")
	fmt.Fprintf(&b, "  Status:          %s\n", s.Status)
	fmt.Fprintf(&b, "  Current Power:   %d W\n", s.CurrentPower)
	fmt.Fprintf(&b, "  Modules:         %d\n", s.Modules)

	todayKWh := float64(s.EnergyToday) / 1000.0
	fmt.Fprintf(&b, "  Energy Today:    %.2f kWh", todayKWh)
	if ratePerKWh > 0 {
		fmt.Fprintf(&b, " ($%.2f)", todayKWh*ratePerKWh)
	}
	b.WriteString("\n")

	lifetimeKWh := float64(s.EnergyLifetime) / 1000.0
	fmt.Fprintf(&b, "  Energy Lifetime: %.1f kWh", lifetimeKWh)
	if ratePerKWh > 0 {
		fmt.Fprintf(&b, " ($%.2f)", lifetimeKWh*ratePerKWh)
	}
	b.WriteString("\n")

	if s.LastReportAt > 0 {
		t := time.Unix(s.LastReportAt, 0)
		fmt.Fprintf(&b, "  Last Report:     %s\n", t.Format("2006-01-02 15:04:05"))
	}

	return b.String()
}

// MonthStats holds computed statistics for a month of energy production.
type MonthStats struct {
	TotalKWh  float64
	DailyAvg  float64
	BestKWh   float64
	WorstKWh  float64
	Above15   int
	Below5    int
	DaysCount int
}

// ComputeMonthStats computes statistics from daily production values (in Wh).
func ComputeMonthStats(production []int) MonthStats {
	n := len(production)
	if n == 0 {
		return MonthStats{}
	}

	total := 0
	best := math.MinInt64
	worst := math.MaxInt64
	above15 := 0
	below5 := 0

	for _, v := range production {
		total += v
		if v > best {
			best = v
		}
		if v < worst {
			worst = v
		}
		if v >= 15000 {
			above15++
		}
		if v < 5000 {
			below5++
		}
	}

	totalKWh := float64(total) / 1000.0
	return MonthStats{
		TotalKWh:  totalKWh,
		DailyAvg:  totalKWh / float64(n),
		BestKWh:   float64(best) / 1000.0,
		WorstKWh:  float64(worst) / 1000.0,
		Above15:   above15,
		Below5:    below5,
		DaysCount: n,
	}
}

// FormatMonthComparison formats two months of energy data as a comparison table,
// matching the output of the old enphase-compare.sh script.
func FormatMonthComparison(systemID, monthA, monthB string, prodA, prodB *EnergyLifetime, ratePerKWh float64) string {
	var b strings.Builder

	fmt.Fprintf(&b, "=== Month Comparison (System %s) ===\n", systemID)
	fmt.Fprintf(&b, "  %s vs %s\n\n", monthA, monthB)

	statsA := ComputeMonthStats(prodA.Production)
	statsB := ComputeMonthStats(prodB.Production)

	fmt.Fprintf(&b, "  %-20s  %12s  %12s\n", "Metric", monthA, monthB)
	fmt.Fprintf(&b, "  %-20s  %12s  %12s\n", "--------------------", "------------", "------------")
	fmt.Fprintf(&b, "  %-20s  %10.1f kWh  %10.1f kWh\n", "Total Production", statsA.TotalKWh, statsB.TotalKWh)
	fmt.Fprintf(&b, "  %-20s  %10.1f kWh  %10.1f kWh\n", "Daily Average", statsA.DailyAvg, statsB.DailyAvg)
	fmt.Fprintf(&b, "  %-20s  %10.1f kWh  %10.1f kWh\n", "Best Day", statsA.BestKWh, statsB.BestKWh)
	fmt.Fprintf(&b, "  %-20s  %10.1f kWh  %10.1f kWh\n", "Worst Day", statsA.WorstKWh, statsB.WorstKWh)
	fmt.Fprintf(&b, "  %-20s  %12d  %12d\n", "Days > 15 kWh", statsA.Above15, statsB.Above15)
	fmt.Fprintf(&b, "  %-20s  %12d  %12d\n", "Days < 5 kWh", statsA.Below5, statsB.Below5)
	fmt.Fprintf(&b, "  %-20s  %12d  %12d\n", "Days Reported", statsA.DaysCount, statsB.DaysCount)

	if ratePerKWh > 0 {
		fmt.Fprintf(&b, "  %-20s  %11s  %11s\n", "Est. Value",
			fmt.Sprintf("$%.2f", statsA.TotalKWh*ratePerKWh),
			fmt.Sprintf("$%.2f", statsB.TotalKWh*ratePerKWh))
	}

	return b.String()
}

// MonthDateRange returns the start and end dates for a YYYY-MM month string.
// The end date is clamped to yesterday if the month extends into the future.
func MonthDateRange(yearMonth string) (start, end string, err error) {
	t, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return "", "", fmt.Errorf("invalid month format %q, expected YYYY-MM", yearMonth)
	}

	start = t.Format("2006-01-02")

	lastDay := t.AddDate(0, 1, -1)
	yesterday := time.Now().AddDate(0, 0, -1)
	if lastDay.After(yesterday) {
		lastDay = yesterday
	}
	end = lastDay.Format("2006-01-02")

	return start, end, nil
}

// LastTwoMonths returns the YYYY-MM strings for the two most recent complete months.
func LastTwoMonths() (twoMonthsAgo, lastMonth string) {
	now := time.Now()
	last := now.AddDate(0, -1, 0)
	twoAgo := now.AddDate(0, -2, 0)
	return twoAgo.Format("2006-01"), last.Format("2006-01")
}
