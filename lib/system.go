package lib

import (
	"fmt"
	"net/http"
)

// ListSystems returns all systems accessible to the authenticated user.
func (c *Client) ListSystems() ([]System, error) {
	url := CloudBaseURL + "/api/v4/systems"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.setCloudHeaders(req)

	var result systemsResponse
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, err
	}

	return result.Systems, nil
}

// GetSystem returns details for a specific system.
func (c *Client) GetSystem(systemID string) (*System, error) {
	var system System
	err := c.cloudGet(CloudBaseURL+"/api/v4/systems/"+systemID, &system)
	if err != nil {
		return nil, err
	}
	return &system, nil
}

// GetSystemSummary returns the production summary for a system.
func (c *Client) GetSystemSummary(systemID string) (*SystemSummary, error) {
	var summary SystemSummary
	err := c.cloudGet(CloudBaseURL+"/api/v4/systems/"+systemID+"/summary", &summary)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

// ListDevices returns devices for a system.
func (c *Client) ListDevices(systemID string) ([]Device, error) {
	url := CloudBaseURL + "/api/v4/systems/" + systemID + "/devices"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.setCloudHeaders(req)

	var result devicesResponse
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, err
	}

	return result.Micro, nil
}
