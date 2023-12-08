package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/multiversx/mx-sdk-go/blockchain"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/urfave/cli"
)

const (
	apiURL = "https://api.multiversx.com"
	shfURL = "https://express-api-shadowfork-four.elrond.ro"

	// Depending on how many transactions one wishes to retrieves. The number of tries needed will increase.
	// See calculateRetryAttempts()
	MINIMUM_NUMBER_OF_RETRIES = 10
	MEDIUM_NUMBER_OF_RETRIES  = 15
	MAXIMUM_NUMBER_OF_RETRIES = 15

	TX_ENDPOINT  = "transactions/%s"
	TXS_ENDPOINT = "transactions?after=%s&size=%d&order=asc"
)

type wrappedTxHashes struct {
	TxHash string `json:"txHash"`
}

type WrappedDifferences struct {
	TxHash      string           `json:"txHash"`
	Differences map[string][]any `json:"differences"`
	Error       string           `json:"error"`
}

type wrappedProxy interface {
	GetHTTP(ctx context.Context, endpoint string) ([]byte, int, error)
}

var (
	//go:embed assets/template.html
	fs embed.FS

	mainProxy wrappedProxy
	shfProxy  wrappedProxy

	retryConfig []retry.Option
)

func main() {
	app := cli.NewApp()
	app.Name = "Accounts Storage Exporter CLI app"
	app.Usage = "This is the entry point for the tool that exports the storage of a given account"
	app.Flags = getFlags()
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return action(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
		return
	}
}

func action(c *cli.Context) error {
	config := getFlagsConfig(c)

	if config.Number%100 != 0 {
		return fmt.Errorf("--number argument must be divisible by 100")
	}

	if config.Number > 10000 {
		return fmt.Errorf("--number argument maxmimum value is 10000")
	}

	retries := calculateRetryAttempts(config.Number)

	//TODO: there is no network provider in mx-sdk-go for api.multiversx.com
	api := blockchain.ArgsProxy{
		ProxyURL:            apiURL,
		Client:              nil,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       false,
		CacheExpirationTime: time.Minute,
		EntityType:          core.Proxy,
	}

	shfArgs := blockchain.ArgsProxy{
		ProxyURL:            shfURL,
		Client:              nil,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       false,
		CacheExpirationTime: time.Minute,
		EntityType:          core.Proxy,
	}

	// Create a client for the mainnet.
	var err error
	mainProxy, err = blockchain.NewProxy(api)
	if err != nil {
		return fmt.Errorf("failed to create proxy: %v", err)
	}

	// Create a client for the shadow-fork.
	shfProxy, err = blockchain.NewProxy(shfArgs)
	if err != nil {
		return fmt.Errorf("failed to create proxy: %v", err)
	}

	// Convert epoch timestamp to actual date.
	timestampTime, err := strconv.ParseInt(config.Timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp: %v", err)
	}
	tm := time.Unix(timestampTime, 0)
	log.Info(fmt.Sprintf("retrieving %d transactions starting from %v", config.Number, tm))

	txHashes := make([]wrappedTxHashes, 0)

	// Retrieve n transactions after a specified timestamp and put them in a slice.
	var allTxsResp []byte
	txsEndpoint := fmt.Sprintf(TXS_ENDPOINT, config.Timestamp, config.Number)
	allTxsResp, _, err = mainProxy.GetHTTP(context.Background(), txsEndpoint)
	if err != nil {
		return fmt.Errorf("failed to retrieve transactions: %v", err)
	}

	err = json.Unmarshal(allTxsResp, &txHashes)
	if err != nil {
		return fmt.Errorf("failed to marshall transactions from mainnet: %v", err)
	}

	retryConfig = []retry.Option{
		retry.OnRetry(func(n uint, err error) {
			log.Info("Retry request", n+1, err)
		}),
		retry.Attempts(retries),
		retry.Delay(5 * time.Second),
	}

	wrappedDiffs := compareTransactions(txHashes)

	jsonBytes, err := json.MarshalIndent(wrappedDiffs, "", " ")
	if err != nil {
		return err
	}

	err = os.WriteFile(config.Outfile, jsonBytes, 0644)
	if err != nil {
		return err
	}

	// Execute the list in the go template.
	tmpl, err := template.ParseFS(fs, "assets/template.html")
	if err != nil {
		panic(err)
	}

	file, _ := os.Create("index.html")
	defer file.Close()

	err = tmpl.Execute(file, wrappedDiffs)
	if err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	log.Info("finished creating report")

	return nil
}

func calculateRetryAttempts(n int) (retriesNo uint) {
	if n >= 1000 && n <= 5000 {
		return MEDIUM_NUMBER_OF_RETRIES
	}

	if n >= 5000 && n <= 10000 {
		return MAXIMUM_NUMBER_OF_RETRIES
	}

	return MINIMUM_NUMBER_OF_RETRIES
}

func compareTransactions(txHashes []wrappedTxHashes) []WrappedDifferences {
	// Iterate over the transactions hashes and fetch all the information contained in either mainnet or shadow-fork
	// and compare each field respectively.
	wg := sync.WaitGroup{}
	wrappedDiffs := make([]WrappedDifferences, len(txHashes))

	for i, t := range txHashes {
		wg.Add(1)
		go func(i int, t string) {
			defer wg.Done()
			var (
				txM *data.TransactionOnNetwork
				txS *data.TransactionOnNetwork
			)

			// Get transaction from shadow-fork in a retry loop.
			err := retry.Do(
				func() error {
					var (
						err error
						wd  *WrappedDifferences
					)
					txM, wd, err = getTransaction(t, "mainnet", mainProxy)
					if err != nil {
						return err
					}

					txS, wd, err = getTransaction(t, "shadow-fork", shfProxy)
					if err != nil {
						return err
					}

					if wd != nil {
						wrappedDiffs[i] = *wd
						return nil
					}

					wrappedDiffs[i] = getDifference(t, *txM, *txS)
					return nil

				}, retryConfig...,
			)

			if err != nil {
				log.Error(err.Error())
			}
		}(i, t.TxHash)
	}

	wg.Wait()
	return wrappedDiffs
}

func getDifference(txHash string, t1, t2 data.TransactionOnNetwork) WrappedDifferences {
	diff := WrappedDifferences{TxHash: txHash}
	diffMap := make(map[string][]any)

	structType := reflect.TypeOf(t1)

	structVal1 := reflect.ValueOf(t1)
	structVal2 := reflect.ValueOf(t2)
	fieldNum := structVal1.NumField()

	for i := 0; i < fieldNum; i++ {
		fieldName := structType.Field(i).Name
		value1 := structVal1.Field(i).Interface()
		value2 := structVal2.Field(i).Interface()

		if !reflect.DeepEqual(value1, value2) {
			slice := []any{value1, value2}
			diffMap[fieldName] = slice
		}
	}

	if len(diffMap) > 0 {
		diff.Differences = diffMap
	}

	return diff
}

func getTransaction(txHash, networkName string, p wrappedProxy) (*data.TransactionOnNetwork, *WrappedDifferences, error) {
	//Retrieve transaction from network.
	resp, code, err := p.GetHTTP(context.Background(), fmt.Sprintf(TX_ENDPOINT, txHash))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get tx %q from %q: %v", txHash, networkName, err)
	}

	wrappedErr := struct {
		Data  interface{} `json:"data"`
		Error string      `json:"error"`
		Code  string      `json:"code"`
	}{}

	// If the status code is different from 200, further investigate.
	if code != http.StatusOK {
		switch code {

		// If the status is 404, we don't want to retry looking for it. It is the only case where we
		// also return a WrappedDifferences struct with the not found error.
		case http.StatusNotFound:
			err = json.Unmarshal(resp, &wrappedErr)
			if err != nil {
				wd := &WrappedDifferences{TxHash: txHash, Error: err.Error()}
				return nil, wd, nil
			}

		// If the status is 429, that means we are sending way too many requests at the moment. We return an error
		// in order to further retry looking for the transaction.
		case http.StatusTooManyRequests:
			tooManyReqErr := errors.New(fmt.Sprintf("too many requests: %q", txHash))
			return nil, nil, tooManyReqErr

		// If the code is something else, we keep looking for the transaction.
		default:
			return nil, nil, errors.New(fmt.Sprintf("got %d while trying to retrieve transaction %q", code, txHash))
		}
	}

	var tx data.TransactionOnNetwork
	err = json.Unmarshal(resp, &tx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshall tx %q from %q: %v", txHash, networkName, err)
	}

	return &tx, nil, nil
}
