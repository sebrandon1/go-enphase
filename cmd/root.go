package cmd

import (
	"fmt"
	"log/slog"
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

	configRatePerKWh    string
	loadedConfig        *lib.Config
	configLoadAttempted bool

	verbose bool

	cloudClient *lib.Client
	localClient *lib.Client
)

var rootCmd = &cobra.Command{
	Use:     "go-enphase",
	Short:   "Enphase CLI interacts with the Enphase cloud API and local Envoy gateway",
	Version: Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		loadConfigIfAvailable()
	},
}

var getCmd = &cobra.Command{
	Use:   cmdGet,
	Short: "Get data from Enphase cloud API",
}

var authCmd = &cobra.Command{
	Use:   cmdAuth,
	Short: "Authentication management",
}

var envoyCmd = &cobra.Command{
	Use:   cmdEnvoy,
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
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format: text or json")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Log HTTP requests to stderr")

	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(envoyCmd)
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("go-enphase version", Version)
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
		fmt.Fprintf(os.Stderr, "Error: --system-id not provided and failed to list systems: %v\n", err)
		os.Exit(1)
	}
	if len(systems) == 0 {
		fmt.Fprintln(os.Stderr, "Error: --system-id not provided and no systems found in account")
		os.Exit(1)
	}

	systemID = fmt.Sprintf("%d", systems[0].SystemID)
}

func getCloudClient() *lib.Client {
	if cloudClient != nil {
		return cloudClient
	}
	client, err := lib.NewClientWithRefresh(apiKey, accessToken, refreshToken, clientID, clientSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}
	client.OnTokenRefresh = func(at, rt string) {
		accessToken = at
		refreshToken = rt
		if loadedConfig != nil {
			_ = loadedConfig.SaveTokens(at, rt)
		}
	}
	if verbose {
		lib.WithLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))(client)
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
		fmt.Fprintf(os.Stderr, "Error creating envoy client: %v\n", err)
		os.Exit(1)
	}
	if verbose {
		lib.WithLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))(client)
	}
	localClient = client
	return client
}

// loadConfigIfAvailable loads config from file then falls back to environment
// variables. CLI flags always take precedence; config file beats env vars.
func loadConfigIfAvailable() {
	if configLoadAttempted {
		return
	}
	configLoadAttempted = true

	cfg, err := lib.LoadConfig(configFile)
	if err == nil {
		loadedConfig = cfg
		applyConfigFile(cfg)
	}

	// Env vars fill in anything the config file left empty.
	applyEnvFallbacks()
}

func applyConfigFile(cfg *lib.Config) {
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
	if configRatePerKWh == "" {
		configRatePerKWh = cfg.RatePerKWh
	}
}

// applyEnvFallbacks fills in any configuration variables still unset after
// CLI flags and config file have been applied.
func applyEnvFallbacks() {
	if apiKey == "" {
		apiKey = os.Getenv("ENPHASE_API_KEY")
	}
	if accessToken == "" {
		accessToken = os.Getenv("ENPHASE_ACCESS_TOKEN")
	}
	if refreshToken == "" {
		refreshToken = os.Getenv("ENPHASE_REFRESH_TOKEN")
	}
	if clientID == "" {
		clientID = os.Getenv("ENPHASE_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("ENPHASE_CLIENT_SECRET")
	}
	if systemID == "" {
		systemID = os.Getenv("ENPHASE_SYSTEM_ID")
	}
	if envoyIP == "" {
		envoyIP = os.Getenv("ENPHASE_ENVOY_IP")
	}
	if envoyToken == "" {
		envoyToken = os.Getenv("ENPHASE_ENVOY_TOKEN")
	}
	if envoySerial == "" {
		envoySerial = os.Getenv("ENPHASE_ENVOY_SERIAL")
	}
	if configRatePerKWh == "" {
		configRatePerKWh = os.Getenv("ENPHASE_RATE_PER_KWH")
	}
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
