package lib

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

// makeTestJWT builds a JWT-shaped string with the given exp Unix timestamp.
func makeTestJWT(exp int64) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`))
	payload, _ := json.Marshal(map[string]int64{"exp": exp})
	return fmt.Sprintf("%s.%s.sig", header, base64.RawURLEncoding.EncodeToString(payload))
}

func TestGetEnvoyTokenHappyPath(t *testing.T) {
	fakeJWT := makeTestJWT(9999999999)

	enlightenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"session_id":"sess123","user_id":1,"user_name":"test"}`))
	}))
	defer enlightenServer.Close()

	entrezServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fakeJWT))
	}))
	defer entrezServer.Close()

	origEnlighten := EnlightenURL
	origEntrez := EntrezURL
	EnlightenURL = enlightenServer.URL
	EntrezURL = entrezServer.URL
	defer func() {
		EnlightenURL = origEnlighten
		EntrezURL = origEntrez
	}()

	client, _ := NewClient("key", "token")
	got, err := client.GetEnvoyToken("user@example.com", "pass", "SERIAL123")
	if err != nil {
		t.Fatalf("GetEnvoyToken failed: %v", err)
	}
	if got != fakeJWT {
		t.Errorf("Expected JWT %q, got %q", fakeJWT, got)
	}
}

func TestEnsureEnvoyTokenCaching(t *testing.T) {
	fakeJWT := makeTestJWT(9999999999)
	entrezCalls := 0

	enlightenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"session_id":"sess","user_id":1,"user_name":"t"}`))
	}))
	defer enlightenServer.Close()

	entrezServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entrezCalls++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fakeJWT))
	}))
	defer entrezServer.Close()

	origEnlighten := EnlightenURL
	origEntrez := EntrezURL
	EnlightenURL = enlightenServer.URL
	EntrezURL = entrezServer.URL
	defer func() {
		EnlightenURL = origEnlighten
		EntrezURL = origEntrez
	}()

	client, _ := NewClient("key", "token")
	if err := client.EnsureEnvoyToken(context.Background(), "u@e.com", "p", "S1"); err != nil {
		t.Fatalf("first EnsureEnvoyToken failed: %v", err)
	}
	if err := client.EnsureEnvoyToken(context.Background(), "u@e.com", "p", "S1"); err != nil {
		t.Fatalf("second EnsureEnvoyToken failed: %v", err)
	}
	if entrezCalls != 1 {
		t.Errorf("Expected Entrez called once (cached), got %d", entrezCalls)
	}
}

func TestEnsureEnvoyTokenRefresh(t *testing.T) {
	fakeJWT := makeTestJWT(9999999999)
	entrezCalls := 0

	enlightenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"session_id":"sess","user_id":1,"user_name":"t"}`))
	}))
	defer enlightenServer.Close()

	entrezServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entrezCalls++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fakeJWT))
	}))
	defer entrezServer.Close()

	origEnlighten := EnlightenURL
	origEntrez := EntrezURL
	EnlightenURL = enlightenServer.URL
	EntrezURL = entrezServer.URL
	defer func() {
		EnlightenURL = origEnlighten
		EntrezURL = origEntrez
	}()

	client, _ := NewClient("key", "token")
	// Pre-set an expired token so EnsureEnvoyToken is forced to refresh.
	client.EnvoyToken = makeTestJWT(1) // exp=1 (far in the past)
	client.envoyTokenExpiry = time.Unix(1, 0)

	if err := client.EnsureEnvoyToken(context.Background(), "u@e.com", "p", "S1"); err != nil {
		t.Fatalf("EnsureEnvoyToken refresh failed: %v", err)
	}
	if entrezCalls != 1 {
		t.Errorf("Expected Entrez called once for refresh, got %d", entrezCalls)
	}
	if client.EnvoyToken != fakeJWT {
		t.Error("Expected token updated to new JWT")
	}
}

func TestParseJWTExpiry(t *testing.T) {
	jwt := makeTestJWT(1700000000)
	expiry, err := parseJWTExpiry(jwt)
	if err != nil {
		t.Fatalf("parseJWTExpiry failed: %v", err)
	}
	if expiry.Unix() != 1700000000 {
		t.Errorf("Expected exp 1700000000, got %d", expiry.Unix())
	}
}

func TestParseJWTExpiryInvalid(t *testing.T) {
	_, err := parseJWTExpiry("not.a.jwt.with.five.parts")
	if err == nil {
		t.Error("Expected error for invalid JWT, got nil")
	}
}
