package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRotateLogFiles(t *testing.T) {
	tempDir := t.TempDir()

	file1 := filepath.Join(tempDir, "log1.log")
	file2 := filepath.Join(tempDir, "log2.log")
	_ = os.WriteFile(file1, make([]byte, 1024*1024*5), 0644)  // 5MB log file
	_ = os.WriteFile(file2, make([]byte, 1024*1024*10), 0644) // 10MB log file

	logEntry := LogEntry{
		Path: Paths{filepath.Join(tempDir, "*.log")},
		Type: "rotate",
		Condition: &Condition{
			Size:     stringPtr("1MB"),
			Compress: boolPtr(true),
			MaxKeep:  intPtr(1),
		},
	}

	rotateLogFiles(logEntry)

	compressedFiles, _ := filepath.Glob(filepath.Join(tempDir, "*.gz"))
	if len(compressedFiles) == 0 {
		t.Errorf("Expected compressed log files, but found none.")
	}

	remainingFiles, _ := filepath.Glob(filepath.Join(tempDir, "*.log"))
	if len(remainingFiles) > 1 {
		t.Errorf("Expected at most 1 log file, but found %d", len(remainingFiles))
	}
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}

func TestCreateTask(t *testing.T) {
	logEntry := LogEntry{
		Path: Paths{"/tmp/test/logs/delete/*.log"},
		Type: "delete",
		Condition: &Condition{
			Age: stringPtr("1h"),
		},
	}

	task := createTask(logEntry)

	logBuf := new(bytes.Buffer)
	log.SetOutput(logBuf)

	task()

	if !strings.Contains(logBuf.String(), "Running task for path: /tmp/test/logs/delete/*.log") {
		t.Errorf("Expected log output for running task, got %s", logBuf.String())
	}
}
