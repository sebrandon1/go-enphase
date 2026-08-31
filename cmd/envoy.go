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
	Run: func(cmd *cobra.Command, args []string) {
		client := getEnvoyClient()
		production, err := client.GetEnvoyProduction()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting envoy status: %v\n", err)
			os.Exit(1)
		}
		if isJSONOutput() {
			printJSON(production)
			return
		}
		fmt.Print(lib.FormatEnvoyProduction(production))
	},
}

var envoySensorsCmd = &cobra.Command{
	Use:   "sensors",
	Short: "Get local Envoy sensor readings",
	Run: func(cmd *cobra.Command, args []string) {
		client := getEnvoyClient()
		sensors, err := client.GetEnvoySensors()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting sensors: %v\n", err)
			os.Exit(1)
		}
		if isJSONOutput() {
			printJSON(sensors)
			return
		}
		fmt.Print(lib.FormatEnvoySensors(sensors))
	},
}

var envoyInvertersCmd = &cobra.Command{
	Use:   cmdInverters,
	Short: "Get per-inverter production data from local Envoy",
	Run: func(cmd *cobra.Command, args []string) {
		client := getEnvoyClient()
		readings, err := client.GetInverterReadings()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting inverter readings: %v\n", err)
			os.Exit(1)
		}
		if isJSONOutput() {
			printJSON(readings)
			return
		}
		fmt.Print(lib.FormatInverterReadings(readings))
	},
}

var envoyMetersCmd = &cobra.Command{
	Use:   cmdMeters,
	Short: "Get meter configuration from local Envoy",
	Run: func(cmd *cobra.Command, args []string) {
		client := getEnvoyClient()
		configs, err := client.GetMeterConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting meter config: %v\n", err)
			os.Exit(1)
		}
		if isJSONOutput() {
			printJSON(configs)
			return
		}
		fmt.Print(lib.FormatMeterConfig(configs))
	},
}

var envoyMeterReadingsCmd = &cobra.Command{
	Use:   cmdMeterReadings,
	Short: "Get latest meter readings from local Envoy",
	Run: func(cmd *cobra.Command, args []string) {
		client := getEnvoyClient()
		data, err := client.GetMeterReadings()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting meter readings: %v\n", err)
			os.Exit(1)
		}
		if isJSONOutput() {
			printJSON(data)
			return
		}
		fmt.Print(lib.FormatMeterData(data))
	},
}

var envoyStreamCmd = &cobra.Command{
	Use:   cmdStream,
	Short: "Stream live meter events from local Envoy (press Ctrl+C to stop)",
	Run: func(cmd *cobra.Command, args []string) {
		client := getEnvoyClient()
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
