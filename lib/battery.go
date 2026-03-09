package lib

// GetBatteryStatus returns battery storage status for a system.
func (c *Client) GetBatteryStatus(systemID string) (*BatteryStatus, error) {
	var status BatteryStatus
	err := c.cloudGet(CloudBaseURL+"/api/v4/systems/"+systemID+"/battery_lifetime", &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}
