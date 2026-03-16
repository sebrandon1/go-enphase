package lib

// System represents an Enphase solar system.
type System struct {
	SystemID    int    `json:"system_id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Modules     int    `json:"modules"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	TimeZone    string `json:"timezone"`
	PublicName  string `json:"public_name"`
	ReferenceID string `json:"reference"`
}

// systemsResponse wraps the API response for listing systems.
type systemsResponse struct {
	Systems []System `json:"systems"`
}

// SystemSummary represents a system's production summary.
type SystemSummary struct {
	SystemID       int    `json:"system_id"`
	Modules        int    `json:"modules"`
	Status         string `json:"status"`
	CurrentPower   int    `json:"current_power"`
	EnergyToday    int    `json:"energy_today"`
	EnergyLifetime int    `json:"energy_lifetime"`
	LastReportAt   int64  `json:"last_report_at"`
	SizeW          int    `json:"size_w"`
	SummaryDate    string `json:"summary_date"`
}

// Device represents a device (microinverter, battery, etc.) in a system.
type Device struct {
	ID           int    `json:"id"`
	SerialNumber string `json:"serial_number"`
	Model        string `json:"model"`
	PartNumber   string `json:"part_number"`
	Status       string `json:"status"`
	LastReportAt int64  `json:"last_report_at"`
}

// devicesResponse wraps the API response for listing devices.
type devicesResponse struct {
	Micro []Device `json:"micro_inverters"`
}

// MeterReading represents a production meter reading.
type MeterReading struct {
	SerialNumber string `json:"serial_number"`
	Value        int    `json:"value"`
	ReadAt       int64  `json:"read_at"`
}

// meterReadingsResponse wraps the API response for production meter readings.
type meterReadingsResponse struct {
	MeterReadings []MeterReading `json:"meter_readings"`
}

// EnergyLifetime represents lifetime energy production data.
type EnergyLifetime struct {
	StartDate  string `json:"start_date"`
	SystemID   int    `json:"system_id"`
	Production []int  `json:"production"`
}

// ConsumptionLifetime represents lifetime energy consumption data.
type ConsumptionLifetime struct {
	StartDate   string `json:"start_date"`
	SystemID    int    `json:"system_id"`
	Consumption []int  `json:"consumption"`
}

// BatteryStatus represents battery storage status for a system.
type BatteryStatus struct {
	Status                 string  `json:"status"`
	BatteryCount           int     `json:"battery_count"`
	EnergyStoredLifetime   int     `json:"energy_stored_lifetime"`
	EnergyConsumedLifetime int     `json:"energy_consumed_lifetime"`
	ChargePercent          float64 `json:"charge_percent"`
}

// TokenInfo represents OAuth2 token information.
type TokenInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// EnlightenSession represents the response from Enlighten login.
type EnlightenSession struct {
	SessionID string `json:"session_id"`
	UserID    int    `json:"user_id"`
	UserName  string `json:"user_name"`
}

// EnvoyProduction represents local Envoy production data.
type EnvoyProduction struct {
	Production  []EnvoyProductionEntry  `json:"production"`
	Consumption []EnvoyConsumptionEntry `json:"consumption"`
}

// EnvoyProductionEntry represents a single production entry from the Envoy.
type EnvoyProductionEntry struct {
	Type             string  `json:"type"`
	ActiveCount      int     `json:"activeCount"`
	ReadingTime      int64   `json:"readingTime"`
	WNow             float64 `json:"wNow"`
	WhLifetime       float64 `json:"whLifetime"`
	WhLastSevenDays  float64 `json:"whLastSevenDays"`
	WhToday          float64 `json:"whToday"`
	VahLifetime      float64 `json:"vahLifetime"`
	RmsCurrent       float64 `json:"rmsCurrent"`
	RmsVoltage       float64 `json:"rmsVoltage"`
	ReactPower       float64 `json:"reactPwr"`
	ApprntPower      float64 `json:"apprntPwr"`
	PwrFactor        float64 `json:"pwrFactor"`
	WhLastSevenDaysR float64 `json:"whLastSevenDaysR,omitempty"`
	WhTodayR         float64 `json:"whTodayR,omitempty"`
	WhLifetimeR      float64 `json:"whLifetimeR,omitempty"`
}

// EnvoyConsumptionEntry represents a single consumption entry from the Envoy.
type EnvoyConsumptionEntry struct {
	Type             string  `json:"type"`
	MeasurementType  string  `json:"measurementType"`
	ActiveCount      int     `json:"activeCount"`
	ReadingTime      int64   `json:"readingTime"`
	WNow             float64 `json:"wNow"`
	WhLifetime       float64 `json:"whLifetime"`
	WhLastSevenDays  float64 `json:"whLastSevenDays"`
	WhToday          float64 `json:"whToday"`
	VahLifetime      float64 `json:"vahLifetime"`
	RmsCurrent       float64 `json:"rmsCurrent"`
	RmsVoltage       float64 `json:"rmsVoltage"`
	ReactPower       float64 `json:"reactPwr"`
	ApprntPower      float64 `json:"apprntPwr"`
	PwrFactor        float64 `json:"pwrFactor"`
	WhLastSevenDaysR float64 `json:"whLastSevenDaysR,omitempty"`
	WhTodayR         float64 `json:"whTodayR,omitempty"`
	WhLifetimeR      float64 `json:"whLifetimeR,omitempty"`
}

// SensorReading represents a sensor reading from the local Envoy.
type SensorReading struct {
	MeasurementType string  `json:"measurementType"`
	ActivePower     float64 `json:"activePower"`
	RmsVoltage      float64 `json:"rmsVoltage"`
	RmsCurrent      float64 `json:"rmsCurrent"`
	PowerFactor     float64 `json:"pwrFactor"`
	Frequency       float64 `json:"frequency"`
}

// sensorReadingsResponse wraps the Envoy sensor readings response.
type sensorReadingsResponse struct {
	Sensors []struct {
		Readings []SensorReading `json:"readings"`
	} `json:"sensors"`
}

// EnvoySimpleProduction represents the basic production summary from /api/v1/production.
type EnvoySimpleProduction struct {
	WattsNow        float64 `json:"wattsNow"`
	WhToday         float64 `json:"whToday"`
	WhLastSevenDays float64 `json:"whLastSevenDays"`
	WhLifetime      float64 `json:"whLifetime"`
}

// InverterReading represents a single inverter's latest report from
// /api/v1/production/inverters.
type InverterReading struct {
	SerialNumber    string `json:"serialNumber"`
	LastReportDate  int64  `json:"lastReportDate"`
	LastReportWatts int    `json:"lastReportWatts"`
	MaxReportWatts  int    `json:"maxReportWatts"`
	DevType         int    `json:"devType"`
}

// MeterConfig represents the configuration of a revenue-grade meter from /ivp/meters.
type MeterConfig struct {
	EID             int    `json:"eid"`
	State           string `json:"state"`
	MeasurementType string `json:"measurementType"`
	PhaseMode       string `json:"phaseMode"`
	PhaseCount      int    `json:"phaseCount"`
}

// MeterData represents a meter reading snapshot from /ivp/meters/readings.
type MeterData struct {
	EID        int     `json:"eid"`
	Timestamp  int64   `json:"timestamp"`
	ActPower   float64 `json:"activePower"`
	ApprntPwr  float64 `json:"apparentPower"`
	ReactPwr   float64 `json:"reactivePower"`
	WhDlvdCum  float64 `json:"whDlvdCum"`
	WhRcvdCum  float64 `json:"whRcvdCum"`
	RmsCurrent float64 `json:"rmsCurrent"`
	RmsVoltage float64 `json:"rmsVoltage"`
	PwrFactor  float64 `json:"pwrFactor"`
	Frequency  float64 `json:"frequency"`
}

// StreamMeterEvent represents a single real-time meter event from the SSE stream at
// /stream/meter.
type StreamMeterEvent struct {
	EID       int     `json:"eid"`
	Timestamp int64   `json:"timestamp"`
	ActPower  float64 `json:"activePower"`
}
