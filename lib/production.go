package lib

import (
	"fmt"
	"net/http"
)

// GetProductionMeterReadings returns production meter readings for a system.
func (c *Client) GetProductionMeterReadings(systemID string) ([]MeterReading, error) {
	url := CloudBaseURL + "/api/v4/systems/" + systemID + "/production_meter_readings"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.setCloudHeaders(req)

	var result meterReadingsResponse
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

	return result.MeterReadings, nil
}

// GetEnergyLifetime returns lifetime energy production data.
func (c *Client) GetEnergyLifetime(systemID, startDate, endDate string) (*EnergyLifetime, error) {
	url := CloudBaseURL + "/api/v4/systems/" + systemID + "/energy_lifetime"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.setCloudHeaders(req)

	params := map[string]string{}
	if startDate != "" {
		params["start_date"] = startDate
	}
	if endDate != "" {
		params["end_date"] = endDate
	}
	if len(params) > 0 {
		addQueryParams(req, params)
	}

	var result EnergyLifetime
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

	return &result, nil
}

// GetConsumptionLifetime returns lifetime energy consumption data.
func (c *Client) GetConsumptionLifetime(systemID, startDate, endDate string) (*ConsumptionLifetime, error) {
	url := CloudBaseURL + "/api/v4/systems/" + systemID + "/consumption_lifetime"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.setCloudHeaders(req)

	params := map[string]string{}
	if startDate != "" {
		params["start_date"] = startDate
	}
	if endDate != "" {
		params["end_date"] = endDate
	}
	if len(params) > 0 {
		addQueryParams(req, params)
	}

	var result ConsumptionLifetime
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

	return &result, nil
}
