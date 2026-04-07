package cmd

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sebrandon1/go-enphase/lib"
	"github.com/spf13/cobra"
)

var dailyDays int
var weekDays int
var historyOutput string
var ratePerKWh float64

type todayJSON struct {
	SystemID       int     `json:"system_id"`
	Status         string  `json:"status"`
	Modules        int     `json:"modules"`
	CurrentPowerW  int     `json:"current_power_w"`
	ProductionWh   int     `json:"production_wh"`
	ProductionKWh  float64 `json:"production_kwh"`
	ConsumptionWh  int     `json:"consumption_wh,omitempty"`
	ConsumptionKWh float64 `json:"consumption_kwh,omitempty"`
	LifetimeKWh    float64 `json:"lifetime_kwh"`
	LastReportAt   int64   `json:"last_report_at,omitempty"`
	RatePerKWh     float64 `json:"rate_per_kwh,omitempty"`
	ValueUSD       float64 `json:"value_usd,omitempty"`
}

type summaryDayJSON struct {
	ProductionKWh  float64 `json:"production_kwh"`
	ConsumptionKWh float64 `json:"consumption_kwh,omitempty"`
	ValueUSD       float64 `json:"value_usd,omitempty"`
}

type summaryTodayJSON struct {
	ProductionKWh  float64 `json:"production_kwh"`
	ConsumptionKWh float64 `json:"consumption_kwh,omitempty"`
	CurrentPowerW  int     `json:"current_power_w"`
	ValueUSD       float64 `json:"value_usd,omitempty"`
}

type summaryMonthJSON struct {
	Label       string  `json:"label"`
	DaysCount   int     `json:"days_count"`
	TotalKWh    float64 `json:"total_kwh"`
	DailyAvgKWh float64 `json:"daily_avg_kwh"`
	BestDayKWh  float64 `json:"best_day_kwh"`
	WorstDayKWh float64 `json:"worst_day_kwh"`
	ValueUSD    float64 `json:"value_usd,omitempty"`
}

type summaryJSON struct {
	SystemID  int               `json:"system_id"`
	Yesterday summaryDayJSON    `json:"yesterday"`
	Today     summaryTodayJSON  `json:"today"`
	ChangeKWh float64           `json:"change_kwh"`
	ChangePct float64           `json:"change_pct"`
	Month     *summaryMonthJSON `json:"month,omitempty"`
}

type dailyJSON struct {
	SystemID      int            `json:"system_id"`
	Status        string         `json:"status"`
	CurrentPowerW int            `json:"current_power_w"`
	TodayKWh      float64        `json:"today_kwh"`
	TodayValueUSD float64        `json:"today_value_usd,omitempty"`
	History       []lib.DailyRow `json:"history"`
}

type weekJSON struct {
	Days         int            `json:"days"`
	StartDate    string         `json:"start_date,omitempty"`
	EndDate      string         `json:"end_date,omitempty"`
	Rows         []lib.DailyRow `json:"rows"`
	TotalProdKWh float64        `json:"total_production_kwh"`
	TotalConsKWh float64        `json:"total_consumption_kwh,omitempty"`
	AvgProdKWh   float64        `json:"avg_production_kwh"`
	AvgConsKWh   float64        `json:"avg_consumption_kwh,omitempty"`
}

type monthConsJSON struct {
	TotalKWh    float64 `json:"total_kwh"`
	DailyAvgKWh float64 `json:"daily_avg_kwh"`
	ValueUSD    float64 `json:"value_usd,omitempty"`
}

type monthJSON struct {
	Month       string         `json:"month"`
	SystemID    string         `json:"system_id"`
	Production  lib.MonthStats `json:"production"`
	Consumption *monthConsJSON `json:"consumption,omitempty"`
	NetKWh      float64        `json:"net_kwh,omitempty"`
	SolarOffset float64        `json:"solar_offset_pct,omitempty"`
	RatePerKWh  float64        `json:"rate_per_kwh,omitempty"`
}

type compareMonthJSON struct {
	Month string         `json:"month"`
	Stats lib.MonthStats `json:"stats"`
}

type compareJSON struct {
	SystemID   string           `json:"system_id"`
	MonthA     compareMonthJSON `json:"month_a"`
	MonthB     compareMonthJSON `json:"month_b"`
	RatePerKWh float64          `json:"rate_per_kwh,omitempty"`
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Formatted text reports for cron/Discord",
}

var reportTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "Today's solar production summary (formatted text)",
	Long:  "Live snapshot of today's production: current power, energy produced so far, and lifetime total. Does not include yesterday's data or historical comparisons — use 'report summary' for day-over-day comparison.",
	Run: func(cmd *cobra.Command, args []string) {
		requireSystemID()
		client := getCloudClient()
		resolveRate()

		today := time.Now().Format("2006-01-02")

		var summary *lib.SystemSummary
		var cons *lib.ConsumptionLifetime
		var errSummary, errCons error

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			summary, errSummary = client.GetSystemSummary(systemID)
		}()
		go func() {
			defer wg.Done()
			cons, errCons = client.GetConsumptionLifetime(systemID, today, today)
		}()
		wg.Wait()

		if errSummary != nil {
			fmt.Printf("Error getting summary: %v\n", errSummary)
			os.Exit(1)
		}
		var todayConsWh int
		if errCons == nil && cons != nil && len(cons.Consumption) > 0 {
			todayConsWh = cons.Consumption[0]
		}
		if isJSONOutput() {
			prodKWh := whToKwh(summary.EnergyToday)
			out := todayJSON{
				SystemID:      summary.SystemID,
				Status:        summary.Status,
				Modules:       summary.Modules,
				CurrentPowerW: summary.CurrentPower,
				ProductionWh:  summary.EnergyToday,
				ProductionKWh: prodKWh,
				LifetimeKWh:   whToKwh(summary.EnergyLifetime),
				LastReportAt:  summary.LastReportAt,
				RatePerKWh:    ratePerKWh,
			}
			if todayConsWh > 0 {
				out.ConsumptionWh = todayConsWh
				out.ConsumptionKWh = whToKwh(todayConsWh)
			}
			if ratePerKWh > 0 {
				out.ValueUSD = prodKWh * ratePerKWh
			}
			printJSON(out)
			return
		}
		fmt.Print(lib.FormatTodaySummary(summary, todayConsWh, ratePerKWh))
	},
}

var reportCompareCmd = &cobra.Command{
	Use:   "compare [YYYY-MM] [YYYY-MM]",
	Short: "Compare two months of solar production (formatted text)",
	Long:  "Compare two calendar months side-by-side: total, daily average, best/worst day. If no months given, compares the two most recent complete months. For a single month's stats use 'report month'. For day-over-day (today vs yesterday) use 'report summary'.",
	Run: func(cmd *cobra.Command, args []string) {
		requireSystemID()
		client := getCloudClient()

		var monthA, monthB string
		if len(args) >= 2 {
			monthA, monthB = args[0], args[1]
		} else {
			monthA, monthB = lib.LastTwoMonths()
		}

		startA, endA, err := lib.MonthDateRange(monthA)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		startB, endB, err := lib.MonthDateRange(monthB)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		var prodA, prodB *lib.EnergyLifetime
		var errA, errB error
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			prodA, errA = client.GetEnergyLifetime(systemID, startA, endA)
		}()
		go func() {
			defer wg.Done()
			prodB, errB = client.GetEnergyLifetime(systemID, startB, endB)
		}()
		wg.Wait()
		if errA != nil {
			fmt.Printf("Error fetching %s: %v\n", monthA, errA)
			os.Exit(1)
		}
		if errB != nil {
			fmt.Printf("Error fetching %s: %v\n", monthB, errB)
			os.Exit(1)
		}

		resolveRate()
		if isJSONOutput() {
			printJSON(compareJSON{
				SystemID:   systemID,
				MonthA:     compareMonthJSON{Month: monthA, Stats: lib.ComputeMonthStats(prodA.Production)},
				MonthB:     compareMonthJSON{Month: monthB, Stats: lib.ComputeMonthStats(prodB.Production)},
				RatePerKWh: ratePerKWh,
			})
			return
		}
		fmt.Print(lib.FormatMonthComparison(systemID, monthA, monthB, prodA, prodB, ratePerKWh))
	},
}

var reportDailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Daily report with production, consumption, and dollar values",
	Long:  "Shows today's live production (in progress) alongside a trailing history table of complete days. Includes yesterday's final total. Use --days to control how many trailing days appear (default 7). For a clean complete-days-only table with totals use 'report week'.",
	Run: func(cmd *cobra.Command, args []string) {
		requireSystemID()
		client := getCloudClient()
		resolveRate()

		// Compute date range for trailing N days (yesterday back).
		yesterday := time.Now().AddDate(0, 0, -1)
		startDate := yesterday.AddDate(0, 0, -(dailyDays - 1))

		// Fetch summary, production, and consumption concurrently.
		var summary *lib.SystemSummary
		var prod *lib.EnergyLifetime
		var cons *lib.ConsumptionLifetime
		var errSummary, errProd, errCons error

		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			defer wg.Done()
			summary, errSummary = client.GetSystemSummary(systemID)
		}()
		go func() {
			defer wg.Done()
			prod, errProd = client.GetEnergyLifetime(systemID,
				startDate.Format("2006-01-02"), yesterday.Format("2006-01-02"))
		}()
		go func() {
			defer wg.Done()
			cons, errCons = client.GetConsumptionLifetime(systemID,
				startDate.Format("2006-01-02"), yesterday.Format("2006-01-02"))
		}()
		wg.Wait()

		if errSummary != nil {
			fmt.Printf("Error getting summary: %v\n", errSummary)
			os.Exit(1)
		}
		if errProd != nil {
			fmt.Printf("Error getting production: %v\n", errProd)
			os.Exit(1)
		}
		if errCons != nil {
			// Consumption may not be available on all systems; warn but continue.
			fmt.Printf("Warning: consumption data unavailable: %v\n", errCons)
			cons = nil
		}

		rows := lib.BuildDailyRows(prod, cons, dailyDays)
		if isJSONOutput() {
			todayKWh := whToKwh(summary.EnergyToday)
			out := dailyJSON{
				SystemID:      summary.SystemID,
				Status:        summary.Status,
				CurrentPowerW: summary.CurrentPower,
				TodayKWh:      todayKWh,
				History:       rows,
			}
			if ratePerKWh > 0 {
				out.TodayValueUSD = todayKWh * ratePerKWh
			}
			printJSON(out)
			return
		}
		fmt.Print(lib.FormatDailyReport(summary, rows, ratePerKWh))
	},
}

var reportHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Export combined production/consumption history to JSON",
	Long:  "Fetches all available production and consumption data and writes a combined JSON file.",
	Run: func(cmd *cobra.Command, args []string) {
		requireSystemID()
		client := getCloudClient()
		resolveRate()

		// Fetch all available data (no date filters = full history).
		var prod *lib.EnergyLifetime
		var cons *lib.ConsumptionLifetime
		var errProd, errCons error

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			prod, errProd = client.GetEnergyLifetime(systemID, "", "")
		}()
		go func() {
			defer wg.Done()
			cons, errCons = client.GetConsumptionLifetime(systemID, "", "")
		}()
		wg.Wait()

		if errProd != nil {
			fmt.Printf("Error getting production: %v\n", errProd)
			os.Exit(1)
		}
		if errCons != nil {
			fmt.Printf("Warning: consumption data unavailable: %v\n", errCons)
			cons = nil
		}

		rows := lib.BuildDailyRows(prod, cons, 0)
		records := lib.BuildHistoryRecords(rows, ratePerKWh)

		systemIDInt := 0
		if v, err := strconv.Atoi(systemID); err == nil {
			systemIDInt = v
		}

		if err := lib.WriteHistoryFile(historyOutput, systemIDInt, ratePerKWh, records); err != nil {
			fmt.Printf("Error writing history: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Wrote %d records to %s\n", len(records), historyOutput)
	},
}

var reportSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Day-over-day comparison: today vs yesterday, plus current month stats",
	Long:  "Answers 'how does today compare to yesterday?' — shows yesterday's final production, today's in-progress total, the kWh/percent change, and current month running statistics (total, daily average, best/worst day). This is the best command for a quick daily check-in.",
	Run: func(cmd *cobra.Command, args []string) {
		requireSystemID()
		client := getCloudClient()
		resolveRate()

		now := time.Now()
		today := now.Format("2006-01-02")
		yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		monthLabel := now.Format("January 2006")

		var summary *lib.SystemSummary
		var yCons *lib.ConsumptionLifetime
		var tCons *lib.ConsumptionLifetime
		var mProd *lib.EnergyLifetime
		var errSummary, errYCons, errTCons, errMProd error

		var wg sync.WaitGroup
		wg.Add(4)
		go func() {
			defer wg.Done()
			summary, errSummary = client.GetSystemSummary(systemID)
		}()
		go func() {
			defer wg.Done()
			yCons, errYCons = client.GetConsumptionLifetime(systemID, yesterday, yesterday)
		}()
		go func() {
			defer wg.Done()
			tCons, errTCons = client.GetConsumptionLifetime(systemID, today, today)
		}()
		go func() {
			defer wg.Done()
			mProd, errMProd = client.GetEnergyLifetime(systemID, monthStart, yesterday)
		}()
		wg.Wait()

		if errSummary != nil {
			fmt.Printf("Error getting summary: %v\n", errSummary)
			os.Exit(1)
		}
		if errMProd != nil {
			fmt.Printf("Error getting production: %v\n", errMProd)
			os.Exit(1)
		}

		// yesterday's production is the last element of the month-range response
		var yesterdayProdWh int
		if mProd != nil && len(mProd.Production) > 0 {
			yesterdayProdWh = mProd.Production[len(mProd.Production)-1]
		}
		var yesterdayConsWh int
		if errYCons == nil && yCons != nil && len(yCons.Consumption) > 0 {
			yesterdayConsWh = yCons.Consumption[len(yCons.Consumption)-1]
		}
		var todayConsWh int
		if errTCons == nil && tCons != nil && len(tCons.Consumption) > 0 {
			todayConsWh = tCons.Consumption[0]
		}
		var monthProdValues []int
		if mProd != nil {
			monthProdValues = mProd.Production
		}

		if isJSONOutput() {
			todayKWh := whToKwh(summary.EnergyToday)
			yesterdayKWh := whToKwh(yesterdayProdWh)
			changeKWh := todayKWh - yesterdayKWh
			var changePct float64
			if yesterdayKWh > 0 {
				changePct = (changeKWh / yesterdayKWh) * 100.0
			}
			out := summaryJSON{
				SystemID: summary.SystemID,
				Yesterday: summaryDayJSON{
					ProductionKWh:  yesterdayKWh,
					ConsumptionKWh: whToKwh(yesterdayConsWh),
				},
				Today: summaryTodayJSON{
					ProductionKWh:  todayKWh,
					ConsumptionKWh: whToKwh(todayConsWh),
					CurrentPowerW:  summary.CurrentPower,
				},
				ChangeKWh: changeKWh,
				ChangePct: changePct,
			}
			if ratePerKWh > 0 {
				out.Yesterday.ValueUSD = yesterdayKWh * ratePerKWh
				out.Today.ValueUSD = todayKWh * ratePerKWh
			}
			if len(monthProdValues) > 0 {
				stats := lib.ComputeMonthStats(monthProdValues)
				m := &summaryMonthJSON{
					Label:       monthLabel,
					DaysCount:   stats.DaysCount,
					TotalKWh:    stats.TotalKWh,
					DailyAvgKWh: stats.DailyAvg,
					BestDayKWh:  stats.BestKWh,
					WorstDayKWh: stats.WorstKWh,
				}
				if ratePerKWh > 0 {
					m.ValueUSD = stats.TotalKWh * ratePerKWh
				}
				out.Month = m
			}
			printJSON(out)
			return
		}
		fmt.Print(lib.FormatSummaryReport(summary, yesterdayProdWh, yesterdayConsWh, todayConsWh, monthProdValues, monthLabel, ratePerKWh))
	},
}

var reportWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "Last N complete days: production, consumption, totals, and averages",
	Long:  "Shows a table of the last N complete calendar days (default 7) with production, consumption, and net values, plus totals and averages. Unlike 'report daily', this contains only fully complete days — no in-progress today data. Use --days to change the window.",
	Run: func(cmd *cobra.Command, args []string) {
		requireSystemID()
		client := getCloudClient()
		resolveRate()

		yesterday := time.Now().AddDate(0, 0, -1)
		startDate := yesterday.AddDate(0, 0, -(weekDays - 1))

		var prod *lib.EnergyLifetime
		var cons *lib.ConsumptionLifetime
		var errProd, errCons error

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			prod, errProd = client.GetEnergyLifetime(systemID,
				startDate.Format("2006-01-02"), yesterday.Format("2006-01-02"))
		}()
		go func() {
			defer wg.Done()
			cons, errCons = client.GetConsumptionLifetime(systemID,
				startDate.Format("2006-01-02"), yesterday.Format("2006-01-02"))
		}()
		wg.Wait()

		if errProd != nil {
			fmt.Printf("Error getting production: %v\n", errProd)
			os.Exit(1)
		}
		if errCons != nil {
			fmt.Printf("Warning: consumption data unavailable: %v\n", errCons)
			cons = nil
		}

		rows := lib.BuildDailyRows(prod, cons, weekDays)
		if isJSONOutput() {
			out := weekJSON{Days: len(rows), Rows: rows}
			if len(rows) > 0 {
				out.StartDate = rows[0].Date
				out.EndDate = rows[len(rows)-1].Date
			}
			for _, r := range rows {
				out.TotalProdKWh += r.ProductionKWh
				out.TotalConsKWh += r.ConsumptionKWh
			}
			if n := float64(len(rows)); n > 0 {
				out.AvgProdKWh = out.TotalProdKWh / n
				out.AvgConsKWh = out.TotalConsKWh / n
			}
			printJSON(out)
			return
		}
		fmt.Print(lib.FormatWeekReport(rows, ratePerKWh))
	},
}

var reportMonthCmd = &cobra.Command{
	Use:   "month [YYYY-MM]",
	Short: "Aggregate stats for a calendar month: total, average, best/worst day",
	Long:  "Shows total production, daily average, best day, worst day, and consumption stats for a single calendar month. Defaults to the current month (partial if in progress). For side-by-side comparison of two months use 'report compare'.",
	Run: func(cmd *cobra.Command, args []string) {
		requireSystemID()
		client := getCloudClient()
		resolveRate()

		var month string
		if len(args) >= 1 {
			month = args[0]
		} else {
			month = time.Now().Format("2006-01")
		}

		start, end, err := lib.MonthDateRange(month)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		var prod *lib.EnergyLifetime
		var cons *lib.ConsumptionLifetime
		var errProd, errCons error

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			prod, errProd = client.GetEnergyLifetime(systemID, start, end)
		}()
		go func() {
			defer wg.Done()
			cons, errCons = client.GetConsumptionLifetime(systemID, start, end)
		}()
		wg.Wait()

		if errProd != nil {
			fmt.Printf("Error getting production: %v\n", errProd)
			os.Exit(1)
		}
		if errCons != nil {
			fmt.Printf("Warning: consumption data unavailable: %v\n", errCons)
			cons = nil
		}

		if isJSONOutput() {
			out := monthJSON{Month: month, SystemID: systemID, RatePerKWh: ratePerKWh}
			if prod != nil && len(prod.Production) > 0 {
				out.Production = lib.ComputeMonthStats(prod.Production)
			}
			if cons != nil && len(cons.Consumption) > 0 {
				cs := lib.ComputeMonthStats(cons.Consumption)
				mc := &monthConsJSON{
					TotalKWh:    cs.TotalKWh,
					DailyAvgKWh: cs.DailyAvg,
				}
				if ratePerKWh > 0 {
					mc.ValueUSD = cs.TotalKWh * ratePerKWh
				}
				out.Consumption = mc
				out.NetKWh = out.Production.TotalKWh - cs.TotalKWh
				if cs.TotalKWh > 0 {
					out.SolarOffset = (out.Production.TotalKWh / cs.TotalKWh) * 100.0
				}
			}
			printJSON(out)
			return
		}
		fmt.Print(lib.FormatMonthReport(systemID, month, prod, cons, ratePerKWh))
	},
}

func resolveRate() {
	if ratePerKWh > 0 {
		return
	}
	if configRatePerKWh != "" {
		if v, err := strconv.ParseFloat(configRatePerKWh, 64); err == nil {
			ratePerKWh = v
		}
	}
}

func init() {
	reportCmd.PersistentFlags().Float64Var(&ratePerKWh, "rate", 0, "Electricity rate per kWh (for dollar estimates)")

	reportDailyCmd.Flags().IntVar(&dailyDays, "days", 7, "Number of trailing days to include in the report")
	reportWeekCmd.Flags().IntVar(&weekDays, "days", 7, "Number of complete days to include in the report")
	reportHistoryCmd.Flags().StringVarP(&historyOutput, "file", "f", "history.json", "Output file path for the history JSON")

	reportCmd.AddCommand(reportTodayCmd)
	reportCmd.AddCommand(reportSummaryCmd)
	reportCmd.AddCommand(reportCompareCmd)
	reportCmd.AddCommand(reportDailyCmd)
	reportCmd.AddCommand(reportWeekCmd)
	reportCmd.AddCommand(reportMonthCmd)
	reportCmd.AddCommand(reportHistoryCmd)
	rootCmd.AddCommand(reportCmd)
}
