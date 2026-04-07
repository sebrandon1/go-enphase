package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestIsJSONOutput(t *testing.T) {
	original := outputFormat
	defer func() { outputFormat = original }()

	outputFormat = "json"
	if !isJSONOutput() {
		t.Error("Expected isJSONOutput() to return true when outputFormat is 'json'")
	}

	outputFormat = "text"
	if isJSONOutput() {
		t.Error("Expected isJSONOutput() to return false when outputFormat is 'text'")
	}

	outputFormat = ""
	if isJSONOutput() {
		t.Error("Expected isJSONOutput() to return false when outputFormat is empty")
	}
}

func TestOutputJSONConstant(t *testing.T) {
	if outputJSON != "json" {
		t.Errorf("Expected outputJSON to be 'json', got '%s'", outputJSON)
	}
}

func TestPrintJSON(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]string{"key": "value"}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("Expected JSON output, got empty string")
	}

	if !strings.Contains(output, "\"key\": \"value\"") {
		t.Errorf("Expected pretty-printed JSON with key/value, got: %s", output)
	}
}

func TestPrintJSONWithStruct(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := testStruct{Name: "test", Age: 30}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "\"name\": \"test\"") {
		t.Errorf("Expected name field in output, got: %s", output)
	}
	if !strings.Contains(output, "\"age\": 30") {
		t.Errorf("Expected age field in output, got: %s", output)
	}
}

func TestPrintJSONWithSlice(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := []string{"a", "b", "c"}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "\"a\"") {
		t.Errorf("Expected element 'a' in output, got: %s", output)
	}
}

func TestPrintJSONWithNil(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printJSON(nil)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "null") {
		t.Errorf("Expected null for nil input, got: %s", output)
	}
}

func TestOutputFlagWiringOnCommandGroups(t *testing.T) {
	original := outputFormat
	defer func() { outputFormat = original }()

	tests := []struct {
		name     string
		args     []string
		want     string
		wantJSON bool
	}{
		{"default is text", []string{"get", "--help"}, "text", false},
		{"json via get group", []string{"get", "-o", "json", "--help"}, "json", true},
		{"json via auth group", []string{"auth", "-o", "json", "--help"}, "json", true},
		{"json via envoy group", []string{"envoy", "-o", "json", "--help"}, "json", true},
		{"json via report group", []string{"report", "-o", "json", "--help"}, "json", true},
		{"text explicit via get", []string{"get", "-o", "text", "--help"}, "text", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputFormat = "text" // reset before each
			rootCmd.SetArgs(tt.args)
			_ = rootCmd.Execute()

			if outputFormat != tt.want {
				t.Errorf("Expected outputFormat=%q, got %q", tt.want, outputFormat)
			}
			if isJSONOutput() != tt.wantJSON {
				t.Errorf("Expected isJSONOutput()=%v, got %v", tt.wantJSON, isJSONOutput())
			}
		})
	}
}

func TestPrintJSONOutputIsIndented(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]string{"key": "value"}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "  ") {
		t.Errorf("Expected indented output with two spaces, got: %s", output)
	}
}
