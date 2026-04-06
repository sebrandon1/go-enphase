package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-enphase/lib"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

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
	configFile   string

	configRatePerKWh string
	loadedConfig     *lib.Config

	cloudClient *lib.Client
	localClient *lib.Client
)

var rootCmd = &cobra.Command{
	Use:     "enphase",
	Short:   "Enphase CLI interacts with the Enphase cloud API and local Envoy gateway",
	Version: Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		loadConfigIfAvailable()
	},
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
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Enphase API key")
	rootCmd.PersistentFlags().StringVar(&accessToken, "access-token", "", "OAuth2 access token")
	rootCmd.PersistentFlags().StringVar(&refreshToken, "refresh-token", "", "OAuth2 refresh token")
	rootCmd.PersistentFlags().StringVar(&clientID, "client-id", "", "OAuth2 client ID")
	rootCmd.PersistentFlags().StringVar(&clientSecret, "client-secret", "", "OAuth2 client secret")
	rootCmd.PersistentFlags().StringVar(&systemID, "system-id", "", "Enphase system ID")
	rootCmd.PersistentFlags().StringVar(&envoyIP, "envoy-ip", "", "Local Envoy gateway IP address")
	rootCmd.PersistentFlags().StringVar(&envoyToken, "envoy-token", "", "Local Envoy JWT token")
	rootCmd.PersistentFlags().StringVar(&envoySerial, "envoy-serial", "", "Envoy serial number")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Config file path (default: ~/.enphase/config)")

	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(envoyCmd)
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("enphase version", Version)
		},
	})
}

func requireSystemID() {
	if systemID != "" {
		return
	}

	client := getCloudClient()
	systems, err := client.ListSystems()
	if err != nil {
		fmt.Printf("Error: --system-id not provided and failed to list systems: %v\n", err)
		os.Exit(1)
	}
	if len(systems) == 0 {
		fmt.Println("Error: --system-id not provided and no systems found in account")
		os.Exit(1)
	}

	systemID = fmt.Sprintf("%d", systems[0].SystemID)
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

// loadConfigIfAvailable loads config from file, using values as defaults
// (CLI flags take precedence).
func loadConfigIfAvailable() {
	if loadedConfig != nil {
		return
	}

	cfg, err := lib.LoadConfig(configFile)
	if err != nil {
		return
	}
	loadedConfig = cfg

	if apiKey == "" {
		apiKey = cfg.APIKey
	}
	if accessToken == "" {
		accessToken = cfg.AccessToken
	}
	if refreshToken == "" {
		refreshToken = cfg.RefreshToken
	}
	if clientID == "" {
		clientID = cfg.ClientID
	}
	if clientSecret == "" {
		clientSecret = cfg.ClientSecret
	}
	if systemID == "" {
		systemID = cfg.SystemID
	}
	if envoyIP == "" {
		envoyIP = cfg.EnvoyIP
	}
	if envoyToken == "" {
		envoyToken = cfg.EnvoyToken
	}
	if envoySerial == "" {
		envoySerial = cfg.EnvoySerial
	}
	configRatePerKWh = cfg.RatePerKWh
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
