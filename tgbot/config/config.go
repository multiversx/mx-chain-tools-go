package config

type GeneralConfig struct {
	BotConfig struct {
		General struct {
			GatewayURL         string `toml:"gatewayURL"`
			Address            string `toml:"address"`
			BalanceThreshold   string `toml:"balanceThreshold"`
			CheckIntervalInMin int    `toml:"checkIntervalInMin"`
			NotificationStep   int    `toml:"notificationStep"`
		} `toml:"general"`
		Telegram struct {
			GroupID string `toml:"groupID"`
			ApiKey  string `toml:"apiKey"`
		} `toml:"telegram"`
	} `toml:"config"`
}
