package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

const (
	systemsPath = "/api/v4/systems"
	systemPath  = "/api/v4/systems/" + testSystemID
)

var systemRoutes = routes{
	systemsPath:             `{"systems":[{"system_id":123,"name":"My System","status":"normal"}]}`,
	systemPath:              `{"system_id":123,"name":"My System","status":"normal"}`,
	systemPath + "/summary": `{"system_id":123,"status":"normal","current_power":1500,"energy_today":9000,"energy_lifetime":1000000}`,
	systemPath + "/devices": `{"devices":{"micros":[{"id":1,"serial_number":"S1"}]}}`,
}

func TestSystemCommands(t *testing.T) {
	tests := []struct {
		name string
		cmd  func() (string, error)
	}{
		{"systems", func() (string, error) { return runCmd(t, systemsCmd) }},
		{"system", func() (string, error) { return runCmd(t, systemCmd) }},
		{"summary", func() (string, error) { return runCmd(t, summaryCmd) }},
		{"devices", func() (string, error) { return runCmd(t, devicesCmd) }},
	}
	for _, tt := range tests {
		t.Run(tt.name+"/text", func(t *testing.T) {
			useCloud(t, systemRoutes, http.StatusOK)
			out, err := tt.cmd()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.TrimSpace(out) == "" {
				t.Error("expected text output, got none")
			}
			if json.Valid([]byte(out)) {
				t.Errorf("text output should not be JSON: %q", out)
			}
		})
		t.Run(tt.name+"/json", func(t *testing.T) {
			useCloud(t, systemRoutes, http.StatusOK)
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
			useCloud(t, systemRoutes, http.StatusBadRequest)
			if _, err := tt.cmd(); err == nil {
				t.Error("expected error on HTTP 400")
			}
		})
	}
}

func TestSystemsListError(t *testing.T) {
	useCloud(t, systemRoutes, http.StatusBadRequest)
	_, err := runCmd(t, systemsCmd)
	if err == nil || !strings.Contains(err.Error(), "error listing systems") {
		t.Errorf("expected wrapped list error, got %v", err)
	}
}

func TestSystemIDAutoDetected(t *testing.T) {
	useCloud(t, systemRoutes, http.StatusOK)
	systemID = ""
	if _, err := runCmd(t, summaryCmd); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if systemID != testSystemID {
		t.Errorf("systemID = %q, want %q", systemID, testSystemID)
	}
}
