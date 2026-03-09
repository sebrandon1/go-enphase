package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

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
