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
	RunE: func(cmd *cobra.Command, args []string) error {
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
		return nil
	},
}

var authRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh access token",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getCloudClient()
		if err != nil {
			return err
		}
		token, err := client.RefreshAccessToken()
		if err != nil {
			return fmt.Errorf("error refreshing token: %w", err)
		}
		if saveTokens && loadedConfig != nil {
			if err := loadedConfig.SaveTokens(token.AccessToken, token.RefreshToken); err != nil {
				return fmt.Errorf("error saving tokens: %w", err)
			}
			fmt.Println("Token refreshed and saved.")
		} else {
			printJSON(token)
		}
		return nil
	},
}

var envoyTokenCmd = &cobra.Command{
	Use:   cmdEnvoyToken,
	Short: "Get Envoy JWT token via Enlighten login",
	RunE: func(cmd *cobra.Command, args []string) error {
		if authPassword == "" && !authPasswordStdin {
			return fmt.Errorf("one of --password or --password-stdin is required")
		}
		if authPassword != "" && authPasswordStdin {
			return fmt.Errorf("--password and --password-stdin are mutually exclusive")
		}

		password := authPassword
		if authPasswordStdin {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("error reading password from stdin: %w", err)
			}
			password = strings.TrimRight(string(data), "\r\n")
		}

		client, err := getCloudClient()
		if err != nil {
			return err
		}
		token, err := client.GetEnvoyToken(authEmail, password, envoySerial)
		if err != nil {
			return fmt.Errorf("error getting envoy token: %w", err)
		}
		fmt.Println(token)
		return nil
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
