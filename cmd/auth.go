package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	authEmail         string
	authPassword      string
	authPasswordStdin bool
	saveTokens        bool
)

const (
	statusAPIKeySet       = "api_key_set"
	statusAccessTokenSet  = "access_token_set"
	statusRefreshTokenSet = "refresh_token_set"
	statusClientIDSet     = "client_id_set"
	statusClientSecretSet = "client_secret_set"
	statusSystemIDSet     = "system_id_set"
	statusEnvoyIPSet      = "envoy_ip_set"
	statusEnvoyTokenSet   = "envoy_token_set"
	statusEnvoySerialSet  = "envoy_serial_set"
)

var authStatusCmd = &cobra.Command{
	Use:   cmdStatus,
	Short: "Show token status (no secrets displayed)",
	Run: func(cmd *cobra.Command, args []string) {
		status := map[string]any{
			statusAPIKeySet:       apiKey != "",
			statusAccessTokenSet:  accessToken != "",
			statusRefreshTokenSet: refreshToken != "",
			statusClientIDSet:     clientID != "",
			statusClientSecretSet: clientSecret != "",
			statusSystemIDSet:     systemID != "",
			statusEnvoyIPSet:      envoyIP != "",
			statusEnvoyTokenSet:   envoyToken != "",
			statusEnvoySerialSet:  envoySerial != "",
		}
		printJSON(status)
	},
}

var authRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh access token",
	Run: func(cmd *cobra.Command, args []string) {
		client := getCloudClient()
		token, err := client.RefreshAccessToken()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error refreshing token: %v\n", err)
			os.Exit(1)
		}
		if saveTokens && loadedConfig != nil {
			if err := loadedConfig.SaveTokens(token.AccessToken, token.RefreshToken); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving tokens: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Token refreshed and saved.")
		} else {
			printJSON(token)
		}
	},
}

var envoyTokenCmd = &cobra.Command{
	Use:   cmdEnvoyToken,
	Short: "Get Envoy JWT token via Enlighten login",
	Run: func(cmd *cobra.Command, args []string) {
		if authPassword == "" && !authPasswordStdin {
			fmt.Fprintln(os.Stderr, "Error: one of --password or --password-stdin is required")
			os.Exit(1)
		}
		if authPassword != "" && authPasswordStdin {
			fmt.Fprintln(os.Stderr, "Error: --password and --password-stdin are mutually exclusive")
			os.Exit(1)
		}

		password := authPassword
		if authPasswordStdin {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading password from stdin: %v\n", err)
				os.Exit(1)
			}
			password = strings.TrimRight(string(data), "\r\n")
		}

		client := getCloudClient()
		token, err := client.GetEnvoyToken(authEmail, password, envoySerial)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting envoy token: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(token)
	},
}

func init() {
	authRefreshCmd.Flags().BoolVar(&saveTokens, "save", false, "Save refreshed tokens to config file")
	envoyTokenCmd.Flags().StringVar(&authEmail, "email", "", "Enlighten account email")
	_ = envoyTokenCmd.MarkFlagRequired("email")
	envoyTokenCmd.Flags().StringVar(&authPassword, "password", "", "Enlighten account password")
	envoyTokenCmd.Flags().BoolVar(&authPasswordStdin, "password-stdin", false, "Read Enlighten password from stdin")

	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authRefreshCmd)
	authCmd.AddCommand(envoyTokenCmd)
}
