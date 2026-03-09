package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	startDate string
	endDate   string
)

var productionCmd = &cobra.Command{
	Use:   "production",
	Short: "Get production meter readings",
	Run: func(cmd *cobra.Command, args []string) {
		requireSystemID()
		client := getCloudClient()
		readings, err := client.GetProductionMeterReadings(systemID)
		if err != nil {
			fmt.Printf("Error getting production: %v\n", err)
			os.Exit(1)
		}
		printJSON(readings)
	},
}

var energyLifetimeCmd = &cobra.Command{
	Use:   "energy-lifetime",
	Short: "Get lifetime energy production history",
	Run: func(cmd *cobra.Command, args []string) {
		requireSystemID()
		client := getCloudClient()
		energy, err := client.GetEnergyLifetime(systemID, startDate, endDate)
		if err != nil {
			fmt.Printf("Error getting energy lifetime: %v\n", err)
			os.Exit(1)
		}
		printJSON(energy)
	},
}

var consumptionCmd = &cobra.Command{
	Use:   "consumption",
	Short: "Get lifetime consumption history",
	Run: func(cmd *cobra.Command, args []string) {
		requireSystemID()
		client := getCloudClient()
		consumption, err := client.GetConsumptionLifetime(systemID, startDate, endDate)
		if err != nil {
			fmt.Printf("Error getting consumption: %v\n", err)
			os.Exit(1)
		}
		printJSON(consumption)
	},
}

var batteryCmd = &cobra.Command{
	Use:   "battery",
	Short: "Get battery status",
	Run: func(cmd *cobra.Command, args []string) {
		requireSystemID()
		client := getCloudClient()
		battery, err := client.GetBatteryStatus(systemID)
		if err != nil {
			fmt.Printf("Error getting battery status: %v\n", err)
			os.Exit(1)
		}
		printJSON(battery)
	},
}

func init() {
	energyLifetimeCmd.Flags().StringVar(&startDate, "start-date", "", "Start date (YYYY-MM-DD)")
	energyLifetimeCmd.Flags().StringVar(&endDate, "end-date", "", "End date (YYYY-MM-DD)")

	consumptionCmd.Flags().StringVar(&startDate, "start-date", "", "Start date (YYYY-MM-DD)")
	consumptionCmd.Flags().StringVar(&endDate, "end-date", "", "End date (YYYY-MM-DD)")

	getCmd.AddCommand(productionCmd)
	getCmd.AddCommand(energyLifetimeCmd)
	getCmd.AddCommand(consumptionCmd)
	getCmd.AddCommand(batteryCmd)
}
