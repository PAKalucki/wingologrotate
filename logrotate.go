package main

import (
	"log"
	"path/filepath"

	"github.com/robfig/cron/v3"
)

var config = filepath.Join(".", "configs", "wingologrotate.yaml")
var logOutput = filepath.Join(".", "logs", "wingologrotate.log")

func runLogRotation() {
	setupLogging(logOutput)
	defer closeLogFile()

	config, err := loadConfig(config)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	c := cron.New()

	// Schedule tasks based on the configuration
	for _, logEntry := range config.Logs {
		schedule := config.Schedule
		task := createTask(logEntry)

		// Add the task to the cron scheduler
		_, err := c.AddFunc(schedule, task)
		if err != nil {
			log.Printf("Failed to schedule task for path %s: %v", logEntry.Path, err)
		} else {
			log.Printf("Scheduled task for path %s with schedule: %s", logEntry.Path, schedule)
		}
	}

	// Start the cron scheduler
	c.Start()

	// Keep the program running
	select {}

}
