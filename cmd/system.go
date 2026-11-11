package cmd

import (
	"fmt"

	"github.com/sebrandon1/go-enphase/lib"
	"github.com/spf13/cobra"
)

var systemsCmd = &cobra.Command{
	Use:   cmdSystems,
	Short: "List all systems",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getCloudClient()
		if err != nil {
			return err
		}
		systems, err := client.ListSystems()
		if err != nil {
			return fmt.Errorf("error listing systems: %w", err)
		}
		if isJSONOutput() {
			printJSON(systems)
			return nil
		}
		fmt.Print(lib.FormatSystems(systems))
		return nil
	},
}

var summaryCmd = &cobra.Command{
	Use:   cmdSummary,
	Short: "Get system summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireSystemID(); err != nil {
			return err
		}
		client, err := getCloudClient()
		if err != nil {
			return err
		}
		summary, err := client.GetSystemSummary(systemID)
		if err != nil {
			return fmt.Errorf("error getting summary: %w", err)
		}
		if isJSONOutput() {
			printJSON(summary)
			return nil
		}
		fmt.Print(lib.FormatSystemSummary(summary))
		return nil
	},
}

var devicesCmd = &cobra.Command{
	Use:   cmdDevices,
	Short: "List devices in a system",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireSystemID(); err != nil {
			return err
		}
		client, err := getCloudClient()
		if err != nil {
			return err
		}
		devices, err := client.ListDevices(systemID)
		if err != nil {
			return fmt.Errorf("error listing devices: %w", err)
		}
		if isJSONOutput() {
			printJSON(devices)
			return nil
		}
		fmt.Print(lib.FormatDevices(devices))
		return nil
	},
}

var systemCmd = &cobra.Command{
	Use:   cmdSystem,
	Short: "Get details for a specific system",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireSystemID(); err != nil {
			return err
		}
		client, err := getCloudClient()
		if err != nil {
			return err
		}
		system, err := client.GetSystem(systemID)
		if err != nil {
			return fmt.Errorf("error getting system: %w", err)
		}
		if isJSONOutput() {
			printJSON(system)
			return nil
		}
		fmt.Print(lib.FormatSystem(system))
		return nil
	},
}

func init() {
	getCmd.AddCommand(systemsCmd)
	getCmd.AddCommand(systemCmd)
	getCmd.AddCommand(summaryCmd)
	getCmd.AddCommand(devicesCmd)
}
