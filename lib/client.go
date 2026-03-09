package lib

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

var (
	CloudBaseURL = "https://api.enphaseenergy.com"
	EnlightenURL = "https://enlighten.enphaseenergy.com"
	EntrezURL    = "https://entrez.enphaseenergy.com"
)

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
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

func newInsecureHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
	}
}

// NewClient creates a new Enphase API client for cloud access.
func NewClient(apiKey, accessToken string) (*Client, error) {
	if apiKey == "" || accessToken == "" {
		return nil, fmt.Errorf("api key and access token are required")
	}

	return &Client{
		APIKey:      apiKey,
		AccessToken: accessToken,
		HTTPClient:  newHTTPClient(),
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
		HTTPClient:   newHTTPClient(),
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
		HTTPClient: newInsecureHTTPClient(),
	}, nil
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

func (c *Client) cloudGet(url string, v any) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	c.setCloudHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return decodeJSON(resp.Body, v)
}

func (c *Client) envoyGet(url string, v any) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	c.setEnvoyHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	return decodeJSON(resp.Body, v)
}

func (c *Client) post(url string, body any, v any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	if v != nil {
		return decodeJSON(resp.Body, v)
	}

	return nil
}

func (c *Client) postForm(url string, formData string, v any) error {
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(formData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	if v != nil {
		return decodeJSON(resp.Body, v)
	}

	return nil
}

func decodeJSON(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

func addQueryParams(req *http.Request, params map[string]string) {
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
}
