package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/sebrandon1/go-enphase/lib"
	"github.com/spf13/cobra"
)

var envoyStatusCmd = &cobra.Command{
	Use:   cmdStatus,
	Short: "Get local Envoy production and consumption",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getEnvoyClient()
		if err != nil {
			return err
		}
		production, err := client.GetEnvoyProduction()
		if err != nil {
			return fmt.Errorf("error getting envoy status: %w", err)
		}
		if isJSONOutput() {
			printJSON(production)
			return nil
		}
		fmt.Print(lib.FormatEnvoyProduction(production))
		return nil
	},
}

var envoySensorsCmd = &cobra.Command{
	Use:   "sensors",
	Short: "Get local Envoy sensor readings",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getEnvoyClient()
		if err != nil {
			return err
		}
		sensors, err := client.GetEnvoySensors()
		if err != nil {
			return fmt.Errorf("error getting sensors: %w", err)
		}
		if isJSONOutput() {
			printJSON(sensors)
			return nil
		}
		fmt.Print(lib.FormatEnvoySensors(sensors))
		return nil
	},
}

var envoyInvertersCmd = &cobra.Command{
	Use:   cmdInverters,
	Short: "Get per-inverter production data from local Envoy",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getEnvoyClient()
		if err != nil {
			return err
		}
		readings, err := client.GetInverterReadings()
		if err != nil {
			return fmt.Errorf("error getting inverter readings: %w", err)
		}
		if isJSONOutput() {
			printJSON(readings)
			return nil
		}
		fmt.Print(lib.FormatInverterReadings(readings))
		return nil
	},
}

var envoyMetersCmd = &cobra.Command{
	Use:   cmdMeters,
	Short: "Get meter configuration from local Envoy",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getEnvoyClient()
		if err != nil {
			return err
		}
		configs, err := client.GetMeterConfig()
		if err != nil {
			return fmt.Errorf("error getting meter config: %w", err)
		}
		if isJSONOutput() {
			printJSON(configs)
			return nil
		}
		fmt.Print(lib.FormatMeterConfig(configs))
		return nil
	},
}

var envoyMeterReadingsCmd = &cobra.Command{
	Use:   cmdMeterReadings,
	Short: "Get latest meter readings from local Envoy",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getEnvoyClient()
		if err != nil {
			return err
		}
		data, err := client.GetMeterReadings()
		if err != nil {
			return fmt.Errorf("error getting meter readings: %w", err)
		}
		if isJSONOutput() {
			printJSON(data)
			return nil
		}
		fmt.Print(lib.FormatMeterData(data))
		return nil
	},
}

var envoyStreamCmd = &cobra.Command{
	Use:   cmdStream,
	Short: "Stream live meter events from local Envoy (press Ctrl+C to stop)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getEnvoyClient()
		if err != nil {
			return err
		}
		jsonOut := isJSONOutput()
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		client.StreamMeter(ctx, func(event *lib.StreamMeterEvent, err error) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
				return
			}
			if jsonOut {
				printJSON(event)
				return
			}
			ts := time.Unix(event.Timestamp, 0).Format("2006-01-02 15:04:05")
			fmt.Printf("%s  eid=%-12d  watts=%8.1f W\n", ts, event.EID, event.ActPower)
		})
		return nil
	},
}

func init() {
	envoyCmd.AddCommand(envoyStatusCmd)
	envoyCmd.AddCommand(envoySensorsCmd)
	envoyCmd.AddCommand(envoyInvertersCmd)
	envoyCmd.AddCommand(envoyMetersCmd)
	envoyCmd.AddCommand(envoyMeterReadingsCmd)
	envoyCmd.AddCommand(envoyStreamCmd)
}
