package lib

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open config: %w", err)
	}
	defer f.Close()

	vars := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok {
			vars[k] = v
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
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

	input, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read config for update: %w", err)
	}

	updates := map[string]string{
		"ENPHASE_ACCESS_TOKEN":  accessToken,
		"ENPHASE_REFRESH_TOKEN": refreshToken,
	}

	var lines []string
	seen := map[string]bool{}
	for _, line := range strings.Split(string(input), "\n") {
		trimmed := strings.TrimSpace(line)
		if k, _, ok := strings.Cut(trimmed, "="); ok {
			if newVal, updating := updates[k]; updating {
				lines = append(lines, k+"="+newVal)
				seen[k] = true
				continue
			}
		}
		lines = append(lines, line)
	}

	for k, v := range updates {
		if !seen[k] {
			lines = append(lines, k+"="+v)
		}
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0600)
}
