package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/multiversx/mx-sdk-go/blockchain"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/urfave/cli"
	"html/template"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"
)

const (
	apiURL = "https://api.multiversx.com"
	shfURL = "https://express-api-shadowfork-four.elrond.ro"
)

type transactions struct {
	TxHash string `json:"txHash"`
}

type Transaction struct {
	TxHash         string `json:"txHash"`
	GasLimit       int    `json:"gasLimit"`
	GasPrice       int    `json:"gasPrice"`
	GasUsed        int    `json:"gasUsed"`
	MiniBlockHash  string `json:"miniBlockHash"`
	Nonce          int    `json:"nonce"`
	Receiver       string `json:"receiver"`
	ReceiverAssets struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	} `json:"receiverAssets"`
	ReceiverShard int    `json:"receiverShard"`
	Round         int    `json:"round"`
	Sender        string `json:"sender"`
	SenderShard   int    `json:"senderShard"`
	Signature     string `json:"signature"`
	Status        string `json:"status"`
	Value         string `json:"value"`
	Fee           string `json:"fee"`
	Timestamp     int    `json:"timestamp"`
	Data          string `json:"data"`
	Function      string `json:"function"`
	Action        struct {
		Category string `json:"category"`
		Name     string `json:"name"`
	} `json:"action"`
	Results []struct {
		Hash         string `json:"hash"`
		Timestamp    int    `json:"timestamp"`
		Nonce        int    `json:"nonce"`
		GasLimit     int    `json:"gasLimit"`
		GasPrice     int    `json:"gasPrice"`
		Value        string `json:"value"`
		Sender       string `json:"sender"`
		Receiver     string `json:"receiver"`
		SenderAssets struct {
			Name string   `json:"name"`
			Tags []string `json:"tags"`
		} `json:"senderAssets"`
		Data           string `json:"data"`
		PrevTxHash     string `json:"prevTxHash"`
		OriginalTxHash string `json:"originalTxHash"`
		CallType       string `json:"callType"`
		MiniBlockHash  string `json:"miniBlockHash"`
		Logs           struct {
			Id      string `json:"id"`
			Address string `json:"address"`
			Events  []struct {
				Identifier string   `json:"identifier"`
				Address    string   `json:"address"`
				Topics     []string `json:"topics"`
				Order      int      `json:"order"`
				Data       string   `json:"data,omitempty"`
			} `json:"events"`
		} `json:"logs"`
	} `json:"results"`
	Price float64 `json:"price"`
	Logs  struct {
		Id            string `json:"id"`
		Address       string `json:"address"`
		AddressAssets struct {
			Name string   `json:"name"`
			Tags []string `json:"tags"`
		} `json:"addressAssets"`
		Events []struct {
			Identifier    string   `json:"identifier"`
			Address       string   `json:"address"`
			Topics        []string `json:"topics"`
			Order         int      `json:"order"`
			AddressAssets struct {
				Name string   `json:"name"`
				Tags []string `json:"tags"`
			} `json:"addressAssets"`
		} `json:"events"`
	} `json:"logs"`
	Operations []struct {
		Id           string `json:"id"`
		Action       string `json:"action"`
		Type         string `json:"type"`
		Sender       string `json:"sender"`
		Receiver     string `json:"receiver"`
		Data         string `json:"data,omitempty"`
		EsdtType     string `json:"esdtType,omitempty"`
		Identifier   string `json:"identifier,omitempty"`
		Ticker       string `json:"ticker,omitempty"`
		Name         string `json:"name,omitempty"`
		Value        string `json:"value,omitempty"`
		Decimals     int    `json:"decimals,omitempty"`
		SvgUrl       string `json:"svgUrl,omitempty"`
		SenderAssets struct {
			Name string   `json:"name"`
			Tags []string `json:"tags"`
		} `json:"senderAssets,omitempty"`
		ValueUSD float64 `json:"valueUSD,omitempty"`
	} `json:"operations"`
}

type WrappedError struct {
	TxHash      string           `json:"txHash"`
	Differences map[string][]any `json:"differences"`
	Error       string           `json:"error"`
}

//go:embed assets/template.html
var fs embed.FS

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
	ep, err := blockchain.NewProxy(api)
	if err != nil {
		return fmt.Errorf("failed to create proxy: %v", err)
	}

	// Create a client for the shadow-fork.
	shf, err := blockchain.NewProxy(shfArgs)
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

	txEndpoint := "transactions/%s"
	wg := sync.WaitGroup{}
	txHashes := make([]transactions, 0)

	// Retrieve n transactions after a specified timestamp and put them in a slice.
	txsEndpoint := fmt.Sprintf("transactions?after=%s&size=%d&order=asc", config.Timestamp, config.Number)
	var allTxsResp []byte
	allTxsResp, _, err = ep.GetHTTP(context.Background(), txsEndpoint)
	if err != nil {
		return fmt.Errorf("failed to retrieve transactions: %v", err)
	}

	err = json.Unmarshal(allTxsResp, &txHashes)
	if err != nil {
		return fmt.Errorf("failed to marshall transactions from mainnet: %v", err)
	}

	retryConfig := []retry.Option{
		retry.OnRetry(func(n uint, err error) {
			log.Info("Retry request", n+1, err)
		}),
		retry.Attempts(retries),
		retry.Delay(5 * time.Second),
	}

	// Iterate over the transactions hashes and fetch all the information contained in either mainnet or shadow-fork
	// and compare each field respectively.
	wrappedErrors := make([]WrappedError, len(txHashes))
	for i, t := range txHashes {
		wg.Add(1)
		go func(i int, t string) {
			var (
				txM  Transaction
				txS  Transaction
				resp []byte
				code int
			)

			// Get transaction from shadow-fork in a retry loop.
			err = retry.Do(
				func() error {
					resp, code, err = shf.GetHTTP(context.Background(), fmt.Sprintf(txEndpoint, t))
					if err != nil {
						wrappedErr := fmt.Errorf("failed to get tx %q from shadow fork: %v", t, err)
						return wrappedErr
					}

					// If the status is different from 200, most probably the transaction could not be found.
					if code != http.StatusOK {
						wrappedErr := struct {
							Data  interface{} `json:"data"`
							Error string      `json:"error"`
							Code  string      `json:"code"`
						}{}

						err = json.Unmarshal(resp, &wrappedErr)
						if err != nil {
							wrappedErrors[i] = WrappedError{TxHash: t, Error: err.Error()}
							return err
						}
						wrappedErrors[i] = WrappedError{TxHash: t, Error: wrappedErr.Error}
						return nil
					}

					return nil
				}, retryConfig...)

			// If there was an error there is no need to perform any other actions. Close the go routine and return.
			if err != nil {
				wg.Done()
				return
			}

			// Marshall the response from the shadow-fork into a variable.
			err = json.Unmarshal(resp, &txS)
			if err != nil {
				wrappedErr := fmt.Errorf("failed to marshall tx %q from shadow fork: %v", t, err)
				wrappedErrors[i] = WrappedError{TxHash: t, Error: wrappedErr.Error()}
				wg.Done()
				return
			}

			// Get transaction from mainnet in a retry loop.
			err = retry.Do(func() error {
				resp, code, err = ep.GetHTTP(context.Background(), fmt.Sprintf(txEndpoint, t))
				if err != nil {
					wrappedErr := fmt.Errorf("failed to get tx %q from mainnet fork: %v", t, err)
					wrappedErrors[i] = WrappedError{TxHash: t, Error: wrappedErr.Error()}
					return wrappedErr
				}

				// If the transaction could not be found add an error.
				if code == http.StatusNotFound {
					wrappedErr := struct {
						Data  interface{} `json:"data"`
						Error string      `json:"error"`
						Code  string      `json:"code"`
					}{}

					err = json.Unmarshal(resp, &wrappedErr)
					if err != nil {
						panic(err)
					}
					wrappedErrors[i] = WrappedError{TxHash: t, Error: wrappedErr.Error}
					return nil
				}

				// If the response code was 429 then try again later as the api cannot handle so many requests.
				if code == http.StatusTooManyRequests {
					tooManyReqErr := errors.New(fmt.Sprintf("too many requests: %q", t))
					wrappedErrors[i] = WrappedError{TxHash: t, Error: tooManyReqErr.Error()}
					return tooManyReqErr
				}

				return nil
			}, retryConfig...)

			// If there was an error there is no need to perform any other actions. Close the go routine and return.
			if err != nil {
				wg.Done()
				return
			}

			// Marshall the response from the shadow-fork into a variable.
			err = json.Unmarshal(resp, &txM)
			if err != nil {
				wrappedErr := fmt.Errorf("failed to get tx %q from mainnet fork: %v", t, err)
				wrappedErrors[i] = WrappedError{TxHash: t, Error: wrappedErr.Error()}
				wg.Done()
				return
			}

			//  Compare differences between the full-transaction on mainnet and on shadowfork.
			diff := getDifference(txM, txS)
			wrappedErrors[i] = diff

			wg.Done()
		}(i, t.TxHash)
	}

	// Wait for all the go routines to finish.
	wg.Wait()

	jsonBytes, err := json.MarshalIndent(wrappedErrors, "", " ")
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

	err = tmpl.Execute(file, wrappedErrors)
	if err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	log.Info("finished creating report")

	return nil
}

func getDifference(t1 Transaction, t2 Transaction) WrappedError {
	diff := WrappedError{TxHash: t1.TxHash}
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

func calculateRetryAttempts(n int) (retriesNo uint) {
	if n >= 1000 && n <= 5000 {
		return 15
	}

	if n >= 5000 && n <= 10000 {
		return 20
	}

	return 10
}
