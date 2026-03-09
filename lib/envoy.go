package lib

import "fmt"

// GetEnvoyProduction returns production and consumption data from the local Envoy.
func (c *Client) GetEnvoyProduction() (*EnvoyProduction, error) {
	if c.EnvoyIP == "" {
		return nil, fmt.Errorf("envoy IP is required")
	}

	var production EnvoyProduction
	err := c.envoyGet("https://"+c.EnvoyIP+"/production.json", &production)
	if err != nil {
		return nil, err
	}
	return &production, nil
}

// GetEnvoySensors returns sensor readings from the local Envoy.
func (c *Client) GetEnvoySensors() ([]SensorReading, error) {
	if c.EnvoyIP == "" {
		return nil, fmt.Errorf("envoy IP is required")
	}

	var result sensorReadingsResponse
	err := c.envoyGet("https://"+c.EnvoyIP+"/ivp/sensors/readings_object", &result)
	if err != nil {
		return nil, err
	}

	var readings []SensorReading
	for _, sensor := range result.Sensors {
		readings = append(readings, sensor.Readings...)
	}
	return readings, nil
}
