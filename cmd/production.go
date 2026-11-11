package cmd

import (
	"fmt"

	"github.com/sebrandon1/go-enphase/lib"
	"github.com/spf13/cobra"
)

var (
	energyStartDate      string
	energyEndDate        string
	consumptionStartDate string
	consumptionEndDate   string
)

var productionCmd = &cobra.Command{
	Use:   cmdProduction,
	Short: "Get production meter readings",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireSystemID(); err != nil {
			return err
		}
		client, err := getCloudClient()
		if err != nil {
			return err
		}
		readings, err := client.GetProductionMeterReadings(systemID)
		if err != nil {
			return fmt.Errorf("error getting production: %w", err)
		}
		if isJSONOutput() {
			printJSON(readings)
			return nil
		}
		fmt.Print(lib.FormatMeterReadings(readings))
		return nil
	},
}

var energyLifetimeCmd = &cobra.Command{
	Use:   cmdEnergyLifetime,
	Short: "Get lifetime energy production history",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateDateFlag("start-date", energyStartDate); err != nil {
			return err
		}
		if err := validateDateFlag("end-date", energyEndDate); err != nil {
			return err
		}
		if err := requireSystemID(); err != nil {
			return err
		}
		client, err := getCloudClient()
		if err != nil {
			return err
		}
		energy, err := client.GetEnergyLifetime(systemID, energyStartDate, energyEndDate)
		if err != nil {
			return fmt.Errorf("error getting energy lifetime: %w", err)
		}
		if isJSONOutput() {
			printJSON(energy)
			return nil
		}
		fmt.Print(lib.FormatEnergyLifetime(energy))
		return nil
	},
}

var consumptionCmd = &cobra.Command{
	Use:   cmdConsumption,
	Short: "Get lifetime consumption history",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateDateFlag("start-date", consumptionStartDate); err != nil {
			return err
		}
		if err := validateDateFlag("end-date", consumptionEndDate); err != nil {
			return err
		}
		if err := requireSystemID(); err != nil {
			return err
		}
		client, err := getCloudClient()
		if err != nil {
			return err
		}
		consumption, err := client.GetConsumptionLifetime(systemID, consumptionStartDate, consumptionEndDate)
		if err != nil {
			return fmt.Errorf("error getting consumption: %w", err)
		}
		if isJSONOutput() {
			printJSON(consumption)
			return nil
		}
		fmt.Print(lib.FormatConsumptionLifetime(consumption))
		return nil
	},
}

var batteryCmd = &cobra.Command{
	Use:   cmdBattery,
	Short: "Get battery status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireSystemID(); err != nil {
			return err
		}
		client, err := getCloudClient()
		if err != nil {
			return err
		}
		battery, err := client.GetBatteryStatus(systemID)
		if err != nil {
			return fmt.Errorf("error getting battery status: %w", err)
		}
		if isJSONOutput() {
			printJSON(battery)
			return nil
		}
		fmt.Print(lib.FormatBatteryStatus(battery))
		return nil
	},
}

func init() {
	energyLifetimeCmd.Flags().StringVar(&energyStartDate, "start-date", "", "Start date (YYYY-MM-DD)")
	energyLifetimeCmd.Flags().StringVar(&energyEndDate, "end-date", "", "End date (YYYY-MM-DD)")

	consumptionCmd.Flags().StringVar(&consumptionStartDate, "start-date", "", "Start date (YYYY-MM-DD)")
	consumptionCmd.Flags().StringVar(&consumptionEndDate, "end-date", "", "End date (YYYY-MM-DD)")

	getCmd.AddCommand(productionCmd)
	getCmd.AddCommand(energyLifetimeCmd)
	getCmd.AddCommand(consumptionCmd)
	getCmd.AddCommand(batteryCmd)
}
