package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommandExists(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}
	if rootCmd.Use != "enphase" {
		t.Errorf("Expected Use 'enphase', got '%s'", rootCmd.Use)
	}
}

func TestRootCommandShort(t *testing.T) {
	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}
}

func TestGetCommandExists(t *testing.T) {
	if getCmd == nil {
		t.Fatal("getCmd should not be nil")
	}
	if getCmd.Use != "get" {
		t.Errorf("Expected Use 'get', got '%s'", getCmd.Use)
	}
}

func TestAuthCommandExists(t *testing.T) {
	if authCmd == nil {
		t.Fatal("authCmd should not be nil")
	}
	if authCmd.Use != "auth" {
		t.Errorf("Expected Use 'auth', got '%s'", authCmd.Use)
	}
}

func TestEnvoyCommandExists(t *testing.T) {
	if envoyCmd == nil {
		t.Fatal("envoyCmd should not be nil")
	}
	if envoyCmd.Use != "envoy" {
		t.Errorf("Expected Use 'envoy', got '%s'", envoyCmd.Use)
	}
}

func TestRootPersistentFlags(t *testing.T) {
	flags := rootCmd.PersistentFlags()

	tests := []struct {
		name string
		flag string
	}{
		{"api-key flag", "api-key"},
		{"access-token flag", "access-token"},
		{"refresh-token flag", "refresh-token"},
		{"client-id flag", "client-id"},
		{"client-secret flag", "client-secret"},
		{"system-id flag", "system-id"},
		{"envoy-ip flag", "envoy-ip"},
		{"envoy-token flag", "envoy-token"},
		{"envoy-serial flag", "envoy-serial"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := flags.Lookup(tt.flag)
			if f == nil {
				t.Errorf("Expected flag '%s' to be registered", tt.flag)
			}
		})
	}
}

func TestRootSubcommands(t *testing.T) {
	subcommands := rootCmd.Commands()

	expected := map[string]bool{
		"get":   false,
		"auth":  false,
		"envoy": false,
	}

	for _, cmd := range subcommands {
		if _, ok := expected[cmd.Use]; ok {
			expected[cmd.Use] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Expected subcommand '%s' under root", name)
		}
	}
}

func TestGetSubcommands(t *testing.T) {
	subcommands := getCmd.Commands()

	expected := map[string]bool{
		"systems":         false,
		"summary":         false,
		"devices":         false,
		"production":      false,
		"energy-lifetime": false,
		"consumption":     false,
		"battery":         false,
	}

	for _, cmd := range subcommands {
		if _, ok := expected[cmd.Use]; ok {
			expected[cmd.Use] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Expected subcommand '%s' under get", name)
		}
	}
}

func TestAuthSubcommands(t *testing.T) {
	subcommands := authCmd.Commands()

	expected := map[string]bool{
		"status":      false,
		"refresh":     false,
		"envoy-token": false,
	}

	for _, cmd := range subcommands {
		if _, ok := expected[cmd.Use]; ok {
			expected[cmd.Use] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Expected subcommand '%s' under auth", name)
		}
	}
}

func TestEnvoySubcommands(t *testing.T) {
	subcommands := envoyCmd.Commands()

	expected := map[string]bool{
		"status":  false,
		"sensors": false,
	}

	for _, cmd := range subcommands {
		if _, ok := expected[cmd.Use]; ok {
			expected[cmd.Use] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Expected subcommand '%s' under envoy", name)
		}
	}
}

func TestEnergyLifetimeFlags(t *testing.T) {
	flags := energyLifetimeCmd.Flags()

	for _, flag := range []string{"start-date", "end-date"} {
		f := flags.Lookup(flag)
		if f == nil {
			t.Errorf("Expected flag '%s' on energy-lifetime command", flag)
		}
	}
}

func TestConsumptionFlags(t *testing.T) {
	flags := consumptionCmd.Flags()

	for _, flag := range []string{"start-date", "end-date"} {
		f := flags.Lookup(flag)
		if f == nil {
			t.Errorf("Expected flag '%s' on consumption command", flag)
		}
	}
}

func TestEnvoyTokenFlags(t *testing.T) {
	flags := envoyTokenCmd.Flags()

	for _, flag := range []string{"email", "password"} {
		f := flags.Lookup(flag)
		if f == nil {
			t.Errorf("Expected flag '%s' on envoy-token command", flag)
		}
	}
}

func TestExecuteReturnsNoError(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute with --help failed: %v", err)
	}
}

func TestAllCommandsHaveShortDescriptions(t *testing.T) {
	cmds := map[string]*cobra.Command{
		"root":             rootCmd,
		"get":              getCmd,
		"auth":             authCmd,
		"envoy":            envoyCmd,
		"systems":          systemsCmd,
		"summary":          summaryCmd,
		"devices":          devicesCmd,
		"production":       productionCmd,
		"energy-lifetime":  energyLifetimeCmd,
		"consumption":      consumptionCmd,
		"battery":          batteryCmd,
		"auth status":      authStatusCmd,
		"auth refresh":     authRefreshCmd,
		"auth envoy-token": envoyTokenCmd,
		"envoy status":     envoyStatusCmd,
		"envoy sensors":    envoySensorsCmd,
	}

	for name, c := range cmds {
		t.Run(name, func(t *testing.T) {
			if c.Short == "" {
				t.Errorf("Command '%s' should have a Short description", name)
			}
		})
	}
}

func TestLeafCommandsHaveRunFunctions(t *testing.T) {
	leafCmds := map[string]*cobra.Command{
		"systems":          systemsCmd,
		"summary":          summaryCmd,
		"devices":          devicesCmd,
		"production":       productionCmd,
		"energy-lifetime":  energyLifetimeCmd,
		"consumption":      consumptionCmd,
		"battery":          batteryCmd,
		"auth status":      authStatusCmd,
		"auth refresh":     authRefreshCmd,
		"auth envoy-token": envoyTokenCmd,
		"envoy status":     envoyStatusCmd,
		"envoy sensors":    envoySensorsCmd,
	}

	for name, c := range leafCmds {
		t.Run(name, func(t *testing.T) {
			if c.Run == nil {
				t.Errorf("Leaf command '%s' should have a Run function", name)
			}
		})
	}
}

func TestParentCommandsHaveNoRunFunction(t *testing.T) {
	parentCmds := map[string]*cobra.Command{
		"root":  rootCmd,
		"get":   getCmd,
		"auth":  authCmd,
		"envoy": envoyCmd,
	}

	for name, c := range parentCmds {
		t.Run(name, func(t *testing.T) {
			if c.Run != nil {
				t.Errorf("Parent command '%s' should not have a Run function", name)
			}
		})
	}
}

func TestGetHelpDoesNotError(t *testing.T) {
	rootCmd.SetArgs([]string{"get", "--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute 'get --help' failed: %v", err)
	}
}

func TestAuthHelpDoesNotError(t *testing.T) {
	rootCmd.SetArgs([]string{"auth", "--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute 'auth --help' failed: %v", err)
	}
}

func TestEnvoyHelpDoesNotError(t *testing.T) {
	rootCmd.SetArgs([]string{"envoy", "--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Execute 'envoy --help' failed: %v", err)
	}
}
