package lib

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// RefreshAccessToken uses the refresh token to obtain a new access token.
func (c *Client) RefreshAccessToken() (*TokenInfo, error) {
	return c.RefreshAccessTokenCtx(context.Background())
}

// RefreshAccessTokenCtx uses the refresh token to obtain a new access token,
// respecting the provided context.
func (c *Client) RefreshAccessTokenCtx(ctx context.Context) (*TokenInfo, error) {
	if c.RefreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}
	if c.ClientID == "" || c.ClientSecret == "" {
		return nil, fmt.Errorf("client ID and client secret are required for token refresh")
	}

	formData := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {c.RefreshToken},
	}.Encode()

	var token TokenInfo
	err := c.postFormWithAuthCtx(ctx, CloudBaseURL+"/oauth/token", formData, c.ClientID, c.ClientSecret, &token)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	c.AccessToken = token.AccessToken
	if token.RefreshToken != "" {
		c.RefreshToken = token.RefreshToken
	}

	return &token, nil
}

// ExchangeAuthCode exchanges an authorization code for tokens.
func (c *Client) ExchangeAuthCode(code, redirectURI string) (*TokenInfo, error) {
	return c.ExchangeAuthCodeCtx(context.Background(), code, redirectURI)
}

// ExchangeAuthCodeCtx exchanges an authorization code for tokens, respecting the
// provided context.
func (c *Client) ExchangeAuthCodeCtx(ctx context.Context, code, redirectURI string) (*TokenInfo, error) {
	if code == "" {
		return nil, fmt.Errorf("authorization code is required")
	}
	if c.ClientID == "" || c.ClientSecret == "" {
		return nil, fmt.Errorf("client ID and client secret are required")
	}

	formData := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {redirectURI},
	}.Encode()

	var token TokenInfo
	err := c.postFormWithAuthCtx(ctx, CloudBaseURL+"/oauth/token", formData, c.ClientID, c.ClientSecret, &token)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange auth code: %w", err)
	}

	c.AccessToken = token.AccessToken
	if token.RefreshToken != "" {
		c.RefreshToken = token.RefreshToken
	}

	return &token, nil
}

// GetEnvoyToken obtains a local Envoy JWT token via the Enlighten login flow.
func (c *Client) GetEnvoyToken(email, password, envoySerial string) (string, error) {
	return c.GetEnvoyTokenCtx(context.Background(), email, password, envoySerial)
}

// GetEnvoyTokenCtx obtains a local Envoy JWT token via the Enlighten login flow,
// respecting the provided context.
func (c *Client) GetEnvoyTokenCtx(ctx context.Context, email, password, envoySerial string) (string, error) {
	if email == "" || password == "" {
		return "", fmt.Errorf("email and password are required")
	}
	if envoySerial == "" {
		return "", fmt.Errorf("envoy serial number is required")
	}

	// Step 1: Login to Enlighten
	loginData := url.Values{
		"user[email]":    {email},
		"user[password]": {password},
	}.Encode()

	var session EnlightenSession
	err := c.postFormWithAuthCtx(ctx, EnlightenURL+"/login/login.json", loginData, "", "", &session)
	if err != nil {
		return "", fmt.Errorf("enlighten login failed: %w", err)
	}

	// Step 2: Get Envoy token from Entrez — returns a bare JWT string, not JSON.
	tokenData := url.Values{
		"session_id": {session.SessionID},
		"serial_num": {envoySerial},
		"username":   {email},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, EntrezURL+"/tokens", bytes.NewBufferString(tokenData))
	if err != nil {
		return "", fmt.Errorf("failed to build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get envoy token: %w", err)
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get envoy token: status %d, response: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read envoy token: %w", err)
	}

	return string(body), nil
}
