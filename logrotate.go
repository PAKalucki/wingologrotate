package main

import (
	"fmt"
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

	for _, logEntry := range config.Logs {
		schedule := config.Schedule
		task := createTask(logEntry)

		_, err := c.AddFunc(schedule, task)
		if err != nil {
			log.Printf("Failed to schedule task for path %s: %v", logEntry.Path, err)
		} else {
			log.Printf("Scheduled task for path %s", logEntry.Path)
		}
	}

	c.Start()

	select {}

}

func createTask(logEntry LogEntry) func() {
	return func() {
		for _, path := range logEntry.Path {
			log.Printf("Running task for path: %s", path)
			matchingFiles, err := filepath.Glob(filepath.Clean(path))
			if err != nil {
				log.Printf("Failed to expand wildcard for path %s: %v", path, err)
				return
			}

			switch logEntry.Type {
			case "delete":
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

			case "rotate":
				rotateLogFiles(logEntry)

			default:
				log.Printf("Unsupported task type: %s", logEntry.Type)
			}
		}
	}
}

func rotateLogFiles(logEntry LogEntry) {
	for _, path := range logEntry.Path {
		log.Printf("Rotating logs for path: %s", path)
		matchingFiles, err := filepath.Glob(filepath.Clean(path))
		if err != nil {
			log.Printf("Failed to expand wildcard for path %s: %v", path, err)
			return
		}

		for _, file := range matchingFiles {
			fileInfo, err := os.Stat(file)
			if err != nil {
				log.Printf("Failed to get file info for %s: %v", file, err)
				continue
			}

			rotateDueToSize := false
			if logEntry.Condition.Size != nil {
				maxSize, err := parseSize(*logEntry.Condition.Size)
				if err != nil {
					log.Printf("Invalid size format for rotation: %v", err)
					continue
				}

				if fileInfo.Size() >= maxSize {
					rotateDueToSize = true
				}
			}

			rotateDueToAge := false
			if logEntry.Condition.Age != nil {
				ageDuration, err := parseDuration(*logEntry.Condition.Age)
				if err != nil {
					log.Printf("Invalid age format for rotation: %v", err)
					continue
				}

				fileAge := time.Since(fileInfo.ModTime())
				if fileAge >= ageDuration {
					rotateDueToAge = true
				}
			}

			if rotateDueToSize || rotateDueToAge {
				rotatedFilePath := fmt.Sprintf("%s.%s", file, time.Now().Format("20060102-150405"))
				if err := os.Rename(file, rotatedFilePath); err != nil {
					log.Printf("Failed to rotate log file %s: %v", file, err)
					continue
				}
				log.Printf("Rotated log file: %s to %s", file, rotatedFilePath)

				if logEntry.Condition.Compress == nil || *logEntry.Condition.Compress {
					if err := compressLogFile(rotatedFilePath, "gzip"); err != nil { // todo read me from config
						log.Printf("Failed to compress rotated log file %s: %v", rotatedFilePath, err)
					} else {
						log.Printf("Compressed log file: %s", rotatedFilePath)
					}
				}

				if logEntry.Condition.MaxKeep != nil {
					if err := removeOldLogFiles(filepath.Dir(file), filepath.Base(file), *logEntry.Condition.MaxKeep); err != nil {
						log.Printf("Failed to remove old log files: %v", err)
					}
				}
			}
		}
	}
}
