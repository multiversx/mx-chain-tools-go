package logParser

import (
	"strings"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type loggerFormatter struct {
	mutErrors    sync.RWMutex
	errorStrings []string
}

// NewLoggerFormatter creates a new logger formatter instance
func NewLoggerFormatter() *loggerFormatter {
	return &loggerFormatter{
		errorStrings: make([]string, 0),
	}
}

// Output will always return nil byte slice but records any type of errors encountered
func (formatter *loggerFormatter) Output(line logger.LogLineHandler) []byte {
	if line.GetLogLevel() != int32(logger.LogError) {
		return nil
	}

	data := append([]string{line.GetMessage()}, line.GetArgs()...)

	formatter.mutErrors.Lock()
	formatter.errorStrings = append(formatter.errorStrings, strings.Join(data, " "))
	formatter.mutErrors.Unlock()

	return nil
}

// GetAllErrorStrings returns a copy of all contained error strings
func (formatter *loggerFormatter) GetAllErrorStrings() []string {
	result := make([]string, len(formatter.errorStrings))

	formatter.mutErrors.RLock()
	copy(result, formatter.errorStrings)
	formatter.mutErrors.RUnlock()

	return result
}

// IsInterfaceNil returns true if there is no value under the interface
func (formatter *loggerFormatter) IsInterfaceNil() bool {
	return formatter == nil
}
