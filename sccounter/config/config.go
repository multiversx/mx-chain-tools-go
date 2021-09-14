package config

// GeneralConfig holds the entire configuration
type GeneralConfig struct {
	SCDeploysConfig SCDeploysConfig `toml:"config"`
}

// SCDeploysConfig holds the configuration related to sc deploys counter
type SCDeploysConfig struct {
	ElasticInstance ElasticInstanceConfig `toml:"elastic"`
}

// ElasticInstanceConfig holds the configuration needed for connecting to an Elasticsearch instance
type ElasticInstanceConfig struct {
	URL      string `toml:"url"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}
