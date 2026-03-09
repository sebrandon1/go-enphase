package lib

import (
	"fmt"
	"net/url"
)

// RefreshAccessToken uses the refresh token to obtain a new access token.
func (c *Client) RefreshAccessToken() (*TokenInfo, error) {
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
	err := c.postFormWithBasicAuth(CloudBaseURL+"/oauth/token", formData, c.ClientID, c.ClientSecret, &token)
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
	err := c.postFormWithBasicAuth(CloudBaseURL+"/oauth/token", formData, c.ClientID, c.ClientSecret, &token)
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
	err := c.postForm(EnlightenURL+"/login/login.json", loginData, &session)
	if err != nil {
		return "", fmt.Errorf("enlighten login failed: %w", err)
	}

	// Step 2: Get Envoy token from Entrez
	tokenData := url.Values{
		"session_id": {session.SessionID},
		"serial_num": {envoySerial},
		"username":   {email},
	}.Encode()

	var tokenResp string
	err = c.postForm(EntrezURL+"/tokens", tokenData, &tokenResp)
	if err != nil {
		return "", fmt.Errorf("failed to get envoy token: %w", err)
	}

	return tokenResp, nil
}
