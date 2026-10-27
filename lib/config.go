package lib

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sebrandon1/go-toolbox/confkit"
)

const defaultConfigPath = ".enphase/config"

// Config holds Enphase configuration loaded from ~/.enphase/config.
type Config struct {
	Path         string
	APIKey       string
	AccessToken  string
	RefreshToken string
	ClientID     string
	ClientSecret string
	SystemID     string
	RatePerKWh   string
	RedirectURI  string
	EnvoyIP      string
	EnvoyToken   string
	EnvoySerial  string
}

// DefaultConfigPath returns the default config file path (~/.enphase/config).
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, defaultConfigPath)
}

// LoadConfig reads an Enphase config file (KEY=VALUE format).
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	vars, err := confkit.LoadKV(path)
	if err != nil {
		return nil, err
	}

	// Support ENVOY_IP / ENVOY_TOKEN / ENVOY_SERIAL without the ENPHASE_ prefix
	// as well as the prefixed variants for consistency.
	envoyIP := vars["ENPHASE_ENVOY_IP"]
	if envoyIP == "" {
		envoyIP = vars["ENVOY_IP"]
	}
	envoyToken := vars["ENPHASE_ENVOY_TOKEN"]
	if envoyToken == "" {
		envoyToken = vars["ENVOY_TOKEN"]
	}
	envoySerial := vars["ENPHASE_ENVOY_SERIAL"]
	if envoySerial == "" {
		envoySerial = vars["ENVOY_SERIAL"]
	}

	return &Config{
		Path:         path,
		APIKey:       vars["ENPHASE_API_KEY"],
		AccessToken:  vars["ENPHASE_ACCESS_TOKEN"],
		RefreshToken: vars["ENPHASE_REFRESH_TOKEN"],
		ClientID:     vars["ENPHASE_CLIENT_ID"],
		ClientSecret: vars["ENPHASE_CLIENT_SECRET"],
		SystemID:     vars["ENPHASE_SYSTEM_ID"],
		RatePerKWh:   vars["ENPHASE_RATE_PER_KWH"],
		RedirectURI:  vars["ENPHASE_REDIRECT_URI"],
		EnvoyIP:      envoyIP,
		EnvoyToken:   envoyToken,
		EnvoySerial:  envoySerial,
	}, nil
}

// SaveTokens updates the access and refresh tokens in the config file,
// preserving all other lines (comments, other vars).
func (c *Config) SaveTokens(accessToken, refreshToken string) error {
	path := c.Path
	if path == "" {
		path = DefaultConfigPath()
	}

	// Keep the historical behavior of failing when the config file does not
	// exist (confkit.UpdateKV would create it).
	if _, err := os.ReadFile(path); err != nil {
		return fmt.Errorf("cannot read config for update: %w", err)
	}

	return confkit.UpdateKV(path, map[string]string{
		"ENPHASE_ACCESS_TOKEN":  accessToken,
		"ENPHASE_REFRESH_TOKEN": refreshToken,
	})
}
