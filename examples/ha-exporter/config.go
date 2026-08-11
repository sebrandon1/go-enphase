package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
)

// Config holds all runtime configuration for the ha-exporter.
type Config struct {
	// Enphase cloud credentials
	APIKey       string `json:"api_key"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	SystemID     string `json:"system_id"`

	// Local Envoy
	EnvoyIP     string `json:"envoy_ip"`
	EnvoyToken  string `json:"envoy_token"`
	EnvoySerial string `json:"envoy_serial"`

	// Collection settings
	PollInterval string `json:"poll_interval"` // e.g. "30s"

	// Prometheus metrics server
	MetricsAddr string `json:"metrics_addr"` // e.g. ":9090"

	// MQTT / Home Assistant
	MQTTBroker      string `json:"mqtt_broker"`        // e.g. "tcp://192.168.1.10:1883"
	MQTTUsername    string `json:"mqtt_username"`
	MQTTPassword    string `json:"mqtt_password"`
	MQTTTopicPrefix string `json:"mqtt_topic_prefix"`  // e.g. "homeassistant"
	MQTTClientID    string `json:"mqtt_client_id"`     // unique per broker; defaults to "go-enphase-<serial>"
}

// SaveTokens atomically rewrites the config file at path with updated tokens.
func (c *Config) SaveTokens(path, accessToken, refreshToken string) error {
	updated := *c
	updated.AccessToken = accessToken
	updated.RefreshToken = refreshToken

	data, err := json.MarshalIndent(&updated, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		os.Remove(tmp) //nolint:errcheck // best-effort cleanup
		return fmt.Errorf("write config: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp) //nolint:errcheck // best-effort cleanup
		return fmt.Errorf("rename config: %w", err)
	}
	c.AccessToken = accessToken
	c.RefreshToken = refreshToken
	return nil
}

// LoadConfig reads a JSON config file from path and returns the parsed Config.
func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	if info, statErr := f.Stat(); statErr == nil && info.Mode().Perm()&0o077 != 0 {
		Warn("config file %s is readable by group or others (permissions %04o); consider chmod 600", path, info.Mode().Perm())
	}

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.MetricsAddr == "" {
		cfg.MetricsAddr = ":9090"
	}
	if cfg.PollInterval == "" {
		cfg.PollInterval = "30s"
	}
	if cfg.MQTTTopicPrefix == "" {
		cfg.MQTTTopicPrefix = "homeassistant"
	}
	if cfg.MQTTClientID == "" {
		if cfg.EnvoySerial != "" {
			cfg.MQTTClientID = "go-enphase-" + cfg.EnvoySerial
		} else {
			cfg.MQTTClientID = fmt.Sprintf("go-enphase-%08x", rand.Uint32())
		}
	}

	return &cfg, nil
}
