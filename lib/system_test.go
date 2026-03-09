package lib

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListSystemsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v4/systems" {
			t.Errorf("Expected path /api/v4/systems, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"systems":[{"system_id":123,"name":"My System","status":"normal"}]}`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	systems, err := client.ListSystems()
	if err != nil {
		t.Fatalf("ListSystems failed: %v", err)
	}

	if len(systems) != 1 {
		t.Fatalf("Expected 1 system, got %d", len(systems))
	}
	if systems[0].SystemID != 123 {
		t.Errorf("Expected SystemID 123, got %d", systems[0].SystemID)
	}
	if systems[0].Name != "My System" {
		t.Errorf("Expected name 'My System', got '%s'", systems[0].Name)
	}
}

func TestListSystemsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.ListSystems()
	if err == nil {
		t.Error("Expected error for 401, got nil")
	}
}

func TestListSystemsInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.ListSystems()
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestListSystemsBadURL(t *testing.T) {
	originalURL := CloudBaseURL
	CloudBaseURL = "://bad"
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.ListSystems()
	if err == nil {
		t.Error("Expected error for bad URL, got nil")
	}
}

func TestGetSystemSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v4/systems/123" {
			t.Errorf("Expected path /api/v4/systems/123, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"system_id":123,"name":"Test System"}`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	system, err := client.GetSystem("123")
	if err != nil {
		t.Fatalf("GetSystem failed: %v", err)
	}

	if system.Name != "Test System" {
		t.Errorf("Expected 'Test System', got '%s'", system.Name)
	}
}

func TestGetSystemSummarySuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v4/systems/123/summary" {
			t.Errorf("Expected path /api/v4/systems/123/summary, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"system_id":123,"current_power":5000,"energy_today":25000}`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	summary, err := client.GetSystemSummary("123")
	if err != nil {
		t.Fatalf("GetSystemSummary failed: %v", err)
	}

	if summary.CurrentPower != 5000 {
		t.Errorf("Expected CurrentPower 5000, got %d", summary.CurrentPower)
	}
	if summary.EnergyToday != 25000 {
		t.Errorf("Expected EnergyToday 25000, got %d", summary.EnergyToday)
	}
}

func TestListDevicesSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v4/systems/123/devices" {
			t.Errorf("Expected path /api/v4/systems/123/devices, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"micro_inverters":[{"id":1,"serial_number":"SN001","model":"IQ8"}]}`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	devices, err := client.ListDevices("123")
	if err != nil {
		t.Fatalf("ListDevices failed: %v", err)
	}

	if len(devices) != 1 {
		t.Fatalf("Expected 1 device, got %d", len(devices))
	}
	if devices[0].SerialNumber != "SN001" {
		t.Errorf("Expected serial 'SN001', got '%s'", devices[0].SerialNumber)
	}
}

func TestListDevicesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.ListDevices("999")
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
}

func TestListDevicesBadURL(t *testing.T) {
	originalURL := CloudBaseURL
	CloudBaseURL = "://bad"
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.ListDevices("123")
	if err == nil {
		t.Error("Expected error for bad URL, got nil")
	}
}

func TestListDevicesInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.ListDevices("123")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestListSystemsConnectionRefused(t *testing.T) {
	originalURL := CloudBaseURL
	CloudBaseURL = "http://localhost:1"
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.ListSystems()
	if err == nil {
		t.Error("Expected error for connection refused, got nil")
	}
}

func TestListDevicesConnectionRefused(t *testing.T) {
	originalURL := CloudBaseURL
	CloudBaseURL = "http://localhost:1"
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.ListDevices("123")
	if err == nil {
		t.Error("Expected error for connection refused, got nil")
	}
}
