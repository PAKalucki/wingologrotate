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

const config = "./config/wingologrotate.yaml"

// LogEntry represents a single log configuration entry
type LogEntry struct {
	Path      string     `yaml:"path"`
	Type      string     `yaml:"type"`
	Schedule  string     `yaml:"schedule,omitempty"`
	Size      string     `yaml:"size,omitempty"`
	MaxKeep   int        `yaml:"max_keep,omitempty"`
	Condition *Condition `yaml:"condition,omitempty"`
}

type Condition struct {
	Age string `yaml:"age,omitempty"`
}

type Config struct {
	Logs []LogEntry `yaml:"logs"`
}

type Schedule struct {
	schedule string `yaml:"schedule"`
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

func main() {
	// Read the YAML file
	yamlFile, err := os.ReadFile(config)
	if err != nil {
		log.Fatalf("Failed to read YAML file: %v", err)
		os.Exit(1)
	}

	// Unmarshal the YAML file into the Config struct
	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Failed to parse YAML file: %v", err)
		os.Exit(1)
	}

	for _, logEntry := range config.Logs {
		fmt.Println("Path:", logEntry.Path)
		matchingFiles, err := filepath.Glob(logEntry.Path)
		if err != nil {
			log.Printf("Failed to expand wildcard for path %s: %v", logEntry.Path, err)
			continue
		}

		if len(matchingFiles) == 0 {
			fmt.Println("No files found for path:", logEntry.Path)
		} else {
			for _, file := range matchingFiles {
				fmt.Println("Matched file:", file)
			}
		}
	}

}
