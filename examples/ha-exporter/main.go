package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	configPath := flag.String("config", "config.json", "path to JSON config file")
	dryRun := flag.Bool("dry-run", false, "log what would be published without making API calls or publishing to MQTT")
	metricsAddr := flag.String("metrics-addr", "", "override metrics listen address (e.g. :9090)")
	flag.Parse()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}
	if *metricsAddr != "" {
		cfg.MetricsAddr = *metricsAddr
	}

	if *dryRun {
		Info("dry-run mode: no API calls or MQTT publishes will be made")
		Info("config: metrics_addr=%s poll_interval=%s mqtt_broker=%s",
			cfg.MetricsAddr, cfg.PollInterval, cfg.MQTTBroker)
		return
	}

	pollInterval, err := time.ParseDuration(cfg.PollInterval)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid poll_interval %q: %v\n", cfg.PollInterval, err)
		os.Exit(1)
	}

	col, err := NewCollector(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating collector: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start MQTT publisher if configured.
	var mqttPub *MQTTPublisher
	if cfg.MQTTBroker != "" && cfg.EnvoySerial != "" {
		mqttPub, err = NewMQTTPublisher(cfg.MQTTBroker, cfg.MQTTUsername, cfg.MQTTPassword,
			cfg.MQTTTopicPrefix, cfg.EnvoySerial)
		if err != nil {
			Warn("MQTT unavailable: %v — continuing without MQTT", err)
		} else {
			defer mqttPub.Disconnect()
			mqttPub.PublishDiscovery()
			Info("MQTT discovery published to %s", cfg.MQTTBroker)
		}
	}

	// Collector goroutine.
	go col.Run(ctx, pollInterval)

	// MQTT state publisher goroutine.
	if mqttPub != nil {
		go func() {
			ticker := time.NewTicker(pollInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					mqttPub.PublishState(col.Snapshot())
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Metrics HTTP server.
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", ServeMetrics(col))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	srv := &http.Server{Addr: cfg.MetricsAddr, Handler: mux}
	go func() {
		Info("metrics server listening on %s", cfg.MetricsAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			Error("metrics server: %v", err)
		}
	}()

	<-ctx.Done()
	Info("shutting down")
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()
	srv.Shutdown(shutCtx) //nolint:errcheck
}
