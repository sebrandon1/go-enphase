package lib

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetEnvoyProductionSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/production.json" {
			t.Errorf("Expected path /production.json, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"production":[{"type":"inverters","activeCount":20,"wNow":5000.0}],"consumption":[{"type":"eim","measurementType":"total-consumption","wNow":3000.0}]}`))
	}))
	defer server.Close()

	// Extract host:port from server URL (strip https://)
	client, _ := NewEnvoyClient("localhost", "")
	// Override to use http test server
	var result EnvoyProduction
	err := client.envoyGet(server.URL+"/production.json", &result)
	if err != nil {
		t.Fatalf("GetEnvoyProduction failed: %v", err)
	}

	if len(result.Production) != 1 {
		t.Fatalf("Expected 1 production entry, got %d", len(result.Production))
	}
	if result.Production[0].WNow != 5000.0 {
		t.Errorf("Expected WNow 5000.0, got %f", result.Production[0].WNow)
	}
}

func TestGetEnvoyProductionNoIP(t *testing.T) {
	client := &Client{HTTPClient: newHTTPClient()}
	_, err := client.GetEnvoyProduction()
	if err == nil {
		t.Error("Expected error for missing envoy IP, got nil")
	}
}

func TestGetEnvoySensorsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ivp/sensors/readings_object" {
			t.Errorf("Expected path /ivp/sensors/readings_object, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"sensors":[{"readings":[{"measurementType":"production","activePower":4500.0,"rmsVoltage":240.5}]}]}`))
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "")
	var result sensorReadingsResponse
	err := client.envoyGet(server.URL+"/ivp/sensors/readings_object", &result)
	if err != nil {
		t.Fatalf("GetEnvoySensors failed: %v", err)
	}

	if len(result.Sensors) != 1 {
		t.Fatalf("Expected 1 sensor, got %d", len(result.Sensors))
	}
	if result.Sensors[0].Readings[0].ActivePower != 4500.0 {
		t.Errorf("Expected ActivePower 4500.0, got %f", result.Sensors[0].Readings[0].ActivePower)
	}
}

func TestGetEnvoySensorsNoIP(t *testing.T) {
	client := &Client{HTTPClient: newHTTPClient()}
	_, err := client.GetEnvoySensors()
	if err == nil {
		t.Error("Expected error for missing envoy IP, got nil")
	}
}

func TestGetEnvoyProductionConnectionRefused(t *testing.T) {
	client, _ := NewEnvoyClient("localhost:1", "")
	_, err := client.GetEnvoyProduction()
	if err == nil {
		t.Error("Expected error for connection refused, got nil")
	}
}

func TestGetEnvoySensorsConnectionRefused(t *testing.T) {
	client, _ := NewEnvoyClient("localhost:1", "")
	_, err := client.GetEnvoySensors()
	if err == nil {
		t.Error("Expected error for connection refused, got nil")
	}
}

func TestGetEnvoySensorsEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"sensors":[]}`))
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "")
	var result sensorReadingsResponse
	err := client.envoyGet(server.URL+"/ivp/sensors/readings_object", &result)
	if err != nil {
		t.Fatalf("GetEnvoySensors with empty failed: %v", err)
	}

	if len(result.Sensors) != 0 {
		t.Errorf("Expected 0 sensors, got %d", len(result.Sensors))
	}
}
