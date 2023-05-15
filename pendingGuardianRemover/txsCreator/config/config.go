package config

// Configs holds the general configuration
type Configs struct {
	API     ApiConfig
	TxGen   TxGeneratorConfig
	TxSaver TxSaverConfig
	Keys    KeysConfig
}

// ApiConfig holds the configuration for the API
type ApiConfig struct {
	NetworkAddress string
}

// KeysConfig holds the configuration for the keys
type KeysConfig struct {
	UserPEMFile     string
	GuardianPEMFile string
}

// TxGeneratorConfig holds the tx generation configuration
type TxGeneratorConfig struct {
	NumOfTransactions uint64
	ServiceUID        string
	ChainID           string
	GasPrice          uint64
	GasLimit          uint64
}

// TxSaverConfig holds the tx saver configuration
type TxSaverConfig struct {
	OutputFile string
}

// ContextFlagsConfig the configuration for flags
type ContextFlagsConfig struct {
	LogLevel          string
	ConfigurationFile string
}
