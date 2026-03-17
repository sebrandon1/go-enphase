package lib

import "context"

// GetProductionMeterReadings returns production meter readings for a system.
func (c *Client) GetProductionMeterReadings(systemID string) ([]MeterReading, error) {
	return c.GetProductionMeterReadingsCtx(context.Background(), systemID)
}

// GetProductionMeterReadingsCtx returns production meter readings for a system,
// respecting the provided context.
func (c *Client) GetProductionMeterReadingsCtx(ctx context.Context, systemID string) ([]MeterReading, error) {
	var result meterReadingsResponse
	err := c.cloudGetCtx(ctx, CloudBaseURL+"/api/v4/systems/"+systemID+"/production_meter_readings", &result)
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
	return c.GetEnergyLifetimeCtx(context.Background(), systemID, startDate, endDate)
}

// GetEnergyLifetimeCtx returns lifetime energy production data, respecting the
// provided context.
func (c *Client) GetEnergyLifetimeCtx(ctx context.Context, systemID, startDate, endDate string) (*EnergyLifetime, error) {
	var result EnergyLifetime
	err := c.cloudGetWithParamsCtx(
		ctx,
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
	return c.GetConsumptionLifetimeCtx(context.Background(), systemID, startDate, endDate)
}

// GetConsumptionLifetimeCtx returns lifetime energy consumption data, respecting the
// provided context.
func (c *Client) GetConsumptionLifetimeCtx(ctx context.Context, systemID, startDate, endDate string) (*ConsumptionLifetime, error) {
	var result ConsumptionLifetime
	err := c.cloudGetWithParamsCtx(
		ctx,
		CloudBaseURL+"/api/v4/systems/"+systemID+"/consumption_lifetime",
		dateParams(startDate, endDate),
		&result,
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
