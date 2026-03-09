package lib

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")

	content := `ENPHASE_API_KEY=test-key
ENPHASE_ACCESS_TOKEN=test-token
ENPHASE_SYSTEM_ID=12345
ENPHASE_RATE_PER_KWH=0.12
# This is a comment
ENPHASE_CLIENT_ID=client-id
ENPHASE_CLIENT_SECRET=client-secret
ENPHASE_REFRESH_TOKEN=refresh-token
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.APIKey != "test-key" {
		t.Errorf("Expected APIKey 'test-key', got %q", cfg.APIKey)
	}
	if cfg.SystemID != "12345" {
		t.Errorf("Expected SystemID '12345', got %q", cfg.SystemID)
	}
	if cfg.RatePerKWh != "0.12" {
		t.Errorf("Expected RatePerKWh '0.12', got %q", cfg.RatePerKWh)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config")
	if err == nil {
		t.Error("Expected error for missing config file")
	}
}

func TestSaveTokens(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")

	content := `ENPHASE_API_KEY=test-key
ENPHASE_ACCESS_TOKEN=old-token
ENPHASE_REFRESH_TOKEN=old-refresh
ENPHASE_SYSTEM_ID=12345
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{Path: configPath}
	if err := cfg.SaveTokens("new-token", "new-refresh"); err != nil {
		t.Fatalf("SaveTokens failed: %v", err)
	}

	updated, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig after save failed: %v", err)
	}

	if updated.AccessToken != "new-token" {
		t.Errorf("Expected AccessToken 'new-token', got %q", updated.AccessToken)
	}
	if updated.RefreshToken != "new-refresh" {
		t.Errorf("Expected RefreshToken 'new-refresh', got %q", updated.RefreshToken)
	}
	if updated.APIKey != "test-key" {
		t.Errorf("Expected APIKey preserved as 'test-key', got %q", updated.APIKey)
	}
	if updated.SystemID != "12345" {
		t.Errorf("Expected SystemID preserved as '12345', got %q", updated.SystemID)
	}
}

func TestSaveTokensAppendsNew(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")

	content := `ENPHASE_API_KEY=test-key
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{Path: configPath}
	if err := cfg.SaveTokens("new-token", "new-refresh"); err != nil {
		t.Fatalf("SaveTokens failed: %v", err)
	}

	updated, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig after save failed: %v", err)
	}

	if updated.AccessToken != "new-token" {
		t.Errorf("Expected AccessToken 'new-token', got %q", updated.AccessToken)
	}
	if updated.RefreshToken != "new-refresh" {
		t.Errorf("Expected RefreshToken 'new-refresh', got %q", updated.RefreshToken)
	}
}
