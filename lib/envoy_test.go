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
	client := &Client{HTTPClient: newHTTPClientWithTLS(false)}
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
	client := &Client{HTTPClient: newHTTPClientWithTLS(false)}
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

// --- New endpoint tests ---

func TestGetEnvoySimpleProductionSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/production" {
			t.Errorf("Expected /api/v1/production, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"wattsNow":4000.0,"whToday":20000.0,"whLastSevenDays":80000.0,"whLifetime":999000.0}`))
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "")
	var result EnvoySimpleProduction
	err := client.envoyGet(server.URL+"/api/v1/production", &result)
	if err != nil {
		t.Fatalf("GetEnvoySimpleProduction failed: %v", err)
	}
	if result.WattsNow != 4000.0 {
		t.Errorf("Expected WattsNow 4000.0, got %f", result.WattsNow)
	}
}

func TestGetEnvoySimpleProductionNoIP(t *testing.T) {
	client := &Client{HTTPClient: newHTTPClientWithTLS(false)}
	_, err := client.GetEnvoySimpleProduction()
	if err == nil {
		t.Error("Expected error for missing envoy IP, got nil")
	}
}

func TestGetEnvoySimpleProductionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "")
	err := client.envoyGet(server.URL+"/api/v1/production", &EnvoySimpleProduction{})
	if err == nil {
		t.Error("Expected error for 401, got nil")
	}
}

func TestGetEnvoySimpleProductionInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "")
	err := client.envoyGet(server.URL+"/api/v1/production", &EnvoySimpleProduction{})
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGetInverterReadingsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/production/inverters" {
			t.Errorf("Expected /api/v1/production/inverters, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"serialNumber":"SN001","lastReportDate":1700000000,"lastReportWatts":250,"maxReportWatts":300,"devType":1}]`))
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "")
	var result []InverterReading
	err := client.envoyGet(server.URL+"/api/v1/production/inverters", &result)
	if err != nil {
		t.Fatalf("GetInverterReadings failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("Expected 1 inverter, got %d", len(result))
	}
	if result[0].SerialNumber != "SN001" {
		t.Errorf("Expected serial SN001, got %s", result[0].SerialNumber)
	}
}

func TestGetInverterReadingsNoIP(t *testing.T) {
	client := &Client{HTTPClient: newHTTPClientWithTLS(false)}
	_, err := client.GetInverterReadings()
	if err == nil {
		t.Error("Expected error for missing envoy IP, got nil")
	}
}

func TestGetInverterReadingsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "")
	err := client.envoyGet(server.URL+"/api/v1/production/inverters", &[]InverterReading{})
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
}

func TestGetInverterReadingsInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "")
	err := client.envoyGet(server.URL+"/api/v1/production/inverters", &[]InverterReading{})
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGetMeterConfigSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ivp/meters" {
			t.Errorf("Expected /ivp/meters, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"eid":704643328,"state":"enabled","measurementType":"production","phaseMode":"three","phaseCount":3}]`))
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "")
	var result []MeterConfig
	err := client.envoyGet(server.URL+"/ivp/meters", &result)
	if err != nil {
		t.Fatalf("GetMeterConfig failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("Expected 1 meter, got %d", len(result))
	}
	if result[0].State != "enabled" {
		t.Errorf("Expected state enabled, got %s", result[0].State)
	}
}

func TestGetMeterConfigNoIP(t *testing.T) {
	client := &Client{HTTPClient: newHTTPClientWithTLS(false)}
	_, err := client.GetMeterConfig()
	if err == nil {
		t.Error("Expected error for missing envoy IP, got nil")
	}
}

func TestGetMeterConfigError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "")
	err := client.envoyGet(server.URL+"/ivp/meters", &[]MeterConfig{})
	if err == nil {
		t.Error("Expected error for 401, got nil")
	}
}

func TestGetMeterConfigInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "")
	err := client.envoyGet(server.URL+"/ivp/meters", &[]MeterConfig{})
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestGetMeterReadingsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ivp/meters/readings" {
			t.Errorf("Expected /ivp/meters/readings, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"eid":704643328,"timestamp":1700000000,"activePower":4000.0,"apparentPower":4100.0,"reactivePower":100.0,"whDlvdCum":999000.0,"whRcvdCum":0.0,"rmsCurrent":16.5,"rmsVoltage":240.0,"pwrFactor":0.98,"frequency":60.0}]`))
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "")
	var result []MeterData
	err := client.envoyGet(server.URL+"/ivp/meters/readings", &result)
	if err != nil {
		t.Fatalf("GetMeterReadings failed: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("Expected 1 meter reading, got %d", len(result))
	}
	if result[0].ActPower != 4000.0 {
		t.Errorf("Expected ActPower 4000.0, got %f", result[0].ActPower)
	}
}

func TestGetMeterReadingsNoIP(t *testing.T) {
	client := &Client{HTTPClient: newHTTPClientWithTLS(false)}
	_, err := client.GetMeterReadings()
	if err == nil {
		t.Error("Expected error for missing envoy IP, got nil")
	}
}

func TestGetMeterReadingsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "")
	err := client.envoyGet(server.URL+"/ivp/meters/readings", &[]MeterData{})
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
}

func TestGetMeterReadingsInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "")
	err := client.envoyGet(server.URL+"/ivp/meters/readings", &[]MeterData{})
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
