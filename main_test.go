package main

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"1d", 24 * time.Hour, false},
		{"1h", time.Hour, false},
		{"30m", 30 * time.Minute, false},
		{"15s", 15 * time.Second, false},
		{"2h", 2 * time.Hour, false},
		{"0m", 0, false},
		{"invalid", 0, true}, // Invalid format
		{"10x", 0, true},     // Invalid unit
		{"", 0, true},        // Empty input
		{"-5m", 0, true},     // Negative value
		{"1h30m", 0, true},   // Unsupported combined format
		{"1.5h", 0, true},    // Unsupported float format
	}

	for _, tt := range tests {
		result, err := parseDuration(tt.input)

		if tt.hasError && err == nil {
			t.Errorf("parseDuration(%s) expected error, got none", tt.input)
		}
		if !tt.hasError && err != nil {
			t.Errorf("parseDuration(%s) unexpected error: %v", tt.input, err)
		}
		if result != tt.expected {
			t.Errorf("parseDuration(%s) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}
