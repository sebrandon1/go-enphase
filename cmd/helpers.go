package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var outputFormat string

const outputJSON = "json"

func isJSONOutput() bool {
	return outputFormat == outputJSON
}

func whToKwh(wh int) float64 {
	return float64(wh) / 1000.0
}

// validateDateFlag returns an error when s is non-empty and does not match the
// YYYY-MM-DD layout expected by the Enphase API.
func validateDateFlag(flagName, s string) error {
	if s == "" {
		return nil
	}
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return fmt.Errorf("--%s must be in YYYY-MM-DD format, got %q", flagName, s)
	}
	return nil
}

func printJSON(data any) {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(output))
}
