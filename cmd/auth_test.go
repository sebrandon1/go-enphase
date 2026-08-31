package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestAuthStatusFields(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		teardown func()
		want     map[string]bool
	}{
		{
			name:     "all fields empty",
			setup:    func() {},
			teardown: func() {},
			want: map[string]bool{
				"api_key_set":       false,
				"access_token_set":  false,
				"refresh_token_set": false,
				"client_id_set":     false,
				"client_secret_set": false,
				"system_id_set":     false,
				"envoy_ip_set":      false,
				"envoy_token_set":   false,
				"envoy_serial_set":  false,
			},
		},
		{
			name: "envoy fields set",
			setup: func() {
				envoyIP = "192.168.1.10"
				envoyToken = "mytoken"
				envoySerial = "123456"
			},
			teardown: func() {
				envoyIP = ""
				envoyToken = ""
				envoySerial = ""
			},
			want: map[string]bool{
				"api_key_set":       false,
				"access_token_set":  false,
				"refresh_token_set": false,
				"client_id_set":     false,
				"client_secret_set": false,
				"system_id_set":     false,
				"envoy_ip_set":      true,
				"envoy_token_set":   true,
				"envoy_serial_set":  true,
			},
		},
		{
			name: "system_id set",
			setup: func() {
				systemID = "67890"
			},
			teardown: func() {
				systemID = ""
			},
			want: map[string]bool{
				"api_key_set":       false,
				"access_token_set":  false,
				"refresh_token_set": false,
				"client_id_set":     false,
				"client_secret_set": false,
				"system_id_set":     true,
				"envoy_ip_set":      false,
				"envoy_token_set":   false,
				"envoy_serial_set":  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.teardown()

			// Capture stdout
			orig := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			authStatusCmd.Run(nil, nil)

			w.Close()
			os.Stdout = orig

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)

			var got map[string]bool
			if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
				t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
			}

			for field, wantVal := range tt.want {
				gotVal, ok := got[field]
				if !ok {
					t.Errorf("field %q missing from output", field)
					continue
				}
				if gotVal != wantVal {
					t.Errorf("field %q: want %v, got %v", field, wantVal, gotVal)
				}
			}
		})
	}
}
