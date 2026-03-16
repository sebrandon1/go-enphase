package lib

import "context"

// ListSystems returns all systems accessible to the authenticated user.
func (c *Client) ListSystems() ([]System, error) {
	return c.ListSystemsCtx(context.Background())
}

// ListSystemsCtx returns all systems accessible to the authenticated user,
// respecting the provided context.
func (c *Client) ListSystemsCtx(ctx context.Context) ([]System, error) {
	var result systemsResponse
	err := c.cloudGetCtx(ctx, CloudBaseURL+"/api/v4/systems", &result)
	if err != nil {
		return nil, err
	}
	return result.Systems, nil
}

// GetSystem returns details for a specific system.
func (c *Client) GetSystem(systemID string) (*System, error) {
	return c.GetSystemCtx(context.Background(), systemID)
}

// GetSystemCtx returns details for a specific system, respecting the provided context.
func (c *Client) GetSystemCtx(ctx context.Context, systemID string) (*System, error) {
	var system System
	err := c.cloudGetCtx(ctx, CloudBaseURL+"/api/v4/systems/"+systemID, &system)
	if err != nil {
		return nil, err
	}
	return &system, nil
}

// GetSystemSummary returns the production summary for a system.
func (c *Client) GetSystemSummary(systemID string) (*SystemSummary, error) {
	return c.GetSystemSummaryCtx(context.Background(), systemID)
}

// GetSystemSummaryCtx returns the production summary for a system, respecting the
// provided context.
func (c *Client) GetSystemSummaryCtx(ctx context.Context, systemID string) (*SystemSummary, error) {
	var summary SystemSummary
	err := c.cloudGetCtx(ctx, CloudBaseURL+"/api/v4/systems/"+systemID+"/summary", &summary)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

// ListDevices returns devices for a system.
func (c *Client) ListDevices(systemID string) ([]Device, error) {
	return c.ListDevicesCtx(context.Background(), systemID)
}

// ListDevicesCtx returns devices for a system, respecting the provided context.
func (c *Client) ListDevicesCtx(ctx context.Context, systemID string) ([]Device, error) {
	var result devicesResponse
	err := c.cloudGetCtx(ctx, CloudBaseURL+"/api/v4/systems/"+systemID+"/devices", &result)
	if err != nil {
		return nil, err
	}
	return result.Micro, nil
}
