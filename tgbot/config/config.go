package config

// GeneralConfig will hold all the configuration for the bot
type GeneralConfig struct {
	BotConfigs []BotConfig `toml:"config"`
}

// BotConfig will hold configuration for an address
type BotConfig struct {
	General struct {
		ExplorerUrl        string `toml:"explorerURL"`
		GatewayURL         string `toml:"gatewayURL"`
		Address            string `toml:"address"`
		Label              string `toml:"label"`
		BalanceThreshold   string `toml:"balanceThreshold"`
		CheckIntervalInMin int    `toml:"checkIntervalInMin"`
		NotificationStep   int    `toml:"notificationStep"`
	} `toml:"general"`
	Telegram struct {
		GroupID string `toml:"groupID"`
		ApiKey  string `toml:"apiKey"`
	} `toml:"telegram"`
}
