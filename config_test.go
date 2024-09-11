package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	yamlContent := `
logs:
  - path: "/path/to/log/*.log"
    type: delete
  - path:
      - "/path/to/log1/*.log"
      - "/path/to/log2/*.log"
    type: rotate
    condition:
      size: "100MB"
schedule: "*/5 * * * *"
`

	tempFile := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(tempFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	config, err := loadConfig(tempFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded config values
	if len(config.Logs) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(config.Logs))
	}

	if config.Logs[0].Type != "delete" {
		t.Errorf("Expected type 'delete', got %s", config.Logs[0].Type)
	}

	if len(config.Logs[1].Path) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(config.Logs[1].Path))
	}

	if config.Schedule != "*/5 * * * *" {
		t.Errorf("Expected schedule '*/5 * * * *', got %s", config.Schedule)
	}
}
