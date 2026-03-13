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
var historyOutput string

var ratePerKWh float64

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Formatted text reports for cron/Discord",
}

var reportTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "Today's solar production summary (formatted text)",
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
		fmt.Print(lib.FormatTodaySummary(summary, todayConsWh, ratePerKWh))
	},
}

var reportCompareCmd = &cobra.Command{
	Use:   "compare [YYYY-MM] [YYYY-MM]",
	Short: "Compare two months of solar production (formatted text)",
	Long:  "Compare two months. If no months given, compares the two most recent complete months.",
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
		fmt.Print(lib.FormatMonthComparison(systemID, monthA, monthB, prodA, prodB, ratePerKWh))
	},
}

var reportDailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Daily report with production, consumption, and dollar values",
	Long:  "Shows today's live production alongside recent daily production vs consumption history.",
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
	reportHistoryCmd.Flags().StringVarP(&historyOutput, "output", "o", "history.json", "Output path for the history JSON file")

	reportCmd.AddCommand(reportTodayCmd)
	reportCmd.AddCommand(reportCompareCmd)
	reportCmd.AddCommand(reportDailyCmd)
	reportCmd.AddCommand(reportHistoryCmd)
	rootCmd.AddCommand(reportCmd)
}
