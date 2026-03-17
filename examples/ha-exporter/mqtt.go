package main

import (
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const haDiscoveryTemplate = `{
  "name": "Solar %s",
  "state_topic": "%s/sensor/solar_%s/state",
  "unit_of_measurement": "%s",
  "value_template": "{{ value_json.%s }}",
  "unique_id": "enphase_solar_%s_%s",
  "device": {
    "identifiers": ["enphase_solar_%s"],
    "name": "Enphase Solar",
    "manufacturer": "Enphase Energy"
  }
}`

// MQTTPublisher publishes solar metrics to an MQTT broker for Home Assistant discovery.
type MQTTPublisher struct {
	client mqtt.Client
	prefix string
	serial string
}

// NewMQTTPublisher creates and connects an MQTTPublisher.
func NewMQTTPublisher(broker, username, password, prefix, serial string) (*MQTTPublisher, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID("go-enphase-exporter").
		SetUsername(username).
		SetPassword(password).
		SetAutoReconnect(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("mqtt connect: %w", token.Error())
	}

	return &MQTTPublisher{client: client, prefix: prefix, serial: serial}, nil
}

// PublishDiscovery sends Home Assistant MQTT auto-discovery config messages.
func (p *MQTTPublisher) PublishDiscovery() {
	sensors := []struct {
		name  string
		unit  string
		field string
	}{
		{"Current Power", "W", "current_power_w"},
		{"Energy Today", "Wh", "energy_today_wh"},
		{"Energy Lifetime", "Wh", "energy_lifetime_wh"},
		{"Net Power", "W", "net_power_w"},
	}

	for _, s := range sensors {
		topic := fmt.Sprintf("%s/sensor/solar_%s/config", p.prefix, p.serial)
		payload := fmt.Sprintf(haDiscoveryTemplate,
			s.name, p.prefix, p.serial, s.unit, s.field,
			p.serial, s.field, p.serial,
		)
		p.client.Publish(topic, 1, true, payload)
	}
}

// PublishState publishes the current metric values as a JSON state payload.
func (p *MQTTPublisher) PublishState(snap Snapshot) {
	state := map[string]float64{
		"current_power_w":    snap.CurrentPowerW,
		"energy_today_wh":    snap.EnergyTodayWh,
		"energy_lifetime_wh": snap.EnergyLifetimeWh,
		"net_power_w":        snap.NetPowerW,
	}

	payload, err := json.Marshal(state)
	if err != nil {
		Error("mqtt marshal error: %v", err)
		return
	}

	topic := fmt.Sprintf("%s/sensor/solar_%s/state", p.prefix, p.serial)
	p.client.Publish(topic, 0, false, payload)
}

// Disconnect cleanly disconnects from the broker.
func (p *MQTTPublisher) Disconnect() {
	p.client.Disconnect(250)
}
