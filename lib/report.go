package lib

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

// FormatTodaySummary formats a SystemSummary as human-readable text.
// todayConsWh is today's consumption in Wh (0 if unavailable).
func FormatTodaySummary(s *SystemSummary, todayConsWh int, ratePerKWh float64) string {
	var b strings.Builder

	fmt.Fprintf(&b, "=== Solar Today (System %d) ===\n", s.SystemID)
	b.WriteString("\n")
	fmt.Fprintf(&b, "  Status:          %s\n", s.Status)
	fmt.Fprintf(&b, "  Current Power:   %d W\n", s.CurrentPower)
	fmt.Fprintf(&b, "  Modules:         %d\n", s.Modules)

	todayKWh := float64(s.EnergyToday) / 1000.0
	fmt.Fprintf(&b, "  Production:      %.2f kWh", todayKWh)
	if ratePerKWh > 0 {
		fmt.Fprintf(&b, " ($%.2f)", todayKWh*ratePerKWh)
	}
	b.WriteString("\n")

	if todayConsWh > 0 {
		consKWh := float64(todayConsWh) / 1000.0
		fmt.Fprintf(&b, "  Consumption:     %.2f kWh", consKWh)
		if ratePerKWh > 0 {
			fmt.Fprintf(&b, " ($%.2f)", consKWh*ratePerKWh)
		}
		b.WriteString("\n")

		netKWh := todayKWh - consKWh
		gridLabel := "Grid Draw"
		if netKWh >= 0 {
			gridLabel = "Grid Export"
		}
		absNet := netKWh
		if absNet < 0 {
			absNet = -absNet
		}
		fmt.Fprintf(&b, "  %s:      %.2f kWh", gridLabel, absNet)
		if ratePerKWh > 0 {
			fmt.Fprintf(&b, " ($%.2f)", absNet*ratePerKWh)
		}
		b.WriteString("\n")

		offsetPct := (todayKWh / consKWh) * 100.0
		fmt.Fprintf(&b, "  Solar Offset:    %.0f%%\n", offsetPct)
	}

	lifetimeKWh := float64(s.EnergyLifetime) / 1000.0
	fmt.Fprintf(&b, "  Energy Lifetime: %.1f kWh", lifetimeKWh)
	if ratePerKWh > 0 {
		fmt.Fprintf(&b, " ($%.2f)", lifetimeKWh*ratePerKWh)
	}
	b.WriteString("\n")

	if s.LastReportAt > 0 {
		t := time.Unix(s.LastReportAt, 0)
		fmt.Fprintf(&b, "  Last Report:     %s\n", t.Format(tsLayout))
	}

	return b.String()
}

// MonthStats holds computed statistics for a month of energy production.
type MonthStats struct {
	TotalKWh  float64 `json:"total_kwh"`
	DailyAvg  float64 `json:"daily_avg_kwh"`
	BestKWh   float64 `json:"best_day_kwh"`
	WorstKWh  float64 `json:"worst_day_kwh"`
	Above15   int     `json:"days_above_15kwh"`
	Below5    int     `json:"days_below_5kwh"`
	DaysCount int     `json:"days_reported"`
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

// DailyRow holds one day of combined production and consumption data.
type DailyRow struct {
	Date           string  `json:"date"`
	ProductionWh   int     `json:"production_wh"`
	ConsumptionWh  int     `json:"consumption_wh"`
	ProductionKWh  float64 `json:"production_kwh"`
	ConsumptionKWh float64 `json:"consumption_kwh"`
	NetKWh         float64 `json:"net_kwh"`
}

// dailyDateRange computes the overall start/end dates spanned by production and
// optional consumption lifetime data.
func dailyDateRange(prod *EnergyLifetime, cons *ConsumptionLifetime) (start, end time.Time, prodStart, consStart time.Time, ok bool) {
	prodStart, err := time.Parse("2006-01-02", prod.StartDate)
	if err != nil {
		return time.Time{}, time.Time{}, time.Time{}, time.Time{}, false
	}

	consStart = prodStart
	if cons != nil {
		if t, err := time.Parse("2006-01-02", cons.StartDate); err == nil {
			consStart = t
		}
	}

	start = prodStart
	end = prodStart.AddDate(0, 0, len(prod.Production)-1)

	if cons != nil {
		if consStart.Before(start) {
			start = consStart
		}
		if len(cons.Consumption) > 0 {
			consEnd := consStart.AddDate(0, 0, len(cons.Consumption)-1)
			if consEnd.After(end) {
				end = consEnd
			}
		}
	}
	return start, end, prodStart, consStart, true
}

// lookupWh returns the Wh value from a daily array given its start date and a target date.
func lookupWh(values []int, seriesStart, target time.Time) int {
	idx := int(target.Sub(seriesStart).Hours() / 24)
	if idx >= 0 && idx < len(values) {
		return values[idx]
	}
	return 0
}

// BuildDailyRows combines production and consumption lifetime data into
// day-by-day rows for the trailing N days. It aligns the two arrays by date
// and returns the most recent `days` entries (or all available if fewer).
func BuildDailyRows(prod *EnergyLifetime, cons *ConsumptionLifetime, days int) []DailyRow {
	start, end, prodStart, consStart, ok := dailyDateRange(prod, cons)
	if !ok {
		return nil
	}

	totalDays := int(end.Sub(start).Hours()/24) + 1
	if totalDays <= 0 {
		return nil
	}

	rows := make([]DailyRow, 0, totalDays)
	for i := 0; i < totalDays; i++ {
		d := start.AddDate(0, 0, i)
		prodWh := lookupWh(prod.Production, prodStart, d)

		var consWh int
		if cons != nil {
			consWh = lookupWh(cons.Consumption, consStart, d)
		}

		prodKWh := float64(prodWh) / 1000.0
		consKWh := float64(consWh) / 1000.0
		rows = append(rows, DailyRow{
			Date:           d.Format("2006-01-02"),
			ProductionWh:   prodWh,
			ConsumptionWh:  consWh,
			ProductionKWh:  prodKWh,
			ConsumptionKWh: consKWh,
			NetKWh:         prodKWh - consKWh,
		})
	}

	if days > 0 && len(rows) > days {
		rows = rows[len(rows)-days:]
	}
	return rows
}

// FormatDailyReport formats a combined daily report with today's live summary
// and a trailing history table of production vs consumption.
func FormatDailyReport(summary *SystemSummary, rows []DailyRow, ratePerKWh float64) string {
	var b strings.Builder

	fmt.Fprintf(&b, "=== Solar Daily Report (System %d) ===\n\n", summary.SystemID)
	fmt.Fprintf(&b, "  Status: %s | Current: %d W | Modules: %d\n\n", summary.Status, summary.CurrentPower, summary.Modules)

	todayKWh := float64(summary.EnergyToday) / 1000.0
	fmt.Fprintf(&b, "  Today (in progress): %.2f kWh produced", todayKWh)
	if ratePerKWh > 0 {
		fmt.Fprintf(&b, " ($%.2f)", todayKWh*ratePerKWh)
	}
	b.WriteString("\n")

	if summary.LastReportAt > 0 {
		t := time.Unix(summary.LastReportAt, 0)
		fmt.Fprintf(&b, "  Last Report:         %s\n", t.Format(tsLayout))
	}

	if len(rows) > 0 {
		b.WriteString("\n  Recent History:\n")
		if ratePerKWh > 0 {
			fmt.Fprintf(&b, "  %-12s  %12s  %8s  %12s  %8s  %10s\n",
				"Date", "Production", "Prod $", "Consumption", "Cons $", "Net")
		} else {
			fmt.Fprintf(&b, "  %-12s  %12s  %12s  %10s\n",
				"Date", "Production", "Consumption", "Net")
		}

		var totalProd, totalCons float64
		for _, r := range rows {
			totalProd += r.ProductionKWh
			totalCons += r.ConsumptionKWh
			if ratePerKWh > 0 {
				fmt.Fprintf(&b, "  %-12s  %9.2f kWh  %7s  %9.2f kWh  %7s  %7.2f kWh\n",
					r.Date, r.ProductionKWh, fmt.Sprintf("$%.2f", r.ProductionKWh*ratePerKWh),
					r.ConsumptionKWh, fmt.Sprintf("$%.2f", r.ConsumptionKWh*ratePerKWh),
					r.NetKWh)
			} else {
				fmt.Fprintf(&b, "  %-12s  %9.2f kWh  %9.2f kWh  %7.2f kWh\n",
					r.Date, r.ProductionKWh, r.ConsumptionKWh, r.NetKWh)
			}
		}

		n := float64(len(rows))
		avgProd := totalProd / n
		avgCons := totalCons / n
		avgNet := avgProd - avgCons
		b.WriteString("  ---\n")
		if ratePerKWh > 0 {
			fmt.Fprintf(&b, "  %-12s  %9.2f kWh  %7s  %9.2f kWh  %7s  %7.2f kWh\n",
				fmt.Sprintf("%d-Day Avg", len(rows)),
				avgProd, fmt.Sprintf("$%.2f", avgProd*ratePerKWh),
				avgCons, fmt.Sprintf("$%.2f", avgCons*ratePerKWh),
				avgNet)
		} else {
			fmt.Fprintf(&b, "  %-12s  %9.2f kWh  %9.2f kWh  %7.2f kWh\n",
				fmt.Sprintf("%d-Day Avg", len(rows)),
				avgProd, avgCons, avgNet)
		}
	}

	return b.String()
}

// HistoryRecord represents one day in the history JSON file.
type HistoryRecord struct {
	Date               string  `json:"date"`
	ProductionWh       int     `json:"production_wh"`
	ConsumptionWh      int     `json:"consumption_wh"`
	ProductionKWh      float64 `json:"production_kwh"`
	ConsumptionKWh     float64 `json:"consumption_kwh"`
	NetKWh             float64 `json:"net_kwh"`
	ProductionValueUSD float64 `json:"production_value_usd"`
}

// HistoryFile represents the top-level structure of the history JSON file.
type HistoryFile struct {
	UpdatedAt    string          `json:"updated_at"`
	RatePerKWh   float64         `json:"rate_per_kwh"`
	SystemID     int             `json:"system_id"`
	TotalRecords int             `json:"total_records"`
	Records      []HistoryRecord `json:"records"`
}

// BuildHistoryRecords converts daily rows into history records with dollar values.
func BuildHistoryRecords(rows []DailyRow, ratePerKWh float64) []HistoryRecord {
	records := make([]HistoryRecord, len(rows))
	for i, r := range rows {
		records[i] = HistoryRecord{
			Date:               r.Date,
			ProductionWh:       r.ProductionWh,
			ConsumptionWh:      r.ConsumptionWh,
			ProductionKWh:      math.Round(r.ProductionKWh*1000) / 1000,
			ConsumptionKWh:     math.Round(r.ConsumptionKWh*1000) / 1000,
			NetKWh:             math.Round(r.NetKWh*1000) / 1000,
			ProductionValueUSD: math.Round(r.ProductionKWh*ratePerKWh*100) / 100,
		}
	}
	return records
}

// FormatSummaryReport formats a day-over-day comparison of today vs yesterday
// plus current month running stats. This is the best command for answering
// "how does today compare to yesterday?" or "how is this month going?"
func FormatSummaryReport(summary *SystemSummary, yesterdayProdWh, yesterdayConsWh, todayConsWh int, monthProd []int, monthLabel string, ratePerKWh float64) string {
	var b strings.Builder

	fmt.Fprintf(&b, "=== Solar Summary (System %d) ===\n\n", summary.SystemID)

	todayKWh := float64(summary.EnergyToday) / 1000.0
	yesterdayKWh := float64(yesterdayProdWh) / 1000.0

	b.WriteString("  Day-over-Day:\n")
	fmt.Fprintf(&b, "    Yesterday: %8.2f kWh", yesterdayKWh)
	if ratePerKWh > 0 {
		fmt.Fprintf(&b, "  ($%.2f)", yesterdayKWh*ratePerKWh)
	}
	if yesterdayConsWh > 0 {
		fmt.Fprintf(&b, "  |  consumed %.2f kWh", float64(yesterdayConsWh)/1000.0)
	}
	b.WriteString("\n")

	fmt.Fprintf(&b, "    Today:     %8.2f kWh", todayKWh)
	if ratePerKWh > 0 {
		fmt.Fprintf(&b, "  ($%.2f)", todayKWh*ratePerKWh)
	}
	fmt.Fprintf(&b, "  [in progress, %d W now]", summary.CurrentPower)
	b.WriteString("\n")

	if yesterdayProdWh > 0 {
		delta := todayKWh - yesterdayKWh
		pct := (delta / yesterdayKWh) * 100.0
		if delta >= 0 {
			fmt.Fprintf(&b, "    Change:    +%.2f kWh  (+%.1f%%)\n", delta, pct)
		} else {
			fmt.Fprintf(&b, "    Change:    %.2f kWh  (%.1f%%)\n", delta, pct)
		}
	}

	if todayConsWh > 0 {
		todayConsKWh := float64(todayConsWh) / 1000.0
		netKWh := todayKWh - todayConsKWh
		if netKWh >= 0 {
			fmt.Fprintf(&b, "    Grid Export:  %.2f kWh\n", netKWh)
		} else {
			fmt.Fprintf(&b, "    Grid Draw:    %.2f kWh\n", -netKWh)
		}
	}

	if len(monthProd) > 0 {
		stats := ComputeMonthStats(monthProd)
		b.WriteString("\n")
		fmt.Fprintf(&b, "  %s (so far, %d days):\n", monthLabel, stats.DaysCount)
		fmt.Fprintf(&b, "    Total:     %8.2f kWh", stats.TotalKWh)
		if ratePerKWh > 0 {
			fmt.Fprintf(&b, "  ($%.2f)", stats.TotalKWh*ratePerKWh)
		}
		b.WriteString("\n")
		fmt.Fprintf(&b, "    Daily Avg: %8.2f kWh/day\n", stats.DailyAvg)
		fmt.Fprintf(&b, "    Best Day:  %8.2f kWh\n", stats.BestKWh)
		fmt.Fprintf(&b, "    Worst Day: %8.2f kWh\n", stats.WorstKWh)
	}

	return b.String()
}

// writeWeekRow writes a single formatted row to b, adapting columns based on
// whether consumption and rate data are available.
func writeWeekRow(b *strings.Builder, label string, prodKWh, consKWh, netKWh, ratePerKWh float64, hasCons bool) {
	switch {
	case ratePerKWh > 0 && hasCons:
		fmt.Fprintf(b, "  %-12s  %9.2f kWh  %7s  %9.2f kWh  %7s  %7.2f kWh\n",
			label, prodKWh, fmt.Sprintf("$%.2f", prodKWh*ratePerKWh),
			consKWh, fmt.Sprintf("$%.2f", consKWh*ratePerKWh), netKWh)
	case hasCons:
		fmt.Fprintf(b, "  %-12s  %9.2f kWh  %9.2f kWh  %7.2f kWh\n",
			label, prodKWh, consKWh, netKWh)
	case ratePerKWh > 0:
		fmt.Fprintf(b, "  %-12s  %9.2f kWh  %7s\n",
			label, prodKWh, fmt.Sprintf("$%.2f", prodKWh*ratePerKWh))
	default:
		fmt.Fprintf(b, "  %-12s  %9.2f kWh\n", label, prodKWh)
	}
}

// FormatWeekReport formats a complete N-day history table with totals and averages.
// Unlike FormatDailyReport, it contains only complete days — no in-progress today data.
func FormatWeekReport(rows []DailyRow, ratePerKWh float64) string {
	var b strings.Builder

	if len(rows) == 0 {
		return "No data available.\n"
	}

	hasCons := false
	for _, r := range rows {
		if r.ConsumptionWh > 0 {
			hasCons = true
			break
		}
	}

	first := rows[0].Date
	last := rows[len(rows)-1].Date
	fmt.Fprintf(&b, "=== %d-Day Report (%s to %s) ===\n\n", len(rows), first, last)

	switch {
	case ratePerKWh > 0 && hasCons:
		fmt.Fprintf(&b, "  %-12s  %12s  %8s  %12s  %8s  %10s\n",
			"Date", "Production", "Prod $", "Consumption", "Cons $", "Net")
	case hasCons:
		fmt.Fprintf(&b, "  %-12s  %12s  %12s  %10s\n", "Date", "Production", "Consumption", "Net")
	case ratePerKWh > 0:
		fmt.Fprintf(&b, "  %-12s  %12s  %8s\n", "Date", "Production", "Value")
	default:
		fmt.Fprintf(&b, "  %-12s  %12s\n", "Date", "Production")
	}

	var totalProd, totalCons float64
	for _, r := range rows {
		totalProd += r.ProductionKWh
		totalCons += r.ConsumptionKWh
		writeWeekRow(&b, r.Date, r.ProductionKWh, r.ConsumptionKWh, r.NetKWh, ratePerKWh, hasCons)
	}

	n := float64(len(rows))
	avgProd := totalProd / n
	avgCons := totalCons / n

	b.WriteString("  ---\n")
	writeWeekRow(&b, "Total", totalProd, totalCons, totalProd-totalCons, ratePerKWh, hasCons)
	writeWeekRow(&b, "Average", avgProd, avgCons, avgProd-avgCons, ratePerKWh, hasCons)

	return b.String()
}

// FormatMonthReport formats aggregate statistics for a single calendar month.
// Use this to get total production, daily average, best/worst day, and consumption
// for any given month.
func FormatMonthReport(systemID, month string, prod *EnergyLifetime, cons *ConsumptionLifetime, ratePerKWh float64) string {
	var b strings.Builder

	fmt.Fprintf(&b, "=== Monthly Report: %s (System %s) ===\n\n", month, systemID)

	if prod == nil || len(prod.Production) == 0 {
		b.WriteString("  No production data available.\n")
		return b.String()
	}

	stats := ComputeMonthStats(prod.Production)

	fmt.Fprintf(&b, "  %-22s  %12s\n", "Metric", "Value")
	fmt.Fprintf(&b, "  %-22s  %12s\n", "----------------------", "------------")
	fmt.Fprintf(&b, "  %-22s  %9.1f kWh\n", "Total Production", stats.TotalKWh)
	if ratePerKWh > 0 {
		fmt.Fprintf(&b, "  %-22s  %11s\n", "Est. Value", fmt.Sprintf("$%.2f", stats.TotalKWh*ratePerKWh))
	}
	fmt.Fprintf(&b, "  %-22s  %9.1f kWh\n", "Daily Average", stats.DailyAvg)
	fmt.Fprintf(&b, "  %-22s  %9.1f kWh\n", "Best Day", stats.BestKWh)
	fmt.Fprintf(&b, "  %-22s  %9.1f kWh\n", "Worst Day", stats.WorstKWh)
	fmt.Fprintf(&b, "  %-22s  %12d\n", "Days > 15 kWh", stats.Above15)
	fmt.Fprintf(&b, "  %-22s  %12d\n", "Days < 5 kWh", stats.Below5)
	fmt.Fprintf(&b, "  %-22s  %12d\n", "Days Reported", stats.DaysCount)

	if cons != nil && len(cons.Consumption) > 0 {
		consStats := ComputeMonthStats(cons.Consumption)
		b.WriteString("\n")
		fmt.Fprintf(&b, "  %-22s  %9.1f kWh\n", "Total Consumption", consStats.TotalKWh)
		if ratePerKWh > 0 {
			fmt.Fprintf(&b, "  %-22s  %11s\n", "Consumption Cost", fmt.Sprintf("$%.2f", consStats.TotalKWh*ratePerKWh))
		}
		netKWh := stats.TotalKWh - consStats.TotalKWh
		fmt.Fprintf(&b, "  %-22s  %9.1f kWh\n", "Net (Prod-Cons)", netKWh)
		offsetPct := (stats.TotalKWh / consStats.TotalKWh) * 100.0
		fmt.Fprintf(&b, "  %-22s  %11.0f%%\n", "Solar Offset", offsetPct)
	}

	return b.String()
}

// WriteHistoryFile writes the history JSON to the given path.
func WriteHistoryFile(path string, systemID int, ratePerKWh float64, records []HistoryRecord) error {
	hf := HistoryFile{
		UpdatedAt:    time.Now().UTC().Format(time.RFC3339),
		RatePerKWh:   ratePerKWh,
		SystemID:     systemID,
		TotalRecords: len(records),
		Records:      records,
	}
	data, err := json.MarshalIndent(hf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling history: %w", err)
	}
	if err := atomicWriteFile(path, data); err != nil {
		return fmt.Errorf("writing history file: %w", err)
	}
	return nil
}
