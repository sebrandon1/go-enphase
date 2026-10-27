package lib

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sebrandon1/go-toolbox/httpx"
)

var (
	CloudBaseURL = "https://api.enphaseenergy.com"
	EnlightenURL = "https://enlighten.enphaseenergy.com"
	EntrezURL    = "https://entrez.enphaseenergy.com"
)

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithTimeout sets the HTTP client request timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.HTTPClient.Timeout = d
	}
}

// WithHTTPClient replaces the underlying HTTP client.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = hc
	}
}

// WithLogger attaches a structured logger to the client. When set, each outgoing
// request and its status code are logged at DEBUG level.
func WithLogger(l *slog.Logger) ClientOption {
	return func(c *Client) {
		c.logger = l
	}
}

// httpStatusError is the user-facing error for an HTTP failure status. It keeps
// the Enphase-specific message while wrapping the httpx typed error, so callers
// can still classify it with errors.As / httpx.Is*.
type httpStatusError struct {
	msg string
	err error
}

func (e *httpStatusError) Error() string { return e.msg }
func (e *httpStatusError) Unwrap() error { return e.err }

// friendlyHTTPError converts a raw HTTP error status into a human-readable message.
func friendlyHTTPError(statusCode int, body []byte) error {
	typed := httpx.CheckStatus(&http.Response{StatusCode: statusCode}, body)

	var msg string
	switch {
	case httpx.IsAuthError(typed):
		msg = "authentication failed (401): check your API key and access token"
	case httpx.IsValidation(typed):
		msg = fmt.Sprintf("invalid request parameters (422): %s", strings.TrimSpace(string(body)))
	case httpx.IsRateLimited(typed):
		msg = "rate limit exceeded (429): wait before retrying"
	case statusCode == http.StatusInternalServerError:
		msg = "enphase server error (500): try again later"
	default:
		msg = fmt.Sprintf("unexpected status code: %d, response: %s", statusCode, string(body))
	}
	return &httpStatusError{msg: msg, err: typed}
}

// WithInsecureSkipVerify controls TLS certificate verification on the underlying transport.
func WithInsecureSkipVerify(skip bool) ClientOption {
	return func(c *Client) {
		inner := innerTransport(c.HTTPClient.Transport)
		if inner == nil {
			return
		}
		if inner.TLSClientConfig == nil {
			inner.TLSClientConfig = &tls.Config{} //nolint:gosec
		}
		inner.TLSClientConfig.InsecureSkipVerify = skip //nolint:gosec
	}
}

// innerTransport extracts the *http.Transport from a (possibly wrapped) RoundTripper.
func innerTransport(rt http.RoundTripper) *http.Transport {
	if t, ok := rt.(*http.Transport); ok {
		return t
	}
	return nil
}

// Client provides access to Enphase cloud and local Envoy APIs.
type Client struct {
	APIKey       string
	AccessToken  string
	RefreshToken string
	ClientID     string
	ClientSecret string
	EnvoyIP      string
	EnvoyToken   string
	HTTPClient   *http.Client

	// OnTokenRefresh is called after a successful automatic token refresh with
	// the new access and refresh tokens, allowing callers to persist them.
	OnTokenRefresh func(accessToken, refreshToken string)

	logger           *slog.Logger
	envoyTokenExpiry time.Time
	envoyTokenMu     sync.Mutex
	refreshMu        sync.Mutex
}

func newHTTPClientWithTLS(insecure bool) *http.Client {
	opts := httpx.Options{
		Timeout:             30 * time.Second,
		DialTimeout:         10 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	if insecure {
		opts.TLSConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}
	return httpx.NewHTTPClient(opts)
}

// NewClient creates a new Enphase API client for cloud access.
func NewClient(apiKey, accessToken string) (*Client, error) {
	if apiKey == "" || accessToken == "" {
		return nil, fmt.Errorf("api key and access token are required")
	}

	return &Client{
		APIKey:      apiKey,
		AccessToken: accessToken,
		HTTPClient:  newHTTPClientWithTLS(false),
	}, nil
}

// NewClientWithRefresh creates a client with refresh token support.
func NewClientWithRefresh(apiKey, accessToken, refreshToken, clientID, clientSecret string) (*Client, error) {
	if apiKey == "" || accessToken == "" {
		return nil, fmt.Errorf("api key and access token are required")
	}

	return &Client{
		APIKey:       apiKey,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		HTTPClient:   newHTTPClientWithTLS(false),
	}, nil
}

// NewEnvoyClient creates a client for local Envoy access.
func NewEnvoyClient(envoyIP, envoyToken string) (*Client, error) {
	if envoyIP == "" {
		return nil, fmt.Errorf("envoy IP is required")
	}

	return &Client{
		EnvoyIP:    envoyIP,
		EnvoyToken: envoyToken,
		HTTPClient: newHTTPClientWithTLS(true),
	}, nil
}

// NewClientWithOptions creates a cloud API client with functional options applied.
func NewClientWithOptions(apiKey, accessToken string, opts ...ClientOption) (*Client, error) {
	if apiKey == "" || accessToken == "" {
		return nil, fmt.Errorf("api key and access token are required")
	}

	c := &Client{
		APIKey:      apiKey,
		AccessToken: accessToken,
		HTTPClient:  newHTTPClientWithTLS(false),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

func (c *Client) setCloudHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")
}

func (c *Client) setEnvoyHeaders(req *http.Request) {
	if c.EnvoyToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.EnvoyToken)
	}
	req.Header.Set("Content-Type", "application/json")
}

// drainAndClose drains and closes a response body, enabling HTTP keep-alive reuse.
func drainAndClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	body.Close()
}

// tryRefreshToken attempts to refresh the access token when a 401 is received.
// oldToken is the token that was used for the failed request; if another goroutine
// has already refreshed (c.AccessToken != oldToken), the refresh is skipped.
// Returns true if the client now has a fresh token (either just refreshed or
// already refreshed by a concurrent goroutine).
// TryLock is used instead of Lock to prevent deadlock when called re-entrantly
// through RefreshAccessToken → postFormWithAuthCtx → tryRefreshToken.
func (c *Client) tryRefreshToken(oldToken string) bool {
	if c.RefreshToken == "" || c.ClientID == "" || c.ClientSecret == "" {
		return false
	}
	if !c.refreshMu.TryLock() {
		return false // refresh already in progress (re-entrant or concurrent)
	}
	defer c.refreshMu.Unlock()
	if c.AccessToken != oldToken {
		return true // already refreshed by another goroutine
	}
	token, err := c.RefreshAccessToken()
	if err != nil {
		return false
	}
	if c.OnTokenRefresh != nil {
		c.OnTokenRefresh(token.AccessToken, token.RefreshToken)
	}
	return true
}

func (c *Client) cloudGetCtx(ctx context.Context, url string, v any) error {
	return c.cloudGetWithParamsCtx(ctx, url, nil, v)
}

func (c *Client) envoyGetCtx(ctx context.Context, url string, v any) error {
	req, err := httpx.NewRequest(ctx, http.MethodGet, url)
	if err != nil {
		return err
	}
	c.setEnvoyHeaders(req)

	resp, err := c.doGet(ctx, req)
	if err != nil {
		return err
	}
	defer drainAndClose(resp.Body)

	if c.logger != nil {
		c.logger.Debug("http request", "method", req.Method, "url", req.URL.String(), "status", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return friendlyHTTPError(resp.StatusCode, []byte(httpx.ReadErrorBody(resp.Body, httpx.DefaultMaxErrorBodyBytes)))
	}
	return httpx.DecodeJSON(resp.Body, v)
}

func (c *Client) cloudGetWithParamsCtx(ctx context.Context, url string, params map[string]string, v any) error {
	buildReq := func() (*http.Request, error) {
		req, err := httpx.NewRequest(ctx, http.MethodGet, url)
		if err != nil {
			return nil, err
		}
		c.setCloudHeaders(req)
		httpx.AddQueryParams(req, params)
		return req, nil
	}

	oldToken := c.AccessToken
	req, err := buildReq()
	if err != nil {
		return err
	}
	resp, err := c.doGet(ctx, req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnauthorized && c.tryRefreshToken(oldToken) {
		drainAndClose(resp.Body)
		req, err = buildReq()
		if err != nil {
			return err
		}
		resp, err = c.doGet(ctx, req)
		if err != nil {
			return err
		}
	}

	defer drainAndClose(resp.Body)

	if c.logger != nil {
		c.logger.Debug("http request", "method", req.Method, "url", req.URL.String(), "status", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return friendlyHTTPError(resp.StatusCode, []byte(httpx.ReadErrorBody(resp.Body, httpx.DefaultMaxErrorBodyBytes)))
	}
	return httpx.DecodeJSON(resp.Body, v)
}

func (c *Client) postFormWithAuthCtx(ctx context.Context, url, formData, username, password string, v any) error {
	buildReq := func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBufferString(formData))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if username != "" {
			req.SetBasicAuth(username, password)
		}
		return req, nil
	}

	oldToken := c.AccessToken
	req, err := buildReq()
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnauthorized && c.tryRefreshToken(oldToken) {
		drainAndClose(resp.Body)
		req, err = buildReq()
		if err != nil {
			return err
		}
		resp, err = c.HTTPClient.Do(req)
		if err != nil {
			return err
		}
	}

	defer drainAndClose(resp.Body)

	if c.logger != nil {
		c.logger.Debug("http request", "method", req.Method, "url", req.URL.String(), "status", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return friendlyHTTPError(resp.StatusCode, []byte(httpx.ReadErrorBody(resp.Body, httpx.DefaultMaxErrorBodyBytes)))
	}

	if v != nil {
		return httpx.DecodeJSON(resp.Body, v)
	}
	return nil
}

// EnsureEnvoyToken ensures a valid Envoy JWT is set on the client, refreshing it if
// it is missing or within 5 minutes of expiry. It is safe for concurrent use.
func (c *Client) EnsureEnvoyToken(ctx context.Context, email, password, serial string) error {
	c.envoyTokenMu.Lock()
	defer c.envoyTokenMu.Unlock()

	if c.EnvoyToken != "" && time.Now().Before(c.envoyTokenExpiry.Add(-5*time.Minute)) {
		return nil
	}

	token, err := c.GetEnvoyTokenCtx(ctx, email, password, serial)
	if err != nil {
		return err
	}

	expiry, err := parseJWTExpiry(token)
	if err != nil {
		expiry = time.Now().Add(time.Hour)
	}

	c.EnvoyToken = token
	c.envoyTokenExpiry = expiry
	return nil
}

// parseJWTExpiry extracts the exp claim from a JWT without external dependencies.
func parseJWTExpiry(token string) (time.Time, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid JWT format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return time.Time{}, fmt.Errorf("failed to parse JWT claims: %w", err)
	}
	if claims.Exp == 0 {
		return time.Time{}, fmt.Errorf("JWT has no exp claim")
	}

	return time.Unix(claims.Exp, 0), nil
}
