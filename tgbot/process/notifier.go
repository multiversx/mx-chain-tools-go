package process

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-tools-go/tgbot/config"
)

var log = logger.GetOrCreate("process")

type notifier struct {
	gatewayURL  string
	explorerURL string

	address            string
	label              string
	balanceThreshold   *big.Int
	checkIntervalInMin int
	notificationStep   int

	telegramBotKey  string
	telegramGroupID string

	notified bool
	counter  int
}

func NewBalanceNotifier(cfg config.BotConfig) (*notifier, error) {
	balanceThreshold, ok := big.NewInt(0).SetString(cfg.General.BalanceThreshold, 10)
	if !ok {
		return nil, errors.New("invalid balance threshold")
	}

	return &notifier{
		explorerURL:        cfg.General.ExplorerUrl,
		gatewayURL:         cfg.General.GatewayURL,
		address:            cfg.General.Address,
		label:              cfg.General.Label,
		balanceThreshold:   balanceThreshold,
		checkIntervalInMin: cfg.General.CheckIntervalInMin,
		notificationStep:   cfg.General.NotificationStep,

		telegramBotKey:  cfg.Telegram.ApiKey,
		telegramGroupID: cfg.Telegram.GroupID,
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
	message := fmt.Sprintf(`⚠<a href="%s"> %s </a> balance is below threshold (<i>%s</i>).
🚨 Current balance: <b> %s </b>`,
		fmt.Sprintf("%s/accounts/%s", n.explorerURL, n.address),
		n.label,
		beautifyAmount(n.balanceThreshold.String()),
		beautifyAmount(currentBalance.String()))

	urlreq := fmt.Sprintf(
		`https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s&parse_mode=HTML&disable_web_page_preview=true`,
		n.telegramBotKey,
		n.telegramGroupID,
		url.QueryEscape(message),
	)

	res, err := http.Get(urlreq)
	if err != nil {
		log.Warn("cannot send message on telegram", "error", err)
	}

	if res.StatusCode >= 400 {
		b, _ := httputil.DumpResponse(res, true)
		log.Warn(string(b))
	}
}
