package config

// Configs holds the general configuration
type Configs struct {
	API      ApiConfig
	TxSender TxSenderConfig
}

// ApiConfig holds the configuration for the API
type ApiConfig struct {
	NetworkAddress string
}

// TxSenderConfig holds the tx sender's configuration
type TxSenderConfig struct {
	TxsFile                string
	DelayBetweenSendsInSec uint64
}

// ContextFlagsConfig the configuration for flags
type ContextFlagsConfig struct {
	LogLevel          string
	ConfigurationFile string
	WorkingDir        string
	SaveLogFile       bool
}
