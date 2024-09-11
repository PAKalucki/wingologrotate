package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Paths []string

type LogEntry struct {
	Path      Paths      `yaml:"path"`
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

func (p *Paths) UnmarshalYAML(value *yaml.Node) error {
	var singlePath string
	if err := value.Decode(&singlePath); err == nil {
		*p = Paths{singlePath}
		return nil
	}

	var multiplePaths []string
	if err := value.Decode(&multiplePaths); err == nil {
		*p = Paths(multiplePaths)
		return nil
	}

	return fmt.Errorf("failed to unmarshal path, expected a string or a list of strings")
}
