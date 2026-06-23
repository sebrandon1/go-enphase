package cmd

import (
	"encoding/json"
	"fmt"
	"os"
)

var outputFormat string

const outputJSON = "json"

func isJSONOutput() bool {
	return outputFormat == outputJSON
}

func whToKwh(wh int) float64 {
	return float64(wh) / 1000.0
}

func printJSON(data any) {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(output))
}
