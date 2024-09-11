package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type LogEntry struct {
	Path      string     `yaml:"path"`
	Type      string     `yaml:"type"`
	Condition *Condition `yaml:"condition,omitempty"`
}

type Condition struct {
	Age      *string `yaml:"age,omitempty"`
	MaxKeep  *int    `yaml:"max_keep,omitempty"`
	Size     *string `yaml:"size,omitempty"`
	Compress *bool   `yaml:"compress,omitempty"`
}

type Config struct {
	Logs     []LogEntry `yaml:"logs"`
	Schedule string     `yaml:"schedule"`
}

func (entry *LogEntry) setDefaults() {
	if entry.Type == "rotate" && entry.Condition == nil {
		entry.Condition = &Condition{}

		if entry.Condition.Compress == nil {
			defaultCompress := true
			entry.Condition.Compress = &defaultCompress
		}
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

	for i := range config.Logs {
		config.Logs[i].setDefaults()
	}

	return config, nil
}
