package lib

// ListSystems returns all systems accessible to the authenticated user.
func (c *Client) ListSystems() ([]System, error) {
	var result systemsResponse
	err := c.cloudGet(CloudBaseURL+"/api/v4/systems", &result)
	if err != nil {
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
	var result devicesResponse
	err := c.cloudGet(CloudBaseURL+"/api/v4/systems/"+systemID+"/devices", &result)
	if err != nil {
		return nil, err
	}
	return result.Micro, nil
}
