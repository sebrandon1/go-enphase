package cmd

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/sebrandon1/go-enphase/lib"
	"github.com/spf13/cobra"
)

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
		summary, err := client.GetSystemSummary(systemID)
		if err != nil {
			fmt.Printf("Error getting summary: %v\n", err)
			os.Exit(1)
		}
		resolveRate()
		fmt.Print(lib.FormatTodaySummary(summary, ratePerKWh))
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

	reportCmd.AddCommand(reportTodayCmd)
	reportCmd.AddCommand(reportCompareCmd)
	rootCmd.AddCommand(reportCmd)
}
