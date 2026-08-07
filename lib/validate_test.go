package lib

import "testing"

func TestValidateSystemID(t *testing.T) {
	valid := []string{"1", "123", "9999999"}
	for _, id := range valid {
		if err := validateSystemID(id); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", id, err)
		}
	}

	invalid := []string{
		"",
		"abc",
		"123abc",
		"123/../../../oauth/token",
		"../etc/passwd",
		"-1",
		"1 2",
	}
	for _, id := range invalid {
		if err := validateSystemID(id); err == nil {
			t.Errorf("expected %q to be invalid, got nil error", id)
		}
	}
}
