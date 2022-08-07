package main

import (
	"fmt"
	"io"
	"os"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common/logging"
)

const (
	defaultLogsPath      = "logs"
	logFilePrefix        = "accounts-exporter"
	logFileLifeSpanInSec = 86400
	logMaxSizeInMB       = 1024
)

var log = logger.GetOrCreate("main")

func initializeLogger(logLevel string) (io.Closer, error) {
	currentDirectory, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	err = logger.SetLogLevel(logLevel)
	if err != nil {
		return nil, err
	}

	logger.ToggleLoggerName(true)

	fileLogging, err := logging.NewFileLogging(logging.ArgsFileLogging{
		WorkingDir:      currentDirectory,
		DefaultLogsPath: defaultLogsPath,
		LogFilePrefix:   logFilePrefix,
	})
	if err != nil {
		return nil, fmt.Errorf("%w creating a log file", err)
	}

	err = fileLogging.ChangeFileLifeSpan(time.Second*time.Duration(logFileLifeSpanInSec), logMaxSizeInMB)
	if err != nil {
		return nil, err
	}

	return fileLogging, nil
}
