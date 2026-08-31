package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStderr runs f and returns anything written to os.Stderr.
func captureStderr(f func()) string {
	orig := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f()

	w.Close()
	os.Stderr = orig

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestNewCollectorWarnsWhenCloudUnconfigured(t *testing.T) {
	cfg := &Config{} // no APIKey or AccessToken

	out := captureStderr(func() {
		_, _ = NewCollector(cfg, "")
	})

	if !strings.Contains(out, "cloud client not configured") {
		t.Errorf("expected cloud-not-configured warning, got: %s", out)
	}
}

func TestNewCollectorWarnsWhenEnvoyUnconfigured(t *testing.T) {
	cfg := &Config{
		APIKey:      "key",
		AccessToken: "token",
		SystemID:    "12345",
		// no EnvoyIP
	}

	out := captureStderr(func() {
		_, _ = NewCollector(cfg, "")
	})

	if !strings.Contains(out, "Envoy client not configured") {
		t.Errorf("expected envoy-not-configured warning, got: %s", out)
	}
}

func TestNewCollectorWarnsWhenSystemIDMissing(t *testing.T) {
	cfg := &Config{
		APIKey:      "key",
		AccessToken: "token",
		// no SystemID, no EnvoyIP
	}

	out := captureStderr(func() {
		_, _ = NewCollector(cfg, "")
	})

	if !strings.Contains(out, "system_id is not set") {
		t.Errorf("expected system_id warning, got: %s", out)
	}
}
