package process

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-tools-go/tgbot/config"
)

var log = logger.GetOrCreate("process")

type notifier struct {
	gatewayURL         string
	address            string
	balanceThreshold   *big.Int
	checkIntervalInMin int
	notificationStep   int

	telegramBotKey  string
	telegramGroupID string

	notified bool
	counter  int
}

func NewBalanceNotifier(cfg *config.GeneralConfig) (*notifier, error) {
	balanceThreshold, ok := big.NewInt(0).SetString(cfg.BotConfig.General.BalanceThreshold, 10)
	if !ok {
		return nil, errors.New("invalid balance threshold")
	}

	return &notifier{
		gatewayURL:         cfg.BotConfig.General.GatewayURL,
		address:            cfg.BotConfig.General.Address,
		balanceThreshold:   balanceThreshold,
		checkIntervalInMin: cfg.BotConfig.General.CheckIntervalInMin,
		notificationStep:   cfg.BotConfig.General.NotificationStep,

		telegramBotKey:  cfg.BotConfig.Telegram.ApiKey,
		telegramGroupID: cfg.BotConfig.Telegram.GroupID,
	}, nil
}

func (n *notifier) StartNotifier() {
	n.checkBalanceAndNotifyIfNeeded()

	checkInterval := time.Duration(n.checkIntervalInMin) * time.Minute
	for {
		select {
		case <-time.After(checkInterval):
			n.checkBalanceAndNotifyIfNeeded()
		}
	}
}

func (n *notifier) checkBalanceAndNotifyIfNeeded() {
	if n.counter == n.notificationStep {
		n.counter = 0
		n.notified = false
	}

	accountBalance, err := GetAddressBalance(n.gatewayURL, n.address)
	if err != nil {
		log.Error("n.checkBalanceAndNotifierIfNeeded cannot get address balance", "error", err)
		return
	}

	if accountBalance.Cmp(n.balanceThreshold) >= 0 {
		n.counter = 0
		n.notified = false

		return
	}

	if !n.notified {
		n.notifyOnTG(accountBalance)
		n.notified = true
	}

	n.counter++
}

func (n *notifier) notifyOnTG(currentBalance *big.Int) {
	message := fmt.Sprintf(`⚠<a href="%s">Hot wallet </a> balance is below threshold (<i>%s</i>).
🚨 Current balance: <b> %s </b>`,
		fmt.Sprintf("https://explorer.elrond.com/accounts/%s", n.address),
		beautifyAmount(n.balanceThreshold.String()),
		beautifyAmount(currentBalance.String()))

	message = url.QueryEscape(message)
	urlreq := fmt.Sprintf(`https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s&parse_mode=HTML&disable_web_page_preview=true`, n.telegramBotKey, n.telegramGroupID, message)

	_, _ = http.Get(urlreq)
}
