package config

// ContextFlagsConfig the configuration for flags
type ContextFlagsConfig struct {
	WorkingDir  string
	DbPath      string
	Shard       uint32
	Epoch       uint32
	LogLevel    string
	SaveLogFile bool
}
