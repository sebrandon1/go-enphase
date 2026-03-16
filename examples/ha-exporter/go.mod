module github.com/sebrandon1/go-enphase/examples/ha-exporter

go 1.24.0

require (
	github.com/eclipse/paho.mqtt.golang v1.5.0
	github.com/sebrandon1/go-enphase v0.0.0
)

replace github.com/sebrandon1/go-enphase => ../../

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
)
