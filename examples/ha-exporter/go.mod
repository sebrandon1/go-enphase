module github.com/sebrandon1/go-enphase/examples/ha-exporter

go 1.27.0

toolchain go1.27.4

require (
	github.com/eclipse/paho.mqtt.golang v1.5.2
	github.com/sebrandon1/go-enphase v0.0.0
	github.com/sebrandon1/go-toolbox v0.1.2
)

replace github.com/sebrandon1/go-enphase => ../../

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
)
