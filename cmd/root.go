package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-enphase/lib"
	"github.com/spf13/cobra"
)

var (
	apiKey       string
	accessToken  string
	refreshToken string
	clientID     string
	clientSecret string
	systemID     string
	envoyIP      string
	envoyToken   string
	envoySerial  string

	cloudClient *lib.Client
	localClient *lib.Client
)

var rootCmd = &cobra.Command{
	Use:   "enphase",
	Short: "Enphase CLI interacts with the Enphase cloud API and local Envoy gateway",
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get data from Enphase cloud API",
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication management",
}

var envoyCmd = &cobra.Command{
	Use:   "envoy",
	Short: "Local Envoy gateway commands",
}

func init() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "help" {
			return nil
		}
		return nil
	}

	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Enphase API key")
	rootCmd.PersistentFlags().StringVar(&accessToken, "access-token", "", "OAuth2 access token")
	rootCmd.PersistentFlags().StringVar(&refreshToken, "refresh-token", "", "OAuth2 refresh token")
	rootCmd.PersistentFlags().StringVar(&clientID, "client-id", "", "OAuth2 client ID")
	rootCmd.PersistentFlags().StringVar(&clientSecret, "client-secret", "", "OAuth2 client secret")
	rootCmd.PersistentFlags().StringVar(&systemID, "system-id", "", "Enphase system ID")
	rootCmd.PersistentFlags().StringVar(&envoyIP, "envoy-ip", "", "Local Envoy gateway IP address")
	rootCmd.PersistentFlags().StringVar(&envoyToken, "envoy-token", "", "Local Envoy JWT token")
	rootCmd.PersistentFlags().StringVar(&envoySerial, "envoy-serial", "", "Envoy serial number")

	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(envoyCmd)
}

func requireSystemID() {
	if systemID == "" {
		fmt.Println("Error: --system-id is required")
		os.Exit(1)
	}
}

func getCloudClient() *lib.Client {
	if cloudClient != nil {
		return cloudClient
	}
	client, err := lib.NewClient(apiKey, accessToken)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	cloudClient = client
	return client
}

func getCloudClientWithRefresh() *lib.Client {
	if cloudClient != nil {
		return cloudClient
	}
	client, err := lib.NewClientWithRefresh(apiKey, accessToken, refreshToken, clientID, clientSecret)
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	cloudClient = client
	return client
}

func getEnvoyClient() *lib.Client {
	if localClient != nil {
		return localClient
	}
	client, err := lib.NewEnvoyClient(envoyIP, envoyToken)
	if err != nil {
		fmt.Printf("Error creating envoy client: %v\n", err)
		os.Exit(1)
	}
	localClient = client
	return client
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
