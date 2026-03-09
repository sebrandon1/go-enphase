package lib

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetProductionMeterReadingsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v4/systems/123/production_meter_readings" {
			t.Errorf("Expected path /api/v4/systems/123/production_meter_readings, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"meter_readings":[{"serial_number":"SN001","value":1000,"read_at":1700000000}]}`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	readings, err := client.GetProductionMeterReadings("123")
	if err != nil {
		t.Fatalf("GetProductionMeterReadings failed: %v", err)
	}

	if len(readings) != 1 {
		t.Fatalf("Expected 1 reading, got %d", len(readings))
	}
	if readings[0].Value != 1000 {
		t.Errorf("Expected value 1000, got %d", readings[0].Value)
	}
}

func TestGetProductionMeterReadingsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.GetProductionMeterReadings("999")
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
}

func TestGetProductionMeterReadingsBadURL(t *testing.T) {
	originalURL := CloudBaseURL
	CloudBaseURL = "://bad"
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.GetProductionMeterReadings("123")
	if err == nil {
		t.Error("Expected error for bad URL, got nil")
	}
}

func TestGetProductionMeterReadingsInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.GetProductionMeterReadings("123")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGetEnergyLifetimeSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v4/systems/123/energy_lifetime" {
			t.Errorf("Expected energy_lifetime path, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("start_date") != "2024-01-01" {
			t.Errorf("Expected start_date '2024-01-01', got '%s'", r.URL.Query().Get("start_date"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"start_date":"2024-01-01","system_id":123,"production":[100,200,300]}`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	energy, err := client.GetEnergyLifetime("123", "2024-01-01", "")
	if err != nil {
		t.Fatalf("GetEnergyLifetime failed: %v", err)
	}

	if len(energy.Production) != 3 {
		t.Errorf("Expected 3 production entries, got %d", len(energy.Production))
	}
	if energy.StartDate != "2024-01-01" {
		t.Errorf("Expected start date '2024-01-01', got '%s'", energy.StartDate)
	}
}

func TestGetEnergyLifetimeNoParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("Expected no query params, got '%s'", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"start_date":"2020-01-01","system_id":123,"production":[]}`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.GetEnergyLifetime("123", "", "")
	if err != nil {
		t.Fatalf("GetEnergyLifetime with no params failed: %v", err)
	}
}

func TestGetEnergyLifetimeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.GetEnergyLifetime("999", "", "")
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
}

func TestGetEnergyLifetimeBadURL(t *testing.T) {
	originalURL := CloudBaseURL
	CloudBaseURL = "://bad"
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.GetEnergyLifetime("123", "", "")
	if err == nil {
		t.Error("Expected error for bad URL, got nil")
	}
}

func TestGetEnergyLifetimeConnectionRefused(t *testing.T) {
	originalURL := CloudBaseURL
	CloudBaseURL = "http://localhost:1"
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.GetEnergyLifetime("123", "", "")
	if err == nil {
		t.Error("Expected error for connection refused, got nil")
	}
}

func TestGetEnergyLifetimeInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.GetEnergyLifetime("123", "", "")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGetConsumptionLifetimeSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v4/systems/123/consumption_lifetime" {
			t.Errorf("Expected consumption_lifetime path, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"start_date":"2024-01-01","system_id":123,"consumption":[500,600]}`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	consumption, err := client.GetConsumptionLifetime("123", "2024-01-01", "2024-12-31")
	if err != nil {
		t.Fatalf("GetConsumptionLifetime failed: %v", err)
	}

	if len(consumption.Consumption) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(consumption.Consumption))
	}
}

func TestGetConsumptionLifetimeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.GetConsumptionLifetime("999", "", "")
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
}

func TestGetConsumptionLifetimeBadURL(t *testing.T) {
	originalURL := CloudBaseURL
	CloudBaseURL = "://bad"
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.GetConsumptionLifetime("123", "", "")
	if err == nil {
		t.Error("Expected error for bad URL, got nil")
	}
}

func TestGetConsumptionLifetimeConnectionRefused(t *testing.T) {
	originalURL := CloudBaseURL
	CloudBaseURL = "http://localhost:1"
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.GetConsumptionLifetime("123", "", "")
	if err == nil {
		t.Error("Expected error for connection refused, got nil")
	}
}

func TestGetConsumptionLifetimeInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.GetConsumptionLifetime("123", "", "")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGetProductionMeterReadingsConnectionRefused(t *testing.T) {
	originalURL := CloudBaseURL
	CloudBaseURL = "http://localhost:1"
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.GetProductionMeterReadings("123")
	if err == nil {
		t.Error("Expected error for connection refused, got nil")
	}
}
