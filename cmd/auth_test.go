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
				statusAPIKeySet:       false,
				statusAccessTokenSet:  false,
				statusRefreshTokenSet: false,
				statusClientIDSet:     false,
				statusClientSecretSet: false,
				statusSystemIDSet:     false,
				statusEnvoyIPSet:      false,
				statusEnvoyTokenSet:   false,
				statusEnvoySerialSet:  false,
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
				statusAPIKeySet:       false,
				statusAccessTokenSet:  false,
				statusRefreshTokenSet: false,
				statusClientIDSet:     false,
				statusClientSecretSet: false,
				statusSystemIDSet:     false,
				statusEnvoyIPSet:      true,
				statusEnvoyTokenSet:   true,
				statusEnvoySerialSet:  true,
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
				statusAPIKeySet:       false,
				statusAccessTokenSet:  false,
				statusRefreshTokenSet: false,
				statusClientIDSet:     false,
				statusClientSecretSet: false,
				statusSystemIDSet:     true,
				statusEnvoyIPSet:      false,
				statusEnvoyTokenSet:   false,
				statusEnvoySerialSet:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.teardown()

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
