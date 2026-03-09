package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var envoyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get local Envoy production and consumption",
	Run: func(cmd *cobra.Command, args []string) {
		client := getEnvoyClient()
		production, err := client.GetEnvoyProduction()
		if err != nil {
			fmt.Printf("Error getting envoy status: %v\n", err)
			os.Exit(1)
		}
		printJSON(production)
	},
}

var envoySensorsCmd = &cobra.Command{
	Use:   "sensors",
	Short: "Get local Envoy sensor readings",
	Run: func(cmd *cobra.Command, args []string) {
		client := getEnvoyClient()
		sensors, err := client.GetEnvoySensors()
		if err != nil {
			fmt.Printf("Error getting sensors: %v\n", err)
			os.Exit(1)
		}
		printJSON(sensors)
	},
}

func init() {
	envoyCmd.AddCommand(envoyStatusCmd)
	envoyCmd.AddCommand(envoySensorsCmd)
}
