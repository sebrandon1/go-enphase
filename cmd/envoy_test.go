package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

var envoyRoutes = routes{
	"/production.json":             `{"production":[{"type":"inverters","activeCount":2,"wNow":1200}],"consumption":[]}`,
	"/ivp/sensors/readings_object": `{"sensors":[]}`,
	"/api/v1/production/inverters": `[{"serialNumber":"I1","lastReportDate":1700000000,"lastReportWatts":250,"maxReportWatts":300,"devType":1}]`,
	"/ivp/meters":                  `[{"eid":1,"state":"enabled","measurementType":"production"}]`,
	"/ivp/meters/readings":         `[{"eid":1,"activePower":1000}]`,
}

func TestEnvoyCommands(t *testing.T) {
	tests := []struct {
		name string
		cmd  func() (string, error)
	}{
		{"status", func() (string, error) { return runCmd(t, envoyStatusCmd) }},
		{"inverters", func() (string, error) { return runCmd(t, envoyInvertersCmd) }},
		{"meters", func() (string, error) { return runCmd(t, envoyMetersCmd) }},
		{"meter-readings", func() (string, error) { return runCmd(t, envoyMeterReadingsCmd) }},
	}
	for _, tt := range tests {
		t.Run(tt.name+"/text", func(t *testing.T) {
			useEnvoy(t, envoyRoutes, http.StatusOK)
			out, err := tt.cmd()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.TrimSpace(out) == "" {
				t.Error("expected text output, got none")
			}
		})
		t.Run(tt.name+"/json", func(t *testing.T) {
			useEnvoy(t, envoyRoutes, http.StatusOK)
			outputFormat = outputJSON
			out, err := tt.cmd()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !json.Valid([]byte(out)) {
				t.Errorf("expected valid JSON, got %q", out)
			}
		})
		t.Run(tt.name+"/server error", func(t *testing.T) {
			useEnvoy(t, envoyRoutes, http.StatusBadRequest)
			if _, err := tt.cmd(); err == nil {
				t.Error("expected error on HTTP 400")
			}
		})
	}
}

func TestEnvoyStatusErrorIsWrapped(t *testing.T) {
	useEnvoy(t, envoyRoutes, http.StatusBadRequest)
	_, err := runCmd(t, envoyStatusCmd)
	if err == nil || !strings.Contains(err.Error(), "error getting envoy status") {
		t.Errorf("expected wrapped status error, got %v", err)
	}
}

func TestEnvoyCommandRequiresIP(t *testing.T) {
	origClient, origIP := localClient, envoyIP
	localClient, envoyIP = nil, ""
	t.Cleanup(func() { localClient, envoyIP = origClient, origIP })
	if _, err := runCmd(t, envoyStatusCmd); err == nil {
		t.Error("expected error when no envoy IP is configured")
	}
}
