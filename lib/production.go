package lib

// GetProductionMeterReadings returns production meter readings for a system.
func (c *Client) GetProductionMeterReadings(systemID string) ([]MeterReading, error) {
	var result meterReadingsResponse
	err := c.cloudGet(CloudBaseURL+"/api/v4/systems/"+systemID+"/production_meter_readings", &result)
	if err != nil {
		return nil, err
	}
	return result.MeterReadings, nil
}

func dateParams(startDate, endDate string) map[string]string {
	params := map[string]string{}
	if startDate != "" {
		params["start_date"] = startDate
	}
	if endDate != "" {
		params["end_date"] = endDate
	}
	return params
}

// GetEnergyLifetime returns lifetime energy production data.
func (c *Client) GetEnergyLifetime(systemID, startDate, endDate string) (*EnergyLifetime, error) {
	var result EnergyLifetime
	err := c.cloudGetWithParams(
		CloudBaseURL+"/api/v4/systems/"+systemID+"/energy_lifetime",
		dateParams(startDate, endDate),
		&result,
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetConsumptionLifetime returns lifetime energy consumption data.
func (c *Client) GetConsumptionLifetime(systemID, startDate, endDate string) (*ConsumptionLifetime, error) {
	var result ConsumptionLifetime
	err := c.cloudGetWithParams(
		CloudBaseURL+"/api/v4/systems/"+systemID+"/consumption_lifetime",
		dateParams(startDate, endDate),
		&result,
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
