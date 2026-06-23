package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeConfigFile(t *testing.T, dir string, cfg map[string]any) string {
	t.Helper()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestLoadConfigBasic(t *testing.T) {
	dir := t.TempDir()
	path := writeConfigFile(t, dir, map[string]any{
		"api_key":      "testkey",
		"access_token": "testtoken",
		"system_id":    "12345",
	})

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.APIKey != "testkey" {
		t.Errorf("Expected api_key 'testkey', got %q", cfg.APIKey)
	}
	if cfg.AccessToken != "testtoken" {
		t.Errorf("Expected access_token 'testtoken', got %q", cfg.AccessToken)
	}
	if cfg.SystemID != "12345" {
		t.Errorf("Expected system_id '12345', got %q", cfg.SystemID)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	path := writeConfigFile(t, dir, map[string]any{
		"api_key":      "k",
		"access_token": "t",
	})

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.MetricsAddr != ":9090" {
		t.Errorf("Expected default MetricsAddr ':9090', got %q", cfg.MetricsAddr)
	}
	if cfg.PollInterval != "30s" {
		t.Errorf("Expected default PollInterval '30s', got %q", cfg.PollInterval)
	}
	if cfg.MQTTTopicPrefix != "homeassistant" {
		t.Errorf("Expected default MQTTTopicPrefix 'homeassistant', got %q", cfg.MQTTTopicPrefix)
	}
}

func TestLoadConfigWithRefreshCredentials(t *testing.T) {
	dir := t.TempDir()
	path := writeConfigFile(t, dir, map[string]any{
		"api_key":       "k",
		"access_token":  "t",
		"refresh_token": "r",
		"client_id":     "cid",
		"client_secret": "csec",
	})

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.RefreshToken != "r" {
		t.Errorf("Expected refresh_token 'r', got %q", cfg.RefreshToken)
	}
	if cfg.ClientID != "cid" {
		t.Errorf("Expected client_id 'cid', got %q", cfg.ClientID)
	}
	if cfg.ClientSecret != "csec" {
		t.Errorf("Expected client_secret 'csec', got %q", cfg.ClientSecret)
	}
}

func TestLoadConfigOverridesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := writeConfigFile(t, dir, map[string]any{
		"api_key":           "k",
		"access_token":      "t",
		"metrics_addr":      ":8080",
		"poll_interval":     "60s",
		"mqtt_topic_prefix": "myprefix",
	})

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.MetricsAddr != ":8080" {
		t.Errorf("Expected MetricsAddr ':8080', got %q", cfg.MetricsAddr)
	}
	if cfg.PollInterval != "60s" {
		t.Errorf("Expected PollInterval '60s', got %q", cfg.PollInterval)
	}
	if cfg.MQTTTopicPrefix != "myprefix" {
		t.Errorf("Expected MQTTTopicPrefix 'myprefix', got %q", cfg.MQTTTopicPrefix)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("not json"), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := LoadConfig(path)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
