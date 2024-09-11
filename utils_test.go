package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
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

func TestParseSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"10KB", 10 * 1024, false},             // 10 KB
		{"20MB", 20 * 1024 * 1024, false},      // 20 MB
		{"5GB", 5 * 1024 * 1024 * 1024, false}, // 5 GB
		{"100", 100, false},                    // 100 bytes (default to bytes if no suffix)
		{"15kb", 15 * 1024, false},             // 15 kb (case insensitive)
		{"2mb", 2 * 1024 * 1024, false},        // 2 mb (case insensitive)
		{"1gb", 1024 * 1024 * 1024, false},     // 1 gb (case insensitive)
		{"", 0, true},                          // Invalid: Empty input
		{"10XB", 0, true},                      // Invalid: Unrecognized suffix
		{"abc", 0, true},                       // Invalid: Non-numeric input
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSize(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("parseSize(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExePath(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tempDir := t.TempDir()

	exeFile := filepath.Join(tempDir, "testexe")
	if err := os.WriteFile(exeFile, []byte{}, 0755); err != nil {
		t.Fatalf("Failed to create test executable: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		expectPath  string
		expectError bool
	}{
		{
			name:        "Valid executable path",
			args:        []string{exeFile},
			expectPath:  exeFile,
			expectError: false,
		},
		{
			name:        "Path to a directory",
			args:        []string{tempDir},
			expectPath:  "",
			expectError: true,
		},
		{
			name:        "Executable without extension",
			args:        []string{exeFile[:len(exeFile)-4]},
			expectPath:  exeFile,
			expectError: false,
		},
		{
			name:        "Non-existent executable",
			args:        []string{filepath.Join(tempDir, "nonexistent.exe")},
			expectPath:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = []string{tt.args[0]}

			got, err := exePath()
			if (err != nil) != tt.expectError {
				t.Errorf("exePath() error = %v, expected error = %v", err, tt.expectError)
				return
			}
			if got != tt.expectPath {
				t.Errorf("exePath() = %v, expected path = %v", got, tt.expectPath)
			}
		})
	}
}

func TestCompressLogFile(t *testing.T) {
	tempDir := t.TempDir()

	originalFilePath := filepath.Join(tempDir, "test.log")
	originalContent := []byte("This is a test log file content.")
	err := os.WriteFile(originalFilePath, originalContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	err = compressLogFile(originalFilePath)
	if err != nil {
		t.Fatalf("compressLogFile() error: %v", err)
	}

	compressedFilePath := originalFilePath + ".gz"
	if _, err := os.Stat(compressedFilePath); os.IsNotExist(err) {
		t.Errorf("Compressed file does not exist: %s", compressedFilePath)
	}

	if _, err := os.Stat(originalFilePath); err == nil {
		t.Errorf("Original file still exists: %s", originalFilePath)
	}

	compressedFile, err := os.Open(compressedFilePath)
	if err != nil {
		t.Fatalf("Failed to open compressed file: %v", err)
	}
	defer compressedFile.Close()

	gzipReader, err := gzip.NewReader(compressedFile)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	var decompressedContent bytes.Buffer
	if _, err := io.Copy(&decompressedContent, gzipReader); err != nil {
		t.Fatalf("Failed to decompress file: %v", err)
	}

	if !bytes.Equal(decompressedContent.Bytes(), originalContent) {
		t.Errorf("Decompressed content does not match original. Got: %s, Want: %s", decompressedContent.String(), string(originalContent))
	}
}
