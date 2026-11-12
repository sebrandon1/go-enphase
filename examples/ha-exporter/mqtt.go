package main

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

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
	client          mqtt.Client
	prefix          string
	serial          string
	disconnectWarn  atomic.Bool // true once a not-connected warning has been logged
}

// NewMQTTPublisher creates an MQTTPublisher and begins connecting to the broker.
// It returns immediately without waiting for the connection to complete; the client
// will retry every 30 s until it succeeds. PublishDiscovery is called automatically
// on every successful (re)connection via the OnConnectHandler.
func NewMQTTPublisher(broker, username, password, prefix, serial, clientID string) (*MQTTPublisher, error) {
	p := &MQTTPublisher{prefix: prefix, serial: serial}

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetUsername(username).
		SetPassword(password).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(30 * time.Second).
		SetOnConnectHandler(func(_ mqtt.Client) {
			Info("MQTT connected to %s; publishing discovery", broker)
			p.PublishDiscovery()
		})

	p.client = mqtt.NewClient(opts)
	p.client.Connect() // non-blocking: retry loop handles the rest
	return p, nil
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
		topic := fmt.Sprintf("%s/sensor/solar_%s_%s/config", p.prefix, p.serial, s.field)
		payload := fmt.Sprintf(haDiscoveryTemplate,
			s.name, p.prefix, p.serial, s.unit, s.field,
			p.serial, s.field, p.serial,
		)
		if token := p.client.Publish(topic, 1, true, payload); token.Wait() && token.Error() != nil {
			Warn("mqtt discovery publish failed for %s: %v", s.field, token.Error())
		}
	}
}

// PublishState publishes the current metric values as a JSON state payload.
// If the client is not connected, it logs a single warning per outage and skips.
func (p *MQTTPublisher) PublishState(snap Snapshot) {
	if !p.client.IsConnected() {
		if p.disconnectWarn.CompareAndSwap(false, true) {
			Warn("MQTT not connected; skipping state publish until reconnected")
		}
		return
	}
	p.disconnectWarn.Store(false)

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
	if token := p.client.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
		Warn("mqtt state publish failed: %v", token.Error())
	}
}

// Disconnect cleanly disconnects from the broker.
func (p *MQTTPublisher) Disconnect() {
	p.client.Disconnect(250)
}
