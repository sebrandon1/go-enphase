package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds all runtime configuration for the ha-exporter.
type Config struct {
	// Enphase cloud credentials
	APIKey      string `json:"api_key"`
	AccessToken string `json:"access_token"`
	SystemID    string `json:"system_id"`

	// Local Envoy
	EnvoyIP     string `json:"envoy_ip"`
	EnvoyToken  string `json:"envoy_token"`
	EnvoySerial string `json:"envoy_serial"`

	// Collection settings
	PollInterval string `json:"poll_interval"` // e.g. "30s"

	// Prometheus metrics server
	MetricsAddr string `json:"metrics_addr"` // e.g. ":9090"

	// MQTT / Home Assistant
	MQTTBroker      string `json:"mqtt_broker"`       // e.g. "tcp://192.168.1.10:1883"
	MQTTUsername    string `json:"mqtt_username"`
	MQTTPassword    string `json:"mqtt_password"`
	MQTTTopicPrefix string `json:"mqtt_topic_prefix"` // e.g. "homeassistant"
}

// LoadConfig reads a JSON config file from path and returns the parsed Config.
func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

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

	return &cfg, nil
}
