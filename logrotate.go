package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/robfig/cron/v3"
)

var (
	exeDir     = getExecutablePath()
	configPath = filepath.Join(exeDir, "configs", "wingologrotate.yaml")
	logOutput  = filepath.Join(exeDir, "logs", "wingologrotate.log")
)

func runLogRotation() {
	setupLogging(logOutput)
	defer closeLogFile()

	config, err := loadConfig(configPath)
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
				if logEntry.Condition != nil && logEntry.Condition.Age != nil {
					ageDuration, err := parseDuration(*logEntry.Condition.Age)
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
