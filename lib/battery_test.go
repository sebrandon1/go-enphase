package lib

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetBatteryStatusSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v4/systems/123/battery_lifetime" {
			t.Errorf("Expected battery_lifetime path, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"normal","battery_count":2,"charge_percent":85.5}`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	status, err := client.GetBatteryStatus("123")
	if err != nil {
		t.Fatalf("GetBatteryStatus failed: %v", err)
	}

	if status.BatteryCount != 2 {
		t.Errorf("Expected 2 batteries, got %d", status.BatteryCount)
	}
	if status.ChargePercent != 85.5 {
		t.Errorf("Expected charge 85.5%%, got %f", status.ChargePercent)
	}
}

func TestGetBatteryStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.GetBatteryStatus("999")
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
}

func TestGetBatteryStatusBadURL(t *testing.T) {
	originalURL := CloudBaseURL
	CloudBaseURL = "://bad"
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.GetBatteryStatus("123")
	if err == nil {
		t.Error("Expected error for bad URL, got nil")
	}
}
