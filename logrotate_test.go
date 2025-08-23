package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRotateLogFiles(t *testing.T) {
	tempDir := t.TempDir()

	file1 := filepath.Join(tempDir, "log1.log")
	originalContent := bytes.Repeat([]byte("a"), 1024*1024*2) // 2MB log file
	if err := os.WriteFile(file1, originalContent, 0644); err != nil {
		t.Fatalf("Failed to write test log file: %v", err)
	}

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
	if len(compressedFiles) != 1 {
		t.Fatalf("Expected 1 compressed log file, but found %d", len(compressedFiles))
	}

	f, err := os.Open(compressedFiles[0])
	if err != nil {
		t.Fatalf("Failed to open compressed file: %v", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gz.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, gz); err != nil {
		t.Fatalf("Failed to read compressed file: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), originalContent) {
		t.Errorf("Rotated log content mismatch")
	}

	remainingFiles, _ := filepath.Glob(filepath.Join(tempDir, "*.log"))
	if len(remainingFiles) != 0 {
		t.Errorf("Expected no uncompressed log files, but found %d", len(remainingFiles))
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

func TestPrePostScripts(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "test.log")
	if err := os.WriteFile(targetFile, []byte("data"), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}
	preFile := filepath.Join(tempDir, "pre.txt")
	postFile := filepath.Join(tempDir, "post.txt")

	logEntry := LogEntry{
		Path:       Paths{targetFile},
		Type:       "delete",
		PreScript:  stringPtr(fmt.Sprintf("echo pre > %s", preFile)),
		PostScript: stringPtr(fmt.Sprintf("echo post > %s", postFile)),
	}

	task := createTask(logEntry)
	task()

	if _, err := os.Stat(preFile); err != nil {
		t.Errorf("pre script did not execute: %v", err)
	}
	if _, err := os.Stat(postFile); err != nil {
		t.Errorf("post script did not execute: %v", err)
	}
}

func TestTimeIntervalCondition(t *testing.T) {
	tempDir := t.TempDir()
	marker := filepath.Join(tempDir, "marker.txt")
	logEntry := LogEntry{
		Path:      Paths{filepath.Join(tempDir, "*.log")},
		Type:      "delete",
		PreScript: stringPtr(fmt.Sprintf("echo run >> %s", marker)),
		Condition: &Condition{TimeInterval: stringPtr("1s")},
	}

	task := createTask(logEntry)
	task()
	task()
	time.Sleep(1 * time.Second)
	task()

	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("failed to read marker: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected script to run twice, ran %d times", len(lines))
	}
}
