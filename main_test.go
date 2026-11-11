package main

import (
	"os"
	"os/exec"
	"testing"
)

// TestMain re-execs the test binary as the CLI when the sentinel env var is set.
func TestMain(m *testing.M) {
	if os.Getenv("GO_ENPHASE_TEST_MAIN") == "1" {
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestExitCodes(t *testing.T) {
	tests := []struct {
		args     []string
		wantCode int
	}{
		{[]string{"bogus"}, 1},
		{[]string{"get", "--nope"}, 1},
		{[]string{"--help"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.args[0], func(t *testing.T) {
			cmd := exec.Command(os.Args[0], tt.args...)
			cmd.Env = append(os.Environ(), "GO_ENPHASE_TEST_MAIN=1", "HOME="+t.TempDir())
			err := cmd.Run()
			code := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					code = exitErr.ExitCode()
				}
			}
			if code != tt.wantCode {
				t.Errorf("args %v: exit code = %d, want %d", tt.args, code, tt.wantCode)
			}
		})
	}
}
