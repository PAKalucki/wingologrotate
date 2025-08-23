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
    pre_script: "echo pre"
    post_script: "echo post"
    condition:
      time_interval: "1h"
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

	if len(config.Logs) != 2 {
		t.Errorf("Expected 2 log entries, got %d", len(config.Logs))
	}

	if config.Logs[0].Type != "delete" {
		t.Errorf("Expected type 'delete', got %s", config.Logs[0].Type)
	}
	if config.Logs[0].PreScript == nil || *config.Logs[0].PreScript != "echo pre" {
		t.Errorf("Expected pre_script 'echo pre', got %v", config.Logs[0].PreScript)
	}
	if config.Logs[0].PostScript == nil || *config.Logs[0].PostScript != "echo post" {
		t.Errorf("Expected post_script 'echo post', got %v", config.Logs[0].PostScript)
	}
	if config.Logs[0].Condition == nil || config.Logs[0].Condition.TimeInterval == nil || *config.Logs[0].Condition.TimeInterval != "1h" {
		t.Errorf("Expected time_interval '1h', got %v", config.Logs[0].Condition)
	}

	if len(config.Logs[1].Path) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(config.Logs[1].Path))
	}

	if config.Schedule != "*/5 * * * *" {
		t.Errorf("Expected schedule '*/5 * * * *', got %s", config.Schedule)
	}
}
