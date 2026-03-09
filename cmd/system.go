package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var systemsCmd = &cobra.Command{
	Use:   "systems",
	Short: "List all systems",
	Run: func(cmd *cobra.Command, args []string) {
		client := getCloudClient()
		systems, err := client.ListSystems()
		if err != nil {
			fmt.Printf("Error listing systems: %v\n", err)
			os.Exit(1)
		}
		printJSON(systems)
	},
}

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Get system summary",
	Run: func(cmd *cobra.Command, args []string) {
		requireSystemID()
		client := getCloudClient()
		summary, err := client.GetSystemSummary(systemID)
		if err != nil {
			fmt.Printf("Error getting summary: %v\n", err)
			os.Exit(1)
		}
		printJSON(summary)
	},
}

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List devices in a system",
	Run: func(cmd *cobra.Command, args []string) {
		requireSystemID()
		client := getCloudClient()
		devices, err := client.ListDevices(systemID)
		if err != nil {
			fmt.Printf("Error listing devices: %v\n", err)
			os.Exit(1)
		}
		printJSON(devices)
	},
}

func init() {
	getCmd.AddCommand(systemsCmd)
	getCmd.AddCommand(summaryCmd)
	getCmd.AddCommand(devicesCmd)
}
