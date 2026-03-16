package lib

import "context"

// GetBatteryStatus returns battery storage status for a system.
func (c *Client) GetBatteryStatus(systemID string) (*BatteryStatus, error) {
	return c.GetBatteryStatusCtx(context.Background(), systemID)
}

// GetBatteryStatusCtx returns battery storage status for a system, respecting the
// provided context.
func (c *Client) GetBatteryStatusCtx(ctx context.Context, systemID string) (*BatteryStatus, error) {
	var status BatteryStatus
	err := c.cloudGetCtx(ctx, CloudBaseURL+"/api/v4/systems/"+systemID+"/battery_lifetime", &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}
