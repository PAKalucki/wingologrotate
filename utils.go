package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

var logFile *os.File

func setupLogging(outputFilePath string) {
	// Create the directory if it doesn't exist
	logDir := filepath.Dir(outputFilePath)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.MkdirAll(logDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create log directory: %v", err)
		}
	}

	logFile, err := os.OpenFile(outputFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// defer logFile.Close()

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func closeLogFile() {
	if logFile != nil {
		err := logFile.Close()
		if err != nil {
			log.Printf("Error closing log file: %v", err)
		}
	}
}

func parseDuration(ageStr string) (time.Duration, error) {
	if len(ageStr) < 2 {
		return 0, fmt.Errorf("invalid age format: %s", ageStr)
	}

	// Determine the time unit (h for hours, m for minutes, etc.)
	unit := ageStr[len(ageStr)-1]
	value := ageStr[:len(ageStr)-1]

	ageValue, err := strconv.Atoi(value)
	if err != nil || ageValue < 0 {
		return 0, fmt.Errorf("invalid age value: %s", value)
	}

	switch unit {
	case 'd':
		return time.Duration(ageValue) * time.Hour * 24, nil
	case 'h':
		return time.Duration(ageValue) * time.Hour, nil
	case 'm':
		return time.Duration(ageValue) * time.Minute, nil
	case 's':
		return time.Duration(ageValue) * time.Second, nil
	default:
		return 0, fmt.Errorf("invalid age unit: %c", unit)
	}
}

func loadConfig(filePath string) (Config, error) {
	yamlFile, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return Config{}, fmt.Errorf("failed to read YAML file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse YAML file: %v", err)
	}

	return config, nil
}

func createTask(logEntry LogEntry) func() {
	return func() {
		log.Printf("Running task for path: %s", logEntry.Path)
		matchingFiles, err := filepath.Glob(filepath.Clean(logEntry.Path))
		if err != nil {
			log.Printf("Failed to expand wildcard for path %s: %v", logEntry.Path, err)
			return
		}

		if logEntry.Type == "delete" {
			for _, file := range matchingFiles {
				if logEntry.Condition != nil && logEntry.Condition.Age != "" {
					ageDuration, err := parseDuration(logEntry.Condition.Age)
					if err != nil {
						log.Printf("Invalid age format for file %s: %v", file, err)
						continue
					}

					fileInfo, err := os.Stat(file)
					if err != nil {
						log.Printf("Failed to get file info for %s: %v", file, err)
						continue
					}
					fileAge := time.Since(fileInfo.ModTime())

					if fileAge < ageDuration {
						// log.Printf("Skipping file %s, does not meet age condition (%s)", file, logEntry.Condition.Age)
						continue
					}
				}

				log.Printf("Deleting file: %s", file)
				err := os.Remove(file)
				if err != nil {
					log.Printf("Failed to delete file %s: %v", file, err)
				} else {
					log.Printf("Successfully deleted file: %s", file)
				}
			}
		}
	}
}

func getExecutablePath() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	return filepath.Dir(exePath)
}
