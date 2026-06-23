package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-enphase/lib"
	"github.com/spf13/cobra"
)

var systemsCmd = &cobra.Command{
	Use:   cmdSystems,
	Short: "List all systems",
	Run: func(cmd *cobra.Command, args []string) {
		client := getCloudClient()
		systems, err := client.ListSystems()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing systems: %v\n", err)
			os.Exit(1)
		}
		if isJSONOutput() {
			printJSON(systems)
			return
		}
		fmt.Print(lib.FormatSystems(systems))
	},
}

var summaryCmd = &cobra.Command{
	Use:   cmdSummary,
	Short: "Get system summary",
	Run: func(cmd *cobra.Command, args []string) {
		requireSystemID()
		client := getCloudClient()
		summary, err := client.GetSystemSummary(systemID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting summary: %v\n", err)
			os.Exit(1)
		}
		if isJSONOutput() {
			printJSON(summary)
			return
		}
		fmt.Print(lib.FormatSystemSummary(summary))
	},
}

var devicesCmd = &cobra.Command{
	Use:   cmdDevices,
	Short: "List devices in a system",
	Run: func(cmd *cobra.Command, args []string) {
		requireSystemID()
		client := getCloudClient()
		devices, err := client.ListDevices(systemID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing devices: %v\n", err)
			os.Exit(1)
		}
		if isJSONOutput() {
			printJSON(devices)
			return
		}
		fmt.Print(lib.FormatDevices(devices))
	},
}

func init() {
	getCmd.AddCommand(systemsCmd)
	getCmd.AddCommand(summaryCmd)
	getCmd.AddCommand(devicesCmd)
}
