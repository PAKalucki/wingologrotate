package main

import (
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var logFile *os.File

func setupLogging(outputFilePath string) {
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

func getExecutablePath() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	return filepath.Dir(exePath)
}

func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	var multiplier int64 = 1

	switch {
	case strings.HasSuffix(sizeStr, "KB"):
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "KB")
	case strings.HasSuffix(sizeStr, "MB"):
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
	case strings.HasSuffix(sizeStr, "GB"):
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "GB")
	}

	sizeValue, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size: %s", sizeStr)
	}

	return sizeValue * multiplier, nil
}

func compressLogFile(filePath string, compressionFormat string) error {
	var compressedFilePath string
	var compressFunc func(input *os.File, output *os.File) error

	switch compressionFormat {
	case "gzip":
		compressedFilePath = filePath + ".gz"
		compressFunc = func(input *os.File, output *os.File) error {
			gzipWriter := gzip.NewWriter(output)
			defer gzipWriter.Close()

			if _, err := io.Copy(gzipWriter, input); err != nil {
				return fmt.Errorf("failed to compress file with gzip: %v", err)
			}
			return nil
		}
	case "zip":
		compressedFilePath = filePath + ".zip"
		compressFunc = func(input *os.File, output *os.File) error {
			archive := zip.NewWriter(output)
			defer archive.Close()

			writer, err := archive.Create(filepath.Base(filePath))
			if err != nil {
				return fmt.Errorf("failed to create zip entry: %v", err)
			}

			if _, err := io.Copy(writer, input); err != nil {
				return fmt.Errorf("failed to write to zip: %v", err)
			}
			return nil
		}
	default:
		return fmt.Errorf("unsupported compression format: %s", compressionFormat)
	}

	inputFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for compression: %v", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(compressedFilePath)
	if err != nil {
		return fmt.Errorf("failed to create compressed file: %v", err)
	}
	defer outputFile.Close()

	if err := compressFunc(inputFile, outputFile); err != nil {
		return err
	}

	if err := inputFile.Close(); err != nil {
		return fmt.Errorf("failed to close input file: %v", err)
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove original file after compression: %v", err)
	}

	return nil
}

func removeOldLogFiles(dir, baseFileName string, maxKeep int) error {
	matches, err := filepath.Glob(filepath.Join(dir, baseFileName+".*"))
	if err != nil {
		return fmt.Errorf("failed to list rotated log files: %v", err)
	}

	sort.Slice(matches, func(i, j int) bool {
		fileInfoI, err := os.Stat(matches[i])
		if err != nil {
			return false
		}
		fileInfoJ, err := os.Stat(matches[j])
		if err != nil {
			return true
		}
		return fileInfoI.ModTime().Before(fileInfoJ.ModTime())
	})

	for len(matches) > maxKeep {
		oldestFile := matches[0]
		if err := os.Remove(oldestFile); err != nil {
			return fmt.Errorf("failed to remove old log file %s: %v", oldestFile, err)
		}
		log.Printf("Removed old log file: %s", oldestFile)
		matches = matches[1:]
	}

	return nil
}

func exePath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}
	return "", err
}
