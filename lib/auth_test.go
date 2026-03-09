package lib

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRefreshAccessTokenSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token":"new-at","refresh_token":"new-rt","token_type":"Bearer","expires_in":3600}`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClientWithRefresh("key", "old-at", "old-rt", "cid", "csecret")
	token, err := client.RefreshAccessToken()
	if err != nil {
		t.Fatalf("RefreshAccessToken failed: %v", err)
	}

	if token.AccessToken != "new-at" {
		t.Errorf("Expected access token 'new-at', got '%s'", token.AccessToken)
	}
	if client.AccessToken != "new-at" {
		t.Errorf("Expected client access token updated to 'new-at', got '%s'", client.AccessToken)
	}
	if client.RefreshToken != "new-rt" {
		t.Errorf("Expected client refresh token updated to 'new-rt', got '%s'", client.RefreshToken)
	}
}

func TestRefreshAccessTokenNoRefreshToken(t *testing.T) {
	client, _ := NewClient("key", "token")
	_, err := client.RefreshAccessToken()
	if err == nil {
		t.Error("Expected error for missing refresh token, got nil")
	}
}

func TestRefreshAccessTokenNoClientCredentials(t *testing.T) {
	client, _ := NewClient("key", "token")
	client.RefreshToken = "rt"
	_, err := client.RefreshAccessToken()
	if err == nil {
		t.Error("Expected error for missing client credentials, got nil")
	}
}

func TestRefreshAccessTokenServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClientWithRefresh("key", "at", "rt", "cid", "cs")
	_, err := client.RefreshAccessToken()
	if err == nil {
		t.Error("Expected error for server error, got nil")
	}
}

func TestExchangeAuthCodeSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token":"at","refresh_token":"rt","token_type":"Bearer","expires_in":3600}`))
	}))
	defer server.Close()

	originalURL := CloudBaseURL
	CloudBaseURL = server.URL
	defer func() { CloudBaseURL = originalURL }()

	client, _ := NewClient("key", "dummy")
	client.ClientID = "cid"
	client.ClientSecret = "cs"
	token, err := client.ExchangeAuthCode("auth-code", "https://redirect.example.com")
	if err != nil {
		t.Fatalf("ExchangeAuthCode failed: %v", err)
	}

	if token.AccessToken != "at" {
		t.Errorf("Expected 'at', got '%s'", token.AccessToken)
	}
}

func TestExchangeAuthCodeEmptyCode(t *testing.T) {
	client, _ := NewClient("key", "token")
	_, err := client.ExchangeAuthCode("", "https://redirect.example.com")
	if err == nil {
		t.Error("Expected error for empty code, got nil")
	}
}

func TestExchangeAuthCodeNoClientCredentials(t *testing.T) {
	client, _ := NewClient("key", "token")
	_, err := client.ExchangeAuthCode("code", "https://redirect.example.com")
	if err == nil {
		t.Error("Expected error for missing client credentials, got nil")
	}
}

func TestGetEnvoyTokenMissingEmail(t *testing.T) {
	client, _ := NewClient("key", "token")
	_, err := client.GetEnvoyToken("", "password", "serial")
	if err == nil {
		t.Error("Expected error for empty email, got nil")
	}
}

func TestGetEnvoyTokenMissingPassword(t *testing.T) {
	client, _ := NewClient("key", "token")
	_, err := client.GetEnvoyToken("email@example.com", "", "serial")
	if err == nil {
		t.Error("Expected error for empty password, got nil")
	}
}

func TestGetEnvoyTokenMissingSerial(t *testing.T) {
	client, _ := NewClient("key", "token")
	_, err := client.GetEnvoyToken("email@example.com", "password", "")
	if err == nil {
		t.Error("Expected error for empty serial, got nil")
	}
}

func TestGetEnvoyTokenLoginFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"bad credentials"}`))
	}))
	defer server.Close()

	originalURL := EnlightenURL
	EnlightenURL = server.URL
	defer func() { EnlightenURL = originalURL }()

	client, _ := NewClient("key", "token")
	_, err := client.GetEnvoyToken("email@example.com", "password", "serial123")
	if err == nil {
		t.Error("Expected error for login failure, got nil")
	}
}
