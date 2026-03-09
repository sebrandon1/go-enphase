package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	authEmail    string
	authPassword string
	saveTokens   bool
)

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show token status (no secrets displayed)",
	Run: func(cmd *cobra.Command, args []string) {
		status := map[string]any{
			"api_key_set":       apiKey != "",
			"access_token_set":  accessToken != "",
			"refresh_token_set": refreshToken != "",
			"client_id_set":     clientID != "",
			"client_secret_set": clientSecret != "",
		}
		printJSON(status)
	},
}

var authRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh access token",
	Run: func(cmd *cobra.Command, args []string) {
		loadConfigIfAvailable()
		client := getCloudClientWithRefresh()
		token, err := client.RefreshAccessToken()
		if err != nil {
			fmt.Printf("Error refreshing token: %v\n", err)
			os.Exit(1)
		}
		if saveTokens && loadedConfig != nil {
			if err := loadedConfig.SaveTokens(token.AccessToken, token.RefreshToken); err != nil {
				fmt.Printf("Error saving tokens: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Token refreshed and saved.")
		} else {
			printJSON(token)
		}
	},
}

var envoyTokenCmd = &cobra.Command{
	Use:   "envoy-token",
	Short: "Get Envoy JWT token via Enlighten login",
	Run: func(cmd *cobra.Command, args []string) {
		client := getCloudClient()
		token, err := client.GetEnvoyToken(authEmail, authPassword, envoySerial)
		if err != nil {
			fmt.Printf("Error getting envoy token: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(token)
	},
}

func init() {
	authRefreshCmd.Flags().BoolVar(&saveTokens, "save", false, "Save refreshed tokens to config file")
	envoyTokenCmd.Flags().StringVar(&authEmail, "email", "", "Enlighten account email")
	envoyTokenCmd.Flags().StringVar(&authPassword, "password", "", "Enlighten account password")

	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authRefreshCmd)
	authCmd.AddCommand(envoyTokenCmd)
}
