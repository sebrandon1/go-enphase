package lib

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("key1", "token1")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client.APIKey != "key1" {
		t.Errorf("Expected APIKey 'key1', got '%s'", client.APIKey)
	}

	if client.AccessToken != "token1" {
		t.Errorf("Expected AccessToken 'token1', got '%s'", client.AccessToken)
	}
}

func TestNewClientEmptyAPIKey(t *testing.T) {
	_, err := NewClient("", "token1")
	if err == nil {
		t.Error("Expected error for empty API key, got nil")
	}
}

func TestNewClientEmptyAccessToken(t *testing.T) {
	_, err := NewClient("key1", "")
	if err == nil {
		t.Error("Expected error for empty access token, got nil")
	}
}

func TestNewClientWithRefresh(t *testing.T) {
	client, err := NewClientWithRefresh("key1", "token1", "refresh1", "cid", "csecret")
	if err != nil {
		t.Fatalf("NewClientWithRefresh failed: %v", err)
	}

	if client.RefreshToken != "refresh1" {
		t.Errorf("Expected RefreshToken 'refresh1', got '%s'", client.RefreshToken)
	}
	if client.ClientID != "cid" {
		t.Errorf("Expected ClientID 'cid', got '%s'", client.ClientID)
	}
}

func TestNewClientWithRefreshEmptyAPIKey(t *testing.T) {
	_, err := NewClientWithRefresh("", "token1", "refresh1", "cid", "csecret")
	if err == nil {
		t.Error("Expected error for empty API key, got nil")
	}
}

func TestNewEnvoyClient(t *testing.T) {
	client, err := NewEnvoyClient("192.168.1.100", "jwt-token")
	if err != nil {
		t.Fatalf("NewEnvoyClient failed: %v", err)
	}

	if client.EnvoyIP != "192.168.1.100" {
		t.Errorf("Expected EnvoyIP '192.168.1.100', got '%s'", client.EnvoyIP)
	}
}

func TestNewEnvoyClientEmptyIP(t *testing.T) {
	_, err := NewEnvoyClient("", "token")
	if err == nil {
		t.Error("Expected error for empty envoy IP, got nil")
	}
}

func TestCloudGetSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok123" {
			t.Errorf("Expected Bearer auth, got '%s'", r.Header.Get("Authorization"))
		}
		if r.Header.Get("key") != "apikey1" {
			t.Errorf("Expected key header 'apikey1', got '%s'", r.Header.Get("key"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"system_id": 123}`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("apikey1", "tok123")

	var result System
	err := client.cloudGet(server.URL+"/test", &result)
	if err != nil {
		t.Fatalf("cloudGet failed: %v", err)
	}
	if result.SystemID != 123 {
		t.Errorf("Expected SystemID 123, got %d", result.SystemID)
	}
}

func TestCloudGetErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	client, _ := NewClient("key", "token")
	var result System
	err := client.cloudGet(server.URL+"/test", &result)
	if err == nil {
		t.Error("Expected error for 401, got nil")
	}
}

func TestCloudGetInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	client, _ := NewClient("key", "token")
	var result System
	err := client.cloudGet(server.URL+"/test", &result)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestCloudGetConnectionRefused(t *testing.T) {
	client, _ := NewClient("key", "token")
	var result System
	err := client.cloudGet("http://localhost:1/test", &result)
	if err == nil {
		t.Error("Expected error for connection refused, got nil")
	}
}

func TestCloudGetBadURL(t *testing.T) {
	client, _ := NewClient("key", "token")
	var result System
	err := client.cloudGet("://bad", &result)
	if err == nil {
		t.Error("Expected error for bad URL, got nil")
	}
}

func TestEnvoyGetSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer envoy-jwt" {
			t.Errorf("Expected Bearer envoy-jwt, got '%s'", r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"production":[],"consumption":[]}`))
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "envoy-jwt")
	var result EnvoyProduction
	err := client.envoyGet(server.URL+"/production.json", &result)
	if err != nil {
		t.Fatalf("envoyGet failed: %v", err)
	}
}

func TestEnvoyGetNoToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Errorf("Expected no Authorization header, got '%s'", r.Header.Get("Authorization"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"production":[],"consumption":[]}`))
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "")
	var result EnvoyProduction
	err := client.envoyGet(server.URL+"/test", &result)
	if err != nil {
		t.Fatalf("envoyGet with no token failed: %v", err)
	}
}

func TestEnvoyGetErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`forbidden`))
	}))
	defer server.Close()

	client, _ := NewEnvoyClient("localhost", "")
	var result EnvoyProduction
	err := client.envoyGet(server.URL+"/test", &result)
	if err == nil {
		t.Error("Expected error for 403, got nil")
	}
}

func TestEnvoyGetBadURL(t *testing.T) {
	client, _ := NewEnvoyClient("localhost", "")
	var result EnvoyProduction
	err := client.envoyGet("://bad", &result)
	if err == nil {
		t.Error("Expected error for bad URL, got nil")
	}
}

func TestPostFormSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Expected form content type, got '%s'", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token":"refreshed"}`))
	}))
	defer server.Close()

	client, _ := NewClient("key", "token")
	var result TokenInfo
	err := client.postForm(server.URL+"/test", "grant_type=refresh_token", &result)
	if err != nil {
		t.Fatalf("postForm failed: %v", err)
	}
	if result.AccessToken != "refreshed" {
		t.Errorf("Expected 'refreshed', got '%s'", result.AccessToken)
	}
}

func TestPostFormErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`unauthorized`))
	}))
	defer server.Close()

	client, _ := NewClient("key", "token")
	var result TokenInfo
	err := client.postForm(server.URL+"/test", "data=val", &result)
	if err == nil {
		t.Error("Expected error for 401, got nil")
	}
}

func TestPostFormNilResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := NewClient("key", "token")
	err := client.postForm(server.URL+"/test", "data=val", nil)
	if err != nil {
		t.Fatalf("postForm with nil response failed: %v", err)
	}
}

func TestPostFormBadURL(t *testing.T) {
	client, _ := NewClient("key", "token")
	err := client.postForm("://bad", "data=val", nil)
	if err == nil {
		t.Error("Expected error for bad URL, got nil")
	}
}

func TestAddQueryParams(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com/test", nil)
	addQueryParams(req, map[string]string{"start_date": "2024-01-01", "end_date": "2024-12-31"})

	query := req.URL.Query()
	if query.Get("start_date") != "2024-01-01" {
		t.Errorf("Expected start_date '2024-01-01', got '%s'", query.Get("start_date"))
	}
	if query.Get("end_date") != "2024-12-31" {
		t.Errorf("Expected end_date '2024-12-31', got '%s'", query.Get("end_date"))
	}
}

func TestAddQueryParamsEmpty(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com/test", nil)
	addQueryParams(req, map[string]string{})

	if req.URL.RawQuery != "" {
		t.Errorf("Expected no query params, got '%s'", req.URL.RawQuery)
	}
}
