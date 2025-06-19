package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func main() {
	pathToLogsDir := "logs"
	regex := regexp.MustCompile(`elapsed time to commit block             time \[s\] = ([\d\.]+[a-zµ]+)`)
	operation := "commit"
	collectTime(pathToLogsDir, regex, operation)
}

func collectTime(logsDir string, re *regexp.Regexp, operation string) {
	files, err := os.ReadDir(logsDir)
	if err != nil {
		fmt.Printf("Failed to read directory %s: %v\n", logsDir, err)
		return
	}

	var total time.Duration
	var count int

	for _, file := range files {
		if file.IsDir() {
			fmt.Printf("Skipping directory %s\n", file.Name())
			continue
		}
		f, err := os.Open(filepath.Join(logsDir, file.Name()))
		if err != nil {
			fmt.Printf("Failed to open file %s: %v\n", file.Name(), err)
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				dur, err := time.ParseDuration(matches[1])
				if err != nil {
					fmt.Println("Failed to parse duration:", matches[1])
					continue
				}
				total += dur
				count++
			}
		}
		err = f.Close()
		if err != nil {
			fmt.Printf("Failed to close file %s: %v\n", file.Name(), err)
			continue
		}
	}

	if count == 0 {
		fmt.Println("No commit times found.")
		return
	}

	average := total / time.Duration(count)
	fmt.Printf("Average time: %s (%d samples), operation %s, directory: %s\n", average, count, operation, logsDir)
}
