package lib

import (
	"context"
	"fmt"
)

// GetEnvoyProduction returns production and consumption data from the local Envoy.
func (c *Client) GetEnvoyProduction() (*EnvoyProduction, error) {
	return c.GetEnvoyProductionCtx(context.Background())
}

// GetEnvoyProductionCtx returns production and consumption data from the local Envoy,
// respecting the provided context.
func (c *Client) GetEnvoyProductionCtx(ctx context.Context) (*EnvoyProduction, error) {
	if c.EnvoyIP == "" {
		return nil, fmt.Errorf("envoy IP is required")
	}

	var production EnvoyProduction
	err := c.envoyGetCtx(ctx, "https://"+c.EnvoyIP+"/production.json", &production)
	if err != nil {
		return nil, err
	}
	return &production, nil
}

// GetEnvoySensors returns sensor readings from the local Envoy.
func (c *Client) GetEnvoySensors() ([]SensorReading, error) {
	return c.GetEnvoySensorsCtx(context.Background())
}

// GetEnvoySensorsCtx returns sensor readings from the local Envoy, respecting the
// provided context.
func (c *Client) GetEnvoySensorsCtx(ctx context.Context) ([]SensorReading, error) {
	if c.EnvoyIP == "" {
		return nil, fmt.Errorf("envoy IP is required")
	}

	var result sensorReadingsResponse
	err := c.envoyGetCtx(ctx, "https://"+c.EnvoyIP+"/ivp/sensors/readings_object", &result)
	if err != nil {
		return nil, err
	}

	var readings []SensorReading
	for _, sensor := range result.Sensors {
		readings = append(readings, sensor.Readings...)
	}
	return readings, nil
}

// GetEnvoySimpleProduction returns the basic production summary from the local Envoy
// at /api/v1/production.
func (c *Client) GetEnvoySimpleProduction() (*EnvoySimpleProduction, error) {
	return c.GetEnvoySimpleProductionCtx(context.Background())
}

// GetEnvoySimpleProductionCtx returns the basic production summary from the local Envoy,
// respecting the provided context.
func (c *Client) GetEnvoySimpleProductionCtx(ctx context.Context) (*EnvoySimpleProduction, error) {
	if c.EnvoyIP == "" {
		return nil, fmt.Errorf("envoy IP is required")
	}

	var result EnvoySimpleProduction
	err := c.envoyGetCtx(ctx, "https://"+c.EnvoyIP+"/api/v1/production", &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetInverterReadings returns per-inverter production data from the local Envoy at
// /api/v1/production/inverters.
func (c *Client) GetInverterReadings() ([]InverterReading, error) {
	return c.GetInverterReadingsCtx(context.Background())
}

// GetInverterReadingsCtx returns per-inverter production data from the local Envoy,
// respecting the provided context.
func (c *Client) GetInverterReadingsCtx(ctx context.Context) ([]InverterReading, error) {
	if c.EnvoyIP == "" {
		return nil, fmt.Errorf("envoy IP is required")
	}

	var result []InverterReading
	err := c.envoyGetCtx(ctx, "https://"+c.EnvoyIP+"/api/v1/production/inverters", &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetMeterConfig returns the configuration of all revenue-grade meters from the local
// Envoy at /ivp/meters.
func (c *Client) GetMeterConfig() ([]MeterConfig, error) {
	return c.GetMeterConfigCtx(context.Background())
}

// GetMeterConfigCtx returns meter configuration from the local Envoy, respecting the
// provided context.
func (c *Client) GetMeterConfigCtx(ctx context.Context) ([]MeterConfig, error) {
	if c.EnvoyIP == "" {
		return nil, fmt.Errorf("envoy IP is required")
	}

	var result []MeterConfig
	err := c.envoyGetCtx(ctx, "https://"+c.EnvoyIP+"/ivp/meters", &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetMeterReadings returns the latest meter readings from the local Envoy at
// /ivp/meters/readings.
func (c *Client) GetMeterReadings() ([]MeterData, error) {
	return c.GetMeterReadingsCtx(context.Background())
}

// GetMeterReadingsCtx returns the latest meter readings from the local Envoy,
// respecting the provided context.
func (c *Client) GetMeterReadingsCtx(ctx context.Context) ([]MeterData, error) {
	if c.EnvoyIP == "" {
		return nil, fmt.Errorf("envoy IP is required")
	}

	var result []MeterData
	err := c.envoyGetCtx(ctx, "https://"+c.EnvoyIP+"/ivp/meters/readings", &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
