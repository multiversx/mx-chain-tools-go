package config

// Config defines the main config
type Config struct {
	Keys []string
	Logs LogsConfig
}

// LogsConfig will hold settings related to the logging sub-system
type LogsConfig struct {
	LogFileLifeSpanInSec int
	LogFileLifeSpanInMB  int
}
